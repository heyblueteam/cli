package fields

import (
	"fmt"
	"strings"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

// CustomFieldPagination represents a paginated list of custom fields
type CustomFieldPagination struct {
	Items    []common.CustomField `json:"items"`
	PageInfo common.OffsetPageInfo `json:"pageInfo"`
}

// CustomFieldsResponse represents the response from the GraphQL query
type CustomFieldsResponse struct {
	CustomFields CustomFieldPagination `json:"customFields"`
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List custom fields in a workspace",
	Long:  "List custom fields in a workspace with pagination. Use --detailed for enhanced reference format with examples.",
	Example: `  blue fields list --workspace <id>
  blue fields list -w <id> --simple
  blue fields list -w <id> --detailed --examples
  blue fields list -w <id> --format json
  blue fields list -w <id> --page 2 --size 25`,
	RunE: runList,
}

var (
	listWorkspace string
	listSimple    bool
	listDetailed  bool
	listExamples  bool
	listFormat    string
	listPage      int
	listPageSize  int
)

func init() {
	listCmd.Flags().StringVarP(&listWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	listCmd.Flags().BoolVarP(&listSimple, "simple", "s", false, "Show only basic custom field information")
	listCmd.Flags().BoolVar(&listDetailed, "detailed", false, "Show enhanced reference format with type descriptions and usage examples")
	listCmd.Flags().BoolVar(&listExamples, "examples", false, "Show example usage for create and update commands (use with --detailed)")
	listCmd.Flags().StringVar(&listFormat, "format", "table", "Output format: table, json, csv")
	listCmd.Flags().IntVar(&listPage, "page", 1, "Page number")
	listCmd.Flags().IntVar(&listPageSize, "size", 50, "Page size")
}

func runList(cmd *cobra.Command, args []string) error {
	if listWorkspace == "" {
		return fmt.Errorf("workspace ID or slug is required. Use --workspace flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)
	client.SetProject(listWorkspace)

	skip := (listPage - 1) * listPageSize
	take := listPageSize

	query := buildCustomFieldsQuery(listWorkspace, skip, take, listDetailed)

	var response CustomFieldsResponse
	if err := client.ExecuteQueryWithResult(query, nil, &response); err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	customFields := response.CustomFields

	if listDetailed {
		return displayDetailed(customFields, listWorkspace, listPage, listPageSize, listSimple, listExamples, listFormat)
	}

	return displayBasic(customFields, listWorkspace, listPage, listPageSize, listSimple, skip)
}

func buildCustomFieldsQuery(projectID string, skip int, take int, detailed bool) string {
	fields := `id
				name
				type
				position
				description
				createdAt
				updatedAt
				customFieldOptions {
					id
					title
					color
				}`

	if detailed {
		fields = `id
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
				}`
	}

	query := fmt.Sprintf(`query {
		customFields(
			filter: { projectId: "%s" }
			skip: %d
			take: %d
		) {
			items {
				%s
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
	}`, projectID, skip, take, fields)

	return query
}

// displayBasic shows the standard project custom fields listing
func displayBasic(customFields CustomFieldPagination, workspace string, page, pageSize int, simple bool, skip int) error {
	fmt.Printf("\n=== Custom Fields in Workspace %s ===\n", workspace)

	totalPages := 1
	if pageSize > 0 && customFields.PageInfo.TotalItems > 0 {
		totalPages = (customFields.PageInfo.TotalItems + pageSize - 1) / pageSize
	}
	fmt.Printf("Page %d of %d (showing %d of %d total)\n\n",
		page, totalPages, len(customFields.Items), customFields.PageInfo.TotalItems)

	if len(customFields.Items) == 0 {
		fmt.Println("No custom fields found in this workspace.")
		return nil
	}

	if simple {
		startNum := skip + 1
		for i, field := range customFields.Items {
			fmt.Printf("%d. %s (%s)\n   ID: %s\n   Position: %.0f\n",
				startNum+i, field.Name, field.Type, field.ID, field.Position)

			if (field.Type == "SELECT_SINGLE" || field.Type == "SELECT_MULTI") && len(field.Options) > 0 {
				fmt.Printf("   Options: ")
				for j, option := range field.Options {
					if j > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s [%s]", option.Title, option.ID)
					if option.Color != "" {
						fmt.Printf(" (%s)", option.Color)
					}
				}
				fmt.Printf("\n")
			}
			fmt.Printf("\n")
		}
	} else {
		startNum := skip + 1
		for i, field := range customFields.Items {
			fmt.Printf("%d. %s\n", startNum+i, field.Name)
			fmt.Printf("   ID: %s\n", field.ID)
			fmt.Printf("   Type: %s\n", field.Type)
			fmt.Printf("   Position: %.0f\n", field.Position)

			if field.Description != "" {
				fmt.Printf("   Description: %s\n", field.Description)
			}

			if (field.Type == "SELECT_SINGLE" || field.Type == "SELECT_MULTI") && len(field.Options) > 0 {
				fmt.Printf("   Available Options (use Option ID or Title for record values):\n")
				for _, option := range field.Options {
					fmt.Printf("     - %s [%s]", option.Title, option.ID)
					if option.Color != "" {
						fmt.Printf(" (%s)", option.Color)
					}
					fmt.Printf("\n")
				}
			}

			fmt.Printf("   Created: %s\n", field.CreatedAt)
			fmt.Printf("   Updated: %s\n", field.UpdatedAt)
			fmt.Println()
		}
	}

	if customFields.PageInfo.HasNextPage || page > 1 {
		fmt.Println("\n=== Navigation ===")
		if page > 1 {
			fmt.Printf("Previous page: blue fields list --workspace %s --page %d", workspace, page-1)
			if simple {
				fmt.Printf(" --simple")
			}
			fmt.Println()
		}
		if customFields.PageInfo.HasNextPage {
			fmt.Printf("Next page: blue fields list --workspace %s --page %d", workspace, page+1)
			if simple {
				fmt.Printf(" --simple")
			}
			fmt.Println()
		}
	}

	return nil
}

// displayDetailed shows the enhanced reference format with type descriptions and examples
func displayDetailed(customFields CustomFieldPagination, workspace string, page, pageSize int, simple, examples bool, format string) error {
	if len(customFields.Items) == 0 {
		fmt.Printf("No custom fields found in workspace %s\n", workspace)
		return nil
	}

	switch format {
	case "json":
		return displayJSON(customFields.Items)
	case "csv":
		return displayCSV(customFields.Items)
	default:
		return displayTable(customFields.Items, workspace, page, pageSize, simple, examples)
	}
}

func displayTable(fields []common.CustomField, workspace string, page, pageSize int, simple, examples bool) error {
	fmt.Printf("\n=== Custom Fields Reference for Workspace %s ===\n", workspace)
	fmt.Printf("Found %d custom fields - Use these Field IDs for record create and update commands\n\n", len(fields))

	if simple {
		fmt.Printf("%-32s | %-25s | %-15s\n", "Field ID", "Field Name", "Type")
		fmt.Printf("%-32s-+-%-25s-+-%-15s\n", strings.Repeat("-", 32), strings.Repeat("-", 25), strings.Repeat("-", 15))

		for _, field := range fields {
			fmt.Printf("%-32s | %-25s | %-15s\n", field.ID, field.Name, field.Type)
		}
	} else {
		for i, field := range fields {
			fmt.Printf("%d. %s\n", i+1, field.Name)
			fmt.Printf("   Field ID: %s\n", field.ID)
			fmt.Printf("   Type: %s\n", field.Type)
			fmt.Printf("   Description: %s\n", getFieldTypeDescription(field))

			if field.Description != "" {
				fmt.Printf("   Notes: %s\n", field.Description)
			}

			if (field.Type == "SELECT_SINGLE" || field.Type == "SELECT_MULTI") && len(field.Options) > 0 {
				fmt.Printf("   Available Options:\n")
				for _, option := range field.Options {
					colorInfo := ""
					if option.Color != "" {
						colorInfo = fmt.Sprintf(" (%s)", option.Color)
					}
					fmt.Printf("      - %s%s\n", option.Title, colorInfo)
				}
			}

			fmt.Printf("   Example Value: %s\n", generateExampleValue(field))

			fmt.Printf("   Usage in Commands:\n")
			fmt.Printf("      records create: --custom-fields \"%s:%s\"\n", field.ID, generateExampleValue(field))
			fmt.Printf("      records update: --custom-fields \"%s:%s\"\n", field.ID, generateExampleValue(field))

			fmt.Println()
		}
	}

	if examples {
		fmt.Println("\n=== Command Examples ===")
		fmt.Println("Create a record with custom field values:")

		var exampleFields []string
		for i, field := range fields {
			if i >= 3 {
				break
			}
			exampleFields = append(exampleFields, fmt.Sprintf("%s:%s", field.ID, generateExampleValue(field)))
		}

		if len(exampleFields) > 0 {
			fmt.Printf("blue records create --list LIST_ID --title \"Sample Record\" --custom-fields \"%s\"\n\n",
				strings.Join(exampleFields, ";"))
		}

		fmt.Println("Update a record's custom fields:")
		if len(exampleFields) > 0 {
			fmt.Printf("blue records update --record RECORD_ID --custom-fields \"%s\"\n\n",
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

			fmt.Printf("blue records list --workspace %s --custom-field \"%s:%s:%s\" --simple\n",
				workspace, field.ID, operator, example)
		}
	}

	return nil
}

func displayJSON(fields []common.CustomField) error {
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

func displayCSV(fields []common.CustomField) error {
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

// getFieldTypeDescription returns a user-friendly description of what values a field accepts
func getFieldTypeDescription(field common.CustomField) string {
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
func generateExampleValue(field common.CustomField) string {
	switch field.Type {
	case "TEXT_SINGLE", "TEXT_MULTI":
		return "\"Sample text\""
	case "NUMBER":
		if field.Min != nil {
			return fmt.Sprintf("%.0f", *field.Min+1)
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
