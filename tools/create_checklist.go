package tools

import (
	"flag"
	"fmt"

	. "demo-builder/common"
)

// Checklist represents a checklist structure
type Checklist struct {
	ID        string `json:"id"`
	UID       string `json:"uid"`
	Title     string `json:"title"`
	Position  float64 `json:"position"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	CreatedBy User   `json:"createdBy"`
}

// CreateChecklistInput represents the input for creating a checklist
type CreateChecklistInput struct {
	TodoID   string  `json:"todoId"`
	Title    string  `json:"title"`
	Position float64 `json:"position"`
}

// CreateChecklistResponse represents the response from creating a checklist
type CreateChecklistResponse struct {
	CreateChecklist Checklist `json:"createChecklist"`
}

// Execute GraphQL mutation to create a checklist
func executeCreateChecklist(client *Client, input CreateChecklistInput) (*Checklist, error) {
	// Build the mutation
	mutation := `
		mutation CreateChecklist($input: CreateChecklistInput!) {
			createChecklist(input: $input) {
				id
				uid
				title
				position
				createdAt
				updatedAt
				createdBy {
					id
					uid
					fullName
					email
				}
			}
		}
	`

	// Prepare variables
	variables := map[string]interface{}{
		"input": input,
	}

	// Execute mutation
	var response CreateChecklistResponse
	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return nil, err
	}

	return &response.CreateChecklist, nil
}

// RunCreateChecklist handles the create-checklist command
func RunCreateChecklist(args []string) error {
	// Define flags
	fs := flag.NewFlagSet("create-checklist", flag.ExitOnError)
	recordID := fs.String("record", "", "Record/Todo ID to add checklist to (required)")
	title := fs.String("title", "", "Checklist title (required)")
	position := fs.Float64("position", 1000.0, "Position of the checklist (default: 1000.0)")
	projectID := fs.String("project", "", "Project ID or slug (optional - for context)")
	simple := fs.Bool("simple", false, "Show simple output")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Validate required flags
	if *recordID == "" {
		return fmt.Errorf("record ID is required")
	}
	if *title == "" {
		return fmt.Errorf("checklist title is required")
	}

	// Load config and create client
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	client := NewClient(config)

	// Set project context if provided
	if *projectID != "" {
		client.SetProject(*projectID)
	}

	// Prepare checklist input
	input := CreateChecklistInput{
		TodoID:   *recordID,
		Title:    *title,
		Position: *position,
	}

	// Display operation details
	if !*simple {
		fmt.Printf("=== Creating Checklist ===\n")
		fmt.Printf("Record ID: %s\n", *recordID)
		fmt.Printf("Title: %s\n", *title)
		fmt.Printf("Position: %.1f\n", *position)
		if *projectID != "" {
			fmt.Printf("Project: %s\n", *projectID)
		}
		fmt.Printf("\n")
	}

	// Execute checklist creation
	checklist, err := executeCreateChecklist(client, input)
	if err != nil {
		return fmt.Errorf("failed to create checklist: %v", err)
	}

	// Display results
	if *simple {
		fmt.Printf("Checklist ID: %s\n", checklist.ID)
	} else {
		fmt.Printf("=== Checklist Created Successfully ===\n")
		fmt.Printf("ID: %s\n", checklist.ID)
		fmt.Printf("UID: %s\n", checklist.UID)
		fmt.Printf("Title: %s\n", checklist.Title)
		fmt.Printf("Position: %.1f\n", checklist.Position)
		fmt.Printf("Created: %s\n", checklist.CreatedAt)
		fmt.Printf("Created By: %s (%s)\n", checklist.CreatedBy.FullName, checklist.CreatedBy.Email)
		fmt.Printf("✅ Checklist created successfully!\n")
	}

	return nil
}
