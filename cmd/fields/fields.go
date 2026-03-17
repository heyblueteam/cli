package fields

import (
	"blue-cli/cmd/fields/groups"
	"blue-cli/cmd/fields/options"

	"github.com/spf13/cobra"
)

// Cmd is the parent command for custom field operations
var Cmd = &cobra.Command{
	Use:     "fields",
	Aliases: []string{"cf"},
	Short:   "Manage custom fields",
	Long:    "Create, list, update, and delete custom fields in your Blue workspaces.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(options.Cmd)
	Cmd.AddCommand(groups.Cmd)
}
