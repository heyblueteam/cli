package dashboards

import (
	"fmt"
	"strings"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

var shareCmd = &cobra.Command{
	Use:   "share",
	Short: "Share a dashboard with users",
	Long:  "Add or update user access to a dashboard.",
	Example: `  blue dashboards share --dashboard <id> --users "user1:EDITOR,user2:VIEWER"`,
	RunE: runShare,
}

var (
	shareDashboard string
	shareUsers     string
)

func init() {
	shareCmd.Flags().StringVar(&shareDashboard, "dashboard", "", "Dashboard ID (required)")
	shareCmd.Flags().StringVar(&shareUsers, "users", "", "Users and roles (format: userId:EDITOR,userId:VIEWER)")
}

func runShare(cmd *cobra.Command, args []string) error {
	if shareDashboard == "" {
		return fmt.Errorf("dashboard ID is required. Use --dashboard flag")
	}
	if shareUsers == "" {
		return fmt.Errorf("users are required. Use --users flag (format: userId:EDITOR,userId:VIEWER)")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := common.NewClient(config)

	// Parse users
	pairs := strings.Split(shareUsers, ",")
	var userInputs []string
	for _, pair := range pairs {
		parts := strings.SplitN(strings.TrimSpace(pair), ":", 2)
		if len(parts) != 2 {
			return fmt.Errorf("invalid user format: %s (expected userId:ROLE)", pair)
		}
		userID := strings.TrimSpace(parts[0])
		role := strings.ToUpper(strings.TrimSpace(parts[1]))
		if role != "EDITOR" && role != "VIEWER" {
			return fmt.Errorf("invalid role '%s'. Must be EDITOR or VIEWER", role)
		}
		userInputs = append(userInputs, fmt.Sprintf(`{userId: "%s", role: %s}`, userID, role))
	}

	mutation := fmt.Sprintf(`
		mutation EditDashboard {
			editDashboard(input: {
				id: "%s"
				dashboardUsers: [%s]
			}) {
				id
				title
				dashboardUsers {
					role
					user {
						id
						fullName
					}
				}
			}
		}
	`, shareDashboard, strings.Join(userInputs, ", "))

	var response struct {
		EditDashboard struct {
			ID             string          `json:"id"`
			Title          string          `json:"title"`
			DashboardUsers []DashboardUser `json:"dashboardUsers"`
		} `json:"editDashboard"`
	}

	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		return fmt.Errorf("failed to share dashboard: %w", err)
	}

	d := response.EditDashboard
	fmt.Printf("Dashboard '%s' sharing updated\n", d.Title)
	if len(d.DashboardUsers) > 0 {
		fmt.Printf("Shared with:\n")
		for _, u := range d.DashboardUsers {
			fmt.Printf("  - %s (%s)\n", u.User.FullName, u.Role)
		}
	}

	return nil
}
