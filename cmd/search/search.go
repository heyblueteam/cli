package search

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

// Cmd searches API-supported resources. Today this is intentionally limited to records.
var Cmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search records by name",
	Long:  "Search records by name in a workspace using the Blue API record search filter.",
	Example: `  blue search "launch" --workspace <id-or-slug>
  blue search "invoice" --workspace <id-or-slug> --format json
  blue search "bug" --workspace <id-or-slug> --done false --limit 50`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

var (
	searchWorkspace string
	searchFormat    string
	searchLimit     int
	searchDone      string
	searchArchived  string
)

type searchRecord struct {
	ID        string `json:"id"`
	UID       string `json:"uid"`
	Title     string `json:"title"`
	Done      bool   `json:"done"`
	Archived  bool   `json:"archived"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
	TodoList  *struct {
		ID    string `json:"id"`
		Title string `json:"title"`
	} `json:"todoList"`
	Users []struct {
		ID       string `json:"id"`
		FullName string `json:"fullName"`
	} `json:"users"`
	Tags []struct {
		ID    string `json:"id"`
		Title string `json:"title"`
		Color string `json:"color"`
	} `json:"tags"`
}

func init() {
	Cmd.Flags().StringVarP(&searchWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	Cmd.Flags().StringVar(&searchFormat, "format", "text", "Output format: text, json, csv")
	Cmd.Flags().IntVar(&searchLimit, "limit", 20, "Maximum records to return")
	Cmd.Flags().StringVar(&searchDone, "done", "", "Filter by completion status (true/false)")
	Cmd.Flags().StringVar(&searchArchived, "archived", "false", "Filter by archived status (true/false)")
}

func runSearch(cmd *cobra.Command, args []string) error {
	if searchWorkspace == "" {
		return fmt.Errorf("workspace ID or slug is required. Use --workspace flag")
	}
	if searchLimit <= 0 {
		return fmt.Errorf("limit must be greater than 0")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)
	client.SetProject(searchWorkspace)
	projectID, err := client.ResolveProjectID(searchWorkspace)
	if err != nil {
		return fmt.Errorf("failed to resolve workspace: %w", err)
	}
	client.SetProject(projectID)

	records, err := searchRecords(client, projectID, args[0])
	if err != nil {
		return err
	}

	return printSearchResults(records)
}

func searchRecords(client *common.Client, projectID, queryText string) ([]searchRecord, error) {
	query := `query SearchRecords($filter: TodosFilter!, $limit: Int!) {
		todoQueries {
			todos(filter: $filter, limit: $limit) {
				items {
					id uid title done archived createdAt updatedAt
					todoList { id title }
					users { id fullName }
					tags { id title color }
				}
			}
		}
	}`

	filter := map[string]interface{}{
		"companyIds": []string{},
		"projectIds": []string{projectID},
		"q":          queryText,
	}
	if searchDone == "true" {
		filter["done"] = true
	} else if searchDone == "false" {
		filter["done"] = false
	} else if searchDone != "" {
		return nil, fmt.Errorf("invalid done value %q. Use true or false", searchDone)
	}
	if searchArchived == "true" {
		filter["archived"] = true
	} else if searchArchived == "false" {
		filter["archived"] = false
	} else if searchArchived != "" {
		return nil, fmt.Errorf("invalid archived value %q. Use true or false", searchArchived)
	}

	variables := map[string]interface{}{"filter": filter, "limit": searchLimit}
	var response struct {
		TodoQueries struct {
			Todos struct {
				Items []searchRecord `json:"items"`
			} `json:"todos"`
		} `json:"todoQueries"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return nil, fmt.Errorf("failed to search records: %w", err)
	}

	return response.TodoQueries.Todos.Items, nil
}

func printSearchResults(records []searchRecord) error {
	switch searchFormat {
	case "text":
		if len(records) == 0 {
			fmt.Println("No matching records found.")
			return nil
		}
		for i, record := range records {
			fmt.Printf("%d. %s\n", i+1, record.Title)
			fmt.Printf("   ID: %s\n", record.ID)
			if record.UID != "" {
				fmt.Printf("   UID: %s\n", record.UID)
			}
			if record.TodoList != nil {
				fmt.Printf("   List: %s (%s)\n", record.TodoList.Title, record.TodoList.ID)
			}
			fmt.Printf("   Done: %v\n", record.Done)
			fmt.Println()
		}
		return nil
	case "json":
		out, err := json.MarshalIndent(records, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(out))
		return nil
	case "csv":
		writer := csv.NewWriter(os.Stdout)
		if err := writer.Write([]string{"id", "uid", "title", "list_id", "list_title", "done", "archived", "created_at", "updated_at"}); err != nil {
			return err
		}
		for _, record := range records {
			listID, listTitle := "", ""
			if record.TodoList != nil {
				listID = record.TodoList.ID
				listTitle = record.TodoList.Title
			}
			if err := writer.Write([]string{
				record.ID,
				record.UID,
				record.Title,
				listID,
				listTitle,
				strconv.FormatBool(record.Done),
				strconv.FormatBool(record.Archived),
				record.CreatedAt,
				record.UpdatedAt,
			}); err != nil {
				return err
			}
		}
		writer.Flush()
		return writer.Error()
	default:
		return fmt.Errorf("invalid format %q. Use text, json, or csv", searchFormat)
	}
}
