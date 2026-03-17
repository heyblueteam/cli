package dependencies

import (
	"fmt"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a dependency between records",
	Long:  "Create a dependency relationship between two records.",
	Example: `  blue dependencies create --record <id> --other-record <id>
  blue deps create -r <id> --other-record <id> --type BLOCKING -w <workspace> -s`,
	RunE: runCreate,
}

var (
	createRecord      string
	createOtherRecord string
	createType        string
	createWorkspace   string
	createSimple      bool
)

func init() {
	createCmd.Flags().StringVarP(&createRecord, "record", "r", "", "Record/Todo ID (required)")
	createCmd.Flags().StringVar(&createOtherRecord, "other-record", "", "Other Record/Todo ID to create dependency with (required)")
	createCmd.Flags().StringVar(&createType, "type", "BLOCKED_BY", "Dependency type: BLOCKING or BLOCKED_BY (default: BLOCKED_BY)")
	createCmd.Flags().StringVarP(&createWorkspace, "workspace", "w", "", "Workspace ID or slug (optional - for context)")
	createCmd.Flags().BoolVarP(&createSimple, "simple", "s", false, "Show simple output")
}

// DependencyTodo represents a minimal todo in dependency responses
type DependencyTodo struct {
	ID    string `json:"id"`
	UID   string `json:"uid"`
	Title string `json:"title"`
}

func runCreate(cmd *cobra.Command, args []string) error {
	if createRecord == "" {
		return fmt.Errorf("record ID is required. Use --record flag")
	}
	if createOtherRecord == "" {
		return fmt.Errorf("other record ID is required. Use --other-record flag")
	}
	if createType != "BLOCKING" && createType != "BLOCKED_BY" {
		return fmt.Errorf("type must be BLOCKING or BLOCKED_BY")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := common.NewClient(config)
	if createWorkspace != "" {
		client.SetProject(createWorkspace)
	}

	mutation := `
		mutation CreateTodoDependency($input: CreateTodoDependencyInput!) {
			createTodoDependency(input: $input) {
				id
				uid
				title
				dependOn {
					id
					uid
					title
				}
				dependBy {
					id
					uid
					title
				}
			}
		}
	`

	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"type":        createType,
			"todoId":      createRecord,
			"otherTodoId": createOtherRecord,
		},
	}

	var response struct {
		CreateTodoDependency struct {
			ID       string           `json:"id"`
			UID      string           `json:"uid"`
			Title    string           `json:"title"`
			DependOn []DependencyTodo `json:"dependOn"`
			DependBy []DependencyTodo `json:"dependBy"`
		} `json:"createTodoDependency"`
	}

	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to create dependency: %w", err)
	}

	todo := response.CreateTodoDependency

	if createSimple {
		fmt.Printf("Dependency created: %s %s %s\n", createRecord, createType, createOtherRecord)
	} else {
		fmt.Printf("=== Dependency Created Successfully ===\n")
		fmt.Printf("Record: %s (%s)\n", todo.Title, todo.ID)
		fmt.Printf("Type: %s\n", createType)
		fmt.Printf("Other Record: %s\n", createOtherRecord)
		fmt.Println()

		if len(todo.DependOn) > 0 {
			fmt.Printf("Depends On (%d):\n", len(todo.DependOn))
			for _, dep := range todo.DependOn {
				fmt.Printf("  -> %s (%s)\n", dep.Title, dep.ID)
			}
		}
		if len(todo.DependBy) > 0 {
			fmt.Printf("Blocked By (%d):\n", len(todo.DependBy))
			for _, dep := range todo.DependBy {
				fmt.Printf("  <- %s (%s)\n", dep.Title, dep.ID)
			}
		}
		fmt.Printf("\nDependency created successfully!\n")
	}

	return nil
}
