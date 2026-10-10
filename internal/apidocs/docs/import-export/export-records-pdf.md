---
title: Export Records as PDF
description: Export workspace records to a single PDF rendered from a Portable Document template and delivered by email, with optional filters and real-time progress.
icon: FileText
order: 4
---

Use the `exportRecordsPdf` mutation to export the records of a workspace as **one PDF**, rendered from a [Portable Document](/api/documents/portable-documents) template. The mutation is asynchronous: it returns `true` once the job is queued, the finished PDF is **emailed to the requesting user**, and progress streams over the `subscribeToImportExportProgress` subscription. Records are `Record` objects, a workspace is a `Workspace`, a record table is a `RecordDatabase`, and a template is a `PortableDocument` in the API.

```
mutation ──► export-pdf queue ──► worker renders one PDF (one page set per record)
                                      │
                                      ├─► progress events ──► subscription + exportPdfProgress query
                                      └─► uploaded file ──► "Download PDF" email
```

<Callout variant="warning" title="Rate-limited, and capped at 250 records">

The mutation accepts **1 request per 50 seconds** per token (GraphQL Shield rate limiter); a second call inside that window is rejected with `RATE_LIMITED`. The worker assembles the PDF fully in memory, so an export is **capped at 250 records** — a filter that matches more fails the job and suggests [CSV export](/api/import-export/export-records) instead.

</Callout>

## Request

Export every record of a workspace through a template:

```graphql
mutation ExportRecordsPdf {
  exportRecordsPdf(input: { projectId: "project_123", pdfId: "clm4n8qwx000008l0g4oxdqn7" })
}
```

`projectId` must be the workspace **ID** — a slug is not accepted and fails with `PROJECT_NOT_FOUND`. `pdfId` is the template's `id` (the CUID returned by the `portableDocuments` query), not its `pdf_…` `uid`. The template must be **active** and belong to that workspace.

## Parameters

### ExportRecordsPdfInput

| Parameter    | Type            | Required | Description                                                                                                                                                                                       |
| ------------ | --------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `projectId`  | `String!`       | Yes      | ID of the workspace to export records from. Must be the workspace ID — a slug is not accepted.                                                                                                    |
| `pdfId`      | `String!`       | Yes      | ID of the Portable Document template to render each matching record with. Must be active, and belong to the workspace.                                                                            |
| `databaseId` | `String`        | No       | Export only this record table of the workspace. Must be a live table in that workspace, and the template must be linked to it (below). If omitted, every live table of the workspace is exported. |
| `filter`     | `RecordsFilter` | No       | Criteria to limit which records are exported. If omitted, all records are exported.                                                                                                               |

### The template's table link

A template fills its overlay placements from a table's fields through the link created with `setPortableDocumentDatabaseLinks`. With `databaseId` set, the template must have a link to that table, and the mutation validates the link against the live template placements and the table fields **you can see** before queueing the job. A template that is not linked to the selected table, or a mapping that no longer fits, fails with `BAD_USER_INPUT` before the job is queued. See [Link a template to record tables](/api/documents/portable-documents#link-a-template-to-record-tables) for the mapping rules and the field types the renderer covers.

With `databaseId` omitted, each matching record still renders through the table it was selected from, using that table's link when the template has one.

### RecordsFilter

Combine any of these fields to narrow the export. Unless noted, fields are ANDed together.

