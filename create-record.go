package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
)

type CreateTodoInput struct {
	TodoListID  string   `json:"todoListId"`
	Title       string   `json:"title"`
	Position    *float64 `json:"position,omitempty"`
	Description *string  `json:"description,omitempty"`
	Placement   *string  `json:"placement,omitempty"`
	AssigneeIDs []string `json:"assigneeIds,omitempty"`
}

type CreateTodoResponse struct {
	CreateTodo struct {
		ID       string  `json:"id"`
		Title    string  `json:"title"`
		Position float64 `json:"position"`
		TodoList struct {
			ID    string `json:"id"`
			Title string `json:"title"`
		} `json:"todoList"`
	} `json:"createTodo"`
}

func main() {
	var projectID = flag.String("project", "", "Project ID (required)")
	var listID = flag.String("list", "", "List ID to create the record in (required)")
	var title = flag.String("title", "", "Title of the record (required)")
	var description = flag.String("description", "", "Description of the record")
	var placement = flag.String("placement", "", "Placement in list: TOP or BOTTOM")
	var assignees = flag.String("assignees", "", "Comma-separated assignee IDs")
	var simple = flag.Bool("simple", false, "Simple output format")

	flag.Parse()

	if *projectID == "" || *listID == "" || *title == "" {
		fmt.Println("Error: -project, -list and -title flags are required")
		fmt.Println("\nUsage:")
		fmt.Println("  go run auth.go create-record.go -project PROJECT_ID -list LIST_ID -title \"Record Title\" [flags]")
		fmt.Println("\nFlags:")
		flag.PrintDefaults()
		return
	}

	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	client := NewClient(config)
	
	// Set project context from the provided flag
	client.SetProjectID(*projectID)

	input := CreateTodoInput{
		TodoListID: *listID,
		Title:      *title,
	}

	if *description != "" {
		input.Description = description
	}

	if *placement != "" {
		input.Placement = placement
	}

	if *assignees != "" {
		assigneeList := strings.Split(*assignees, ",")
		for i, assignee := range assigneeList {
			assigneeList[i] = strings.TrimSpace(assignee)
		}
		input.AssigneeIDs = assigneeList
	}

	// Build the mutation with optional fields
	var descriptionField string
	if input.Description != nil {
		descriptionField = fmt.Sprintf(`description: "%s"`, *input.Description)
	}
	
	var placementField string
	if input.Placement != nil {
		placementField = fmt.Sprintf(`placement: %s`, *input.Placement)
	}

	mutation := fmt.Sprintf(`
		mutation CreateTodo {
			createTodo(input: {
				todoListId: "%s"
				title: "%s"
				%s
				%s
			}) {
				id
				title
				position
				todoList {
					id
					title
				}
			}
		}
	`, input.TodoListID, input.Title, descriptionField, placementField)

	var response CreateTodoResponse
	if err := client.ExecuteQueryWithResult(mutation, nil, &response); err != nil {
		log.Fatalf("Request failed: %v", err)
	}

	record := response.CreateTodo

	if *simple {
		fmt.Printf("Created record: %s (ID: %s)\n", record.Title, record.ID)
	} else {
		fmt.Printf("=== Record Created Successfully ===\n")
		fmt.Printf("ID: %s\n", record.ID)
		fmt.Printf("Title: %s\n", record.Title)
		fmt.Printf("Position: %.0f\n", record.Position)
		fmt.Printf("List: %s (%s)\n", record.TodoList.Title, record.TodoList.ID)
	}
}