package tools

import (
	"flag"
	"fmt"
	
	"cli/common"
)

// CustomFieldValue represents a custom field value in record details
// RecordCustomFieldValue is now defined in common/types.go

// SimpleCustomFieldValue represents a custom field value with simplified structure  
type SimpleCustomFieldValue struct {
	ID    string      `json:"id"`
	Value interface{} `json:"value"`
}

// RecordCustomFieldInfo holds field metadata for display
type RecordCustomFieldInfo struct {
	ID   string
	Name string
	Type string
}

// DetailedRecord represents a record with all possible fields including custom fields
type DetailedRecord struct {
	ID                      string                     `json:"id"`
	UID                     string                     `json:"uid"`
	Position                float64                    `json:"position"`
	Title                   string                     `json:"title"`
	Text                    string                     `json:"text,omitempty"`
	HTML                    string                     `json:"html,omitempty"`
	StartedAt               string                     `json:"startedAt,omitempty"`
	DuedAt                  string                     `json:"duedAt,omitempty"`
	Timezone                string                     `json:"timezone,omitempty"`
	Color                   string                     `json:"color,omitempty"`
	Cover                   string                     `json:"cover,omitempty"`
	CoverLocked             bool                       `json:"coverLocked,omitempty"`
	Archived                bool                       `json:"archived"`
	Done                    bool                       `json:"done"`
	CommentCount            int                        `json:"commentCount,omitempty"`
	ChecklistCount          int                        `json:"checklistCount,omitempty"`
	ChecklistCompletedCount int                        `json:"checklistCompletedCount,omitempty"`
	IsRepeating             bool                       `json:"isRepeating,omitempty"`
	IsRead                  bool                       `json:"isRead,omitempty"`
	IsSeen                  bool                       `json:"isSeen,omitempty"`
	CreatedAt               string                     `json:"createdAt"`
	UpdatedAt               string                     `json:"updatedAt"`
	Users                   []common.User              `json:"users,omitempty"`
	Tags                    []common.Tag               `json:"tags,omitempty"`
	TodoList                *common.TodoListInfo       `json:"todoList,omitempty"`
	CustomFields            []SimpleCustomFieldValue   `json:"customFields,omitempty"`
}

// TodoRecordResponse represents the response from the single record GraphQL query
type TodoRecordResponse struct {
	Todo DetailedRecord `json:"todo"`
}

func RunReadRecord(args []string) error {
	fs := flag.NewFlagSet("read-record", flag.ExitOnError)
	recordID := fs.String("record", "", "Record ID (required)")
	projectID := fs.String("project", "", "Project ID or slug (required)")
	simple := fs.Bool("simple", false, "Show only basic record information")
	fs.Parse(args)

	// Validate required parameters
	if *recordID == "" || *projectID == "" {
		if *recordID == "" {
			fmt.Println("Error: -record flag is required")
		}
		if *projectID == "" {
			fmt.Println("Error: -project flag is required")
		}
		fmt.Println("\nUsage:")
		fmt.Println("  go run . read-record -record RECORD_ID -project PROJECT_ID [flags]")
		fmt.Println("\nFlags:")
		fs.PrintDefaults()
		fmt.Println("\nExamples:")
		fmt.Println("  # Get detailed record information")
		fmt.Println("  go run . read-record -record RECORD_ID -project PROJECT_ID")
		fmt.Println("")
		fmt.Println("  # Get simple record information")
		fmt.Println("  go run . read-record -record RECORD_ID -project PROJECT_ID -simple")
		return fmt.Errorf("required flags missing")
	}

	// Load configuration
	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Create client
	client := common.NewClient(config)

	// Set project context (now required)
	client.SetProject(*projectID)
	
	// Fetch custom field metadata
	var customFieldInfo map[string]RecordCustomFieldInfo
	info, err := getRecordCustomFieldInfo(client, *projectID)
	if err != nil {
		// Don't fail if we can't get field info, just log it
		fmt.Printf("Warning: Could not fetch custom field info: %v\n", err)
		customFieldInfo = make(map[string]RecordCustomFieldInfo)
	} else {
		customFieldInfo = info
	}

	// Build the GraphQL query with embedded ID (like test_custom_fields.go)
	query := buildRecordDetailQuery(*simple, *recordID)

	// Execute query without variables (ID is embedded)
	var response TodoRecordResponse
	if err := client.ExecuteQueryWithResult(query, nil, &response); err != nil {
		return fmt.Errorf("failed to execute query: %v", err)
	}

	record := response.Todo
	
	// Check if record was found
	if record.ID == "" {
		fmt.Printf("Record with ID '%s' not found.\n", *recordID)
		return nil
	}

	// Display results
	displayRecordDetails(record, *simple, customFieldInfo)
	
	return nil
}

