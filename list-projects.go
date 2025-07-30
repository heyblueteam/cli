package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// GraphQL request structure
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// Response structures
type Project struct {
	ID          string    `json:"id"`
	UID         string    `json:"uid"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Archived    bool      `json:"archived"`
	Color       string    `json:"color"`
	Icon        string    `json:"icon"`
	CreatedAt   string    `json:"createdAt"`
	UpdatedAt   string    `json:"updatedAt"`
	Position    float64   `json:"position"`
	IsTemplate  bool      `json:"isTemplate"`
}

type PageInfo struct {
	TotalPages      int  `json:"totalPages"`
	TotalItems      int  `json:"totalItems"`
	Page            int  `json:"page"`
	PerPage         int  `json:"perPage"`
	HasNextPage     bool `json:"hasNextPage"`
	HasPreviousPage bool `json:"hasPreviousPage"`
}

type ProjectList struct {
	Items    []Project `json:"items"`
	PageInfo PageInfo  `json:"pageInfo"`
}

type GraphQLResponse struct {
	Data   map[string]ProjectList `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// Config holds API configuration
type Config struct {
	APIUrl    string
	AuthToken string
	ClientID  string
	CompanyID string
}

// Load configuration from .env
func loadConfig() (*Config, error) {
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

// Execute GraphQL query
func executeQuery(config *Config, query string) (*GraphQLResponse, error) {
	// Create request body
	reqBody := GraphQLRequest{
		Query: query,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", config.APIUrl, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Bloo-Token-ID", config.ClientID)
	req.Header.Set("X-Bloo-Token-Secret", config.AuthToken)
	req.Header.Set("X-Bloo-Company-ID", config.CompanyID)

	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Parse response
	var graphQLResp GraphQLResponse
	if err := json.Unmarshal(body, &graphQLResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	// Check for GraphQL errors
	if len(graphQLResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", graphQLResp.Errors[0].Message)
	}

	return &graphQLResp, nil
}

// Queries
const (
	fullQuery = `query ProjectListQuery {
		projectList(filter: { companyIds: ["%s"] }) {
			items {
				id
				uid
				slug
				name
				description
				archived
				color
				icon
				createdAt
				updatedAt
				position
				isTemplate
			}
			pageInfo {
				totalPages
				totalItems
				page
				perPage
				hasNextPage
				hasPreviousPage
			}
		}
	}`

	simpleQuery = `query ProjectListQuery {
		projectList(filter: { companyIds: ["%s"] }) {
			items {
				id
				name
			}
			pageInfo {
				totalItems
			}
		}
	}`
)

func main() {
	// Parse command line flags
	simple := flag.Bool("simple", false, "Show only project names and IDs")
	flag.Parse()

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Select query based on flag
	var query string
	if *simple {
		query = fmt.Sprintf(simpleQuery, config.CompanyID)
	} else {
		query = fmt.Sprintf(fullQuery, config.CompanyID)
	}

	// Execute query
	response, err := executeQuery(config, query)
	if err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}

	// Display results
	projectList, ok := response.Data["projectList"]
	if !ok {
		log.Fatal("No project list in response")
	}

	fmt.Printf("\n=== Projects in %s ===\n", config.CompanyID)
	fmt.Printf("Total projects: %d\n\n", projectList.PageInfo.TotalItems)

	if *simple {
		// Simple output
		for i, project := range projectList.Items {
			fmt.Printf("%d. %s\n   ID: %s\n\n", i+1, project.Name, project.ID)
		}
	} else {
		// Detailed output
		for i, project := range projectList.Items {
			fmt.Printf("%d. %s\n", i+1, project.Name)
			fmt.Printf("   ID: %s\n", project.ID)
			fmt.Printf("   Slug: %s\n", project.Slug)
			fmt.Printf("   Archived: %v\n", project.Archived)
			fmt.Printf("   Template: %v\n", project.IsTemplate)
			if project.Description != "" {
				fmt.Printf("   Description: %s\n", project.Description)
			}
			if project.Color != "" {
				fmt.Printf("   Color: %s\n", project.Color)
			}
			if project.Icon != "" {
				fmt.Printf("   Icon: %s\n", project.Icon)
			}
			fmt.Printf("   Created: %s\n", project.CreatedAt)
			fmt.Printf("   Updated: %s\n", project.UpdatedAt)
			fmt.Println()
		}
	}

	// Show pagination info if there are more pages
	if projectList.PageInfo.HasNextPage {
		fmt.Printf("\nNote: Showing page %d of %d. Total items: %d\n", 
			projectList.PageInfo.Page, 
			projectList.PageInfo.TotalPages, 
			projectList.PageInfo.TotalItems)
	}
}