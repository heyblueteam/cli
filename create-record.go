package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"strings"
)

// CustomFieldValue represents a custom field value input
type CustomFieldValue struct {
	CustomFieldID string      `json:"customFieldId"`
	Value         interface{} `json:"value"`
}

// CreateTodoInput represents the input for creating a todo with optional custom fields
type CreateTodoInput struct {
	TodoListID        string              `json:"todoListId"`
	Title             string              `json:"title"`
	Position          *float64            `json:"position,omitempty"`
	Description       *string             `json:"description,omitempty"`
	Placement         *string             `json:"placement,omitempty"`
	AssigneeIDs       []string            `json:"assigneeIds,omitempty"`
	CustomFieldValues []CustomFieldValue  `json:"customFieldValues,omitempty"`
}

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
		CustomFieldValues []struct {
			ID            string `json:"id"`
			CustomFieldID string `json:"customFieldId"`
			Value         interface{} `json:"value"`
			CustomField   struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"customField"`
		} `json:"customFieldValues"`
	} `json:"createTodo"`
}

// parseCustomFieldValues parses the custom field values from command line arguments
func parseCustomFieldValues(customFieldsStr string) ([]CustomFieldValue, error) {
	if customFieldsStr == "" {
		return nil, nil
	}

	var customFieldValues []CustomFieldValue
	
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
		
		customFieldValues = append(customFieldValues, CustomFieldValue{
			CustomFieldID: fieldID,
			Value:         value,
		})
	}
	
	return customFieldValues, nil
}

// buildCustomFieldValuesString builds the GraphQL custom field values string
func buildCustomFieldValuesString(customFieldValues []CustomFieldValue) string {
	if len(customFieldValues) == 0 {
		return ""
	}
	
	var valueStrings []string
	
	for _, cfv := range customFieldValues {
		var valueStr string
		
		switch v := cfv.Value.(type) {
		case string:
			valueStr = fmt.Sprintf(`"%s"`, strings.ReplaceAll(v, `"`, `\"`))
		case float64:
			valueStr = fmt.Sprintf(`%g`, v)
		case bool:
			valueStr = fmt.Sprintf(`%t`, v)
		case []string:
			var arrayItems []string
			for _, item := range v {
				arrayItems = append(arrayItems, fmt.Sprintf(`"%s"`, strings.ReplaceAll(item, `"`, `\"`)))
			}
			valueStr = fmt.Sprintf(`[%s]`, strings.Join(arrayItems, ", "))
		default:
			// Fallback to JSON marshaling
			if jsonBytes, err := json.Marshal(v); err == nil {
				valueStr = string(jsonBytes)
			} else {
				valueStr = fmt.Sprintf(`"%v"`, v)
			}
		}
		
		valueStrings = append(valueStrings, fmt.Sprintf(`{
			customFieldId: "%s"
			value: %s
		}`, cfv.CustomFieldID, valueStr))
	}
	
	return fmt.Sprintf(`customFields: [%s]`, strings.Join(valueStrings, ", "))
}

