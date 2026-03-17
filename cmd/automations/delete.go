package automations

import (
	"encoding/json"
	"fmt"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

type DeleteAutomationResponse struct {
	DeleteAutomation bool `json:"deleteAutomation"`
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete an automation",
	Long:  "Permanently delete an automation. This action cannot be undone.",
	Example: `  blue automations delete -w <id> --automation <aid> --confirm
  blue automations delete -w <id> --automation <aid> -y --simple`,
	RunE: runDelete,
}

var (
	deleteAutomationID string
	deleteWorkspace    string
	deleteConfirm      bool
	deleteSimple       bool
)

func init() {
	deleteCmd.Flags().StringVar(&deleteAutomationID, "automation", "", "Automation ID (required)")
	deleteCmd.Flags().StringVarP(&deleteWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	deleteCmd.Flags().BoolVarP(&deleteConfirm, "confirm", "y", false, "Confirm deletion (required)")
	deleteCmd.Flags().BoolVarP(&deleteSimple, "simple", "s", false, "Simple output format")
}

func runDelete(cmd *cobra.Command, args []string) error {
	if deleteAutomationID == "" {
		return fmt.Errorf("automation ID is required. Use --automation flag")
	}
	if deleteWorkspace == "" {
		return fmt.Errorf("workspace ID is required. Use --workspace flag")
	}
	if !deleteConfirm {
		return fmt.Errorf("deletion requires confirmation. Use --confirm or -y flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)
	client.SetProject(deleteWorkspace)

	mutation := `
		mutation DeleteAutomation($id: String!) {
			deleteAutomation(id: $id)
		}
	`

	variables := map[string]interface{}{
		"id": deleteAutomationID,
	}

	result, err := client.ExecuteQuery(mutation, variables)
	if err != nil {
		return fmt.Errorf("failed to delete automation: %w", err)
	}

	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	var response DeleteAutomationResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if deleteSimple {
		if response.DeleteAutomation {
			fmt.Printf("Deleted automation: %s\n", deleteAutomationID)
		} else {
			fmt.Printf("Failed to delete automation: %s\n", deleteAutomationID)
		}
	} else {
		if response.DeleteAutomation {
			fmt.Printf("Successfully deleted automation\n\n")
			fmt.Printf("Automation ID: %s\n", deleteAutomationID)
			fmt.Printf("Status: Permanently deleted\n")
			fmt.Printf("\nThis action cannot be undone. The automation has been permanently removed.\n")
		} else {
			fmt.Printf("Failed to delete automation\n\n")
			fmt.Printf("Automation ID: %s\n", deleteAutomationID)
			fmt.Printf("Status: Deletion failed\n")
			fmt.Printf("\nThe automation may not exist or you may not have permission to delete it.\n")
		}
	}

	return nil
}
