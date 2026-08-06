package dashboards

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for dashboard operations
var Cmd = &cobra.Command{
	Use:     "dashboards",
	Aliases: []string{"dash"},
	Short:   "Manage dashboards",
	Long:    "Create, list, view, update, delete, and share dashboards.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(shareCmd)
}
