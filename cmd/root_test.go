package cmd

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/spf13/cobra"
)

// TestSourceHintsResolveToRealFiles walks the full command tree and checks
// that every RunE-bearing command produced a non-empty source hint pointing
// at a file that actually exists. This is the safety net for addSourceHints:
// if a future Go toolchain change ever broke runtime.FuncForPC's file/line
// resolution, this test — not a silently wrong error message — would catch it.
func TestSourceHintsResolveToRealFiles(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine test file location")
	}
	cmdDir := filepath.Dir(thisFile) // .../cli/cmd
	repoRoot := filepath.Dir(cmdDir) // .../cli

	var checked int
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		for _, sub := range c.Commands() {
			walk(sub)
			if sub.RunE == nil {
				continue
			}
			hint := sourceHint(sub.RunE)
			if hint == "" {
				t.Errorf("%s: no source hint resolved", sub.CommandPath())
				continue
			}
			checked++
			if _, err := os.Stat(filepath.Join(repoRoot, hint)); err != nil {
				t.Errorf("%s: hint %q does not exist: %v", sub.CommandPath(), hint, err)
			}
		}
	}
	walk(rootCmd)

	if checked == 0 {
		t.Fatal("no RunE commands were checked — test is not exercising anything")
	}
}
