package api

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

var queryCmd = &cobra.Command{
	Use:   "query",
	Short: "Run a raw GraphQL query",
	Example: `  blue api query --raw 'query { company(id: "acme") { id name } }'
  blue api query --file query.graphql --variables '{"id":"workspace_123"}'
  blue gql query --raw 'query { __typename }'`,
	RunE: runQuery,
}

var (
	queryRaw       string
	queryFile      string
	queryVariables string
	queryWorkspace string
)

func init() {
	queryCmd.Flags().StringVar(&queryRaw, "raw", "", "Raw GraphQL query or mutation")
	queryCmd.Flags().StringVar(&queryFile, "file", "", "Path to .graphql file")
	queryCmd.Flags().StringVar(&queryVariables, "variables", "", "Variables JSON object")
	queryCmd.Flags().StringVarP(&queryWorkspace, "workspace", "w", "", "Optional workspace context")
}

func runQuery(cmd *cobra.Command, args []string) error {
	if queryRaw == "" && queryFile == "" {
		return fmt.Errorf("query is required. Use --raw or --file")
	}
	if queryRaw != "" && queryFile != "" {
		return fmt.Errorf("use only one of --raw or --file")
	}

	query := queryRaw
	if queryFile != "" {
		data, err := os.ReadFile(queryFile)
		if err != nil {
			return fmt.Errorf("failed to read query file: %w", err)
		}
		query = string(data)
	}

	variables := map[string]interface{}{}
	if queryVariables != "" {
		if err := json.Unmarshal([]byte(queryVariables), &variables); err != nil {
			return fmt.Errorf("invalid --variables JSON: %w", err)
		}
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)
	if queryWorkspace != "" {
		client.SetProject(queryWorkspace)
	}

	data, err := client.ExecuteQuery(query, variables)
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
