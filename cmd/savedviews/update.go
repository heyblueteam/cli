package savedviews

import (
	"fmt"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a saved view",
	Example: `  blue saved-views update --view <id> --name "Sprint Board"
  blue saved-views update --view <id> --shared true --config-json '{"searchQuery":"launch"}'`,
	RunE: runUpdate,
}

var (
	updateView       string
	updateName       string
	updateIcon       string
	updateShared     string
	updateConfigJSON string
)

func init() {
	updateCmd.Flags().StringVar(&updateView, "view", "", "Saved view ID (required)")
	updateCmd.Flags().StringVar(&updateName, "name", "", "New view name")
	updateCmd.Flags().StringVar(&updateIcon, "icon", "", "New icon")
	updateCmd.Flags().StringVar(&updateShared, "shared", "", "Set shared state (true or false)")
	updateCmd.Flags().StringVar(&updateConfigJSON, "config-json", "", "Raw replacement viewConfig JSON")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if updateView == "" {
		return fmt.Errorf("saved view ID is required. Use --view flag")
	}
	input := map[string]interface{}{"id": updateView}
	if updateName != "" {
		input["name"] = updateName
	}
	if updateIcon != "" {
		input["icon"] = updateIcon
	}
	if updateShared != "" {
		switch updateShared {
		case "true":
			input["isShared"] = true
		case "false":
			input["isShared"] = false
		default:
			return fmt.Errorf("--shared must be true or false")
		}
	}
	if updateConfigJSON != "" {
		config, err := parseJSONFlag("--config-json", updateConfigJSON)
		if err != nil {
			return err
		}
		input["viewConfig"] = config
	}
	if len(input) == 1 {
		return fmt.Errorf("nothing to update. Pass at least one field flag")
	}
	client, err := clientFor("")
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`mutation EditSavedView($input: EditSavedViewInput!) { editSavedView(input: $input) { %s } }`, savedViewFields)
	var response struct {
		EditSavedView SavedView `json:"editSavedView"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"input": input}, &response); err != nil {
		return fmt.Errorf("failed to update saved view: %w", err)
	}
	printView(response.EditSavedView)
	return nil
}
