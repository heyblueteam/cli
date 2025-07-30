package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
)

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

type CreateTodoListResponse struct {
	CreateTodoList CreatedTodoList `json:"createTodoList"`
}

type MaxPositionResponse struct {
	TodoLists []struct {
		Position float64 `json:"position"`
	} `json:"todoLists"`
}

// Get current max position for a project
func getMaxPosition(client *Client, projectID string) (float64, error) {
	query := `query GetProjectLists($projectId: String!) {
		todoLists(projectId: $projectId) {
			position
		}
	}`

	variables := map[string]interface{}{
		"projectId": projectID,
	}

	var response MaxPositionResponse
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return 0, err
	}

	// Find the max position
	maxPos := 0.0
	for _, list := range response.TodoLists {
		if list.Position > maxPos {
			maxPos = list.Position
		}
	}

	return maxPos, nil
}

// Execute GraphQL mutation to create a single list
func createTodoList(client *Client, input CreateTodoListInput) (*CreatedTodoList, error) {
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

	var response CreateTodoListResponse
	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		return nil, err
	}

	return &response.CreateTodoList, nil
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
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create client
	client := NewClient(config)

	// Get current max position
	fmt.Printf("Getting current lists in project %s...\n", *projectID)
	maxPos, err := getMaxPosition(client, *projectID)
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
		
		list, err := createTodoList(client, input)
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