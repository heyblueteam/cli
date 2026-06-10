package ids

import (
	"fmt"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

var fieldCmd = &cobra.Command{
	Use:     "field",
	Aliases: []string{"fields", "cf"},
	Short:   "Look up custom field IDs",
	Example: `  blue ids field --workspace <id>
  blue ids field --workspace <id> --search Priority`,
	RunE: runFieldIDs,
}

func init() {
	addCommonFlags(fieldCmd, true)
}

func runFieldIDs(cmd *cobra.Command, args []string) error {
	if err := requireWorkspace(); err != nil {
		return err
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	client.SetProject(idsWorkspace)
	projectID, err := client.ResolveProjectID(idsWorkspace)
	if err != nil {
		return fmt.Errorf("failed to resolve workspace: %w", err)
	}
	client.SetProject(projectID)

	query := `query FieldIDs($projectId: String!, $take: Int!) {
		customFields(filter: { projectId: $projectId }, skip: 0, take: $take) {
			items { id uid name type customFieldOptions { id title } }
		}
	}`
	variables := map[string]interface{}{"projectId": projectID, "take": idsLimit}

	var response struct {
		CustomFields struct {
			Items []common.CustomField `json:"items"`
		} `json:"customFields"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to look up field IDs: %w", err)
	}

	var rows []idRow
	for _, field := range response.CustomFields.Items {
		if !matchesSearch(field.Name, field.Type) {
			continue
		}
		rows = append(rows, idRow{Type: "field", Name: field.Name, ID: field.ID, UID: field.UID, Extra: "type:" + field.Type, Workspace: projectID})
		for _, option := range field.Options {
			if matchesSearch(option.Title, field.Name) {
				rows = append(rows, idRow{Type: "field-option", Name: field.Name + " / " + option.Title, ID: option.ID, Extra: "field:" + field.ID, Workspace: projectID})
			}
		}
	}
	return printRows(limitRows(rows))
}
