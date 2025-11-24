package tools

import (
	"flag"
	"fmt"

	. "demo-builder/common"
)

// ChecklistItem represents a checklist item structure
type ChecklistItem struct {
	ID        string  `json:"id"`
	UID       string  `json:"uid"`
	Title     string  `json:"title"`
	Position  float64 `json:"position"`
	Done      bool    `json:"done"`
	StartedAt *string `json:"startedAt"`
	DuedAt    *string `json:"duedAt"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
	CreatedBy User    `json:"createdBy"`
}

// CreateChecklistItemInput represents the input for creating a checklist item
type CreateChecklistItemInput struct {
	ChecklistID string  `json:"checklistId"`
	Title       string  `json:"title"`
	Position    float64 `json:"position"`
}

// CreateChecklistItemResponse represents the response from creating a checklist item
type CreateChecklistItemResponse struct {
	CreateChecklistItem ChecklistItem `json:"createChecklistItem"`
}

// Execute GraphQL mutation to create a checklist item
func executeCreateChecklistItem(client *Client, input CreateChecklistItemInput) (*ChecklistItem, error) {
	// Build the mutation
	mutation := `
		mutation CreateChecklistItem($input: CreateChecklistItemInput!) {
			createChecklistItem(input: $input) {
				id
				uid
				title
				position
				done
				startedAt
				duedAt
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
	var response CreateChecklistItemResponse
	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return nil, err
	}

	return &response.CreateChecklistItem, nil
}

// RunCreateChecklistItem handles the create-checklist-item command
func RunCreateChecklistItem(args []string) error {
	// Define flags
	fs := flag.NewFlagSet("create-checklist-item", flag.ExitOnError)
	checklistID := fs.String("checklist", "", "Checklist ID to add item to (required)")
	title := fs.String("title", "", "Checklist item title (required)")
	position := fs.Float64("position", 1000.0, "Position of the checklist item (default: 1000.0)")
	projectID := fs.String("project", "", "Project ID or slug (optional - for context)")
	simple := fs.Bool("simple", false, "Show simple output")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Validate required flags
	if *checklistID == "" {
		return fmt.Errorf("checklist ID is required")
	}
	if *title == "" {
		return fmt.Errorf("checklist item title is required")
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

	// Prepare checklist item input
	input := CreateChecklistItemInput{
		ChecklistID: *checklistID,
		Title:       *title,
		Position:    *position,
	}

	// Display operation details
	if !*simple {
		fmt.Printf("=== Creating Checklist Item ===\n")
		fmt.Printf("Checklist ID: %s\n", *checklistID)
		fmt.Printf("Title: %s\n", *title)
		fmt.Printf("Position: %.1f\n", *position)
		if *projectID != "" {
			fmt.Printf("Project: %s\n", *projectID)
		}
		fmt.Printf("\n")
	}

	// Execute checklist item creation
	item, err := executeCreateChecklistItem(client, input)
	if err != nil {
		return fmt.Errorf("failed to create checklist item: %v", err)
	}

	// Display results
	if *simple {
		fmt.Printf("Checklist Item ID: %s\n", item.ID)
	} else {
		fmt.Printf("=== Checklist Item Created Successfully ===\n")
		fmt.Printf("ID: %s\n", item.ID)
		fmt.Printf("UID: %s\n", item.UID)
		fmt.Printf("Title: %s\n", item.Title)
		fmt.Printf("Position: %.1f\n", item.Position)
		fmt.Printf("Done: %t\n", item.Done)
		if item.StartedAt != nil {
			fmt.Printf("Started: %s\n", *item.StartedAt)
		}
		if item.DuedAt != nil {
			fmt.Printf("Due: %s\n", *item.DuedAt)
		}
		fmt.Printf("Created: %s\n", item.CreatedAt)
		fmt.Printf("Created By: %s (%s)\n", item.CreatedBy.FullName, item.CreatedBy.Email)
		fmt.Printf("✅ Checklist item created successfully!\n")
	}

	return nil
}
