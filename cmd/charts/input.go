package charts

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

// Flags shared by `charts create` and `charts preview`. Both commands take the
// same description of a chart; one saves it, the other only computes it.
var (
	inDashboard   string
	inType        string
	inTitle       string
	inWorkspace   string
	inField       string
	inGroupByCF   string
	inFunction    string
	inGroupBy     string
	inInterval    string
	inDisplay     string
	inCurrency    string
	inPrecision   float64
	inRenderStyle string
	inPosition    float64
	inSources     []string
	inFormula     string

	inTarget        float64
	inTargetSegment string
	inDirection     string
	inBands         string
	inStatStyle     string

	inShowCompleted bool
	inArchived      bool
	inUnassigned    bool
	inAssignees     string
	inTags          string
	inLists         string
	inDueStart      string
	inDueEnd        string
	inQuery         string
	inFilterJSON    string
)

const groupByHelp = "Group by dimension (BAR/PIE): ASSIGNEE, TAG, TODO_LIST, TODO_STATUS, PROJECT, CUSTOM_FIELD, TODO_DUE_DATE, TODO_CREATED_AT, TODO_UPDATED_AT, TODO_COMPLETED_AT"

// registerChartInputFlags attaches the shared chart-description flags. The
// backing variables are package-level, which is safe because cobra runs exactly
// one command per process.
func registerChartInputFlags(cmd *cobra.Command) {
	f := cmd.Flags()
	f.StringVar(&inDashboard, "dashboard", "", "Dashboard ID (required)")
	f.StringVar(&inType, "type", "", "Chart type: STAT, BAR, or PIE (required)")
	f.StringVarP(&inTitle, "title", "t", "", "Chart title (required)")
	f.StringVarP(&inWorkspace, "workspace", "w", "", "Workspace ID or slug for the data source")
	f.StringVar(&inField, "field", "", "Custom field ID or name to measure (omit to count records)")
	f.StringVar(&inGroupByCF, "group-by-field", "", "Custom field ID or name to group by (with --group-by CUSTOM_FIELD)")
	f.StringVar(&inFunction, "function", "COUNT", "Aggregation: COUNT, COUNTA, SUM, AVERAGE, AVERAGEA, MIN, MAX")
	f.StringVar(&inGroupBy, "group-by", "", groupByHelp)
	f.StringVar(&inInterval, "interval", "MONTH", "Time interval for date grouping: DAY, WEEK, MONTH, QUARTER, YEAR")
	f.StringVar(&inDisplay, "display", "number", "Display format: number, currency, percentage")
	f.StringVar(&inCurrency, "currency", "USD", "Currency code (when --display currency)")
	f.Float64Var(&inPrecision, "precision", 0, "Decimal precision")
	f.StringVar(&inRenderStyle, "render-style", "BAR", "How a BAR chart is drawn: BAR, LINE, or AREA")
	f.Float64Var(&inPosition, "position", 0, "Position on the dashboard (0 appends)")
	f.StringArrayVar(&inSources, "source", nil, "Extra STAT data source: 'workspace=<id>;field=<id|name>;function=SUM;title=<label>' (repeatable)")
	f.StringVar(&inFormula, "formula", "", "STAT formula over the sources, referencing them by title (e.g. 'Won / Total * 100')")

	f.Float64Var(&inTarget, "target", 0, "STAT goal target (a fixed number)")
	f.StringVar(&inTargetSegment, "target-segment", "", "STAT goal target computed by another segment, by uid")
	f.StringVar(&inDirection, "direction", "", "Whether exceeding the target is good: HIGHER_IS_BETTER or LOWER_IS_BETTER")
	f.StringVar(&inBands, "bands", "", "Goal thresholds as '<atRisk>,<onTrack>' ratios (e.g. '0.5,0.9')")
	f.StringVar(&inStatStyle, "stat-style", "", "How a STAT card is drawn: VALUE, PROGRESS, or GAUGE")

	f.BoolVar(&inShowCompleted, "show-completed", false, "Include completed records")
	f.BoolVar(&inArchived, "archived", false, "Count archived records instead of active ones")
	f.BoolVar(&inUnassigned, "unassigned", false, "Only records with no assignee")
	f.StringVar(&inAssignees, "assignees", "", "Comma-separated assignee user IDs")
	f.StringVar(&inTags, "tags", "", "Comma-separated tag IDs")
	f.StringVar(&inLists, "lists", "", "Comma-separated list IDs")
	f.StringVar(&inDueStart, "due-start", "", "Only records due on or after this date (YYYY-MM-DD)")
	f.StringVar(&inDueEnd, "due-end", "", "Only records due on or before this date (YYYY-MM-DD)")
	f.StringVar(&inQuery, "q", "", "Only records matching this text")
	f.StringVar(&inFilterJSON, "filter-json", "", "Raw TodoFilterInput JSON, merged last (for per-field conditions and nested groups)")
}

