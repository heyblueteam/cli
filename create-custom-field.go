package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
)

// Custom field option input for creating options after field creation
type CustomFieldOptionInput struct {
	Title string `json:"title"`
	Color string `json:"color,omitempty"`
}

// Input for creating multiple custom field options
type CreateCustomFieldOptionsInput struct {
	CustomFieldID      string                    `json:"customFieldId"`
	CustomFieldOptions []CustomFieldOptionInput `json:"customFieldOptions"`
}

// Custom field creation input
type CreateCustomFieldInput struct {
	Name                      string                    `json:"name"`
	Type                      string                    `json:"type"`
	Description               string                    `json:"description,omitempty"`
	ButtonType                string                    `json:"buttonType,omitempty"`
	ButtonConfirmText         string                    `json:"buttonConfirmText,omitempty"`
	CurrencyFieldID           string                    `json:"currencyFieldId,omitempty"`
	ConversionDate            string                    `json:"conversionDate,omitempty"`
	ConversionDateType        string                    `json:"conversionDateType,omitempty"`
	Min                       *float64                  `json:"min,omitempty"`
	Max                       *float64                  `json:"max,omitempty"`
	Currency                  string                    `json:"currency,omitempty"`
	Prefix                    string                    `json:"prefix,omitempty"`
	IsDueDate                 *bool                     `json:"isDueDate,omitempty"`
	Formula                   interface{}               `json:"formula,omitempty"`
	Metadata                  interface{}               `json:"metadata,omitempty"`
	TimeDurationDisplay       string                    `json:"timeDurationDisplay,omitempty"`
	TimeDurationTargetTime    *float64                  `json:"timeDurationTargetTime,omitempty"`
	TimeDurationStartInput    *CustomFieldTimeDurationInput `json:"timeDurationStartInput,omitempty"`
	TimeDurationEndInput      *CustomFieldTimeDurationInput `json:"timeDurationEndInput,omitempty"`
	ReferenceProjectID        string                    `json:"referenceProjectId,omitempty"`
	ReferenceFilter           interface{}               `json:"referenceFilter,omitempty"`
	ReferenceMultiple         *bool                    `json:"referenceMultiple,omitempty"`
	LookupOption              *CustomFieldLookupOptionInput `json:"lookupOption,omitempty"`
	UseSequenceUniqueID       *bool                    `json:"useSequenceUniqueId,omitempty"`
	SequenceDigits            *int                     `json:"sequenceDigits,omitempty"`
	SequenceStartingNumber    *int                     `json:"sequenceStartingNumber,omitempty"`
}

// Custom field time duration input
type CustomFieldTimeDurationInput struct {
	Type              string   `json:"type"`
	Condition         string   `json:"condition"`
	CustomFieldID     string   `json:"customFieldId,omitempty"`
	CustomFieldOptionIDs []string `json:"customFieldOptionIds,omitempty"`
	TodoListID        string   `json:"todoListId,omitempty"`
	TagID             string   `json:"tagId,omitempty"`
	AssigneeID        string   `json:"assigneeId,omitempty"`
}

// Custom field lookup option input
type CustomFieldLookupOptionInput struct {
	ReferenceID string `json:"referenceId"`
	LookupID    string `json:"lookupId,omitempty"`
	LookupType  string `json:"lookupType"`
}

// Response structures
type CreatedCustomField struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
}

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

// Execute GraphQL mutation
func executeCreateCustomField(client *Client, input CreateCustomFieldInput) (*CreatedCustomField, error) {
	// Build the mutation
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


	// Execute mutation
	var response CreateCustomFieldResponse
	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		return nil, err
	}

	return &response.CreateCustomField, nil
}

