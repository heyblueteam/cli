package records

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

type UpdateRecordResponse struct {
	EditTodo struct {
		ID        string  `json:"id"`
		Title     string  `json:"title"`
		Position  float64 `json:"position"`
		Color     string  `json:"color,omitempty"`
		StartedAt string  `json:"startedAt,omitempty"`
		DuedAt    string  `json:"duedAt,omitempty"`
		TodoList  struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"todoList"`
		Users []struct {
			ID        string `json:"id"`
			FirstName string `json:"firstName"`
			LastName  string `json:"lastName"`
		} `json:"users"`
		Tags []struct {
			ID    string `json:"id"`
			Title string `json:"title"`
			Color string `json:"color"`
		} `json:"tags"`
	} `json:"editTodo"`
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a record",
	Long:  "Update record properties, assignees, tags, and custom fields.",
	Example: `  blue records update --record <id> --title "New Title"
  blue records update -r <id> -w <id> --assignees "user1,user2"
  blue records update -r <id> -w <id> --custom-fields "cf123:value;cf456:42"
  blue records update -r <id> --due-date "2026-12-31"`,
	RunE: runUpdate,
}

var (
	updateRecord       string
	updateWorkspace    string
	updateTitle        string
	updateDescription  string
	updateHTML         string
	updatePosition     string
	updateListID       string
	updateStartDate    string
	updateDueDate      string
	updateTimezone     string
	updateColor        string
	updateCover        string
	updateAssignees    string
	updateTagIds       string
	updateTagTitles    string
	updateCustomFields string
	updateSimple       bool
)

func init() {
	updateCmd.Flags().StringVarP(&updateRecord, "record", "r", "", "Record ID to update (required)")
	updateCmd.Flags().StringVarP(&updateWorkspace, "workspace", "w", "", "Workspace ID (required for tags, custom fields, assignees)")
	updateCmd.Flags().StringVarP(&updateTitle, "title", "t", "", "New title")
	updateCmd.Flags().StringVarP(&updateDescription, "description", "d", "", "New description")
	updateCmd.Flags().StringVar(&updateHTML, "html", "", "New HTML content")
	updateCmd.Flags().StringVar(&updatePosition, "position", "", "New position (float)")
	updateCmd.Flags().StringVarP(&updateListID, "list", "l", "", "Move to different list ID")
	updateCmd.Flags().StringVar(&updateStartDate, "start-date", "", "Start date (ISO format or YYYY-MM-DD)")
	updateCmd.Flags().StringVar(&updateDueDate, "due-date", "", "Due date (ISO format or YYYY-MM-DD)")
	updateCmd.Flags().StringVar(&updateTimezone, "timezone", "", "Timezone for dates (e.g., UTC, America/New_York)")
	updateCmd.Flags().StringVar(&updateColor, "color", "", "Record color")
	updateCmd.Flags().StringVar(&updateCover, "cover", "", "Cover image")
	updateCmd.Flags().StringVar(&updateAssignees, "assignees", "", "Comma-separated assignee IDs")
	updateCmd.Flags().StringVar(&updateTagIds, "tag-ids", "", "Comma-separated tag IDs")
	updateCmd.Flags().StringVar(&updateTagTitles, "tag-titles", "", "Comma-separated tag titles")
	updateCmd.Flags().StringVar(&updateCustomFields, "custom-fields", "", "Custom field values (format: field_id1:value1;field_id2:value2)")
	updateCmd.Flags().BoolVarP(&updateSimple, "simple", "s", false, "Simple output format")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if updateRecord == "" {
		return fmt.Errorf("record ID is required. Use --record flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := common.NewClient(config)

	if updateWorkspace != "" {
		client.SetProject(updateWorkspace)
	}

	// Build editTodo fields
	var fields []string

	if updateListID != "" {
		fields = append(fields, fmt.Sprintf(`todoListId: "%s"`, updateListID))
	}
	if updatePosition != "" {
		pos, err := strconv.ParseFloat(updatePosition, 64)
		if err != nil {
			return fmt.Errorf("invalid position value: %s", updatePosition)
		}
		fields = append(fields, fmt.Sprintf(`position: %g`, pos))
	}
	if updateTitle != "" {
		fields = append(fields, fmt.Sprintf(`title: "%s"`, strings.ReplaceAll(updateTitle, `"`, `\"`)))
	}
	if updateHTML != "" {
		fields = append(fields, fmt.Sprintf(`html: "%s"`, strings.ReplaceAll(updateHTML, `"`, `\"`)))
	}
	if updateDescription != "" {
		fields = append(fields, fmt.Sprintf(`text: "%s"`, strings.ReplaceAll(updateDescription, `"`, `\"`)))
	}
	if updateColor != "" {
		fields = append(fields, fmt.Sprintf(`color: "%s"`, updateColor))
	}
	if updateCover != "" {
		fields = append(fields, fmt.Sprintf(`cover: "%s"`, updateCover))
	}

	// Handle dates with timezone
	resolvedStart := updateStartDate
	resolvedDue := updateDueDate
	resolvedTz := updateTimezone

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

	// If timezone is provided with dates, use dedicated mutation
	if resolvedTz != "" && (resolvedStart != "" || resolvedDue != "") {
		dueDateMutation := fmt.Sprintf(`
			mutation UpdateTodoDueDate {
				updateTodoDueDate(
					todoId: "%s"
					startedAt: "%s"
					duedAt: "%s"
					timezone: "%s"
				) { id startedAt duedAt timezone }
			}
		`, updateRecord, resolvedStart, resolvedDue, resolvedTz)

		var dueDateResponse struct {
			UpdateTodoDueDate struct {
				ID string `json:"id"`
			} `json:"updateTodoDueDate"`
		}
		if err := client.ExecuteQueryWithResult(dueDateMutation, nil, &dueDateResponse); err != nil {
			return fmt.Errorf("failed to update dates: %w", err)
		}
	} else {
		if resolvedStart != "" {
			fields = append(fields, fmt.Sprintf(`startedAt: "%s"`, resolvedStart))
		}
		if resolvedDue != "" {
			fields = append(fields, fmt.Sprintf(`duedAt: "%s"`, resolvedDue))
		}
	}

	// Execute editTodo mutation
	mutation := fmt.Sprintf(`
		mutation EditTodo {
			editTodo(input: {
				todoId: "%s"
				%s
			}) {
				id title position color startedAt duedAt
				todoList { id title }
				users { id firstName lastName }
				tags { id title color }
			}
		}
	`, updateRecord, strings.Join(fields, "\n\t\t\t\t"))

	var response UpdateRecordResponse
	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}

	// Check if workspace required for advanced ops
	hasAdvanced := updateAssignees != "" || updateTagIds != "" || updateTagTitles != "" || updateCustomFields != ""
	if hasAdvanced && updateWorkspace == "" {
		return fmt.Errorf("workspace is required for assignee, tag, or custom field updates. Use --workspace flag")
	}

	// Set assignees
	if updateAssignees != "" {
		assigneeList := strings.Split(updateAssignees, ",")
		for i, a := range assigneeList {
			assigneeList[i] = strings.TrimSpace(a)
		}
		var assigneeStrings []string
		for _, a := range assigneeList {
			assigneeStrings = append(assigneeStrings, fmt.Sprintf(`"%s"`, a))
		}
		assigneeMutation := fmt.Sprintf(`
			mutation SetTodoAssignees {
				setTodoAssignees(input: {
					todoId: "%s"
					assigneeIds: [%s]
				}) { success }
			}
		`, updateRecord, strings.Join(assigneeStrings, ", "))

		var assigneeResp struct {
			SetTodoAssignees struct {
				Success bool `json:"success"`
			} `json:"setTodoAssignees"`
		}
		if err := client.ExecuteQueryWithResult(assigneeMutation, nil, &assigneeResp); err != nil {
			return fmt.Errorf("failed to update assignees: %w", err)
		}
	}

	// Set tags
	if updateTagIds != "" || updateTagTitles != "" {
		var tagFields []string
		if updateTagIds != "" {
			tagList := strings.Split(updateTagIds, ",")
			var tagStrings []string
			for _, t := range tagList {
				tagStrings = append(tagStrings, fmt.Sprintf(`"%s"`, strings.TrimSpace(t)))
			}
			tagFields = append(tagFields, fmt.Sprintf(`tagIds: [%s]`, strings.Join(tagStrings, ", ")))
		}
		if updateTagTitles != "" {
			titleList := strings.Split(updateTagTitles, ",")
			var titleStrings []string
			for _, t := range titleList {
				titleStrings = append(titleStrings, fmt.Sprintf(`"%s"`, strings.TrimSpace(t)))
			}
			tagFields = append(tagFields, fmt.Sprintf(`tagTitles: [%s]`, strings.Join(titleStrings, ", ")))
		}

		tagMutation := fmt.Sprintf(`
			mutation SetTodoTags {
				setTodoTags(input: {
					todoId: "%s"
					%s
				})
			}
		`, updateRecord, strings.Join(tagFields, "\n\t\t\t\t"))

		var tagResp struct {
			SetTodoTags bool `json:"setTodoTags"`
		}
		if err := client.ExecuteQueryWithResult(tagMutation, nil, &tagResp); err != nil {
			return fmt.Errorf("failed to update tags: %w", err)
		}
	}

	// Set custom fields
	if updateCustomFields != "" {
		cfValues, err := parseUpdateCustomFieldValues(updateCustomFields)
		if err != nil {
			return fmt.Errorf("failed to parse custom fields: %w", err)
		}
		if err := executeSetCustomFields(client, updateRecord, cfValues); err != nil {
			return fmt.Errorf("failed to update custom fields: %w", err)
		}
	}

	record := response.EditTodo

	if updateSimple {
		fmt.Printf("Updated record: %s (ID: %s)\n", record.Title, record.ID)
	} else {
		fmt.Printf("=== Record Updated Successfully ===\n")
		fmt.Printf("ID: %s\n", record.ID)
		fmt.Printf("Title: %s\n", record.Title)
		fmt.Printf("Position: %.0f\n", record.Position)
		fmt.Printf("List: %s (%s)\n", record.TodoList.Title, record.TodoList.ID)
		if record.Color != "" {
			fmt.Printf("Color: %s\n", record.Color)
		}
		if record.DuedAt != "" {
			fmt.Printf("Due At: %s\n", record.DuedAt)
		}
		if len(record.Users) > 0 {
			fmt.Printf("Assignees: %d\n", len(record.Users))
		}
		if len(record.Tags) > 0 {
			fmt.Printf("Tags: %d\n", len(record.Tags))
		}
	}

	return nil
}

// parseUpdateCustomFieldValues handles typed parsing (numbers, booleans, arrays)
func parseUpdateCustomFieldValues(customFieldsStr string) ([]common.CustomFieldValue, error) {
	var values []common.CustomFieldValue
	fieldPairs := strings.Split(customFieldsStr, ";")

	for _, pair := range fieldPairs {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid custom field format: %s (expected field_id:value)", pair)
		}

		fieldID := strings.TrimSpace(parts[0])
		valueStr := strings.TrimSpace(parts[1])

		var value interface{}
		if valueStr == "true" || valueStr == "false" {
			value, _ = strconv.ParseBool(valueStr)
		} else if floatVal, err := strconv.ParseFloat(valueStr, 64); err == nil {
			value = floatVal
		} else if strings.HasPrefix(valueStr, "[") && strings.HasSuffix(valueStr, "]") {
			var arr []string
			if err := json.Unmarshal([]byte(valueStr), &arr); err == nil {
				value = arr
			} else {
				value = valueStr
			}
		} else {
			value = valueStr
		}

		values = append(values, common.CustomFieldValue{
			CustomFieldID: fieldID,
			Value:         value,
		})
	}

	return values, nil
}
