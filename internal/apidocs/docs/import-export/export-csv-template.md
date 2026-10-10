---
title: Export CSV Template
description: Download a pre-formatted CSV template with the correct column headers for bulk-importing records into a workspace.
icon: FileSpreadsheet
order: 2
---

Use the `exportCSVTemplate` mutation to generate a blank CSV file pre-populated with the correct column headers for a workspace — including a column for every custom field. Fill in the template and upload it with `importRecords` to bulk-create records. Workspaces are `Workspace` objects and records are `Record` objects in the API.

The mutation returns a signed download URL as a plain `String!`. The URL forces a browser download (via `Content-Disposition: attachment`) and expires after 24 hours.

## Request

```graphql
mutation ExportTemplate {
  exportCSVTemplate(input: { projectId: "project_123" })
}
```

The `projectId` argument accepts a workspace ID or slug. Pass `databaseId` to scope the template to a single database (table) of the workspace instead of the whole workspace.

## Parameters

### ExportCSVTemplateInput

| Parameter    | Type      | Required | Description                                                                             |
| ------------ | --------- | -------- | --------------------------------------------------------------------------------------- |
| `projectId`  | `String!` | Yes      | ID or slug of the workspace to generate the template for.                               |
| `databaseId` | `String`  | No       | Scope the template to this database (table). Must be a live database in that workspace. |

## Response

`exportCSVTemplate` returns a signed download URL as a single `String!`. Opening it downloads the file directly because the URL is signed with `Content-Disposition: attachment`.

```json
{
  "data": {
    "exportCSVTemplate": "https://api.blue.app/storage/acme/marketing-plan/marketing-plan-template-301520032025-a1b2c3.csv?X-Amz-Expires=86400&X-Amz-Signature=..."
  }
}
```

The filename is derived deterministically from the workspace (or, with `databaseId`, the database's) name and a timestamp: `slugify("<name> Template <mmHHDDMMYYYY>") + ".csv"`. For a workspace named "Marketing Plan", a template generated on 20 March 2025 produces a name like `marketing-plan-template-301520032025.csv`.

### Returns

| Field               | Type      | Description                                                                                                                          |
| ------------------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `exportCSVTemplate` | `String!` | A signed download URL for the generated CSV template. The URL expires 24 hours after it is issued and forces a download when opened. |

## Template columns

The generated CSV contains a single header row followed by one empty sample row. Without `databaseId`, the columns are:

| Column                 | Description                                                        |
| ---------------------- | ------------------------------------------------------------------ |
| `Title`                | Record title.                                                      |
| `List`                 | Name of the list to place the record in.                           |
| `Done`                 | Completion status (`true` / `false`).                              |
| `Start Date`           | Record start date.                                                 |
| `Due Date`             | Record due date.                                                   |
| `Description`          | Record description.                                                |
| `Assignees`            | Comma-separated assignee emails.                                   |
| `Created At`           | Original creation timestamp.                                       |
| `Updated At`           | Original last-updated timestamp.                                   |
| `Created By`           | Original creator name or email.                                    |
| `Color`                | Record color.                                                      |
| `Project`              | Workspace name.                                                    |
| `Tags`                 | Comma-separated tag titles.                                        |
| _Custom field columns_ | One column per custom field in the workspace, ordered by position. |

<Callout variant="info" title="The sample row spans the full header">

The empty sample row underneath the header has one empty cell per column — custom fields included. It is a formatting hint, not a per-column placeholder. Add your own values across all columns when you fill in the template.

</Callout>

<Callout variant="info" title="A database-scoped template lists the database's fields">

With `databaseId`, the header is one column per field mounted on that database (duplicate field names get a ` (2)`, ` (3)` … suffix), named after the database instead of the workspace, and the base columns above are absent.

</Callout>

## Errors

| Code                | When                                                                                         |
| ------------------- | -------------------------------------------------------------------------------------------- |
| `PROJECT_NOT_FOUND` | No workspace exists for the given `projectId`, or `databaseId` is not a live database in it. |

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
4. Call `importRecords` with the `s3Key`, the `headers` mapping, and the `projectId`.
5. Subscribe to `subscribeToImportExportProgress` to track import progress.

## Related

- [Export Records](/api/import-export/export-records)
- [Uploading Files](/api/files)
- [Error Codes](/api/start-guide/error-codes)
