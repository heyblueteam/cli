package charts

import (
	"fmt"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

var previewCmd = &cobra.Command{
	Use:   "preview",
	Short: "Compute a chart without saving it",
	Long: `Compute what a chart would show, without creating it.

Takes the same flags as 'charts create' — use it to check a grouping or a
filter before committing it to a dashboard. --dashboard is still required,
because the preview is scoped to that dashboard's organization.`,
	Example: `  blue charts preview --dashboard <id> --type PIE --title "By Status" \
    --workspace <id> --group-by TODO_STATUS

  blue charts preview --dashboard <id> --type BAR --title "Open by List" \
    --workspace <id> --group-by TODO_LIST --field "Amount" --function SUM \
    --format json`,
	RunE: runPreview,
}

var previewFormat string

func init() {
	registerChartInputFlags(previewCmd)
	previewCmd.Flags().StringVar(&previewFormat, "format", "", "Output format (json)")
}

const previewChartQuery = `
	query PreviewChart($input: CreateChartInput!) {
		previewChart(input: $input) {
			title
			type
			chartSegments {
				uid
				title
				color
				formulaResult
				formula {
					logic { text html }
					display { type precision function currency { code name } }
				}
			}
		}
	}
`

type previewSegment struct {
	UID           string   `json:"uid"`
	Title         string   `json:"title"`
	Color         *string  `json:"color"`
	FormulaResult *float64 `json:"formulaResult"`
	Formula       *struct {
		Logic *struct {
			Text string `json:"text"`
			HTML string `json:"html"`
		} `json:"logic"`
		Display map[string]interface{} `json:"display"`
	} `json:"formula"`
}

type previewResponse struct {
	PreviewChart struct {
		Title         string           `json:"title"`
		Type          string           `json:"type"`
		ChartSegments []previewSegment `json:"chartSegments"`
	} `json:"previewChart"`
}

func runPreview(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	input, err := buildChartInput(cmd, client)
	if err != nil {
		return err
	}

	var response previewResponse
	variables := map[string]interface{}{"input": input}
	if err := client.ExecuteQueryWithResult(previewChartQuery, variables, &response); err != nil {
		return fmt.Errorf("failed to preview chart: %w", err)
	}

	p := response.PreviewChart

	if previewFormat == "json" {
		return printJSON(p)
	}

	fmt.Printf("%s [%s]\n", p.Title, p.Type)
	if len(p.ChartSegments) == 0 {
		fmt.Printf("  (no data)\n")
		return nil
	}
	for _, segment := range p.ChartSegments {
		// A preview is computed synchronously, so there is no pending state:
		// the API stores a zero result as null, and the dashboard renders those
		// as 0. Print the same thing rather than inventing a third state.
		value := 0.0
		if segment.FormulaResult != nil {
			value = *segment.FormulaResult
		}
		fmt.Printf("  %-32s %s\n", common.TruncateString(segment.Title, 32), common.FormatNumber(value))
	}

	return nil
}

// previewSegments computes an auto chart's segments so they can be saved with
// the chart, giving it values before the background recalculation lands.
//
// The saved segments carry the computed number in the formula and the colour
// the API picked, matching what the dashboard editor saves.
func previewSegments(client *common.Client, input map[string]interface{}) ([]interface{}, error) {
	var response previewResponse
	variables := map[string]interface{}{"input": input}
	if err := client.ExecuteQueryWithResult(previewChartQuery, variables, &response); err != nil {
		return nil, err
	}

	display := input["display"]
	segments := make([]interface{}, 0, len(response.PreviewChart.ChartSegments))
	for index, segment := range response.PreviewChart.ChartSegments {
		result := 0.0
		if segment.FormulaResult != nil {
			result = *segment.FormulaResult
		}
		text := common.FormatNumber(result)

		title := segment.Title
		if title == "" {
			title = fmt.Sprintf("Segment %d", index+1)
		}

		saved := map[string]interface{}{
			"title": title,
			// The values are already computed, so the saved segment carries the
			// number itself rather than a reference to re-evaluate.
			"chartSegmentValues": []interface{}{},
			"formula": map[string]interface{}{
				"logic":   map[string]interface{}{"text": text, "html": text},
				"display": display,
			},
		}
		if segment.Color != nil && *segment.Color != "" {
			saved["color"] = *segment.Color
		}
		segments = append(segments, saved)
	}

	return segments, nil
}
