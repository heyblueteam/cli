---
title: Export CSV Template
description: Download a pre-formatted CSV template with the correct column headers for bulk-importing records into a workspace.
icon: FileSpreadsheet
order: 2
---

Use the `exportCSVTemplate` mutation to generate a blank CSV file pre-populated with the correct column headers for a workspace — including a column for every custom field. Fill in the template and upload it with `importTodos` to bulk-create records. Workspaces are `Project` objects and records are `Todo` objects in the API.

The mutation returns a signed download URL as a plain `String!`. The URL forces a browser download (via `Content-Disposition: attachment`) and expires after 24 hours.

## Request

```graphql
mutation ExportTemplate {
  exportCSVTemplate(input: { projectId: "project_123" })
}
```

The `projectId` argument accepts a workspace ID or slug.

## Parameters

### ExportCSVTemplateInput

| Parameter   | Type      | Required | Description                                               |
| ----------- | --------- | -------- | --------------------------------------------------------- |
| `projectId` | `String!` | Yes      | ID or slug of the workspace to generate the template for. |

## Response

`exportCSVTemplate` returns a signed download URL as a single `String!`. Opening it downloads the file directly because the URL is signed with `Content-Disposition: attachment`.

```json
{
  "data": {
    "exportCSVTemplate": "https://api.blue.app/storage/acme/marketing-plan/marketing-plan-template-301520032025-a1b2c3.csv?X-Amz-Expires=86400&X-Amz-Signature=..."
  }
}
```

The filename is derived deterministically from the workspace name and a timestamp: `slugify("<workspace name> Template <mmHHDDMMYYYY>") + ".csv"`. For a workspace named "Marketing Plan", a template generated on 20 March 2025 produces a name like `marketing-plan-template-301520032025.csv`.

### Returns

| Field               | Type      | Description                                                                                                                          |
| ------------------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `exportCSVTemplate` | `String!` | A signed download URL for the generated CSV template. The URL expires 24 hours after it is issued and forces a download when opened. |

## Template columns

The generated CSV contains a single header row followed by one empty sample row.

| Column                 | Description                                                                                                                                                                             |
| ---------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `Title`                | Record title.                                                                                                                                                                           |
| `List`                 | Name of the list to place the record in.                                                                                                                                                |
| `Done`                 | Completion status (`true` / `false`).                                                                                                                                                   |
| `Start Date`           | Record start date.                                                                                                                                                                      |
| `Due Date`             | Record due date.                                                                                                                                                                        |
| `Description`          | Record description.                                                                                                                                                                     |
| `Assignees`            | Comma-separated assignee emails.                                                                                                                                                        |
| `Created At`           | Original creation timestamp.                                                                                                                                                            |
| `Updated At`           | Original last-updated timestamp.                                                                                                                                                        |
| `Created By`           | Original creator name or email.                                                                                                                                                         |
| `Color`                | Record color.                                                                                                                                                                           |
| `Project`              | Workspace name.                                                                                                                                                                         |
| `Tags`                 | Comma-separated tag titles.                                                                                                                                                             |
| _Custom field columns_ | One column per custom field in the workspace, ordered by position. A currency-type field adds a `Currency Types` column immediately after it for the currency code (e.g. `USD`, `EUR`). |

<Callout variant="info" title="The sample row spans only the base columns">

The empty sample row underneath the header covers the 13 base columns above. When the workspace has custom fields, that row is shorter than the full header — it's only a formatting hint, not a per-column placeholder. Add your own values across all columns when you fill in the template.

</Callout>

## Errors

| Code                | When                                           |
| ------------------- | ---------------------------------------------- |
| `PROJECT_NOT_FOUND` | No workspace exists for the given `projectId`. |

```json
{
  "errors": [
    {
      "message": "Project was not found.",
      "extensions": { "code": "PROJECT_NOT_FOUND" }
    }
  ]
}
```

## Permissions

Requires project-level `OWNER`, `ADMIN`, or `MEMBER` access, and the workspace must be active.

| Access level   | Can export template |
| -------------- | ------------------- |
| `OWNER`        | Yes                 |
| `ADMIN`        | Yes                 |
| `MEMBER`       | Yes                 |
| `CLIENT`       | No                  |
| `COMMENT_ONLY` | No                  |
| `VIEW_ONLY`    | No                  |

## Bulk-import workflow

1. Call `exportCSVTemplate` and download the template from the returned URL.
2. Fill in record data, one row per record.
3. Upload the completed CSV with the [file upload API](/api/files) to obtain an `s3Key`.
4. Call `importTodos` with the `s3Key`, the `headers` mapping, and the `projectId`.
5. Subscribe to `subscribeToImportExportProgress` to track import progress.

## Related

- [Export Records](/api/import-export/export-records)
- [Uploading Files](/api/files)
- [Error Codes](/api/start-guide/error-codes)
