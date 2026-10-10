package databases

import (
	"github.com/spf13/cobra"
)

// newFieldsCmd builds the fields command group. Fields live on the record
// engine's databases and are shared across databases in a workspace: create
// a field once, then mount it on other databases.
func newFieldsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fields",
		Short: "Manage fields on a database",
		Long: `Manage the fields on a database via the REST API.

Fields are shared across databases in a workspace: create a field once, then
mount it on other databases with 'fields mount' and remove it again with
'fields unmount'. Field identities come from 'blue databases fields list'.`,
	}

	cmd.AddCommand(newFieldsListCmd())
	cmd.AddCommand(newFieldsCreateCmd())
	cmd.AddCommand(newFieldsRenameCmd())
	cmd.AddCommand(newFieldsDeleteCmd())
	cmd.AddCommand(newFieldsMountCmd())
	cmd.AddCommand(newFieldsUnmountCmd())

	return cmd
}
