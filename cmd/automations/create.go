package automations

import (
	"encoding/json"
	"fmt"
	"strings"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

// Input structures for automation creation
type CreateAutomationInput struct {
	Trigger CreateAutomationTriggerInput  `json:"trigger"`
	Actions []CreateAutomationActionInput `json:"actions"`
}

type CreateAutomationTriggerInput struct {
	Type                 string                          `json:"type"`
	TodoListID           string                          `json:"todoListId,omitempty"`
	Color                *string                         `json:"color,omitempty"`
	Metadata             *AutomationTriggerMetadataInput `json:"metadata"`
	TagIDs               []string                        `json:"tagIds,omitempty"`
	AssigneeIDs          []string                        `json:"assigneeIds,omitempty"`
	CustomFieldOptionIDs []string                        `json:"customFieldOptionIds,omitempty"`
	TodoIDs              []string                        `json:"todoIds,omitempty"`
}

type AutomationTriggerMetadataInput struct {
	IncompleteOnly *bool `json:"incompleteOnly,omitempty"`
}

type CreateAutomationActionInput struct {
	Type                 string                         `json:"type"`
	DuedIn               *int                           `json:"duedIn,omitempty"`
	Color                *string                        `json:"color,omitempty"`
	AssigneeTriggerer    *string                        `json:"assigneeTriggerer,omitempty"`
	TodoListID           string                         `json:"todoListId,omitempty"`
	TagIDs               []string                       `json:"tagIds,omitempty"`
	AssigneeIDs          []string                       `json:"assigneeIds,omitempty"`
	CustomFieldOptionIDs []string                       `json:"customFieldOptionIds,omitempty"`
	Metadata             *AutomationActionMetadataInput `json:"metadata"`
	HttpOption           *HttpOptionInput               `json:"httpOption"`
}

type AutomationActionMetadataInput struct {
	CopyTodoOptions []string                         `json:"copyTodoOptions,omitempty"`
	Email           *AutomationEmailInput            `json:"email,omitempty"`
	Checklists      []AutomationChecklistInput       `json:"checklists,omitempty"`
}

type AutomationEmailInput struct {
	From        string                          `json:"from"`
	To          []string                        `json:"to"`
	Cc          []string                        `json:"cc,omitempty"`
	Bcc         []string                        `json:"bcc,omitempty"`
	ReplyTo     []string                        `json:"replyTo,omitempty"`
	Subject     string                          `json:"subject"`
	Content     string                          `json:"content"`
	Attachments []AutomationEmailAttachmentInput `json:"attachments,omitempty"`
}

type AutomationEmailAttachmentInput struct {
	UID       string  `json:"uid"`
	Name      string  `json:"name"`
	Size      float64 `json:"size"`
	Type      string  `json:"type"`
	Extension string  `json:"extension"`
}

type AutomationChecklistInput struct {
	Title          string                        `json:"title"`
	Position       float64                       `json:"position"`
	ChecklistItems []AutomationChecklistItemInput `json:"checklistItems,omitempty"`
}

type AutomationChecklistItemInput struct {
	Title       string   `json:"title"`
	Position    float64  `json:"position"`
	DuedIn      *int     `json:"duedIn,omitempty"`
	AssigneeIds []string `json:"assigneeIds,omitempty"`
}

type HttpOptionInput struct {
	URL                      string             `json:"url"`
	Method                   string             `json:"method"`
	ContentType              string             `json:"contentType"`
	Headers                  []HttpHeaderInput  `json:"headers,omitempty"`
	Parameters               []HttpParameterInput `json:"parameters,omitempty"`
	Body                     string             `json:"body,omitempty"`
	AuthorizationType        string             `json:"authorizationType,omitempty"`
	AuthorizationBearerToken string             `json:"authorizationBearerToken,omitempty"`
	AuthorizationBasicAuth   *HttpBasicAuthInput `json:"authorizationBasicAuth,omitempty"`
	AuthorizationApiKey      *HttpApiKeyInput   `json:"authorizationApiKey,omitempty"`
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
	PassBy string `json:"passBy"`
}

// Response structures
type CreatedAutomation struct {
	ID        string                     `json:"id"`
	IsActive  bool                       `json:"isActive"`
	Trigger   AutomationTriggerResponse  `json:"trigger"`
	Actions   []AutomationActionResponse `json:"actions"`
	CreatedAt string                     `json:"createdAt"`
	UpdatedAt string                     `json:"updatedAt"`
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

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new automation",
	Long: `Create a new automation in a workspace. Supports single or multi-action automations.

For single actions, use --action-type with related flags.
For multiple actions, use numbered flags like --action1-type, --action2-type, --action3-type.`,
	Example: `  # Simple email automation
  blue automations create -w <id> --trigger-type TODO_MARKED_AS_COMPLETE \
    --action-type SEND_EMAIL --email-to "user@example.com"

  # HTTP webhook when tag is added
  blue automations create -w <id> --trigger-type TAG_ADDED --trigger-tags "TAG_ID" \
    --action-type MAKE_HTTP_REQUEST --http-url "https://example.com/webhook"

  # Multi-action: email + color change
  blue automations create -w <id> --trigger-type TAG_ADDED --trigger-tags "TAG_ID" \
    --action1-type SEND_EMAIL --action1-email-to "mgr@co.com" \
    --action2-type ADD_COLOR --action2-color "#ff0000"`,
	RunE: runCreate,
}

var (
	createWorkspace          string
	createSimple             bool
	// Trigger flags
	createTriggerType           string
	createTriggerTodoList       string
	createTriggerTags           string
	createTriggerAssignees      string
	createTriggerColor          string
	createTriggerIncompleteOnly bool
	// Unnumbered action flags (single action or action1 fallback)
	createActionType       string
	createActionDueIn      int
	createActionColor      string
	createActionTodoList   string
	createActionTags       string
	createActionAssignees  string
	// Unnumbered email flags
	createEmailFrom    string
	createEmailTo      string
	createEmailSubject string
	createEmailContent string
	// Unnumbered HTTP flags
	createHttpURL         string
	createHttpMethod      string
	createHttpContentType string
	createHttpBody        string
	createHttpHeaders     string
	createHttpParams      string
	createHttpAuthType    string
	createHttpAuthValue   string
	// Action1 flags
	createAction1Type       string
	createAction1Color      string
	createAction1TodoList   string
	createAction1Tags       string
	createAction1Assignees  string
	createAction1EmailFrom    string
	createAction1EmailTo      string
	createAction1EmailSubject string
	createAction1EmailContent string
	createAction1HttpURL         string
	createAction1HttpMethod      string
	createAction1HttpContentType string
	createAction1HttpBody        string
	// Action2 flags
	createAction2Type       string
	createAction2Color      string
	createAction2TodoList   string
	createAction2Tags       string
	createAction2Assignees  string
	createAction2EmailFrom    string
	createAction2EmailTo      string
	createAction2EmailSubject string
	createAction2EmailContent string
	createAction2HttpURL         string
	createAction2HttpMethod      string
	createAction2HttpContentType string
	createAction2HttpBody        string
	// Action3 flags
	createAction3Type       string
	createAction3Color      string
	createAction3TodoList   string
	createAction3Tags       string
	createAction3Assignees  string
	createAction3EmailFrom    string
	createAction3EmailTo      string
	createAction3EmailSubject string
	createAction3EmailContent string
	createAction3HttpURL         string
	createAction3HttpMethod      string
	createAction3HttpContentType string
	createAction3HttpBody        string
)

func init() {
	createCmd.Flags().StringVarP(&createWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	createCmd.Flags().BoolVarP(&createSimple, "simple", "s", false, "Simple output format")

	// Trigger flags
	createCmd.Flags().StringVar(&createTriggerType, "trigger-type", "", "Trigger type (required)")
	createCmd.Flags().StringVar(&createTriggerTodoList, "trigger-todo-list", "", "Todo list ID for trigger")
	createCmd.Flags().StringVar(&createTriggerTags, "trigger-tags", "", "Comma-separated tag IDs")
	createCmd.Flags().StringVar(&createTriggerAssignees, "trigger-assignees", "", "Comma-separated assignee IDs")
	createCmd.Flags().StringVar(&createTriggerColor, "trigger-color", "", "Trigger color")
	createCmd.Flags().BoolVar(&createTriggerIncompleteOnly, "trigger-incomplete-only", false, "Only trigger for incomplete todos")

	// Unnumbered action flags
	createCmd.Flags().StringVar(&createActionType, "action-type", "", "Action type")
	createCmd.Flags().IntVar(&createActionDueIn, "action-due-in", 0, "Due in days for action")
	createCmd.Flags().StringVar(&createActionColor, "action-color", "", "Action color")
	createCmd.Flags().StringVar(&createActionTodoList, "action-todo-list", "", "Todo list ID for action")
	createCmd.Flags().StringVar(&createActionTags, "action-tags", "", "Comma-separated tag IDs")
	createCmd.Flags().StringVar(&createActionAssignees, "action-assignees", "", "Comma-separated assignee IDs")

	// Unnumbered email flags
	createCmd.Flags().StringVar(&createEmailFrom, "email-from", "<p>Blue</p>", "Email from address")
	createCmd.Flags().StringVar(&createEmailTo, "email-to", "", "Comma-separated email addresses")
	createCmd.Flags().StringVar(&createEmailSubject, "email-subject", "", "Email subject")
	createCmd.Flags().StringVar(&createEmailContent, "email-content", "", "Email content (HTML)")

	// Unnumbered HTTP flags
	createCmd.Flags().StringVar(&createHttpURL, "http-url", "", "HTTP request URL")
	createCmd.Flags().StringVar(&createHttpMethod, "http-method", "GET", "HTTP method")
	createCmd.Flags().StringVar(&createHttpContentType, "http-content-type", "JSON", "HTTP content type")
	createCmd.Flags().StringVar(&createHttpBody, "http-body", "", "HTTP request body")
	createCmd.Flags().StringVar(&createHttpHeaders, "http-headers", "", "HTTP headers (key1:value1,key2:value2)")
	createCmd.Flags().StringVar(&createHttpParams, "http-params", "", "HTTP parameters (key1:value1,key2:value2)")
	createCmd.Flags().StringVar(&createHttpAuthType, "http-auth-type", "", "Authorization type (API_KEY, BEARER_TOKEN, BASIC_AUTH)")
	createCmd.Flags().StringVar(&createHttpAuthValue, "http-auth-value", "", "Authorization value")

	// Action1 flags
	createCmd.Flags().StringVar(&createAction1Type, "action1-type", "", "First action type")
	createCmd.Flags().StringVar(&createAction1Color, "action1-color", "", "First action color")
	createCmd.Flags().StringVar(&createAction1TodoList, "action1-todo-list", "", "First action todo list ID")
	createCmd.Flags().StringVar(&createAction1Tags, "action1-tags", "", "First action tag IDs")
	createCmd.Flags().StringVar(&createAction1Assignees, "action1-assignees", "", "First action assignee IDs")
	createCmd.Flags().StringVar(&createAction1EmailFrom, "action1-email-from", "<p>Blue</p>", "First action email from")
	createCmd.Flags().StringVar(&createAction1EmailTo, "action1-email-to", "", "First action email to")
	createCmd.Flags().StringVar(&createAction1EmailSubject, "action1-email-subject", "", "First action email subject")
	createCmd.Flags().StringVar(&createAction1EmailContent, "action1-email-content", "", "First action email content")
	createCmd.Flags().StringVar(&createAction1HttpURL, "action1-http-url", "", "First action HTTP URL")
	createCmd.Flags().StringVar(&createAction1HttpMethod, "action1-http-method", "GET", "First action HTTP method")
	createCmd.Flags().StringVar(&createAction1HttpContentType, "action1-http-content-type", "JSON", "First action HTTP content type")
	createCmd.Flags().StringVar(&createAction1HttpBody, "action1-http-body", "", "First action HTTP body")

	// Action2 flags
	createCmd.Flags().StringVar(&createAction2Type, "action2-type", "", "Second action type")
	createCmd.Flags().StringVar(&createAction2Color, "action2-color", "", "Second action color")
	createCmd.Flags().StringVar(&createAction2TodoList, "action2-todo-list", "", "Second action todo list ID")
	createCmd.Flags().StringVar(&createAction2Tags, "action2-tags", "", "Second action tag IDs")
	createCmd.Flags().StringVar(&createAction2Assignees, "action2-assignees", "", "Second action assignee IDs")
	createCmd.Flags().StringVar(&createAction2EmailFrom, "action2-email-from", "<p>Blue</p>", "Second action email from")
	createCmd.Flags().StringVar(&createAction2EmailTo, "action2-email-to", "", "Second action email to")
	createCmd.Flags().StringVar(&createAction2EmailSubject, "action2-email-subject", "", "Second action email subject")
	createCmd.Flags().StringVar(&createAction2EmailContent, "action2-email-content", "", "Second action email content")
	createCmd.Flags().StringVar(&createAction2HttpURL, "action2-http-url", "", "Second action HTTP URL")
	createCmd.Flags().StringVar(&createAction2HttpMethod, "action2-http-method", "GET", "Second action HTTP method")
	createCmd.Flags().StringVar(&createAction2HttpContentType, "action2-http-content-type", "JSON", "Second action HTTP content type")
	createCmd.Flags().StringVar(&createAction2HttpBody, "action2-http-body", "", "Second action HTTP body")

	// Action3 flags
	createCmd.Flags().StringVar(&createAction3Type, "action3-type", "", "Third action type")
	createCmd.Flags().StringVar(&createAction3Color, "action3-color", "", "Third action color")
	createCmd.Flags().StringVar(&createAction3TodoList, "action3-todo-list", "", "Third action todo list ID")
	createCmd.Flags().StringVar(&createAction3Tags, "action3-tags", "", "Third action tag IDs")
	createCmd.Flags().StringVar(&createAction3Assignees, "action3-assignees", "", "Third action assignee IDs")
	createCmd.Flags().StringVar(&createAction3EmailFrom, "action3-email-from", "<p>Blue</p>", "Third action email from")
	createCmd.Flags().StringVar(&createAction3EmailTo, "action3-email-to", "", "Third action email to")
	createCmd.Flags().StringVar(&createAction3EmailSubject, "action3-email-subject", "", "Third action email subject")
	createCmd.Flags().StringVar(&createAction3EmailContent, "action3-email-content", "", "Third action email content")
	createCmd.Flags().StringVar(&createAction3HttpURL, "action3-http-url", "", "Third action HTTP URL")
	createCmd.Flags().StringVar(&createAction3HttpMethod, "action3-http-method", "GET", "Third action HTTP method")
	createCmd.Flags().StringVar(&createAction3HttpContentType, "action3-http-content-type", "JSON", "Third action HTTP content type")
	createCmd.Flags().StringVar(&createAction3HttpBody, "action3-http-body", "", "Third action HTTP body")
}

func buildCreateAction(actionType, color, todoList, tags, assignees, emailFrom, emailTo, emailSubject, emailContent, httpURL, httpMethod, httpContentType, httpBody, httpHeaders, httpParams, httpAuthType, httpAuthValue string) CreateAutomationActionInput {
	action := CreateAutomationActionInput{
		Type:       actionType,
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
				httpOption.AuthorizationBearerToken = httpAuthValue
			case "API_KEY":
				httpOption.AuthorizationApiKey = &HttpApiKeyInput{
					Key:    "Authorization",
					Value:  httpAuthValue,
					PassBy: "HEADER",
				}
			case "BASIC_AUTH":
				parts := strings.SplitN(httpAuthValue, ":", 2)
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

	return action
}

func executeCreateAutomation(client *common.Client, input CreateAutomationInput) (*CreatedAutomation, error) {
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

	result, err := client.ExecuteQuery(mutation, variables)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(result)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal response: %w", err)
	}

	var response CreateAutomationResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &response.CreateAutomation, nil
}

func runCreate(cmd *cobra.Command, args []string) error {
	if createWorkspace == "" {
		return fmt.Errorf("workspace ID is required. Use --workspace flag")
	}
	if createTriggerType == "" {
		return fmt.Errorf("trigger type is required. Use --trigger-type flag")
	}

	// Determine if this is multi-action (numbered flags) or single action
	isMultiAction := createAction1Type != "" || createAction2Type != "" || createAction3Type != ""

	if !isMultiAction && createActionType == "" {
		return fmt.Errorf("at least --action-type or --action1-type is required")
	}

	// Validate action-specific requirements for single action mode
	if !isMultiAction {
		if createActionType == "SEND_EMAIL" && createEmailTo == "" {
			return fmt.Errorf("--email-to is required for SEND_EMAIL actions")
		}
		if createActionType == "MAKE_HTTP_REQUEST" && createHttpURL == "" {
			return fmt.Errorf("--http-url is required for MAKE_HTTP_REQUEST actions")
		}
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}

	client := common.NewClient(config)
	client.SetProject(createWorkspace)

	// Build trigger
	trigger := CreateAutomationTriggerInput{
		Type:       createTriggerType,
		TodoListID: createTriggerTodoList,
		Metadata:   nil,
	}

	if createTriggerColor != "" {
		trigger.Color = &createTriggerColor
	}
	if createTriggerIncompleteOnly {
		trigger.Metadata = &AutomationTriggerMetadataInput{
			IncompleteOnly: &createTriggerIncompleteOnly,
		}
	}
	if createTriggerTags != "" {
		trigger.TagIDs = strings.Split(createTriggerTags, ",")
	}
	if createTriggerAssignees != "" {
		trigger.AssigneeIDs = strings.Split(createTriggerAssignees, ",")
	}

	// Build actions
	var actions []CreateAutomationActionInput

	if isMultiAction {
		// Multi-action mode: use numbered flags, with unnumbered as fallback for action1
		act1Type := createAction1Type
		act1Color := createAction1Color
		act1TodoList := createAction1TodoList
		act1Tags := createAction1Tags
		act1Assignees := createAction1Assignees
		act1EmailFrom := createAction1EmailFrom
		act1EmailTo := createAction1EmailTo
		act1EmailSubject := createAction1EmailSubject
		act1EmailContent := createAction1EmailContent
		act1HttpURL := createAction1HttpURL
		act1HttpMethod := createAction1HttpMethod
		act1HttpContentType := createAction1HttpContentType
		act1HttpBody := createAction1HttpBody

		// Fallback to unnumbered flags for action1
		if act1Type == "" && createActionType != "" {
			act1Type = createActionType
		}
		if act1Color == "" && createActionColor != "" {
			act1Color = createActionColor
		}
		if act1TodoList == "" && createActionTodoList != "" {
			act1TodoList = createActionTodoList
		}
		if act1Tags == "" && createActionTags != "" {
			act1Tags = createActionTags
		}
		if act1Assignees == "" && createActionAssignees != "" {
			act1Assignees = createActionAssignees
		}
		if act1EmailTo == "" && createEmailTo != "" {
			act1EmailFrom = createEmailFrom
			act1EmailTo = createEmailTo
			act1EmailSubject = createEmailSubject
			act1EmailContent = createEmailContent
		}
		if act1HttpURL == "" && createHttpURL != "" {
			act1HttpURL = createHttpURL
			act1HttpMethod = createHttpMethod
			act1HttpContentType = createHttpContentType
			act1HttpBody = createHttpBody
		}

		if act1Type != "" {
			actions = append(actions, buildCreateAction(
				act1Type, act1Color, act1TodoList, act1Tags, act1Assignees,
				act1EmailFrom, act1EmailTo, act1EmailSubject, act1EmailContent,
				act1HttpURL, act1HttpMethod, act1HttpContentType, act1HttpBody,
				"", "", "", "",
			))
		}
		if createAction2Type != "" {
			actions = append(actions, buildCreateAction(
				createAction2Type, createAction2Color, createAction2TodoList, createAction2Tags, createAction2Assignees,
				createAction2EmailFrom, createAction2EmailTo, createAction2EmailSubject, createAction2EmailContent,
				createAction2HttpURL, createAction2HttpMethod, createAction2HttpContentType, createAction2HttpBody,
				"", "", "", "",
			))
		}
		if createAction3Type != "" {
			actions = append(actions, buildCreateAction(
				createAction3Type, createAction3Color, createAction3TodoList, createAction3Tags, createAction3Assignees,
				createAction3EmailFrom, createAction3EmailTo, createAction3EmailSubject, createAction3EmailContent,
				createAction3HttpURL, createAction3HttpMethod, createAction3HttpContentType, createAction3HttpBody,
				"", "", "", "",
			))
		}
	} else {
		// Single action mode
		action := buildCreateAction(
			createActionType, createActionColor, createActionTodoList, createActionTags, createActionAssignees,
			createEmailFrom, createEmailTo, createEmailSubject, createEmailContent,
			createHttpURL, createHttpMethod, createHttpContentType, createHttpBody,
			createHttpHeaders, createHttpParams, createHttpAuthType, createHttpAuthValue,
		)
		if createActionDueIn > 0 {
			action.DuedIn = &createActionDueIn
		}
		actions = append(actions, action)
	}

	input := CreateAutomationInput{
		Trigger: trigger,
		Actions: actions,
	}

	automation, err := executeCreateAutomation(client, input)
	if err != nil {
		return fmt.Errorf("failed to create automation: %w", err)
	}

	if createSimple {
		fmt.Printf("Created automation: %s\n", automation.ID)
		fmt.Printf("Trigger: %s\n", automation.Trigger.Type)
		for i, action := range automation.Actions {
			fmt.Printf("Action %d: %s\n", i+1, action.Type)
		}
	} else {
		fmt.Printf("Successfully created automation\n\n")
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
