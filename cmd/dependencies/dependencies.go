package dependencies

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for dependency operations
var Cmd = &cobra.Command{
	Use:     "dependencies",
	Aliases: []string{"deps"},
	Short:   "Manage record dependencies",
	Long:    "Create, update, and delete dependencies between records.",
}

func init() {
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(deleteCmd)
}
