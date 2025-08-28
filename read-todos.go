package main

import (
	"flag"
	"fmt"
	"log"
)

// Todo represents a todo item in the system
type Todo struct {
	ID                    string   `json:"id"`
	UID                   string   `json:"uid"`
	Position              float64  `json:"position"`
	Title                 string   `json:"title"`
	Text                  string   `json:"text"`
	HTML                  string   `json:"html"`
	StartedAt             string   `json:"startedAt"`
	DuedAt                string   `json:"duedAt"`
	Timezone              string   `json:"timezone"`
	Color                 string   `json:"color"`
	Cover                 string   `json:"cover"`
	CoverLocked           bool     `json:"coverLocked"`
	Archived              bool     `json:"archived"`
	Done                  bool     `json:"done"`
	CommentCount          int      `json:"commentCount"`
	ChecklistCount        int      `json:"checklistCount"`
	ChecklistCompletedCount int    `json:"checklistCompletedCount"`
	IsRepeating           bool     `json:"isRepeating"`
	IsRead                bool     `json:"isRead"`
	IsSeen                bool     `json:"isSeen"`
	CreatedAt             string   `json:"createdAt"`
	UpdatedAt             string   `json:"updatedAt"`
	Users                 []User   `json:"users"`
	Tags                  []Tag    `json:"tags"`
}

