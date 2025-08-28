package main

import (
	"flag"
	"fmt"
	"log"
)

// CustomField represents a custom field in the system
type CustomField struct {
	ID                    string                 `json:"id"`
	UID                   string                 `json:"uid"`
	Name                  string                 `json:"name"`
	Type                  string                 `json:"type"`
	Position              float64                `json:"position"`
	ButtonType            string                 `json:"buttonType"`
	ButtonConfirmText     string                 `json:"buttonConfirmText"`
	CurrencyFieldID       string                 `json:"currencyFieldId"`
	ConversionDateType    string                 `json:"conversionDateType"`
	ConversionDate        string                 `json:"conversionDate"`
	Description           string                 `json:"description"`
	Min                   *float64               `json:"min"`
	Max                   *float64               `json:"max"`
	Latitude              *float64               `json:"latitude"`
	Longitude             *float64               `json:"longitude"`
	StartDate             string                 `json:"startDate"`
	EndDate               string                 `json:"endDate"`
	Timezone              string                 `json:"timezone"`
	Currency              string                 `json:"currency"`
	Prefix                string                 `json:"prefix"`
	IsDueDate             bool                   `json:"isDueDate"`
	Formula               interface{}            `json:"formula"`
	CreatedAt             string                 `json:"createdAt"`
	UpdatedAt             string                 `json:"updatedAt"`
	RegionCode            string                 `json:"regionCode"`
	CountryCodes          []string               `json:"countryCodes"`
	Text                  string                 `json:"text"`
	Number                *float64               `json:"number"`
	Checked               bool                   `json:"checked"`
	Editable              bool                   `json:"editable"`
	UseSequenceUniqueID   bool                   `json:"useSequenceUniqueId"`
	SequenceDigits        *int                   `json:"sequenceDigits"`
	SequenceStartingNumber *int                  `json:"sequenceStartingNumber"`
	SequenceID            *int                   `json:"sequenceId"`
	ReferenceProject      *Project               `json:"referenceProject"`
	ReferenceFilter       interface{}            `json:"referenceFilter"`
	ReferenceMultiple     bool                   `json:"referenceMultiple"`
	Project               Project                 `json:"project"`
}

// Project represents a project in the system
type Project struct {
	ID          string `json:"id"`
	UID         string `json:"uid"`
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Archived    bool   `json:"archived"`
	Color       string `json:"color"`
	Icon        string `json:"icon"`
}

// CustomFieldOption represents a custom field option
type CustomFieldOption struct {
	ID                    string `json:"id"`
	Title                 string `json:"title"`
	ButtonType            string `json:"buttonType"`
	ButtonConfirmText     string `json:"buttonConfirmText"`
	Color                 string `json:"color"`
	CurrencyConversionTo  string `json:"currencyConversionTo"`
}

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
	Items    []CustomField `json:"items"`
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

func main() {
	// Parse command line flags
	projectID := flag.String("project", "", "Project ID (required)")
	page := flag.Int("page", 1, "Page number (default: 1)")
	pageSize := flag.Int("size", 50, "Page size (default: 50)")
	simple := flag.Bool("simple", false, "Show only basic custom field information")
	flag.Parse()

	// Validate required parameters
	if *projectID == "" {
		log.Fatal("Project ID is required. Use -project flag")
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create client
	client := NewClient(config)

	// Calculate skip value from page
	skip := (*page - 1) * *pageSize
	take := *pageSize

	// Build and execute query
	query := buildQuery(*projectID, skip, take)

	// Execute query
	var response CustomFieldsResponse
	if err := client.ExecuteQueryWithResult(query, nil, &response); err != nil {
		log.Fatalf("Failed to execute query: %v", err)
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
		return
	}

	if *simple {
		// Simple output
		startNum := skip + 1
		for i, field := range customFields.Items {
			fmt.Printf("%d. %s (%s)\n   ID: %s\n   Position: %.0f\n\n", 
				startNum+i, field.Name, field.Type, field.ID, field.Position)
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
}
