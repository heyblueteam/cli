package automations

import (
	"encoding/json"
	"fmt"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

// Response structures
type AutomationListResponse struct {
	AutomationList struct {
		Items      []AutomationItem    `json:"items"`
		TotalCount int                 `json:"totalCount"`
		PageInfo   *AutomationPageInfo `json:"pageInfo,omitempty"`
	} `json:"automationList"`
}

type AutomationPageInfo struct {
	HasNext     bool `json:"hasNext"`
	HasPrevious bool `json:"hasPrevious"`
	TotalItems  int  `json:"totalItems"`
	CurrentPage int  `json:"currentPage"`
	TotalPages  int  `json:"totalPages"`
}

type AutomationItem struct {
	ID        string                     `json:"id"`
	UID       string                     `json:"uid"`
	IsActive  bool                       `json:"isActive"`
	CreatedAt string                     `json:"createdAt"`
	UpdatedAt string                     `json:"updatedAt"`
	Trigger   AutomationTriggerDetailed  `json:"trigger"`
	Actions   []AutomationActionDetailed `json:"actions"`
}

type AutomationTriggerDetailed struct {
	ID                 string                              `json:"id"`
	Type               string                              `json:"type"`
	Metadata           *AutomationTriggerMetadataDetailed  `json:"metadata,omitempty"`
	CustomField        *common.CustomField                 `json:"customField,omitempty"`
	CustomFieldOptions []common.CustomFieldOption           `json:"customFieldOptions,omitempty"`
	TodoList           *common.TodoList                    `json:"todoList,omitempty"`
	Tags               []common.Tag                        `json:"tags,omitempty"`
	Assignees          []common.User                       `json:"assignees,omitempty"`
	Color              *string                             `json:"color,omitempty"`
}

type AutomationActionDetailed struct {
	ID                 string                             `json:"id"`
	Type               string                             `json:"type"`
	DuedIn             *int                               `json:"duedIn,omitempty"`
	Metadata           *AutomationActionMetadataDetailed  `json:"metadata,omitempty"`
	CustomField        *common.CustomField                `json:"customField,omitempty"`
	CustomFieldOptions []common.CustomFieldOption          `json:"customFieldOptions,omitempty"`
	TodoList           *common.TodoList                   `json:"todoList,omitempty"`
	Tags               []common.Tag                       `json:"tags,omitempty"`
	Assignees          []common.User                      `json:"assignees,omitempty"`
	Color              *string                            `json:"color,omitempty"`
	AssigneeTriggerer  *string                            `json:"assigneeTriggerer,omitempty"`
	HttpOption         *HttpOption                        `json:"httpOption,omitempty"`
}

type AutomationTriggerMetadataDetailed struct {
	IncompleteOnly *bool `json:"incompleteOnly,omitempty"`
}

type AutomationActionMetadataDetailed struct {
	Checklists      []AutomationChecklist `json:"checklists,omitempty"`
	CopyTodoOptions []string              `json:"copyTodoOptions,omitempty"`
	Email           *AutomationEmail      `json:"email,omitempty"`
}

type AutomationChecklist struct {
	Title          string                    `json:"title"`
	Position       float64                   `json:"position"`
	ChecklistItems []AutomationChecklistItem `json:"checklistItems,omitempty"`
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
	ID                string          `json:"id"`
	UID               string          `json:"uid"`
	URL               string          `json:"url"`
	Method            string          `json:"method"`
	Headers           []HttpHeader    `json:"headers,omitempty"`
	Parameters        []HttpParameter `json:"parameters,omitempty"`
	Body              *string         `json:"body,omitempty"`
	ContentType       *string         `json:"contentType,omitempty"`
	AuthorizationType *string         `json:"authorizationType,omitempty"`
}

type HttpHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type HttpParameter struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List automations in a workspace",
	Long:  "List automations in a workspace with pagination support.",
	Example: `  blue automations list --workspace <id>
  blue automations list -w <id> --simple
  blue automations list -w <id> --page 2 --size 10
  blue automations list -w <id> --skip 20 --limit 5`,
	RunE: runList,
}

var (
	listWorkspace string
	listSimple    bool
	listPage      int
	listSize      int
	listSkip      int
	listLimit     int
)

