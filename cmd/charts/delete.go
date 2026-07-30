package charts

import (
	"fmt"

	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete a chart",
	Long:    "Permanently delete a chart from a dashboard.",
	Example: `  blue charts delete --chart <id> --confirm`,
	RunE:    runDelete,
}

var (
	deleteChart   string
	deleteConfirm bool
)

func init() {
	deleteCmd.Flags().StringVar(&deleteChart, "chart", "", "Chart ID (required)")
	deleteCmd.Flags().BoolVarP(&deleteConfirm, "confirm", "y", false, "Confirm deletion (required)")
}

func runDelete(cmd *cobra.Command, args []string) error {
	if deleteChart == "" {
		return fmt.Errorf("chart ID is required. Use --chart flag")
	}
	if !deleteConfirm {
		return fmt.Errorf("deletion confirmation is required. Use --confirm flag")
	}

	client, err := newClient()
	if err != nil {
		return err
	}

	const mutation = `
		mutation DeleteChart($id: String!) {
			deleteChart(id: $id) {
				success
			}
		}
	`

	var response struct {
		DeleteChart struct {
			Success bool `json:"success"`
		} `json:"deleteChart"`
	}

	variables := map[string]interface{}{"id": deleteChart}
	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to delete chart: %w", err)
	}

	if response.DeleteChart.Success {
		fmt.Printf("Chart %s deleted\n", deleteChart)
	} else {
		return fmt.Errorf("failed to delete chart")
	}

	return nil
}
