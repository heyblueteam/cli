package domains

import (
	"fmt"

	"github.com/spf13/cobra"
)

var smtpListCmd = &cobra.Command{Use: "list", Short: "List SMTP credentials", RunE: runSmtpList}
var smtpVerifyCmd = &cobra.Command{Use: "verify", Short: "Verify SMTP connection settings", RunE: runSmtpVerify}

var smtpHost, smtpUsername, smtpPassword, smtpSenderName, smtpSenderEmail string
var smtpPort int

func init() {
	smtpVerifyCmd.Flags().StringVar(&smtpHost, "host", "", "SMTP host (required)")
	smtpVerifyCmd.Flags().IntVar(&smtpPort, "port", 587, "SMTP port")
	smtpVerifyCmd.Flags().StringVar(&smtpUsername, "username", "", "SMTP username (required)")
	smtpVerifyCmd.Flags().StringVar(&smtpPassword, "password", "", "SMTP password (required)")
	smtpVerifyCmd.Flags().StringVar(&smtpSenderName, "sender-name", "", "Sender name")
	smtpVerifyCmd.Flags().StringVar(&smtpSenderEmail, "sender-email", "", "Sender email")
}

func runSmtpList(cmd *cobra.Command, args []string) error {
	client, err := newClient()
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`query SmtpCredentials { smtpCredentials(skip: 0, take: 50) { items { %s } } }`, smtpFields)
	var response struct {
		SmtpCredentials struct {
			Items []SmtpCredential `json:"items"`
		} `json:"smtpCredentials"`
	}
	if err := client.ExecuteQueryWithResult(query, nil, &response); err != nil {
		return fmt.Errorf("failed to list SMTP credentials: %w", err)
	}
	for _, s := range response.SmtpCredentials.Items {
		fmt.Printf("%s  %s:%d  %s  verified=%t\n", s.ID, s.Host, s.Port, s.SenderEmail, s.VerifiedAt != "")
	}
	return nil
}

func runSmtpVerify(cmd *cobra.Command, args []string) error {
	if smtpHost == "" || smtpUsername == "" || smtpPassword == "" {
		return fmt.Errorf("--host, --username, and --password are required")
	}
	client, err := newClient()
	if err != nil {
		return err
	}
	input := map[string]interface{}{"host": smtpHost, "port": smtpPort, "username": smtpUsername, "password": smtpPassword}
	if smtpSenderName != "" {
		input["senderName"] = smtpSenderName
	}
	if smtpSenderEmail != "" {
		input["senderEmail"] = smtpSenderEmail
	}
	var response struct {
		VerifySmtpCredential bool `json:"verifySmtpCredential"`
	}
	if err := client.ExecuteQueryWithResult(`mutation VerifySmtp($input: VerifySmtpCredentialInput!) { verifySmtpCredential(input: $input) }`, map[string]interface{}{"input": input}, &response); err != nil {
		return fmt.Errorf("failed to verify SMTP settings: %w", err)
	}
	fmt.Printf("Verified: %t\n", response.VerifySmtpCredential)
	return nil
}
