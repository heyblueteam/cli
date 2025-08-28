package tools

import (
	"flag"
	"fmt"
	"strings"
	
	. "demo-builder/common"
)

// Response structures - MutationResult is already in types.go
type DeleteProjectResponse struct {
	DeleteProject MutationResult `json:"deleteProject"`
}

// Execute GraphQL mutation to delete project
func executeDeleteProject(client *Client, projectID string) (*MutationResult, error) {
	// Set project context header
	client.SetProjectID(projectID)
	
	// Build the mutation
	mutation := fmt.Sprintf(`
		mutation DeleteProject {
			deleteProject(id: "%s") {
				success
				operationId
			}
		}
	`, projectID)

	// Execute mutation
	var response DeleteProjectResponse
	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		return nil, err
	}

	return &response.DeleteProject, nil
}

// RunDeleteProject deletes a project
func RunDeleteProject(args []string) error {
	// Create flag set for this command
	fs := flag.NewFlagSet("delete-project", flag.ExitOnError)
	
	// Parse command line flags
	projectID := fs.String("project", "", "Project ID to delete (required)")
	confirm := fs.Bool("confirm", false, "Confirm deletion (required for safety)")
	
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Validate required parameters
	if *projectID == "" {
		return fmt.Errorf("project ID is required. Use -project flag")
	}

	if !*confirm {
		return fmt.Errorf("deletion confirmation is required. Use -confirm flag to confirm deletion")
	}

	// Show warning but proceed with -confirm flag
	fmt.Printf("⚠️  WARNING: Deleting project '%s' (this action cannot be undone)\n", *projectID)

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create client
	client := NewClient(config)

	// Execute deletion
	fmt.Printf("Deleting project '%s'...\n", *projectID)
	
	result, err := executeDeleteProject(client, *projectID)
	if err != nil {
		if strings.Contains(err.Error(), "not authorized") {
			return fmt.Errorf("failed to delete project: %w\n\nNote: Project deletion requires special permissions. Contact your administrator if you need to delete projects", err)
		}
		return fmt.Errorf("failed to delete project: %w", err)
	}

	// Display results
	if result.Success {
		fmt.Println("\n✅ Project deleted successfully!")
		if result.OperationID != "" {
			fmt.Printf("Operation ID: %s\n", result.OperationID)
		}
	} else {
		fmt.Println("\n❌ Project deletion failed")
		if result.OperationID != "" {
			fmt.Printf("Operation ID: %s\n", result.OperationID)
		}
	}
	
	return nil
}