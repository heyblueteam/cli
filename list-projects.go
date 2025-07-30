package main

import (
	"flag"
	"fmt"
	"log"
)

// Response structures
type Project struct {
	ID          string    `json:"id"`
	UID         string    `json:"uid"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Archived    bool      `json:"archived"`
	Color       string    `json:"color"`
	Icon        string    `json:"icon"`
	CreatedAt   string    `json:"createdAt"`
	UpdatedAt   string    `json:"updatedAt"`
	Position    float64   `json:"position"`
	IsTemplate  bool      `json:"isTemplate"`
}

type PageInfo struct {
	TotalPages      int  `json:"totalPages"`
	TotalItems      int  `json:"totalItems"`
	Page            int  `json:"page"`
	PerPage         int  `json:"perPage"`
	HasNextPage     bool `json:"hasNextPage"`
	HasPreviousPage bool `json:"hasPreviousPage"`
}

type ProjectList struct {
	Items    []Project `json:"items"`
	PageInfo PageInfo  `json:"pageInfo"`
}

type ProjectListResponse struct {
	ProjectList ProjectList `json:"projectList"`
}

// Queries
const (
	fullQuery = `query ProjectListQuery {
		projectList(filter: { companyIds: ["%s"] }) {
			items {
				id
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
				isTemplate
			}
			pageInfo {
				totalPages
				totalItems
				page
				perPage
				hasNextPage
				hasPreviousPage
			}
		}
	}`

	simpleQuery = `query ProjectListQuery {
		projectList(filter: { companyIds: ["%s"] }) {
			items {
				id
				name
			}
			pageInfo {
				totalItems
			}
		}
	}`
)

func main() {
	// Parse command line flags
	simple := flag.Bool("simple", false, "Show only project names and IDs")
	flag.Parse()

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create client
	client := NewClient(config)

	// Select query based on flag
	var query string
	if *simple {
		query = fmt.Sprintf(simpleQuery, client.GetCompanyID())
	} else {
		query = fmt.Sprintf(fullQuery, client.GetCompanyID())
	}

	// Execute query
	var response ProjectListResponse
	if err := client.ExecuteQueryWithResult(query, nil, &response); err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}

	// Get project list
	projectList := response.ProjectList

	fmt.Printf("\n=== Projects in %s ===\n", client.GetCompanyID())
	fmt.Printf("Total projects: %d\n\n", projectList.PageInfo.TotalItems)

	if *simple {
		// Simple output
		for i, project := range projectList.Items {
			fmt.Printf("%d. %s\n   ID: %s\n\n", i+1, project.Name, project.ID)
		}
	} else {
		// Detailed output
		for i, project := range projectList.Items {
			fmt.Printf("%d. %s\n", i+1, project.Name)
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

	// Show pagination info if there are more pages
	if projectList.PageInfo.HasNextPage {
		fmt.Printf("\nNote: Showing page %d of %d. Total items: %d\n", 
			projectList.PageInfo.Page, 
			projectList.PageInfo.TotalPages, 
			projectList.PageInfo.TotalItems)
	}
}