package tables

import (
	"encoding/csv"
	"fmt"

	"github.com/spf13/cobra"
)

type viewDatabase struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type viewEntry struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Database viewDatabase `json:"database"`
	ViewType *string      `json:"viewType"`
}

type viewRow struct {
	ID   string    `json:"id"`
	Rank string    `json:"rank"`
	View viewEntry `json:"view"`
}

type viewsOptions struct {
	workspace string
	format    string
}

func newViewsCmd() *cobra.Command {
	opts := viewsOptions{}
	cmd := &cobra.Command{
		Use:   "views",
		Short: "List the workspace's tab rail views",
		Long:  "List every view on the workspace's tab rail, grouped by its table.",
		Example: `  blue tables views -w <workspace>
  blue tables views -w <workspace> --format csv`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runViews(cmd, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.workspace, "workspace", "w", "", "Workspace ID or slug (required)")
	cmd.Flags().StringVar(&opts.format, "format", "text", "Output format: text, json, csv")

	return cmd
}

func runViews(cmd *cobra.Command, opts viewsOptions) error {
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

	var rows []viewRow
	body := map[string]string{"workspaceId": workspaceID}
	if err := client.REST("POST", "/list-database-views", nil, body, &rows); err != nil {
		return fmt.Errorf("failed to list views: %w", err)
	}

	switch opts.format {
	case "json":
		return printJSON(rows)
	case "csv":
		return writeViewsCSV(cmd, rows)
	default:
		return printViewsText(cmd, rows)
	}
}

func viewKind(view viewEntry) string {
	if view.ViewType == nil {
		return "-"
	}
	return *view.ViewType
}

func viewKindCSV(view viewEntry) string {
	if view.ViewType == nil {
		return ""
	}
	return *view.ViewType
}

func printViewsText(cmd *cobra.Command, rows []viewRow) error {
	out := cmd.OutOrStdout()
	if len(rows) == 0 {
		fmt.Fprintln(out, "No views found.")
		return nil
	}

	fmt.Fprintln(out, "=== Views ===")
	for i, row := range rows {
		if i > 0 {
			fmt.Fprintln(out)
		}
		fmt.Fprintf(out, "%d. %s\n", i+1, row.View.Name)
		fmt.Fprintf(out, "   ID: %s\n", row.View.ID)
		fmt.Fprintf(out, "   Table: %s (%s)\n", row.View.Database.Name, row.View.Database.ID)
		fmt.Fprintf(out, "   Kind: %s\n", viewKind(row.View))
	}
	return nil
}

func writeViewsCSV(cmd *cobra.Command, rows []viewRow) error {
	writer := csv.NewWriter(cmd.OutOrStdout())
	if err := writer.Write([]string{"id", "name", "databaseId", "databaseName", "viewType"}); err != nil {
		return err
	}
	for _, row := range rows {
		record := []string{
			row.View.ID,
			row.View.Name,
			row.View.Database.ID,
			row.View.Database.Name,
			viewKindCSV(row.View),
		}
		if err := writer.Write(record); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}
