package tools

import (
	"flag"
	"fmt"
	"strings"
	
	. "demo-builder/common"
)

// Tag is already defined in common/types.go

func RunCreateTags(args []string) error {
	fs := flag.NewFlagSet("create-tags", flag.ExitOnError)
	projectID := fs.String("project", "", "Project ID (required)")
	title := fs.String("title", "", "Tag title (required)")
	color := fs.String("color", "", "Tag color (required)")
	fs.Parse(args)

	if *projectID == "" {
		return fmt.Errorf("project ID is required. Use -project flag")
	}
	if *title == "" {
		return fmt.Errorf("tag title is required. Use -title flag")
	}
	if *color == "" {
		return fmt.Errorf("tag color is required. Use -color flag")
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Create client using shared auth
	client := NewClient(config)
	
	// Set project context for tag creation
	client.SetProjectID(*projectID)

	// GraphQL mutation for creating a tag
	mutation := `
		mutation CreateTag($input: CreateTagInput!) {
			createTag(input: $input) {
				id
				uid
				title
				color
				createdAt
				updatedAt
			}
		}
	`

	// Variables
	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"title": strings.TrimSpace(*title),
			"color": strings.TrimSpace(*color),
		},
	}

	// Execute mutation
	fmt.Printf("=== Creating Tag ===\n")

	var tagResponse struct {
		CreateTag Tag `json:"createTag"`
	}

	if err := client.ExecuteQueryWithResult(mutation, variables, &tagResponse); err != nil {
		return fmt.Errorf("failed to create tag: %v", err)
	}

	// Display results
	fmt.Printf("✅ Tag created successfully!\n\n")
	fmt.Printf("Title: %s\n", tagResponse.CreateTag.Title)
	fmt.Printf("ID: %s\n", tagResponse.CreateTag.ID)
	fmt.Printf("UID: %s\n", tagResponse.CreateTag.UID)
	fmt.Printf("Color: %s\n", tagResponse.CreateTag.Color)
	fmt.Printf("Created: %s\n", tagResponse.CreateTag.CreatedAt)

	return nil
}