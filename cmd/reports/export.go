package reports

import (
	"fmt"

	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:     "export",
	Short:   "Queue a report CSV export",
	Example: `  blue reports export --report <id>`,
	RunE:    runExport,
}

var exportReport string

func init() {
	exportCmd.Flags().StringVar(&exportReport, "report", "", "Report ID (required)")
}

func runExport(cmd *cobra.Command, args []string) error {
	if exportReport == "" {
		return fmt.Errorf("report ID is required. Use --report flag")
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	query := `mutation ExportReport($input: ExportReportInput!) { exportReport(input: $input) }`
	var response struct {
		ExportReport bool `json:"exportReport"`
	}
	variables := map[string]interface{}{"input": map[string]interface{}{"reportId": exportReport}}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to queue report export: %w", err)
	}
	if !response.ExportReport {
		return fmt.Errorf("report export was not queued")
	}
	fmt.Println("Report export queued. Blue will email the finished CSV to your account.")
	return nil
}
