package tools

import (
	"flag"
	"fmt"
	"strings"
	
	. "cli/common"
)

// ReadCustomFieldsResponse represents the response from the GraphQL query
type ReadCustomFieldsResponse struct {
	CustomFields struct {
		Items    []CustomField  `json:"items"`
		PageInfo OffsetPageInfo `json:"pageInfo"`
	} `json:"customFields"`
}

// Build enhanced query to fetch custom fields with all details needed for record operations
func buildReadCustomFieldsQuery(projectID string, skip int, take int) string {
	query := fmt.Sprintf(`query {
		customFields(
			filter: { projectId: "%s" }
			skip: %d
			take: %d
		) {
			items {
				id
				name
				type
				position
				description
				min
				max
				currency
				prefix
				createdAt
				updatedAt
				customFieldOptions {
					id
					title
					color
				}
			}
			pageInfo {
				totalPages
				totalItems
				page
				perPage
				hasNextPage
				hasPreviousPage
			}
		}
	}`, projectID, skip, take)

	return query
}

// getFieldTypeDescription returns a user-friendly description of what values a field accepts
func getFieldTypeDescription(field CustomField) string {
	switch field.Type {
	case "TEXT_SINGLE":
		return "Single line of text"
	case "TEXT_MULTI":
		return "Multi-line text"
	case "NUMBER":
		desc := "Numeric value"
		if field.Min != nil && field.Max != nil {
			desc += fmt.Sprintf(" (range: %.0f - %.0f)", *field.Min, *field.Max)
		} else if field.Min != nil {
			desc += fmt.Sprintf(" (min: %.0f)", *field.Min)
		} else if field.Max != nil {
			desc += fmt.Sprintf(" (max: %.0f)", *field.Max)
		}
		return desc
	case "CURRENCY":
		currency := "USD"
		if field.Currency != "" {
			currency = field.Currency
		}
		return fmt.Sprintf("Currency amount (%s)", currency)
	case "PERCENT":
		return "Percentage value (0-100)"
	case "EMAIL":
		return "Email address"
	case "PHONE":
		return "Phone number"
	case "URL":
		return "Web URL"
	case "CHECKBOX":
		return "Boolean value (true/false)"
	case "SELECT_SINGLE":
		return "Single selection from options"
	case "SELECT_MULTI":
		return "Multiple selections from options"
	case "RATING":
		return "Rating value (typically 1-5)"
	case "DATE":
		return "Date (YYYY-MM-DD)"
	case "DATETIME":
		return "Date and time (ISO 8601)"
	case "FILE":
		return "File attachment"
	default:
		return field.Type
	}
}

// generateExampleValue creates an example value for the field type
func generateExampleValue(field CustomField) string {
	switch field.Type {
	case "TEXT_SINGLE", "TEXT_MULTI":
		return "\"Sample text\""
	case "NUMBER":
		if field.Min != nil {
			return fmt.Sprintf("%.0f", *field.Min + 1)
		}
		return "42"
	case "CURRENCY":
		return "1000.50"
	case "PERCENT":
		return "75.5"
	case "EMAIL":
		return "\"user@example.com\""
	case "PHONE":
		return "\"+1-555-123-4567\""
	case "URL":
		return "\"https://example.com\""
	case "CHECKBOX":
		return "true"
	case "SELECT_SINGLE":
		if len(field.Options) > 0 {
			return "\"" + field.Options[0].Title + "\""
		}
		return "\"option_value\""
	case "SELECT_MULTI":
		if len(field.Options) >= 2 {
			return "\"" + field.Options[0].Title + "," + field.Options[1].Title + "\""
		} else if len(field.Options) == 1 {
			return "\"" + field.Options[0].Title + "\""
		}
		return "\"option1,option2\""
	case "RATING":
		return "4"
	case "DATE":
		return "\"2025-12-31\""
	case "DATETIME":
		return "\"2025-12-31T14:30:00Z\""
	default:
		return "\"value\""
	}
}

