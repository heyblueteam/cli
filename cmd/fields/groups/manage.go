package groups

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

// ManageFieldGroupsResponse for project mutations
type ManageFieldGroupsResponse struct {
	Project struct {
		ID         string             `json:"id"`
		Name       string             `json:"name"`
		TodoFields []common.TodoField `json:"todoFields"`
	} `json:"editProject"`
}

var manageCmd = &cobra.Command{
	Use:   "manage",
	Short: "Manage custom field groups",
	Long: `Manage custom field groups within a workspace. Supports creating, deleting,
renaming, recoloring groups, and moving fields in and out of groups.`,
	Example: `  blue fields groups manage -w <id> --action create --name "Financial" --color blue
  blue fields groups manage -w <id> --action delete --group <group-id>
  blue fields groups manage -w <id> --action rename --group <group-id> --name "New Name"
  blue fields groups manage -w <id> --action recolor --group <group-id> --color red
  blue fields groups manage -w <id> --action add-field --field <field-id>
  blue fields groups manage -w <id> --action move-in --field <field-id> --group <group-id>
  blue fields groups manage -w <id> --action move-out --field <field-id>`,
	RunE: runManage,
}

var (
	manageWorkspace string
	manageAction    string
	manageName      string
	manageColor     string
	manageGroup     string
	manageField     string
)

