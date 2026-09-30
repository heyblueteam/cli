package tables

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"reflect"
	"strings"
	"testing"
)

// fieldsTestRequest is one REST request observed by fieldsTestServer.
type fieldsTestRequest struct {
	Method string
	Path   string
	Query  url.Values
	Body   map[string]interface{}
}

// fieldsTestServer starts a server that answers the GraphQL workspace
// resolution on POST "/" with {"data":{"project":{"id":"w_real"}}} and hands
// every other request to respond. The config env vars point at the server.
func fieldsTestServer(t *testing.T, respond func(req fieldsTestRequest) (int, string)) {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":{"project":{"id":"w_real"}}}`))
			return
		}
		req := fieldsTestRequest{Method: r.Method, Path: r.URL.Path, Query: r.URL.Query()}
		if r.Body != nil {
			raw, err := io.ReadAll(r.Body)
			if err != nil {
				t.Errorf("read request body: %v", err)
			}
			if len(raw) > 0 {
				req.Body = map[string]interface{}{}
				if err := json.Unmarshal(raw, &req.Body); err != nil {
					t.Errorf("request body is not JSON: %v", err)
				}
			}
		}
		status, payload := respond(req)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(payload))
	}))
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("API_URL", server.URL)
	t.Setenv("AUTH_TOKEN", "token")
	t.Setenv("CLIENT_ID", "client")
	t.Setenv("COMPANY_ID", "company")
	t.Cleanup(server.Close)
}

// fieldsFailingServer refuses every request; tests that expect a client-side
// error before any network traffic use it.
func fieldsFailingServer(t *testing.T) {
	fieldsTestServer(t, func(req fieldsTestRequest) (int, string) {
		t.Errorf("unexpected request: %s %s", req.Method, req.Path)
		return http.StatusInternalServerError, `{"error":"unexpected request"}`
	})
}

// captureFieldsStdout swaps os.Stdout for a pipe and returns a restore
// function that yields everything written while it was active. The shared
// helpers (common.PrintSuccess, common.PrintInfo, printJSON) write to
// os.Stdout directly.
func captureFieldsStdout(t *testing.T) func() string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	original := os.Stdout
	os.Stdout = w
	done := make(chan struct{})
	var buf bytes.Buffer
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()
	return func() string {
		os.Stdout = original
		_ = w.Close()
		<-done
		_ = r.Close()
		return buf.String()
	}
}

// runFieldsCommand executes `blue tables fields <args...>` and returns the
// combined command output (cmd.OutOrStdout) and process stdout.
func runFieldsCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := newFieldsCmd()
	cmd.SetArgs(args)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	var out bytes.Buffer
	cmd.SetOut(&out)
	capture := captureFieldsStdout(t)
	err := cmd.Execute()
	return out.String() + capture(), err
}

func TestFieldsListText(t *testing.T) {
	var query url.Values
	fieldsTestServer(t, func(req fieldsTestRequest) (int, string) {
		if req.Method != http.MethodGet || req.Path != "/v1/databases/table-1/fields" {
			t.Errorf("unexpected request: %s %s", req.Method, req.Path)
		}
		query = req.Query
		return http.StatusOK, `[{"id":"f_1","name":"Priority","type":"SELECT_SINGLE","configuration":{"options":["a","b"]},"rank":"a0"},{"id":"f_2","name":"Notes","type":"TEXT_SINGLE","configuration":null,"rank":"a1"}]`
	})

	output, err := runFieldsCommand(t, "list", "--workspace", "w-slug", "--table", "table-1")
	if err != nil {
		t.Fatal(err)
	}

	if got := query.Get("workspaceId"); got != "w_real" {
		t.Fatalf("workspaceId query = %q", got)
	}
	for _, want := range []string{
		"RANK",
		"ID",
		"NAME",
		"TYPE",
		"a0",
		"f_1",
		"Priority",
		"SELECT_SINGLE",
		"f_2",
		"Notes",
		"TEXT_SINGLE",
	} {
		if !strings.Contains(output, want) {
			t.Fatalf("output %q does not contain %q", output, want)
		}
	}
	if strings.Contains(output, "No fields found") {
		t.Fatalf("output %q claims no fields found", output)
	}
}

