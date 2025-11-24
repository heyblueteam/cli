package tools

import (
	"flag"
	"fmt"

	. "demo-builder/common"
)

// DeleteChecklistResponse represents the response from deleting a checklist
type DeleteChecklistResponse struct {
	DeleteChecklist bool `json:"deleteChecklist"`
}

// Execute GraphQL mutation to delete a checklist
func executeDeleteChecklist(client *Client, checklistID string) (bool, error) {
	// Build the mutation
	mutation := `
		mutation DeleteChecklist($id: String!) {
			deleteChecklist(id: $id)
		}
	`

	// Prepare variables
	variables := map[string]interface{}{
		"id": checklistID,
	}

	// Execute mutation
	var response DeleteChecklistResponse
	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return false, err
	}

	return response.DeleteChecklist, nil
}

// RunDeleteChecklist handles the delete-checklist command
func RunDeleteChecklist(args []string) error {
	// Define flags
	fs := flag.NewFlagSet("delete-checklist", flag.ExitOnError)
	checklistID := fs.String("checklist", "", "Checklist ID to delete (required)")
	projectID := fs.String("project", "", "Project ID or slug (optional - for context)")
	confirm := fs.Bool("confirm", false, "Confirm deletion (required for safety)")
	simple := fs.Bool("simple", false, "Show simple output")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Validate required flags
	if *checklistID == "" {
		return fmt.Errorf("checklist ID is required")
	}

	if !*confirm {
		return fmt.Errorf("deletion requires -confirm flag for safety")
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

	// Display operation details
	if !*simple {
		fmt.Printf("=== Deleting Checklist ===\n")
		fmt.Printf("Checklist ID: %s\n", *checklistID)
		if *projectID != "" {
			fmt.Printf("Project: %s\n", *projectID)
		}
		fmt.Printf("⚠️  This will permanently delete the checklist and all its items!\n\n")
	}

	// Execute deletion
	success, err := executeDeleteChecklist(client, *checklistID)
	if err != nil {
		return fmt.Errorf("failed to delete checklist: %v", err)
	}

	// Display results
	if *simple {
		if success {
			fmt.Printf("Checklist deleted: %s\n", *checklistID)
		} else {
			fmt.Printf("Failed to delete checklist: %s\n", *checklistID)
		}
	} else {
		if success {
			fmt.Printf("=== Checklist Deleted Successfully ===\n")
			fmt.Printf("Checklist ID: %s\n", *checklistID)
			fmt.Printf("✅ Checklist and all its items have been permanently deleted.\n")
		} else {
			fmt.Printf("❌ Failed to delete checklist %s\n", *checklistID)
		}
	}

	return nil
}
