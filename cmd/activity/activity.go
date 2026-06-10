package activity

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"html"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

// Cmd shows recent company or workspace activity.
var Cmd = &cobra.Command{
	Use:   "activity",
	Short: "Show recent activity",
	Long:  "Show recent company or workspace activity from the Blue activity feed.",
	Example: `  blue activity
  blue activity --workspace <id-or-slug> --since 7d
  blue activity --workspace <id-or-slug> --category CREATE_TODO,CREATE_COMMENT
  blue activity record <record-id> --workspace <id-or-slug>
  blue activity --format json`,
	RunE: runActivity,
}

var (
	activityWorkspace  string
	activitySince      string
	activityStart      string
	activityEnd        string
	activityUser       string
	activityCategories string
	activityLimit      int
	activityAfter      string
	activityFormat     string
)

type activityItem struct {
	ID        string `json:"id"`
	UID       string `json:"uid"`
	Category  string `json:"category"`
	Text      string `json:"text"`
	HTML      string `json:"html"`
	IsSeen    bool   `json:"isSeen"`
	IsRead    bool   `json:"isRead"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	CreatedBy struct {
		ID       string `json:"id"`
		FullName string `json:"fullName"`
		Email    string `json:"email"`
	} `json:"createdBy"`
	AffectedBy *struct {
		ID       string `json:"id"`
		FullName string `json:"fullName"`
		Email    string `json:"email"`
	} `json:"affectedBy"`
	Project *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Slug string `json:"slug"`
	} `json:"project"`
	Todo *struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	} `json:"todo"`
	TodoList *struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	} `json:"todoList"`
	InviteeEmail string `json:"inviteeEmail"`
	Metadata     string `json:"metadata"`
}

type activityResult struct {
	Activities []activityItem `json:"activities"`
	PageInfo   struct {
		HasNextPage bool   `json:"hasNextPage"`
		EndCursor   string `json:"endCursor,omitempty"`
	} `json:"pageInfo"`
	TotalCount int `json:"totalCount"`
}

var htmlTagPattern = regexp.MustCompile(`<[^>]*>`)

func init() {
	Cmd.Flags().StringVarP(&activityWorkspace, "workspace", "w", "", "Workspace ID or slug; omit for company-wide activity")
	Cmd.Flags().StringVar(&activitySince, "since", "", "Start time shortcut (for example 24h, 7d, 30d) or ISO date/time")
	Cmd.Flags().StringVar(&activityStart, "start", "", "Start date/time as ISO 8601")
	Cmd.Flags().StringVar(&activityEnd, "end", "", "End date/time as ISO 8601")
	Cmd.Flags().StringVar(&activityUser, "user", "", "Filter by user ID")
	Cmd.Flags().StringVar(&activityCategories, "category", "", "Filter by activity categories (comma-separated)")
	Cmd.Flags().IntVar(&activityLimit, "limit", 20, "Maximum activities to return")
	Cmd.Flags().StringVar(&activityAfter, "after", "", "Cursor activity ID for the next page")
	Cmd.Flags().StringVar(&activityFormat, "format", "text", "Output format: text, json, csv")
}

func runActivity(cmd *cobra.Command, args []string) error {
	if activityLimit <= 0 {
		return fmt.Errorf("limit must be greater than 0")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)

	projectID := ""
	if activityWorkspace != "" {
		client.SetProject(activityWorkspace)
		projectID, err = client.ResolveProjectID(activityWorkspace)
		if err != nil {
			return fmt.Errorf("failed to resolve workspace: %w", err)
		}
		client.SetProject(projectID)
	}

	startDate := activityStart
	if activitySince != "" {
		startDate, err = parseSince(activitySince)
		if err != nil {
			return err
		}
	}

	result, err := fetchActivity(client, projectID, startDate)
	if err != nil {
		return err
	}

	return printActivity(result)
}

func fetchActivity(client *common.Client, projectID, startDate string) (activityResult, error) {
	query := `query ActivityFeed(
		$projectId: String
		$userId: String
		$categories: [ActivityCategory!]
		$startDate: DateTime
		$endDate: DateTime
		$first: Int
		$after: String
	) {
		activityList(
			projectId: $projectId
			userId: $userId
			categories: $categories
			startDate: $startDate
			endDate: $endDate
			first: $first
			after: $after
			orderBy: createdAt_DESC
		) {
			activities {
				id uid category text html isSeen isRead createdAt updatedAt inviteeEmail metadata
				createdBy { id fullName email }
				affectedBy { id fullName email }
				project { id name slug }
				todo { id title }
				todoList { id title }
			}
			pageInfo { hasNextPage endCursor }
			totalCount
		}
	}`

	variables := map[string]interface{}{
		"projectId":  nil,
		"userId":     nil,
		"categories": nil,
		"startDate":  nil,
		"endDate":    nil,
		"first":      activityLimit,
		"after":      nil,
	}
	if projectID != "" {
		variables["projectId"] = projectID
	}
	if activityUser != "" {
		variables["userId"] = activityUser
	}
	if activityCategories != "" {
		variables["categories"] = splitCSV(activityCategories)
	}
	if startDate != "" {
		variables["startDate"] = startDate
	}
	if activityEnd != "" {
		variables["endDate"] = activityEnd
	}
	if activityAfter != "" {
		variables["after"] = activityAfter
	}

	var response struct {
		ActivityList activityResult `json:"activityList"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return activityResult{}, fmt.Errorf("failed to fetch activity: %w", err)
	}

	return response.ActivityList, nil
}

