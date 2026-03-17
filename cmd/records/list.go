package records

import (
	"fmt"
	"strconv"
	"strings"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

// EnhancedRecord with custom field values
type EnhancedRecord struct {
	common.Record
	CustomFieldValues []SimpleFieldValue `json:"customFields"`
}

type RecordsResponse struct {
	TodoQueries struct {
		Todos struct {
			Items    []EnhancedRecord     `json:"items"`
			PageInfo common.CursorPageInfo `json:"pageInfo"`
		} `json:"todos"`
	} `json:"todoQueries"`
}

type NumericalStats struct {
	FieldName string
	FieldID   string
	Count     int
	Sum       float64
	Average   float64
	Min       float64
	Max       float64
}

// Client-side filter
type ClientSideFilter struct {
	FieldID  string
	Operator string
	Value    interface{}
	ValueStr string
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List records with filtering",
	Long:  "Query records across workspaces with advanced filtering, sorting, and statistics.",
	Example: `  blue records list --workspace <id>
  blue records list --workspace <id> --done false --simple
  blue records list --workspace <id> --custom-field "cf123:GT:50000" --stats
  blue records list --list <id>`,
	RunE: runList,
}

var (
	listWorkspace   string
	listListID      string
	listAssignee    string
	listTags        string
	listDone        string
	listArchived    string
	listOrder       string
	listLimit       int
	listSkip        int
	listSimple      bool
	listCustomField string
	listStats       bool
	listCalc        bool
	listCalcFields  string
)

func init() {
	listCmd.Flags().StringVarP(&listWorkspace, "workspace", "w", "", "Workspace ID to filter records")
	listCmd.Flags().StringVarP(&listListID, "list", "l", "", "List ID to filter records")
	listCmd.Flags().StringVar(&listAssignee, "assignee", "", "Filter by assignee ID")
	listCmd.Flags().StringVar(&listTags, "tags", "", "Filter by tag IDs (comma-separated)")
	listCmd.Flags().StringVar(&listDone, "done", "", "Filter by completion status (true/false)")
	listCmd.Flags().StringVar(&listArchived, "archived", "", "Filter by archived status (true/false)")
	listCmd.Flags().StringVar(&listOrder, "order", "updatedAt_DESC", "Order by field")
	listCmd.Flags().IntVar(&listLimit, "limit", 20, "Maximum number of records to return")
	listCmd.Flags().IntVar(&listSkip, "skip", 0, "Number of records to skip")
	listCmd.Flags().BoolVarP(&listSimple, "simple", "s", false, "Show only basic record information")
	listCmd.Flags().StringVar(&listCustomField, "custom-field", "", "Filter by custom field: 'field_id:operator:value'")
	listCmd.Flags().BoolVar(&listStats, "stats", false, "Show numerical statistics")
	listCmd.Flags().BoolVar(&listCalc, "calc", false, "Auto-calculate stats for numerical fields")
	listCmd.Flags().StringVar(&listCalcFields, "calc-fields", "", "Custom field IDs to calculate stats for")
}

func runList(cmd *cobra.Command, args []string) error {
	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := common.NewClient(config)

	// Fetch custom field metadata if workspace is provided
	var fieldInfoMap map[string]FieldInfo
	if listWorkspace != "" {
		info, err := fetchFieldInfo(client)
		if err != nil {
			fieldInfoMap = make(map[string]FieldInfo)
		} else {
			fieldInfoMap = info
		}
		client.SetProject(listWorkspace)
	} else {
		fieldInfoMap = make(map[string]FieldInfo)
	}

	query := buildRecordsQuery(listSimple)

	filter := map[string]interface{}{
		"companyIds": []string{},
	}

	if listWorkspace != "" {
		filter["projectIds"] = []string{listWorkspace}
	}
	if listListID != "" {
		filter["todoListIds"] = []string{listListID}
	}
	if listAssignee != "" {
		filter["assigneeIds"] = []string{listAssignee}
	}
	if listTags != "" {
		tagList := strings.Split(listTags, ",")
		for i, tag := range tagList {
			tagList[i] = strings.TrimSpace(tag)
		}
		filter["tagIds"] = tagList
	}

	// Parse client-side filter
	var clientFilter *ClientSideFilter
	if listCustomField != "" {
		parsed, err := parseClientSideFilter(listCustomField)
		if err != nil {
			return fmt.Errorf("failed to parse custom field filter: %w", err)
		}
		clientFilter = parsed
	}

	if listDone == "true" {
		filter["done"] = true
	} else if listDone == "false" {
		filter["done"] = false
	}
	if listArchived == "true" {
		filter["archived"] = true
	} else if listArchived == "false" {
		filter["archived"] = false
	}

	variables := map[string]interface{}{
		"filter": filter,
		"limit":  listLimit,
	}

	var response RecordsResponse
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to list records: %w", err)
	}

	// Apply client-side filtering
	result := response.TodoQueries.Todos
	originalCount := len(result.Items)
	if clientFilter != nil {
		result.Items = applyClientSideFilter(result.Items, clientFilter)
	}

	fmt.Printf("\n=== Records Query Results ===\n")
	if listWorkspace != "" {
		fmt.Printf("Workspace ID: %s\n", listWorkspace)
	}
	if listListID != "" {
		fmt.Printf("List ID: %s\n", listListID)
	}
	if listCustomField != "" {
		fmt.Printf("Custom Field Filter: %s\n", listCustomField)
		if clientFilter != nil {
			fmt.Printf("Filter Applied: %d -> %d records (client-side)\n", originalCount, len(result.Items))
		}
	}
	fmt.Printf("Showing: %d records (skip: %d, limit: %d)\n", len(result.Items), listSkip, listLimit)
	fmt.Println()

	// Calculate stats if requested
	if (listStats || listCalc) && len(result.Items) > 0 {
		calcFieldsToUse := listCalcFields
		if listCalc && calcFieldsToUse == "" {
			calcFieldsToUse = autoDetectNumericalFields(result.Items, fieldInfoMap)
		}
		if calcFieldsToUse != "" {
			stats := calculateStats(result.Items, calcFieldsToUse, fieldInfoMap)
			displayRecordStats(stats)
		}
	}

	if len(result.Items) == 0 {
		fmt.Println("No records found matching the criteria.")
		return nil
	}

	for i, record := range result.Items {
		recordNum := listSkip + i + 1
		if listSimple {
			fmt.Printf("%d. %s\n", recordNum, record.Title)
			fmt.Printf("   ID: %s\n", record.ID)
			if record.TodoList != nil {
				fmt.Printf("   List: %s\n", record.TodoList.Title)
			}
			status := "Active"
			if record.Archived {
				status = "Archived"
			} else if record.Done {
				status = "Completed"
			}
			fmt.Printf("   Status: %s\n", status)
			if record.DuedAt != "" {
				fmt.Printf("   Due: %s\n", record.DuedAt)
			}
			fmt.Println()
		} else {
			fmt.Printf("%d. %s\n", recordNum, record.Title)
			fmt.Printf("   ID: %s\n", record.ID)
			if record.TodoList != nil {
				fmt.Printf("   List: %s (%s)\n", record.TodoList.Title, record.TodoList.ID)
			}
			status := "Active"
			if record.Archived {
				status = "Archived"
			} else if record.Done {
				status = "Completed"
			}
			fmt.Printf("   Status: %s\n", status)
			if record.Text != "" {
				fmt.Printf("   Description: %s\n", common.TruncateString(record.Text, 100))
			}
			if record.DuedAt != "" {
				fmt.Printf("   Due: %s\n", record.DuedAt)
			}
			if len(record.Users) > 0 {
				fmt.Printf("   Assignees: ")
				for j, user := range record.Users {
					if j > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s", user.FullName)
				}
				fmt.Println()
			}
			if len(record.Tags) > 0 {
				fmt.Printf("   Tags: ")
				for j, tag := range record.Tags {
					if j > 0 {
						fmt.Printf(", ")
					}
					fmt.Printf("%s", tag.Title)
				}
				fmt.Println()
			}
			if len(record.CustomFieldValues) > 0 {
				hasVisible := false
				for _, cfv := range record.CustomFieldValues {
					val := parseFieldValue(cfv.Value)
					if val != nil {
						if !hasVisible {
							fmt.Printf("   Custom Fields:\n")
							hasVisible = true
						}
						fieldDisplay := cfv.ID
						if info, exists := fieldInfoMap[cfv.ID]; exists {
							fieldDisplay = fmt.Sprintf("%s (%s)", info.Name, info.Type)
						}
						fmt.Printf("     %s: %v\n", fieldDisplay, val)
					}
				}
			}
			fmt.Printf("   Created: %s\n", record.CreatedAt)
			fmt.Printf("   Updated: %s\n", record.UpdatedAt)
			fmt.Println()
		}
	}

	if result.PageInfo.HasNextPage {
		fmt.Printf("To see more records, use: --skip %d\n", listSkip+listLimit)
	}

	return nil
}

