package databases

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/heyblueteam/cli/common"
)

// resolveWorkspaceID resolves the --workspace flag (ID or slug) to the real
// workspace ID. The REST tools require the true workspace ID in their input:
// it is re-authenticated server-side and must match the caller's scoped
// workspace.
func resolveWorkspaceID(client *common.Client, workspace string) (string, error) {
	client.SetProject(workspace)
	return client.ResolveProjectID(workspace)
}

func printJSON(value interface{}) error {
	out, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

func requireFlag(value, flag string) error {
	if value == "" {
		return fmt.Errorf("%s is required. Use --%s flag", flag, flag)
	}
	return nil
}

// databasesClient loads config and builds a client scoped to the --workspace
// flag, returning the resolved workspace ID for REST input.
func databasesClient(workspace string) (*common.Client, string, error) {
	config, err := common.LoadConfig()
	if err != nil {
		return nil, "", fmt.Errorf("failed to load config: %w", err)
	}
	client := common.NewClient(config)
	workspaceID, err := resolveWorkspaceID(client, workspace)
	if err != nil {
		return nil, "", err
	}
	return client, workspaceID, nil
}

// queryValues builds the query string for GET resource routes.
func queryValues(pairs ...string) url.Values {
	v := url.Values{}
	for i := 0; i+1 < len(pairs); i += 2 {
		v.Set(pairs[i], pairs[i+1])
	}
	return v
}
