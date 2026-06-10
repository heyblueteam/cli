package bootstrap

import (
	"fmt"

	"github.com/spf13/cobra"
)

var applyCmd = &cobra.Command{
	Use:     "apply",
	Short:   "Apply a bootstrap JSON config",
	Example: `  blue bootstrap apply --file workspace.json --confirm`,
	RunE:    runApply,
}

var (
	applyFile    string
	applyConfirm bool
)

func init() {
	applyCmd.Flags().StringVar(&applyFile, "file", "", "Bootstrap JSON file (required)")
	applyCmd.Flags().BoolVarP(&applyConfirm, "confirm", "y", false, "Actually create resources")
}

func runApply(cmd *cobra.Command, args []string) error {
	if applyFile == "" {
		return fmt.Errorf("bootstrap file is required. Use --file")
	}
	cfg, err := readConfig(applyFile)
	if err != nil {
		return err
	}
	if cfg.Workspace.Name == "" {
		return fmt.Errorf("workspace.name is required")
	}
	if !applyConfirm {
		fmt.Printf("Would create workspace %q with %d lists, %d tags, and %d fields. Pass --confirm to apply.\n", cfg.Workspace.Name, len(cfg.Lists), len(cfg.Tags), len(cfg.Fields))
		return nil
	}

	client, err := newClient()
	if err != nil {
		return err
	}
	workspace, err := createWorkspace(client, cfg.Workspace)
	if err != nil {
		return err
	}
	fmt.Printf("Created workspace %s (%s)\n", workspace.Name, workspace.ID)
	client.SetProject(workspace.ID)

	for i, list := range cfg.Lists {
		if compact(list.Title) == "" {
			continue
		}
		created, err := createList(client, workspace.ID, list, i)
		if err != nil {
			return err
		}
		fmt.Printf("Created list %s (%s)\n", created.Title, created.ID)
	}
	for _, tag := range cfg.Tags {
		if compact(tag.Title) == "" || compact(tag.Color) == "" {
			continue
		}
		created, err := createTag(client, tag)
		if err != nil {
			return err
		}
		fmt.Printf("Created tag %s (%s)\n", created.Title, created.ID)
	}
	for _, field := range cfg.Fields {
		if compact(field.Name) == "" || compact(field.Type) == "" {
			continue
		}
		created, err := createField(client, field)
		if err != nil {
			return err
		}
		fmt.Printf("Created field %s (%s)\n", created.Name, created.ID)
	}

	fmt.Printf("\nBootstrap complete. Workspace ID: %s\n", workspace.ID)
	return nil
}

type createdWorkspace struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type createdList struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type createdTag struct {
	ID    string `json:"id"`
	Title string `json:"title"`
}

type createdField struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func createWorkspace(client graphqlClient, cfg WorkspaceConfig) (*createdWorkspace, error) {
	input := map[string]interface{}{"companyId": client.GetCompanyID(), "name": cfg.Name}
	if cfg.Description != "" {
		input["description"] = cfg.Description
	}
	if cfg.Color != "" {
		input["color"] = cfg.Color
	}
	if cfg.Icon != "" {
		input["icon"] = cfg.Icon
	}
	if cfg.Category != "" {
		input["category"] = cfg.Category
	}
	if cfg.TemplateID != "" {
		input["templateId"] = cfg.TemplateID
	}
	query := `mutation CreateWorkspace($input: CreateProjectInput!) { createProject(input: $input) { id name slug } }`
	var response struct {
		CreateProject createdWorkspace `json:"createProject"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"input": input}, &response); err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}
	return &response.CreateProject, nil
}

func createList(client graphqlClient, workspaceID string, cfg ListConfig, index int) (*createdList, error) {
	query := `mutation CreateList($input: CreateTodoListInput!) { createTodoList(input: $input) { id title } }`
	variables := map[string]interface{}{"input": map[string]interface{}{"projectId": workspaceID, "title": cfg.Title, "position": float64(index+1) * 65535}}
	var response struct {
		CreateTodoList createdList `json:"createTodoList"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return nil, fmt.Errorf("failed to create list %q: %w", cfg.Title, err)
	}
	return &response.CreateTodoList, nil
}

func createTag(client graphqlClient, cfg TagConfig) (*createdTag, error) {
	query := `mutation CreateTag($input: CreateTagInput!) { createTag(input: $input) { id title } }`
	variables := map[string]interface{}{"input": map[string]interface{}{"title": cfg.Title, "color": cfg.Color}}
	var response struct {
		CreateTag createdTag `json:"createTag"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return nil, fmt.Errorf("failed to create tag %q: %w", cfg.Title, err)
	}
	return &response.CreateTag, nil
}

func createField(client graphqlClient, cfg FieldConfig) (*createdField, error) {
	input := map[string]interface{}{"name": cfg.Name, "type": cfg.Type}
	if cfg.Description != "" {
		input["description"] = cfg.Description
	}
	for key, value := range cfg.Settings {
		input[key] = value
	}
	query := `mutation CreateField($input: CreateCustomFieldInput!) { createCustomField(input: $input) { id name type } }`
	var response struct {
		CreateCustomField createdField `json:"createCustomField"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"input": input}, &response); err != nil {
		return nil, fmt.Errorf("failed to create field %q: %w", cfg.Name, err)
	}
	if len(cfg.Options) > 0 {
		if err := createFieldOptions(client, response.CreateCustomField.ID, cfg.Options); err != nil {
			return nil, err
		}
	}
	return &response.CreateCustomField, nil
}

func createFieldOptions(client graphqlClient, fieldID string, options []map[string]interface{}) error {
	query := `mutation CreateFieldOptions($input: CreateCustomFieldOptionsInput!) { createCustomFieldOptions(input: $input) { id title } }`
	variables := map[string]interface{}{"input": map[string]interface{}{"customFieldId": fieldID, "customFieldOptions": options}}
	var response struct{}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to create options for field %s: %w", fieldID, err)
	}
	return nil
}

type graphqlClient interface {
	GetCompanyID() string
	SetProject(string)
	ExecuteQueryWithResult(string, map[string]interface{}, interface{}) error
}
