package tools

import (
	"flag"
	"fmt"
	
	. "demo-builder/common"
)

// Record and TodoList are already defined in common/types.go
// Using TodoListSimple for this use case
type ProjectTodoList struct {
	ID     string   `json:"id"`
	Title  string   `json:"title"`
	Todos  []Record `json:"todos"`
}

type ProjectResponse struct {
	TodoLists []ProjectTodoList `json:"todoLists"`
}

func RunReadProjectRecords(args []string) error {
	fs := flag.NewFlagSet("read-project-records", flag.ExitOnError)
	projectID := fs.String("project", "", "Project ID (required)")
	fs.Parse(args)

	if *projectID == "" {
		return fmt.Errorf("project ID is required. Use -project flag")
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Create client
	client := NewClient(config)

	// First get all lists
	listQuery := `
		query GetProjectLists($projectId: String!) {
			todoLists(projectId: $projectId) {
				id
				title
				todosCount
			}
		}
	`

	variables := map[string]interface{}{
		"projectId": *projectID,
	}

	fmt.Printf("=== Records in Project %s ===\n\n", *projectID)

	type ListResponse struct {
		TodoLists []struct {
			ID         string `json:"id"`
			Title      string `json:"title"`
			TodosCount int    `json:"todosCount"`
		} `json:"todoLists"`
	}

	var listResponse ListResponse
	if err := client.ExecuteQueryWithResult(listQuery, variables, &listResponse); err != nil {
		return fmt.Errorf("failed to query lists: %v", err)
	}

	totalRecords := 0

	// For each list, get its todos using the todoList.todos field
	for _, list := range listResponse.TodoLists {
		fmt.Printf("📋 **%s** (%d records)\n", list.Title, list.TodosCount)
		
		if list.TodosCount == 0 {
			fmt.Printf("   (No records)\n\n")
		} else {
			// Query todos using the todoList field structure from schema
			recordQuery := `
				query GetListRecords($listId: String!) {
					todoList(id: $listId) {
						todos(first: 50) {
							id
							uid
							title
							position
							done
							createdAt
							updatedAt
						}
					}
				}
			`

			recordVariables := map[string]interface{}{
				"listId": list.ID,
			}

			var recordResponse struct {
				TodoList struct {
					Todos []Record `json:"todos"`
				} `json:"todoList"`
			}

			if err := client.ExecuteQueryWithResult(recordQuery, recordVariables, &recordResponse); err != nil {
				fmt.Printf("   Error fetching todos: %v\n\n", err)
				continue
			}

			for i, record := range recordResponse.TodoList.Todos {
				status := "⭕"
				if record.Done {
					status = "✅"
				}
				fmt.Printf("   %d. %s %s\n", i+1, status, record.Title)
				fmt.Printf("      ID: %s\n", record.ID)
			}
			fmt.Println()
		}
		totalRecords += list.TodosCount
	}

	fmt.Printf("Total records across all lists: %d\n", totalRecords)

	return nil
}