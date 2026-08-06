package charts

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

type chartInputOptions struct {
	dashboard, chartType, title, workspace, workspaces string
	field, function, groupBy, groupField               string
	breakout, breakoutField, stackMode                 string
	interval, display, displayType, currency, filter   string
	precision                                          float64
	input, format                                      string
}

var createCmd = newChartInputCommand(false)
var previewCmd = newChartInputCommand(true)

func newChartInputCommand(preview bool) *cobra.Command {
	verb := "create"
	short := "Create a chart in a dashboard"
	if preview {
		verb, short = "preview", "Preview a chart without saving it"
	}
	o := &chartInputOptions{}
	cmd := &cobra.Command{
		Use:   verb,
		Short: short,
		Example: `  blue charts ` + verb + ` --dashboard <id> --title "By status" --display-type bar --workspace <id> --group-by TODO_STATUS
  blue charts ` + verb + ` --input chart.json --format json
  cat chart.json | blue charts ` + verb + ` --input - --format json`,
		RunE: func(cmd *cobra.Command, args []string) error { return runChartInput(cmd, o, preview) },
	}
	f := cmd.Flags()
	f.StringVar(&o.dashboard, "dashboard", "", "Dashboard ID")
	f.StringVar(&o.chartType, "type", "", "Legacy GraphQL type: STAT, BAR, or PIE")
	f.StringVarP(&o.title, "title", "t", "", "Chart title")
	f.StringVarP(&o.workspace, "workspace", "w", "", "Workspace ID or slug")
	f.StringVar(&o.workspaces, "workspaces", "", "Comma-separated workspace IDs or slugs")
	f.StringVar(&o.field, "field", "", "Custom field ID to measure")
	f.StringVar(&o.function, "function", "COUNT", "COUNT, COUNTA, SUM, AVERAGE, AVERAGEA, MIN, or MAX")
	f.StringVar(&o.groupBy, "group-by", "", "Dimension: PROJECT, ASSIGNEE, TAG, CUSTOM_FIELD, TODO, TODO_LIST, TODO_STATUS, or a TODO_* date")
	f.StringVar(&o.groupField, "group-field", "", "Custom field ID when --group-by is CUSTOM_FIELD")
	f.StringVar(&o.breakout, "breakout", "", "Breakdown dimension: PROJECT, ASSIGNEE, TAG, CUSTOM_FIELD, TODO_LIST, or TODO_STATUS")
	f.StringVar(&o.breakoutField, "breakout-field", "", "Custom field ID when --breakout is CUSTOM_FIELD")
	f.StringVar(&o.stackMode, "stack-mode", "", "STACKED or PERCENT")
	f.StringVar(&o.interval, "interval", "MONTH", "Date interval: DAY, WEEK, MONTH, QUARTER, or YEAR")
	f.StringVar(&o.display, "display", "number", "Number format: number, currency, or percentage")
	f.StringVar(&o.displayType, "display-type", "", "bar, line, area, row, leaderboard, table, pie, funnel, combo, stat, progress, or gauge")
	f.StringVar(&o.currency, "currency", "USD", "Currency code")
	f.Float64Var(&o.precision, "precision", 0, "Decimal precision")
	f.StringVar(&o.filter, "filter-json", "", "Chart-level TodoFilterInput JSON")
	f.StringVar(&o.input, "input", "", "Exact CreateChartInput JSON file, or - for stdin")
	f.StringVar(&o.format, "format", "", "Output format (json)")
	return cmd
}

func runChartInput(cmd *cobra.Command, o *chartInputOptions, preview bool) error {
	input, err := chartInput(cmd, o)
	if err != nil {
		return err
	}
	client, err := chartClient()
	if err != nil {
		return err
	}
	operation, field := "mutation CreateChart($input: CreateChartInput!)", "createChart"
	if preview {
		operation, field = "query PreviewChart($input: CreateChartInput!)", "previewChart"
	}
	query := fmt.Sprintf(`%s { %s(input: $input) { %s } }`, operation, field, chartFields)
	var response map[string]Chart
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"input": input}, &response); err != nil {
		return fmt.Errorf("failed to %s chart: %w", map[bool]string{true: "preview", false: "create"}[preview], err)
	}
	chart := response[field]
	if strings.EqualFold(o.format, "json") {
		return printJSON(chart)
	}
	if preview {
		fmt.Println("Chart preview ready")
	} else {
		fmt.Println("Chart created")
	}
	printChartSummary(chart)
	return nil
}

func chartInput(cmd *cobra.Command, o *chartInputOptions) (map[string]interface{}, error) {
	if o.input != "" {
		for _, name := range []string{"dashboard", "type", "title", "workspace", "workspaces", "field", "function", "group-by", "group-field", "breakout", "breakout-field", "stack-mode", "interval", "display", "display-type", "currency", "precision", "filter-json"} {
			if cmd.Flags().Changed(name) {
				return nil, fmt.Errorf("--input cannot be combined with --%s", name)
			}
		}
		return loadJSONInput(o.input)
	}
	return buildCommonChartInput(o)
}

