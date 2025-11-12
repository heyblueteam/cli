package tools

import (
	"flag"
	"fmt"
	"strings"

	. "cli/common"
)

// Multi-action automation creation
func RunCreateAutomationMulti(args []string) error {
	fs := flag.NewFlagSet("create-automation-multi", flag.ExitOnError)
	
	projectID := fs.String("project", "", "Project ID or slug (required)")
	simple := fs.Bool("simple", false, "Simple output format")
	
	// Trigger options
	triggerType := fs.String("trigger-type", "", "Trigger type (required)")
	triggerTodoList := fs.String("trigger-todo-list", "", "Todo list ID for trigger")
	triggerTags := fs.String("trigger-tags", "", "Comma-separated tag IDs")
	triggerAssignees := fs.String("trigger-assignees", "", "Comma-separated assignee IDs")
	triggerColor := fs.String("trigger-color", "", "Trigger color")
	triggerIncompleteOnly := fs.Bool("trigger-incomplete-only", false, "Only trigger for incomplete todos")
	
	// Multiple action support - numbered and unnumbered flags
	// Unnumbered flags for single action convenience
	actionType := fs.String("action-type", "", "Action type (same as action1-type)")
	actionColor := fs.String("action-color", "", "Action color (same as action1-color)")
	actionTodoList := fs.String("action-todo-list", "", "Action todo list ID (same as action1-todo-list)")
	actionTags := fs.String("action-tags", "", "Action comma-separated tag IDs (same as action1-tags)")
	actionAssignees := fs.String("action-assignees", "", "Action comma-separated assignee IDs (same as action1-assignees)")
	
	// Numbered action flags
	action1Type := fs.String("action1-type", "", "First action type")
	action1Color := fs.String("action1-color", "", "First action color")
	action1TodoList := fs.String("action1-todo-list", "", "First action todo list ID")
	action1Tags := fs.String("action1-tags", "", "First action comma-separated tag IDs")
	action1Assignees := fs.String("action1-assignees", "", "First action comma-separated assignee IDs")
	
	action2Type := fs.String("action2-type", "", "Second action type")
	action2Color := fs.String("action2-color", "", "Second action color") 
	action2TodoList := fs.String("action2-todo-list", "", "Second action todo list ID")
	action2Tags := fs.String("action2-tags", "", "Second action comma-separated tag IDs")
	action2Assignees := fs.String("action2-assignees", "", "Second action comma-separated assignee IDs")
	
	action3Type := fs.String("action3-type", "", "Third action type")
	action3Color := fs.String("action3-color", "", "Third action color")
	action3TodoList := fs.String("action3-todo-list", "", "Third action todo list ID")
	action3Tags := fs.String("action3-tags", "", "Third action comma-separated tag IDs")
	action3Assignees := fs.String("action3-assignees", "", "Third action comma-separated assignee IDs")
	
	// Per-action email options
	// Unnumbered email for single action convenience  
	emailFrom := fs.String("email-from", "<p>Blue</p>", "Email from address (same as action1-email-from)")
	emailTo := fs.String("email-to", "", "Comma-separated email addresses (same as action1-email-to)")
	emailSubject := fs.String("email-subject", "", "Email subject (same as action1-email-subject)")
	emailContent := fs.String("email-content", "", "Email content HTML (same as action1-email-content)")
	
	// Action1 email options
	action1EmailFrom := fs.String("action1-email-from", "<p>Blue</p>", "First action email from address")
	action1EmailTo := fs.String("action1-email-to", "", "First action comma-separated email addresses")
	action1EmailSubject := fs.String("action1-email-subject", "", "First action email subject")
	action1EmailContent := fs.String("action1-email-content", "", "First action email content HTML")
	
	// Action2 email options
	action2EmailFrom := fs.String("action2-email-from", "<p>Blue</p>", "Second action email from address")
	action2EmailTo := fs.String("action2-email-to", "", "Second action comma-separated email addresses")
	action2EmailSubject := fs.String("action2-email-subject", "", "Second action email subject")
	action2EmailContent := fs.String("action2-email-content", "", "Second action email content HTML")
	
	// Action3 email options
	action3EmailFrom := fs.String("action3-email-from", "<p>Blue</p>", "Third action email from address")
	action3EmailTo := fs.String("action3-email-to", "", "Third action comma-separated email addresses")
	action3EmailSubject := fs.String("action3-email-subject", "", "Third action email subject")
	action3EmailContent := fs.String("action3-email-content", "", "Third action email content HTML")
	
	// Per-action HTTP options
	// Unnumbered HTTP for single action convenience
	httpURL := fs.String("http-url", "", "HTTP request URL (same as action1-http-url)")
	httpMethod := fs.String("http-method", "GET", "HTTP method (same as action1-http-method)")
	httpContentType := fs.String("http-content-type", "JSON", "HTTP content type (same as action1-http-content-type)")
	httpBody := fs.String("http-body", "", "HTTP request body (same as action1-http-body)")
	
	// Action1 HTTP options
	action1HttpURL := fs.String("action1-http-url", "", "First action HTTP request URL")
	action1HttpMethod := fs.String("action1-http-method", "GET", "First action HTTP method")
	action1HttpContentType := fs.String("action1-http-content-type", "JSON", "First action HTTP content type")
	action1HttpBody := fs.String("action1-http-body", "", "First action HTTP request body")
	
	// Action2 HTTP options
	action2HttpURL := fs.String("action2-http-url", "", "Second action HTTP request URL")
	action2HttpMethod := fs.String("action2-http-method", "GET", "Second action HTTP method")
	action2HttpContentType := fs.String("action2-http-content-type", "JSON", "Second action HTTP content type")
	action2HttpBody := fs.String("action2-http-body", "", "Second action HTTP request body")
	
	// Action3 HTTP options
	action3HttpURL := fs.String("action3-http-url", "", "Third action HTTP request URL")
	action3HttpMethod := fs.String("action3-http-method", "GET", "Third action HTTP method")
	action3HttpContentType := fs.String("action3-http-content-type", "JSON", "Third action HTTP content type")
	action3HttpBody := fs.String("action3-http-body", "", "Third action HTTP request body")

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
	if *action1Type == "" && *actionType == "" {
		return fmt.Errorf("at least action1-type or action-type is required")
	}

	// Load configuration
	config, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Create client
	client := NewClient(config)
	client.SetProject(*projectID)

	// Build trigger
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

	// Build actions array
	var actions []CreateAutomationActionInput

	// Helper function to create an action with per-action email/HTTP settings
	createAction := func(actionType, color, todoList, tags, assignees, emailFrom, emailTo, emailSubject, emailContent, httpURL, httpMethod, httpContentType, httpBody string) CreateAutomationActionInput {
		action := CreateAutomationActionInput{
			Type: actionType,
			TodoListID: todoList,
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
			action.HttpOption = httpOption
		}

		return action
	}

	// Priority: Use numbered flags first, fall back to unnumbered for action1
	// Action 1 (or unnumbered action)
	act1Type := *action1Type
	act1Color := *action1Color
	act1TodoList := *action1TodoList
	act1Tags := *action1Tags
	act1Assignees := *action1Assignees
	act1EmailFrom := *action1EmailFrom
	act1EmailTo := *action1EmailTo
	act1EmailSubject := *action1EmailSubject
	act1EmailContent := *action1EmailContent
	act1HttpURL := *action1HttpURL
	act1HttpMethod := *action1HttpMethod
	act1HttpContentType := *action1HttpContentType
	act1HttpBody := *action1HttpBody

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
	if act1EmailTo == "" && *emailTo != "" {
		act1EmailFrom = *emailFrom
		act1EmailTo = *emailTo
		act1EmailSubject = *emailSubject
		act1EmailContent = *emailContent
	}
	if act1HttpURL == "" && *httpURL != "" {
		act1HttpURL = *httpURL
		act1HttpMethod = *httpMethod
		act1HttpContentType = *httpContentType
		act1HttpBody = *httpBody
	}

	// Add actions based on what's specified
	if act1Type != "" {
		actions = append(actions, createAction(
			act1Type, act1Color, act1TodoList, act1Tags, act1Assignees,
			act1EmailFrom, act1EmailTo, act1EmailSubject, act1EmailContent,
			act1HttpURL, act1HttpMethod, act1HttpContentType, act1HttpBody,
		))
	}
	
	if *action2Type != "" {
		actions = append(actions, createAction(
			*action2Type, *action2Color, *action2TodoList, *action2Tags, *action2Assignees,
			*action2EmailFrom, *action2EmailTo, *action2EmailSubject, *action2EmailContent,
			*action2HttpURL, *action2HttpMethod, *action2HttpContentType, *action2HttpBody,
		))
	}
	
	if *action3Type != "" {
		actions = append(actions, createAction(
			*action3Type, *action3Color, *action3TodoList, *action3Tags, *action3Assignees,
			*action3EmailFrom, *action3EmailTo, *action3EmailSubject, *action3EmailContent,
			*action3HttpURL, *action3HttpMethod, *action3HttpContentType, *action3HttpBody,
		))
	}

	// Create automation input
	input := CreateAutomationInput{
		Trigger: trigger,
		Actions: actions,
	}

	// Execute creation using the existing function
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
		fmt.Printf("✅ Successfully created multi-action automation\n\n")
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