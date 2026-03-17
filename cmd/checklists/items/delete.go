package items

import (
	"fmt"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a checklist item",
	Long:  "Permanently delete a checklist item.",
	Example: `  blue checklists items delete --item <id> --confirm
  blue checklists items delete --item <id> --workspace <id> --confirm --simple`,
	RunE: runDelete,
}

var (
	deleteItem      string
	deleteWorkspace string
	deleteConfirm   bool
	deleteSimple    bool
)

func init() {
	deleteCmd.Flags().StringVar(&deleteItem, "item", "", "Checklist item ID to delete (required)")
	deleteCmd.Flags().StringVarP(&deleteWorkspace, "workspace", "w", "", "Workspace ID or slug (optional)")
	deleteCmd.Flags().BoolVarP(&deleteConfirm, "confirm", "y", false, "Confirm deletion (required for safety)")
	deleteCmd.Flags().BoolVarP(&deleteSimple, "simple", "s", false, "Simple output format")
}

func runDelete(cmd *cobra.Command, args []string) error {
	if deleteItem == "" {
		return fmt.Errorf("checklist item ID is required. Use --item flag")
	}
	if !deleteConfirm {
		return fmt.Errorf("deletion requires --confirm flag for safety")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)

	if deleteWorkspace != "" {
		client.SetProject(deleteWorkspace)
	}

	if !deleteSimple {
		fmt.Printf("=== Deleting Checklist Item ===\n")
		fmt.Printf("Item ID: %s\n", deleteItem)
		if deleteWorkspace != "" {
			fmt.Printf("Workspace: %s\n", deleteWorkspace)
		}
		fmt.Printf("WARNING: This will permanently delete the checklist item!\n\n")
	}

	mutation := `
		mutation DeleteChecklistItem($id: String!) {
			deleteChecklistItem(id: $id)
		}
	`

	variables := map[string]interface{}{
		"id": deleteItem,
	}

	var response struct {
		DeleteChecklistItem bool `json:"deleteChecklistItem"`
	}

	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to delete checklist item: %w", err)
	}

	if deleteSimple {
		if response.DeleteChecklistItem {
			fmt.Printf("Checklist item deleted: %s\n", deleteItem)
		} else {
			fmt.Printf("Failed to delete checklist item: %s\n", deleteItem)
		}
	} else {
		if response.DeleteChecklistItem {
			fmt.Printf("=== Checklist Item Deleted Successfully ===\n")
			fmt.Printf("Item ID: %s\n", deleteItem)
			fmt.Printf("Checklist item has been permanently deleted.\n")
		} else {
			fmt.Printf("Failed to delete checklist item %s\n", deleteItem)
		}
	}

	return nil
}
