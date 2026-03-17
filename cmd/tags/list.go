package tags

import (
	"fmt"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List tags in a workspace",
	Long:  "List all tags within a specific workspace.",
	Example: `  blue tags list --workspace <id>`,
	RunE: runList,
}

var (
	listWorkspace string
)

func init() {
	listCmd.Flags().StringVarP(&listWorkspace, "workspace", "w", "", "Workspace ID (required)")
}

func runList(cmd *cobra.Command, args []string) error {
	if listWorkspace == "" {
		return fmt.Errorf("workspace ID is required. Use --workspace flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)

	query := fmt.Sprintf(`
		query ListTags {
			tagList(
				filter: { projectIds: ["%s"] }
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
	`, listWorkspace)

	var response struct {
		TagList struct {
			Items      []common.Tag `json:"items"`
			TotalCount int          `json:"totalCount"`
		} `json:"tagList"`
	}

	if err := client.ExecuteQueryWithResult(query, nil, &response); err != nil {
		return fmt.Errorf("failed to list tags: %w", err)
	}

	fmt.Printf("=== Tags in Workspace %s ===\n", listWorkspace)
	fmt.Printf("Total tags: %d\n\n", response.TagList.TotalCount)

	if len(response.TagList.Items) == 0 {
		fmt.Println("No tags found.")
		return nil
	}

	for i, tag := range response.TagList.Items {
		fmt.Printf("%d. %s\n", i+1, tag.Title)
		fmt.Printf("   ID: %s\n", tag.ID)
		fmt.Printf("   Color: %s\n", tag.Color)
		fmt.Println()
	}

	return nil
}
