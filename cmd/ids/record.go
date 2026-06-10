package ids

import (
	"fmt"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

var recordCmd = &cobra.Command{
	Use:     "record",
	Aliases: []string{"records", "rec"},
	Short:   "Look up record IDs",
	Example: `  blue ids record --workspace <id>
  blue ids record --workspace <id> --search "Launch plan"`,
	RunE: runRecordIDs,
}

func init() {
	addCommonFlags(recordCmd, true)
}

func runRecordIDs(cmd *cobra.Command, args []string) error {
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

	query := `query RecordIDs($filter: TodosFilter!, $limit: Int!) {
		todoQueries {
			todos(filter: $filter, limit: $limit) {
				items { id uid title todoList { id title } }
			}
		}
	}`
	filter := map[string]interface{}{
		"companyIds": []string{},
		"projectIds": []string{projectID},
	}
	if idsSearch != "" {
		filter["q"] = idsSearch
	}
	variables := map[string]interface{}{"filter": filter, "limit": idsLimit}

	var response struct {
		TodoQueries struct {
			Todos struct {
				Items []common.Record `json:"items"`
			} `json:"todos"`
		} `json:"todoQueries"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to look up record IDs: %w", err)
	}

	var rows []idRow
	for _, record := range response.TodoQueries.Todos.Items {
		if !matchesSearch(record.Title) {
			continue
		}
		extra := ""
		if record.TodoList != nil {
			extra = "list:" + record.TodoList.Title + " " + record.TodoList.ID
		}
		rows = append(rows, idRow{Type: "record", Name: record.Title, ID: record.ID, UID: record.UID, Extra: extra, Workspace: projectID})
	}
	return printRows(limitRows(rows))
}
