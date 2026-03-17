package charts

import (
	"fmt"
	"strings"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a chart in a dashboard",
	Long: `Create a chart within a dashboard. Supports three chart types:

  STAT  - Single number/statistic (e.g., total revenue, record count)
  BAR   - Bar chart grouped by a dimension (e.g., records by assignee)
  PIE   - Pie chart grouped by a dimension (e.g., records by status)

For BAR and PIE charts, use --group-by to set the grouping dimension.
The API automatically generates segments based on the data.`,
	Example: `  # Stat card: count of records
  blue charts create --dashboard <id> --type STAT --title "Total Records" \
    --workspace <id> --function COUNT

  # Stat card: sum of a currency field
  blue charts create --dashboard <id> --type STAT --title "Total Revenue" \
    --workspace <id> --field <field_id> --function SUM --display currency --currency USD

  # Bar chart: records grouped by assignee
  blue charts create --dashboard <id> --type BAR --title "By Assignee" \
    --workspace <id> --group-by ASSIGNEE --function COUNT

  # Pie chart: records grouped by status
  blue charts create --dashboard <id> --type PIE --title "By Status" \
    --workspace <id> --group-by TODO_STATUS --function COUNT

  # Bar chart: sum of field grouped by list
  blue charts create --dashboard <id> --type BAR --title "Points by List" \
    --workspace <id> --group-by TODO_LIST --field <field_id> --function SUM`,
	RunE: runCreate,
}

var (
	createDashboard string
	createType      string
	createTitle     string
	createWorkspace string
	createField     string
	createFunction  string
	createGroupBy   string
	createInterval  string
	createDisplay   string
	createCurrency  string
	createPrecision float64
)

func init() {
	createCmd.Flags().StringVar(&createDashboard, "dashboard", "", "Dashboard ID (required)")
	createCmd.Flags().StringVar(&createType, "type", "", "Chart type: STAT, BAR, or PIE (required)")
	createCmd.Flags().StringVarP(&createTitle, "title", "t", "", "Chart title (required)")
	createCmd.Flags().StringVarP(&createWorkspace, "workspace", "w", "", "Workspace ID for data source (required)")
	createCmd.Flags().StringVar(&createField, "field", "", "Custom field ID to measure (omit for record count)")
	createCmd.Flags().StringVar(&createFunction, "function", "COUNT", "Aggregation: COUNT, COUNTA, SUM, AVERAGE, AVERAGEA, MIN, MAX")
	createCmd.Flags().StringVar(&createGroupBy, "group-by", "", "Group by dimension (BAR/PIE): ASSIGNEE, TAG, TODO_LIST, TODO_STATUS, PROJECT, CUSTOM_FIELD, TODO_DUE_DATE, TODO_CREATED_AT, TODO_UPDATED_AT")
	createCmd.Flags().StringVar(&createInterval, "interval", "MONTH", "Time interval for date grouping: DAY, WEEK, MONTH, QUARTER, YEAR")
	createCmd.Flags().StringVar(&createDisplay, "display", "number", "Display format: number, currency, percentage")
	createCmd.Flags().StringVar(&createCurrency, "currency", "USD", "Currency code (when --display currency)")
	createCmd.Flags().Float64Var(&createPrecision, "precision", 0, "Decimal precision")
}

func runCreate(cmd *cobra.Command, args []string) error {
	if createDashboard == "" {
		return fmt.Errorf("dashboard ID is required. Use --dashboard flag")
	}
	if createType == "" {
		return fmt.Errorf("chart type is required. Use --type flag (STAT, BAR, or PIE)")
	}
	if createTitle == "" {
		return fmt.Errorf("chart title is required. Use --title flag")
	}
	if createWorkspace == "" {
		return fmt.Errorf("workspace ID is required. Use --workspace flag")
	}

	createType = strings.ToUpper(createType)
	if createType != "STAT" && createType != "BAR" && createType != "PIE" {
		return fmt.Errorf("invalid chart type '%s'. Must be STAT, BAR, or PIE", createType)
	}

	createFunction = strings.ToUpper(createFunction)

	if (createType == "BAR" || createType == "PIE") && createGroupBy == "" {
		return fmt.Errorf("--group-by is required for %s charts", createType)
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := common.NewClient(config)
	client.SetProject(createWorkspace)

	// Build display settings
	displayType := strings.ToUpper(createDisplay)
	if displayType == "NUMBER" || displayType == "" {
		displayType = "NUMBER"
	} else if displayType == "CURRENCY" {
		displayType = "CURRENCY"
	} else if displayType == "PERCENTAGE" {
		displayType = "PERCENTAGE"
	}

	var input string
	switch createType {
	case "STAT":
		input = buildStatInput()
	case "BAR":
		input = buildBarInput(displayType)
	case "PIE":
		input = buildPieInput(displayType)
	}

	mutation := fmt.Sprintf(`
		mutation CreateChart {
			createChart(input: %s) {
				id
				title
				type
				isCalculating
				chartSegments {
					id
					title
					formulaResult
				}
			}
		}
	`, input)

	var response struct {
		CreateChart struct {
			ID            string `json:"id"`
			Title         string `json:"title"`
			Type          string `json:"type"`
			IsCalculating bool   `json:"isCalculating"`
			ChartSegments []struct {
				ID            string   `json:"id"`
				Title         string   `json:"title"`
				FormulaResult *float64 `json:"formulaResult"`
			} `json:"chartSegments"`
		} `json:"createChart"`
	}

	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		return fmt.Errorf("failed to create chart: %w", err)
	}

	c := response.CreateChart
	fmt.Printf("Chart created!\n")
	fmt.Printf("ID: %s\n", c.ID)
	fmt.Printf("Title: %s\n", c.Title)
	fmt.Printf("Type: %s\n", c.Type)
	if c.IsCalculating {
		fmt.Printf("Status: Calculating...\n")
	}
	if len(c.ChartSegments) > 0 {
		fmt.Printf("Segments: %d\n", len(c.ChartSegments))
	}

	return nil
}

