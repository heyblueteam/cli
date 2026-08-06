package charts

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var getChart, getFormat string
var getCmd = &cobra.Command{Use: "get", Short: "Get a chart and its current configuration", Example: `  blue charts get --chart <id> --format json`, RunE: runGet}

func init() {
	getCmd.Flags().StringVar(&getChart, "chart", "", "Chart ID (required)")
	getCmd.Flags().StringVar(&getFormat, "format", "", "Output format (json)")
}
func runGet(cmd *cobra.Command, args []string) error {
	if getChart == "" {
		return fmt.Errorf("chart ID is required. Use --chart")
	}
	client, err := chartClient()
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`query Chart($id:String!){ chart(id:$id){ %s } }`, chartFields)
	var response struct {
		Chart Chart `json:"chart"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"id": getChart}, &response); err != nil {
		return fmt.Errorf("failed to get chart: %w", err)
	}
	if strings.EqualFold(getFormat, "json") {
		return printJSON(response.Chart)
	}
	printChartSummary(response.Chart)
	fmt.Printf("Segments: %d\n", len(response.Chart.ChartSegments))
	return nil
}
