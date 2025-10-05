package tools

import (
	"flag"
	"fmt"

	"demo-builder/common"
)

// MoveRecordInput represents the input for moving a record
type MoveRecordInput struct {
	TodoID     string `json:"todoId"`
	TodoListID string `json:"todoListId"`
}

// MoveRecordResponse represents the response from the move operation
type MoveRecordResponse struct {
	EditTodo struct {
		ID       string `json:"id"`
		Title    string `json:"title"`
		Position float64 `json:"position"`
		TodoList struct {
			ID    string `json:"id"`
			Title string `json:"title"`
			Project struct {
				ID   string `json:"id"`
				Name string `json:"name"`
			} `json:"project"`
		} `json:"todoList"`
	} `json:"editTodo"`
}

func RunMoveRecord(args []string) error {
	fs := flag.NewFlagSet("move-record", flag.ExitOnError)

	// Required flags
	recordID := fs.String("record", "", "Record ID to move (required)")
	listID := fs.String("list", "", "Destination list ID (required)")

	// Optional flags
	simple := fs.Bool("simple", false, "Simple output format")

	fs.Parse(args)

	if *recordID == "" || *listID == "" {
		fmt.Println("Error: Both -record and -list flags are required")
		fmt.Println("\nUsage:")
		fmt.Println("  go run . move-record -record RECORD_ID -list LIST_ID [flags]")
		fmt.Println("\nFlags:")
		fs.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  # Move record to different list (same or different project)")
		fmt.Println("  go run . move-record -record rec_123456 -list list_789012")
		fmt.Println("")
		fmt.Println("  # Move record with simple output")
		fmt.Println("  go run . move-record -record rec_123456 -list list_789012 -simple")
		fmt.Println("")
		fmt.Println("Note: The record will be moved to the specified list, which can be in")
		fmt.Println("      the same project or a different project. The system automatically")
		fmt.Println("      handles cross-project moves.")
		return fmt.Errorf("required flags missing")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	client := common.NewClient(config)

	// Execute the move operation using editTodo mutation
	input := MoveRecordInput{
		TodoID:     *recordID,
		TodoListID: *listID,
	}

	response, err := executeMoveRecord(client, input)
	if err != nil {
		return fmt.Errorf("failed to move record: %v", err)
	}

	record := response.EditTodo

	if *simple {
		fmt.Printf("Moved record: %s (ID: %s) to list: %s (%s)\n",
			record.Title, record.ID, record.TodoList.Title, record.TodoList.ID)
		if record.TodoList.Project.Name != "" {
			fmt.Printf("Project: %s (%s)\n", record.TodoList.Project.Name, record.TodoList.Project.ID)
		}
	} else {
		fmt.Printf("=== Record Moved Successfully ===\n")
		fmt.Printf("Record ID: %s\n", record.ID)
		fmt.Printf("Title: %s\n", record.Title)
		fmt.Printf("Position: %.0f\n", record.Position)
		fmt.Printf("\n=== Destination ===\n")
		fmt.Printf("List: %s (%s)\n", record.TodoList.Title, record.TodoList.ID)
		if record.TodoList.Project.Name != "" {
			fmt.Printf("Project: %s (%s)\n", record.TodoList.Project.Name, record.TodoList.Project.ID)
		}
	}

	return nil
}

// executeMoveRecord performs the move operation using the editTodo mutation
func executeMoveRecord(client *common.Client, input MoveRecordInput) (*MoveRecordResponse, error) {
	mutation := fmt.Sprintf(`
		mutation EditTodo {
			editTodo(input: {
				todoId: "%s"
				todoListId: "%s"
			}) {
				id
				title
				position
				todoList {
					id
					title
					project {
						id
						name
					}
				}
			}
		}
	`, input.TodoID, input.TodoListID)

	var response MoveRecordResponse
	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		return nil, err
	}

	return &response, nil
}