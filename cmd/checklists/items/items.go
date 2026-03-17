package items

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for checklist item operations
var Cmd = &cobra.Command{
	Use:   "items",
	Short: "Manage checklist items",
	Long:  "Create, update, and delete checklist items.",
}

func init() {
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(deleteCmd)
}
