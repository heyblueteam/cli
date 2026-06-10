package api

import (
	"fmt"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

var docsCmd = &cobra.Command{
	Use:   "docs",
	Short: "Open Blue API documentation",
	Example: `  blue api docs
  blue api docs --print`,
	RunE: runDocs,
}

var docsPrint bool

const docsURL = "https://blue.cc/api"

func init() {
	docsCmd.Flags().BoolVar(&docsPrint, "print", false, "Print the docs URL instead of opening it")
}

func runDocs(cmd *cobra.Command, args []string) error {
	if docsPrint {
		fmt.Println(docsURL)
		return nil
	}

	var openCmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		openCmd = exec.Command("open", docsURL)
	case "windows":
		openCmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", docsURL)
	default:
		openCmd = exec.Command("xdg-open", docsURL)
	}
	if err := openCmd.Start(); err != nil {
		return fmt.Errorf("failed to open docs; visit %s: %w", docsURL, err)
	}
	fmt.Println(docsURL)
	return nil
}
