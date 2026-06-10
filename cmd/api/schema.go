package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

var schemaCmd = &cobra.Command{
	Use:   "schema",
	Short: "Print the GraphQL schema",
	Example: `  blue api schema
  blue api schema --introspect`,
	RunE: runSchema,
}

var schemaIntrospect bool

func init() {
	schemaCmd.Flags().BoolVar(&schemaIntrospect, "introspect", false, "Fetch schema via GraphQL introspection instead of printing local schema.graphql")
}

func runSchema(cmd *cobra.Command, args []string) error {
	if !schemaIntrospect {
		for _, path := range []string{"schema.graphql", filepath.Join("..", "api", "src", "schema.graphql")} {
			data, err := os.ReadFile(path)
			if err == nil {
				fmt.Print(string(data))
				return nil
			}
		}
		return fmt.Errorf("local schema.graphql not found; use --introspect to fetch from the API")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)
	query := `query IntrospectionQuery { __schema { queryType { name } mutationType { name } subscriptionType { name } types { kind name description fields { name description args { name description type { kind name ofType { kind name ofType { kind name } } } defaultValue } type { kind name ofType { kind name ofType { kind name } } } } inputFields { name description type { kind name ofType { kind name ofType { kind name } } } defaultValue } enumValues { name description isDeprecated deprecationReason } } directives { name description locations args { name description type { kind name ofType { kind name } } defaultValue } } } }`
	data, err := client.ExecuteQuery(query, nil)
	if err != nil {
		return err
	}
	out, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}
