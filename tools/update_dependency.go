package tools

import (
	"flag"
	"fmt"

	. "demo-builder/common"
)

// UpdateTodoDependencyInput represents the input for updating a dependency
type UpdateTodoDependencyInput struct {
	Type        string `json:"type"`
	TodoID      string `json:"todoId"`
	OtherTodoID string `json:"otherTodoId"`
}

// UpdateDependencyResponse represents the response from updating a dependency
type UpdateDependencyResponse struct {
	UpdateTodoDependency struct {
		ID        string           `json:"id"`
		UID       string           `json:"uid"`
		Title     string           `json:"title"`
		DependOn  []DependencyTodo `json:"dependOn"`
		DependBy  []DependencyTodo `json:"dependBy"`
	} `json:"updateTodoDependency"`
}

func RunUpdateDependency(args []string) error {
	fs := flag.NewFlagSet("update-dependency", flag.ExitOnError)
	todoID := fs.String("record", "", "Record/Todo ID (required)")
	otherTodoID := fs.String("other-record", "", "Other Record/Todo ID (required)")
	depType := fs.String("type", "", "New dependency type: BLOCKING or BLOCKED_BY (required)")
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
	if *depType == "" {
		return fmt.Errorf("dependency type is required (-type)")
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
		mutation UpdateTodoDependency($input: UpdateTodoDependencyInput!) {
			updateTodoDependency(input: $input) {
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
		"input": UpdateTodoDependencyInput{
			Type:        *depType,
			TodoID:      *todoID,
			OtherTodoID: *otherTodoID,
		},
	}

	var response UpdateDependencyResponse
	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to update dependency: %v", err)
	}

	todo := response.UpdateTodoDependency

	if *simple {
		fmt.Printf("✅ Dependency updated: %s %s %s\n", *todoID, *depType, *otherTodoID)
	} else {
		fmt.Printf("=== Dependency Updated Successfully ===\n")
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
		fmt.Printf("\n✅ Dependency updated successfully!\n")
	}

	return nil
}
