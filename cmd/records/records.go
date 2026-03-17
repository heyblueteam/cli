package records

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for record operations
var Cmd = &cobra.Command{
	Use:     "records",
	Aliases: []string{"rec"},
	Short:   "Manage records",
	Long:    "Create, list, get, update, delete, move, and count records within workspaces.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(moveCmd)
	Cmd.AddCommand(countCmd)
}
