package records

import (
	"fmt"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

// DetailedRecord represents a record with all possible fields
type DetailedRecord struct {
	ID                      string               `json:"id"`
	UID                     string               `json:"uid"`
	Position                float64              `json:"position"`
	Title                   string               `json:"title"`
	Text                    string               `json:"text,omitempty"`
	HTML                    string               `json:"html,omitempty"`
	StartedAt               string               `json:"startedAt,omitempty"`
	DuedAt                  string               `json:"duedAt,omitempty"`
	Timezone                string               `json:"timezone,omitempty"`
	Color                   string               `json:"color,omitempty"`
	Cover                   string               `json:"cover,omitempty"`
	CoverLocked             bool                 `json:"coverLocked,omitempty"`
	Archived                bool                 `json:"archived"`
	Done                    bool                 `json:"done"`
	CommentCount            int                  `json:"commentCount,omitempty"`
	ChecklistCount          int                  `json:"checklistCount,omitempty"`
	ChecklistCompletedCount int                  `json:"checklistCompletedCount,omitempty"`
	IsRepeating             bool                 `json:"isRepeating,omitempty"`
	CreatedAt               string               `json:"createdAt"`
	UpdatedAt               string               `json:"updatedAt"`
	Users                   []common.User        `json:"users,omitempty"`
	Tags                    []common.Tag         `json:"tags,omitempty"`
	TodoList                *common.TodoListInfo `json:"todoList,omitempty"`
	CustomFields            []SimpleFieldValue   `json:"customFields,omitempty"`
}

type SimpleFieldValue struct {
	ID    string      `json:"id"`
	Value interface{} `json:"value"`
}

type FieldInfo struct {
	ID   string
	Name string
	Type string
}

type TodoRecordResponse struct {
	Todo DetailedRecord `json:"todo"`
}

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get detailed record information",
	Long:  "Retrieve detailed information about a specific record.",
	Example: `  blue records get --record <id> --workspace <id>
  blue records get -r <id> -w <id> --simple`,
	RunE: runGet,
}

var (
	getRecord    string
	getWorkspace string
	getSimple    bool
)

