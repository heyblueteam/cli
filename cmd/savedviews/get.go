package savedviews

import (
	"fmt"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a saved view",
	Example: `  blue saved-views get --view <id>
  blue saved-views get --view <id> --format json`,
	RunE: runGet,
}

var (
	getView   string
	getFormat string
)

func init() {
	getCmd.Flags().StringVar(&getView, "view", "", "Saved view ID (required)")
	getCmd.Flags().StringVar(&getFormat, "format", "", "Output format (json)")
}

func runGet(cmd *cobra.Command, args []string) error {
	if getView == "" {
		return fmt.Errorf("saved view ID is required. Use --view flag")
	}
	client, err := clientFor("")
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`query GetSavedView($id: String!) { savedView(id: $id) { %s } }`, savedViewFields)
	var response struct {
		SavedView SavedView `json:"savedView"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"id": getView}, &response); err != nil {
		return fmt.Errorf("failed to get saved view: %w", err)
	}
	if getFormat == "json" {
		return printJSON(response.SavedView)
	}
	printView(response.SavedView)
	return nil
}
