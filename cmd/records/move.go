package records

import (
	"fmt"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

type MoveRecordResponse struct {
	UpdateTodos bool `json:"updateTodos"`
}

var moveCmd = &cobra.Command{
	Use:   "move",
	Short: "Move a record to a different list",
	Long:  "Move a record to a different list, optionally across workspaces.",
	Example: `  blue records move --record <id> --list <id> --workspace <id>
  blue records move -r <id> -l <id> -w <id> --simple`,
	RunE: runMove,
}

var (
	moveRecord    string
	moveList      string
	moveWorkspace string
	moveSimple    bool
)

func init() {
	moveCmd.Flags().StringVarP(&moveRecord, "record", "r", "", "Record ID to move (required)")
	moveCmd.Flags().StringVarP(&moveList, "list", "l", "", "Destination list ID (required)")
	moveCmd.Flags().StringVarP(&moveWorkspace, "workspace", "w", "", "Source workspace ID (required)")
	moveCmd.Flags().BoolVarP(&moveSimple, "simple", "s", false, "Simple output format")
}

func runMove(cmd *cobra.Command, args []string) error {
	if moveRecord == "" || moveList == "" || moveWorkspace == "" {
		return fmt.Errorf("--record, --list, and --workspace flags are required")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := common.NewClient(config)
	client.SetProject(moveWorkspace)

	mutation := fmt.Sprintf(`
		mutation UpdateTodos {
			updateTodos(input: {
				todoListId: "%s"
				filter: {
					todoIds: ["%s"]
				}
			})
		}
	`, moveList, moveRecord)

	var response MoveRecordResponse
	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		return fmt.Errorf("failed to move record: %w", err)
	}

	if !response.UpdateTodos {
		return fmt.Errorf("failed to move record: operation returned false")
	}

	if moveSimple {
		fmt.Printf("Moved record %s to list %s\n", moveRecord, moveList)
	} else {
		fmt.Printf("=== Record Moved Successfully ===\n")
		fmt.Printf("Record ID: %s\n", moveRecord)
		fmt.Printf("Destination List ID: %s\n", moveList)
	}

	return nil
}
