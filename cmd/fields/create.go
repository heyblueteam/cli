package fields

import (
	"fmt"
	"strings"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

// LocalCreateCustomFieldInput includes all fields needed for custom field creation
type LocalCreateCustomFieldInput struct {
	Name                   string                        `json:"name"`
	Type                   string                        `json:"type"`
	Description            string                        `json:"description,omitempty"`
	ButtonType             string                        `json:"buttonType,omitempty"`
	ButtonConfirmText      string                        `json:"buttonConfirmText,omitempty"`
	CurrencyFieldID        string                        `json:"currencyFieldId,omitempty"`
	ConversionDate         string                        `json:"conversionDate,omitempty"`
	ConversionDateType     string                        `json:"conversionDateType,omitempty"`
	Min                    *float64                      `json:"min,omitempty"`
	Max                    *float64                      `json:"max,omitempty"`
	Currency               string                        `json:"currency,omitempty"`
	Prefix                 string                        `json:"prefix,omitempty"`
	IsDueDate              *bool                         `json:"isDueDate,omitempty"`
	Formula                interface{}                   `json:"formula,omitempty"`
	Metadata               interface{}                   `json:"metadata,omitempty"`
	TimeDurationDisplay    string                        `json:"timeDurationDisplay,omitempty"`
	TimeDurationTargetTime *float64                      `json:"timeDurationTargetTime,omitempty"`
	TimeDurationStartInput *CustomFieldTimeDurationInput `json:"timeDurationStartInput,omitempty"`
	TimeDurationEndInput   *CustomFieldTimeDurationInput `json:"timeDurationEndInput,omitempty"`
	ReferenceProjectID     string                        `json:"referenceProjectId,omitempty"`
	ReferenceFilter        interface{}                   `json:"referenceFilter,omitempty"`
	ReferenceMultiple      *bool                         `json:"referenceMultiple,omitempty"`
	LookupOption           *CustomFieldLookupOptionInput `json:"lookupOption,omitempty"`
	UseSequenceUniqueID    *bool                         `json:"useSequenceUniqueId,omitempty"`
	SequenceDigits         *int                          `json:"sequenceDigits,omitempty"`
	SequenceStartingNumber *int                          `json:"sequenceStartingNumber,omitempty"`
}

// CustomFieldTimeDurationInput for time duration fields
type CustomFieldTimeDurationInput struct {
	Type                 string   `json:"type"`
	Condition            string   `json:"condition"`
	CustomFieldID        string   `json:"customFieldId,omitempty"`
	CustomFieldOptionIDs []string `json:"customFieldOptionIds,omitempty"`
	TodoListID           string   `json:"todoListId,omitempty"`
	TagID                string   `json:"tagId,omitempty"`
	AssigneeID           string   `json:"assigneeId,omitempty"`
}

// CustomFieldLookupOptionInput for lookup fields
type CustomFieldLookupOptionInput struct {
	ReferenceID string `json:"referenceId"`
	LookupID    string `json:"lookupId,omitempty"`
	LookupType  string `json:"lookupType"`
}

// CreateCustomFieldOptionsInput for creating options after field creation
type CreateCustomFieldOptionsInput struct {
	CustomFieldID      string                          `json:"customFieldId"`
	CustomFieldOptions []common.CustomFieldOptionInput `json:"customFieldOptions"`
}

// CreatedCustomField represents the response from field creation
type CreatedCustomField struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

// CreateCustomFieldResponse wraps the mutation response
type CreateCustomFieldResponse struct {
	CreateCustomField CreatedCustomField `json:"createCustomField"`
}

// Available custom field types
var customFieldTypes = []string{
	"CHECKBOX", "CURRENCY", "EMAIL", "LOCATION", "NUMBER", "PERCENT",
	"PHONE", "RATING", "SELECT_MULTI", "SELECT_SINGLE", "TEXT_MULTI",
	"TEXT_SINGLE", "UNIQUE_ID", "URL", "FILE", "COUNTRY", "DATE",
	"FORMULA", "REFERENCE", "LOOKUP", "TIME_DURATION", "BUTTON",
	"CURRENCY_CONVERSION",
}

// Common currencies
var currencies = []string{
	"USD", "EUR", "GBP", "JPY", "CAD", "AUD", "CHF", "CNY", "INR", "BRL",
}

// Time duration types
var timeDurationTypes = []string{
	"CREATED_AT", "DUE_DATE", "STARTED_AT", "COMPLETED_AT", "CUSTOM_FIELD",
}

// Time duration conditions
var timeDurationConditions = []string{
	"EQUALS", "NOT_EQUALS", "GREATER_THAN", "LESS_THAN", "BETWEEN",
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a custom field",
	Long:  "Create a new custom field in a workspace.",
	Example: `  blue fields create --workspace <id> --name "Priority" --type SELECT_SINGLE --options "High:red,Medium:yellow,Low:green"
  blue fields create -w <id> --name "Story Points" --type NUMBER --min 1 --max 13
  blue fields create -w <id> --name "Cost" --type CURRENCY --currency USD
  blue fields create --list-types`,
	RunE: runCreate,
}

var (
	createWorkspace            string
	createName                 string
	createType                 string
	createDescription          string
	createOptions              string
	createButtonType           string
	createButtonConfirmText    string
	createCurrencyFieldID      string
	createConversionDate       string
	createConversionDateType   string
	createMin                  float64
	createMax                  float64
	createCurrency             string
	createPrefix               string
	createIsDueDate            bool
	createTimeDurationDisplay  string
	createTimeDurationTarget   float64
	createReferenceProject     string
	createReferenceMultiple    bool
	createUseSequence          bool
	createSequenceDigits       int
	createSequenceStart        int
	createListTypes            bool
)

func init() {
	createCmd.Flags().StringVarP(&createWorkspace, "workspace", "w", "", "Workspace ID (required)")
	createCmd.Flags().StringVar(&createName, "name", "", "Custom field name (required)")
	createCmd.Flags().StringVar(&createType, "type", "", "Custom field type (required)")
	createCmd.Flags().StringVar(&createDescription, "description", "", "Custom field description")
	createCmd.Flags().StringVar(&createOptions, "options", "", "Options for SELECT fields (format: 'value1:color1,value2:color2')")
	createCmd.Flags().StringVar(&createButtonType, "button-type", "", "Button type for BUTTON field type")
	createCmd.Flags().StringVar(&createButtonConfirmText, "button-confirm-text", "", "Button confirmation text")
	createCmd.Flags().StringVar(&createCurrencyFieldID, "currency-field-id", "", "Currency field ID for CURRENCY_CONVERSION type")
	createCmd.Flags().StringVar(&createConversionDate, "conversion-date", "", "Conversion date")
	createCmd.Flags().StringVar(&createConversionDateType, "conversion-date-type", "", "Conversion date type")
	createCmd.Flags().Float64Var(&createMin, "min", 0, "Minimum value for NUMBER type")
	createCmd.Flags().Float64Var(&createMax, "max", 0, "Maximum value for NUMBER type")
	createCmd.Flags().StringVar(&createCurrency, "currency", "USD", "Currency code")
	createCmd.Flags().StringVar(&createPrefix, "prefix", "", "Field prefix")
	createCmd.Flags().BoolVar(&createIsDueDate, "is-due-date", false, "Whether this field represents a due date")
	createCmd.Flags().StringVar(&createTimeDurationDisplay, "time-duration-display", "", "Time duration display type")
	createCmd.Flags().Float64Var(&createTimeDurationTarget, "time-duration-target", 0, "Time duration target time")
	createCmd.Flags().StringVar(&createReferenceProject, "reference-project", "", "Reference project ID for REFERENCE type")
	createCmd.Flags().BoolVar(&createReferenceMultiple, "reference-multiple", false, "Allow multiple references")
	createCmd.Flags().BoolVar(&createUseSequence, "use-sequence", false, "Use sequence unique ID")
	createCmd.Flags().IntVar(&createSequenceDigits, "sequence-digits", 6, "Number of digits in sequence")
	createCmd.Flags().IntVar(&createSequenceStart, "sequence-start", 1, "Starting number for sequence")
	createCmd.Flags().BoolVar(&createListTypes, "list-types", false, "List available field types and options")
}

func runCreate(cmd *cobra.Command, args []string) error {
	// Show available options if requested
	if createListTypes {
		fmt.Println("\n=== Available Custom Field Types ===")
		for _, t := range customFieldTypes {
			fmt.Printf("  - %s\n", t)
		}

		fmt.Println("\n=== Available Currencies ===")
		for _, c := range currencies {
			fmt.Printf("  - %s\n", c)
		}

		fmt.Println("\n=== Available Time Duration Types ===")
		for _, t := range timeDurationTypes {
			fmt.Printf("  - %s\n", t)
		}

		fmt.Println("\n=== Available Time Duration Conditions ===")
		for _, c := range timeDurationConditions {
			fmt.Printf("  - %s\n", c)
		}
		return nil
	}

	if createName == "" {
		return fmt.Errorf("custom field name is required. Use --name flag")
	}
	if createType == "" {
		return fmt.Errorf("custom field type is required. Use --type flag")
	}
	if createWorkspace == "" {
		return fmt.Errorf("workspace ID is required. Use --workspace flag")
	}

	// Validate field type
	validType := false
	for _, t := range customFieldTypes {
		if createType == t {
			validType = true
			break
		}
	}
	if !validType {
		return fmt.Errorf("invalid field type '%s'. Use --list-types flag to see available types", createType)
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)
	client.SetProject(createWorkspace)

	// Parse options if provided
	parsedOptions, err := parseOptions(createOptions)
	if err != nil {
		return fmt.Errorf("failed to parse options: %w", err)
	}

	// Create custom field input
	input := LocalCreateCustomFieldInput{
		Name:                   createName,
		Type:                   createType,
		Description:            createDescription,
		ButtonType:             createButtonType,
		ButtonConfirmText:      createButtonConfirmText,
		CurrencyFieldID:        createCurrencyFieldID,
		ConversionDate:         createConversionDate,
		ConversionDateType:     createConversionDateType,
		Currency:               createCurrency,
		Prefix:                 createPrefix,
		IsDueDate:              &createIsDueDate,
		TimeDurationDisplay:    createTimeDurationDisplay,
		ReferenceProjectID:     createReferenceProject,
		ReferenceMultiple:      &createReferenceMultiple,
		UseSequenceUniqueID:    &createUseSequence,
		SequenceDigits:         &createSequenceDigits,
		SequenceStartingNumber: &createSequenceStart,
	}

	if createMin != 0 {
		input.Min = &createMin
	}
	if createMax != 0 {
		input.Max = &createMax
	}
	if createTimeDurationTarget != 0 {
		input.TimeDurationTargetTime = &createTimeDurationTarget
	}

	fmt.Printf("Creating custom field '%s' of type '%s'...\n", input.Name, input.Type)

	customField, err := executeCreateCustomField(client, input)
	if err != nil {
		return fmt.Errorf("failed to create custom field: %w", err)
	}

	fmt.Println("\nCustom field created successfully!")
	fmt.Printf("\nCustom Field Details:\n")
	fmt.Printf("  ID:          %s\n", customField.ID)
	fmt.Printf("  Name:        %s\n", customField.Name)
	fmt.Printf("  Type:        %s\n", customField.Type)
	if customField.Description != "" {
		fmt.Printf("  Description: %s\n", customField.Description)
	}

	// Create options if provided and field type supports them
	if len(parsedOptions) > 0 && (createType == "SELECT_SINGLE" || createType == "SELECT_MULTI") {
		fmt.Printf("\nCreating %d options for the field...\n", len(parsedOptions))

		if err := createCustomFieldOptions(client, customField.ID, parsedOptions); err != nil {
			fmt.Printf("Warning: Field created successfully but failed to create options: %v\n", err)
			fmt.Printf("You can manually add options later.\n")
		} else {
			fmt.Printf("Options created successfully!\n")
			fmt.Printf("\nOptions created:\n")
			for _, option := range parsedOptions {
				if option.Color != "" {
					fmt.Printf("  - %s (color: %s)\n", option.Title, option.Color)
				} else {
					fmt.Printf("  - %s\n", option.Title)
				}
			}
		}
	} else if len(parsedOptions) > 0 {
		fmt.Printf("\nWarning: Options provided but field type '%s' doesn't support options. Options were ignored.\n", createType)
	}

	fmt.Printf("\nYou can now use this custom field in your records.\n")

	return nil
}

// executeCreateCustomField executes the GraphQL mutation
func executeCreateCustomField(client *common.Client, input LocalCreateCustomFieldInput) (*CreatedCustomField, error) {
	optionalFields := buildOptionalFields(input)
	if optionalFields != "" {
		optionalFields = "\n\t\t\t\t" + optionalFields
	}

	mutation := fmt.Sprintf(`
		mutation CreateCustomField {
			createCustomField(input: {
				name: "%s"
				type: %s%s
			}) {
				id
				name
				type
				description
			}
		}
	`, input.Name, input.Type, optionalFields)

	var response CreateCustomFieldResponse
	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		return nil, err
	}

	return &response.CreateCustomField, nil
}

