package fields

import (
	"fmt"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a field to a form",
	Long: `Add a field to a form via upsertFormField with an empty formFieldId.

For type=custom, --custom-field is required.`,
	Example: `  blue forms fields add --form <id> --type title --name "Full name" --required --position 1000
  blue forms fields add --form <id> --type custom --custom-field <cf-id> \
    --name "Priority" --placeholder "Pick one" --required --position 2000 --add-to-description`,
	RunE: runAdd,
}

var (
	addForm             string
	addWorkspace        string
	addFieldType        string
	addCustomField      string
	addName             string
	addPlaceholder      string
	addPosition         float64
	addRequired         bool
	addHidden           bool
	addExtraInfo        string
	addAddToDescription bool
)

var validFieldTypes = map[string]bool{
	"title": true, "description": true, "tags": true,
	"startedAt": true, "duedAt": true, "custom": true,
}

func init() {
	addCmd.Flags().StringVarP(&addForm, "form", "f", "", "Form ID (required)")
	addCmd.Flags().StringVarP(&addWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	addCmd.Flags().StringVar(&addFieldType, "type", "", "Field type (title|description|tags|startedAt|duedAt|custom) (required)")
	addCmd.Flags().StringVar(&addCustomField, "custom-field", "", "Custom field ID (required when --type=custom)")
	addCmd.Flags().StringVar(&addName, "name", "", "Display name (required)")
	addCmd.Flags().StringVar(&addPlaceholder, "placeholder", "", "Placeholder text")
	addCmd.Flags().Float64Var(&addPosition, "position", 0, "Position (sort key)")
	addCmd.Flags().BoolVar(&addRequired, "required", false, "Mark as required")
	addCmd.Flags().BoolVar(&addHidden, "hidden", false, "Hide the field")
	addCmd.Flags().StringVar(&addExtraInfo, "extra-info", "", "Helper text shown under the field")
	addCmd.Flags().BoolVar(&addAddToDescription, "add-to-description", false, "Append the answer to the record description")
}

func runAdd(cmd *cobra.Command, args []string) error {
	if addForm == "" {
		return fmt.Errorf("form ID is required. Use --form flag")
	}
	if addFieldType == "" {
		return fmt.Errorf("field type is required. Use --type flag")
	}
	if !validFieldTypes[addFieldType] {
		return fmt.Errorf("invalid type %q (use title|description|tags|startedAt|duedAt|custom)", addFieldType)
	}
	if addFieldType == "custom" && addCustomField == "" {
		return fmt.Errorf("--custom-field is required when --type=custom")
	}
	if addName == "" {
		return fmt.Errorf("--name is required")
	}
	if addWorkspace == "" {
		return fmt.Errorf("workspace is required. Use --workspace flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)
	client.SetProject(addWorkspace)

	input := map[string]interface{}{
		"formId":      addForm,
		"formFieldId": common.NewCuid(),
		"field":       addFieldType,
		"name":        addName,
	}
	if addCustomField != "" {
		input["customFieldId"] = addCustomField
	}
	if cmd.Flags().Changed("placeholder") {
		input["placeholder"] = addPlaceholder
	}
	if cmd.Flags().Changed("position") {
		input["position"] = addPosition
	}
	if cmd.Flags().Changed("required") {
		input["required"] = addRequired
	}
	// API defaults `hidden` to true when omitted — override so CLI-added
	// fields are visible by default.
	if cmd.Flags().Changed("hidden") {
		input["hidden"] = addHidden
	} else {
		input["hidden"] = false
	}
	if cmd.Flags().Changed("extra-info") {
		input["extraInfo"] = addExtraInfo
	}
	if cmd.Flags().Changed("add-to-description") {
		input["addToDescription"] = addAddToDescription
	}

	mutation := `
		mutation UpsertFormField($input: UpsertFormFieldInput!) {
			upsertFormField(input: $input) {
				id
				field
				name
				required
				position
				hidden
				addToDescription
				customField { id name type }
			}
		}
	`
	var resp struct {
		UpsertFormField struct {
			ID       string  `json:"id"`
			Field    string  `json:"field"`
			Name     string  `json:"name"`
			Required bool    `json:"required"`
			Position float64 `json:"position"`
		} `json:"upsertFormField"`
	}
	if err := client.ExecuteQueryWithResult(mutation, map[string]interface{}{"input": input}, &resp); err != nil {
		return fmt.Errorf("upsertFormField failed: %w", err)
	}
	fmt.Printf("Added field %s (id=%s) on form %s\n", resp.UpsertFormField.Name, resp.UpsertFormField.ID, addForm)
	return nil
}