func buildStatInput() string {
	// Build display
	display := buildDisplayString()

	// Build formula — for a single-source stat, the formula just references the value UID
	uid := "csv-1"
	formula := fmt.Sprintf(`{
		logic: {
			text: "{\"chartSegmentValueUID\": \"%s\"}"
			html: "<p>{\"chartSegmentValueUID\": \"%s\"}</p>"
		}
		display: %s
	}`, uid, uid, display)

	return fmt.Sprintf(`{
		dashboardId: "%s"
		title: "%s"
		type: STAT
		display: %s
		chartSegments: [{
			title: "%s"
			color: "#3B82F6"
			uid: "cs-1"
			formula: %s
			chartSegmentValues: [{
				uid: "%s"
				title: "%s"
				projectId: "%s"
				%s
				function: %s
			}]
		}]
	}`, createDashboard, createTitle, display, createTitle, formula, uid, createTitle,
		createWorkspace, fieldParam(), createFunction)
}

func buildBarInput(displayType string) string {
	display := buildDisplayString()

	// For bar charts, use metadata to let the API auto-generate segments
	xAxisType := strings.ToUpper(createGroupBy)

	// Build interval for date-based grouping
	intervalStr := ""
	if isDateGroupBy(xAxisType) {
		intervalStr = fmt.Sprintf(`interval: %s`, strings.ToUpper(createInterval))
	}

	yAxisField := ""
	if createField != "" {
		yAxisField = fmt.Sprintf(`customFieldName: "%s"`, createField)
	}

	return fmt.Sprintf(`{
		dashboardId: "%s"
		title: "%s"
		type: BAR
		display: %s
		metadata: {
			barChart: {
				xAxis: {
					title: "Category"
					type: %s
					%s
				}
				yAxis: {
					title: "Value"
					function: %s
					%s
					filter: {
						projectIds: ["%s"]
					}
				}
			}
		}
	}`, createDashboard, createTitle, display, xAxisType, intervalStr,
		createFunction, yAxisField, createWorkspace)
}

func buildPieInput(displayType string) string {
	display := buildDisplayString()

	// Pie charts use the same metadata structure as bar charts internally
	groupByType := strings.ToUpper(createGroupBy)

	valueField := ""
	if createField != "" {
		valueField = fmt.Sprintf(`customFieldName: "%s"`, createField)
	}

	return fmt.Sprintf(`{
		dashboardId: "%s"
		title: "%s"
		type: PIE
		display: %s
		metadata: {
			barChart: {
				xAxis: {
					title: "Segment"
					type: %s
				}
				yAxis: {
					title: "Value"
					function: %s
					%s
					filter: {
						projectIds: ["%s"]
					}
				}
			}
		}
	}`, createDashboard, createTitle, display, groupByType,
		createFunction, valueField, createWorkspace)
}

func buildDisplayString() string {
	displayType := strings.ToUpper(createDisplay)
	switch displayType {
	case "CURRENCY":
		return fmt.Sprintf(`{
			type: CURRENCY
			currency: { code: "%s", name: "%s" }
			precision: %g
			function: %s
		}`, createCurrency, createCurrency, createPrecision, createFunction)
	case "PERCENTAGE":
		return fmt.Sprintf(`{
			type: PERCENTAGE
			precision: %g
			function: %s
		}`, createPrecision, createFunction)
	default:
		return fmt.Sprintf(`{
			type: NUMBER
			precision: %g
			function: %s
		}`, createPrecision, createFunction)
	}
}

func fieldParam() string {
	if createField != "" {
		return fmt.Sprintf(`customFieldId: "%s"`, createField)
	}
	return ""
}

func isDateGroupBy(groupBy string) bool {
	return groupBy == "TODO_DUE_DATE" || groupBy == "TODO_CREATED_AT" || groupBy == "TODO_UPDATED_AT"
}
