package tables

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for generic table operations (the record
// engine's tables, fields and records). These commands use the REST /v1 API
// and work on any table in a workspace, including tables with no list
// binding.
var Cmd = &cobra.Command{
	Use:     "tables",
	Aliases: []string{"table"},
	Short:   "Manage tables, fields, and records via the record engine",
	Long: `Manage generic tables, their fields, and their records via the REST API.

Unlike 'blue records' (legacy lists), these commands work on any table in a
workspace, including tables that have no list binding. Table and field
identities come from 'blue tables list' and 'blue tables fields list'.`,
}

func init() {
	Cmd.AddCommand(newListCmd())
	Cmd.AddCommand(newViewsCmd())
	Cmd.AddCommand(newFieldsCmd())
	Cmd.AddCommand(newRecordsCmd())
}
