package charts

import (
	"fmt"

	"github.com/spf13/cobra"
)

var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Duplicate a chart",
	Long: `Duplicate a chart into a dashboard.

The copy keeps the original's configuration and recalculates its own values.
Pass the source dashboard's own ID to duplicate a chart in place.`,
	Example: `  blue charts copy --chart <id> --dashboard <destination-dashboard-id>
  blue charts copy --chart <id> --dashboard <id> --title "Revenue (copy)"`,
	RunE: runCopy,
}

var (
	copyChart     string
	copyDashboard string
	copyTitle     string
)

func init() {
	copyCmd.Flags().StringVar(&copyChart, "chart", "", "Chart ID to copy (required)")
	copyCmd.Flags().StringVar(&copyDashboard, "dashboard", "", "Destination dashboard ID (required)")
	copyCmd.Flags().StringVarP(&copyTitle, "title", "t", "", "Title for the copy")
}

const copyChartMutation = `
	mutation CopyChart($input: CopyChartInput!) {
		copyChart(input: $input) {
			id
			title
			type
			isCalculating
		}
	}
`

func runCopy(cmd *cobra.Command, args []string) error {
	if copyChart == "" {
		return fmt.Errorf("chart ID is required. Use --chart flag")
	}
	// The API needs an explicit destination, and a Chart doesn't expose the
	// dashboard it belongs to, so there is nothing to default to.
	if copyDashboard == "" {
		return fmt.Errorf("destination dashboard is required. Use --dashboard flag (pass the source dashboard to copy in place)")
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	input := map[string]interface{}{
		"chartId":     copyChart,
		"dashboardId": copyDashboard,
	}
	if copyTitle != "" {
		input["title"] = copyTitle
	}

	var response struct {
		CopyChart struct {
			ID            string `json:"id"`
			Title         string `json:"title"`
			Type          string `json:"type"`
			IsCalculating bool   `json:"isCalculating"`
		} `json:"copyChart"`
	}

	variables := map[string]interface{}{"input": input}
	if err := client.ExecuteQueryWithResult(copyChartMutation, variables, &response); err != nil {
		return fmt.Errorf("failed to copy chart: %w", err)
	}

	c := response.CopyChart
	fmt.Printf("Chart copied\n")
	fmt.Printf("ID: %s\n", c.ID)
	fmt.Printf("Title: %s\n", c.Title)
	fmt.Printf("Type: %s\n", c.Type)
	if c.IsCalculating {
		fmt.Printf("Status: Calculating...\n")
	}

	return nil
}
