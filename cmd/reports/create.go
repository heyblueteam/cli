package reports

import (
	"fmt"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a report",
	Example: `  blue reports create --title "Open work"
  blue reports create --title "Open work" --workspaces ws1,ws2 --filter-json '{"done":false}'
  blue reports create --title "Custom" --data-sources-json '[{"sourceType":"TODOS"}]'`,
	RunE: runCreate,
}

var (
	createTitle           string
	createDescription     string
	createWorkspaces      string
	createFilterJSON      string
	createDataSourcesJSON string
	createFormat          string
)

func init() {
	createCmd.Flags().StringVarP(&createTitle, "title", "t", "", "Report title (required)")
	createCmd.Flags().StringVar(&createDescription, "description", "", "Report description")
	createCmd.Flags().StringVar(&createWorkspaces, "workspaces", "", "Comma-separated workspace IDs or slugs for the default data source")
	createCmd.Flags().StringVar(&createFilterJSON, "filter-json", "", "Filter JSON for the default data source")
	createCmd.Flags().StringVar(&createDataSourcesJSON, "data-sources-json", "", "Raw dataSources JSON; overrides --workspaces and --filter-json")
	createCmd.Flags().StringVar(&createFormat, "format", "", "Output format (json)")
}

func runCreate(cmd *cobra.Command, args []string) error {
	if createTitle == "" {
		return fmt.Errorf("title is required. Use --title flag")
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	input := map[string]interface{}{"title": createTitle}
	if createDescription != "" {
		input["description"] = createDescription
	}
	if createDataSourcesJSON != "" {
		dataSources, err := parseJSONFlag("--data-sources-json", createDataSourcesJSON)
		if err != nil {
			return err
		}
		input["dataSources"] = dataSources
	} else {
		dataSource := map[string]interface{}{"sourceType": "TODOS"}
		projectIDs, err := resolveProjectIDs(client, createWorkspaces)
		if err != nil {
			return err
		}
		if len(projectIDs) > 0 {
			dataSource["projectIds"] = projectIDs
		}
		filter, err := parseJSONFlag("--filter-json", createFilterJSON)
		if err != nil {
			return err
		}
		if filter != nil {
			dataSource["filters"] = filter
		}
		input["dataSources"] = []interface{}{dataSource}
	}
	query := fmt.Sprintf(`mutation CreateReport($input: CreateReportInput!) { createReport(input: $input) { %s } }`, reportFields)
	var response struct {
		CreateReport Report `json:"createReport"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"input": input}, &response); err != nil {
		return fmt.Errorf("failed to create report: %w", err)
	}
	if createFormat == "json" {
		return printJSON(response.CreateReport)
	}
	fmt.Println("Report created")
	printReport(response.CreateReport)
	return nil
}
