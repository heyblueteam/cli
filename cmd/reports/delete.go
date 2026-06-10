package reports

import (
	"fmt"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete a report",
	Example: `  blue reports delete --report <id> --confirm`,
	RunE:    runDelete,
}

var (
	deleteReport  string
	deleteConfirm bool
)

func init() {
	deleteCmd.Flags().StringVar(&deleteReport, "report", "", "Report ID (required)")
	deleteCmd.Flags().BoolVarP(&deleteConfirm, "confirm", "y", false, "Confirm deletion (required)")
}

func runDelete(cmd *cobra.Command, args []string) error {
	if deleteReport == "" {
		return fmt.Errorf("report ID is required. Use --report flag")
	}
	if !deleteConfirm {
		return fmt.Errorf("deletion confirmation is required. Use --confirm flag")
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	query := `mutation DeleteReport($id: String!) { deleteReport(id: $id) }`
	var response struct {
		DeleteReport bool `json:"deleteReport"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"id": deleteReport}, &response); err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}
	if !response.DeleteReport {
		return fmt.Errorf("report was not deleted")
	}
	common.PrintSuccess(fmt.Sprintf("Deleted report %s", deleteReport))
	return nil
}