// User represents a user assigned to a todo
type User struct {
	ID        string `json:"id"`
	UID       string `json:"uid"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	FullName  string `json:"fullName"`
	Email     string `json:"email"`
}

// Tag represents a tag associated with a todo
type Tag struct {
	ID    string `json:"id"`
	UID   string `json:"uid"`
	Title string `json:"title"`
	Color string `json:"color"`
}

// TodoList represents a todo list with its todos
type TodoList struct {
	ID               string  `json:"id"`
	UID              string  `json:"uid"`
	Title            string  `json:"title"`
	Position         float64 `json:"position"`
	TodosCount       int     `json:"todosCount"`
	TodosMaxPosition float64 `json:"todosMaxPosition"`
	CreatedAt        string  `json:"createdAt"`
	UpdatedAt        string  `json:"updatedAt"`
	IsDisabled       bool    `json:"isDisabled"`
	IsLocked         bool    `json:"isLocked"`
	Completed        bool    `json:"completed"`
	Editable         bool    `json:"editable"`
	Deletable        bool    `json:"deletable"`
	Todos            []Todo  `json:"todos"`
}

// TodoListResponse represents the response from the GraphQL query
type TodoListResponse struct {
	TodoList TodoList `json:"todoList"`
}

func main() {
	// Parse command line flags
	todoListID := flag.String("list", "", "Todo List ID (required)")
	search := flag.String("search", "", "Search todos by title or description")
	assigneeID := flag.String("assignee", "", "Filter by assignee ID")
	tagIDs := flag.String("tags", "", "Filter by tag IDs (comma-separated)")
	done := flag.String("done", "", "Filter by completion status (true/false)")
	orderBy := flag.String("order", "position_ASC", "Order by field (position_ASC, position_DESC, title_ASC, title_DESC, createdAt_ASC, createdAt_DESC, duedAt_ASC, duedAt_DESC)")
	limit := flag.Int("limit", 50, "Maximum number of todos to return")
	simple := flag.Bool("simple", false, "Show only basic todo information")
	flag.Parse()

	// Validate required parameters
	if *todoListID == "" {
		log.Fatal("Todo List ID is required. Use -list flag")
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create client
	client := NewClient(config)

	// Build the GraphQL query
	query := buildTodoListQuery(*simple)

	// Build variables
	variables := map[string]interface{}{
		"todoListId": *todoListID,
		"limit":      *limit,
		"orderBy":    *orderBy,
	}

	// Add optional filters
	if *search != "" {
		variables["search"] = *search
	}
	if *assigneeID != "" {
		variables["assigneeId"] = *assigneeID
	}
	if *tagIDs != "" {
		// Parse comma-separated tag IDs
		// Note: This would need proper parsing in a real implementation
		variables["tagIds"] = []string{*tagIDs} // Simplified for now
	}
	if *done != "" {
		if *done == "true" {
			variables["done"] = true
		} else if *done == "false" {
			variables["done"] = false
		}
	}

	// Execute query
	var response TodoListResponse
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}

	// Display results
	todoList := response.TodoList
	fmt.Printf("\n=== Todos in List: %s ===\n", todoList.Title)
	fmt.Printf("List ID: %s\n", todoList.ID)
	fmt.Printf("Total todos: %d\n", todoList.TodosCount)
	fmt.Printf("Max position: %.0f\n", todoList.TodosMaxPosition)
	fmt.Printf("List status: %s\n", getListStatus(todoList))
	fmt.Println()

	if len(todoList.Todos) == 0 {
		fmt.Println("No todos found in this list.")
		return
	}

	// Display todos
	for i, todo := range todoList.Todos {
		if *simple {
			// Simple output
			fmt.Printf("%d. %s\n", i+1, todo.Title)
			fmt.Printf("   ID: %s\n", todo.ID)
			fmt.Printf("   Position: %.0f\n", todo.Position)
			fmt.Printf("   Status: %s\n", getTodoStatus(todo))
			if todo.DuedAt != "" {
				fmt.Printf("   Due: %s\n", todo.DuedAt)
			}
			fmt.Println()
		} else {
			// Detailed output
			fmt.Printf("%d. %s\n", i+1, todo.Title)
			fmt.Printf("   ID: %s\n", todo.ID)
			fmt.Printf("   UID: %s\n", todo.UID)
			fmt.Printf("   Position: %.0f\n", todo.Position)
			fmt.Printf("   Status: %s\n", getTodoStatus(todo))
			if todo.Text != "" {
				fmt.Printf("   Description: %s\n", truncateString(todo.Text, 100))
			}
			if todo.StartedAt != "" {
				fmt.Printf("   Started: %s\n", todo.StartedAt)
			}
			if todo.DuedAt != "" {
				fmt.Printf("   Due: %s\n", todo.DuedAt)
			}
			if todo.Color != "" {
				fmt.Printf("   Color: %s\n", todo.Color)
			}
			if todo.Cover != "" {
				fmt.Printf("   Has cover: Yes\n")
			}
			fmt.Printf("   Comments: %d\n", todo.CommentCount)
			fmt.Printf("   Checklists: %d/%d completed\n", todo.ChecklistCompletedCount, todo.ChecklistCount)
			if todo.IsRepeating {
				fmt.Printf("   Repeating: Yes\n")
			}
			
			// Display assignees
			if len(todo.Users) > 0 {
				fmt.Printf("   Assignees: ")
				for j, user := range todo.Users {
					if j > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s", user.FullName)
				}
				fmt.Println()
			}

			// Display tags
			if len(todo.Tags) > 0 {
				fmt.Printf("   Tags: ")
				for j, tag := range todo.Tags {
					if j > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s", tag.Title)
				}
				fmt.Println()
			}

			fmt.Printf("   Created: %s\n", todo.CreatedAt)
			fmt.Printf("   Updated: %s\n", todo.UpdatedAt)
			fmt.Println()
		}
	}
}

// buildTodoListQuery builds the GraphQL query based on the detail level
func buildTodoListQuery(simple bool) string {
	if simple {
		return `
			query GetTodoList($todoListId: String!, $search: String, $assigneeId: String, $tagIds: [String!], $done: Boolean, $orderBy: TodoOrderByInput, $limit: Int) {
				todoList(id: $todoListId) {
					id
					uid
					title
					position
					todosCount
					todosMaxPosition
					createdAt
					updatedAt
					isDisabled
					isLocked
					completed
					editable
					deletable
					todos(
						search: $search
						assigneeId: $assigneeId
						tagIds: $tagIds
						done: $done
						orderBy: $orderBy
						first: $limit
					) {
						id
						uid
						position
						title
						duedAt
						done
						archived
						commentCount
						checklistCount
						checklistCompletedCount
						isRepeating
						createdAt
						updatedAt
						users {
							id
							uid
							firstName
							lastName
							fullName
							email
						}
						tags {
							id
							uid
							title
							color
						}
					}
				}
			}
		`
	}

	return `
		query GetTodoList($todoListId: String!, $search: String, $assigneeId: String, $tagIds: [String!], $done: Boolean, $orderBy: TodoOrderByInput, $limit: Int) {
			todoList(id: $todoListId) {
				id
				uid
				title
				position
				todosCount
				todosMaxPosition
				createdAt
				updatedAt
				isDisabled
				isLocked
				completed
				editable
				deletable
				todos(
					search: $search
					assigneeId: $assigneeId
					tagIds: $tagIds
					done: $done
					orderBy: $orderBy
					first: $limit
				) {
					id
					uid
					position
					title
					text
					html
					startedAt
					duedAt
					timezone
					color
					cover
					coverLocked
					archived
					done
					commentCount
					checklistCount
					checklistCompletedCount
					isRepeating
					isRead
					isSeen
					createdAt
					updatedAt
					users {
						id
						uid
						firstName
						lastName
						fullName
						email
					}
					tags {
						id
						uid
						title
						color
					}
				}
			}
		}
	`
}

// getTodoStatus returns a human-readable status for a todo
func getTodoStatus(todo Todo) string {
	if todo.Archived {
		return "Archived"
	}
	if todo.Done {
		return "Completed"
	}
	return "Active"
}

// getListStatus returns a human-readable status for a todo list
func getListStatus(list TodoList) string {
	if list.IsDisabled {
		return "Disabled"
	}
	if list.IsLocked {
		return "Locked"
	}
	if list.Completed {
		return "Completed"
	}
	return "Active"
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
