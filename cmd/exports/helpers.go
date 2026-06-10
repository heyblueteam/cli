package exports

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

func printQueued(kind string) {
	fmt.Printf("%s export queued. Blue will email the finished CSV to your account.\n", kind)
	fmt.Println("Exports for records and reports are rate-limited to one request per 50 seconds per token.")
}
