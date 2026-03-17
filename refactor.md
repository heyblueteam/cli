# Blue CLI Refactor Plan

## Goal

Transform the Blue CLI from a flat `go run . <command>` tool into a proper, distributable CLI with nested subcommands, auto-generated help, shell completions, and professional packaging.

**Binary name:** `blue`
**Terminology change:** "projects" becomes **"workspaces"** throughout the entire CLI.

---

## Phase 1: Cobra Setup & Project Structure

### 1.1 Add Cobra dependency

```bash
go get github.com/spf13/cobra@latest
```

### 1.2 New directory structure

```
cli/
├── main.go                        # Minimal: calls cmd.Execute()
├── cmd/
│   ├── root.go                    # Root command, global flags, config loading
│   ├── version.go                 # blue version
│   ├── workspaces/
│   │   ├── workspaces.go          # blue workspaces (parent)
│   │   ├── list.go                # blue workspaces list
│   │   ├── create.go              # blue workspaces create
│   │   ├── update.go              # blue workspaces update
│   │   ├── delete.go              # blue workspaces delete
│   ├── records/
│   │   ├── records.go             # blue records (parent)
│   │   ├── list.go                # blue records list
│   │   ├── get.go                 # blue records get
│   │   ├── create.go              # blue records create
│   │   ├── update.go              # blue records update
│   │   ├── delete.go              # blue records delete
│   │   ├── move.go                # blue records move
│   │   ├── count.go               # blue records count
│   ├── lists/
│   │   ├── lists.go               # blue lists (parent)
│   │   ├── list.go                # blue lists list
│   │   ├── create.go              # blue lists create
│   │   ├── update.go              # blue lists update
│   │   ├── delete.go              # blue lists delete
│   ├── tags/
│   │   ├── tags.go                # blue tags (parent)
│   │   ├── list.go                # blue tags list
│   │   ├── create.go              # blue tags create
│   │   ├── add.go                 # blue tags add (add tags to record)
│   ├── fields/
│   │   ├── fields.go              # blue fields (parent)
│   │   ├── list.go                # blue fields list
│   │   ├── create.go              # blue fields create
│   │   ├── update.go              # blue fields update
│   │   ├── delete.go              # blue fields delete
│   │   ├── options/
│   │   │   ├── options.go         # blue fields options (parent)
│   │   │   ├── create.go          # blue fields options create
│   │   │   ├── delete.go          # blue fields options delete
│   │   ├── groups/
│   │   │   ├── groups.go          # blue fields groups (parent)
│   │   │   ├── list.go            # blue fields groups list
│   │   │   ├── manage.go          # blue fields groups manage
│   ├── automations/
│   │   ├── automations.go         # blue automations (parent)
│   │   ├── list.go                # blue automations list
│   │   ├── create.go              # blue automations create
│   │   ├── update.go              # blue automations update
│   │   ├── delete.go              # blue automations delete
│   ├── checklists/
│   │   ├── checklists.go          # blue checklists (parent)
│   │   ├── list.go                # blue checklists list
│   │   ├── create.go              # blue checklists create
│   │   ├── delete.go              # blue checklists delete
│   │   ├── items/
│   │   │   ├── items.go           # blue checklists items (parent)
│   │   │   ├── create.go          # blue checklists items create
│   │   │   ├── update.go          # blue checklists items update
│   │   │   ├── delete.go          # blue checklists items delete
│   ├── comments/
│   │   ├── comments.go            # blue comments (parent)
│   │   ├── create.go              # blue comments create
│   │   ├── update.go              # blue comments update
│   ├── users/
│   │   ├── users.go               # blue users (parent)
│   │   ├── list.go                # blue users list
│   │   ├── invite.go              # blue users invite
│   │   ├── roles.go               # blue users roles
│   ├── dependencies/
│   │   ├── dependencies.go        # blue dependencies (parent)
│   │   ├── create.go              # blue dependencies create
│   │   ├── update.go              # blue dependencies update
│   │   ├── delete.go              # blue dependencies delete
│   ├── files/
│   │   ├── files.go               # blue files (parent)
│   │   ├── download.go            # blue files download
├── common/                        # Stays as-is (auth.go, types.go, utils.go)
├── tools/                         # KEEP during migration, delete when done
├── test/
│   └── e2e.go                     # Update to use new command structure
```

### 1.3 Root command (`cmd/root.go`)

Global persistent flags available to all subcommands:
- `--workspace` / `-w` — workspace ID or slug (replaces per-command `-project`)
- `--format` — output format: `table` (default), `json`, `csv`
- `--simple` / `-s` — simplified output
- `--config` — path to config file (default: `.env`)

