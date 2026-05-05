package forms

import (
	"encoding/json"
	"fmt"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a single form by ID",
	Example: `  blue forms get --form <id>
  blue forms get --form <id> --workspace <id>
  blue forms get --form <id> --simple
  blue forms get --form <id> --format json`,
	RunE: runGet,
}

var (
	getForm      string
	getWorkspace string
	getSimple    bool
	getFormat    string
)

func init() {
	getCmd.Flags().StringVarP(&getForm, "form", "f", "", "Form ID (required)")
	getCmd.Flags().StringVarP(&getWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	getCmd.Flags().BoolVarP(&getSimple, "simple", "s", false, "Simple output format")
	getCmd.Flags().StringVar(&getFormat, "format", "", "Output format (json)")
}

func runGet(cmd *cobra.Command, args []string) error {
	if getForm == "" {
		return fmt.Errorf("form ID is required. Use --form flag")
	}
	if getWorkspace == "" {
		return fmt.Errorf("workspace is required. Use --workspace flag")
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)
	client.SetProject(getWorkspace)

	form, err := fetchForm(client, getForm)
	if err != nil {
		return err
	}

	if getFormat == "json" {
		out, err := json.MarshalIndent(form, "", "  ")
		if err != nil {
			return err
		}
		fmt.Println(string(out))
		return nil
	}

	printFormDetail(form, getSimple)
	return nil
}

func fetchForm(client *common.Client, id string) (*FormDetail, error) {
	query := `
		query GetForm($id: String!) {
			form(id: $id) {
				...FormDetailFields
			}
		}
	` + formDetailFragment

	variables := map[string]interface{}{"id": id}

	var response struct {
		Form FormDetail `json:"form"`
	}
	if err := client.ExecuteQueryWithResult(query, variables, &response); err != nil {
		return nil, fmt.Errorf("failed to get form: %w", err)
	}
	return &response.Form, nil
}

func printFormDetail(form *FormDetail, simple bool) {
	if simple {
		fmt.Printf("%s  %s  (uid=%s, active=%t)\n", form.ID, form.Title, form.UID, form.IsActive)
		fmt.Printf("Fields: %d\n", len(form.FormFields))
		for i, f := range form.FormFields {
			cf := ""
			if f.CustomField != nil {
				cf = fmt.Sprintf(" -> %s (%s)", f.CustomField.Name, f.CustomField.ID)
			}
			fmt.Printf("  %d. [%s] %s%s\n", i+1, f.Field, f.Name, cf)
		}
		return
	}

	fmt.Printf("=== Form: %s ===\n", form.Title)
	fmt.Printf("ID:            %s\n", form.ID)
	fmt.Printf("UID:           %s\n", form.UID)
	fmt.Printf("Active:        %t\n", form.IsActive)
	fmt.Printf("Theme:         %s\n", form.Theme)
	fmt.Printf("PrimaryColor:  %s\n", form.PrimaryColor)
	fmt.Printf("HideBranding:  %t\n", form.HideBranding)
	if form.Description != nil {
		fmt.Printf("Description:   %s\n", *form.Description)
	}
	if form.SubmitText != nil {
		fmt.Printf("SubmitText:    %s\n", *form.SubmitText)
	}
	if form.ResponseText != nil {
		fmt.Printf("ResponseText:  %s\n", *form.ResponseText)
	}
	if form.RedirectURL != nil {
		fmt.Printf("RedirectURL:   %s\n", *form.RedirectURL)
	}
	if form.ImageURL != nil {
		fmt.Printf("ImageURL:      %s\n", *form.ImageURL)
	}
	if form.FooterText != nil {
		fmt.Printf("FooterText:    %s\n", *form.FooterText)
	}
	if form.TodoList != nil {
		fmt.Printf("List:          %s (%s)\n", form.TodoList.Title, form.TodoList.ID)
	}
	fmt.Printf("Created:       %s\n", form.CreatedAt)
	fmt.Printf("Updated:       %s\n", form.UpdatedAt)

	fmt.Printf("\nFields (%d):\n", len(form.FormFields))
	for i, f := range form.FormFields {
		fmt.Printf("  %d. [%s] %s\n", i+1, f.Field, f.Name)
		fmt.Printf("     ID:           %s\n", f.ID)
		fmt.Printf("     Required:     %t\n", f.Required)
		fmt.Printf("     Position:     %.1f\n", f.Position)
		fmt.Printf("     Hidden:       %t\n", f.Hidden)
		if f.Placeholder != "" {
			fmt.Printf("     Placeholder:  %s\n", f.Placeholder)
		}
		if f.ExtraInfo != nil && *f.ExtraInfo != "" {
			fmt.Printf("     ExtraInfo:    %s\n", *f.ExtraInfo)
		}
		if f.CustomField != nil {
			fmt.Printf("     CustomField:  %s (%s, %s)\n", f.CustomField.Name, f.CustomField.ID, f.CustomField.Type)
		}
	}
}
