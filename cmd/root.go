package cmd

import (
	"fmt"
	"os"
	"runtime/debug"

	"github.com/heyblueteam/cli/cmd/automations"
	"github.com/heyblueteam/cli/cmd/charts"
	"github.com/heyblueteam/cli/cmd/checklists"
	"github.com/heyblueteam/cli/cmd/comments"
	"github.com/heyblueteam/cli/cmd/company"
	"github.com/heyblueteam/cli/cmd/dashboards"
	"github.com/heyblueteam/cli/cmd/dependencies"
	"github.com/heyblueteam/cli/cmd/exports"
	"github.com/heyblueteam/cli/cmd/fields"
	"github.com/heyblueteam/cli/cmd/files"
	"github.com/heyblueteam/cli/cmd/forms"
	"github.com/heyblueteam/cli/cmd/lists"
	"github.com/heyblueteam/cli/cmd/records"
	"github.com/heyblueteam/cli/cmd/tags"
	"github.com/heyblueteam/cli/cmd/users"
	"github.com/heyblueteam/cli/cmd/webhooks"
	"github.com/heyblueteam/cli/cmd/workspaces"

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
	rootCmd.AddCommand(exports.Cmd)
	rootCmd.AddCommand(fields.Cmd)
	rootCmd.AddCommand(files.Cmd)
	rootCmd.AddCommand(forms.Cmd)
	rootCmd.AddCommand(webhooks.Cmd)

	// Version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			v, c, d := resolveVersion()
			fmt.Printf("blue %s (commit: %s, built: %s)\n", v, c, d)
		},
	})
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// resolveVersion returns (version, commit, date), preferring values injected
// by GoReleaser via ldflags. For `go install` builds (where ldflags aren't
// set), it falls back to the module version and VCS metadata that Go embeds
// in the binary.
func resolveVersion() (string, string, string) {
	v, c, d := version, commit, date

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return v, c, d
	}

	if v == "dev" && info.Main.Version != "" && info.Main.Version != "(devel)" {
		v = info.Main.Version
	}

	for _, s := range info.Settings {
		switch s.Key {
		case "vcs.revision":
			if c == "none" && s.Value != "" {
				c = s.Value
				if len(c) > 7 {
					c = c[:7]
				}
			}
		case "vcs.time":
			if d == "unknown" && s.Value != "" {
				d = s.Value
			}
		}
	}

	return v, c, d
}
