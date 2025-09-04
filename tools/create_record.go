package tools

import (
	"flag"
	"fmt"
	"strings"

	"demo-builder/common"
)

// CustomFieldValue and CreateTodoInput are already defined in common/types.go

// CreateTodoResponse represents the GraphQL response
type CreateTodoResponse struct {
	CreateTodo struct {
		ID       string  `json:"id"`
		Title    string  `json:"title"`
		Position float64 `json:"position"`
		TodoList struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"todoList"`
	} `json:"createTodo"`
}

// parseCustomFieldValues parses the custom field values from command line arguments
func parseCustomFieldValues(customFieldsStr string) ([]common.CustomFieldValue, error) {
	if customFieldsStr == "" {
		return nil, nil
	}

	var customFieldValues []common.CustomFieldValue

	// Split by semicolon for multiple custom fields
	fieldPairs := strings.Split(customFieldsStr, ";")

	for _, pair := range fieldPairs {
		// Split by colon for field_id:value
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid custom field format: %s (expected field_id:value)", pair)
		}

		fieldID := strings.TrimSpace(parts[0])
		valueStr := strings.TrimSpace(parts[1])

		// For now, always use string values (API might expect strings even for numbers)
		value := valueStr

		customFieldValues = append(customFieldValues, common.CustomFieldValue{
			CustomFieldID: fieldID,
			Value:         value,
		})
	}

	return customFieldValues, nil
}


func RunCreateRecord(args []string) error {
	fs := flag.NewFlagSet("create-record", flag.ExitOnError)
	projectID := fs.String("project", "", "Project ID or Project slug (required)")
	listID := fs.String("list", "", "List ID to create the record in (required)")
	title := fs.String("title", "", "Title of the record (required)")
	description := fs.String("description", "", "Description of the record")
	placement := fs.String("placement", "", "Placement in list: TOP or BOTTOM")
	assignees := fs.String("assignees", "", "Comma-separated assignee IDs")
	customFields := fs.String("custom-fields", "", "Custom field values in format: field_id1:value1;field_id2:value2")
	simple := fs.Bool("simple", false, "Simple output format")
	fs.Parse(args)

	if *projectID == "" || *listID == "" || *title == "" {
		fmt.Println("Error: -project, -list and -title flags are required")
		fmt.Println("\nUsage:")
		fmt.Println("  go run auth.go create-record.go -project PROJECT_ID_OR_SLUG -list LIST_ID -title \"Record Title\" [flags]")
		fmt.Println("\nFlags:")
		fs.PrintDefaults()
		fmt.Println("\nCustom Fields Format:")
		fmt.Println("  -custom-fields \"field_id1:value1;field_id2:value2\"")
		fmt.Println("  Examples:")
		fmt.Println("    Text field: -custom-fields \"cf123:Hello World\"")
		fmt.Println("    Number field: -custom-fields \"cf456:42.5\"")
		fmt.Println("    Boolean field: -custom-fields \"cf789:true\"")
		fmt.Println("    Multi-select: -custom-fields 'cf000:[\"option1\",\"option2\"]'")
		fmt.Println("    Multiple fields: -custom-fields \"cf123:Hello;cf456:42;cf789:true\"")
		return fmt.Errorf("required flags missing")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	client := common.NewClient(config)

	// Set project context from the provided flag (auto-detects ID vs slug)
	client.SetProject(*projectID)

	input := common.CreateTodoInput{
		TodoListID: *listID,
		Title:      *title,
	}

	if *description != "" {
		input.Description = *description
	}

	if *placement != "" {
		input.TodoListPlacement = *placement
	}

	if *assignees != "" {
		assigneeList := strings.Split(*assignees, ",")
		for i, assignee := range assigneeList {
			assigneeList[i] = strings.TrimSpace(assignee)
		}
		input.AssigneeIds = assigneeList
	}

	// Parse custom field values
	if *customFields != "" {
		customFieldValues, err := parseCustomFieldValues(*customFields)
		if err != nil {
			return fmt.Errorf("failed to parse custom fields: %v", err)
		}
		input.CustomFieldValues = customFieldValues
	}

	// Build the mutation with optional fields
	var descriptionField string
	if input.Description != "" {
		descriptionField = fmt.Sprintf(`description: "%s"`, strings.ReplaceAll(input.Description, `"`, `\"`))
	}

	var placementField string
	if input.TodoListPlacement != "" {
		placementField = fmt.Sprintf(`placement: %s`, input.TodoListPlacement)
	}

	var assigneesField string
	if len(input.AssigneeIds) > 0 {
		var assigneeStrings []string
		for _, assigneeID := range input.AssigneeIds {
			assigneeStrings = append(assigneeStrings, fmt.Sprintf(`"%s"`, assigneeID))
		}
		assigneesField = fmt.Sprintf(`assigneeIds: [%s]`, strings.Join(assigneeStrings, ", "))
	}

	// Create the basic mutation without custom fields
	mutation := fmt.Sprintf(`
		mutation CreateTodo {
			createTodo(input: {
				todoListId: "%s"
				title: "%s"
				%s
				%s
				%s
			}) {
				id
				title
				position
				todoList {
					id
					title
				}
			}
		}
	`, input.TodoListID, strings.ReplaceAll(input.Title, `"`, `\"`), descriptionField, placementField, assigneesField)

	var response CreateTodoResponse
	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		return fmt.Errorf("request failed: %v", err)
	}

	record := response.CreateTodo

	// Set custom fields if provided
	if len(input.CustomFieldValues) > 0 {
		if err := executeSetCustomFields(client, record.ID, input.CustomFieldValues); err != nil {
			return fmt.Errorf("record created but failed to set custom fields: %v", err)
		}
	}

	if *simple {
		fmt.Printf("Created record: %s (ID: %s)\n", record.Title, record.ID)
		if len(input.CustomFieldValues) > 0 {
			fmt.Printf("Custom fields set: %d\n", len(input.CustomFieldValues))
		}
	} else {
		fmt.Printf("=== Record Created Successfully ===\n")
		fmt.Printf("ID: %s\n", record.ID)
		fmt.Printf("Title: %s\n", record.Title)
		fmt.Printf("Position: %.0f\n", record.Position)
		fmt.Printf("List: %s (%s)\n", record.TodoList.Title, record.TodoList.ID)

		if len(input.CustomFieldValues) > 0 {
			fmt.Printf("Custom fields set: %d\n", len(input.CustomFieldValues))
		}
	}

	return nil
}
