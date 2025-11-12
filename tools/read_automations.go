package tools

import (
	"encoding/json"
	"flag"
	"fmt"

	. "cli/common"
)

// Helper function
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Response structures
type AutomationListResponse struct {
	AutomationList struct {
		Items      []AutomationItem `json:"items"`
		TotalCount int              `json:"totalCount"`
		PageInfo   *AutomationPageInfo `json:"pageInfo,omitempty"`
	} `json:"automationList"`
}

type AutomationPageInfo struct {
	HasNext       bool `json:"hasNext"`
	HasPrevious   bool `json:"hasPrevious"`
	TotalItems    int  `json:"totalItems"`
	CurrentPage   int  `json:"currentPage"`
	TotalPages    int  `json:"totalPages"`
}

type AutomationItem struct {
	ID        string                   `json:"id"`
	UID       string                   `json:"uid"`
	IsActive  bool                     `json:"isActive"`
	CreatedAt string                   `json:"createdAt"`
	UpdatedAt string                   `json:"updatedAt"`
	Trigger   AutomationTriggerDetailed `json:"trigger"`
	Actions   []AutomationActionDetailed `json:"actions"`
}

// Extended structures for detailed automation reading
type AutomationTriggerDetailed struct {
	ID                   string                            `json:"id"`
	Type                 string                            `json:"type"`
	Metadata             *AutomationTriggerMetadataDetailed `json:"metadata,omitempty"`
	CustomField          *CustomField                      `json:"customField,omitempty"`
	CustomFieldOptions   []CustomFieldOption               `json:"customFieldOptions,omitempty"`
	TodoList             *TodoList                         `json:"todoList,omitempty"`
	Tags                 []Tag                             `json:"tags,omitempty"`
	Assignees            []User                            `json:"assignees,omitempty"`
	Color                *string                           `json:"color,omitempty"`
}

type AutomationActionDetailed struct {
	ID                   string                           `json:"id"`
	Type                 string                           `json:"type"`
	DuedIn               *int                             `json:"duedIn,omitempty"`
	Metadata             *AutomationActionMetadataDetailed `json:"metadata,omitempty"`
	CustomField          *CustomField                     `json:"customField,omitempty"`
	CustomFieldOptions   []CustomFieldOption              `json:"customFieldOptions,omitempty"`
	TodoList             *TodoList                        `json:"todoList,omitempty"`
	Tags                 []Tag                            `json:"tags,omitempty"`
	Assignees            []User                           `json:"assignees,omitempty"`
	Color                *string                          `json:"color,omitempty"`
	AssigneeTriggerer    *string                          `json:"assigneeTriggerer,omitempty"`
	HttpOption           *HttpOption                      `json:"httpOption,omitempty"`
}

// Metadata structures
type AutomationTriggerMetadataDetailed struct {
	IncompleteOnly *bool `json:"incompleteOnly,omitempty"`
}

type AutomationActionMetadataDetailed struct {
	Checklists       []AutomationChecklist `json:"checklists,omitempty"`
	CopyTodoOptions  []string              `json:"copyTodoOptions,omitempty"`
	Email            *AutomationEmail      `json:"email,omitempty"`
}

type AutomationChecklist struct {
	Title          string                         `json:"title"`
	Position       float64                        `json:"position"`
	ChecklistItems []AutomationChecklistItem      `json:"checklistItems,omitempty"`
}

type AutomationChecklistItem struct {
	Title       string   `json:"title"`
	Position    float64  `json:"position"`
	DuedIn      *int     `json:"duedIn,omitempty"`
	AssigneeIds []string `json:"assigneeIds,omitempty"`
}

type AutomationEmail struct {
	From        *string                    `json:"from,omitempty"`
	To          []string                   `json:"to"`
	Bcc         []string                   `json:"bcc,omitempty"`
	Cc          []string                   `json:"cc,omitempty"`
	Subject     string                     `json:"subject"`
	Content     string                     `json:"content"`
	Attachments []AutomationEmailAttachment `json:"attachments,omitempty"`
}

type AutomationEmailAttachment struct {
	UID       string  `json:"uid"`
	Name      string  `json:"name"`
	Size      float64 `json:"size"`
	Type      string  `json:"type"`
	Extension string  `json:"extension"`
}

