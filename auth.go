package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds API configuration
type Config struct {
	APIUrl    string
	AuthToken string
	ClientID  string
	CompanyID string
}

// GraphQLRequest represents a GraphQL request
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// GraphQLError represents a GraphQL error
type GraphQLError struct {
	Message    string                 `json:"message"`
	Extensions map[string]interface{} `json:"extensions,omitempty"`
}

// Client handles Blue API communication
type Client struct {
	config        *Config
	httpClient    *http.Client
	projectID     string
	projectSlug   string
}

// LoadConfig loads configuration from .env file
func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("error loading .env file: %w", err)
	}

	config := &Config{
		APIUrl:    os.Getenv("API_URL"),
		AuthToken: os.Getenv("AUTH_TOKEN"),
		ClientID:  os.Getenv("CLIENT_ID"),
		CompanyID: os.Getenv("COMPANY_ID"),
	}

	if config.APIUrl == "" || config.AuthToken == "" || config.ClientID == "" || config.CompanyID == "" {
		return nil, fmt.Errorf("missing required environment variables")
	}

	return config, nil
}

// NewClient creates a new Blue API client
func NewClient(config *Config) *Client {
	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ExecuteQuery executes a GraphQL query and returns the raw response
func (c *Client) ExecuteQuery(query string, variables map[string]interface{}) (map[string]interface{}, error) {
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", c.config.APIUrl, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Bloo-Token-ID", c.config.ClientID)
	req.Header.Set("X-Bloo-Token-Secret", c.config.AuthToken)
	req.Header.Set("X-Bloo-Company-ID", c.config.CompanyID)
	
	// Include project context header if project context is set
	// Use Project ID if available, otherwise use Project slug
	if c.projectID != "" {
		req.Header.Set("X-Bloo-Project-Id", c.projectID)
	} else if c.projectSlug != "" {
		req.Header.Set("X-Bloo-Project-Id", c.projectSlug)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var response struct {
		Data   map[string]interface{} `json:"data"`
		Errors []GraphQLError         `json:"errors"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	if len(response.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", response.Errors[0].Message)
	}

	if response.Data == nil {
		return nil, fmt.Errorf("no data in response")
	}

	return response.Data, nil
}

// ExecuteQueryWithResult executes a GraphQL query and unmarshals the result
func (c *Client) ExecuteQueryWithResult(query string, variables map[string]interface{}, result interface{}) error {
	data, err := c.ExecuteQuery(query, variables)
	if err != nil {
		return err
	}

	// Marshal and unmarshal to convert to the result type
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("error marshaling data: %w", err)
	}

	if err := json.Unmarshal(jsonData, result); err != nil {
		return fmt.Errorf("error unmarshaling result: %w", err)
	}

	return nil
}

// SetProjectID sets the project ID for requests that require project context
func (c *Client) SetProjectID(projectID string) {
	c.projectID = projectID
	c.projectSlug = "" // Clear slug when setting ID
}

// SetProjectSlug sets the project slug for requests that require project context
func (c *Client) SetProjectSlug(projectSlug string) {
	c.projectSlug = projectSlug
	c.projectID = "" // Clear ID when setting slug
}

// SetProject sets the project ID or slug for requests that require project context
// This method automatically detects whether the input is an ID or slug
func (c *Client) SetProject(project string) {
	// Simple heuristic: if it looks like a UUID/ID (contains hyphens), treat as ID
	// Otherwise treat as slug
	if len(project) > 20 && (project[8] == '-' || project[13] == '-' || project[18] == '-') {
		c.SetProjectID(project)
	} else {
		c.SetProjectSlug(project)
	}
}

// GetProjectID returns the current project ID
func (c *Client) GetProjectID() string {
	return c.projectID
}

// GetProjectSlug returns the current project slug
func (c *Client) GetProjectSlug() string {
	return c.projectSlug
}

// GetProjectContext returns the current project context (ID or slug)
func (c *Client) GetProjectContext() string {
	if c.projectID != "" {
		return c.projectID
	}
	return c.projectSlug
}

// GetCompanyID returns the configured company ID
func (c *Client) GetCompanyID() string {
	return c.config.CompanyID
}