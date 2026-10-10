package databases

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

type recordListOptions struct {
	database  string
	workspace string
	limit     int
	after     string
	archived  bool
	format    string
}

type recordListRow struct {
	Record struct {
		ID        string `json:"id"`
		CreatedAt string `json:"createdAt"`
		UpdatedAt string `json:"updatedAt"`
	} `json:"record"`
	Membership struct {
		Archived bool `json:"archived"`
	} `json:"membership"`
	Fields []recordListField `json:"fields"`
}

type recordListField struct {
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Type  string          `json:"type"`
	Value json.RawMessage `json:"value"`
}

type recordListResponse struct {
	Rows       []recordListRow `json:"rows"`
	NextCursor string          `json:"nextCursor"`
	HasMore    bool            `json:"hasMore"`
}

func newRecordsListCmd() *cobra.Command {
	opts := recordListOptions{}
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List records on a database",
		Long: `List records on a database with their field values.

Values are rendered per field type; use --format csv for a spreadsheet or
--format json for the raw API response. Pagination is cursor based: pass
--after with the cursor from the previous page.`,
		Example: `  blue databases records list --database <id> --workspace <id>
  blue databases records list --database <id> --workspace <id> --limit 100
  blue databases records list --database <id> --workspace <id> --after <cursor>
  blue databases records list --database <id> --workspace <id> --archived --format csv`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRecordsList(cmd, opts)
		},
	}

	cmd.Flags().StringVar(&opts.database, "database", "", "Database ID (required)")
	cmd.Flags().StringVarP(&opts.workspace, "workspace", "w", "", "Workspace ID or slug (required)")
	cmd.Flags().IntVar(&opts.limit, "limit", 50, "Maximum records to return (1-200; server maximum is 200)")
	cmd.Flags().StringVar(&opts.after, "after", "", "Cursor from a previous page")
	cmd.Flags().BoolVar(&opts.archived, "archived", false, "Include archived records")
	cmd.Flags().StringVar(&opts.format, "format", "text", "Output format: text, json, csv")

	return cmd
}

func runRecordsList(cmd *cobra.Command, opts recordListOptions) error {
	if err := requireFlag(opts.database, "database"); err != nil {
		return err
	}
	if err := requireFlag(opts.workspace, "workspace"); err != nil {
		return err
	}
	if opts.limit < 1 || opts.limit > 200 {
		return fmt.Errorf("limit must be between 1 and 200 (server maximum is 200), got %d", opts.limit)
	}
	switch opts.format {
	case "text", "json", "csv":
	default:
		return fmt.Errorf("invalid format %q. Use text, json, or csv", opts.format)
	}

	client, workspaceID, err := databasesClient(opts.workspace)
	if err != nil {
		return err
	}

	query := queryValues("workspaceId", workspaceID, "first", strconv.Itoa(opts.limit))
	if opts.after != "" {
		query.Set("cursor", opts.after)
	}
	if opts.archived {
		query.Set("includeArchived", "true")
	}

	var raw json.RawMessage
	if err := client.REST("GET", "/databases/"+opts.database+"/records", query, nil, &raw); err != nil {
		return fmt.Errorf("failed to list database records: %w", err)
	}

	switch opts.format {
	case "json":
		return printJSON(raw)
	case "csv":
		return printRecordsCSV(cmd, raw)
	default:
		return printRecordsText(cmd, raw, opts)
	}
}

func printRecordsText(cmd *cobra.Command, raw json.RawMessage, opts recordListOptions) error {
	out := cmd.OutOrStdout()
	parsed, err := decodeRecordsResponse(raw)
	if err != nil {
		return err
	}

	fmt.Fprintln(out, "=== Records ===")
	if len(parsed.Rows) == 0 {
		fmt.Fprintln(out, "No records.")
		return nil
	}

	for i, row := range parsed.Rows {
		fmt.Fprintf(out, "%d. %s\n", i+1, row.Record.ID)
		fmt.Fprintf(out, "   Created: %s\n", row.Record.CreatedAt)
		fmt.Fprintf(out, "   Updated: %s\n", row.Record.UpdatedAt)
		for _, field := range row.Fields {
			if recordIsNullJSON(field.Value) {
				continue
			}
			fmt.Fprintf(out, "   %s: %s\n", field.Name, recordRenderFieldValue(field.Type, field.Value))
		}
	}

	if parsed.HasMore {
		fmt.Fprintln(out)
		fmt.Fprintf(out, "Next page: blue databases records list --database %s --after %s --workspace %s\n", opts.database, parsed.NextCursor, opts.workspace)
	}
	return nil
}

