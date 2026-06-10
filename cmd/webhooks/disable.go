package webhooks

import (
	"fmt"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

var disableCmd = &cobra.Command{
	Use:     "disable",
	Short:   "Disable a webhook",
	Example: `  blue webhooks disable --webhook <id>`,
	RunE:    runDisable,
}

var disableWebhook string

func init() {
	disableCmd.Flags().StringVar(&disableWebhook, "webhook", "", "Webhook ID (required)")
}

func runDisable(cmd *cobra.Command, args []string) error {
	if disableWebhook == "" {
		return fmt.Errorf("webhook ID is required. Use --webhook flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)

	query := `
		mutation DisableWebhook($input: DisableWebhookInput!) {
			disableWebhook(input: $input) { success operationId }
		}
	`
	var response struct {
		DisableWebhook mutationResult `json:"disableWebhook"`
	}
	variables := map[string]interface{}{"input": map[string]interface{}{"webhookId": disableWebhook}}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to disable webhook: %w", err)
	}
	if !response.DisableWebhook.Success {
		return fmt.Errorf("webhook was not disabled")
	}
	common.PrintSuccess(fmt.Sprintf("Disabled webhook %s", disableWebhook))
	return nil
}
