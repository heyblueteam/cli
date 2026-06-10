package bootstrap

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "bootstrap",
	Short: "Bootstrap workspaces from JSON",
	Long:  "Generate, apply, and export JSON workspace bootstrap configs for lists, tags, and custom fields.",
}

type Config struct {
	Workspace WorkspaceConfig `json:"workspace"`
	Lists     []ListConfig    `json:"lists"`
	Tags      []TagConfig     `json:"tags"`
	Fields    []FieldConfig   `json:"fields"`
}

type WorkspaceConfig struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Category    string `json:"category,omitempty"`
	TemplateID  string `json:"templateId,omitempty"`
}

type ListConfig struct {
	Title string `json:"title"`
}

type TagConfig struct {
	Title string `json:"title"`
	Color string `json:"color"`
}

type FieldConfig struct {
	Name        string                   `json:"name"`
	Type        string                   `json:"type"`
	Description string                   `json:"description,omitempty"`
	Options     []map[string]interface{} `json:"options,omitempty"`
	Settings    map[string]interface{}   `json:"settings,omitempty"`
}

func init() {
	Cmd.AddCommand(templateCmd)
	Cmd.AddCommand(applyCmd)
	Cmd.AddCommand(exportCmd)
}

func newClient() (*common.Client, error) {
	config, err := common.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	return common.NewClient(config), nil
}

func printJSON(value interface{}) error {
	out, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func readConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("invalid bootstrap JSON: %w", err)
	}
	return &cfg, nil
}

func compact(value string) string {
	return strings.TrimSpace(value)
}
