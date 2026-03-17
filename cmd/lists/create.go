package lists

import (
	"fmt"
	"strings"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

type CreateTodoListResponse struct {
	CreateTodoList struct {
		ID       string  `json:"id"`
		UID      string  `json:"uid"`
		Title    string  `json:"title"`
		Position float64 `json:"position"`
	} `json:"createTodoList"`
}

type MaxPositionResponse struct {
	TodoLists []struct {
		Position float64 `json:"position"`
	} `json:"todoLists"`
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

	// Get current max position
	fmt.Printf("Getting current lists in workspace %s...\n", createWorkspace)
	maxPos, err := getMaxPosition(client, createWorkspace)
	if err != nil {
		return fmt.Errorf("failed to get max position: %w", err)
	}

	increment := 65535.0
	startPos := maxPos + increment

	if createReverse {
		for i, j := 0, len(validNames)-1; i < j; i, j = i+1, j-1 {
			validNames[i], validNames[j] = validNames[j], validNames[i]
		}
	}

	fmt.Printf("\nCreating %d lists...\n", len(validNames))
	var createdCount int

	for i, name := range validNames {
		position := startPos + (float64(i) * increment)

		mutation := fmt.Sprintf(`
			mutation CreateTodoList {
				createTodoList(input: {
					projectId: "%s"
					title: "%s"
					position: %f
				}) {
					id
					uid
					title
					position
				}
			}
		`, createWorkspace, name, position)

		var response CreateTodoListResponse
		if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
			fmt.Printf("Failed to create list '%s': %v\n", name, err)
			continue
		}

		list := response.CreateTodoList
		createdCount++
		fmt.Printf("Created list '%s' (ID: %s)\n", list.Title, list.ID)
	}

	fmt.Printf("\nSuccessfully created %d out of %d lists\n", createdCount, len(validNames))

	return nil
}

func getMaxPosition(client *common.Client, workspaceID string) (float64, error) {
	query := `query GetProjectLists($projectId: String!) {
		todoLists(projectId: $projectId) {
			position
		}
	}`

	variables := map[string]interface{}{
		"projectId": workspaceID,
	}

	var response MaxPositionResponse
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return 0, err
	}

	maxPos := 0.0
	for _, list := range response.TodoLists {
		if list.Position > maxPos {
			maxPos = list.Position
		}
	}

	return maxPos, nil
}
