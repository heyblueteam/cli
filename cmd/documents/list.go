package documents

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List documents in a workspace",
	Example: `  blue documents list --workspace <id>
  blue documents list --workspace <id> --wiki true --simple
  blue documents list --workspace <id> --format json`,
	RunE: runList,
}

var (
	listWorkspace string
	listWiki      string
	listSimple    bool
	listPage      int
	listSize      int
	listSort      string
	listFormat    string
)

func init() {
	listCmd.Flags().StringVarP(&listWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	listCmd.Flags().StringVar(&listWiki, "wiki", "", "Filter wiki pages (true or false)")
	listCmd.Flags().BoolVarP(&listSimple, "simple", "s", false, "Simple output format")
	listCmd.Flags().IntVar(&listPage, "page", 1, "Page number (1-indexed)")
	listCmd.Flags().IntVar(&listSize, "size", 20, "Page size")
	listCmd.Flags().StringVar(&listSort, "sort", "updatedAt_DESC", "Sort order")
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
	client, err := newClient(listWorkspace)
	if err != nil {
		return err
	}
	projectID, err := client.ResolveProjectID(listWorkspace)
	if err != nil {
		return err
	}
	client.SetProject(projectID)
	filter := map[string]interface{}{"projectId": projectID}
	if listWiki != "" {
		switch listWiki {
		case "true":
			filter["wiki"] = true
		case "false":
			filter["wiki"] = false
		default:
			return fmt.Errorf("--wiki must be true or false")
		}
	}
	query := fmt.Sprintf(`
		query ListDocuments($filter: DocumentFilterInput!, $sort: [DocumentSort!], $skip: Int, $take: Int) {
			documents(filter: $filter, sort: $sort, skip: $skip, take: $take) {
				items { %s }
				pageInfo { totalItems hasNextPage }
			}
		}
	`, documentFields)
	variables := map[string]interface{}{"filter": filter, "sort": []string{listSort}, "skip": (listPage - 1) * listSize, "take": listSize}
	var response struct {
		Documents struct {
			Items    []Document `json:"items"`
			PageInfo struct {
				TotalItems  int  `json:"totalItems"`
				HasNextPage bool `json:"hasNextPage"`
			} `json:"pageInfo"`
		} `json:"documents"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to list documents: %w", err)
	}
	if listFormat == "json" {
		return printJSON(response.Documents.Items)
	}
	if listSimple {
		for _, d := range response.Documents.Items {
			fmt.Printf("%s  %s\n", d.ID, d.Title)
		}
		return nil
	}
	fmt.Printf("=== Documents (page %d, %d shown, %d total) ===\n\n", listPage, len(response.Documents.Items), response.Documents.PageInfo.TotalItems)
	for i, d := range response.Documents.Items {
		fmt.Printf("%d. ", i+1)
		printDocument(d, false)
		fmt.Println()
	}
	if response.Documents.PageInfo.HasNextPage {
		fmt.Printf("More results available. Pass --page %d to fetch the next page.\n", listPage+1)
	}
	return nil
}
