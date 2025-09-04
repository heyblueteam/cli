package tools

import (
	"demo-builder/common"
	"flag"
	"fmt"
	"strconv"
)

// Update list input
type UpdateTodoListInput struct {
	TodoListID string   `json:"todoListId"`
	Title      string   `json:"title,omitempty"`
	Position   *float64 `json:"position,omitempty"`
	IsLocked   *bool    `json:"isLocked,omitempty"`
}

// Response structures
type EditedTodoList struct {
	ID       string  `json:"id"`
	UID      string  `json:"uid"`
	Title    string  `json:"title"`
	Position float64 `json:"position"`
	IsLocked bool    `json:"isLocked"`
}

type UpdateTodoListResponse struct {
	EditTodoList EditedTodoList `json:"editTodoList"`
}

func RunUpdateList(args []string) error {
	fs := flag.NewFlagSet("update-list", flag.ExitOnError)
	listID := fs.String("list", "", "List ID (required)")
	projectID := fs.String("project", "", "Project ID (optional for context)")
	title := fs.String("title", "", "New title for the list")
	positionStr := fs.String("position", "", "New position for the list (float)")
	lockedStr := fs.String("locked", "", "Lock status (true/false)")
	simple := fs.Bool("simple", false, "Simple output format")
	fs.Parse(args)

	// Validate required parameters
	if *listID == "" {
		return fmt.Errorf("list ID is required. Use -list flag")
	}

	// Check if at least one field is being updated
	if *title == "" && *positionStr == "" && *lockedStr == "" {
		return fmt.Errorf("at least one field must be specified for update (-title, -position, or -locked)")
	}

	// Load configuration
	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Create client
	client := common.NewClient(config)

	// Set project context if provided
	if *projectID != "" {
		client.SetProjectID(*projectID)
	}

	// Prepare input
	input := UpdateTodoListInput{
		TodoListID: *listID,
	}

	// Set optional fields
	if *title != "" {
		input.Title = *title
	}

	if *positionStr != "" {
		position, err := strconv.ParseFloat(*positionStr, 64)
		if err != nil {
			return fmt.Errorf("invalid position value '%s': %v", *positionStr, err)
		}
		input.Position = &position
	}

	if *lockedStr != "" {
		locked, err := strconv.ParseBool(*lockedStr)
		if err != nil {
			return fmt.Errorf("invalid locked value '%s': %v", *lockedStr, err)
		}
		input.IsLocked = &locked
	}

	// Execute mutation using variables
	mutation := `
		mutation EditTodoList($input: EditTodoListInput!) {
			editTodoList(input: $input) {
				id
				uid
				title
				position
				isLocked
			}
		}`

	variables := map[string]interface{}{
		"input": input,
	}

	var response UpdateTodoListResponse
	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to edit list: %v", err)
	}

	// Display results
	list := response.EditTodoList
	if *simple {
		fmt.Printf("List updated: %s (ID: %s)\n", list.Title, list.ID)
	} else {
		fmt.Printf("=== List Updated Successfully ===\n")
		fmt.Printf("ID: %s\n", list.ID)
		fmt.Printf("UID: %s\n", list.UID)
		fmt.Printf("Title: %s\n", list.Title)
		fmt.Printf("Position: %.0f\n", list.Position)
		fmt.Printf("Is Locked: %t\n", list.IsLocked)
		
		fmt.Printf("\nUpdated fields:\n")
		if input.Title != "" {
			fmt.Printf("- Title: %s\n", input.Title)
		}
		if input.Position != nil {
			fmt.Printf("- Position: %.0f\n", *input.Position)
		}
		if input.IsLocked != nil {
			fmt.Printf("- Locked: %t\n", *input.IsLocked)
		}
	}

	return nil
}