func init() {
	listCmd.Flags().StringVarP(&listWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	listCmd.Flags().BoolVarP(&listSimple, "simple", "s", false, "Simple output format")
	listCmd.Flags().IntVar(&listPage, "page", 1, "Page number (default: 1)")
	listCmd.Flags().IntVar(&listSize, "size", 50, "Page size (default: 50, max: 100)")
	listCmd.Flags().IntVar(&listSkip, "skip", 0, "Number of items to skip (overrides page)")
	listCmd.Flags().IntVar(&listLimit, "limit", 0, "Max items to return (overrides size)")
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func runList(cmd *cobra.Command, args []string) error {
	if listWorkspace == "" {
		return fmt.Errorf("workspace ID is required. Use --workspace flag")
	}

	// Validate page size
	if listSize > 100 {
		listSize = 100
	}
	if listSize <= 0 {
		listSize = 50
	}

	// Calculate skip and take
	var skipValue, takeValue int
	if listSkip > 0 {
		skipValue = listSkip
	} else {
		skipValue = (listPage - 1) * listSize
	}
	if listLimit > 0 {
		takeValue = listLimit
	} else {
		takeValue = listSize
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)
	client.SetProject(listWorkspace)

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
	`, skipValue, takeValue)

	result, err := client.ExecuteQuery(query, nil)
	if err != nil {
		return fmt.Errorf("failed to read automations: %w", err)
	}

	data, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	var response AutomationListResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return fmt.Errorf("failed to unmarshal response: %w", err)
	}

	automations := response.AutomationList.Items
	totalCount := response.AutomationList.TotalCount
	pageInfo := response.AutomationList.PageInfo

	if listSimple {
		fmt.Printf("=== Automations in Workspace %s ===\n", listWorkspace)
		fmt.Printf("Total automations: %d\n", totalCount)
		if pageInfo != nil {
			fmt.Printf("Showing items %d-%d (Page %d of %d)\n",
				skipValue+1,
				minInt(skipValue+len(automations), totalCount),
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
		fmt.Printf("=== Automations in Workspace %s ===\n", listWorkspace)
		fmt.Printf("Total automations: %d\n", totalCount)
		if pageInfo != nil {
			fmt.Printf("Showing items %d-%d (Page %d of %d)\n",
				skipValue+1,
				minInt(skipValue+len(automations), totalCount),
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
			statusIcon := "X"
			if automation.IsActive {
				status = "Active"
				statusIcon = "+"
			}
			fmt.Printf("[%s] %d. Automation %s - %s\n", statusIcon, i+1, automation.UID, status)
			fmt.Printf("    ID: %s\n", automation.ID)
			fmt.Printf("    UID: %s\n", automation.UID)
			fmt.Printf("    Active: %t\n", automation.IsActive)
			fmt.Printf("    Created: %s\n", automation.CreatedAt)
			fmt.Printf("    Updated: %s\n", automation.UpdatedAt)
			fmt.Printf("\n")

			fmt.Printf("    Trigger:\n")
			fmt.Printf("      ID: %s\n", automation.Trigger.ID)
			fmt.Printf("      Type: %s\n", automation.Trigger.Type)

			if automation.Trigger.Metadata != nil {
				if automation.Trigger.Metadata.IncompleteOnly != nil {
					fmt.Printf("      Incomplete Only: %t\n", *automation.Trigger.Metadata.IncompleteOnly)
				}
			}
			if automation.Trigger.CustomField != nil {
				fmt.Printf("      Custom Field: %s (%s)\n", automation.Trigger.CustomField.Name, automation.Trigger.CustomField.ID)
			}
			if len(automation.Trigger.CustomFieldOptions) > 0 {
				fmt.Printf("      Custom Field Options:\n")
				for _, option := range automation.Trigger.CustomFieldOptions {
					fmt.Printf("        - %s (%s) [%s]\n", option.Title, option.ID, option.Color)
				}
			}
			if automation.Trigger.TodoList != nil {
				fmt.Printf("      List: %s (%s)\n", automation.Trigger.TodoList.Title, automation.Trigger.TodoList.ID)
			}
			if len(automation.Trigger.Tags) > 0 {
				fmt.Printf("      Tags:\n")
				for _, tag := range automation.Trigger.Tags {
					fmt.Printf("        - %s (%s) [%s]\n", tag.Title, tag.ID, tag.Color)
				}
			}
			if len(automation.Trigger.Assignees) > 0 {
				fmt.Printf("      Assignees:\n")
				for _, assignee := range automation.Trigger.Assignees {
					fmt.Printf("        - %s (%s)\n", assignee.FullName, assignee.ID)
				}
			}
			if automation.Trigger.Color != nil {
				fmt.Printf("      Color: %s\n", *automation.Trigger.Color)
			}

			fmt.Printf("\n    Actions (%d):\n", len(automation.Actions))
			if len(automation.Actions) == 0 {
				fmt.Printf("      No actions defined\n")
			} else {
				for j, action := range automation.Actions {
					fmt.Printf("      %d. ID: %s\n", j+1, action.ID)
					fmt.Printf("         Type: %s\n", action.Type)

					if action.DuedIn != nil {
						fmt.Printf("         Due in: %d days\n", *action.DuedIn)
					}

					if action.Metadata != nil {
						if len(action.Metadata.Checklists) > 0 {
							fmt.Printf("         Checklists:\n")
							for k, checklist := range action.Metadata.Checklists {
								fmt.Printf("           %d. %s (pos: %.1f)\n", k+1, checklist.Title, checklist.Position)
								for l, item := range checklist.ChecklistItems {
									fmt.Printf("              %d.%d. %s (pos: %.1f)", k+1, l+1, item.Title, item.Position)
									if item.DuedIn != nil {
										fmt.Printf(" [due: %d days]", *item.DuedIn)
									}
									if len(item.AssigneeIds) > 0 {
										fmt.Printf(" [assignees: %v]", item.AssigneeIds)
									}
									fmt.Printf("\n")
								}
							}
						}
						if len(action.Metadata.CopyTodoOptions) > 0 {
							fmt.Printf("         Copy Todo Options: %v\n", action.Metadata.CopyTodoOptions)
						}
						if action.Metadata.Email != nil {
							email := action.Metadata.Email
							fmt.Printf("         Email:\n")
							if email.From != nil {
								fmt.Printf("           From: %s\n", *email.From)
							}
							fmt.Printf("           To: %v\n", email.To)
							if len(email.Cc) > 0 {
								fmt.Printf("           Cc: %v\n", email.Cc)
							}
							if len(email.Bcc) > 0 {
								fmt.Printf("           Bcc: %v\n", email.Bcc)
							}
							fmt.Printf("           Subject: %s\n", email.Subject)
							fmt.Printf("           Content: %s\n", email.Content)
							if len(email.Attachments) > 0 {
								fmt.Printf("           Attachments:\n")
								for _, attachment := range email.Attachments {
									fmt.Printf("             - %s (%s, %.2f bytes) [%s]\n",
										attachment.Name, attachment.Type, attachment.Size, attachment.UID)
								}
							}
						}
					}

					if action.CustomField != nil {
						fmt.Printf("         Custom Field: %s (%s)\n", action.CustomField.Name, action.CustomField.ID)
					}
					if len(action.CustomFieldOptions) > 0 {
						fmt.Printf("         Custom Field Options:\n")
						for _, option := range action.CustomFieldOptions {
							fmt.Printf("           - %s (%s) [%s]\n", option.Title, option.ID, option.Color)
						}
					}
					if action.TodoList != nil {
						fmt.Printf("         List: %s (%s)\n", action.TodoList.Title, action.TodoList.ID)
					}
					if len(action.Tags) > 0 {
						fmt.Printf("         Tags:\n")
						for _, tag := range action.Tags {
							fmt.Printf("           - %s (%s) [%s]\n", tag.Title, tag.ID, tag.Color)
						}
					}
					if len(action.Assignees) > 0 {
						fmt.Printf("         Assignees:\n")
						for _, assignee := range action.Assignees {
							fmt.Printf("           - %s (%s)\n", assignee.FullName, assignee.ID)
						}
					}
					if action.Color != nil {
						fmt.Printf("         Color: %s\n", *action.Color)
					}
					if action.AssigneeTriggerer != nil {
						fmt.Printf("         Assignee Triggerer: %s\n", *action.AssigneeTriggerer)
					}
					if action.HttpOption != nil {
						http := action.HttpOption
						fmt.Printf("         HTTP Webhook:\n")
						fmt.Printf("           ID: %s\n", http.ID)
						fmt.Printf("           UID: %s\n", http.UID)
						fmt.Printf("           URL: %s\n", http.URL)
						fmt.Printf("           Method: %s\n", http.Method)
						if len(http.Headers) > 0 {
							fmt.Printf("           Headers:\n")
							for _, header := range http.Headers {
								fmt.Printf("             %s: %s\n", header.Key, header.Value)
							}
						}
						if len(http.Parameters) > 0 {
							fmt.Printf("           Parameters:\n")
							for _, param := range http.Parameters {
								fmt.Printf("             %s: %s\n", param.Key, param.Value)
							}
						}
						if http.Body != nil {
							fmt.Printf("           Body: %s\n", *http.Body)
						}
						if http.ContentType != nil {
							fmt.Printf("           Content Type: %s\n", *http.ContentType)
						}
						if http.AuthorizationType != nil {
							fmt.Printf("           Authorization Type: %s\n", *http.AuthorizationType)
						}
					}
				}
			}
			fmt.Printf("\n")
		}
	}

	if totalCount == 0 {
		fmt.Printf("No automations found in this workspace.\n")
	} else if pageInfo != nil {
		fmt.Printf("---\n")
		fmt.Printf("Pagination: Total items: %d | Page %d of %d | Items shown: %d-%d\n",
			pageInfo.TotalItems, pageInfo.CurrentPage, pageInfo.TotalPages,
			skipValue+1, minInt(skipValue+len(automations), totalCount))
		if pageInfo.HasPrevious {
			fmt.Printf("  Previous page available\n")
		}
		if pageInfo.HasNext {
			fmt.Printf("  Next page available\n")
		}
	}

	return nil
}
