package tools

import (
	"flag"
	"fmt"
	"strconv"
	"strings"

	"demo-builder/common"
)

// Custom field filter structure
type CustomFieldFilter struct {
	Type           string      `json:"type"`
	CustomFieldID  string      `json:"customFieldId"`
	CustomFieldType string     `json:"customFieldType"`
	Op             string      `json:"op"`
	Values         interface{} `json:"values"`
}

// Fields filter structure
type FieldsFilter struct {
	Fields []CustomFieldFilter `json:"fields"`
	Op     string              `json:"op"`
}

// CustomFieldValueAlt represents a custom field value with simplified structure
type CustomFieldValueAlt struct {
	ID    string      `json:"id"`
	Value interface{} `json:"value"`
}

// CustomFieldInfo holds field metadata for display
type CustomFieldInfo struct {
	ID   string
	Name string
	Type string
}

// CustomFieldFilterParsed represents a parsed client-side custom field filter
type CustomFieldFilterParsed struct {
	FieldID   string
	Operator  string
	Value     interface{}
	ValueStr  string
}

// Enhanced record with custom field values
type EnhancedRecord struct {
	common.Record
	CustomFieldValues []CustomFieldValueAlt `json:"customFields"`
}

// Numerical stats for calculations
type NumericalStats struct {
	FieldName string  `json:"fieldName"`
	FieldID   string  `json:"fieldId"`
	Count     int     `json:"count"`
	Sum       float64 `json:"sum"`
	Average   float64 `json:"average"`
	Min       float64 `json:"min"`
	Max       float64 `json:"max"`
}

// RecordsResult represents the paginated response from the GraphQL query
type RecordsResult struct {
	Items    []EnhancedRecord          `json:"items"`
	PageInfo common.CursorPageInfo    `json:"pageInfo"`
}

// RecordsResponse represents the response from the GraphQL query
type RecordsResponse struct {
	TodoQueries struct {
		Todos RecordsResult `json:"todos"`
	} `json:"todoQueries"`
}

