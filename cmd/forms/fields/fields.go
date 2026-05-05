package fields

import "github.com/spf13/cobra"

var Cmd = &cobra.Command{
	Use:     "fields",
	Aliases: []string{"field"},
	Short:   "Manage form fields",
	Long:    "Add, list, update, and delete fields on a form.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(addCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(deleteCmd)
}
