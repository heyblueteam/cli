package documents

import (
	"fmt"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update a document",
	Example: `  blue documents update --document <id> --title "New title"
  blue documents update --document <id> --content-file updated.html`,
	RunE: runUpdate,
}

var (
	updateDocument      string
	updateTitle         string
	updateContent       string
	updateContentFile   string
	updateContentBase64 string
	updateWiki          string
	updateFormat        string
)

func init() {
	updateCmd.Flags().StringVar(&updateDocument, "document", "", "Document ID (required)")
	updateCmd.Flags().StringVar(&updateTitle, "title", "", "New title")
	updateCmd.Flags().StringVar(&updateContent, "content", "", "New HTML content")
	updateCmd.Flags().StringVar(&updateContentFile, "content-file", "", "File containing new HTML content")
	updateCmd.Flags().StringVar(&updateContentBase64, "content-base64", "", "New base64 collaboration snapshot")
	updateCmd.Flags().StringVar(&updateWiki, "wiki", "", "Set wiki state (true or false)")
	updateCmd.Flags().StringVar(&updateFormat, "format", "", "Output format (json)")
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if updateDocument == "" {
		return fmt.Errorf("document ID is required. Use --document flag")
	}
	content, err := readContent(updateContent, updateContentFile)
	if err != nil {
		return err
	}
	input := map[string]interface{}{"id": updateDocument}
	if updateTitle != "" {
		input["title"] = updateTitle
	}
	if content != "" {
		input["content"] = content
	}
	if updateContentBase64 != "" {
		input["contentBase64"] = updateContentBase64
	}
	if updateWiki != "" {
		switch updateWiki {
		case "true":
			input["wiki"] = true
		case "false":
			input["wiki"] = false
		default:
			return fmt.Errorf("--wiki must be true or false")
		}
	}
	if len(input) == 1 {
		return fmt.Errorf("nothing to update. Pass at least one field flag")
	}
	client, err := newClient("")
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`mutation UpdateDocument($input: UpdateDocumentInput!) { updateDocument(input: $input) { %s } }`, documentFields)
	var response struct {
		UpdateDocument Document `json:"updateDocument"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"input": input}, &response); err != nil {
		return fmt.Errorf("failed to update document: %w", err)
	}
	if updateFormat == "json" {
		return printJSON(response.UpdateDocument)
	}
	printDocument(response.UpdateDocument, false)
	return nil
}
