package fields

import (
	"fmt"
	"strconv"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing form field",
	Long: `Update a form field via upsertFormField with a known formFieldId.

The field's --type and --custom-field cannot be changed after creation; to
change those, delete and re-add.`,
	Example: `  blue forms fields update --field <ff-id> --form <form-id> --name "New label"
  blue forms fields update --field <ff-id> --form <form-id> --required true --position 1500`,
	RunE: runUpdate,
}

var (
	updField             string
	updForm              string
	updWorkspace         string
	updFieldType         string
	updCustomField       string
	updName              string
	updPlaceholder       string
	updPosition          float64
	updRequired          string
	updHidden            string
	updExtraInfo         string
	updAddToDescription  string
)

func init() {
	updateCmd.Flags().StringVar(&updField, "field", "", "Form field ID (required)")
	updateCmd.Flags().StringVarP(&updForm, "form", "f", "", "Form ID (required)")
	updateCmd.Flags().StringVarP(&updWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	updateCmd.Flags().StringVar(&updFieldType, "type", "", "Field type (only set when re-asserting; cannot change after creation)")
	updateCmd.Flags().StringVar(&updCustomField, "custom-field", "", "Custom field ID (only set when re-asserting)")
	updateCmd.Flags().StringVar(&updName, "name", "", "Display name")
	updateCmd.Flags().StringVar(&updPlaceholder, "placeholder", "", "Placeholder text")
	updateCmd.Flags().Float64Var(&updPosition, "position", 0, "Position (sort key)")
	updateCmd.Flags().StringVar(&updRequired, "required", "", "true|false")
	updateCmd.Flags().StringVar(&updHidden, "hidden", "", "true|false")
	updateCmd.Flags().StringVar(&updExtraInfo, "extra-info", "", "Helper text shown under the field")
	updateCmd.Flags().StringVar(&updAddToDescription, "add-to-description", "", "true|false")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if updField == "" {
		return fmt.Errorf("form field ID is required. Use --field flag")
	}
	if updForm == "" {
		return fmt.Errorf("form ID is required. Use --form flag")
	}
	if updWorkspace == "" {
		return fmt.Errorf("workspace is required. Use --workspace flag")
	}

	// upsertFormField requires `field` even on update — fetch the current
	// value if the caller didn't provide one, so we don't accidentally change
	// the field type.
	fieldType := updFieldType
	customFieldID := updCustomField
	if fieldType == "" || (fieldType == "custom" && customFieldID == "") {
		current, err := fetchFormField(updForm, updWorkspace, updField)
		if err != nil {
			return err
		}
		if fieldType == "" {
			fieldType = current.Field
		}
		if customFieldID == "" && current.CustomField != nil {
			customFieldID = current.CustomField.ID
		}
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)
	client.SetProject(updWorkspace)

	input := map[string]interface{}{
		"formId":      updForm,
		"formFieldId": updField,
		"field":       fieldType,
	}
	if customFieldID != "" {
		input["customFieldId"] = customFieldID
	}
	if updName != "" {
		input["name"] = updName
	}
	if cmd.Flags().Changed("placeholder") {
		input["placeholder"] = updPlaceholder
	}
	if cmd.Flags().Changed("position") {
		input["position"] = updPosition
	}
	if updRequired != "" {
		b, err := strconv.ParseBool(updRequired)
		if err != nil {
			return fmt.Errorf("invalid --required %q (use true|false)", updRequired)
		}
		input["required"] = b
	}
	if updHidden != "" {
		b, err := strconv.ParseBool(updHidden)
		if err != nil {
			return fmt.Errorf("invalid --hidden %q (use true|false)", updHidden)
		}
		input["hidden"] = b
	}
	if updAddToDescription != "" {
		b, err := strconv.ParseBool(updAddToDescription)
		if err != nil {
			return fmt.Errorf("invalid --add-to-description %q (use true|false)", updAddToDescription)
		}
		input["addToDescription"] = b
	}
	if cmd.Flags().Changed("extra-info") {
		input["extraInfo"] = updExtraInfo
	}

	mutation := `
		mutation UpsertFormField($input: UpsertFormFieldInput!) {
			upsertFormField(input: $input) { id name }
		}
	`
	var resp struct {
		UpsertFormField struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		} `json:"upsertFormField"`
	}
	if err := client.ExecuteQueryWithResult(mutation, map[string]interface{}{"input": input}, &resp); err != nil {
		return fmt.Errorf("upsertFormField failed: %w", err)
	}
	fmt.Printf("Updated field %s (id=%s)\n", resp.UpsertFormField.Name, resp.UpsertFormField.ID)
	return nil
}

type currentField struct {
	Field       string `json:"field"`
	CustomField *struct {
		ID string `json:"id"`
	} `json:"customField"`
}

func fetchFormField(formID, workspace, fieldID string) (*currentField, error) {
	config, err := common.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)
	client.SetProject(workspace)
	query := `
		query FormFields($filter: FormFieldFilterInput) {
			formFields(filter: $filter) {
				id
				field
				customField { id }
			}
		}
	`
	var resp struct {
		FormFields []struct {
			ID          string `json:"id"`
			Field       string `json:"field"`
			CustomField *struct {
				ID string `json:"id"`
			} `json:"customField"`
		} `json:"formFields"`
	}
	variables := map[string]interface{}{
		"filter": map[string]interface{}{"formId": formID},
	}
	if err := client.ExecuteQueryWithResult(query, variables, &resp); err != nil {
		return nil, fmt.Errorf("failed to read form fields: %w", err)
	}
	for _, f := range resp.FormFields {
		if f.ID == fieldID {
			return &currentField{Field: f.Field, CustomField: f.CustomField}, nil
		}
	}
	return nil, fmt.Errorf("form field %s not found on form %s", fieldID, formID)
}
