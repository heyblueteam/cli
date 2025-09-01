package tools

import (
	"flag"
	"fmt"
	"strings"

	. "demo-builder/common"
)

// Comment represents a comment structure
type Comment struct {
	ID        string `json:"id"`
	UID       string `json:"uid"`
	HTML      string `json:"html"`
	Text      string `json:"text"`
	Category  string `json:"category"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	User      User   `json:"user"`
}

// CreateCommentInput represents the input for creating a comment
type CreateCommentInput struct {
	HTML       string `json:"html"`
	Text       string `json:"text"`
	Category   string `json:"category"`
	CategoryID string `json:"categoryId"`
	Tiptap     bool   `json:"tiptap,omitempty"`
}

// CreateCommentResponse represents the response from creating a comment
type CreateCommentResponse struct {
	CreateComment Comment `json:"createComment"`
}

// Execute GraphQL mutation to create a comment
func executeCreateComment(client *Client, input CreateCommentInput) (*Comment, error) {
	// Build the mutation
	mutation := `
		mutation CreateComment($input: CreateCommentInput!) {
			createComment(input: $input) {
				id
				uid
				html
				text
				category
				createdAt
				updatedAt
				user {
					id
					uid
					fullName
					email
				}
			}
		}
	`

	// Prepare variables
	variables := map[string]interface{}{
		"input": input,
	}

	// Execute mutation
	var response CreateCommentResponse
	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return nil, err
	}

	return &response.CreateComment, nil
}

// RunCreateComment handles the create-comment command
func RunCreateComment(args []string) error {
	// Define flags
	fs := flag.NewFlagSet("create-comment", flag.ExitOnError)
	recordID := fs.String("record", "", "Record ID to comment on (required)")
	text := fs.String("text", "", "Comment text content (required)")
	html := fs.String("html", "", "Comment HTML content (optional - will use text if not provided)")
	projectID := fs.String("project", "", "Project ID or slug (optional - for context)")
	simple := fs.Bool("simple", false, "Show simple output")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Validate required flags
	if *recordID == "" {
		return fmt.Errorf("record ID is required")
	}
	if *text == "" {
		return fmt.Errorf("comment text is required")
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

	// Prepare comment input
	input := CreateCommentInput{
		Text:       *text,
		HTML:       *html,
		Category:   "TODO", // Comments on records are TODO category
		CategoryID: *recordID,
		Tiptap:     false,
	}

	// If HTML is not provided, use text as HTML (with basic formatting)
	if input.HTML == "" {
		// Convert basic text to HTML - replace newlines with <br>
		input.HTML = strings.ReplaceAll(*text, "\n", "<br>")
	}

	// Display operation details
	if !*simple {
		fmt.Printf("=== Creating Comment ===\n")
		fmt.Printf("Record ID: %s\n", *recordID)
		fmt.Printf("Text: %s\n", *text)
		if *projectID != "" {
			fmt.Printf("Project: %s\n", *projectID)
		}
		fmt.Printf("\n")
	}

	// Execute comment creation
	comment, err := executeCreateComment(client, input)
	if err != nil {
		return fmt.Errorf("failed to create comment: %v", err)
	}

	// Display results
	if *simple {
		fmt.Printf("Comment ID: %s\n", comment.ID)
	} else {
		fmt.Printf("=== Comment Created Successfully ===\n")
		fmt.Printf("ID: %s\n", comment.ID)
		fmt.Printf("UID: %s\n", comment.UID)
		fmt.Printf("Category: %s\n", comment.Category)
		fmt.Printf("Text: %s\n", comment.Text)
		if comment.HTML != comment.Text {
			fmt.Printf("HTML: %s\n", comment.HTML)
		}
		fmt.Printf("Created: %s\n", comment.CreatedAt)
		fmt.Printf("User: %s (%s)\n", comment.User.FullName, comment.User.Email)
		fmt.Printf("✅ Comment added to record successfully!\n")
	}

	return nil
}