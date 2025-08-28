package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
)

func main() {
	var recordID = flag.String("record", "", "Record/Todo ID to add tags to (required)")
	var tagIDs = flag.String("tag-ids", "", "Comma-separated list of existing tag IDs to add")
	var tagTitles = flag.String("tag-titles", "", "Comma-separated list of tag titles to add (will create if not exist)")
	var projectID = flag.String("project", "", "Project ID (required for tag title lookup)")
	var simple = flag.Bool("simple", false, "Simple output format")

	flag.Parse()

	if *recordID == "" {
		fmt.Println("Error: -record flag is required")
		fmt.Println("\nUsage:")
		fmt.Println("  go run auth.go add-tags-to-record.go -record RECORD_ID [flags]")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  # Add existing tags by ID")
		fmt.Println("  go run auth.go add-tags-to-record.go -record cm7abc123 -tag-ids \"tag1,tag2\"")
		fmt.Println("")
		fmt.Println("  # Add tags by title (will create if needed)")
		fmt.Println("  go run auth.go add-tags-to-record.go -record cm7abc123 -tag-titles \"Bug,Priority\" -project PROJECT_ID")
		return
	}

	if *tagIDs == "" && *tagTitles == "" {
		fmt.Println("Error: Either -tag-ids or -tag-titles must be provided")
		return
	}

	if *tagTitles != "" && *projectID == "" {
		fmt.Println("Error: -project flag is required when using -tag-titles")
		return
	}

	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	client := NewClient(config)

	// Set project context if provided
	if *projectID != "" {
		client.SetProjectID(*projectID)
	}

	// Prepare tag IDs and titles arrays
	var tagIDsList []string
	var tagTitlesList []string

	if *tagIDs != "" {
		tagIDsList = strings.Split(*tagIDs, ",")
		for i, id := range tagIDsList {
			tagIDsList[i] = strings.TrimSpace(id)
		}
	}

	if *tagTitles != "" {
		tagTitlesList = strings.Split(*tagTitles, ",")
		for i, title := range tagTitlesList {
			tagTitlesList[i] = strings.TrimSpace(title)
		}
	}

	// GraphQL mutation for setting todo tags
	mutation := `
		mutation SetTodoTags($input: SetTodoTagsInput!) {
			setTodoTags(input: $input)
		}
	`

	// Variables
	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"todoId": *recordID,
		},
	}

	// Add tag IDs if provided
	if len(tagIDsList) > 0 {
		variables["input"].(map[string]interface{})["tagIds"] = tagIDsList
	}

	// Add tag titles if provided
	if len(tagTitlesList) > 0 {
		variables["input"].(map[string]interface{})["tagTitles"] = tagTitlesList
	}

	// Execute mutation
	if !*simple {
		fmt.Printf("=== Adding Tags to Record ===\n")
		fmt.Printf("Record ID: %s\n", *recordID)
		if len(tagIDsList) > 0 {
			fmt.Printf("Tag IDs: %s\n", strings.Join(tagIDsList, ", "))
		}
		if len(tagTitlesList) > 0 {
			fmt.Printf("Tag Titles: %s\n", strings.Join(tagTitlesList, ", "))
		}
		fmt.Printf("\n")
	}

	var response struct {
		SetTodoTags bool `json:"setTodoTags"`
	}

	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		log.Fatalf("Failed to add tags to record: %v", err)
	}

	// Display results
	if response.SetTodoTags {
		if *simple {
			fmt.Printf("Tags added to record %s\n", *recordID)
		} else {
			fmt.Printf("✅ Tags successfully added to record!\n")
		}
	} else {
		if *simple {
			fmt.Printf("Failed to add tags to record %s\n", *recordID)
		} else {
			fmt.Printf("❌ Failed to add tags to record\n")
		}
	}
}