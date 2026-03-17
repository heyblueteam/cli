package charts

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for chart operations
var Cmd = &cobra.Command{
	Use:   "charts",
	Short: "Manage charts within dashboards",
	Long:  "Create, list, delete, and export charts. Charts display data from workspaces as stats, bar charts, or pie charts.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(recalcCmd)
}
