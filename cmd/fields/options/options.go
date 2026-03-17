package options

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for custom field option operations
var Cmd = &cobra.Command{
	Use:   "options",
	Short: "Manage custom field options",
	Long:  "Create and delete options for select-type custom fields.",
}

func init() {
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(deleteCmd)
}