func init() {
	manageCmd.Flags().StringVarP(&manageWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	manageCmd.Flags().StringVar(&manageAction, "action", "", "Action to perform: create, add-field, delete, rename, recolor, move-in, move-out (required)")
	manageCmd.Flags().StringVar(&manageName, "name", "", "Group name (for create, rename)")
	manageCmd.Flags().StringVar(&manageColor, "color", "", "Group color (for create, recolor)")
	manageCmd.Flags().StringVar(&manageGroup, "group", "", "Group ID (for delete, rename, recolor, move-in)")
	manageCmd.Flags().StringVar(&manageField, "field", "", "Field ID (for add-field, move-in, move-out)")
}

func runManage(cmd *cobra.Command, args []string) error {
	if manageWorkspace == "" {
		return fmt.Errorf("workspace ID is required. Use --workspace flag")
	}

	if manageAction == "" {
		return fmt.Errorf("action is required. Use --action flag (create, add-field, delete, rename, recolor, move-in, move-out)")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)
	client.SetProject(manageWorkspace)

	switch strings.ToLower(manageAction) {
	case "create":
		if manageName == "" {
			return fmt.Errorf("name is required for create action. Use --name flag")
		}
		if manageColor == "" {
			manageColor = "default"
		}
		return actionCreate(client, manageWorkspace, manageName, manageColor)

	case "add-field":
		if manageField == "" {
			return fmt.Errorf("field ID is required for add-field action. Use --field flag")
		}
		return actionAddField(client, manageWorkspace, manageField)

	case "delete":
		if manageGroup == "" {
			return fmt.Errorf("group ID is required for delete action. Use --group flag")
		}
		return actionDelete(client, manageWorkspace, manageGroup)

	case "rename":
		if manageGroup == "" || manageName == "" {
			return fmt.Errorf("group ID and name are required for rename action. Use --group and --name flags")
		}
		return actionRename(client, manageWorkspace, manageGroup, manageName)

	case "recolor":
		if manageGroup == "" || manageColor == "" {
			return fmt.Errorf("group ID and color are required for recolor action. Use --group and --color flags")
		}
		return actionRecolor(client, manageWorkspace, manageGroup, manageColor)

	case "move-in":
		if manageField == "" || manageGroup == "" {
			return fmt.Errorf("field ID and group ID are required for move-in action. Use --field and --group flags")
		}
		return actionMoveIn(client, manageWorkspace, manageField, manageGroup)

	case "move-out":
		if manageField == "" {
			return fmt.Errorf("field ID is required for move-out action. Use --field flag")
		}
		return actionMoveOut(client, manageWorkspace, manageField)

	default:
		return fmt.Errorf("invalid action '%s'. Valid actions: create, add-field, delete, rename, recolor, move-in, move-out", manageAction)
	}
}

// generateGroupID creates a unique ID for a new group
func generateGroupID() string {
	b := make([]byte, 12)
	rand.Read(b)
	return "grp_" + hex.EncodeToString(b)
}

// fetchProjectTodoFields retrieves the current todoFields configuration
func fetchProjectTodoFields(client *common.Client, projectID string) ([]common.TodoField, error) {
	infoQuery := fmt.Sprintf(`
		query GetProjectInfo {
			project(id: "%s") {
				id
				slug
			}
		}
	`, projectID)

	var projectInfo struct {
		Project struct {
			ID   string `json:"id"`
			Slug string `json:"slug"`
		} `json:"project"`
	}

	if err := client.ExecuteQueryWithResult(infoQuery, nil, &projectInfo); err != nil {
		return nil, err
	}

	queryIdentifier := projectInfo.Project.Slug
	if projectID == projectInfo.Project.Slug {
		queryIdentifier = projectInfo.Project.ID
	}

	query := fmt.Sprintf(`
		query GetProjectTodoFields {
			project(id: "%s") {
				id
				todoFields {
					type
					customFieldId
					name
					color
					todoFields {
						type
						customFieldId
						name
						color
					}
				}
			}
		}
	`, queryIdentifier)

	var response ProjectWithTodoFieldsResponse
	if err := client.ExecuteQueryWithResult(query, nil, &response); err != nil {
		return nil, err
	}

	return response.Project.TodoFields, nil
}

// updateProjectTodoFields updates the project's todoFields configuration
func updateProjectTodoFields(client *common.Client, projectID string, todoFields []common.TodoFieldInput) error {
	mutation := `
		mutation EditProject($projectId: String!, $todoFields: [TodoFieldInput]) {
			editProject(input: {
				projectId: $projectId
				todoFields: $todoFields
			}) {
				id
				name
				todoFields {
					type
					customFieldId
					name
					color
					todoFields {
						type
						customFieldId
					}
				}
			}
		}
	`

	variables := map[string]interface{}{
		"projectId":  projectID,
		"todoFields": todoFields,
	}

	var response ManageFieldGroupsResponse
	return client.ExecuteQueryWithResult(mutation, variables, &response)
}

// convertToInput converts TodoField to TodoFieldInput
func convertToInput(fields []common.TodoField) []common.TodoFieldInput {
	inputs := make([]common.TodoFieldInput, len(fields))
	for i, field := range fields {
		inputs[i] = common.TodoFieldInput{
			Type:          field.Type,
			CustomFieldID: field.CustomFieldID,
			Name:          field.Name,
			Color:         field.Color,
		}

		if len(field.TodoFields) > 0 {
			inputs[i].TodoFields = convertToInput(field.TodoFields)
		}
	}
	return inputs
}

// findFieldIndex finds a field by ID in the todoFields array
func findFieldIndex(fields []common.TodoField, fieldID string) (int, int) {
	for i, field := range fields {
		if field.CustomFieldID != nil && *field.CustomFieldID == fieldID {
			return i, -1
		}
		if field.Type == "CUSTOM_FIELD_GROUP" {
			for j, nestedField := range field.TodoFields {
				if nestedField.CustomFieldID != nil && *nestedField.CustomFieldID == fieldID {
					return i, j
				}
			}
		}
	}
	return -1, -1
}

// findGroupIndex finds a group by ID
func findGroupIndex(fields []common.TodoField, groupID string) int {
	for i, field := range fields {
		if field.Type == "CUSTOM_FIELD_GROUP" && field.CustomFieldID != nil && *field.CustomFieldID == groupID {
			return i
		}
	}
	return -1
}

// actionCreate creates a new group
func actionCreate(client *common.Client, projectID, name, color string) error {
	fields, err := fetchProjectTodoFields(client, projectID)
	if err != nil {
		return err
	}

	groupID := generateGroupID()
	newGroup := common.TodoField{
		Type:          "CUSTOM_FIELD_GROUP",
		CustomFieldID: &groupID,
		Name:          &name,
		Color:         &color,
		TodoFields:    []common.TodoField{},
	}

	fields = append(fields, newGroup)
	inputs := convertToInput(fields)

	if err := updateProjectTodoFields(client, projectID, inputs); err != nil {
		return err
	}

	fmt.Printf("Group '%s' created successfully!\n", name)
	fmt.Printf("Group ID: %s\n", groupID)
	return nil
}

// actionDelete deletes a group and moves its fields to root level
func actionDelete(client *common.Client, projectID, groupID string) error {
	fields, err := fetchProjectTodoFields(client, projectID)
	if err != nil {
		return err
	}

	groupIdx := findGroupIndex(fields, groupID)
	if groupIdx == -1 {
		return fmt.Errorf("group with ID '%s' not found", groupID)
	}

	group := fields[groupIdx]
	nestedFields := group.TodoFields

	fields = append(fields[:groupIdx], fields[groupIdx+1:]...)

	for _, nestedField := range nestedFields {
		fields = append(fields, nestedField)
	}

	inputs := convertToInput(fields)
	if err := updateProjectTodoFields(client, projectID, inputs); err != nil {
		return err
	}

	fmt.Printf("Group deleted successfully. %d field(s) moved to root level.\n", len(nestedFields))
	return nil
}

// actionRename renames a group
func actionRename(client *common.Client, projectID, groupID, newName string) error {
	fields, err := fetchProjectTodoFields(client, projectID)
	if err != nil {
		return err
	}

	groupIdx := findGroupIndex(fields, groupID)
	if groupIdx == -1 {
		return fmt.Errorf("group with ID '%s' not found", groupID)
	}

	fields[groupIdx].Name = &newName

	inputs := convertToInput(fields)
	if err := updateProjectTodoFields(client, projectID, inputs); err != nil {
		return err
	}

	fmt.Printf("Group renamed to '%s'\n", newName)
	return nil
}

// actionRecolor changes a group's color
func actionRecolor(client *common.Client, projectID, groupID, newColor string) error {
	fields, err := fetchProjectTodoFields(client, projectID)
	if err != nil {
		return err
	}

	groupIdx := findGroupIndex(fields, groupID)
	if groupIdx == -1 {
		return fmt.Errorf("group with ID '%s' not found", groupID)
	}

	fields[groupIdx].Color = &newColor

	inputs := convertToInput(fields)
	if err := updateProjectTodoFields(client, projectID, inputs); err != nil {
		return err
	}

	fmt.Printf("Group color changed to '%s'\n", newColor)
	return nil
}

// actionAddField adds a custom field to the root level of todoFields
func actionAddField(client *common.Client, projectID, fieldID string) error {
	fields, err := fetchProjectTodoFields(client, projectID)
	if err != nil {
		return err
	}

	fieldIdx, nestedIdx := findFieldIndex(fields, fieldID)
	if fieldIdx != -1 {
		if nestedIdx == -1 {
			return fmt.Errorf("field already exists at root level")
		}
		return fmt.Errorf("field already exists in a group")
	}

	newField := common.TodoField{
		Type:          "CUSTOM_FIELD",
		CustomFieldID: &fieldID,
	}

	fields = append(fields, newField)
	inputs := convertToInput(fields)

	if err := updateProjectTodoFields(client, projectID, inputs); err != nil {
		return err
	}

	fmt.Printf("Custom field added to root level successfully!\n")
	return nil
}

// actionMoveIn moves a field into a group
func actionMoveIn(client *common.Client, projectID, fieldID, groupID string) error {
	fields, err := fetchProjectTodoFields(client, projectID)
	if err != nil {
		return err
	}

	fieldIdx, nestedIdx := findFieldIndex(fields, fieldID)
	if fieldIdx == -1 {
		return fmt.Errorf("field with ID '%s' not found", fieldID)
	}

	groupIdx := findGroupIndex(fields, groupID)
	if groupIdx == -1 {
		return fmt.Errorf("group with ID '%s' not found", groupID)
	}

	var fieldToMove common.TodoField

	if nestedIdx == -1 {
		fieldToMove = fields[fieldIdx]
		fields = append(fields[:fieldIdx], fields[fieldIdx+1:]...)
	} else {
		fieldToMove = fields[fieldIdx].TodoFields[nestedIdx]
		fields[fieldIdx].TodoFields = append(
			fields[fieldIdx].TodoFields[:nestedIdx],
			fields[fieldIdx].TodoFields[nestedIdx+1:]...,
		)
	}

	if nestedIdx == -1 && fieldIdx < groupIdx {
		groupIdx--
	}

	fields[groupIdx].TodoFields = append(fields[groupIdx].TodoFields, fieldToMove)

	inputs := convertToInput(fields)
	if err := updateProjectTodoFields(client, projectID, inputs); err != nil {
		return err
	}

	fmt.Printf("Field moved into group successfully!\n")
	return nil
}

// actionMoveOut moves a field out of a group to root level
func actionMoveOut(client *common.Client, projectID, fieldID string) error {
	fields, err := fetchProjectTodoFields(client, projectID)
	if err != nil {
		return err
	}

	fieldIdx, nestedIdx := findFieldIndex(fields, fieldID)
	if fieldIdx == -1 {
		return fmt.Errorf("field with ID '%s' not found", fieldID)
	}

	if nestedIdx == -1 {
		return fmt.Errorf("field is already at root level")
	}

	fieldToMove := fields[fieldIdx].TodoFields[nestedIdx]
	fields[fieldIdx].TodoFields = append(
		fields[fieldIdx].TodoFields[:nestedIdx],
		fields[fieldIdx].TodoFields[nestedIdx+1:]...,
	)

	fields = append(fields, fieldToMove)

	inputs := convertToInput(fields)
	if err := updateProjectTodoFields(client, projectID, inputs); err != nil {
		return err
	}

	fmt.Printf("Field moved to root level successfully!\n")
	return nil
}