type HttpOption struct {
	ID                  string          `json:"id"`
	UID                 string          `json:"uid"`
	URL                 string          `json:"url"`
	Method              string          `json:"method"`
	Headers             []HttpHeader    `json:"headers,omitempty"`
	Parameters          []HttpParameter `json:"parameters,omitempty"`
	Body                *string         `json:"body,omitempty"`
	ContentType         *string         `json:"contentType,omitempty"`
	AuthorizationType   *string         `json:"authorizationType,omitempty"`
}

type HttpHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type HttpParameter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}



// Execute GraphQL query
func executeReadAutomations(client *Client, projectID string, skip int, take int) (*AutomationListResponse, error) {
	query := fmt.Sprintf(`
		query AutomationList {
			automationList(skip: %d, take: %d) {
				totalCount
				items {
					id
					uid
					isActive
					createdAt
					updatedAt
					trigger {
						id
						type
						metadata {
							... on AutomationTriggerMetadataTodoOverdue {
								incompleteOnly
							}
						}
						customField {
							id
							name
						}
						customFieldOptions {
							id
							title
							color
						}
						todoList {
							id
							title
						}
						tags {
							id
							title
							color
						}
						assignees {
							id
							fullName
						}
						color
					}
					actions {
						id
						type
						duedIn
						metadata {
							... on AutomationActionMetadataCreateChecklist {
								checklists {
									title
									position
									checklistItems {
										title
										position
										duedIn
										assigneeIds
									}
								}
							}
							... on AutomationActionMetadataCopyTodo {
								copyTodoOptions
							}
							... on AutomationActionMetadataSendEmail {
								email {
									from
									to
									bcc
									cc
									subject
									content
									attachments {
										uid
										name
										size
										type
										extension
									}
								}
							}
						}
						customField {
							id
							name
						}
						customFieldOptions {
							id
							title
							color
						}
						todoList {
							id
							title
						}
						tags {
							id
							title
							color
						}
						assignees {
							id
							fullName
						}
						color
						assigneeTriggerer
						httpOption {
							id
							uid
							url
							method
							headers {
								key
								value
							}
							parameters {
								key
								value
							}
							body
							contentType
							authorizationType
						}
					}
				}
			}
		}
	`, skip, take)

	var response AutomationListResponse
	result, err := client.ExecuteQuery(query, nil)
	if err != nil {
		return nil, err
	}

	// Parse the response
	data, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %v", err)
	}

	err = json.Unmarshal(data, &response)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return &response, nil
}

