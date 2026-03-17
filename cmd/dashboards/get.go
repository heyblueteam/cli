package dashboards

import (
	"fmt"
	"math"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

type ChartSegmentValue struct {
	ID            string  `json:"id"`
	UID           string  `json:"uid"`
	Title         string  `json:"title"`
	ProjectID     string  `json:"projectId"`
	CustomFieldID *string `json:"customFieldId"`
	Function      string  `json:"function"`
}

type ChartSegment struct {
	ID                 string              `json:"id"`
	UID                string              `json:"uid"`
	Title              string              `json:"title"`
	Color              string              `json:"color"`
	FormulaResult      *float64            `json:"formulaResult"`
	ChartSegmentValues []ChartSegmentValue `json:"chartSegmentValues"`
	Formula            *struct {
		Display *struct {
			Type      string `json:"type"`
			Precision *float64 `json:"precision"`
			Currency  *struct {
				Code string `json:"code"`
				Name string `json:"name"`
			} `json:"currency"`
			Function string `json:"function"`
		} `json:"display"`
	} `json:"formula"`
}

type ChartItem struct {
	ID                       string         `json:"id"`
	Title                    string         `json:"title"`
	Position                 float64        `json:"position"`
	Type                     string         `json:"type"`
	IsCalculating            bool           `json:"isCalculating"`
	IsCalculatingWithFilter  bool           `json:"isCalculatingWithFilter"`
	NeedCalculation          bool           `json:"needCalculation"`
	ChartSegments            []ChartSegment `json:"chartSegments"`
	Display                  *struct {
		Type      string `json:"type"`
		Precision *float64 `json:"precision"`
		Currency  *struct {
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"currency"`
		Function string `json:"function"`
	} `json:"display"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type ChartsResponse struct {
	Charts struct {
		Items    []ChartItem `json:"items"`
		PageInfo struct {
			HasNextPage bool `json:"hasNextPage"`
			TotalItems  int  `json:"totalItems"`
		} `json:"pageInfo"`
	} `json:"charts"`
}

type DashboardResponse struct {
	Dashboard DashboardItem `json:"dashboard"`
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "View a dashboard with chart data",
	Long:  "Get a dashboard and display all its charts with their current values.",
	Example: `  blue dashboards get --dashboard <id>
  blue dashboards get --dashboard <id> --simple`,
	RunE: runGet,
}

var (
	getDashboard string
	getSimple    bool
)

func init() {
	getCmd.Flags().StringVar(&getDashboard, "dashboard", "", "Dashboard ID (required)")
	getCmd.Flags().BoolVarP(&getSimple, "simple", "s", false, "Simple output format")
}

func runGet(cmd *cobra.Command, args []string) error {
	if getDashboard == "" {
		return fmt.Errorf("dashboard ID is required. Use --dashboard flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := common.NewClient(config)

	// Fetch dashboard metadata
	dashQuery := fmt.Sprintf(`
		query Dashboard {
			dashboard(id: "%s") {
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
		}
	`, getDashboard)

	var dashResponse DashboardResponse
	if err := client.ExecuteQueryWithResult(dashQuery, nil, &dashResponse); err != nil {
		return fmt.Errorf("failed to get dashboard: %w", err)
	}

	dash := dashResponse.Dashboard
	if dash.ID == "" {
		return fmt.Errorf("dashboard '%s' not found", getDashboard)
	}

	// Fetch charts
	chartsQuery := fmt.Sprintf(`
		query Charts {
			charts(
				filter: { dashboardId: "%s" }
				sort: [position_ASC]
				take: 100
			) {
				items {
					id
					title
					position
					type
					isCalculating
					isCalculatingWithFilter
					needCalculation
					display {
						type
						precision
						currency {
							code
							name
						}
						function
					}
					chartSegments {
						id
						uid
						title
						color
						formulaResult
						formula {
							display {
								type
								precision
								currency {
									code
									name
								}
								function
							}
						}
						chartSegmentValues {
							id
							uid
							title
							projectId
							customFieldId
							function
						}
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
	`, getDashboard)

	var chartsResponse ChartsResponse
	if err := client.ExecuteQueryWithResult(chartsQuery, nil, &chartsResponse); err != nil {
		return fmt.Errorf("failed to get charts: %w", err)
	}

	charts := chartsResponse.Charts.Items

	// Display dashboard
	fmt.Printf("\n=== %s ===\n", dash.Title)
	if !getSimple {
		fmt.Printf("ID: %s\n", dash.ID)
		fmt.Printf("Created by: %s\n", dash.CreatedBy.FullName)
		if len(dash.DashboardUsers) > 0 {
			fmt.Printf("Shared with: ")
			for i, u := range dash.DashboardUsers {
				if i > 0 {
					fmt.Printf(", ")
				}
				fmt.Printf("%s (%s)", u.User.FullName, u.Role)
			}
			fmt.Println()
		}
	}
	fmt.Printf("Charts: %d\n\n", len(charts))

	if len(charts) == 0 {
		fmt.Println("No charts in this dashboard.")
		return nil
	}

	// Display each chart with values
	for i, chart := range charts {
		status := ""
		if chart.IsCalculating || chart.IsCalculatingWithFilter {
			status = " (calculating...)"
		} else if chart.NeedCalculation {
			status = " (needs recalculation)"
		}

		fmt.Printf("%d. %s [%s]%s\n", i+1, chart.Title, chart.Type, status)

		if !getSimple {
			fmt.Printf("   ID: %s\n", chart.ID)
		}

		if len(chart.ChartSegments) == 0 {
			fmt.Printf("   (no data)\n\n")
			continue
		}

		switch chart.Type {
		case "STAT":
			displayStatChart(chart)
		case "BAR":
			displayBarChart(chart, getSimple)
		case "PIE":
			displayPieChart(chart, getSimple)
		default:
			displayGenericChart(chart, getSimple)
		}

		fmt.Println()
	}

	return nil
}

func displayStatChart(chart ChartItem) {
	for _, seg := range chart.ChartSegments {
		if seg.FormulaResult != nil {
			fmt.Printf("   Value: %s\n", formatValue(*seg.FormulaResult, chart.Display, seg.Formula))
		} else {
			fmt.Printf("   Value: (pending)\n")
		}
	}
}

func displayBarChart(chart ChartItem, simple bool) {
	// Find max value for bar rendering
	maxVal := 0.0
	for _, seg := range chart.ChartSegments {
		if seg.FormulaResult != nil && *seg.FormulaResult > maxVal {
			maxVal = *seg.FormulaResult
		}
	}

	for _, seg := range chart.ChartSegments {
		val := "(pending)"
		barWidth := 0
		if seg.FormulaResult != nil {
			val = formatValue(*seg.FormulaResult, chart.Display, seg.Formula)
			if maxVal > 0 {
				barWidth = int((*seg.FormulaResult / maxVal) * 30)
			}
		}

		bar := ""
		for j := 0; j < barWidth; j++ {
			bar += "█"
		}

		if simple {
			fmt.Printf("   %s: %s\n", seg.Title, val)
		} else {
			fmt.Printf("   %s %s %s\n", bar, val, seg.Title)
		}
	}
}

func displayPieChart(chart ChartItem, simple bool) {
	// Calculate total for percentages
	total := 0.0
	for _, seg := range chart.ChartSegments {
		if seg.FormulaResult != nil {
			total += *seg.FormulaResult
		}
	}

	for _, seg := range chart.ChartSegments {
		if seg.FormulaResult != nil {
			pct := 0.0
			if total > 0 {
				pct = (*seg.FormulaResult / total) * 100
			}
			val := formatValue(*seg.FormulaResult, chart.Display, seg.Formula)
			if simple {
				fmt.Printf("   %s: %s (%.0f%%)\n", seg.Title, val, pct)
			} else {
				fmt.Printf("   %s: %s (%.1f%%)\n", seg.Title, val, pct)
			}
		} else {
			fmt.Printf("   %s: (pending)\n", seg.Title)
		}
	}
}

func displayGenericChart(chart ChartItem, simple bool) {
	for _, seg := range chart.ChartSegments {
		if seg.FormulaResult != nil {
			val := formatValue(*seg.FormulaResult, chart.Display, seg.Formula)
			fmt.Printf("   %s: %s\n", seg.Title, val)
		} else {
			fmt.Printf("   %s: (pending)\n", seg.Title)
		}
	}
}

type displayConfig struct {
	Type      string
	Precision *float64
	Currency  *struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}
}

func formatValue(val float64, chartDisplay *struct {
	Type      string   `json:"type"`
	Precision *float64 `json:"precision"`
	Currency  *struct {
		Code string `json:"code"`
		Name string `json:"name"`
	} `json:"currency"`
	Function string `json:"function"`
}, segFormula *struct {
	Display *struct {
		Type      string   `json:"type"`
		Precision *float64 `json:"precision"`
		Currency  *struct {
			Code string `json:"code"`
			Name string `json:"name"`
		} `json:"currency"`
		Function string `json:"function"`
	} `json:"display"`
}) string {
	// Use segment formula display if available, else chart display
	displayType := "NUMBER"
	var precision float64 = 0
	currencyCode := ""

	if segFormula != nil && segFormula.Display != nil {
		displayType = segFormula.Display.Type
		if segFormula.Display.Precision != nil {
			precision = *segFormula.Display.Precision
		}
		if segFormula.Display.Currency != nil {
			currencyCode = segFormula.Display.Currency.Code
		}
	} else if chartDisplay != nil {
		displayType = chartDisplay.Type
		if chartDisplay.Precision != nil {
			precision = *chartDisplay.Precision
		}
		if chartDisplay.Currency != nil {
			currencyCode = chartDisplay.Currency.Code
		}
	}

	// Round to precision
	if precision > 0 {
		factor := math.Pow(10, precision)
		val = math.Round(val*factor) / factor
	}

	switch displayType {
	case "CURRENCY":
		if currencyCode != "" {
			return fmt.Sprintf("%s %."+fmt.Sprintf("%.0f", precision)+"f", currencyCode, val)
		}
		return fmt.Sprintf("%."+fmt.Sprintf("%.0f", precision)+"f", val)
	case "PERCENTAGE":
		return fmt.Sprintf("%."+fmt.Sprintf("%.0f", precision)+"f%%", val)
	default:
		if precision > 0 {
			return fmt.Sprintf("%."+fmt.Sprintf("%.0f", precision)+"f", val)
		}
		// Clean integer display
		if val == math.Trunc(val) {
			return fmt.Sprintf("%.0f", val)
		}
		return fmt.Sprintf("%g", val)
	}
}
