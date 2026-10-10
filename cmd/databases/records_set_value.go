package databases

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

type recordSetValueOptions struct {
	database         string
	record           string
	field            string
	kind             string
	value            string
	clientMutationID string
	entitySequence   int64
	workspace        string
	format           string
}

type recordWriteRejection struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type recordWriteResponse struct {
	Outcome         string                `json:"outcome"`
	TargetCursor    string                `json:"targetCursor"`
	EntityRevision  json.Number           `json:"entityRevision"`
	EntityUpdatedAt string                `json:"entityUpdatedAt"`
	Rejection       *recordWriteRejection `json:"rejection"`
}

func newRecordsSetValueCmd() *cobra.Command {
	opts := recordSetValueOptions{}
	cmd := &cobra.Command{
		Use:   "set-value",
		Short: "Set one field value on a database record",
		Long: `Set one field value on a database record.

Writes are idempotent: the server answers ACCEPTED, REPLAYED, REJECTED, or
CONFLICT based on the clientMutationId. Pass --client-mutation-id to make
retries of the same write collapse into one REPLAYED answer.

A --value that parses as a JSON object or array is sent as-is. Any other
--value is wrapped per --kind:

  text          {"text": "the raw string"}
  checkbox      strict true or false
  number        {"number": 1.5} (add "currency": "USD" for CURRENCY fields)
  select        comma-separated option IDs
  assignee      comma-separated user IDs
  reference     comma-separated record IDs
  file          comma-separated file IDs
  country       comma-separated country codes
  select-delta  JSON only: {"add":["opt_1"],"remove":["opt_2"]}
  date          JSON only: {"start":"2026-09-13","end":null,"timezone":"UTC","granularity":"ALL_DAY"}
  contact       JSON only: {"text":"+1 555 0100","regionCode":"US"}
  location      JSON only: {"text":"1 Main St","latitude":38.9,"longitude":-77.03}`,
		Example: `  blue databases records set-value --database <id> --record <id> --field <id> --kind text --value "Hello" --workspace <id>
  blue databases records set-value --database <id> --record <id> --field <id> --kind select --value "opt_1,opt_2" --workspace <id>
  blue databases records set-value --database <id> --record <id> --field <id> --kind number --value 42 --workspace <id>
  blue databases records set-value --database <id> --record <id> --field <id> --kind date --value '{"start":"2026-09-13","end":null,"timezone":"UTC","granularity":"ALL_DAY"}' --workspace <id>`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRecordsSetValue(cmd, opts)
		},
	}

	cmd.Flags().StringVar(&opts.database, "database", "", "Database ID (required)")
	cmd.Flags().StringVar(&opts.record, "record", "", "Record ID (required)")
	cmd.Flags().StringVar(&opts.field, "field", "", "Field ID (required)")
	cmd.Flags().StringVar(&opts.kind, "kind", "", "Value kind: text, number, date, checkbox, select, select-delta, assignee, reference, file, country, contact, location (required)")
	cmd.Flags().StringVar(&opts.value, "value", "", "Raw value; JSON object/array is sent as-is, otherwise wrapped per --kind (required)")
	cmd.Flags().StringVar(&opts.clientMutationID, "client-mutation-id", "", "Idempotency key (default: generated cuid)")
	cmd.Flags().Int64Var(&opts.entitySequence, "entity-sequence", 0, "Entity sequence number (default: current Unix millis)")
	cmd.Flags().StringVarP(&opts.workspace, "workspace", "w", "", "Workspace ID or slug (required)")
	cmd.Flags().StringVar(&opts.format, "format", "text", "Output format: text or json")

	return cmd
}

