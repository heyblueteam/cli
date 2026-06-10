package exports

import (
	"fmt"

	"github.com/spf13/cobra"
)

var recordsCmd = &cobra.Command{
	Use:   "records",
	Short: "Export workspace records",
	Example: `  blue exports records --workspace <id>
  blue exports records --workspace <id> --done false --q launch --assignees u1,u2 --tags tag1
  blue exports records --workspace <id> --filter-json '{"hasDueDate":true}'`,
	RunE: runRecords,
}

var (
	recordsWorkspace  string
	recordsDone       string
	recordsQ          string
	recordsAssignees  string
	recordsTags       string
	recordsLists      string
	recordsFilterJSON string
)

func init() {
	recordsCmd.Flags().StringVarP(&recordsWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	recordsCmd.Flags().StringVar(&recordsDone, "done", "", "Filter by completion status (true or false)")
	recordsCmd.Flags().StringVar(&recordsQ, "q", "", "Quick search query")
	recordsCmd.Flags().StringVar(&recordsAssignees, "assignees", "", "Comma-separated assignee IDs")
	recordsCmd.Flags().StringVar(&recordsTags, "tags", "", "Comma-separated tag IDs")
	recordsCmd.Flags().StringVar(&recordsLists, "lists", "", "Comma-separated list IDs")
	recordsCmd.Flags().StringVar(&recordsFilterJSON, "filter-json", "", "Raw TodosFilter JSON merged into the filter")
}

func runRecords(cmd *cobra.Command, args []string) error {
	if recordsWorkspace == "" {
		return fmt.Errorf("workspace ID is required. Use --workspace flag")
	}

	client, err := newClient()
	if err != nil {
		return err
	}
	client.SetProject(recordsWorkspace)
	projectID, err := client.ResolveProjectID(recordsWorkspace)
	if err != nil {
		return err
	}
	client.SetProject(projectID)
	companyID, err := client.ResolveCompanyID()
	if err != nil {
		return err
	}

	filter := map[string]interface{}{"companyIds": []string{companyID}}
	if recordsQ != "" {
		filter["q"] = recordsQ
	}
	if recordsAssignees != "" {
		filter["assigneeIds"] = splitCSV(recordsAssignees)
	}
	if recordsTags != "" {
		filter["tagIds"] = splitCSV(recordsTags)
	}
	if recordsLists != "" {
		filter["todoListIds"] = splitCSV(recordsLists)
	}
	if recordsDone != "" {
		switch recordsDone {
		case "true":
			filter["done"] = true
		case "false":
			filter["done"] = false
		default:
			return fmt.Errorf("--done must be true or false")
		}
	}
	if recordsFilterJSON != "" {
		parsed, err := parseJSONFlag("--filter-json", recordsFilterJSON)
		if err != nil {
			return err
		}
		parsedMap, ok := parsed.(map[string]interface{})
		if !ok {
			return fmt.Errorf("--filter-json must be a JSON object")
		}
		for key, value := range parsedMap {
			filter[key] = value
		}
	}

	query := `
		mutation ExportRecords($input: ExportTodosInput!) {
			exportTodos(input: $input)
		}
	`
	variables := map[string]interface{}{"input": map[string]interface{}{"projectId": projectID, "filter": filter}}

	var response struct {
		ExportTodos bool `json:"exportTodos"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to queue records export: %w", err)
	}
	if !response.ExportTodos {
		return fmt.Errorf("records export was not queued")
	}
	printQueued("Records")
	return nil
}
