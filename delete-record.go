package main

import (
	"flag"
	"fmt"
	"log"
	"os"
)

type DeleteTodoInput struct {
	TodoID string `json:"todoId"`
}

type DeleteTodoResponse struct {
	Data struct {
		DeleteTodo struct {
			Success     bool   `json:"success"`
			OperationID string `json:"operationId"`
		} `json:"deleteTodo"`
	} `json:"data"`
}

func main() {
	var recordID string
	var confirm bool

	flag.StringVar(&recordID, "record", "", "Record/Todo ID to delete")
	flag.BoolVar(&confirm, "confirm", false, "Confirm deletion (required for safety)")
	flag.Parse()

	if recordID == "" {
		fmt.Println("Error: -record flag is required")
		fmt.Println("Usage: go run auth.go delete-record.go -record RECORD_ID -confirm")
		os.Exit(1)
	}

	if !confirm {
		fmt.Println("Error: -confirm flag is required for safety")
		fmt.Println("This will permanently delete the record/todo.")
		fmt.Println("Usage: go run auth.go delete-record.go -record RECORD_ID -confirm")
		os.Exit(1)
	}

	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Error loading config: %v", err)
	}

	client := NewClient(config)

	mutation := `
		mutation DeleteTodo($input: DeleteTodoInput!) {
			deleteTodo(input: $input) {
				success
				operationId
			}
		}
	`

	variables := map[string]interface{}{
		"input": DeleteTodoInput{
			TodoID: recordID,
		},
	}

	data, err := client.ExecuteQuery(mutation, variables)
	if err != nil {
		log.Fatalf("Error deleting record: %v", err)
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
		} else {
			fmt.Printf("Failed to delete record %s (success: %v)\n", recordID, success)
			os.Exit(1)
		}
	} else {
		fmt.Printf("Unexpected response format: %+v\n", data)
		os.Exit(1)
	}
}