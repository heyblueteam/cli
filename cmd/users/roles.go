package users

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

var rolesCmd = &cobra.Command{
	Use:   "roles",
	Short: "List project/workspace user roles",
	Long:  "List custom user roles for one or more workspaces.",
	Example: `  blue users roles --workspace <id> -s
  blue users roles --workspaces "id1,id2" --format json
  blue users roles -w <id> --format csv`,
	RunE: runRoles,
}

var (
	rolesWorkspace  string
	rolesWorkspaces string
	rolesSimple     bool
	rolesFormat     string
)

func init() {
	rolesCmd.Flags().StringVarP(&rolesWorkspace, "workspace", "w", "", "Workspace ID to get roles for")
	rolesCmd.Flags().StringVar(&rolesWorkspaces, "workspaces", "", "Comma-separated list of workspace IDs to get roles for")
	rolesCmd.Flags().BoolVarP(&rolesSimple, "simple", "s", false, "Show only basic role info")
	rolesCmd.Flags().StringVar(&rolesFormat, "format", "table", "Output format: table, json, csv")
}

// ProjectUserRole represents a custom user role in a project
type ProjectUserRole struct {
	ID                        string    `json:"id"`
	UID                       string    `json:"uid"`
	Name                      string    `json:"name"`
	Description               string    `json:"description"`
	AllowInviteOthers         bool      `json:"allowInviteOthers"`
	AllowMarkRecordsAsDone    bool      `json:"allowMarkRecordsAsDone"`
	ShowOnlyAssignedTodos     bool      `json:"showOnlyAssignedTodos"`
	ShowOnlyMentionedComments bool      `json:"showOnlyMentionedComments"`
	IsActivityEnabled         bool      `json:"isActivityEnabled"`
	IsChatEnabled             bool      `json:"isChatEnabled"`
	IsDocsEnabled             bool      `json:"isDocsEnabled"`
	IsFormsEnabled            bool      `json:"isFormsEnabled"`
	IsWikiEnabled             bool      `json:"isWikiEnabled"`
	IsFilesEnabled            bool      `json:"isFilesEnabled"`
	IsRecordsEnabled          bool      `json:"isRecordsEnabled"`
	IsPeopleEnabled           bool      `json:"isPeopleEnabled"`
	CanDeleteRecords          bool      `json:"canDeleteRecords"`
	CreatedAt                 time.Time `json:"createdAt"`
	UpdatedAt                 time.Time `json:"updatedAt"`
	Project                   struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"project"`
	CustomFields []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"customFields"`
	TodoLists []struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Position int    `json:"position"`
	} `json:"todoLists"`
}

