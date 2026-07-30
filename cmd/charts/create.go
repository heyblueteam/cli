package charts

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/heyblueteam/cli/common"

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
The API automatically generates segments based on the data.

--field accepts a custom field ID or its exact name. Grouping by CUSTOM_FIELD
only works for field types the API can group on: SELECT_SINGLE, SELECT_MULTI,
CHECKBOX, COUNTRY, DATE, REFERENCE, REFERENCED_BY, ASSIGNEE.`,
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

  # Bar chart: sum of a field grouped by list
  blue charts create --dashboard <id> --type BAR --title "Points by List" \
    --workspace <id> --group-by TODO_LIST --field "Story Points" --function SUM

  # Bar chart: grouped by a select field
  blue charts create --dashboard <id> --type BAR --title "By Priority" \
    --workspace <id> --group-by CUSTOM_FIELD --group-by-field "Priority"`,
	RunE: runCreate,
}

var (
	createDashboard   string
	createType        string
	createTitle       string
	createWorkspace   string
	createField       string
	createGroupByCF   string
	createFunction    string
	createGroupBy     string
	createInterval    string
	createDisplay     string
	createCurrency    string
	createPrecision   float64
	createRenderStyle string
)

func init() {
	createCmd.Flags().StringVar(&createDashboard, "dashboard", "", "Dashboard ID (required)")
	createCmd.Flags().StringVar(&createType, "type", "", "Chart type: STAT, BAR, or PIE (required)")
	createCmd.Flags().StringVarP(&createTitle, "title", "t", "", "Chart title (required)")
	createCmd.Flags().StringVarP(&createWorkspace, "workspace", "w", "", "Workspace ID or slug for the data source (required)")
	createCmd.Flags().StringVar(&createField, "field", "", "Custom field ID or name to measure (omit to count records)")
	createCmd.Flags().StringVar(&createGroupByCF, "group-by-field", "", "Custom field ID or name to group by (with --group-by CUSTOM_FIELD)")
	createCmd.Flags().StringVar(&createFunction, "function", "COUNT", "Aggregation: COUNT, COUNTA, SUM, AVERAGE, AVERAGEA, MIN, MAX")
	createCmd.Flags().StringVar(&createGroupBy, "group-by", "", "Group by dimension (BAR/PIE): ASSIGNEE, TAG, TODO_LIST, TODO_STATUS, PROJECT, CUSTOM_FIELD, TODO_DUE_DATE, TODO_CREATED_AT, TODO_UPDATED_AT, TODO_COMPLETED_AT")
	createCmd.Flags().StringVar(&createInterval, "interval", "MONTH", "Time interval for date grouping: DAY, WEEK, MONTH, QUARTER, YEAR")
	createCmd.Flags().StringVar(&createDisplay, "display", "number", "Display format: number, currency, percentage")
	createCmd.Flags().StringVar(&createCurrency, "currency", "USD", "Currency code (when --display currency)")
	createCmd.Flags().Float64Var(&createPrecision, "precision", 0, "Decimal precision")
	createCmd.Flags().StringVar(&createRenderStyle, "render-style", "BAR", "How a BAR chart is drawn: BAR, LINE, or AREA")
}

