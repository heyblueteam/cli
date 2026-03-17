package fields

import (
	"fmt"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

// DeleteCustomFieldResponse represents the response from the delete mutation
type DeleteCustomFieldResponse struct {
	DeleteCustomField bool `json:"deleteCustomField"`
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a custom field",
	Long:  "Permanently delete a custom field and remove it from all records. This action cannot be undone.",
	Example: `  blue fields delete --workspace <id> --field <field-id> --confirm
  blue fields delete -w <id> --field <field-id> -y --simple`,
	RunE: runDelete,
}

var (
	deleteWorkspace string
	deleteField     string
	deleteConfirm   bool
	deleteSimple    bool
)

func init() {
	deleteCmd.Flags().StringVarP(&deleteWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	deleteCmd.Flags().StringVar(&deleteField, "field", "", "Custom field ID to delete (required)")
	deleteCmd.Flags().BoolVarP(&deleteConfirm, "confirm", "y", false, "Confirm deletion (required for safety)")
	deleteCmd.Flags().BoolVarP(&deleteSimple, "simple", "s", false, "Simple output format")
}

func runDelete(cmd *cobra.Command, args []string) error {
	if deleteField == "" {
		return fmt.Errorf("--field parameter is required")
	}

	if deleteWorkspace == "" {
		return fmt.Errorf("--workspace parameter is required for authorization")
	}

	if !deleteConfirm {
		return fmt.Errorf("--confirm flag is required for safety. This operation will permanently delete the custom field and all its data")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := common.NewClient(config)
	client.SetProject(deleteWorkspace)

	// Fetch custom field details for confirmation display
	if !deleteSimple {
		fmt.Printf("Fetching custom field details for %s...\n", deleteField)
		field, err := getCustomFieldDetails(client, deleteField)
		if err != nil {
			return fmt.Errorf("failed to fetch custom field details: %w", err)
		}

		fmt.Printf("\nAbout to delete custom field:\n")
		fmt.Printf("  ID:   %s\n", field.ID)
		fmt.Printf("  Name: %s\n", field.Name)
		fmt.Printf("  Type: %s\n", field.Type)
		if field.Description != "" {
			fmt.Printf("  Description: %s\n", field.Description)
		}
		if len(field.Options) > 0 {
			fmt.Printf("  Options: %d\n", len(field.Options))
		}
		fmt.Printf("\nThis will permanently delete this custom field and remove it from all records!\n\n")
	}

	mutation := `
		mutation DeleteCustomField($id: String!) {
			deleteCustomField(id: $id)
		}
	`

	variables := map[string]interface{}{
		"id": deleteField,
	}

	var result DeleteCustomFieldResponse
	if err := client.ExecuteQueryWithResult(mutation, variables, &result); err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	if result.DeleteCustomField {
		if deleteSimple {
			fmt.Printf("Deleted custom field %s\n", deleteField)
		} else {
			fmt.Printf("Custom field deleted successfully!\n")
			fmt.Printf("Custom field %s has been permanently removed.\n", deleteField)
			fmt.Printf("All record data associated with this field has been cleared.\n")
		}
	} else {
		if deleteSimple {
			fmt.Printf("Failed to delete custom field %s\n", deleteField)
		} else {
			fmt.Printf("Custom field was not deleted.\n")
			fmt.Printf("This may indicate that the field doesn't exist or cannot be deleted.\n")
		}
		return fmt.Errorf("custom field was not deleted")
	}

	return nil
}

// getCustomFieldDetails fetches custom field details for confirmation display
func getCustomFieldDetails(client *common.Client, customFieldID string) (*common.CustomField, error) {
	query := `
		query GetCustomField($customFieldId: String!) {
			customField(id: $customFieldId) {
				id
				uid
				name
				type
				description
				position
				customFieldOptions {
					id
					title
					color
				}
			}
		}
	`

	variables := map[string]interface{}{
		"customFieldId": customFieldID,
	}

	type CustomFieldResponse struct {
		CustomField common.CustomField `json:"customField"`
	}

	var result CustomFieldResponse
	if err := client.ExecuteQueryWithResult(query, variables, &result); err != nil {
		return nil, err
	}

	return &result.CustomField, nil
}
