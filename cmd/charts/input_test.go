package charts

import (
	"strings"
	"testing"
)

func TestBuildStatFormulaSingleSourceReferencesItsValue(t *testing.T) {
	inFormula = ""
	sources := []statSource{{Title: "Total"}}
	uids := map[string]string{"Total": "uid-total"}

	got, err := buildStatFormula(sources, uids)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := `{"chartSegmentValueUID":"uid-total"}`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestBuildStatFormulaSeveralSourcesRequireAFormula(t *testing.T) {
	inFormula = ""
	sources := []statSource{{Title: "Won"}, {Title: "Total"}}
	uids := map[string]string{"Won": "uid-won", "Total": "uid-total"}

	if _, err := buildStatFormula(sources, uids); err == nil {
		t.Fatal("expected an error when several sources have no formula")
	}
}

// A title that is a prefix of another must not shadow it — replacing "Won"
// first would corrupt "Won Deals" into a reference followed by " Deals".
func TestBuildStatFormulaPrefersLongerTitles(t *testing.T) {
	inFormula = "Won Deals / Won * 100"
	sources := []statSource{{Title: "Won"}, {Title: "Won Deals"}}
	uids := map[string]string{"Won": "uid-won", "Won Deals": "uid-won-deals"}

	got, err := buildStatFormula(sources, uids)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(got, " Deals") {
		t.Errorf("longer title was shadowed by the shorter one: %q", got)
	}
	if !strings.Contains(got, "uid-won-deals") || !strings.Contains(got, "uid-won\"") {
		t.Errorf("both sources should be referenced, got %q", got)
	}
}

func TestBuildStatFormulaRejectsUnknownNames(t *testing.T) {
	inFormula = "Alpha / Beta"
	sources := []statSource{{Title: "Won"}, {Title: "Total"}}
	uids := map[string]string{"Won": "uid-won", "Total": "uid-total"}

	if _, err := buildStatFormula(sources, uids); err == nil {
		t.Fatal("expected an error when the formula names no source")
	}
}

func TestParseBands(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		atRisk  float64
		onTrack float64
		wantErr bool
	}{
		{name: "ordered pair", input: "0.5,0.9", atRisk: 0.5, onTrack: 0.9},
		{name: "spaces tolerated", input: " 0.4 , 0.8 ", atRisk: 0.4, onTrack: 0.8},
		{name: "one value", input: "0.5", wantErr: true},
		{name: "three values", input: "0.1,0.5,0.9", wantErr: true},
		{name: "not a number", input: "half,0.9", wantErr: true},
		{name: "negative", input: "-0.5,0.9", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			atRisk, onTrack, err := parseBands(test.input)
			if test.wantErr {
				if err == nil {
					t.Fatalf("expected an error for %q", test.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if atRisk != test.atRisk || onTrack != test.onTrack {
				t.Errorf("got (%v, %v), want (%v, %v)", atRisk, onTrack, test.atRisk, test.onTrack)
			}
		})
	}
}

func TestIsDateGroupBy(t *testing.T) {
	for _, groupBy := range []string{"TODO_DUE_DATE", "TODO_CREATED_AT", "TODO_UPDATED_AT", "TODO_COMPLETED_AT"} {
		if !isDateGroupBy(groupBy) {
			t.Errorf("%s should take an interval", groupBy)
		}
	}
	for _, groupBy := range []string{"ASSIGNEE", "TAG", "TODO_LIST", "TODO_STATUS", "PROJECT", "CUSTOM_FIELD"} {
		if isDateGroupBy(groupBy) {
			t.Errorf("%s should not take an interval", groupBy)
		}
	}
}

// Metadata read back from the API carries __typename keys the input types
// reject, and a null for every unset filter field. Both must go, but a false
// or a zero is a real value.
func TestStripTypenamesRemovesTypenamesAndNulls(t *testing.T) {
	metadata := map[string]interface{}{
		"__typename":  "ChartMetadataBarChart",
		"renderStyle": "LINE",
		"yAxis": map[string]interface{}{
			"__typename": "ChartMetadataBarChartYAxis",
			"title":      "Value",
			"function":   nil,
			"filter": map[string]interface{}{
				"projectIds":    []interface{}{"p1"},
				"showCompleted": false,
				"q":             nil,
			},
		},
	}

	stripTypenames(metadata)

	if _, found := metadata["__typename"]; found {
		t.Error("__typename should be removed")
	}
	yAxis := metadata["yAxis"].(map[string]interface{})
	if _, found := yAxis["__typename"]; found {
		t.Error("nested __typename should be removed")
	}
	if _, found := yAxis["function"]; found {
		t.Error("null values should be removed")
	}
	if yAxis["title"] != "Value" {
		t.Error("set values should be kept")
	}

	filter := yAxis["filter"].(map[string]interface{})
	if _, found := filter["q"]; found {
		t.Error("null filter fields should be removed")
	}
	if showCompleted, found := filter["showCompleted"]; !found || showCompleted != false {
		t.Error("a false is a real value and must be kept")
	}
	if _, found := filter["projectIds"]; !found {
		t.Error("the workspace scope must survive")
	}
}
