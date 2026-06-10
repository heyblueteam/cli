package savedviews

import (
	"encoding/json"
	"fmt"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:     "saved-views",
	Aliases: []string{"views"},
	Short:   "Manage saved views",
	Long:    "List, inspect, update, delete, and apply saved view configurations.",
}

type SavedView struct {
	ID         string                 `json:"id"`
	UID        string                 `json:"uid"`
	Name       string                 `json:"name"`
	Icon       string                 `json:"icon"`
	Position   float64                `json:"position"`
	IsShared   bool                   `json:"isShared"`
	ViewType   string                 `json:"viewType"`
	ViewConfig map[string]interface{} `json:"viewConfig"`
	CreatedAt  string                 `json:"createdAt"`
	UpdatedAt  string                 `json:"updatedAt"`
}

var savedViewFields = `id uid name icon position isShared viewType viewConfig createdAt updatedAt`

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(getCmd)
	Cmd.AddCommand(updateCmd)
	Cmd.AddCommand(deleteCmd)
	Cmd.AddCommand(applyCmd)
}

func clientFor(workspace string) (*common.Client, error) {
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

func resolveWorkspaceID(client *common.Client, workspace string) (string, error) {
	client.SetProject(workspace)
	id, err := client.ResolveProjectID(workspace)
	if err != nil {
		return "", err
	}
	client.SetProject(id)
	return id, nil
}

func printJSON(value interface{}) error {
	out, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func parseJSONFlag(name, value string) (interface{}, error) {
	if value == "" {
		return nil, nil
	}
	var parsed interface{}
	if err := json.Unmarshal([]byte(value), &parsed); err != nil {
		return nil, fmt.Errorf("invalid JSON for %s: %w", name, err)
	}
	return parsed, nil
}

func printView(v SavedView) {
	fmt.Printf("%s\n", v.Name)
	fmt.Printf("  ID:       %s\n", v.ID)
	fmt.Printf("  Type:     %s\n", v.ViewType)
	fmt.Printf("  Shared:   %t\n", v.IsShared)
	fmt.Printf("  Position: %.2f\n", v.Position)
	fmt.Printf("  Updated:  %s\n", v.UpdatedAt)
}