func RunReadCustomFields(args []string) error {
	fs := flag.NewFlagSet("read-custom-fields", flag.ExitOnError)
	projectID := fs.String("project", "", "Project ID or slug (required)")
	page := fs.Int("page", 1, "Page number (default: 1)")
	pageSize := fs.Int("size", 50, "Page size (default: 50)")
	simple := fs.Bool("simple", false, "Show only essential information for record creation")
	examples := fs.Bool("examples", false, "Show example usage for create-record and update-record commands")
	format := fs.String("format", "table", "Output format: table, json, csv")
	fs.Parse(args)

	// Validate required parameters
	if *projectID == "" {
		return fmt.Errorf("project ID or slug is required. Use -project flag")
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Create client and set project context
	client := NewClient(config)
	client.SetProject(*projectID)

	// Calculate skip value from page
	skip := (*page - 1) * *pageSize
	take := *pageSize

	// Build and execute query
	query := buildReadCustomFieldsQuery(*projectID, skip, take)

	// Execute query
	var response ReadCustomFieldsResponse
	if err := client.ExecuteQueryWithResult(query, nil, &response); err != nil {
		return fmt.Errorf("failed to execute query: %v", err)
	}

	// Get custom fields list
	customFields := response.CustomFields

	if len(customFields.Items) == 0 {
		fmt.Printf("No custom fields found in project %s\n", *projectID)
		return nil
	}

	// Display based on format
	switch *format {
	case "json":
		return displayJSON(customFields.Items)
	case "csv":
		return displayCSV(customFields.Items)
	default:
		return displayTable(customFields.Items, *projectID, *page, *pageSize, *simple, *examples)
	}
}

func displayTable(fields []CustomField, projectID string, page, pageSize int, simple, examples bool) error {
	// Display header
	fmt.Printf("\n=== Custom Fields Reference for Project %s ===\n", projectID)
	fmt.Printf("📋 Found %d custom fields - Use these Field IDs for create-record and update-record commands\n\n", len(fields))

	if simple {
		// Simple format - just ID, name, type for quick reference
		fmt.Printf("%-32s | %-25s | %-15s\n", "Field ID", "Field Name", "Type")
		fmt.Printf("%-32s-+-%-25s-+-%-15s\n", strings.Repeat("-", 32), strings.Repeat("-", 25), strings.Repeat("-", 15))
		
		for _, field := range fields {
			fmt.Printf("%-32s | %-25s | %-15s\n", field.ID, field.Name, field.Type)
		}
	} else {
		// Detailed format with examples and descriptions
		for i, field := range fields {
			fmt.Printf("%d. %s\n", i+1, field.Name)
			fmt.Printf("   🔑 Field ID: %s\n", field.ID)
			fmt.Printf("   📝 Type: %s\n", field.Type)
			fmt.Printf("   💡 Description: %s\n", getFieldTypeDescription(field))
			
			if field.Description != "" {
				fmt.Printf("   📖 Notes: %s\n", field.Description)
			}
			
			// Show options for SELECT fields
			if (field.Type == "SELECT_SINGLE" || field.Type == "SELECT_MULTI") && len(field.Options) > 0 {
				fmt.Printf("   🎯 Available Options:\n")
				for _, option := range field.Options {
					colorInfo := ""
					if option.Color != "" {
						colorInfo = fmt.Sprintf(" (%s)", option.Color)
					}
					fmt.Printf("      • %s%s\n", option.Title, colorInfo)
				}
			}
			
			// Show example value
			fmt.Printf("   📋 Example Value: %s\n", generateExampleValue(field))
			
			// Show command usage example
			fmt.Printf("   ⚡ Usage in Commands:\n")
			fmt.Printf("      create-record: -custom-fields \"%s:%s\"\n", field.ID, generateExampleValue(field))
			fmt.Printf("      update-record: -custom-fields \"%s:%s\"\n", field.ID, generateExampleValue(field))
			
			fmt.Println()
		}
	}

	if examples {
		fmt.Println("\n=== 📚 Command Examples ===")
		fmt.Println("Create a record with custom field values:")
		
		// Generate a comprehensive example using multiple fields
		var exampleFields []string
		for i, field := range fields {
			if i >= 3 { // Limit to first 3 fields for readability
				break
			}
			exampleFields = append(exampleFields, fmt.Sprintf("%s:%s", field.ID, generateExampleValue(field)))
		}
		
		if len(exampleFields) > 0 {
			fmt.Printf("go run . create-record -list LIST_ID -title \"Sample Record\" -custom-fields \"%s\"\n\n", 
				strings.Join(exampleFields, ";"))
		}
		
		fmt.Println("Update a record's custom fields:")
		if len(exampleFields) > 0 {
			fmt.Printf("go run . update-record -record RECORD_ID -custom-fields \"%s\"\n\n", 
				strings.Join(exampleFields, ";"))
		}
		
		fmt.Println("Query records by custom field values (client-side filtering):")
		if len(fields) > 0 {
			field := fields[0]
			example := "value"
			operator := "EQ"
			
			switch field.Type {
			case "NUMBER", "CURRENCY", "RATING", "PERCENT":
				example = "1000"
				operator = "GT"
			case "TEXT_SINGLE", "TEXT_MULTI":
				example = "search term"
				operator = "CONTAINS"
			case "SELECT_SINGLE", "SELECT_MULTI":
				if len(field.Options) > 0 {
					example = field.Options[0].Title
				}
				operator = "EQ"
			case "CHECKBOX":
				example = "true"
				operator = "EQ"
			}
			
			fmt.Printf("go run . read-records -project %s -custom-field \"%s:%s:%s\" -simple\n", 
				projectID, field.ID, operator, example)
		}
	}

	return nil
}

func displayJSON(fields []CustomField) error {
	fmt.Println("[")
	for i, field := range fields {
		fmt.Printf("  {\n")
		fmt.Printf("    \"id\": \"%s\",\n", field.ID)
		fmt.Printf("    \"name\": \"%s\",\n", field.Name)
		fmt.Printf("    \"type\": \"%s\",\n", field.Type)
		fmt.Printf("    \"description\": \"%s\",\n", field.Description)
		
		if len(field.Options) > 0 {
			fmt.Printf("    \"options\": [\n")
			for j, option := range field.Options {
				fmt.Printf("      {\"title\": \"%s\", \"color\": \"%s\"}", option.Title, option.Color)
				if j < len(field.Options)-1 {
					fmt.Printf(",")
				}
				fmt.Printf("\n")
			}
			fmt.Printf("    ],\n")
		}
		
		fmt.Printf("    \"exampleValue\": %s\n", generateExampleValue(field))
		fmt.Printf("  }")
		if i < len(fields)-1 {
			fmt.Printf(",")
		}
		fmt.Printf("\n")
	}
	fmt.Println("]")
	return nil
}

func displayCSV(fields []CustomField) error {
	fmt.Println("field_id,name,type,description,example_value,options")
	for _, field := range fields {
		options := ""
		if len(field.Options) > 0 {
			var optionStrings []string
			for _, option := range field.Options {
				optionStrings = append(optionStrings, fmt.Sprintf("%s:%s", option.Title, option.Color))
			}
			options = strings.Join(optionStrings, "|")
		}
		
		fmt.Printf("%s,%s,%s,\"%s\",%s,\"%s\"\n", 
			field.ID, field.Name, field.Type, field.Description, generateExampleValue(field), options)
	}
	return nil
}