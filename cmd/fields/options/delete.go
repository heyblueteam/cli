package options

import (
	"fmt"
	"strings"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

// DeleteCustomFieldOptionResponse represents the response from the delete mutation
type DeleteCustomFieldOptionResponse struct {
	DeleteCustomFieldOption bool `json:"deleteCustomFieldOption"`
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete options from a custom field",
	Long:  "Delete one or more options from a select-type custom field. Options can be specified by ID or title.",
	Example: `  blue fields options delete --field <field-id> --option-ids "id1,id2" --confirm
  blue fields options delete --field <field-id> --option-titles "High,Low" --confirm
  blue fields options delete --field <field-id> --option-ids "id1" -y --simple`,
	RunE: runDelete,
}

var (
	deleteField        string
	deleteWorkspace    string
	deleteOptionIDs    string
	deleteOptionTitles string
	deleteTodoID       string
	deleteConfirm      bool
	deleteSimple       bool
)

func init() {
	deleteCmd.Flags().StringVar(&deleteField, "field", "", "Custom field ID containing the options (required)")
	deleteCmd.Flags().StringVarP(&deleteWorkspace, "workspace", "w", "", "Workspace ID or slug (optional - improves authorization)")
	deleteCmd.Flags().StringVar(&deleteOptionIDs, "option-ids", "", "Comma-separated list of option IDs to delete")
	deleteCmd.Flags().StringVar(&deleteOptionTitles, "option-titles", "", "Comma-separated list of option titles to delete")
	deleteCmd.Flags().StringVar(&deleteTodoID, "todo", "", "Todo ID (optional - used for option dependency tracking)")
	deleteCmd.Flags().BoolVarP(&deleteConfirm, "confirm", "y", false, "Confirm deletion (required for safety)")
	deleteCmd.Flags().BoolVarP(&deleteSimple, "simple", "s", false, "Simple output format")
}

func runDelete(cmd *cobra.Command, args []string) error {
	if deleteField == "" {
		return fmt.Errorf("--field parameter is required")
	}

	if deleteOptionIDs == "" && deleteOptionTitles == "" {
		return fmt.Errorf("either --option-ids or --option-titles parameter is required")
	}

	if !deleteConfirm {
		return fmt.Errorf("--confirm flag is required for safety")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := common.NewClient(config)

	if deleteWorkspace != "" {
		client.SetProject(deleteWorkspace)
	}

	var optionsToDelete []string

	// If option titles are provided, resolve them to IDs first
	if deleteOptionTitles != "" {
		resolvedIDs, err := resolveOptionTitlesToIDs(client, deleteField, deleteOptionTitles)
		if err != nil {
			return fmt.Errorf("resolving option titles: %w", err)
		}
		optionsToDelete = resolvedIDs
	} else {
		optionsToDelete = strings.Split(deleteOptionIDs, ",")
		for i := range optionsToDelete {
			optionsToDelete[i] = strings.TrimSpace(optionsToDelete[i])
		}
	}

	if len(optionsToDelete) == 0 {
		return fmt.Errorf("no valid options found to delete")
	}

	mutation := `
		mutation DeleteCustomFieldOption($customFieldId: String!, $optionId: String!, $todoId: String) {
			deleteCustomFieldOption(customFieldId: $customFieldId, optionId: $optionId, todoId: $todoId)
		}
	`

	var deletedCount int
	var errors []string

	if !deleteSimple {
		fmt.Printf("Deleting %d option(s) from custom field %s...\n\n", len(optionsToDelete), deleteField)
	}

	for _, optionID := range optionsToDelete {
		if optionID == "" {
			continue
		}

		variables := map[string]interface{}{
			"customFieldId": deleteField,
			"optionId":      optionID,
		}
		if deleteTodoID != "" {
			variables["todoId"] = deleteTodoID
		}

		var result DeleteCustomFieldOptionResponse
		err := client.ExecuteQueryWithResult(mutation, variables, &result)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to delete option %s: %v", optionID, err))
			if !deleteSimple {
				fmt.Printf("Failed to delete option %s: %v\n", optionID, err)
			}
		} else if result.DeleteCustomFieldOption {
			deletedCount++
			if !deleteSimple {
				fmt.Printf("Deleted option %s\n", optionID)
			}
		} else {
			errors = append(errors, fmt.Sprintf("Option %s was not deleted (may not exist or be in use)", optionID))
			if !deleteSimple {
				fmt.Printf("Option %s was not deleted (may not exist or be in use)\n", optionID)
			}
		}
	}

	// Summary output
	if deleteSimple {
		if len(errors) == 0 {
			fmt.Printf("Deleted %d options from custom field %s\n", deletedCount, deleteField)
		} else {
			fmt.Printf("Deleted %d options, %d errors occurred\n", deletedCount, len(errors))
		}
	} else {
		fmt.Printf("\n=== Summary ===\n")
		fmt.Printf("Deleted: %d options\n", deletedCount)
		if len(errors) > 0 {
			fmt.Printf("Errors: %d\n", len(errors))
			fmt.Println("\nError details:")
			for _, err := range errors {
				fmt.Printf("  - %s\n", err)
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("some deletions failed")
	}

	return nil
}

// resolveOptionTitlesToIDs fetches the custom field and resolves option titles to their IDs
func resolveOptionTitlesToIDs(client *common.Client, customFieldID, titlesStr string) ([]string, error) {
	query := `
		query GetCustomField($customFieldId: String!) {
			customField(id: $customFieldId) {
				id
				name
				customFieldOptions {
					id
					title
				}
			}
		}
	`

	variables := map[string]interface{}{
		"customFieldId": customFieldID,
	}

	type CustomFieldResponse struct {
		CustomField struct {
			ID      string `json:"id"`
			Name    string `json:"name"`
			Options []struct {
				ID    string `json:"id"`
				Title string `json:"title"`
			} `json:"customFieldOptions"`
		} `json:"customField"`
	}

	var result CustomFieldResponse
	if err := client.ExecuteQueryWithResult(query, variables, &result); err != nil {
		return nil, fmt.Errorf("failed to fetch custom field: %w", err)
	}

	requestedTitles := strings.Split(titlesStr, ",")
	for i := range requestedTitles {
		requestedTitles[i] = strings.TrimSpace(requestedTitles[i])
	}

	var resolvedIDs []string
	var notFound []string

	for _, requestedTitle := range requestedTitles {
		if requestedTitle == "" {
			continue
		}

		found := false
		for _, option := range result.CustomField.Options {
			if option.Title == requestedTitle {
				resolvedIDs = append(resolvedIDs, option.ID)
				found = true
				break
			}
		}

		if !found {
			notFound = append(notFound, requestedTitle)
		}
	}

	if len(notFound) > 0 {
		return nil, fmt.Errorf("option titles not found: %v", notFound)
	}

	return resolvedIDs, nil
}
