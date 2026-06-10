package savedviews

import (
	"fmt"

	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:     "apply",
	Short:   "Print a saved view configuration",
	Long:    "Print the saved view's viewConfig JSON so it can be inspected or reused in automation.",
	Example: `  blue saved-views apply --view <id>`,
	RunE:    runApply,
}

var applyView string

func init() {
	applyCmd.Flags().StringVar(&applyView, "view", "", "Saved view ID (required)")
}

func runApply(cmd *cobra.Command, args []string) error {
	if applyView == "" {
		return fmt.Errorf("saved view ID is required. Use --view flag")
	}
	client, err := clientFor("")
	if err != nil {
		return err
	}
	query := `query GetSavedViewConfig($id: String!) { savedView(id: $id) { viewConfig } }`
	var response struct {
		SavedView SavedView `json:"savedView"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"id": applyView}, &response); err != nil {
		return fmt.Errorf("failed to get saved view: %w", err)
	}
	return printJSON(response.SavedView.ViewConfig)
}
