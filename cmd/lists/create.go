package lists

import (
	"fmt"
	"strings"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

type CreateTodoListResponse struct {
	CreateTodoList struct {
		ID    string `json:"id"`
		UID   string `json:"uid"`
		Title string `json:"title"`
	} `json:"createTodoList"`
}

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create lists in a workspace",
	Long:  "Create one or more lists in a workspace.",
	Example: `  blue lists create --workspace <id> --names "To Do,In Progress,Done"
  blue lists create --workspace <id> --names "Backlog" --reverse`,
	RunE: runCreate,
}

var (
	createWorkspace string
	createNames     string
	createReverse   bool
)

func init() {
	createCmd.Flags().StringVarP(&createWorkspace, "workspace", "w", "", "Workspace ID (required)")
	createCmd.Flags().StringVar(&createNames, "names", "", "Comma-separated list names (required)")
	createCmd.Flags().BoolVar(&createReverse, "reverse", false, "Create lists in reverse order")
}

func runCreate(cmd *cobra.Command, args []string) error {
	if createWorkspace == "" {
		return fmt.Errorf("workspace is required. Use --workspace flag")
	}
	if createNames == "" {
		return fmt.Errorf("list names are required. Use --names flag with comma-separated values")
	}

	listNames := strings.Split(createNames, ",")
	var validNames []string
	for _, name := range listNames {
		name = strings.TrimSpace(name)
		if name != "" {
			validNames = append(validNames, name)
		}
	}

	if len(validNames) == 0 {
		return fmt.Errorf("no valid list names provided")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)

	if createReverse {
		for i, j := 0, len(validNames)-1; i < j; i, j = i+1, j-1 {
			validNames[i], validNames[j] = validNames[j], validNames[i]
		}
	}

	fmt.Printf("\nCreating %d lists...\n", len(validNames))
	createdCount, failures := createLists(client, createWorkspace, validNames, func(name string, list CreateTodoListResponse) {
		fmt.Printf("Created list '%s' (ID: %s)\n", list.CreateTodoList.Title, list.CreateTodoList.ID)
	})

	if len(failures) > 0 {
		fmt.Printf("\nCreated %d out of %d lists\n", createdCount, len(validNames))
		for _, failure := range failures {
			fmt.Printf("Failed to create list '%s': %v\n", failure.name, failure.err)
		}
		return fmt.Errorf("failed to create %d of %d lists", len(failures), len(validNames))
	}

	fmt.Printf("\nSuccessfully created %d out of %d lists\n", createdCount, len(validNames))

	return nil
}

const createTodoListMutation = `
	mutation CreateTodoList($input: CreateTodoListInput!) {
		createTodoList(input: $input) {
			id
			uid
			title
		}
	}
`

type listGraphQLClient interface {
	ExecuteQueryWithResult(query string, variables map[string]interface{}, result interface{}) error
}

type listCreateFailure struct {
	name string
	err  error
}

func createLists(
	client listGraphQLClient,
	workspaceID string,
	names []string,
	onCreated func(name string, response CreateTodoListResponse),
) (int, []listCreateFailure) {
	var previousID string
	createdCount := 0
	var failures []listCreateFailure

	for _, name := range names {
		input := map[string]interface{}{
			"projectId": workspaceID,
			"title":     name,
		}
		if previousID != "" {
			input["previousId"] = previousID
		}

		var response CreateTodoListResponse
		if err := client.ExecuteQueryWithResult(createTodoListMutation, map[string]interface{}{"input": input}, &response); err != nil {
			failures = append(failures, listCreateFailure{name: name, err: err})
			continue
		}
		if response.CreateTodoList.ID == "" {
			failures = append(failures, listCreateFailure{
				name: name,
				err:  fmt.Errorf("API response did not include a list ID"),
			})
			continue
		}

		createdCount++
		previousID = response.CreateTodoList.ID
		if onCreated != nil {
			onCreated(name, response)
		}
	}

	return createdCount, failures
}
