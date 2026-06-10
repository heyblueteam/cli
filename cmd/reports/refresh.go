package reports

import (
	"fmt"

	"github.com/spf13/cobra"
)

var refreshCmd = &cobra.Command{
	Use:     "refresh",
	Short:   "Refresh report aggregation cache",
	Example: `  blue reports refresh --report <id>`,
	RunE:    runRefresh,
}

var refreshReport string

func init() {
	refreshCmd.Flags().StringVar(&refreshReport, "report", "", "Report ID (required)")
}

func runRefresh(cmd *cobra.Command, args []string) error {
	if refreshReport == "" {
		return fmt.Errorf("report ID is required. Use --report flag")
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	query := `mutation RefreshReport($reportId: String!) { refreshReportAggregations(reportId: $reportId) { id lastGeneratedAt } }`
	var response struct {
		RefreshReportAggregations Report `json:"refreshReportAggregations"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"reportId": refreshReport}, &response); err != nil {
		return fmt.Errorf("failed to refresh report aggregations: %w", err)
	}
	fmt.Printf("Report aggregations invalidated at %s\n", response.RefreshReportAggregations.LastGeneratedAt)
	return nil
}
