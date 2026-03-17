package workspaces

import (
	"fmt"
	"log"
	"strconv"
	"strings"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

// Edit project input
type EditProjectInput struct {
	ProjectID               string                       `json:"projectId"`
	Name                    string                       `json:"name,omitempty"`
	Slug                    string                       `json:"slug,omitempty"`
	Description             string                       `json:"description,omitempty"`
	Color                   string                       `json:"color,omitempty"`
	Icon                    string                       `json:"icon,omitempty"`
	Category                string                       `json:"category,omitempty"`
	TodoAlias               string                       `json:"todoAlias,omitempty"`
	HideRecordCount         *bool                        `json:"hideRecordCount,omitempty"`
	ShowTimeSpentInTodoList *bool                        `json:"showTimeSpentInTodoList,omitempty"`
	ShowTimeSpentInProject  *bool                        `json:"showTimeSpentInProject,omitempty"`
	Features                []common.ProjectFeatureInput `json:"features,omitempty"`
}

type EditedProject struct {
	ID                      string                  `json:"id"`
	Name                    string                  `json:"name"`
	Slug                    string                  `json:"slug"`
	Description             string                  `json:"description"`
	Color                   string                  `json:"color"`
	Icon                    string                  `json:"icon"`
	Category                string                  `json:"category"`
	TodoAlias               string                  `json:"todoAlias"`
	HideRecordCount         bool                    `json:"hideRecordCount"`
	ShowTimeSpentInTodoList bool                    `json:"showTimeSpentInTodoList"`
	ShowTimeSpentInProject  bool                    `json:"showTimeSpentInProject"`
	Features                []common.ProjectFeature `json:"features"`
}

type EditProjectResponse struct {
	EditProject EditedProject `json:"editProject"`
}

var featureTypes = []string{
	"Activity", "Todo", "Wiki", "Chat", "Docs", "Forms", "Files", "People",
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a workspace",
	Long:  "Update workspace settings, features, and metadata.",
	Example: `  blue workspaces update --workspace <id> --name "New Name"
  blue workspaces update --workspace <id> --features "Chat:true,Files:false"
  blue workspaces update --workspace <id> --color green --icon chart`,
	RunE: runUpdate,
}

var (
	updateWorkspace             string
	updateName                  string
	updateSlug                  string
	updateDescription           string
	updateColor                 string
	updateIcon                  string
	updateCategory              string
	updateTodoAlias             string
	updateHideRecordCount       string
	updateShowTimeSpentTodoList string
	updateShowTimeSpentProject  string
	updateFeaturesStr           string
	updateListOptions           bool
	updateSimple                bool
)

func init() {
	updateCmd.Flags().StringVarP(&updateWorkspace, "workspace", "w", "", "Workspace ID to update (required)")
	updateCmd.Flags().StringVarP(&updateName, "name", "n", "", "New workspace name")
	updateCmd.Flags().StringVar(&updateSlug, "slug", "", "New workspace slug")
	updateCmd.Flags().StringVarP(&updateDescription, "description", "d", "", "New workspace description")
	updateCmd.Flags().StringVar(&updateColor, "color", "", "New workspace color")
	updateCmd.Flags().StringVar(&updateIcon, "icon", "", "New workspace icon")
	updateCmd.Flags().StringVar(&updateCategory, "category", "", "New workspace category")
	updateCmd.Flags().StringVar(&updateTodoAlias, "todo-alias", "", "Custom name for todos/records")
	updateCmd.Flags().StringVar(&updateHideRecordCount, "hide-record-count", "", "Hide record count (true/false)")
	updateCmd.Flags().StringVar(&updateShowTimeSpentTodoList, "show-time-todo-list", "", "Show time spent in todo list (true/false)")
	updateCmd.Flags().StringVar(&updateShowTimeSpentProject, "show-time-project", "", "Show time spent in project (true/false)")
	updateCmd.Flags().StringVar(&updateFeaturesStr, "features", "", "Features to toggle (format: TYPE:true/false,TYPE:true/false)")
	updateCmd.Flags().BoolVar(&updateListOptions, "list-options", false, "List available options")
	updateCmd.Flags().BoolVarP(&updateSimple, "simple", "s", false, "Simple output format")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if updateListOptions {
		fmt.Println("\n=== Available Options ===")
		fmt.Println("\nFeature Types:")
		for _, feature := range featureTypes {
			fmt.Printf("  - %s\n", feature)
		}
		fmt.Println("\nCategories:")
		for _, cat := range common.ProjectCategories {
			fmt.Printf("  - %s\n", cat)
		}
		fmt.Println("\nExample feature toggles:")
		fmt.Printf("  --features \"Chat:true,Files:false,Wiki:true\"\n")
		return nil
	}

	if updateWorkspace == "" {
		return fmt.Errorf("workspace ID is required. Use --workspace flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)
	client.SetProject(updateWorkspace)

	features := parseFeatures(updateFeaturesStr)

	input := EditProjectInput{
		ProjectID:               updateWorkspace,
		Name:                    updateName,
		Slug:                    updateSlug,
		Description:             updateDescription,
		Color:                   updateColor,
		Icon:                    updateIcon,
		Category:                updateCategory,
		TodoAlias:               updateTodoAlias,
		HideRecordCount:         parseBoolPtr(updateHideRecordCount),
		ShowTimeSpentInTodoList: parseBoolPtr(updateShowTimeSpentTodoList),
		ShowTimeSpentInProject:  parseBoolPtr(updateShowTimeSpentProject),
		Features:                features,
	}

	if !updateSimple {
		fmt.Printf("Updating workspace %s...\n", updateWorkspace)
	}

	project, err := executeEditProject(client, input)
	if err != nil {
		return fmt.Errorf("failed to update workspace: %w", err)
	}

	if updateSimple {
		fmt.Printf("Workspace %s updated successfully\n", project.ID)
	} else {
		fmt.Println("\nWorkspace updated successfully!")
		fmt.Printf("\nWorkspace Details:\n")
		fmt.Printf("  ID:                        %s\n", project.ID)
		fmt.Printf("  Name:                      %s\n", project.Name)
		fmt.Printf("  Slug:                      %s\n", project.Slug)
		if project.Description != "" {
			fmt.Printf("  Description:               %s\n", project.Description)
		}
		if project.Color != "" {
			fmt.Printf("  Color:                     %s\n", project.Color)
		}
		if project.Icon != "" {
			fmt.Printf("  Icon:                      %s\n", project.Icon)
		}
		fmt.Printf("  Category:                  %s\n", project.Category)
		if project.TodoAlias != "" {
			fmt.Printf("  Todo Alias:                %s\n", project.TodoAlias)
		}
		fmt.Printf("  Hide Record Count:         %t\n", project.HideRecordCount)
		fmt.Printf("  Show Time Spent (Lists):   %t\n", project.ShowTimeSpentInTodoList)
		fmt.Printf("  Show Time Spent (Project): %t\n", project.ShowTimeSpentInProject)

		if len(project.Features) > 0 {
			fmt.Printf("\nWorkspace Features:\n")
			for _, feature := range project.Features {
				status := "Disabled"
				if feature.Enabled {
					status = "Enabled"
				}
				fmt.Printf("  %-20s %s\n", feature.Type, status)
			}
		}
	}

	return nil
}

func getCurrentProject(client *common.Client, projectID string) (*EditedProject, error) {
	query := fmt.Sprintf(`
		query GetProject {
			project(id: "%s") {
				id
				name
				slug
				description
				color
				icon
				category
				todoAlias
				hideRecordCount
				showTimeSpentInTodoList
				showTimeSpentInProject
				features {
					type
					enabled
				}
			}
		}
	`, projectID)

	var response struct {
		Project EditedProject `json:"project"`
	}
	if err := client.ExecuteQueryWithResult(query, nil, &response); err != nil {
		return nil, err
	}

	return &response.Project, nil
}

func mergeFeatures(existingFeatures []common.ProjectFeature, newFeatures []common.ProjectFeatureInput) []common.ProjectFeatureInput {
	featureMap := make(map[string]bool)
	for _, featureType := range featureTypes {
		featureMap[featureType] = true
	}
	for _, feature := range existingFeatures {
		featureMap[feature.Type] = feature.Enabled
	}
	for _, feature := range newFeatures {
		featureMap[feature.Type] = feature.Enabled
	}

	var result []common.ProjectFeatureInput
	for _, featureType := range featureTypes {
		result = append(result, common.ProjectFeatureInput{
			Type:    featureType,
			Enabled: featureMap[featureType],
		})
	}
	return result
}

func executeEditProject(client *common.Client, input EditProjectInput) (*EditedProject, error) {
	if len(input.Features) > 0 {
		currentProject, err := getCurrentProject(client, input.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to get current workspace state: %v", err)
		}
		input.Features = mergeFeatures(currentProject.Features, input.Features)
	}

	mutation := fmt.Sprintf(`
		mutation EditProject {
			editProject(input: {
				projectId: "%s"
				%s
			}) {
				id
				name
				slug
				description
				color
				icon
				category
				todoAlias
				hideRecordCount
				showTimeSpentInTodoList
				showTimeSpentInProject
				features {
					type
					enabled
				}
			}
		}
	`, input.ProjectID, buildEditFields(input))

	var response EditProjectResponse
	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		return nil, err
	}

	return &response.EditProject, nil
}

func buildEditFields(input EditProjectInput) string {
	var fields []string

	if input.Name != "" {
		fields = append(fields, fmt.Sprintf(`name: "%s"`, input.Name))
	}
	if input.Slug != "" {
		fields = append(fields, fmt.Sprintf(`slug: "%s"`, input.Slug))
	}
	if input.Description != "" {
		fields = append(fields, fmt.Sprintf(`description: "%s"`, input.Description))
	}
	if input.Color != "" {
		fields = append(fields, fmt.Sprintf(`color: "%s"`, input.Color))
	}
	if input.Icon != "" {
		fields = append(fields, fmt.Sprintf(`icon: "%s"`, input.Icon))
	}
	if input.Category != "" {
		fields = append(fields, fmt.Sprintf(`category: %s`, input.Category))
	}
	if input.TodoAlias != "" {
		fields = append(fields, fmt.Sprintf(`todoAlias: "%s"`, input.TodoAlias))
	}
	if input.HideRecordCount != nil {
		fields = append(fields, fmt.Sprintf(`hideRecordCount: %t`, *input.HideRecordCount))
	}
	if input.ShowTimeSpentInTodoList != nil {
		fields = append(fields, fmt.Sprintf(`showTimeSpentInTodoList: %t`, *input.ShowTimeSpentInTodoList))
	}
	if input.ShowTimeSpentInProject != nil {
		fields = append(fields, fmt.Sprintf(`showTimeSpentInProject: %t`, *input.ShowTimeSpentInProject))
	}
	if len(input.Features) > 0 {
		var featureStrings []string
		for _, feature := range input.Features {
			featureStrings = append(featureStrings, fmt.Sprintf(`{type: "%s", enabled: %t}`, feature.Type, feature.Enabled))
		}
		fields = append(fields, fmt.Sprintf(`features: [%s]`, strings.Join(featureStrings, ", ")))
	}

	return strings.Join(fields, "\n\t\t\t\t")
}

func parseFeatures(featuresStr string) []common.ProjectFeatureInput {
	if featuresStr == "" {
		return nil
	}

	var features []common.ProjectFeatureInput
	pairs := strings.Split(featuresStr, ",")

	for _, pair := range pairs {
		parts := strings.Split(strings.TrimSpace(pair), ":")
		if len(parts) != 2 {
			log.Printf("Warning: Invalid feature format '%s', expected 'TYPE:true/false'", pair)
			continue
		}

		featureType := strings.TrimSpace(parts[0])
		if len(featureType) > 0 {
			featureType = strings.ToUpper(string(featureType[0])) + strings.ToLower(featureType[1:])
		}
		enabledStr := strings.TrimSpace(strings.ToLower(parts[1]))
		enabled := enabledStr == "true" || enabledStr == "1" || enabledStr == "yes" || enabledStr == "on"

		features = append(features, common.ProjectFeatureInput{
			Type:    featureType,
			Enabled: enabled,
		})
	}

	return features
}

func parseBoolPtr(value string) *bool {
	if value == "" {
		return nil
	}
	b, err := strconv.ParseBool(value)
	if err != nil {
		log.Printf("Warning: Invalid boolean value '%s', ignoring", value)
		return nil
	}
	return &b
}
