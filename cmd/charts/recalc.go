package charts

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var recalcCmd = &cobra.Command{
	Use:     "recalculate",
	Short:   "Recalculate chart data",
	Long:    "Force recalculation of one or more charts.",
	Example: `  blue charts recalculate --charts "chart1,chart2"`,
	RunE:    runRecalc,
}

var (
	recalcCharts string
)

func init() {
	recalcCmd.Flags().StringVar(&recalcCharts, "charts", "", "Comma-separated chart IDs (required)")
}

func runRecalc(cmd *cobra.Command, args []string) error {
	if recalcCharts == "" {
		return fmt.Errorf("chart IDs are required. Use --charts flag")
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	var ids []string
	for _, id := range strings.Split(recalcCharts, ",") {
		if id = strings.TrimSpace(id); id != "" {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return fmt.Errorf("chart IDs are required. Use --charts flag")
	}

	const mutation = `
		mutation RecalculateCharts($input: RecalculateChartsInput!) {
			recalculateCharts(input: $input)
		}
	`

	var response struct {
		RecalculateCharts bool `json:"recalculateCharts"`
	}

	variables := map[string]interface{}{"input": map[string]interface{}{"chartIds": ids}}
	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to recalculate charts: %w", err)
	}

	fmt.Printf("Recalculation triggered for %d chart(s)\n", len(ids))

	return nil
}
