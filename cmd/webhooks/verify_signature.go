package webhooks

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var verifySignatureCmd = &cobra.Command{
	Use:   "verify-signature",
	Short: "Verify a webhook HMAC signature",
	Example: `  blue webhooks verify-signature --secret whsec_123 --signature <hex> --body-file payload.json
  blue webhooks verify-signature --secret whsec_123 --signature <hex> --body '{"event":"TODO_CREATED"}'`,
	RunE: runVerifySignature,
}

var (
	verifySecret    string
	verifySignature string
	verifyBody      string
	verifyBodyFile  string
)

func init() {
	verifySignatureCmd.Flags().StringVar(&verifySecret, "secret", "", "Webhook signing secret (required)")
	verifySignatureCmd.Flags().StringVar(&verifySignature, "signature", "", "X-Signature header value (required)")
	verifySignatureCmd.Flags().StringVar(&verifyBody, "body", "", "Raw request body")
	verifySignatureCmd.Flags().StringVar(&verifyBodyFile, "body-file", "", "Path to file containing the raw request body")
}

func runVerifySignature(cmd *cobra.Command, args []string) error {
	if verifySecret == "" {
		return fmt.Errorf("secret is required. Use --secret flag")
	}
	if verifySignature == "" {
		return fmt.Errorf("signature is required. Use --signature flag")
	}
	if verifyBody == "" && verifyBodyFile == "" {
		return fmt.Errorf("request body is required. Use --body or --body-file")
	}
	if verifyBody != "" && verifyBodyFile != "" {
		return fmt.Errorf("use only one of --body or --body-file")
	}

	body := []byte(verifyBody)
	if verifyBodyFile != "" {
		data, err := os.ReadFile(verifyBodyFile)
		if err != nil {
			return fmt.Errorf("failed to read body file: %w", err)
		}
		body = data
	}

	if !signaturesMatch(body, verifySecret, verifySignature) {
		return fmt.Errorf("signature does not match")
	}
	fmt.Println("Signature verified")
	return nil
}
