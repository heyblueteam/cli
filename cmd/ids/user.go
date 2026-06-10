package ids

import (
	"fmt"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

var userCmd = &cobra.Command{
	Use:     "user",
	Aliases: []string{"users"},
	Short:   "Look up user IDs",
	Example: `  blue ids user
  blue ids user --workspace <id>
  blue ids user --search alex`,
	RunE: runUserIDs,
}

func init() {
	addCommonFlags(userCmd, false)
}

func runUserIDs(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}

	var rows []idRow
	if idsWorkspace != "" {
		client.SetProject(idsWorkspace)
		projectID, err := client.ResolveProjectID(idsWorkspace)
		if err != nil {
			return fmt.Errorf("failed to resolve workspace: %w", err)
		}
		client.SetProject(projectID)
		rows, err = fetchProjectUsers(client, projectID)
		if err != nil {
			return err
		}
	} else {
		rows, err = fetchCompanyUsers(client)
		if err != nil {
			return err
		}
	}

	return printRows(limitRows(rows))
}

func fetchProjectUsers(client *common.Client, projectID string) ([]idRow, error) {
	query := `query ProjectUserIDs($projectId: String!, $first: Int!) {
		projectUserList(filter: { projectIds: [$projectId] }, first: $first, orderBy: firstName_ASC) {
			items { id uid firstName lastName fullName email }
		}
	}`
	variables := map[string]interface{}{"projectId": projectID, "first": idsLimit}
	var response struct {
		ProjectUserList struct {
			Items []common.User `json:"items"`
		} `json:"projectUserList"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return nil, fmt.Errorf("failed to look up user IDs: %w", err)
	}
	return userRows(response.ProjectUserList.Items, projectID), nil
}

func fetchCompanyUsers(client *common.Client) ([]idRow, error) {
	query := `query CompanyUserIDs($companyId: String!, $search: String, $first: Int!) {
		companyUserList(companyId: $companyId, search: $search, first: $first, orderBy: firstName_ASC) {
			users { id firstName lastName fullName email }
		}
	}`
	variables := map[string]interface{}{"companyId": client.GetCompanyID(), "search": nil, "first": idsLimit}
	if idsSearch != "" {
		variables["search"] = idsSearch
	}
	var response struct {
		CompanyUserList struct {
			Users []common.User `json:"users"`
		} `json:"companyUserList"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return nil, fmt.Errorf("failed to look up user IDs: %w", err)
	}
	return userRows(response.CompanyUserList.Users, ""), nil
}

func userRows(users []common.User, workspace string) []idRow {
	rows := make([]idRow, 0, len(users))
	for _, user := range users {
		name := user.FullName
		if name == "" {
			name = user.FirstName + " " + user.LastName
		}
		if !matchesSearch(name, user.Email) {
			continue
		}
		rows = append(rows, idRow{Type: "user", Name: name, ID: user.ID, UID: user.UID, Extra: user.Email, Workspace: workspace})
	}
	return rows
}
