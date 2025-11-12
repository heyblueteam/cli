package tools

import (
	"flag"
	"fmt"
	"strings"
	
	"cli/common"
)

// RunInviteUser invites a user to the company or project with specified role
func RunInviteUser(args []string) error {
	// Create flag set for this command
	fs := flag.NewFlagSet("invite-user", flag.ExitOnError)
	
	// Required flags
	email := fs.String("email", "", "Email address of user to invite (required)")
	accessLevel := fs.String("access-level", "", "User access level: OWNER, ADMIN, MEMBER, CLIENT, COMMENT_ONLY (required)")
	
	// Optional flags
	projectID := fs.String("project", "", "Project ID to invite user to (optional, for project-specific invitation)")
	projectIDs := fs.String("projects", "", "Comma-separated list of project IDs to invite user to")
	companyID := fs.String("company", "", "Company ID (uses default from config if not specified)")
	roleID := fs.String("role", "", "Custom role ID (for project-specific roles)")
	
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}
	
	// Validate required parameters
	if *email == "" {
		return fmt.Errorf("email is required")
	}
	
	if *accessLevel == "" {
		return fmt.Errorf("access-level is required (OWNER, ADMIN, MEMBER, CLIENT, COMMENT_ONLY)")
	}
	
	// Validate access level
	validLevels := []string{"OWNER", "ADMIN", "MEMBER", "CLIENT", "COMMENT_ONLY"}
	isValid := false
	for _, level := range validLevels {
		if *accessLevel == level {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("invalid access-level: %s. Valid options: %s", *accessLevel, strings.Join(validLevels, ", "))
	}
	
	// Load configuration
	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Create client using shared auth
	client := common.NewClient(config)
	
	// Build input object
	input := map[string]interface{}{
		"email":       *email,
		"accessLevel": *accessLevel,
	}
	
	// Determine target company ID
	targetCompanyID := *companyID
	if targetCompanyID == "" {
		targetCompanyID = client.GetCompanyID()
	}
	
	// Note: For company-wide invitations, we don't specify companyId in input
	// The company context comes from the X-Bloo-Company-ID header
	
	// Add project-specific parameters
	if *projectID != "" {
		input["projectId"] = *projectID
	}
	
	if *projectIDs != "" {
		projectList := strings.Split(*projectIDs, ",")
		cleanProjectList := make([]string, len(projectList))
		for i, pid := range projectList {
			cleanProjectList[i] = strings.TrimSpace(pid)
		}
		input["projectIds"] = cleanProjectList
	}
	
	if *roleID != "" {
		input["roleId"] = *roleID
	}
	
	// Build GraphQL mutation
	mutation := `
		mutation InviteUser($input: InviteUserInput!) {
			inviteUser(input: $input)
		}
	`
	
	// Prepare variables
	variables := map[string]interface{}{
		"input": input,
	}
	
	fmt.Printf("Inviting user %s with access level %s...\n", *email, *accessLevel)
	if *projectID == "" && *projectIDs == "" {
		fmt.Printf("Target company: %s (company-wide invitation)\n", targetCompanyID)
	}
	if *projectID != "" {
		fmt.Printf("Target project: %s\n", *projectID)
	}
	if *projectIDs != "" {
		fmt.Printf("Target projects: %s\n", *projectIDs)
	}
	if *roleID != "" {
		fmt.Printf("Custom role: %s\n", *roleID)
	}
	
	// Execute mutation
	var response struct {
		InviteUser bool `json:"inviteUser"`
	}
	
	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to invite user: %v", err)
	}
	
	if response.InviteUser {
		fmt.Printf("✅ Successfully sent invitation to %s\n", *email)
		
		// Show invitation details
		fmt.Printf("\nInvitation Details:\n")
		fmt.Printf("• Email: %s\n", *email)
		fmt.Printf("• Access Level: %s\n", *accessLevel)
		if *projectID == "" && *projectIDs == "" {
			fmt.Printf("• Company: %s (company-wide)\n", targetCompanyID)
		}
		
		if *projectID != "" {
			fmt.Printf("• Project: %s\n", *projectID)
		}
		if *projectIDs != "" {
			fmt.Printf("• Projects: %s\n", *projectIDs)
		}
		if *roleID != "" {
			fmt.Printf("• Custom Role: %s\n", *roleID)
		}
		
		fmt.Printf("\nThe invited user will receive an email with instructions to join.\n")
	} else {
		return fmt.Errorf("invitation failed - the API returned false")
	}
	
	return nil
}