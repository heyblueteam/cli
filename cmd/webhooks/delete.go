package webhooks

import (
	"fmt"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete a webhook",
	Example: `  blue webhooks delete --webhook <id> --confirm`,
	RunE:    runDelete,
}

var (
	deleteWebhook string
	deleteConfirm bool
)

func init() {
	deleteCmd.Flags().StringVar(&deleteWebhook, "webhook", "", "Webhook ID (required)")
	deleteCmd.Flags().BoolVarP(&deleteConfirm, "confirm", "y", false, "Confirm deletion (required)")
}

func runDelete(cmd *cobra.Command, args []string) error {
	if deleteWebhook == "" {
		return fmt.Errorf("webhook ID is required. Use --webhook flag")
	}
	if !deleteConfirm {
		return fmt.Errorf("deletion confirmation is required. Use --confirm flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)

	query := `
		mutation DeleteWebhook($input: DeleteWebhookInput!) {
			deleteWebhook(input: $input) { success operationId }
		}
	`
	var response struct {
		DeleteWebhook mutationResult `json:"deleteWebhook"`
	}
	variables := map[string]interface{}{"input": map[string]interface{}{"webhookId": deleteWebhook}}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to delete webhook: %w", err)
	}
	if !response.DeleteWebhook.Success {
		return fmt.Errorf("webhook was not deleted")
	}
	common.PrintSuccess(fmt.Sprintf("Deleted webhook %s", deleteWebhook))
	return nil
}
