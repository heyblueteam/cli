package tools

import (
	"flag"
	"fmt"
	"strings"

	. "demo-builder/common"
)

// Enhanced multi-action automation update
func RunUpdateAutomationMulti(args []string) error {
	fs := flag.NewFlagSet("update-automation-multi", flag.ExitOnError)
	
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
	
	// Multiple action support - numbered and unnumbered flags
	// Unnumbered flags for single action convenience
	actionType := fs.String("action-type", "", "Action type (same as action1-type)")
	actionColor := fs.String("action-color", "", "Action color (same as action1-color)")
	actionTodoList := fs.String("action-todo-list", "", "Action todo list ID (same as action1-todo-list)")
	actionTags := fs.String("action-tags", "", "Action comma-separated tag IDs (same as action1-tags)")
	actionAssignees := fs.String("action-assignees", "", "Action comma-separated assignee IDs (same as action1-assignees)")
	actionDueIn := fs.Int("action-due-in", -1, "Due in days for action (-1 to keep current)")
	
	// Numbered action flags
	action1Type := fs.String("action1-type", "", "First action type")
	action1Color := fs.String("action1-color", "", "First action color")
	action1TodoList := fs.String("action1-todo-list", "", "First action todo list ID")
	action1Tags := fs.String("action1-tags", "", "First action comma-separated tag IDs")
	action1Assignees := fs.String("action1-assignees", "", "First action comma-separated assignee IDs")
	action1DueIn := fs.Int("action1-due-in", -1, "First action due in days (-1 to keep current)")
	
	action2Type := fs.String("action2-type", "", "Second action type")
	action2Color := fs.String("action2-color", "", "Second action color") 
	action2TodoList := fs.String("action2-todo-list", "", "Second action todo list ID")
	action2Tags := fs.String("action2-tags", "", "Second action comma-separated tag IDs")
	action2Assignees := fs.String("action2-assignees", "", "Second action comma-separated assignee IDs")
	action2DueIn := fs.Int("action2-due-in", -1, "Second action due in days (-1 to keep current)")
	
	action3Type := fs.String("action3-type", "", "Third action type")
	action3Color := fs.String("action3-color", "", "Third action color")
	action3TodoList := fs.String("action3-todo-list", "", "Third action todo list ID")
	action3Tags := fs.String("action3-tags", "", "Third action comma-separated tag IDs")
	action3Assignees := fs.String("action3-assignees", "", "Third action comma-separated assignee IDs")
	action3DueIn := fs.Int("action3-due-in", -1, "Third action due in days (-1 to keep current)")
	
	// Per-action email options
	// Unnumbered email for single action convenience  
	emailFrom := fs.String("email-from", "", "Email from address (same as action1-email-from)")
	emailTo := fs.String("email-to", "", "Comma-separated email addresses (same as action1-email-to)")
	emailSubject := fs.String("email-subject", "", "Email subject (same as action1-email-subject)")
	emailContent := fs.String("email-content", "", "Email content HTML (same as action1-email-content)")
	
	// Action1 email options
	action1EmailFrom := fs.String("action1-email-from", "", "First action email from address")
	action1EmailTo := fs.String("action1-email-to", "", "First action comma-separated email addresses")
	action1EmailSubject := fs.String("action1-email-subject", "", "First action email subject")
	action1EmailContent := fs.String("action1-email-content", "", "First action email content HTML")
	
	// Action2 email options
	action2EmailFrom := fs.String("action2-email-from", "", "Second action email from address")
	action2EmailTo := fs.String("action2-email-to", "", "Second action comma-separated email addresses")
	action2EmailSubject := fs.String("action2-email-subject", "", "Second action email subject")
	action2EmailContent := fs.String("action2-email-content", "", "Second action email content HTML")
	
	// Action3 email options
	action3EmailFrom := fs.String("action3-email-from", "", "Third action email from address")
	action3EmailTo := fs.String("action3-email-to", "", "Third action comma-separated email addresses")
	action3EmailSubject := fs.String("action3-email-subject", "", "Third action email subject")
	action3EmailContent := fs.String("action3-email-content", "", "Third action email content HTML")
	
	// Per-action HTTP options
	// Unnumbered HTTP for single action convenience
	httpURL := fs.String("http-url", "", "HTTP request URL (same as action1-http-url)")
	httpMethod := fs.String("http-method", "", "HTTP method (same as action1-http-method)")
	httpContentType := fs.String("http-content-type", "", "HTTP content type (same as action1-http-content-type)")
	httpBody := fs.String("http-body", "", "HTTP request body (same as action1-http-body)")
	httpHeaders := fs.String("http-headers", "", "HTTP headers (same as action1-http-headers)")
	httpParams := fs.String("http-params", "", "HTTP parameters (same as action1-http-params)")
	httpAuthType := fs.String("http-auth-type", "", "Authorization type (same as action1-http-auth-type)")
	httpAuthValue := fs.String("http-auth-value", "", "Authorization value (same as action1-http-auth-value)")
	
	// Action1 HTTP options
	action1HttpURL := fs.String("action1-http-url", "", "First action HTTP request URL")
	action1HttpMethod := fs.String("action1-http-method", "", "First action HTTP method")
	action1HttpContentType := fs.String("action1-http-content-type", "", "First action HTTP content type")
	action1HttpBody := fs.String("action1-http-body", "", "First action HTTP request body")
	action1HttpHeaders := fs.String("action1-http-headers", "", "First action HTTP headers")
	action1HttpParams := fs.String("action1-http-params", "", "First action HTTP parameters")
	action1HttpAuthType := fs.String("action1-http-auth-type", "", "First action authorization type")
	action1HttpAuthValue := fs.String("action1-http-auth-value", "", "First action authorization value")
	
	// Action2 HTTP options
	action2HttpURL := fs.String("action2-http-url", "", "Second action HTTP request URL")
	action2HttpMethod := fs.String("action2-http-method", "", "Second action HTTP method")
	action2HttpContentType := fs.String("action2-http-content-type", "", "Second action HTTP content type")
	action2HttpBody := fs.String("action2-http-body", "", "Second action HTTP request body")
	action2HttpHeaders := fs.String("action2-http-headers", "", "Second action HTTP headers")
	action2HttpParams := fs.String("action2-http-params", "", "Second action HTTP parameters")
	action2HttpAuthType := fs.String("action2-http-auth-type", "", "Second action authorization type")
	action2HttpAuthValue := fs.String("action2-http-auth-value", "", "Second action authorization value")
	
	// Action3 HTTP options
	action3HttpURL := fs.String("action3-http-url", "", "Third action HTTP request URL")
	action3HttpMethod := fs.String("action3-http-method", "", "Third action HTTP method")
	action3HttpContentType := fs.String("action3-http-content-type", "", "Third action HTTP content type")
	action3HttpBody := fs.String("action3-http-body", "", "Third action HTTP request body")
	action3HttpHeaders := fs.String("action3-http-headers", "", "Third action HTTP headers")
	action3HttpParams := fs.String("action3-http-params", "", "Third action HTTP parameters")
	action3HttpAuthType := fs.String("action3-http-auth-type", "", "Third action authorization type")
	action3HttpAuthValue := fs.String("action3-http-auth-value", "", "Third action authorization value")

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

	// Helper function to create an action with per-action settings
	createAction := func(actionType, color, todoList, tags, assignees string, dueIn int, emailFrom, emailTo, emailSubject, emailContent, httpURL, httpMethod, httpContentType, httpBody, httpHeaders, httpParams, httpAuthType, httpAuthValue string) CreateAutomationActionInput {
		action := CreateAutomationActionInput{
			TodoListID: todoList,
		}

		if actionType != "" {
			action.Type = actionType
		}
		if color != "" {
			action.Color = &color
		}
		if tags != "" {
			action.TagIDs = strings.Split(tags, ",")
		}
		if assignees != "" {
			action.AssigneeIDs = strings.Split(assignees, ",")
		}
		if dueIn != -1 {
			action.DuedIn = &dueIn
		}

		// Handle SEND_EMAIL action with per-action email settings
		if actionType == "SEND_EMAIL" && emailTo != "" {
			emailMetadata := &AutomationEmailInput{
				From: emailFrom,
				To: strings.Split(emailTo, ","),
				Subject: emailSubject,
				Content: emailContent,
				Cc: []string{},
				Bcc: []string{},
				ReplyTo: []string{},
				Attachments: []AutomationEmailAttachmentInput{},
			}
			action.Metadata = &AutomationActionMetadataInput{
				Email: emailMetadata,
			}
		}

		// Handle MAKE_HTTP_REQUEST action with per-action HTTP settings
		if actionType == "MAKE_HTTP_REQUEST" && httpURL != "" {
			httpOption := &HttpOptionInput{
				URL: httpURL,
				Method: httpMethod,
				ContentType: httpContentType,
				Body: httpBody,
			}

			// Parse headers
			if httpHeaders != "" {
				pairs := strings.Split(httpHeaders, ",")
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
			if httpParams != "" {
				pairs := strings.Split(httpParams, ",")
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
			if httpAuthType != "" {
				httpOption.AuthorizationType = httpAuthType
				switch httpAuthType {
				case "BEARER_TOKEN":
					if httpAuthValue != "" {
						httpOption.AuthorizationBearerToken = httpAuthValue
					}
				case "API_KEY":
					if httpAuthValue != "" {
						httpOption.AuthorizationApiKey = &HttpApiKeyInput{
							Key: "Authorization",
							Value: httpAuthValue,
							PassBy: "HEADER",
						}
					}
				case "BASIC_AUTH":
					if httpAuthValue != "" {
						parts := strings.SplitN(httpAuthValue, ":", 2)
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

		return action
	}

	// Check if any action is being updated
	var actions []CreateAutomationActionInput

	// Priority: Use numbered flags first, fall back to unnumbered for action1
	// Action 1 (or unnumbered action)
	act1Type := *action1Type
	act1Color := *action1Color
	act1TodoList := *action1TodoList
	act1Tags := *action1Tags
	act1Assignees := *action1Assignees
	act1DueIn := *action1DueIn
	act1EmailFrom := *action1EmailFrom
	act1EmailTo := *action1EmailTo
	act1EmailSubject := *action1EmailSubject
	act1EmailContent := *action1EmailContent
	act1HttpURL := *action1HttpURL
	act1HttpMethod := *action1HttpMethod
	act1HttpContentType := *action1HttpContentType
	act1HttpBody := *action1HttpBody
	act1HttpHeaders := *action1HttpHeaders
	act1HttpParams := *action1HttpParams
	act1HttpAuthType := *action1HttpAuthType
	act1HttpAuthValue := *action1HttpAuthValue

	// If numbered flags are empty, use unnumbered flags for action1
	if act1Type == "" && *actionType != "" {
		act1Type = *actionType
	}
	if act1Color == "" && *actionColor != "" {
		act1Color = *actionColor
	}
	if act1TodoList == "" && *actionTodoList != "" {
		act1TodoList = *actionTodoList
	}
	if act1Tags == "" && *actionTags != "" {
		act1Tags = *actionTags
	}
	if act1Assignees == "" && *actionAssignees != "" {
		act1Assignees = *actionAssignees
	}
	if act1DueIn == -1 && *actionDueIn != -1 {
		act1DueIn = *actionDueIn
	}
	if act1EmailTo == "" && *emailTo != "" {
		if *emailFrom != "" {
			act1EmailFrom = *emailFrom
		}
		act1EmailTo = *emailTo
		if *emailSubject != "" {
			act1EmailSubject = *emailSubject
		}
		if *emailContent != "" {
			act1EmailContent = *emailContent
		}
	}
	if act1HttpURL == "" && *httpURL != "" {
		act1HttpURL = *httpURL
		if *httpMethod != "" {
			act1HttpMethod = *httpMethod
		}
		if *httpContentType != "" {
			act1HttpContentType = *httpContentType
		}
		if *httpBody != "" {
			act1HttpBody = *httpBody
		}
		if *httpHeaders != "" {
			act1HttpHeaders = *httpHeaders
		}
		if *httpParams != "" {
			act1HttpParams = *httpParams
		}
		if *httpAuthType != "" {
			act1HttpAuthType = *httpAuthType
		}
		if *httpAuthValue != "" {
			act1HttpAuthValue = *httpAuthValue
		}
	}

	// Check if any action1 field is provided
	action1Provided := act1Type != "" || act1Color != "" || act1TodoList != "" || act1Tags != "" || 
		act1Assignees != "" || act1DueIn != -1 || act1EmailTo != "" || act1HttpURL != ""

	if action1Provided {
		actions = append(actions, createAction(
			act1Type, act1Color, act1TodoList, act1Tags, act1Assignees, act1DueIn,
			act1EmailFrom, act1EmailTo, act1EmailSubject, act1EmailContent,
			act1HttpURL, act1HttpMethod, act1HttpContentType, act1HttpBody,
			act1HttpHeaders, act1HttpParams, act1HttpAuthType, act1HttpAuthValue,
		))
	}

	// Check if any action2 field is provided
	action2Provided := *action2Type != "" || *action2Color != "" || *action2TodoList != "" || 
		*action2Tags != "" || *action2Assignees != "" || *action2DueIn != -1 || 
		*action2EmailTo != "" || *action2HttpURL != ""

	if action2Provided {
		actions = append(actions, createAction(
			*action2Type, *action2Color, *action2TodoList, *action2Tags, *action2Assignees, *action2DueIn,
			*action2EmailFrom, *action2EmailTo, *action2EmailSubject, *action2EmailContent,
			*action2HttpURL, *action2HttpMethod, *action2HttpContentType, *action2HttpBody,
			*action2HttpHeaders, *action2HttpParams, *action2HttpAuthType, *action2HttpAuthValue,
		))
	}

	// Check if any action3 field is provided
	action3Provided := *action3Type != "" || *action3Color != "" || *action3TodoList != "" || 
		*action3Tags != "" || *action3Assignees != "" || *action3DueIn != -1 || 
		*action3EmailTo != "" || *action3HttpURL != ""

	if action3Provided {
		actions = append(actions, createAction(
			*action3Type, *action3Color, *action3TodoList, *action3Tags, *action3Assignees, *action3DueIn,
			*action3EmailFrom, *action3EmailTo, *action3EmailSubject, *action3EmailContent,
			*action3HttpURL, *action3HttpMethod, *action3HttpContentType, *action3HttpBody,
			*action3HttpHeaders, *action3HttpParams, *action3HttpAuthType, *action3HttpAuthValue,
		))
	}

	if len(actions) > 0 {
		input.Actions = actions
		
		// IMPORTANT WARNING: The EditAutomation API replaces the entire actions array
		// If the automation has multiple actions, you must specify ALL actions to avoid losing any
		fmt.Printf("⚠️  WARNING: Action update will REPLACE all actions. If this automation has multiple actions, specify ALL of them to avoid data loss.\n\n")
	}

	// Validate that at least one field is being updated
	if input.Trigger == nil && input.Actions == nil && input.IsActive == nil {
		return fmt.Errorf("at least one field must be provided to update (trigger, action, or active status)")
	}

	// Execute update using the existing function
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
		fmt.Printf("✅ Successfully updated multi-action automation\n\n")
		fmt.Printf("Automation Details:\n")
		fmt.Printf("  ID: %s\n", automation.ID)
		fmt.Printf("  Active: %t\n", automation.IsActive)
		fmt.Printf("  Created: %s\n", automation.CreatedAt)
		fmt.Printf("  Updated: %s\n\n", automation.UpdatedAt)
		
		fmt.Printf("Trigger:\n")
		fmt.Printf("  Type: %s\n", automation.Trigger.Type)
		fmt.Printf("  ID: %s\n", automation.Trigger.ID)
		
		fmt.Printf("\nActions (%d):\n", len(automation.Actions))
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