package comments

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for comment operations
var Cmd = &cobra.Command{
	Use:   "comments",
	Short: "Manage comments",
	Long:  "Create, list, and update comments on records.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(createCmd)
	Cmd.AddCommand(updateCmd)
}
