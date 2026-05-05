package forms

import (
	"fmt"
	"strings"

	"github.com/heyblueteam/cli/common"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new form",
	Long: `Create a new form in a workspace.

The Blue API splits form creation across two mutations: createForm only accepts
title/description/primaryColor/hideBranding, and everything else is set via
updateForm. This command hides that — pass any flag and it runs the second
mutation transparently. Form fields are added via upsertFormField after the
form exists.

If the upstream calls fail partway through, the form ID is printed to stderr so
you can recover with 'blue forms update' or 'blue forms delete'.`,
	Example: `  # Minimal — title only
  blue forms create -w <workspace-id> --title "Contact us"

  # Branded form on a specific list with a few inline fields
  blue forms create -w <workspace-id> --title "Lead intake" \
    --description "Tell us about your project" \
    --primary-color "#0066ff" --theme dark --hide-branding \
    --submit-text "Send" --response-text "Thanks!" \
    --redirect-url "https://example.com/thanks" \
    --list <list-id> --active \
    --field "type=title;name=Full name;required=true;position=1000" \
    --field "type=description;name=Project details;placeholder=Tell us more;position=2000" \
    --field "type=custom;customField=cf_xxx;name=Budget;required=true;position=3000"

  # Realistic form — fields defined in JSON
  blue forms create -w <workspace-id> --title "Lead intake" --fields-file ./form-fields.json`,
	RunE: runCreate,
}

var (
	createWorkspace    string
	createTitle        string
	createDescription  string
	createPrimaryColor string
	createHideBranding bool
	createTheme        string
	createSubmitText   string
	createResponseText string
	createRedirectURL  string
	createImageURL     string
	createFooterText   string
	createShowFooter   bool
	createListID       string
	createAssignees    string
	createTagIDs       string
	createActive       bool
	createInactive     bool
	createFields       []string
	createFieldsFile   string
	createSimple       bool
)

