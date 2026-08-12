package common

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestFetchFieldTypesPaginatesUntilRequestedFieldsAreFound(t *testing.T) {
	var skips []int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request GraphQLRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		skipValue, skipOK := request.Variables["skip"].(float64)
		takeValue, takeOK := request.Variables["take"].(float64)
		if !skipOK || !takeOK {
			http.Error(w, "skip and take variables are required", http.StatusBadRequest)
			return
		}
		skip := int(skipValue)
		take := int(takeValue)
		skips = append(skips, skip)
		if take != customFieldTypePageSize {
			http.Error(w, fmt.Sprintf("take = %d, want %d", take, customFieldTypePageSize), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if skip == 0 {
			_, _ = fmt.Fprint(w, `{"data":{"customFields":{"items":[{"id":"field-early","type":"TEXT_SINGLE"}],"pageInfo":{"hasNextPage":true}}}}`)
			return
		}
		_, _ = fmt.Fprint(w, `{"data":{"customFields":{"items":[{"id":"field-date","type":"DATE"},{"id":"field-checkbox","type":"CHECKBOX"}],"pageInfo":{"hasNextPage":false}}}}`)
	}))
	defer server.Close()

	client := NewClient(&Config{
		APIUrl:    server.URL,
		AuthToken: "token",
		ClientID:  "client",
		CompanyID: "company",
	})
	client.SetProject("workspace")

	got, err := FetchFieldTypes(client, []string{"field-date", "field-checkbox"})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{"field-date": "DATE", "field-checkbox": "CHECKBOX"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("FetchFieldTypes() = %#v, want %#v", got, want)
	}
	if !reflect.DeepEqual(skips, []int{0, customFieldTypePageSize}) {
		t.Fatalf("page skips = %v, want [0 %d]", skips, customFieldTypePageSize)
	}
}

func TestFetchFieldTypesRejectsFieldsOutsideWorkspace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"data":{"customFields":{"items":[],"pageInfo":{"hasNextPage":false}}}}`)
	}))
	defer server.Close()

	client := NewClient(&Config{
		APIUrl:    server.URL,
		AuthToken: "token",
		ClientID:  "client",
		CompanyID: "company",
	})
	client.SetProject("workspace")

	_, err := FetchFieldTypes(client, []string{"field-missing"})
	if err == nil || !strings.Contains(err.Error(), "custom field(s) not found in workspace: field-missing") {
		t.Fatalf("FetchFieldTypes() error = %v, want missing-field error", err)
	}
}
