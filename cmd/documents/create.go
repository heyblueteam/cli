package documents

import (
	"fmt"

	"github.com/spf13/cobra"
)

var createCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a document or wiki page",
	Example: `  blue documents create --workspace <id> --title "Runbook" --content '<h1>Runbook</h1>'
  blue documents create --workspace <id> --title "Handbook" --wiki --content-file handbook.html`,
	RunE: runCreate,
}

var (
	createWorkspace   string
	createTitle       string
	createContent     string
	createContentFile string
	createWiki        bool
	createFormat      string
)

func init() {
	createCmd.Flags().StringVarP(&createWorkspace, "workspace", "w", "", "Workspace ID or slug (required)")
	createCmd.Flags().StringVarP(&createTitle, "title", "t", "", "Document title")
	createCmd.Flags().StringVar(&createContent, "content", "", "HTML content")
	createCmd.Flags().StringVar(&createContentFile, "content-file", "", "File containing HTML content")
	createCmd.Flags().BoolVar(&createWiki, "wiki", false, "Create as a wiki page")
	createCmd.Flags().StringVar(&createFormat, "format", "", "Output format (json)")
}

func runCreate(cmd *cobra.Command, args []string) error {
	if createWorkspace == "" {
		return fmt.Errorf("workspace ID is required. Use --workspace flag")
	}
	content, err := readContent(createContent, createContentFile)
	if err != nil {
		return err
	}
	client, err := newClient(createWorkspace)
	if err != nil {
		return err
	}
	projectID, err := client.ResolveProjectID(createWorkspace)
	if err != nil {
		return err
	}
	client.SetProject(projectID)
	input := map[string]interface{}{"projectId": projectID, "wiki": createWiki}
	if createTitle != "" {
		input["title"] = createTitle
	}
	if content != "" {
		input["content"] = content
	}
	query := fmt.Sprintf(`mutation CreateDocument($input: CreateDocumentInput!) { createDocument(input: $input) { %s } }`, documentFields)
	var response struct {
		CreateDocument Document `json:"createDocument"`
	}
	if err := client.ExecuteQueryWithResult(query, map[string]interface{}{"input": input}, &response); err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}
	if createFormat == "json" {
		return printJSON(response.CreateDocument)
	}
	fmt.Println("Document created")
	printDocument(response.CreateDocument, false)
	return nil
}
