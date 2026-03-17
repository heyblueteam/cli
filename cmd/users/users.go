package users

import (
	"github.com/spf13/cobra"
)

// Cmd is the parent command for user operations
var Cmd = &cobra.Command{
	Use:   "users",
	Short: "Manage users",
	Long:  "List user profiles, invite users, and manage project user roles.",
}

func init() {
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(inviteCmd)
	Cmd.AddCommand(rolesCmd)
}
