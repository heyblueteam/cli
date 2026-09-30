package tables

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

type fieldsRenameOptions struct {
	workspace string
	field     string
	name      string
	format    string
}

func newFieldsRenameCmd() *cobra.Command {
	opts := fieldsRenameOptions{}
	cmd := &cobra.Command{
		Use:     "rename",
		Short:   "Rename a field",
		Example: `  blue tables fields rename --workspace <id> --field <id> --name "New name"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFieldsRename(cmd, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.workspace, "workspace", "w", "", "Workspace ID or slug (required)")
	cmd.Flags().StringVar(&opts.field, "field", "", "Field ID (required)")
	cmd.Flags().StringVar(&opts.name, "name", "", "New field name (required)")
	cmd.Flags().StringVar(&opts.format, "format", "text", "Output format: text, json")

	return cmd
}

func runFieldsRename(cmd *cobra.Command, opts fieldsRenameOptions) error {
	opts.workspace = strings.TrimSpace(opts.workspace)
	opts.field = strings.TrimSpace(opts.field)
	opts.name = strings.TrimSpace(opts.name)

	if err := requireFlag(opts.workspace, "workspace"); err != nil {
		return err
	}
	if err := requireFlag(opts.field, "field"); err != nil {
		return err
	}
	if err := requireFlag(opts.name, "name"); err != nil {
		return err
	}

	client, workspaceID, err := tablesClient(opts.workspace)
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"workspaceId": workspaceID,
		"fieldId":     opts.field,
		"name":        opts.name,
	}
	var response struct {
		FieldID string `json:"fieldId"`
		Renamed bool   `json:"renamed"`
	}
	if err := client.REST(http.MethodPost, "/rename-database-field", nil, body, &response); err != nil {
		return fmt.Errorf("failed to rename field: %w", err)
	}

	if opts.format == "json" {
		return printJSON(response)
	}
	common.PrintSuccess("Field renamed")
	return nil
}

type fieldsDeleteOptions struct {
	workspace string
	field     string
	confirm   bool
	format    string
}

func newFieldsDeleteCmd() *cobra.Command {
	opts := fieldsDeleteOptions{}
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a field",
		Long: `Permanently delete a field from the workspace. This action cannot be undone.

The server refuses deletion while the field is still in use, for example as
a table's display field or when another field depends on it.`,
		Example: `  blue tables fields delete --workspace <id> --field <id> --confirm
  blue tables fields delete -w <id> --field <id> -y`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFieldsDelete(cmd, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.workspace, "workspace", "w", "", "Workspace ID or slug (required)")
	cmd.Flags().StringVar(&opts.field, "field", "", "Field ID (required)")
	cmd.Flags().BoolVarP(&opts.confirm, "confirm", "y", false, "Skip confirmation prompt")
	cmd.Flags().StringVar(&opts.format, "format", "text", "Output format: text, json")

	return cmd
}

func runFieldsDelete(cmd *cobra.Command, opts fieldsDeleteOptions) error {
	opts.workspace = strings.TrimSpace(opts.workspace)
	opts.field = strings.TrimSpace(opts.field)

	if err := requireFlag(opts.workspace, "workspace"); err != nil {
		return err
	}
	if err := requireFlag(opts.field, "field"); err != nil {
		return err
	}
	if !opts.confirm {
		return fmt.Errorf("deletion confirmation is required. Use --confirm flag")
	}

	client, workspaceID, err := tablesClient(opts.workspace)
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"workspaceId": workspaceID,
		"fieldId":     opts.field,
	}
	var response struct {
		FieldID string `json:"fieldId"`
		Deleted bool   `json:"deleted"`
	}
	if err := client.REST(http.MethodPost, "/delete-database-field", nil, body, &response); err != nil {
		return fmt.Errorf("failed to delete field: %w", err)
	}

	if opts.format == "json" {
		return printJSON(response)
	}
	common.PrintSuccess("Field deleted")
	return nil
}

type fieldsMountOptions struct {
	workspace string
	table     string
	field     string
	format    string
}

