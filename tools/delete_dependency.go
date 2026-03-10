package tools

import (
	"flag"
	"fmt"

	. "demo-builder/common"
)

// DeleteTodoDependencyInput represents the input for deleting a dependency
type DeleteTodoDependencyInput struct {
	TodoID      string `json:"todoId"`
	OtherTodoID string `json:"otherTodoId"`
}

// DeleteDependencyResponse represents the response from deleting a dependency
type DeleteDependencyResponse struct {
	DeleteTodoDependency bool `json:"deleteTodoDependency"`
}

func RunDeleteDependency(args []string) error {
	fs := flag.NewFlagSet("delete-dependency", flag.ExitOnError)
	todoID := fs.String("record", "", "Record/Todo ID (required)")
	otherTodoID := fs.String("other-record", "", "Other Record/Todo ID (required)")
	confirm := fs.Bool("confirm", false, "Confirm deletion (required)")
	projectID := fs.String("project", "", "Project ID or slug (optional - for context)")
	simple := fs.Bool("simple", false, "Show simple output")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *todoID == "" {
		return fmt.Errorf("record ID is required (-record)")
	}
	if *otherTodoID == "" {
		return fmt.Errorf("other record ID is required (-other-record)")
	}
	if !*confirm {
		return fmt.Errorf("confirmation required: use -confirm flag to delete dependency")
	}

	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	client := NewClient(config)
	if *projectID != "" {
		client.SetProject(*projectID)
	}

	mutation := `
		mutation DeleteTodoDependency($input: DeleteTodoDependencyInput!) {
			deleteTodoDependency(input: $input)
		}
	`

	variables := map[string]any{
		"input": DeleteTodoDependencyInput{
			TodoID:      *todoID,
			OtherTodoID: *otherTodoID,
		},
	}

	var response DeleteDependencyResponse
	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to delete dependency: %v", err)
	}

	if *simple {
		fmt.Printf("✅ Dependency deleted: %s ↔ %s\n", *todoID, *otherTodoID)
	} else {
		fmt.Printf("=== Dependency Deleted Successfully ===\n")
		fmt.Printf("Record: %s\n", *todoID)
		fmt.Printf("Other Record: %s\n", *otherTodoID)
		fmt.Printf("Result: %v\n", response.DeleteTodoDependency)
		fmt.Printf("\n✅ Dependency deleted successfully!\n")
	}

	return nil
}