func runRoles(cmd *cobra.Command, args []string) error {
	if rolesWorkspace == "" && rolesWorkspaces == "" {
		return fmt.Errorf("either --workspace or --workspaces is required")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)

	if rolesWorkspace != "" {
		client.SetProject(rolesWorkspace)
	}

	// Build filter based on provided workspaces
	filter := make(map[string]interface{})

	if rolesWorkspace != "" {
		filter["projectId"] = rolesWorkspace
	} else if rolesWorkspaces != "" {
		projectList := strings.Split(rolesWorkspaces, ",")
		cleanProjectList := make([]string, len(projectList))
		for i, pid := range projectList {
			cleanProjectList[i] = strings.TrimSpace(pid)
		}
		filter["projectIds"] = cleanProjectList
	}

	query := `
		query ProjectUserRoles($filter: ProjectUserRoleFilter!) {
			projectUserRoles(filter: $filter) {
				id
				uid
				name
				description
				allowInviteOthers
				allowMarkRecordsAsDone
				showOnlyAssignedTodos
				showOnlyMentionedComments
				isActivityEnabled
				isChatEnabled
				isDocsEnabled
				isFormsEnabled
				isWikiEnabled
				isFilesEnabled
				isRecordsEnabled
				isPeopleEnabled
				canDeleteRecords
				createdAt
				updatedAt
				project {
					id
					name
					slug
				}
			}
		}
	`

	variables := map[string]interface{}{
		"filter": filter,
	}

	var response struct {
		ProjectUserRoles []ProjectUserRole `json:"projectUserRoles"`
	}

	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to fetch project user roles: %w", err)
	}

	roles := response.ProjectUserRoles

	if len(roles) == 0 {
		fmt.Println("No custom user roles found.")
		return nil
	}

	// Handle different output formats
	switch rolesFormat {
	case "json":
		jsonData, err := json.MarshalIndent(roles, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal JSON: %v", err)
		}
		fmt.Println(string(jsonData))
		return nil

	case "csv":
		fmt.Println("ID,UID,Name,Description,Project,ProjectID,AllowInvite,CanMarkDone,ShowOnlyAssigned,ShowOnlyMentioned,Activity,Chat,Docs,Forms,Wiki,Files,Records,People,CanDelete,Created,Updated")
		for _, role := range roles {
			fmt.Printf("%s,%s,%q,%q,%s,%s,%t,%t,%t,%t,%t,%t,%t,%t,%t,%t,%t,%t,%t,%s,%s\n",
				role.ID, role.UID, role.Name, role.Description,
				role.Project.Name, role.Project.ID,
				role.AllowInviteOthers, role.AllowMarkRecordsAsDone,
				role.ShowOnlyAssignedTodos, role.ShowOnlyMentionedComments,
				role.IsActivityEnabled, role.IsChatEnabled,
				role.IsDocsEnabled, role.IsFormsEnabled,
				role.IsWikiEnabled, role.IsFilesEnabled,
				role.IsRecordsEnabled, role.IsPeopleEnabled,
				role.CanDeleteRecords,
				role.CreatedAt.Format("2006-01-02"),
				role.UpdatedAt.Format("2006-01-02"))
		}
		return nil
	}

	// Table format (default)
	fmt.Printf("Found %d custom user role(s)\n\n", len(roles))

	for i, role := range roles {
		if i > 0 {
			fmt.Println()
		}

		fmt.Printf("  %s\n", role.Name)
		fmt.Printf("    ID: %s\n", role.ID)
		fmt.Printf("    UID: %s\n", role.UID)
		if role.Description != "" {
			fmt.Printf("    Description: %s\n", role.Description)
		}
		fmt.Printf("    Project: %s (%s)\n", role.Project.Name, role.Project.ID)
		fmt.Printf("    Created: %s\n", role.CreatedAt.Format("2006-01-02 15:04:05"))
		if !role.UpdatedAt.IsZero() {
			fmt.Printf("    Updated: %s\n", role.UpdatedAt.Format("2006-01-02 15:04:05"))
		}

		if rolesSimple {
			continue
		}

		// Permissions section
		fmt.Printf("\n    Permissions:\n")
		fmt.Printf("       General: Invite Others (%s), Mark Records Done (%s), Delete Records (%s)\n",
			formatBool(role.AllowInviteOthers),
			formatBool(role.AllowMarkRecordsAsDone),
			formatBool(role.CanDeleteRecords))

		fmt.Printf("       Visibility: Show Only Assigned (%s), Show Only Mentioned Comments (%s)\n",
			formatBool(role.ShowOnlyAssignedTodos),
			formatBool(role.ShowOnlyMentionedComments))

		fmt.Printf("       Features: Activity (%s), Chat (%s), Docs (%s)\n",
			formatBool(role.IsActivityEnabled),
			formatBool(role.IsChatEnabled),
			formatBool(role.IsDocsEnabled))

		fmt.Printf("                Forms (%s), Wiki (%s), Files (%s)\n",
			formatBool(role.IsFormsEnabled),
			formatBool(role.IsWikiEnabled),
			formatBool(role.IsFilesEnabled))

		fmt.Printf("                Records (%s), People (%s)\n",
			formatBool(role.IsRecordsEnabled),
			formatBool(role.IsPeopleEnabled))

		if len(role.CustomFields) > 0 {
			fmt.Printf("\n    Custom Fields Access (%d):\n", len(role.CustomFields))
			for _, cf := range role.CustomFields {
				fmt.Printf("       - %s (%s) [%s]\n", cf.Name, cf.Type, cf.ID)
			}
		}

		if len(role.TodoLists) > 0 {
			fmt.Printf("\n    Todo Lists Access (%d):\n", len(role.TodoLists))
			for _, tl := range role.TodoLists {
				fmt.Printf("       - %s (pos: %d) [%s]\n", tl.Title, tl.Position, tl.ID)
			}
		}
	}

	return nil
}

// formatBool returns yes/no for boolean values
func formatBool(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}
