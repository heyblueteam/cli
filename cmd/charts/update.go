package charts

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a chart's title, position, display, or configuration",
	Long: `Change a saved chart without recreating it.

Covers the chart's own settings: its title, where it sits on the dashboard,
how its number is formatted, how it is drawn, and its goal configuration.

Changing what a chart measures (its grouping, field, or filter) means
replacing its configuration wholesale — pass --metadata-json for that, or
create a new chart.`,
	Example: `  blue charts update --chart <id> --title "Revenue this quarter"
  blue charts update --chart <id> --position 3000
  blue charts update --chart <id> --display currency --currency EUR --precision 2
  blue charts update --chart <id> --render-style LINE
  blue charts update --chart <id> --target 500000 --direction HIGHER_IS_BETTER \
    --bands 0.5,0.9 --stat-style PROGRESS
  blue charts update --chart <id> --clear-target`,
	RunE: runUpdate,
}

var (
	updateChart       string
	updateTitle       string
	updatePosition    float64
	updateDisplay     string
	updateCurrency    string
	updatePrecision   float64
	updateFunction    string
	updateRenderStyle string
	updateMetadata    string
	updateClearTarget bool

	updateTarget        float64
	updateTargetSegment string
	updateDirection     string
	updateBands         string
	updateStatStyle     string
)

func init() {
	f := updateCmd.Flags()
	f.StringVar(&updateChart, "chart", "", "Chart ID (required)")
	f.StringVarP(&updateTitle, "title", "t", "", "New chart title")
	f.Float64Var(&updatePosition, "position", 0, "New position on the dashboard")
	f.StringVar(&updateDisplay, "display", "", "Display format: number, currency, percentage")
	f.StringVar(&updateCurrency, "currency", "USD", "Currency code (when --display currency)")
	f.Float64Var(&updatePrecision, "precision", 0, "Decimal precision")
	f.StringVar(&updateFunction, "function", "", "Aggregation shown in the display: COUNT, SUM, AVERAGE, MIN, MAX, COUNTA, AVERAGEA")
	f.StringVar(&updateRenderStyle, "render-style", "", "How a BAR chart is drawn: BAR, LINE, or AREA")
	f.StringVar(&updateMetadata, "metadata-json", "", "Raw ChartMetadataInput JSON, replacing the chart's configuration")
	f.BoolVar(&updateClearTarget, "clear-target", false, "Remove a stat card's goal, leaving a plain stat card")

	f.Float64Var(&updateTarget, "target", 0, "Stat card goal target (a fixed number)")
	f.StringVar(&updateTargetSegment, "target-segment", "", "Stat card goal target computed by another segment, by uid")
	f.StringVar(&updateDirection, "direction", "", "Whether exceeding the target is good: HIGHER_IS_BETTER or LOWER_IS_BETTER")
	f.StringVar(&updateBands, "bands", "", "Goal thresholds as '<atRisk>,<onTrack>' ratios (e.g. '0.5,0.9')")
	f.StringVar(&updateStatStyle, "stat-style", "", "How a stat card is drawn: VALUE, PROGRESS, or GAUGE")
}

const editChartMutation = `
	mutation EditChart($input: EditChartInput!) {
		editChart(input: $input) {
			id
			title
			type
			position
			isCalculating
		}
	}
`

// todoFilterFields mirrors every field of TodoFilter that TodoFilterInput also
// accepts. Changing a chart's render style means reading its configuration and
// writing it back, so anything omitted here would be silently dropped — a
// chart filtered to "open deals this quarter" would quietly start counting
// everything. Keep in step with TodoFilter in schema.graphql.
const todoFilterFields = `
	assigneeIds unassigned dueEnd dueStart showCompleted projectIds q
	tagIds tagColors tagTitles todoListIds todoListTitles
	fields op groups groupLinks
	notAssigneeIds notTagIds notTodoListIds notProjectIds colors notColors
	hasTag hasColor hasDueDate hasDescription hasChecklist hasDependency hasReference
	createdStart createdEnd completedStart completedEnd updatedAt_gt updatedAt_gte
	recordName recordNameOp
	lastUpdatedByUserIds lastUpdatedByAutomationIds lastUpdatedByActorTypes
`

var chartMetadataQuery = `
	query Chart($id: String!) {
		chart(id: $id) {
			id
			type
			metadata {
				__typename
				... on ChartMetadataBarChart {
					barChart {
						renderStyle
						xAxis { title type interval customFieldName customFieldType }
						yAxis {
							title function customFieldName customFieldType
							filter { ` + todoFilterFields + ` }
						}
					}
				}
			}
		}
	}
`

