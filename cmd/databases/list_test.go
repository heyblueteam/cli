package databases

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func setDatabasesEnv(t *testing.T, server *httptest.Server) {
	t.Helper()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("API_URL", server.URL)
	t.Setenv("AUTH_TOKEN", "secret")
	t.Setenv("CLIENT_ID", "client")
	t.Setenv("COMPANY_ID", "company")
}

// captureStdout collects everything printed to os.Stdout while fn runs.
// printJSON writes straight to os.Stdout, so JSON-format tests need this.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = writer
	defer func() { os.Stdout = orig }()

	fn()

	writer.Close()
	out, err := io.ReadAll(reader)
	if err != nil {
		t.Fatal(err)
	}
	return string(out)
}

func newDatabasesServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"data":{"project":{"id":"w_real"}}}`))
			return
		}
		handler(w, r)
	}))
}

func TestListDatabasesJSON(t *testing.T) {
	var method, path, workspace string
	server := newDatabasesServer(func(w http.ResponseWriter, r *http.Request) {
		method = r.Method
		path = r.URL.Path
		workspace = r.URL.Query().Get("workspaceId")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"t_1","name":"Deals"},{"id":"t_2","name":"Projects"}]`))
	})
	defer server.Close()
	setDatabasesEnv(t, server)

	cmd := newListCmd()
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"--workspace", "w_slug", "--format", "json"})

	output := captureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatal(err)
		}
	})

	if method != "GET" || path != "/v1/databases" {
		t.Fatalf("request = %s %s", method, path)
	}
	if workspace != "w_real" {
		t.Fatalf("workspaceId = %q", workspace)
	}
	if !strings.Contains(output, "Deals") || !strings.Contains(output, "t_1") {
		t.Fatalf("json output %q does not contain databases", output)
	}
}

func TestListDatabasesText(t *testing.T) {
	server := newDatabasesServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"t_1","name":"Deals"},{"id":"t_2","name":"Projects"}]`))
	})
	defer server.Close()
	setDatabasesEnv(t, server)

	cmd := newListCmd()
	cmd.SilenceUsage = true
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--workspace", "w_slug"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"=== Databases ===",
		"1. Deals",
		"   ID: t_1",
		"\n\n2. Projects",
		"   ID: t_2",
	} {
		if !strings.Contains(out.String(), want) {
			t.Fatalf("output %q does not contain %q", out.String(), want)
		}
	}
}

func TestListDatabasesCSV(t *testing.T) {
	server := newDatabasesServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[{"id":"t_1","name":"Deals"}]`))
	})
	defer server.Close()
	setDatabasesEnv(t, server)

	cmd := newListCmd()
	cmd.SilenceUsage = true
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--workspace", "w_slug", "--format", "csv"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	want := "id,name\nt_1,Deals\n"
	if out.String() != want {
		t.Fatalf("csv output = %q, want %q", out.String(), want)
	}
}

func TestListDatabasesEmpty(t *testing.T) {
	server := newDatabasesServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`[]`))
	})
	defer server.Close()
	setDatabasesEnv(t, server)

	cmd := newListCmd()
	cmd.SilenceUsage = true
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetArgs([]string{"--workspace", "w_slug"})
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}
	if got := out.String(); got != "No databases found.\n" {
		t.Fatalf("empty output = %q", got)
	}
}

func TestListDatabasesRequiresWorkspace(t *testing.T) {
	cmd := newListCmd()
	cmd.SilenceUsage = true
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "workspace is required. Use --workspace flag") {
		t.Fatalf("missing workspace error = %v", err)
	}
}

func TestListDatabasesInvalidFormat(t *testing.T) {
	cmd := newListCmd()
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"--workspace", "w_slug", "--format", "yaml"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), `invalid format "yaml"`) {
		t.Fatalf("invalid format error = %v", err)
	}
}

func TestListDatabasesAPIError(t *testing.T) {
	server := newDatabasesServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		_, _ = w.Write([]byte(`{"error":"Invalid database identity."}`))
	})
	defer server.Close()
	setDatabasesEnv(t, server)

	cmd := newListCmd()
	cmd.SilenceUsage = true
	cmd.SetArgs([]string{"--workspace", "w_slug"})
	err := cmd.Execute()
	if err == nil || !strings.Contains(err.Error(), "Invalid database identity.") {
		t.Fatalf("API error = %v", err)
	}
}