```
blue [command] [subcommand] [flags]

Available Commands:
  workspaces   Manage workspaces
  records      Manage records
  lists        Manage lists
  tags         Manage tags
  fields       Manage custom fields
  automations  Manage automations
  checklists   Manage checklists
  comments     Manage comments
  users        Manage users
  dependencies Manage record dependencies
  files        Manage files
  version      Print version information
  completion   Generate shell completions

Flags:
  -w, --workspace string   Workspace ID or slug
  -s, --simple             Simplified output
      --format string      Output format: table, json, csv (default "table")
      --config string      Config file path (default ".env")
  -h, --help               Help for blue
```

---

## Phase 2: Command Migration

### 2.1 Full command mapping (old → new)

Each existing `tools.RunXxx()` function becomes a cobra command's `RunE` handler. The business logic inside each function stays identical — only the flag wiring changes.

| Old Command | New Command | Tool File |
|---|---|---|
| `read-projects` | `blue workspaces list` | read_projects.go |
| `create-project` | `blue workspaces create` | create_project.go |
| `update-project` | `blue workspaces update` | update_project.go |
| `delete-project` | `blue workspaces delete` | delete_project.go |
| `read-lists` | `blue lists list` | read_lists.go |
| `create-list` | `blue lists create` | create_list.go |
| `update-list` | `blue lists update` | update_list.go |
| `delete-list` | `blue lists delete` | delete_list.go |
| `read-record` | `blue records get` | read_record.go |
| `read-records` | `blue records list` | read_records.go |
| `read-list-records` | `blue records list --list <id>` | read_list_records.go |
| `read-project-records` | `blue records list --all` | read_project_records.go |
| `read-records-count` | `blue records count` | read_records_count.go |
| `create-record` | `blue records create` | create_record.go |
| `update-record` | `blue records update` | update_record.go |
| `delete-record` | `blue records delete` | delete_record.go |
| `move-record` | `blue records move` | move_record.go |
| `read-tags` | `blue tags list` | read_tags.go |
| `create-tags` | `blue tags create` | create_tags.go |
| `create-record-tags` | `blue tags add` | create_record_tags.go |
| `read-project-custom-fields` | `blue fields list` | read_project_custom_fields.go |
| `read-custom-fields` | `blue fields list --detailed` | read_custom_fields.go |
| `create-custom-field` | `blue fields create` | create_custom_field.go |
| `update-custom-field` | `blue fields update` | update_custom_field.go |
| `delete-custom-field` | `blue fields delete` | delete_custom_field.go |
| `create-custom-field-options` | `blue fields options create` | create_custom_field_options.go |
| `delete-custom-field-options` | `blue fields options delete` | delete_custom_field_options.go |
| `read-field-groups` | `blue fields groups list` | read_custom_field_groups.go |
| `manage-field-groups` | `blue fields groups manage` | manage_custom_field_groups.go |
| `read-automations` | `blue automations list` | read_automations.go |
| `create-automation` | `blue automations create` | create_automation.go |
| `create-automation-multi` | `blue automations create --multi` | create_automation_multi.go |
| `update-automation` | `blue automations update` | update_automation.go |
| `update-automation-multi` | `blue automations update --multi` | update_automation_multi.go |
| `delete-automation` | `blue automations delete` | delete_automation.go |
| `read-checklists` | `blue checklists list` | read_checklists.go |
| `create-checklist` | `blue checklists create` | create_checklist.go |
| `delete-checklist` | `blue checklists delete` | delete_checklist.go |
| `create-checklist-item` | `blue checklists items create` | create_checklist_item.go |
| `update-checklist-item` | `blue checklists items update` | update_checklist_item.go |
| `delete-checklist-item` | `blue checklists items delete` | delete_checklist_item.go |
| `create-comment` | `blue comments create` | create_comment.go |
| `update-comment` | `blue comments update` | update_comment.go |
| `read-user-profiles` | `blue users list` | read_user_profiles.go |
| `invite-user` | `blue users invite` | invite_user.go |
| `read-project-user-roles` | `blue users roles` | read_project_user_roles.go |
| `create-dependency` | `blue dependencies create` | create_dependency.go |
| `update-dependency` | `blue dependencies update` | update_dependency.go |
| `delete-dependency` | `blue dependencies delete` | delete_dependency.go |
| `download-files` | `blue files download` | download_files.go |
| `test-custom-fields` | (remove or move to `blue fields test`) | test_custom_fields.go |
| `e2e` | `blue test` (hidden/dev-only) | test/e2e.go |

