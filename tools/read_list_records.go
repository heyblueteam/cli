package tools

import (
	"flag"
	"fmt"
	
	"demo-builder/common"
)

// User, Tag, TodoList, and Record are already defined in common/types.go
// Using Record instead of Todo for consistency

type TodoListWithRecords struct {
	ID               string   `json:"id"`
	UID              string   `json:"uid"`
	Title            string   `json:"title"`
	Position         float64  `json:"position"`
	TodosCount       int      `json:"todosCount"`
	TodosMaxPosition float64  `json:"todosMaxPosition"`
	CreatedAt        string   `json:"createdAt"`
	UpdatedAt        string   `json:"updatedAt"`
	IsDisabled       bool     `json:"isDisabled"`
	IsLocked         bool     `json:"isLocked"`
	Completed        bool     `json:"completed"`
	Editable         bool     `json:"editable"`
	Deletable        bool     `json:"deletable"`
	Todos            []common.Record `json:"todos"`
}

// TodoListResponse represents the response from the GraphQL query
type TodoListResponse struct {
	TodoList TodoListWithRecords `json:"todoList"`
}

func RunReadTodos(args []string) error {
	fs := flag.NewFlagSet("read-todos", flag.ExitOnError)
	todoListID := fs.String("list", "", "Todo List ID (required)")
	search := fs.String("search", "", "Search todos by title or description")
	assigneeID := fs.String("assignee", "", "Filter by assignee ID")
	tagIDs := fs.String("tags", "", "Filter by tag IDs (comma-separated)")
	done := fs.String("done", "", "Filter by completion status (true/false)")
	orderBy := fs.String("order", "position_ASC", "Order by field (position_ASC, position_DESC, title_ASC, title_DESC, createdAt_ASC, createdAt_DESC, duedAt_ASC, duedAt_DESC)")
	limit := fs.Int("limit", 50, "Maximum number of todos to return")
	simple := fs.Bool("simple", false, "Show only basic todo information")
	fs.Parse(args)

	// Validate required parameters
	if *todoListID == "" {
		return fmt.Errorf("todo List ID is required. Use -list flag")
	}

	// Load configuration
	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Create client
	client := common.NewClient(config)

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
		return fmt.Errorf("failed to execute query: %v", err)
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
		return nil
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
				fmt.Printf("   Description: %s\n", common.TruncateString(todo.Text, 100))
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

	return nil
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
func getTodoStatus(todo common.Record) string {
	if todo.Archived {
		return "Archived"
	}
	if todo.Done {
		return "Completed"
	}
	return "Active"
}

// getListStatus returns a human-readable status for a todo list
func getListStatus(list TodoListWithRecords) string {
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

// TruncateString is now in common package