---
title: Import Records
description: Bulk-import records into a workspace from a CSV file and track progress through a real-time subscription.
icon: Upload
order: 3
---

Use the `importRecords` mutation to bulk-import records into a workspace from a CSV file that has already been uploaded to storage. The file is parsed in streaming chunks and records are created in the background. Records are `Record` objects and a workspace is a `Workspace` in the API.

The import runs asynchronously: the mutation returns `true` once the job is accepted, then you follow its progress with the `subscribeToImportExportProgress` subscription. Use `cancelRecordImport` to stop an import that is still running.

Before calling `importRecords`, upload the CSV and obtain its storage key. See [Uploading Files](/api/files) for both upload paths.

## Request

The smallest call needs the storage key of the uploaded file, an ordered array of column headers, and the workspace `id`.

```graphql
mutation ImportRecords {
  importRecords(
    input: {
      s3Key: "uploads/org_123/project_123/import.csv"
      headers: ["Title", "List", "Done", "Due Date"]
      projectId: "project_123"
    }
  )
}
```

`projectId` is the workspace CUID (the `id` field of `Workspace`). The headers map each CSV column, in order, to a record field; see [Recognized headers](#recognized-headers) below.

## Parameters

### ImportRecordsInput

| Parameter       | Type      | Required | Description                                                                                                                                                           |
| --------------- | --------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `s3Key`         | `String!` | Yes      | Storage key of the uploaded CSV file. Validated server-side: keys containing `..`, a leading `/`, or characters outside `^[a-zA-Z0-9][a-zA-Z0-9_./-]*$` are rejected. |
| `headers`       | `JSON!`   | Yes      | Ordered JSON array of column header names. Each entry maps the CSV column at that position to a record field, matched by name.                                        |
| `projectId`     | `String!` | Yes      | The workspace `id` (CUID) to import records into.                                                                                                                     |
| `useRustImport` | `Boolean` | No       | **Deprecated** — no longer used and has no effect. Do not set it.                                                                                                     |

### Recognized headers

The `headers` array tells the importer how to interpret each CSV column. These built-in names are recognized; any other header is matched to a custom field of the same name in the workspace.

| Header        | Maps to                                                                |
| ------------- | ---------------------------------------------------------------------- |
| `Title`       | Record title                                                           |
| `List`        | Name of the list to place the record in (created if it does not exist) |
| `Done`        | Completion status (`true` / `false`)                                   |
| `Start Date`  | Record start date                                                      |
| `Due Date`    | Record due date                                                        |
| `Description` | Record description                                                     |
| `Assignees`   | Comma-separated assignee names or emails                               |
| `Created At`  | Original creation timestamp                                            |
| `Updated At`  | Original last-updated timestamp                                        |
| `Created By`  | Original creator name or email                                         |
| `Color`       | Record color                                                           |
| `Tags`        | Comma-separated tag names (created if they do not exist)               |

To produce a CSV that already lines up with a workspace's columns, download a [CSV template](/api/import-export/export-csv-template) first.

## Response

`importRecords` returns `Boolean!` — `true` once the import has been accepted and queued.

```json
{ "data": { "importRecords": true } }
```

A `true` response means the job is enqueued, not finished. Subscribe to [import progress](#track-import-progress) to know when it completes.

## Full example

Custom-field columns are imported by adding their exact field name as a header alongside the built-in ones. Here the workspace has `Budget` and `Priority` custom fields:

```graphql
mutation ImportRecordsWithCustomFields {
  importRecords(
    input: {
      s3Key: "uploads/org_123/project_123/import.csv"
      headers: ["Title", "List", "Done", "Due Date", "Assignees", "Tags", "Budget", "Priority"]
      projectId: "project_123"
    }
  )
}
```

## Track import progress

Subscribe to `subscribeToImportExportProgress` to receive real-time progress while an import (or [export](/api/import-export/export-records)) runs. The field returns `JSON` (nullable); each event is an untyped JSON object with the shape below.

```graphql
subscription TrackImportProgress {
  subscribeToImportExportProgress(projectId: "project_123", userId: "user_123")
}
```

The stream is filtered by `userId` only: you receive every import and export event for that user. `projectId` is required by the schema but does not currently scope the events.

### Event payload

| Field         | Type     | Description                                                |
| ------------- | -------- | ---------------------------------------------------------- |
| `type`        | `String` | `"IMPORT"` or `"EXPORT"`                                   |
| `userId`      | `String` | `id` of the user who initiated the operation               |
| `username`    | `String` | First name of that user                                    |
| `projectId`   | `String` | Workspace **slug** (not the CUID you pass to the mutation) |
| `projectName` | `String` | Workspace name                                             |
| `status`      | `String` | `IN_PROGRESS`, `COMPLETE`, `CANCELLED`, or `ERROR`         |
| `progress`    | `Float`  | Percentage complete, 0–100                                 |
| `error`       | `String` | Error message, present only when `status` is `ERROR`       |

A sample event:

```json
{
  "data": {
    "subscribeToImportExportProgress": {
      "type": "IMPORT",
      "userId": "clm4n8qwx000008l0g4oxdqn7",
      "username": "Ada",
      "projectId": "marketing-ops",
      "projectName": "Marketing Ops",
      "status": "IN_PROGRESS",
      "progress": 42
    }
  }
}
```

Note the identifier spaces differ: you pass the workspace **CUID** as `projectId` to `importRecords`, but the payload's `projectId` is the workspace **slug**.

## Cancel an import

Use `cancelRecordImport` to stop an import that is still running. The worker checks for the cancel flag on each chunk and aborts.

```graphql
mutation CancelImport {
  cancelRecordImport(projectId: "project_123")
}
```

`projectId` accepts either the workspace `id` (CUID) or its slug. The mutation returns `Boolean!` — `true` once the cancellation has been recorded.

```json
{ "data": { "cancelRecordImport": true } }
```

## Errors

| Code                         | When                                                                                                                                                                                                                                                                                                                            |
| ---------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `BAD_USER_INPUT`             | `s3Key` is missing or fails validation (contains `..`, starts with `/`, or has disallowed characters). Message: `Invalid s3Key`.                                                                                                                                                                                                |
| `PROJECT_NOT_FOUND`          | The `projectId` does not resolve to a workspace, or the caller cannot access it. Message: `Project was not found.`                                                                                                                                                                                                              |
| `PLAN_LIMIT_REACHED`         | Importing would push the workspace over its plan's per-workspace record cap. Checked before the import starts (pre-flight), against the exact same `recordsPerWorkspace` limit enforced everywhere else records are counted. See [error codes](/api/start-guide/error-codes#plan_limit_reached) for the structured error shape. |
| `IMPORT_ALREADY_IN_PROGRESS` | An import is already running for this workspace. Message: `An import is already running for this workspace. Please wait for it to finish before starting another.`                                                                                                                                                              |
| `FORBIDDEN`                  | The caller lacks permission for the operation (see [Permissions](#permissions)).                                                                                                                                                                                                                                                |

## Permissions

| Operation            | Requirement                                                                                                                                                                       |
| -------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `importRecords`      | Project-level `OWNER`, `ADMIN`, or `MEMBER` (view-only and comment-only roles are rejected), the workspace must be active, and any custom role must have record creation enabled. |
| `cancelRecordImport` | Project-level `OWNER` or `ADMIN` on an active workspace. Enforced in the resolver; `projectId` may be an `id` or a slug.                                                          |

<Callout variant="info" title="One import at a time">

Only one import can run per workspace. A second `importRecords` call while one is active throws `IMPORT_ALREADY_IN_PROGRESS`. This guard is unconditional — it is enforced by a Redis lock plus a per-workspace BullMQ job ID for every import, with no path that silently succeeds.

</Callout>

<Callout variant="warning" title="Import limit scales with your plan">

An import can't push a workspace's record count past your plan's `recordsPerWorkspace` cap — the same limit enforced for every other way of creating records, not a separate import-only number. Upgrading your plan raises the cap for imports along with everything else.

</Callout>

### Behavior notes

- **Lists and tags auto-create.** Lists and tags referenced in the CSV that don't already exist in the workspace are created during the import.
- **Custom fields match by name.** Any header that isn't a built-in name is matched to a custom field of the same name. Add custom-field columns to the `headers` array to import their values.
- **Upload first.** The CSV must be uploaded to storage before you call `importRecords`. See [Uploading Files](/api/files) to get the `s3Key`.

## Related

- [Export records](/api/import-export/export-records)
- [Export a CSV template](/api/import-export/export-csv-template)
- [Uploading files](/api/files)
- [Error codes](/api/start-guide/error-codes)
