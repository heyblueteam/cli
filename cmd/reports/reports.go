package reports

import "github.com/spf13/cobra"

// Cmd is the parent command for report operations.
var Cmd = &cobra.Command{
	Use:   "reports",
	Short: "Manage reports",
	Long:  "Create, list, update, share, read, duplicate, delete, and export Blue reports.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(shareCmd)
	Cmd.AddCommand(dataCmd)
	Cmd.AddCommand(aggregateCmd)
	Cmd.AddCommand(refreshCmd)
	Cmd.AddCommand(duplicateCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(exportCmd)
}