func buildRecordsQuery(simple bool) string {
	fields := `id uid position title text startedAt duedAt timezone color cover archived done
		commentCount checklistCount checklistCompletedCount isRepeating createdAt updatedAt
		users { id uid firstName lastName fullName email }
		tags { id uid title color }
		todoList { id uid title }
		customFields { id value }`
	if simple {
		fields = `id uid position title duedAt done archived createdAt updatedAt
			users { id fullName }
			tags { id title color }
			todoList { id title }
			customFields { id value }`
	}

	return fmt.Sprintf(`
		query GetRecords($filter: TodosFilter!, $limit: Int) {
			todoQueries {
				todos(filter: $filter, limit: $limit) {
					items { %s }
					pageInfo {
						hasNextPage
						hasPreviousPage
						startCursor
						endCursor
					}
				}
			}
		}
	`, fields)
}

func parseClientSideFilter(filterStr string) (*ClientSideFilter, error) {
	parts := strings.SplitN(filterStr, ":", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid filter format. Use 'field_id:operator:value'")
	}

	fieldID := strings.TrimSpace(parts[0])
	operator := strings.ToUpper(strings.TrimSpace(parts[1]))
	valueStr := strings.TrimSpace(parts[2])

	var value interface{}
	if numValue, err := strconv.ParseFloat(valueStr, 64); err == nil {
		value = numValue
	} else if boolValue, err := strconv.ParseBool(valueStr); err == nil {
		value = boolValue
	} else {
		value = valueStr
	}

	return &ClientSideFilter{
		FieldID:  fieldID,
		Operator: operator,
		Value:    value,
		ValueStr: valueStr,
	}, nil
}

