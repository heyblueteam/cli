package fields

import (
	"encoding/json"
	"fmt"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List fields on a form",
	Example: `  blue forms fields list --form <id>
  blue forms fields list --form <id> --simple
  blue forms fields list --form <id> --format json`,
	RunE: runList,
}

var (
	listForm      string
	listWorkspace string
	listSimple    bool
	listFormat    string
)

func init() {
	listCmd.Flags().StringVarP(&listForm, "form", "f", "", "Form ID (required)")
	listCmd.Flags().StringVarP(&listWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	listCmd.Flags().BoolVarP(&listSimple, "simple", "s", false, "Simple output format")
	listCmd.Flags().StringVar(&listFormat, "format", "", "Output format (json)")
}

type formFieldRow struct {
	ID               string  `json:"id"`
	UID              string  `json:"uid"`
	Field            string  `json:"field"`
	Name             string  `json:"name"`
	Placeholder      string  `json:"placeholder"`
	Required         bool    `json:"required"`
	Position         float64 `json:"position"`
	AddToDescription bool    `json:"addToDescription"`
	Hidden           bool    `json:"hidden"`
	ExtraInfo        *string `json:"extraInfo"`
	CustomField      *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		Type string `json:"type"`
	} `json:"customField"`
}

func runList(cmd *cobra.Command, args []string) error {
	if listForm == "" {
		return fmt.Errorf("form ID is required. Use --form flag")
	}
	if listWorkspace == "" {
		return fmt.Errorf("workspace is required. Use --workspace flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)
	client.SetProject(listWorkspace)

	query := `
		query FormFields($filter: FormFieldFilterInput) {
			formFields(filter: $filter) {
				id
				uid
				field
				name
				placeholder
				required
				position
				addToDescription
				hidden
				extraInfo
				customField {
					id
					name
					type
				}
			}
		}
	`
	variables := map[string]interface{}{
		"filter": map[string]interface{}{"formId": listForm},
	}
	var resp struct {
		FormFields []formFieldRow `json:"formFields"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &resp); err != nil {
		return fmt.Errorf("failed to list form fields: %w", err)
	}

	if listFormat == "json" {
		out, err := json.MarshalIndent(resp.FormFields, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(out))
		return nil
	}

	if listSimple {
		for i, f := range resp.FormFields {
			cf := ""
			if f.CustomField != nil {
				cf = fmt.Sprintf(" -> %s", f.CustomField.Name)
			}
			fmt.Printf("%d. %s  [%s] %s%s  (id=%s)\n", i+1, f.Name, f.Field, requiredFlag(f.Required), cf, f.ID)
		}
		return nil
	}

	fmt.Printf("=== Form fields (%d) ===\n\n", len(resp.FormFields))
	for i, f := range resp.FormFields {
		fmt.Printf("%d. %s\n", i+1, f.Name)
		fmt.Printf("   ID:           %s\n", f.ID)
		fmt.Printf("   Type:         %s\n", f.Field)
		fmt.Printf("   Required:     %t\n", f.Required)
		fmt.Printf("   Position:     %.1f\n", f.Position)
		fmt.Printf("   Hidden:       %t\n", f.Hidden)
		if f.Placeholder != "" {
			fmt.Printf("   Placeholder:  %s\n", f.Placeholder)
		}
		if f.ExtraInfo != nil && *f.ExtraInfo != "" {
			fmt.Printf("   ExtraInfo:    %s\n", *f.ExtraInfo)
		}
		if f.CustomField != nil {
			fmt.Printf("   CustomField:  %s (%s, %s)\n", f.CustomField.Name, f.CustomField.ID, f.CustomField.Type)
		}
		fmt.Println()
	}
	return nil
}

func requiredFlag(b bool) string {
	if b {
		return "required"
	}
	return "optional"
}
