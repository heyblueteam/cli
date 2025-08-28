package main

import (
	"flag"
	"fmt"
	"log"
)

// CountResponse represents the response from the GraphQL query
type CountResponse struct {
	Todos struct {
		TotalCount int `json:"totalCount"`
	} `json:"todos"`
}

func main() {
	// Parse command line flags
	projectID := flag.String("project", "", "Project ID to count records (required)")
	todoListID := flag.String("list", "", "Todo List ID to filter records (optional)")
	done := flag.String("done", "", "Filter by completion status (true/false, optional)")
	archived := flag.String("archived", "", "Filter by archived status (true/false, optional)")
	flag.Parse()

	// Validate required parameters
	if *projectID == "" {
		log.Fatal("Project ID is required. Use -project flag.")
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create client
	client := NewClient(config)

	// Build the GraphQL query - using deprecated todos query which has totalCount
	query := `
		query CountRecords($filter: TodosFilter!) {
			todos(filter: $filter, first: 1) {
				totalCount
			}
		}
	`

	// Build filter variables
	filter := make(map[string]interface{})
	
	// companyIds is required for TodosFilter
	filter["companyIds"] = []string{}
	
	// Add project filter
	filter["projectIds"] = []string{*projectID}
	
	// Add optional filters
	if *todoListID != "" {
		filter["todoListIds"] = []string{*todoListID}
	}
	
	if *done != "" {
		if *done == "true" {
			filter["done"] = true
		} else if *done == "false" {
			filter["done"] = false
		}
	}
	
	if *archived != "" {
		if *archived == "true" {
			filter["archived"] = true
		} else if *archived == "false" {
			filter["archived"] = false
		}
	}

	// Build variables
	variables := map[string]interface{}{
		"filter": filter,
	}

	// Execute query
	var response CountResponse
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}

	// Display result
	totalCount := response.Todos.TotalCount
	
	fmt.Printf("\n=== Record Count ===\n")
	fmt.Printf("Project ID: %s\n", *projectID)
	if *todoListID != "" {
		fmt.Printf("List ID: %s\n", *todoListID)
	}
	if *done != "" {
		fmt.Printf("Completion filter: %s\n", *done)
	}
	if *archived != "" {
		fmt.Printf("Archived filter: %s\n", *archived)
	}
	fmt.Printf("\nTotal Records: %d\n", totalCount)
}