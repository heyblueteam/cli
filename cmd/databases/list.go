package databases

import (
	"encoding/csv"
	"fmt"

	"github.com/spf13/cobra"
)

type databaseEntry struct {
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
		Short: "List databases in a workspace",
		Long:  "List every database in a workspace.",
		Example: `  blue databases list -w <workspace>
  blue databases list -w <workspace> --format json`,
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

	client, workspaceID, err := databasesClient(opts.workspace)
	if err != nil {
		return err
	}

	var databases []databaseEntry
	if err := client.REST("GET", "/databases", queryValues("workspaceId", workspaceID), nil, &databases); err != nil {
		return fmt.Errorf("failed to list databases: %w", err)
	}

	switch opts.format {
	case "json":
		return printJSON(databases)
	case "csv":
		return writeDatabasesCSV(cmd, databases)
	default:
		return printDatabasesText(cmd, databases)
	}
}

func printDatabasesText(cmd *cobra.Command, databases []databaseEntry) error {
	out := cmd.OutOrStdout()
	if len(databases) == 0 {
		fmt.Fprintln(out, "No databases found.")
		return nil
	}

	fmt.Fprintln(out, "=== Databases ===")
	for i, database := range databases {
		if i > 0 {
			fmt.Fprintln(out)
		}
		fmt.Fprintf(out, "%d. %s\n", i+1, database.Name)
		fmt.Fprintf(out, "   ID: %s\n", database.ID)
	}
	return nil
}

func writeDatabasesCSV(cmd *cobra.Command, databases []databaseEntry) error {
	writer := csv.NewWriter(cmd.OutOrStdout())
	if err := writer.Write([]string{"id", "name"}); err != nil {
		return err
	}
	for _, database := range databases {
		if err := writer.Write([]string{database.ID, database.Name}); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}
