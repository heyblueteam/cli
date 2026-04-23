package company

import (
	"fmt"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List known companies",
	RunE: func(cmd *cobra.Command, args []string) error {
		companies, err := common.GetCompanies()
		if err != nil {
			return err
		}

		if len(companies) == 0 {
			common.PrintInfo("No companies configured. Run 'blue company add <slug>' or 'blue init'.")
			return nil
		}

		active, _ := common.GetActiveCompany()

		for _, c := range companies {
			if c == active {
				fmt.Printf("  * %s (active)\n", c)
			} else {
				fmt.Printf("    %s\n", c)
			}
		}

		return nil
	},
}