func printRecordsCSV(cmd *cobra.Command, raw json.RawMessage) error {
	parsed, err := decodeRecordsResponse(raw)
	if err != nil {
		return err
	}

	fieldNames := make([]string, 0)
	fieldIndex := make(map[string]int)
	for _, row := range parsed.Rows {
		for _, field := range row.Fields {
			if _, ok := fieldIndex[field.Name]; !ok {
				fieldIndex[field.Name] = len(fieldNames)
				fieldNames = append(fieldNames, field.Name)
			}
		}
	}

	header := []string{"recordId", "createdAt", "updatedAt", "archived"}
	header = append(header, fieldNames...)

	writer := csv.NewWriter(cmd.OutOrStdout())
	if err := writer.Write(header); err != nil {
		return err
	}
	for _, row := range parsed.Rows {
		cells := []string{
			row.Record.ID,
			row.Record.CreatedAt,
			row.Record.UpdatedAt,
			strconv.FormatBool(row.Membership.Archived),
		}
		for range fieldNames {
			cells = append(cells, "")
		}
		for _, field := range row.Fields {
			if recordIsNullJSON(field.Value) {
				continue
			}
			cells[4+fieldIndex[field.Name]] = recordCompactJSON(field.Value)
		}
		if err := writer.Write(cells); err != nil {
			return err
		}
	}
	writer.Flush()
	return writer.Error()
}

// decodeRecordsResponse decodes the raw list response. Field values stay as
// json.RawMessage so number formatting survives for JSON and CSV passthrough.
func decodeRecordsResponse(raw json.RawMessage) (recordListResponse, error) {
	var parsed recordListResponse
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := dec.Decode(&parsed); err != nil {
		return parsed, fmt.Errorf("failed to parse records response: %w", err)
	}
	return parsed, nil
}

func recordIsNullJSON(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	return len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null"))
}

func recordCompactJSON(raw json.RawMessage) string {
	out, err := json.Marshal(raw)
	if err != nil {
		return string(raw)
	}
	return string(out)
}

// recordRenderFieldValue renders one field value for text output per its
// type. Anything without a known shape falls back to compact JSON.
func recordRenderFieldValue(fieldType string, raw json.RawMessage) string {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var value interface{}
	if err := dec.Decode(&value); err != nil {
		return recordCompactJSON(raw)
	}
	m, ok := value.(map[string]interface{})
	if !ok {
		return recordCompactJSON(raw)
	}

	switch fieldType {
	case "TEXT_SINGLE", "EMAIL", "URL":
		return recordAnyString(m["text"])
	case "NUMBER", "PERCENT", "RATING":
		return recordAnyString(m["number"])
	case "CURRENCY":
		number := recordAnyString(m["number"])
		currency := recordAnyString(m["currency"])
		if number == "" {
			return ""
		}
		if currency == "" {
			return number
		}
		return number + " " + currency
	case "PHONE":
		text := recordAnyString(m["text"])
		region := recordAnyString(m["regionCode"])
		if region == "" {
			return text
		}
		return text + " (" + region + ")"
	case "CHECKBOX":
		if checked, ok := m["checked"].(bool); ok {
			return strconv.FormatBool(checked)
		}
		return recordAnyString(m["checked"])
	case "COUNTRY":
		return recordJoinTitles(m["countryCodes"], "")
	case "LOCATION":
		if text := recordAnyString(m["text"]); text != "" {
			return text
		}
		return strings.Trim(recordAnyString(m["latitude"])+", "+recordAnyString(m["longitude"]), ", ")
	case "DATE":
		start := recordAnyString(m["startDate"])
		end := recordAnyString(m["endDate"])
		if end == "" {
			return start
		}
		return start + " → " + end
	case "SELECT_SINGLE", "SELECT_MULTI":
		return recordJoinTitles(m["options"], "title")
	case "ASSIGNEE":
		return recordJoinTitles(m["users"], "name")
	case "FILE":
		files := recordJoinTitles(m["files"], "title")
		folders := recordJoinTitles(m["folders"], "title")
		if files == "" {
			return folders
		}
		if folders == "" {
			return files
		}
		return files + ", " + folders
	case "REFERENCE":
		return recordJoinTitles(m["records"], "label")
	default:
		return recordCompactJSON(raw)
	}
}

func recordAnyString(value interface{}) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case json.Number:
		return v.String()
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// recordJoinTitles joins a JSON array of objects by one key. A plain array of
// strings (country codes) passes key "".
func recordJoinTitles(value interface{}, key string) string {
	items, ok := value.([]interface{})
	if !ok {
		return ""
	}
	parts := make([]string, 0, len(items))
	for _, item := range items {
		if key == "" {
			if s := recordAnyString(item); s != "" {
				parts = append(parts, s)
			}
			continue
		}
		m, ok := item.(map[string]interface{})
		if !ok {
			continue
		}
		if title := recordAnyString(m[key]); title != "" {
			parts = append(parts, title)
		}
	}
	return strings.Join(parts, ", ")
}
