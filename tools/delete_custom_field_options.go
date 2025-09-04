package tools

import (
	"flag"
	"fmt"
	"strings"

	"demo-builder/common"
)

// DeleteCustomFieldOptionResponse represents the response from the delete mutation
type DeleteCustomFieldOptionResponse struct {
	DeleteCustomFieldOption bool `json:"deleteCustomFieldOption"`
}

// RunDeleteCustomFieldOptions executes the delete custom field options command
func RunDeleteCustomFieldOptions(args []string) error {
	flagSet := flag.NewFlagSet("delete-custom-field-options", flag.ExitOnError)
	var (
		customFieldID = flagSet.String("field", "", "Custom field ID containing the options (required)")
		projectID     = flagSet.String("project", "", "Project ID or slug (optional - improves authorization)")
		optionIDs     = flagSet.String("option-ids", "", "Comma-separated list of option IDs to delete")
		optionTitles  = flagSet.String("option-titles", "", "Comma-separated list of option titles to delete")
		todoID        = flagSet.String("todo", "", "Todo ID (optional - used for option dependency tracking)")
		confirm       = flagSet.Bool("confirm", false, "Confirm deletion (required for safety)")
		simple        = flagSet.Bool("simple", false, "Simple output format")
	)
	flagSet.Parse(args)

	if *customFieldID == "" {
		return fmt.Errorf("-field parameter is required")
	}

	if *optionIDs == "" && *optionTitles == "" {
		return fmt.Errorf("either -option-ids or -option-titles parameter is required")
	}

	if !*confirm {
		return fmt.Errorf("-confirm flag is required for safety")
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

	var optionsToDelete []string

	// If option titles are provided, resolve them to IDs first
	if *optionTitles != "" {
		resolvedIDs, err := resolveOptionTitlesToIDs(client, *customFieldID, *optionTitles)
		if err != nil {
			return fmt.Errorf("resolving option titles: %v", err)
		}
		optionsToDelete = resolvedIDs
	} else {
		// Use provided option IDs directly
		optionsToDelete = strings.Split(*optionIDs, ",")
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

	if !*simple {
		fmt.Printf("Deleting %d option(s) from custom field %s...\n\n", len(optionsToDelete), *customFieldID)
	}

	for _, optionID := range optionsToDelete {
		if optionID == "" {
			continue
		}

		variables := map[string]interface{}{
			"customFieldId": *customFieldID,
			"optionId":      optionID,
		}
		if *todoID != "" {
			variables["todoId"] = *todoID
		}

		var result DeleteCustomFieldOptionResponse
		err := client.ExecuteQueryWithResult(mutation, variables, &result)
		if err != nil {
			errors = append(errors, fmt.Sprintf("Failed to delete option %s: %v", optionID, err))
			if !*simple {
				fmt.Printf("❌ Failed to delete option %s: %v\n", optionID, err)
			}
		} else if result.DeleteCustomFieldOption {
			deletedCount++
			if !*simple {
				fmt.Printf("✅ Deleted option %s\n", optionID)
			}
		} else {
			errors = append(errors, fmt.Sprintf("Option %s was not deleted (may not exist or be in use)", optionID))
			if !*simple {
				fmt.Printf("⚠️  Option %s was not deleted (may not exist or be in use)\n", optionID)
			}
		}
	}

	// Summary output
	if *simple {
		if len(errors) == 0 {
			fmt.Printf("✅ Deleted %d options from custom field %s\n", deletedCount, *customFieldID)
		} else {
			fmt.Printf("⚠️  Deleted %d options, %d errors occurred\n", deletedCount, len(errors))
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
	err := client.ExecuteQueryWithResult(query, variables, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch custom field: %v", err)
	}

	// Parse requested titles
	requestedTitles := strings.Split(titlesStr, ",")
	for i := range requestedTitles {
		requestedTitles[i] = strings.TrimSpace(requestedTitles[i])
	}

	// Map titles to IDs
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