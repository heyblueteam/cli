package company

import "github.com/spf13/cobra"

// Cmd is the parent command for company operations
var Cmd = &cobra.Command{
	Use:   "company",
	Short: "Manage known companies",
	Long:  "Add, remove, list, and switch between company accounts.",
}

func init() {
	Cmd.AddCommand(addCmd)
	Cmd.AddCommand(useCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(removeCmd)
}
