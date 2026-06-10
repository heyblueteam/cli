package reports

import (
	"fmt"

	"github.com/spf13/cobra"
)

var aggregateCmd = &cobra.Command{
	Use:   "aggregate",
	Short: "Read report field aggregations",
	Example: `  blue reports aggregate --report <id> --field field_123:number
  blue reports aggregate --report <id> --fields-json '[{"field":"field_123","fieldType":"number"}]'`,
	RunE: runAggregate,
}

var (
	aggReport     string
	aggField      string
	aggFieldsJSON string
	aggFilterJSON string
)

func init() {
	aggregateCmd.Flags().StringVar(&aggReport, "report", "", "Report ID (required)")
	aggregateCmd.Flags().StringVar(&aggField, "field", "", "Single field as fieldId[:fieldType]")
	aggregateCmd.Flags().StringVar(&aggFieldsJSON, "fields-json", "", "Raw AggregationFieldInput array JSON")
	aggregateCmd.Flags().StringVar(&aggFilterJSON, "filter-json", "", "Raw TodosFilter JSON")
}

func runAggregate(cmd *cobra.Command, args []string) error {
	if aggReport == "" {
		return fmt.Errorf("report ID is required. Use --report flag")
	}
	var fields interface{}
	if aggFieldsJSON != "" {
		parsed, err := parseJSONFlag("--fields-json", aggFieldsJSON)
		if err != nil {
			return err
		}
		fields = parsed
	} else if aggField != "" {
		parts := splitCSV(aggField)
		field := parts[0]
		fieldType := ""
		if colon := stringsIndex(field, ':'); colon >= 0 {
			fieldType = field[colon+1:]
			field = field[:colon]
		}
		entry := map[string]interface{}{"field": field}
		if fieldType != "" {
			entry["fieldType"] = fieldType
		}
		fields = []interface{}{entry}
	} else {
		return fmt.Errorf("aggregation field is required. Use --field or --fields-json")
	}
	filter, err := parseJSONFlag("--filter-json", aggFilterJSON)
	if err != nil {
		return err
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	query := `query ReportAggregations($id: String!, $fields: [AggregationFieldInput!]!, $filter: TodosFilter) { report(id: $id) { aggregations(fields: $fields, filter: $filter) { field fieldName fieldType sum avg min max count } } }`
	var response struct {
		Report struct {
			Aggregations []FieldAggregation `json:"aggregations"`
		} `json:"report"`
	}
	variables := map[string]interface{}{"id": aggReport, "fields": fields, "filter": filter}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to read report aggregations: %w", err)
	}
	return printJSON(response.Report.Aggregations)
}

func stringsIndex(value string, sep rune) int {
	for i, r := range value {
		if r == sep {
			return i
		}
	}
	return -1
}
