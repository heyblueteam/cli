package options

import (
	"fmt"
	"strings"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

// AddCustomFieldOptionsInput for adding options to existing custom fields
type AddCustomFieldOptionsInput struct {
	CustomFieldID      string                          `json:"customFieldId"`
	CustomFieldOptions []common.CustomFieldOptionInput `json:"customFieldOptions"`
}

// AddCustomFieldOptionsResponse represents the response from the mutation
type AddCustomFieldOptionsResponse struct {
	CreateCustomFieldOptions []common.CustomFieldOption `json:"createCustomFieldOptions"`
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create options for a custom field",
	Long:  "Add new options to an existing select-type custom field.",
	Example: `  blue fields options create --field <field-id> --options "High:red,Medium:yellow,Low:green"
  blue fields options create --field <field-id> --workspace <id> --options "Option1,Option2" --simple`,
	RunE: runCreate,
}

var (
	createField     string
	createWorkspace string
	createOptions   string
	createSimple    bool
)

func init() {
	createCmd.Flags().StringVar(&createField, "field", "", "Custom field ID to add options to (required)")
	createCmd.Flags().StringVarP(&createWorkspace, "workspace", "w", "", "Workspace ID or slug (optional - improves authorization)")
	createCmd.Flags().StringVar(&createOptions, "options", "", "Options in format 'Title1:color1,Title2:color2' (required)")
	createCmd.Flags().BoolVarP(&createSimple, "simple", "s", false, "Simple output format")
}

func runCreate(cmd *cobra.Command, args []string) error {
	if createField == "" {
		return fmt.Errorf("--field parameter is required")
	}

	if createOptions == "" {
		return fmt.Errorf("--options parameter is required. Format: 'Option1:red,Option2:blue'")
	}

	// Parse options string
	optionInputs, err := parseOptionsFromString(createOptions)
	if err != nil {
		return fmt.Errorf("parsing options: %w", err)
	}

	if len(optionInputs) == 0 {
		return fmt.Errorf("no valid options provided")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := common.NewClient(config)

	if createWorkspace != "" {
		client.SetProject(createWorkspace)
	}

	mutation := `
		mutation CreateCustomFieldOptions($input: CreateCustomFieldOptionsInput!) {
			createCustomFieldOptions(input: $input) {
				id
				title
				color
			}
		}
	`

	variables := map[string]interface{}{
		"input": AddCustomFieldOptionsInput{
			CustomFieldID:      createField,
			CustomFieldOptions: optionInputs,
		},
	}

	var result AddCustomFieldOptionsResponse
	if err := client.ExecuteQueryWithResult(mutation, variables, &result); err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	if createSimple {
		fmt.Printf("Added %d options to custom field %s\n", len(result.CreateCustomFieldOptions), createField)
	} else {
		fmt.Printf("Adding %d options to custom field '%s'...\n\n", len(optionInputs), createField)
		fmt.Println("Options added successfully!")
		fmt.Println("\nOptions created:")
		for _, option := range result.CreateCustomFieldOptions {
			if option.Color != "" {
				fmt.Printf("  - %s (color: %s) [ID: %s]\n", option.Title, option.Color, option.ID)
			} else {
				fmt.Printf("  - %s [ID: %s]\n", option.Title, option.ID)
			}
		}
		fmt.Printf("\nYou can now use these options when creating or updating records.\n")
	}

	return nil
}

// parseOptionsFromString parses the options string format "Option1:color1,Option2:color2"
func parseOptionsFromString(optionsStr string) ([]common.CustomFieldOptionInput, error) {
	var options []common.CustomFieldOptionInput

	if optionsStr == "" {
		return options, nil
	}

	parts := strings.Split(optionsStr, ",")
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		titleColor := strings.SplitN(part, ":", 2)
		title := strings.TrimSpace(titleColor[0])

		if title == "" {
			continue
		}

		option := common.CustomFieldOptionInput{
			Title: title,
		}

		if len(titleColor) > 1 {
			color := strings.TrimSpace(titleColor[1])
			if color != "" {
				option.Color = color
			}
		}

		options = append(options, option)
	}

	return options, nil
}
