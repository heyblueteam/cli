package company

import (
	"fmt"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add <slug>",
	Short: "Add a company to the known list",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		slug := args[0]

		if err := common.AddCompany(slug); err != nil {
			return err
		}

		common.PrintSuccess(fmt.Sprintf("Added company %q", slug))

		active, _ := common.GetActiveCompany()
		if active == "" {
			common.PrintInfo(fmt.Sprintf("Run 'blue company use %s' to set it as active", slug))
		}

		return nil
	},
}
