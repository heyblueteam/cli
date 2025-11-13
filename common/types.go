package common

// ============================================================================
// CORE ENTITY TYPES
// ============================================================================

// User represents a user in the Blue system
type User struct {
	ID        string `json:"id"`
	UID       string `json:"uid"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	FullName  string `json:"fullName"`
	Email     string `json:"email"`
}

// Tag represents a tag that can be associated with records
type Tag struct {
	ID        string `json:"id"`
	UID       string `json:"uid"`
	Title     string `json:"title"`
	Color     string `json:"color"`
	CreatedAt string `json:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt,omitempty"`
}

// Project represents a Blue project with all possible fields
type Project struct {
	ID                      string           `json:"id"`
	UID                     string           `json:"uid"`
	Slug                    string           `json:"slug"`
	Name                    string           `json:"name"`
	Description             string           `json:"description"`
	Archived                bool             `json:"archived"`
	Color                   string           `json:"color"`
	Icon                    string           `json:"icon"`
	Category                string           `json:"category,omitempty"`
	TodoAlias               string           `json:"todoAlias,omitempty"`
	HideRecordCount         bool             `json:"hideRecordCount,omitempty"`
	ShowTimeSpentInTodoList bool             `json:"showTimeSpentInTodoList,omitempty"`
	ShowTimeSpentInProject  bool             `json:"showTimeSpentInProject,omitempty"`
	CreatedAt               string           `json:"createdAt,omitempty"`
	UpdatedAt               string           `json:"updatedAt,omitempty"`
	Position                float64          `json:"position,omitempty"`
	IsTemplate              bool             `json:"isTemplate,omitempty"`
	Features                []ProjectFeature `json:"features,omitempty"`
	TodoFields              []TodoField      `json:"todoFields,omitempty"`
}

