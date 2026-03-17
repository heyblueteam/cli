package records

import (
	"fmt"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

type CountResponse struct {
	TodoQueries struct {
		Todos struct {
			Items []struct {
				ID string `json:"id"`
			} `json:"items"`
			PageInfo struct {
				HasNextPage bool `json:"hasNextPage"`
			} `json:"pageInfo"`
		} `json:"todos"`
	} `json:"todoQueries"`
}

var countCmd = &cobra.Command{
	Use:   "count",
	Short: "Count records in a workspace",
	Long:  "Count the total number of records with optional filtering.",
	Example: `  blue records count --workspace <id>
  blue records count --workspace <id> --done false
  blue records count --workspace <id> --list <id>`,
	RunE: runCount,
}

var (
	countWorkspace string
	countList      string
	countDone      string
	countArchived  string
)

func init() {
	countCmd.Flags().StringVarP(&countWorkspace, "workspace", "w", "", "Workspace ID (required)")
	countCmd.Flags().StringVarP(&countList, "list", "l", "", "List ID to filter (optional)")
	countCmd.Flags().StringVar(&countDone, "done", "", "Filter by completion status (true/false)")
	countCmd.Flags().StringVar(&countArchived, "archived", "", "Filter by archived status (true/false)")
}

func runCount(cmd *cobra.Command, args []string) error {
	if countWorkspace == "" {
		return fmt.Errorf("workspace ID is required. Use --workspace flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := common.NewClient(config)

	query := `
		query CountRecords($filter: TodosFilter!, $limit: Int) {
			todoQueries {
				todos(filter: $filter, limit: $limit) {
					items { id }
					pageInfo { hasNextPage }
				}
			}
		}
	`

	filter := map[string]interface{}{
		"companyIds": []string{},
		"projectIds": []string{countWorkspace},
	}

	if countList != "" {
		filter["todoListIds"] = []string{countList}
	}
	if countDone == "true" {
		filter["done"] = true
	} else if countDone == "false" {
		filter["done"] = false
	}
	if countArchived == "true" {
		filter["archived"] = true
	} else if countArchived == "false" {
		filter["archived"] = false
	}

	variables := map[string]interface{}{
		"filter": filter,
		"limit":  500,
	}

	var response CountResponse
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to count records: %w", err)
	}

	totalCount := len(response.TodoQueries.Todos.Items)
	hasMore := response.TodoQueries.Todos.PageInfo.HasNextPage

	fmt.Printf("\n=== Record Count ===\n")
	fmt.Printf("Workspace ID: %s\n", countWorkspace)
	if countList != "" {
		fmt.Printf("List ID: %s\n", countList)
	}
	if countDone != "" {
		fmt.Printf("Completion filter: %s\n", countDone)
	}
	if countArchived != "" {
		fmt.Printf("Archived filter: %s\n", countArchived)
	}
	if hasMore {
		fmt.Printf("\nTotal Records: %d+ (more available)\n", totalCount)
	} else {
		fmt.Printf("\nTotal Records: %d\n", totalCount)
	}

	return nil
}
