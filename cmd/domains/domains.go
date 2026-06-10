package domains

import (
	"encoding/json"
	"fmt"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{Use: "domains", Short: "Manage custom domains and email settings"}

var domainsCmd = &cobra.Command{Use: "domains", Short: "Manage custom domains"}
var smtpCmd = &cobra.Command{Use: "smtp", Short: "Manage SMTP credentials"}
var templatesCmd = &cobra.Command{Use: "templates", Short: "Manage email templates"}

func init() {
	domainsCmd.AddCommand(domainsListCmd, domainsCreateCmd, domainsVerifyCmd, domainsUpdateCmd, domainsDeleteCmd)
	smtpCmd.AddCommand(smtpListCmd, smtpVerifyCmd)
	templatesCmd.AddCommand(templatesListCmd, templatesGetCmd, templatesTestCmd)
	Cmd.AddCommand(domainsCmd, smtpCmd, templatesCmd)
}

func newClient() (*common.Client, error) {
	config, err := common.LoadConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to load configuration: %w", err)
	}
	return common.NewClient(config), nil
}

func printJSON(value interface{}) error {
	out, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	fmt.Println(string(out))
	return nil
}

type CustomDomain struct {
	ID                 string `json:"id"`
	UID                string `json:"uid"`
	Name               string `json:"name"`
	ApplicationType    string `json:"applicationType"`
	VerificationStatus string `json:"verificationStatus"`
	CreatedAt          string `json:"createdAt"`
	UpdatedAt          string `json:"updatedAt"`
}

type SmtpCredential struct {
	ID          string `json:"id"`
	UID         string `json:"uid"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	SenderName  string `json:"senderName"`
	SenderEmail string `json:"senderEmail"`
	VerifiedAt  string `json:"verifiedAt"`
}

type EmailTemplate struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Enabled   bool   `json:"enabled"`
	Subject   string `json:"subject"`
	Body      string `json:"body"`
	CtaText   string `json:"ctaText"`
	CtaLink   string `json:"ctaLink"`
	Footer    string `json:"footer"`
	UpdatedAt string `json:"updatedAt"`
}

const customDomainFields = `id uid name applicationType verificationStatus createdAt updatedAt`
const smtpFields = `id uid host port username senderName senderEmail verifiedAt`
const templateFields = `id type enabled subject body ctaText ctaLink footer updatedAt`
