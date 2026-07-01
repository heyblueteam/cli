package workspaces

import (
	"strings"
	"testing"

	"github.com/heyblueteam/cli/common"
)

func TestBuildProjectOptionalFields_EscapesDescription(t *testing.T) {
	input := common.CreateProjectInput{
		Name:        "Sprint Planning",
		CompanyID:   "company-1",
		Description: "line one\nline two with a \"quote\" and a back\\slash",
	}

	got := buildProjectOptionalFields(input)

	want := `description: "line one\nline two with a \"quote\" and a back\\slash"`
	if !strings.Contains(got, want) {
		t.Errorf("buildProjectOptionalFields() = %q, want it to contain %q", got, want)
	}
}
