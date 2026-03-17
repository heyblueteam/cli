package fields

import (
	"fmt"
	"strconv"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

// UpdateCustomFieldInput for updating custom field properties
type UpdateCustomFieldInput struct {
	CustomFieldID          string   `json:"customFieldId"`
	Position               *float64 `json:"position,omitempty"`
	Name                   string   `json:"name,omitempty"`
	Description            string   `json:"description,omitempty"`
	Min                    *float64 `json:"min,omitempty"`
	Max                    *float64 `json:"max,omitempty"`
	Currency               string   `json:"currency,omitempty"`
	Prefix                 string   `json:"prefix,omitempty"`
	SequenceStartingNumber *int     `json:"sequenceStartingNumber,omitempty"`
}

// UpdateCustomFieldResponse represents the response from the update mutation
type UpdateCustomFieldResponse struct {
	EditCustomField common.CustomField `json:"editCustomField"`
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a custom field",
	Long:  "Update properties of an existing custom field.",
	Example: `  blue fields update --workspace <id> --field <field-id> --name "New Name"
  blue fields update -w <id> --field <field-id> --description "Updated description"
  blue fields update -w <id> --field <field-id> --min 0 --max 100
  blue fields update -w <id> --field <field-id> --currency EUR --simple`,
	RunE: runUpdate,
}

var (
	updateWorkspace            string
	updateField                string
	updateName                 string
	updateDescription          string
	updatePosition             string
	updateMin                  string
	updateMax                  string
	updateCurrency             string
	updatePrefix               string
	updateSequenceStart        string
	updateSimple               bool
)

func init() {
	updateCmd.Flags().StringVarP(&updateWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	updateCmd.Flags().StringVar(&updateField, "field", "", "Custom field ID to update (required)")
	updateCmd.Flags().StringVar(&updateName, "name", "", "New name for the custom field")
	updateCmd.Flags().StringVar(&updateDescription, "description", "", "New description for the custom field")
	updateCmd.Flags().StringVar(&updatePosition, "position", "", "New position for the custom field")
	updateCmd.Flags().StringVar(&updateMin, "min", "", "New minimum value (for NUMBER fields)")
	updateCmd.Flags().StringVar(&updateMax, "max", "", "New maximum value (for NUMBER fields)")
	updateCmd.Flags().StringVar(&updateCurrency, "currency", "", "New currency code (for CURRENCY fields)")
	updateCmd.Flags().StringVar(&updatePrefix, "prefix", "", "New prefix (for TEXT fields)")
	updateCmd.Flags().StringVar(&updateSequenceStart, "sequence-start", "", "New sequence starting number (for SEQUENCE fields)")
	updateCmd.Flags().BoolVarP(&updateSimple, "simple", "s", false, "Simple output format")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if updateField == "" {
		return fmt.Errorf("--field parameter is required")
	}

	if updateWorkspace == "" {
		return fmt.Errorf("--workspace parameter is required for authorization")
	}

	// Check if at least one field is being updated
	if updateName == "" && updateDescription == "" && updatePosition == "" && updateMin == "" &&
		updateMax == "" && updateCurrency == "" && updatePrefix == "" && updateSequenceStart == "" {
		return fmt.Errorf("at least one field must be specified for update")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := common.NewClient(config)
	client.SetProject(updateWorkspace)

	input := UpdateCustomFieldInput{
		CustomFieldID: updateField,
	}

	if updateName != "" {
		input.Name = updateName
	}
	if updateDescription != "" {
		input.Description = updateDescription
	}
	if updateCurrency != "" {
		input.Currency = updateCurrency
	}
	if updatePrefix != "" {
		input.Prefix = updatePrefix
	}

	if updatePosition != "" {
		pos, err := strconv.ParseFloat(updatePosition, 64)
		if err != nil {
			return fmt.Errorf("invalid position value '%s'. Must be a number", updatePosition)
		}
		input.Position = &pos
	}

	if updateMin != "" {
		min, err := strconv.ParseFloat(updateMin, 64)
		if err != nil {
			return fmt.Errorf("invalid min value '%s'. Must be a number", updateMin)
		}
		input.Min = &min
	}

	if updateMax != "" {
		max, err := strconv.ParseFloat(updateMax, 64)
		if err != nil {
			return fmt.Errorf("invalid max value '%s'. Must be a number", updateMax)
		}
		input.Max = &max
	}

	if updateSequenceStart != "" {
		seq, err := strconv.Atoi(updateSequenceStart)
		if err != nil {
			return fmt.Errorf("invalid sequence starting number '%s'. Must be an integer", updateSequenceStart)
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
	if err := client.ExecuteQueryWithResult(mutation, variables, &result); err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	field := result.EditCustomField
	if updateSimple {
		fmt.Printf("Updated custom field %s\n", field.ID)
	} else {
		fmt.Printf("Updating custom field '%s'...\n\n", updateField)
		fmt.Println("Custom field updated successfully!")
		fmt.Println("\nUpdated Field Details:")
		fmt.Printf("  ID:          %s\n", field.ID)
		fmt.Printf("  UID:         %s\n", field.UID)
		fmt.Printf("  Name:        %s\n", field.Name)
		fmt.Printf("  Type:        %s\n", field.Type)
		if field.Description != "" {
			fmt.Printf("  Description: %s\n", field.Description)
		}
		fmt.Printf("  Position:    %.0f\n", field.Position)

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
