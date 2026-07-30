package common

import (
	"testing"
)

// The chart worker reads a stored filter with Object.entries(), which throws on
// null, so "no filter" must serialize as an empty object rather than nothing.
func TestBuildTodoFilterIsNeverNil(t *testing.T) {
	filter, err := BuildTodoFilter(TodoFilterOptions{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filter == nil {
		t.Fatal("an empty filter must still be an object")
	}
	if len(filter) != 0 {
		t.Errorf("expected no keys, got %v", filter)
	}
}

func TestBuildTodoFilterScopesToProjects(t *testing.T) {
	filter, err := BuildTodoFilter(TodoFilterOptions{}, []string{"p1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ids, ok := filter["projectIds"].([]string)
	if !ok || len(ids) != 1 || ids[0] != "p1" {
		t.Errorf("expected the workspace scope, got %v", filter["projectIds"])
	}
}

func TestBuildTodoFilterOmitsUnsetBooleans(t *testing.T) {
	filter, err := BuildTodoFilter(TodoFilterOptions{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, key := range []string{"showCompleted", "archived", "unassigned"} {
		if _, found := filter[key]; found {
			t.Errorf("%s should be absent when the flag was not passed", key)
		}
	}
}

func TestBuildTodoFilterSplitsCSVFlags(t *testing.T) {
	filter, err := BuildTodoFilter(TodoFilterOptions{
		Assignees: "u1, u2",
		Tags:      "t1,,t2",
		Lists:     " l1 ",
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := filter["assigneeIds"].([]string); len(got) != 2 || got[1] != "u2" {
		t.Errorf("assignees not split/trimmed: %v", got)
	}
	if got := filter["tagIds"].([]string); len(got) != 2 {
		t.Errorf("empty entries should be dropped: %v", got)
	}
	if got := filter["todoListIds"].([]string); len(got) != 1 || got[0] != "l1" {
		t.Errorf("lists not trimmed: %v", got)
	}
}

// --filter-json is the escape hatch for anything the flags can't express, so it
// is merged last and overrides a flag that set the same key.
func TestBuildTodoFilterJSONOverridesFlags(t *testing.T) {
	filter, err := BuildTodoFilter(TodoFilterOptions{
		Query:      "from flag",
		FilterJSON: `{"q":"from json","fields":[{"type":"CUSTOM_FIELD"}]}`,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if filter["q"] != "from json" {
		t.Errorf("--filter-json should win, got %v", filter["q"])
	}
	if _, found := filter["fields"]; !found {
		t.Error("--filter-json should be able to add keys the flags don't cover")
	}
}

func TestBuildTodoFilterRejectsBadJSON(t *testing.T) {
	if _, err := BuildTodoFilter(TodoFilterOptions{FilterJSON: "{nope}"}, nil); err == nil {
		t.Fatal("expected an error for malformed --filter-json")
	}
}

func TestFormatNumber(t *testing.T) {
	tests := []struct {
		value float64
		want  string
	}{
		{value: 0, want: "0"},
		{value: 23, want: "23"},
		{value: 7710000, want: "7710000"},
		{value: 47.5, want: "47.5"},
		{value: -3, want: "-3"},
	}
	for _, test := range tests {
		if got := FormatNumber(test.value); got != test.want {
			t.Errorf("FormatNumber(%v) = %q, want %q", test.value, got, test.want)
		}
	}
}