// buildOptionalFields builds optional fields for the mutation
func buildOptionalFields(input LocalCreateCustomFieldInput) string {
	var fields []string

	if input.Description != "" {
		fields = append(fields, fmt.Sprintf(`description: "%s"`, input.Description))
	}
	if input.ButtonType != "" {
		fields = append(fields, fmt.Sprintf(`buttonType: "%s"`, input.ButtonType))
	}
	if input.ButtonConfirmText != "" {
		fields = append(fields, fmt.Sprintf(`buttonConfirmText: "%s"`, input.ButtonConfirmText))
	}
	if input.CurrencyFieldID != "" {
		fields = append(fields, fmt.Sprintf(`currencyFieldId: "%s"`, input.CurrencyFieldID))
	}
	if input.ConversionDate != "" {
		fields = append(fields, fmt.Sprintf(`conversionDate: "%s"`, input.ConversionDate))
	}
	if input.ConversionDateType != "" {
		fields = append(fields, fmt.Sprintf(`conversionDateType: "%s"`, input.ConversionDateType))
	}
	if input.Min != nil {
		fields = append(fields, fmt.Sprintf(`min: %f`, *input.Min))
	}
	if input.Max != nil {
		fields = append(fields, fmt.Sprintf(`max: %f`, *input.Max))
	}
	if input.Currency != "" && input.Currency != "USD" {
		fields = append(fields, fmt.Sprintf(`currency: "%s"`, input.Currency))
	}
	if input.Prefix != "" {
		fields = append(fields, fmt.Sprintf(`prefix: "%s"`, input.Prefix))
	}
	if input.IsDueDate != nil && *input.IsDueDate {
		fields = append(fields, fmt.Sprintf(`isDueDate: %t`, *input.IsDueDate))
	}
	if input.TimeDurationDisplay != "" {
		fields = append(fields, fmt.Sprintf(`timeDurationDisplay: %s`, input.TimeDurationDisplay))
	}
	if input.TimeDurationTargetTime != nil {
		fields = append(fields, fmt.Sprintf(`timeDurationTargetTime: %f`, *input.TimeDurationTargetTime))
	}
	if input.ReferenceProjectID != "" {
		fields = append(fields, fmt.Sprintf(`referenceProjectId: "%s"`, input.ReferenceProjectID))
	}
	if input.ReferenceMultiple != nil && *input.ReferenceMultiple {
		fields = append(fields, fmt.Sprintf(`referenceMultiple: %t`, *input.ReferenceMultiple))
	}
	if input.UseSequenceUniqueID != nil && *input.UseSequenceUniqueID {
		fields = append(fields, fmt.Sprintf(`useSequenceUniqueId: %t`, *input.UseSequenceUniqueID))
	}
	if input.SequenceDigits != nil && *input.SequenceDigits != 6 {
		fields = append(fields, fmt.Sprintf(`sequenceDigits: %d`, *input.SequenceDigits))
	}
	if input.SequenceStartingNumber != nil && *input.SequenceStartingNumber != 1 {
		fields = append(fields, fmt.Sprintf(`sequenceStartingNumber: %d`, *input.SequenceStartingNumber))
	}

	if input.TimeDurationStartInput != nil {
		startFields := buildTimeDurationInput(input.TimeDurationStartInput)
		if startFields != "" {
			fields = append(fields, fmt.Sprintf(`timeDurationStartInput: { %s }`, startFields))
		}
	}

	if input.TimeDurationEndInput != nil {
		endFields := buildTimeDurationInput(input.TimeDurationEndInput)
		if endFields != "" {
			fields = append(fields, fmt.Sprintf(`timeDurationEndInput: { %s }`, endFields))
		}
	}

	if input.LookupOption != nil {
		lookupFields := buildLookupOptionInput(input.LookupOption)
		if lookupFields != "" {
			fields = append(fields, fmt.Sprintf(`lookupOption: { %s }`, lookupFields))
		}
	}

	return strings.Join(fields, "\n\t\t\t\t")
}

