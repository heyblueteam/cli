package tools

import (
	"flag"
	"fmt"

	. "demo-builder/common"
)

// DeleteChecklistItemResponse represents the response from deleting a checklist item
type DeleteChecklistItemResponse struct {
	DeleteChecklistItem bool `json:"deleteChecklistItem"`
}

// Execute GraphQL mutation to delete a checklist item
func executeDeleteChecklistItem(client *Client, itemID string) (bool, error) {
	// Build the mutation
	mutation := `
		mutation DeleteChecklistItem($id: String!) {
			deleteChecklistItem(id: $id)
		}
	`

	// Prepare variables
	variables := map[string]interface{}{
		"id": itemID,
	}

	// Execute mutation
	var response DeleteChecklistItemResponse
	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return false, err
	}

	return response.DeleteChecklistItem, nil
}

// RunDeleteChecklistItem handles the delete-checklist-item command
func RunDeleteChecklistItem(args []string) error {
	// Define flags
	fs := flag.NewFlagSet("delete-checklist-item", flag.ExitOnError)
	itemID := fs.String("item", "", "Checklist item ID to delete (required)")
	projectID := fs.String("project", "", "Project ID or slug (optional - for context)")
	confirm := fs.Bool("confirm", false, "Confirm deletion (required for safety)")
	simple := fs.Bool("simple", false, "Show simple output")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Validate required flags
	if *itemID == "" {
		return fmt.Errorf("checklist item ID is required")
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
		fmt.Printf("=== Deleting Checklist Item ===\n")
		fmt.Printf("Item ID: %s\n", *itemID)
		if *projectID != "" {
			fmt.Printf("Project: %s\n", *projectID)
		}
		fmt.Printf("⚠️  This will permanently delete the checklist item!\n\n")
	}

	// Execute deletion
	success, err := executeDeleteChecklistItem(client, *itemID)
	if err != nil {
		return fmt.Errorf("failed to delete checklist item: %v", err)
	}

	// Display results
	if *simple {
		if success {
			fmt.Printf("Checklist item deleted: %s\n", *itemID)
		} else {
			fmt.Printf("Failed to delete checklist item: %s\n", *itemID)
		}
	} else {
		if success {
			fmt.Printf("=== Checklist Item Deleted Successfully ===\n")
			fmt.Printf("Item ID: %s\n", *itemID)
			fmt.Printf("✅ Checklist item has been permanently deleted.\n")
		} else {
			fmt.Printf("❌ Failed to delete checklist item %s\n", *itemID)
		}
	}

	return nil
}
