package workspaces

import (
	"strings"
	"testing"
)

func TestBuildEditFields_EscapesFreeTextFields(t *testing.T) {
	input := EditProjectInput{
		ProjectID:   "project-1",
		Name:        `Client "Acme" Onboarding`,
		Slug:        "acme-onboarding",
		Description: "line one\nline two",
		TodoAlias:   `Ticket "#"`,
	}

	got := buildEditFields(input)

	for _, want := range []string{
		`name: "Client \"Acme\" Onboarding"`,
		`description: "line one\nline two"`,
		`todoAlias: "Ticket \"#\""`,
	} {
		if !strings.Contains(got, want) {
			t.Errorf("buildEditFields() = %q, want it to contain %q", got, want)
		}
	}
}
