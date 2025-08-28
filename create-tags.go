package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
)

// Tag represents a tag in the system
type Tag struct {
	ID        string `json:"id"`
	UID       string `json:"uid"`
	Title     string `json:"title"`
	Color     string `json:"color"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

func main() {
	var projectID, title, color string
	flag.StringVar(&projectID, "project", "", "Project ID (required)")
	flag.StringVar(&title, "title", "", "Tag title (required)")
	flag.StringVar(&color, "color", "", "Tag color (required)")
	flag.Parse()

	if projectID == "" {
		log.Fatal("Project ID is required. Use -project flag.")
	}
	if title == "" {
		log.Fatal("Tag title is required. Use -title flag.")
	}
	if color == "" {
		log.Fatal("Tag color is required. Use -color flag.")
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create client using shared auth
	client := NewClient(config)
	
	// Set project context for tag creation
	client.SetProjectID(projectID)

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
			"title": strings.TrimSpace(title),
			"color": strings.TrimSpace(color),
		},
	}

	// Execute mutation
	fmt.Printf("=== Creating Tag ===\n")

	var tagResponse struct {
		CreateTag Tag `json:"createTag"`
	}

	if err := client.ExecuteQueryWithResult(mutation, variables, &tagResponse); err != nil {
		log.Fatalf("Failed to create tag: %v", err)
	}

	// Display results
	fmt.Printf("✅ Tag created successfully!\n\n")
	fmt.Printf("Title: %s\n", tagResponse.CreateTag.Title)
	fmt.Printf("ID: %s\n", tagResponse.CreateTag.ID)
	fmt.Printf("UID: %s\n", tagResponse.CreateTag.UID)
	fmt.Printf("Color: %s\n", tagResponse.CreateTag.Color)
	fmt.Printf("Created: %s\n", tagResponse.CreateTag.CreatedAt)
}