### 2.2 Migration strategy

Migrate one command group at a time. Order:

1. **Root + workspaces** — smallest group, proves the pattern
2. **lists** — simple, few commands
3. **records** — largest group, most flags
4. **tags** — small
5. **fields** (including options + groups subcommands)
6. **automations**
7. **checklists** (including items subcommand)
8. **comments**
9. **users**
10. **dependencies**
11. **files**

For each group:
1. Create the `cmd/<group>/` directory and parent command
2. Create each subcommand, wiring cobra flags to the existing `tools.RunXxx()` logic
3. Verify `blue <group> <subcommand> --help` works
4. Test the command actually runs against the API
5. Remove the old case from `main.go`

### 2.3 Flag conventions (cobra style)

| Old (stdlib flag) | New (cobra) | Notes |
|---|---|---|
| `-project` | `--workspace` / `-w` | Global persistent flag, renamed |
| `-simple` | `--simple` / `-s` | Global persistent flag |
| `-record` | `--record` / `-r` | Per-command |
| `-list` | `--list` / `-l` | Per-command |
| `-confirm` | `--confirm` / `-y` | Per-command (like -y for yes) |
| `-page` | `--page` | Per-command |
| `-size` | `--size` | Per-command |
| `-title` | `--title` / `-t` | Per-command |

Single-dash flags (`-project`) become double-dash (`--workspace`). Short flags added where useful.

---

## Phase 3: Terminology Rename (projects → workspaces)

### 3.1 User-facing changes
- All command names: `workspaces` not `projects`
- All flag names: `--workspace` not `--project`
- All help text, descriptions, error messages
- All output text ("Workspace created", "Workspace ID:", etc.)

### 3.2 Internal code
- **DO NOT rename** GraphQL queries/mutations — the API still uses "project"
- **DO NOT rename** the `common.Project` type or API field names
- Only rename user-facing strings (CLI flags, help text, printed output)
- Add a comment where the mapping happens: `// "workspace" in CLI = "project" in API`