func runRecordsSetValue(cmd *cobra.Command, opts recordSetValueOptions) error {
	if err := requireFlag(opts.database, "database"); err != nil {
		return err
	}
	if err := requireFlag(opts.record, "record"); err != nil {
		return err
	}
	if err := requireFlag(opts.field, "field"); err != nil {
		return err
	}
	if err := requireFlag(opts.kind, "kind"); err != nil {
		return err
	}
	if err := requireFlag(opts.value, "value"); err != nil {
		return err
	}
	if err := requireFlag(opts.workspace, "workspace"); err != nil {
		return err
	}
	switch opts.format {
	case "text", "json":
	default:
		return fmt.Errorf("invalid format %q. Use text or json", opts.format)
	}

	value, err := recordParseWriteValue(opts.kind, opts.value)
	if err != nil {
		return err
	}

	client, workspaceID, err := databasesClient(opts.workspace)
	if err != nil {
		return err
	}

	clientMutationID := opts.clientMutationID
	if clientMutationID == "" {
		clientMutationID = common.NewCuid()
	}
	entitySequence := opts.entitySequence
	if entitySequence == 0 {
		entitySequence = common.NowMillis()
	}

	body := map[string]interface{}{
		"workspaceId":      workspaceID,
		"kind":             opts.kind,
		"value":            value,
		"clientMutationId": clientMutationID,
		"entitySequence":   entitySequence,
	}

	path := fmt.Sprintf("/databases/%s/records/%s/fields/%s", opts.database, opts.record, opts.field)
	var raw json.RawMessage
	if err := client.REST("PATCH", path, nil, body, &raw); err != nil {
		return fmt.Errorf("failed to set database record value: %w", err)
	}

	if opts.format == "json" {
		return printJSON(raw)
	}

	var response recordWriteResponse
	if err := json.Unmarshal(raw, &response); err != nil {
		return fmt.Errorf("failed to parse write response: %w", err)
	}

	switch response.Outcome {
	case "ACCEPTED":
		common.PrintSuccess("Value accepted")
	case "REPLAYED":
		common.PrintSuccess("Write already applied (idempotent replay)")
	default:
		return fmt.Errorf("write %s: %s", response.Outcome, recordRejectionMessage(response))
	}
	return nil
}

// recordParseWriteValue turns the raw --value string into the JSON value to
// send. JSON objects and arrays pass through; everything else is wrapped per
// kind. date, contact, and location require JSON.
func recordParseWriteValue(kind, raw string) (interface{}, error) {
	trimmed := strings.TrimSpace(raw)
	if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") {
		dec := json.NewDecoder(strings.NewReader(trimmed))
		dec.UseNumber()
		var parsed interface{}
		if err := dec.Decode(&parsed); err != nil {
			return nil, fmt.Errorf("invalid JSON for --kind %s: %w", kind, err)
		}
		switch parsed.(type) {
		case map[string]interface{}, []interface{}:
			return parsed, nil
		}
	}

	switch kind {
	case "text":
		return map[string]interface{}{"text": raw}, nil
	case "checkbox":
		parsed, err := strconv.ParseBool(trimmed)
		if err != nil {
			return nil, fmt.Errorf("--kind checkbox requires true or false, got %q", raw)
		}
		return parsed, nil
	case "number":
		if _, err := strconv.ParseFloat(trimmed, 64); err != nil {
			return nil, fmt.Errorf("--kind number requires a number, got %q", raw)
		}
		return map[string]interface{}{"number": json.Number(trimmed)}, nil
	case "select", "assignee", "reference", "file", "country":
		return recordSplitCommaList(trimmed), nil
	case "select-delta":
		return nil, fmt.Errorf(`--kind select-delta requires JSON like {"add":["opt_1"],"remove":["opt_2"]}`)
	case "date":
		return nil, fmt.Errorf(`--kind date requires JSON like {"start":"2026-09-13","end":null,"timezone":"America/New_York","granularity":"ALL_DAY"}`)
	case "contact":
		return nil, fmt.Errorf(`--kind contact requires JSON like {"text":"+1 555 0100","regionCode":"US"}`)
	case "location":
		return nil, fmt.Errorf(`--kind location requires JSON like {"text":"1600 Pennsylvania Ave","latitude":38.8977,"longitude":-77.0365}`)
	default:
		return nil, fmt.Errorf("unknown --kind %q. Use text, number, date, checkbox, select, select-delta, assignee, reference, file, country, contact, or location", kind)
	}
}

func recordSplitCommaList(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func recordRejectionMessage(response recordWriteResponse) string {
	if response.Rejection == nil {
		return response.Outcome
	}
	if response.Rejection.Code == "" {
		return response.Rejection.Message
	}
	return response.Rejection.Code + ": " + response.Rejection.Message
}
