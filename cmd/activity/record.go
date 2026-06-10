package activity

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

var recordCmd = &cobra.Command{
	Use:   "record <record-id>",
	Short: "Show record-level activity",
	Long:  "Show raw-but-readable comments and actions from a single record's activity feed.",
	Example: `  blue activity record <record-id> --workspace <id-or-slug>
  blue activity record <record-id> --workspace <id-or-slug> --type comments
  blue activity record <record-id> --workspace <id-or-slug> --format json`,
	Args: cobra.ExactArgs(1),
	RunE: runRecordActivity,
}

var (
	recordActivityWorkspace string
	recordActivityType      string
	recordActivityLimit     int
	recordActivitySkip      int
	recordActivityFormat    string
	recordActivityRecordID  string
)

type recordActivityResult struct {
	Items      []recordActivityItem `json:"items"`
	TotalCount int                  `json:"totalCount"`
}

type recordActivityItem struct {
	Typename  string `json:"__typename"`
	ID        string `json:"id"`
	CreatedAt string `json:"createdAt"`

	// Comment fields.
	HTML        string `json:"html,omitempty"`
	DeletedAt   string `json:"deletedAt,omitempty"`
	CommentUser *struct {
		ID       string `json:"id"`
		FullName string `json:"fullName"`
	} `json:"commentUser,omitempty"`

	// TodoAction fields.
	ActionType string `json:"type,omitempty"`
	OldValue   string `json:"oldValue,omitempty"`
	NewValue   string `json:"newValue,omitempty"`
	Automated  bool   `json:"automated,omitempty"`
	ActionUser *struct {
		ID       string `json:"id"`
		FullName string `json:"fullName"`
	} `json:"actionUser,omitempty"`
	AffectedBy *struct {
		ID       string `json:"id"`
		FullName string `json:"fullName"`
	} `json:"affectedBy,omitempty"`
	CustomField *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"customField,omitempty"`
}

func init() {
	Cmd.AddCommand(recordCmd)
	recordCmd.Flags().StringVarP(&recordActivityWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	recordCmd.Flags().StringVar(&recordActivityType, "type", "all", "Activity type: all, comments, actions")
	recordCmd.Flags().IntVar(&recordActivityLimit, "limit", 20, "Maximum items to return (1-24; API caps record activity pages)")
	recordCmd.Flags().IntVar(&recordActivitySkip, "skip", 0, "Number of items to skip")
	recordCmd.Flags().StringVar(&recordActivityFormat, "format", "text", "Output format: text, json, csv")
}

func runRecordActivity(cmd *cobra.Command, args []string) error {
	if recordActivityWorkspace == "" {
		return fmt.Errorf("workspace ID or slug is required. Use --workspace flag")
	}
	if recordActivityLimit <= 0 || recordActivityLimit > 24 {
		return fmt.Errorf("limit must be between 1 and 24")
	}
	if recordActivitySkip < 0 {
		return fmt.Errorf("skip must be 0 or greater")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)
	client.SetProject(recordActivityWorkspace)
	projectID, err := client.ResolveProjectID(recordActivityWorkspace)
	if err != nil {
		return fmt.Errorf("failed to resolve workspace: %w", err)
	}
	client.SetProject(projectID)

	recordActivityRecordID = args[0]
	result, err := fetchRecordActivity(client, config.CompanyID, projectID, recordActivityRecordID)
	if err != nil {
		return err
	}

	return printRecordActivity(result)
}

func fetchRecordActivity(client *common.Client, companyID, projectID, recordID string) (recordActivityResult, error) {
	query := `query RecordActivity($filter: TodoActivityFilter!, $orderBy: TodoActivityOrderBy, $limit: Int, $skip: Int) {
		todoActivity(filter: $filter, orderBy: $orderBy, limit: $limit, skip: $skip) {
			items {
				__typename
				... on Comment {
					id html createdAt deletedAt
					commentUser: user { id fullName }
				}
				... on TodoAction {
					id type oldValue newValue automated createdAt
					actionUser: user { id fullName }
					affectedBy { id fullName }
					customField { id name }
				}
			}
			totalCount
		}
	}`

	filter := map[string]interface{}{
		"companyId": companyID,
		"projectId": projectID,
		"todoId":    recordID,
	}
	switch recordActivityType {
	case "all", "":
	case "comments":
		filter["type"] = "comments"
	case "actions":
		filter["type"] = "activities"
	default:
		return recordActivityResult{}, fmt.Errorf("invalid type %q. Use all, comments, or actions", recordActivityType)
	}

	variables := map[string]interface{}{
		"filter":  filter,
		"orderBy": map[string]interface{}{"createdAt": "DESC", "updatedAt": "DESC"},
		"limit":   recordActivityLimit,
		"skip":    recordActivitySkip,
	}

	var response struct {
		TodoActivity recordActivityResult `json:"todoActivity"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return recordActivityResult{}, fmt.Errorf("failed to fetch record activity: %w", err)
	}

	return response.TodoActivity, nil
}

