package dashboards

import (
	"fmt"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Duplicate a dashboard",
	Long: `Duplicate a dashboard along with all of its charts.

Each copied chart recalculates its own values. A goal card whose target is
computed by another segment is rewired to the copy's own segments.`,
	Example: `  blue dashboards copy --dashboard <id>
  blue dashboards copy --dashboard <id> --title "Sales Dashboard (Q4)"`,
	RunE: runCopy,
}

var (
	copyDashboard string
	copyTitle     string
)

func init() {
	copyCmd.Flags().StringVar(&copyDashboard, "dashboard", "", "Dashboard ID to copy (required)")
	copyCmd.Flags().StringVarP(&copyTitle, "title", "t", "", "Title for the copy (defaults to the API's naming)")
}

func runCopy(cmd *cobra.Command, args []string) error {
	if copyDashboard == "" {
		return fmt.Errorf("dashboard ID is required. Use --dashboard flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	client := common.NewClient(config)

	input := map[string]interface{}{"dashboardId": copyDashboard}
	if copyTitle != "" {
		input["title"] = copyTitle
	}

	const mutation = `
		mutation CopyDashboard($input: CopyDashboardInput!) {
			copyDashboard(input: $input) {
				id
				title
				createdAt
			}
		}
	`

	var response struct {
		CopyDashboard struct {
			ID        string `json:"id"`
			Title     string `json:"title"`
			CreatedAt string `json:"createdAt"`
		} `json:"copyDashboard"`
	}

	variables := map[string]interface{}{"input": input}
	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to copy dashboard: %w", err)
	}

	d := response.CopyDashboard
	fmt.Printf("Dashboard copied\n")
	fmt.Printf("ID: %s\n", d.ID)
	fmt.Printf("Title: %s\n", d.Title)

	return nil
}
