package context

import (
	"fmt"
	"os"
	"strings"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "context",
	Short: "Manage default company and workspace context",
	Long:  "Manage the default company and workspace used by workspace-scoped commands.",
	Example: `  blue context current
  blue context list
  blue context use acme
  blue context use acme/development
  blue context set-workspace development
  blue context clear`,
}

var currentCmd = &cobra.Command{
	Use:   "current",
	Short: "Show active default context",
	RunE: func(cmd *cobra.Command, args []string) error {
		company, workspace, err := readContext()
		if err != nil {
			return err
		}
		fmt.Println("Current context")
		fmt.Printf("  Company:   %s\n", valueOrNone(company))
		fmt.Printf("  Workspace: %s\n", valueOrNone(workspace))
		fmt.Printf("  Config:    %s\n", common.ConfigPath())
		return nil
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List known companies and active defaults",
	RunE: func(cmd *cobra.Command, args []string) error {
		company, workspace, err := readContext()
		if err != nil {
			return err
		}
		companies, err := common.GetCompanies()
		if err != nil {
			return err
		}
		fmt.Println("Known companies")
		if len(companies) == 0 {
			fmt.Println("  none")
		} else {
			for _, item := range companies {
				marker := " "
				if item == company {
					marker = "*"
				}
				fmt.Printf("  %s %s\n", marker, item)
			}
		}
		fmt.Println("\nDefaults")
		fmt.Printf("  Company:   %s\n", valueOrNone(company))
		fmt.Printf("  Workspace: %s\n", valueOrNone(workspace))
		return nil
	},
}

var useCmd = &cobra.Command{
	Use:   "use <company>[/<workspace>]",
	Short: "Set default company and optionally workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		company, workspace, hasWorkspace := strings.Cut(strings.TrimSpace(args[0]), "/")
		company = strings.TrimSpace(company)
		workspace = strings.TrimSpace(workspace)
		if company == "" {
			return fmt.Errorf("company is required")
		}
		if err := common.SetActiveCompany(company); err != nil {
			return err
		}
		os.Setenv("COMPANY_ID", company)
		if err := common.AddCompany(company); err != nil {
			return err
		}
		if !hasWorkspace {
			if err := common.ClearDefaultWorkspace(); err != nil {
				return err
			}
			fmt.Printf("Using company %s. Default workspace cleared.\n", company)
			return nil
		}
		if workspace == "" {
			return fmt.Errorf("workspace is required after slash")
		}
		resolved, err := resolveWorkspace(workspace)
		if err != nil {
			return err
		}
		if err := common.SetDefaultWorkspace(resolved); err != nil {
			return err
		}
		fmt.Printf("Using company %s and workspace %s.\n", company, resolved)
		return nil
	},
}

var setWorkspaceCmd = &cobra.Command{
	Use:   "set-workspace <workspace-id-or-slug>",
	Short: "Set default workspace for the active company",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		resolved, err := resolveWorkspace(args[0])
		if err != nil {
			return err
		}
		if err := common.SetDefaultWorkspace(resolved); err != nil {
			return err
		}
		fmt.Printf("Default workspace set to %s.\n", resolved)
		return nil
	},
}

var clearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear default workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := common.ClearDefaultWorkspace(); err != nil {
			return err
		}
		fmt.Println("Default workspace cleared.")
		return nil
	},
}

func init() {
	Cmd.AddCommand(currentCmd)
	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(useCmd)
	Cmd.AddCommand(setWorkspaceCmd)
	Cmd.AddCommand(clearCmd)
}

func readContext() (string, string, error) {
	envMap, err := common.ReadConfigFile()
	if err != nil {
		return "", "", err
	}
	return envMap["COMPANY_ID"], envMap["DEFAULT_WORKSPACE_ID"], nil
}

func resolveWorkspace(workspace string) (string, error) {
	config, err := common.LoadConfig()
	if err != nil {
		return "", fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)
	client.SetProject(workspace)
	projectID, err := client.ResolveProjectID(workspace)
	if err != nil {
		return "", fmt.Errorf("failed to resolve workspace %q: %w", workspace, err)
	}
	return projectID, nil
}

func valueOrNone(value string) string {
	if strings.TrimSpace(value) == "" {
		return "none"
	}
	return value
}
