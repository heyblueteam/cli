package documents

type Document struct {
	ID        string `json:"id"`
	UID       string `json:"uid"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Wiki      bool   `json:"wiki"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	Project   struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"project"`
	CreatedBy struct {
		ID       string `json:"id"`
		FullName string `json:"fullName"`
		Email    string `json:"email"`
	} `json:"createdBy"`
}

var documentFields = `
	id
	uid
	title
	content
	wiki
	createdAt
	updatedAt
	project { id name }
	createdBy { id fullName email }
`
