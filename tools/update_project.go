package tools

import (
	"flag"
	"fmt"
	"log"
	"strconv"
	"strings"
	"demo-builder/common"
)

// Edit project input
type EditProjectInput struct {
	ProjectID                 string                 `json:"projectId"`
	Name                      string                 `json:"name,omitempty"`
	Slug                      string                 `json:"slug,omitempty"`
	Description               string                 `json:"description,omitempty"`
	Color                     string                 `json:"color,omitempty"`
	Icon                      string                 `json:"icon,omitempty"`
	Category                  string                 `json:"category,omitempty"`
	TodoAlias                 string                 `json:"todoAlias,omitempty"`
	HideRecordCount           *bool                  `json:"hideRecordCount,omitempty"`
	ShowTimeSpentInTodoList   *bool                  `json:"showTimeSpentInTodoList,omitempty"`
	ShowTimeSpentInProject    *bool                  `json:"showTimeSpentInProject,omitempty"`
	Features                  []common.ProjectFeatureInput  `json:"features,omitempty"`
}

// Response structures
type EditedProject struct {
	ID                        string           `json:"id"`
	Name                      string           `json:"name"`
	Slug                      string           `json:"slug"`
	Description               string           `json:"description"`
	Color                     string           `json:"color"`
	Icon                      string           `json:"icon"`
	Category                  string           `json:"category"`
	TodoAlias                 string           `json:"todoAlias"`
	HideRecordCount           bool             `json:"hideRecordCount"`
	ShowTimeSpentInTodoList   bool             `json:"showTimeSpentInTodoList"`
	ShowTimeSpentInProject    bool             `json:"showTimeSpentInProject"`
	Features                  []common.ProjectFeature `json:"features"`
}

type EditProjectResponse struct {
	EditProject EditedProject `json:"editProject"`
}

// Allowed feature types for Blue projects
var featureTypes = []string{
	"Activity",
	"Todo",
	"Wiki",
	"Chat",
	"Docs",
	"Forms",
	"Files",
	"People",
}

// Available project categories
var projectCategories = []string{
	"GENERAL", "CRM", "MARKETING", "ENGINEERING", "PRODUCT", "SALES",
	"DESIGN", "FINANCE", "HR", "LEGAL", "OPERATIONS", "SUPPORT",
}

// Get current project data to merge features
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

// Merge existing features with user-specified changes
func mergeFeatures(existingFeatures []common.ProjectFeature, newFeatures []common.ProjectFeatureInput) []common.ProjectFeatureInput {
	// Create a map of all possible feature types with default enabled=true
	featureMap := make(map[string]bool)
	for _, featureType := range featureTypes {
		featureMap[featureType] = true
	}
	
	// Apply existing feature states
	for _, feature := range existingFeatures {
		featureMap[feature.Type] = feature.Enabled
	}
	
	// Apply user-specified changes
	for _, feature := range newFeatures {
		featureMap[feature.Type] = feature.Enabled
	}
	
	// Convert back to array with all feature types
	var result []common.ProjectFeatureInput
	for _, featureType := range featureTypes {
		result = append(result, common.ProjectFeatureInput{
			Type:    featureType,
			Enabled: featureMap[featureType],
		})
	}
	
	return result
}

// Execute GraphQL mutation
func executeEditProject(client *common.Client, input EditProjectInput) (*EditedProject, error) {
	// If features are being updated, we need to merge with existing features
	if len(input.Features) > 0 {
		currentProject, err := getCurrentProject(client, input.ProjectID)
		if err != nil {
			return nil, fmt.Errorf("failed to get current project state: %v", err)
		}
		
		// Merge features
		input.Features = mergeFeatures(currentProject.Features, input.Features)
	}

	// Build the mutation
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

	// Execute mutation
	var response EditProjectResponse
	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		return nil, err
	}

	return &response.EditProject, nil
}

// Build optional fields for the mutation
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
		featuresStr := buildFeaturesString(input.Features)
		fields = append(fields, fmt.Sprintf(`features: [%s]`, featuresStr))
	}

	return strings.Join(fields, "\n\t\t\t\t")
}

