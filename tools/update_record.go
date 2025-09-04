package tools

import (
	"encoding/json"
	"flag"
	"fmt"
	"strconv"
	"strings"

	"demo-builder/common"
)

// UpdateRecordInput combines all possible update operations
type UpdateRecordInput struct {
	TodoID      string                  `json:"todoId"`
	TodoListID  string                  `json:"todoListId,omitempty"`
	Position    *float64                `json:"position,omitempty"`
	Title       string                  `json:"title,omitempty"`
	HTML        string                  `json:"html,omitempty"`
	Text        string                  `json:"text,omitempty"`
	StartedAt   string                  `json:"startedAt,omitempty"`
	DuedAt      string                  `json:"duedAt,omitempty"`
	Color       string                  `json:"color,omitempty"`
	Cover       string                  `json:"cover,omitempty"`
	AssigneeIds []string                `json:"assigneeIds,omitempty"`
	TagIds      []string                `json:"tagIds,omitempty"`
	TagTitles   []string                `json:"tagTitles,omitempty"`
	CustomFields []common.CustomFieldValue `json:"customFields,omitempty"`
}

// Response structures
type UpdateRecordResponse struct {
	EditTodo struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Position float64 `json:"position"`
		Color    string `json:"color,omitempty"`
		StartedAt string `json:"startedAt,omitempty"`
		DuedAt   string `json:"duedAt,omitempty"`
		TodoList struct {
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

type MutationResultResponse struct {
	SetTodoAssignees struct {
		Success     bool   `json:"success"`
		OperationID string `json:"operationId"`
	} `json:"setTodoAssignees"`
}

type SetTagsResponse struct {
	SetTodoTags bool `json:"setTodoTags"`
}


// parseCustomFieldValues parses the custom field values from command line arguments
func parseUpdateCustomFieldValues(customFieldsStr string) ([]common.CustomFieldValue, error) {
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

		// Try to parse as different types
		var value interface{}
		
		// Try boolean first
		if valueStr == "true" || valueStr == "false" {
			value, _ = strconv.ParseBool(valueStr)
		} else if floatVal, err := strconv.ParseFloat(valueStr, 64); err == nil {
			// Try as number
			value = floatVal
		} else if strings.HasPrefix(valueStr, "[") && strings.HasSuffix(valueStr, "]") {
			// Try as array
			var arr []string
			if err := json.Unmarshal([]byte(valueStr), &arr); err == nil {
				value = arr
			} else {
				value = valueStr // fallback to string
			}
		} else {
			// Default to string
			value = valueStr
		}

		customFieldValues = append(customFieldValues, common.CustomFieldValue{
			CustomFieldID: fieldID,
			Value:         value,
		})
	}

	return customFieldValues, nil
}


// executeEditTodo performs the main record update
func executeEditTodo(client *common.Client, input UpdateRecordInput) (*UpdateRecordResponse, error) {
	// Build optional fields
	var fields []string

	if input.TodoListID != "" {
		fields = append(fields, fmt.Sprintf(`todoListId: "%s"`, input.TodoListID))
	}
	if input.Position != nil {
		fields = append(fields, fmt.Sprintf(`position: %g`, *input.Position))
	}
	if input.Title != "" {
		fields = append(fields, fmt.Sprintf(`title: "%s"`, strings.ReplaceAll(input.Title, `"`, `\"`)))
	}
	if input.HTML != "" {
		fields = append(fields, fmt.Sprintf(`html: "%s"`, strings.ReplaceAll(input.HTML, `"`, `\"`)))
	}
	if input.Text != "" {
		fields = append(fields, fmt.Sprintf(`text: "%s"`, strings.ReplaceAll(input.Text, `"`, `\"`)))
	}
	if input.StartedAt != "" {
		fields = append(fields, fmt.Sprintf(`startedAt: "%s"`, input.StartedAt))
	}
	if input.DuedAt != "" {
		fields = append(fields, fmt.Sprintf(`duedAt: "%s"`, input.DuedAt))
	}
	if input.Color != "" {
		fields = append(fields, fmt.Sprintf(`color: "%s"`, input.Color))
	}
	if input.Cover != "" {
		fields = append(fields, fmt.Sprintf(`cover: "%s"`, input.Cover))
	}

	fieldsStr := ""
	if len(fields) > 0 {
		fieldsStr = strings.Join(fields, "\n\t\t\t")
	}

	mutation := fmt.Sprintf(`
		mutation EditTodo {
			editTodo(input: {
				todoId: "%s"
				%s
			}) {
				id
				title
				position
				color
				startedAt
				duedAt
				todoList {
					id
					title
				}
				users {
					id
					firstName
					lastName
				}
				tags {
					id
					title
					color
				}
			}
		}
	`, input.TodoID, fieldsStr)

	var response UpdateRecordResponse
	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// executeSetAssignees updates record assignees
func executeSetAssignees(client *common.Client, todoID string, assigneeIds []string) error {
	if len(assigneeIds) == 0 {
		return nil
	}

	var assigneeStrings []string
	for _, assigneeID := range assigneeIds {
		assigneeStrings = append(assigneeStrings, fmt.Sprintf(`"%s"`, assigneeID))
	}

	mutation := fmt.Sprintf(`
		mutation SetTodoAssignees {
			setTodoAssignees(input: {
				todoId: "%s"
				assigneeIds: [%s]
			}) {
				success
				operationId
			}
		}
	`, todoID, strings.Join(assigneeStrings, ", "))

	var response MutationResultResponse
	return client.ExecuteQueryWithResult(mutation, nil, &response)
}

// executeSetTags updates record tags
func executeSetTags(client *common.Client, todoID string, tagIds []string, tagTitles []string) error {
	if len(tagIds) == 0 && len(tagTitles) == 0 {
		return nil
	}

	var fields []string

	if len(tagIds) > 0 {
		var tagIdStrings []string
		for _, tagID := range tagIds {
			tagIdStrings = append(tagIdStrings, fmt.Sprintf(`"%s"`, tagID))
		}
		fields = append(fields, fmt.Sprintf(`tagIds: [%s]`, strings.Join(tagIdStrings, ", ")))
	}

	if len(tagTitles) > 0 {
		var tagTitleStrings []string
		for _, title := range tagTitles {
			tagTitleStrings = append(tagTitleStrings, fmt.Sprintf(`"%s"`, strings.ReplaceAll(title, `"`, `\"`)))
		}
		fields = append(fields, fmt.Sprintf(`tagTitles: [%s]`, strings.Join(tagTitleStrings, ", ")))
	}

	mutation := fmt.Sprintf(`
		mutation SetTodoTags {
			setTodoTags(input: {
				todoId: "%s"
				%s
			})
		}
	`, todoID, strings.Join(fields, "\n\t\t\t"))

	var response SetTagsResponse
	return client.ExecuteQueryWithResult(mutation, nil, &response)
}

// executeSetCustomFields updates custom field values

// getProjectIDFromRecord retrieves project ID from a record
func getProjectIDFromRecord(client *common.Client, todoID string) (string, error) {
	query := fmt.Sprintf(`
		query GetTodo {
			todo(id: "%s") {
				id
				todoList {
					project {
						id
					}
				}
			}
		}
	`, todoID)

	var response struct {
		Todo struct {
			ID       string `json:"id"`
			TodoList struct {
				Project struct {
					ID string `json:"id"`
				} `json:"project"`
			} `json:"todoList"`
		} `json:"todo"`
	}

	if err := client.ExecuteQueryWithResult(query, nil, &response); err != nil {
		return "", fmt.Errorf("failed to get record details: %v", err)
	}

	if response.Todo.ID == "" {
		return "", fmt.Errorf("record not found: %s", todoID)
	}

	return response.Todo.TodoList.Project.ID, nil
}

func RunUpdateRecord(args []string) error {
	fs := flag.NewFlagSet("update-record", flag.ExitOnError)
	
	// Required
	todoID := fs.String("record", "", "Record ID to update (required)")
	
	// Project context (required for tags, custom fields, assignees)
	projectID := fs.String("project", "", "Project ID or slug (required for tag, custom field, or assignee updates)")
	
	// Basic fields
	title := fs.String("title", "", "New title")
	description := fs.String("description", "", "New description (text)")
	htmlContent := fs.String("html", "", "New HTML content")
	position := fs.String("position", "", "New position (float)")
	listID := fs.String("list", "", "Move to different list ID")
	
	// Date fields
	startDate := fs.String("start-date", "", "Start date (ISO format)")
	dueDate := fs.String("due-date", "", "Due date (ISO format)")
	
	// Visual fields
	color := fs.String("color", "", "Record color")
	cover := fs.String("cover", "", "Cover image")
	
	// Relationships
	assignees := fs.String("assignees", "", "Comma-separated assignee IDs")
	tagIds := fs.String("tag-ids", "", "Comma-separated tag IDs")
	tagTitles := fs.String("tag-titles", "", "Comma-separated tag titles")
	
	// Custom fields
	customFields := fs.String("custom-fields", "", "Custom field values (format: field_id1:value1;field_id2:value2)")
	
	// Options
	simple := fs.Bool("simple", false, "Simple output format")

	fs.Parse(args)

	if *todoID == "" {
		fmt.Println("Error: -record flag is required")
		fmt.Println("\nUsage:")
		fmt.Println("  go run . update-record -record RECORD_ID [flags]")
		fmt.Println("\nFlags:")
		fs.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  # Update title and description")
		fmt.Println("  go run . update-record -record rec123 -title \"New Title\" -description \"New description\"")
		fmt.Println("")
		fmt.Println("  # Update assignees and tags (requires project ID)")
		fmt.Println("  go run . update-record -record rec123 -project proj456 -assignees \"user1,user2\" -tag-titles \"Bug,Priority\"")
		fmt.Println("")
		fmt.Println("  # Update custom fields (requires project ID)")
		fmt.Println("  go run . update-record -record rec123 -project proj456 -custom-fields \"cf123:High Priority;cf456:42.5;cf789:true\"")
		fmt.Println("")
		fmt.Println("  # Move to different list with new due date")
		fmt.Println("  go run . update-record -record rec123 -list list456 -due-date \"2024-12-31T23:59:59Z\"")
		return fmt.Errorf("required flags missing")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	client := common.NewClient(config)

	// Set project context if provided
	if *projectID != "" {
		client.SetProject(*projectID)
	}

	// Parse position if provided
	var positionFloat *float64
	if *position != "" {
		if pos, err := strconv.ParseFloat(*position, 64); err == nil {
			positionFloat = &pos
		} else {
			return fmt.Errorf("invalid position value: %s", *position)
		}
	}

	// Parse assignees
	var assigneeIds []string
	if *assignees != "" {
		assigneeList := strings.Split(*assignees, ",")
		for _, assignee := range assigneeList {
			assigneeIds = append(assigneeIds, strings.TrimSpace(assignee))
		}
	}

	// Parse tag IDs
	var tagIdList []string
	if *tagIds != "" {
		tagList := strings.Split(*tagIds, ",")
		for _, tag := range tagList {
			tagIdList = append(tagIdList, strings.TrimSpace(tag))
		}
	}

	// Parse tag titles
	var tagTitleList []string
	if *tagTitles != "" {
		titleList := strings.Split(*tagTitles, ",")
		for _, title := range titleList {
			tagTitleList = append(tagTitleList, strings.TrimSpace(title))
		}
	}

	// Parse custom fields
	var customFieldValues []common.CustomFieldValue
	if *customFields != "" {
		customFieldValues, err = parseUpdateCustomFieldValues(*customFields)
		if err != nil {
			return fmt.Errorf("failed to parse custom fields: %v", err)
		}
	}

	// Build update input
	input := UpdateRecordInput{
		TodoID:       *todoID,
		TodoListID:   *listID,
		Position:     positionFloat,
		Title:        *title,
		HTML:         *htmlContent,
		Text:         *description,
		StartedAt:    *startDate,
		DuedAt:       *dueDate,
		Color:        *color,
		Cover:        *cover,
		AssigneeIds:  assigneeIds,
		TagIds:       tagIdList,
		TagTitles:    tagTitleList,
		CustomFields: customFieldValues,
	}

	// Execute the main record update (editTodo)
	response, err := executeEditTodo(client, input)
	if err != nil {
		return fmt.Errorf("failed to update record: %v", err)
	}

	// Check if project context is required for advanced operations
	requiresProject := len(input.AssigneeIds) > 0 || len(input.TagIds) > 0 || len(input.TagTitles) > 0 || len(input.CustomFields) > 0
	if requiresProject && *projectID == "" {
		return fmt.Errorf("project ID is required for assignee, tag, or custom field updates. Use -project flag")
	}

	// Execute assignees update if needed
	if len(input.AssigneeIds) > 0 {
		if err := executeSetAssignees(client, input.TodoID, input.AssigneeIds); err != nil {
			return fmt.Errorf("failed to update assignees: %v", err)
		}
	}

	// Execute tags update if needed
	if len(input.TagIds) > 0 || len(input.TagTitles) > 0 {
		if err := executeSetTags(client, input.TodoID, input.TagIds, input.TagTitles); err != nil {
			return fmt.Errorf("failed to update tags: %v", err)
		}
	}

	// Execute custom fields update if needed
	if len(input.CustomFields) > 0 {
		if err := executeSetCustomFields(client, input.TodoID, input.CustomFields); err != nil {
			return fmt.Errorf("failed to update custom fields: %v", err)
		}
	}

	record := response.EditTodo

	if *simple {
		fmt.Printf("Updated record: %s (ID: %s)\n", record.Title, record.ID)
		if len(input.AssigneeIds) > 0 {
			fmt.Printf("Assignees updated: %d\n", len(record.Users))
		}
		if len(input.TagIds) > 0 || len(input.TagTitles) > 0 {
			fmt.Printf("Tags updated: %d\n", len(record.Tags))
		}
		if len(input.CustomFields) > 0 {
			fmt.Printf("Custom fields updated: %d\n", len(input.CustomFields))
		}
	} else {
		fmt.Printf("=== Record Updated Successfully ===\n")
		fmt.Printf("ID: %s\n", record.ID)
		fmt.Printf("Title: %s\n", record.Title)
		fmt.Printf("Position: %.0f\n", record.Position)
		fmt.Printf("List: %s (%s)\n", record.TodoList.Title, record.TodoList.ID)

		if record.Color != "" {
			fmt.Printf("Color: %s\n", record.Color)
		}
		if record.StartedAt != "" {
			fmt.Printf("Started At: %s\n", record.StartedAt)
		}
		if record.DuedAt != "" {
			fmt.Printf("Due At: %s\n", record.DuedAt)
		}

		if len(record.Users) > 0 {
			fmt.Printf("\n=== Assignees (%d) ===\n", len(record.Users))
			for _, user := range record.Users {
				fmt.Printf("%s %s (%s)\n", user.FirstName, user.LastName, user.ID)
			}
		}

		if len(record.Tags) > 0 {
			fmt.Printf("\n=== Tags (%d) ===\n", len(record.Tags))
			for _, tag := range record.Tags {
				color := tag.Color
				if color == "" {
					color = "default"
				}
				fmt.Printf("%s (%s) [%s]\n", tag.Title, tag.ID, color)
			}
		}

		if len(input.CustomFields) > 0 {
			fmt.Printf("\n=== Custom Fields Updated (%d) ===\n", len(input.CustomFields))
			for _, cf := range input.CustomFields {
				fmt.Printf("%s: %v\n", cf.CustomFieldID, cf.Value)
			}
		}
	}

	return nil
}