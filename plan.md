# Complete Restructuring Plan: Tools Package + Types Centralization

## Final Structure
```
cli/
├── main.go           (router - single entry point)
├── auth.go           (package main - shared auth client)
├── types.go          (package main - all shared types)
├── e2e.go           (package main - end-to-end tests)
├── tools/           (new folder)
│   ├── create_project.go
│   ├── read_projects.go
│   ├── delete_project.go
│   └── ... (all 18 tools)
└── [keep existing docs, go.mod, etc.]
```

## Phase 1: Create Tools Package Structure

### Step 1: Create tools directory ✅
- Create `tools/` folder

### Step 2: Create types.go in root ✅
Create centralized types with smart handling of variations:

**Types to include:**
- **User, Tag** (single version for all uses)
- **PageInfo variations:**
  - `CursorPageInfo` (for GraphQL relay-style)
  - `OffsetPageInfo` (for traditional pagination)
- **TodoList variations:**
  - `TodoList` (full version)
  - `TodoListSimple` (just ID, Title, Todos)
- **Project, Record, CustomField** (comprehensive versions)
- **Common response types**

## Phase 2: Migrate Tools to Package

### Step 3: Convert each tool file (18 files) ⬜
For each tool file:
1. **Copy** to `tools/` with underscores (e.g., `create_project.go`)
2. **Change package** from `package main` to `package tools`
3. **Convert main()** to exported function:
   - `func main()` → `func RunCreateProject(args []string)`
4. **Update imports:**
   - Remove local type definitions
   - Import types from parent package
   - Keep auth client import
5. **Move flag parsing** inside the Run function
6. **Return error** instead of log.Fatal

Example transformation:
```go
// Before (create-project.go)
package main
func main() {
    flag.Parse()
    // ... logic
}

// After (tools/create_project.go)
package tools
import "github.com/blue/cli"
func RunCreateProject(args []string) error {
    fs := flag.NewFlagSet("create-project", flag.ExitOnError)
    // ... parse args
    fs.Parse(args)
    // ... logic (return errors)
}
```

**Tools to migrate:**
- ⬜ create-project.go → create_project.go
- ✅ read-projects.go → read_projects.go
- ⬜ delete-project.go → delete_project.go
- ⬜ create-list.go → create_list.go
- ⬜ read-lists.go → read_lists.go
- ⬜ create-tags.go → create_tags.go
- ⬜ read-tags.go → read_tags.go
- ⬜ create-custom-field.go → create_custom_field.go
- ⬜ read-project-custom-fields.go → read_project_custom_fields.go
- ⬜ create-record.go → create_record.go
- ⬜ read-todos.go → read_todos.go
- ⬜ read-project-todos.go → read_project_todos.go
- ⬜ read-records.go → read_records.go
- ⬜ read-records-count.go → read_records_count.go
- ⬜ create-record-tags.go → create_record_tags.go
- ⬜ update-project.go → update_project.go
- ⬜ delete-record.go → delete_record.go

## Phase 3: Create Main Router

### Step 4: Create main.go router ⬜
```go
package main

import (
    "os"
    "fmt"
    "github.com/blue/cli/tools"
)

func main() {
    if len(os.Args) < 2 {
        printUsage()
        os.Exit(1)
    }

    command := os.Args[1]
    args := os.Args[2:]

    var err error
    switch command {
    case "create-project":
        err = tools.RunCreateProject(args)
    case "read-projects":
        err = tools.RunReadProjects(args)
    // ... all 18 commands
    case "e2e":
        err = runE2E(args)
    default:
        fmt.Printf("Unknown command: %s\n", command)
        printUsage()
        os.Exit(1)
    }

    if err != nil {
        fmt.Fprintf(os.Stderr, "Error: %v\n", err)
        os.Exit(1)
    }
}
```

## Phase 4: Update E2E Test

### Step 5: Modify e2e.go ⬜
- Keep in root as `package main`
- Update to use new command structure:
  - Change: `exec.Command("go", "run", "auth.go", "create-project.go", ...)`
  - To: `exec.Command("go", "run", ".", "create-project", ...)`

## Phase 5: Clean Up

### Step 6: Remove old files ⬜
- Delete the 18 original tool files from root
- Keep only: main.go, auth.go, types.go, e2e.go, docs, configs

### Step 7: Update documentation ⬜
- Update README.md with new usage:
  ```bash
  # Old: go run auth.go create-project.go -name "Demo"
  # New: go run . create-project -name "Demo"
  ```
- Update CLAUDE.md similarly

## Phase 6: Testing

### Step 8: Validate everything works ⬜
1. Run `go build .` - should create single binary
2. Test a few commands:
   - `go run . create-project -name "Test"`
   - `go run . read-projects`
3. Run full e2e test: `go run . e2e`

## Benefits of This Approach

1. **Fixes "main redeclared" errors** - only one main() function
2. **Centralizes types** - no more duplicates
3. **Cleaner usage** - `go run . <command>` instead of multiple files
4. **Single binary** - can build and distribute one executable
5. **Better organization** - clear separation of concerns
6. **Maintains backwards compatibility** - same flags, same functionality
7. **Easier to extend** - add new tools by adding to tools/ and updating router

## Migration Order (minimize risk)

1. **First:** Create types.go and tools/ folder
2. **Second:** Migrate one simple read tool as proof of concept
3. **Third:** Create basic main.go router with just that one tool
4. **Fourth:** Test that one tool works
5. **Fifth:** Migrate remaining tools in batches
6. **Last:** Update e2e test and documentation

## Total files to modify/create:
- **Create:** 3 new files (main.go, types.go, tools/ folder)
- **Migrate:** 18 tool files
- **Update:** 3 files (e2e.go, README.md, CLAUDE.md)
- **Delete:** 18 old tool files from root