package tools

import (
	"demo-builder/common"
	"flag"
	"fmt"
)

// Delete list input
type DeleteTodoListInput struct {
	ProjectID  string `json:"projectId"`
	TodoListID string `json:"todoListId"`
}

// Response structures
type DeleteTodoListResponse struct {
	DeleteTodoList common.MutationResult `json:"deleteTodoList"`
}

func RunDeleteList(args []string) error {
	fs := flag.NewFlagSet("delete-list", flag.ExitOnError)
	projectID := fs.String("project", "", "Project ID (required)")
	listID := fs.String("list", "", "List ID (required)")
	confirm := fs.Bool("confirm", false, "Confirm deletion (required for safety)")
	simple := fs.Bool("simple", false, "Simple output format")
	fs.Parse(args)

	// Validate required parameters
	if *listID == "" {
		fmt.Println("Error: -list flag is required")
		fmt.Println("Usage: go run . delete-list -project PROJECT_ID -list LIST_ID -confirm")
		return fmt.Errorf("list ID is required")
	}

	if *projectID == "" {
		fmt.Println("Error: -project flag is required")
		fmt.Println("Usage: go run . delete-list -project PROJECT_ID -list LIST_ID -confirm")
		return fmt.Errorf("project ID is required")
	}

	if !*confirm {
		fmt.Println("Error: -confirm flag is required for safety")
		fmt.Println("This will permanently delete the todo list and may affect records in this list.")
		fmt.Println("Usage: go run . delete-list -project PROJECT_ID -list LIST_ID -confirm")
		return fmt.Errorf("confirmation required for deletion")
	}

	// Load configuration
	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Create client
	client := common.NewClient(config)

	// Set project context
	client.SetProjectID(*projectID)

	// Prepare input
	input := DeleteTodoListInput{
		ProjectID:  *projectID,
		TodoListID: *listID,
	}

	// Execute mutation
	mutation := `
		mutation DeleteTodoList($input: DeleteTodoListInput!) {
			deleteTodoList(input: $input) {
				success
				operationId
			}
		}`

	variables := map[string]interface{}{
		"input": input,
	}

	var response DeleteTodoListResponse
	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to delete list: %v", err)
	}

	// Display results
	result := response.DeleteTodoList
	if result.Success {
		if *simple {
			fmt.Printf("List %s deleted successfully\n", *listID)
		} else {
			fmt.Printf("=== List Deleted Successfully ===\n")
			fmt.Printf("List ID: %s\n", *listID)
			fmt.Printf("Project ID: %s\n", *projectID)
			if result.OperationID != "" {
				fmt.Printf("Operation ID: %s\n", result.OperationID)
			}
			fmt.Printf("\n⚠️  WARNING: This list has been permanently deleted.\n")
			fmt.Printf("Any records/todos in this list may have been affected.\n")
		}
	} else {
		return fmt.Errorf("failed to delete list %s (success: %v)", *listID, result.Success)
	}

	return nil
}