// statSource is one measured quantity on a stat card. A card with several of
// them combines their results with --formula.
//
// The narrowing keys matter for the common shape: a rate is two counts over the
// same records distinguished only by a tag or a list, so each source needs its
// own filter rather than sharing the card's.
type statSource struct {
	Workspace     string
	Field         string
	Function      string
	Title         string
	Lists         string
	Tags          string
	Assignees     string
	ShowCompleted *bool
}

// buildChartInput validates the flags and assembles a CreateChartInput.
// previewChart and createChart take the identical input, so both commands
// share this.
func buildChartInput(cmd *cobra.Command, client *common.Client) (map[string]interface{}, error) {
	if inDashboard == "" {
		return nil, fmt.Errorf("dashboard ID is required. Use --dashboard flag")
	}
	if inType == "" {
		return nil, fmt.Errorf("chart type is required. Use --type flag (STAT, BAR, or PIE)")
	}
	if inTitle == "" {
		return nil, fmt.Errorf("chart title is required. Use --title flag")
	}

	inType = strings.ToUpper(inType)
	if inType != "STAT" && inType != "BAR" && inType != "PIE" {
		return nil, fmt.Errorf("invalid chart type '%s'. Must be STAT, BAR, or PIE", inType)
	}
	inFunction = strings.ToUpper(inFunction)
	inGroupBy = strings.ToUpper(inGroupBy)

	if (inType == "BAR" || inType == "PIE") && inGroupBy == "" {
		return nil, fmt.Errorf("--group-by is required for %s charts", inType)
	}

	sources, err := parseStatSources(cmd)
	if err != nil {
		return nil, err
	}
	if inType != "STAT" && len(inSources) > 0 {
		return nil, fmt.Errorf("--source only applies to STAT charts")
	}
	if inType != "STAT" && inFormula != "" {
		return nil, fmt.Errorf("--formula only applies to STAT charts")
	}

	display := buildDisplay()

	switch inType {
	case "STAT":
		return buildStatInput(cmd, client, sources, display)
	case "BAR":
		return buildAutoInput(cmd, client, display, false)
	default:
		return buildAutoInput(cmd, client, display, true)
	}
}

// parseStatSources resolves the default source (from --workspace/--field) plus
// any --source entries into one ordered list.
func parseStatSources(cmd *cobra.Command) ([]statSource, error) {
	var sources []statSource

	if inWorkspace != "" {
		sources = append(sources, statSource{
			Workspace: inWorkspace,
			Field:     inField,
			Function:  inFunction,
			Title:     inTitle,
		})
	}

	for _, raw := range inSources {
		source := statSource{Function: inFunction}
		for _, pair := range strings.Split(raw, ";") {
			pair = strings.TrimSpace(pair)
			if pair == "" {
				continue
			}
			key, value, found := strings.Cut(pair, "=")
			if !found {
				return nil, fmt.Errorf("invalid --source entry %q: expected key=value pairs separated by ';'", raw)
			}
			value = strings.TrimSpace(value)
			switch strings.ToLower(strings.TrimSpace(key)) {
			case "workspace":
				source.Workspace = value
			case "field":
				source.Field = value
			case "function":
				source.Function = strings.ToUpper(value)
			case "title":
				source.Title = value
			case "lists":
				source.Lists = value
			case "tags":
				source.Tags = value
			case "assignees":
				source.Assignees = value
			case "show-completed", "showcompleted":
				parsed, err := strconv.ParseBool(value)
				if err != nil {
					return nil, fmt.Errorf("--source show-completed must be true or false, got %q", value)
				}
				source.ShowCompleted = &parsed
			default:
				return nil, fmt.Errorf(
					"unknown key %q in --source. Valid keys: workspace, field, function, title, lists, tags, assignees, show-completed",
					key,
				)
			}
		}
		if source.Workspace == "" {
			return nil, fmt.Errorf("--source entry %q is missing workspace=", raw)
		}
		if source.Title == "" {
			source.Title = fmt.Sprintf("Source %d", len(sources)+1)
		}
		sources = append(sources, source)
	}

	if len(sources) == 0 {
		return nil, fmt.Errorf("a data source is required. Use --workspace (or --source for several)")
	}
	return sources, nil
}

