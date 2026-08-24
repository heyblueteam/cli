package records

import (
	"fmt"
	"strings"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

type linkPROptions struct {
	recordID  string
	pr        string
	workspace string
	simple    bool
}

type gitHubEntityLink struct {
	ID           string `json:"id"`
	Number       int    `json:"number"`
	RepoFullName string `json:"repoFullName"`
}

func newLinkPRCmd() *cobra.Command {
	opts := linkPROptions{}
	cmd := &cobra.Command{
		Use:   "link-pr",
		Short: "Link a GitHub pull request to a record",
		Long: `Link a GitHub pull request to a record through Blue's GitHub integration.

The record's workspace must have an active linked GitHub repository. The pull
request can be a number, a number prefixed with #, or a full GitHub URL.`,
		Example: `  blue records link-pr --record <id> --pr 2016
  blue records link-pr -r <id> --pr '#2016' --workspace <id>
  blue records link-pr -r <id> --pr https://github.com/owner/repo/pull/2016 --simple`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runLinkPR(cmd, opts)
		},
	}

	cmd.Flags().StringVarP(&opts.recordID, "record", "r", "", "Record ID (required)")
	cmd.Flags().StringVar(&opts.pr, "pr", "", "PR number or full GitHub URL (required)")
	cmd.Flags().StringVarP(&opts.workspace, "workspace", "w", "", "Workspace ID or slug (optional context)")
	cmd.Flags().BoolVarP(&opts.simple, "simple", "s", false, "Show concise output")

	return cmd
}

func runLinkPR(cmd *cobra.Command, opts linkPROptions) error {
	recordID := strings.TrimSpace(opts.recordID)
	if recordID == "" {
		return fmt.Errorf("record ID is required. Use --record flag")
	}

	pr := strings.TrimSpace(opts.pr)
	if pr == "" {
		return fmt.Errorf("pull request is required. Use --pr flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	client := common.NewClient(config)
	if opts.workspace != "" {
		client.SetProject(opts.workspace)
	}

	mutation := `
		mutation AttachGitHubPr($input: AttachGitHubPrInput!) {
			attachGitHubPr(input: $input) {
				id
				number
				repoFullName
			}
		}
	`
	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"recordId": recordID,
			"pr":       pr,
		},
	}

	var response struct {
		AttachGitHubPR gitHubEntityLink `json:"attachGitHubPr"`
	}
	if err := client.ExecuteQueryWithResult(mutation, variables, &response); err != nil {
		return fmt.Errorf("failed to link pull request: %w", err)
	}

	link := response.AttachGitHubPR
	if opts.simple {
		fmt.Fprintf(cmd.OutOrStdout(), "GitHub link ID: %s\n", link.ID)
		return nil
	}

	fmt.Fprintln(cmd.OutOrStdout(), "Pull request linked successfully")
	fmt.Fprintf(cmd.OutOrStdout(), "Link ID: %s\n", link.ID)
	fmt.Fprintf(cmd.OutOrStdout(), "Repository: %s\n", link.RepoFullName)
	fmt.Fprintf(cmd.OutOrStdout(), "Pull request: #%d\n", link.Number)
	return nil
}
