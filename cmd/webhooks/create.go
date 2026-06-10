package webhooks

import (
	"fmt"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a webhook",
	Example: `  blue webhooks create --url https://example.com/webhooks/blue
  blue webhooks create --name "Production sync" --url https://example.com/webhooks/blue --events TODO_CREATED,COMMENT_CREATED --workspaces ws1,ws2`,
	RunE: runCreate,
}

var (
	createName       string
	createURL        string
	createEvents     string
	createWorkspaces string
	createFormat     string
)

func init() {
	createCmd.Flags().StringVar(&createName, "name", "", "Webhook name")
	createCmd.Flags().StringVar(&createURL, "url", "", "Webhook endpoint URL (required)")
	createCmd.Flags().StringVar(&createEvents, "events", "", "Comma-separated event names (empty means all events)")
	createCmd.Flags().StringVar(&createWorkspaces, "workspaces", "", "Comma-separated workspace IDs (empty means all workspaces)")
	createCmd.Flags().StringVar(&createFormat, "format", "", "Output format (json)")
}

func runCreate(cmd *cobra.Command, args []string) error {
	if createURL == "" {
		return fmt.Errorf("webhook URL is required. Use --url flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)

	input := map[string]interface{}{"url": createURL}
	if createName != "" {
		input["name"] = createName
	}
	if createEvents != "" {
		input["events"] = splitCSV(createEvents)
	}
	if createWorkspaces != "" {
		input["projectIds"] = splitCSV(createWorkspaces)
	}

	query := fmt.Sprintf(`
		mutation CreateWebhook($input: CreateWebhookInput!) {
			createWebhook(input: $input) { %s }
		}
	`, webhookFields)

	var response struct {
		CreateWebhook Webhook `json:"createWebhook"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"input": input}, &response); err != nil {
		return fmt.Errorf("failed to create webhook: %w", err)
	}

	if createFormat == "json" {
		return printJSON(response.CreateWebhook)
	}
	fmt.Println("Webhook created. Store the secret now; it is only returned once.")
	printWebhook(response.CreateWebhook, true)
	return nil
}
