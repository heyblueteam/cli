package tools

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	. "cli/common"
)

// Response structures for automation creation
type CreatedAutomation struct {
	ID       string                     `json:"id"`
	IsActive bool                       `json:"isActive"`
	Trigger  AutomationTriggerResponse  `json:"trigger"`
	Actions  []AutomationActionResponse `json:"actions"`
	CreatedAt string                    `json:"createdAt"`
	UpdatedAt string                    `json:"updatedAt"`
}

type AutomationTriggerResponse struct {
	ID   string `json:"id"`
	Type string `json:"type"`
}

type AutomationActionResponse struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	DuedIn *int   `json:"duedIn,omitempty"`
}

type CreateAutomationResponse struct {
	CreateAutomation CreatedAutomation `json:"createAutomation"`
}

// Input structures matching the GraphQL schema from working examples
type CreateAutomationInput struct {
	Trigger CreateAutomationTriggerInput  `json:"trigger"`
	Actions []CreateAutomationActionInput `json:"actions"`
}

type CreateAutomationTriggerInput struct {
	Type                  string                             `json:"type"`
	TodoListID            string                             `json:"todoListId,omitempty"`
	Color                 *string                            `json:"color,omitempty"`
	Metadata              *AutomationTriggerMetadataInput    `json:"metadata"`
	TagIDs                []string                           `json:"tagIds,omitempty"`
	AssigneeIDs           []string                           `json:"assigneeIds,omitempty"`
	CustomFieldOptionIDs  []string                           `json:"customFieldOptionIds,omitempty"`
	TodoIDs               []string                           `json:"todoIds,omitempty"`
}

type AutomationTriggerMetadataInput struct {
	IncompleteOnly *bool `json:"incompleteOnly,omitempty"`
}

type CreateAutomationActionInput struct {
	Type                  string                           `json:"type"`
	DuedIn                *int                             `json:"duedIn,omitempty"`
	Color                 *string                          `json:"color,omitempty"`
	AssigneeTriggerer     *string                          `json:"assigneeTriggerer,omitempty"`
	TodoListID            string                           `json:"todoListId,omitempty"`
	TagIDs                []string                         `json:"tagIds,omitempty"`
	AssigneeIDs           []string                         `json:"assigneeIds,omitempty"`
	CustomFieldOptionIDs  []string                         `json:"customFieldOptionIds,omitempty"`
	Metadata              *AutomationActionMetadataInput   `json:"metadata"`
	HttpOption            *HttpOptionInput                 `json:"httpOption"`
}

type AutomationActionMetadataInput struct {
	CopyTodoOptions []string                          `json:"copyTodoOptions,omitempty"`
	Email           *AutomationEmailInput             `json:"email,omitempty"`
	Checklists      []AutomationChecklistInput        `json:"checklists,omitempty"`
}

type AutomationEmailInput struct {
	From        string                            `json:"from"`
	To          []string                          `json:"to"`
	Cc          []string                          `json:"cc,omitempty"`
	Bcc         []string                          `json:"bcc,omitempty"`
	ReplyTo     []string                          `json:"replyTo,omitempty"`
	Subject     string                            `json:"subject"`
	Content     string                            `json:"content"`
	Attachments []AutomationEmailAttachmentInput  `json:"attachments,omitempty"`
}

type AutomationEmailAttachmentInput struct {
	UID       string  `json:"uid"`
	Name      string  `json:"name"`
	Size      float64 `json:"size"`
	Type      string  `json:"type"`
	Extension string  `json:"extension"`
}

type AutomationChecklistInput struct {
	Title          string                            `json:"title"`
	Position       float64                           `json:"position"`
	ChecklistItems []AutomationChecklistItemInput    `json:"checklistItems,omitempty"`
}

type AutomationChecklistItemInput struct {
	Title       string   `json:"title"`
	Position    float64  `json:"position"`
	DuedIn      *int     `json:"duedIn,omitempty"`
	AssigneeIds []string `json:"assigneeIds,omitempty"`
}

type HttpOptionInput struct {
	URL                      string                         `json:"url"`
	Method                   string                         `json:"method"`
	ContentType              string                         `json:"contentType"`
	Headers                  []HttpHeaderInput              `json:"headers,omitempty"`
	Parameters               []HttpParameterInput           `json:"parameters,omitempty"`
	Body                     string                         `json:"body,omitempty"`
	AuthorizationType        string                         `json:"authorizationType,omitempty"`
	AuthorizationBearerToken string                         `json:"authorizationBearerToken,omitempty"`
	AuthorizationBasicAuth   *HttpBasicAuthInput            `json:"authorizationBasicAuth,omitempty"`
	AuthorizationApiKey      *HttpApiKeyInput               `json:"authorizationApiKey,omitempty"`
}

