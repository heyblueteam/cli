package tools

import (
	"flag"
	"fmt"
	"strings"

	. "demo-builder/common"
)

// EditCommentInput represents the input for editing a comment
type EditCommentInput struct {
	ID   string `json:"id"`
	HTML string `json:"html"`
	Text string `json:"text"`
}

// EditCommentResponse represents the response from editing a comment
type EditCommentResponse struct {
	EditComment Comment `json:"editComment"`
}

// Execute GraphQL mutation to update a comment
func executeEditComment(client *Client, input EditCommentInput) (*Comment, error) {
	// Build the mutation
	mutation := `
		mutation EditComment($input: EditCommentInput!) {
			editComment(input: $input) {
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
	var response EditCommentResponse
	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return nil, err
	}

	return &response.EditComment, nil
}

// RunUpdateComment handles the update-comment command
func RunUpdateComment(args []string) error {
	// Define flags
	fs := flag.NewFlagSet("update-comment", flag.ExitOnError)
	commentID := fs.String("comment", "", "Comment ID to update (required)")
	text := fs.String("text", "", "Updated comment text content (required)")
	html := fs.String("html", "", "Updated comment HTML content (optional - will use text if not provided)")
	projectID := fs.String("project", "", "Project ID or slug (optional - for context)")
	simple := fs.Bool("simple", false, "Show simple output")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Validate required flags
	if *commentID == "" {
		return fmt.Errorf("comment ID is required")
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
	input := EditCommentInput{
		ID:   *commentID,
		Text: *text,
		HTML: *html,
	}

	// If HTML is not provided, use text as HTML (with basic formatting)
	if input.HTML == "" {
		// Convert basic text to HTML - replace newlines with <br>
		input.HTML = strings.ReplaceAll(*text, "\n", "<br>")
	}

	// Display operation details
	if !*simple {
		fmt.Printf("=== Updating Comment ===\n")
		fmt.Printf("Comment ID: %s\n", *commentID)
		fmt.Printf("New Text: %s\n", *text)
		if *projectID != "" {
			fmt.Printf("Project: %s\n", *projectID)
		}
		fmt.Printf("\n")
	}

	// Execute comment update
	comment, err := executeEditComment(client, input)
	if err != nil {
		return fmt.Errorf("failed to update comment: %v", err)
	}

	// Display results
	if *simple {
		fmt.Printf("Comment updated: %s\n", comment.ID)
	} else {
		fmt.Printf("=== Comment Updated Successfully ===\n")
		fmt.Printf("ID: %s\n", comment.ID)
		fmt.Printf("UID: %s\n", comment.UID)
		fmt.Printf("Category: %s\n", comment.Category)
		fmt.Printf("Text: %s\n", comment.Text)
		if comment.HTML != comment.Text {
			fmt.Printf("HTML: %s\n", comment.HTML)
		}
		fmt.Printf("Created: %s\n", comment.CreatedAt)
		fmt.Printf("Updated: %s\n", comment.UpdatedAt)
		fmt.Printf("User: %s (%s)\n", comment.User.FullName, comment.User.Email)
		fmt.Printf("✅ Comment updated successfully!\n")
	}

	return nil
}