func loadJSONInput(path string) (map[string]interface{}, error) {
	var data []byte
	var err error
	if path == "-" {
		data, err = io.ReadAll(os.Stdin)
	} else {
		data, err = os.ReadFile(path)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read chart input %q: %w", path, err)
	}
	var input map[string]interface{}
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("invalid chart input JSON: %w", err)
	}
	if input == nil {
		return nil, fmt.Errorf("chart input must be a JSON object")
	}
	return input, nil
}

func buildCommonChartInput(o *chartInputOptions) (map[string]interface{}, error) {
	if o.dashboard == "" || o.title == "" {
		return nil, fmt.Errorf("--dashboard and --title are required without --input")
	}
	displayType, chartType, err := resolveDisplayAndChartType(o.displayType, o.chartType)
	if err != nil {
		return nil, err
	}
	if displayType == "combo" {
		return nil, fmt.Errorf("combo charts require --input with at least two metrics")
	}
	if displayType == "progress" || displayType == "gauge" {
		return nil, fmt.Errorf("%s charts require --input with a target", displayType)
	}
	input := map[string]interface{}{
		"dashboardId": o.dashboard, "title": o.title, "type": chartType, "displayType": displayType,
		"display": displayInput(o.display, o.currency, o.precision),
	}
	if isGroupedDisplay(displayType) {
		if o.groupBy == "" {
			return nil, fmt.Errorf("--group-by is required for %s charts", displayType)
		}
		workspaceRefs := splitCSV(o.workspace + "," + o.workspaces)
		if len(workspaceRefs) == 0 {
			return nil, fmt.Errorf("--workspace or --workspaces is required")
		}
		client, err := chartClient()
		if err != nil {
			return nil, err
		}
		projectIDs, err := resolveProjectIDs(client, workspaceRefs)
		if err != nil {
			return nil, err
		}
		dimension, err := dimensionInput(client, strings.ToUpper(o.groupBy), o.groupField, o.interval)
		if err != nil {
			return nil, err
		}
		metric := map[string]interface{}{"key": "value", "title": o.title}
		fn := strings.ToUpper(o.function)
		if o.field == "" {
			if fn != "COUNT" {
				return nil, fmt.Errorf("--field is required for %s", fn)
			}
		} else {
			field, err := resolveCustomField(client, o.field)
			if err != nil {
				return nil, err
			}
			applyFieldMetadata(metric, field)
			metric["function"] = fn
		}
		filters := map[string]interface{}{"projectIds": projectIDs}
		if o.filter != "" {
			if err := json.Unmarshal([]byte(o.filter), &filters); err != nil {
				return nil, fmt.Errorf("invalid --filter-json: %w", err)
			}
			filters["projectIds"] = projectIDs
		}
		query := map[string]interface{}{"dimensions": []interface{}{dimension}, "metrics": []interface{}{metric}, "filters": filters}
		metadata := map[string]interface{}{"query": query}
		if o.breakout != "" {
			breakoutType := strings.ToUpper(o.breakout)
			allowedBreakouts := map[string]bool{"PROJECT": true, "ASSIGNEE": true, "TAG": true, "CUSTOM_FIELD": true, "TODO_LIST": true, "TODO_STATUS": true}
			if !allowedBreakouts[breakoutType] {
				return nil, fmt.Errorf("invalid breakout dimension %q", o.breakout)
			}
			breakout, err := dimensionInput(client, breakoutType, o.breakoutField, "")
			if err != nil {
				return nil, err
			}
			delete(breakout, "interval")
			query["breakout"] = breakout
			mode := strings.ToUpper(o.stackMode)
			if mode == "" {
				mode = "STACKED"
			}
			if mode != "STACKED" && mode != "PERCENT" {
				return nil, fmt.Errorf("invalid --stack-mode %q", o.stackMode)
			}
			metadata["presentation"] = map[string]interface{}{"stackMode": mode}
		} else if o.stackMode != "" {
			return nil, fmt.Errorf("--stack-mode requires --breakout")
		}
		input["metadata"] = metadata
		return input, nil
	}
	workspaceRefs := splitCSV(o.workspace + "," + o.workspaces)
	if len(workspaceRefs) != 1 {
		return nil, fmt.Errorf("manual %s charts require one --workspace", displayType)
	}
	client, err := chartClient()
	if err != nil {
		return nil, err
	}
	projectIDs, err := resolveProjectIDs(client, workspaceRefs)
	if err != nil {
		return nil, err
	}
	uid := "value-1"
	value := map[string]interface{}{"uid": uid, "title": o.title, "projectId": projectIDs[0]}
	fn := strings.ToUpper(o.function)
	if o.field != "" {
		value["customFieldId"] = o.field
		value["function"] = fn
	} else if fn != "COUNT" {
		return nil, fmt.Errorf("--field is required for %s", fn)
	} else {
		value["function"] = "COUNT"
	}
	if o.filter != "" {
		var filter map[string]interface{}
		if err := json.Unmarshal([]byte(o.filter), &filter); err != nil {
			return nil, fmt.Errorf("invalid --filter-json: %w", err)
		}
		value["filter"] = filter
	}
	formulaDisplay := displayInput(o.display, o.currency, o.precision)
	input["chartSegments"] = []interface{}{map[string]interface{}{
		"title": o.title, "color": "#3B82F6", "uid": "segment-1",
		"formula":            map[string]interface{}{"logic": map[string]interface{}{"text": fmt.Sprintf(`{"chartSegmentValueUID":"%s"}`, uid), "html": fmt.Sprintf(`<p>{"chartSegmentValueUID":"%s"}</p>`, uid)}, "display": formulaDisplay},
		"chartSegmentValues": []interface{}{value},
	}}
	return input, nil
}

