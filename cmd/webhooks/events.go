package webhooks

import (
	"fmt"

	"github.com/spf13/cobra"
)

var eventsCmd = &cobra.Command{
	Use:   "events",
	Short: "List webhook events",
	Example: `  blue webhooks events
  blue webhooks events --format json`,
	RunE: runEvents,
}

var eventsFormat string

func init() {
	eventsCmd.Flags().StringVar(&eventsFormat, "format", "", "Output format (json)")
}

func runEvents(cmd *cobra.Command, args []string) error {
	if eventsFormat == "json" {
		return printJSON(webhookEvents)
	}
	for _, event := range webhookEvents {
		fmt.Printf("%s\n  %s\n", event.Name, event.Description)
	}
	return nil
}
