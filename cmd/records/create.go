package records

import (
	"fmt"
	"strings"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

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

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new record",
	Long:  "Create a new record/todo within a list.",
	Example: `  blue records create --workspace <id> --list <id> --title "Fix login bug"
  blue records create -w <id> -l <id> -t "Task" --description "Details here"
  blue records create -w <id> -l <id> -t "Task" --custom-fields "cf123:value;cf456:42"`,
	RunE: runCreate,
}

var (
	createWorkspace    string
	createList         string
	createTitle        string
	createDescription  string
	createPlacement    string
	createAssignees    string
	createCustomFields string
	createSimple       bool
)

func init() {
	createCmd.Flags().StringVarP(&createWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	createCmd.Flags().StringVarP(&createList, "list", "l", "", "List ID to create the record in (required)")
	createCmd.Flags().StringVarP(&createTitle, "title", "t", "", "Title of the record (required)")
	createCmd.Flags().StringVarP(&createDescription, "description", "d", "", "Description of the record")
	createCmd.Flags().StringVar(&createPlacement, "placement", "", "Placement in list: TOP or BOTTOM")
	createCmd.Flags().StringVar(&createAssignees, "assignees", "", "Comma-separated assignee IDs")
	createCmd.Flags().StringVar(&createCustomFields, "custom-fields", "", "Custom field values (format: field_id1:value1;field_id2:value2)")
	createCmd.Flags().BoolVarP(&createSimple, "simple", "s", false, "Simple output format")
}

func runCreate(cmd *cobra.Command, args []string) error {
	if createWorkspace == "" || createList == "" || createTitle == "" {
		return fmt.Errorf("--workspace, --list and --title flags are required")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := common.NewClient(config)
	client.SetProject(createWorkspace)

	// Build optional fields
	var optionalFields []string
	if createDescription != "" {
		optionalFields = append(optionalFields, fmt.Sprintf(`description: "%s"`, strings.ReplaceAll(createDescription, `"`, `\"`)))
	}
	if createPlacement != "" {
		optionalFields = append(optionalFields, fmt.Sprintf(`placement: %s`, createPlacement))
	}
	if createAssignees != "" {
		assigneeList := strings.Split(createAssignees, ",")
		var assigneeStrings []string
		for _, a := range assigneeList {
			assigneeStrings = append(assigneeStrings, fmt.Sprintf(`"%s"`, strings.TrimSpace(a)))
		}
		optionalFields = append(optionalFields, fmt.Sprintf(`assigneeIds: [%s]`, strings.Join(assigneeStrings, ", ")))
	}

	mutation := fmt.Sprintf(`
		mutation CreateTodo {
			createTodo(input: {
				todoListId: "%s"
				title: "%s"
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
	`, createList, strings.ReplaceAll(createTitle, `"`, `\"`), strings.Join(optionalFields, "\n\t\t\t\t"))

	var response CreateTodoResponse
	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		return fmt.Errorf("failed to create record: %w", err)
	}

	record := response.CreateTodo

	// Set custom fields if provided
	if createCustomFields != "" {
		customFieldValues, err := parseCustomFieldValues(createCustomFields)
		if err != nil {
			return fmt.Errorf("failed to parse custom fields: %w", err)
		}
		if err := executeSetCustomFields(client, record.ID, customFieldValues); err != nil {
			return fmt.Errorf("record created but failed to set custom fields: %w", err)
		}
	}

	if createSimple {
		fmt.Printf("Created record: %s (ID: %s)\n", record.Title, record.ID)
	} else {
		fmt.Printf("=== Record Created Successfully ===\n")
		fmt.Printf("ID: %s\n", record.ID)
		fmt.Printf("Title: %s\n", record.Title)
		fmt.Printf("Position: %.0f\n", record.Position)
		fmt.Printf("List: %s (%s)\n", record.TodoList.Title, record.TodoList.ID)
	}

	return nil
}

// parseCustomFieldValues parses custom field values from the CLI format
func parseCustomFieldValues(customFieldsStr string) ([]common.CustomFieldValue, error) {
	if customFieldsStr == "" {
		return nil, nil
	}

	var values []common.CustomFieldValue
	fieldPairs := strings.Split(customFieldsStr, ";")

	for _, pair := range fieldPairs {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid custom field format: %s (expected field_id:value)", pair)
		}

		values = append(values, common.CustomFieldValue{
			CustomFieldID: strings.TrimSpace(parts[0]),
			Value:         strings.TrimSpace(parts[1]),
		})
	}

	return values, nil
}

// SetCustomFieldResponse represents the response from setTodoCustomField mutation
type SetCustomFieldResponse struct {
	SetTodoCustomField bool `json:"setTodoCustomField"`
}

// executeSetCustomFields sets custom field values on a record
func executeSetCustomFields(client *common.Client, todoID string, customFields []common.CustomFieldValue) error {
	for _, cfv := range customFields {
		var valueStr string

		switch v := cfv.Value.(type) {
		case string:
			if strings.Contains(v, ",") {
				items := strings.Split(v, ",")
				var arrayItems []string
				for _, item := range items {
					arrayItems = append(arrayItems, fmt.Sprintf(`"%s"`, strings.TrimSpace(item)))
				}
				valueStr = fmt.Sprintf(`customFieldOptionIds: [%s]`, strings.Join(arrayItems, ", "))
			} else {
				valueStr = fmt.Sprintf(`text: "%s"`, strings.ReplaceAll(v, `"`, `\"`))
			}
		case float64:
			valueStr = fmt.Sprintf(`number: %g`, v)
		case bool:
			valueStr = fmt.Sprintf(`checked: %t`, v)
		default:
			valueStr = fmt.Sprintf(`text: "%v"`, v)
		}

		mutation := fmt.Sprintf(`
			mutation SetTodoCustomField {
				setTodoCustomField(input: {
					todoId: "%s"
					customFieldId: "%s"
					%s
				})
			}
		`, todoID, cfv.CustomFieldID, valueStr)

		var response SetCustomFieldResponse
		if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
			return fmt.Errorf("failed to set custom field %s: %w", cfv.CustomFieldID, err)
		}
	}

	return nil
}
