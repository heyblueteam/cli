package charts

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var listDashboard, listFormat string
var listSimple bool

var listCmd = &cobra.Command{
	Use: "list", Short: "List charts in a dashboard",
	Example: `  blue charts list --dashboard <id>
  blue charts list --dashboard <id> --format json`,
	RunE: runList,
}

func init() {
	listCmd.Flags().StringVar(&listDashboard, "dashboard", "", "Dashboard ID (required)")
	listCmd.Flags().BoolVarP(&listSimple, "simple", "s", false, "Simple output")
	listCmd.Flags().StringVar(&listFormat, "format", "", "Output format (json)")
}

func runList(cmd *cobra.Command, args []string) error {
	if listDashboard == "" {
		return fmt.Errorf("dashboard ID is required. Use --dashboard")
	}
	client, err := chartClient()
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`query Charts($filter: ChartFilterInput!) { charts(filter:$filter,sort:[position_ASC],take:100) { items { %s } pageInfo { hasNextPage totalItems } } }`, chartFields)
	var response struct {
		Charts struct {
			Items    []Chart `json:"items"`
			PageInfo struct {
				HasNextPage bool `json:"hasNextPage"`
				TotalItems  int  `json:"totalItems"`
			} `json:"pageInfo"`
		} `json:"charts"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"filter": map[string]interface{}{"dashboardId": listDashboard}}, &response); err != nil {
		return fmt.Errorf("failed to list charts: %w", err)
	}
	if strings.EqualFold(listFormat, "json") {
		return printJSON(response.Charts)
	}
	fmt.Printf("Charts: %d\n", response.Charts.PageInfo.TotalItems)
	for _, chart := range response.Charts.Items {
		printChartSummary(chart)
		if !listSimple {
			fmt.Printf("Segments: %d\n", len(chart.ChartSegments))
		}
		fmt.Println()
	}
	return nil
}
