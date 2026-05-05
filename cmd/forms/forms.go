package forms

import (
	"github.com/heyblueteam/cli/cmd/forms/fields"

	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "forms",
	Aliases: []string{"form"},
	Short:   "Manage forms",
	Long:    "Create, list, update, copy, and delete forms, and manage their fields.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(copyCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(urlCmd)
	Cmd.AddCommand(fields.Cmd)
}
