package forms

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// FormFieldsField mirrors the GraphQL enum.
var validFieldTypes = map[string]bool{
	"title":       true,
	"description": true,
	"tags":        true,
	"startedAt":   true,
	"duedAt":      true,
	"custom":      true,
}

// FieldSpec is the parsed form-field specification used by both the inline
// --field flag and the --fields-file JSON.
type FieldSpec struct {
	ID               string   `json:"id,omitempty"`
	Field            string   `json:"field"`
	CustomFieldID    string   `json:"customFieldId,omitempty"`
	Name             string   `json:"name"`
	Placeholder      string   `json:"placeholder,omitempty"`
	Position         *float64 `json:"position,omitempty"`
	Required         *bool    `json:"required,omitempty"`
	Hidden           *bool    `json:"hidden,omitempty"`
	ExtraInfo        *string  `json:"extraInfo,omitempty"`
	AddToDescription *bool    `json:"addToDescription,omitempty"`
}

// parseInlineField turns "type=custom;customField=cf_xxx;name=Priority;..."
// into a FieldSpec.
func parseInlineField(s string) (FieldSpec, error) {
	spec := FieldSpec{}
	for _, raw := range strings.Split(s, ";") {
		raw = strings.TrimSpace(raw)
		if raw == "" {
			continue
		}
		kv := strings.SplitN(raw, "=", 2)
		if len(kv) != 2 {
			return spec, fmt.Errorf("invalid field segment %q (expected key=value)", raw)
		}
		key := strings.TrimSpace(kv[0])
		val := strings.TrimSpace(kv[1])
		switch key {
		case "type", "field":
			spec.Field = val
		case "customField", "customFieldId", "cf":
			spec.CustomFieldID = val
		case "id", "formFieldId":
			spec.ID = val
		case "name":
			spec.Name = val
		case "placeholder":
			spec.Placeholder = val
		case "position":
			f, err := strconv.ParseFloat(val, 64)
			if err != nil {
				return spec, fmt.Errorf("invalid position %q: %w", val, err)
			}
			spec.Position = &f
		case "required":
			b, err := strconv.ParseBool(val)
			if err != nil {
				return spec, fmt.Errorf("invalid required %q: %w", val, err)
			}
			spec.Required = &b
		case "hidden":
			b, err := strconv.ParseBool(val)
			if err != nil {
				return spec, fmt.Errorf("invalid hidden %q: %w", val, err)
			}
			spec.Hidden = &b
		case "addToDescription":
			b, err := strconv.ParseBool(val)
			if err != nil {
				return spec, fmt.Errorf("invalid addToDescription %q: %w", val, err)
			}
			spec.AddToDescription = &b
		case "extraInfo":
			v := val
			spec.ExtraInfo = &v
		default:
			return spec, fmt.Errorf("unknown field key %q", key)
		}
	}
	return spec, nil
}

// loadFieldsFile parses a JSON file containing an array of FieldSpec entries.
func loadFieldsFile(path string) ([]FieldSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read fields file %q: %w", path, err)
	}
	var specs []FieldSpec
	if err := json.Unmarshal(data, &specs); err != nil {
		return nil, fmt.Errorf("failed to parse fields file %q: %w", path, err)
	}
	return specs, nil
}

// validateFieldSpec returns an error if the spec is structurally invalid.
func validateFieldSpec(spec FieldSpec) error {
	if spec.Field == "" {
		return fmt.Errorf("field type is required (use type=title|description|tags|startedAt|duedAt|custom)")
	}
	if !validFieldTypes[spec.Field] {
		return fmt.Errorf("invalid field type %q (must be one of: title, description, tags, startedAt, duedAt, custom)", spec.Field)
	}
	if spec.Field == "custom" && spec.CustomFieldID == "" {
		return fmt.Errorf("customField is required when type=custom")
	}
	if spec.Name == "" {
		return fmt.Errorf("name is required for each field")
	}
	return nil
}

// collectFieldSpecs merges --fields-file (first) with repeated --field flags
// (after), validating each entry.
func collectFieldSpecs(filePath string, inline []string) ([]FieldSpec, error) {
	var specs []FieldSpec
	if filePath != "" {
		fileSpecs, err := loadFieldsFile(filePath)
		if err != nil {
			return nil, err
		}
		specs = append(specs, fileSpecs...)
	}
	for _, raw := range inline {
		spec, err := parseInlineField(raw)
		if err != nil {
			return nil, err
		}
		specs = append(specs, spec)
	}
	for i, spec := range specs {
		if err := validateFieldSpec(spec); err != nil {
			return nil, fmt.Errorf("field #%d: %w", i+1, err)
		}
	}
	return specs, nil
}

// FormSummary is the minimal projection used by list and create output.
type FormSummary struct {
	ID            string  `json:"id"`
	UID           string  `json:"uid"`
	Title         string  `json:"title"`
	Description   *string `json:"description"`
	IsActive      bool    `json:"isActive"`
	Theme         string  `json:"theme"`
	PrimaryColor  string  `json:"primaryColor"`
	HideBranding  bool    `json:"hideBranding"`
	CreatedAt     string  `json:"createdAt"`
	UpdatedAt     string  `json:"updatedAt"`
}

// FormDetail is the full projection used by get/create/update output.
type FormDetail struct {
	ID            string         `json:"id"`
	UID           string         `json:"uid"`
	Title         string         `json:"title"`
	Description   *string        `json:"description"`
	FooterText    *string        `json:"footerText"`
	ShowFooter    *bool          `json:"showFooter"`
	IsActive      bool           `json:"isActive"`
	Theme         string         `json:"theme"`
	PrimaryColor  string         `json:"primaryColor"`
	HideBranding  bool           `json:"hideBranding"`
	ResponseText  *string        `json:"responseText"`
	SubmitText    *string        `json:"submitText"`
	ImageURL      *string        `json:"imageURL"`
	RedirectURL   *string        `json:"redirectURL"`
	SnapshotURL   *string        `json:"snapshotURL"`
	CreatedAt     string         `json:"createdAt"`
	UpdatedAt     string         `json:"updatedAt"`
	TodoList      *FormTodoList  `json:"todoList"`
	FormFields    []FormFieldOut `json:"formFields"`
}

type FormTodoList struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type FormFieldOut struct {
	ID               string                `json:"id"`
	UID              string                `json:"uid"`
	Field            string                `json:"field"`
	Name             string                `json:"name"`
	Placeholder      string                `json:"placeholder"`
	Required         bool                  `json:"required"`
	Position         float64               `json:"position"`
	AddToDescription bool                  `json:"addToDescription"`
	Hidden           bool                  `json:"hidden"`
	ExtraInfo        *string               `json:"extraInfo"`
	CustomField      *FormFieldCustomField `json:"customField"`
}

type FormFieldCustomField struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}
