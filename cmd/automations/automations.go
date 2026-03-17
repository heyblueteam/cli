package automations

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for automation operations
var Cmd = &cobra.Command{
	Use:     "automations",
	Aliases: []string{"auto"},
	Short:   "Manage automations",
	Long:    "Create, list, update, and delete automations within workspaces.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(deleteCmd)
}