func init() {
	createCmd.Flags().StringVarP(&createWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	createCmd.Flags().StringVarP(&createTitle, "title", "t", "", "Form title (required)")
	createCmd.Flags().StringVar(&createDescription, "description", "", "Form description")
	createCmd.Flags().StringVar(&createPrimaryColor, "primary-color", "", "Hex primary color (e.g. #0066ff)")
	createCmd.Flags().BoolVar(&createHideBranding, "hide-branding", false, "Hide Blue branding")
	createCmd.Flags().StringVar(&createTheme, "theme", "", "Theme (light|dark)")
	createCmd.Flags().StringVar(&createSubmitText, "submit-text", "", "Submit button label")
	createCmd.Flags().StringVar(&createResponseText, "response-text", "", "Post-submit response text")
	createCmd.Flags().StringVar(&createRedirectURL, "redirect-url", "", "URL to redirect to after submit")
	createCmd.Flags().StringVar(&createImageURL, "image-url", "", "Header image URL")
	createCmd.Flags().StringVar(&createFooterText, "footer-text", "", "Footer text")
	createCmd.Flags().BoolVar(&createShowFooter, "show-footer", false, "Show footer")
	createCmd.Flags().StringVar(&createListID, "list", "", "List ID submissions land in")
	createCmd.Flags().StringVar(&createAssignees, "assignees", "", "Comma-separated user IDs to assign to new records")
	createCmd.Flags().StringVar(&createTagIDs, "tag-ids", "", "Comma-separated tag IDs to apply to new records")
	createCmd.Flags().BoolVar(&createActive, "active", false, "Activate the form (default: inactive)")
	createCmd.Flags().BoolVar(&createInactive, "inactive", false, "Force inactive (overrides --active)")
	createCmd.Flags().StringArrayVar(&createFields, "field", nil, "Repeatable: field spec (e.g. \"type=title;name=Full name;required=true\")")
	createCmd.Flags().StringVar(&createFieldsFile, "fields-file", "", "Path to JSON file with array of field specs")
	createCmd.Flags().BoolVarP(&createSimple, "simple", "s", false, "Simple output format")
}

type createFormResp struct {
	CreateForm struct {
		ID string `json:"id"`
	} `json:"createForm"`
}

func runCreate(cmd *cobra.Command, args []string) error {
	if createWorkspace == "" {
		return fmt.Errorf("workspace is required. Use --workspace flag")
	}
	if createTitle == "" {
		return fmt.Errorf("title is required. Use --title flag")
	}

	specs, err := collectFieldSpecs(createFieldsFile, createFields)
	if err != nil {
		return err
	}

	config, err := common.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	client := common.NewClient(config)
	client.SetProject(createWorkspace)

	// Step 1: createForm with the minimal accepted input.
	createInput := map[string]interface{}{
		"projectId": createWorkspace,
		"title":     createTitle,
	}
	if createDescription != "" {
		createInput["description"] = createDescription
	}
	if createPrimaryColor != "" {
		createInput["primaryColor"] = createPrimaryColor
	}
	if createHideBranding {
		createInput["hideBranding"] = true
	}

	createMutation := `
		mutation CreateForm($input: CreateFormInput!) {
			createForm(input: $input) { id }
		}
	`

	var created createFormResp
	if err := client.ExecuteQueryWithResult(createMutation, map[string]interface{}{"input": createInput}, &created); err != nil {
		return fmt.Errorf("createForm failed: %w", err)
	}
	formID := created.CreateForm.ID
	fmt.Fprintf(cmd.OutOrStderr(), "[1/3] Created form %s\n", formID)

	// Step 2: updateForm — push everything else (theme, copy, list, assignees,
	// tags, active state, primitive fields). Custom-field linkage is set in
	// step 3 because UpdateFormInput.formFields cannot carry customFieldId.
	updateInput := map[string]interface{}{"id": formID}
	if createTheme != "" {
		updateInput["theme"] = createTheme
	}
	if createSubmitText != "" {
		updateInput["submitText"] = createSubmitText
	}
	if createResponseText != "" {
		updateInput["responseText"] = createResponseText
	}
	if createRedirectURL != "" {
		updateInput["redirectURL"] = createRedirectURL
	}
	if createImageURL != "" {
		updateInput["imageURL"] = createImageURL
	}
	if createFooterText != "" {
		updateInput["footerText"] = createFooterText
	}
	if cmd.Flags().Changed("show-footer") {
		updateInput["showFooter"] = createShowFooter
	}
	if createListID != "" {
		updateInput["todoListId"] = createListID
	}
	if createAssignees != "" {
		updateInput["assigneeIds"] = splitCSV(createAssignees)
	}
	if createTagIDs != "" {
		updateInput["tagIds"] = splitCSV(createTagIDs)
	}
	if createInactive {
		updateInput["isActive"] = false
	} else if createActive {
		updateInput["isActive"] = true
	}

	// updateForm ignores its formFields input — every field (built-in and
	// custom) goes through upsertFormField in step 3.
	hasUpdateWork := len(updateInput) > 1
	if hasUpdateWork {
		if err := executeUpdateForm(client, updateInput); err != nil {
			return fmt.Errorf("form %s created but updateForm failed: %w", formID, err)
		}
		fmt.Fprintf(cmd.OutOrStderr(), "[2/3] Configured form\n")
	} else {
		fmt.Fprintf(cmd.OutOrStderr(), "[2/3] Skipped configure (no extra options)\n")
	}

	// Step 3: upsertFormField for every field — both built-in (title,
	// description, tags, startedAt, duedAt) and custom.
	for _, s := range specs {
		if err := executeUpsertFormField(client, formID, common.NewCuid(), s); err != nil {
			return fmt.Errorf("form %s configured but upsertFormField for %q failed: %w", formID, s.Name, err)
		}
	}
	fmt.Fprintf(cmd.OutOrStderr(), "[3/3] Added %d field(s)\n", len(specs))

	form, err := fetchForm(client, formID)
	if err != nil {
		return fmt.Errorf("form created (id=%s) but failed to read it back: %w", formID, err)
	}

	fmt.Fprintln(cmd.OutOrStderr())
	printFormDetail(form, createSimple)
	return nil
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func executeUpdateForm(client *common.Client, input map[string]interface{}) error {
	mutation := `
		mutation UpdateForm($input: UpdateFormInput!) {
			updateForm(input: $input) { id }
		}
	`
	var resp struct {
		UpdateForm struct {
			ID string `json:"id"`
		} `json:"updateForm"`
	}
	return client.ExecuteQueryWithResult(mutation, map[string]interface{}{"input": input}, &resp)
}

func executeUpsertFormField(client *common.Client, formID, formFieldID string, spec FieldSpec) error {
	input := map[string]interface{}{
		"formId":      formID,
		"formFieldId": formFieldID,
		"field":       spec.Field,
	}
	if spec.CustomFieldID != "" {
		input["customFieldId"] = spec.CustomFieldID
	}
	if spec.Name != "" {
		input["name"] = spec.Name
	}
	if spec.Placeholder != "" {
		input["placeholder"] = spec.Placeholder
	}
	if spec.Position != nil {
		input["position"] = *spec.Position
	}
	if spec.Required != nil {
		input["required"] = *spec.Required
	}
	if spec.Hidden != nil {
		input["hidden"] = *spec.Hidden
	} else {
		// API defaults `hidden` to true when omitted, which makes the field
		// invisible on the public form — surprising for CLI users.
		input["hidden"] = false
	}
	if spec.AddToDescription != nil {
		input["addToDescription"] = *spec.AddToDescription
	}
	if spec.ExtraInfo != nil {
		input["extraInfo"] = *spec.ExtraInfo
	}

	mutation := `
		mutation UpsertFormField($input: UpsertFormFieldInput!) {
			upsertFormField(input: $input) { id }
		}
	`
	var resp struct {
		UpsertFormField struct {
			ID string `json:"id"`
		} `json:"upsertFormField"`
	}
	return client.ExecuteQueryWithResult(mutation, map[string]interface{}{"input": input}, &resp)
}
