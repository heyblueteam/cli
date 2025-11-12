package tools

import (
	"fmt"
	"strings"
	
	"cli/common"
)

// SetCustomFieldResponse represents the response from setTodoCustomField mutation
type SetCustomFieldResponse struct {
	SetTodoCustomField bool `json:"setTodoCustomField"`
}

// buildCustomFieldMutation builds individual custom field mutations
func buildCustomFieldMutation(todoID string, cfv common.CustomFieldValue) string {
	var valueStr string

	switch v := cfv.Value.(type) {
	case string:
		// Check if it looks like an array of option IDs or titles for multi-select
		if strings.Contains(v, ",") {
			// Split and treat as array for multi-select
			items := strings.Split(v, ",")
			var arrayItems []string
			for _, item := range items {
				arrayItems = append(arrayItems, fmt.Sprintf(`"%s"`, strings.TrimSpace(item)))
			}
			valueStr = fmt.Sprintf(`customFieldOptionIds: [%s]`, strings.Join(arrayItems, ", "))
		} else {
			// Single value - could be text, option ID, or option title
			valueStr = fmt.Sprintf(`text: "%s"`, strings.ReplaceAll(v, `"`, `\"`))
		}
	case float64:
		valueStr = fmt.Sprintf(`number: %g`, v)
	case bool:
		valueStr = fmt.Sprintf(`checked: %t`, v)
	case []string:
		var arrayItems []string
		for _, item := range v {
			arrayItems = append(arrayItems, fmt.Sprintf(`"%s"`, strings.ReplaceAll(item, `"`, `\"`)))
		}
		valueStr = fmt.Sprintf(`customFieldOptionIds: [%s]`, strings.Join(arrayItems, ", "))
	default:
		// Fallback to text
		valueStr = fmt.Sprintf(`text: "%v"`, v)
	}
	
	return fmt.Sprintf(`
		mutation SetTodoCustomField {
			setTodoCustomField(input: {
				todoId: "%s"
				customFieldId: "%s"
				%s
			})
		}
	`, todoID, cfv.CustomFieldID, valueStr)
}

// executeSetCustomFields sets custom field values on a record
func executeSetCustomFields(client *common.Client, todoID string, customFields []common.CustomFieldValue) error {
	if len(customFields) == 0 {
		return nil
	}
	
	for _, cfv := range customFields {
		mutation := buildCustomFieldMutation(todoID, cfv)
		var response SetCustomFieldResponse
		if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
			return fmt.Errorf("failed to set custom field %s: %v", cfv.CustomFieldID, err)
		}
	}
	
	return nil
}