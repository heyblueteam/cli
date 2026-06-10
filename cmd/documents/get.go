package documents

import (
	"fmt"

	"github.com/spf13/cobra"
)

var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get a document",
	Example: `  blue documents get --document <id>
  blue documents get --document <id> --content --format json`,
	RunE: runGet,
}

var (
	getDocument string
	getContent  bool
	getFormat   string
)

func init() {
	getCmd.Flags().StringVar(&getDocument, "document", "", "Document ID (required)")
	getCmd.Flags().BoolVar(&getContent, "content", false, "Print HTML content")
	getCmd.Flags().StringVar(&getFormat, "format", "", "Output format (json)")
}

func runGet(cmd *cobra.Command, args []string) error {
	if getDocument == "" {
		return fmt.Errorf("document ID is required. Use --document flag")
	}
	client, err := newClient("")
	if err != nil {
		return err
	}
	query := fmt.Sprintf(`query GetDocument($id: String!) { document(id: $id) { %s } }`, documentFields)
	var response struct {
		Document Document `json:"document"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"id": getDocument}, &response); err != nil {
		return fmt.Errorf("failed to get document: %w", err)
	}
	if getFormat == "json" {
		return printJSON(response.Document)
	}
	printDocument(response.Document, getContent)
	return nil
}
