package exports

import "github.com/spf13/cobra"

// Cmd is the parent command for export operations.
var Cmd = &cobra.Command{
	Use:   "exports",
	Short: "Queue CSV exports",
	Long:  "Queue record, report, chart, and CSV template exports through the Blue API.",
}

func init() {
	Cmd.AddCommand(recordsCmd)
	Cmd.AddCommand(reportCmd)
	Cmd.AddCommand(chartCmd)
	Cmd.AddCommand(templateCmd)
}
