package open

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"strings"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

// Cmd opens Blue app pages from IDs or existing URLs.
var Cmd = &cobra.Command{
	Use:   "open [url]",
	Short: "Open Blue pages in a browser",
	Long:  "Build and open Blue app URLs for workspaces, records, forms, documents, dashboards, reports, and files.",
	Example: `  blue open https://blue.app/org/acme
  blue open workspace <id-or-slug>
  blue open record <id> --workspace <id-or-slug>
  blue open dashboard <id>
  blue open form <id> --workspace <id-or-slug> --print`,
	Args: cobra.MaximumNArgs(1),
	RunE: runOpenURL,
}

var (
	openWorkspace string
	openBaseURL   string
	openPrintOnly bool
)

type workspaceRef struct {
	ID   string `json:"id"`
	Slug string `json:"slug"`
}

func init() {
	Cmd.PersistentFlags().StringVarP(&openWorkspace, "workspace", "w", "", "Workspace ID or slug for workspace-scoped entities")
	Cmd.PersistentFlags().StringVar(&openBaseURL, "base-url", "https://blue.app", "Blue app base URL")
	Cmd.PersistentFlags().BoolVar(&openPrintOnly, "print", false, "Print the URL instead of opening it")

	Cmd.AddCommand(workspaceCmd)
	Cmd.AddCommand(recordCmd)
	Cmd.AddCommand(formCmd)
	Cmd.AddCommand(documentCmd)
	Cmd.AddCommand(dashboardCmd)
	Cmd.AddCommand(reportCmd)
	Cmd.AddCommand(filesCmd)
	Cmd.AddCommand(folderCmd)
}

func runOpenURL(cmd *cobra.Command, args []string) error {
	if len(args) == 0 {
		return cmd.Help()
	}
	input := args[0]
	if !isURL(input) {
		return fmt.Errorf("%q is not a URL. Use a typed command like 'blue open record <id> --workspace <id>'", input)
	}
	return outputURL(input)
}

func newClient() (*common.Client, error) {
	config, err := common.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	return common.NewClient(config), nil
}

func appURL(path string) (string, error) {
	base, err := url.Parse(strings.TrimRight(openBaseURL, "/"))
	if err != nil || base.Scheme == "" || base.Host == "" {
		return "", fmt.Errorf("invalid base-url %q", openBaseURL)
	}
	base.Path = strings.TrimRight(base.Path, "/") + "/" + strings.TrimLeft(path, "/")
	return base.String(), nil
}

func orgPath(client *common.Client, path string) (string, error) {
	return appURL("/org/" + url.PathEscape(client.GetCompanyID()) + "/" + strings.TrimLeft(path, "/"))
}

func workspacePath(client *common.Client, workspace, path string) (string, error) {
	ref, err := resolveWorkspace(client, workspace)
	if err != nil {
		return "", err
	}
	return orgPath(client, "workspace/"+url.PathEscape(ref.Slug)+"/"+strings.TrimLeft(path, "/"))
}

func resolveWorkspace(client *common.Client, workspace string) (workspaceRef, error) {
	if workspace == "" {
		return workspaceRef{}, fmt.Errorf("workspace ID or slug is required. Use --workspace flag")
	}
	client.SetProject(workspace)

	query := `query ResolveOpenWorkspace($workspace: String!) {
		project(id: $workspace) { id slug }
	}`
	variables := map[string]interface{}{"workspace": workspace}
	var response struct {
		Project workspaceRef `json:"project"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return workspaceRef{}, fmt.Errorf("failed to resolve workspace: %w", err)
	}
	if response.Project.ID == "" || response.Project.Slug == "" {
		return workspaceRef{}, fmt.Errorf("could not resolve workspace from %q", workspace)
	}
	return response.Project, nil
}

func outputURL(targetURL string) error {
	if openPrintOnly {
		fmt.Println(targetURL)
		return nil
	}
	return openBrowser(targetURL)
}

func openBrowser(targetURL string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", targetURL)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", targetURL)
	default:
		cmd = exec.Command("xdg-open", targetURL)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}
	return nil
}

func isURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme != "" && parsed.Host != ""
}