func init() {
	getCmd.Flags().StringVarP(&getRecord, "record", "r", "", "Record ID (required)")
	getCmd.Flags().StringVarP(&getWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	getCmd.Flags().BoolVarP(&getSimple, "simple", "s", false, "Show only basic record information")
}

func runGet(cmd *cobra.Command, args []string) error {
	if getRecord == "" || getWorkspace == "" {
		return fmt.Errorf("--record and --workspace flags are required")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := common.NewClient(config)
	client.SetProject(getWorkspace)

	// Fetch custom field metadata
	fieldInfo, err := fetchFieldInfo(client)
	if err != nil {
		fieldInfo = make(map[string]FieldInfo)
	}

	query := buildRecordDetailQuery(getSimple, getRecord)

	var response TodoRecordResponse
	if err := client.ExecuteQueryWithResult(query, nil, &response); err != nil {
		return fmt.Errorf("failed to get record: %w", err)
	}

	record := response.Todo
	if record.ID == "" {
		return fmt.Errorf("record '%s' not found", getRecord)
	}

	displayRecordDetails(record, getSimple, fieldInfo)
	return nil
}

func displayRecordDetails(record DetailedRecord, simple bool, fieldInfo map[string]FieldInfo) {
	fmt.Printf("\n=== Record Details ===\n")
	fmt.Printf("ID: %s\n", record.ID)
	fmt.Printf("Title: %s\n", record.Title)

	if record.TodoList != nil {
		fmt.Printf("List: %s (%s)\n", record.TodoList.Title, record.TodoList.ID)
	}

	status := "Active"
	if record.Archived {
		status = "Archived"
	} else if record.Done {
		status = "Completed"
	}
	fmt.Printf("Status: %s\n", status)

	if !simple {
		if record.Text != "" {
			fmt.Printf("Description: %s\n", record.Text)
		}
		if record.StartedAt != "" {
			fmt.Printf("Started At: %s\n", record.StartedAt)
		}
		if record.DuedAt != "" {
			fmt.Printf("Due At: %s\n", record.DuedAt)
		}
		if record.Color != "" {
			fmt.Printf("Color: %s\n", record.Color)
		}
		fmt.Printf("Comments: %d\n", record.CommentCount)
		fmt.Printf("Checklists: %d/%d completed\n", record.ChecklistCompletedCount, record.ChecklistCount)

		if len(record.Users) > 0 {
			fmt.Printf("\n=== Assignees (%d) ===\n", len(record.Users))
			for _, user := range record.Users {
				fmt.Printf("- %s (%s) [%s]\n", user.FullName, user.Email, user.ID)
			}
		}

		if len(record.Tags) > 0 {
			fmt.Printf("\n=== Tags (%d) ===\n", len(record.Tags))
			for _, tag := range record.Tags {
				fmt.Printf("- %s [%s] (%s)\n", tag.Title, tag.Color, tag.ID)
			}
		}

		if len(record.CustomFields) > 0 {
			fmt.Printf("\n=== Custom Fields (%d) ===\n", len(record.CustomFields))
			for _, cfv := range record.CustomFields {
				fieldDisplay := cfv.ID
				if info, exists := fieldInfo[cfv.ID]; exists {
					fieldDisplay = fmt.Sprintf("%s (%s) [%s]", info.Name, info.Type, cfv.ID)
				}
				parsedValue := parseFieldValue(cfv.Value)
				if parsedValue != nil {
					fmt.Printf("- %s: %v\n", fieldDisplay, parsedValue)
				} else {
					fmt.Printf("- %s: (empty)\n", fieldDisplay)
				}
			}
		}

		fmt.Printf("Created At: %s\n", record.CreatedAt)
		fmt.Printf("Updated At: %s\n", record.UpdatedAt)
	} else {
		if record.DuedAt != "" {
			fmt.Printf("Due: %s\n", record.DuedAt)
		}
		if len(record.Users) > 0 {
			fmt.Printf("Assignees: %d\n", len(record.Users))
		}
		if len(record.Tags) > 0 {
			fmt.Printf("Tags: %d\n", len(record.Tags))
		}
	}
}

func buildRecordDetailQuery(simple bool, recordID string) string {
	if simple {
		return fmt.Sprintf(`
			query GetRecord {
				todo(id: "%s") {
					id
					uid
					title
					done
					archived
					duedAt
					users { id fullName email }
					tags { id title color }
					todoList { id title }
					customFields { id value }
				}
			}
		`, recordID)
	}

	return fmt.Sprintf(`
		query GetRecord {
			todo(id: "%s") {
				id
				uid
				position
				title
				text
				html
				startedAt
				duedAt
				timezone
				color
				cover
				coverLocked
				archived
				done
				commentCount
				checklistCount
				checklistCompletedCount
				isRepeating
				createdAt
				updatedAt
				users { id uid firstName lastName fullName email }
				tags { id uid title color }
				todoList { id uid title }
				customFields { id value }
			}
		}
	`, recordID)
}

func fetchFieldInfo(client *common.Client) (map[string]FieldInfo, error) {
	query := `
		query GetProjectCustomFields {
			customFields {
				items { id name type }
			}
		}
	`

	var response struct {
		CustomFields struct {
			Items []FieldInfo `json:"items"`
		} `json:"customFields"`
	}

	if err := client.ExecuteQueryWithResult(query, nil, &response); err != nil {
		return nil, err
	}

	fieldMap := make(map[string]FieldInfo)
	for _, field := range response.CustomFields.Items {
		fieldMap[field.ID] = field
	}
	return fieldMap, nil
}

func parseFieldValue(value interface{}) interface{} {
	if value == nil {
		return nil
	}
	if valueMap, ok := value.(map[string]interface{}); ok {
		for _, key := range []string{"number", "currency", "text", "date", "boolean"} {
			if v := valueMap[key]; v != nil {
				return v
			}
		}
		return value
	}
	return value
}
