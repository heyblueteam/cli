package records

import (
	"fmt"
	"strings"

	"github.com/heyblueteam/cli/common"

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
  blue records create -w <id> -l <id> -t "Task" --due-date "2026-12-31"
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
	createDueDate      string
	createStartDate    string
	createTimezone     string
)

func init() {
	createCmd.Flags().StringVarP(&createWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	createCmd.Flags().StringVarP(&createList, "list", "l", "", "List ID to create the record in (required)")
	createCmd.Flags().StringVarP(&createTitle, "title", "t", "", "Title of the record (required)")
	createCmd.Flags().StringVarP(&createDescription, "description", "d", "", "Description of the record")
	createCmd.Flags().StringVar(&createPlacement, "placement", "", "Placement in list: TOP or BOTTOM")
	createCmd.Flags().StringVar(&createAssignees, "assignees", "", "Comma-separated assignee IDs")
	createCmd.Flags().StringVar(&createCustomFields, "custom-fields", "", "Custom field values (format: field_id1:value1;field_id2:value2)")
	createCmd.Flags().StringVar(&createDueDate, "due-date", "", "Due date (ISO format or YYYY-MM-DD)")
	createCmd.Flags().StringVar(&createStartDate, "start-date", "", "Start date (ISO format or YYYY-MM-DD)")
	createCmd.Flags().StringVar(&createTimezone, "timezone", "", "Timezone for dates (e.g., UTC, America/New_York)")
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

	// Set due date / start date if provided
	if createDueDate != "" || createStartDate != "" {
		resolvedStart := createStartDate
		resolvedDue := createDueDate
		resolvedTz := createTimezone

		if resolvedStart != "" && len(resolvedStart) == 10 {
			resolvedStart = resolvedStart + "T00:00:00Z"
			if resolvedTz == "" {
				resolvedTz = "UTC"
			}
		}
		if resolvedDue != "" && len(resolvedDue) == 10 {
			resolvedDue = resolvedDue + "T23:59:00Z"
			if resolvedTz == "" {
				resolvedTz = "UTC"
			}
		}

		var dateFields []string
		dateFields = append(dateFields, fmt.Sprintf(`todoId: "%s"`, record.ID))
		if resolvedStart != "" {
			dateFields = append(dateFields, fmt.Sprintf(`startedAt: "%s"`, resolvedStart))
		}
		if resolvedDue != "" {
			dateFields = append(dateFields, fmt.Sprintf(`duedAt: "%s"`, resolvedDue))
		}
		if resolvedTz != "" {
			dateFields = append(dateFields, fmt.Sprintf(`timezone: "%s"`, resolvedTz))
		}

		dueDateMutation := fmt.Sprintf(`
			mutation UpdateTodoDueDate {
				updateTodoDueDate(
					%s
				) { id startedAt duedAt timezone }
			}
		`, strings.Join(dateFields, "\n\t\t\t\t\t"))

		var dueDateResponse struct {
			UpdateTodoDueDate struct {
				ID string `json:"id"`
			} `json:"updateTodoDueDate"`
		}
		if err := client.ExecuteQueryWithResult(dueDateMutation, nil, &dueDateResponse); err != nil {
			return fmt.Errorf("record created but failed to set dates: %w", err)
		}
	}

	// Set custom fields if provided
	if createCustomFields != "" {
		customFieldValues, err := common.ParseCustomFieldValues(createCustomFields)
		if err != nil {
			return fmt.Errorf("failed to parse custom fields: %w", err)
		}
		if err := common.SetCustomFields(client, record.ID, customFieldValues); err != nil {
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

