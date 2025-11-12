package tools

import (
	"flag"
	"fmt"

	"cli/common"
)

// RunDeleteRecord deletes a record/todo by ID
func RunDeleteRecord(args []string) error {
	fs := flag.NewFlagSet("delete-record", flag.ExitOnError)

	var recordID string
	var confirm bool

	fs.StringVar(&recordID, "record", "", "Record/Todo ID to delete")
	fs.BoolVar(&confirm, "confirm", false, "Confirm deletion (required for safety)")

	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("error parsing flags: %v", err)
	}

	if recordID == "" {
		fmt.Println("Error: -record flag is required")
		fmt.Println("Usage: go run main.go delete-record -record RECORD_ID -confirm")
		return fmt.Errorf("record ID is required")
	}

	if !confirm {
		fmt.Println("Error: -confirm flag is required for safety")
		fmt.Println("This will permanently delete the record/todo.")
		fmt.Println("Usage: go run main.go delete-record -record RECORD_ID -confirm")
		return fmt.Errorf("confirmation required for deletion")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("error loading config: %v", err)
	}

	client := common.NewClient(config)

	mutation := `
		mutation DeleteTodo($input: DeleteTodoInput!) {
			deleteTodo(input: $input) {
				success
				operationId
			}
		}
	`

	variables := map[string]interface{}{
		"input": common.DeleteTodoInput{
			TodoID: recordID,
		},
	}

	data, err := client.ExecuteQuery(mutation, variables)
	if err != nil {
		return fmt.Errorf("error deleting record: %v", err)
	}

	// Extract the deleteTodo result
	if deleteTodoData, ok := data["deleteTodo"].(map[string]interface{}); ok {
		success, hasSuccess := deleteTodoData["success"].(bool)
		operationID, _ := deleteTodoData["operationId"].(string)

		if hasSuccess && success {
			fmt.Printf("Record %s deleted successfully\n", recordID)
			if operationID != "" {
				fmt.Printf("Operation ID: %s\n", operationID)
			}
			return nil
		} else {
			return fmt.Errorf("failed to delete record %s (success: %v)", recordID, success)
		}
	} else {
		return fmt.Errorf("unexpected response format: %+v", data)
	}
}
