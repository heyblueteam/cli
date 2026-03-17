package workspaces

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for workspace operations
var Cmd = &cobra.Command{
	Use:     "workspaces",
	Aliases: []string{"ws"},
	Short:   "Manage workspaces",
	Long:    "Create, list, update, and delete workspaces in your Blue company.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(deleteCmd)
}