func RunReadRecords(args []string) error {
	// Create flag set for this command
	fs := flag.NewFlagSet("read-records", flag.ExitOnError)
	
	// Parse command line flags
	projectID := fs.String("project", "", "Project ID to filter records")
	todoListID := fs.String("list", "", "Todo List ID to filter records")
	assigneeID := fs.String("assignee", "", "Filter by assignee ID")
	tagIDs := fs.String("tags", "", "Filter by tag IDs (comma-separated)")
	done := fs.String("done", "", "Filter by completion status (true/false)")
	archived := fs.String("archived", "", "Filter by archived status (true/false)")
	orderBy := fs.String("order", "updatedAt_DESC", "Order by field (position_ASC, position_DESC, title_ASC, title_DESC, createdAt_ASC, createdAt_DESC, updatedAt_ASC, updatedAt_DESC, duedAt_ASC, duedAt_DESC)")
	limit := fs.Int("limit", 20, "Maximum number of records to return")
	skip := fs.Int("skip", 0, "Number of records to skip (for pagination)")
	simple := fs.Bool("simple", false, "Show only basic record information")
	
	// Custom field filtering flags
	customFieldFilter := fs.String("custom-field", "", "Filter by custom field: 'field_id:operator:value' (e.g., 'cf123:GT:50000' or 'cf456:CONTAINS:urgent')")
	showStats := fs.Bool("stats", false, "Show numerical statistics for custom fields (sum, average, min, max)")
	calcFields := fs.String("calc-fields", "", "Comma-separated list of custom field IDs to calculate stats for (optional - auto-detects numerical fields if not specified)")
	quickCalc := fs.Bool("calc", false, "Automatically calculate and display stats for all numerical fields found in results")
	
	if err := fs.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %v", err)
	}

	// Show help if needed
	if len(args) == 0 || (len(args) == 1 && (args[0] == "-h" || args[0] == "--help")) {
		fmt.Println("Read records with advanced filtering and statistics")
		fmt.Println()
		fmt.Println("Usage:")
		fmt.Println("  go run . read-records [flags]")
		fmt.Println()
		fmt.Println("Flags:")
		fs.PrintDefaults()
		fmt.Println()
		fmt.Println("Custom Field Filter Examples:")
		fmt.Println("  # Find records with amount over $50,000")
		fmt.Println("  go run . read-records -project PROJECT_ID -custom-field \"cf123:GT:50000\"")
		fmt.Println()
		fmt.Println("  # Find records containing 'urgent' in a text field")
		fmt.Println("  go run . read-records -project PROJECT_ID -custom-field \"cf456:CONTAINS:urgent\"")
		fmt.Println()
		fmt.Println("  # Show numerical statistics for custom fields")
		fmt.Println("  go run . read-records -project PROJECT_ID -stats")
		fmt.Println()
		fmt.Println("Operators: EQ, NE, GT, GTE, LT, LTE, IN, NIN, CONTAINS, IS, NOT")
		return nil
	}

	// Load configuration
	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %v", err)
	}

	// Create client
	client := common.NewClient(config)

	// Fetch custom field metadata if project ID is provided
	var customFieldInfo map[string]CustomFieldInfo
	if *projectID != "" {
		info, err := getCustomFieldInfo(client, *projectID)
		if err != nil {
			// Don't fail if we can't get field info, just log it
			fmt.Printf("Warning: Could not fetch custom field info: %v\n", err)
			customFieldInfo = make(map[string]CustomFieldInfo)
		} else {
			customFieldInfo = info
		}
	} else {
		customFieldInfo = make(map[string]CustomFieldInfo)
	}

	// Build the GraphQL query
	query := buildRecordsQuery(*simple)

	// Build filter variables - TodosFilter requires companyIds and uses different field names
	filter := make(map[string]interface{})
	
	// companyIds is required for TodosFilter - get from config or leave empty to match all companies
	filter["companyIds"] = []string{} // This will match all companies the user has access to
	
	if *projectID != "" {
		filter["projectIds"] = []string{*projectID}
	}
	if *todoListID != "" {
		filter["todoListIds"] = []string{*todoListID}
	}
	if *assigneeID != "" {
		filter["assigneeIds"] = []string{*assigneeID}
	}
	if *tagIDs != "" {
		// Parse comma-separated tag IDs
		tagList := strings.Split(*tagIDs, ",")
		for i, tag := range tagList {
			tagList[i] = strings.TrimSpace(tag)
		}
		filter["tagIds"] = tagList
	}
	
	// Note: Server-side custom field filtering is not working, so we'll do it client-side
	// Don't add the fields filter to avoid confusing the server
	var clientSideFilter *CustomFieldFilterParsed
	if *customFieldFilter != "" {
		parsed, err := parseClientSideCustomFieldFilter(*customFieldFilter)
		if err != nil {
			return fmt.Errorf("failed to parse custom field filter: %v", err)
		}
		clientSideFilter = parsed
	}
	if *done != "" {
		if *done == "true" {
			filter["done"] = true
		} else if *done == "false" {
			filter["done"] = false
		}
	}
	if *archived != "" {
		if *archived == "true" {
			filter["archived"] = true
		} else if *archived == "false" {
			filter["archived"] = false
		}
	}

	// Build sort array based on orderBy string
	var sort []string
	if *orderBy != "" {
		// Convert from TodoOrderByInput format to TodosSort format
		switch *orderBy {
		case "position_ASC":
			sort = append(sort, "position_ASC")
		case "position_DESC":
			sort = append(sort, "position_DESC")
		case "title_ASC":
			sort = append(sort, "title_ASC")
		case "title_DESC":
			sort = append(sort, "title_DESC")
		case "createdAt_ASC":
			sort = append(sort, "createdAt_ASC")
		case "createdAt_DESC":
			sort = append(sort, "createdAt_DESC")
		case "updatedAt_ASC":
			// No updatedAt in TodosSort, use createdAt instead
			sort = append(sort, "createdAt_ASC")
		case "updatedAt_DESC":
			// No updatedAt in TodosSort, use createdAt instead
			sort = append(sort, "createdAt_DESC")
		case "duedAt_ASC":
			sort = append(sort, "duedAt_ASC")
		case "duedAt_DESC":
			sort = append(sort, "duedAt_DESC")
		default:
			sort = append(sort, "createdAt_DESC")
		}
	}

	// Build variables
	variables := map[string]interface{}{
		"filter": filter,
		"limit":  *limit,
	}


	// Execute query
	var response RecordsResponse
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return fmt.Errorf("failed to execute query: %v", err)
	}
	
	// Apply client-side filtering if specified
	originalCount := len(response.TodoQueries.Todos.Items)
	if clientSideFilter != nil {
		response.TodoQueries.Todos.Items = applyClientSideFilter(response.TodoQueries.Todos.Items, clientSideFilter)
		// Update pagination info since we filtered client-side
		response.TodoQueries.Todos.PageInfo.HasNextPage = false // Can't determine this after client-side filtering
		response.TodoQueries.Todos.PageInfo.HasPreviousPage = *skip > 0
	}

	// Display results
	result := response.TodoQueries.Todos
	fmt.Printf("\n=== Records Query Results ===\n")
	if *projectID != "" {
		fmt.Printf("Project ID: %s\n", *projectID)
	}
	if *todoListID != "" {
		fmt.Printf("List ID: %s\n", *todoListID)
	}
	if *assigneeID != "" {
		fmt.Printf("Assignee ID: %s\n", *assigneeID)
	}
	if *tagIDs != "" {
		fmt.Printf("Tag IDs: %s\n", *tagIDs)
	}
	if *customFieldFilter != "" {
		fmt.Printf("Custom Field Filter: %s\n", *customFieldFilter)
		if clientSideFilter != nil {
			fmt.Printf("Filter Applied: %d → %d records (client-side)\n", originalCount, len(result.Items))
		}
	}
	fmt.Printf("Showing: %d records (skip: %d, limit: %d)\n", len(result.Items), *skip, *limit)
	fmt.Printf("Has next page: %t\n", result.PageInfo.HasNextPage)
	fmt.Printf("Has previous page: %t\n", result.PageInfo.HasPreviousPage)
	fmt.Println()
	
	// Calculate and display statistics if requested
	if (*showStats || *quickCalc) && len(result.Items) > 0 {
		var calcFieldsToUse string = *calcFields
		if *quickCalc && calcFieldsToUse == "" {
			// Auto-detect numerical fields from the results
			calcFieldsToUse = autoDetectNumericalFields(result.Items, customFieldInfo)
		}
		if calcFieldsToUse != "" {
			stats := calculateNumericalStats(result.Items, calcFieldsToUse, customFieldInfo)
			displayStats(stats)
		}
	}

	if len(result.Items) == 0 {
		fmt.Println("No records found matching the criteria.")
		return nil
	}

	// Display records
	for i, record := range result.Items {
		recordNum := *skip + i + 1
		if *simple {
			// Simple output
			fmt.Printf("%d. %s\n", recordNum, record.Title)
			fmt.Printf("   ID: %s\n", record.ID)
			if record.TodoList != nil {
				fmt.Printf("   List: %s\n", record.TodoList.Title)
			}
			fmt.Printf("   Status: %s\n", getRecordStatus(record.Record))
			if record.DuedAt != "" {
				fmt.Printf("   Due: %s\n", record.DuedAt)
			}
			
			// Show custom field values in simple mode too
			if len(record.CustomFieldValues) > 0 {
				fmt.Printf("   Custom Fields: ")
				nonEmptyFields := 0
				for _, cfv := range record.CustomFieldValues {
					actualValue := parseCustomFieldValue(cfv.Value)
					if actualValue == nil {
						continue
					}
					
					if nonEmptyFields > 0 {
						fmt.Printf(", ")
					}
					
					// Use field name with ID if available, otherwise just ID
					fieldDisplay := cfv.ID
					if info, exists := customFieldInfo[cfv.ID]; exists {
						fieldDisplay = fmt.Sprintf("%s (%s)", info.Name, cfv.ID)
					}
					
					fmt.Printf("%s=%v", fieldDisplay, actualValue)
					nonEmptyFields++
				}
				if nonEmptyFields == 0 {
					fmt.Printf("(none set)")
				}
				fmt.Println()
			}
			fmt.Println()
		} else {
			// Detailed output
			fmt.Printf("%d. %s\n", recordNum, record.Title)
			fmt.Printf("   ID: %s\n", record.ID)
			fmt.Printf("   UID: %s\n", record.UID)
			if record.TodoList != nil {
				fmt.Printf("   List: %s (%s)\n", record.TodoList.Title, record.TodoList.ID)
			}
			fmt.Printf("   Position: %.0f\n", record.Position)
			fmt.Printf("   Status: %s\n", getRecordStatus(record.Record))
			
			if record.Text != "" {
				fmt.Printf("   Description: %s\n", common.TruncateString(record.Text, 100))
			}
			if record.StartedAt != "" {
				fmt.Printf("   Started: %s\n", record.StartedAt)
			}
			if record.DuedAt != "" {
				fmt.Printf("   Due: %s\n", record.DuedAt)
			}
			if record.Color != "" {
				fmt.Printf("   Color: %s\n", record.Color)
			}
			if record.Cover != "" {
				fmt.Printf("   Has cover: Yes\n")
			}
			fmt.Printf("   Comments: %d\n", record.CommentCount)
			fmt.Printf("   Checklists: %d/%d completed\n", record.ChecklistCompletedCount, record.ChecklistCount)
			if record.IsRepeating {
				fmt.Printf("   Repeating: Yes\n")
			}
			
			// Display assignees
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

			// Display tags
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
			
			// Display custom fields in detailed mode
			if len(record.CustomFieldValues) > 0 {
				hasVisibleFields := false
				for _, cfv := range record.CustomFieldValues {
					actualValue := parseCustomFieldValue(cfv.Value)
					if actualValue != nil {
						if !hasVisibleFields {
							fmt.Printf("   Custom Fields:\n")
							hasVisibleFields = true
						}
						
						// Use field name with ID and type if available, otherwise just ID
						fieldDisplay := cfv.ID
						if info, exists := customFieldInfo[cfv.ID]; exists {
							fieldDisplay = fmt.Sprintf("%s (%s) [%s]", info.Name, info.Type, cfv.ID)
						}
						
						fmt.Printf("     %s: %v\n", fieldDisplay, actualValue)
					}
				}
			}

			fmt.Printf("   Created: %s\n", record.CreatedAt)
			fmt.Printf("   Updated: %s\n", record.UpdatedAt)
			fmt.Println()
		}
	}

	// Display pagination info
	if result.PageInfo.HasNextPage {
		nextSkip := *skip + *limit
		fmt.Printf("To see more records, use: -skip %d\n", nextSkip)
	}

	return nil
}