func TestFieldsListJSON(t *testing.T) {
	fieldsTestServer(t, func(req fieldsTestRequest) (int, string) {
		return http.StatusOK, `[{"id":"f_1","name":"Priority","type":"SELECT_SINGLE","configuration":{"options":["a"]},"rank":"a0"}]`
	})

	output, err := runFieldsCommand(t, "list", "-w", "w-slug", "--table", "table-1", "--format", "json")
	if err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{`"id": "f_1"`, `"name": "Priority"`, `"type": "SELECT_SINGLE"`, `"rank": "a0"`, `"options"`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output %q does not contain %q", output, want)
		}
	}
	if strings.Contains(output, "RANK") {
		t.Fatalf("json output %q contains the text table header", output)
	}
}

func TestFieldsListCSV(t *testing.T) {
	fieldsTestServer(t, func(req fieldsTestRequest) (int, string) {
		return http.StatusOK, `[{"id":"f_1","name":"Priority","type":"SELECT_SINGLE","configuration":{"min":0},"rank":"a0"},{"id":"f_2","name":"Notes","type":"TEXT_SINGLE","configuration":null,"rank":"a1"}]`
	})

	output, err := runFieldsCommand(t, "list", "-w", "w-slug", "--table", "table-1", "--format", "csv")
	if err != nil {
		t.Fatal(err)
	}

	want := "id,name,type,rank,configuration\n" +
		"f_1,Priority,SELECT_SINGLE,a0,\"{\"\"min\"\":0}\"\n" +
		"f_2,Notes,TEXT_SINGLE,a1,\n"
	if output != want {
		t.Fatalf("csv output = %q, want %q", output, want)
	}
}

func TestFieldsListEmpty(t *testing.T) {
	fieldsTestServer(t, func(req fieldsTestRequest) (int, string) {
		return http.StatusOK, `[]`
	})

	output, err := runFieldsCommand(t, "list", "-w", "w-slug", "--table", "table-1")
	if err != nil {
		t.Fatal(err)
	}
	if output != "No fields found.\n" {
		t.Fatalf("empty output = %q", output)
	}
}

func TestFieldsListRequiresWorkspaceAndTable(t *testing.T) {
	fieldsFailingServer(t)

	_, err := runFieldsCommand(t, "list", "-w", "w-slug")
	if err == nil || !strings.Contains(err.Error(), "table is required. Use --table flag") {
		t.Fatalf("missing table error = %v", err)
	}

	_, err = runFieldsCommand(t, "list", "--table", "table-1")
	if err == nil || !strings.Contains(err.Error(), "workspace is required. Use --workspace flag") {
		t.Fatalf("missing workspace error = %v", err)
	}
}

func TestFieldsListServerError(t *testing.T) {
	fieldsTestServer(t, func(req fieldsTestRequest) (int, string) {
		return http.StatusUnprocessableEntity, `{"error":"Invalid table identity."}`
	})

	_, err := runFieldsCommand(t, "list", "-w", "w-slug", "--table", "table-1")
	if err == nil || !strings.Contains(err.Error(), "failed to list fields") || !strings.Contains(err.Error(), "Invalid table identity.") {
		t.Fatalf("server error = %v", err)
	}
}

func TestFieldsCreate(t *testing.T) {
	var body map[string]interface{}
	fieldsTestServer(t, func(req fieldsTestRequest) (int, string) {
		if req.Method != http.MethodPost || req.Path != "/v1/databases/table-1/fields" {
			t.Errorf("unexpected request: %s %s", req.Method, req.Path)
		}
		body = req.Body
		return http.StatusOK, `{"fieldId":"f_new","fieldCreated":true,"mounted":true,"rank":"a0"}`
	})

	output, err := runFieldsCommand(t, "create", "-w", "w-slug", "--table", "table-1",
		"--name", "Priority", "--type", "SELECT_SINGLE", "--field-id", "f_new")
	if err != nil {
		t.Fatal(err)
	}

	if body["workspaceId"] != "w_real" || body["fieldId"] != "f_new" || body["name"] != "Priority" || body["type"] != "SELECT_SINGLE" {
		t.Fatalf("unexpected body: %#v", body)
	}
	if _, hasConfiguration := body["configuration"]; hasConfiguration {
		t.Fatalf("configuration sent without --configuration: %#v", body)
	}
	for _, want := range []string{"✓ Field created", "Field ID: f_new", "Rank: a0", "Mounted: true"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output %q does not contain %q", output, want)
		}
	}
}

