// Command sync-schema refreshes the vendored copy of the Blue GraphQL schema
// from the api checkout.
//
// The runtime schema the API actually serves is not any single file: api's
// lib/schema.ts merges src/schema.graphql with two generated artifacts
// (aliases.graphql, which carries the Record/Workspace/Organization vocabulary,
// and prisma.graphql, which carries the filter and where inputs). Vendoring
// only src/schema.graphql would drop the majority of the types the CLI's own
// queries use, so this concatenates the same three sources in the same order.
//
// The vendored copy is what `blue api schema` prints for users without a
// monorepo checkout. Nothing consumes it at build time, so nothing catches it
// going stale — hence this explicit step, run via `make schema`.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// sources are joined in the order api's lib/schema.ts merges them.
var sources = []string{
	"src/schema.graphql",
	"src/generated/aliases.graphql",
	"src/generated/prisma.graphql",
}

func main() {
	defaultAPI := filepath.Clean("../api")
	apiDir := flag.String("api", defaultAPI, "Path to the api checkout")
	out := flag.String("out", "schema.graphql", "Output path for the vendored schema")
	check := flag.Bool("check", false, "Exit non-zero if the vendored schema is out of date, without writing")
	flag.Parse()

	if err := run(*apiDir, *out, *check); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(apiDir, out string, check bool) error {
	var merged bytes.Buffer
	for _, source := range sources {
		path := filepath.Join(apiDir, source)
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read %s: %w", path, err)
		}
		if len(data) == 0 {
			return fmt.Errorf("%s is empty — run `npm run generate` in the api checkout", path)
		}
		fmt.Fprintf(&merged, "# --- %s ---\n\n", source)
		merged.Write(data)
		merged.WriteString("\n")
	}

	current, err := os.ReadFile(out)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("read vendored schema: %w", err)
	}

	if bytes.Equal(current, merged.Bytes()) {
		fmt.Printf("schema.graphql is up to date (%d bytes)\n", merged.Len())
		return nil
	}

	if check {
		return fmt.Errorf("schema.graphql is out of date with %s — run `make schema`", apiDir)
	}

	if err := os.WriteFile(out, merged.Bytes(), 0o644); err != nil {
		return fmt.Errorf("write vendored schema: %w", err)
	}

	fmt.Printf("schema.graphql updated from %s (%d bytes)\n", apiDir, merged.Len())
	return nil
}
