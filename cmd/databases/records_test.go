package databases

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"
)

type recordsTestRequest struct {
	Method string
	Path   string
	Query  url.Values
	Body   map[string]interface{}
}

// recordsStartServer runs one mock server for both the GraphQL workspace
// resolution (POST "/") and the REST /v1 routes. The canned REST response is
// returned for every /v1 request; the last REST request is recorded in got.
func recordsStartServer(t *testing.T, response string, got *recordsTestRequest) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":{"project":{"id":"w_real"}}}`))
			return
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
			return
		}
		got.Method = r.Method
		got.Path = r.URL.Path
		got.Query = r.URL.Query()
		if len(body) > 0 {
			if err := json.Unmarshal(body, &got.Body); err != nil {
				t.Errorf("request body is not JSON: %v", err)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(response))
	}))
	t.Cleanup(server.Close)

	recordsSetTestEnv(t, server.URL)
}

// recordsSetTestEnv points the client at an unreachable API for tests that
// must fail before any request is made.
func recordsSetTestEnv(t *testing.T, apiURL string) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("API_URL", apiURL)
	t.Setenv("AUTH_TOKEN", "token")
	t.Setenv("CLIENT_ID", "client")
	t.Setenv("COMPANY_ID", "company")
}

// recordsCaptureStdout runs fn while capturing everything written to
// os.Stdout (common.PrintSuccess writes there, not to cmd.OutOrStdout()).
func recordsCaptureStdout(t *testing.T, fn func()) string {
	t.Helper()
	original := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w
	defer func() { os.Stdout = original }()

	fn()

	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	data, err := io.ReadAll(r)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestRecordsListTextRendersFieldsAndNextPage(t *testing.T) {
	var got recordsTestRequest
	recordsStartServer(t, `{"rows":[
		{"record":{"id":"r_1","createdAt":"2026-09-13T06:00:00.000Z","updatedAt":"2026-09-14T07:00:00.000Z"},
		 "databaseId":"t_1","membership":{"archived":false,"pinned":false,"editable":true},
		 "fields":[
			{"id":"f_1","name":"Priority","type":"SELECT_SINGLE","value":{"options":[{"id":"opt_1","title":"High","color":"red","icon":null,"position":0}]},"editable":true},
			{"id":"f_2","name":"Notes","type":"TEXT_MULTI","value":null},
			{"id":"f_3","name":"Budget","type":"NUMBER","value":{"number":"1.5"}},
			{"id":"f_4","name":"Due","type":"DATE","value":{"startDate":"2026-09-13","endDate":null,"timezone":"America/New_York","dateGranularity":"ALL_DAY"}}
		 ]},
		{"record":{"id":"r_2","createdAt":"2026-09-15T06:00:00.000Z","updatedAt":"2026-09-15T06:00:00.000Z"},
		 "databaseId":"t_1","membership":{"archived":false,"pinned":false,"editable":true},"fields":[]}
	],"nextCursor":"cur_2","hasMore":true}`, &got)

	cmd := newRecordsListCmd()
	cmd.SetArgs([]string{"--database", "t_1", "--workspace", "ws_1"})
	cmd.SilenceUsage = true
	var output bytes.Buffer
	cmd.SetOut(&output)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	if got.Method != http.MethodGet || got.Path != "/v1/databases/t_1/records" {
		t.Fatalf("request was %s %s, want GET /v1/databases/t_1/records", got.Method, got.Path)
	}
	if got.Query.Get("workspaceId") != "w_real" {
		t.Fatalf("workspaceId = %q, want w_real", got.Query.Get("workspaceId"))
	}
	if got.Query.Get("first") != "50" {
		t.Fatalf("first = %q, want 50", got.Query.Get("first"))
	}
	if got.Query.Get("cursor") != "" || got.Query.Get("includeArchived") != "" {
		t.Fatalf("unexpected query params: %v", got.Query)
	}

	out := output.String()
	for _, want := range []string{
		"=== Records ===",
		"1. r_1",
		"   Created: 2026-09-13T06:00:00.000Z",
		"   Updated: 2026-09-14T07:00:00.000Z",
		"   Priority: High",
		"   Budget: 1.5",
		"   Due: 2026-09-13",
		"2. r_2",
		"\nNext page: blue databases records list --database t_1 --after cur_2 --workspace ws_1\n",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output %q does not contain %q", out, want)
		}
	}
	if strings.Contains(out, "Notes") {
		t.Fatalf("null field value was not skipped: %q", out)
	}
}