func printActivity(result activityResult) error {
	switch activityFormat {
	case "text":
		if len(result.Activities) == 0 {
			fmt.Println("No activity found.")
			return nil
		}
		for i, item := range result.Activities {
			fmt.Printf("%d. %s\n", i+1, item.DisplayText())
			fmt.Printf("   Category: %s\n", item.Category)
			fmt.Printf("   Created:  %s\n", item.CreatedAt)
			fmt.Printf("   Actor:    %s <%s>\n", item.CreatedBy.FullName, item.CreatedBy.Email)
			if item.Project != nil {
				fmt.Printf("   Workspace: %s (%s)\n", item.Project.Name, item.Project.ID)
			}
			if item.Todo != nil {
				fmt.Printf("   Record:   %s (%s)\n", item.Todo.Title, item.Todo.ID)
			}
			fmt.Printf("   ID:       %s\n\n", item.ID)
		}
		if result.PageInfo.HasNextPage {
			cursor := result.PageInfo.EndCursor
			if cursor == "" && len(result.Activities) > 0 {
				cursor = result.Activities[len(result.Activities)-1].ID
			}
			fmt.Printf("Next page: %s --after %s\n", nextPageCommand(), cursor)
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
		if err := writer.Write([]string{"id", "category", "text", "created_at", "actor_id", "actor_name", "actor_email", "workspace_id", "workspace_name", "record_id", "record_title", "seen", "read"}); err != nil {
			return err
		}
		for _, item := range result.Activities {
			workspaceID, workspaceName := "", ""
			if item.Project != nil {
				workspaceID = item.Project.ID
				workspaceName = item.Project.Name
			}
			recordID, recordTitle := "", ""
			if item.Todo != nil {
				recordID = item.Todo.ID
				recordTitle = item.Todo.Title
			}
			if err := writer.Write([]string{item.ID, item.Category, item.DisplayText(), item.CreatedAt, item.CreatedBy.ID, item.CreatedBy.FullName, item.CreatedBy.Email, workspaceID, workspaceName, recordID, recordTitle, strconv.FormatBool(item.IsSeen), strconv.FormatBool(item.IsRead)}); err != nil {
				return err
			}
		}
		writer.Flush()
		return writer.Error()
	default:
		return fmt.Errorf("invalid format %q. Use text, json, or csv", activityFormat)
	}
}

func nextPageCommand() string {
	parts := []string{"blue", "activity"}
	if activityWorkspace != "" {
		parts = append(parts, "--workspace", activityWorkspace)
	}
	if activitySince != "" {
		parts = append(parts, "--since", activitySince)
	}
	if activityStart != "" {
		parts = append(parts, "--start", activityStart)
	}
	if activityEnd != "" {
		parts = append(parts, "--end", activityEnd)
	}
	if activityUser != "" {
		parts = append(parts, "--user", activityUser)
	}
	if activityCategories != "" {
		parts = append(parts, "--category", activityCategories)
	}
	if activityLimit != 20 {
		parts = append(parts, "--limit", strconv.Itoa(activityLimit))
	}
	return strings.Join(parts, " ")
}

func (item activityItem) DisplayText() string {
	if strings.TrimSpace(item.Text) != "" {
		return item.Text
	}
	return plainFromHTML(item.HTML)
}

func plainFromHTML(value string) string {
	plain := htmlTagPattern.ReplaceAllString(value, "")
	plain = html.UnescapeString(plain)
	return strings.Join(strings.Fields(plain), " ")
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func parseSince(value string) (string, error) {
	value = strings.TrimSpace(value)
	if strings.HasSuffix(value, "d") {
		days, err := strconv.Atoi(strings.TrimSuffix(value, "d"))
		if err != nil || days < 0 {
			return "", fmt.Errorf("invalid since value %q", value)
		}
		return time.Now().UTC().AddDate(0, 0, -days).Format(time.RFC3339), nil
	}
	if duration, err := time.ParseDuration(value); err == nil {
		return time.Now().UTC().Add(-duration).Format(time.RFC3339), nil
	}
	if parsed, err := time.Parse(time.RFC3339, value); err == nil {
		return parsed.UTC().Format(time.RFC3339), nil
	}
	if parsed, err := time.Parse("2006-01-02", value); err == nil {
		return parsed.UTC().Format(time.RFC3339), nil
	}
	return "", fmt.Errorf("invalid since value %q. Use values like 24h, 7d, 2026-06-10, or RFC3339", value)
}
