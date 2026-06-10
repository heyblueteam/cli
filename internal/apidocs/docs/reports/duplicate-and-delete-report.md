---
title: Duplicate & Delete a Report
description: Clone a report and its data sources, or hard-delete a report and its data sources, with the duplicateReport and deleteReport mutations.
icon: Copy
order: 5
---

Use the `duplicateReport` mutation to clone a report, and the `deleteReport` mutation to remove one. Duplicating copies the report's title, description, config, and every data source — but **not** its shared collaborators; the duplicate starts private to you, and you become its creator. Deleting hard-deletes the report and cascade-deletes its data sources, and is restricted to the report's creator.

A report aggregates records (`Todo` objects) across one or more workspaces (`Project` objects); see the [Reports overview](/api/reports) for the full model.

## Request

Duplicate a report. The `input` argument is optional — omit it to accept the default title.

```graphql
mutation DuplicateReport {
  duplicateReport(id: "report_123") {
    id
    title
    createdAt
  }
}
```

Pass a `title` to name the copy explicitly.

```graphql
mutation DuplicateReportWithTitle {
  duplicateReport(id: "report_123", input: { title: "Q4 Pipeline (working copy)" }) {
    id
    title
  }
}
```

Delete a report. `deleteReport` returns `Boolean!`, so it takes no sub-selection.

```graphql
mutation DeleteReport {
  deleteReport(id: "report_123")
}
```

## Parameters

### duplicateReport

| Parameter | Type                   | Required | Description                                                                                |
| --------- | ---------------------- | -------- | ------------------------------------------------------------------------------------------ |
| `id`      | `String!`              | Yes      | ID of the report to duplicate. You must have view access (creator, `EDITOR`, or `VIEWER`). |
| `input`   | `DuplicateReportInput` | No       | Options for the copy. Omit entirely to use defaults.                                       |

### DuplicateReportInput

| Parameter | Type     | Required | Description                                                                      |
| --------- | -------- | -------- | -------------------------------------------------------------------------------- |
| `title`   | `String` | No       | Title for the copy. If omitted, the copy is titled `"Copy of <original title>"`. |

### deleteReport

| Parameter | Type      | Required | Description                                                               |
| --------- | --------- | -------- | ------------------------------------------------------------------------- |
| `id`      | `String!` | Yes      | ID of the report to delete. You must be the report's creator (see below). |

## Response

`duplicateReport` returns the newly created `Report`.

```json
{
  "data": {
    "duplicateReport": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Copy of Q4 Pipeline",
      "createdAt": "2026-05-29T14:03:11.000Z"
    }
  }
}
```

`deleteReport` returns `true` on success.

```json
{
  "data": {
    "deleteReport": true
  }
}
```

### Returns — duplicateReport

The mutation returns the new `Report`. The duplicate lives in the same organization as the original, and you are its creator.