// buildTimeDurationInput builds time duration input fields
func buildTimeDurationInput(input *CustomFieldTimeDurationInput) string {
	var fields []string

	if input.Type != "" {
		fields = append(fields, fmt.Sprintf(`type: %s`, input.Type))
	}
	if input.Condition != "" {
		fields = append(fields, fmt.Sprintf(`condition: %s`, input.Condition))
	}
	if input.CustomFieldID != "" {
		fields = append(fields, fmt.Sprintf(`customFieldId: "%s"`, input.CustomFieldID))
	}
	if len(input.CustomFieldOptionIDs) > 0 {
		optionIDs := make([]string, len(input.CustomFieldOptionIDs))
		for i, id := range input.CustomFieldOptionIDs {
			optionIDs[i] = fmt.Sprintf(`"%s"`, id)
		}
		fields = append(fields, fmt.Sprintf(`customFieldOptionIds: [%s]`, strings.Join(optionIDs, ", ")))
	}
	if input.TodoListID != "" {
		fields = append(fields, fmt.Sprintf(`todoListId: "%s"`, input.TodoListID))
	}
	if input.TagID != "" {
		fields = append(fields, fmt.Sprintf(`tagId: "%s"`, input.TagID))
	}
	if input.AssigneeID != "" {
		fields = append(fields, fmt.Sprintf(`assigneeId: "%s"`, input.AssigneeID))
	}

	return strings.Join(fields, "\n\t\t\t\t\t")
}