func buildStatInput(
	cmd *cobra.Command,
	client *common.Client,
	sources []statSource,
	display map[string]interface{},
) (map[string]interface{}, error) {
	segmentUID := common.NewCuid()

	// Each source becomes one segment value with its own uid; the formula
	// combines them. Uids are globally unique, so they are minted per call.
	values := make([]interface{}, 0, len(sources))
	titleToUID := map[string]string{}

	for _, source := range sources {
		client.SetProject(source.Workspace)
		workspaceID, err := client.ResolveProjectID(source.Workspace)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve workspace %q: %w", source.Workspace, err)
		}

		filter, err := buildSourceFilter(source, workspaceID)
		if err != nil {
			return nil, err
		}

		valueUID := common.NewCuid()
		titleToUID[source.Title] = valueUID

		value := map[string]interface{}{
			"uid":       valueUID,
			"title":     source.Title,
			"projectId": workspaceID,
			"function":  source.Function,
			"filter":    filter,
		}
		if source.Field != "" {
			field, err := common.ResolveCustomField(client, source.Field)
			if err != nil {
				return nil, fmt.Errorf("source %q: %w", source.Title, err)
			}
			value["customFieldId"] = field.ID
		}
		values = append(values, value)
	}

	logicText, err := buildStatFormula(sources, titleToUID)
	if err != nil {
		return nil, err
	}

	input := map[string]interface{}{
		"dashboardId": inDashboard,
		"title":       inTitle,
		"type":        "STAT",
		"display":     display,
		"chartSegments": []interface{}{
			map[string]interface{}{
				"title":              inTitle,
				"color":              "#3B82F6",
				"uid":                segmentUID,
				"formula":            map[string]interface{}{"logic": map[string]interface{}{"text": logicText, "html": "<p>" + logicText + "</p>"}, "display": display},
				"chartSegmentValues": values,
			},
		},
	}
	if inPosition != 0 {
		input["position"] = inPosition
	}

	statCard, err := buildStatCardMetadata(cmd, segmentUID)
	if err != nil {
		return nil, err
	}
	if statCard != nil {
		input["metadata"] = map[string]interface{}{"statCard": statCard}
	}

	return input, nil
}

// buildStatFormula produces the formula text. With one source and no --formula
// it is a bare reference to that source; with --formula, each source title in
// the expression is swapped for its uid reference.
func buildStatFormula(sources []statSource, titleToUID map[string]string) (string, error) {
	reference := func(uid string) string {
		encoded, _ := json.Marshal(map[string]string{"chartSegmentValueUID": uid})
		return string(encoded)
	}

	if inFormula == "" {
		if len(sources) > 1 {
			return "", fmt.Errorf("%d data sources need a --formula saying how to combine them", len(sources))
		}
		return reference(titleToUID[sources[0].Title]), nil
	}

	// Longest title first, so a title that is a prefix of another ("Won" vs
	// "Won Deals") can't shadow it.
	titles := make([]string, 0, len(titleToUID))
	for title := range titleToUID {
		titles = append(titles, title)
	}
	for i := range titles {
		for j := i + 1; j < len(titles); j++ {
			if len(titles[j]) > len(titles[i]) {
				titles[i], titles[j] = titles[j], titles[i]
			}
		}
	}

	formula := inFormula
	replaced := false
	for _, title := range titles {
		if strings.Contains(formula, title) {
			formula = strings.ReplaceAll(formula, title, reference(titleToUID[title]))
			replaced = true
		}
	}
	if !replaced {
		return "", fmt.Errorf("--formula does not mention any source by title. Sources: %s", strings.Join(titles, ", "))
	}
	return formula, nil
}

