package company

import (
	"fmt"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

var useCmd = &cobra.Command{
	Use:   "use <slug>",
	Short: "Set the active company",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		slug := args[0]

		// Add to known list if not already there
		if err := common.AddCompany(slug); err != nil {
			return err
		}

		if err := common.SetActiveCompany(slug); err != nil {
			return err
		}

		common.PrintSuccess(fmt.Sprintf("Switched to company %q", slug))
		return nil
	},
}
