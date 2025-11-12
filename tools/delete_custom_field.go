package tools

import (
	"flag"
	"fmt"

	"cli/common"
)

// DeleteCustomFieldResponse represents the response from the delete mutation
type DeleteCustomFieldResponse struct {
	DeleteCustomField bool `json:"deleteCustomField"`
}

// RunDeleteCustomField executes the delete custom field command
func RunDeleteCustomField(args []string) error {
	flagSet := flag.NewFlagSet("delete-custom-field", flag.ExitOnError)
	var (
		customFieldID = flagSet.String("field", "", "Custom field ID to delete (required)")
		projectID     = flagSet.String("project", "", "Project ID or slug (required for authorization)")
		confirm       = flagSet.Bool("confirm", false, "Confirm deletion (required for safety)")
		simple        = flagSet.Bool("simple", false, "Simple output format")
	)
	flagSet.Parse(args)

	if *customFieldID == "" {
		return fmt.Errorf("-field parameter is required")
	}

	if *projectID == "" {
		return fmt.Errorf("-project parameter is required for authorization")
	}

	if !*confirm {
		return fmt.Errorf("-confirm flag is required for safety. This operation will permanently delete the custom field and all its data")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	client := common.NewClient(config)

	// Set project context for authorization
	client.SetProjectID(*projectID)

	// First, get the custom field details for confirmation
	if !*simple {
		fmt.Printf("Fetching custom field details for %s...\n", *customFieldID)
		field, err := getCustomFieldDetails(client, *customFieldID)
		if err != nil {
			return fmt.Errorf("failed to fetch custom field details: %v", err)
		}
		
		fmt.Printf("\n⚠️  About to delete custom field:\n")
		fmt.Printf("  ID:   %s\n", field.ID)
		fmt.Printf("  Name: %s\n", field.Name)
		fmt.Printf("  Type: %s\n", field.Type)
		if field.Description != "" {
			fmt.Printf("  Description: %s\n", field.Description)
		}
		if len(field.Options) > 0 {
			fmt.Printf("  Options: %d\n", len(field.Options))
		}
		fmt.Printf("\n🚨 This will permanently delete this custom field and remove it from all records!\n\n")
	}

	mutation := `
		mutation DeleteCustomField($id: String!) {
			deleteCustomField(id: $id)
		}
	`

	variables := map[string]interface{}{
		"id": *customFieldID,
	}

	var result DeleteCustomFieldResponse
	err = client.ExecuteQueryWithResult(mutation, variables, &result)
	if err != nil {
		return fmt.Errorf("failed to execute query: %v", err)
	}

	// Output results
	if result.DeleteCustomField {
		if *simple {
			fmt.Printf("✅ Deleted custom field %s\n", *customFieldID)
		} else {
			fmt.Printf("✅ Custom field deleted successfully!\n")
			fmt.Printf("Custom field %s has been permanently removed.\n", *customFieldID)
			fmt.Printf("All record data associated with this field has been cleared.\n")
		}
	} else {
		if *simple {
			fmt.Printf("❌ Failed to delete custom field %s\n", *customFieldID)
		} else {
			fmt.Printf("❌ Custom field was not deleted.\n")
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
	err := client.ExecuteQueryWithResult(query, variables, &result)
	if err != nil {
		return nil, err
	}

	return &result.CustomField, nil
}