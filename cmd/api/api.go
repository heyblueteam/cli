package api

import "github.com/spf13/cobra"

// Cmd is the parent command for raw Blue API access.
var Cmd = &cobra.Command{
	Use:     "api",
	Aliases: []string{"graphql", "gql"},
	Short:   "Run raw Blue API requests",
	Long:    "Run raw GraphQL queries, inspect the schema, and open Blue API documentation.",
}

func init() {
	Cmd.AddCommand(queryCmd)
	Cmd.AddCommand(schemaCmd)
	Cmd.AddCommand(docsCmd)
}
