package domains

import (
	"fmt"

	"github.com/spf13/cobra"
)

var templatesListCmd = &cobra.Command{Use: "list", Short: "List email templates", RunE: runTemplatesList}
var templatesGetCmd = &cobra.Command{Use: "get", Short: "Get an email template by type", RunE: runTemplatesGet}
var templatesTestCmd = &cobra.Command{Use: "test", Short: "Send a test email template", RunE: runTemplatesTest}

var templateType, templateID, templateEmail string

func init() {
	templatesListCmd.Flags().StringVar(&templateType, "type", "", "Template type filter, e.g. INVITATION")
	templatesGetCmd.Flags().StringVar(&templateType, "type", "INVITATION", "Template type")
	templatesTestCmd.Flags().StringVar(&templateID, "template", "", "Email template ID (required)")
	templatesTestCmd.Flags().StringVar(&templateEmail, "email", "", "Recipient email (required)")
}

func runTemplatesList(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}
	filter := map[string]interface{}{}
	if templateType != "" {
		filter["type"] = templateType
	}
	query := fmt.Sprintf(`query EmailTemplates($filter: EmailTemplateFilterInput) { emailTemplates(filter: $filter, skip: 0, take: 50) { items { %s } } }`, templateFields)
	var response struct {
		EmailTemplates struct {
			Items []EmailTemplate `json:"items"`
		} `json:"emailTemplates"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"filter": filter}, &response); err != nil {
		return fmt.Errorf("failed to list email templates: %w", err)
	}
	return printJSON(response.EmailTemplates.Items)
}

func runTemplatesGet(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`query EmailTemplate($type: EmailTemplateType!) { emailTemplate(type: $type) { %s } }`, templateFields)
	var response struct {
		EmailTemplate *EmailTemplate `json:"emailTemplate"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"type": templateType}, &response); err != nil {
		return fmt.Errorf("failed to get email template: %w", err)
	}
	return printJSON(response.EmailTemplate)
}

func runTemplatesTest(cmd *cobra.Command, args []string) error {
	if templateID == "" || templateEmail == "" {
		return fmt.Errorf("--template and --email are required")
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	var response struct {
		SendTestEmail bool `json:"sendTestEmail"`
	}
	variables := map[string]interface{}{"input": map[string]interface{}{"emailTemplateId": templateID, "email": templateEmail}}
	if err := client.ExecuteQueryWithResult(`mutation SendTestEmail($input: SendTestEmailInput!) { sendTestEmail(input: $input) }`, variables, &response); err != nil {
		return fmt.Errorf("failed to send test email: %w", err)
	}
	fmt.Printf("Sent: %t\n", response.SendTestEmail)
	return nil
}
