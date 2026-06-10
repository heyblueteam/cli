package exports

import (
	"fmt"

	"github.com/spf13/cobra"
)

var reportCmd = &cobra.Command{
	Use:     "report",
	Short:   "Export a report",
	Example: `  blue exports report --report <id>`,
	RunE:    runReport,
}

var reportID string

func init() {
	reportCmd.Flags().StringVar(&reportID, "report", "", "Report ID (required)")
}

func runReport(cmd *cobra.Command, args []string) error {
	if reportID == "" {
		return fmt.Errorf("report ID is required. Use --report flag")
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	query := `mutation ExportReport($input: ExportReportInput!) { exportReport(input: $input) }`
	variables := map[string]interface{}{"input": map[string]interface{}{"reportId": reportID}}
	var response struct {
		ExportReport bool `json:"exportReport"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to queue report export: %w", err)
	}
	if !response.ExportReport {
		return fmt.Errorf("report export was not queued")
	}
	printQueued("Report")
	return nil
}
