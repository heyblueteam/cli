package cmd

import (
	"fmt"
	"os"

	"blue-cli/cmd/automations"
	"blue-cli/cmd/charts"
	"blue-cli/cmd/checklists"
	"blue-cli/cmd/company"
	"blue-cli/cmd/comments"
	"blue-cli/cmd/dashboards"
	"blue-cli/cmd/dependencies"
	"blue-cli/cmd/fields"
	"blue-cli/cmd/files"
	"blue-cli/cmd/lists"
	"blue-cli/cmd/records"
	"blue-cli/cmd/tags"
	"blue-cli/cmd/users"
	"blue-cli/cmd/workspaces"

	"github.com/spf13/cobra"
)

// Version information (injected at build time via ldflags)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "blue",
	Short: "Blue CLI - Manage your Blue workspaces from the command line",
	Long: `Blue CLI is a command-line tool for interacting with the Blue API.
Manage workspaces, records, lists, tags, custom fields, automations, and more.`,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if companyFlag, _ := cmd.Flags().GetString("company"); companyFlag != "" {
			os.Setenv("COMPANY_ID", companyFlag)
		}
	},
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().String("company", "", "Override the active company for this command")

	// Add subcommand groups
	rootCmd.AddCommand(automations.Cmd)
	rootCmd.AddCommand(charts.Cmd)
	rootCmd.AddCommand(checklists.Cmd)
	rootCmd.AddCommand(company.Cmd)
	rootCmd.AddCommand(dashboards.Cmd)
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
			fmt.Printf("blue %s (commit: %s, built: %s)\n", version, commit, date)
		},
	})
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
