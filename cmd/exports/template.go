package exports

import (
	"fmt"

	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:     "template",
	Short:   "Export a CSV import template",
	Example: `  blue exports template --workspace <id>`,
	RunE:    runTemplate,
}

var templateWorkspace string

func init() {
	templateCmd.Flags().StringVarP(&templateWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
}

func runTemplate(cmd *cobra.Command, args []string) error {
	if templateWorkspace == "" {
		return fmt.Errorf("workspace ID is required. Use --workspace flag")
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	client.SetProject(templateWorkspace)
	projectID, err := client.ResolveProjectID(templateWorkspace)
	if err != nil {
		return err
	}
	client.SetProject(projectID)
	query := `mutation ExportTemplate($input: ExportCSVTemplateInput!) { exportCSVTemplate(input: $input) }`
	variables := map[string]interface{}{"input": map[string]interface{}{"projectId": projectID}}
	var response struct {
		ExportCSVTemplate string `json:"exportCSVTemplate"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to export CSV template: %w", err)
	}
	fmt.Println(response.ExportCSVTemplate)
	return nil
}