// buildLookupOptionInput builds lookup option input fields
func buildLookupOptionInput(input *CustomFieldLookupOptionInput) string {
	var fields []string

	if input.ReferenceID != "" {
		fields = append(fields, fmt.Sprintf(`referenceId: "%s"`, input.ReferenceID))
	}
	if input.LookupID != "" {
		fields = append(fields, fmt.Sprintf(`lookupId: "%s"`, input.LookupID))
	}
	if input.LookupType != "" {
		fields = append(fields, fmt.Sprintf(`lookupType: %s`, input.LookupType))
	}

	return strings.Join(fields, "\n\t\t\t\t\t")
}

// parseOptions parses options string into CustomFieldOptionInput slice
func parseOptions(optionsStr string) ([]common.CustomFieldOptionInput, error) {
	if optionsStr == "" {
		return nil, nil
	}

	var options []common.CustomFieldOptionInput
	pairs := strings.Split(optionsStr, ",")

	for _, pair := range pairs {
		parts := strings.Split(strings.TrimSpace(pair), ":")
		if len(parts) < 1 || parts[0] == "" {
			continue
		}

		option := common.CustomFieldOptionInput{
			Title: parts[0],
		}

		if len(parts) > 1 && parts[1] != "" {
			option.Color = parts[1]
		}

		options = append(options, option)
	}

	return options, nil
}

// createCustomFieldOptions creates options after field creation
func createCustomFieldOptions(client *common.Client, customFieldID string, options []common.CustomFieldOptionInput) error {
	if len(options) == 0 {
		return nil
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
		"input": CreateCustomFieldOptionsInput{
			CustomFieldID:      customFieldID,
			CustomFieldOptions: options,
		},
	}

	var response map[string]interface{}
	return client.ExecuteQueryWithResult(mutation, variables, &response)
}
