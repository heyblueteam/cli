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
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// GraphQL request structure
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

// List creation input
type CreateTodoListInput struct {
	ProjectID string  `json:"projectId"`
	Title     string  `json:"title"`
	Position  float64 `json:"position"`
}

// Response structures
type CreatedTodoList struct {
	ID       string  `json:"id"`
	UID      string  `json:"uid"`
	Title    string  `json:"title"`
	Position float64 `json:"position"`
}

type GraphQLResponse struct {
	Data   map[string]CreatedTodoList `json:"data"`
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

// Get current max position for a project
func getMaxPosition(config *Config, projectID string) (float64, error) {
	query := `query GetProjectLists($projectId: String!) {
		todoLists(projectId: $projectId) {
			position
		}
	}`

	reqBody := GraphQLRequest{
		Query: query,
		Variables: map[string]interface{}{
			"projectId": projectID,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return 0, err
	}

	req, err := http.NewRequest("POST", config.APIUrl, bytes.NewBuffer(jsonBody))
	if err != nil {
		return 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Bloo-Token-ID", config.ClientID)
	req.Header.Set("X-Bloo-Token-Secret", config.AuthToken)
	req.Header.Set("X-Bloo-Company-ID", config.CompanyID)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var response struct {
		Data struct {
			TodoLists []struct {
				Position float64 `json:"position"`
			} `json:"todoLists"`
		} `json:"data"`
		Errors []struct {
			Message string `json:"message"`
		} `json:"errors"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return 0, err
	}

	if len(response.Errors) > 0 {
		return 0, fmt.Errorf("GraphQL error: %s", response.Errors[0].Message)
	}

	// Find the max position
	maxPos := 0.0
	for _, list := range response.Data.TodoLists {
		if list.Position > maxPos {
			maxPos = list.Position
		}
	}

	return maxPos, nil
}

// Execute GraphQL mutation to create a single list
func createTodoList(config *Config, input CreateTodoListInput) (*CreatedTodoList, error) {
	mutation := fmt.Sprintf(`
		mutation CreateTodoList {
			createTodoList(input: {
				projectId: "%s"
				title: "%s"
				position: %f
			}) {
				id
				uid
				title
				position
			}
		}
	`, input.ProjectID, input.Title, input.Position)

	reqBody := GraphQLRequest{
		Query: mutation,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	req, err := http.NewRequest("POST", config.APIUrl, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Bloo-Token-ID", config.ClientID)
	req.Header.Set("X-Bloo-Token-Secret", config.AuthToken)
	req.Header.Set("X-Bloo-Company-ID", config.CompanyID)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	var graphQLResp GraphQLResponse
	if err := json.Unmarshal(body, &graphQLResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	if len(graphQLResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", graphQLResp.Errors[0].Message)
	}

	list, ok := graphQLResp.Data["createTodoList"]
	if !ok {
		return nil, fmt.Errorf("no createTodoList in response")
	}

	return &list, nil
}

func main() {
	// Parse command line flags
	projectID := flag.String("project", "", "Project ID (required)")
	names := flag.String("names", "", "Comma-separated list names (required)")
	reverse := flag.Bool("reverse", false, "Create lists in reverse order")
	flag.Parse()

	// Validate required parameters
	if *projectID == "" {
		log.Fatal("Project ID is required. Use -project flag")
	}
	if *names == "" {
		log.Fatal("List names are required. Use -names flag with comma-separated values")
	}

	// Parse list names
	listNames := strings.Split(*names, ",")
	for i := range listNames {
		listNames[i] = strings.TrimSpace(listNames[i])
	}

	// Filter out empty names
	var validNames []string
	for _, name := range listNames {
		if name != "" {
			validNames = append(validNames, name)
		}
	}

	if len(validNames) == 0 {
		log.Fatal("No valid list names provided")
	}

	// Load configuration
	config, err := loadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Get current max position
	fmt.Printf("Getting current lists in project %s...\n", *projectID)
	maxPos, err := getMaxPosition(config, *projectID)
	if err != nil {
		log.Fatalf("Failed to get max position: %v", err)
	}

	// Calculate positions for new lists
	// Standard increment is 65535.0 as per Blue's implementation
	increment := 65535.0
	startPos := maxPos + increment

	// Reverse the order if requested
	if *reverse {
		for i, j := 0, len(validNames)-1; i < j; i, j = i+1, j-1 {
			validNames[i], validNames[j] = validNames[j], validNames[i]
		}
	}

	// Create lists
	fmt.Printf("\nCreating %d lists...\n", len(validNames))
	var createdLists []*CreatedTodoList

	for i, name := range validNames {
		position := startPos + (float64(i) * increment)
		
		input := CreateTodoListInput{
			ProjectID: *projectID,
			Title:     name,
			Position:  position,
		}

		fmt.Printf("Creating list '%s' at position %.0f...\n", name, position)
		
		list, err := createTodoList(config, input)
		if err != nil {
			log.Printf("Failed to create list '%s': %v", name, err)
			continue
		}

		createdLists = append(createdLists, list)
		fmt.Printf("✅ Created list '%s' (ID: %s)\n", list.Title, list.ID)
	}

	// Summary
	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Successfully created %d out of %d lists\n", len(createdLists), len(validNames))
	
	if len(createdLists) > 0 {
		fmt.Printf("\nCreated lists:\n")
		for i, list := range createdLists {
			fmt.Printf("%d. %s (ID: %s, Position: %.0f)\n", i+1, list.Title, list.ID, list.Position)
		}
		
		fmt.Printf("\nYou can now add records to these lists using:\n")
		fmt.Printf("  go run create-records.go -list %s -records \"Task 1,Task 2,Task 3\"\n", createdLists[0].ID)
	}
}