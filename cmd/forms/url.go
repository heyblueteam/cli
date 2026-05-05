package forms

import (
	"fmt"
	"os"
	"strings"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

const defaultFormsBaseURL = "https://blue.cc"

var urlCmd = &cobra.Command{
	Use:   "url",
	Short: "Print the public submit URL for a form",
	Long: `Print the public submit URL using the form's uid.

Defaults to https://blue.cc/forms/<uid>. Override the base via --base-url or
the BLUE_FORMS_BASE_URL environment variable for white-label / self-hosted
deployments.`,
	Example: `  blue forms url --form <id>
  blue forms url --form <id> --base-url https://forms.acme.com
  BLUE_FORMS_BASE_URL=https://forms.acme.com blue forms url --form <id>`,
	RunE: runURL,
}

var (
	urlForm      string
	urlWorkspace string
	urlBase      string
)

func init() {
	urlCmd.Flags().StringVarP(&urlForm, "form", "f", "", "Form ID (required)")
	urlCmd.Flags().StringVarP(&urlWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	urlCmd.Flags().StringVar(&urlBase, "base-url", "", "Base URL for forms (default: $BLUE_FORMS_BASE_URL or https://blue.cc)")
}

func runURL(cmd *cobra.Command, args []string) error {
	if urlForm == "" {
		return fmt.Errorf("form ID is required. Use --form flag")
	}
	if urlWorkspace == "" {
		return fmt.Errorf("workspace is required. Use --workspace flag")
	}

	base := urlBase
	if base == "" {
		base = os.Getenv("BLUE_FORMS_BASE_URL")
	}
	if base == "" {
		base = defaultFormsBaseURL
	}
	base = strings.TrimRight(base, "/")

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)
	client.SetProject(urlWorkspace)

	query := `
		query FormUID($id: String!) {
			form(id: $id) { uid }
		}
	`
	var resp struct {
		Form struct {
			UID string `json:"uid"`
		} `json:"form"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"id": urlForm}, &resp); err != nil {
		return fmt.Errorf("failed to read form: %w", err)
	}
	fmt.Printf("%s/forms/%s\n", base, resp.Form.UID)
	return nil
}
