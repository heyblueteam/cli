package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
)

// Project creation input
type CreateProjectInput struct {
	Name        string `json:"name"`
	CompanyID   string `json:"companyId"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Category    string `json:"category,omitempty"`
	TemplateID  string `json:"templateId,omitempty"`
}

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

// Available project categories
var projectCategories = []string{
	"GENERAL", "CRM", "MARKETING", "ENGINEERING", "PRODUCT", "SALES",
	"DESIGN", "FINANCE", "HR", "LEGAL", "OPERATIONS", "SUPPORT",
}

// Common project colors
var projectColors = map[string]string{
	"blue":   "#3B82F6",
	"red":    "#EF4444",
	"green":  "#10B981",
	"purple": "#8B5CF6",
	"yellow": "#F59E0B",
	"pink":   "#EC4899",
	"indigo": "#6366F1",
	"gray":   "#6B7280",
}

// Available project icons
var projectIcons = []string{
	"mdi-cash",
	"mdi-shield-bug-outline",
	"mdi-account-box-outline",
	"mdi-account-group-outline",
	"mdi-alarm-panel-outline",
	"mdi-animation-play-outline",
	"mdi-application-brackets-outline",
	"mdi-archive-arrow-up-outline",
	"mdi-badge-account-horizontal-outline",
	"mdi-bank-outline",
	"mdi-basket-outline",
	"mdi-book-open-outline",
	"mdi-briefcase-variant-outline",
	"mdi-car-outline",
	"mdi-cake-variant-outline",
	"mdi-calendar-account-outline",
	"mdi-camera-outline",
	"mdi-card-account-mail-outline",
	"mdi-cards-club-outline",
	"mdi-cards-heart-outline",
	"mdi-cellphone-basic",
	"mdi-chart-line",
	"mdi-flag-variant-outline",
	"mdi-chat-outline",
	"mdi-cloud-check-outline",
	"mdi-clipboard-list-outline",
	"mdi-clock-time-eight-outline",
	"mdi-video-outline",
	"mdi-gamepad-round-outline",
	"mdi-earth",
	"mdi-image-frame",
	"mdi-laptop",
	"mdi-microphone-outline",
	"mdi-music-note",
	"mdi-cog-outline",
	"mdi-compass-outline",
	"mdi-home-outline",
	"mdi-airplane-takeoff",
	"mdi-gamepad-variant-outline",
	"mdi-key-outline",
	"mdi-folder",
	"mdi-folder-search-outline",
}

// Execute GraphQL mutation
func executeCreateProject(client *Client, input CreateProjectInput) (*CreatedProject, error) {
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
	`, input.Name, input.CompanyID, buildOptionalFields(input))

	// Execute mutation
	var response CreateProjectResponse
	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		return nil, err
	}

	return &response.CreateProject, nil
}

// Build optional fields for the mutation
func buildOptionalFields(input CreateProjectInput) string {
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

func main() {
	// Parse command line flags
	name := flag.String("name", "", "Project name (required)")
	description := flag.String("description", "", "Project description")
	color := flag.String("color", "", "Project color (e.g., blue, red, #3B82F6)")
	icon := flag.String("icon", "mdi-briefcase-variant-outline", "Project icon")
	category := flag.String("category", "GENERAL", "Project category")
	templateID := flag.String("template", "", "Template ID to create from")
	listOptions := flag.Bool("list", false, "List available options")
	flag.Parse()

	// Show available options if requested
	if *listOptions {
		fmt.Println("\n=== Available Options ===")
		fmt.Println("\nCategories:")
		for _, cat := range projectCategories {
			fmt.Printf("  - %s\n", cat)
		}
		fmt.Println("\nColors:")
		for name, hex := range projectColors {
			fmt.Printf("  - %s: %s\n", name, hex)
		}
		fmt.Println("\nIcons:")
		for _, ico := range projectIcons {
			fmt.Printf("  - %s\n", ico)
		}
		return
	}

	// Validate required parameters
	if *name == "" {
		log.Fatal("Project name is required. Use -name flag")
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create client
	client := NewClient(config)

	// Process color input
	colorValue := *color
	if colorValue != "" && !strings.HasPrefix(colorValue, "#") {
		if hex, ok := projectColors[colorValue]; ok {
			colorValue = hex
		}
	}

	// Create project input
	input := CreateProjectInput{
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
		log.Fatalf("Failed to create project: %v", err)
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
	fmt.Printf("  go run create-list.go -project %s -names \"To Do,In Progress,Done\"\n", project.ID)
}