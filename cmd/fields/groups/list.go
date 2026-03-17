package groups

import (
	"fmt"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

// ProjectWithTodoFieldsResponse for fetching project with todoFields
type ProjectWithTodoFieldsResponse struct {
	Project struct {
		ID         string             `json:"id"`
		Name       string             `json:"name"`
		TodoFields []common.TodoField `json:"todoFields"`
	} `json:"project"`
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List custom field groups",
	Long:  "Display custom field groups and their organization within a workspace.",
	Example: `  blue fields groups list --workspace <id>
  blue fields groups list -w <slug>`,
	RunE: runList,
}

var (
	listWorkspace string
)

func init() {
	listCmd.Flags().StringVarP(&listWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
}

func runList(cmd *cobra.Command, args []string) error {
	if listWorkspace == "" {
		return fmt.Errorf("workspace ID or slug is required. Use --workspace flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)
	client.SetProject(listWorkspace)

	// First, get project basic info to retrieve both ID and slug
	infoQuery := fmt.Sprintf(`
		query GetProjectInfo {
			project(id: "%s") {
				id
				slug
			}
		}
	`, listWorkspace)

	var projectInfo struct {
		Project struct {
			ID   string `json:"id"`
			Slug string `json:"slug"`
		} `json:"project"`
	}

	if err := client.ExecuteQueryWithResult(infoQuery, nil, &projectInfo); err != nil {
		return fmt.Errorf("failed to fetch project info: %w", err)
	}

	// Query using opposite identifier to force DB lookup with todoFields
	queryIdentifier := projectInfo.Project.Slug
	if listWorkspace == projectInfo.Project.Slug {
		queryIdentifier = projectInfo.Project.ID
	}

	query := fmt.Sprintf(`
		query GetProjectTodoFields {
			project(id: "%s") {
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
	`, queryIdentifier)

	var response ProjectWithTodoFieldsResponse
	if err := client.ExecuteQueryWithResult(query, nil, &response); err != nil {
		return fmt.Errorf("failed to fetch project: %w", err)
	}

	fmt.Printf("\n=== Custom Field Groups for Workspace: %s ===\n", response.Project.Name)
	fmt.Printf("Workspace ID: %s\n\n", response.Project.ID)

	if len(response.Project.TodoFields) == 0 {
		fmt.Println("No field configuration found.")
		return nil
	}

	groupCount := 0
	ungroupedCount := 0

	for i, field := range response.Project.TodoFields {
		if field.Type == "CUSTOM_FIELD_GROUP" {
			groupCount++
			groupName := "Unnamed Group"
			if field.Name != nil {
				groupName = *field.Name
			}

			color := "default"
			if field.Color != nil {
				color = *field.Color
			}

			groupID := "no-id"
			if field.CustomFieldID != nil {
				groupID = *field.CustomFieldID
			}

			fmt.Printf("%d. [GROUP] %s (color: %s) [Group ID: %s]\n", i+1, groupName, color, groupID)

			if len(field.TodoFields) > 0 {
				for j, nestedField := range field.TodoFields {
					fieldID := "no-id"
					if nestedField.CustomFieldID != nil {
						fieldID = *nestedField.CustomFieldID
					}
					fmt.Printf("   %d.%d  -- %s [Field ID: %s]\n", i+1, j+1, nestedField.Type, fieldID)
				}
			} else {
				fmt.Printf("   (empty group)\n")
			}
		} else if field.Type == "CUSTOM_FIELD" {
			ungroupedCount++
			fieldID := "no-id"
			if field.CustomFieldID != nil {
				fieldID = *field.CustomFieldID
			}
			fmt.Printf("%d. [FIELD] %s [Field ID: %s]\n", i+1, field.Type, fieldID)
		} else {
			fieldID := "no-id"
			if field.CustomFieldID != nil {
				fieldID = *field.CustomFieldID
			}
			fmt.Printf("%d. [OTHER] %s [ID: %s]\n", i+1, field.Type, fieldID)
		}
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Total Groups: %d\n", groupCount)
	fmt.Printf("Ungrouped Custom Fields: %d\n", ungroupedCount)
	fmt.Printf("Total Field Configurations: %d\n", len(response.Project.TodoFields))
	fmt.Println()

	return nil
}
