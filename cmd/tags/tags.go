package tags

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for tag operations
var Cmd = &cobra.Command{
	Use:   "tags",
	Short: "Manage tags",
	Long:  "Create, list, and add tags to records within workspaces.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(addCmd)
}
