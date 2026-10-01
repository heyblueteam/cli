package databases

import (
	"github.com/spf13/cobra"
)

// newRecordsCmd is the parent for record operations on the record engine's
// databases via the REST /v1 API.
func newRecordsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "records",
		Short: "Manage records on a database",
		Long: `Read and write records on any database via the record engine.

Works on any database in a workspace. Database and field IDs come from
'blue databases list' and 'blue databases fields list'.`,
	}

	cmd.AddCommand(newRecordsListCmd())
	cmd.AddCommand(newRecordsCreateCmd())
	cmd.AddCommand(newRecordsSetValueCmd())
	cmd.AddCommand(newRecordsLinkPRCmd())

	return cmd
}