type HttpHeaderInput struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type HttpParameterInput struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type HttpBasicAuthInput struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type HttpApiKeyInput struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	PassBy string `json:"passBy"` // "HEADER" or "PARAMETER"
}

// Execute GraphQL mutation using the exact structure from working examples
func executeCreateAutomation(client *Client, input CreateAutomationInput) (*CreatedAutomation, error) {
	// Use the complete GraphQL fragments from working examples
	mutation := `
		mutation CreateAutomation($input: CreateAutomationInput!) {
			createAutomation(input: $input) {
				...AutomationFields
				trigger {
					...AutomationTriggerFields
					__typename
				}
				actions {
					...AutomationActionFields
					__typename
				}
				__typename
			}
		}

		fragment AutomationFields on Automation {
			id
			isActive
			updatedAt
			createdAt
			__typename
		}

		fragment AutomationTriggerFields on AutomationTrigger {
			id
			type
			metadata {
				... on AutomationTriggerMetadataTodoOverdue {
					incompleteOnly
					__typename
				}
				__typename
			}
			color
			customField {
				id
				name
				type
				referenceProject {
					id
					__typename
				}
				__typename
			}
			customFieldOptions {
				id
				title
				__typename
			}
			todoList {
				id
				title
				__typename
			}
			assignees {
				id
				fullName
				image {
					id
					thumbnail
					__typename
				}
				__typename
			}
			tags {
				id
				title
				color
				__typename
			}
			todos {
				id
				title
				__typename
			}
			__typename
		}

		fragment AutomationActionFields on AutomationAction {
			id
			type
			duedIn
			color
			assigneeTriggerer
			portableDocument {
				id
				name
				__typename
			}
			customField {
				id
				name
				type
				__typename
			}
			customFieldOptions {
				id
				title
				__typename
			}
			todoList {
				id
				title
				project {
					id
					name
					__typename
				}
				__typename
			}
			metadata {
				... on AutomationActionMetadataCopyTodo {
					copyTodoOptions
					__typename
				}
				... on AutomationActionMetadataCreateChecklist {
					checklists {
						title
						position
						checklistItems {
							title
							position
							duedIn
							assigneeIds
							__typename
						}
						__typename
					}
					__typename
				}
				... on AutomationActionMetadataSendEmail {
					email {
						from
						to
						bcc
						cc
						content
						subject
						replyTo
						attachments {
							uid
							name
							size
							type
							extension
							__typename
						}
						__typename
					}
					__typename
				}
				__typename
			}
			assignees {
				id
				fullName
				image {
					id
					thumbnail
					__typename
				}
				__typename
			}
			tags {
				id
				title
				color
				__typename
			}
			httpOption {
				url
				method
				contentType
				headers {
					key
					value
					__typename
				}
				parameters {
					key
					value
					__typename
				}
				authorizationType
				authorizationBearerToken
				authorizationBasicAuth {
					username
					password
					__typename
				}
				authorizationApiKey {
					key
					value
					passBy
					__typename
				}
				body
				__typename
			}
			__typename
		}
	`

	variables := map[string]interface{}{
		"input": input,
	}

	// Execute mutation
	var response CreateAutomationResponse
	result, err := client.ExecuteQuery(mutation, variables)
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

	return &response.CreateAutomation, nil
}

