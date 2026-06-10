package ids

import (
	"fmt"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

var tagCmd = &cobra.Command{
	Use:     "tag",
	Aliases: []string{"tags"},
	Short:   "Look up tag IDs",
	Example: `  blue ids tag --workspace <id>
  blue ids tag --workspace <id> --search Bug`,
	RunE: runTagIDs,
}

func init() {
	addCommonFlags(tagCmd, true)
}

func runTagIDs(cmd *cobra.Command, args []string) error {
	if err := requireWorkspace(); err != nil {
		return err
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	client.SetProject(idsWorkspace)
	projectID, err := client.ResolveProjectID(idsWorkspace)
	if err != nil {
		return fmt.Errorf("failed to resolve workspace: %w", err)
	}
	client.SetProject(projectID)

	query := `query TagIDs($projectId: String!, $first: Int!) {
		tagList(filter: { projectIds: [$projectId] }, first: $first, orderBy: title_ASC) {
			items { id uid title color }
		}
	}`
	variables := map[string]interface{}{"projectId": projectID, "first": idsLimit}

	var response struct {
		TagList struct {
			Items []common.Tag `json:"items"`
		} `json:"tagList"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to look up tag IDs: %w", err)
	}

	var rows []idRow
	for _, tag := range response.TagList.Items {
		if !matchesSearch(tag.Title, tag.Color) {
			continue
		}
		rows = append(rows, idRow{Type: "tag", Name: tag.Title, ID: tag.ID, UID: tag.UID, Extra: "color:" + tag.Color, Workspace: projectID})
	}
	return printRows(limitRows(rows))
}
