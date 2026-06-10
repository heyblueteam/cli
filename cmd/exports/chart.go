package exports

import (
	"fmt"

	"github.com/spf13/cobra"
)

var chartCmd = &cobra.Command{
	Use:   "chart",
	Short: "Export chart data",
	Example: `  blue exports chart --chart <id>
  blue exports chart --chart <id> --filter-json '{"showCompleted":false}'`,
	RunE: runChart,
}

var (
	chartID         string
	chartFilterJSON string
)

func init() {
	chartCmd.Flags().StringVar(&chartID, "chart", "", "Chart ID (required)")
	chartCmd.Flags().StringVar(&chartFilterJSON, "filter-json", "", "Raw TodoFilterInput JSON")
}

func runChart(cmd *cobra.Command, args []string) error {
	if chartID == "" {
		return fmt.Errorf("chart ID is required. Use --chart flag")
	}
	filter, err := parseJSONFlag("--filter-json", chartFilterJSON)
	if err != nil {
		return err
	}

	client, err := newClient()
	if err != nil {
		return err
	}
	query := `mutation ExportChart($chartId: ID!, $filter: TodoFilterInput) { exportChartCSV(chartId: $chartId, filter: $filter) }`
	variables := map[string]interface{}{"chartId": chartID, "filter": filter}
	var response struct {
		ExportChartCSV bool `json:"exportChartCSV"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to queue chart export: %w", err)
	}
	if !response.ExportChartCSV {
		return fmt.Errorf("chart export was not queued")
	}
	printQueued("Chart")
	return nil
}