// ProjectFeature represents a project feature toggle
type ProjectFeature struct {
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

// ProjectFeatureInput for mutations
type ProjectFeatureInput struct {
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

// TodoField represents a field configuration in a project (includes custom field groups)
type TodoField struct {
	Type          string       `json:"type"`           // TodoFieldType enum (CUSTOM_FIELD, CUSTOM_FIELD_GROUP, etc.)
	CustomFieldID *string      `json:"customFieldId"`  // Group ID or field ID
	Name          *string      `json:"name,omitempty"` // Group name (for CUSTOM_FIELD_GROUP)
	Color         *string      `json:"color,omitempty"` // Group color (for CUSTOM_FIELD_GROUP)
	TodoFields    []TodoField  `json:"todoFields,omitempty"` // Nested fields (for CUSTOM_FIELD_GROUP)
}

// TodoFieldInput for mutation input
type TodoFieldInput struct {
	Type          string            `json:"type"`
	CustomFieldID *string           `json:"customFieldId,omitempty"`
	Name          *string           `json:"name,omitempty"`
	Color         *string           `json:"color,omitempty"`
	TodoFields    []TodoFieldInput  `json:"todoFields,omitempty"`
}

// TodoList represents a todo list with all possible fields
type TodoList struct {
	ID               string   `json:"id"`
	UID              string   `json:"uid"`
	Title            string   `json:"title"`
	Position         float64  `json:"position"`
	TodosCount       int      `json:"todosCount,omitempty"`
	TodosMaxPosition float64  `json:"todosMaxPosition,omitempty"`
	CreatedAt        string   `json:"createdAt,omitempty"`
	UpdatedAt        string   `json:"updatedAt,omitempty"`
	IsDisabled       bool     `json:"isDisabled,omitempty"`
	IsLocked         bool     `json:"isLocked,omitempty"`
	Completed        bool     `json:"completed,omitempty"`
	Editable         bool     `json:"editable,omitempty"`
	Deletable        bool     `json:"deletable,omitempty"`
	Todos            []Record `json:"todos,omitempty"`
}

// TodoListSimple for minimal use cases (like read-project-todos)
type TodoListSimple struct {
	ID    string   `json:"id"`
	Title string   `json:"title"`
	Todos []Record `json:"todos"`
}

// Record represents a todo/record with all possible fields
type Record struct {
	ID                      string        `json:"id"`
	UID                     string        `json:"uid"`
	Position                float64       `json:"position"`
	Title                   string        `json:"title"`
	Text                    string        `json:"text,omitempty"`
	HTML                    string        `json:"html,omitempty"`
	StartedAt               string        `json:"startedAt,omitempty"`
	DuedAt                  string        `json:"duedAt,omitempty"`
	Timezone                string        `json:"timezone,omitempty"`
	Color                   string        `json:"color,omitempty"`
	Cover                   string        `json:"cover,omitempty"`
	CoverLocked             bool          `json:"coverLocked,omitempty"`
	Archived                bool          `json:"archived"`
	Done                    bool          `json:"done"`
	CommentCount            int           `json:"commentCount,omitempty"`
	ChecklistCount          int           `json:"checklistCount,omitempty"`
	ChecklistCompletedCount int           `json:"checklistCompletedCount,omitempty"`
	IsRepeating             bool          `json:"isRepeating,omitempty"`
	IsRead                  bool          `json:"isRead,omitempty"`
	IsSeen                  bool          `json:"isSeen,omitempty"`
	CreatedAt               string        `json:"createdAt"`
	UpdatedAt               string        `json:"updatedAt"`
	Users                   []User        `json:"users,omitempty"`
	Tags                    []Tag         `json:"tags,omitempty"`
	TodoList                *TodoListInfo `json:"todoList,omitempty"`
}

// TodoListInfo represents basic info about a todo list (for nested references)
type TodoListInfo struct {
	ID    string `json:"id"`
	UID   string `json:"uid"`
	Title string `json:"title"`
}

// CustomField represents a custom field with all possible properties
type CustomField struct {
	ID                    string              `json:"id"`
	UID                   string              `json:"uid"`
	Name                  string              `json:"name"`
	Type                  string              `json:"type"`
	Position              float64             `json:"position"`
	Description           string              `json:"description,omitempty"`
	ButtonType            string              `json:"buttonType,omitempty"`
	ButtonConfirmText     string              `json:"buttonConfirmText,omitempty"`
	CurrencyFieldID       string              `json:"currencyFieldId,omitempty"`
	ConversionDateType    string              `json:"conversionDateType,omitempty"`
	ConversionDate        string              `json:"conversionDate,omitempty"`
	Min                   *float64            `json:"min,omitempty"`
	Max                   *float64            `json:"max,omitempty"`
	Latitude              *float64            `json:"latitude,omitempty"`
	Longitude             *float64            `json:"longitude,omitempty"`
	StartDate             string              `json:"startDate,omitempty"`
	EndDate               string              `json:"endDate,omitempty"`
	Timezone              string              `json:"timezone,omitempty"`
	Currency              string              `json:"currency,omitempty"`
	Prefix                string              `json:"prefix,omitempty"`
	IsDueDate             bool                `json:"isDueDate,omitempty"`
	Formula               interface{}         `json:"formula,omitempty"`
	CreatedAt             string              `json:"createdAt"`
	UpdatedAt             string              `json:"updatedAt"`
	RegionCode            string              `json:"regionCode,omitempty"`
	CountryCodes          []string            `json:"countryCodes,omitempty"`
	Text                  string              `json:"text,omitempty"`
	Number                *float64            `json:"number,omitempty"`
	Checked               bool                `json:"checked,omitempty"`
	Editable              bool                `json:"editable"`
	Options               []CustomFieldOption `json:"customFieldOptions,omitempty"`
}

// CustomFieldOption represents an option for select-type custom fields
type CustomFieldOption struct {
	ID                   string `json:"id"`
	Title                string `json:"title"`
	ButtonType           string `json:"buttonType,omitempty"`
	ButtonConfirmText    string `json:"buttonConfirmText,omitempty"`
	Color                string `json:"color,omitempty"`
	CurrencyConversionTo string `json:"currencyConversionTo,omitempty"`
}

// CustomFieldValue represents a value assigned to a custom field
type CustomFieldValue struct {
	CustomFieldID string      `json:"customFieldId"`
	Value         interface{} `json:"value"`
}

// ============================================================================
// PAGINATION TYPES
// ============================================================================

// CursorPageInfo represents cursor-based pagination (GraphQL Relay-style)
type CursorPageInfo struct {
	HasNextPage     bool   `json:"hasNextPage"`
	HasPreviousPage bool   `json:"hasPreviousPage"`
	StartCursor     string `json:"startCursor"`
	EndCursor       string `json:"endCursor"`
}

// OffsetPageInfo represents offset-based pagination
type OffsetPageInfo struct {
	TotalPages      int  `json:"totalPages"`
	TotalItems      int  `json:"totalItems"`
	Page            int  `json:"page"`
	PerPage         int  `json:"perPage"`
	HasNextPage     bool `json:"hasNextPage"`
	HasPreviousPage bool `json:"hasPreviousPage"`
}

// ============================================================================
// RESPONSE WRAPPER TYPES
// ============================================================================

// MutationResult represents the result of a mutation operation
type MutationResult struct {
	Success     bool   `json:"success"`
	OperationID string `json:"operationId"`
}

// ============================================================================
// INPUT TYPES FOR MUTATIONS
// ============================================================================

// CreateProjectInput for project creation
type CreateProjectInput struct {
	Name        string `json:"name"`
	CompanyID   string `json:"companyId"`
	Description string `json:"description,omitempty"`
	Color       string `json:"color,omitempty"`
	Icon        string `json:"icon,omitempty"`
	Category    string `json:"category,omitempty"`
	TemplateID  string `json:"templateId,omitempty"`
}

// CreateTodoInput for record/todo creation
type CreateTodoInput struct {
	Title              string             `json:"title"`
	Description        string             `json:"description,omitempty"`
	TodoListID         string             `json:"todoListId"`
	ProjectID          string             `json:"projectId"`
	AssigneeIds        []string           `json:"assigneeIds,omitempty"`
	CustomFieldValues  []CustomFieldValue `json:"customFieldValues,omitempty"`
	TodoListPlacement  string             `json:"todoListPlacement,omitempty"`
	TodoListPlaceAfter string             `json:"todoListPlaceAfter,omitempty"`
}

// CreateCustomFieldInput for custom field creation
type CreateCustomFieldInput struct {
	Name                      string                      `json:"name"`
	Type                      string                      `json:"type"`
	ProjectID                 string                      `json:"projectId"`
	Description               string                      `json:"description,omitempty"`
	Min                       *float64                    `json:"min,omitempty"`
	Max                       *float64                    `json:"max,omitempty"`
	Currency                  string                      `json:"currency,omitempty"`
	Options                   []CustomFieldOptionInput    `json:"options,omitempty"`
	AllowOtherOption          bool                        `json:"allowOtherOption,omitempty"`
	OtherOptionPlaceholder    string                      `json:"otherOptionPlaceholder,omitempty"`
	ConversionRatesFromFields []ConversionRateFieldInput  `json:"conversionRatesFromFields,omitempty"`
	ConversionRatesFromStatic []ConversionRateStaticInput `json:"conversionRatesFromStatic,omitempty"`
}

// CustomFieldOptionInput for select field options
type CustomFieldOptionInput struct {
	Title string `json:"title"`
	Color string `json:"color,omitempty"`
}

// ConversionRateFieldInput for currency conversion from fields
type ConversionRateFieldInput struct {
	CurrencyFieldID string `json:"currencyFieldId"`
}

// ConversionRateStaticInput for static currency conversion
type ConversionRateStaticInput struct {
	Currency string  `json:"currency"`
	Rate     float64 `json:"rate"`
}

// DeleteTodoInput for record/todo deletion
type DeleteTodoInput struct {
	TodoID string `json:"todoId"`
}

// RecordCustomFieldValue represents a custom field value for a record
type RecordCustomFieldValue struct {
	ID            string      `json:"id"`
	CustomFieldID string      `json:"customFieldId"`
	Value         interface{} `json:"value"`
	CustomField   struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"customField"`
}

// ============================================================================
// FILE TYPES
// ============================================================================

// File represents a file in the Blue system
type File struct {
	ID        string `json:"id"`
	UID       string `json:"uid"`
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	Type      string `json:"type"`
	Extension string `json:"extension"`
}

// FilesResponse represents a response containing files
type FilesResponse struct {
	Files struct {
		Items    []File `json:"items"`
		PageInfo struct {
			TotalItems  int  `json:"totalItems"`
			HasNextPage bool `json:"hasNextPage"`
		} `json:"pageInfo"`
	} `json:"files"`
}