package bootstrap

import (
	"fmt"

	"github.com/spf13/cobra"
)

var exportCmd = &cobra.Command{
	Use:     "export",
	Short:   "Export a workspace bootstrap config",
	Example: `  blue bootstrap export --workspace <id> > workspace.json`,
	RunE:    runExport,
}

var exportWorkspace string

func init() {
	exportCmd.Flags().StringVarP(&exportWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
}

func runExport(cmd *cobra.Command, args []string) error {
	if exportWorkspace == "" {
		return fmt.Errorf("workspace ID is required. Use --workspace")
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	client.SetProject(exportWorkspace)
	projectID, err := client.ResolveProjectID(exportWorkspace)
	if err != nil {
		return err
	}
	client.SetProject(projectID)

	query := `
		query BootstrapExport($projectId: String!) {
			project(id: $projectId) { name description color icon category }
			todoLists(projectId: $projectId) { title }
			tags { title color }
			customFields(filter: { projectId: $projectId }, skip: 0, take: 500) { items { name type description } }
		}
	`
	var response struct {
		Project struct {
			Name        string `json:"name"`
			Description string `json:"description"`
			Color       string `json:"color"`
			Icon        string `json:"icon"`
			Category    string `json:"category"`
		} `json:"project"`
		TodoLists    []ListConfig `json:"todoLists"`
		Tags         []TagConfig  `json:"tags"`
		CustomFields struct {
			Items []FieldConfig `json:"items"`
		} `json:"customFields"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"projectId": projectID}, &response); err != nil {
		return fmt.Errorf("failed to export workspace: %w", err)
	}
	return printJSON(Config{
		Workspace: WorkspaceConfig{
			Name:        response.Project.Name,
			Description: response.Project.Description,
			Color:       response.Project.Color,
			Icon:        response.Project.Icon,
			Category:    response.Project.Category,
		},
		Lists:  response.TodoLists,
		Tags:   response.Tags,
		Fields: response.CustomFields.Items,
	})
}
