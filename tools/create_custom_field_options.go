package tools

import (
	"flag"
	"fmt"
	"strings"

	"demo-builder/common"
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

// RunCreateCustomFieldOptions executes the create custom field options command
func RunCreateCustomFieldOptions(args []string) error {
	flagSet := flag.NewFlagSet("create-custom-field-options", flag.ExitOnError)
	var (
		customFieldID = flagSet.String("field", "", "Custom field ID to add options to (required)")
		projectID     = flagSet.String("project", "", "Project ID or slug (optional - improves authorization)")
		options       = flagSet.String("options", "", "Options in format 'Title1:color1,Title2:color2' (required)")
		simple        = flagSet.Bool("simple", false, "Simple output format")
	)
	flagSet.Parse(args)

	if *customFieldID == "" {
		return fmt.Errorf("-field parameter is required. Usage: go run . create-custom-field-options -field FIELD_ID -options 'Option1:red,Option2:blue'")
	}

	if *options == "" {
		return fmt.Errorf("-options parameter is required. Usage: go run . create-custom-field-options -field FIELD_ID -options 'Option1:red,Option2:blue'")
	}

	// Parse options string into CustomFieldOptionInput array
	optionInputs, err := parseOptionsFromString(*options)
	if err != nil {
		return fmt.Errorf("parsing options: %v", err)
	}

	if len(optionInputs) == 0 {
		return fmt.Errorf("no valid options provided")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	client := common.NewClient(config)

	// Set project context if provided (improves authorization)
	if *projectID != "" {
		client.SetProjectID(*projectID)
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
			CustomFieldID:      *customFieldID,
			CustomFieldOptions: optionInputs,
		},
	}

	var result AddCustomFieldOptionsResponse
	err = client.ExecuteQueryWithResult(mutation, variables, &result)
	if err != nil {
		return fmt.Errorf("failed to execute query: %v", err)
	}

	// Output results
	if *simple {
		fmt.Printf("✅ Added %d options to custom field %s\n", len(result.CreateCustomFieldOptions), *customFieldID)
	} else {
		fmt.Printf("Adding %d options to custom field '%s'...\n\n", len(optionInputs), *customFieldID)
		fmt.Println("✅ Options added successfully!")
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

		// Split by colon to separate title and color
		titleColor := strings.SplitN(part, ":", 2)
		title := strings.TrimSpace(titleColor[0])
		
		if title == "" {
			continue
		}

		option := common.CustomFieldOptionInput{
			Title: title,
		}

		// Add color if provided
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