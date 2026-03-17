package workspaces

import (
	"fmt"
	"strings"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

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

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new workspace",
	Long:  "Create a new workspace in your Blue company with customizable settings.",
	Example: `  blue workspaces create --name "My Workspace"
  blue workspaces create --name "Sprint Planning" --color blue --icon rocket --category ENGINEERING
  blue workspaces create --list-options`,
	RunE: runCreate,
}

var (
	createName        string
	createDescription string
	createColor       string
	createIcon        string
	createCategory    string
	createTemplateID  string
	createListOptions bool
)

func init() {
	createCmd.Flags().StringVarP(&createName, "name", "n", "", "Workspace name (required)")
	createCmd.Flags().StringVarP(&createDescription, "description", "d", "", "Workspace description")
	createCmd.Flags().StringVar(&createColor, "color", "", "Workspace color (e.g., blue, red, #3B82F6)")
	createCmd.Flags().StringVar(&createIcon, "icon", "mdi-briefcase-variant-outline", "Workspace icon")
	createCmd.Flags().StringVar(&createCategory, "category", "GENERAL", "Workspace category")
	createCmd.Flags().StringVar(&createTemplateID, "template", "", "Template ID to create from")
	createCmd.Flags().BoolVar(&createListOptions, "list-options", false, "List available options")
}

func runCreate(cmd *cobra.Command, args []string) error {
	if createListOptions {
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

	if createName == "" {
		return fmt.Errorf("workspace name is required. Use --name flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)

	// Process color input
	colorValue := createColor
	if colorValue != "" && !strings.HasPrefix(colorValue, "#") {
		if hex, ok := common.ProjectColors[colorValue]; ok {
			colorValue = hex
		}
	}

	input := common.CreateProjectInput{
		Name:        createName,
		CompanyID:   client.GetCompanyID(),
		Description: createDescription,
		Color:       colorValue,
		Icon:        createIcon,
		Category:    createCategory,
		TemplateID:  createTemplateID,
	}

	fmt.Printf("Creating workspace '%s'...\n", input.Name)

	project, err := executeCreateProject(client, input)
	if err != nil {
		return fmt.Errorf("failed to create workspace: %w", err)
	}

	fmt.Println("\nWorkspace created successfully!")
	fmt.Printf("\nWorkspace Details:\n")
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

	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  blue lists create --workspace %s --names \"To Do,In Progress,Done\"\n", project.ID)

	return nil
}

func executeCreateProject(client *common.Client, input common.CreateProjectInput) (*CreatedProject, error) {
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

	var response CreateProjectResponse
	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		return nil, err
	}

	return &response.CreateProject, nil
}

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
