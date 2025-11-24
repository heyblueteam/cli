package tools

import (
	"flag"
	"fmt"

	. "demo-builder/common"
)

// EditChecklistItemInput represents the input for editing a checklist item
type EditChecklistItemInput struct {
	ChecklistItemID string   `json:"checklistItemId"`
	ChecklistID     *string  `json:"checklistId,omitempty"`
	Title           *string  `json:"title,omitempty"`
	Position        *float64 `json:"position,omitempty"`
	Done            *bool    `json:"done,omitempty"`
}

// EditChecklistItemResponse represents the response from editing a checklist item
type EditChecklistItemResponse struct {
	EditChecklistItem ChecklistItem `json:"editChecklistItem"`
}

// Execute GraphQL mutation to update a checklist item
func executeUpdateChecklistItem(client *Client, input EditChecklistItemInput) (*ChecklistItem, error) {
	// Build the mutation
	mutation := `
		mutation EditChecklistItem($input: EditChecklistItemInput!) {
			editChecklistItem(input: $input) {
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
	var response EditChecklistItemResponse
	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return nil, err
	}

	return &response.EditChecklistItem, nil
}

// RunUpdateChecklistItem handles the update-checklist-item command
func RunUpdateChecklistItem(args []string) error {
	// Define flags
	fs := flag.NewFlagSet("update-checklist-item", flag.ExitOnError)
	itemID := fs.String("item", "", "Checklist item ID to update (required)")
	title := fs.String("title", "", "New title for the checklist item")
	position := fs.Float64("position", -1, "New position for the checklist item")
	checklistID := fs.String("move-to-checklist", "", "Move item to a different checklist (checklist ID)")
	done := fs.String("done", "", "Mark item as done (true/false)")
	projectID := fs.String("project", "", "Project ID or slug (optional - for context)")
	simple := fs.Bool("simple", false, "Show simple output")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Validate required flags
	if *itemID == "" {
		return fmt.Errorf("checklist item ID is required")
	}

	// Check if at least one update field is provided
	if *title == "" && *position == -1 && *checklistID == "" && *done == "" {
		return fmt.Errorf("at least one field to update must be provided (title, position, move-to-checklist, or done)")
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

	// Prepare update input
	input := EditChecklistItemInput{
		ChecklistItemID: *itemID,
	}

	// Add optional fields if provided
	if *title != "" {
		input.Title = title
	}
	if *position != -1 {
		input.Position = position
	}
	if *checklistID != "" {
		input.ChecklistID = checklistID
	}
	if *done != "" {
		if *done == "true" {
			doneVal := true
			input.Done = &doneVal
		} else if *done == "false" {
			doneVal := false
			input.Done = &doneVal
		} else {
			return fmt.Errorf("done flag must be 'true' or 'false'")
		}
	}

	// Display operation details
	if !*simple {
		fmt.Printf("=== Updating Checklist Item ===\n")
		fmt.Printf("Item ID: %s\n", *itemID)
		if *title != "" {
			fmt.Printf("New Title: %s\n", *title)
		}
		if *position != -1 {
			fmt.Printf("New Position: %.1f\n", *position)
		}
		if *checklistID != "" {
			fmt.Printf("Move to Checklist: %s\n", *checklistID)
		}
		if *done != "" {
			fmt.Printf("Done Status: %s\n", *done)
		}
		if *projectID != "" {
			fmt.Printf("Project: %s\n", *projectID)
		}
		fmt.Printf("\n")
	}

	// Execute update
	item, err := executeUpdateChecklistItem(client, input)
	if err != nil {
		return fmt.Errorf("failed to update checklist item: %v", err)
	}

	// Display results
	if *simple {
		fmt.Printf("Checklist Item Updated: %s\n", item.ID)
	} else {
		fmt.Printf("=== Checklist Item Updated Successfully ===\n")
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
		fmt.Printf("Updated: %s\n", item.UpdatedAt)
		fmt.Printf("Created By: %s (%s)\n", item.CreatedBy.FullName, item.CreatedBy.Email)
		fmt.Printf("✅ Checklist item updated successfully!\n")
	}

	return nil
}