func buildAutoInput(
	cmd *cobra.Command,
	client *common.Client,
	display map[string]interface{},
	isPie bool,
) (map[string]interface{}, error) {
	if inWorkspace == "" {
		return nil, fmt.Errorf("workspace is required. Use --workspace flag")
	}

	client.SetProject(inWorkspace)
	workspaceID, err := client.ResolveProjectID(inWorkspace)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve workspace: %w", err)
	}

	measure, err := resolveMeasureField(client)
	if err != nil {
		return nil, err
	}
	groupField, err := resolveGroupByField(client)
	if err != nil {
		return nil, err
	}

	filter, err := buildFilter([]string{workspaceID})
	if err != nil {
		return nil, err
	}

	groupBy := map[string]interface{}{
		"title": "Category",
		"type":  inGroupBy,
	}
	if groupField != nil {
		groupBy["customFieldName"] = groupField.Name
		groupBy["customFieldType"] = groupField.Type
	}

	value := map[string]interface{}{
		"title":  "Value",
		"filter": filter,
	}
	// Only send function + field when measuring a custom field. Omitting them
	// counts records and lets the API skip the custom-field JOIN.
	if measure != nil {
		value["function"] = inFunction
		value["customFieldName"] = measure.Name
		value["customFieldType"] = measure.Type
	}

	input := map[string]interface{}{
		"dashboardId": inDashboard,
		"title":       inTitle,
		"display":     display,
	}
	if inPosition != 0 {
		input["position"] = inPosition
	}

	if isPie {
		input["type"] = "PIE"
		groupBy["title"] = "Segment"
		input["metadata"] = map[string]interface{}{
			"pieChart": map[string]interface{}{"groupBy": groupBy, "value": value},
		}
		return input, nil
	}

	if isDateGroupBy(inGroupBy) {
		groupBy["interval"] = strings.ToUpper(inInterval)
	}
	renderStyle := strings.ToUpper(inRenderStyle)
	switch renderStyle {
	case "BAR", "LINE", "AREA":
	default:
		return nil, fmt.Errorf("invalid --render-style %q. Must be BAR, LINE, or AREA", inRenderStyle)
	}

	input["type"] = "BAR"
	input["metadata"] = map[string]interface{}{
		"barChart": map[string]interface{}{
			"renderStyle": renderStyle,
			"xAxis":       groupBy,
			"yAxis":       value,
		},
	}
	return input, nil
}

// buildStatCardMetadata assembles the goal configuration, or returns nil when
// the card has no goal.
//
// A card with a target must send its target inside a present statCard object;
// "no target" is `target: null` rather than absent metadata, because the API's
// metadata union cannot resolve an empty object and an unresolvable member
// fails the whole charts query.
func buildStatCardMetadata(cmd *cobra.Command, valueSegmentUID string) (map[string]interface{}, error) {
	hasTarget := cmd.Flags().Changed("target")
	hasSegmentTarget := inTargetSegment != ""
	hasStyle := inStatStyle != ""
	hasBands := inBands != ""

	if !hasTarget && !hasSegmentTarget && !hasStyle && !hasBands {
		return nil, nil
	}
	if hasTarget && hasSegmentTarget {
		return nil, fmt.Errorf("--target and --target-segment are alternatives; pass only one")
	}

	statCard := map[string]interface{}{}

	switch {
	case hasTarget:
		if inTarget <= 0 {
			return nil, fmt.Errorf("--target must be greater than zero")
		}
		statCard["target"] = map[string]interface{}{"mode": "STATIC", "value": inTarget}
	case hasSegmentTarget:
		if inTargetSegment == valueSegmentUID {
			return nil, fmt.Errorf("--target-segment must name a different segment from the one showing the value")
		}
		statCard["target"] = map[string]interface{}{"mode": "SEGMENT", "segmentUid": inTargetSegment}
	default:
		statCard["target"] = nil
	}

	direction := strings.ToUpper(inDirection)
	if direction != "" && direction != "HIGHER_IS_BETTER" && direction != "LOWER_IS_BETTER" {
		return nil, fmt.Errorf("invalid --direction %q. Must be HIGHER_IS_BETTER or LOWER_IS_BETTER", inDirection)
	}

	if hasBands {
		// The API requires direction alongside bands: without it a
		// LOWER_IS_BETTER pair would be graded by HIGHER_IS_BETTER rules and
		// the card would read backwards.
		if direction == "" {
			direction = "HIGHER_IS_BETTER"
		}
		atRisk, onTrack, err := parseBands(inBands)
		if err != nil {
			return nil, err
		}
		if direction == "LOWER_IS_BETTER" && onTrack > atRisk {
			return nil, fmt.Errorf("when lower is better, the on-track threshold must be at or below the at-risk threshold")
		}
		if direction == "HIGHER_IS_BETTER" && atRisk > onTrack {
			return nil, fmt.Errorf("when higher is better, the at-risk threshold must be at or below the on-track threshold")
		}
		statCard["bands"] = map[string]interface{}{"atRisk": atRisk, "onTrack": onTrack}
	}

	if direction != "" {
		statCard["direction"] = direction
	}

	if hasStyle {
		style := strings.ToUpper(inStatStyle)
		switch style {
		case "VALUE", "PROGRESS", "GAUGE":
		default:
			return nil, fmt.Errorf("invalid --stat-style %q. Must be VALUE, PROGRESS, or GAUGE", inStatStyle)
		}
		statCard["renderStyle"] = style
	}

	return statCard, nil
}

