package tools

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	. "demo-builder/common"
)

// Response structures for automation update
type UpdateAutomationResponse struct {
	EditAutomation CreatedAutomation `json:"editAutomation"`
}

// Input structure matching the GraphQL schema
type EditAutomationInput struct {
	AutomationID string                          `json:"automationId"`
	Trigger      *CreateAutomationTriggerInput   `json:"trigger,omitempty"`
	Actions      []CreateAutomationActionInput   `json:"actions,omitempty"`
	IsActive     *bool                           `json:"isActive,omitempty"`
}

// Execute GraphQL mutation
func executeUpdateAutomation(client *Client, input EditAutomationInput) (*CreatedAutomation, error) {
	// Use the same fragments as create automation for consistency
	mutation := `
		mutation EditAutomation($input: EditAutomationInput!) {
			editAutomation(input: $input) {
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
	var response UpdateAutomationResponse
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

	return &response.EditAutomation, nil
}

// Command-line interface
func RunUpdateAutomation(args []string) error {
	fs := flag.NewFlagSet("update-automation", flag.ExitOnError)
	
	automationID := fs.String("automation", "", "Automation ID (required)")
	projectID := fs.String("project", "", "Project ID or slug (required for project context)")
	simple := fs.Bool("simple", false, "Simple output format")
	isActive := fs.String("active", "", "Set automation active status (true/false)")
	
	// Trigger options
	triggerType := fs.String("trigger-type", "", "Update trigger type")
	triggerTodoList := fs.String("trigger-todo-list", "", "Todo list ID for trigger")
	triggerTags := fs.String("trigger-tags", "", "Comma-separated tag IDs")
	triggerAssignees := fs.String("trigger-assignees", "", "Comma-separated assignee IDs")
	triggerColor := fs.String("trigger-color", "", "Trigger color")
	triggerIncompleteOnly := fs.String("trigger-incomplete-only", "", "Only trigger for incomplete todos (true/false)")
	
	// Action options
	actionType := fs.String("action-type", "", "Update action type")
	actionDueIn := fs.Int("action-due-in", -1, "Due in days for action (-1 to keep current)")
	actionColor := fs.String("action-color", "", "Action color")
	actionTodoList := fs.String("action-todo-list", "", "Todo list ID for action")
	actionTags := fs.String("action-tags", "", "Comma-separated tag IDs")
	actionAssignees := fs.String("action-assignees", "", "Comma-separated assignee IDs")
	
	// Email options (for SEND_EMAIL actions)
	emailFrom := fs.String("email-from", "", "Email from address")
	emailTo := fs.String("email-to", "", "Comma-separated email addresses")
	emailSubject := fs.String("email-subject", "", "Email subject")
	emailContent := fs.String("email-content", "", "Email content (HTML)")
	
	// HTTP options (for MAKE_HTTP_REQUEST actions)
	httpURL := fs.String("http-url", "", "HTTP request URL")
	httpMethod := fs.String("http-method", "", "HTTP method (GET, POST, PUT, DELETE)")
	httpContentType := fs.String("http-content-type", "", "HTTP content type")
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
	if *automationID == "" {
		return fmt.Errorf("automation ID is required")
	}
	if *projectID == "" {
		return fmt.Errorf("project ID is required for project context")
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Create client
	client := NewClient(config)
	client.SetProject(*projectID)

	// Build update input
	input := EditAutomationInput{
		AutomationID: *automationID,
	}

	// Set active status if provided
	if *isActive != "" {
		if *isActive == "true" {
			active := true
			input.IsActive = &active
		} else if *isActive == "false" {
			active := false
			input.IsActive = &active
		} else {
			return fmt.Errorf("active must be 'true' or 'false', got: %s", *isActive)
		}
	}

	// Build trigger input if any trigger options are provided
	triggerProvided := *triggerType != "" || *triggerTodoList != "" || *triggerTags != "" || 
		*triggerAssignees != "" || *triggerColor != "" || *triggerIncompleteOnly != ""

	if triggerProvided {
		trigger := &CreateAutomationTriggerInput{
			TodoListID: *triggerTodoList,
			Metadata: nil,
		}

		if *triggerType != "" {
			trigger.Type = *triggerType
		}
		if *triggerColor != "" {
			trigger.Color = triggerColor
		}
		if *triggerIncompleteOnly != "" {
			if *triggerIncompleteOnly == "true" {
				incompleteOnly := true
				trigger.Metadata = &AutomationTriggerMetadataInput{
					IncompleteOnly: &incompleteOnly,
				}
			} else if *triggerIncompleteOnly == "false" {
				incompleteOnly := false
				trigger.Metadata = &AutomationTriggerMetadataInput{
					IncompleteOnly: &incompleteOnly,
				}
			} else {
				return fmt.Errorf("trigger-incomplete-only must be 'true' or 'false', got: %s", *triggerIncompleteOnly)
			}
		}
		if *triggerTags != "" {
			trigger.TagIDs = strings.Split(*triggerTags, ",")
		}
		if *triggerAssignees != "" {
			trigger.AssigneeIDs = strings.Split(*triggerAssignees, ",")
		}

		input.Trigger = trigger
	}

	// Build action input if any action options are provided
	actionProvided := *actionType != "" || *actionDueIn != -1 || *actionColor != "" || 
		*actionTodoList != "" || *actionTags != "" || *actionAssignees != "" ||
		*emailFrom != "" || *emailTo != "" || *emailSubject != "" || *emailContent != "" ||
		*httpURL != "" || *httpMethod != "" || *httpContentType != "" || *httpBody != "" ||
		*httpHeaders != "" || *httpParams != "" || *httpAuthType != "" || *httpAuthValue != ""

	if actionProvided {
		action := CreateAutomationActionInput{
			TodoListID: *actionTodoList,
			Metadata: nil,
			HttpOption: nil,
		}

		if *actionType != "" {
			action.Type = *actionType
		}
		if *actionDueIn != -1 {
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

		// Handle SEND_EMAIL action updates
		emailProvided := *emailFrom != "" || *emailTo != "" || *emailSubject != "" || *emailContent != ""
		if emailProvided {
			emailMetadata := &AutomationEmailInput{
				Cc: []string{},
				Bcc: []string{},
				ReplyTo: []string{},
				Attachments: []AutomationEmailAttachmentInput{},
			}

			if *emailFrom != "" {
				emailMetadata.From = *emailFrom
			}
			if *emailTo != "" {
				emailMetadata.To = strings.Split(*emailTo, ",")
			}
			if *emailSubject != "" {
				emailMetadata.Subject = *emailSubject
			}
			if *emailContent != "" {
				emailMetadata.Content = *emailContent
			}

			action.Metadata = &AutomationActionMetadataInput{
				Email: emailMetadata,
			}
		}

		// Handle MAKE_HTTP_REQUEST action updates
		httpProvided := *httpURL != "" || *httpMethod != "" || *httpContentType != "" || 
			*httpBody != "" || *httpHeaders != "" || *httpParams != "" || 
			*httpAuthType != "" || *httpAuthValue != ""

		if httpProvided {
			httpOption := &HttpOptionInput{}

			if *httpURL != "" {
				httpOption.URL = *httpURL
			}
			if *httpMethod != "" {
				httpOption.Method = *httpMethod
			}
			if *httpContentType != "" {
				httpOption.ContentType = *httpContentType
			}
			if *httpBody != "" {
				httpOption.Body = *httpBody
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
					if *httpAuthValue != "" {
						httpOption.AuthorizationBearerToken = *httpAuthValue
					}
				case "API_KEY":
					if *httpAuthValue != "" {
						httpOption.AuthorizationApiKey = &HttpApiKeyInput{
							Key: "Authorization",
							Value: *httpAuthValue,
							PassBy: "HEADER",
						}
					}
				case "BASIC_AUTH":
					if *httpAuthValue != "" {
						parts := strings.SplitN(*httpAuthValue, ":", 2)
						if len(parts) == 2 {
							httpOption.AuthorizationBasicAuth = &HttpBasicAuthInput{
								Username: parts[0],
								Password: parts[1],
							}
						}
					}
				}
			}

			action.HttpOption = httpOption
		}

		input.Actions = []CreateAutomationActionInput{action}
	}

	// Validate that at least one field is being updated
	if input.Trigger == nil && input.Actions == nil && input.IsActive == nil {
		return fmt.Errorf("at least one field must be provided to update (trigger, action, or active status)")
	}

	// Execute update
	automation, err := executeUpdateAutomation(client, input)
	if err != nil {
		return fmt.Errorf("failed to update automation: %v", err)
	}

	// Output results
	if *simple {
		fmt.Printf("Updated automation: %s\n", automation.ID)
		fmt.Printf("Active: %t\n", automation.IsActive)
		fmt.Printf("Trigger: %s\n", automation.Trigger.Type)
		for i, action := range automation.Actions {
			fmt.Printf("Action %d: %s\n", i+1, action.Type)
		}
	} else {
		fmt.Printf("✅ Successfully updated automation\n\n")
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