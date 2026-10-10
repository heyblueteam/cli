package databases

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newViewsServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":{"project":{"id":"w_real"}}}`))
			return
		}
		handler(w, r)
	}))
	return server
}

const viewsFixture = `[
	{"id":"r_1","rank":"a0","view":{"id":"v_1","name":"All deals","database":{"id":"t_1","name":"Deals"},"viewType":null}},
	{"id":"r_2","rank":"a1","view":{"id":"v_2","name":"Calendar","database":{"id":"t_1","name":"Deals"},"viewType":"CALENDAR"}}
]`

func TestViewsJSON(t *testing.T) {
	var method, path string
	var body struct {
		WorkspaceID string `json:"workspaceId"`
	}
	server := newViewsServer(t, func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Error(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(viewsFixture))
	})
	defer server.Close()
	setDatabasesEnv(t, server)

	cmd := newViewsCmd()
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"--workspace", "w_slug", "--format", "json"})

	output := captureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if method != "POST" || path != "/v1/list-database-views" {
		t.Fatalf("request = %s %s", method, path)
	}
	if body.WorkspaceID != "w_real" {
		t.Fatalf("workspaceId body = %q", body.WorkspaceID)
	}
	if !strings.Contains(output, "All deals") || !strings.Contains(output, "v_1") {
		t.Fatalf("json output does not contain views")
	}
}

func TestViewsText(t *testing.T) {
	server := newViewsServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(viewsFixture))
	})
	defer server.Close()
	setDatabasesEnv(t, server)

	cmd := newViewsCmd()
	cmd.SilenceUsage = true
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--workspace", "w_slug"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"=== Views ===",
		"1. All deals",
		"   ID: v_1",
		"   Database: Deals (t_1)",
		"   Kind: -",
		"\n\n2. Calendar",
		"   Kind: CALENDAR",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("output %q does not contain %q", out.String(), want)
		}
	}
}

func TestViewsCSV(t *testing.T) {
	server := newViewsServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(viewsFixture))
	})
	defer server.Close()
	setDatabasesEnv(t, server)

	cmd := newViewsCmd()
	cmd.SilenceUsage = true
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--workspace", "w_slug", "--format", "csv"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	want := "id,name,databaseId,databaseName,viewType\nv_1,All deals,t_1,Deals,\nv_2,Calendar,t_1,Deals,CALENDAR\n"
	if out.String() != want {
		t.Fatalf("csv output = %q, want %q", out.String(), want)
	}
}

func TestViewsRequiresWorkspace(t *testing.T) {
	cmd := newViewsCmd()
	cmd.SilenceUsage = true
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "workspace is required. Use --workspace flag") {
		t.Fatalf("missing workspace error = %v", err)
	}
}
