package charts

import (
	"fmt"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a chart in a dashboard",
	Long: `Create a chart within a dashboard. Supports three chart types:

  STAT  - Single number/statistic (e.g., total revenue, record count)
  BAR   - Bar chart grouped by a dimension, drawn as bars, a line, or an area
  PIE   - Pie chart grouped by a dimension (e.g., records by status)

For BAR and PIE charts, use --group-by to set the grouping dimension.
The API automatically generates segments based on the data.

--field and --group-by-field accept a custom field ID or its exact name.
Grouping by CUSTOM_FIELD only works for field types the API can group on:
SELECT_SINGLE, SELECT_MULTI, CHECKBOX, COUNTRY, DATE, REFERENCE,
REFERENCED_BY, ASSIGNEE.

Record filters (--show-completed, --assignees, --tags, ...) narrow what the
chart measures. Anything they can't express goes in --filter-json, which is
merged last.

A stat card becomes a goal card when given --target (or --target-segment).
Add --bands and --stat-style PROGRESS or GAUGE to draw progress against it.`,
	Example: `  # Stat card: count of records
  blue charts create --dashboard <id> --type STAT --title "Total Records" \
    --workspace <id> --function COUNT

  # Stat card: open revenue only
  blue charts create --dashboard <id> --type STAT --title "Open Revenue" \
    --workspace <id> --field "Amount" --function SUM \
    --display currency --currency USD

  # Goal card: revenue against a quota, drawn as a progress bar
  blue charts create --dashboard <id> --type STAT --title "Q3 Revenue" \
    --workspace <id> --field "Amount" --function SUM \
    --target 500000 --direction HIGHER_IS_BETTER --bands 0.5,0.9 \
    --stat-style PROGRESS

  # Win rate across two sources
  blue charts create --dashboard <id> --type STAT --title "Win Rate" \
    --source "workspace=<id>;title=Won;tags=<won-tag>" \
    --source "workspace=<id>;title=Total" \
    --formula "Won / Total * 100" --display percentage

  # Records created per month, drawn as a line
  blue charts create --dashboard <id> --type BAR --title "Created" \
    --workspace <id> --group-by TODO_CREATED_AT --interval MONTH \
    --render-style LINE

  # Pie grouped by a select field
  blue charts create --dashboard <id> --type PIE --title "By Priority" \
    --workspace <id> --group-by CUSTOM_FIELD --group-by-field "Priority"`,
	RunE: runCreate,
}

var createNoPreview bool

func init() {
	registerChartInputFlags(createCmd)
	createCmd.Flags().BoolVar(&createNoPreview, "no-preview", false, "Skip the preview pass that seeds the chart's initial values")
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
	client, err := newClient()
	if err != nil {
		return err
	}

	input, err := buildChartInput(cmd, client)
	if err != nil {
		return err
	}

	// Auto charts (bar/pie) have their segments generated server-side, which
	// happens in the background — the chart renders empty until that lands.
	// Computing a preview first and saving its segments gives the chart its
	// numbers immediately, which is what the dashboard editor does too.
	if !createNoPreview && input["metadata"] != nil && input["chartSegments"] == nil {
		segments, err := previewSegments(client, input)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Warning: could not precompute values (%v); the chart will fill in shortly.\n", err)
		} else if len(segments) > 0 {
			input["chartSegments"] = segments
		}
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
