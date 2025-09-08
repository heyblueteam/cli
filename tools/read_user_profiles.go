package tools

import (
	"flag"
	"fmt"
	"sort"
	"strings"
	
	"demo-builder/common"
)

// CompanyInfo holds company display information
type CompanyInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// ProjectInfo holds project display information
type ProjectInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// RunReadUserProfiles lists user profiles with company/project options
func RunReadUserProfiles(args []string) error {
	// Create flag set for this command
	fs := flag.NewFlagSet("read-user-profiles", flag.ExitOnError)
	
	// Parse command line flags
	simple := fs.Bool("simple", false, "Show only basic user info")
	projectID := fs.String("project", "", "Project ID to get users from (if not specified, attempts company-wide)")
	companyID := fs.String("company", "", "Company ID (optional, uses default from config if not specified)")
	search := fs.String("search", "", "Search users by name or email")
	first := fs.Int("first", 50, "Number of users to fetch (default: 50)")
	
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}
	
	// Load configuration
	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Create client using shared auth
	client := common.NewClient(config)
	
	// Use provided company ID or default from config
	targetCompanyID := *companyID
	if targetCompanyID == "" {
		targetCompanyID = client.GetCompanyID()
	}
	
	// Get company information for better display
	companyInfo := getCompanyInfo(client, targetCompanyID)
	
	var users []common.User
	var totalCount int
	var err2 error
	var projectInfo *ProjectInfo
	
	if *projectID != "" {
		// Get project information for better display
		projectInfo = getProjectInfo(client, *projectID)
		projectDisplay := getProjectDisplayName(projectInfo)
		fmt.Printf("👥 Fetching users from project %s...\n", projectDisplay)
		users, totalCount, err2 = getProjectUsers(client, *projectID, *first, *search)
	} else {
		// Company-wide user listing
		companyDisplay := getCompanyDisplayName(companyInfo)
		fmt.Printf("👥 Fetching company-wide users (%s)...\n", companyDisplay)
		users, totalCount, err2 = getCompanyUsers(client, targetCompanyID, *first, *search)
	}
	
	if err2 != nil {
		return err2
	}
	
	// Apply search filter
	var filteredUsers []common.User
	for _, user := range users {
		if *search != "" {
			searchTerm := strings.ToLower(*search)
			if !strings.Contains(strings.ToLower(user.FullName), searchTerm) &&
			   !strings.Contains(strings.ToLower(user.FirstName), searchTerm) &&
			   !strings.Contains(strings.ToLower(user.LastName), searchTerm) &&
			   !strings.Contains(strings.ToLower(user.Email), searchTerm) {
				continue
			}
		}
		filteredUsers = append(filteredUsers, user)
	}
	
	// Sort users by name
	sort.Slice(filteredUsers, func(i, j int) bool {
		nameI := filteredUsers[i].FullName
		if nameI == "" {
			nameI = fmt.Sprintf("%s %s", filteredUsers[i].FirstName, filteredUsers[i].LastName)
		}
		nameJ := filteredUsers[j].FullName
		if nameJ == "" {
			nameJ = fmt.Sprintf("%s %s", filteredUsers[j].FirstName, filteredUsers[j].LastName)
		}
		return strings.ToLower(nameI) < strings.ToLower(nameJ)
	})
	
	// Display results
	if len(filteredUsers) == 0 {
		fmt.Println("No users found.")
		if *search != "" {
			fmt.Printf("Search term: \"%s\"\n", *search)
		}
		return nil
	}
	
	if *search != "" {
		fmt.Printf("🔍 Search: \"%s\" - Found %d users (of %d total)\n\n", *search, len(filteredUsers), totalCount)
	} else {
		fmt.Printf("\n👥 Found %d users:\n\n", len(filteredUsers))
	}
	
	if *simple {
		// Simple tabular format with full IDs
		fmt.Printf("%-25s %-25s %-35s\n", "ID", "Name", "Email")
		fmt.Printf("%-25s %-25s %-35s\n", "-------------------------", "-------------------------", "-----------------------------------")
		
		for _, user := range filteredUsers {
			name := user.FullName
			if name == "" {
				name = strings.TrimSpace(fmt.Sprintf("%s %s", user.FirstName, user.LastName))
			}
			if name == "" {
				name = "N/A"
			}
			
			fmt.Printf("%-25s %-25s %-35s\n", 
				user.ID,  // No truncation
				truncateString(name, 25),
				truncateString(user.Email, 35))
		}
	} else {
		// Detailed format
		for i, user := range filteredUsers {
			fmt.Printf("👤 User %d:\n", i+1)
			fmt.Printf("   ID: %s\n", user.ID)
			if user.UID != "" {
				fmt.Printf("   UID: %s\n", user.UID)
			}
			
			// Display name info
			if user.FullName != "" {
				fmt.Printf("   Full Name: %s\n", user.FullName)
			}
			if user.FirstName != "" {
				fmt.Printf("   First Name: %s\n", user.FirstName)
			}
			if user.LastName != "" {
				fmt.Printf("   Last Name: %s\n", user.LastName)
			}
			if user.Email != "" {
				fmt.Printf("   Email: %s\n", user.Email)
			}
			
			if i < len(filteredUsers)-1 {
				fmt.Println()
			}
		}
	}
	
	// Show summary
	fmt.Printf("\n📊 Summary: %d users displayed", len(filteredUsers))
	if *search != "" {
		fmt.Printf(" (filtered from %d total)", totalCount)
	}
	if *projectID != "" {
		projectDisplay := getProjectDisplayName(projectInfo)
		fmt.Printf(" from project %s", projectDisplay)
	} else {
		companyDisplay := getCompanyDisplayName(companyInfo)
		fmt.Printf(" from %s", companyDisplay)
	}
	fmt.Println()
	
	return nil
}

