package webhooks

import (
	"fmt"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a webhook",
	Example: `  blue webhooks update --webhook <id> --name "Production sync"
  blue webhooks update --webhook <id> --url https://example.com/webhooks/blue --enabled true
  blue webhooks update --webhook <id> --events TODO_CREATED,COMMENT_CREATED --workspaces ws1,ws2`,
	RunE: runUpdate,
}

var (
	updateWebhook    string
	updateName       string
	updateURL        string
	updateEvents     string
	updateWorkspaces string
	updateEnabled    string
	updateFormat     string
)

func init() {
	updateCmd.Flags().StringVar(&updateWebhook, "webhook", "", "Webhook ID (required)")
	updateCmd.Flags().StringVar(&updateName, "name", "", "New webhook name")
	updateCmd.Flags().StringVar(&updateURL, "url", "", "New webhook endpoint URL")
	updateCmd.Flags().StringVar(&updateEvents, "events", "", "Comma-separated event names; pass empty string to leave unchanged")
	updateCmd.Flags().StringVar(&updateWorkspaces, "workspaces", "", "Comma-separated workspace IDs; pass empty string to leave unchanged")
	updateCmd.Flags().StringVar(&updateEnabled, "enabled", "", "Set enabled state (true or false)")
	updateCmd.Flags().StringVar(&updateFormat, "format", "", "Output format (json)")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if updateWebhook == "" {
		return fmt.Errorf("webhook ID is required. Use --webhook flag")
	}

	input := map[string]interface{}{"webhookId": updateWebhook}
	if updateName != "" {
		input["name"] = updateName
	}
	if updateURL != "" {
		input["url"] = updateURL
	}
	if updateEvents != "" {
		input["events"] = splitCSV(updateEvents)
	}
	if updateWorkspaces != "" {
		input["projectIds"] = splitCSV(updateWorkspaces)
	}
	if updateEnabled != "" {
		switch updateEnabled {
		case "true":
			input["enabled"] = true
		case "false":
			input["enabled"] = false
		default:
			return fmt.Errorf("--enabled must be true or false")
		}
	}
	if len(input) == 1 {
		return fmt.Errorf("nothing to update. Pass at least one field flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)

	query := fmt.Sprintf(`
		mutation UpdateWebhook($input: UpdateWebhookInput!) {
			updateWebhook(input: $input) { %s }
		}
	`, webhookFields)

	var response struct {
		UpdateWebhook Webhook `json:"updateWebhook"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"input": input}, &response); err != nil {
		return fmt.Errorf("failed to update webhook: %w", err)
	}

	if updateFormat == "json" {
		return printJSON(response.UpdateWebhook)
	}
	printWebhook(response.UpdateWebhook, false)
	return nil
}
