package tools

import (
	"encoding/json"
	"flag"
	"fmt"

	. "cli/common"
)

// Response structure for automation deletion
type DeleteAutomationResponse struct {
	DeleteAutomation bool `json:"deleteAutomation"`
}

// Execute GraphQL mutation to delete automation
func executeDeleteAutomation(client *Client, automationID string) (bool, error) {
	mutation := `
		mutation DeleteAutomation($id: String!) {
			deleteAutomation(id: $id)
		}
	`

	variables := map[string]interface{}{
		"id": automationID,
	}

	// Execute mutation
	var response DeleteAutomationResponse
	result, err := client.ExecuteQuery(mutation, variables)
	if err != nil {
		return false, err
	}

	// Parse the response
	data, err := json.Marshal(result)
	if err != nil {
		return false, fmt.Errorf("failed to marshal response: %v", err)
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return false, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return response.DeleteAutomation, nil
}

// Command-line interface
func RunDeleteAutomation(args []string) error {
	fs := flag.NewFlagSet("delete-automation", flag.ExitOnError)
	
	automationID := fs.String("automation", "", "Automation ID (required)")
	projectID := fs.String("project", "", "Project ID or slug (required for project context)")
	confirm := fs.Bool("confirm", false, "Confirm deletion (required for safety)")
	simple := fs.Bool("simple", false, "Simple output format")

	err := fs.Parse(args)
	if err != nil {
		return err
	}

	// Validate required fields
	if *automationID == "" {
		return fmt.Errorf("automation ID is required")
	}
	if *projectID == "" {
		return fmt.Errorf("project ID is required for project context")
	}
	if !*confirm {
		return fmt.Errorf("deletion requires confirmation flag -confirm for safety")
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Create client
	client := NewClient(config)
	client.SetProject(*projectID)

	// Execute deletion
	success, err := executeDeleteAutomation(client, *automationID)
	if err != nil {
		return fmt.Errorf("failed to delete automation: %v", err)
	}

	// Output results
	if *simple {
		if success {
			fmt.Printf("Deleted automation: %s\n", *automationID)
		} else {
			fmt.Printf("Failed to delete automation: %s\n", *automationID)
		}
	} else {
		if success {
			fmt.Printf("✅ Successfully deleted automation\n\n")
			fmt.Printf("Automation ID: %s\n", *automationID)
			fmt.Printf("Status: Permanently deleted\n")
			fmt.Printf("\n⚠️  This action cannot be undone. The automation has been permanently removed.\n")
		} else {
			fmt.Printf("❌ Failed to delete automation\n\n")
			fmt.Printf("Automation ID: %s\n", *automationID)
			fmt.Printf("Status: Deletion failed\n")
			fmt.Printf("\nThe automation may not exist or you may not have permission to delete it.\n")
		}
	}

	return nil
}