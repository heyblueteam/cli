package forms

import (
	"fmt"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

var copyCmd = &cobra.Command{
	Use:   "copy",
	Short: "Duplicate a form",
	Example: `  blue forms copy --form <id>
  blue forms copy --form <id> --workspace <id>`,
	RunE: runCopy,
}

var (
	copyForm      string
	copyWorkspace string
	copySimple    bool
)

func init() {
	copyCmd.Flags().StringVarP(&copyForm, "form", "f", "", "Form ID to copy (required)")
	copyCmd.Flags().StringVarP(&copyWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	copyCmd.Flags().BoolVarP(&copySimple, "simple", "s", false, "Simple output format")
}

func runCopy(cmd *cobra.Command, args []string) error {
	if copyForm == "" {
		return fmt.Errorf("form ID is required. Use --form flag")
	}
	if copyWorkspace == "" {
		return fmt.Errorf("workspace is required. Use --workspace flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)
	client.SetProject(copyWorkspace)

	mutation := `
		mutation CopyForm($input: CopyFormInput!) {
			copyForm(input: $input) {
				...FormDetailFields
			}
		}
	` + formDetailFragment

	variables := map[string]interface{}{
		"input": map[string]interface{}{"formId": copyForm},
	}

	var resp struct {
		CopyForm FormDetail `json:"copyForm"`
	}
	if err := client.ExecuteQueryWithResult(mutation, variables, &resp); err != nil {
		return fmt.Errorf("copyForm failed: %w", err)
	}
	printFormDetail(&resp.CopyForm, copySimple)
	return nil
}
