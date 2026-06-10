package reports

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/heyblueteam/cli/common"
)

func newClient() (*common.Client, error) {
	config, err := common.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	return common.NewClient(config), nil
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			items = append(items, part)
		}
	}
	return items
}

func parseJSONFlag(name, value string) (interface{}, error) {
	if strings.TrimSpace(value) == "" {
		return nil, nil
	}
	var parsed interface{}
	if err := json.Unmarshal([]byte(value), &parsed); err != nil {
		return nil, fmt.Errorf("invalid JSON for %s: %w", name, err)
	}
	return parsed, nil
}

func printJSON(value interface{}) error {
	out, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func printReport(r Report) {
	fmt.Printf("%s\n", r.Title)
	fmt.Printf("  ID:          %s\n", r.ID)
	if r.Description != "" {
		fmt.Printf("  Description: %s\n", r.Description)
	}
	fmt.Printf("  Creator:     %s\n", r.CreatedBy.FullName)
	fmt.Printf("  Sources:     %d\n", len(r.DataSources))
	fmt.Printf("  Shared:      %d users\n", len(r.ReportUsers))
	fmt.Printf("  Updated:     %s\n", r.UpdatedAt)
}

func resolveProjectIDs(client *common.Client, value string) ([]string, error) {
	items := splitCSV(value)
	if len(items) == 0 {
		return nil, nil
	}
	resolved := make([]string, 0, len(items))
	for _, item := range items {
		client.SetProject(item)
		id, err := client.ResolveProjectID(item)
		if err != nil {
			return nil, err
		}
		resolved = append(resolved, id)
	}
	return resolved, nil
}
