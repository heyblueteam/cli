package tables

import (
	"github.com/spf13/cobra"
)

// newRecordsCmd is the parent for record operations on the record engine's
// generic tables. Unlike 'blue records' (legacy lists), these commands work
// on any table in a workspace via the REST /v1 API.
func newRecordsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "records",
		Short: "Manage records on a table",
		Long: `Read and write records on any table via the record engine.

Works on any table in a workspace. Table and field IDs come from
'blue tables list' and 'blue tables fields list'.`,
	}

	cmd.AddCommand(newRecordsListCmd())
	cmd.AddCommand(newRecordsCreateCmd())
	cmd.AddCommand(newRecordsSetValueCmd())

	return cmd
}
