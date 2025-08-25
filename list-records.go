package main

import (
	"flag"
	"fmt"
	"log"
	"strings"
)

// Record represents a todo/record item in the system
type Record struct {
	ID                      string   `json:"id"`
	UID                     string   `json:"uid"`
	Position                float64  `json:"position"`
	Title                   string   `json:"title"`
	Text                    string   `json:"text"`
	HTML                    string   `json:"html"`
	StartedAt               string   `json:"startedAt"`
	DuedAt                  string   `json:"duedAt"`
	Timezone                string   `json:"timezone"`
	Color                   string   `json:"color"`
	Cover                   string   `json:"cover"`
	CoverLocked             bool     `json:"coverLocked"`
	Archived                bool     `json:"archived"`
	Done                    bool     `json:"done"`
	CommentCount            int      `json:"commentCount"`
	ChecklistCount          int      `json:"checklistCount"`
	ChecklistCompletedCount int      `json:"checklistCompletedCount"`
	IsRepeating             bool     `json:"isRepeating"`
	IsRead                  bool     `json:"isRead"`
	IsSeen                  bool     `json:"isSeen"`
	CreatedAt               string   `json:"createdAt"`
	UpdatedAt               string   `json:"updatedAt"`
	Users                   []User   `json:"users"`
	Tags                    []Tag    `json:"tags"`
	TodoList                TodoListInfo `json:"todoList"`
}

