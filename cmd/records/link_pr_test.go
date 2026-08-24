package records

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestLinkPRRequiresRecordAndPR(t *testing.T) {
	cmd := newLinkPRCmd()
	cmd.SilenceUsage = true
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "record ID is required") {
		t.Fatalf("missing record error = %v", err)
	}

	cmd = newLinkPRCmd()
	cmd.SetArgs([]string{"--record", "record-1"})
	cmd.SilenceUsage = true
	if err := cmd.Execute(); err == nil || !strings.Contains(err.Error(), "pull request is required") {
		t.Fatalf("missing PR error = %v", err)
	}
}

func TestLinkPRUsesGraphQLVariablesAndSimpleOutput(t *testing.T) {
	var request struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			t.Error(err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"attachGitHubPr":{"id":"link-1","number":2016,"repoFullName":"heyblueteam/blue","inert":false}}}`))
	}))
	defer server.Close()

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("API_URL", server.URL)
	t.Setenv("AUTH_TOKEN", "token")
	t.Setenv("CLIENT_ID", "client")
	t.Setenv("COMPANY_ID", "company")

	prURL := "https://github.com/heyblueteam/blue/pull/2016"
	cmd := newLinkPRCmd()
	cmd.SetArgs([]string{"--record", "record-1", "--pr", prURL, "--workspace", "product", "--simple"})
	cmd.SilenceUsage = true
	var output bytes.Buffer
	cmd.SetOut(&output)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	input, ok := request.Variables["input"].(map[string]interface{})
	if !ok {
		t.Fatalf("input variable = %#v", request.Variables["input"])
	}
	if input["recordId"] != "record-1" || input["pr"] != prURL {
		t.Fatalf("unexpected input: %#v", input)
	}
	if strings.Contains(request.Query, prURL) {
		t.Fatal("PR URL was interpolated into the GraphQL query")
	}
	if got := output.String(); got != "GitHub link ID: link-1\n" {
		t.Fatalf("simple output = %q", got)
	}
}

func TestLinkPRDetailedOutput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":{"attachGitHubPr":{"id":"link-1","number":42,"repoFullName":"owner/repo","inert":false}}}`))
	}))
	defer server.Close()

	t.Setenv("XDG_CONFIG_HOME", t.TempDir())
	t.Setenv("API_URL", server.URL)
	t.Setenv("AUTH_TOKEN", "token")
	t.Setenv("CLIENT_ID", "client")
	t.Setenv("COMPANY_ID", "company")

	cmd := newLinkPRCmd()
	cmd.SetArgs([]string{"--record", "record-1", "--pr", "#42"})
	cmd.SilenceUsage = true
	var output bytes.Buffer
	cmd.SetOut(&output)
	if err := cmd.Execute(); err != nil {
		t.Fatal(err)
	}

	for _, want := range []string{
		"Pull request linked successfully",
		"Link ID: link-1",
		"Repository: owner/repo",
		"Pull request: #42",
	} {
		if !strings.Contains(output.String(), want) {
			t.Fatalf("output %q does not contain %q", output.String(), want)
		}
	}
}
