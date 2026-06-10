package webhooks

import "github.com/spf13/cobra"

// Cmd is the parent command for webhook operations.
var Cmd = &cobra.Command{
	Use:     "webhooks",
	Aliases: []string{"wh"},
	Short:   "Manage webhooks",
	Long:    "Create, list, update, disable, delete, inspect, and test Blue webhooks.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(disableCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(eventsCmd)
	Cmd.AddCommand(verifySignatureCmd)
	Cmd.AddCommand(listenCmd)
}
