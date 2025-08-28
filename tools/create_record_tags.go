package tools

import (
	"flag"
	"fmt"
	"strings"
	. "demo-builder/common"
)

func RunCreateRecordTags(args []string) error {
	fs := flag.NewFlagSet("create-record-tags", flag.ExitOnError)
	
	var recordID = fs.String("record", "", "Record/Todo ID to add tags to (required)")
	var tagIDs = fs.String("tag-ids", "", "Comma-separated list of existing tag IDs to add")
	var tagTitles = fs.String("tag-titles", "", "Comma-separated list of tag titles to add (will create if not exist)")
	var projectID = fs.String("project", "", "Project ID (required for tag title lookup)")
	var simple = fs.Bool("simple", false, "Simple output format")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse arguments: %v", err)
	}

	if *recordID == "" {
		fmt.Println("Error: -record flag is required")
		fmt.Println("\nUsage:")
		fmt.Println("  go run auth.go add-tags-to-record.go -record RECORD_ID [flags]")
		fmt.Println("\nFlags:")
		fs.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  # Add existing tags by ID")
		fmt.Println("  go run auth.go add-tags-to-record.go -record cm7abc123 -tag-ids \"tag1,tag2\"")
		fmt.Println("")
		fmt.Println("  # Add tags by title (will create if needed)")
		fmt.Println("  go run auth.go add-tags-to-record.go -record cm7abc123 -tag-titles \"Bug,Priority\" -project PROJECT_ID")
		return fmt.Errorf("record flag is required")
	}

	if *tagIDs == "" && *tagTitles == "" {
		return fmt.Errorf("either -tag-ids or -tag-titles must be provided")
	}

	if *tagTitles != "" && *projectID == "" {
		return fmt.Errorf("project flag is required when using -tag-titles")
	}

	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
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
		return fmt.Errorf("failed to add tags to record: %v", err)
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

	return nil
}