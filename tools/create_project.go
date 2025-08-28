package tools

import (
	"flag"
	"fmt"
	"strings"
	
	"demo-builder/common"
)

// Response structures
type CreatedProject struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
	Category    string `json:"category"`
}

type CreateProjectResponse struct {
	CreateProject CreatedProject `json:"createProject"`
}

// Using common package for project constants

// Execute GraphQL mutation
func executeCreateProject(client *common.Client, input common.CreateProjectInput) (*CreatedProject, error) {
	// Build the mutation
	mutation := fmt.Sprintf(`
		mutation CreateProject {
			createProject(input: {
				name: "%s"
				companyId: "%s"
				%s
			}) {
				id
				name
				slug
				description
				color
				icon
				category
			}
		}
	`, input.Name, input.CompanyID, buildProjectOptionalFields(input))

	// Execute mutation
	var response CreateProjectResponse
	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		return nil, err
	}

	return &response.CreateProject, nil
}

// Build optional fields for the mutation
func buildProjectOptionalFields(input common.CreateProjectInput) string {
	var fields []string

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
	if input.TemplateID != "" {
		fields = append(fields, fmt.Sprintf(`templateId: "%s"`, input.TemplateID))
	}

	return strings.Join(fields, "\n\t\t\t\t")
}

// RunCreateProject creates a new project
func RunCreateProject(args []string) error {
	// Create flag set for this command
	fs := flag.NewFlagSet("create-project", flag.ExitOnError)
	
	// Parse command line flags
	name := fs.String("name", "", "Project name (required)")
	description := fs.String("description", "", "Project description")
	color := fs.String("color", "", "Project color (e.g., blue, red, #3B82F6)")
	icon := fs.String("icon", "mdi-briefcase-variant-outline", "Project icon")
	category := fs.String("category", "GENERAL", "Project category")
	templateID := fs.String("template", "", "Template ID to create from")
	listOptions := fs.Bool("list", false, "List available options")
	
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Show available options if requested
	if *listOptions {
		fmt.Println("\n=== Available Options ===")
		fmt.Println("\nCategories:")
		for _, cat := range common.ProjectCategories {
			fmt.Printf("  - %s\n", cat)
		}
		fmt.Println("\nColors:")
		for name, hex := range common.ProjectColors {
			fmt.Printf("  - %s: %s\n", name, hex)
		}
		fmt.Println("\nIcons:")
		for _, ico := range common.ProjectIcons {
			fmt.Printf("  - %s\n", ico)
		}
		return nil
	}

	// Validate required parameters
	if *name == "" {
		return fmt.Errorf("project name is required. Use -name flag")
	}

	// Load configuration
	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create client
	client := common.NewClient(config)

	// Process color input
	colorValue := *color
	if colorValue != "" && !strings.HasPrefix(colorValue, "#") {
		if hex, ok := common.ProjectColors[colorValue]; ok {
			colorValue = hex
		}
	}

	// Create project input
	input := common.CreateProjectInput{
		Name:        *name,
		CompanyID:   client.GetCompanyID(),
		Description: *description,
		Color:       colorValue,
		Icon:        *icon,
		Category:    *category,
		TemplateID:  *templateID,
	}

	// Execute creation
	fmt.Printf("Creating project '%s' in company '%s'...\n", input.Name, client.GetCompanyID())
	
	project, err := executeCreateProject(client, input)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	// Display results
	fmt.Println("\n✅ Project created successfully!")
	fmt.Printf("\nProject Details:\n")
	fmt.Printf("  ID:          %s\n", project.ID)
	fmt.Printf("  Name:        %s\n", project.Name)
	fmt.Printf("  Slug:        %s\n", project.Slug)
	if project.Description != "" {
		fmt.Printf("  Description: %s\n", project.Description)
	}
	if project.Color != "" {
		fmt.Printf("  Color:       %s\n", project.Color)
	}
	if project.Icon != "" {
		fmt.Printf("  Icon:        %s\n", project.Icon)
	}
	fmt.Printf("  Category:    %s\n", project.Category)
	
	fmt.Printf("\nYou can now create lists in this project using:\n")
	fmt.Printf("  go run . create-list -project %s -names \"To Do,In Progress,Done\"\n", project.ID)
	
	return nil
}