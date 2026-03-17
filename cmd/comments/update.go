package comments

import (
	"fmt"
	"strings"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a comment",
	Long:  "Update an existing comment's text and/or HTML content.",
	Example: `  blue comments update --comment <id> --text "Updated text"
  blue comments update --comment <id> --text "Updated" --html "<p><em>Updated</em></p>" -w <workspace> -s`,
	RunE: runUpdate,
}

var (
	updateComment   string
	updateText      string
	updateWorkspace string
	updateSimple    bool
)

func init() {
	updateCmd.Flags().StringVar(&updateComment, "comment", "", "Comment ID to update (required)")
	updateCmd.Flags().StringVarP(&updateText, "text", "t", "", "Updated comment text content (required)")
	updateCmd.Flags().String("html", "", "Updated comment HTML content (optional - will use text if not provided)")
	updateCmd.Flags().StringVarP(&updateWorkspace, "workspace", "w", "", "Workspace ID or slug (optional - for context)")
	updateCmd.Flags().BoolVarP(&updateSimple, "simple", "s", false, "Show simple output")
}

// EditCommentInput represents the input for editing a comment
type EditCommentInput struct {
	ID   string `json:"id"`
	HTML string `json:"html"`
	Text string `json:"text"`
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if updateComment == "" {
		return fmt.Errorf("comment ID is required. Use --comment flag")
	}
	if updateText == "" {
		return fmt.Errorf("comment text is required. Use --text flag")
	}

	htmlVal, _ := cmd.Flags().GetString("html")

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := common.NewClient(config)

	if updateWorkspace != "" {
		client.SetProject(updateWorkspace)
	}

	input := EditCommentInput{
		ID:   updateComment,
		Text: updateText,
		HTML: htmlVal,
	}

	if input.HTML == "" {
		input.HTML = strings.ReplaceAll(updateText, "\n", "<br>")
	}

	if !updateSimple {
		fmt.Printf("=== Updating Comment ===\n")
		fmt.Printf("Comment ID: %s\n", updateComment)
		fmt.Printf("New Text: %s\n", updateText)
		if updateWorkspace != "" {
			fmt.Printf("Workspace: %s\n", updateWorkspace)
		}
		fmt.Println()
	}

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

	variables := map[string]interface{}{
		"input": input,
	}

	var response struct {
		EditComment Comment `json:"editComment"`
	}

	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to update comment: %w", err)
	}

	comment := response.EditComment

	if updateSimple {
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
		fmt.Printf("Comment updated successfully!\n")
	}

	return nil
}