func runUpdate(cmd *cobra.Command, args []string) error {
	if updateChart == "" {
		return fmt.Errorf("chart ID is required. Use --chart flag")
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	input := map[string]interface{}{"id": updateChart}

	if updateTitle != "" {
		input["title"] = updateTitle
	}
	if cmd.Flags().Changed("position") {
		input["position"] = updatePosition
	}
	if updateDisplay != "" {
		display := map[string]interface{}{"precision": updatePrecision}
		switch strings.ToUpper(updateDisplay) {
		case "CURRENCY":
			display["type"] = "CURRENCY"
			display["currency"] = map[string]interface{}{"code": updateCurrency, "name": updateCurrency}
		case "PERCENTAGE":
			display["type"] = "PERCENTAGE"
		case "NUMBER":
			display["type"] = "NUMBER"
		default:
			return fmt.Errorf("invalid --display %q. Must be number, currency, or percentage", updateDisplay)
		}
		if updateFunction != "" {
			display["function"] = strings.ToUpper(updateFunction)
		}
		input["display"] = display
	}

	metadata, err := buildUpdateMetadata(cmd, client)
	if err != nil {
		return err
	}
	if metadata != nil {
		input["metadata"] = metadata
	}

	if len(input) == 1 {
		return fmt.Errorf("nothing to update. Pass --title, --position, --display, --render-style, a goal flag, or --metadata-json")
	}

	var response struct {
		EditChart struct {
			ID            string  `json:"id"`
			Title         string  `json:"title"`
			Type          string  `json:"type"`
			Position      float64 `json:"position"`
			IsCalculating bool    `json:"isCalculating"`
		} `json:"editChart"`
	}

	variables := map[string]interface{}{"input": input}
	if err := client.ExecuteQueryWithResult(editChartMutation, variables, &response); err != nil {
		return fmt.Errorf("failed to update chart: %w", err)
	}

	c := response.EditChart
	fmt.Printf("Chart updated\n")
	fmt.Printf("ID: %s\n", c.ID)
	fmt.Printf("Title: %s\n", c.Title)
	if c.IsCalculating {
		fmt.Printf("Status: Calculating...\n")
	}

	return nil
}

// buildUpdateMetadata assembles the metadata patch, or returns nil when none of
// the flags that touch metadata were given.
//
// Metadata is replaced wholesale rather than merged, so changing only the
// render style means reading the chart's current configuration back first —
// otherwise the axes and filter would be dropped.
func buildUpdateMetadata(cmd *cobra.Command, client interface {
	ExecuteQueryWithResult(string, map[string]interface{}, interface{}) error
}) (map[string]interface{}, error) {
	if updateMetadata != "" {
		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(updateMetadata), &raw); err != nil {
			return nil, fmt.Errorf("invalid JSON for --metadata-json: %w", err)
		}
		return raw, nil
	}

	hasGoal := cmd.Flags().Changed("target") || updateTargetSegment != "" ||
		updateBands != "" || updateStatStyle != "" || updateDirection != "" || updateClearTarget

	if updateRenderStyle != "" {
		style := strings.ToUpper(updateRenderStyle)
		switch style {
		case "BAR", "LINE", "AREA":
		default:
			return nil, fmt.Errorf("invalid --render-style %q. Must be BAR, LINE, or AREA", updateRenderStyle)
		}
		if hasGoal {
			return nil, fmt.Errorf("--render-style applies to bar charts and the goal flags to stat cards; set them separately")
		}
		return buildRenderStyleMetadata(client, style)
	}

	if !hasGoal {
		return nil, nil
	}

	return buildGoalMetadata()
}

func buildRenderStyleMetadata(client interface {
	ExecuteQueryWithResult(string, map[string]interface{}, interface{}) error
}, style string) (map[string]interface{}, error) {
	var current struct {
		Chart struct {
			Type     string `json:"type"`
			Metadata *struct {
				TypeName string                 `json:"__typename"`
				BarChart map[string]interface{} `json:"barChart"`
			} `json:"metadata"`
		} `json:"chart"`
	}

	if err := client.ExecuteQueryWithResult(chartMetadataQuery, map[string]interface{}{"id": updateChart}, &current); err != nil {
		return nil, fmt.Errorf("failed to read the chart's current configuration: %w", err)
	}
	if current.Chart.Metadata == nil || current.Chart.Metadata.BarChart == nil {
		return nil, fmt.Errorf("--render-style applies to automatic bar charts; chart %s is a %s chart", updateChart, current.Chart.Type)
	}

	barChart := current.Chart.Metadata.BarChart
	barChart["renderStyle"] = style
	stripTypenames(barChart)

	return map[string]interface{}{"barChart": barChart}, nil
}

func buildGoalMetadata() (map[string]interface{}, error) {
	statCard := map[string]interface{}{}

	switch {
	case updateClearTarget:
		// "No goal" is a null target inside a present statCard, not absent
		// metadata: the API's metadata union cannot resolve an empty object.
		statCard["target"] = nil
		statCard["renderStyle"] = "VALUE"
	case updateTargetSegment != "":
		statCard["target"] = map[string]interface{}{"mode": "SEGMENT", "segmentUid": updateTargetSegment}
	case updateTarget != 0:
		if updateTarget <= 0 {
			return nil, fmt.Errorf("--target must be greater than zero")
		}
		statCard["target"] = map[string]interface{}{"mode": "STATIC", "value": updateTarget}
	}

	direction := strings.ToUpper(updateDirection)
	if direction != "" && direction != "HIGHER_IS_BETTER" && direction != "LOWER_IS_BETTER" {
		return nil, fmt.Errorf("invalid --direction %q. Must be HIGHER_IS_BETTER or LOWER_IS_BETTER", updateDirection)
	}

	if updateBands != "" {
		if direction == "" {
			direction = "HIGHER_IS_BETTER"
		}
		atRisk, onTrack, err := parseBands(updateBands)
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

	if updateStatStyle != "" {
		style := strings.ToUpper(updateStatStyle)
		switch style {
		case "VALUE", "PROGRESS", "GAUGE":
		default:
			return nil, fmt.Errorf("invalid --stat-style %q. Must be VALUE, PROGRESS, or GAUGE", updateStatStyle)
		}
		statCard["renderStyle"] = style
	}

	return map[string]interface{}{"statCard": statCard}, nil
}

// stripTypenames prepares metadata read back from the API for writing to the
// matching input type: it drops the __typename keys the input rejects, and the
// nulls that a wide selection set produces for every unset filter field. A
// false or a zero is a real value and is kept.
func stripTypenames(value interface{}) {
	switch typed := value.(type) {
	case map[string]interface{}:
		delete(typed, "__typename")
		for key, nested := range typed {
			if nested == nil {
				delete(typed, key)
				continue
			}
			stripTypenames(nested)
		}
	case []interface{}:
		for _, nested := range typed {
			stripTypenames(nested)
		}
	}
}
