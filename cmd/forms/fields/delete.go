package fields

import (
	"fmt"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a form field",
	Example: `  blue forms fields delete --field <ff-id> --confirm`,
	RunE: runDelete,
}

var (
	delField     string
	delWorkspace string
	delConfirm   bool
)

func init() {
	deleteCmd.Flags().StringVar(&delField, "field", "", "Form field ID (required)")
	deleteCmd.Flags().StringVarP(&delWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	deleteCmd.Flags().BoolVarP(&delConfirm, "confirm", "y", false, "Skip confirmation prompt (required)")
}

func runDelete(cmd *cobra.Command, args []string) error {
	if delField == "" {
		return fmt.Errorf("form field ID is required. Use --field flag")
	}
	if delWorkspace == "" {
		return fmt.Errorf("workspace is required. Use --workspace flag")
	}
	if !delConfirm {
		return fmt.Errorf("destructive operation — pass --confirm to proceed")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)
	client.SetProject(delWorkspace)

	mutation := `
		mutation DeleteFormField($id: String!) {
			deleteFormField(id: $id)
		}
	`
	if _, err := client.ExecuteQuery(mutation, map[string]interface{}{"id": delField}); err != nil {
		return fmt.Errorf("deleteFormField failed: %w", err)
	}
	fmt.Printf("Deleted form field %s\n", delField)
	return nil
}
