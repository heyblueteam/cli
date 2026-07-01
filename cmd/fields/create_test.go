package fields

import (
	"strings"
	"testing"
)

func TestBuildOptionalFields_EscapesFreeText(t *testing.T) {
	input := LocalCreateCustomFieldInput{
		Name:              `Priority`,
		Type:              "TEXT_SINGLE",
		Description:       "line one\nline two with a \"quote\"",
		ButtonConfirmText: `Are you sure you want to say "yes"?`,
	}

	got := buildOptionalFields(input)

	wantDescription := `description: "line one\nline two with a \"quote\""`
	if !strings.Contains(got, wantDescription) {
		t.Errorf("buildOptionalFields() = %q, want it to contain %q", got, wantDescription)
	}

	wantButtonConfirm := `buttonConfirmText: "Are you sure you want to say \"yes\"?"`
	if !strings.Contains(got, wantButtonConfirm) {
		t.Errorf("buildOptionalFields() = %q, want it to contain %q", got, wantButtonConfirm)
	}
}
