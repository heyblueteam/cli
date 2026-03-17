package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/joho/godotenv"
)

// Default API URL
const DefaultAPIUrl = "https://api.blue.cc/graphql"

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

// ConfigDir returns the path to the Blue CLI config directory
func ConfigDir() string {
	// Check XDG_CONFIG_HOME first
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "blue")
	}
	// Fall back to ~/.config/blue
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "blue")
}

// ConfigPath returns the path to the Blue CLI config file
func ConfigPath() string {
	dir := ConfigDir()
	if dir == "" {
		return ""
	}
	return filepath.Join(dir, "config.env")
}

// LoadConfig loads configuration with the following priority:
// 1. Environment variables (always win)
// 2. .env file in current directory
// 3. ~/.config/blue/config.env (global config from "blue init")
func LoadConfig() (*Config, error) {
	// Try loading .env from current directory (silently ignore if not found)
	_ = godotenv.Load()

	// Try loading global config (silently ignore if not found)
	globalConfig := ConfigPath()
	if globalConfig != "" {
		_ = godotenv.Load(globalConfig)
	}

	config := &Config{
		APIUrl:    os.Getenv("API_URL"),
		AuthToken: os.Getenv("AUTH_TOKEN"),
		ClientID:  os.Getenv("CLIENT_ID"),
		CompanyID: os.Getenv("COMPANY_ID"),
	}

	// Default API URL
	if config.APIUrl == "" {
		config.APIUrl = DefaultAPIUrl
	}

	if config.AuthToken == "" || config.ClientID == "" || config.CompanyID == "" {
		return nil, fmt.Errorf("missing credentials. Run 'blue init' to set up your configuration")
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
	c.projectSlug = ""
}

// SetProjectSlug sets the project slug for requests that require project context
func (c *Client) SetProjectSlug(projectSlug string) {
	c.projectSlug = projectSlug
	c.projectID = ""
}

// SetProject sets the project ID or slug for requests that require project context
func (c *Client) SetProject(project string) {
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

// GetCompanyID returns the configured company ID (slug)
func (c *Client) GetCompanyID() string {
	return c.config.CompanyID
}

// ResolveCompanyID resolves the company slug to the actual company ID (cuid).
// Some GraphQL queries (like dashboards) require the real ID, not the slug.
func (c *Client) ResolveCompanyID() (string, error) {
	query := fmt.Sprintf(`query { company(id: "%s") { id } }`, c.config.CompanyID)
	data, err := c.ExecuteQuery(query, nil)
	if err != nil {
		return "", fmt.Errorf("failed to resolve company ID: %w", err)
	}
	if company, ok := data["company"].(map[string]interface{}); ok {
		if id, ok := company["id"].(string); ok {
			return id, nil
		}
	}
	return "", fmt.Errorf("could not resolve company ID from slug '%s'", c.config.CompanyID)
}

// DownloadFile downloads a file from the given URL using the authenticated client
func (c *Client) DownloadFile(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("X-Bloo-Token-ID", c.config.ClientID)
	req.Header.Set("X-Bloo-Token-Secret", c.config.AuthToken)
	req.Header.Set("X-Bloo-Company-ID", c.config.CompanyID)

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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	return data, nil
}
