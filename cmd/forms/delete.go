package forms

import (
	"fmt"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a form",
	Example: `  blue forms delete --form <id> --confirm`,
	RunE: runDelete,
}

var (
	deleteForm      string
	deleteWorkspace string
	deleteConfirm   bool
)

func init() {
	deleteCmd.Flags().StringVarP(&deleteForm, "form", "f", "", "Form ID (required)")
	deleteCmd.Flags().StringVarP(&deleteWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	deleteCmd.Flags().BoolVarP(&deleteConfirm, "confirm", "y", false, "Skip confirmation prompt (required)")
}

func runDelete(cmd *cobra.Command, args []string) error {
	if deleteForm == "" {
		return fmt.Errorf("form ID is required. Use --form flag")
	}
	if deleteWorkspace == "" {
		return fmt.Errorf("workspace is required. Use --workspace flag")
	}
	if !deleteConfirm {
		return fmt.Errorf("destructive operation — pass --confirm to proceed")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)
	client.SetProject(deleteWorkspace)

	mutation := `
		mutation DeleteForm($id: String!) {
			deleteForm(id: $id)
		}
	`
	if _, err := client.ExecuteQuery(mutation, map[string]interface{}{"id": deleteForm}); err != nil {
		return fmt.Errorf("deleteForm failed: %w", err)
	}
	fmt.Printf("Deleted form %s\n", deleteForm)
	return nil
}
