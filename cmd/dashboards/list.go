package dashboards

import (
	"fmt"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

type DashboardUser struct {
	ID   string `json:"id"`
	Role string `json:"role"`
	User struct {
		ID       string `json:"id"`
		FullName string `json:"fullName"`
	} `json:"user"`
}

type DashboardItem struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	CreatedBy struct {
		ID       string `json:"id"`
		FullName string `json:"fullName"`
	} `json:"createdBy"`
	DashboardUsers []DashboardUser `json:"dashboardUsers"`
}

type DashboardsResponse struct {
	Dashboards struct {
		Items    []DashboardItem `json:"items"`
		PageInfo struct {
			HasNextPage bool `json:"hasNextPage"`
			TotalItems  int  `json:"totalItems"`
		} `json:"pageInfo"`
	} `json:"dashboards"`
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List dashboards",
	Long:  "List dashboards in a company, optionally filtered by workspace.",
	Example: `  blue dashboards list
  blue dashboards list --workspace <id>
  blue dashboards list --simple`,
	RunE: runList,
}

var (
	listWorkspace string
	listSimple    bool
	listPage      int
	listSize      int
)

func init() {
	listCmd.Flags().StringVarP(&listWorkspace, "workspace", "w", "", "Filter by workspace ID (optional)")
	listCmd.Flags().BoolVarP(&listSimple, "simple", "s", false, "Simple output format")
	listCmd.Flags().IntVar(&listPage, "page", 1, "Page number")
	listCmd.Flags().IntVar(&listSize, "size", 20, "Page size")
}

func runList(cmd *cobra.Command, args []string) error {
	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := common.NewClient(config)

	skip := (listPage - 1) * listSize

	query := `
		query Dashboards($filter: DashboardFilterInput!, $skip: Int, $take: Int) {
			dashboards(filter: $filter, skip: $skip, take: $take) {
				items {
					id
					title
					createdAt
					updatedAt
					createdBy {
						id
						fullName
					}
					dashboardUsers {
						id
						role
						user {
							id
							fullName
						}
					}
				}
				pageInfo {
					hasNextPage
					totalItems
				}
			}
		}
	`

	companyID, err := client.ResolveCompanyID()
	if err != nil {
		return fmt.Errorf("failed to resolve company: %w", err)
	}

	filter := map[string]interface{}{
		"companyId": companyID,
	}
	if listWorkspace != "" {
		filter["projectId"] = listWorkspace
	}

	variables := map[string]interface{}{
		"filter": filter,
		"skip":   skip,
		"take":   listSize,
	}

	var response DashboardsResponse
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to list dashboards: %w", err)
	}

	items := response.Dashboards.Items
	total := response.Dashboards.PageInfo.TotalItems

	fmt.Printf("\n=== Dashboards ===\n")
	fmt.Printf("Page %d (showing %d of %d total)\n\n", listPage, len(items), total)

	if len(items) == 0 {
		fmt.Println("No dashboards found.")
		return nil
	}

	for i, d := range items {
		num := skip + i + 1
		if listSimple {
			fmt.Printf("%d. %s\n   ID: %s\n\n", num, d.Title, d.ID)
		} else {
			fmt.Printf("%d. %s\n", num, d.Title)
			fmt.Printf("   ID: %s\n", d.ID)
			fmt.Printf("   Created by: %s\n", d.CreatedBy.FullName)
			if len(d.DashboardUsers) > 0 {
				fmt.Printf("   Shared with: ")
				for j, u := range d.DashboardUsers {
					if j > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s (%s)", u.User.FullName, u.Role)
				}
				fmt.Println()
			}
			fmt.Printf("   Updated: %s\n", d.UpdatedAt)
			fmt.Println()
		}
	}

	return nil
}
