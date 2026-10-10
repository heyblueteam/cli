package databases

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

type recordCreateOptions struct {
	database   string
	title      string
	group      string
	fieldValue []string
	workspace  string
	format     string
}

type recordCreateResponse struct {
	RecordID   string  `json:"recordId"`
	DatabaseID string  `json:"databaseId"`
	Created    bool    `json:"created"`
	Joined     bool    `json:"joined"`
	Rank       *string `json:"rank"`
}

func newRecordsCreateCmd() *cobra.Command {
	opts := recordCreateOptions{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a record on a database",
		Long: `Create a record on any database.

Works on any database in a workspace. --group places the record in a group
option on the database's grouping field. A database with a required creation
form rejects missing fields server-side.

Repeat --field-value once per field. Each value is one JSON object:
{"fieldId":"f_1","kind":"text","value":{"text":"hi"}}`,
		Example: `  blue databases records create --database <id> --workspace <id> --title "New deal"
  blue databases records create --database <id> --workspace <id> --title "New deal" --group <option-id>
  blue databases records create --database <id> --workspace <id> --title "New deal" \
    --field-value '{"fieldId":"f_1","kind":"text","value":{"text":"hi"}}'`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRecordsCreate(cmd, opts)
		},
	}

	cmd.Flags().StringVar(&opts.database, "database", "", "Database ID (required)")
	cmd.Flags().StringVarP(&opts.title, "title", "t", "", "Record title (required)")
	cmd.Flags().StringVar(&opts.group, "group", "", "Group option ID on the database's grouping field")
	cmd.Flags().StringArrayVar(&opts.fieldValue, "field-value", nil, "Field value as one JSON object; repeat the flag per field")
	cmd.Flags().StringVarP(&opts.workspace, "workspace", "w", "", "Workspace ID or slug (required)")
	cmd.Flags().StringVar(&opts.format, "format", "text", "Output format: text or json")

	return cmd
}

func runRecordsCreate(cmd *cobra.Command, opts recordCreateOptions) error {
	if err := requireFlag(opts.database, "database"); err != nil {
		return err
	}
	if err := requireFlag(opts.title, "title"); err != nil {
		return err
	}
	if err := requireFlag(opts.workspace, "workspace"); err != nil {
		return err
	}
	switch opts.format {
	case "text", "json":
	default:
		return fmt.Errorf("invalid format %q. Use text or json", opts.format)
	}

	fieldValues := make([]map[string]interface{}, 0, len(opts.fieldValue))
	for i, raw := range opts.fieldValue {
		parsed, err := recordParseFieldValue(raw)
		if err != nil {
			return fmt.Errorf("invalid --field-value JSON (--field-value[%d]): %w", i, err)
		}
		fieldValues = append(fieldValues, parsed)
	}

	client, workspaceID, err := databasesClient(opts.workspace)
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"workspaceId": workspaceID,
		"title":       opts.title,
	}
	if opts.group != "" {
		body["groupOptionId"] = opts.group
	}
	if len(fieldValues) > 0 {
		body["fieldValues"] = fieldValues
	}

	var response recordCreateResponse
	if err := client.REST("POST", "/databases/"+opts.database+"/records", nil, body, &response); err != nil {
		return fmt.Errorf("failed to create database record: %w", err)
	}

	if opts.format == "json" {
		return printJSON(response)
	}

	if response.Joined {
		common.PrintSuccess("Existing record joined to database")
	} else {
		common.PrintSuccess("Record created")
	}

	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "  Record ID: %s\n", response.RecordID)
	if response.Rank != nil {
		fmt.Fprintf(out, "  Rank: %v\n", *response.Rank)
	}
	return nil
}

func recordParseFieldValue(raw string) (map[string]interface{}, error) {
	dec := json.NewDecoder(strings.NewReader(raw))
	dec.UseNumber()
	var parsed map[string]interface{}
	if err := dec.Decode(&parsed); err != nil {
		return nil, err
	}
	for _, key := range []string{"fieldId", "kind", "value"} {
		if _, ok := parsed[key]; !ok {
			return nil, fmt.Errorf("missing the %q key", key)
		}
	}
	return parsed, nil
}
