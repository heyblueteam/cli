package lists

import (
	"fmt"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

type DeleteTodoListResponse struct {
	DeleteTodoList common.MutationResult `json:"deleteTodoList"`
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a list",
	Long:  "Permanently delete a list from a workspace. This action cannot be undone.",
	Example: `  blue lists delete --workspace <id> --list <id> --confirm`,
	RunE: runDelete,
}

var (
	deleteWorkspace string
	deleteList      string
	deleteConfirm   bool
	deleteSimple    bool
)

func init() {
	deleteCmd.Flags().StringVarP(&deleteWorkspace, "workspace", "w", "", "Workspace ID (required)")
	deleteCmd.Flags().StringVarP(&deleteList, "list", "l", "", "List ID (required)")
	deleteCmd.Flags().BoolVarP(&deleteConfirm, "confirm", "y", false, "Confirm deletion (required for safety)")
	deleteCmd.Flags().BoolVarP(&deleteSimple, "simple", "s", false, "Simple output format")
}

func runDelete(cmd *cobra.Command, args []string) error {
	if deleteList == "" {
		return fmt.Errorf("list ID is required. Use --list flag")
	}
	if deleteWorkspace == "" {
		return fmt.Errorf("workspace ID is required. Use --workspace flag")
	}
	if !deleteConfirm {
		return fmt.Errorf("deletion confirmation is required. Use --confirm flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)
	client.SetProject(deleteWorkspace)

	input := map[string]string{
		"projectId":  deleteWorkspace,
		"todoListId": deleteList,
	}

	mutation := `
		mutation DeleteTodoList($input: DeleteTodoListInput!) {
			deleteTodoList(input: $input) {
				success
				operationId
			}
		}`

	variables := map[string]interface{}{
		"input": input,
	}

	var response DeleteTodoListResponse
	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to delete list: %w", err)
	}

	result := response.DeleteTodoList
	if result.Success {
		if deleteSimple {
			fmt.Printf("List %s deleted\n", deleteList)
		} else {
			fmt.Printf("=== List Deleted Successfully ===\n")
			fmt.Printf("List ID: %s\n", deleteList)
			fmt.Printf("Workspace ID: %s\n", deleteWorkspace)
		}
	} else {
		return fmt.Errorf("failed to delete list %s", deleteList)
	}

	return nil
}