// buildRecordsQuery builds the GraphQL query based on the detail level  
func buildRecordsQuery(simple bool) string {
	// Try using a different GraphQL endpoint - let's try the todoQueries approach
	if simple {
		return `
			query GetRecords($filter: TodosFilter!, $limit: Int) {
				todoQueries {
					todos(filter: $filter, limit: $limit) {
						items {
							id
							uid
							position
							title
							duedAt
							done
							archived
							commentCount
							checklistCount
							checklistCompletedCount
							isRepeating
							createdAt
							updatedAt
							users {
								id
								uid
								firstName
								lastName
								fullName
								email
							}
							tags {
								id
								uid
								title
								color
							}
							todoList {
								id
								uid
								title
							}
							customFields {
								id
								value
							}
						}
						pageInfo {
							hasNextPage
							hasPreviousPage
							startCursor
							endCursor
						}
					}
				}
			}
		`
	}

	return `
		query GetRecords($filter: TodosFilter!, $limit: Int) {
			todoQueries {
				todos(filter: $filter, limit: $limit) {
					items {
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
						isRead
						isSeen
						createdAt
						updatedAt
						users {
							id
							uid
							firstName
							lastName
							fullName
							email
						}
						tags {
							id
							uid
							title
							color
						}
						todoList {
							id
							uid
							title
						}
						customFields {
							id
							value
						}
					}
					pageInfo {
						hasNextPage
						hasPreviousPage
						startCursor
						endCursor
					}
				}
			}
		}
	`
}