| Parameter                                                                                                   | Type                        | Required | Description                                                                                                             |
| ----------------------------------------------------------------------------------------------------------- | --------------------------- | -------- | ----------------------------------------------------------------------------------------------------------------------- |
| `companyIds`                                                                                                | `[String!]!`                | Yes      | Organization IDs that scope the filter. Required by the type even though the workspace is already given by `projectId`. |
| `todoIds`                                                                                                   | `[String!]`                 | No       | Export only these specific record IDs.                                                                                  |
| `assigneeIds`                                                                                               | `[String!]`                 | No       | Include records assigned to any of these users.                                                                         |
| `assigneeMatch`                                                                                             | `AssigneeMatch`             | No       | How multiple `assigneeIds` combine — `ANY` (default) or `ALL`.                                                          |
| `hasAnyAssignee`                                                                                            | `Boolean`                   | No       | `true` → at least one assignee; `false` → none.                                                                         |
| `unassigned`                                                                                                | `Boolean`                   | No       | When `true`, include only records with no assignee.                                                                     |
| `tagIds`                                                                                                    | `[String!]`                 | No       | Include records carrying any of these tag IDs.                                                                          |
| `tagColors`                                                                                                 | `[String!]`                 | No       | Include records carrying a tag of any of these hex colors.                                                              |
| `tagTitles`                                                                                                 | `[String!]`                 | No       | Include records carrying a tag with any of these titles.                                                                |
| `todoListIds`                                                                                               | `[String!]`                 | No       | Include records in any of these lists (by ID).                                                                          |
| `todoListTitles`                                                                                            | `[String!]`                 | No       | Include records in any of these lists (by title).                                                                       |
| `done`                                                                                                      | `Boolean`                   | No       | Filter by completion status (`true` = completed).                                                                       |
| `showCompleted`                                                                                             | `Boolean`                   | No       | Include completed records in the result set.                                                                            |
| `startedAt`                                                                                                 | `DateTime`                  | No       | Include records started on or after this timestamp.                                                                     |
| `duedAt`                                                                                                    | `DateTime`                  | No       | Include records due on or before this timestamp.                                                                        |
| `dueStart` / `dueEnd`                                                                                       | `DateTime`                  | No       | Due-date range start and end.                                                                                           |
| `duedAtStart` / `duedAtEnd`                                                                                 | `DateTime`                  | No       | Alternative due-date range start and end.                                                                               |
| `createdStart` / `createdEnd`                                                                               | `DateTime`                  | No       | Include records created on or after / on or before this timestamp.                                                      |
| `createdByIds`                                                                                              | `[String!]`                 | No       | Include records created by any of these users.                                                                          |
| `updatedAt_gt` / `updatedAt_gte`                                                                            | `DateTime`                  | No       | Include records updated after / at or after this timestamp.                                                             |
| `archived`                                                                                                  | `Boolean`                   | No       | Archive scope. Omitted or `false` → active records only; `true` → archived records only.                                |
| `search`                                                                                                    | `String`                    | No       | Full-text search query.                                                                                                 |
| `q`                                                                                                         | `String`                    | No       | Quick search query (matches the record title).                                                                          |
| `recordName`                                                                                                | `String`                    | No       | Record-title substring search. Prefer over `q` for name-only search.                                                    |
| `recordNameOp`                                                                                              | `FilterComparisonOperator`  | No       | Comparison operator applied to `recordName` (`CONTAINS`, `NOT_CONTAINS`, `IS`, …).                                      |
| `notAssigneeIds`                                                                                            | `[String!]`                 | No       | Exclude records assigned to any of these users.                                                                         |
| `notTagIds`                                                                                                 | `[String!]`                 | No       | Exclude records carrying any of these tag IDs.                                                                          |
| `notColors`                                                                                                 | `[String!]`                 | No       | Exclude records of any of these colors.                                                                                 |
| `notTodoListIds`                                                                                            | `[String!]`                 | No       | Exclude records in any of these lists.                                                                                  |
| `hasTag` / `hasColor` / `hasDueDate` / `hasDescription` / `hasChecklist` / `hasDependency` / `hasReference` | `Boolean`                   | No       | Existence filters; when `true`, include only records that have the named attribute.                                     |
| `fields`                                                                                                    | `JSON`                      | No       | Custom-field filter conditions. See [custom-field filters](#custom-field-filters).                                      |
| `op`                                                                                                        | `FilterLogicalOperator`     | No       | Logical operator applied to the `fields` conditions. Defaults to `AND`.                                                 |
| `groups`                                                                                                    | `[RecordFilterGroupInput!]` | No       | Nested filter groups. **When provided, the flat fields above are ignored in favor of the groups.**                      |
| `groupLinks`                                                                                                | `[FilterLogicalOperator!]`  | No       | Operators linking adjacent groups; length should be `groups.length - 1`. If shorter, the last operator is repeated.     |
| `projectIds`                                                                                                | `[String!]`                 | No       | No effect here — the PDF export covers only the one workspace named by `projectId`.                                     |
| `excludeArchivedProjects`                                                                                   | `Boolean`                   | No       | No effect here — the export is scoped to the one workspace.                                                             |
| `rankBefore`                                                                                                | `RecordRankCursorInput`     | No       | Keyset cursor for paginated list reads. No effect on the export.                                                        |
| `coordinates`                                                                                               | `JSON`                      | No       | Map-view bounding box. No effect on the export.                                                                         |

Filter conditions resolve against each exported table's mounted fields, so a condition on a custom field matches the records of every table that mounts a field it resolves to.

#### Custom-field filters

Each entry in the `fields` array **must include a `type`** — the query layer routes on it, and entries without `type` are silently dropped. The most common type is `CUSTOM_FIELD`:

```json
[
  {
    "type": "CUSTOM_FIELD",
    "customFieldId": "field_123",
    "customFieldType": "CHECKBOX",
    "op": "IS",
    "values": true
  }
]
```

Other valid `type` values are `TIME_IN_LIST`, `RELATIVE_DUEDATE`, and `FIELD`. The operators available per `customFieldType` depend on the field type.

## Response

`exportRecordsPdf` returns a bare boolean — `true` once the job is queued. The PDF is generated in the background and emailed to the requesting user.

```json
{ "data": { "exportRecordsPdf": true } }
```

### Returns

| Field              | Type       | Description                                 |
| ------------------ | ---------- | ------------------------------------------- |
| `exportRecordsPdf` | `Boolean!` | `true` when the export job has been queued. |

## Full example

Export the records of one table — incomplete, assigned to two users, tagged, due in Q1, matching a search term and a custom-field value — through a linked template:

```graphql
mutation ExportFilteredRecordsPdf {
  exportRecordsPdf(
    input: {
      projectId: "project_123"
      pdfId: "clm4n8qwx000008l0g4oxdqn7"
      databaseId: "clm4n8qwx000008l0t9table1"
      filter: {
        companyIds: ["company_123"]
        done: false
        assigneeIds: ["user_123", "user_456"]
        tagIds: ["tag_123"]
        dueStart: "2026-01-01T00:00:00Z"
        dueEnd: "2026-03-31T23:59:59Z"
        q: "launch"
        fields: [
          {
            type: "CUSTOM_FIELD"
            customFieldId: "field_123"
            customFieldType: "SELECT_SINGLE"
            op: "IS"
            values: ["option_123"]
          }
        ]
        op: AND
      }
    }
  )
}
```

## Which records are selected

- **With `databaseId`** — only that table's records are considered.
- **Without `databaseId`** — every live table of the workspace is walked in table order. Each matching record prints through the table it was selected from, and a record that sits in several tables prints once.

The selection runs through the exporting user's live record and field policy: a record, or a field on a record, that your role cannot view is left out of the PDF. An empty result fails the job — there is no "0 pages" PDF (see [Job-time failures](#job-time-failures)).

## Delivery

The worker uploads the finished PDF and publishes a `COMPLETE` progress event carrying the download URL, then emails the requesting user a **Download PDF** link for the file. The URL is a signed export link, so it expires; the email is the durable delivery, and the same URL is available from the progress channel below.

## Track export progress

Subscribe to `subscribeToImportExportProgress` to receive real-time progress for the current user's imports and exports. The subscription filters server-side by `userId`, so pass the ID of the user who started the export.

```graphql
subscription TrackPdfExportProgress {
  subscribeToImportExportProgress(projectId: "project_123", userId: "user_123")
}
```

PDF export events carry `type: "EXPORT_PDF"`:

```json
{
  "data": {
    "subscribeToImportExportProgress": {
      "type": "EXPORT_PDF",
      "status": "IN_PROGRESS",
      "userId": "user_123",
      "projectId": "project_123",
      "projectName": "Q1 Launch",
      "current": 120,
      "total": 250,
      "progress": 48
    }
  }
}
```

| Field                            | Description                                                                                   |
| -------------------------------- | --------------------------------------------------------------------------------------------- |
| `type`                           | Always `"EXPORT_PDF"` for this export.                                                        |
| `status`                         | `IN_PROGRESS` while rendering, `COMPLETE` with `fileUrl`, or `ERROR` with an `error` message. |
| `projectId`                      | The workspace **ID** you passed to the mutation.                                              |
| `current` / `total` / `progress` | Records rendered so far, total records, and the percentage. `ERROR` events omit them.         |
| `fileUrl`                        | Download URL of the finished PDF — present only on `COMPLETE`.                                |

Live events are fire-and-forget: a client that navigates away or refreshes misses everything sent in between. The `exportPdfProgress` query reads the durable server-side snapshot (kept for about 1 hour, scoped to the calling user) to re-sync on mount:

```graphql
query ExportPdfProgress {
  exportPdfProgress(projectId: "project_123")
}
```

```json
{
  "data": {
    "exportPdfProgress": {
      "status": "COMPLETE",
      "current": 250,
      "total": 250,
      "fileUrl": "https://api.blue.app/exports/eyJ...",
      "updatedAt": "2026-05-29T14:03:02.000Z"
    }
  }
}
```

The snapshot's `status` is `GENERATING`, `COMPLETE`, or `ERROR`; on `ERROR` it carries `error` instead of `fileUrl`.

## Errors

| Code                          | When                                                                                                                                                                                                             |
| ----------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `PROJECT_NOT_FOUND`           | The workspace does not exist, `projectId` is a slug instead of the ID, or the requesting user has no membership row in it. `databaseId` raises the same code when it is not a live table in that workspace.      |
| `PORTABLE_DOCUMENT_NOT_FOUND` | No template with that `pdfId` in the workspace — an inactive template is indistinguishable from a missing or foreign one.                                                                                        |
| `BAD_USER_INPUT`              | The template is not linked to the selected `databaseId`, or the link fails validation: a placement has no mapping, the mapped field is not visible to you on that table, or its type does not fit the placement. |
| `RATE_LIMITED`                | More than one `exportRecordsPdf` request within 50 seconds by the same token. The 50-second window is per operation, so it is separate from `exportRecords`' own window.                                         |

## Job-time failures

Everything above rejects the mutation itself. The queued job can still fail; the failure surfaces as an `ERROR` progress event plus a failure email to the requesting user, never as a GraphQL error:

| Situation                                              | What the user is told                                                          |
| ------------------------------------------------------ | ------------------------------------------------------------------------------ |
| The filter matched 0 records.                          | `No records matched your active filters, so there was nothing to export.`      |
| The filter matched more than 250 records.              | The count and the cap, with the advice to narrow the filter or use CSV export. |
| The template was deleted or deactivated meanwhile.     | A generic failure message.                                                     |
| A mapped field is no longer mounted on the table.      | A generic failure message.                                                     |
| A filter condition the record engine does not support. | A generic failure message.                                                     |

## Permissions

| Check            | Requirement                                                                                                           |
| ---------------- | --------------------------------------------------------------------------------------------------------------------- |
| Field permission | Project-level `OWNER`, `ADMIN`, or `MEMBER` on an **active** workspace.                                               |
| Resolver         | The caller must have a membership row in the workspace (part of the project lookup).                                  |
| Worker           | The selection applies the exporting user's record and field visibility, so the PDF shows only what that user can see. |

There is no batch or multi-workspace PDF export — one call covers one workspace.

## Notes

- **One PDF, not one per table.** Unlike the CSV export, which builds a file per database, `exportRecordsPdf` appends every matching record's pages to a single document.
- **The template must be built first.** Create and place fields on the template as described in [Build Portable Document templates](/api/documents/portable-documents), and print a single record with `printPortableDocument` to check the layout before exporting hundreds of records.
- **Rate limit and tier limits stack.** The per-operation 1-per-50-seconds limit runs inside GraphQL on top of the tier limits described in [Rate Limits](/api/start-guide/rate-limits).
- **Imports and CSV exports are covered elsewhere.** See the [Import/Export overview](/api/import-export) and [Export records](/api/import-export/export-records).

## Related

- [Export records](/api/import-export/export-records) — the CSV equivalent, with the full filter behavior
- [Build Portable Document templates](/api/documents/portable-documents) — create the template and link it to record tables
- [Map fields and print Portable Documents](/api/documents/portable-document-fields) — place overlay fields and print one record
- [Export a CSV template](/api/import-export/export-csv-template) — download the workspace's columns before importing
- [Rate Limits](/api/start-guide/rate-limits)
