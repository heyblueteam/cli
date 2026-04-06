package comments

import (
	"fmt"
	"strings"
	"time"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List comments on a record",
	Long:  "List all comments and replies for a specific record.",
	Example: `  blue comments list --record <id>
  blue comments list --record <id> --workspace <id>
  blue comments list --record <id> --limit 50
  blue comments list --record <id> --simple`,
	RunE: runList,
}

var (
	listRecord    string
	listWorkspace string
	listSimple    bool
	listLimit     int
	listSkip      int
)

func init() {
	listCmd.Flags().StringVarP(&listRecord, "record", "r", "", "Record ID to list comments for (required)")
	listCmd.Flags().StringVarP(&listWorkspace, "workspace", "w", "", "Workspace ID or slug (optional)")
	listCmd.Flags().BoolVarP(&listSimple, "simple", "s", false, "Simple output format")
	listCmd.Flags().IntVarP(&listLimit, "limit", "l", 20, "Maximum number of comments to return")
	listCmd.Flags().IntVar(&listSkip, "skip", 0, "Number of comments to skip (for pagination)")
}

// CommentWithReplies represents a comment with nested replies
type CommentWithReplies struct {
	ID        string      `json:"id"`
	UID       string      `json:"uid"`
	HTML      string      `json:"html"`
	Text      string      `json:"text"`
	Category  string      `json:"category"`
	CreatedAt string      `json:"createdAt"`
	UpdatedAt string      `json:"updatedAt"`
	User      common.User `json:"user"`
	Replies   []Comment   `json:"replies"`
}

// PageInfo represents pagination information
type PageInfo struct {
	HasNextPage     bool `json:"hasNextPage"`
	HasPreviousPage bool `json:"hasPreviousPage"`
	Page            int  `json:"page"`
	TotalPages      int  `json:"totalPages"`
	PerPage         int  `json:"perPage"`
}

func runList(cmd *cobra.Command, args []string) error {
	if listRecord == "" {
		return fmt.Errorf("record ID is required. Use --record flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)

	if listWorkspace != "" {
		client.SetProject(listWorkspace)
	}

	query := `
		query CommentList($categoryId: String!, $category: CommentCategory!, $first: Int, $skip: Int, $orderBy: DiscussionOrderByInput) {
			commentList(categoryId: $categoryId, category: $category, first: $first, skip: $skip, orderBy: $orderBy) {
				comments {
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
					replies {
						id
						text
						createdAt
						user {
							id
							fullName
							email
						}
					}
				}
				totalCount
				pageInfo {
					hasNextPage
					hasPreviousPage
					page
					totalPages
					perPage
				}
			}
		}
	`

	variables := map[string]interface{}{
		"categoryId": listRecord,
		"category":   "TODO",
		"first":      listLimit,
		"skip":       listSkip,
		"orderBy":    "createdAt_ASC",
	}

	var response struct {
		CommentList struct {
			Comments   []CommentWithReplies `json:"comments"`
			TotalCount int                  `json:"totalCount"`
			PageInfo   PageInfo             `json:"pageInfo"`
		} `json:"commentList"`
	}

	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to list comments: %w", err)
	}

	comments := response.CommentList.Comments
	totalCount := response.CommentList.TotalCount
	pageInfo := response.CommentList.PageInfo

	if listSimple {
		fmt.Printf("Record: %s\n", listRecord)
		fmt.Printf("Comments: %d\n\n", totalCount)
		for i, c := range comments {
			fmt.Printf("%d. [%s] %s: %s\n", i+1+listSkip, formatTimestamp(c.CreatedAt), c.User.FullName, truncateText(c.Text, 120))
			for _, r := range c.Replies {
				fmt.Printf("   -> [%s] %s: %s\n", formatTimestamp(r.CreatedAt), r.User.FullName, truncateText(r.Text, 100))
			}
		}
		return nil
	}

	fmt.Printf("=== Comments for Record ===\n")
	fmt.Printf("Record ID: %s\n", listRecord)
	fmt.Printf("Showing %d of %d comments", len(comments), totalCount)
	if pageInfo.TotalPages > 1 {
		fmt.Printf(" (page %d of %d)", pageInfo.Page, pageInfo.TotalPages)
	}
	fmt.Println()
	fmt.Println()

	if len(comments) == 0 {
		fmt.Println("No comments found for this record.")
		return nil
	}

	for i, c := range comments {
		num := i + 1 + listSkip
		fmt.Printf("%d. %s (%s)\n", num, c.User.FullName, c.User.Email)
		fmt.Printf("   Date: %s\n", formatTimestamp(c.CreatedAt))
		fmt.Printf("   ID: %s\n", c.ID)
		fmt.Printf("   %s\n", c.Text)

		if len(c.Replies) > 0 {
			for _, r := range c.Replies {
				fmt.Printf("   \u2514\u2500 Reply by %s (%s)\n", r.User.FullName, formatTimestamp(r.CreatedAt))
				fmt.Printf("      %s\n", r.Text)
			}
		}
		fmt.Println()
	}

	if pageInfo.HasNextPage {
		nextSkip := listSkip + listLimit
		fmt.Printf("--- More comments available. Use --skip %d to see the next page ---\n", nextSkip)
	}

	return nil
}

// formatTimestamp formats an ISO timestamp into a readable date string
func formatTimestamp(ts string) string {
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		t, err = time.Parse(time.RFC3339, ts)
		if err != nil {
			return ts
		}
	}
	return t.Format("2006-01-02 15:04")
}

// truncateText shortens text to maxLen characters with an ellipsis
func truncateText(text string, maxLen int) string {
	text = strings.ReplaceAll(text, "\n", " ")
	if len(text) > maxLen {
		return text[:maxLen-3] + "..."
	}
	return text
}
