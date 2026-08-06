# Blue CLI Roadmap

## v1.0 Readiness

`v1.0` should mean the CLI is stable, trustworthy, and polished enough to treat command names, config, and output shapes as supported interfaces.

Must-have before v1:

- Stable command naming and aliases; avoid further breaking renames after v1.
- Stable config format for `~/.config/blue/config.env`.
- Strong `--help` coverage for every major command, with examples and required flags.
- Verified release pipeline for GitHub Releases, Homebrew, and Scoop.
- `blue doctor` covers common auth, config, API, and workspace-context failures.
- `blue docs` and `blue api` stay reliable because they make the CLI self-serve.
- README covers install, auth, common workflows, shell completions, and troubleshooting.
- Production smoke tests for auth/init, workspace and record basics, raw GraphQL, docs search, activity, exports, automations, webhooks, forms, documents, and reports.
- Consistent output behavior across major commands: `--format json`, `--simple`, errors, and exit codes.

Strong candidates before v1:

- `blue context` to manage default company/workspace and reduce repeated `--workspace` flags.
- Generated CLI command docs from Cobra help, either as Markdown files or `blue help markdown`.
- Better automation ergonomics, especially `blue automations create --file automation.json` for HTTP and multi-action automations.
- `blue inspect` for pasting any Blue ID or URL and identifying the entity.
- `blue describe` for rich metadata on known entity types.
- `blue wait` for script-friendly async job polling.

Do not block v1 on large workflow features like `blue backup`, `blue apply`, `blue clone`, `blue import`, or `blue files upload`; those can ship in `v1.x`.

Recommended path:

1. `v0.10`: `blue context`, generated CLI docs, automation `--file` support.
2. `v0.11`: `blue inspect`, `blue describe`, output consistency cleanup.
3. `v1.0`: stabilization release once the above feels solid.

## Feature Backlog

1. `blue inspect`
Human/agent-friendly entity resolver.
Commands:
`blue inspect <id-or-url>`, `blue inspect <id-or-url> --format json`.
Why: detect whether an ID or Blue URL is a record, workspace, form, dashboard, report, document, saved view, file, or user, then print useful metadata and next commands.

2. `blue backup`
Export a workspace configuration snapshot.
Commands:
`blue backup workspace <id> --output backup.json`, `blue backup workspace <id> --include-doc-content`, `blue backup workspace <id> --include-records`.
Why: deeper than `bootstrap export`; include lists, tags, fields, automations, forms, saved views, dashboards, reports, and document metadata/content.

3. `blue diff`
Compare workspace setup or saved snapshots.
Commands:
`blue diff workspace <source> <target>`, `blue diff backup backup-a.json backup-b.json`.
Why: useful for template/process standardization and spotting configuration drift.

4. `blue apply`
Apply declarative workspace configuration.
Commands:
`blue apply -f workspace-config.json --dry-run`, `blue apply -f workspace-config.json --workspace <id>`.
Why: pairs with `blue backup` and `blue diff` to make workspace setup repeatable.

5. `blue permissions`
Explain effective access.
Commands:
`blue permissions user <email-or-id> --workspace <id>`, `blue permissions workspace <id>`.
Why: show role, feature access, custom role constraints, and why a user can or cannot do something.

6. `blue describe`
Deep metadata view for a known entity type.
Commands:
`blue describe workspace <id>`, `blue describe record <id> --workspace <id>`, `blue describe form <id> --workspace <id>`.
Why: `inspect` identifies unknown inputs; `describe` gives a richer known-type view, like `kubectl describe`.

7. `blue audit`
Workspace/admin risk summary.
Commands:
`blue audit workspace <id>`, `blue audit company`.
Why: summarize users, roles, public forms/views, webhooks, automations, integrations, domains, SMTP, and risky configuration.

8. `blue events`
Follow or inspect live/recent workspace events.
Commands:
`blue events --workspace <id> --follow`, `blue events record <id>`, `blue webhooks listen`.
Why: live debugging for integrations, webhooks, comments, records, and files.

9. `blue trigger`
Trigger test events or automation checks.
Commands:
`blue trigger webhook record.created --workspace <id>`, `blue trigger automation <id> --record <id>`.
Why: test webhooks and automations without clicking around in the UI; may require backend support.

10. `blue wait`
Wait for async Blue jobs.
Commands:
`blue wait export <id>`, `blue wait import <id>`, `blue wait report <id>`.
Why: script-friendly handling for exports, imports, reports, and other async operations.

11. `blue usage`
Show company or workspace usage.
Commands:
`blue usage company`, `blue usage workspace <id>`, `blue usage files`, `blue usage records`.
Why: quick admin visibility into users, records, files, storage, and activity volume.

12. `blue alias`
Manage user-defined command aliases.
Commands:
`blue alias set mybugs 'records list --workspace abc --tags bug --done false'`, `blue alias list`, `blue alias delete mybugs`.
Why: make repeated workflows faster and more personal, similar to GitHub CLI aliases.

13. `blue runbook`
Agent-oriented automation recipes.
Commands:
`blue runbook onboarding --workspace <id>`, `blue runbook sales-crm --workspace <id>`, `blue runbook support-queue --workspace <id>`.
Why: create standard lists/tags/fields/forms/automations from named presets.

14. `blue clone`
Clone process configuration.
Commands:
`blue clone workspace <source> --name "New workspace"`, with flags to include/exclude records, forms, automations, dashboards, and docs.
Why: faster workspace setup using a known-good process.

15. `blue import`
CLI-friendly import flow.
Commands:
`blue import records --workspace <id> --csv file.csv`.
Why: pairs with `blue exports template`; needs file upload support first.

16. `blue files upload`
Upload and manage files from the CLI.
Commands:
`blue files upload`, `blue files upload-large`, `blue files share`, `blue files url`, `blue files move`.
Why: file upload is a natural CLI use case, especially large files via presigned PUT. This was intentionally skipped in the first implementation pass.
