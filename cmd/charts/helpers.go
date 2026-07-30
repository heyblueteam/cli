package charts

import (
	"encoding/json"
	"fmt"

	"github.com/heyblueteam/cli/common"
)

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
