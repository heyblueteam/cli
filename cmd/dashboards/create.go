package dashboards

import (
	"fmt"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new dashboard",
	Long:  "Create a new dashboard, optionally scoped to a workspace.",
	Example: `  blue dashboards create --title "Sales Dashboard"
  blue dashboards create --title "Sprint Metrics" --workspace <id>`,
	RunE: runCreate,
}

var (
	createTitle     string
	createWorkspace string
)

func init() {
	createCmd.Flags().StringVarP(&createTitle, "title", "t", "", "Dashboard title (required)")
	createCmd.Flags().StringVarP(&createWorkspace, "workspace", "w", "", "Workspace ID (optional, scopes dashboard to workspace)")
}

func runCreate(cmd *cobra.Command, args []string) error {
	if createTitle == "" {
		return fmt.Errorf("title is required. Use --title flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := common.NewClient(config)

	companyID, err := client.ResolveCompanyID()
	if err != nil {
		return fmt.Errorf("failed to resolve company: %w", err)
	}

	input := map[string]interface{}{
		"companyId": companyID,
		"title":     createTitle,
	}
	if createWorkspace != "" {
		client.SetProject(createWorkspace)
		workspaceID, err := client.ResolveProjectID(createWorkspace)
		if err != nil {
			return fmt.Errorf("failed to resolve workspace: %w", err)
		}
		input["projectId"] = workspaceID
	}

	const mutation = `
		mutation CreateDashboard($input: CreateDashboardInput!) {
			createDashboard(input: $input) {
				id
				title
				createdAt
				createdBy {
					id
					fullName
				}
			}
		}
	`

	var response struct {
		CreateDashboard struct {
			ID        string `json:"id"`
			Title     string `json:"title"`
			CreatedAt string `json:"createdAt"`
			CreatedBy struct {
				FullName string `json:"fullName"`
			} `json:"createdBy"`
		} `json:"createDashboard"`
	}

	variables := map[string]interface{}{"input": input}
	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to create dashboard: %w", err)
	}

	d := response.CreateDashboard
	fmt.Printf("Dashboard created!\n")
	fmt.Printf("ID: %s\n", d.ID)
	fmt.Printf("Title: %s\n", d.Title)

	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  blue charts create --dashboard %s --type STAT --title \"Total Records\" --workspace <id> --function COUNT\n", d.ID)

	return nil
}
