package charts

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for chart operations
var Cmd = &cobra.Command{
	Use:   "charts",
	Short: "Manage charts within dashboards",
	Long:  "Create, preview, inspect, edit, recalculate, and delete dashboard charts.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(previewCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(editCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(recalcCmd)
}