// parseCustomFieldFilter parses a custom field filter string
// Format: "field_id:operator:value" (e.g., "cf123:GT:50000" or "cf456:CONTAINS:urgent")
func parseCustomFieldFilter(filterStr string) (*FieldsFilter, error) {
	parts := strings.SplitN(filterStr, ":", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid filter format. Use 'field_id:operator:value'")
	}
	
	fieldID := strings.TrimSpace(parts[0])
	operator := strings.TrimSpace(strings.ToUpper(parts[1]))
	valueStr := strings.TrimSpace(parts[2])
	
	// Validate operator
	validOps := map[string]bool{
		"EQ": true, "NE": true, "GT": true, "GTE": true, "LT": true, "LTE": true,
		"IN": true, "NIN": true, "CONTAINS": true, "IS": true, "NOT": true,
	}
	if !validOps[operator] {
		return nil, fmt.Errorf("invalid operator '%s'. Valid operators: EQ, NE, GT, GTE, LT, LTE, IN, NIN, CONTAINS, IS, NOT", operator)
	}
	
	// Parse value based on operator
	var value interface{}
	switch operator {
	case "IN", "NIN":
		// Array values
		values := strings.Split(valueStr, ",")
		for i, v := range values {
			values[i] = strings.TrimSpace(v)
		}
		value = values
	case "GT", "GTE", "LT", "LTE":
		// Try to parse as number first
		if numVal, err := strconv.ParseFloat(valueStr, 64); err == nil {
			value = numVal
		} else {
			value = valueStr // Keep as string for date/time fields
		}
	default:
		// Try to parse as number, boolean, or keep as string
		if numVal, err := strconv.ParseFloat(valueStr, 64); err == nil {
			value = numVal
		} else if boolVal, err := strconv.ParseBool(valueStr); err == nil {
			value = boolVal
		} else {
			value = valueStr
		}
	}
	
	filter := &FieldsFilter{
		Fields: []CustomFieldFilter{{
			Type:            "CUSTOM_FIELD",
			CustomFieldID:   fieldID,
			CustomFieldType: "AUTO_DETECT", // Let the API auto-detect the field type
			Op:              operator,
			Values:          value,
		}},
		Op: "AND",
	}
	
	return filter, nil
}