// getProjectUsers fetches users from a specific project using projectUserList
func getProjectUsers(client *common.Client, projectID string, first int, search string) ([]common.User, int, error) {
	// Try projectUserList first
	query := fmt.Sprintf(`
		query ProjectUserList {
			projectUserList(
				filter: { 
					projectIds: ["%s"] 
				}
				first: %d
				orderBy: firstName_ASC
			) {
				items {
					id
					uid
					firstName
					lastName
					fullName
					email
				}
				totalCount
			}
		}
	`, projectID, first)
	
	variables := map[string]interface{}{}
	
	// Response structure for projectUserList
	var response struct {
		ProjectUserList struct {
			Items      []common.User `json:"items"`
			TotalCount int          `json:"totalCount"`
		} `json:"projectUserList"`
	}
	
	if err := client.ExecuteQueryWithResult(query, variables, &response); err == nil {
		return response.ProjectUserList.Items, response.ProjectUserList.TotalCount, nil
	}
	
	// Fallback: Try userList with project filter
	return getUsersWithProjectFilter(client, projectID, first, search)
}

// getCompanyUsers fetches users from a company using companyUserList or userList
func getCompanyUsers(client *common.Client, companyID string, first int, search string) ([]common.User, int, error) {
	// Try companyUserList with correct structure matching frontend
	query := `
		query CompanyUserList($companyId: String!, $notInProjectId: String, $search: String, $first: Int, $after: String, $orderBy: UserOrderByInput, $skip: Int) {
			companyUserList(
				companyId: $companyId
				notInProjectId: $notInProjectId
				search: $search
				first: $first
				after: $after
				orderBy: $orderBy
				skip: $skip
			) {
				users {
					id
					email
					firstName
					lastName
					fullName
					image {
						id
						thumbnail
					}
					isOnline
					lastActiveAt
				}
				pageInfo {
					hasNextPage
					endCursor
				}
				totalCount
			}
		}
	`
	
	variables := map[string]interface{}{
		"companyId": companyID,
		"first":     first,
		"orderBy":   "firstName_ASC",
	}
	
	// Add search parameter if provided
	if search != "" {
		variables["search"] = search
	}
	
	// Response structure for companyUserList (matching frontend)
	type CompanyUserListUser struct {
		ID           string `json:"id"`
		Email        string `json:"email"`
		FirstName    string `json:"firstName"`
		LastName     string `json:"lastName"`
		FullName     string `json:"fullName"`
		Image        *struct {
			ID        string `json:"id"`
			Thumbnail string `json:"thumbnail"`
		} `json:"image"`
		IsOnline     bool   `json:"isOnline"`
		LastActiveAt string `json:"lastActiveAt"`
	}
	
	var companyResponse struct {
		CompanyUserList struct {
			Users    []CompanyUserListUser `json:"users"`
			PageInfo struct {
				HasNextPage bool   `json:"hasNextPage"`
				EndCursor   string `json:"endCursor"`
			} `json:"pageInfo"`
			TotalCount int `json:"totalCount"`
		} `json:"companyUserList"`
	}
	
	if err := client.ExecuteQueryWithResult(query, variables, &companyResponse); err == nil {
		// Convert to common.User format
		var users []common.User
		for _, user := range companyResponse.CompanyUserList.Users {
			users = append(users, common.User{
				ID:        user.ID,
				UID:       "", // Not available in this query
				FirstName: user.FirstName,
				LastName:  user.LastName,
				FullName:  user.FullName,
				Email:     user.Email,
			})
		}
		return users, companyResponse.CompanyUserList.TotalCount, nil
	}
	
	// Fallback: Try userList with company filter
	return getUsersWithCompanyFilter(client, companyID, first, search)
}

