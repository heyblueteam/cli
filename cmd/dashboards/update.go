package dashboards

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

var updateDashboard, updateTitle, updateFormat string
var updateAllowViewerChartData string

var updateCmd = &cobra.Command{
	Use: "update", Short: "Update a dashboard",
	Example: `  blue dashboards update --dashboard <id> --title "New title"
  blue dashboards update --dashboard <id> --allow-viewer-chart-data true --format json`,
	RunE: runUpdate,
}

func init() {
	updateCmd.Flags().StringVar(&updateDashboard, "dashboard", "", "Dashboard ID (required)")
	updateCmd.Flags().StringVar(&updateTitle, "title", "", "New title")
	updateCmd.Flags().StringVar(&updateAllowViewerChartData, "allow-viewer-chart-data", "", "Allow VIEWER shares to inspect chart records (true or false)")
	updateCmd.Flags().StringVar(&updateFormat, "format", "", "Output format (json)")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if updateDashboard == "" {
		return fmt.Errorf("dashboard ID is required. Use --dashboard")
	}
	input := map[string]interface{}{"id": updateDashboard}
	if cmd.Flags().Changed("title") {
		input["title"] = updateTitle
	}
	if updateAllowViewerChartData != "" {
		var value bool
		if err := json.Unmarshal([]byte(strings.ToLower(updateAllowViewerChartData)), &value); err != nil {
			return fmt.Errorf("--allow-viewer-chart-data must be true or false")
		}
		input["allowViewerChartData"] = value
	}
	if len(input) == 1 {
		return fmt.Errorf("nothing to update")
	}
	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	client := common.NewClient(config)
	query := `mutation EditDashboard($input:EditDashboardInput!){editDashboard(input:$input){id title allowViewerChartData createdAt updatedAt createdBy{id fullName} dashboardUsers{id role user{id fullName}}}}`
	var response struct {
		Dashboard DashboardItem `json:"editDashboard"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"input": input}, &response); err != nil {
		return fmt.Errorf("failed to update dashboard: %w", err)
	}
	if strings.EqualFold(updateFormat, "json") {
		data, err := json.MarshalIndent(response.Dashboard, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(data))
		return nil
	}
	fmt.Printf("Dashboard updated\nID: %s\nTitle: %s\nViewer chart data: %t\n", response.Dashboard.ID, response.Dashboard.Title, response.Dashboard.AllowViewerChartData)
	return nil
}
