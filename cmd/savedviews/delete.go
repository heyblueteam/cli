package savedviews

import (
	"fmt"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete a saved view",
	Example: `  blue saved-views delete --view <id> --confirm`,
	RunE:    runDelete,
}

var (
	deleteView    string
	deleteConfirm bool
)

func init() {
	deleteCmd.Flags().StringVar(&deleteView, "view", "", "Saved view ID (required)")
	deleteCmd.Flags().BoolVarP(&deleteConfirm, "confirm", "y", false, "Confirm deletion (required)")
}

func runDelete(cmd *cobra.Command, args []string) error {
	if deleteView == "" {
		return fmt.Errorf("saved view ID is required. Use --view flag")
	}
	if !deleteConfirm {
		return fmt.Errorf("deletion confirmation is required. Use --confirm flag")
	}
	client, err := clientFor("")
	if err != nil {
		return err
	}
	query := `mutation DeleteSavedView($id: String!) { deleteSavedView(id: $id) { success operationId } }`
	var response struct {
		DeleteSavedView struct {
			Success bool `json:"success"`
		} `json:"deleteSavedView"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"id": deleteView}, &response); err != nil {
		return fmt.Errorf("failed to delete saved view: %w", err)
	}
	if !response.DeleteSavedView.Success {
		return fmt.Errorf("saved view was not deleted")
	}
	common.PrintSuccess(fmt.Sprintf("Deleted saved view %s", deleteView))
	return nil
}
