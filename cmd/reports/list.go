package reports

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List reports",
	Example: `  blue reports list --simple
  blue reports list --page 2 --size 50 --format json`,
	RunE: runList,
}

var (
	listSimple bool
	listPage   int
	listSize   int
	listFormat string
)

func init() {
	listCmd.Flags().BoolVarP(&listSimple, "simple", "s", false, "Simple output format")
	listCmd.Flags().IntVar(&listPage, "page", 1, "Page number (1-indexed)")
	listCmd.Flags().IntVar(&listSize, "size", 20, "Page size")
	listCmd.Flags().StringVar(&listFormat, "format", "", "Output format (json)")
}

func runList(cmd *cobra.Command, args []string) error {
	if listPage < 1 {
		listPage = 1
	}
	if listSize < 1 {
		listSize = 20
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`
		query ListReports($filter: ReportFilter!, $skip: Int, $take: Int) {
			reports(filter: $filter, skip: $skip, take: $take) {
				totalCount
				hasNextPage
				hasPreviousPage
				items { %s }
			}
		}
	`, reportFields)
	variables := map[string]interface{}{
		"filter": map[string]interface{}{"companyId": client.GetCompanyID()},
		"skip":   (listPage - 1) * listSize,
		"take":   listSize,
	}
	var response struct {
		Reports struct {
			TotalCount  int      `json:"totalCount"`
			HasNextPage bool     `json:"hasNextPage"`
			Items       []Report `json:"items"`
		} `json:"reports"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to list reports: %w", err)
	}
	if listFormat == "json" {
		return printJSON(response.Reports.Items)
	}
	if listSimple {
		for _, r := range response.Reports.Items {
			fmt.Printf("%s  %s\n", r.ID, r.Title)
		}
		return nil
	}
	fmt.Printf("=== Reports (page %d, %d shown, %d total) ===\n\n", listPage, len(response.Reports.Items), response.Reports.TotalCount)
	for i, r := range response.Reports.Items {
		fmt.Printf("%d. ", i+1)
		printReport(r)
		fmt.Println()
	}
	if response.Reports.HasNextPage {
		fmt.Printf("More results available. Pass --page %d to fetch the next page.\n", listPage+1)
	}
	return nil
}
