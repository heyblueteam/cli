# Blue CLI Roadmap

1. `blue webhooks`
Highest leverage for developers.
Commands:
`list`, `get`, `create`, `update`, `disable`, `delete`, `events`, `verify-signature`, maybe `listen`.
Why: API docs have a full webhook surface, but CLI has none. A CLI can make webhook setup/test loops much faster.

2. `blue exports`
Unify async exports.
Commands:
`records`, `report`, `chart`, `template`, maybe `wait`.
Why: export APIs email results and have progress subscriptions/rate limits. CLI can wrap that complexity and prevent accidental rate-limit failures.

3. `blue reports`
Useful for admins and automation.
Commands:
`list`, `get`, `create`, `update`, `duplicate`, `delete`, `share`, `data`, `aggregate`, `export`.
Why: reports are a whole API section and are currently absent from CLI.

4. `blue documents`
Especially useful if customers generate docs/PDFs from records.
Commands:
`list`, `get`, `create`, `update`, `delete`, `wiki list`, `portable list/create/fields/print`.
Why: the API supports rich documents, wiki pages, and portable PDF templates, but CLI has no surface for them.

5. `blue files upload`
Current CLI has `files download`, but docs heavily cover upload flows.
Commands:
`upload`, `upload-large`, `share`, `url`, `move`.
Why: file upload is a natural CLI use case, especially large files via presigned PUT.

6. `blue saved-views`
Commands:
`list`, `update`, `delete`, maybe `apply`.
Why: saved views are useful for scripting repeatable record filters.

7. `blue domains`
White-label/customer ops tool.
Commands:
`domains list/create/verify/delete`, `smtp list/create/verify/delete`, `email-templates list/update/test`.
Why: great for customer onboarding/support, less broadly used than webhooks/reports.

8. `blue gql`
Generic API escape hatch.
Commands:
`query --file`, `query --raw`, `schema`, `docs`.
Why: lets power users use new API features before a dedicated CLI command exists.

9. `blue doctor`
Diagnostics for credentials and access.
Checks:
auth headers, active company, workspace access, rate limits, API URL, token validity.
Why: would reduce support/debugging friction for CLI users.

10. `blue bootstrap`
Opinionated setup workflows.
Examples:
create workspace from JSON/YAML, create lists/tags/fields/forms/automations in one run, export config from existing workspace.
Why: this is where the CLI can become much more valuable than direct GraphQL CRUD.
