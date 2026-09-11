package lists

import (
	"errors"
	"strings"
	"testing"
)

type recordedListCreate struct {
	query     string
	variables map[string]interface{}
}

type recordingListClient struct {
	calls  []recordedListCreate
	failOn map[string]error
}

func (c *recordingListClient) ExecuteQueryWithResult(
	query string,
	variables map[string]interface{},
	result interface{},
) error {
	input := variables["input"].(map[string]interface{})
	c.calls = append(c.calls, recordedListCreate{query: query, variables: variables})

	if err := c.failOn[input["title"].(string)]; err != nil {
		return err
	}

	response := result.(*CreateTodoListResponse)
	title := input["title"].(string)
	response.CreateTodoList.ID = "list-" + title
	response.CreateTodoList.Title = title
	return nil
}

func TestCreateListsUsesSupportedPlacementAndRequestedOrder(t *testing.T) {
	client := &recordingListClient{}
	created, failures := createLists(client, "workspace-1", []string{"A", "B", "C"}, nil)
	if created != 3 {
		t.Fatalf("created = %d, want 3", created)
	}
	if len(failures) != 0 {
		t.Fatalf("failures = %#v, want none", failures)
	}

	wantNames := []string{"A", "B", "C"}
	for i, wantName := range wantNames {
		call := client.calls[i]
		if strings.Contains(call.query, "position") {
			t.Fatalf("call %d query contains unsupported position field: %s", i, call.query)
		}

		input := call.variables["input"].(map[string]interface{})
		if _, ok := input["position"]; ok {
			t.Fatalf("call %d input contains unsupported position field: %#v", i, input)
		}
		if got := input["title"]; got != wantName {
			t.Fatalf("call %d title = %v, want %q", i, got, wantName)
		}
		if got := input["projectId"]; got != "workspace-1" {
			t.Fatalf("call %d projectId = %v, want workspace-1", i, got)
		}

		if i == 0 {
			if _, ok := input["previousId"]; ok {
				t.Fatalf("first call unexpectedly has previousId: %#v", input)
			}
			continue
		}
		if got := input["previousId"]; got != "list-"+wantNames[i-1] {
			t.Fatalf("call %d previousId = %v, want list-%s", i, got, wantNames[i-1])
		}
	}
}

func TestCreateListsReportsFailuresAndContinuesOrderChain(t *testing.T) {
	client := &recordingListClient{failOn: map[string]error{"B": errors.New("schema rejected input")}}
	created, failures := createLists(client, "workspace-1", []string{"A", "B", "C"}, nil)
	if created != 2 {
		t.Fatalf("created = %d, want 2", created)
	}
	if len(failures) != 1 || failures[0].name != "B" || failures[0].err.Error() != "schema rejected input" {
		t.Fatalf("failures = %#v, want B/schema rejected input", failures)
	}

	input := client.calls[2].variables["input"].(map[string]interface{})
	if got := input["title"]; got != "C" {
		t.Fatalf("third call title = %v, want C", got)
	}
	if got := input["previousId"]; got != "list-A" {
		t.Fatalf("third call previousId = %v, want list-A", got)
	}
}
