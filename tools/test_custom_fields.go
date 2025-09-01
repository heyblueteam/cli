package tools

import (
	"encoding/json"
	"flag"
	"fmt"

	"demo-builder/common"
)

// CustomFieldTestResponse structure
type CustomFieldTestResponse struct {
	Todo struct {
		ID                string `json:"id"`
		Title             string `json:"title"`
		CustomFieldValues []struct {
			ID            string      `json:"id"`
			CustomFieldID string      `json:"customFieldId"`
			Value         interface{} `json:"value"`
			CustomField   struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"customField"`
		} `json:"customFieldValues"`
	} `json:"todo"`
}

func RunTestCustomFields(args []string) error {
	fs := flag.NewFlagSet("test-custom-fields", flag.ExitOnError)
	recordID := fs.String("record", "", "Record ID to check custom fields (required)")
	projectID := fs.String("project", "", "Project ID or slug (required)")
	fs.Parse(args)

	if *recordID == "" || *projectID == "" {
		fmt.Println("Error: both -record and -project flags are required")
		return fmt.Errorf("required flags missing")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	client := common.NewClient(config)
	client.SetProject(*projectID)

	// Query to get all custom field values
	query := fmt.Sprintf(`
		query GetTodoCustomFields {
			todo(id: "%s") {
				id
				title
				customFieldValues {
					id
					customFieldId
					value
					customField {
						id
						name
						type
					}
				}
			}
		}
	`, *recordID)

	var response CustomFieldTestResponse
	if err := client.ExecuteQueryWithResult(query, nil, &response); err != nil {
		return fmt.Errorf("failed to get custom fields: %v", err)
	}

	fmt.Printf("=== Custom Field Values for Record %s ===\n", *recordID)
	fmt.Printf("Record: %s\n", response.Todo.Title)
	fmt.Printf("Total Custom Fields Set: %d\n\n", len(response.Todo.CustomFieldValues))

	if len(response.Todo.CustomFieldValues) == 0 {
		fmt.Println("No custom field values are set on this record.")
		return nil
	}

	for i, cfv := range response.Todo.CustomFieldValues {
		fmt.Printf("%d. %s (%s)\n", i+1, cfv.CustomField.Name, cfv.CustomField.Type)
		fmt.Printf("   Field ID: %s\n", cfv.CustomFieldID)
		fmt.Printf("   Value ID: %s\n", cfv.ID)
		
		// Pretty print the value based on type
		switch v := cfv.Value.(type) {
		case string:
			fmt.Printf("   Value: \"%s\"\n", v)
		case float64:
			fmt.Printf("   Value: %g\n", v)
		case bool:
			fmt.Printf("   Value: %t\n", v)
		case []interface{}:
			fmt.Printf("   Value: %v\n", v)
		case map[string]interface{}:
			if jsonBytes, err := json.MarshalIndent(v, "   ", "  "); err == nil {
				fmt.Printf("   Value: %s\n", string(jsonBytes))
			} else {
				fmt.Printf("   Value: %v\n", v)
			}
		default:
			fmt.Printf("   Value: %v (type: %T)\n", v, v)
		}
		fmt.Println()
	}

	return nil
}