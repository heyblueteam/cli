package main

import (
	"flag"
	"fmt"
	"log"
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
	var projectID string
	flag.StringVar(&projectID, "project", "", "Project ID (required)")
	flag.Parse()

	if projectID == "" {
		log.Fatal("Project ID is required. Use -project flag.")
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create client using shared auth
	client := NewClient(config)

	// GraphQL query for listing tags
	query := fmt.Sprintf(`
		query ListTags {
			tagList(
				filter: { 
					projectIds: ["%s"] 
				}
				first: 50
				orderBy: title_ASC
			) {
				items {
					id
					uid
					title
					color
					createdAt
					updatedAt
				}
				totalCount
			}
		}
	`, projectID)

	// Variables
	variables := map[string]interface{}{}

	// Execute query
	fmt.Printf("=== Tags in Project %s ===\n", projectID)

	// Execute query
	var tagResponse struct {
		TagList struct {
			Items      []Tag `json:"items"`
			TotalCount int   `json:"totalCount"`
		} `json:"tagList"`
	}

	if err := client.ExecuteQueryWithResult(query, variables, &tagResponse); err != nil {
		log.Fatalf("Failed to query tags: %v", err)
	}

	// Display results
	fmt.Printf("Total tags: %d\n\n", tagResponse.TagList.TotalCount)

	if len(tagResponse.TagList.Items) == 0 {
		fmt.Println("No tags found for this project.")
	} else {
		for i, tag := range tagResponse.TagList.Items {
			fmt.Printf("%d. %s\n", i+1, tag.Title)
			fmt.Printf("   ID: %s\n", tag.ID)
			fmt.Printf("   UID: %s\n", tag.UID)
			fmt.Printf("   Color: %s\n", tag.Color)
			fmt.Println()
		}
	}
}