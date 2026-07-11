package cmd

import (
	"fmt"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"

	"github.com/heyblueteam/cli/cmd/activity"
	"github.com/heyblueteam/cli/cmd/api"
	"github.com/heyblueteam/cli/cmd/automations"
	"github.com/heyblueteam/cli/cmd/bootstrap"
	"github.com/heyblueteam/cli/cmd/charts"
	"github.com/heyblueteam/cli/cmd/checklists"
	"github.com/heyblueteam/cli/cmd/comments"
	"github.com/heyblueteam/cli/cmd/company"
	"github.com/heyblueteam/cli/cmd/dashboards"
	"github.com/heyblueteam/cli/cmd/dependencies"
	"github.com/heyblueteam/cli/cmd/docs"
	"github.com/heyblueteam/cli/cmd/doctor"
	"github.com/heyblueteam/cli/cmd/documents"
	"github.com/heyblueteam/cli/cmd/domains"
	"github.com/heyblueteam/cli/cmd/exports"
	"github.com/heyblueteam/cli/cmd/fields"
	"github.com/heyblueteam/cli/cmd/files"
	"github.com/heyblueteam/cli/cmd/forms"
	"github.com/heyblueteam/cli/cmd/ids"
	"github.com/heyblueteam/cli/cmd/lists"
	"github.com/heyblueteam/cli/cmd/open"
	"github.com/heyblueteam/cli/cmd/records"
	"github.com/heyblueteam/cli/cmd/reports"
	"github.com/heyblueteam/cli/cmd/savedviews"
	"github.com/heyblueteam/cli/cmd/search"
	"github.com/heyblueteam/cli/cmd/tags"
	"github.com/heyblueteam/cli/cmd/users"
	"github.com/heyblueteam/cli/cmd/webhooks"
	"github.com/heyblueteam/cli/cmd/whoami"
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
	rootCmd.AddCommand(api.Cmd)
	rootCmd.AddCommand(activity.Cmd)
	rootCmd.AddCommand(bootstrap.Cmd)
	rootCmd.AddCommand(charts.Cmd)
	rootCmd.AddCommand(checklists.Cmd)
	rootCmd.AddCommand(company.Cmd)
	rootCmd.AddCommand(dashboards.Cmd)
	rootCmd.AddCommand(workspaces.Cmd)
	rootCmd.AddCommand(lists.Cmd)
	rootCmd.AddCommand(open.Cmd)
	rootCmd.AddCommand(records.Cmd)
	rootCmd.AddCommand(reports.Cmd)
	rootCmd.AddCommand(savedviews.Cmd)
	rootCmd.AddCommand(search.Cmd)
	rootCmd.AddCommand(tags.Cmd)
	rootCmd.AddCommand(comments.Cmd)
	rootCmd.AddCommand(users.Cmd)
	rootCmd.AddCommand(dependencies.Cmd)
	rootCmd.AddCommand(documents.Cmd)
	rootCmd.AddCommand(doctor.Cmd)
	rootCmd.AddCommand(docs.Cmd)
	rootCmd.AddCommand(domains.Cmd)
	rootCmd.AddCommand(exports.Cmd)
	rootCmd.AddCommand(fields.Cmd)
	rootCmd.AddCommand(files.Cmd)
	rootCmd.AddCommand(forms.Cmd)
	rootCmd.AddCommand(ids.Cmd)
	rootCmd.AddCommand(webhooks.Cmd)
	rootCmd.AddCommand(whoami.Cmd)

	// Version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			v, c, d := resolveVersion()
			fmt.Printf("blue %s (commit: %s, built: %s)\n", v, c, d)
		},
	})

	// Append the defining source file to every command's errors, so an agent
	// hitting an error can go straight to the implementation instead of
	// guessing from the command name.
	addSourceHints(rootCmd)
}

// addSourceHints walks the command tree and wraps each RunE so a failure's
// error message is suffixed with the file it came from. The location is
// read from the compiled binary's function metadata (runtime.FuncForPC),
// not guessed from the command name, so it's accurate even where a file
// holds multiple subcommands (e.g. cmd/docs/docs.go) or a command's Use
// differs from its filename — and it survives release builds (-ldflags
// "-s -w" strips symbols/DWARF but not the pcln table this relies on).
func addSourceHints(c *cobra.Command) {
	for _, sub := range c.Commands() {
		addSourceHints(sub)

		if sub.RunE == nil {
			continue
		}
		hint := sourceHint(sub.RunE)
		original := sub.RunE
		sub.RunE = func(cmd *cobra.Command, args []string) error {
			err := original(cmd, args)
			if err != nil && hint != "" {
				return fmt.Errorf("%w (see %s)", err, hint)
			}
			return err
		}
	}
}

// sourceHint returns the repo-relative path (e.g. "cmd/records/list.go")
// where fn is defined, or "" if it can't be determined.
func sourceHint(fn func(*cobra.Command, []string) error) string {
	pc := reflect.ValueOf(fn).Pointer()
	f := runtime.FuncForPC(pc)
	if f == nil {
		return ""
	}
	file, _ := f.FileLine(pc)
	if idx := strings.LastIndex(file, "/cmd/"); idx != -1 {
		return file[idx+1:]
	}
	return ""
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
