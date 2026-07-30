package common

import (
	"fmt"
	"slices"
	"strings"
)

// CustomFieldRef is a custom field resolved to everything the chart APIs need.
// Charts are inconsistent about which identifier they take: a stat card's
// segment value references a field by `customFieldId`, while an auto-generated
// bar or pie chart matches on `customFieldName` *and* `customFieldType`
// together. Resolving once into this triple lets a caller supply whichever the
// endpoint wants without the user having to know the difference.
type CustomFieldRef struct {
	ID   string
	Name string
	Type string
}

// GroupableCustomFieldTypes are the custom field types the chart API can build
// an x-axis or pie grouping from. Mirrors SUPPORTED_GROUP_BY_CUSTOM_FIELD_TYPES
// in the api. Grouping by anything else has no column behind it and comes back
// as an empty chart rather than an error, so callers check this first and fail
// with something a user can act on.
var GroupableCustomFieldTypes = []string{
	"SELECT_SINGLE",
	"SELECT_MULTI",
	"CHECKBOX",
	"COUNTRY",
	"DATE",
	"REFERENCE",
	"REFERENCED_BY",
	"ASSIGNEE",
}

// IsGroupableCustomFieldType reports whether a chart can group by this field type.
func IsGroupableCustomFieldType(fieldType string) bool {
	return slices.Contains(GroupableCustomFieldTypes, fieldType)
}

// ResolveCustomField looks up a custom field in the client's current workspace
// by ID or by exact name (case-insensitive). The workspace must already be set
// via SetProject.
//
// A name that matches more than one field is an error rather than an arbitrary
// pick: charts match on name, so silently choosing one of two same-named fields
// would produce a chart measuring something other than what was asked for.
func ResolveCustomField(client *Client, ref string) (*CustomFieldRef, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, fmt.Errorf("no custom field given")
	}

	fields, err := FetchCustomFields(client)
	if err != nil {
		return nil, err
	}

	var byName []CustomFieldRef
	for _, field := range fields {
		if field.ID == ref {
			return &field, nil
		}
		if strings.EqualFold(field.Name, ref) {
			byName = append(byName, field)
		}
	}

	switch len(byName) {
	case 0:
		return nil, fmt.Errorf("no custom field named or with ID %q in this workspace", ref)
	case 1:
		return &byName[0], nil
	default:
		return nil, fmt.Errorf(
			"%d custom fields are named %q in this workspace — pass the field ID instead",
			len(byName), ref,
		)
	}
}

// FetchCustomFields returns every custom field in the client's current workspace.
func FetchCustomFields(client *Client) ([]CustomFieldRef, error) {
	const query = `
		query CustomFields($skip: Int, $take: Int) {
			customFields(skip: $skip, take: $take) {
				items {
					id
					name
					type
				}
			}
		}
	`

	var response struct {
		CustomFields struct {
			Items []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"items"`
		} `json:"customFields"`
	}

	variables := map[string]interface{}{"skip": 0, "take": 200}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return nil, fmt.Errorf("failed to fetch custom fields: %w", err)
	}

	fields := make([]CustomFieldRef, 0, len(response.CustomFields.Items))
	for _, item := range response.CustomFields.Items {
		fields = append(fields, CustomFieldRef{ID: item.ID, Name: item.Name, Type: item.Type})
	}
	return fields, nil
}
