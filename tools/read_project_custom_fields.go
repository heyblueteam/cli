package tools

import (
	"flag"
	"fmt"
	
	"cli/common"
)

// CustomField, Project, and CustomFieldOption are already defined in common/types.go
// Note: This file may need adjustments if the CustomField in common/types.go
// doesn't have all the fields needed (like SequenceDigits, ReferenceProject, etc.)

// PageInfo represents pagination information
type PageInfo struct {
	TotalPages      int  `json:"totalPages"`
	TotalItems      int  `json:"totalItems"`
	Page            int  `json:"page"`
	PerPage         int  `json:"perPage"`
	HasNextPage     bool `json:"hasNextPage"`
	HasPreviousPage bool `json:"hasPreviousPage"`
}

// CustomFieldPagination represents a paginated list of custom fields
type CustomFieldPagination struct {
	Items    []common.CustomField `json:"items"`
	PageInfo PageInfo      `json:"pageInfo"`
}

// CustomFieldsResponse represents the response from the GraphQL query
type CustomFieldsResponse struct {
	CustomFields CustomFieldPagination `json:"customFields"`
}

// Build query to fetch custom fields for a project
func buildQuery(projectID string, skip int, take int) string {
	query := fmt.Sprintf(`query {
		customFields(
			filter: { projectId: "%s" }
			skip: %d
			take: %d
		) {
			items {
				id
				name
				type
				position
				description
				createdAt
				updatedAt
				customFieldOptions {
					id
					title
					color
				}
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
	}`, projectID, skip, take)

	return query
}

func RunReadProjectCustomFields(args []string) error {
	fs := flag.NewFlagSet("read-project-custom-fields", flag.ExitOnError)
	projectID := fs.String("project", "", "Project ID (required)")
	page := fs.Int("page", 1, "Page number (default: 1)")
	pageSize := fs.Int("size", 50, "Page size (default: 50)")
	simple := fs.Bool("simple", false, "Show only basic custom field information")
	fs.Parse(args)

	// Validate required parameters
	if *projectID == "" {
		return fmt.Errorf("project ID is required. Use -project flag")
	}

	// Load configuration
	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Create client
	client := common.NewClient(config)

	// Calculate skip value from page
	skip := (*page - 1) * *pageSize
	take := *pageSize

	// Build and execute query
	query := buildQuery(*projectID, skip, take)

	// Execute query
	var response CustomFieldsResponse
	if err := client.ExecuteQueryWithResult(query, nil, &response); err != nil {
		return fmt.Errorf("failed to execute query: %v", err)
	}

	// Get custom fields list
	customFields := response.CustomFields

	// Display header
	fmt.Printf("\n=== Custom Fields in Project %s ===\n", *projectID)
	fmt.Printf("Page %d of %d (showing %d of %d total)\n\n", 
		*page, 
		(customFields.PageInfo.TotalItems + *pageSize - 1) / *pageSize,
		len(customFields.Items),
		customFields.PageInfo.TotalItems)

	if len(customFields.Items) == 0 {
		fmt.Println("No custom fields found in this project.")
		return nil
	}

	if *simple {
		// Simple output
		startNum := skip + 1
		for i, field := range customFields.Items {
			fmt.Printf("%d. %s (%s)\n   ID: %s\n   Position: %.0f\n", 
				startNum+i, field.Name, field.Type, field.ID, field.Position)
			
			// Show options for SELECT fields
			if (field.Type == "SELECT_SINGLE" || field.Type == "SELECT_MULTI") && len(field.Options) > 0 {
				fmt.Printf("   Options: ")
				for j, option := range field.Options {
					if j > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s [%s]", option.Title, option.ID)
					if option.Color != "" {
						fmt.Printf(" (%s)", option.Color)
					}
				}
				fmt.Printf("\n")
			}
			fmt.Printf("\n")
		}
	} else {
		// Detailed output
		startNum := skip + 1
		for i, field := range customFields.Items {
			fmt.Printf("%d. %s\n", startNum+i, field.Name)
			fmt.Printf("   ID: %s\n", field.ID)
			fmt.Printf("   Type: %s\n", field.Type)
			fmt.Printf("   Position: %.0f\n", field.Position)
			
			if field.Description != "" {
				fmt.Printf("   Description: %s\n", field.Description)
			}
			
			// Show options for SELECT fields
			if (field.Type == "SELECT_SINGLE" || field.Type == "SELECT_MULTI") && len(field.Options) > 0 {
				fmt.Printf("   Available Options (use Option ID or Title for record values):\n")
				for _, option := range field.Options {
					fmt.Printf("     - %s [%s]", option.Title, option.ID)
					if option.Color != "" {
						fmt.Printf(" (%s)", option.Color)
					}
					fmt.Printf("\n")
				}
			}
			
			fmt.Printf("   Created: %s\n", field.CreatedAt)
			fmt.Printf("   Updated: %s\n", field.UpdatedAt)
			fmt.Println()
		}
	}

	// Show pagination help
	if customFields.PageInfo.HasNextPage || *page > 1 {
		fmt.Println("\n=== Navigation ===")
		if *page > 1 {
			fmt.Printf("Previous page: go run auth.go list-project-custom-fields.go -project %s -page %d", *projectID, *page-1)
			if *simple {
				fmt.Printf(" -simple")
			}
			fmt.Println()
		}
		if customFields.PageInfo.HasNextPage {
			fmt.Printf("Next page: go run auth.go list-project-custom-fields.go -project %s -page %d", *projectID, *page+1)
			if *simple {
				fmt.Printf(" -simple")
			}
			fmt.Println()
		}
	}

	return nil
}