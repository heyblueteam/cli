package webhooks

import (
	"fmt"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a webhook",
	Example: `  blue webhooks get --webhook <id>
  blue webhooks get --webhook <id> --format json`,
	RunE: runGet,
}

var (
	getWebhook string
	getFormat  string
)

func init() {
	getCmd.Flags().StringVar(&getWebhook, "webhook", "", "Webhook ID (required)")
	getCmd.Flags().StringVar(&getFormat, "format", "", "Output format (json)")
}

func runGet(cmd *cobra.Command, args []string) error {
	if getWebhook == "" {
		return fmt.Errorf("webhook ID is required. Use --webhook flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)

	query := fmt.Sprintf(`
		query GetWebhook($id: String!) {
			webhook(id: $id) { %s }
		}
	`, webhookFields)

	var response struct {
		Webhook Webhook `json:"webhook"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"id": getWebhook}, &response); err != nil {
		return fmt.Errorf("failed to get webhook: %w", err)
	}

	if getFormat == "json" {
		return printJSON(response.Webhook)
	}
	printWebhook(response.Webhook, false)
	return nil
}