// Command-line interface
func RunCreateAutomation(args []string) error {
	fs := flag.NewFlagSet("create-automation", flag.ExitOnError)
	
	projectID := fs.String("project", "", "Project ID or slug (required)")
	simple := fs.Bool("simple", false, "Simple output format")
	
	// Trigger options
	triggerType := fs.String("trigger-type", "", "Trigger type (required)")
	triggerTodoList := fs.String("trigger-todo-list", "", "Todo list ID for trigger")
	triggerTags := fs.String("trigger-tags", "", "Comma-separated tag IDs")
	triggerAssignees := fs.String("trigger-assignees", "", "Comma-separated assignee IDs")
	triggerColor := fs.String("trigger-color", "", "Trigger color")
	triggerIncompleteOnly := fs.Bool("trigger-incomplete-only", false, "Only trigger for incomplete todos")
	
	// Action options
	actionType := fs.String("action-type", "", "Action type (required)")
	actionDueIn := fs.Int("action-due-in", 0, "Due in days for action")
	actionColor := fs.String("action-color", "", "Action color")
	actionTodoList := fs.String("action-todo-list", "", "Todo list ID for action")
	actionTags := fs.String("action-tags", "", "Comma-separated tag IDs")
	actionAssignees := fs.String("action-assignees", "", "Comma-separated assignee IDs")
	
	// Email options (for SEND_EMAIL actions)
	emailFrom := fs.String("email-from", "<p>Blue</p>", "Email from address")
	emailTo := fs.String("email-to", "", "Comma-separated email addresses")
	emailSubject := fs.String("email-subject", "", "Email subject")
	emailContent := fs.String("email-content", "", "Email content (HTML)")
	
	// HTTP options (for MAKE_HTTP_REQUEST actions)
	httpURL := fs.String("http-url", "", "HTTP request URL")
	httpMethod := fs.String("http-method", "GET", "HTTP method (GET, POST, PUT, DELETE)")
	httpContentType := fs.String("http-content-type", "JSON", "HTTP content type")
	httpBody := fs.String("http-body", "", "HTTP request body")
	httpHeaders := fs.String("http-headers", "", "HTTP headers (key1:value1,key2:value2)")
	httpParams := fs.String("http-params", "", "HTTP parameters (key1:value1,key2:value2)")
	httpAuthType := fs.String("http-auth-type", "", "Authorization type (API_KEY, BEARER_TOKEN, BASIC_AUTH)")
	httpAuthValue := fs.String("http-auth-value", "", "Authorization value")

	err := fs.Parse(args)
	if err != nil {
		return err
	}

	// Validate required fields
	if *projectID == "" {
		return fmt.Errorf("project ID is required")
	}
	if *triggerType == "" {
		return fmt.Errorf("trigger type is required")
	}
	if *actionType == "" {
		return fmt.Errorf("action type is required")
	}

	// Validate action-specific requirements
	if *actionType == "SEND_EMAIL" && *emailTo == "" {
		return fmt.Errorf("email-to is required for SEND_EMAIL actions")
	}
	if *actionType == "MAKE_HTTP_REQUEST" && *httpURL == "" {
		return fmt.Errorf("http-url is required for MAKE_HTTP_REQUEST actions")
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Create client
	client := NewClient(config)
	client.SetProject(*projectID)

	// Build trigger input
	trigger := CreateAutomationTriggerInput{
		Type: *triggerType,
		TodoListID: *triggerTodoList,
		Metadata: nil,
	}

	if *triggerColor != "" {
		trigger.Color = triggerColor
	}
	if *triggerIncompleteOnly {
		trigger.Metadata = &AutomationTriggerMetadataInput{
			IncompleteOnly: triggerIncompleteOnly,
		}
	}
	if *triggerTags != "" {
		trigger.TagIDs = strings.Split(*triggerTags, ",")
	}
	if *triggerAssignees != "" {
		trigger.AssigneeIDs = strings.Split(*triggerAssignees, ",")
	}

	// Build action input
	action := CreateAutomationActionInput{
		Type: *actionType,
		TodoListID: *actionTodoList,
		Metadata: nil,
		HttpOption: nil,
	}

	if *actionDueIn > 0 {
		action.DuedIn = actionDueIn
	}
	if *actionColor != "" {
		action.Color = actionColor
	}
	if *actionTags != "" {
		action.TagIDs = strings.Split(*actionTags, ",")
	}
	if *actionAssignees != "" {
		action.AssigneeIDs = strings.Split(*actionAssignees, ",")
	}

	// Handle SEND_EMAIL action
	if *actionType == "SEND_EMAIL" {
		emailMetadata := &AutomationEmailInput{
			From: *emailFrom,
			To: strings.Split(*emailTo, ","),
			Subject: *emailSubject,
			Content: *emailContent,
			Cc: []string{},
			Bcc: []string{},
			ReplyTo: []string{},
			Attachments: []AutomationEmailAttachmentInput{},
		}
		action.Metadata = &AutomationActionMetadataInput{
			Email: emailMetadata,
		}
	}

	// Handle MAKE_HTTP_REQUEST action
	if *actionType == "MAKE_HTTP_REQUEST" {
		httpOption := &HttpOptionInput{
			URL: *httpURL,
			Method: *httpMethod,
			ContentType: *httpContentType,
			Body: *httpBody,
		}

		// Parse headers
		if *httpHeaders != "" {
			pairs := strings.Split(*httpHeaders, ",")
			for _, pair := range pairs {
				parts := strings.SplitN(pair, ":", 2)
				if len(parts) == 2 {
					httpOption.Headers = append(httpOption.Headers, HttpHeaderInput{
						Key: strings.TrimSpace(parts[0]),
						Value: strings.TrimSpace(parts[1]),
					})
				}
			}
		}

		// Parse parameters
		if *httpParams != "" {
			pairs := strings.Split(*httpParams, ",")
			for _, pair := range pairs {
				parts := strings.SplitN(pair, ":", 2)
				if len(parts) == 2 {
					httpOption.Parameters = append(httpOption.Parameters, HttpParameterInput{
						Key: strings.TrimSpace(parts[0]),
						Value: strings.TrimSpace(parts[1]),
					})
				}
			}
		}

		// Handle authentication
		if *httpAuthType != "" {
			httpOption.AuthorizationType = *httpAuthType
			switch *httpAuthType {
			case "BEARER_TOKEN":
				httpOption.AuthorizationBearerToken = *httpAuthValue
			case "API_KEY":
				httpOption.AuthorizationApiKey = &HttpApiKeyInput{
					Key: "Authorization",
					Value: *httpAuthValue,
					PassBy: "HEADER",
				}
			case "BASIC_AUTH":
				parts := strings.SplitN(*httpAuthValue, ":", 2)
				if len(parts) == 2 {
					httpOption.AuthorizationBasicAuth = &HttpBasicAuthInput{
						Username: parts[0],
						Password: parts[1],
					}
				}
			}
		}

		action.HttpOption = httpOption
	}

	// Create automation input
	input := CreateAutomationInput{
		Trigger: trigger,
		Actions: []CreateAutomationActionInput{action},
	}

	// Execute creation
	automation, err := executeCreateAutomation(client, input)
	if err != nil {
		return fmt.Errorf("failed to create automation: %v", err)
	}

	// Output results
	if *simple {
		fmt.Printf("Created automation: %s\n", automation.ID)
		fmt.Printf("Trigger: %s\n", automation.Trigger.Type)
		for i, action := range automation.Actions {
			fmt.Printf("Action %d: %s\n", i+1, action.Type)
		}
	} else {
		fmt.Printf("✅ Successfully created automation\n\n")
		fmt.Printf("Automation Details:\n")
		fmt.Printf("  ID: %s\n", automation.ID)
		fmt.Printf("  Active: %t\n", automation.IsActive)
		fmt.Printf("  Created: %s\n", automation.CreatedAt)
		fmt.Printf("  Updated: %s\n\n", automation.UpdatedAt)
		
		fmt.Printf("Trigger:\n")
		fmt.Printf("  Type: %s\n", automation.Trigger.Type)
		fmt.Printf("  ID: %s\n", automation.Trigger.ID)
		
		fmt.Printf("\nActions:\n")
		for i, action := range automation.Actions {
			fmt.Printf("  %d. Type: %s\n", i+1, action.Type)
			fmt.Printf("     ID: %s\n", action.ID)
			if action.DuedIn != nil {
				fmt.Printf("     Due In: %d days\n", *action.DuedIn)
			}
		}
	}

	return nil
}

// Helper function to print usage examples
func printAutomationExamples() {
	fmt.Println("Examples:")
	fmt.Println()
	fmt.Println("Simple email automation:")
	fmt.Println("  go run . create-automation -project PROJECT_ID \\")
	fmt.Println("    -trigger-type \"TODO_MARKED_AS_COMPLETE\" \\")
	fmt.Println("    -action-type \"SEND_EMAIL\" \\")
	fmt.Println("    -email-to \"user@example.com\" \\")
	fmt.Println("    -email-subject \"Task completed\" \\")
	fmt.Println("    -email-content \"<p>Task has been completed!</p>\"")
	fmt.Println()
	fmt.Println("HTTP webhook on tag added:")
	fmt.Println("  go run . create-automation -project PROJECT_ID \\")
	fmt.Println("    -trigger-type \"TAG_ADDED\" -trigger-tags \"tag_id\" \\")
	fmt.Println("    -action-type \"MAKE_HTTP_REQUEST\" \\")
	fmt.Println("    -http-url \"https://example.com/webhook\" \\")
	fmt.Println("    -http-method \"POST\" \\")
	fmt.Println("    -http-body '{\"event\": \"tag_added\"}'")
	fmt.Println()
	fmt.Println("Add tag and assign user:")
	fmt.Println("  go run . create-automation -project PROJECT_ID \\")
	fmt.Println("    -trigger-type \"TODO_CREATED\" -trigger-todo-list \"list_id\" \\")
	fmt.Println("    -action-type \"ADD_TAG\" \\")
	fmt.Println("    -action-tags \"tag_id\"")
}