package webhooks

import (
	"fmt"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List webhooks",
	Example: `  blue webhooks list
  blue webhooks list --enabled --simple
  blue webhooks list --page 2 --size 50 --format json`,
	RunE: runList,
}

var (
	listEnabled bool
	listSimple  bool
	listPage    int
	listSize    int
	listFormat  string
)

func init() {
	listCmd.Flags().BoolVar(&listEnabled, "enabled", false, "Only show enabled webhooks")
	listCmd.Flags().BoolVarP(&listSimple, "simple", "s", false, "Simple output format")
	listCmd.Flags().IntVar(&listPage, "page", 1, "Page number (1-indexed)")
	listCmd.Flags().IntVar(&listSize, "size", 20, "Page size")
	listCmd.Flags().StringVar(&listFormat, "format", "", "Output format (json)")
}

func runList(cmd *cobra.Command, args []string) error {
	if listPage < 1 {
		listPage = 1
	}
	if listSize < 1 {
		listSize = 20
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)

	query := fmt.Sprintf(`
		query ListWebhooks($filter: WebhookFilter, $skip: Int, $take: Int) {
			webhooks(filter: $filter, skip: $skip, take: $take) {
				items { %s }
				pageInfo { totalItems totalPages hasNextPage hasPreviousPage }
			}
		}
	`, webhookFields)

	variables := map[string]interface{}{
		"skip": (listPage - 1) * listSize,
		"take": listSize,
	}
	if listEnabled {
		variables["filter"] = map[string]interface{}{"enabled": true}
	}

	var response struct {
		Webhooks struct {
			Items    []Webhook `json:"items"`
			PageInfo struct {
				TotalItems  int  `json:"totalItems"`
				TotalPages  int  `json:"totalPages"`
				HasNextPage bool `json:"hasNextPage"`
			} `json:"pageInfo"`
		} `json:"webhooks"`
	}

	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to list webhooks: %w", err)
	}

	if listFormat == "json" {
		return printJSON(response.Webhooks.Items)
	}

	if listSimple {
		for _, w := range response.Webhooks.Items {
			state := "disabled"
			if w.Enabled {
				state = "enabled"
			}
			fmt.Printf("%s  %s  %s  %s\n", w.ID, state, w.Status, w.URL)
		}
		return nil
	}

	fmt.Printf("=== Webhooks (page %d, %d shown, %d total) ===\n\n", listPage, len(response.Webhooks.Items), response.Webhooks.PageInfo.TotalItems)
	for i, w := range response.Webhooks.Items {
		fmt.Printf("%d. ", i+1)
		printWebhook(w, false)
		fmt.Println()
	}
	if response.Webhooks.PageInfo.HasNextPage {
		fmt.Printf("More results available. Pass --page %d to fetch the next page.\n", listPage+1)
	}
	return nil
}
