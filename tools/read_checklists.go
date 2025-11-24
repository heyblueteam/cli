package tools

import (
	"flag"
	"fmt"

	. "demo-builder/common"
)

// ChecklistItemWithUsers represents a checklist item with user assignments
type ChecklistItemWithUsers struct {
	ID        string  `json:"id"`
	UID       string  `json:"uid"`
	Title     string  `json:"title"`
	Position  float64 `json:"position"`
	Done      bool    `json:"done"`
	StartedAt *string `json:"startedAt"`
	DuedAt    *string `json:"duedAt"`
	CreatedAt string  `json:"createdAt"`
	UpdatedAt string  `json:"updatedAt"`
	CreatedBy User    `json:"createdBy"`
	Users     []User  `json:"users"`
}

// ChecklistWithItems represents a checklist with its items
type ChecklistWithItems struct {
	ID             string                   `json:"id"`
	UID            string                   `json:"uid"`
	Title          string                   `json:"title"`
	Position       float64                  `json:"position"`
	CreatedAt      string                   `json:"createdAt"`
	UpdatedAt      string                   `json:"updatedAt"`
	CreatedBy      User                     `json:"createdBy"`
	ChecklistItems []ChecklistItemWithUsers `json:"checklistItems"`
}

// ReadChecklistsResponse represents the response from querying a todo with checklists
type ReadChecklistsResponse struct {
	Todo struct {
		ID         string               `json:"id"`
		Title      string               `json:"title"`
		Checklists []ChecklistWithItems `json:"checklists"`
	} `json:"todo"`
}

// Execute GraphQL query to read checklists from a record
func executeReadChecklists(client *Client, recordID string) (*ReadChecklistsResponse, error) {
	// Build the query
	query := `
		query GetTodoChecklists($id: String!) {
			todo(id: $id) {
				id
				title
				checklists(orderBy: position_ASC) {
					id
					uid
					title
					position
					createdAt
					updatedAt
					createdBy {
						id
						uid
						fullName
						email
					}
					checklistItems(orderBy: position_ASC) {
						id
						uid
						title
						position
						done
						startedAt
						duedAt
						createdAt
						updatedAt
						createdBy {
							id
							uid
							fullName
							email
						}
						users {
							id
							uid
							fullName
							email
						}
					}
				}
			}
		}
	`

	// Prepare variables
	variables := map[string]interface{}{
		"id": recordID,
	}

	// Execute query
	var response ReadChecklistsResponse
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return nil, err
	}

	return &response, nil
}

// RunReadChecklists handles the read-checklists command
func RunReadChecklists(args []string) error {
	// Define flags
	fs := flag.NewFlagSet("read-checklists", flag.ExitOnError)
	recordID := fs.String("record", "", "Record/Todo ID to read checklists from (required)")
	projectID := fs.String("project", "", "Project ID or slug (optional - for context)")
	simple := fs.Bool("simple", false, "Show simple output")
	showItems := fs.Bool("items", true, "Show checklist items (default: true)")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Validate required flags
	if *recordID == "" {
		return fmt.Errorf("record ID is required")
	}

	// Load config and create client
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	client := NewClient(config)

	// Set project context if provided
	if *projectID != "" {
		client.SetProject(*projectID)
	}

	// Execute query
	response, err := executeReadChecklists(client, *recordID)
	if err != nil {
		return fmt.Errorf("failed to read checklists: %v", err)
	}

	// Display results
	if *simple {
		fmt.Printf("Record: %s (%s)\n", response.Todo.Title, response.Todo.ID)
		fmt.Printf("Checklists: %d\n\n", len(response.Todo.Checklists))
		for i, checklist := range response.Todo.Checklists {
			fmt.Printf("%d. %s (ID: %s)\n", i+1, checklist.Title, checklist.ID)
			if *showItems {
				for j, item := range checklist.ChecklistItems {
					status := "☐"
					if item.Done {
						status = "☑"
					}
					fmt.Printf("   %s %d.%d %s\n", status, i+1, j+1, item.Title)
				}
			}
		}
	} else {
		fmt.Printf("=== Checklists for Record: %s ===\n", response.Todo.Title)
		fmt.Printf("Record ID: %s\n", response.Todo.ID)
		fmt.Printf("Total Checklists: %d\n\n", len(response.Todo.Checklists))

		if len(response.Todo.Checklists) == 0 {
			fmt.Printf("No checklists found for this record.\n")
			return nil
		}

		for i, checklist := range response.Todo.Checklists {
			fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			fmt.Printf("📋 Checklist #%d: %s\n", i+1, checklist.Title)
			fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
			fmt.Printf("ID: %s\n", checklist.ID)
			fmt.Printf("UID: %s\n", checklist.UID)
			fmt.Printf("Position: %.1f\n", checklist.Position)
			// Calculate progress
			completedCount := 0
			for _, item := range checklist.ChecklistItems {
				if item.Done {
					completedCount++
				}
			}
			fmt.Printf("Progress: %d/%d completed\n", completedCount, len(checklist.ChecklistItems))
			fmt.Printf("Created: %s\n", checklist.CreatedAt)
			fmt.Printf("Updated: %s\n", checklist.UpdatedAt)
			fmt.Printf("Created By: %s (%s)\n", checklist.CreatedBy.FullName, checklist.CreatedBy.Email)

			if *showItems && len(checklist.ChecklistItems) > 0 {
				fmt.Printf("\n Items (%d):\n", len(checklist.ChecklistItems))
				for j, item := range checklist.ChecklistItems {
					status := "☐ Pending"
					if item.Done {
						status = "☑ Done"
					}
					fmt.Printf("\n   %d. %s %s\n", j+1, status, item.Title)
					fmt.Printf("      ID: %s\n", item.ID)
					fmt.Printf("      Position: %.1f\n", item.Position)
					if item.StartedAt != nil {
						fmt.Printf("      Started: %s\n", *item.StartedAt)
					}
					if item.DuedAt != nil {
						fmt.Printf("      Due: %s\n", *item.DuedAt)
					}
					if len(item.Users) > 0 {
						fmt.Printf("      Assigned to: ")
						for k, user := range item.Users {
							if k > 0 {
								fmt.Printf(", ")
							}
							fmt.Printf("%s", user.FullName)
						}
						fmt.Printf("\n")
					}
					fmt.Printf("      Created: %s by %s\n", item.CreatedAt, item.CreatedBy.FullName)
				}
			}
			fmt.Printf("\n")
		}
	}

	return nil
}