func displayInput(kind, currency string, precision float64) map[string]interface{} {
	t := strings.ToUpper(kind)
	if t == "" {
		t = "NUMBER"
	}
	result := map[string]interface{}{"type": t, "precision": precision}
	if t == "CURRENCY" {
		result["currency"] = map[string]interface{}{"code": currency, "name": currency}
	}
	return result
}

func resolveDisplayAndChartType(displayType, chartType string) (string, string, error) {
	d := strings.ToLower(displayType)
	legacy := strings.ToUpper(chartType)
	if legacy != "" && legacy != "STAT" && legacy != "PIE" && legacy != "BAR" {
		return "", "", fmt.Errorf("invalid --type %q", chartType)
	}
	if d == "" {
		d = map[string]string{"STAT": "stat", "PIE": "pie", "BAR": "bar"}[legacy]
	}
	if d == "" {
		return "", "", fmt.Errorf("--display-type or --type is required")
	}
	known := map[string]bool{"bar": true, "line": true, "area": true, "row": true, "leaderboard": true, "table": true, "pie": true, "funnel": true, "combo": true, "stat": true, "progress": true, "gauge": true}
	if !known[d] {
		return "", "", fmt.Errorf("invalid --display-type %q", displayType)
	}
	grouped := isGroupedDisplay(d)
	derived := "STAT"
	if grouped {
		derived = "BAR"
	}
	if d == "pie" {
		derived = "PIE"
	}
	if legacy != "" && legacy != derived {
		return "", "", fmt.Errorf("--type %s conflicts with --display-type %s", legacy, d)
	}
	return d, derived, nil
}

func isGroupedDisplay(value string) bool {
	switch value {
	case "bar", "line", "area", "row", "leaderboard", "table", "pie", "funnel", "combo":
		return true
	case "stat", "progress", "gauge":
		return false
	}
	return false
}

type chartFieldMetadata struct{ Name, Type, ReferenceProjectID string }

func dimensionInput(client *common.Client, kind, fieldID, interval string) (map[string]interface{}, error) {
	allowed := map[string]bool{"PROJECT": true, "ASSIGNEE": true, "TAG": true, "CUSTOM_FIELD": true, "TODO": true, "TODO_LIST": true, "TODO_STATUS": true, "TODO_DUE_DATE": true, "TODO_CREATED_AT": true, "TODO_UPDATED_AT": true, "TODO_COMPLETED_AT": true}
	if !allowed[kind] {
		return nil, fmt.Errorf("invalid chart dimension %q", kind)
	}
	result := map[string]interface{}{"type": kind}
	if strings.HasPrefix(kind, "TODO_") && kind != "TODO_LIST" && kind != "TODO_STATUS" {
		result["interval"] = strings.ToUpper(interval)
	}
	if kind == "CUSTOM_FIELD" {
		if fieldID == "" {
			return nil, fmt.Errorf("a custom field ID is required for CUSTOM_FIELD")
		}
		field, err := resolveCustomField(client, fieldID)
		if err != nil {
			return nil, err
		}
		applyFieldMetadata(result, field)
	}
	return result, nil
}

func resolveCustomField(client *common.Client, id string) (chartFieldMetadata, error) {
	query := `query ChartCustomField($id: String!) { customField(id: $id) { name type referenceProject { id } } }`
	var response struct {
		CustomField struct {
			Name, Type       string
			ReferenceProject *struct {
				ID string `json:"id"`
			} `json:"referenceProject"`
		} `json:"customField"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"id": id}, &response); err != nil {
		return chartFieldMetadata{}, fmt.Errorf("failed to resolve custom field %q: %w", id, err)
	}
	field := chartFieldMetadata{Name: response.CustomField.Name, Type: response.CustomField.Type}
	if response.CustomField.ReferenceProject != nil {
		field.ReferenceProjectID = response.CustomField.ReferenceProject.ID
	}
	if field.Name == "" {
		return field, fmt.Errorf("custom field %q was not found", id)
	}
	return field, nil
}

func applyFieldMetadata(target map[string]interface{}, field chartFieldMetadata) {
	target["customFieldName"], target["customFieldType"] = field.Name, field.Type
	if field.ReferenceProjectID != "" {
		target["customFieldReferenceProjectId"] = field.ReferenceProjectID
	}
}
