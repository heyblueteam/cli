package lists

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for list operations
var Cmd = &cobra.Command{
	Use:   "lists",
	Short: "Manage lists",
	Long:  "Create, list, update, and delete lists within workspaces.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(deleteCmd)
}