// Command-line interface
func RunReadAutomations(args []string) error {
	fs := flag.NewFlagSet("read-automations", flag.ExitOnError)
	
	projectID := fs.String("project", "", "Project ID or slug (required)")
	simple := fs.Bool("simple", false, "Simple output format")
	page := fs.Int("page", 1, "Page number (default: 1)")
	size := fs.Int("size", 50, "Page size - number of items per page (default: 50, max: 100)")
	skip := fs.Int("skip", 0, "Number of items to skip (overrides page if set)")
	limit := fs.Int("limit", 0, "Maximum number of items to return (overrides size if set)")

	err := fs.Parse(args)
	if err != nil {
		return err
	}

	// Validate required fields
	if *projectID == "" {
		return fmt.Errorf("project ID is required")
	}

	// Validate page size
	if *size > 100 {
		*size = 100
	}
	if *size <= 0 {
		*size = 50
	}

	// Calculate skip and take values
	var skipValue, takeValue int
	
	if *skip > 0 {
		// Use skip directly if provided
		skipValue = *skip
	} else {
		// Calculate skip from page number
		skipValue = (*page - 1) * *size
	}
	
	if *limit > 0 {
		// Use limit if provided
		takeValue = *limit
	} else {
		// Use size
		takeValue = *size
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Create client
	client := NewClient(config)

	// Set project context
	client.SetProject(*projectID)

	// Execute query
	response, err := executeReadAutomations(client, *projectID, skipValue, takeValue)
	if err != nil {
		return fmt.Errorf("failed to read automations: %v", err)
	}

	// Output results
	automations := response.AutomationList.Items
	totalCount := response.AutomationList.TotalCount
	pageInfo := response.AutomationList.PageInfo

	if *simple {
		fmt.Printf("=== Automations in Project %s ===\n", *projectID)
		fmt.Printf("Total automations: %d\n", totalCount)
		if pageInfo != nil {
			fmt.Printf("Showing items %d-%d (Page %d of %d)\n", 
				skipValue+1, 
				min(skipValue+len(automations), totalCount), 
				pageInfo.CurrentPage, 
				pageInfo.TotalPages)
		} else {
			fmt.Printf("Showing %d items (skip: %d, take: %d)\n", len(automations), skipValue, takeValue)
		}
		fmt.Printf("\n")
		
		for i, automation := range automations {
			status := "Inactive"
			if automation.IsActive {
				status = "Active"
			}
			fmt.Printf("%d. Automation %s (%s)\n", i+1, automation.UID, status)
			fmt.Printf("   ID: %s\n", automation.ID)
			fmt.Printf("   Trigger: %s\n", automation.Trigger.Type)
			if len(automation.Actions) > 0 {
				fmt.Printf("   Action: %s\n", automation.Actions[0].Type)
			}
			fmt.Printf("\n")
		}
	} else {
		fmt.Printf("=== Automations in Project %s ===\n", *projectID)
		fmt.Printf("Total automations: %d\n", totalCount)
		if pageInfo != nil {
			fmt.Printf("Showing items %d-%d (Page %d of %d)\n", 
				skipValue+1, 
				min(skipValue+len(automations), totalCount), 
				pageInfo.CurrentPage, 
				pageInfo.TotalPages)
			if pageInfo.HasPrevious {
				fmt.Printf("Has previous page: Yes\n")
			}
			if pageInfo.HasNext {
				fmt.Printf("Has next page: Yes\n")
			}
		} else {
			fmt.Printf("Showing %d items (skip: %d, take: %d)\n", len(automations), skipValue, takeValue)
		}
		fmt.Printf("\n")
		
		for i, automation := range automations {
			status := "Inactive"
			statusIcon := "❌"
			if automation.IsActive {
				status = "Active"
				statusIcon = "✅"
			}
			fmt.Printf("╭─ %d. Automation %s %s %s\n", i+1, automation.UID, statusIcon, status)
			fmt.Printf("│  ID: %s\n", automation.ID)
			fmt.Printf("│  UID: %s\n", automation.UID)
			fmt.Printf("│  Active: %t\n", automation.IsActive)
			fmt.Printf("│  Created: %s\n", automation.CreatedAt)
			fmt.Printf("│  Updated: %s\n", automation.UpdatedAt)
			fmt.Printf("│\n")
			
			fmt.Printf("├─ 🎯 Trigger:\n")
			fmt.Printf("│  │  ID: %s\n", automation.Trigger.ID)
			fmt.Printf("│  │  Type: %s\n", automation.Trigger.Type)
			
			// Trigger metadata
			if automation.Trigger.Metadata != nil {
				fmt.Printf("│  │  📋 Metadata:\n")
				if automation.Trigger.Metadata.IncompleteOnly != nil {
					fmt.Printf("│  │     Incomplete Only: %t\n", *automation.Trigger.Metadata.IncompleteOnly)
				}
			}
			
			// Trigger custom field
			if automation.Trigger.CustomField != nil {
				fmt.Printf("│  │  🏷️  Custom Field: %s (%s)\n", automation.Trigger.CustomField.Name, automation.Trigger.CustomField.ID)
			}
			
			// Trigger custom field options
			if len(automation.Trigger.CustomFieldOptions) > 0 {
				fmt.Printf("│  │  🔧 Custom Field Options:\n")
				for _, option := range automation.Trigger.CustomFieldOptions {
					fmt.Printf("│  │     - %s (%s) [%s]\n", option.Title, option.ID, option.Color)
				}
			}
			
			// Trigger todo list
			if automation.Trigger.TodoList != nil {
				fmt.Printf("│  │  📝 List: %s (%s)\n", automation.Trigger.TodoList.Title, automation.Trigger.TodoList.ID)
			}
			
			// Trigger tags
			if len(automation.Trigger.Tags) > 0 {
				fmt.Printf("│  │  🏷️  Tags:\n")
				for _, tag := range automation.Trigger.Tags {
					fmt.Printf("│  │     - %s (%s) [%s]\n", tag.Title, tag.ID, tag.Color)
				}
			}
			
			// Trigger assignees
			if len(automation.Trigger.Assignees) > 0 {
				fmt.Printf("│  │  👥 Assignees:\n")
				for _, assignee := range automation.Trigger.Assignees {
					fmt.Printf("│  │     - %s (%s)\n", assignee.FullName, assignee.ID)
				}
			}
			
			// Trigger color
			if automation.Trigger.Color != nil {
				fmt.Printf("│  │  🎨 Color: %s\n", *automation.Trigger.Color)
			}
			
			fmt.Printf("│\n")
			fmt.Printf("└─ ⚡ Actions (%d):\n", len(automation.Actions))
			if len(automation.Actions) == 0 {
				fmt.Printf("   No actions defined\n")
			} else {
				for j, action := range automation.Actions {
					isLast := j == len(automation.Actions)-1
					connector := "├─"
					prefix := "│  "
					if isLast {
						connector = "└─"
						prefix = "   "
					}
					fmt.Printf("   %s %d. ID: %s\n", connector, j+1, action.ID)
					fmt.Printf("   %s    Type: %s\n", prefix, action.Type)
					
					if action.DuedIn != nil {
						fmt.Printf("   %s    ⏰ Due in: %d days\n", prefix, *action.DuedIn)
					}
					
					// Action metadata
					if action.Metadata != nil {
						fmt.Printf("   %s    📋 Metadata:\n", prefix)
						
						// Checklist metadata
						if len(action.Metadata.Checklists) > 0 {
							fmt.Printf("   %s       ✅ Checklists:\n", prefix)
							for k, checklist := range action.Metadata.Checklists {
								fmt.Printf("   %s          %d. %s (pos: %.1f)\n", prefix, k+1, checklist.Title, checklist.Position)
								for l, item := range checklist.ChecklistItems {
									fmt.Printf("   %s             %d.%d. %s (pos: %.1f)", prefix, k+1, l+1, item.Title, item.Position)
									if item.DuedIn != nil {
										fmt.Printf(" [⏰ due: %d days]", *item.DuedIn)
									}
									if len(item.AssigneeIds) > 0 {
										fmt.Printf(" [👥 assignees: %v]", item.AssigneeIds)
									}
									fmt.Printf("\n")
								}
							}
						}
						
						// Copy todo metadata
						if len(action.Metadata.CopyTodoOptions) > 0 {
							fmt.Printf("   %s       📋 Copy Todo Options: %v\n", prefix, action.Metadata.CopyTodoOptions)
						}
						
						// Email metadata
						if action.Metadata.Email != nil {
							email := action.Metadata.Email
							fmt.Printf("   %s       📧 Email:\n", prefix)
							if email.From != nil {
								fmt.Printf("   %s          From: %s\n", prefix, *email.From)
							}
							fmt.Printf("   %s          To: %v\n", prefix, email.To)
							if len(email.Cc) > 0 {
								fmt.Printf("   %s          Cc: %v\n", prefix, email.Cc)
							}
							if len(email.Bcc) > 0 {
								fmt.Printf("   %s          Bcc: %v\n", prefix, email.Bcc)
							}
							fmt.Printf("   %s          Subject: %s\n", prefix, email.Subject)
							fmt.Printf("   %s          Content: %s\n", prefix, email.Content)
							if len(email.Attachments) > 0 {
								fmt.Printf("   %s          📎 Attachments:\n", prefix)
								for _, attachment := range email.Attachments {
									fmt.Printf("   %s            - %s (%s, %.2f bytes) [%s]\n", prefix,
										attachment.Name, attachment.Type, attachment.Size, attachment.UID)
								}
							}
						}
					}
					
					// Action custom field
					if action.CustomField != nil {
						fmt.Printf("   %s    🏷️  Custom Field: %s (%s)\n", prefix, action.CustomField.Name, action.CustomField.ID)
					}
					
					// Action custom field options
					if len(action.CustomFieldOptions) > 0 {
						fmt.Printf("   %s    🔧 Custom Field Options:\n", prefix)
						for _, option := range action.CustomFieldOptions {
							fmt.Printf("   %s       - %s (%s) [%s]\n", prefix, option.Title, option.ID, option.Color)
						}
					}
					
					// Action todo list
					if action.TodoList != nil {
						fmt.Printf("   %s    📝 List: %s (%s)\n", prefix, action.TodoList.Title, action.TodoList.ID)
					}
					
					// Action tags
					if len(action.Tags) > 0 {
						fmt.Printf("   %s    🏷️  Tags:\n", prefix)
						for _, tag := range action.Tags {
							fmt.Printf("   %s       - %s (%s) [%s]\n", prefix, tag.Title, tag.ID, tag.Color)
						}
					}
					
					// Action assignees
					if len(action.Assignees) > 0 {
						fmt.Printf("   %s    👥 Assignees:\n", prefix)
						for _, assignee := range action.Assignees {
							fmt.Printf("   %s       - %s (%s)\n", prefix, assignee.FullName, assignee.ID)
						}
					}
					
					// Action color
					if action.Color != nil {
						fmt.Printf("   %s    🎨 Color: %s\n", prefix, *action.Color)
					}
					
					// Assignee triggerer
					if action.AssigneeTriggerer != nil {
						fmt.Printf("   %s    👤 Assignee Triggerer: %s\n", prefix, *action.AssigneeTriggerer)
					}
					
					// HTTP options
					if action.HttpOption != nil {
						http := action.HttpOption
						fmt.Printf("   %s    🌐 HTTP Webhook:\n", prefix)
						fmt.Printf("   %s       ID: %s\n", prefix, http.ID)
						fmt.Printf("   %s       UID: %s\n", prefix, http.UID)
						fmt.Printf("   %s       URL: %s\n", prefix, http.URL)
						fmt.Printf("   %s       Method: %s\n", prefix, http.Method)
						if len(http.Headers) > 0 {
							fmt.Printf("   %s       📋 Headers:\n", prefix)
							for _, header := range http.Headers {
								fmt.Printf("   %s          %s: %s\n", prefix, header.Key, header.Value)
							}
						}
						if len(http.Parameters) > 0 {
							fmt.Printf("   %s       🔧 Parameters:\n", prefix)
							for _, param := range http.Parameters {
								fmt.Printf("   %s          %s: %s\n", prefix, param.Key, param.Value)
							}
						}
						if http.Body != nil {
							fmt.Printf("   %s       📄 Body: %s\n", prefix, *http.Body)
						}
						if http.ContentType != nil {
							fmt.Printf("   %s       📝 Content Type: %s\n", prefix, *http.ContentType)
						}
						if http.AuthorizationType != nil {
							fmt.Printf("   %s       🔐 Authorization Type: %s\n", prefix, *http.AuthorizationType)
						}
					}
				}
			}
			fmt.Printf("\n")
		}
	}

	if totalCount == 0 {
		fmt.Printf("No automations found in this project.\n")
	} else if pageInfo != nil {
		fmt.Printf("═══════════════════════════════════════════════\n")
		fmt.Printf("📊 Pagination Summary:\n")
		fmt.Printf("   Total items: %d\n", pageInfo.TotalItems)
		fmt.Printf("   Current page: %d of %d\n", pageInfo.CurrentPage, pageInfo.TotalPages)
		fmt.Printf("   Items shown: %d-%d\n", skipValue+1, min(skipValue+len(automations), totalCount))
		if pageInfo.HasPrevious {
			fmt.Printf("   ⬅️  Previous page available\n")
		}
		if pageInfo.HasNext {
			fmt.Printf("   ➡️  Next page available\n")
		}
		fmt.Printf("═══════════════════════════════════════════════\n")
		fmt.Printf("\n💡 Usage Examples:\n")
		fmt.Printf("   Next page: go run . read-automations -project %s -page %d\n", *projectID, pageInfo.CurrentPage+1)
		if pageInfo.HasPrevious {
			fmt.Printf("   Prev page: go run . read-automations -project %s -page %d\n", *projectID, pageInfo.CurrentPage-1)
		}
		fmt.Printf("   Custom size: go run . read-automations -project %s -size 10\n", *projectID)
		fmt.Printf("   Skip/limit: go run . read-automations -project %s -skip %d -limit 25\n", *projectID, skipValue+takeValue)
	}

	return nil
}