const createChartMutation = `
	mutation CreateChart($input: CreateChartInput!) {
		createChart(input: $input) {
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
`

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
	createGroupBy = strings.ToUpper(createGroupBy)

	if (createType == "BAR" || createType == "PIE") && createGroupBy == "" {
		return fmt.Errorf("--group-by is required for %s charts", createType)
	}

	client, err := newClient()
	if err != nil {
		return err
	}
	client.SetProject(createWorkspace)

	workspaceID, err := client.ResolveProjectID(createWorkspace)
	if err != nil {
		return fmt.Errorf("failed to resolve workspace: %w", err)
	}

	// The measured field and the grouping field are resolved separately: the
	// API matches them by *name* plus type (not ID), so an unresolved reference
	// silently degrades a SUM into a record count rather than erroring.
	measure, err := resolveMeasureField(client)
	if err != nil {
		return err
	}
	groupField, err := resolveGroupByField(client)
	if err != nil {
		return err
	}

	var input map[string]interface{}
	switch createType {
	case "STAT":
		input = buildStatInput(workspaceID, measure)
	case "BAR":
		input = buildBarInput(workspaceID, measure, groupField)
	case "PIE":
		input = buildPieInput(workspaceID, measure, groupField)
	}

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

	variables := map[string]interface{}{"input": input}
	if err := client.ExecuteQueryWithResult(createChartMutation, variables, &response); err != nil {
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

// resolveMeasureField resolves --field into the id/name/type triple the API
// needs. Auto charts key off name+type; stat cards key off the ID.
func resolveMeasureField(client *common.Client) (*common.CustomFieldRef, error) {
	if createField == "" {
		return nil, nil
	}
	field, err := common.ResolveCustomField(client, createField)
	if err != nil {
		return nil, fmt.Errorf("--field: %w", err)
	}
	return field, nil
}

// resolveGroupByField resolves --group-by-field, and enforces that the field
// is one the API can actually group on. Without this check an ungroupable type
// returns an empty chart with no error.
func resolveGroupByField(client *common.Client) (*common.CustomFieldRef, error) {
	if createGroupBy != "CUSTOM_FIELD" {
		if createGroupByCF != "" {
			return nil, fmt.Errorf("--group-by-field only applies with --group-by CUSTOM_FIELD")
		}
		return nil, nil
	}

	ref := createGroupByCF
	if ref == "" {
		return nil, fmt.Errorf("--group-by CUSTOM_FIELD requires --group-by-field")
	}

	field, err := common.ResolveCustomField(client, ref)
	if err != nil {
		return nil, fmt.Errorf("--group-by-field: %w", err)
	}
	if !common.IsGroupableCustomFieldType(field.Type) {
		return nil, fmt.Errorf(
			"custom field %q is a %s field, which the API cannot group by. Groupable types: %s",
			field.Name, field.Type, strings.Join(common.GroupableCustomFieldTypes, ", "),
		)
	}
	return field, nil
}

func buildStatInput(workspaceID string, measure *common.CustomFieldRef) map[string]interface{} {
	display := buildDisplay()

	// Segment uids are globally unique, not per-chart, so they must be minted
	// fresh every time. Fixed placeholders collide with the first chart ever
	// created with them and the mutation fails on the unique index.
	segmentUID := common.NewCuid()
	valueUID := common.NewCuid()

	// A single-source stat's formula is just a reference to its one value UID.
	reference, _ := json.Marshal(map[string]string{"chartSegmentValueUID": valueUID})

	segmentValue := map[string]interface{}{
		"uid":       valueUID,
		"title":     createTitle,
		"projectId": workspaceID,
		"function":  createFunction,
		// Always an object, never null: the chart worker reads this with
		// Object.entries() and a null filter throws, leaving the card
		// permanently blank.
		"filter": map[string]interface{}{},
	}
	if measure != nil {
		segmentValue["customFieldId"] = measure.ID
	}

	return map[string]interface{}{
		"dashboardId": createDashboard,
		"title":       createTitle,
		"type":        "STAT",
		"display":     display,
		"chartSegments": []interface{}{
			map[string]interface{}{
				"title": createTitle,
				"color": "#3B82F6",
				"uid":   segmentUID,
				"formula": map[string]interface{}{
					"logic": map[string]interface{}{
						"text": string(reference),
						"html": "<p>" + string(reference) + "</p>",
					},
					"display": display,
				},
				"chartSegmentValues": []interface{}{segmentValue},
			},
		},
	}
}

func buildBarInput(workspaceID string, measure, groupField *common.CustomFieldRef) map[string]interface{} {
	display := buildDisplay()

	xAxis := map[string]interface{}{
		"title": "Category",
		"type":  createGroupBy,
	}
	if isDateGroupBy(createGroupBy) {
		xAxis["interval"] = strings.ToUpper(createInterval)
	}
	if groupField != nil {
		xAxis["customFieldName"] = groupField.Name
		xAxis["customFieldType"] = groupField.Type
	}

	yAxis := map[string]interface{}{
		"title":  "Value",
		"filter": map[string]interface{}{"projectIds": []string{workspaceID}},
	}
	// Only send function + field when measuring a custom field. Omitting them
	// counts records, and lets the API skip the custom-field JOIN.
	if measure != nil {
		yAxis["function"] = createFunction
		yAxis["customFieldName"] = measure.Name
		yAxis["customFieldType"] = measure.Type
	}

	return map[string]interface{}{
		"dashboardId": createDashboard,
		"title":       createTitle,
		"type":        "BAR",
		"display":     display,
		"metadata": map[string]interface{}{
			"barChart": map[string]interface{}{
				"renderStyle": strings.ToUpper(createRenderStyle),
				"xAxis":       xAxis,
				"yAxis":       yAxis,
			},
		},
	}
}

func buildPieInput(workspaceID string, measure, groupField *common.CustomFieldRef) map[string]interface{} {
	display := buildDisplay()

	groupBy := map[string]interface{}{
		"title": "Segment",
		"type":  createGroupBy,
	}
	if groupField != nil {
		groupBy["customFieldName"] = groupField.Name
		groupBy["customFieldType"] = groupField.Type
	}

	value := map[string]interface{}{
		"title":  "Value",
		"filter": map[string]interface{}{"projectIds": []string{workspaceID}},
	}
	if measure != nil {
		value["function"] = createFunction
		value["customFieldName"] = measure.Name
		value["customFieldType"] = measure.Type
	}

	return map[string]interface{}{
		"dashboardId": createDashboard,
		"title":       createTitle,
		"type":        "PIE",
		"display":     display,
		// Pie charts have their own metadata shape. Sending barChart here
		// computes, but the app only treats metadata.pieChart as an
		// auto-generated pie and cannot open the chart for editing.
		"metadata": map[string]interface{}{
			"pieChart": map[string]interface{}{
				"groupBy": groupBy,
				"value":   value,
			},
		},
	}
}

func buildDisplay() map[string]interface{} {
	display := map[string]interface{}{
		"precision": createPrecision,
		"function":  createFunction,
	}

	switch strings.ToUpper(createDisplay) {
	case "CURRENCY":
		display["type"] = "CURRENCY"
		display["currency"] = map[string]interface{}{"code": createCurrency, "name": createCurrency}
	case "PERCENTAGE":
		display["type"] = "PERCENTAGE"
	default:
		display["type"] = "NUMBER"
	}

	return display
}

func isDateGroupBy(groupBy string) bool {
	switch groupBy {
	case "TODO_DUE_DATE", "TODO_CREATED_AT", "TODO_UPDATED_AT", "TODO_COMPLETED_AT":
		return true
	}
	return false
}