func TestFieldsCreateWithConfiguration(t *testing.T) {
	var body map[string]interface{}
	fieldsTestServer(t, func(req fieldsTestRequest) (int, string) {
		body = req.Body
		return http.StatusOK, `{"fieldId":"f_new","fieldCreated":true,"mounted":true,"rank":"a0"}`
	})

	output, err := runFieldsCommand(t, "create", "-w", "w-slug", "--table", "table-1",
		"--name", "Priority", "--type", "SELECT_SINGLE", "--configuration", `{"options":["a","b"]}`, "--format", "json")
	if err != nil {
		t.Fatal(err)
	}

	wantConfiguration := map[string]interface{}{"options": []interface{}{"a", "b"}}
	if !reflect.DeepEqual(body["configuration"], wantConfiguration) {
		t.Fatalf("configuration = %#v, want %#v", body["configuration"], wantConfiguration)
	}
	for _, want := range []string{`"fieldId": "f_new"`, `"fieldCreated": true`, `"mounted": true`} {
		if !strings.Contains(output, want) {
			t.Fatalf("output %q does not contain %q", output, want)
		}
	}
}

func TestFieldsCreateDefaultFieldID(t *testing.T) {
	var body map[string]interface{}
	fieldsTestServer(t, func(req fieldsTestRequest) (int, string) {
		body = req.Body
		return http.StatusOK, `{"fieldId":"c_generated","fieldCreated":true,"mounted":true,"rank":"a0"}`
	})

	if _, err := runFieldsCommand(t, "create", "-w", "w-slug", "--table", "table-1",
		"--name", "Estimate", "--type", "NUMBER"); err != nil {
		t.Fatal(err)
	}

	fieldID, ok := body["fieldId"].(string)
	if !ok || fieldID == "" {
		t.Fatalf("fieldId = %#v, want a generated non-empty string", body["fieldId"])
	}
	if !strings.HasPrefix(fieldID, "c") {
		t.Fatalf("generated fieldId = %q, want a cuid starting with 'c'", fieldID)
	}
}

func TestFieldsCreateInvalidConfiguration(t *testing.T) {
	fieldsFailingServer(t)

	_, err := runFieldsCommand(t, "create", "-w", "w-slug", "--table", "table-1",
		"--name", "Estimate", "--type", "NUMBER", "--configuration", "{not json")
	if err == nil || !strings.Contains(err.Error(), "invalid --configuration JSON") {
		t.Fatalf("invalid configuration error = %v", err)
	}
}

func TestFieldsCreateAlreadyExists(t *testing.T) {
	fieldsTestServer(t, func(req fieldsTestRequest) (int, string) {
		return http.StatusOK, `{"fieldId":"f_new","fieldCreated":false,"mounted":true,"rank":"a0"}`
	})

	output, err := runFieldsCommand(t, "create", "-w", "w-slug", "--table", "table-1",
		"--name", "Priority", "--type", "SELECT_SINGLE", "--field-id", "f_new")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"✓ Field created", "already existed"} {
		if !strings.Contains(output, want) {
			t.Fatalf("output %q does not contain %q", output, want)
		}
	}
}

func TestFieldsRename(t *testing.T) {
	var body map[string]interface{}
	fieldsTestServer(t, func(req fieldsTestRequest) (int, string) {
		if req.Method != http.MethodPost || req.Path != "/v1/rename-database-field" {
			t.Errorf("unexpected request: %s %s", req.Method, req.Path)
		}
		body = req.Body
		return http.StatusOK, `{"fieldId":"f_1","renamed":true}`
	})

	output, err := runFieldsCommand(t, "rename", "-w", "w-slug", "--field", "f_1", "--name", "New name")
	if err != nil {
		t.Fatal(err)
	}

	if body["workspaceId"] != "w_real" || body["fieldId"] != "f_1" || body["name"] != "New name" {
		t.Fatalf("unexpected body: %#v", body)
	}
	if !strings.Contains(output, "✓ Field renamed") {
		t.Fatalf("output = %q", output)
	}

	jsonOutput, err := runFieldsCommand(t, "rename", "-w", "w-slug", "--field", "f_1", "--name", "New name", "--format", "json")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(jsonOutput, `"renamed": true`) {
		t.Fatalf("json output = %q", jsonOutput)
	}
}

