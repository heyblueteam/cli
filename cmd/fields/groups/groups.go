package groups

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for custom field group operations
var Cmd = &cobra.Command{
	Use:   "groups",
	Short: "Manage custom field groups",
	Long:  "List and manage custom field groups within workspaces.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(manageCmd)
}
