package charts

import (
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
