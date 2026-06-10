package reports

import (
	"fmt"

	"github.com/spf13/cobra"
)

var duplicateCmd = &cobra.Command{
	Use:   "duplicate",
	Short: "Duplicate a report",
	Example: `  blue reports duplicate --report <id>
  blue reports duplicate --report <id> --title "Working copy"`,
	RunE: runDuplicate,
}

var (
	duplicateReport string
	duplicateTitle  string
)

func init() {
	duplicateCmd.Flags().StringVar(&duplicateReport, "report", "", "Report ID (required)")
	duplicateCmd.Flags().StringVar(&duplicateTitle, "title", "", "Title for the copy")
}

func runDuplicate(cmd *cobra.Command, args []string) error {
	if duplicateReport == "" {
		return fmt.Errorf("report ID is required. Use --report flag")
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	input := map[string]interface{}{}
	if duplicateTitle != "" {
		input["title"] = duplicateTitle
	}
	query := fmt.Sprintf(`mutation DuplicateReport($id: String!, $input: DuplicateReportInput) { duplicateReport(id: $id, input: $input) { %s } }`, reportFields)
	var response struct {
		DuplicateReport Report `json:"duplicateReport"`
	}
	variables := map[string]interface{}{"id": duplicateReport, "input": input}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to duplicate report: %w", err)
	}
	fmt.Println("Report duplicated")
	printReport(response.DuplicateReport)
	return nil
}
