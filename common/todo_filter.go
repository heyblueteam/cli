package common

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

// TodoFilterOptions are the record-filter flags a command exposes. They cover
// the common cases; anything the flags can't express (per-field conditions,
// nested groups) goes through FilterJSON, which is merged in last and wins.
type TodoFilterOptions struct {
	ShowCompleted *bool
	Archived      *bool
	Unassigned    *bool
	Assignees     string
	Tags          string
	Lists         string
	DueStart      string
	DueEnd        string
	Query         string
	FilterJSON    string
}

// BuildTodoFilter turns the flags into a TodoFilterInput map.
//
// Always returns a non-nil map, even when nothing is set: the chart worker
// reads a segment's stored filter with Object.entries(), which throws on null,
// so an empty object is the correct "no filter" value rather than an omission.
//
// projectIDs scopes the filter to the chart's workspace. Callers pass the
// resolved workspace ID; a segment value is scoped to one project anyway, but
// the auto-chart axes need it stated explicitly.
func BuildTodoFilter(opts TodoFilterOptions, projectIDs []string) (map[string]interface{}, error) {
	filter := map[string]interface{}{}

	if len(projectIDs) > 0 {
		filter["projectIds"] = projectIDs
	}
	if opts.ShowCompleted != nil {
		filter["showCompleted"] = *opts.ShowCompleted
	}
	if opts.Archived != nil {
		filter["archived"] = *opts.Archived
	}
	if opts.Unassigned != nil {
		filter["unassigned"] = *opts.Unassigned
	}
	if ids := SplitCSV(opts.Assignees); len(ids) > 0 {
		filter["assigneeIds"] = ids
	}
	if ids := SplitCSV(opts.Tags); len(ids) > 0 {
		filter["tagIds"] = ids
	}
	if ids := SplitCSV(opts.Lists); len(ids) > 0 {
		filter["todoListIds"] = ids
	}
	if opts.DueStart != "" {
		filter["dueStart"] = opts.DueStart
	}
	if opts.DueEnd != "" {
		filter["dueEnd"] = opts.DueEnd
	}
	if opts.Query != "" {
		filter["q"] = opts.Query
	}

	if strings.TrimSpace(opts.FilterJSON) != "" {
		var raw map[string]interface{}
		if err := json.Unmarshal([]byte(opts.FilterJSON), &raw); err != nil {
			return nil, fmt.Errorf("invalid JSON for --filter-json: %w", err)
		}
		// Last write wins, key by key — so --filter-json can both add keys the
		// flags don't cover and override one a flag set.
		for key, value := range raw {
			filter[key] = value
		}
	}

	return filter, nil
}

// FormatNumber renders a chart value for display: whole numbers without a
// trailing ".0", fractions with the digits they need and no padding.
func FormatNumber(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64)
}

// SplitCSV splits a comma-separated flag value, trimming blanks.
func SplitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	var items []string
	for _, part := range strings.Split(value, ",") {
		if part = strings.TrimSpace(part); part != "" {
			items = append(items, part)
		}
	}
	return items
}
