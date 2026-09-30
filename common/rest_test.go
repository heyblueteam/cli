package common

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func restTestClient(t *testing.T, serverURL string) *Client {
	t.Helper()
	return NewClient(&Config{
		APIUrl:    serverURL,
		AuthToken: "secret",
		ClientID:  "client",
		CompanyID: "company",
	})
}

func TestRESTBaseURLStripsGraphQLSuffix(t *testing.T) {
	client := NewClient(&Config{APIUrl: "https://api.blue.app/graphql"})
	if got := client.restBaseURL(); got != "https://api.blue.app/v1" {
		t.Fatalf("restBaseURL() = %q, want https://api.blue.app/v1", got)
	}
	client = NewClient(&Config{APIUrl: "https://api.blue.app/"})
	if got := client.restBaseURL(); got != "https://api.blue.app/v1" {
		t.Fatalf("restBaseURL() = %q, want https://api.blue.app/v1", got)
	}
	client = NewClient(&Config{APIUrl: "https://api.blue.app"})
	if got := client.restBaseURL(); got != "https://api.blue.app/v1" {
		t.Fatalf("restBaseURL() = %q, want https://api.blue.app/v1", got)
	}
}

func TestRESTSendsAuthAndCompanyHeaders(t *testing.T) {
	var gotPath, gotMethod string
	var gotTokenID, gotTokenSecret, gotCompany string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotMethod = r.Method
		gotTokenID = r.Header.Get("X-Bloo-Token-ID")
		gotTokenSecret = r.Header.Get("X-Bloo-Token-Secret")
		gotCompany = r.Header.Get("X-Bloo-Company-ID")
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"t_1","name":"Deals"}`))
	}))
	defer server.Close()

	client := restTestClient(t, server.URL)
	var out struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
	if err := client.REST("GET", "/databases", url.Values{"workspaceId": {"w1"}}, nil, &out); err != nil {
		t.Fatal(err)
	}
	if gotPath != "/v1/databases" || gotMethod != "GET" {
		t.Fatalf("request was %s %s, want GET /v1/databases", gotMethod, gotPath)
	}
	if gotTokenID != "client" || gotTokenSecret != "secret" || gotCompany != "company" {
		t.Fatalf("missing auth headers: id=%q secret=%q company=%q", gotTokenID, gotTokenSecret, gotCompany)
	}
	if out.ID != "t_1" || out.Name != "Deals" {
		t.Fatalf("bad decode: %+v", out)
	}
}

func TestRESTPostBodyAndContentType(t *testing.T) {
	var gotBody map[string]interface{}
	var gotContentType string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotContentType = r.Header.Get("Content-Type")
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Error(err)
		}
		w.Write([]byte(`{"fieldId":"f_1","renamed":true}`))
	}))
	defer server.Close()

	client := restTestClient(t, server.URL)
	var out struct {
		FieldID string `json:"fieldId"`
		Renamed bool   `json:"renamed"`
	}
	if err := client.REST("POST", "/rename-database-field", nil, map[string]interface{}{
		"workspaceId": "w1",
		"fieldId":     "f_1",
		"name":        "New name",
	}, &out); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(gotContentType, "application/json") {
		t.Fatalf("Content-Type = %q, want application/json", gotContentType)
	}
	if gotBody["name"] != "New name" {
		t.Fatalf("body = %+v", gotBody)
	}
	if out.Renamed != true {
		t.Fatalf("out = %+v", out)
	}
}

func TestRESTErrorParsing(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`{"error":"Invalid input","issues":[{"path":["first"],"message":"must be between 1 and 200"}]}`))
	}))
	defer server.Close()

	client := restTestClient(t, server.URL)
	err := client.REST("POST", "/list-database-records", nil, map[string]interface{}{"first": 0}, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	msg := err.Error()
	if !strings.Contains(msg, "422") || !strings.Contains(msg, "Invalid input") || !strings.Contains(msg, "first: must be between 1 and 200") {
		t.Fatalf("error message incomplete: %q", msg)
	}
	var restErr *RESTError
	if !errors.As(err, &restErr) || restErr.Status != 422 {
		t.Fatalf("expected *RESTError with status 422, got %v", err)
	}
}

func TestRESTPlainErrorMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error":"Unauthorized"}`))
	}))
	defer server.Close()

	client := restTestClient(t, server.URL)
	err := client.REST("GET", "/databases", nil, nil, nil)
	if err == nil || !strings.Contains(err.Error(), "Unauthorized") {
		t.Fatalf("expected Unauthorized in error, got %v", err)
	}
}
