package ids

import (
	"fmt"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

var workspaceCmd = &cobra.Command{
	Use:     "workspace",
	Aliases: []string{"workspaces", "ws"},
	Short:   "Look up workspace IDs",
	Example: `  blue ids workspace
  blue ids workspace --search CRM
  blue ids workspace --format json`,
	RunE: runWorkspaceIDs,
}

func init() {
	addCommonFlags(workspaceCmd, false)
}

func runWorkspaceIDs(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	query := `query WorkspaceIDs($companyId: String!, $search: String, $take: Int!) {
		projectList(
			filter: { companyIds: [$companyId], search: $search, archived: false, isTemplate: false }
			skip: 0
			take: $take
			sort: [name_ASC]
		) {
			items { id uid slug name }
		}
	}`
	variables := map[string]interface{}{
		"companyId": client.GetCompanyID(),
		"search":    nil,
		"take":      idsLimit,
	}
	if idsSearch != "" {
		variables["search"] = idsSearch
	}

	var response struct {
		ProjectList struct {
			Items []common.Project `json:"items"`
		} `json:"projectList"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to look up workspace IDs: %w", err)
	}

	rows := make([]idRow, 0, len(response.ProjectList.Items))
	for _, workspace := range response.ProjectList.Items {
		if !matchesSearch(workspace.Name, workspace.Slug) {
			continue
		}
		rows = append(rows, idRow{Type: "workspace", Name: workspace.Name, ID: workspace.ID, UID: workspace.UID, Extra: "slug:" + workspace.Slug})
	}
	return printRows(limitRows(rows))
}
