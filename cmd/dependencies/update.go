package dependencies

import (
	"fmt"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a dependency between records",
	Long:  "Update the type of an existing dependency relationship between two records.",
	Example: `  blue dependencies update --record <id> --other-record <id> --type BLOCKING
  blue deps update -r <id> --other-record <id> --type BLOCKED_BY -w <workspace> -s`,
	RunE: runUpdate,
}

var (
	updateRecord      string
	updateOtherRecord string
	updateType        string
	updateWorkspace   string
	updateSimple      bool
)

func init() {
	updateCmd.Flags().StringVarP(&updateRecord, "record", "r", "", "Record/Todo ID (required)")
	updateCmd.Flags().StringVar(&updateOtherRecord, "other-record", "", "Other Record/Todo ID (required)")
	updateCmd.Flags().StringVar(&updateType, "type", "", "New dependency type: BLOCKING or BLOCKED_BY (required)")
	updateCmd.Flags().StringVarP(&updateWorkspace, "workspace", "w", "", "Workspace ID or slug (optional - for context)")
	updateCmd.Flags().BoolVarP(&updateSimple, "simple", "s", false, "Show simple output")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if updateRecord == "" {
		return fmt.Errorf("record ID is required. Use --record flag")
	}
	if updateOtherRecord == "" {
		return fmt.Errorf("other record ID is required. Use --other-record flag")
	}
	if updateType == "" {
		return fmt.Errorf("dependency type is required. Use --type flag")
	}
	if updateType != "BLOCKING" && updateType != "BLOCKED_BY" {
		return fmt.Errorf("type must be BLOCKING or BLOCKED_BY")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := common.NewClient(config)
	if updateWorkspace != "" {
		client.SetProject(updateWorkspace)
	}

	mutation := `
		mutation UpdateTodoDependency($input: UpdateTodoDependencyInput!) {
			updateTodoDependency(input: $input) {
				id
				uid
				title
				dependOn {
					id
					uid
					title
				}
				dependBy {
					id
					uid
					title
				}
			}
		}
	`

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"type":        updateType,
			"todoId":      updateRecord,
			"otherTodoId": updateOtherRecord,
		},
	}

	var response struct {
		UpdateTodoDependency struct {
			ID       string           `json:"id"`
			UID      string           `json:"uid"`
			Title    string           `json:"title"`
			DependOn []DependencyTodo `json:"dependOn"`
			DependBy []DependencyTodo `json:"dependBy"`
		} `json:"updateTodoDependency"`
	}

	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to update dependency: %w", err)
	}

	todo := response.UpdateTodoDependency

	if updateSimple {
		fmt.Printf("Dependency updated: %s %s %s\n", updateRecord, updateType, updateOtherRecord)
	} else {
		fmt.Printf("=== Dependency Updated Successfully ===\n")
		fmt.Printf("Record: %s (%s)\n", todo.Title, todo.ID)
		fmt.Printf("Type: %s\n", updateType)
		fmt.Printf("Other Record: %s\n", updateOtherRecord)
		fmt.Println()

		if len(todo.DependOn) > 0 {
			fmt.Printf("Depends On (%d):\n", len(todo.DependOn))
			for _, dep := range todo.DependOn {
				fmt.Printf("  -> %s (%s)\n", dep.Title, dep.ID)
			}
		}
		if len(todo.DependBy) > 0 {
			fmt.Printf("Blocked By (%d):\n", len(todo.DependBy))
			for _, dep := range todo.DependBy {
				fmt.Printf("  <- %s (%s)\n", dep.Title, dep.ID)
			}
		}
		fmt.Printf("\nDependency updated successfully!\n")
	}

	return nil
}
