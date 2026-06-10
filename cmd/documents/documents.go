package documents

import "github.com/spf13/cobra"

// Cmd is the parent command for document operations.
var Cmd = &cobra.Command{
	Use:   "documents",
	Short: "Manage documents and wiki pages",
	Long:  "Create, list, update, and delete rich-text Documents and Wiki pages.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(deleteCmd)
}
