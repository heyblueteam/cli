package charts

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveDisplayAndChartType(t *testing.T) {
	tests := map[string]string{
		"bar": "BAR", "line": "BAR", "area": "BAR", "row": "BAR", "leaderboard": "BAR",
		"table": "BAR", "funnel": "BAR", "combo": "BAR", "pie": "PIE",
		"stat": "STAT", "progress": "STAT", "gauge": "STAT",
	}
	for display, wantType := range tests {
		t.Run(display, func(t *testing.T) {
			gotDisplay, gotType, err := resolveDisplayAndChartType(display, "")
			if err != nil {
				t.Fatal(err)
			}
			if gotDisplay != display || gotType != wantType {
				t.Fatalf("got %s/%s, want %s/%s", gotDisplay, gotType, display, wantType)
			}
		})
	}
	if _, _, err := resolveDisplayAndChartType("unknown", ""); err == nil {
		t.Fatal("unknown display type was accepted")
	}
	if _, _, err := resolveDisplayAndChartType("line", "PIE"); err == nil {
		t.Fatal("conflicting legacy type was accepted")
	}
	if display, chartType, err := resolveDisplayAndChartType("", "BAR"); err != nil || display != "bar" || chartType != "BAR" {
		t.Fatalf("legacy BAR: %s/%s/%v", display, chartType, err)
	}
}

func TestLoadJSONInput(t *testing.T) {
	path := filepath.Join(t.TempDir(), "chart.json")
	if err := os.WriteFile(path, []byte(`{"dashboardId":"d1","title":"A chart","type":"BAR"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	input, err := loadJSONInput(path)
	if err != nil {
		t.Fatal(err)
	}
	if input["dashboardId"] != "d1" {
		t.Fatalf("unexpected input: %#v", input)
	}
	if err := os.WriteFile(path, []byte(`[]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := loadJSONInput(path); err == nil {
		t.Fatal("array input was accepted")
	}
}

func TestCreateWithJSONUsesVariables(t *testing.T) {
	var request struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		if err := json.Unmarshal(body, &request); err != nil {
			t.Error(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"createChart":{"id":"chart-1","title":"Quoted \"chart\"","type":"BAR","displayType":"line","isCalculating":true,"chartSegments":[]}}}`))
	}))
	defer server.Close()
	t.Setenv("API_URL", server.URL)
	t.Setenv("AUTH_TOKEN", "token")
	t.Setenv("CLIENT_ID", "client")
	t.Setenv("COMPANY_ID", "company")
	path := filepath.Join(t.TempDir(), "chart.json")
	payload := `{"dashboardId":"d1","title":"Quoted \"chart\"","type":"BAR","displayType":"line","metadata":{"query":{"dimensions":[{"type":"TODO_STATUS"}],"metrics":[{"key":"value"}]}}}`
	if err := os.WriteFile(path, []byte(payload), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := newChartInputCommand(false)
	cmd.SetArgs([]string{"--input", path, "--format", "json"})
	cmd.SilenceUsage = true
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	input := request.Variables["input"].(map[string]interface{})
	if input["title"] != `Quoted "chart"` {
		t.Fatalf("title did not survive variables: %#v", input["title"])
	}
	if request.Query == "" {
		t.Fatal("no GraphQL query captured")
	}
}

func TestCreateWithFlagsBuildsCurrentQueryMetadata(t *testing.T) {
	var createInput map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Query     string                 `json:"query"`
			Variables map[string]interface{} `json:"variables"`
		}
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		w.Header().Set("Content-Type", "application/json")
		if request.Variables["project"] != nil {
			_, _ = w.Write([]byte(`{"data":{"project":{"id":"workspace-1"}}}`))
			return
		}
		createInput = request.Variables["input"].(map[string]interface{})
		_, _ = w.Write([]byte(`{"data":{"createChart":{"id":"chart-1","title":"Completed by month","type":"BAR","displayType":"line","chartSegments":[]}}}`))
	}))
	defer server.Close()
	t.Setenv("API_URL", server.URL)
	t.Setenv("AUTH_TOKEN", "token")
	t.Setenv("CLIENT_ID", "client")
	t.Setenv("COMPANY_ID", "company")
	cmd := newChartInputCommand(false)
	cmd.SetArgs([]string{"--dashboard", "d1", "--title", "Completed by month", "--display-type", "line", "--workspace", "delivery", "--group-by", "TODO_COMPLETED_AT", "--interval", "MONTH", "--filter-json", `{"showCompleted":true}`})
	cmd.SilenceUsage = true
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	metadata := createInput["metadata"].(map[string]interface{})
	query := metadata["query"].(map[string]interface{})
	dimension := query["dimensions"].([]interface{})[0].(map[string]interface{})
	if dimension["type"] != "TODO_COMPLETED_AT" || dimension["interval"] != "MONTH" {
		t.Fatalf("unexpected dimension: %#v", dimension)
	}
	metric := query["metrics"].([]interface{})[0].(map[string]interface{})
	if _, exists := metric["function"]; exists {
		t.Fatalf("record count must omit function: %#v", metric)
	}
	filters := query["filters"].(map[string]interface{})
	if filters["showCompleted"] != true {
		t.Fatalf("filter was lost: %#v", filters)
	}
	projects := filters["projectIds"].([]interface{})
	if len(projects) != 1 || projects[0] != "workspace-1" {
		t.Fatalf("workspace was not resolved: %#v", projects)
	}
}

func TestInputRejectsPayloadFlags(t *testing.T) {
	path := filepath.Join(t.TempDir(), "chart.json")
	if err := os.WriteFile(path, []byte(`{"dashboardId":"d1"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	cmd := newChartInputCommand(false)
	cmd.SetArgs([]string{"--input", path, "--title", "conflict"})
	cmd.SilenceUsage = true
	if err := cmd.Execute(); err == nil {
		t.Fatal("mixed JSON and flag input was accepted")
	}
}