func parseBands(value string) (float64, float64, error) {
	parts := strings.Split(value, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("--bands must be two numbers, '<atRisk>,<onTrack>' (e.g. '0.5,0.9')")
	}
	atRisk, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("--bands at-risk threshold is not a number: %w", err)
	}
	onTrack, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, 0, fmt.Errorf("--bands on-track threshold is not a number: %w", err)
	}
	if atRisk < 0 || onTrack < 0 {
		return 0, 0, fmt.Errorf("--bands thresholds cannot be negative")
	}
	return atRisk, onTrack, nil
}

// buildSourceFilter narrows one stat source: the card's own filter flags, with
// the source's narrowing keys layered on top so each source can measure a
// different slice of the same records.
func buildSourceFilter(source statSource, workspaceID string) (map[string]interface{}, error) {
	filter, err := buildFilter([]string{workspaceID})
	if err != nil {
		return nil, err
	}
	if ids := common.SplitCSV(source.Lists); len(ids) > 0 {
		filter["todoListIds"] = ids
	}
	if ids := common.SplitCSV(source.Tags); len(ids) > 0 {
		filter["tagIds"] = ids
	}
	if ids := common.SplitCSV(source.Assignees); len(ids) > 0 {
		filter["assigneeIds"] = ids
	}
	if source.ShowCompleted != nil {
		filter["showCompleted"] = *source.ShowCompleted
	}
	return filter, nil
}

func buildFilter(projectIDs []string) (map[string]interface{}, error) {
	opts := common.TodoFilterOptions{
		Assignees:  inAssignees,
		Tags:       inTags,
		Lists:      inLists,
		DueStart:   inDueStart,
		DueEnd:     inDueEnd,
		Query:      inQuery,
		FilterJSON: inFilterJSON,
	}
	if inShowCompleted {
		opts.ShowCompleted = &inShowCompleted
	}
	if inArchived {
		opts.Archived = &inArchived
	}
	if inUnassigned {
		opts.Unassigned = &inUnassigned
	}
	return common.BuildTodoFilter(opts, projectIDs)
}

// resolveMeasureField resolves --field into the id/name/type triple the API
// needs. Auto charts key off name+type; stat cards key off the ID.
func resolveMeasureField(client *common.Client) (*common.CustomFieldRef, error) {
	if inField == "" {
		return nil, nil
	}
	field, err := common.ResolveCustomField(client, inField)
	if err != nil {
		return nil, fmt.Errorf("--field: %w", err)
	}
	return field, nil
}

// resolveGroupByField resolves --group-by-field, and enforces that the field is
// one the API can actually group on. Without this check an ungroupable type
// returns an empty chart with no error.
func resolveGroupByField(client *common.Client) (*common.CustomFieldRef, error) {
	if inGroupBy != "CUSTOM_FIELD" {
		if inGroupByCF != "" {
			return nil, fmt.Errorf("--group-by-field only applies with --group-by CUSTOM_FIELD")
		}
		return nil, nil
	}
	if inGroupByCF == "" {
		return nil, fmt.Errorf("--group-by CUSTOM_FIELD requires --group-by-field")
	}

	field, err := common.ResolveCustomField(client, inGroupByCF)
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

func buildDisplay() map[string]interface{} {
	display := map[string]interface{}{
		"precision": inPrecision,
		"function":  inFunction,
	}

	switch strings.ToUpper(inDisplay) {
	case "CURRENCY":
		display["type"] = "CURRENCY"
		display["currency"] = map[string]interface{}{"code": inCurrency, "name": inCurrency}
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
