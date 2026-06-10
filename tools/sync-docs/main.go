package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

type section struct {
	Slug  string `json:"slug"`
	Title string `json:"title"`
	Order int    `json:"order"`
}

type document struct {
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

type index struct {
	Source    string     `json:"source"`
	Sections  []section  `json:"sections"`
	Documents []document `json:"documents"`
}

func main() {
	defaultSource := filepath.Clean("../app/src/content/api")
	source := flag.String("source", defaultSource, "Path to app/src/content/api")
	out := flag.String("out", "internal/apidocs/docs", "Output directory for embedded docs")
	flag.Parse()

	if err := run(*source, *out); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(source, out string) error {
	info, err := os.Stat(source)
	if err != nil {
		return fmt.Errorf("source docs not found: %w", err)
	}
	if !info.IsDir() {
		return fmt.Errorf("source docs path is not a directory: %s", source)
	}

	if err := os.RemoveAll(out); err != nil {
		return fmt.Errorf("clear output directory: %w", err)
	}
	if err := os.MkdirAll(out, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}

	sections, err := readSections(source)
	if err != nil {
		return err
	}
	sectionBySlug := map[string]section{}
	for _, section := range sections {
		sectionBySlug[section.Slug] = section
	}

	docs := []document{}
	err = filepath.WalkDir(source, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".md" {
			return nil
		}

		rel, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		frontmatter, body := splitFrontmatter(string(content))
		slug := strings.TrimSuffix(filepath.ToSlash(rel), ".md")
		if strings.HasSuffix(slug, "/index") {
			slug = strings.TrimSuffix(slug, "/index")
		}
		sectionSlug := strings.Split(slug, "/")[0]
		sectionInfo := sectionBySlug[sectionSlug]
		doc := document{
			Slug:         slug,
			Title:        firstNonEmpty(frontmatter["title"], titleFromSlug(slug)),
			Description:  frontmatter["description"],
			Section:      sectionSlug,
			SectionTitle: firstNonEmpty(sectionInfo.Title, titleFromSlug(sectionSlug)),
			Path:         filepath.ToSlash(rel),
			URL:          "https://blue.app/api/" + slug,
			Order:        parseInt(frontmatter["order"], 999),
			Headings:     extractHeadings(body),
		}
		docs = append(docs, doc)

		destination := filepath.Join(out, rel)
		if err := os.MkdirAll(filepath.Dir(destination), 0o755); err != nil {
			return err
		}
		return os.WriteFile(destination, content, 0o644)
	})
	if err != nil {
		return fmt.Errorf("copy docs: %w", err)
	}

	sort.Slice(docs, func(i, j int) bool {
		if docs[i].Section != docs[j].Section {
			return docs[i].Section < docs[j].Section
		}
		if docs[i].Order != docs[j].Order {
			return docs[i].Order < docs[j].Order
		}
		return docs[i].Title < docs[j].Title
	})

	idx := index{
		Source:    source,
		Sections:  sections,
		Documents: docs,
	}
	encoded, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return fmt.Errorf("encode index: %w", err)
	}
	if err := os.WriteFile(filepath.Join(out, "index.json"), append(encoded, '\n'), 0o644); err != nil {
		return fmt.Errorf("write index: %w", err)
	}

	fmt.Printf("Synced %d API docs across %d sections into %s\n", len(docs), len(sections), out)
	return nil
}

func readSections(source string) ([]section, error) {
	sections := []section{}
	entries, err := os.ReadDir(source)
	if err != nil {
		return nil, fmt.Errorf("read docs sections: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		slug := entry.Name()
		metadata := map[string]string{}
		data, err := os.ReadFile(filepath.Join(source, slug, "_dir.yml"))
		if err == nil {
			metadata = parseSimpleYAML(string(data))
		}
		sections = append(sections, section{
			Slug:  slug,
			Title: firstNonEmpty(metadata["title"], titleFromSlug(slug)),
			Order: parseInt(metadata["order"], 999),
		})
	}
	sort.Slice(sections, func(i, j int) bool {
		if sections[i].Order != sections[j].Order {
			return sections[i].Order < sections[j].Order
		}
		return sections[i].Title < sections[j].Title
	})
	for i := range sections {
		if sections[i].Order == 999 {
			sections[i].Order = i
		}
	}
	return sections, nil
}

func splitFrontmatter(markdown string) (map[string]string, string) {
	if !strings.HasPrefix(markdown, "---\n") {
		return map[string]string{}, markdown
	}
	end := strings.Index(markdown[4:], "\n---")
	if end == -1 {
		return map[string]string{}, markdown
	}
	frontmatterEnd := 4 + end
	frontmatter := markdown[4:frontmatterEnd]
	bodyStart := frontmatterEnd + len("\n---")
	return parseSimpleYAML(frontmatter), strings.TrimLeft(markdown[bodyStart:], "\r\n")
}

func parseSimpleYAML(value string) map[string]string {
	result := map[string]string{}
	for _, line := range strings.Split(value, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, raw, ok := strings.Cut(line, ":")
		if !ok {
			continue
		}
		result[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(raw), "\"'")
	}
	return result
}

func extractHeadings(markdown string) []string {
	headings := []string{}
	for _, line := range strings.Split(markdown, "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "#") {
			continue
		}
		text := strings.TrimSpace(strings.TrimLeft(line, "#"))
		if text != "" {
			headings = append(headings, text)
		}
	}
	return headings
}

func titleFromSlug(slug string) string {
	parts := strings.FieldsFunc(slug, func(r rune) bool { return r == '-' || r == '_' || r == '/' })
	for i, part := range parts {
		if part == "api" || part == "id" || part == "url" || part == "csv" || part == "oauth" {
			parts[i] = strings.ToUpper(part)
			continue
		}
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.Join(parts, " ")
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func parseInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}
