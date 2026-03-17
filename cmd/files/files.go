package files

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for file operations
var Cmd = &cobra.Command{
	Use:   "files",
	Short: "Manage files",
	Long:  "Download and manage files from workspaces.",
}

func init() {
	Cmd.AddCommand(downloadCmd)
}