// getCustomFieldInfo fetches custom field metadata from the project
func getCustomFieldInfo(client *common.Client, projectID string) (map[string]CustomFieldInfo, error) {
	query := `
		query GetProjectCustomFields {
			customFields {
				items {
					id
					name
					type
				}
			}
		}
	`
	
	// Set project context
	client.SetProject(projectID)
	
	type CustomFieldsResponse struct {
		CustomFields struct {
			Items []struct {
				ID   string `json:"id"`
				Name string `json:"name"`
				Type string `json:"type"`
			} `json:"items"`
		} `json:"customFields"`
	}
	
	var response CustomFieldsResponse
	if err := client.ExecuteQueryWithResult(query, nil, &response); err != nil {
		return nil, fmt.Errorf("failed to fetch custom field info: %v", err)
	}
	
	fieldInfo := make(map[string]CustomFieldInfo)
	for _, field := range response.CustomFields.Items {
		fieldInfo[field.ID] = CustomFieldInfo{
			ID:   field.ID,
			Name: field.Name,
			Type: field.Type,
		}
	}
	
	return fieldInfo, nil
}

// parseCustomFieldValue extracts the actual value from the complex value structure
func parseCustomFieldValue(value interface{}) interface{} {
	if value == nil {
		return nil
	}
	
	// Handle map structure like map[currency:<nil> number:75000 text:<nil>]
	if valueMap, ok := value.(map[string]interface{}); ok {
		// Check for different value types in priority order
		if num := valueMap["number"]; num != nil {
			return num
		}
		if curr := valueMap["currency"]; curr != nil {
			return curr
		}
		if txt := valueMap["text"]; txt != nil {
			return txt
		}
		if date := valueMap["date"]; date != nil {
			return date
		}
		if boolean := valueMap["boolean"]; boolean != nil {
			return boolean
		}
		// Return the whole map if we can't find a specific type
		return value
	}
	
	// Return as-is for simple values
	return value
}