// Build optional fields for the mutation
func buildOptionalFields(input CreateCustomFieldInput) string {
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

	// Handle complex nested objects
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

// Build time duration input fields
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

// Build lookup option input fields
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

// Parse options string into CustomFieldOptionInput slice
func parseOptions(optionsStr string) ([]CustomFieldOptionInput, error) {
	if optionsStr == "" {
		return nil, nil
	}

	var options []CustomFieldOptionInput
	pairs := strings.Split(optionsStr, ",")
	
	for _, pair := range pairs {
		parts := strings.Split(strings.TrimSpace(pair), ":")
		if len(parts) < 1 || parts[0] == "" {
			continue
		}
		
		option := CustomFieldOptionInput{
			Title: parts[0],
		}
		
		// Add color if provided
		if len(parts) > 1 && parts[1] != "" {
			option.Color = parts[1]
		}
		
		options = append(options, option)
	}
	
	return options, nil
}

// Create custom field options after field creation
func createCustomFieldOptions(client *Client, customFieldID string, options []CustomFieldOptionInput) error {
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

func main() {
	// Parse command line flags
	name := flag.String("name", "", "Custom field name (required)")
	fieldType := flag.String("type", "", "Custom field type (required)")
	projectID := flag.String("project", "", "Project ID (required for project-level custom fields)")
	description := flag.String("description", "", "Custom field description")
	options := flag.String("options", "", "Options for SELECT fields (format: 'value1:color1,value2:color2')")
	buttonType := flag.String("button-type", "", "Button type for BUTTON field type")
	buttonConfirmText := flag.String("button-confirm-text", "", "Button confirmation text")
	currencyFieldID := flag.String("currency-field-id", "", "Currency field ID for CURRENCY_CONVERSION type")
	conversionDate := flag.String("conversion-date", "", "Conversion date")
	conversionDateType := flag.String("conversion-date-type", "", "Conversion date type")
	min := flag.Float64("min", 0, "Minimum value for NUMBER type")
	max := flag.Float64("max", 0, "Maximum value for NUMBER type")
	currency := flag.String("currency", "USD", "Currency code")
	prefix := flag.String("prefix", "", "Field prefix")
	isDueDate := flag.Bool("is-due-date", false, "Whether this field represents a due date")
	timeDurationDisplay := flag.String("time-duration-display", "", "Time duration display type")
	timeDurationTargetTime := flag.Float64("time-duration-target", 0, "Time duration target time")
	referenceProjectID := flag.String("reference-project", "", "Reference project ID for REFERENCE type")
	referenceMultiple := flag.Bool("reference-multiple", false, "Allow multiple references")
	useSequenceUniqueID := flag.Bool("use-sequence", false, "Use sequence unique ID")
	sequenceDigits := flag.Int("sequence-digits", 6, "Number of digits in sequence")
	sequenceStartingNumber := flag.Int("sequence-start", 1, "Starting number for sequence")
	listOptions := flag.Bool("list", false, "List available options")
	flag.Parse()

	// Show available options if requested
	if *listOptions {
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
		return
	}

	// Validate required parameters
	if *name == "" {
		log.Fatal("Custom field name is required. Use -name flag")
	}
	if *fieldType == "" {
		log.Fatal("Custom field type is required. Use -type flag")
	}
	if *projectID == "" {
		log.Fatal("Project ID is required. Use -project flag")
	}

	// Validate field type
	validType := false
	for _, t := range customFieldTypes {
		if *fieldType == t {
			validType = true
			break
		}
	}
	if !validType {
		log.Fatalf("Invalid field type '%s'. Use -list flag to see available types", *fieldType)
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create client
	client := NewClient(config)
	
	// Set project context for the request
	client.SetProjectID(*projectID)

	// Parse options if provided
	parsedOptions, err := parseOptions(*options)
	if err != nil {
		log.Fatalf("Failed to parse options: %v", err)
	}

	// Create custom field input
	input := CreateCustomFieldInput{
		Name:                   *name,
		Type:                   *fieldType,
		Description:            *description,
		ButtonType:             *buttonType,
		ButtonConfirmText:      *buttonConfirmText,
		CurrencyFieldID:        *currencyFieldID,
		ConversionDate:         *conversionDate,
		ConversionDateType:     *conversionDateType,
		Currency:               *currency,
		Prefix:                 *prefix,
		IsDueDate:              isDueDate,
		TimeDurationDisplay:    *timeDurationDisplay,
		ReferenceProjectID:     *referenceProjectID,
		ReferenceMultiple:      referenceMultiple,
		UseSequenceUniqueID:    useSequenceUniqueID,
		SequenceDigits:         sequenceDigits,
		SequenceStartingNumber: sequenceStartingNumber,
	}

	// Handle numeric fields - only set if non-default values
	if *min != 0 {
		input.Min = min
	}
	if *max != 0 {
		input.Max = max
	}
	if *timeDurationTargetTime != 0 {
		input.TimeDurationTargetTime = timeDurationTargetTime
	}

	// Execute creation
	fmt.Printf("Creating custom field '%s' of type '%s'...\n", input.Name, input.Type)
	
	customField, err := executeCreateCustomField(client, input)
	if err != nil {
		log.Fatalf("Failed to create custom field: %v", err)
	}

	// Display results
	fmt.Println("\n✅ Custom field created successfully!")
	fmt.Printf("\nCustom Field Details:\n")
	fmt.Printf("  ID:          %s\n", customField.ID)
	fmt.Printf("  Name:        %s\n", customField.Name)
	fmt.Printf("  Type:        %s\n", customField.Type)
	if customField.Description != "" {
		fmt.Printf("  Description: %s\n", customField.Description)
	}

	// Create options if provided and field type supports them
	if len(parsedOptions) > 0 && (*fieldType == "SELECT_SINGLE" || *fieldType == "SELECT_MULTI") {
		fmt.Printf("\nCreating %d options for the field...\n", len(parsedOptions))
		
		if err := createCustomFieldOptions(client, customField.ID, parsedOptions); err != nil {
			fmt.Printf("⚠️  Warning: Field created successfully but failed to create options: %v\n", err)
			fmt.Printf("You can manually add options later.\n")
		} else {
			fmt.Printf("✅ Options created successfully!\n")
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
		fmt.Printf("\n⚠️  Warning: Options provided but field type '%s' doesn't support options. Options were ignored.\n", *fieldType)
	}
	
	fmt.Printf("\nYou can now use this custom field in your todos and projects.\n")
}
