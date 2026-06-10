package documents

import (
	"fmt"

	"github.com/heyblueteam/cli/common"
	"github.com/spf13/cobra"
)

var deleteCmd = &cobra.Command{
	Use:     "delete",
	Short:   "Delete a document",
	Example: `  blue documents delete --document <id> --confirm`,
	RunE:    runDelete,
}

var (
	deleteDocument string
	deleteConfirm  bool
)

func init() {
	deleteCmd.Flags().StringVar(&deleteDocument, "document", "", "Document ID (required)")
	deleteCmd.Flags().BoolVarP(&deleteConfirm, "confirm", "y", false, "Confirm deletion (required)")
}

func runDelete(cmd *cobra.Command, args []string) error {
	if deleteDocument == "" {
		return fmt.Errorf("document ID is required. Use --document flag")
	}
	if !deleteConfirm {
		return fmt.Errorf("deletion confirmation is required. Use --confirm flag")
	}
	client, err := newClient("")
	if err != nil {
		return err
	}
	query := `mutation DeleteDocument($id: String!) { deleteDocument(id: $id) }`
	var response struct {
		DeleteDocument bool `json:"deleteDocument"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"id": deleteDocument}, &response); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	if !response.DeleteDocument {
		return fmt.Errorf("document was not deleted")
	}
	common.PrintSuccess(fmt.Sprintf("Deleted document %s", deleteDocument))
	return nil
}