func printRecordActivity(result recordActivityResult) error {
	switch recordActivityFormat {
	case "text":
		if len(result.Items) == 0 {
			fmt.Println("No record activity found.")
			return nil
		}
		for i, item := range result.Items {
			fmt.Printf("%d. %s\n", i+1, item.Summary())
			fmt.Printf("   Kind:    %s\n", item.Kind())
			fmt.Printf("   Created: %s\n", item.CreatedAt)
			if actor := item.Actor(); actor != nil {
				fmt.Printf("   Actor:   %s (%s)\n", actor.FullName, actor.ID)
			}
			if item.AffectedBy != nil {
				fmt.Printf("   Affected: %s (%s)\n", item.AffectedBy.FullName, item.AffectedBy.ID)
			}
			if item.CustomField != nil {
				fmt.Printf("   Field:   %s (%s)\n", item.CustomField.Name, item.CustomField.ID)
			}
			if item.OldValue != "" || item.NewValue != "" {
				fmt.Printf("   Old:     %s\n", item.OldValue)
				fmt.Printf("   New:     %s\n", item.NewValue)
			}
			fmt.Printf("   ID:      %s\n\n", item.ID)
		}
		if recordActivitySkip+len(result.Items) < result.TotalCount {
			fmt.Printf("Next page: %s --skip %d\n", recordActivityNextPageCommand(), recordActivitySkip+len(result.Items))
		}
		return nil
	case "json":
		out, err := json.MarshalIndent(result, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(out))
		return nil
	case "csv":
		writer := csv.NewWriter(os.Stdout)
		if err := writer.Write([]string{"id", "kind", "summary", "created_at", "actor_id", "actor_name", "type", "old_value", "new_value", "custom_field_id", "custom_field_name", "automated"}); err != nil {
			return err
		}
		for _, item := range result.Items {
			actorID, actorName := "", ""
			if actor := item.Actor(); actor != nil {
				actorID = actor.ID
				actorName = actor.FullName
			}
			fieldID, fieldName := "", ""
			if item.CustomField != nil {
				fieldID = item.CustomField.ID
				fieldName = item.CustomField.Name
			}
			if err := writer.Write([]string{item.ID, item.Kind(), item.Summary(), item.CreatedAt, actorID, actorName, item.ActionType, item.OldValue, item.NewValue, fieldID, fieldName, strconv.FormatBool(item.Automated)}); err != nil {
				return err
			}
		}
		writer.Flush()
		return writer.Error()
	default:
		return fmt.Errorf("invalid format %q. Use text, json, or csv", recordActivityFormat)
	}
}

func recordActivityNextPageCommand() string {
	parts := []string{"blue", "activity", "record", recordActivityRecordID, "--workspace", recordActivityWorkspace}
	if recordActivityType != "all" && recordActivityType != "" {
		parts = append(parts, "--type", recordActivityType)
	}
	if recordActivityLimit != 20 {
		parts = append(parts, "--limit", strconv.Itoa(recordActivityLimit))
	}
	return strings.Join(parts, " ")
}

func (item recordActivityItem) Kind() string {
	switch item.Typename {
	case "Comment":
		return "comment"
	case "TodoAction":
		return "action"
	default:
		return item.Typename
	}
}

func (item recordActivityItem) Actor() *struct {
	ID       string `json:"id"`
	FullName string `json:"fullName"`
} {
	if item.CommentUser != nil {
		return item.CommentUser
	}
	return item.ActionUser
}

func (item recordActivityItem) Summary() string {
	if item.Typename == "Comment" {
		text := plainFromHTML(item.HTML)
		if text == "" {
			text = "Comment"
		}
		return text
	}

	parts := []string{item.ActionType}
	if item.OldValue != "" || item.NewValue != "" {
		parts = append(parts, fmt.Sprintf("old=%q new=%q", item.OldValue, item.NewValue))
	}
	return strings.Join(parts, " ")
}
