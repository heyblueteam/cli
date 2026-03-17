package cmd

import (
	"fmt"
	"os"

	"demo-builder/cmd/automations"
	"demo-builder/cmd/checklists"
	"demo-builder/cmd/comments"
	"demo-builder/cmd/dependencies"
	"demo-builder/cmd/fields"
	"demo-builder/cmd/files"
	"demo-builder/cmd/lists"
	"demo-builder/cmd/records"
	"demo-builder/cmd/tags"
	"demo-builder/cmd/users"
	"demo-builder/cmd/workspaces"

	"github.com/spf13/cobra"
)

// Version information (injected at build time via ldflags)
var (
	version = "dev"
	commit  = "none"
)

var rootCmd = &cobra.Command{
	Use:   "blue",
	Short: "Blue CLI - Manage your Blue workspaces from the command line",
	Long: `Blue CLI is a command-line tool for interacting with the Blue API.
Manage workspaces, records, lists, tags, custom fields, automations, and more.`,
}

func init() {
	// Add subcommand groups
	rootCmd.AddCommand(automations.Cmd)
	rootCmd.AddCommand(checklists.Cmd)
	rootCmd.AddCommand(workspaces.Cmd)
	rootCmd.AddCommand(lists.Cmd)
	rootCmd.AddCommand(records.Cmd)
	rootCmd.AddCommand(tags.Cmd)
	rootCmd.AddCommand(comments.Cmd)
	rootCmd.AddCommand(users.Cmd)
	rootCmd.AddCommand(dependencies.Cmd)
	rootCmd.AddCommand(fields.Cmd)
	rootCmd.AddCommand(files.Cmd)

	// Version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("blue %s (commit: %s)\n", version, commit)
		},
	})
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
