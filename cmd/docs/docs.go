package docs

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"unicode"

	"github.com/heyblueteam/cli/internal/apidocs"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "docs [slug]",
	Short: "Browse and search Blue API docs",
	Long:  "Browse the embedded Blue API documentation snapshot from the terminal.",
	Example: `  blue docs
  blue docs list records
  blue docs search "automation trigger"
  blue docs show records/list-records
  blue docs records/list-records --open`,
	Args: cobra.MaximumNArgs(1),
	RunE: runDocs,
}

var (
	docsOpen     bool
	docsPrintURL bool
	docsLimit    int
)

func init() {
	Cmd.PersistentFlags().BoolVar(&docsOpen, "open", false, "Open the matching doc on blue.app")
	Cmd.PersistentFlags().BoolVar(&docsPrintURL, "print-url", false, "Print the canonical docs URL")
	Cmd.PersistentFlags().IntVar(&docsLimit, "limit", 10, "Maximum search results to show")

	Cmd.AddCommand(listCmd)
	Cmd.AddCommand(searchCmd)
	Cmd.AddCommand(showCmd)
}

func runDocs(cmd *cobra.Command, args []string) error {
	index, err := apidocs.LoadIndex()
	if err != nil {
		return err
	}
	if len(args) == 0 {
		return printSections(index)
	}

	input := normalizeSlug(args[0])
	if doc, ok := findDocument(index, input); ok {
		return outputDocument(doc)
	}
	if hasSection(index, input) {
		return printDocuments(index, input)
	}
	return fmt.Errorf("docs page or section %q not found. Try 'blue docs search %q'", args[0], args[0])
}

var listCmd = &cobra.Command{
	Use:   "list [section]",
	Short: "List API docs sections or pages",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := apidocs.LoadIndex()
		if err != nil {
			return err
		}
		if len(args) == 0 {
			return printSections(index)
		}
		return printDocuments(index, normalizeSlug(args[0]))
	},
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search API docs",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := apidocs.LoadIndex()
		if err != nil {
			return err
		}
		return printSearchResults(index, args[0])
	},
}

var showCmd = &cobra.Command{
	Use:   "show <slug>",
	Short: "Show an API docs page",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		index, err := apidocs.LoadIndex()
		if err != nil {
			return err
		}
		doc, ok := findDocument(index, normalizeSlug(args[0]))
		if !ok {
			return fmt.Errorf("docs page %q not found", args[0])
		}
		return outputDocument(doc)
	},
}

func printSections(index apidocs.Index) error {
	fmt.Println("Blue API docs sections:")
	for _, section := range index.Sections {
		count := 0
		for _, doc := range index.Documents {
			if doc.Section == section.Slug {
				count++
			}
		}
		fmt.Printf("  %-28s %s (%d pages)\n", section.Slug, section.Title, count)
	}
	fmt.Println("\nUse 'blue docs list <section>', 'blue docs show <slug>', or 'blue docs search <query>'.")
	return nil
}

func printDocuments(index apidocs.Index, sectionSlug string) error {
	if !hasSection(index, sectionSlug) {
		return fmt.Errorf("docs section %q not found", sectionSlug)
	}
	for _, doc := range index.Documents {
		if doc.Section != sectionSlug {
			continue
		}
		fmt.Printf("%-40s %s\n", doc.Slug, doc.Title)
		if doc.Description != "" {
			fmt.Printf("  %s\n", doc.Description)
		}
	}
	return nil
}

func outputDocument(doc apidocs.Document) error {
	if docsPrintURL {
		fmt.Println(doc.URL)
		return nil
	}
	if docsOpen {
		return openBrowser(doc.URL)
	}
	markdown, err := apidocs.ReadMarkdown(doc)
	if err != nil {
		return err
	}
	fmt.Printf("# %s\n\n", doc.Title)
	if doc.Description != "" {
		fmt.Printf("%s\n\n", doc.Description)
	}
	fmt.Print(markdown)
	if !strings.HasSuffix(markdown, "\n") {
		fmt.Println()
	}
	return nil
}

type searchResult struct {
	doc     apidocs.Document
	score   int
	snippet string
}

