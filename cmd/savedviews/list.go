package savedviews

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List saved views",
	Example: `  blue saved-views list --workspace <id>
  blue saved-views list --workspace <id> --simple`,
	RunE: runList,
}

var (
	listWorkspace string
	listSimple    bool
	listPage      int
	listSize      int
	listFormat    string
)

func init() {
	listCmd.Flags().StringVarP(&listWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	listCmd.Flags().BoolVarP(&listSimple, "simple", "s", false, "Simple output")
	listCmd.Flags().IntVar(&listPage, "page", 1, "Page number")
	listCmd.Flags().IntVar(&listSize, "size", 100, "Page size")
	listCmd.Flags().StringVar(&listFormat, "format", "", "Output format (json)")
}

func runList(cmd *cobra.Command, args []string) error {
	if listWorkspace == "" {
		return fmt.Errorf("workspace ID is required. Use --workspace flag")
	}
	client, err := clientFor(listWorkspace)
	if err != nil {
		return err
	}
	projectID, err := resolveWorkspaceID(client, listWorkspace)
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`query ListSavedViews($filter: SavedViewFilterInput!, $skip: Int, $take: Int) { savedViews(filter: $filter, skip: $skip, take: $take) { items { %s } pageInfo { totalItems hasNextPage } } }`, savedViewFields)
	variables := map[string]interface{}{"filter": map[string]interface{}{"projectId": projectID}, "skip": (listPage - 1) * listSize, "take": listSize}
	var response struct {
		SavedViews struct {
			Items    []SavedView `json:"items"`
			PageInfo struct {
				TotalItems  int  `json:"totalItems"`
				HasNextPage bool `json:"hasNextPage"`
			} `json:"pageInfo"`
		} `json:"savedViews"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to list saved views: %w", err)
	}
	if listFormat == "json" {
		return printJSON(response.SavedViews.Items)
	}
	if listSimple {
		for _, view := range response.SavedViews.Items {
			fmt.Printf("%s  %s  %s\n", view.ID, view.ViewType, view.Name)
		}
		return nil
	}
	fmt.Printf("=== Saved Views (%d total) ===\n\n", response.SavedViews.PageInfo.TotalItems)
	for i, view := range response.SavedViews.Items {
		fmt.Printf("%d. ", i+1)
		printView(view)
		fmt.Println()
	}
	return nil
}
