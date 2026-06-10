# Blue CLI Roadmap

1. `blue ids`
Resolve names/slugs to IDs.
Commands:
`workspace`, `field --workspace <id>`, `list --workspace <id>`, `tag --workspace <id>`, `user`, `record --workspace <id>`.
Why: agents constantly need stable IDs before calling API mutations, and humans rarely have them memorized.

2. `blue inspect`
Human/agent-friendly entity resolver.
Commands:
`blue inspect <id-or-url>`.
Why: detect whether an ID or Blue URL is a record, workspace, form, dashboard, report, document, saved view, etc., then print useful metadata.

3. `blue backup`
Export a workspace configuration snapshot.
Commands:
`workspace <id> --output backup.json`.
Why: deeper than `bootstrap export`; include lists, tags, fields, automations, forms, saved views, dashboards, and document metadata.

4. `blue search`
Unified cross-resource search.
Commands:
`blue search "launch" --workspace <id>`.
Why: search records, docs, comments, files, users, tags, and fields from one agent-friendly command.

5. `blue permissions`
Explain effective access.
Commands:
`user <email-or-id> --workspace <id>`, `workspace <id>`.
Why: show role, feature access, and why a user can or cannot do something.

6. `blue open`
Open Blue URLs from IDs.
Commands:
`record <id>`, `workspace <id>`, `form <id>`, `dashboard <id>`, `report <id>`, `document <id>`.
Why: CLI commands return IDs, but humans often need the matching app page.

7. `blue audit`
Workspace/admin audit.
Commands:
`workspace <id>`, `company`.
Why: summarize users, roles, public forms/views, webhooks, automations, integrations, domains, SMTP, and risky configuration.

8. `blue diff`
Compare workspace setup.
Commands:
`workspace <source> <target>`.
Why: useful for template/process standardization and spotting configuration drift.

9. `blue clone`
Clone process configuration.
Commands:
`workspace <source> --name "New workspace"`, with flags to include/exclude records, forms, automations, dashboards, and docs.
Why: faster workspace setup using a known-good process.

10. `blue import`
CLI-friendly import flow.
Commands:
`records --workspace <id> --csv file.csv`.
Why: pairs with `blue exports template`; needs file upload support first.

11. `blue runbook`
Agent-oriented automation recipes.
Commands:
`onboarding --workspace <id>`, `sales-crm --workspace <id>`, `support-queue --workspace <id>`.
Why: create standard lists/tags/fields/forms/automations from named presets.

12. `blue activity`
Audit recent changes.
Commands:
`--workspace <id> --since 7d`, `record <id>`.
Why: useful for support, debugging, and customer success.

13. `blue files upload`
Upload and manage files from the CLI.
Commands:
`upload`, `upload-large`, `share`, `url`, `move`.
Why: file upload is a natural CLI use case, especially large files via presigned PUT. This was intentionally skipped in the first implementation pass.