// getUsersWithProjectFilter tries userList with project filter
func getUsersWithProjectFilter(client *common.Client, projectID string, first int, search string) ([]common.User, int, error) {
	query := fmt.Sprintf(`
		query UserList {
			userList(
				filter: { 
					projectIds: ["%s"] 
				}
				first: %d
				orderBy: firstName_ASC
			) {
				items {
					id
					uid
					firstName
					lastName
					fullName
					email
				}
				totalCount
			}
		}
	`, projectID, first)
	
	variables := map[string]interface{}{}
	
	var response struct {
		UserList struct {
			Items      []common.User `json:"items"`
			TotalCount int          `json:"totalCount"`
		} `json:"userList"`
	}
	
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return nil, 0, fmt.Errorf("failed to fetch project users: %v", err)
	}
	
	return response.UserList.Items, response.UserList.TotalCount, nil
}

// getUsersWithCompanyFilter tries userList with company filter
func getUsersWithCompanyFilter(client *common.Client, companyID string, first int, search string) ([]common.User, int, error) {
	query := fmt.Sprintf(`
		query UserList {
			userList(
				filter: { 
					companyIds: ["%s"] 
				}
				first: %d
				orderBy: firstName_ASC
			) {
				items {
					id
					uid
					firstName
					lastName
					fullName
					email
				}
				totalCount
			}
		}
	`, companyID, first)
	
	variables := map[string]interface{}{}
	
	var response struct {
		UserList struct {
			Items      []common.User `json:"items"`
			TotalCount int          `json:"totalCount"`
		} `json:"userList"`
	}
	
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return nil, 0, fmt.Errorf("failed to fetch company users: %v", err)
	}
	
	return response.UserList.Items, response.UserList.TotalCount, nil
}

// getCompanyInfo tries to fetch company information for better display
func getCompanyInfo(client *common.Client, companyID string) *CompanyInfo {
	// Try to get company information
	query := `
		query GetCompany($companyId: String!) {
			company(id: $companyId) {
				id
				name
				slug
			}
		}
	`
	
	variables := map[string]interface{}{
		"companyId": companyID,
	}
	
	var response struct {
		Company *CompanyInfo `json:"company"`
	}
	
	if err := client.ExecuteQueryWithResult(query, variables, &response); err == nil && response.Company != nil {
		return response.Company
	}
	
	// Fallback: return minimal info with just ID
	return &CompanyInfo{
		ID:   companyID,
		Name: "",
		Slug: companyID, // Use ID as slug fallback
	}
}

// getProjectInfo tries to fetch project information for better display
func getProjectInfo(client *common.Client, projectID string) *ProjectInfo {
	// Try using projectList to find the specific project
	// We'll get all projects and find the one we want (since we can't filter by specific project ID easily)
	query := `
		query GetProjectInfo($companyIds: [String!]!) {
			projectList(
				filter: { companyIds: $companyIds }
				take: 100
			) {
				items {
					id
					name
					slug
				}
			}
		}
	`
	
	// We need to get the company ID to query projects
	companyID := ""
	if config, err := common.LoadConfig(); err == nil {
		if newClient := common.NewClient(config); newClient != nil {
			companyID = newClient.GetCompanyID()
		}
	}
	
	variables := map[string]interface{}{
		"companyIds": []string{companyID},
	}
	
	var response struct {
		ProjectList struct {
			Items []ProjectInfo `json:"items"`
		} `json:"projectList"`
	}
	
	if err := client.ExecuteQueryWithResult(query, variables, &response); err == nil {
		// Find the specific project by ID
		for _, project := range response.ProjectList.Items {
			if project.ID == projectID {
				return &project
			}
		}
	}
	
	// Fallback: return minimal info with just ID
	return &ProjectInfo{
		ID:   projectID,
		Name: "",
		Slug: projectID, // Use ID as slug fallback
	}
}

// getProjectDisplayName formats project information for display
func getProjectDisplayName(info *ProjectInfo) string {
	if info == nil {
		return "Unknown Project"
	}
	
	if info.Name != "" {
		return info.Name
	} else if info.Slug != "" {
		return info.Slug
	}
	
	return info.ID
}

// getCompanyDisplayName formats company information for display (no slug)
func getCompanyDisplayName(info *CompanyInfo) string {
	if info == nil {
		return "Unknown Company"
	}
	
	if info.Name != "" {
		return info.Name
	} else if info.Slug != "" {
		return info.Slug
	}
	
	return info.ID
}

// Helper function to truncate strings for table display
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen < 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}