// calculateNumericalStats calculates statistics for numerical custom fields
func calculateNumericalStats(records []EnhancedRecord, calcFieldsStr string, customFieldInfo map[string]CustomFieldInfo) []NumericalStats {
	var targetFields []string
	if calcFieldsStr != "" {
		targetFields = strings.Split(calcFieldsStr, ",")
		for i, field := range targetFields {
			targetFields[i] = strings.TrimSpace(field)
		}
	}
	
	// Collect field values by field ID
	fieldValues := make(map[string][]float64)
	fieldNames := make(map[string]string)
	
	for _, record := range records {
		for _, cfv := range record.CustomFieldValues {
			// Check if we should calculate stats for this field
			if len(targetFields) > 0 {
				found := false
				for _, targetField := range targetFields {
					if cfv.ID == targetField {
						found = true
						break
					}
				}
				if !found {
					continue
				}
			}
			
			// Check if field is numerical using field info
			fieldInfo, exists := customFieldInfo[cfv.ID]
			if !exists {
				continue // Skip if we don't have field info
			}
			
			isNumerical := false
			switch fieldInfo.Type {
			case "NUMBER", "CURRENCY", "PERCENT", "RATING":
				isNumerical = true
			}
			
			if !isNumerical {
				continue
			}
			
			// Parse numerical value using our value parser
			actualValue := parseCustomFieldValue(cfv.Value)
			if actualValue != nil {
				var numVal float64
				switch v := actualValue.(type) {
				case float64:
					numVal = v
				case float32:
					numVal = float64(v)
				case int:
					numVal = float64(v)
				case int64:
					numVal = float64(v)
				case string:
					if parsed, err := strconv.ParseFloat(v, 64); err == nil {
						numVal = parsed
					} else {
						continue // Skip non-numeric string values
					}
				default:
					continue // Skip unknown value types
				}
				
				fieldValues[cfv.ID] = append(fieldValues[cfv.ID], numVal)
				fieldNames[cfv.ID] = fieldInfo.Name
			}
		}
	}
	
	// Calculate statistics
	var stats []NumericalStats
	for fieldID, values := range fieldValues {
		if len(values) == 0 {
			continue
		}
		
		// Calculate sum, min, max
		sum := 0.0
		min := values[0]
		max := values[0]
		
		for _, val := range values {
			sum += val
			if val < min {
				min = val
			}
			if val > max {
				max = val
			}
		}
		
		average := sum / float64(len(values))
		
		stats = append(stats, NumericalStats{
			FieldName: fieldNames[fieldID],
			FieldID:   fieldID,
			Count:     len(values),
			Sum:       sum,
			Average:   average,
			Min:       min,
			Max:       max,
		})
	}
	
	return stats
}

// displayStats displays numerical statistics
func displayStats(stats []NumericalStats) {
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

// getRecordStatus returns a human-readable status for a record
func getRecordStatus(record common.Record) string {
	if record.Archived {
		return "Archived"
	}
	if record.Done {
		return "Completed"
	}
	return "Active"
}

// parseClientSideCustomFieldFilter parses a custom field filter string for client-side filtering
func parseClientSideCustomFieldFilter(filterStr string) (*CustomFieldFilterParsed, error) {
	parts := strings.SplitN(filterStr, ":", 3)
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid filter format. Use 'field_id:operator:value'")
	}
	
	fieldID := strings.TrimSpace(parts[0])
	operator := strings.ToUpper(strings.TrimSpace(parts[1]))
	valueStr := strings.TrimSpace(parts[2])
	
	// Try to parse the value as a number
	var value interface{}
	if numValue, err := strconv.ParseFloat(valueStr, 64); err == nil {
		value = numValue
	} else if boolValue, err := strconv.ParseBool(valueStr); err == nil {
		value = boolValue
	} else {
		value = valueStr
	}
	
	return &CustomFieldFilterParsed{
		FieldID:  fieldID,
		Operator: operator,
		Value:    value,
		ValueStr: valueStr,
	}, nil
}

// applyClientSideFilter filters records based on custom field values
func applyClientSideFilter(records []EnhancedRecord, filter *CustomFieldFilterParsed) []EnhancedRecord {
	if filter == nil {
		return records
	}
	
	var filtered []EnhancedRecord
	for _, record := range records {
		if matchesFilter(record, filter) {
			filtered = append(filtered, record)
		}
	}
	return filtered
}

