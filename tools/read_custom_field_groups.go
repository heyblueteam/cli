package tools

import (
	"cli/common"
	"flag"
	"fmt"
)

// ProjectWithTodoFieldsResponse for fetching project with todoFields
type ProjectWithTodoFieldsResponse struct {
	Project struct {
		ID         string             `json:"id"`
		Name       string             `json:"name"`
		TodoFields []common.TodoField `json:"todoFields"`
	} `json:"project"`
}

// RunReadCustomFieldGroups displays custom field groups and their organization
func RunReadCustomFieldGroups(args []string) error {
	fs := flag.NewFlagSet("read-field-groups", flag.ExitOnError)
	projectID := fs.String("project", "", "Project ID or slug (required)")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	if *projectID == "" {
		return fmt.Errorf("project ID or slug is required. Use -project flag")
	}

	// Load configuration
	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Create client
	client := common.NewClient(config)

	// Set project context for authorization
	client.SetProjectID(*projectID)

	// First, get project basic info to retrieve both ID and slug
	// This is necessary because the API returns a cached project without todoFields
	// when the query ID matches the context ID
	infoQuery := fmt.Sprintf(`
		query GetProjectInfo {
			project(id: "%s") {
				id
				slug
			}
		}
	`, *projectID)

	var projectInfo struct {
		Project struct {
			ID   string `json:"id"`
			Slug string `json:"slug"`
		} `json:"project"`
	}

	if err := client.ExecuteQueryWithResult(infoQuery, nil, &projectInfo); err != nil {
		return fmt.Errorf("failed to fetch project info: %v", err)
	}

	// Query using slug if user provided ID, or ID if user provided slug
	// This ensures the query identifier doesn't match context identifier,
	// forcing a full database lookup with todoFields
	queryIdentifier := projectInfo.Project.Slug
	if *projectID == projectInfo.Project.Slug {
		// User provided slug, query by ID
		queryIdentifier = projectInfo.Project.ID
	}

	// Build GraphQL query (use different identifier than context)
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

	// Execute query
	var response ProjectWithTodoFieldsResponse
	if err := client.ExecuteQueryWithResult(query, nil, &response); err != nil {
		return fmt.Errorf("failed to fetch project: %v", err)
	}

	// Display results
	fmt.Printf("\n=== Custom Field Groups for Project: %s ===\n", response.Project.Name)
	fmt.Printf("Project ID: %s\n\n", response.Project.ID)

	if len(response.Project.TodoFields) == 0 {
		fmt.Println("No field configuration found.")
		return nil
	}

	// Display hierarchy
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

			fmt.Printf("%d. 📁 %s (color: %s) [Group ID: %s]\n", i+1, groupName, color, groupID)

			if len(field.TodoFields) > 0 {
				for j, nestedField := range field.TodoFields {
					fieldID := "no-id"
					if nestedField.CustomFieldID != nil {
						fieldID = *nestedField.CustomFieldID
					}
					fmt.Printf("   %d.%d  └─ %s [Field ID: %s]\n", i+1, j+1, nestedField.Type, fieldID)
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
			fmt.Printf("%d. 📄 %s [Field ID: %s]\n", i+1, field.Type, fieldID)
		} else {
			fieldID := "no-id"
			if field.CustomFieldID != nil {
				fieldID = *field.CustomFieldID
			}
			fmt.Printf("%d. 📌 %s [ID: %s]\n", i+1, field.Type, fieldID)
		}
	}

	fmt.Printf("\n=== Summary ===\n")
	fmt.Printf("Total Groups: %d\n", groupCount)
	fmt.Printf("Ungrouped Custom Fields: %d\n", ungroupedCount)
	fmt.Printf("Total Field Configurations: %d\n", len(response.Project.TodoFields))
	fmt.Println()

	return nil
}