// Build features array string for GraphQL
func buildFeaturesString(features []common.ProjectFeatureInput) string {
	var featureStrings []string
	for _, feature := range features {
		featureStrings = append(featureStrings, fmt.Sprintf(`{type: "%s", enabled: %t}`, feature.Type, feature.Enabled))
	}
	return strings.Join(featureStrings, ", ")
}

// Parse features from command line string
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
		// Capitalize first letter to match expected format
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

// Helper to parse boolean flags
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

func RunUpdateProject(args []string) error {
	// Create flagset for this tool
	fs := flag.NewFlagSet("update-project", flag.ExitOnError)
	
	// Parse command line flags
	projectID := fs.String("project", "", "Project ID to edit (required)")
	name := fs.String("name", "", "New project name")
	slug := fs.String("slug", "", "New project slug")
	description := fs.String("description", "", "New project description")
	color := fs.String("color", "", "New project color")
	icon := fs.String("icon", "", "New project icon")
	category := fs.String("category", "", "New project category")
	todoAlias := fs.String("todo-alias", "", "Custom name for todos/records")
	hideRecordCount := fs.String("hide-record-count", "", "Hide record count (true/false)")
	showTimeSpentInTodoList := fs.String("show-time-todo-list", "", "Show time spent in todo list (true/false)")
	showTimeSpentInProject := fs.String("show-time-project", "", "Show time spent in project (true/false)")
	featuresStr := fs.String("features", "", "Features to toggle (comma-separated, format: TYPE:true/false)")
	listOptions := fs.Bool("list", false, "List available options")
	simple := fs.Bool("simple", false, "Simple output format")
	
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %v", err)
	}

	// Show available options if requested
	if *listOptions {
		fmt.Println("\n=== Available Options ===")
		fmt.Println("\nFeature Types:")
		for _, feature := range featureTypes {
			fmt.Printf("  - %s\n", feature)
		}
		fmt.Println("\nCategories:")
		for _, cat := range projectCategories {
			fmt.Printf("  - %s\n", cat)
		}
		fmt.Println("\nExample feature toggles:")
		fmt.Printf("  -features \"Chat:true,Files:false,Wiki:true\"\n")
		fmt.Println("\nExample boolean flags:")
		fmt.Printf("  -hide-record-count true\n")
		fmt.Printf("  -show-time-project false\n")
		return nil
	}

	// Validate required parameters
	if *projectID == "" {
		return fmt.Errorf("Project ID is required. Use -project flag")
	}

	// Load configuration
	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Create client and set project context
	client := common.NewClient(config)
	client.SetProjectID(*projectID)

	// Parse features
	features := parseFeatures(*featuresStr)

	// Create edit input
	input := EditProjectInput{
		ProjectID:                 *projectID,
		Name:                      *name,
		Slug:                      *slug,
		Description:               *description,
		Color:                     *color,
		Icon:                      *icon,
		Category:                  *category,
		TodoAlias:                 *todoAlias,
		HideRecordCount:           parseBoolPtr(*hideRecordCount),
		ShowTimeSpentInTodoList:   parseBoolPtr(*showTimeSpentInTodoList),
		ShowTimeSpentInProject:    parseBoolPtr(*showTimeSpentInProject),
		Features:                  features,
	}

	// Execute edit
	if !*simple {
		fmt.Printf("Editing project %s...\n", *projectID)
	}
	
	project, err := executeEditProject(client, input)
	if err != nil {
		return fmt.Errorf("failed to edit project: %v", err)
	}

	// Display results
	if *simple {
		fmt.Printf("Project %s updated successfully\n", project.ID)
	} else {
		fmt.Println("\n✅ Project updated successfully!")
		fmt.Printf("\nProject Details:\n")
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
			fmt.Printf("\nProject Features:\n")
			for _, feature := range project.Features {
				status := "❌ Disabled"
				if feature.Enabled {
					status = "✅ Enabled"
				}
				fmt.Printf("  %-20s %s\n", feature.Type, status)
			}
		}
	}

	return nil
}