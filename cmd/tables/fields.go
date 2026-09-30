package tables

import (
	"github.com/spf13/cobra"
)

// newFieldsCmd builds the fields command group. Fields live on the record
// engine's tables and are shared across tables in a workspace: create a
// field once, then mount it on other tables.
func newFieldsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fields",
		Short: "Manage fields on a table",
		Long: `Manage fields on the record engine's tables via the REST API.

Fields are shared across tables in a workspace: create a field on one table,
then mount it on other tables with 'fields mount' and remove it again with
'fields unmount'. Field identities come from 'blue tables fields list'.`,
	}

	cmd.AddCommand(newFieldsListCmd())
	cmd.AddCommand(newFieldsCreateCmd())
	cmd.AddCommand(newFieldsRenameCmd())
	cmd.AddCommand(newFieldsDeleteCmd())
	cmd.AddCommand(newFieldsMountCmd())
	cmd.AddCommand(newFieldsUnmountCmd())

	return cmd
}
