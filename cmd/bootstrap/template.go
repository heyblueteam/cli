package bootstrap

import "github.com/spf13/cobra"

var templateCmd = &cobra.Command{
	Use:     "template",
	Short:   "Print a bootstrap JSON template",
	Example: `  blue bootstrap template > workspace.json`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return printJSON(Config{
			Workspace: WorkspaceConfig{
				Name:        "New Process",
				Description: "Workspace created from the Blue CLI",
				Color:       "#3B82F6",
				Icon:        "mdi-briefcase-variant-outline",
				Category:    "GENERAL",
			},
			Lists: []ListConfig{{Title: "To Do"}, {Title: "In Progress"}, {Title: "Done"}},
			Tags:  []TagConfig{{Title: "Bug", Color: "#EF4444"}, {Title: "Priority", Color: "#F59E0B"}},
			Fields: []FieldConfig{
				{Name: "Priority", Type: "SELECT_SINGLE", Options: []map[string]interface{}{{"title": "High", "color": "#EF4444"}, {"title": "Medium", "color": "#F59E0B"}, {"title": "Low", "color": "#10B981"}}},
				{Name: "Estimate", Type: "NUMBER", Settings: map[string]interface{}{"min": 0}},
			},
		})
	},
}