// User represents a user assigned to a record
type User struct {
	ID        string `json:"id"`
	UID       string `json:"uid"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	FullName  string `json:"fullName"`
	Email     string `json:"email"`
}

// Tag represents a tag associated with a record
type Tag struct {
	ID    string `json:"id"`
	UID   string `json:"uid"`
	Title string `json:"title"`
	Color string `json:"color"`
}

// TodoListInfo represents basic info about the todo list containing the record
type TodoListInfo struct {
	ID    string `json:"id"`
	UID   string `json:"uid"`
	Title string `json:"title"`
}

// RecordsResult represents the paginated response from the GraphQL query
type RecordsResult struct {
	Items    []Record `json:"items"`
	PageInfo PageInfo `json:"pageInfo"`
}

// PageInfo represents pagination information
type PageInfo struct {
	HasNextPage     bool   `json:"hasNextPage"`
	HasPreviousPage bool   `json:"hasPreviousPage"`
	StartCursor     string `json:"startCursor"`
	EndCursor       string `json:"endCursor"`
}

// RecordsResponse represents the response from the GraphQL query
type RecordsResponse struct {
	TodoQueries struct {
		Todos RecordsResult `json:"todos"`
	} `json:"todoQueries"`
}

func main() {
	// Parse command line flags
	projectID := flag.String("project", "", "Project ID to filter records")
	todoListID := flag.String("list", "", "Todo List ID to filter records")
	assigneeID := flag.String("assignee", "", "Filter by assignee ID")
	tagIDs := flag.String("tags", "", "Filter by tag IDs (comma-separated)")
	done := flag.String("done", "", "Filter by completion status (true/false)")
	archived := flag.String("archived", "", "Filter by archived status (true/false)")
	orderBy := flag.String("order", "updatedAt_DESC", "Order by field (position_ASC, position_DESC, title_ASC, title_DESC, createdAt_ASC, createdAt_DESC, updatedAt_ASC, updatedAt_DESC, duedAt_ASC, duedAt_DESC)")
	limit := flag.Int("limit", 20, "Maximum number of records to return")
	skip := flag.Int("skip", 0, "Number of records to skip (for pagination)")
	simple := flag.Bool("simple", false, "Show only basic record information")
	flag.Parse()

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Create client
	client := NewClient(config)

	// Build the GraphQL query
	query := buildRecordsQuery(*simple)

	// Build filter variables - TodosFilter requires companyIds and uses different field names
	filter := make(map[string]interface{})
	
	// companyIds is required for TodosFilter - get from config or leave empty to match all companies
	filter["companyIds"] = []string{} // This will match all companies the user has access to
	
	if *projectID != "" {
		filter["projectIds"] = []string{*projectID}
	}
	if *todoListID != "" {
		filter["todoListIds"] = []string{*todoListID}
	}
	if *assigneeID != "" {
		filter["assigneeIds"] = []string{*assigneeID}
	}
	if *tagIDs != "" {
		// Parse comma-separated tag IDs
		tagList := strings.Split(*tagIDs, ",")
		for i, tag := range tagList {
			tagList[i] = strings.TrimSpace(tag)
		}
		filter["tagIds"] = tagList
	}
	if *done != "" {
		if *done == "true" {
			filter["done"] = true
		} else if *done == "false" {
			filter["done"] = false
		}
	}
	if *archived != "" {
		if *archived == "true" {
			filter["archived"] = true
		} else if *archived == "false" {
			filter["archived"] = false
		}
	}

	// Build sort array based on orderBy string
	var sort []string
	if *orderBy != "" {
		// Convert from TodoOrderByInput format to TodosSort format
		switch *orderBy {
		case "position_ASC":
			sort = append(sort, "position_ASC")
		case "position_DESC":
			sort = append(sort, "position_DESC")
		case "title_ASC":
			sort = append(sort, "title_ASC")
		case "title_DESC":
			sort = append(sort, "title_DESC")
		case "createdAt_ASC":
			sort = append(sort, "createdAt_ASC")
		case "createdAt_DESC":
			sort = append(sort, "createdAt_DESC")
		case "updatedAt_ASC":
			// No updatedAt in TodosSort, use createdAt instead
			sort = append(sort, "createdAt_ASC")
		case "updatedAt_DESC":
			// No updatedAt in TodosSort, use createdAt instead
			sort = append(sort, "createdAt_DESC")
		case "duedAt_ASC":
			sort = append(sort, "duedAt_ASC")
		case "duedAt_DESC":
			sort = append(sort, "duedAt_DESC")
		default:
			sort = append(sort, "createdAt_DESC")
		}
	}

	// Build variables
	variables := map[string]interface{}{
		"filter": filter,
		"sort":   sort,
		"limit":  *limit,
		"skip":   *skip,
	}

	// Execute query
	var response RecordsResponse
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		log.Fatalf("Failed to execute query: %v", err)
	}

	// Display results
	result := response.TodoQueries.Todos
	fmt.Printf("\n=== Records Query Results ===\n")
	if *projectID != "" {
		fmt.Printf("Project ID: %s\n", *projectID)
	}
	if *todoListID != "" {
		fmt.Printf("List ID: %s\n", *todoListID)
	}
	if *assigneeID != "" {
		fmt.Printf("Assignee ID: %s\n", *assigneeID)
	}
	if *tagIDs != "" {
		fmt.Printf("Tag IDs: %s\n", *tagIDs)
	}
	fmt.Printf("Showing: %d records (skip: %d, limit: %d)\n", len(result.Items), *skip, *limit)
	fmt.Printf("Has next page: %t\n", result.PageInfo.HasNextPage)
	fmt.Printf("Has previous page: %t\n", result.PageInfo.HasPreviousPage)
	fmt.Println()

	if len(result.Items) == 0 {
		fmt.Println("No records found matching the criteria.")
		return
	}

	// Display records
	for i, record := range result.Items {
		recordNum := *skip + i + 1
		if *simple {
			// Simple output
			fmt.Printf("%d. %s\n", recordNum, record.Title)
			fmt.Printf("   ID: %s\n", record.ID)
			fmt.Printf("   List: %s\n", record.TodoList.Title)
			fmt.Printf("   Status: %s\n", getRecordStatus(record))
			if record.DuedAt != "" {
				fmt.Printf("   Due: %s\n", record.DuedAt)
			}
			fmt.Println()
		} else {
			// Detailed output
			fmt.Printf("%d. %s\n", recordNum, record.Title)
			fmt.Printf("   ID: %s\n", record.ID)
			fmt.Printf("   UID: %s\n", record.UID)
			fmt.Printf("   List: %s (%s)\n", record.TodoList.Title, record.TodoList.ID)
			fmt.Printf("   Position: %.0f\n", record.Position)
			fmt.Printf("   Status: %s\n", getRecordStatus(record))
			
			if record.Text != "" {
				fmt.Printf("   Description: %s\n", truncateString(record.Text, 100))
			}
			if record.StartedAt != "" {
				fmt.Printf("   Started: %s\n", record.StartedAt)
			}
			if record.DuedAt != "" {
				fmt.Printf("   Due: %s\n", record.DuedAt)
			}
			if record.Color != "" {
				fmt.Printf("   Color: %s\n", record.Color)
			}
			if record.Cover != "" {
				fmt.Printf("   Has cover: Yes\n")
			}
			fmt.Printf("   Comments: %d\n", record.CommentCount)
			fmt.Printf("   Checklists: %d/%d completed\n", record.ChecklistCompletedCount, record.ChecklistCount)
			if record.IsRepeating {
				fmt.Printf("   Repeating: Yes\n")
			}
			
			// Display assignees
			if len(record.Users) > 0 {
				fmt.Printf("   Assignees: ")
				for j, user := range record.Users {
					if j > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s", user.FullName)
				}
				fmt.Println()
			}

			// Display tags
			if len(record.Tags) > 0 {
				fmt.Printf("   Tags: ")
				for j, tag := range record.Tags {
					if j > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s", tag.Title)
				}
				fmt.Println()
			}

			fmt.Printf("   Created: %s\n", record.CreatedAt)
			fmt.Printf("   Updated: %s\n", record.UpdatedAt)
			fmt.Println()
		}
	}

	// Display pagination info
	if result.PageInfo.HasNextPage {
		nextSkip := *skip + *limit
		fmt.Printf("To see more records, use: -skip %d\n", nextSkip)
	}
}

// buildRecordsQuery builds the GraphQL query based on the detail level
func buildRecordsQuery(simple bool) string {
	if simple {
		return `
			query GetRecords($filter: TodosFilter!, $sort: [TodosSort!], $limit: Int, $skip: Int) {
				todoQueries {
					todos(filter: $filter, sort: $sort, limit: $limit, skip: $skip) {
						items {
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
							todoList {
								id
								uid
								title
							}
						}
						pageInfo {
							hasNextPage
							hasPreviousPage
							startCursor
							endCursor
						}
					}
				}
			}
		`
	}

	return `
		query GetRecords($filter: TodosFilter!, $sort: [TodosSort!], $limit: Int, $skip: Int) {
			todoQueries {
				todos(filter: $filter, sort: $sort, limit: $limit, skip: $skip) {
					items {
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
						todoList {
							id
							uid
							title
						}
					}
					pageInfo {
						hasNextPage
						hasPreviousPage
						startCursor
						endCursor
					}
				}
			}
		}
	`
}

// getRecordStatus returns a human-readable status for a record
func getRecordStatus(record Record) string {
	if record.Archived {
		return "Archived"
	}
	if record.Done {
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