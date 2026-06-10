package ids

import (
	"fmt"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"lists"},
	Short:   "Look up list IDs",
	Example: `  blue ids list --workspace <id>
  blue ids list --workspace <id> --search Done`,
	RunE: runListIDs,
}

func init() {
	addCommonFlags(listCmd, true)
}

func runListIDs(cmd *cobra.Command, args []string) error {
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

	query := `query ListIDs($projectId: String!) {
		todoLists(projectId: $projectId) { id uid title position todosCount }
	}`
	variables := map[string]interface{}{"projectId": projectID}

	var response struct {
		TodoLists []common.TodoList `json:"todoLists"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to look up list IDs: %w", err)
	}

	var rows []idRow
	for _, list := range response.TodoLists {
		if !matchesSearch(list.Title) {
			continue
		}
		rows = append(rows, idRow{Type: "list", Name: list.Title, ID: list.ID, UID: list.UID, Extra: fmt.Sprintf("records:%d", list.TodosCount), Workspace: projectID})
	}
	return printRows(limitRows(rows))
}
