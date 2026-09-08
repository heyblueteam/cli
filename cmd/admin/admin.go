package admin

import "github.com/spf13/cobra"

// Cmd is the parent command for platform administration operations.
var Cmd = &cobra.Command{
	Use:   "admin",
	Short: "Manage Blue platform settings",
}

func init() {
	Cmd.AddCommand(noticeCmd)
}
