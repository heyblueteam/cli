package lists

import (
	"fmt"
	"strconv"
	"strings"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

type UpdateTodoListResponse struct {
	EditTodoList struct {
		ID       string  `json:"id"`
		UID      string  `json:"uid"`
		Title    string  `json:"title"`
		Position float64 `json:"position"`
		IsLocked bool    `json:"isLocked"`
	} `json:"editTodoList"`
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a list",
	Long:  "Update list properties like title, position, and lock status.",
	Example: `  blue lists update --list <id> --title "New Title"
  blue lists update --list <id> --locked true
  blue lists update --list <id> --position 1000.0 --workspace <id>`,
	RunE: runUpdate,
}

var (
	updateList      string
	updateWorkspace string
	updateTitle     string
	updatePosition  string
	updateLocked    string
	updateSimple    bool
)

func init() {
	updateCmd.Flags().StringVarP(&updateList, "list", "l", "", "List ID (required)")
	updateCmd.Flags().StringVarP(&updateWorkspace, "workspace", "w", "", "Workspace ID (optional, for context)")
	updateCmd.Flags().StringVarP(&updateTitle, "title", "t", "", "New title for the list")
	updateCmd.Flags().StringVar(&updatePosition, "position", "", "New position for the list (float)")
	updateCmd.Flags().StringVar(&updateLocked, "locked", "", "Lock status (true/false)")
	updateCmd.Flags().BoolVarP(&updateSimple, "simple", "s", false, "Simple output format")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if updateList == "" {
		return fmt.Errorf("list ID is required. Use --list flag")
	}

	if updateTitle == "" && updatePosition == "" && updateLocked == "" {
		return fmt.Errorf("at least one field must be specified for update (--title, --position, or --locked)")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)

	if updateWorkspace != "" {
		client.SetProject(updateWorkspace)
	}

	var fields []string

	if updateTitle != "" {
		fields = append(fields, fmt.Sprintf("title: \"%s\"", updateTitle))
	}

	if updatePosition != "" {
		position, err := strconv.ParseFloat(updatePosition, 64)
		if err != nil {
			return fmt.Errorf("invalid position value '%s': %w", updatePosition, err)
		}
		fields = append(fields, fmt.Sprintf("position: %g", position))
	}

	if updateLocked != "" {
		locked, err := strconv.ParseBool(updateLocked)
		if err != nil {
			return fmt.Errorf("invalid locked value '%s': %w", updateLocked, err)
		}
		fields = append(fields, fmt.Sprintf("isLocked: %t", locked))
	}

	mutation := fmt.Sprintf(`
		mutation EditTodoList {
			editTodoList(input: {
				todoListId: "%s"
				%s
			}) {
				id
				uid
				title
				position
				isLocked
			}
		}`, updateList, strings.Join(fields, "\n\t\t\t\t"))

	var response UpdateTodoListResponse
	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		return fmt.Errorf("failed to update list: %w", err)
	}

	list := response.EditTodoList
	if updateSimple {
		fmt.Printf("List updated: %s (ID: %s)\n", list.Title, list.ID)
	} else {
		fmt.Printf("=== List Updated Successfully ===\n")
		fmt.Printf("ID: %s\n", list.ID)
		fmt.Printf("UID: %s\n", list.UID)
		fmt.Printf("Title: %s\n", list.Title)
		fmt.Printf("Position: %.0f\n", list.Position)
		fmt.Printf("Is Locked: %t\n", list.IsLocked)
	}

	return nil
}
