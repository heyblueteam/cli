package reports

type Report struct {
	ID              string                 `json:"id"`
	UID             string                 `json:"uid"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Config          map[string]interface{} `json:"config"`
	ProjectIDs      []string               `json:"projectIds"`
	LastGeneratedAt string                 `json:"lastGeneratedAt"`
	CreatedAt       string                 `json:"createdAt"`
	UpdatedAt       string                 `json:"updatedAt"`
	CreatedBy       User                   `json:"createdBy"`
	DataSources     []ReportDataSource     `json:"dataSources"`
	ReportUsers     []ReportUser           `json:"reportUsers"`
}

type User struct {
	ID       string `json:"id"`
	FullName string `json:"fullName"`
	Email    string `json:"email"`
}

type ReportDataSource struct {
	ID         string                 `json:"id"`
	UID        string                 `json:"uid"`
	Name       string                 `json:"name"`
	SourceType string                 `json:"sourceType"`
	ProjectIDs []string               `json:"projectIds"`
	Filters    map[string]interface{} `json:"filters"`
	Order      int                    `json:"order"`
	CreatedAt  string                 `json:"createdAt"`
	UpdatedAt  string                 `json:"updatedAt"`
}

type ReportUser struct {
	ID   string `json:"id"`
	Role string `json:"role"`
	User User   `json:"user"`
}

type TodoSummary struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Done      bool   `json:"done"`
	StartedAt string `json:"startedAt"`
	DuedAt    string `json:"duedAt"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

type FieldAggregation struct {
	Field     string   `json:"field"`
	FieldName string   `json:"fieldName"`
	FieldType string   `json:"fieldType"`
	Sum       *float64 `json:"sum"`
	Avg       *float64 `json:"avg"`
	Min       *float64 `json:"min"`
	Max       *float64 `json:"max"`
	Count     int      `json:"count"`
}

var reportFields = `
	id
	uid
	title
	description
	config
	projectIds
	lastGeneratedAt
	createdAt
	updatedAt
	createdBy { id fullName email }
	dataSources { id uid name sourceType projectIds filters order createdAt updatedAt }
	reportUsers { id role user { id fullName email } }
`
