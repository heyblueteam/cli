package company

import (
	"fmt"

	"blue-cli/common"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove <slug>",
	Short: "Remove a company from the known list",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		slug := args[0]

		if err := common.RemoveCompany(slug); err != nil {
			return err
		}

		common.PrintSuccess(fmt.Sprintf("Removed company %q", slug))

		active, _ := common.GetActiveCompany()
		if active == slug {
			common.PrintInfo("Warning: removed the active company. Run 'blue company use <slug>' to set a new one.")
		}

		return nil
	},
}
