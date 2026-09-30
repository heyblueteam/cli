package tables

import (
	"encoding/csv"
	"fmt"

	"github.com/spf13/cobra"
)

type tableEntry struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type listOptions struct {
	workspace string
	format    string
}

func newListCmd() *cobra.Command {
	opts := listOptions{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tables in a workspace",
		Long:  "List every table in a workspace, including tables with no list binding.",
		Example: `  blue tables list -w <workspace>
  blue tables list -w <workspace> --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runList(cmd, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.workspace, "workspace", "w", "", "Workspace ID or slug (required)")
	cmd.Flags().StringVar(&opts.format, "format", "text", "Output format: text, json, csv")

	return cmd
}

func runList(cmd *cobra.Command, opts listOptions) error {
	if err := requireFlag(opts.workspace, "workspace"); err != nil {
		return err
	}

	switch opts.format {
	case "text", "json", "csv":
	default:
		return fmt.Errorf("invalid format %q. Use text, json, or csv", opts.format)
	}

	client, workspaceID, err := tablesClient(opts.workspace)
	if err != nil {
		return err
	}

	var tables []tableEntry
	if err := client.REST("GET", "/databases", queryValues("workspaceId", workspaceID), nil, &tables); err != nil {
		return fmt.Errorf("failed to list tables: %w", err)
	}

	switch opts.format {
	case "json":
		return printJSON(tables)
	case "csv":
		return writeTablesCSV(cmd, tables)
	default:
		return printTablesText(cmd, tables)
	}
}

func printTablesText(cmd *cobra.Command, tables []tableEntry) error {
	out := cmd.OutOrStdout()
	if len(tables) == 0 {
		fmt.Fprintln(out, "No tables found.")
		return nil
	}

	fmt.Fprintln(out, "=== Tables ===")
	for i, table := range tables {
		if i > 0 {
			fmt.Fprintln(out)
		}
		fmt.Fprintf(out, "%d. %s\n", i+1, table.Name)
		fmt.Fprintf(out, "   ID: %s\n", table.ID)
	}
	return nil
}

func writeTablesCSV(cmd *cobra.Command, tables []tableEntry) error {
	writer := csv.NewWriter(cmd.OutOrStdout())
	if err := writer.Write([]string{"id", "name"}); err != nil {
		return err
	}
	for _, table := range tables {
		if err := writer.Write([]string{table.ID, table.Name}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}
