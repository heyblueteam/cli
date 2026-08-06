package doctor

import (
	"fmt"
	"os"
	"strings"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

// Cmd is the doctor command.
var Cmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check CLI configuration and API access",
	Long:  "Run non-destructive checks for local config, credentials, API connectivity, company context, and optional workspace access.",
	Example: `  blue doctor
  blue doctor --workspace <id-or-slug>`,
	RunE: runDoctor,
}

var doctorWorkspace string

func init() {
	Cmd.Flags().StringVarP(&doctorWorkspace, "workspace", "w", "", "Optional workspace ID or slug to verify")
}

func runDoctor(cmd *cobra.Command, args []string) error {
	failures := 0

	configPath := common.ConfigPath()
	if configPath == "" {
		failures++
		printFail("Config path", "could not determine config path")
	} else if _, err := os.Stat(configPath); err != nil {
		failures++
		printFail("Config file", fmt.Sprintf("%s (%v)", configPath, err))
	} else {
		printOK("Config file", configPath)
	}

	config, err := common.LoadConfig()
	if err != nil {
		failures++
		printFail("Load config", err.Error())
		return finish(failures)
	}
	printOK("API URL", config.APIUrl)
	printPresence("Client ID", config.ClientID, &failures)
	printPresence("Auth token", config.AuthToken, &failures)
	printPresence("Company", config.CompanyID, &failures)
	if config.DefaultWorkspace != "" {
		printOK("Default workspace", config.DefaultWorkspace)
	}

	client := common.NewClient(config)

	if _, err := client.ExecuteQuery(`query Doctor { __typename }`, nil); err != nil {
		failures++
		printFail("API connectivity", err.Error())
	} else {
		printOK("API connectivity", "authenticated GraphQL request succeeded")
	}

	if companyID, err := client.ResolveCompanyID(); err != nil {
		failures++
		printFail("Company access", err.Error())
	} else {
		printOK("Company access", fmt.Sprintf("resolved %s", companyID))
	}

	workspaceToCheck := doctorWorkspace
	if workspaceToCheck == "" {
		workspaceToCheck = config.DefaultWorkspace
	}
	if workspaceToCheck != "" {
		client.SetProject(workspaceToCheck)
		projectID, err := client.ResolveProjectID(workspaceToCheck)
		if err != nil {
			failures++
			printFail("Workspace access", err.Error())
		} else {
			printOK("Workspace access", fmt.Sprintf("resolved %s", projectID))
		}
	}

	return finish(failures)
}

func printPresence(label, value string, failures *int) {
	if strings.TrimSpace(value) == "" {
		*failures++
		printFail(label, "missing")
		return
	}
	printOK(label, redact(value))
}

func redact(value string) string {
	if len(value) <= 8 {
		return "present"
	}
	return value[:4] + "..." + value[len(value)-4:]
}

func printOK(label, message string) {
	fmt.Printf("✓ %-18s %s\n", label, message)
}

func printFail(label, message string) {
	fmt.Printf("✗ %-18s %s\n", label, message)
}

func finish(failures int) error {
	if failures > 0 {
		return fmt.Errorf("doctor found %d issue(s)", failures)
	}
	fmt.Println("\nAll checks passed.")
	return nil
}
