package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
)

// Response structures
type MutationResult struct {
	Success     bool   `json:"success"`
	OperationID string `json:"operationId"`
}

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

func main() {
	// Parse command line flags
	projectID := flag.String("project", "", "Project ID to delete (required)")
	confirm := flag.Bool("confirm", false, "Confirm deletion (required for safety)")
	flag.Parse()

	// Validate required parameters
	if *projectID == "" {
		log.Fatal("Project ID is required. Use -project flag")
	}

	if !*confirm {
		log.Fatal("Deletion confirmation is required. Use -confirm flag to confirm deletion")
	}

	// Show warning but proceed with -confirm flag
	fmt.Printf("⚠️  WARNING: Deleting project '%s' (this action cannot be undone)\n", *projectID)

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create client
	client := NewClient(config)

	// Execute deletion
	fmt.Printf("Deleting project '%s'...\n", *projectID)
	
	result, err := executeDeleteProject(client, *projectID)
	if err != nil {
		if strings.Contains(err.Error(), "not authorized") {
			log.Fatalf("Failed to delete project: %v\n\nNote: Project deletion requires special permissions. Contact your administrator if you need to delete projects.", err)
		}
		log.Fatalf("Failed to delete project: %v", err)
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
}