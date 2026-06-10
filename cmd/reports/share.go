package reports

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "Replace report sharing users",
	Example: `  blue reports share --report <id> --users "user1:EDITOR,user2:VIEWER"
  blue reports share --report <id> --users ""`,
	RunE: runShare,
}

var (
	shareReport string
	shareUsers  string
)

func init() {
	shareCmd.Flags().StringVar(&shareReport, "report", "", "Report ID (required)")
	shareCmd.Flags().StringVar(&shareUsers, "users", "", "Comma-separated userId:ROLE entries; empty clears sharing")
}

func runShare(cmd *cobra.Command, args []string) error {
	if shareReport == "" {
		return fmt.Errorf("report ID is required. Use --report flag")
	}
	users := []interface{}{}
	for _, pair := range splitCSV(shareUsers) {
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid user format %q; expected userId:EDITOR or userId:VIEWER", pair)
		}
		role := strings.ToUpper(strings.TrimSpace(parts[1]))
		if role != "EDITOR" && role != "VIEWER" {
			return fmt.Errorf("invalid role %q; expected EDITOR or VIEWER", role)
		}
		users = append(users, map[string]interface{}{"userId": strings.TrimSpace(parts[0]), "role": role})
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`mutation ShareReport($id: String!, $input: UpdateReportInput!) { updateReport(id: $id, input: $input) { %s } }`, reportFields)
	var response struct {
		UpdateReport Report `json:"updateReport"`
	}
	variables := map[string]interface{}{"id": shareReport, "input": map[string]interface{}{"reportUsers": users}}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to share report: %w", err)
	}
	fmt.Printf("Report sharing updated (%d users)\n", len(response.UpdateReport.ReportUsers))
	return nil
}