func printSearchResults(index apidocs.Index, query string) error {
	query = strings.TrimSpace(query)
	if query == "" {
		return fmt.Errorf("search query is required")
	}

	results := []searchResult{}
	for _, doc := range index.Documents {
		markdown, err := apidocs.ReadMarkdown(doc)
		if err != nil {
			return err
		}
		score, snippet := scoreDocument(doc, markdown, query)
		if score > 0 {
			results = append(results, searchResult{doc: doc, score: score, snippet: snippet})
		}
	}

	sort.Slice(results, func(i, j int) bool {
		if results[i].score != results[j].score {
			return results[i].score > results[j].score
		}
		return results[i].doc.Slug < results[j].doc.Slug
	})

	if docsLimit <= 0 {
		docsLimit = 10
	}
	if len(results) == 0 {
		fmt.Printf("No API docs matched %q.\n", query)
		return nil
	}
	if len(results) < docsLimit {
		docsLimit = len(results)
	}

	for i := 0; i < docsLimit; i++ {
		result := results[i]
		fmt.Printf("%d. %s\n", i+1, result.doc.Title)
		fmt.Printf("   %s\n", result.doc.Slug)
		if result.doc.Description != "" {
			fmt.Printf("   %s\n", result.doc.Description)
		}
		if result.snippet != "" {
			fmt.Printf("   Match: %s\n", result.snippet)
		}
		fmt.Printf("   %s\n\n", result.doc.URL)
	}
	return nil
}

func scoreDocument(doc apidocs.Document, markdown, query string) (int, string) {
	queryLower := strings.ToLower(query)
	terms := tokenize(queryLower)
	score := 0

	title := strings.ToLower(doc.Title)
	description := strings.ToLower(doc.Description)
	headings := strings.ToLower(strings.Join(doc.Headings, " "))
	body := strings.ToLower(markdown)

	if strings.Contains(title, queryLower) {
		score += 100
	}
	if strings.Contains(description, queryLower) {
		score += 70
	}
	if strings.Contains(headings, queryLower) {
		score += 50
	}
	if strings.Contains(body, queryLower) {
		score += 30
	}

	for _, term := range terms {
		if strings.Contains(title, term) {
			score += 20
		}
		if strings.Contains(description, term) {
			score += 12
		}
		if strings.Contains(headings, term) {
			score += 8
		}
		if strings.Contains(body, term) {
			score += 4
		}
	}

	if score == 0 {
		return 0, ""
	}
	return score, snippet(markdown, queryLower, terms)
}

func snippet(markdown, query string, terms []string) string {
	plain := cleanMarkdown(markdown)
	plainLower := strings.ToLower(plain)
	idx := strings.Index(plainLower, query)
	if idx == -1 {
		for _, term := range terms {
			idx = strings.Index(plainLower, term)
			if idx != -1 {
				break
			}
		}
	}
	if idx == -1 {
		return ""
	}
	start := idx - 80
	if start < 0 {
		start = 0
	}
	end := idx + 160
	if end > len(plain) {
		end = len(plain)
	}
	value := strings.TrimSpace(plain[start:end])
	if start > 0 {
		value = "..." + value
	}
	if end < len(plain) {
		value += "..."
	}
	return value
}

func cleanMarkdown(markdown string) string {
	markdown = strings.ReplaceAll(markdown, "`", "")
	markdown = strings.ReplaceAll(markdown, "#", "")
	markdown = strings.ReplaceAll(markdown, "|", " ")
	markdown = strings.ReplaceAll(markdown, "[", "")
	markdown = strings.ReplaceAll(markdown, "]", "")
	markdown = strings.ReplaceAll(markdown, "(", " ")
	markdown = strings.ReplaceAll(markdown, ")", " ")
	return strings.Join(strings.Fields(markdown), " ")
}

func tokenize(value string) []string {
	fields := strings.FieldsFunc(value, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsNumber(r)
	})
	terms := []string{}
	seen := map[string]bool{}
	for _, field := range fields {
		field = strings.ToLower(strings.TrimSpace(field))
		if len(field) < 2 || seen[field] {
			continue
		}
		seen[field] = true
		terms = append(terms, field)
	}
	return terms
}

func findDocument(index apidocs.Index, slug string) (apidocs.Document, bool) {
	slug = normalizeSlug(slug)
	for _, doc := range index.Documents {
		if doc.Slug == slug || strings.TrimPrefix(doc.URL, "https://blue.app/api/") == slug {
			return doc, true
		}
	}
	return apidocs.Document{}, false
}

func hasSection(index apidocs.Index, slug string) bool {
	for _, section := range index.Sections {
		if section.Slug == slug {
			return true
		}
	}
	return false
}

func normalizeSlug(value string) string {
	value = strings.TrimSpace(value)
	if parsed, err := url.Parse(value); err == nil && parsed.Host != "" {
		value = parsed.Path
	}
	value = strings.Trim(value, "/")
	value = strings.TrimPrefix(value, "api/")
	value = strings.TrimSuffix(value, "/index")
	return value
}

func openBrowser(targetURL string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", targetURL)
	case "windows":
		cmd = exec.Command("rundll32", "url.dll,FileProtocolHandler", targetURL)
	default:
		cmd = exec.Command("xdg-open", targetURL)
	}
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to open browser: %w", err)
	}
	return nil
}
