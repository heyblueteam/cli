package webhooks

type Webhook struct {
	ID         string                 `json:"id"`
	UID        string                 `json:"uid"`
	Name       string                 `json:"name"`
	URL        string                 `json:"url"`
	Secret     string                 `json:"secret,omitempty"`
	Status     string                 `json:"status"`
	Events     []string               `json:"events"`
	ProjectIDs []string               `json:"projectIds"`
	Enabled    bool                   `json:"enabled"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  string                 `json:"createdAt"`
	UpdatedAt  string                 `json:"updatedAt"`
}

type mutationResult struct {
	Success     bool   `json:"success"`
	OperationID string `json:"operationId"`
}

var webhookFields = `
	id
	uid
	name
	url
	secret
	status
	events
	projectIds
	enabled
	metadata
	createdAt
	updatedAt
`

type webhookEvent struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

var webhookEvents = []webhookEvent{
	{"TODO_CREATED", "A record was created."},
	{"TODO_DELETED", "A record was deleted."},
	{"TODO_MOVED", "A record was moved to a different list."},
	{"TODO_NAME_CHANGED", "A record's title was changed."},
	{"TODO_DONE_STATUS_UPDATED", "A record was marked done or undone."},
	{"TODO_DUE_DATE_ADDED", "A due date was added to a record."},
	{"TODO_DUE_DATE_UPDATED", "A record's due date was changed."},
	{"TODO_DUE_DATE_REMOVED", "A due date was removed from a record."},
	{"TODO_ASSIGNEE_ADDED", "An assignee was added to a record."},
	{"TODO_ASSIGNEE_REMOVED", "An assignee was removed from a record."},
	{"TODO_TAG_ADDED", "A tag was added to a record."},
	{"TODO_TAG_REMOVED", "A tag was removed from a record."},
	{"TODO_CUSTOM_FIELD_UPDATED", "A custom field value changed on a record."},
	{"TODO_CHECKLIST_CREATED", "A checklist was created on a record."},
	{"TODO_CHECKLIST_NAME_CHANGED", "A checklist was renamed."},
	{"TODO_CHECKLIST_DELETED", "A checklist was deleted."},
	{"TODO_CHECKLIST_ITEM_CREATED", "A checklist item was created."},
	{"TODO_CHECKLIST_ITEM_NAME_CHANGED", "A checklist item was renamed."},
	{"TODO_CHECKLIST_ITEM_DELETED", "A checklist item was deleted."},
	{"TODO_CHECKLIST_ITEM_DUE_DATE_ADDED", "A due date was added to a checklist item."},
	{"TODO_CHECKLIST_ITEM_DUE_DATE_UPDATED", "A checklist item's due date was changed."},
	{"TODO_CHECKLIST_ITEM_DUE_DATE_REMOVED", "A due date was removed from a checklist item."},
	{"TODO_CHECKLIST_ITEM_ASSIGNEE_ADDED", "An assignee was added to a checklist item."},
	{"TODO_CHECKLIST_ITEM_ASSIGNEE_REMOVED", "An assignee was removed from a checklist item."},
	{"TODO_CHECKLIST_ITEM_DONE_STATUS_UPDATED", "A checklist item was marked done or undone."},
	{"TODO_LIST_CREATED", "A list was created in a workspace."},
	{"TODO_LIST_DELETED", "A list was deleted."},
	{"TODO_LIST_NAME_CHANGED", "A list was renamed."},
	{"CUSTOM_FIELD_CREATED", "A custom field was created."},
	{"CUSTOM_FIELD_DELETED", "A custom field was deleted."},
	{"CUSTOM_FIELD_UPDATED", "A custom field definition was updated."},
	{"TAG_CREATED", "A tag was created."},
	{"TAG_DELETED", "A tag was deleted."},
	{"TAG_UPDATED", "A tag was updated."},
	{"COMMENT_CREATED", "A comment was created."},
	{"COMMENT_DELETED", "A comment was deleted."},
	{"COMMENT_UPDATED", "A comment was updated."},
	{"WEBHOOK_HEALTH_CHECK", "Delivery-only health-check event sent when a webhook is re-enabled."},
}
