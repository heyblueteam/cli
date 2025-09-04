package tools

import (
	"flag"
	"fmt"
	"strconv"

	"demo-builder/common"
)

// UpdateCustomFieldInput for updating custom field properties
type UpdateCustomFieldInput struct {
	CustomFieldID           string   `json:"customFieldId"`
	Position                *float64 `json:"position,omitempty"`
	Name                    string   `json:"name,omitempty"`
	Description             string   `json:"description,omitempty"`
	Min                     *float64 `json:"min,omitempty"`
	Max                     *float64 `json:"max,omitempty"`
	Currency                string   `json:"currency,omitempty"`
	Prefix                  string   `json:"prefix,omitempty"`
	SequenceStartingNumber  *int     `json:"sequenceStartingNumber,omitempty"`
}

// UpdateCustomFieldResponse represents the response from the update mutation
type UpdateCustomFieldResponse struct {
	EditCustomField common.CustomField `json:"editCustomField"`
}

// RunUpdateCustomField executes the update custom field command
func RunUpdateCustomField(args []string) error {
	flagSet := flag.NewFlagSet("update-custom-field", flag.ExitOnError)
	var (
		customFieldID          = flagSet.String("field", "", "Custom field ID to edit (required)")
		projectID              = flagSet.String("project", "", "Project ID or slug (required for authorization)")
		name                   = flagSet.String("name", "", "New name for the custom field")
		description            = flagSet.String("description", "", "New description for the custom field")
		position               = flagSet.String("position", "", "New position for the custom field")
		minValue               = flagSet.String("min", "", "New minimum value (for NUMBER fields)")
		maxValue               = flagSet.String("max", "", "New maximum value (for NUMBER fields)")
		currency               = flagSet.String("currency", "", "New currency code (for CURRENCY fields)")
		prefix                 = flagSet.String("prefix", "", "New prefix (for TEXT fields)")
		sequenceStartingNumber = flagSet.String("sequence-start", "", "New sequence starting number (for SEQUENCE fields)")
		simple                 = flagSet.Bool("simple", false, "Simple output format")
	)
	flagSet.Parse(args)

	if *customFieldID == "" {
		return fmt.Errorf("-field parameter is required")
	}
	
	if *projectID == "" {
		return fmt.Errorf("-project parameter is required for authorization")
	}

	// Check if at least one field is being updated
	if *name == "" && *description == "" && *position == "" && *minValue == "" && 
	   *maxValue == "" && *currency == "" && *prefix == "" && *sequenceStartingNumber == "" {
		return fmt.Errorf("at least one field must be specified for update")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	client := common.NewClient(config)

	// Set project context for authorization
	client.SetProjectID(*projectID)

	// Build the input structure
	input := UpdateCustomFieldInput{
		CustomFieldID: *customFieldID,
	}

	if *name != "" {
		input.Name = *name
	}
	if *description != "" {
		input.Description = *description
	}
	if *currency != "" {
		input.Currency = *currency
	}
	if *prefix != "" {
		input.Prefix = *prefix
	}

	// Parse numeric values
	if *position != "" {
		pos, err := strconv.ParseFloat(*position, 64)
		if err != nil {
			return fmt.Errorf("invalid position value '%s'. Must be a number", *position)
		}
		input.Position = &pos
	}

	if *minValue != "" {
		min, err := strconv.ParseFloat(*minValue, 64)
		if err != nil {
			return fmt.Errorf("invalid min value '%s'. Must be a number", *minValue)
		}
		input.Min = &min
	}

	if *maxValue != "" {
		max, err := strconv.ParseFloat(*maxValue, 64)
		if err != nil {
			return fmt.Errorf("invalid max value '%s'. Must be a number", *maxValue)
		}
		input.Max = &max
	}

	if *sequenceStartingNumber != "" {
		seq, err := strconv.Atoi(*sequenceStartingNumber)
		if err != nil {
			return fmt.Errorf("invalid sequence starting number '%s'. Must be an integer", *sequenceStartingNumber)
		}
		input.SequenceStartingNumber = &seq
	}

	mutation := `
		mutation EditCustomField($input: EditCustomFieldInput!) {
			editCustomField(input: $input) {
				id
				uid
				name
				type
				description
				position
				min
				max
				currency
				prefix
				updatedAt
				customFieldOptions {
					id
					title
					color
				}
			}
		}
	`

	variables := map[string]interface{}{
		"input": input,
	}

	var result UpdateCustomFieldResponse
	err = client.ExecuteQueryWithResult(mutation, variables, &result)
	if err != nil {
		return fmt.Errorf("failed to execute query: %v", err)
	}

	// Output results
	field := result.EditCustomField
	if *simple {
		fmt.Printf("✅ Updated custom field %s\n", field.ID)
	} else {
		fmt.Printf("Updating custom field '%s'...\n\n", *customFieldID)
		fmt.Println("✅ Custom field updated successfully!")
		fmt.Println("\nUpdated Field Details:")
		fmt.Printf("  ID:          %s\n", field.ID)
		fmt.Printf("  UID:         %s\n", field.UID)
		fmt.Printf("  Name:        %s\n", field.Name)
		fmt.Printf("  Type:        %s\n", field.Type)
		if field.Description != "" {
			fmt.Printf("  Description: %s\n", field.Description)
		}
		fmt.Printf("  Position:    %.0f\n", field.Position)
		
		// Show type-specific properties
		if field.Min != nil {
			fmt.Printf("  Min Value:   %.2f\n", *field.Min)
		}
		if field.Max != nil {
			fmt.Printf("  Max Value:   %.2f\n", *field.Max)
		}
		if field.Currency != "" {
			fmt.Printf("  Currency:    %s\n", field.Currency)
		}
		if field.Prefix != "" {
			fmt.Printf("  Prefix:      %s\n", field.Prefix)
		}
		
		fmt.Printf("  Updated:     %s\n", field.UpdatedAt)

		// Show options if available
		if len(field.Options) > 0 {
			fmt.Printf("\nCurrent Options (%d):\n", len(field.Options))
			for _, option := range field.Options {
				if option.Color != "" {
					fmt.Printf("  - %s (color: %s) [ID: %s]\n", option.Title, option.Color, option.ID)
				} else {
					fmt.Printf("  - %s [ID: %s]\n", option.Title, option.ID)
				}
			}
		}

		fmt.Printf("\nThe custom field can now be used in records with its updated properties.\n")
	}

	return nil
}