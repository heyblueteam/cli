package tools

import (
	"flag"
	"fmt"

	. "demo-builder/common"
)

// CreateTodoDependencyInput represents the input for creating a dependency
type CreateTodoDependencyInput struct {
	Type        string `json:"type"`
	TodoID      string `json:"todoId"`
	OtherTodoID string `json:"otherTodoId"`
}

// DependencyTodo represents a minimal todo in dependency responses
type DependencyTodo struct {
	ID    string `json:"id"`
	UID   string `json:"uid"`
	Title string `json:"title"`
}

// CreateDependencyResponse represents the response from creating a dependency
type CreateDependencyResponse struct {
	CreateTodoDependency struct {
		ID        string           `json:"id"`
		UID       string           `json:"uid"`
		Title     string           `json:"title"`
		DependOn  []DependencyTodo `json:"dependOn"`
		DependBy  []DependencyTodo `json:"dependBy"`
	} `json:"createTodoDependency"`
}

func RunCreateDependency(args []string) error {
	fs := flag.NewFlagSet("create-dependency", flag.ExitOnError)
	todoID := fs.String("record", "", "Record/Todo ID (required)")
	otherTodoID := fs.String("other-record", "", "Other Record/Todo ID to create dependency with (required)")
	depType := fs.String("type", "BLOCKED_BY", "Dependency type: BLOCKING or BLOCKED_BY (default: BLOCKED_BY)")
	projectID := fs.String("project", "", "Project ID or slug (optional - for context)")
	simple := fs.Bool("simple", false, "Show simple output")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *todoID == "" {
		return fmt.Errorf("record ID is required (-record)")
	}
	if *otherTodoID == "" {
		return fmt.Errorf("other record ID is required (-other-record)")
	}
	if *depType != "BLOCKING" && *depType != "BLOCKED_BY" {
		return fmt.Errorf("type must be BLOCKING or BLOCKED_BY")
	}

	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	client := NewClient(config)
	if *projectID != "" {
		client.SetProject(*projectID)
	}

	mutation := `
		mutation CreateTodoDependency($input: CreateTodoDependencyInput!) {
			createTodoDependency(input: $input) {
				id
				uid
				title
				dependOn {
					id
					uid
					title
				}
				dependBy {
					id
					uid
					title
				}
			}
		}
	`

	variables := map[string]any{
		"input": CreateTodoDependencyInput{
			Type:        *depType,
			TodoID:      *todoID,
			OtherTodoID: *otherTodoID,
		},
	}

	var response CreateDependencyResponse
	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to create dependency: %v", err)
	}

	todo := response.CreateTodoDependency

	if *simple {
		fmt.Printf("✅ Dependency created: %s %s %s\n", *todoID, *depType, *otherTodoID)
	} else {
		fmt.Printf("=== Dependency Created Successfully ===\n")
		fmt.Printf("Record: %s (%s)\n", todo.Title, todo.ID)
		fmt.Printf("Type: %s\n", *depType)
		fmt.Printf("Other Record: %s\n", *otherTodoID)
		fmt.Println()

		if len(todo.DependOn) > 0 {
			fmt.Printf("Depends On (%d):\n", len(todo.DependOn))
			for _, dep := range todo.DependOn {
				fmt.Printf("  → %s (%s)\n", dep.Title, dep.ID)
			}
		}
		if len(todo.DependBy) > 0 {
			fmt.Printf("Blocked By (%d):\n", len(todo.DependBy))
			for _, dep := range todo.DependBy {
				fmt.Printf("  ← %s (%s)\n", dep.Title, dep.ID)
			}
		}
		fmt.Printf("\n✅ Dependency created successfully!\n")
	}

	return nil
}