func main() {
	var projectID = flag.String("project", "", "Project ID or Project slug (required)")
	var listID = flag.String("list", "", "List ID to create the record in (required)")
	var title = flag.String("title", "", "Title of the record (required)")
	var description = flag.String("description", "", "Description of the record")
	var placement = flag.String("placement", "", "Placement in list: TOP or BOTTOM")
	var assignees = flag.String("assignees", "", "Comma-separated assignee IDs")
	var customFields = flag.String("custom-fields", "", "Custom field values in format: field_id1:value1;field_id2:value2")
	var simple = flag.Bool("simple", false, "Simple output format")

	flag.Parse()

	if *projectID == "" || *listID == "" || *title == "" {
		fmt.Println("Error: -project, -list and -title flags are required")
		fmt.Println("\nUsage:")
		fmt.Println("  go run auth.go create-record.go -project PROJECT_ID_OR_SLUG -list LIST_ID -title \"Record Title\" [flags]")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
		fmt.Println("\nCustom Fields Format:")
		fmt.Println("  -custom-fields \"field_id1:value1;field_id2:value2\"")
		fmt.Println("  Examples:")
		fmt.Println("    Text field: -custom-fields \"cf123:Hello World\"")
		fmt.Println("    Number field: -custom-fields \"cf456:42.5\"")
		fmt.Println("    Boolean field: -custom-fields \"cf789:true\"")
		fmt.Println("    Multi-select: -custom-fields 'cf000:[\"option1\",\"option2\"]'")
		fmt.Println("    Multiple fields: -custom-fields \"cf123:Hello;cf456:42;cf789:true\"")
		return
	}

	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	client := NewClient(config)
	
	// Set project context from the provided flag (auto-detects ID vs slug)
	client.SetProject(*projectID)

	input := CreateTodoInput{
		TodoListID: *listID,
		Title:      *title,
	}

	if *description != "" {
		input.Description = description
	}

	if *placement != "" {
		input.Placement = placement
	}

	if *assignees != "" {
		assigneeList := strings.Split(*assignees, ",")
		for i, assignee := range assigneeList {
			assigneeList[i] = strings.TrimSpace(assignee)
		}
		input.AssigneeIDs = assigneeList
	}

	// Parse custom field values
	if *customFields != "" {
		customFieldValues, err := parseCustomFieldValues(*customFields)
		if err != nil {
			log.Fatalf("Failed to parse custom fields: %v", err)
		}
		input.CustomFieldValues = customFieldValues
	}

	// Build the mutation with optional fields
	var descriptionField string
	if input.Description != nil {
		descriptionField = fmt.Sprintf(`description: "%s"`, strings.ReplaceAll(*input.Description, `"`, `\"`))
	}
	
	var placementField string
	if input.Placement != nil {
		placementField = fmt.Sprintf(`placement: %s`, *input.Placement)
	}

	var assigneesField string
	if len(input.AssigneeIDs) > 0 {
		var assigneeStrings []string
		for _, assigneeID := range input.AssigneeIDs {
			assigneeStrings = append(assigneeStrings, fmt.Sprintf(`"%s"`, assigneeID))
		}
		assigneesField = fmt.Sprintf(`assigneeIds: [%s]`, strings.Join(assigneeStrings, ", "))
	}

	var customFieldValuesField string
	if len(input.CustomFieldValues) > 0 {
		customFieldValuesField = buildCustomFieldValuesString(input.CustomFieldValues)
	}

	// Determine response fields based on whether custom fields are requested
	responseFields := `
		id
		title
		position
		todoList {
			id
			title
		}`

	if len(input.CustomFieldValues) > 0 {
		responseFields += `
		customFieldValues {
			id
			customFieldId
			value
			customField {
				id
				name
				type
			}
		}`
	}

	mutation := fmt.Sprintf(`
		mutation CreateTodo {
			createTodo(input: {
				todoListId: "%s"
				title: "%s"
				%s
				%s
				%s
				%s
			}) {%s
			}
		}
	`, input.TodoListID, strings.ReplaceAll(input.Title, `"`, `\"`), descriptionField, placementField, assigneesField, customFieldValuesField, responseFields)

	var response CreateTodoResponse
	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	record := response.CreateTodo

	if *simple {
		fmt.Printf("Created record: %s (ID: %s)\n", record.Title, record.ID)
		if len(input.CustomFieldValues) > 0 {
			fmt.Printf("Custom fields set: %d\n", len(record.CustomFieldValues))
		}
	} else {
		fmt.Printf("=== Record Created Successfully ===\n")
		fmt.Printf("ID: %s\n", record.ID)
		fmt.Printf("Title: %s\n", record.Title)
		fmt.Printf("Position: %.0f\n", record.Position)
		fmt.Printf("List: %s (%s)\n", record.TodoList.Title, record.TodoList.ID)
		
		if len(record.CustomFieldValues) > 0 {
			fmt.Printf("\n=== Custom Field Values ===\n")
			for _, cfv := range record.CustomFieldValues {
				fmt.Printf("%s (%s): %v\n", cfv.CustomField.Name, cfv.CustomField.Type, cfv.Value)
			}
		}
	}
}