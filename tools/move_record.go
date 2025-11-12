package tools

import (
	"flag"
	"fmt"

	"cli/common"
)

// MoveRecordInput represents the input for moving a record
type MoveRecordInput struct {
	TodoListID string   `json:"todoListId"`
	TodoIDs    []string `json:"todoIds"`
}

// MoveRecordResponse represents the response from the move operation
type MoveRecordResponse struct {
	UpdateTodos bool `json:"updateTodos"`
}

func RunMoveRecord(args []string) error {
	fs := flag.NewFlagSet("move-record", flag.ExitOnError)

	// Required flags
	recordID := fs.String("record", "", "Record ID to move (required)")
	listID := fs.String("list", "", "Destination list ID (required)")
	projectID := fs.String("project", "", "Source project ID where record currently exists (required)")

	// Optional flags
	simple := fs.Bool("simple", false, "Simple output format")

	fs.Parse(args)

	if *recordID == "" || *listID == "" || *projectID == "" {
		fmt.Println("Error: -record, -list, and -project flags are required")
		fmt.Println("\nUsage:")
		fmt.Println("  go run . move-record -record RECORD_ID -list LIST_ID -project PROJECT_ID [flags]")
		fmt.Println("\nFlags:")
		fs.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  # Move record to different list in same project")
		fmt.Println("  go run . move-record -record rec_123456 -list list_789012 -project proj_abc")
		fmt.Println("")
		fmt.Println("  # Move record to different project (cross-project move)")
		fmt.Println("  go run . move-record -record rec_123456 -list list_in_other_proj -project proj_abc -simple")
		fmt.Println("")
		fmt.Println("Note: -project should be the SOURCE project ID (where the record currently is).")
		fmt.Println("      Cross-project moves are handled automatically based on the destination list ID.")
		return fmt.Errorf("required flags missing")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	client := common.NewClient(config)

	// Set the source project as context for the updateTodos mutation
	client.SetProject(*projectID)

	// Execute the move operation using updateTodos mutation
	input := MoveRecordInput{
		TodoListID: *listID,
		TodoIDs:    []string{*recordID},
	}

	response, err := executeMoveRecord(client, input)
	if err != nil {
		return fmt.Errorf("failed to move record: %v", err)
	}

	if !response.UpdateTodos {
		return fmt.Errorf("failed to move record: updateTodos returned false")
	}

	if *simple {
		fmt.Printf("Moved record %s to list %s\n", *recordID, *listID)
	} else {
		fmt.Printf("=== Record Moved Successfully ===\n")
		fmt.Printf("Record ID: %s\n", *recordID)
		fmt.Printf("Destination List ID: %s\n", *listID)
	}

	return nil
}

// executeMoveRecord performs the move operation using the updateTodos mutation
func executeMoveRecord(client *common.Client, input MoveRecordInput) (*MoveRecordResponse, error) {
	mutation := fmt.Sprintf(`
		mutation UpdateTodos {
			updateTodos(input: {
				todoListId: "%s"
				filter: {
					todoIds: ["%s"]
				}
			})
		}
	`, input.TodoListID, input.TodoIDs[0])

	var response MoveRecordResponse
	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		return nil, err
	}

	return &response, nil
}