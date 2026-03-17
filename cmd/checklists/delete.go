package checklists

import (
	"fmt"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a checklist",
	Long:  "Permanently delete a checklist and all its items.",
	Example: `  blue checklists delete --checklist <id> --confirm
  blue checklists delete --checklist <id> --workspace <id> --confirm --simple`,
	RunE: runDelete,
}

var (
	deleteChecklist string
	deleteWorkspace string
	deleteConfirm   bool
	deleteSimple    bool
)

func init() {
	deleteCmd.Flags().StringVar(&deleteChecklist, "checklist", "", "Checklist ID to delete (required)")
	deleteCmd.Flags().StringVarP(&deleteWorkspace, "workspace", "w", "", "Workspace ID or slug (optional)")
	deleteCmd.Flags().BoolVarP(&deleteConfirm, "confirm", "y", false, "Confirm deletion (required for safety)")
	deleteCmd.Flags().BoolVarP(&deleteSimple, "simple", "s", false, "Simple output format")
}

func runDelete(cmd *cobra.Command, args []string) error {
	if deleteChecklist == "" {
		return fmt.Errorf("checklist ID is required. Use --checklist flag")
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
		fmt.Printf("=== Deleting Checklist ===\n")
		fmt.Printf("Checklist ID: %s\n", deleteChecklist)
		if deleteWorkspace != "" {
			fmt.Printf("Workspace: %s\n", deleteWorkspace)
		}
		fmt.Printf("WARNING: This will permanently delete the checklist and all its items!\n\n")
	}

	mutation := `
		mutation DeleteChecklist($id: String!) {
			deleteChecklist(id: $id)
		}
	`

	variables := map[string]interface{}{
		"id": deleteChecklist,
	}

	var response struct {
		DeleteChecklist bool `json:"deleteChecklist"`
	}

	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to delete checklist: %w", err)
	}

	if deleteSimple {
		if response.DeleteChecklist {
			fmt.Printf("Checklist deleted: %s\n", deleteChecklist)
		} else {
			fmt.Printf("Failed to delete checklist: %s\n", deleteChecklist)
		}
	} else {
		if response.DeleteChecklist {
			fmt.Printf("=== Checklist Deleted Successfully ===\n")
			fmt.Printf("Checklist ID: %s\n", deleteChecklist)
			fmt.Printf("Checklist and all its items have been permanently deleted.\n")
		} else {
			fmt.Printf("Failed to delete checklist %s\n", deleteChecklist)
		}
	}

	return nil
}
