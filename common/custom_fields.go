package common

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// FetchFieldTypes fetches custom field ID → type mapping for the current workspace.
// The workspace must be set on the client via SetProject before calling this.
func FetchFieldTypes(client *Client, fieldIDs []string) (map[string]string, error) {
	query := fmt.Sprintf(`
		query {
			customFields(
				skip: 0
				take: %d
			) {
				items {
					id
					type
				}
			}
		}
	`, len(fieldIDs)+20) // fetch extra to cover all fields in workspace

	var response struct {
		CustomFields struct {
			Items []struct {
				ID   string `json:"id"`
				Type string `json:"type"`
			} `json:"items"`
		} `json:"customFields"`
	}

	if err := client.ExecuteQueryWithResult(query, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to fetch custom field types: %w", err)
	}

	result := make(map[string]string)
	for _, item := range response.CustomFields.Items {
		result[item.ID] = item.Type
	}
	return result, nil
}

// ParseCustomFieldValues parses the CLI custom-fields string into CustomFieldValue slice.
// All values are stored as strings; type-aware formatting happens in SetCustomFields.
func ParseCustomFieldValues(customFieldsStr string) ([]CustomFieldValue, error) {
	if customFieldsStr == "" {
		return nil, nil
	}

	var values []CustomFieldValue
	fieldPairs := strings.Split(customFieldsStr, ";")

	for _, pair := range fieldPairs {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid custom field format: %s (expected field_id:value)", pair)
		}

		values = append(values, CustomFieldValue{
			CustomFieldID: strings.TrimSpace(parts[0]),
			Value:         strings.TrimSpace(parts[1]),
		})
	}

	return values, nil
}

// SetCustomFields sets custom field values on a record, using the correct
// GraphQL mutation format based on each field's type from the API.
func SetCustomFields(client *Client, todoID string, customFields []CustomFieldValue) error {
	if len(customFields) == 0 {
		return nil
	}

	// Collect field IDs to look up
	var fieldIDs []string
	for _, cf := range customFields {
		fieldIDs = append(fieldIDs, cf.CustomFieldID)
	}

	// Fetch field types from the API
	fieldTypes, err := FetchFieldTypes(client, fieldIDs)
	if err != nil {
		return err
	}

	for _, cfv := range customFields {
		valueStr, ok := cfv.Value.(string)
		if !ok {
			valueStr = fmt.Sprintf("%v", cfv.Value)
		}

		fieldType := fieldTypes[cfv.CustomFieldID]
		mutationValue, err := formatFieldValue(fieldType, valueStr)
		if err != nil {
			return fmt.Errorf("failed to format value for field %s (type %s): %w", cfv.CustomFieldID, fieldType, err)
		}

		mutation := fmt.Sprintf(`
			mutation SetTodoCustomField {
				setTodoCustomField(input: {
					todoId: "%s"
					customFieldId: "%s"
					%s
				})
			}
		`, todoID, cfv.CustomFieldID, mutationValue)

		var response struct {
			SetTodoCustomField bool `json:"setTodoCustomField"`
		}
		if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
			return fmt.Errorf("failed to set custom field %s: %w", cfv.CustomFieldID, err)
		}
	}

	return nil
}

// formatFieldValue converts a string value to the correct GraphQL mutation
// parameter based on the field type.
func formatFieldValue(fieldType, value string) (string, error) {
	switch fieldType {
	case "NUMBER":
		n, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return "", fmt.Errorf("invalid number value %q: %w", value, err)
		}
		return fmt.Sprintf("number: %g", n), nil

	case "PERCENT":
		n, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return "", fmt.Errorf("invalid percent value %q: %w", value, err)
		}
		return fmt.Sprintf("number: %g", n), nil

	case "CURRENCY":
		n, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return "", fmt.Errorf("invalid currency value %q: %w", value, err)
		}
		return fmt.Sprintf("number: %g", n), nil

	case "COUNTRY":
		codes := parseCommaSeparated(value)
		return fmt.Sprintf("countryCodes: [%s]", quoteAndJoin(codes)), nil

	case "SELECT_SINGLE":
		// Value should be an option ID (trailing comma from old format is trimmed)
		optionID := strings.TrimRight(value, ",")
		return fmt.Sprintf(`customFieldOptionIds: ["%s"]`, optionID), nil

	case "SELECT_MULTI":
		ids := parseCommaSeparated(value)
		return fmt.Sprintf("customFieldOptionIds: [%s]", quoteAndJoin(ids)), nil

	case "CHECKBOX":
		b, err := strconv.ParseBool(value)
		if err != nil {
			return "", fmt.Errorf("invalid checkbox value %q (expected true/false): %w", value, err)
		}
		return fmt.Sprintf("checked: %t", b), nil

	case "RATING":
		n, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return "", fmt.Errorf("invalid rating value %q: %w", value, err)
		}
		return fmt.Sprintf("number: %g", n), nil

	case "DATE":
		return fmt.Sprintf(`text: "%s"`, escapeGraphQL(value)), nil

	case "EMAIL", "PHONE", "URL", "TEXT_SINGLE", "TEXT_MULTI":
		return fmt.Sprintf(`text: "%s"`, escapeGraphQL(value)), nil

	case "LOCATION":
		// Expect "lat,lng" format
		parts := strings.SplitN(value, ",", 2)
		if len(parts) != 2 {
			return "", fmt.Errorf("invalid location value %q (expected lat,lng)", value)
		}
		lat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
		if err != nil {
			return "", fmt.Errorf("invalid latitude in %q: %w", value, err)
		}
		lng, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return "", fmt.Errorf("invalid longitude in %q: %w", value, err)
		}
		return fmt.Sprintf("latitude: %g, longitude: %g", lat, lng), nil

	default:
		// Unknown or unsupported type: fall back to text
		if value == "" {
			return `text: ""`, nil
		}
		// Try number first, then fall back to text
		if n, err := strconv.ParseFloat(value, 64); err == nil {
			return fmt.Sprintf("number: %g", n), nil
		}
		return fmt.Sprintf(`text: "%s"`, escapeGraphQL(value)), nil
	}
}

// parseCommaSeparated splits a comma-separated string, trimming whitespace.
func parseCommaSeparated(s string) []string {
	parts := strings.Split(s, ",")
	var result []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

// quoteAndJoin wraps each string in quotes and joins with commas.
func quoteAndJoin(items []string) string {
	var quoted []string
	for _, item := range items {
		quoted = append(quoted, fmt.Sprintf(`"%s"`, item))
	}
	return strings.Join(quoted, ", ")
}

// escapeGraphQL escapes special characters for GraphQL string values.
func escapeGraphQL(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}

// FormatCustomFieldValueForJSON formats a custom field value for JSON output.
func FormatCustomFieldValueForJSON(cfValue interface{}) interface{} {
	if cfValue == nil {
		return nil
	}
	// Try to parse as JSON first
	if str, ok := cfValue.(string); ok {
		var parsed interface{}
		if err := json.Unmarshal([]byte(str), &parsed); err == nil {
			return parsed
		}
		return str
	}
	return cfValue
}