// matchesFilter checks if a record matches the custom field filter
func matchesFilter(record EnhancedRecord, filter *CustomFieldFilterParsed) bool {
	// Find the custom field value
	var fieldValue interface{}
	found := false
	
	for _, cfv := range record.CustomFieldValues {
		if cfv.ID == filter.FieldID {
			fieldValue = parseCustomFieldValue(cfv.Value)
			found = true
			break
		}
	}
	
	// If field not found or value is nil, handle based on operator
	if !found || fieldValue == nil {
		switch filter.Operator {
		case "IS":
			return filter.ValueStr == "null" || filter.ValueStr == "empty"
		case "NOT":
			return filter.ValueStr != "null" && filter.ValueStr != "empty"
		default:
			return false
		}
	}
	
	// Apply the filter based on operator
	return compareValues(fieldValue, filter.Operator, filter.Value)
}

// compareValues compares two values based on the given operator
func compareValues(fieldValue interface{}, operator string, filterValue interface{}) bool {
	switch operator {
	case "EQ":
		return fmt.Sprintf("%v", fieldValue) == fmt.Sprintf("%v", filterValue)
	case "NE":
		return fmt.Sprintf("%v", fieldValue) != fmt.Sprintf("%v", filterValue)
	case "CONTAINS":
		fieldStr := fmt.Sprintf("%v", fieldValue)
		filterStr := fmt.Sprintf("%v", filterValue)
		return strings.Contains(strings.ToLower(fieldStr), strings.ToLower(filterStr))
	case "GT", "GTE", "LT", "LTE":
		return compareNumerical(fieldValue, operator, filterValue)
	case "IN":
		filterStr := fmt.Sprintf("%v", filterValue)
		values := strings.Split(filterStr, ",")
		fieldStr := fmt.Sprintf("%v", fieldValue)
		for _, val := range values {
			if strings.TrimSpace(val) == fieldStr {
				return true
			}
		}
		return false
	case "NIN":
		filterStr := fmt.Sprintf("%v", filterValue)
		values := strings.Split(filterStr, ",")
		fieldStr := fmt.Sprintf("%v", fieldValue)
		for _, val := range values {
			if strings.TrimSpace(val) == fieldStr {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// compareNumerical compares numerical values
func compareNumerical(fieldValue interface{}, operator string, filterValue interface{}) bool {
	fieldNum, fieldOk := toFloat64(fieldValue)
	filterNum, filterOk := toFloat64(filterValue)
	
	if !fieldOk || !filterOk {
		return false
	}
	
	switch operator {
	case "GT":
		return fieldNum > filterNum
	case "GTE":
		return fieldNum >= filterNum
	case "LT":
		return fieldNum < filterNum
	case "LTE":
		return fieldNum <= filterNum
	default:
		return false
	}
}

// toFloat64 converts a value to float64 if possible
func toFloat64(value interface{}) (float64, bool) {
	switch v := value.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	case int32:
		return float64(v), true
	case string:
		if num, err := strconv.ParseFloat(v, 64); err == nil {
			return num, true
		}
	}
	return 0, false
}

// autoDetectNumericalFields automatically detects numerical fields from the record results
func autoDetectNumericalFields(records []EnhancedRecord, customFieldInfo map[string]CustomFieldInfo) string {
	fieldCounts := make(map[string]int)
	numericalFieldTypes := map[string]bool{
		"NUMBER":   true,
		"CURRENCY": true,
		"PERCENT":  true,
		"RATING":   true,
	}
	
	// Find fields that appear in records and are numerical
	for _, record := range records {
		for _, cfv := range record.CustomFieldValues {
			if info, exists := customFieldInfo[cfv.ID]; exists {
				if numericalFieldTypes[info.Type] {
					// Check if the field has a numerical value
					parsedValue := parseCustomFieldValue(cfv.Value)
					if _, ok := toFloat64(parsedValue); ok {
						fieldCounts[cfv.ID]++
					}
				}
			}
		}
	}
	
	// Convert to comma-separated string of fields that have at least one numerical value
	var detectedFields []string
	for fieldID, count := range fieldCounts {
		if count > 0 {
			detectedFields = append(detectedFields, fieldID)
		}
	}
	
	return strings.Join(detectedFields, ",")
}