package tools

import (
	"flag"
	"fmt"
	
	. "demo-builder/common" // Import types and auth from common
)

// ProjectList represents a paginated list of projects
type ProjectList struct {
	Items    []Project       `json:"items"`
	PageInfo OffsetPageInfo  `json:"pageInfo"`
}

type ProjectListResponse struct {
	ProjectList ProjectList `json:"projectList"`
}

// Build query with pagination and search
func buildProjectQuery(companyID string, simple bool, skip int, take int, search string, showArchived bool, showTemplates bool) string {
	fields := "id name"
	if !simple {
		fields = `id
				uid
				slug
				name
				description
				archived
				color
				icon
				createdAt
				updatedAt
				position
				isTemplate`
	}

	// Build filter
	filter := fmt.Sprintf(`companyIds: ["%s"]`, companyID)
	if search != "" {
		filter += fmt.Sprintf(`, search: "%s"`, search)
	}
	if !showArchived {
		filter += `, archived: false`
	}
	if !showTemplates {
		filter += `, isTemplate: false`
	}

	query := fmt.Sprintf(`query ProjectListQuery {
		projectList(
			filter: { %s }
			skip: %d
			take: %d
			sort: [name_ASC]
		) {
			items {
				%s
			}
			pageInfo {
				totalPages
				totalItems
				page
				perPage
				hasNextPage
				hasPreviousPage
			}
			totalCount
		}
	}`, filter, skip, take, fields)

	return query
}

// RunReadProjects lists all projects with optional filtering
func RunReadProjects(args []string) error {
	// Create flag set for this command
	fs := flag.NewFlagSet("read-projects", flag.ExitOnError)
	
	// Parse command line flags
	simple := fs.Bool("simple", false, "Show only project names and IDs")
	page := fs.Int("page", 1, "Page number (default: 1)")
	pageSize := fs.Int("size", 20, "Page size (default: 20)")
	search := fs.String("search", "", "Search projects by name")
	all := fs.Bool("all", false, "Show all projects (including archived and templates)")
	showArchived := fs.Bool("archived", false, "Include archived projects")
	showTemplates := fs.Bool("templates", false, "Include template projects")
	
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	// Create client
	client := NewClient(config)

	// Calculate skip value from page
	skip := (*page - 1) * *pageSize
	take := *pageSize

	// Override archived/templates flags if -all is set
	if *all {
		*showArchived = true
		*showTemplates = true
	}

	// Build and execute query
	query := buildProjectQuery(client.GetCompanyID(), *simple, skip, take, *search, *showArchived, *showTemplates)

	// Execute query
	var response ProjectListResponse
	if err := client.ExecuteQueryWithResult(query, nil, &response); err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	// Get project list
	projectList := response.ProjectList

	// Display header
	fmt.Printf("\n=== Projects in %s ===\n", client.GetCompanyID())
	if *search != "" {
		fmt.Printf("Search: '%s'\n", *search)
	}
	fmt.Printf("Page %d of %d (showing %d of %d total)\n\n", 
		*page, 
		(projectList.PageInfo.TotalItems + *pageSize - 1) / *pageSize,
		len(projectList.Items),
		projectList.PageInfo.TotalItems)

	if len(projectList.Items) == 0 {
		fmt.Println("No projects found.")
		return nil
	}

	if *simple {
		// Simple output
		startNum := skip + 1
		for i, project := range projectList.Items {
			fmt.Printf("%d. %s\n   ID: %s\n\n", startNum+i, project.Name, project.ID)
		}
	} else {
		// Detailed output
		startNum := skip + 1
		for i, project := range projectList.Items {
			fmt.Printf("%d. %s\n", startNum+i, project.Name)
			fmt.Printf("   ID: %s\n", project.ID)
			fmt.Printf("   Slug: %s\n", project.Slug)
			fmt.Printf("   Archived: %v\n", project.Archived)
			fmt.Printf("   Template: %v\n", project.IsTemplate)
			if project.Description != "" {
				fmt.Printf("   Description: %s\n", project.Description)
			}
			if project.Color != "" {
				fmt.Printf("   Color: %s\n", project.Color)
			}
			if project.Icon != "" {
				fmt.Printf("   Icon: %s\n", project.Icon)
			}
			fmt.Printf("   Created: %s\n", project.CreatedAt)
			fmt.Printf("   Updated: %s\n", project.UpdatedAt)
			fmt.Println()
		}
	}

	// Show pagination help
	if projectList.PageInfo.HasNextPage || *page > 1 {
		fmt.Println("\n=== Navigation ===")
		if *page > 1 {
			fmt.Printf("Previous page: go run . read-projects -page %d", *page-1)
			if *search != "" {
				fmt.Printf(" -search \"%s\"", *search)
			}
			fmt.Println()
		}
		if projectList.PageInfo.HasNextPage {
			fmt.Printf("Next page: go run . read-projects -page %d", *page+1)
			if *search != "" {
				fmt.Printf(" -search \"%s\"", *search)
			}
			fmt.Println()
		}
	}
	
	return nil
}