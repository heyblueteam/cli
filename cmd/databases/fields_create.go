package databases

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

type fieldsCreateOptions struct {
	workspace     string
	database      string
	name          string
	fieldType     string
	configuration string
	fieldID       string
	format        string
}

type createFieldResponse struct {
	FieldID      string `json:"fieldId"`
	FieldCreated bool   `json:"fieldCreated"`
	Mounted      bool   `json:"mounted"`
	Rank         string `json:"rank"`
}

func newFieldsCreateCmd() *cobra.Command {
	opts := fieldsCreateOptions{}
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a field on a database",
		Long: `Create a field on a database.

The field type is an uppercase enum string, e.g. TEXT_SINGLE, NUMBER,
SELECT_SINGLE, DATE, CHECKBOX. The optional configuration is a JSON object
whose shape depends on the field type. Creating a field that already exists
is an idempotent retry: the command still succeeds and reports the field.`,
		Example: `  blue databases fields create --workspace <id> --database <id> --name "Priority" --type SELECT_SINGLE
  blue databases fields create -w <id> --database <id> --name "Estimate" --type NUMBER --configuration '{"precision":2}' --format json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runFieldsCreate(cmd, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.workspace, "workspace", "w", "", "Workspace ID or slug (required)")
	cmd.Flags().StringVar(&opts.database, "database", "", "Database ID (required)")
	cmd.Flags().StringVar(&opts.name, "name", "", "Field name (required)")
	cmd.Flags().StringVar(&opts.fieldType, "type", "", "Field type, e.g. TEXT_SINGLE, NUMBER, SELECT_SINGLE, DATE, CHECKBOX (required)")
	cmd.Flags().StringVar(&opts.configuration, "configuration", "", "Field configuration as a JSON object string")
	cmd.Flags().StringVar(&opts.fieldID, "field-id", "", "Field ID to use (defaults to a generated ID)")
	cmd.Flags().StringVar(&opts.format, "format", "text", "Output format: text, json")

	return cmd
}

func runFieldsCreate(cmd *cobra.Command, opts fieldsCreateOptions) error {
	opts.workspace = strings.TrimSpace(opts.workspace)
	opts.database = strings.TrimSpace(opts.database)
	opts.name = strings.TrimSpace(opts.name)
	opts.fieldType = strings.TrimSpace(opts.fieldType)

	if err := requireFlag(opts.workspace, "workspace"); err != nil {
		return err
	}
	if err := requireFlag(opts.database, "database"); err != nil {
		return err
	}
	if err := requireFlag(opts.name, "name"); err != nil {
		return err
	}
	if err := requireFlag(opts.fieldType, "type"); err != nil {
		return err
	}

	var configuration map[string]interface{}
	if strings.TrimSpace(opts.configuration) != "" {
		if err := json.Unmarshal([]byte(opts.configuration), &configuration); err != nil {
			return fmt.Errorf("invalid --configuration JSON: %w", err)
		}
	}

	fieldID := strings.TrimSpace(opts.fieldID)
	if fieldID == "" {
		fieldID = common.NewCuid()
	}

	client, workspaceID, err := databasesClient(opts.workspace)
	if err != nil {
		return err
	}

	body := map[string]interface{}{
		"workspaceId": workspaceID,
		"fieldId":     fieldID,
		"name":        opts.name,
		"type":        opts.fieldType,
	}
	if configuration != nil {
		body["configuration"] = configuration
	}

	var response createFieldResponse
	path := fmt.Sprintf("/databases/%s/fields", opts.database)
	if err := client.REST(http.MethodPost, path, nil, body, &response); err != nil {
		return fmt.Errorf("failed to create field: %w", err)
	}

	if opts.format == "json" {
		return printJSON(response)
	}

	common.PrintSuccess("Field created")
	out := cmd.OutOrStdout()
	fmt.Fprintf(out, "  Field ID: %s\n", response.FieldID)
	fmt.Fprintf(out, "  Rank: %s\n", response.Rank)
	fmt.Fprintf(out, "  Mounted: %t\n", response.Mounted)
	if !response.FieldCreated {
		common.PrintInfo("Field already existed (idempotent retry)")
	}
	return nil
}