func TestRecordsListCSV(t *testing.T) {
	var got recordsTestRequest
	recordsStartServer(t, `{"rows":[
		{"record":{"id":"r_1","createdAt":"2026-09-13T06:00:00.000Z","updatedAt":"2026-09-14T07:00:00.000Z"},"membership":{"archived":false},"fields":[
			{"id":"f_1","name":"Name","type":"TEXT_SINGLE","value":{ "text": "Deal A" }},
			{"id":"f_2","name":"Score","type":"NUMBER","value":{ "number": "7" }}
		]},
		{"record":{"id":"r_2","createdAt":"2026-09-15T06:00:00.000Z","updatedAt":"2026-09-15T06:00:00.000Z"},"membership":{"archived":true},"fields":[
			{"id":"f_2","name":"Score","type":"NUMBER","value":null}
		]}
	],"nextCursor":null,"hasMore":false}`, &got)

	cmd := newRecordsListCmd()
	cmd.SetArgs([]string{"--database", "t_1", "--workspace", "ws_1", "--format", "csv"})
	cmd.SilenceUsage = true
	var output bytes.Buffer
	cmd.SetOut(&output)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	rows, err := csv.NewReader(strings.NewReader(output.String())).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	want := [][]string{
		{"recordId", "createdAt", "updatedAt", "archived", "Name", "Score"},
		{"r_1", "2026-09-13T06:00:00.000Z", "2026-09-14T07:00:00.000Z", "false", `{"text":"Deal A"}`, `{"number":"7"}`},
		{"r_2", "2026-09-15T06:00:00.000Z", "2026-09-15T06:00:00.000Z", "true", "", ""},
	}
	if len(rows) != len(want) {
		t.Fatalf("got %d rows, want %d: %v", len(rows), len(want), rows)
	}
	for i, row := range rows {
		for j, cell := range row {
			if cell != want[i][j] {
				t.Fatalf("row %d cell %d = %q, want %q (rows: %v)", i, j, cell, want[i][j], rows)
			}
		}
	}
}

func TestRecordsListValidatesLimit(t *testing.T) {
	recordsSetTestEnv(t, "http://127.0.0.1:1")

	cmd := newRecordsListCmd()
	cmd.SetArgs([]string{"--database", "t_1", "--workspace", "ws_1", "--limit", "0"})
	cmd.SilenceUsage = true
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "between 1 and 200") {
		t.Fatalf("--limit 0 error = %v", err)
	}

	cmd = newRecordsListCmd()
	cmd.SetArgs([]string{"--database", "t_1", "--workspace", "ws_1", "--limit", "300"})
	cmd.SilenceUsage = true
	err = cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "between 1 and 200") {
		t.Fatalf("--limit 300 error = %v", err)
	}

	cmd = newRecordsListCmd()
	cmd.SetArgs([]string{"--workspace", "ws_1"})
	cmd.SilenceUsage = true
	err = cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "--database") {
		t.Fatalf("missing --database error = %v", err)
	}
}

func TestRecordsCreatePostsFieldValues(t *testing.T) {
	var got recordsTestRequest
	recordsStartServer(t, `{"recordId":"r_new","databaseId":"t_1","created":true,"joined":false,"rank":"a0"}`, &got)

	cmd := newRecordsCreateCmd()
	cmd.SetArgs([]string{
		"--database", "t_1",
		"--title", "New deal",
		"--group", "grp_1",
		"--field-value", `{"fieldId":"f_1","kind":"text","value":{"text":"hi"}}`,
		"--workspace", "ws_1",
	})
	cmd.SilenceUsage = true
	output := recordsCaptureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if got.Method != http.MethodPost || got.Path != "/v1/databases/t_1/records" {
		t.Fatalf("request was %s %s, want POST /v1/databases/t_1/records", got.Method, got.Path)
	}
	if got.Body["workspaceId"] != "w_real" || got.Body["title"] != "New deal" || got.Body["groupOptionId"] != "grp_1" {
		t.Fatalf("unexpected body: %#v", got.Body)
	}
	fieldValues, ok := got.Body["fieldValues"].([]interface{})
	if !ok || len(fieldValues) != 1 {
		t.Fatalf("fieldValues = %#v", got.Body["fieldValues"])
	}
	field, _ := fieldValues[0].(map[string]interface{})
	if field["fieldId"] != "f_1" || field["kind"] != "text" {
		t.Fatalf("fieldValues[0] = %#v", fieldValues[0])
	}
	value, _ := field["value"].(map[string]interface{})
	if value["text"] != "hi" {
		t.Fatalf("fieldValues[0].value = %#v", field["value"])
	}

	for _, want := range []string{"✓ Record created", "  Record ID: r_new", "  Rank: a0"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output %q does not contain %q", output, want)
		}
	}
}

func TestRecordsCreateValidatesFieldValueJSON(t *testing.T) {
	recordsSetTestEnv(t, "http://127.0.0.1:1")

	cmd := newRecordsCreateCmd()
	cmd.SetArgs([]string{"--database", "t_1", "--title", "New deal", "--workspace", "ws_1", "--field-value", "{bad"})
	cmd.SilenceUsage = true
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "invalid --field-value JSON") {
		t.Fatalf("bad JSON error = %v", err)
	}

	cmd = newRecordsCreateCmd()
	cmd.SetArgs([]string{"--database", "t_1", "--title", "New deal", "--workspace", "ws_1", "--field-value", `{"fieldId":"f_1"}`})
	cmd.SilenceUsage = true
	err = cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "invalid --field-value JSON") || !strings.Contains(err.Error(), `missing the "kind" key`) {
		t.Fatalf("missing key error = %v", err)
	}
}

