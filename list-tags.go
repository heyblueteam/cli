package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
)

// Tag represents a tag in the system
type Tag struct {
	ID        string `json:"id"`
	UID       string `json:"uid"`
	Title     string `json:"title"`
	Color     string `json:"color"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// GraphQL request/response structures
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   json.RawMessage `json:"data,omitempty"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors,omitempty"`
}

// BlueAPIClient implements GraphQL requests to Blue API
type BlueAPIClient struct {
	url   string
	token string
	http  *http.Client
}

func NewBlueAPIClient(url, token string) *BlueAPIClient {
	return &BlueAPIClient{
		url:   url,
		token: token,
		http:  &http.Client{},
	}
}

func (c *BlueAPIClient) Query(ctx context.Context, query string, variables map[string]interface{}) (json.RawMessage, error) {
	// Create request
	reqBody := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Bloo-Token-ID", os.Getenv("CLIENT_ID"))
	req.Header.Set("X-Bloo-Token-Secret", c.token)
	req.Header.Set("X-Bloo-Company-ID", os.Getenv("COMPANY_ID"))

	// Execute request
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var graphQLResp GraphQLResponse
	if err := json.Unmarshal(body, &graphQLResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check for errors
	if len(graphQLResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", graphQLResp.Errors[0].Message)
	}

	return graphQLResp.Data, nil
}

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	apiURL := os.Getenv("API_URL")
	authToken := os.Getenv("AUTH_TOKEN")

	if apiURL == "" || authToken == "" {
		log.Fatal("API_URL and AUTH_TOKEN must be set in .env file")
	}

	// Create API client
	client := NewBlueAPIClient(apiURL, authToken)

	// GraphQL query for listing tags
	projectID := "cmdpb3mfj247psf2jj8x5yg3g"
	
	query := `
		query ListTags {
			tagList(
				filter: { 
					projectIds: ["` + projectID + `"] 
				}
				first: 50
				orderBy: title_ASC
			) {
				items {
					id
					uid
					title
					color
					createdAt
					updatedAt
				}
				totalCount
			}
		}
	`

	// Variables
	variables := map[string]interface{}{}

	// Execute query
	ctx := context.Background()
	fmt.Printf("Listing tags for project ID: %s\n", projectID)
	fmt.Println("==========================================")

	data, err := client.Query(ctx, query, variables)
	if err != nil {
		log.Fatalf("Failed to query tags: %v", err)
	}

	// Parse response
	var response struct {
		TagList struct {
			Items      []Tag `json:"items"`
			TotalCount int   `json:"totalCount"`
		} `json:"tagList"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		log.Fatalf("Failed to parse response: %v", err)
	}

	// Display results
	fmt.Printf("\nFound %d tags (Total: %d):\n", len(response.TagList.Items), response.TagList.TotalCount)
	fmt.Println("------------------------------------------")

	if len(response.TagList.Items) == 0 {
		fmt.Println("No tags found for this project.")
	} else {
		for i, tag := range response.TagList.Items {
			fmt.Printf("%d. %s\n", i+1, tag.Title)
			fmt.Printf("   ID: %s\n", tag.ID)
			fmt.Printf("   UID: %s\n", tag.UID)
			fmt.Printf("   Color: %s\n", tag.Color)
			fmt.Println()
		}
	}
}