package checklists

import (
	"blue-cli/cmd/checklists/items"

	"github.com/spf13/cobra"
)

// Cmd is the parent command for checklist operations
var Cmd = &cobra.Command{
	Use:   "checklists",
	Short: "Manage checklists",
	Long:  "Create, list, and delete checklists on records, and manage checklist items.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(items.Cmd)
}
