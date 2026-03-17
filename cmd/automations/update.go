package automations

import (
	"encoding/json"
	"fmt"
	"strings"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

// Input structure for automation update
type EditAutomationInput struct {
	AutomationID string                        `json:"automationId"`
	Trigger      *CreateAutomationTriggerInput `json:"trigger,omitempty"`
	Actions      []CreateAutomationActionInput `json:"actions,omitempty"`
	IsActive     *bool                         `json:"isActive,omitempty"`
}

type UpdateAutomationResponse struct {
	EditAutomation CreatedAutomation `json:"editAutomation"`
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an automation",
	Long: `Update an existing automation. Supports partial updates and multi-action automations.

WARNING: Action updates REPLACE the entire actions array. For multi-action automations,
you must specify ALL existing actions to avoid losing any.`,
	Example: `  # Enable/disable automation
  blue automations update -w <id> --automation <aid> --active true

  # Update trigger
  blue automations update -w <id> --automation <aid> \
    --trigger-type TODO_MARKED_AS_COMPLETE

  # Update single action email settings
  blue automations update -w <id> --automation <aid> \
    --action-type SEND_EMAIL --email-to "new@example.com"

  # Update multiple actions
  blue automations update -w <id> --automation <aid> \
    --action1-type SEND_EMAIL --action1-email-to "mgr@co.com" \
    --action2-type ADD_COLOR --action2-color "#00ff00"`,
	RunE: runUpdate,
}

var (
	updateAutomationID string
	updateWorkspace    string
	updateSimple       bool
	updateIsActive     string
	// Trigger flags
	updateTriggerType           string
	updateTriggerTodoList       string
	updateTriggerTags           string
	updateTriggerAssignees      string
	updateTriggerColor          string
	updateTriggerIncompleteOnly string
	// Unnumbered action flags
	updateActionType      string
	updateActionDueIn     int
	updateActionColor     string
	updateActionTodoList  string
	updateActionTags      string
	updateActionAssignees string
	// Unnumbered email flags
	updateEmailFrom    string
	updateEmailTo      string
	updateEmailSubject string
	updateEmailContent string
	// Unnumbered HTTP flags
	updateHttpURL         string
	updateHttpMethod      string
	updateHttpContentType string
	updateHttpBody        string
	updateHttpHeaders     string
	updateHttpParams      string
	updateHttpAuthType    string
	updateHttpAuthValue   string
	// Action1 flags
	updateAction1Type      string
	updateAction1Color     string
	updateAction1TodoList  string
	updateAction1Tags      string
	updateAction1Assignees string
	updateAction1DueIn     int
	updateAction1EmailFrom    string
	updateAction1EmailTo      string
	updateAction1EmailSubject string
	updateAction1EmailContent string
	updateAction1HttpURL         string
	updateAction1HttpMethod      string
	updateAction1HttpContentType string
	updateAction1HttpBody        string
	updateAction1HttpHeaders     string
	updateAction1HttpParams      string
	updateAction1HttpAuthType    string
	updateAction1HttpAuthValue   string
	// Action2 flags
	updateAction2Type      string
	updateAction2Color     string
	updateAction2TodoList  string
	updateAction2Tags      string
	updateAction2Assignees string
	updateAction2DueIn     int
	updateAction2EmailFrom    string
	updateAction2EmailTo      string
	updateAction2EmailSubject string
	updateAction2EmailContent string
	updateAction2HttpURL         string
	updateAction2HttpMethod      string
	updateAction2HttpContentType string
	updateAction2HttpBody        string
	updateAction2HttpHeaders     string
	updateAction2HttpParams      string
	updateAction2HttpAuthType    string
	updateAction2HttpAuthValue   string
	// Action3 flags
	updateAction3Type      string
	updateAction3Color     string
	updateAction3TodoList  string
	updateAction3Tags      string
	updateAction3Assignees string
	updateAction3DueIn     int
	updateAction3EmailFrom    string
	updateAction3EmailTo      string
	updateAction3EmailSubject string
	updateAction3EmailContent string
	updateAction3HttpURL         string
	updateAction3HttpMethod      string
	updateAction3HttpContentType string
	updateAction3HttpBody        string
	updateAction3HttpHeaders     string
	updateAction3HttpParams      string
	updateAction3HttpAuthType    string
	updateAction3HttpAuthValue   string
)

func init() {
	updateCmd.Flags().StringVar(&updateAutomationID, "automation", "", "Automation ID (required)")
	updateCmd.Flags().StringVarP(&updateWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	updateCmd.Flags().BoolVarP(&updateSimple, "simple", "s", false, "Simple output format")
	updateCmd.Flags().StringVar(&updateIsActive, "active", "", "Set active status (true/false)")

	// Trigger flags
	updateCmd.Flags().StringVar(&updateTriggerType, "trigger-type", "", "Update trigger type")
	updateCmd.Flags().StringVar(&updateTriggerTodoList, "trigger-todo-list", "", "Todo list ID for trigger")
	updateCmd.Flags().StringVar(&updateTriggerTags, "trigger-tags", "", "Comma-separated tag IDs")
	updateCmd.Flags().StringVar(&updateTriggerAssignees, "trigger-assignees", "", "Comma-separated assignee IDs")
	updateCmd.Flags().StringVar(&updateTriggerColor, "trigger-color", "", "Trigger color")
	updateCmd.Flags().StringVar(&updateTriggerIncompleteOnly, "trigger-incomplete-only", "", "Only trigger for incomplete todos (true/false)")

	// Unnumbered action flags
	updateCmd.Flags().StringVar(&updateActionType, "action-type", "", "Action type")
	updateCmd.Flags().IntVar(&updateActionDueIn, "action-due-in", -1, "Due in days (-1 to keep current)")
	updateCmd.Flags().StringVar(&updateActionColor, "action-color", "", "Action color")
	updateCmd.Flags().StringVar(&updateActionTodoList, "action-todo-list", "", "Todo list ID for action")
	updateCmd.Flags().StringVar(&updateActionTags, "action-tags", "", "Comma-separated tag IDs")
	updateCmd.Flags().StringVar(&updateActionAssignees, "action-assignees", "", "Comma-separated assignee IDs")

	// Unnumbered email flags
	updateCmd.Flags().StringVar(&updateEmailFrom, "email-from", "", "Email from address")
	updateCmd.Flags().StringVar(&updateEmailTo, "email-to", "", "Comma-separated email addresses")
	updateCmd.Flags().StringVar(&updateEmailSubject, "email-subject", "", "Email subject")
	updateCmd.Flags().StringVar(&updateEmailContent, "email-content", "", "Email content (HTML)")

	// Unnumbered HTTP flags
	updateCmd.Flags().StringVar(&updateHttpURL, "http-url", "", "HTTP request URL")
	updateCmd.Flags().StringVar(&updateHttpMethod, "http-method", "", "HTTP method")
	updateCmd.Flags().StringVar(&updateHttpContentType, "http-content-type", "", "HTTP content type")
	updateCmd.Flags().StringVar(&updateHttpBody, "http-body", "", "HTTP request body")
	updateCmd.Flags().StringVar(&updateHttpHeaders, "http-headers", "", "HTTP headers (key1:value1,key2:value2)")
	updateCmd.Flags().StringVar(&updateHttpParams, "http-params", "", "HTTP parameters (key1:value1,key2:value2)")
	updateCmd.Flags().StringVar(&updateHttpAuthType, "http-auth-type", "", "Authorization type")
	updateCmd.Flags().StringVar(&updateHttpAuthValue, "http-auth-value", "", "Authorization value")

	// Action1 flags
	updateCmd.Flags().StringVar(&updateAction1Type, "action1-type", "", "First action type")
	updateCmd.Flags().StringVar(&updateAction1Color, "action1-color", "", "First action color")
	updateCmd.Flags().StringVar(&updateAction1TodoList, "action1-todo-list", "", "First action todo list ID")
	updateCmd.Flags().StringVar(&updateAction1Tags, "action1-tags", "", "First action tag IDs")
	updateCmd.Flags().StringVar(&updateAction1Assignees, "action1-assignees", "", "First action assignee IDs")
	updateCmd.Flags().IntVar(&updateAction1DueIn, "action1-due-in", -1, "First action due in days")
	updateCmd.Flags().StringVar(&updateAction1EmailFrom, "action1-email-from", "", "First action email from")
	updateCmd.Flags().StringVar(&updateAction1EmailTo, "action1-email-to", "", "First action email to")
	updateCmd.Flags().StringVar(&updateAction1EmailSubject, "action1-email-subject", "", "First action email subject")
	updateCmd.Flags().StringVar(&updateAction1EmailContent, "action1-email-content", "", "First action email content")
	updateCmd.Flags().StringVar(&updateAction1HttpURL, "action1-http-url", "", "First action HTTP URL")
	updateCmd.Flags().StringVar(&updateAction1HttpMethod, "action1-http-method", "", "First action HTTP method")
	updateCmd.Flags().StringVar(&updateAction1HttpContentType, "action1-http-content-type", "", "First action HTTP content type")
	updateCmd.Flags().StringVar(&updateAction1HttpBody, "action1-http-body", "", "First action HTTP body")
	updateCmd.Flags().StringVar(&updateAction1HttpHeaders, "action1-http-headers", "", "First action HTTP headers")
	updateCmd.Flags().StringVar(&updateAction1HttpParams, "action1-http-params", "", "First action HTTP parameters")
	updateCmd.Flags().StringVar(&updateAction1HttpAuthType, "action1-http-auth-type", "", "First action auth type")
	updateCmd.Flags().StringVar(&updateAction1HttpAuthValue, "action1-http-auth-value", "", "First action auth value")

	// Action2 flags
	updateCmd.Flags().StringVar(&updateAction2Type, "action2-type", "", "Second action type")
	updateCmd.Flags().StringVar(&updateAction2Color, "action2-color", "", "Second action color")
	updateCmd.Flags().StringVar(&updateAction2TodoList, "action2-todo-list", "", "Second action todo list ID")
	updateCmd.Flags().StringVar(&updateAction2Tags, "action2-tags", "", "Second action tag IDs")
	updateCmd.Flags().StringVar(&updateAction2Assignees, "action2-assignees", "", "Second action assignee IDs")
	updateCmd.Flags().IntVar(&updateAction2DueIn, "action2-due-in", -1, "Second action due in days")
	updateCmd.Flags().StringVar(&updateAction2EmailFrom, "action2-email-from", "", "Second action email from")
	updateCmd.Flags().StringVar(&updateAction2EmailTo, "action2-email-to", "", "Second action email to")
	updateCmd.Flags().StringVar(&updateAction2EmailSubject, "action2-email-subject", "", "Second action email subject")
	updateCmd.Flags().StringVar(&updateAction2EmailContent, "action2-email-content", "", "Second action email content")
	updateCmd.Flags().StringVar(&updateAction2HttpURL, "action2-http-url", "", "Second action HTTP URL")
	updateCmd.Flags().StringVar(&updateAction2HttpMethod, "action2-http-method", "", "Second action HTTP method")
	updateCmd.Flags().StringVar(&updateAction2HttpContentType, "action2-http-content-type", "", "Second action HTTP content type")
	updateCmd.Flags().StringVar(&updateAction2HttpBody, "action2-http-body", "", "Second action HTTP body")
	updateCmd.Flags().StringVar(&updateAction2HttpHeaders, "action2-http-headers", "", "Second action HTTP headers")
	updateCmd.Flags().StringVar(&updateAction2HttpParams, "action2-http-params", "", "Second action HTTP parameters")
	updateCmd.Flags().StringVar(&updateAction2HttpAuthType, "action2-http-auth-type", "", "Second action auth type")
	updateCmd.Flags().StringVar(&updateAction2HttpAuthValue, "action2-http-auth-value", "", "Second action auth value")

	// Action3 flags
	updateCmd.Flags().StringVar(&updateAction3Type, "action3-type", "", "Third action type")
	updateCmd.Flags().StringVar(&updateAction3Color, "action3-color", "", "Third action color")
	updateCmd.Flags().StringVar(&updateAction3TodoList, "action3-todo-list", "", "Third action todo list ID")
	updateCmd.Flags().StringVar(&updateAction3Tags, "action3-tags", "", "Third action tag IDs")
	updateCmd.Flags().StringVar(&updateAction3Assignees, "action3-assignees", "", "Third action assignee IDs")
	updateCmd.Flags().IntVar(&updateAction3DueIn, "action3-due-in", -1, "Third action due in days")
	updateCmd.Flags().StringVar(&updateAction3EmailFrom, "action3-email-from", "", "Third action email from")
	updateCmd.Flags().StringVar(&updateAction3EmailTo, "action3-email-to", "", "Third action email to")
	updateCmd.Flags().StringVar(&updateAction3EmailSubject, "action3-email-subject", "", "Third action email subject")
	updateCmd.Flags().StringVar(&updateAction3EmailContent, "action3-email-content", "", "Third action email content")
	updateCmd.Flags().StringVar(&updateAction3HttpURL, "action3-http-url", "", "Third action HTTP URL")
	updateCmd.Flags().StringVar(&updateAction3HttpMethod, "action3-http-method", "", "Third action HTTP method")
	updateCmd.Flags().StringVar(&updateAction3HttpContentType, "action3-http-content-type", "", "Third action HTTP content type")
	updateCmd.Flags().StringVar(&updateAction3HttpBody, "action3-http-body", "", "Third action HTTP body")
	updateCmd.Flags().StringVar(&updateAction3HttpHeaders, "action3-http-headers", "", "Third action HTTP headers")
	updateCmd.Flags().StringVar(&updateAction3HttpParams, "action3-http-params", "", "Third action HTTP parameters")
	updateCmd.Flags().StringVar(&updateAction3HttpAuthType, "action3-http-auth-type", "", "Third action auth type")
	updateCmd.Flags().StringVar(&updateAction3HttpAuthValue, "action3-http-auth-value", "", "Third action auth value")
}

func buildUpdateAction(actionType, color, todoList, tags, assignees string, dueIn int, emailFrom, emailTo, emailSubject, emailContent, httpURL, httpMethod, httpContentType, httpBody, httpHeaders, httpParams, httpAuthType, httpAuthValue string) CreateAutomationActionInput {
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

	// Handle SEND_EMAIL
	if actionType == "SEND_EMAIL" && emailTo != "" {
		emailMetadata := &AutomationEmailInput{
			From:        emailFrom,
			To:          strings.Split(emailTo, ","),
			Subject:     emailSubject,
			Content:     emailContent,
			Cc:          []string{},
			Bcc:         []string{},
			ReplyTo:     []string{},
			Attachments: []AutomationEmailAttachmentInput{},
		}
		action.Metadata = &AutomationActionMetadataInput{
			Email: emailMetadata,
		}
	}

	// Handle MAKE_HTTP_REQUEST
	if actionType == "MAKE_HTTP_REQUEST" && httpURL != "" {
		httpOption := &HttpOptionInput{
			URL:         httpURL,
			Method:      httpMethod,
			ContentType: httpContentType,
			Body:        httpBody,
		}

		if httpHeaders != "" {
			pairs := strings.Split(httpHeaders, ",")
			for _, pair := range pairs {
				parts := strings.SplitN(pair, ":", 2)
				if len(parts) == 2 {
					httpOption.Headers = append(httpOption.Headers, HttpHeaderInput{
						Key:   strings.TrimSpace(parts[0]),
						Value: strings.TrimSpace(parts[1]),
					})
				}
			}
		}

		if httpParams != "" {
			pairs := strings.Split(httpParams, ",")
			for _, pair := range pairs {
				parts := strings.SplitN(pair, ":", 2)
				if len(parts) == 2 {
					httpOption.Parameters = append(httpOption.Parameters, HttpParameterInput{
						Key:   strings.TrimSpace(parts[0]),
						Value: strings.TrimSpace(parts[1]),
					})
				}
			}
		}

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
						Key:    "Authorization",
						Value:  httpAuthValue,
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

func executeUpdateAutomation(client *common.Client, input EditAutomationInput) (*CreatedAutomation, error) {
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

	result, err := client.ExecuteQuery(mutation, variables)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	var response UpdateAutomationResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response.EditAutomation, nil
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if updateAutomationID == "" {
		return fmt.Errorf("automation ID is required. Use --automation flag")
	}
	if updateWorkspace == "" {
		return fmt.Errorf("workspace ID is required. Use --workspace flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)
	client.SetProject(updateWorkspace)

	input := EditAutomationInput{
		AutomationID: updateAutomationID,
	}

	// Set active status
	if updateIsActive != "" {
		if updateIsActive == "true" {
			active := true
			input.IsActive = &active
		} else if updateIsActive == "false" {
			active := false
			input.IsActive = &active
		} else {
			return fmt.Errorf("--active must be 'true' or 'false', got: %s", updateIsActive)
		}
	}

	// Build trigger if any trigger options provided
	triggerProvided := updateTriggerType != "" || updateTriggerTodoList != "" || updateTriggerTags != "" ||
		updateTriggerAssignees != "" || updateTriggerColor != "" || updateTriggerIncompleteOnly != ""

	if triggerProvided {
		trigger := &CreateAutomationTriggerInput{
			TodoListID: updateTriggerTodoList,
			Metadata:   nil,
		}

		if updateTriggerType != "" {
			trigger.Type = updateTriggerType
		}
		if updateTriggerColor != "" {
			trigger.Color = &updateTriggerColor
		}
		if updateTriggerIncompleteOnly != "" {
			if updateTriggerIncompleteOnly == "true" {
				incompleteOnly := true
				trigger.Metadata = &AutomationTriggerMetadataInput{
					IncompleteOnly: &incompleteOnly,
				}
			} else if updateTriggerIncompleteOnly == "false" {
				incompleteOnly := false
				trigger.Metadata = &AutomationTriggerMetadataInput{
					IncompleteOnly: &incompleteOnly,
				}
			} else {
				return fmt.Errorf("--trigger-incomplete-only must be 'true' or 'false', got: %s", updateTriggerIncompleteOnly)
			}
		}
		if updateTriggerTags != "" {
			trigger.TagIDs = strings.Split(updateTriggerTags, ",")
		}
		if updateTriggerAssignees != "" {
			trigger.AssigneeIDs = strings.Split(updateTriggerAssignees, ",")
		}

		input.Trigger = trigger
	}

	// Build actions
	var actions []CreateAutomationActionInput

	// Resolve action1 from numbered + unnumbered fallback
	act1Type := updateAction1Type
	act1Color := updateAction1Color
	act1TodoList := updateAction1TodoList
	act1Tags := updateAction1Tags
	act1Assignees := updateAction1Assignees
	act1DueIn := updateAction1DueIn
	act1EmailFrom := updateAction1EmailFrom
	act1EmailTo := updateAction1EmailTo
	act1EmailSubject := updateAction1EmailSubject
	act1EmailContent := updateAction1EmailContent
	act1HttpURL := updateAction1HttpURL
	act1HttpMethod := updateAction1HttpMethod
	act1HttpContentType := updateAction1HttpContentType
	act1HttpBody := updateAction1HttpBody
	act1HttpHeaders := updateAction1HttpHeaders
	act1HttpParams := updateAction1HttpParams
	act1HttpAuthType := updateAction1HttpAuthType
	act1HttpAuthValue := updateAction1HttpAuthValue

	if act1Type == "" && updateActionType != "" {
		act1Type = updateActionType
	}
	if act1Color == "" && updateActionColor != "" {
		act1Color = updateActionColor
	}
	if act1TodoList == "" && updateActionTodoList != "" {
		act1TodoList = updateActionTodoList
	}
	if act1Tags == "" && updateActionTags != "" {
		act1Tags = updateActionTags
	}
	if act1Assignees == "" && updateActionAssignees != "" {
		act1Assignees = updateActionAssignees
	}
	if act1DueIn == -1 && updateActionDueIn != -1 {
		act1DueIn = updateActionDueIn
	}
	if act1EmailTo == "" && updateEmailTo != "" {
		if updateEmailFrom != "" {
			act1EmailFrom = updateEmailFrom
		}
		act1EmailTo = updateEmailTo
		if updateEmailSubject != "" {
			act1EmailSubject = updateEmailSubject
		}
		if updateEmailContent != "" {
			act1EmailContent = updateEmailContent
		}
	}
	if act1HttpURL == "" && updateHttpURL != "" {
		act1HttpURL = updateHttpURL
		if updateHttpMethod != "" {
			act1HttpMethod = updateHttpMethod
		}
		if updateHttpContentType != "" {
			act1HttpContentType = updateHttpContentType
		}
		if updateHttpBody != "" {
			act1HttpBody = updateHttpBody
		}
		if updateHttpHeaders != "" {
			act1HttpHeaders = updateHttpHeaders
		}
		if updateHttpParams != "" {
			act1HttpParams = updateHttpParams
		}
		if updateHttpAuthType != "" {
			act1HttpAuthType = updateHttpAuthType
		}
		if updateHttpAuthValue != "" {
			act1HttpAuthValue = updateHttpAuthValue
		}
	}

	// Check if action1 has any fields
	action1Provided := act1Type != "" || act1Color != "" || act1TodoList != "" || act1Tags != "" ||
		act1Assignees != "" || act1DueIn != -1 || act1EmailTo != "" || act1HttpURL != ""

	if action1Provided {
		actions = append(actions, buildUpdateAction(
			act1Type, act1Color, act1TodoList, act1Tags, act1Assignees, act1DueIn,
			act1EmailFrom, act1EmailTo, act1EmailSubject, act1EmailContent,
			act1HttpURL, act1HttpMethod, act1HttpContentType, act1HttpBody,
			act1HttpHeaders, act1HttpParams, act1HttpAuthType, act1HttpAuthValue,
		))
	}

	// Action2
	action2Provided := updateAction2Type != "" || updateAction2Color != "" || updateAction2TodoList != "" ||
		updateAction2Tags != "" || updateAction2Assignees != "" || updateAction2DueIn != -1 ||
		updateAction2EmailTo != "" || updateAction2HttpURL != ""

	if action2Provided {
		actions = append(actions, buildUpdateAction(
			updateAction2Type, updateAction2Color, updateAction2TodoList, updateAction2Tags, updateAction2Assignees, updateAction2DueIn,
			updateAction2EmailFrom, updateAction2EmailTo, updateAction2EmailSubject, updateAction2EmailContent,
			updateAction2HttpURL, updateAction2HttpMethod, updateAction2HttpContentType, updateAction2HttpBody,
			updateAction2HttpHeaders, updateAction2HttpParams, updateAction2HttpAuthType, updateAction2HttpAuthValue,
		))
	}

	// Action3
	action3Provided := updateAction3Type != "" || updateAction3Color != "" || updateAction3TodoList != "" ||
		updateAction3Tags != "" || updateAction3Assignees != "" || updateAction3DueIn != -1 ||
		updateAction3EmailTo != "" || updateAction3HttpURL != ""

	if action3Provided {
		actions = append(actions, buildUpdateAction(
			updateAction3Type, updateAction3Color, updateAction3TodoList, updateAction3Tags, updateAction3Assignees, updateAction3DueIn,
			updateAction3EmailFrom, updateAction3EmailTo, updateAction3EmailSubject, updateAction3EmailContent,
			updateAction3HttpURL, updateAction3HttpMethod, updateAction3HttpContentType, updateAction3HttpBody,
			updateAction3HttpHeaders, updateAction3HttpParams, updateAction3HttpAuthType, updateAction3HttpAuthValue,
		))
	}

	if len(actions) > 0 {
		input.Actions = actions
		fmt.Printf("WARNING: Action update will REPLACE all actions. If this automation has multiple actions, specify ALL of them to avoid data loss.\n\n")
	}

	// Validate at least one field is being updated
	if input.Trigger == nil && input.Actions == nil && input.IsActive == nil {
		return fmt.Errorf("at least one field must be provided to update (trigger, action, or active status)")
	}

	automation, err := executeUpdateAutomation(client, input)
	if err != nil {
		return fmt.Errorf("failed to update automation: %w", err)
	}

	if updateSimple {
		fmt.Printf("Updated automation: %s\n", automation.ID)
		fmt.Printf("Active: %t\n", automation.IsActive)
		fmt.Printf("Trigger: %s\n", automation.Trigger.Type)
		for i, action := range automation.Actions {
			fmt.Printf("Action %d: %s\n", i+1, action.Type)
		}
	} else {
		fmt.Printf("Successfully updated automation\n\n")
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
