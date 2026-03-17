package comments

import (
	"fmt"
	"strings"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a comment on a record",
	Long:  "Create a new comment on a record/todo.",
	Example: `  blue comments create --record <id> --text "Progress update"
  blue comments create -r <id> --text "Update" --html "<p><strong>Update</strong></p>" -w <workspace> -s`,
	RunE: runCreate,
}

var (
	createRecord    string
	createText      string
	createHTML      string
	createWorkspace string
	createSimple    bool
)

func init() {
	createCmd.Flags().StringVarP(&createRecord, "record", "r", "", "Record ID to comment on (required)")
	createCmd.Flags().StringVarP(&createText, "text", "t", "", "Comment text content (required)")
	createCmd.Flags().String("html", "", "Comment HTML content (optional - will use text if not provided)")
	createCmd.Flags().StringVarP(&createWorkspace, "workspace", "w", "", "Workspace ID or slug (optional - for context)")
	createCmd.Flags().BoolVarP(&createSimple, "simple", "s", false, "Show simple output")
}

// Comment represents a comment structure
type Comment struct {
	ID        string      `json:"id"`
	UID       string      `json:"uid"`
	HTML      string      `json:"html"`
	Text      string      `json:"text"`
	Category  string      `json:"category"`
	CreatedAt string      `json:"createdAt"`
	UpdatedAt string      `json:"updatedAt"`
	User      common.User `json:"user"`
}

// CreateCommentInput represents the input for creating a comment
type CreateCommentInput struct {
	HTML       string `json:"html"`
	Text       string `json:"text"`
	Category   string `json:"category"`
	CategoryID string `json:"categoryId"`
	Tiptap     bool   `json:"tiptap,omitempty"`
}

func runCreate(cmd *cobra.Command, args []string) error {
	if createRecord == "" {
		return fmt.Errorf("record ID is required. Use --record flag")
	}
	if createText == "" {
		return fmt.Errorf("comment text is required. Use --text flag")
	}

	htmlVal, _ := cmd.Flags().GetString("html")

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := common.NewClient(config)

	if createWorkspace != "" {
		client.SetProject(createWorkspace)
	}

	// Prepare comment input
	input := CreateCommentInput{
		Text:       createText,
		HTML:       htmlVal,
		Category:   "TODO",
		CategoryID: createRecord,
		Tiptap:     false,
	}

	// If HTML is not provided, use text as HTML (with basic formatting)
	if input.HTML == "" {
		input.HTML = strings.ReplaceAll(createText, "\n", "<br>")
	}

	if !createSimple {
		fmt.Printf("=== Creating Comment ===\n")
		fmt.Printf("Record ID: %s\n", createRecord)
		fmt.Printf("Text: %s\n", createText)
		if createWorkspace != "" {
			fmt.Printf("Workspace: %s\n", createWorkspace)
		}
		fmt.Println()
	}

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

	variables := map[string]interface{}{
		"input": input,
	}

	var response struct {
		CreateComment Comment `json:"createComment"`
	}

	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to create comment: %w", err)
	}

	comment := response.CreateComment

	if createSimple {
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
		fmt.Printf("Comment added to record successfully!\n")
	}

	return nil
}
