package forms

import (
	"fmt"
	"strconv"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update an existing form",
	Long: `Update form metadata, copy, branding, target list, assignees, tags, and active state.

Form fields can be added or modified inline (--field / --fields-file). When
--fields-file is given, --field flags are ignored. To remove a field, use
'blue forms fields delete'.`,
	Example: `  # Toggle active state
  blue forms update --form <id> --active true

  # Rebrand and redirect
  blue forms update --form <id> --primary-color "#ff0000" --redirect-url "https://example.com/done"

  # Move submissions to a different list
  blue forms update --form <id> --list <list-id>

  # Replace assignees
  blue forms update --form <id> --assignees "u1,u2"

  # Add a custom field
  blue forms update --form <id> --field "type=custom;customField=cf_xxx;name=Phone;required=true"`,
	RunE: runUpdate,
}

var (
	updateForm         string
	updateWorkspace    string
	updateTitle        string
	updateDescription  string
	updatePrimaryColor string
	updateHideBranding string
	updateTheme        string
	updateSubmitText   string
	updateResponseText string
	updateRedirectURL  string
	updateImageURL     string
	updateFooterText   string
	updateShowFooter   string
	updateListID       string
	updateAssignees    string
	updateTagIDs       string
	updateActive       string
	updateFields       []string
	updateFieldsFile   string
	updateSimple       bool
)

func init() {
	updateCmd.Flags().StringVarP(&updateForm, "form", "f", "", "Form ID (required)")
	updateCmd.Flags().StringVarP(&updateWorkspace, "workspace", "w", "", "Workspace ID or slug (optional)")
	updateCmd.Flags().StringVarP(&updateTitle, "title", "t", "", "Form title")
	updateCmd.Flags().StringVar(&updateDescription, "description", "", "Form description")
	updateCmd.Flags().StringVar(&updatePrimaryColor, "primary-color", "", "Hex primary color")
	updateCmd.Flags().StringVar(&updateHideBranding, "hide-branding", "", "true|false")
	updateCmd.Flags().StringVar(&updateTheme, "theme", "", "Theme (light|dark)")
	updateCmd.Flags().StringVar(&updateSubmitText, "submit-text", "", "Submit button label")
	updateCmd.Flags().StringVar(&updateResponseText, "response-text", "", "Post-submit response text")
	updateCmd.Flags().StringVar(&updateRedirectURL, "redirect-url", "", "Redirect URL after submit")
	updateCmd.Flags().StringVar(&updateImageURL, "image-url", "", "Header image URL")
	updateCmd.Flags().StringVar(&updateFooterText, "footer-text", "", "Footer text")
	updateCmd.Flags().StringVar(&updateShowFooter, "show-footer", "", "true|false")
	updateCmd.Flags().StringVar(&updateListID, "list", "", "List ID submissions land in")
	updateCmd.Flags().StringVar(&updateAssignees, "assignees", "", "Comma-separated user IDs (replaces existing)")
	updateCmd.Flags().StringVar(&updateTagIDs, "tag-ids", "", "Comma-separated tag IDs (replaces existing)")
	updateCmd.Flags().StringVar(&updateActive, "active", "", "true|false")
	updateCmd.Flags().StringArrayVar(&updateFields, "field", nil, "Repeatable: field spec (e.g. \"type=title;name=Full name\")")
	updateCmd.Flags().StringVar(&updateFieldsFile, "fields-file", "", "Path to JSON file with array of field specs")
	updateCmd.Flags().BoolVarP(&updateSimple, "simple", "s", false, "Simple output format")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if updateForm == "" {
		return fmt.Errorf("form ID is required. Use --form flag")
	}

	// Field handling — file wins over inline if both are present.
	var specs []FieldSpec
	if updateFieldsFile != "" || len(updateFields) > 0 {
		inline := updateFields
		if updateFieldsFile != "" && len(updateFields) > 0 {
			fmt.Fprintln(cmd.OutOrStderr(), "warning: --fields-file given, ignoring --field flags")
			inline = nil
		}
		var err error
		specs, err = collectFieldSpecs(updateFieldsFile, inline)
		if err != nil {
			return err
		}
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)
	if updateWorkspace != "" {
		client.SetProject(updateWorkspace)
	}

	updateInput := map[string]interface{}{"id": updateForm}
	if updateTitle != "" {
		updateInput["title"] = updateTitle
	}
	if cmd.Flags().Changed("description") {
		updateInput["description"] = updateDescription
	}
	if updatePrimaryColor != "" {
		updateInput["primaryColor"] = updatePrimaryColor
	}
	if updateHideBranding != "" {
		b, err := strconv.ParseBool(updateHideBranding)
		if err != nil {
			return fmt.Errorf("invalid --hide-branding %q (use true|false)", updateHideBranding)
		}
		updateInput["hideBranding"] = b
	}
	if updateTheme != "" {
		updateInput["theme"] = updateTheme
	}
	if cmd.Flags().Changed("submit-text") {
		updateInput["submitText"] = updateSubmitText
	}
	if cmd.Flags().Changed("response-text") {
		updateInput["responseText"] = updateResponseText
	}
	if cmd.Flags().Changed("redirect-url") {
		updateInput["redirectURL"] = updateRedirectURL
	}
	if cmd.Flags().Changed("image-url") {
		updateInput["imageURL"] = updateImageURL
	}
	if cmd.Flags().Changed("footer-text") {
		updateInput["footerText"] = updateFooterText
	}
	if updateShowFooter != "" {
		b, err := strconv.ParseBool(updateShowFooter)
		if err != nil {
			return fmt.Errorf("invalid --show-footer %q (use true|false)", updateShowFooter)
		}
		updateInput["showFooter"] = b
	}
	if updateListID != "" {
		updateInput["todoListId"] = updateListID
	}
	if cmd.Flags().Changed("assignees") {
		updateInput["assigneeIds"] = splitCSV(updateAssignees)
	}
	if cmd.Flags().Changed("tag-ids") {
		updateInput["tagIds"] = splitCSV(updateTagIDs)
	}
	if updateActive != "" {
		b, err := strconv.ParseBool(updateActive)
		if err != nil {
			return fmt.Errorf("invalid --active %q (use true|false)", updateActive)
		}
		updateInput["isActive"] = b
	}

	// updateForm ignores formFields — every field goes through upsertFormField.
	if len(updateInput) == 1 && len(specs) == 0 {
		return fmt.Errorf("nothing to update — pass at least one flag")
	}

	if len(updateInput) > 1 {
		if err := executeUpdateForm(client, updateInput); err != nil {
			return fmt.Errorf("updateForm failed: %w", err)
		}
	}

	for _, s := range specs {
		fieldID := s.ID
		if fieldID == "" {
			fieldID = common.NewCuid()
		}
		if err := executeUpsertFormField(client, updateForm, fieldID, s); err != nil {
			return fmt.Errorf("upsertFormField for %q failed: %w", s.Name, err)
		}
	}

	form, err := fetchForm(client, updateForm)
	if err != nil {
		return fmt.Errorf("update succeeded but failed to read back: %w", err)
	}
	printFormDetail(form, updateSimple)
	return nil
}
