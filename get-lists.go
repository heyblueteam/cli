package main

import (
	"flag"
	"fmt"
	"log"
)

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

type TodoListsResponse struct {
	TodoLists []TodoList `json:"todoLists"`
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
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create client
	client := NewClient(config)

	// Select query based on flag
	var query string
	if *simple {
		query = simpleQuery
	} else {
		query = detailedQuery
	}

	// Build variables
	variables := map[string]interface{}{
		"projectId": *projectID,
	}

	// Execute query
	var response TodoListsResponse
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}

	// Get lists
	lists := response.TodoLists

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