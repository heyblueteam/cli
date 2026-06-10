package open

import (
	"net/url"

	"github.com/spf13/cobra"
)

var workspaceCmd = &cobra.Command{
	Use:   "workspace <id-or-slug>",
	Short: "Open a workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		target, err := workspacePath(client, args[0], "")
		if err != nil {
			return err
		}
		return outputURL(target)
	},
}

var recordCmd = &cobra.Command{
	Use:     "record <id>",
	Aliases: []string{"rec"},
	Short:   "Open a record",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return openWorkspaceEntity("records/board/"+url.PathEscape(args[0]), openWorkspace)
	},
}

var formCmd = &cobra.Command{
	Use:   "form <id>",
	Short: "Open a form editor",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return openWorkspaceEntity("forms/"+url.PathEscape(args[0]), openWorkspace)
	},
}

var documentCmd = &cobra.Command{
	Use:     "document <id>",
	Aliases: []string{"doc"},
	Short:   "Open a document",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return openWorkspaceEntity("docs/"+url.PathEscape(args[0]), openWorkspace)
	},
}

var dashboardCmd = &cobra.Command{
	Use:     "dashboard <id>",
	Aliases: []string{"dash"},
	Short:   "Open a dashboard",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		target, err := orgPath(client, "dashboards/"+url.PathEscape(args[0]))
		if err != nil {
			return err
		}
		return outputURL(target)
	},
}

var reportCmd = &cobra.Command{
	Use:   "report <id>",
	Short: "Open a report",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := newClient()
		if err != nil {
			return err
		}
		target, err := orgPath(client, "reports/"+url.PathEscape(args[0]))
		if err != nil {
			return err
		}
		return outputURL(target)
	},
}

var filesCmd = &cobra.Command{
	Use:   "files",
	Short: "Open workspace files",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return openWorkspaceEntity("files", openWorkspace)
	},
}

var folderCmd = &cobra.Command{
	Use:   "folder <id>",
	Short: "Open a workspace file folder",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return openWorkspaceEntity("files/folder/"+url.PathEscape(args[0]), openWorkspace)
	},
}

func openWorkspaceEntity(path string, workspace string) error {
	client, err := newClient()
	if err != nil {
		return err
	}
	client.SetProject(workspace)
	target, err := workspacePath(client, workspace, path)
	if err != nil {
		return err
	}
	return outputURL(target)
}