// displayRecordDetails displays the record information
func displayRecordDetails(record DetailedRecord, simple bool, customFieldInfo map[string]RecordCustomFieldInfo) {
	fmt.Printf("\n=== Record Details ===\n")
	fmt.Printf("ID: %s\n", record.ID)
	fmt.Printf("UID: %s\n", record.UID)
	fmt.Printf("Title: %s\n", record.Title)

	if record.TodoList != nil {
		fmt.Printf("List: %s (%s)\n", record.TodoList.Title, record.TodoList.ID)
	}

	fmt.Printf("Position: %.0f\n", record.Position)
	fmt.Printf("Status: %s\n", getDetailedRecordStatus(record))

	if !simple {
		// Description/Content
		if record.Text != "" {
			fmt.Printf("Description: %s\n", record.Text)
		}
		if record.HTML != "" && record.HTML != record.Text {
			fmt.Printf("HTML Content: %s\n", common.TruncateString(record.HTML, 200))
		}

		// Dates
		if record.StartedAt != "" {
			fmt.Printf("Started At: %s\n", record.StartedAt)
		}
		if record.DuedAt != "" {
			fmt.Printf("Due At: %s\n", record.DuedAt)
		}
		if record.Timezone != "" {
			fmt.Printf("Timezone: %s\n", record.Timezone)
		}

		// Visual properties
		if record.Color != "" {
			fmt.Printf("Color: %s\n", record.Color)
		}
		if record.Cover != "" {
			fmt.Printf("Cover: Yes%s\n", func() string {
				if record.CoverLocked {
					return " (locked)"
				}
				return ""
			}())
		}

		// Counts and flags
		fmt.Printf("Comments: %d\n", record.CommentCount)
		fmt.Printf("Checklists: %d/%d completed\n", record.ChecklistCompletedCount, record.ChecklistCount)
		
		if record.IsRepeating {
			fmt.Printf("Repeating: Yes\n")
		}
		if record.IsRead {
			fmt.Printf("Read: Yes\n")
		}
		if record.IsSeen {
			fmt.Printf("Seen: Yes\n")
		}

		// Assignees
		if len(record.Users) > 0 {
			fmt.Printf("\n=== Assignees (%d) ===\n", len(record.Users))
			for _, user := range record.Users {
				fmt.Printf("- %s (%s) [%s]\n", user.FullName, user.Email, user.ID)
			}
		}

		// Tags
		if len(record.Tags) > 0 {
			fmt.Printf("\n=== Tags (%d) ===\n", len(record.Tags))
			for _, tag := range record.Tags {
				color := tag.Color
				if color == "" {
					color = "default"
				}
				fmt.Printf("- %s [%s] (%s)\n", tag.Title, color, tag.ID)
			}
		}

		// Custom Fields
		if len(record.CustomFields) > 0 {
			fmt.Printf("\n=== Custom Fields (%d) ===\n", len(record.CustomFields))
			for _, cfv := range record.CustomFields {
				// Use field name with type and ID if available, otherwise just ID
				fieldDisplay := cfv.ID
				if info, exists := customFieldInfo[cfv.ID]; exists {
					fieldDisplay = fmt.Sprintf("%s (%s) [%s]", info.Name, info.Type, cfv.ID)
				}
				
				parsedValue := parseRecordCustomFieldValue(cfv.Value)
				if parsedValue != nil {
					fmt.Printf("- %s: %v\n", fieldDisplay, parsedValue)
				} else {
					fmt.Printf("- %s: (empty)\n", fieldDisplay)
				}
			}
		}

		// Timestamps
		fmt.Printf("Created At: %s\n", record.CreatedAt)
		fmt.Printf("Updated At: %s\n", record.UpdatedAt)
	} else {
		// Simple output - just show key info
		if record.DuedAt != "" {
			fmt.Printf("Due: %s\n", record.DuedAt)
		}
		if len(record.Users) > 0 {
			fmt.Printf("Assignees: %d\n", len(record.Users))
		}
		if len(record.Tags) > 0 {
			fmt.Printf("Tags: %d\n", len(record.Tags))
		}
		if len(record.CustomFields) > 0 {
			fmt.Printf("Custom Fields: %d\n", len(record.CustomFields))
		}
	}
}

// buildRecordDetailQuery builds the GraphQL query based on the detail level
func buildRecordDetailQuery(simple bool, recordID string) string {
	if simple {
		return fmt.Sprintf(`
			query GetRecord {
				todo(id: "%s") {
					id
					uid
					title
					done
					customFields {
						id
						value
					}
				}
			}
		`, recordID)
	}

	return fmt.Sprintf(`
		query GetRecord {
			todo(id: "%s") {
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
				customFields {
					id
					value
				}
			}
		}
	`, recordID)
}

// getDetailedRecordStatus returns a human-readable status for a detailed record
func getDetailedRecordStatus(record DetailedRecord) string {
	if record.Archived {
		return "Archived"
	}
	if record.Done {
		return "Completed"
	}
	return "Active"
}

// parseRecordCustomFieldValue parses complex custom field values to extract the meaningful data
func parseRecordCustomFieldValue(value interface{}) interface{} {
	if value == nil {
		return nil
	}
	
	// Handle map structure like map[currency:<nil> number:75000 text:<nil>]
	if valueMap, ok := value.(map[string]interface{}); ok {
		// Check for different value types in priority order
		if num := valueMap["number"]; num != nil {
			return num
		}
		if curr := valueMap["currency"]; curr != nil {
			return curr
		}
		if txt := valueMap["text"]; txt != nil {
			return txt
		}
		if date := valueMap["date"]; date != nil {
			return date
		}
		if boolean := valueMap["boolean"]; boolean != nil {
			return boolean
		}
		// Return the whole map if we can't find a specific type
		return value
	}
	
	// Return as-is for simple values
	return value
}

// getRecordCustomFieldInfo fetches custom field metadata from the project
func getRecordCustomFieldInfo(client *common.Client, projectID string) (map[string]RecordCustomFieldInfo, error) {
	query := `
		query GetProjectCustomFields {
			customFields {
				items {
					id
					name
					type
				}
			}
		}
	`
	
	// Set project context
	client.SetProject(projectID)
	
	var response struct {
		CustomFields struct {
			Items []RecordCustomFieldInfo `json:"items"`
		} `json:"customFields"`
	}
	
	if err := client.ExecuteQueryWithResult(query, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to fetch custom fields: %v", err)
	}
	
	// Convert to map for quick lookup
	fieldMap := make(map[string]RecordCustomFieldInfo)
	for _, field := range response.CustomFields.Items {
		fieldMap[field.ID] = field
	}
	
	return fieldMap, nil
}