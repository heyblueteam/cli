package tables

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
)

// tableField is one entry of GET /v1/databases/:tableId/fields. The array is
// in display order; configuration is arbitrary field-type specific JSON.
type tableField struct {
	ID            string          `json:"id"`
	Name          string          `json:"name"`
	Type          string          `json:"type"`
	Configuration json.RawMessage `json:"configuration,omitempty"`
	Rank          string          `json:"rank"`
}

type fieldsListOptions struct {
	workspace string
	table     string
	format    string
}

func newFieldsListCmd() *cobra.Command {
	opts := fieldsListOptions{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List the fields on a table",
		Long: `List the fields on a table in display order.

Use the field IDs from this listing for rename, delete, mount and unmount.`,
		Example: `  blue tables fields list --workspace <id> --table <id>
  blue tables fields list -w <id> --table <id> --format json
  blue tables fields list -w <id> --table <id> --format csv`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFieldsList(cmd, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.workspace, "workspace", "w", "", "Workspace ID or slug (required)")
	cmd.Flags().StringVar(&opts.table, "table", "", "Table ID (required)")
	cmd.Flags().StringVar(&opts.format, "format", "text", "Output format: text, json, csv")

	return cmd
}

func runFieldsList(cmd *cobra.Command, opts fieldsListOptions) error {
	opts.workspace = strings.TrimSpace(opts.workspace)
	opts.table = strings.TrimSpace(opts.table)

	if err := requireFlag(opts.workspace, "workspace"); err != nil {
		return err
	}
	if err := requireFlag(opts.table, "table"); err != nil {
		return err
	}

	client, workspaceID, err := tablesClient(opts.workspace)
	if err != nil {
		return err
	}

	var fields []tableField
	path := fmt.Sprintf("/databases/%s/fields", opts.table)
	if err := client.REST(http.MethodGet, path, queryValues("workspaceId", workspaceID), nil, &fields); err != nil {
		return fmt.Errorf("failed to list fields: %w", err)
	}

	switch opts.format {
	case "json":
		return printJSON(fields)
	case "csv":
		return writeFieldsCSV(cmd.OutOrStdout(), fields)
	default:
		return printFieldsTable(cmd.OutOrStdout(), fields)
	}
}

// printFieldsTable renders the fields as a fixed-width table, mirroring the
// display style of the workspace custom fields listing.
func printFieldsTable(out io.Writer, fields []tableField) error {
	if len(fields) == 0 {
		fmt.Fprintln(out, "No fields found.")
		return nil
	}

	fmt.Fprintf(out, "%-6s | %-32s | %-25s | %-15s\n", "RANK", "ID", "NAME", "TYPE")
	fmt.Fprintf(out, "%s-+-%s-+-%s-+-%s\n",
		strings.Repeat("-", 6), strings.Repeat("-", 32), strings.Repeat("-", 25), strings.Repeat("-", 15))

	for _, field := range fields {
		fmt.Fprintf(out, "%-6s | %-32s | %-25s | %-15s\n", field.Rank, field.ID, field.Name, field.Type)
	}
	return nil
}

// writeFieldsCSV renders the fields as CSV. The configuration column holds
// the compact JSON string, empty when the field has no configuration.
func writeFieldsCSV(out io.Writer, fields []tableField) error {
	csvWriter := csv.NewWriter(out)
	if err := csvWriter.Write([]string{"id", "name", "type", "rank", "configuration"}); err != nil {
		return err
	}

	for _, field := range fields {
		if err := csvWriter.Write([]string{
			field.ID,
			field.Name,
			field.Type,
			field.Rank,
			fieldsConfigurationString(field.Configuration),
		}); err != nil {
			return err
		}
	}

	csvWriter.Flush()
	return csvWriter.Error()
}

func fieldsConfigurationString(raw json.RawMessage) string {
	if len(raw) == 0 || string(raw) == "null" {
		return ""
	}
	var buf bytes.Buffer
	if err := json.Compact(&buf, raw); err != nil {
		return string(raw)
	}
	return buf.String()
}
