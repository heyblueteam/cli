package reports

import (
	"fmt"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a report",
	Example: `  blue reports get --report <id>
  blue reports get --report <id> --format json`,
	RunE: runGet,
}

var (
	getReport string
	getFormat string
)

func init() {
	getCmd.Flags().StringVar(&getReport, "report", "", "Report ID (required)")
	getCmd.Flags().StringVar(&getFormat, "format", "", "Output format (json)")
}

func runGet(cmd *cobra.Command, args []string) error {
	if getReport == "" {
		return fmt.Errorf("report ID is required. Use --report flag")
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`query GetReport($id: String!) { report(id: $id) { %s } }`, reportFields)
	var response struct {
		Report Report `json:"report"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"id": getReport}, &response); err != nil {
		return fmt.Errorf("failed to get report: %w", err)
	}
	if getFormat == "json" {
		return printJSON(response.Report)
	}
	printReport(response.Report)
	return nil
}
