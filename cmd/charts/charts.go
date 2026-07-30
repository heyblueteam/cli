package charts

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for chart operations
var Cmd = &cobra.Command{
	Use:   "charts",
	Short: "Manage charts within dashboards",
	Long:  "Create, preview, update, copy, list, delete, and recalculate charts. Charts display data from workspaces as stat cards, bar charts, or pie charts.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(previewCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(copyCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(recalcCmd)
}
