package records

import (
	"fmt"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a record",
	Long:  "Permanently delete a record. This action cannot be undone.",
	Example: `  blue records delete --record <id> --confirm
  blue records delete -r <id> -y`,
	RunE: runDelete,
}

var (
	deleteRecord  string
	deleteConfirm bool
)

func init() {
	deleteCmd.Flags().StringVarP(&deleteRecord, "record", "r", "", "Record ID to delete (required)")
	deleteCmd.Flags().BoolVarP(&deleteConfirm, "confirm", "y", false, "Confirm deletion (required for safety)")
}

func runDelete(cmd *cobra.Command, args []string) error {
	if deleteRecord == "" {
		return fmt.Errorf("record ID is required. Use --record flag")
	}
	if !deleteConfirm {
		return fmt.Errorf("deletion confirmation is required. Use --confirm flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := common.NewClient(config)

	mutation := `
		mutation DeleteTodo($input: DeleteTodoInput!) {
			deleteTodo(input: $input) {
				success
				operationId
			}
		}
	`

	variables := map[string]interface{}{
		"input": common.DeleteTodoInput{
			TodoID: deleteRecord,
		},
	}

	data, err := client.ExecuteQuery(mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}

	if deleteTodoData, ok := data["deleteTodo"].(map[string]interface{}); ok {
		success, _ := deleteTodoData["success"].(bool)
		if success {
			fmt.Printf("Record %s deleted successfully\n", deleteRecord)
			return nil
		}
		return fmt.Errorf("failed to delete record %s", deleteRecord)
	}

	return fmt.Errorf("unexpected response format")
}