### 3.3 Header mapping
- `X-Bloo-Project-Id` header stays as-is (it's an API contract)
- The `--workspace` flag value gets passed to `client.SetProject()`

---

## Phase 4: Build & Distribution

### 4.1 Module rename

Change `go.mod` module from `demo-builder` to `github.com/user/blue-cli` (or appropriate path). Update all imports.

### 4.2 GoReleaser setup

Create `.goreleaser.yml`:

```yaml
project_name: blue
builds:
  - main: ./main.go
    binary: blue
    env:
      - CGO_ENABLED=0
    goos:
      - darwin
      - linux
      - windows
    goarch:
      - amd64
      - arm64
    ldflags:
      - -s -w -X main.version={{.Version}} -X main.commit={{.Commit}}
archives:
  - format: tar.gz
    name_template: "blue_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip
brews:
  - repository:
      owner: <github-org>
      name: homebrew-tap
    homepage: https://blue.cc
    description: CLI for Blue workspace management
    install: |
      bin.install "blue"
checksum:
  name_template: "checksums.txt"
release:
  github:
    owner: <github-org>
    name: blue-cli
```

### 4.3 GitHub Actions CI/CD

Create `.github/workflows/release.yml`:

```yaml
name: Release
on:
  push:
    tags: ["v*"]
jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - uses: goreleaser/goreleaser-action@v5
        with:
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### 4.4 Distribution channels (priority order)

1. **GitHub Releases** (automatic via GoReleaser) — cross-platform binaries
2. **Homebrew tap** (automatic via GoReleaser) — `brew install <org>/tap/blue`
3. **curl install script** (optional, later) — `curl -fsSL https://cli.blue.cc/install.sh | sh`

**NOT recommended:**
- npm — adds unnecessary Node.js dependency for a Go binary
- Snap/Flatpak — overkill for a CLI tool

### 4.5 Version command

```
$ blue version
blue v1.0.0 (commit: abc1234, built: 2026-03-17)
```

Inject version at build time via ldflags.

---

## Phase 5: Auth & Config Improvements

### 5.1 Config file location

Move from `.env` in the project directory to a proper config location:

```
~/.config/blue/config.yaml    # Linux/Mac (XDG)
~/.blue/config.yaml            # Fallback
.env                           # Still supported for backwards compat
```

### 5.2 Login command (future)

```bash
blue login                     # Interactive setup
blue login --token <token> --client-id <id> --company <slug>
```

Stores credentials in `~/.config/blue/config.yaml`. This is a nice-to-have for v2.

### 5.3 Environment variable override

Env vars should still work and take precedence over config file:
```
BLUE_API_URL, BLUE_AUTH_TOKEN, BLUE_CLIENT_ID, BLUE_COMPANY_ID
```

---

## Phase 6: Polish

### 6.1 Shell completions (free with cobra)

```bash
blue completion bash > /etc/bash_completion.d/blue
blue completion zsh > "${fpath[1]}/_blue"
blue completion fish > ~/.config/fish/completions/blue.fish
blue completion powershell > blue.ps1
```

### 6.2 Aliases

Register common aliases in cobra:
- `blue ws` → `blue workspaces`
- `blue rec` → `blue records`
- `blue auto` → `blue automations`

### 6.3 JSON output mode

All commands should respect `--format json` for machine-readable output. This is critical for agent/automation use cases.

### 6.4 Exit codes

Standardize:
- `0` — success
- `1` — general error
- `2` — usage error (bad flags, missing args)

---

## Documentation Updates Required

The following files MUST be updated after the refactor. These are critical because **agents and automations are the primary users of this CLI**, not just humans.

### Files to update:

| File | What to change |
|---|---|
| `README.md` | Full rewrite — new install instructions (`brew install`, binary download), new command syntax (`blue workspaces list`), new examples. Remove all `go run .` references. |
| `CLAUDE.md` | Full rewrite — this is the primary reference for AI agents. Every command example must use new syntax. Update project structure, development commands, all usage examples. Rename all "project" references to "workspace" in user-facing text. |
| `CUSTOM_FIELDS_README.md` | Update all command examples to new syntax. |
| `implementation-status.md` | Either delete (outdated) or rewrite to reflect current state. |
| Root `../CLAUDE.md` | Update the `cli/` section if it references old command patterns. |

### Agent/automation considerations:

- The CLI is primarily used by AI agents (Claude, etc.) via CLAUDE.md instructions
- All command examples in CLAUDE.md must be copy-pasteable with the new `blue` binary
- JSON output (`--format json`) should be documented prominently for agent consumption
- Error messages should be parseable (structured) not just human-readable
- Consider adding a `blue` man page or `blue help <topic>` for extended docs that agents can reference

---

## Effort Estimate

| Phase | Scope | Effort |
|---|---|---|
| Phase 1: Cobra setup | Root command, directory structure | Small |
| Phase 2: Command migration | 50 commands, mechanical refactor | Medium-Large (bulk of the work) |
| Phase 3: Terminology rename | Search-and-replace in user-facing strings | Small |
| Phase 4: Build & distribution | GoReleaser, CI/CD, Homebrew | Small-Medium |
| Phase 5: Auth improvements | Config file, env var standardization | Small |
| Phase 6: Polish | Completions, aliases, JSON output | Small |
| Documentation | README, CLAUDE.md, CUSTOM_FIELDS_README | Medium |

**Total: Medium project.** The migration is mechanical — each command is a self-contained function that just needs its flag wiring changed from stdlib `flag` to cobra. No business logic changes. The hardest part is the sheer number of commands (50), but each one is a 10-15 minute conversion.

---

## Migration Checklist

- [x] Add cobra dependency
- [x] Create `cmd/root.go` with global flags
- [x] Migrate workspaces (projects) commands — tested: list, create, update, delete
- [x] Migrate lists commands — tested: list, create, update, delete
- [x] Migrate records commands — tested: list, get, create, update, delete, move, count
- [x] Migrate tags commands — tested: list, create, add
- [x] Migrate fields commands (including options + groups) — tested: list, create
- [x] Migrate automations commands — tested: list
- [x] Migrate checklists commands (including items) — tested: list, create, items create/update
- [x] Migrate comments commands — tested: create, update
- [x] Migrate users commands — tested: list
- [x] Migrate dependencies commands — created, help verified
- [x] Migrate files commands — created, help verified
- [x] Remove old `main.go` switch statement — replaced with cobra
- [x] Rename module from `demo-builder` to `blue-cli`
- [x] Add version injection via ldflags
- [x] Set up `.goreleaser.yml`
- [x] Set up GitHub Actions release workflow
- [x] Add `blue completion` command — built-in with cobra
- [x] Add command aliases (ws, rec, auto, deps)
- [ ] Ensure `--format json` works on all commands
- [ ] Rewrite `README.md`
- [ ] Rewrite `CLAUDE.md`
- [ ] Update `CUSTOM_FIELDS_README.md`
- [ ] Delete or rewrite `implementation-status.md`
- [ ] End-to-end test with new command structure
- [ ] Tag and release v1.0.0
