package charts

import (
	"fmt"
	"strings"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

var recalcCmd = &cobra.Command{
	Use:   "recalculate",
	Short: "Recalculate chart data",
	Long:  "Force recalculation of one or more charts.",
	Example: `  blue charts recalculate --charts "chart1,chart2"`,
	RunE: runRecalc,
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

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := common.NewClient(config)

	ids := strings.Split(recalcCharts, ",")
	var idStrings []string
	for _, id := range ids {
		idStrings = append(idStrings, fmt.Sprintf(`"%s"`, strings.TrimSpace(id)))
	}

	mutation := fmt.Sprintf(`
		mutation RecalculateCharts {
			recalculateCharts(input: {
				chartIds: [%s]
			})
		}
	`, strings.Join(idStrings, ", "))

	var response struct {
		RecalculateCharts bool `json:"recalculateCharts"`
	}

	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		return fmt.Errorf("failed to recalculate charts: %w", err)
	}

	fmt.Printf("Recalculation triggered for %d chart(s)\n", len(ids))

	return nil
}
