package workspaces

import (
	"fmt"
	"strings"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

type DeleteProjectResponse struct {
	DeleteProject common.MutationResult `json:"deleteProject"`
}

var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete a workspace",
	Long:  "Permanently delete a workspace. This action cannot be undone.",
	Example: `  blue workspaces delete --workspace <id> --confirm
  blue workspaces delete -w <id> -y`,
	RunE: runDelete,
}

var (
	deleteWorkspace string
	deleteConfirm   bool
	deleteSimple    bool
)

func init() {
	deleteCmd.Flags().StringVarP(&deleteWorkspace, "workspace", "w", "", "Workspace ID to delete (required)")
	deleteCmd.Flags().BoolVarP(&deleteConfirm, "confirm", "y", false, "Confirm deletion (required for safety)")
	deleteCmd.Flags().BoolVarP(&deleteSimple, "simple", "s", false, "Simple output format")
}

func runDelete(cmd *cobra.Command, args []string) error {
	if deleteWorkspace == "" {
		return fmt.Errorf("workspace ID is required. Use --workspace flag")
	}

	if !deleteConfirm {
		return fmt.Errorf("deletion confirmation is required. Use --confirm flag")
	}

	fmt.Printf("WARNING: Deleting workspace '%s' (this action cannot be undone)\n", deleteWorkspace)

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)

	if !deleteSimple {
		fmt.Printf("Deleting workspace '%s'...\n", deleteWorkspace)
	}

	result, err := executeDeleteProject(client, deleteWorkspace)
	if err != nil {
		if strings.Contains(err.Error(), "not authorized") {
			return fmt.Errorf("failed to delete workspace: %w\n\nNote: Workspace deletion requires special permissions. Contact your administrator", err)
		}
		return fmt.Errorf("failed to delete workspace: %w", err)
	}

	if result.Success {
		if deleteSimple {
			fmt.Printf("Workspace %s deleted\n", deleteWorkspace)
		} else {
			fmt.Println("\nWorkspace deleted successfully!")
			if result.OperationID != "" {
				fmt.Printf("Operation ID: %s\n", result.OperationID)
			}
		}
	} else {
		fmt.Println("\nWorkspace deletion failed")
		if result.OperationID != "" {
			fmt.Printf("Operation ID: %s\n", result.OperationID)
		}
	}

	return nil
}

func executeDeleteProject(client *common.Client, projectID string) (*common.MutationResult, error) {
	client.SetProject(projectID)

	mutation := fmt.Sprintf(`
		mutation DeleteProject {
			deleteProject(id: "%s") {
				success
				operationId
			}
		}
	`, projectID)

	var response DeleteProjectResponse
	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		return nil, err
	}

	return &response.DeleteProject, nil
}
