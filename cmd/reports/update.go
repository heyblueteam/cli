package reports

import (
	"fmt"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update report metadata or configuration",
	Example: `  blue reports update --report <id> --title "New title"
  blue reports update --report <id> --config-json '{"layout":"table"}'
  blue reports update --report <id> --data-sources-json '[{"sourceType":"TODOS"}]'`,
	RunE: runUpdate,
}

var (
	updateReport          string
	updateTitle           string
	updateDescription     string
	updateConfigJSON      string
	updateDataSourcesJSON string
	updateFormat          string
)

func init() {
	updateCmd.Flags().StringVar(&updateReport, "report", "", "Report ID (required)")
	updateCmd.Flags().StringVar(&updateTitle, "title", "", "New report title")
	updateCmd.Flags().StringVar(&updateDescription, "description", "", "New report description")
	updateCmd.Flags().StringVar(&updateConfigJSON, "config-json", "", "Raw config JSON")
	updateCmd.Flags().StringVar(&updateDataSourcesJSON, "data-sources-json", "", "Raw replacement dataSources JSON")
	updateCmd.Flags().StringVar(&updateFormat, "format", "", "Output format (json)")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if updateReport == "" {
		return fmt.Errorf("report ID is required. Use --report flag")
	}
	input := map[string]interface{}{}
	if updateTitle != "" {
		input["title"] = updateTitle
	}
	if updateDescription != "" {
		input["description"] = updateDescription
	}
	if updateConfigJSON != "" {
		config, err := parseJSONFlag("--config-json", updateConfigJSON)
		if err != nil {
			return err
		}
		input["config"] = config
	}
	if updateDataSourcesJSON != "" {
		dataSources, err := parseJSONFlag("--data-sources-json", updateDataSourcesJSON)
		if err != nil {
			return err
		}
		input["dataSources"] = dataSources
	}
	if len(input) == 0 {
		return fmt.Errorf("nothing to update. Pass at least one field flag")
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`mutation UpdateReport($id: String!, $input: UpdateReportInput!) { updateReport(id: $id, input: $input) { %s } }`, reportFields)
	var response struct {
		UpdateReport Report `json:"updateReport"`
	}
	variables := map[string]interface{}{"id": updateReport, "input": input}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to update report: %w", err)
	}
	if updateFormat == "json" {
		return printJSON(response.UpdateReport)
	}
	printReport(response.UpdateReport)
	return nil
}
