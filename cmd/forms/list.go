package forms

import (
	"encoding/json"
	"fmt"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List forms in a workspace",
	Example: `  blue forms list --workspace <id>
  blue forms list --workspace <id> --simple
  blue forms list --workspace <id> --sort title_ASC --page 2 --size 50
  blue forms list --workspace <id> --format json`,
	RunE: runList,
}

var (
	listWorkspace string
	listSimple    bool
	listSort      string
	listPage      int
	listSize      int
	listFormat    string
)

func init() {
	listCmd.Flags().StringVarP(&listWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	listCmd.Flags().BoolVarP(&listSimple, "simple", "s", false, "Simple output format")
	listCmd.Flags().StringVar(&listSort, "sort", "updatedAt_DESC", "Sort order (updatedAt_DESC, title_ASC)")
	listCmd.Flags().IntVar(&listPage, "page", 1, "Page number (1-indexed)")
	listCmd.Flags().IntVar(&listSize, "size", 20, "Page size")
	listCmd.Flags().StringVar(&listFormat, "format", "", "Output format (json)")
}

func runList(cmd *cobra.Command, args []string) error {
	if listWorkspace == "" {
		return fmt.Errorf("workspace ID is required. Use --workspace flag")
	}
	if listPage < 1 {
		listPage = 1
	}
	if listSize < 1 {
		listSize = 20
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)
	client.SetProject(listWorkspace)

	query := `
		query ListForms($filter: FormFilterInput, $sort: FormSort, $skip: Int, $take: Int) {
			forms(filter: $filter, sort: $sort, skip: $skip, take: $take) {
				items {
					id
					uid
					title
					description
					isActive
					theme
					primaryColor
					hideBranding
					createdAt
					updatedAt
				}
				pageInfo {
					hasNextPage
					totalItems
				}
			}
		}
	`

	variables := map[string]interface{}{
		"filter": map[string]interface{}{},
		"sort":   listSort,
		"skip":   (listPage - 1) * listSize,
		"take":   listSize,
	}

	var response struct {
		Forms struct {
			Items    []FormSummary `json:"items"`
			PageInfo struct {
				HasNextPage bool `json:"hasNextPage"`
				TotalItems  int  `json:"totalItems"`
			} `json:"pageInfo"`
		} `json:"forms"`
	}

	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to list forms: %w", err)
	}

	if listFormat == "json" {
		out, err := json.MarshalIndent(response.Forms.Items, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(out))
		return nil
	}

	if listSimple {
		for _, f := range response.Forms.Items {
			active := ""
			if !f.IsActive {
				active = " [inactive]"
			}
			fmt.Printf("%s  %s%s\n", f.ID, f.Title, active)
		}
		return nil
	}

	fmt.Printf("=== Forms (page %d, %d shown, %d total) ===\n\n", listPage, len(response.Forms.Items), response.Forms.PageInfo.TotalItems)
	for i, f := range response.Forms.Items {
		fmt.Printf("%d. %s\n", i+1, f.Title)
		fmt.Printf("   ID:           %s\n", f.ID)
		fmt.Printf("   UID:          %s\n", f.UID)
		fmt.Printf("   Active:       %t\n", f.IsActive)
		fmt.Printf("   Theme:        %s\n", f.Theme)
		fmt.Printf("   PrimaryColor: %s\n", f.PrimaryColor)
		fmt.Printf("   HideBranding: %t\n", f.HideBranding)
		fmt.Printf("   Updated:      %s\n", f.UpdatedAt)
		fmt.Println()
	}
	if response.Forms.PageInfo.HasNextPage {
		fmt.Printf("More results available — pass --page %d to fetch the next page.\n", listPage+1)
	}
	return nil
}
