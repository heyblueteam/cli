package users

import (
	"fmt"
	"strings"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

var inviteCmd = &cobra.Command{
	Use:   "invite",
	Short: "Invite a user",
	Long:  "Invite a user to the company or a specific workspace with a specified access level.",
	Example: `  blue users invite --email "user@example.com" --access-level MEMBER
  blue users invite --email "admin@example.com" --access-level ADMIN --workspace <id>
  blue users invite --email "client@example.com" --access-level CLIENT --workspaces "id1,id2"`,
	RunE: runInvite,
}

var (
	inviteEmail       string
	inviteAccessLevel string
	inviteWorkspace   string
	inviteWorkspaces  string
	inviteCompany     string
	inviteRole        string
)

func init() {
	inviteCmd.Flags().StringVar(&inviteEmail, "email", "", "Email address of user to invite (required)")
	inviteCmd.Flags().StringVar(&inviteAccessLevel, "access-level", "", "User access level: OWNER, ADMIN, MEMBER, CLIENT, COMMENT_ONLY (required)")
	inviteCmd.Flags().StringVarP(&inviteWorkspace, "workspace", "w", "", "Workspace ID to invite user to (optional)")
	inviteCmd.Flags().StringVar(&inviteWorkspaces, "workspaces", "", "Comma-separated list of workspace IDs to invite user to")
	inviteCmd.Flags().StringVar(&inviteCompany, "company", "", "Company ID (uses default from config if not specified)")
	inviteCmd.Flags().StringVar(&inviteRole, "role", "", "Custom role ID (for workspace-specific roles)")
}

func runInvite(cmd *cobra.Command, args []string) error {
	if inviteEmail == "" {
		return fmt.Errorf("email is required. Use --email flag")
	}
	if inviteAccessLevel == "" {
		return fmt.Errorf("access-level is required (OWNER, ADMIN, MEMBER, CLIENT, COMMENT_ONLY). Use --access-level flag")
	}

	// Validate access level
	validLevels := []string{"OWNER", "ADMIN", "MEMBER", "CLIENT", "COMMENT_ONLY"}
	isValid := false
	for _, level := range validLevels {
		if inviteAccessLevel == level {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid access-level: %s. Valid options: %s", inviteAccessLevel, strings.Join(validLevels, ", "))
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)

	// Build input object
	input := map[string]interface{}{
		"email":       inviteEmail,
		"accessLevel": inviteAccessLevel,
	}

	targetCompanyID := inviteCompany
	if targetCompanyID == "" {
		targetCompanyID = client.GetCompanyID()
	}

	if inviteWorkspace != "" {
		input["projectId"] = inviteWorkspace
	}

	if inviteWorkspaces != "" {
		projectList := strings.Split(inviteWorkspaces, ",")
		cleanProjectList := make([]string, len(projectList))
		for i, pid := range projectList {
			cleanProjectList[i] = strings.TrimSpace(pid)
		}
		input["projectIds"] = cleanProjectList
	}

	if inviteRole != "" {
		input["roleId"] = inviteRole
	}

	mutation := `
		mutation InviteUser($input: InviteUserInput!) {
			inviteUser(input: $input)
		}
	`

	variables := map[string]interface{}{
		"input": input,
	}

	fmt.Printf("Inviting user %s with access level %s...\n", inviteEmail, inviteAccessLevel)
	if inviteWorkspace == "" && inviteWorkspaces == "" {
		fmt.Printf("Target company: %s (company-wide invitation)\n", targetCompanyID)
	}
	if inviteWorkspace != "" {
		fmt.Printf("Target workspace: %s\n", inviteWorkspace)
	}
	if inviteWorkspaces != "" {
		fmt.Printf("Target workspaces: %s\n", inviteWorkspaces)
	}
	if inviteRole != "" {
		fmt.Printf("Custom role: %s\n", inviteRole)
	}

	var response struct {
		InviteUser bool `json:"inviteUser"`
	}

	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to invite user: %w", err)
	}

	if response.InviteUser {
		fmt.Printf("Successfully sent invitation to %s\n", inviteEmail)

		fmt.Printf("\nInvitation Details:\n")
		fmt.Printf("  Email: %s\n", inviteEmail)
		fmt.Printf("  Access Level: %s\n", inviteAccessLevel)
		if inviteWorkspace == "" && inviteWorkspaces == "" {
			fmt.Printf("  Company: %s (company-wide)\n", targetCompanyID)
		}
		if inviteWorkspace != "" {
			fmt.Printf("  Workspace: %s\n", inviteWorkspace)
		}
		if inviteWorkspaces != "" {
			fmt.Printf("  Workspaces: %s\n", inviteWorkspaces)
		}
		if inviteRole != "" {
			fmt.Printf("  Custom Role: %s\n", inviteRole)
		}

		fmt.Printf("\nThe invited user will receive an email with instructions to join.\n")
	} else {
		return fmt.Errorf("invitation failed - the API returned false")
	}

	return nil
}
