package databases

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for database operations (the record engine's
// databases, fields and records). These commands use the REST /v1 API and
// work on any database in a workspace.
var Cmd = &cobra.Command{
	Use:     "databases",
	Aliases: []string{"database", "db"},
	Short:   "Manage databases, fields, and records via the record engine",
	Long: `Manage databases, their fields, and their records via the REST API.

Works on any database in a workspace. Database and field identities come from
'blue databases list' and 'blue databases fields list'.`,
}

func init() {
	Cmd.AddCommand(newListCmd())
	Cmd.AddCommand(newViewsCmd())
	Cmd.AddCommand(newFieldsCmd())
	Cmd.AddCommand(newRecordsCmd())
}