func applyClientSideFilter(records []EnhancedRecord, filter *ClientSideFilter) []EnhancedRecord {
	var filtered []EnhancedRecord
	for _, record := range records {
		if matchesFilter(record, filter) {
			filtered = append(filtered, record)
		}
	}
	return filtered
}

func matchesFilter(record EnhancedRecord, filter *ClientSideFilter) bool {
	var fieldValue interface{}
	found := false

	for _, cfv := range record.CustomFieldValues {
		if cfv.ID == filter.FieldID {
			fieldValue = parseFieldValue(cfv.Value)
			found = true
			break
		}
	}

	if !found || fieldValue == nil {
		return filter.Operator == "IS" && (filter.ValueStr == "null" || filter.ValueStr == "empty")
	}

	return compareValues(fieldValue, filter.Operator, filter.Value)
}

func compareValues(fieldValue interface{}, operator string, filterValue interface{}) bool {
	switch operator {
	case "EQ":
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", filterValue)
	case "NE":
		return fmt.Sprintf("%v", fieldValue) != fmt.Sprintf("%v", filterValue)
	case "CONTAINS":
		return strings.Contains(strings.ToLower(fmt.Sprintf("%v", fieldValue)), strings.ToLower(fmt.Sprintf("%v", filterValue)))
	case "GT", "GTE", "LT", "LTE":
		fv, fvOk := toFloat64(fieldValue)
		ffv, ffvOk := toFloat64(filterValue)
		if !fvOk || !ffvOk {
			return false
		}
		switch operator {
		case "GT":
			return fv > ffv
		case "GTE":
			return fv >= ffv
		case "LT":
			return fv < ffv
		case "LTE":
			return fv <= ffv
		}
	}
	return false
}

func toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case string:
		if num, err := strconv.ParseFloat(v, 64); err == nil {
			return num, true
		}
	}
	return 0, false
}

func autoDetectNumericalFields(records []EnhancedRecord, fieldInfoMap map[string]FieldInfo) string {
	numericalTypes := map[string]bool{"NUMBER": true, "CURRENCY": true, "PERCENT": true, "RATING": true}
	fieldCounts := make(map[string]int)

	for _, record := range records {
		for _, cfv := range record.CustomFieldValues {
			if info, exists := fieldInfoMap[cfv.ID]; exists {
				if numericalTypes[info.Type] {
					if v := parseFieldValue(cfv.Value); v != nil {
						if _, ok := toFloat64(v); ok {
							fieldCounts[cfv.ID]++
						}
					}
				}
			}
		}
	}

	var detected []string
	for fieldID, count := range fieldCounts {
		if count > 0 {
			detected = append(detected, fieldID)
		}
	}
	return strings.Join(detected, ",")
}

func calculateStats(records []EnhancedRecord, calcFieldsStr string, fieldInfoMap map[string]FieldInfo) []NumericalStats {
	targetFields := strings.Split(calcFieldsStr, ",")
	for i, f := range targetFields {
		targetFields[i] = strings.TrimSpace(f)
	}

	fieldValues := make(map[string][]float64)
	fieldNames := make(map[string]string)

	for _, record := range records {
		for _, cfv := range record.CustomFieldValues {
			found := false
			for _, tf := range targetFields {
				if cfv.ID == tf {
					found = true
					break
				}
			}
			if !found {
				continue
			}
			if info, exists := fieldInfoMap[cfv.ID]; exists {
				fieldNames[cfv.ID] = info.Name
			}
			val := parseFieldValue(cfv.Value)
			if numVal, ok := toFloat64(val); ok {
				fieldValues[cfv.ID] = append(fieldValues[cfv.ID], numVal)
			}
		}
	}

	var stats []NumericalStats
	for fieldID, values := range fieldValues {
		if len(values) == 0 {
			continue
		}
		sum, min, max := 0.0, values[0], values[0]
		for _, v := range values {
			sum += v
			if v < min {
				min = v
			}
			if v > max {
				max = v
			}
		}
		stats = append(stats, NumericalStats{
			FieldName: fieldNames[fieldID],
			FieldID:   fieldID,
			Count:     len(values),
			Sum:       sum,
			Average:   sum / float64(len(values)),
			Min:       min,
			Max:       max,
		})
	}
	return stats
}

func displayRecordStats(stats []NumericalStats) {
	if len(stats) == 0 {
		fmt.Println("No numerical custom fields found for statistics calculation.")
		return
	}
	fmt.Printf("\n=== Numerical Statistics ===\n")
	for _, stat := range stats {
		fmt.Printf("\nField: %s (%s)\n", stat.FieldName, stat.FieldID)
		fmt.Printf("  Records with values: %d\n", stat.Count)
		fmt.Printf("  Sum: %.2f\n", stat.Sum)
		fmt.Printf("  Average: %.2f\n", stat.Average)
		fmt.Printf("  Min: %.2f\n", stat.Min)
		fmt.Printf("  Max: %.2f\n", stat.Max)
	}
	fmt.Println()
}