func TestFieldsDeleteWithoutConfirm(t *testing.T) {
	fieldsFailingServer(t)

	_, err := runFieldsCommand(t, "delete", "-w", "w-slug", "--field", "f_1")
	if err == nil || !strings.Contains(err.Error(), "deletion confirmation is required") {
		t.Fatalf("missing confirm error = %v", err)
	}
}

func TestFieldsDeleteWithConfirm(t *testing.T) {
	var body map[string]interface{}
	fieldsTestServer(t, func(req fieldsTestRequest) (int, string) {
		if req.Method != http.MethodPost || req.Path != "/v1/delete-database-field" {
			t.Errorf("unexpected request: %s %s", req.Method, req.Path)
		}
		body = req.Body
		return http.StatusOK, `{"fieldId":"f_1","deleted":true}`
	})

	output, err := runFieldsCommand(t, "delete", "-w", "w-slug", "--field", "f_1", "--confirm")
	if err != nil {
		t.Fatal(err)
	}

	if body["workspaceId"] != "w_real" || body["fieldId"] != "f_1" {
		t.Fatalf("unexpected body: %#v", body)
	}
	if !strings.Contains(output, "✓ Field deleted") {
		t.Fatalf("output = %q", output)
	}
}

func TestFieldsMount(t *testing.T) {
	var body map[string]interface{}
	fieldsTestServer(t, func(req fieldsTestRequest) (int, string) {
		if req.Method != http.MethodPost || req.Path != "/v1/mount-database-field" {
			t.Errorf("unexpected request: %s %s", req.Method, req.Path)
		}
		body = req.Body
		return http.StatusOK, `{"databaseId":"table-2","fieldId":"f_1","mounted":true}`
	})

	output, err := runFieldsCommand(t, "mount", "-w", "w-slug", "--table", "table-2", "--field", "f_1")
	if err != nil {
		t.Fatal(err)
	}

	if body["workspaceId"] != "w_real" || body["databaseId"] != "table-2" || body["fieldId"] != "f_1" {
		t.Fatalf("unexpected body: %#v", body)
	}
	if !strings.Contains(output, "✓ Field mounted on table") {
		t.Fatalf("output = %q", output)
	}
}

func TestFieldsMountAlreadyMounted(t *testing.T) {
	fieldsTestServer(t, func(req fieldsTestRequest) (int, string) {
		return http.StatusOK, `{"databaseId":"table-2","fieldId":"f_1","mounted":false}`
	})

	output, err := runFieldsCommand(t, "mount", "-w", "w-slug", "--table", "table-2", "--field", "f_1")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output, "✓") {
		t.Fatalf("no-op mount output = %q", output)
	}
	if !strings.Contains(output, "Field already mounted on table") {
		t.Fatalf("output = %q", output)
	}
}

func TestFieldsUnmount(t *testing.T) {
	var body map[string]interface{}
	fieldsTestServer(t, func(req fieldsTestRequest) (int, string) {
		if req.Method != http.MethodPost || req.Path != "/v1/unmount-database-field" {
			t.Errorf("unexpected request: %s %s", req.Method, req.Path)
		}
		body = req.Body
		return http.StatusOK, `{"databaseId":"table-2","fieldId":"f_1","unmounted":true}`
	})

	output, err := runFieldsCommand(t, "unmount", "-w", "w-slug", "--table", "table-2", "--field", "f_1")
	if err != nil {
		t.Fatal(err)
	}

	if body["workspaceId"] != "w_real" || body["databaseId"] != "table-2" || body["fieldId"] != "f_1" {
		t.Fatalf("unexpected body: %#v", body)
	}
	if !strings.Contains(output, "✓ Field unmounted from table") {
		t.Fatalf("output = %q", output)
	}
}

func TestFieldsUnmountAlreadyUnmounted(t *testing.T) {
	fieldsTestServer(t, func(req fieldsTestRequest) (int, string) {
		return http.StatusOK, `{"databaseId":"table-2","fieldId":"f_1","unmounted":false}`
	})

	output, err := runFieldsCommand(t, "unmount", "-w", "w-slug", "--table", "table-2", "--field", "f_1")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output, "✓") {
		t.Fatalf("no-op unmount output = %q", output)
	}
	if !strings.Contains(output, "Field already not mounted on table") {
		t.Fatalf("output = %q", output)
	}
}
