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

// List structures
type TodoList struct {
	ID               string  `json:"id"`
	UID              string  `json:"uid"`
	Title            string  `json:"title"`
	Position         float64 `json:"position"`
	TodosCount       int     `json:"todosCount"`
	TodosMaxPosition float64 `json:"todosMaxPosition"`
	CreatedAt        string  `json:"createdAt"`
	UpdatedAt        string  `json:"updatedAt"`
	IsDisabled       bool    `json:"isDisabled"`
	IsLocked         bool    `json:"isLocked"`
	Completed        bool    `json:"completed"`
	Editable         bool    `json:"editable"`
	Deletable        bool    `json:"deletable"`
}

type GraphQLResponse struct {
	Data   map[string]interface{} `json:"data"`
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

// Parse lists from response
func parseLists(data map[string]interface{}) ([]TodoList, error) {
	listsData, ok := data["todoLists"]
	if !ok {
		return nil, fmt.Errorf("no todoLists data in response")
	}

	// Marshal and unmarshal to convert to our struct
	jsonData, err := json.Marshal(listsData)
	if err != nil {
		return nil, err
	}

	var lists []TodoList
	if err := json.Unmarshal(jsonData, &lists); err != nil {
		return nil, err
	}

	return lists, nil
}

// Queries
const (
	detailedQuery = `query GetProjectLists($projectId: String!) {
		todoLists(projectId: $projectId) {
			id
			uid
			title
			position
			createdAt
			updatedAt
			isDisabled
			isLocked
			completed
			editable
			deletable
			todosCount
			todosMaxPosition
		}
	}`

	simpleQuery = `query GetProjectLists($projectId: String!) {
		todoLists(projectId: $projectId) {
			id
			title
			position
			todosCount
		}
	}`
)

func main() {
	// Parse command line flags
	projectID := flag.String("project", "", "Project ID (required)")
	simple := flag.Bool("simple", false, "Show only basic list information")
	flag.Parse()

	// Validate required parameters
	if *projectID == "" {
		log.Fatal("Project ID is required. Use -project flag")
	}

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Select query based on flag
	var query string
	if *simple {
		query = simpleQuery
	} else {
		query = detailedQuery
	}

	// Build request with variables
	reqBody := GraphQLRequest{
		Query: query,
		Variables: map[string]interface{}{
			"projectId": *projectID,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		log.Fatalf("Error marshaling request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", config.APIUrl, bytes.NewBuffer(jsonBody))
	if err != nil {
		log.Fatalf("Error creating request: %v", err)
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
		log.Fatalf("Error executing request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response: %v", err)
	}

	// Parse response
	var graphQLResp GraphQLResponse
	if err := json.Unmarshal(body, &graphQLResp); err != nil {
		log.Fatalf("Error parsing response: %v", err)
	}

	// Check for GraphQL errors
	if len(graphQLResp.Errors) > 0 {
		log.Fatalf("GraphQL error: %s", graphQLResp.Errors[0].Message)
	}

	// Parse lists data
	lists, err := parseLists(graphQLResp.Data)
	if err != nil {
		log.Fatalf("Failed to parse lists: %v", err)
	}

	// Display results
	fmt.Printf("\n=== Lists in Project %s ===\n", *projectID)
	fmt.Printf("Total lists: %d\n\n", len(lists))

	if len(lists) == 0 {
		fmt.Println("No lists found in this project.")
		fmt.Printf("\nCreate lists using:\n")
		fmt.Printf("  go run create-list.go -project %s -names \"To Do,In Progress,Done\"\n", *projectID)
		return
	}

	// Sort lists by position
	for i, list := range lists {
		if *simple {
			// Simple output
			fmt.Printf("%d. %s\n", i+1, list.Title)
			fmt.Printf("   ID: %s\n", list.ID)
			fmt.Printf("   Position: %.0f\n", list.Position)
			fmt.Printf("   Tasks: %d\n\n", list.TodosCount)
		} else {
			// Detailed output
			fmt.Printf("%d. %s\n", i+1, list.Title)
			fmt.Printf("   ID: %s\n", list.ID)
			fmt.Printf("   UID: %s\n", list.UID)
			fmt.Printf("   Position: %.0f\n", list.Position)
			fmt.Printf("   Total tasks: %d\n", list.TodosCount)
			fmt.Printf("   Max position: %.0f\n", list.TodosMaxPosition)
			fmt.Printf("   Disabled: %v\n", list.IsDisabled)
			fmt.Printf("   Locked: %v\n", list.IsLocked)
			fmt.Printf("   Completed: %v\n", list.Completed)
			fmt.Printf("   Editable: %v\n", list.Editable)
			fmt.Printf("   Deletable: %v\n", list.Deletable)
			fmt.Printf("   Created: %s\n", list.CreatedAt)
			fmt.Printf("   Updated: %s\n", list.UpdatedAt)
			fmt.Println()
		}
	}
}