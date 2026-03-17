package dashboards

import (
	"fmt"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a dashboard",
	Long:  "Permanently delete a dashboard and all its charts.",
	Example: `  blue dashboards delete --dashboard <id> --confirm`,
	RunE: runDelete,
}

var (
	deleteDashboard string
	deleteConfirm   bool
)

func init() {
	deleteCmd.Flags().StringVar(&deleteDashboard, "dashboard", "", "Dashboard ID (required)")
	deleteCmd.Flags().BoolVarP(&deleteConfirm, "confirm", "y", false, "Confirm deletion (required)")
}

func runDelete(cmd *cobra.Command, args []string) error {
	if deleteDashboard == "" {
		return fmt.Errorf("dashboard ID is required. Use --dashboard flag")
	}
	if !deleteConfirm {
		return fmt.Errorf("deletion confirmation is required. Use --confirm flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := common.NewClient(config)

	mutation := fmt.Sprintf(`
		mutation DeleteDashboard {
			deleteDashboard(id: "%s") {
				success
			}
		}
	`, deleteDashboard)

	var response struct {
		DeleteDashboard struct {
			Success bool `json:"success"`
		} `json:"deleteDashboard"`
	}

	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		return fmt.Errorf("failed to delete dashboard: %w", err)
	}

	if response.DeleteDashboard.Success {
		fmt.Printf("Dashboard %s deleted\n", deleteDashboard)
	} else {
		return fmt.Errorf("failed to delete dashboard")
	}

	return nil
}
