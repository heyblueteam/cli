package documents

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/heyblueteam/cli/common"
)

func newClient(workspace string) (*common.Client, error) {
	config, err := common.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)
	if workspace != "" {
		client.SetProject(workspace)
	}
	return client, nil
}

func printJSON(value interface{}) error {
	out, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func readContent(value, file string) (string, error) {
	if value != "" && file != "" {
		return "", fmt.Errorf("use only one of --content or --content-file")
	}
	if file == "" {
		return value, nil
	}
	data, err := os.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("failed to read content file: %w", err)
	}
	return string(data), nil
}

func printDocument(d Document, includeContent bool) {
	typeName := "Document"
	if d.Wiki {
		typeName = "Wiki"
	}
	fmt.Printf("%s: %s\n", typeName, d.Title)
	fmt.Printf("  ID:        %s\n", d.ID)
	fmt.Printf("  Workspace: %s\n", d.Project.Name)
	fmt.Printf("  Author:    %s\n", d.CreatedBy.FullName)
	fmt.Printf("  Updated:   %s\n", d.UpdatedAt)
	if includeContent && d.Content != "" {
		fmt.Printf("\n%s\n", d.Content)
	}
}
