package apidocs

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"strings"
)

//go:embed docs/index.json docs/**/*.md
var content embed.FS

type Section struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Order int    `json:"order"`
}

type Document struct {
	Slug         string   `json:"slug"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	Section      string   `json:"section"`
	SectionTitle string   `json:"sectionTitle"`
	Path         string   `json:"path"`
	URL          string   `json:"url"`
	Order        int      `json:"order"`
	Headings     []string `json:"headings"`
}

type Index struct {
	Source    string     `json:"source"`
	Sections  []Section  `json:"sections"`
	Documents []Document `json:"documents"`
}

func LoadIndex() (Index, error) {
	data, err := content.ReadFile("docs/index.json")
	if err != nil {
		return Index{}, fmt.Errorf("read embedded docs index: %w", err)
	}
	var index Index
	if err := json.Unmarshal(data, &index); err != nil {
		return Index{}, fmt.Errorf("parse embedded docs index: %w", err)
	}
	return index, nil
}

func ReadMarkdown(doc Document) (string, error) {
	data, err := content.ReadFile("docs/" + doc.Path)
	if err != nil {
		return "", fmt.Errorf("read embedded doc %s: %w", doc.Slug, err)
	}
	return StripFrontmatter(string(data)), nil
}

func StripFrontmatter(markdown string) string {
	if !strings.HasPrefix(markdown, "---\n") {
		return markdown
	}
	end := strings.Index(markdown[4:], "\n---")
	if end == -1 {
		return markdown
	}
	start := 4 + end + len("\n---")
	return strings.TrimLeft(markdown[start:], "\r\n")
}

func Files() fs.FS {
	docs, err := fs.Sub(content, "docs")
	if err != nil {
		return content
	}
	return docs
}
