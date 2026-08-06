package charts

import "testing"

func TestEditRequiresCompleteValidSize(t *testing.T) {
	cmd := newEditCommand()
	cmd.SetArgs([]string{"--chart", "chart-1", "--width", "4"})
	cmd.SilenceUsage = true
	if err := cmd.Execute(); err == nil {
		t.Fatal("half a card size was accepted")
	}

	cmd = newEditCommand()
	cmd.SetArgs([]string{"--chart", "chart-1", "--width", "5", "--height", "2"})
	cmd.SilenceUsage = true
	if err := cmd.Execute(); err == nil {
		t.Fatal("oversized card was accepted")
	}
}

func TestEditFormattingFlagsAreExplicit(t *testing.T) {
	cmd := newEditCommand()
	cmd.SetArgs([]string{"--chart", "chart-1", "--precision", "2"})
	cmd.SilenceUsage = true
	if err := cmd.Execute(); err == nil {
		t.Fatal("precision without display was accepted")
	}
}
