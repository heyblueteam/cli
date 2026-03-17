package workspaces

import (
	"fmt"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

// ProjectList represents a paginated list of projects
type ProjectList struct {
	Items    []common.Project      `json:"items"`
	PageInfo common.OffsetPageInfo `json:"pageInfo"`
}

type ProjectListResponse struct {
	ProjectList ProjectList `json:"projectList"`
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List workspaces",
	Long:  "List workspaces in your Blue company with pagination, search, and filtering.",
	Example: `  blue workspaces list
  blue workspaces list --simple
  blue workspaces list --search "CRM" --sort updatedAt_DESC
  blue workspaces list --page 2 --size 50
  blue workspaces list --all --archived --templates`,
	RunE: runList,
}

var (
	listSimple        bool
	listPage          int
	listPageSize      int
	listSearch        string
	listSortBy        string
	listAll           bool
	listShowArchived  bool
	listShowTemplates bool
)

func init() {
	listCmd.Flags().BoolVarP(&listSimple, "simple", "s", false, "Show only workspace names and IDs")
	listCmd.Flags().IntVar(&listPage, "page", 1, "Page number")
	listCmd.Flags().IntVar(&listPageSize, "size", 20, "Page size")
	listCmd.Flags().StringVar(&listSearch, "search", "", "Search workspaces by name")
	listCmd.Flags().StringVar(&listSortBy, "sort", "name_ASC", "Sort by field (name_ASC, name_DESC, createdAt_ASC, createdAt_DESC, updatedAt_ASC, updatedAt_DESC, position_ASC, position_DESC)")
	listCmd.Flags().BoolVar(&listAll, "all", false, "Show all workspaces (including archived and templates)")
	listCmd.Flags().BoolVar(&listShowArchived, "archived", false, "Include archived workspaces")
	listCmd.Flags().BoolVar(&listShowTemplates, "templates", false, "Include template workspaces")
}

func runList(cmd *cobra.Command, args []string) error {
	// Validate sort option
	validSortOptions := []string{
		"name_ASC", "name_DESC",
		"createdAt_ASC", "createdAt_DESC",
		"updatedAt_ASC", "updatedAt_DESC",
		"position_ASC", "position_DESC",
	}
	isValidSort := false
	for _, validSort := range validSortOptions {
		if listSortBy == validSort {
			isValidSort = true
			break
		}
	}
	if !isValidSort {
		return fmt.Errorf("invalid sort option '%s'. Valid options: %v", listSortBy, validSortOptions)
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)

	skip := (listPage - 1) * listPageSize
	take := listPageSize

	if listAll {
		listShowArchived = true
		listShowTemplates = true
	}

	query := buildProjectQuery(client.GetCompanyID(), listSimple, skip, take, listSearch, listShowArchived, listShowTemplates, listSortBy)

	var response ProjectListResponse
	if err := client.ExecuteQueryWithResult(query, nil, &response); err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	projectList := response.ProjectList

	fmt.Printf("\n=== Workspaces in %s ===\n", client.GetCompanyID())
	if listSearch != "" {
		fmt.Printf("Search: '%s'\n", listSearch)
	}
	if listSortBy != "name_ASC" {
		fmt.Printf("Sort: %s\n", listSortBy)
	}
	if listShowArchived && !listAll {
		fmt.Printf("Filter: Including archived workspaces\n")
	}
	if listShowTemplates && !listAll {
		fmt.Printf("Filter: Including template workspaces\n")
	}
	if listAll {
		fmt.Printf("Filter: All workspaces (including archived and templates)\n")
	}

	totalPages := 1
	if listPageSize > 0 && projectList.PageInfo.TotalItems > 0 {
		totalPages = (projectList.PageInfo.TotalItems + listPageSize - 1) / listPageSize
	}
	fmt.Printf("Page %d of %d (showing %d of %d total)\n\n",
		listPage, totalPages, len(projectList.Items), projectList.PageInfo.TotalItems)

	if len(projectList.Items) == 0 {
		fmt.Println("No workspaces found.")
		return nil
	}

	if listSimple {
		startNum := skip + 1
		for i, project := range projectList.Items {
			fmt.Printf("%d. %s\n   ID: %s\n\n", startNum+i, project.Name, project.ID)
		}
	} else {
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

	if projectList.PageInfo.HasNextPage || listPage > 1 {
		fmt.Println("\n=== Navigation ===")
		if listPage > 1 {
			fmt.Printf("Previous page: blue workspaces list --page %d", listPage-1)
			if listSearch != "" {
				fmt.Printf(" --search \"%s\"", listSearch)
			}
			fmt.Println()
		}
		if projectList.PageInfo.HasNextPage {
			fmt.Printf("Next page: blue workspaces list --page %d", listPage+1)
			if listSearch != "" {
				fmt.Printf(" --search \"%s\"", listSearch)
			}
			fmt.Println()
		}
	}

	return nil
}

func buildProjectQuery(companyID string, simple bool, skip int, take int, search string, showArchived bool, showTemplates bool, sortBy string) string {
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
			sort: [%s]
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
	}`, filter, skip, take, sortBy, fields)

	return query
}