func TestRecordsSetValueSendsDefaultsAndJSONValue(t *testing.T) {
	var got recordsTestRequest
	recordsStartServer(t, `{"outcome":"ACCEPTED","targetCursor":"tc_1","entityRevision":3,"entityUpdatedAt":"2026-09-30T00:00:00Z","rejection":null}`, &got)

	cmd := newRecordsSetValueCmd()
	cmd.SetArgs([]string{
		"--database", "t_1",
		"--record", "r_1",
		"--field", "f_1",
		"--kind", "text",
		"--value", `{"text":"hi"}`,
		"--workspace", "ws_1",
	})
	cmd.SilenceUsage = true
	output := recordsCaptureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if got.Method != http.MethodPatch || got.Path != "/v1/databases/t_1/records/r_1/fields/f_1" {
		t.Fatalf("request was %s %s, want PATCH /v1/databases/t_1/records/r_1/fields/f_1", got.Method, got.Path)
	}
	if got.Body["workspaceId"] != "w_real" || got.Body["kind"] != "text" {
		t.Fatalf("unexpected body: %#v", got.Body)
	}
	value, _ := got.Body["value"].(map[string]interface{})
	if value["text"] != "hi" {
		t.Fatalf("value = %#v", got.Body["value"])
	}
	mutationID, _ := got.Body["clientMutationId"].(string)
	if mutationID == "" {
		t.Fatalf("clientMutationId = %#v, want a generated cuid", got.Body["clientMutationId"])
	}
	sequence, _ := got.Body["entitySequence"].(float64)
	if sequence <= 0 {
		t.Fatalf("entitySequence = %#v, want > 0", got.Body["entitySequence"])
	}
	if !strings.Contains(output, "✓ Value accepted") {
		t.Fatalf("output %q does not contain success line", output)
	}
}

func TestRecordsSetValueWrapsBareText(t *testing.T) {
	var got recordsTestRequest
	recordsStartServer(t, `{"outcome":"ACCEPTED"}`, &got)

	cmd := newRecordsSetValueCmd()
	cmd.SetArgs([]string{
		"--database", "t_1", "--record", "r_1", "--field", "f_1",
		"--kind", "text", "--value", "hello", "--workspace", "ws_1",
	})
	cmd.SilenceUsage = true
	recordsCaptureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	value, ok := got.Body["value"].(map[string]interface{})
	if !ok || value["text"] != "hello" {
		t.Fatalf("value = %#v, want map text=hello", got.Body["value"])
	}
}

func TestRecordsSetValueWrapsCommaList(t *testing.T) {
	var got recordsTestRequest
	recordsStartServer(t, `{"outcome":"ACCEPTED"}`, &got)

	cmd := newRecordsSetValueCmd()
	cmd.SetArgs([]string{
		"--database", "t_1", "--record", "r_1", "--field", "f_1",
		"--kind", "select", "--value", "opt_1, opt_2", "--workspace", "ws_1",
	})
	cmd.SilenceUsage = true
	recordsCaptureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	list, ok := got.Body["value"].([]interface{})
	if !ok || len(list) != 2 || list[0] != "opt_1" || list[1] != "opt_2" {
		t.Fatalf("value = %#v, want [opt_1 opt_2]", got.Body["value"])
	}
}

func TestRecordsSetValueReportsRejection(t *testing.T) {
	var got recordsTestRequest
	recordsStartServer(t, `{"outcome":"REJECTED","targetCursor":"","rejection":{"code":"VALIDATION","message":"Field is required"}}`, &got)

	cmd := newRecordsSetValueCmd()
	cmd.SetArgs([]string{
		"--database", "t_1", "--record", "r_1", "--field", "f_1",
		"--kind", "text", "--value", "hello", "--workspace", "ws_1",
	})
	cmd.SilenceUsage = true
	err := cmd.Execute()
	if err == nil {
		t.Fatal("REJECTED outcome must return an error")
	}
	for _, want := range []string{"write REJECTED", "VALIDATION", "Field is required"} {
		if !strings.Contains(err.Error(), want) {
			t.Fatalf("error %q does not contain %q", err.Error(), want)
		}
	}
}

func TestRecordsSetValueDateRequiresJSON(t *testing.T) {
	recordsSetTestEnv(t, "http://127.0.0.1:1")

	cmd := newRecordsSetValueCmd()
	cmd.SetArgs([]string{
		"--database", "t_1", "--record", "r_1", "--field", "f_1",
		"--kind", "date", "--value", "2026-09-13", "--workspace", "ws_1",
	})
	cmd.SilenceUsage = true
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), `--kind date requires JSON like {"start":"2026-09-13"`) {
		t.Fatalf("date non-JSON error = %v", err)
	}
}
