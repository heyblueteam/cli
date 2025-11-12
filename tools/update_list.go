package tools

import (
	"cli/common"
	"flag"
	"fmt"
	"strconv"
	"strings"
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

	// Build the mutation fields dynamically
	var fields []string

	if *title != "" {
		fields = append(fields, fmt.Sprintf("title: \"%s\"", *title))
	}

	if *positionStr != "" {
		position, err := strconv.ParseFloat(*positionStr, 64)
		if err != nil {
			return fmt.Errorf("invalid position value '%s': %v", *positionStr, err)
		}
		fields = append(fields, fmt.Sprintf("position: %g", position))
	}

	if *lockedStr != "" {
		locked, err := strconv.ParseBool(*lockedStr)
		if err != nil {
			return fmt.Errorf("invalid locked value '%s': %v", *lockedStr, err)
		}
		fields = append(fields, fmt.Sprintf("isLocked: %t", locked))
	}

	// Build the fields string
	fieldsStr := ""
	if len(fields) > 0 {
		fieldsStr = strings.Join(fields, "\n\t\t\t\t")
	}

	// Execute mutation using string interpolation like other working commands
	mutation := fmt.Sprintf(`
		mutation EditTodoList {
			editTodoList(input: {
				todoListId: "%s"
				%s
			}) {
				id
				uid
				title
				position
				isLocked
			}
		}`, *listID, fieldsStr)

	var response UpdateTodoListResponse
	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		return fmt.Errorf("failed to edit list: %v. Note: Try providing -project flag for proper authorization", err)
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
		if *title != "" {
			fmt.Printf("- Title: %s\n", *title)
		}
		if *positionStr != "" {
			fmt.Printf("- Position: %s\n", *positionStr)
		}
		if *lockedStr != "" {
			fmt.Printf("- Locked: %s\n", *lockedStr)
		}
	}

	return nil
}