| Field             | Type                   | Description                                                                               |
| ----------------- | ---------------------- | ----------------------------------------------------------------------------------------- |
| `id`              | `ID!`                  | Unique identifier of the new report.                                                      |
| `uid`             | `String!`              | Stable UID for the report.                                                                |
| `title`           | `String!`              | Title of the copy.                                                                        |
| `description`     | `String`               | Description copied from the original.                                                     |
| `config`          | `JSONObject`           | Opaque view configuration copied from the original (see [note](#json-fields-are-opaque)). |
| `createdBy`       | `User!`                | The user who ran the duplication. The caller always becomes the creator.                  |
| `dataSources`     | `[ReportDataSource!]!` | Data sources copied from the original, each recreated with a fresh `uid`.                 |
| `reportUsers`     | `[ReportUser!]!`       | Empty — collaborators are **not** copied. The duplicate starts private to you.            |
| `lastGeneratedAt` | `DateTime`             | `null` on a fresh duplicate until aggregations are first computed.                        |
| `createdAt`       | `DateTime!`            | When the copy was created.                                                                |
| `updatedAt`       | `DateTime!`            | When the copy was last modified.                                                          |

For the full `Report` and `ReportDataSource` field reference (including the computed `projectIds` union and the `todos` / `todoCount` / `aggregations` query surface), see [Query Reports](/api/reports/query-reports) and [Report Data & Aggregations](/api/reports/report-data-aggregations).

### Returns — deleteReport

| Type       | Description                                                             |
| ---------- | ----------------------------------------------------------------------- |
| `Boolean!` | `true` when the report and its data sources are deleted. Never `false`. |

## What gets copied

`duplicateReport` deep-copies the report into a new record in the same organization:

- **The report** — a new `Report` with a fresh `id` and `uid`. The `title` (defaulting to `"Copy of <title>"`), `description`, and `config` are copied verbatim.
- **All data sources** — every `ReportDataSource` is recreated with a fresh `uid`, preserving its `name`, `sourceType`, `projectIds`, `filters`, and `order`. A data source with `projectIds: null` (all workspaces the creator can access) stays `null` on the copy.
- **Creator** — the calling user becomes `createdBy`, regardless of who created the original.

It deliberately does **not** copy:

- **Shared collaborators** — `reportUsers` is not carried over. The duplicate starts private to you; re-share it with [`updateReport`](/api/reports/manage-report-access) if needed.

<Callout variant="info" title="Data sources reference the new creator's workspaces">

Because each data source resolves `projectIds: null` against the report **creator's** workspace access, a copied data source on the duplicate resolves against *your* access, not the original creator's. If you have access to fewer workspaces than the original creator, the duplicate's record set can differ from the original's.

</Callout>

<Callout variant="danger" title="deleteReport is a permanent, cascading delete">

`deleteReport` hard-deletes the report — there is no soft-delete, archive, or undo. Its data sources are cascade-deleted with it. The underlying records (`Todo` objects) and workspaces are untouched; only the report and its configuration are removed.

</Callout>

## Errors

| Code               | When                                                                                                                                                                                                                         |
| ------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `UNAUTHENTICATED`  | No valid credentials on the request.                                                                                                                                                                                         |
| `REPORT_NOT_FOUND` | No report matches `id` for the caller. For `duplicateReport` this also covers no view access; for `deleteReport` it also covers any caller who is not the creator (an `EDITOR` or `VIEWER` simply finds no matching report). |

```json
{
  "errors": [
    {
      "message": "Report was not found.",
      "extensions": { "code": "REPORT_NOT_FOUND" }
    }
  ]
}
```

Omitting the required `id` is rejected by GraphQL as a non-null validation error before the resolver runs — it is not a Blue error code.

## Permissions

The two mutations sit at opposite ends of the report permission model.

| Mutation          | Required access                                                                 |
| ----------------- | ------------------------------------------------------------------------------- |
| `duplicateReport` | View access — creator, or a shared `reportUser` with role `EDITOR` or `VIEWER`. |
| `deleteReport`    | Creator **only**. Shared `EDITOR` and `VIEWER` collaborators cannot delete.     |

Anyone who can see a report can duplicate it (the duplicate is theirs, sharing nothing back to the original). Deleting is the most restricted report operation — even an `EDITOR` who can [update](/api/reports/manage-report-access) the report cannot delete it. A caller who lacks the required access receives `REPORT_NOT_FOUND` rather than a distinct permission error.

<a id="json-fields-are-opaque"></a>

The `config` field is a `JSONObject` — an opaque key/value blob whose internal shape is defined by the app, not the GraphQL schema. `duplicateReport` copies it as-is.

## Related

- [Reports overview](/api/reports) — the report model, data sources, and how reports relate to records.
- [Create a Report](/api/reports/create-report) — build a report from scratch.
- [Query Reports](/api/reports/query-reports) — fetch a report by id or list reports, with the full `Report` field reference.
- [Manage Report Access](/api/reports/manage-report-access) — update a report and share it with `EDITOR` / `VIEWER` collaborators.
- [Report Data & Aggregations](/api/reports/report-data-aggregations) — read the records, counts, and aggregations a report produces.
