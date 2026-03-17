package charts

import (
	"fmt"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

type ChartSegment struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	FormulaResult *float64 `json:"formulaResult"`
}

type ChartListItem struct {
	ID              string         `json:"id"`
	Title           string         `json:"title"`
	Type            string         `json:"type"`
	Position        float64        `json:"position"`
	IsCalculating   bool           `json:"isCalculating"`
	NeedCalculation bool           `json:"needCalculation"`
	ChartSegments   []ChartSegment `json:"chartSegments"`
	CreatedAt       string         `json:"createdAt"`
	UpdatedAt       string         `json:"updatedAt"`
}

type ChartsListResponse struct {
	Charts struct {
		Items    []ChartListItem `json:"items"`
		PageInfo struct {
			HasNextPage bool `json:"hasNextPage"`
			TotalItems  int  `json:"totalItems"`
		} `json:"pageInfo"`
	} `json:"charts"`
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List charts in a dashboard",
	Long:  "List all charts within a dashboard with their types and segment counts.",
	Example: `  blue charts list --dashboard <id>
  blue charts list --dashboard <id> --simple`,
	RunE: runList,
}

var (
	listDashboard string
	listSimple    bool
)

func init() {
	listCmd.Flags().StringVar(&listDashboard, "dashboard", "", "Dashboard ID (required)")
	listCmd.Flags().BoolVarP(&listSimple, "simple", "s", false, "Simple output format")
}

func runList(cmd *cobra.Command, args []string) error {
	if listDashboard == "" {
		return fmt.Errorf("dashboard ID is required. Use --dashboard flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := common.NewClient(config)

	query := `
		query Charts($filter: ChartFilterInput!) {
			charts(filter: $filter, sort: [position_ASC], take: 100) {
				items {
					id
					title
					type
					position
					isCalculating
					needCalculation
					chartSegments {
						id
						title
						formulaResult
					}
					createdAt
					updatedAt
				}
				pageInfo {
					hasNextPage
					totalItems
				}
			}
		}
	`

	variables := map[string]interface{}{
		"filter": map[string]interface{}{
			"dashboardId": listDashboard,
		},
	}

	var response ChartsListResponse
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to list charts: %w", err)
	}

	items := response.Charts.Items
	total := response.Charts.PageInfo.TotalItems

	fmt.Printf("\n=== Charts in Dashboard ===\n")
	fmt.Printf("Total: %d\n\n", total)

	if len(items) == 0 {
		fmt.Println("No charts found.")
		return nil
	}

	for i, c := range items {
		status := ""
		if c.IsCalculating {
			status = " (calculating...)"
		} else if c.NeedCalculation {
			status = " (needs recalculation)"
		}

		if listSimple {
			fmt.Printf("%d. %s [%s]%s\n   ID: %s\n\n", i+1, c.Title, c.Type, status, c.ID)
		} else {
			fmt.Printf("%d. %s [%s]%s\n", i+1, c.Title, c.Type, status)
			fmt.Printf("   ID: %s\n", c.ID)
			fmt.Printf("   Segments: %d\n", len(c.ChartSegments))
			for _, seg := range c.ChartSegments {
				if seg.FormulaResult != nil {
					fmt.Printf("     - %s: %g\n", seg.Title, *seg.FormulaResult)
				} else {
					fmt.Printf("     - %s: (pending)\n", seg.Title)
				}
			}
			fmt.Println()
		}
	}

	return nil
}