func newFieldsMountCmd() *cobra.Command {
	opts := fieldsMountOptions{}
	cmd := &cobra.Command{
		Use:   "mount",
		Short: "Mount a field on a table",
		Long: `Mount a field on a table so the table's records can use the field.

Fields are shared across tables in a workspace; mounting adds an existing
field to another table. Mounting an already-mounted field is a no-op.`,
		Example: `  blue tables fields mount --workspace <id> --table <id> --field <id>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFieldsMount(cmd, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.workspace, "workspace", "w", "", "Workspace ID or slug (required)")
	cmd.Flags().StringVar(&opts.table, "table", "", "Table ID (required)")
	cmd.Flags().StringVar(&opts.field, "field", "", "Field ID (required)")
	cmd.Flags().StringVar(&opts.format, "format", "text", "Output format: text, json")

	return cmd
}

func runFieldsMount(cmd *cobra.Command, opts fieldsMountOptions) error {
	opts.workspace = strings.TrimSpace(opts.workspace)
	opts.table = strings.TrimSpace(opts.table)
	opts.field = strings.TrimSpace(opts.field)

	if err := requireFlag(opts.workspace, "workspace"); err != nil {
		return err
	}
	if err := requireFlag(opts.table, "table"); err != nil {
		return err
	}
	if err := requireFlag(opts.field, "field"); err != nil {
		return err
	}

	client, workspaceID, err := tablesClient(opts.workspace)
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"workspaceId": workspaceID,
		"databaseId":  opts.table,
		"fieldId":     opts.field,
	}
	var response struct {
		DatabaseID string `json:"databaseId"`
		FieldID    string `json:"fieldId"`
		Mounted    bool   `json:"mounted"`
	}
	if err := client.REST(http.MethodPost, "/mount-database-field", nil, body, &response); err != nil {
		return fmt.Errorf("failed to mount field: %w", err)
	}

	if opts.format == "json" {
		return printJSON(response)
	}
	if response.Mounted {
		common.PrintSuccess("Field mounted on table")
	} else {
		common.PrintInfo("Field already mounted on table")
	}
	return nil
}

type fieldsUnmountOptions struct {
	workspace string
	table     string
	field     string
	format    string
}

func newFieldsUnmountCmd() *cobra.Command {
	opts := fieldsUnmountOptions{}
	cmd := &cobra.Command{
		Use:   "unmount",
		Short: "Unmount a field from a table",
		Long: `Unmount a field from a table so the table's records can no longer use it.

The field itself stays in the workspace and remains mounted on other tables.
Unmounting an already-unmounted field is a no-op.`,
		Example: `  blue tables fields unmount --workspace <id> --table <id> --field <id>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFieldsUnmount(cmd, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.workspace, "workspace", "w", "", "Workspace ID or slug (required)")
	cmd.Flags().StringVar(&opts.table, "table", "", "Table ID (required)")
	cmd.Flags().StringVar(&opts.field, "field", "", "Field ID (required)")
	cmd.Flags().StringVar(&opts.format, "format", "text", "Output format: text, json")

	return cmd
}

func runFieldsUnmount(cmd *cobra.Command, opts fieldsUnmountOptions) error {
	opts.workspace = strings.TrimSpace(opts.workspace)
	opts.table = strings.TrimSpace(opts.table)
	opts.field = strings.TrimSpace(opts.field)

	if err := requireFlag(opts.workspace, "workspace"); err != nil {
		return err
	}
	if err := requireFlag(opts.table, "table"); err != nil {
		return err
	}
	if err := requireFlag(opts.field, "field"); err != nil {
		return err
	}

	client, workspaceID, err := tablesClient(opts.workspace)
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"workspaceId": workspaceID,
		"databaseId":  opts.table,
		"fieldId":     opts.field,
	}
	var response struct {
		DatabaseID string `json:"databaseId"`
		FieldID    string `json:"fieldId"`
		Unmounted  bool   `json:"unmounted"`
	}
	if err := client.REST(http.MethodPost, "/unmount-database-field", nil, body, &response); err != nil {
		return fmt.Errorf("failed to unmount field: %w", err)
	}

	if opts.format == "json" {
		return printJSON(response)
	}
	if response.Unmounted {
		common.PrintSuccess("Field unmounted from table")
	} else {
		common.PrintInfo("Field already not mounted on table")
	}
	return nil
}
