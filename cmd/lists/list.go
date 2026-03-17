package lists

import (
	"fmt"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

type TodoListsResponse struct {
	TodoLists []common.TodoList `json:"todoLists"`
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List lists in a workspace",
	Long:  "Get all lists in a specific workspace with optional detail levels.",
	Example: `  blue lists list --workspace <id>
  blue lists list --workspace <id> --simple`,
	RunE: runList,
}

var (
	listWorkspace string
	listSimple    bool
)

func init() {
	listCmd.Flags().StringVarP(&listWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	listCmd.Flags().BoolVarP(&listSimple, "simple", "s", false, "Show only basic list information")
}

func runList(cmd *cobra.Command, args []string) error {
	if listWorkspace == "" {
		return fmt.Errorf("workspace is required. Use --workspace flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)

	var query string
	if listSimple {
		query = `query GetProjectLists($projectId: String!) {
			todoLists(projectId: $projectId) {
				id
				title
				position
				todosCount
			}
		}`
	} else {
		query = `query GetProjectLists($projectId: String!) {
			todoLists(projectId: $projectId) {
				id
				uid
				title
				position
				createdAt
				updatedAt
				isDisabled
				isLocked
				completed
				editable
				deletable
				todosCount
				todosMaxPosition
			}
		}`
	}

	variables := map[string]interface{}{
		"projectId": listWorkspace,
	}

	var response TodoListsResponse
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}

	lists := response.TodoLists

	fmt.Printf("\n=== Lists in Workspace %s ===\n", listWorkspace)
	fmt.Printf("Total lists: %d\n\n", len(lists))

	if len(lists) == 0 {
		fmt.Println("No lists found in this workspace.")
		fmt.Printf("\nCreate lists using:\n")
		fmt.Printf("  blue lists create --workspace %s --names \"To Do,In Progress,Done\"\n", listWorkspace)
		return nil
	}

	for i, list := range lists {
		if listSimple {
			fmt.Printf("%d. %s\n", i+1, list.Title)
			fmt.Printf("   ID: %s\n", list.ID)
			fmt.Printf("   Position: %.0f\n", list.Position)
			fmt.Printf("   Tasks: %d\n\n", list.TodosCount)
		} else {
			fmt.Printf("%d. %s\n", i+1, list.Title)
			fmt.Printf("   ID: %s\n", list.ID)
			fmt.Printf("   UID: %s\n", list.UID)
			fmt.Printf("   Position: %.0f\n", list.Position)
			fmt.Printf("   Total tasks: %d\n", list.TodosCount)
			fmt.Printf("   Max position: %.0f\n", list.TodosMaxPosition)
			fmt.Printf("   Disabled: %v\n", list.IsDisabled)
			fmt.Printf("   Locked: %v\n", list.IsLocked)
			fmt.Printf("   Completed: %v\n", list.Completed)
			fmt.Printf("   Editable: %v\n", list.Editable)
			fmt.Printf("   Deletable: %v\n", list.Deletable)
			fmt.Printf("   Created: %s\n", list.CreatedAt)
			fmt.Printf("   Updated: %s\n", list.UpdatedAt)
			fmt.Println()
		}
	}

	return nil
}
