# Blue CLI Roadmap

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

5. `blue activity`
Audit recent changes.
Commands:
`blue activity --workspace <id> --since 7d`, `blue activity record <id>`, `blue activity company --since 24h`.
Why: useful for support, debugging, customer success, and answering "what changed?".

6. `blue permissions`
Explain effective access.
Commands:
`blue permissions user <email-or-id> --workspace <id>`, `blue permissions workspace <id>`.
Why: show role, feature access, custom role constraints, and why a user can or cannot do something.

7. `blue describe`
Deep metadata view for a known entity type.
Commands:
`blue describe workspace <id>`, `blue describe record <id> --workspace <id>`, `blue describe form <id> --workspace <id>`.
Why: `inspect` identifies unknown inputs; `describe` gives a richer known-type view, like `kubectl describe`.

8. `blue audit`
Workspace/admin risk summary.
Commands:
`blue audit workspace <id>`, `blue audit company`.
Why: summarize users, roles, public forms/views, webhooks, automations, integrations, domains, SMTP, and risky configuration.

9. `blue context`
Manage default company/workspace context.
Commands:
`blue context list`, `blue context use <company>/<workspace>`, `blue context current`, `blue config set defaultWorkspace <id>`.
Why: reduce repeated `--workspace` flags and improve multi-company workflows.

10. `blue whoami`
Show authenticated identity and active context.
Commands:
`blue whoami`, `blue whoami --format json`.
Why: quick sanity check for scripts and humans before running commands.

11. `blue docs`
Browse Blue API docs from the terminal.
Commands:
`blue docs records`, `blue docs webhooks`, `blue docs custom-fields`, `blue docs search "automation trigger"`.
Why: API docs in terminal are useful for developers and agents.

12. `blue events`
Follow or inspect live/recent workspace events.
Commands:
`blue events --workspace <id> --follow`, `blue events record <id>`, `blue webhooks listen`.
Why: live debugging for integrations, webhooks, comments, records, and files.

13. `blue trigger`
Trigger test events or automation checks.
Commands:
`blue trigger webhook record.created --workspace <id>`, `blue trigger automation <id> --record <id>`.
Why: test webhooks and automations without clicking around in the UI; may require backend support.

14. `blue wait`
Wait for async Blue jobs.
Commands:
`blue wait export <id>`, `blue wait import <id>`, `blue wait report <id>`.
Why: script-friendly handling for exports, imports, reports, and other async operations.

15. `blue usage`
Show company or workspace usage.
Commands:
`blue usage company`, `blue usage workspace <id>`, `blue usage files`, `blue usage records`.
Why: quick admin visibility into users, records, files, storage, and activity volume.

16. `blue alias`
Manage user-defined command aliases.
Commands:
`blue alias set mybugs 'records list --workspace abc --tags bug --done false'`, `blue alias list`, `blue alias delete mybugs`.
Why: make repeated workflows faster and more personal, similar to GitHub CLI aliases.

17. `blue runbook`
Agent-oriented automation recipes.
Commands:
`blue runbook onboarding --workspace <id>`, `blue runbook sales-crm --workspace <id>`, `blue runbook support-queue --workspace <id>`.
Why: create standard lists/tags/fields/forms/automations from named presets.

18. `blue clone`
Clone process configuration.
Commands:
`blue clone workspace <source> --name "New workspace"`, with flags to include/exclude records, forms, automations, dashboards, and docs.
Why: faster workspace setup using a known-good process.

19. `blue import`
CLI-friendly import flow.
Commands:
`blue import records --workspace <id> --csv file.csv`.
Why: pairs with `blue exports template`; needs file upload support first.

20. `blue files upload`
Upload and manage files from the CLI.
Commands:
`blue files upload`, `blue files upload-large`, `blue files share`, `blue files url`, `blue files move`.
Why: file upload is a natural CLI use case, especially large files via presigned PUT. This was intentionally skipped in the first implementation pass.
