package dependencies

import (
	"fmt"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a dependency between records",
	Long:  "Permanently delete a dependency relationship between two records.",
	Example: `  blue dependencies delete --record <id> --other-record <id> --confirm
  blue deps delete -r <id> --other-record <id> -y -s`,
	RunE: runDelete,
}

var (
	deleteRecord      string
	deleteOtherRecord string
	deleteConfirm     bool
	deleteWorkspace   string
	deleteSimple      bool
)

func init() {
	deleteCmd.Flags().StringVarP(&deleteRecord, "record", "r", "", "Record/Todo ID (required)")
	deleteCmd.Flags().StringVar(&deleteOtherRecord, "other-record", "", "Other Record/Todo ID (required)")
	deleteCmd.Flags().BoolVarP(&deleteConfirm, "confirm", "y", false, "Confirm deletion (required for safety)")
	deleteCmd.Flags().StringVarP(&deleteWorkspace, "workspace", "w", "", "Workspace ID or slug (optional - for context)")
	deleteCmd.Flags().BoolVarP(&deleteSimple, "simple", "s", false, "Show simple output")
}

func runDelete(cmd *cobra.Command, args []string) error {
	if deleteRecord == "" {
		return fmt.Errorf("record ID is required. Use --record flag")
	}
	if deleteOtherRecord == "" {
		return fmt.Errorf("other record ID is required. Use --other-record flag")
	}
	if !deleteConfirm {
		return fmt.Errorf("deletion confirmation is required. Use --confirm flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := common.NewClient(config)
	if deleteWorkspace != "" {
		client.SetProject(deleteWorkspace)
	}

	mutation := `
		mutation DeleteTodoDependency($input: DeleteTodoDependencyInput!) {
			deleteTodoDependency(input: $input)
		}
	`

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"todoId":      deleteRecord,
			"otherTodoId": deleteOtherRecord,
		},
	}

	var response struct {
		DeleteTodoDependency bool `json:"deleteTodoDependency"`
	}

	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to delete dependency: %w", err)
	}

	if deleteSimple {
		fmt.Printf("Dependency deleted: %s <-> %s\n", deleteRecord, deleteOtherRecord)
	} else {
		fmt.Printf("=== Dependency Deleted Successfully ===\n")
		fmt.Printf("Record: %s\n", deleteRecord)
		fmt.Printf("Other Record: %s\n", deleteOtherRecord)
		fmt.Printf("Result: %v\n", response.DeleteTodoDependency)
		fmt.Printf("\nDependency deleted successfully!\n")
	}

	return nil
}
