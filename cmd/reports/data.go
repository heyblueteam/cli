package reports

import (
	"fmt"

	"github.com/spf13/cobra"
)

var dataCmd = &cobra.Command{
	Use:   "data",
	Short: "Read report records and count",
	Example: `  blue reports data --report <id> --limit 50
  blue reports data --report <id> --filter-json '{"done":false}' --format json`,
	RunE: runData,
}

var (
	dataReport     string
	dataLimit      int
	dataSkip       int
	dataSort       string
	dataFilterJSON string
	dataFormat     string
)

func init() {
	dataCmd.Flags().StringVar(&dataReport, "report", "", "Report ID (required)")
	dataCmd.Flags().IntVar(&dataLimit, "limit", 50, "Maximum records to return")
	dataCmd.Flags().IntVar(&dataSkip, "skip", 0, "Records to skip")
	dataCmd.Flags().StringVar(&dataSort, "sort", "", "Comma-separated TodosSort values")
	dataCmd.Flags().StringVar(&dataFilterJSON, "filter-json", "", "Raw TodosFilter JSON")
	dataCmd.Flags().StringVar(&dataFormat, "format", "", "Output format (json)")
}

func runData(cmd *cobra.Command, args []string) error {
	if dataReport == "" {
		return fmt.Errorf("report ID is required. Use --report flag")
	}
	filter, err := parseJSONFlag("--filter-json", dataFilterJSON)
	if err != nil {
		return err
	}
	var sort interface{}
	if dataSort != "" {
		sort = splitCSV(dataSort)
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	query := `
		query ReportData($id: String!, $filter: TodosFilter, $sort: [TodosSort!], $limit: Int, $skip: Int) {
			report(id: $id) {
				todoCount(filter: $filter)
				todos(filter: $filter, sort: $sort, limit: $limit, skip: $skip) {
					id title done startedAt duedAt createdAt updatedAt
				}
			}
		}
	`
	variables := map[string]interface{}{"id": dataReport, "filter": filter, "sort": sort, "limit": dataLimit, "skip": dataSkip}
	var response struct {
		Report struct {
			TodoCount int           `json:"todoCount"`
			Todos     []TodoSummary `json:"todos"`
		} `json:"report"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to read report data: %w", err)
	}
	if dataFormat == "json" {
		return printJSON(response.Report)
	}
	fmt.Printf("Total matching records: %d\n", response.Report.TodoCount)
	for _, todo := range response.Report.Todos {
		fmt.Printf("%s  [%t] %s\n", todo.ID, todo.Done, todo.Title)
	}
	return nil
}
