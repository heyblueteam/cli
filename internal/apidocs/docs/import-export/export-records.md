---
title: Export Records
description: Export workspace records to a CSV file delivered by email, with optional filters for assignees, tags, dates, and custom fields.
icon: Download
order: 1
---

Use the `exportRecords` mutation to start an asynchronous CSV export of records from a workspace. The mutation returns immediately with `true` to confirm the job was queued; the finished CSV is **emailed to the requesting user** as a download link. There is no API-returned URL — the emailed link is the only way to fetch the file. Track progress in real time with the `subscribeToImportExportProgress` subscription. (Records are `Record` objects in the API; a workspace is a `Workspace`.)

This page also covers two related exports:

- `exportReport` — export records across every workspace referenced by a report's data sources.
- `exportChartCSV` — export the underlying data of a dashboard chart.

<Callout variant="warning" title="Exports are rate-limited">

`exportRecords` and `exportReport` are limited to **1 request per 50 seconds** per token (GraphQL Shield rate limiter). A second export call inside that window is rejected with `RATE_LIMITED`. Wait at least 50 seconds between exports.

</Callout>

## Request

Export all records from a workspace. Only `projectId` is required.

```graphql
mutation ExportRecords {
  exportRecords(input: { projectId: "project_123" })
}
```

`projectId` accepts a workspace ID or slug.

## Parameters

### ExportRecordsInput

| Parameter       | Type          | Required | Description                                                                                          |
| --------------- | ------------- | -------- | ---------------------------------------------------------------------------------------------------- |
| `projectId`     | `String!`     | Yes      | ID or slug of the workspace to export records from.                                                  |
| `filter`        | `TodosFilter` | No       | Criteria to limit which records are exported. If omitted, all records in the workspace are exported. |
| `useRustExport` | `Boolean`     | No       | **Deprecated — no longer used; has no effect.** The resolver ignores this field.                     |

### TodosFilter

Combine any of these fields to narrow the export. Unless noted, fields are ANDed together.

| Parameter                 | Type                       | Required | Description                                                                                                             |
| ------------------------- | -------------------------- | -------- | ----------------------------------------------------------------------------------------------------------------------- |
| `companyIds`              | `[String!]!`               | Yes      | Organization IDs that scope the export. Required by the type even though the workspace is already given by `projectId`. |
| `projectIds`              | `[String!]`                | No       | Additional workspace IDs to include.                                                                                    |
| `todoIds`                 | `[String!]`                | No       | Export only these specific record IDs.                                                                                  |
| `assigneeIds`             | `[String!]`                | No       | Include records assigned to any of these users.                                                                         |
| `unassigned`              | `Boolean`                  | No       | When `true`, include only records with no assignee.                                                                     |
| `tagIds`                  | `[String!]`                | No       | Include records carrying any of these tag IDs.                                                                          |
| `tagColors`               | `[String!]`                | No       | Include records carrying a tag of any of these hex colors.                                                              |
| `tagTitles`               | `[String!]`                | No       | Include records carrying a tag with any of these titles.                                                                |
| `todoListIds`             | `[String!]`                | No       | Include records in any of these lists (by ID).                                                                          |
| `todoListTitles`          | `[String!]`                | No       | Include records in any of these lists (by title).                                                                       |
| `done`                    | `Boolean`                  | No       | Filter by completion status (`true` = completed).                                                                       |
| `showCompleted`           | `Boolean`                  | No       | Include completed records in the result set.                                                                            |
| `startedAt`               | `DateTime`                 | No       | Include records started on or after this timestamp.                                                                     |
| `duedAt`                  | `DateTime`                 | No       | Include records due on or before this timestamp.                                                                        |
| `dueStart`                | `DateTime`                 | No       | Due-date range start.                                                                                                   |
| `dueEnd`                  | `DateTime`                 | No       | Due-date range end.                                                                                                     |
| `duedAtStart`             | `DateTime`                 | No       | Alternative due-date range start.                                                                                       |
| `duedAtEnd`               | `DateTime`                 | No       | Alternative due-date range end.                                                                                         |
| `createdStart`            | `DateTime`                 | No       | Include records created on or after this timestamp.                                                                     |
| `createdEnd`              | `DateTime`                 | No       | Include records created on or before this timestamp.                                                                    |
| `updatedAt_gt`            | `DateTime`                 | No       | Include records updated after this timestamp.                                                                           |
| `updatedAt_gte`           | `DateTime`                 | No       | Include records updated at or after this timestamp.                                                                     |
| `notAssigneeIds`          | `[String!]`                | No       | Exclude records assigned to any of these users.                                                                         |
| `notTagIds`               | `[String!]`                | No       | Exclude records carrying any of these tag IDs.                                                                          |
| `notColors`               | `[String!]`                | No       | Exclude records of any of these colors.                                                                                 |
| `notTodoListIds`          | `[String!]`                | No       | Exclude records in any of these lists.                                                                                  |
| `hasTag`                  | `Boolean`                  | No       | When `true`, include only records that have at least one tag.                                                           |
| `hasColor`                | `Boolean`                  | No       | When `true`, include only records that have a color set.                                                                |
| `hasDueDate`              | `Boolean`                  | No       | When `true`, include only records that have a due date.                                                                 |
| `hasDescription`          | `Boolean`                  | No       | When `true`, include only records that have a description.                                                              |
| `hasChecklist`            | `Boolean`                  | No       | When `true`, include only records that have a checklist.                                                                |
| `hasDependency`           | `Boolean`                  | No       | When `true`, include only records that have a dependency.                                                               |
| `hasReference`            | `Boolean`                  | No       | When `true`, include only records that have a reference.                                                                |
| `search`                  | `String`                   | No       | Full-text search query.                                                                                                 |
| `q`                       | `String`                   | No       | Quick search query (matches the record title).                                                                          |
| `excludeArchivedProjects` | `Boolean`                  | No       | When `true`, exclude records from archived workspaces.                                                                  |
| `fields`                  | `JSON`                     | No       | Custom-field filter conditions. See [custom-field filters](#custom-field-filters).                                      |
| `op`                      | `FilterLogicalOperator`    | No       | Logical operator applied to the `fields` conditions. Defaults to `AND`.                                                 |
| `coordinates`             | `JSON`                     | No       | Geographic bounding box for map-view filtering (raw JSON; no fixed schema).                                             |
| `groups`                  | `[TodoFilterGroupInput!]`  | No       | Nested filter groups. **When provided, the flat fields above are ignored in favor of the groups.**                      |
| `groupLinks`              | `[FilterLogicalOperator!]` | No       | Operators linking adjacent groups; length should be `groups.length - 1`. If shorter, the last operator is repeated.     |

### FilterLogicalOperator

| Value | Description                              |
| ----- | ---------------------------------------- |
| `AND` | All conditions must match (the default). |
| `OR`  | Any condition may match.                 |

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

`exportRecords` returns a bare boolean — `true` once the job is queued. The CSV is generated in the background and emailed to the requesting user; the API never returns the file or a URL.

```json
{ "data": { "exportRecords": true } }
```

### Returns

| Field           | Type       | Description                                 |
| --------------- | ---------- | ------------------------------------------- |
| `exportRecords` | `Boolean!` | `true` when the export job has been queued. |

## Full example

Export incomplete records assigned to two users, tagged, due in Q1, matching a search term and a custom-field value:

```graphql
mutation ExportFilteredRecords {
  exportRecords(
    input: {
      projectId: "project_123"
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

## Export a report

Use the `exportReport` mutation to export records across every workspace referenced by a report's data sources. The export combines the records from all of those workspaces and applies any report-level filters; like `exportRecords`, it returns `true` and delivers the CSV by email.

```graphql
mutation ExportReport {
  exportReport(input: { reportId: "report_123" })
}
```

```json
{ "data": { "exportReport": true } }
```

### ExportReportInput

| Parameter  | Type      | Required | Description                 |
| ---------- | --------- | -------- | --------------------------- |
| `reportId` | `String!` | Yes      | ID of the report to export. |

## Export a chart

Use the `exportChartCSV` mutation to export the underlying data of a dashboard chart as a CSV file. It returns `true` and emails the file to the requesting user.

```graphql
mutation ExportChart {
  exportChartCSV(chartId: "chart_123", filter: { assigneeIds: ["user_123"], showCompleted: false })
}
```

```json
{ "data": { "exportChartCSV": true } }
```

| Parameter | Type              | Required | Description                                   |
| --------- | ----------------- | -------- | --------------------------------------------- |
| `chartId` | `ID!`             | Yes      | ID of the chart to export.                    |
| `filter`  | `TodoFilterInput` | No       | Filter to scope the chart data before export. |

### TodoFilterInput

The chart filter is a **different type** from `TodosFilter` (which `exportRecords` uses). `TodoFilterInput` has no `companyIds`, `todoIds`, `search`, `startedAt`, `duedAt`, `coordinates`, or `excludeArchivedProjects`; instead it uses `colors`, and adds completed-date and last-updated-by axes. Fields shared with `TodosFilter` (`assigneeIds`, `tagIds`, `showCompleted`, `q`, …) behave the same — but a key like `companyIds` is not accepted here.

| Parameter                                                                                                   | Type                       | Required | Description                                                                         |
| ----------------------------------------------------------------------------------------------------------- | -------------------------- | -------- | ----------------------------------------------------------------------------------- |
| `assigneeIds`                                                                                               | `[String!]`                | No       | Include records assigned to any of these users.                                     |
| `unassigned`                                                                                                | `Boolean`                  | No       | When `true`, include only unassigned records.                                       |
| `colors`                                                                                                    | `[String!]`                | No       | Include records of any of these colors.                                             |
| `dueStart`                                                                                                  | `DateTime`                 | No       | Due-date range start.                                                               |
| `dueEnd`                                                                                                    | `DateTime`                 | No       | Due-date range end.                                                                 |
| `showCompleted`                                                                                             | `Boolean`                  | No       | Include completed records.                                                          |
| `projectIds`                                                                                                | `[String!]`                | No       | Scope to these workspace IDs.                                                       |
| `q`                                                                                                         | `String`                   | No       | Quick search query.                                                                 |
| `tagIds`                                                                                                    | `[String!]`                | No       | Include records carrying any of these tag IDs.                                      |
| `tagColors`                                                                                                 | `[String!]`                | No       | Include records carrying a tag of any of these hex colors.                          |
| `tagTitles`                                                                                                 | `[String!]`                | No       | Include records carrying a tag with any of these titles.                            |
| `todoListIds`                                                                                               | `[String!]`                | No       | Include records in any of these lists (by ID).                                      |
| `todoListTitles`                                                                                            | `[String!]`                | No       | Include records in any of these lists (by title).                                   |
| `updatedAt_gt`                                                                                              | `DateTime`                 | No       | Include records updated after this timestamp.                                       |
| `updatedAt_gte`                                                                                             | `DateTime`                 | No       | Include records updated at or after this timestamp.                                 |
| `createdStart`                                                                                              | `DateTime`                 | No       | Include records created on or after this timestamp.                                 |
| `createdEnd`                                                                                                | `DateTime`                 | No       | Include records created on or before this timestamp.                                |
| `completedStart`                                                                                            | `DateTime`                 | No       | Include records completed on or after this timestamp.                               |
| `completedEnd`                                                                                              | `DateTime`                 | No       | Include records completed on or before this timestamp.                              |
| `recordName`                                                                                                | `String`                   | No       | Filter by record title (full-text match).                                           |
| `recordNameOp`                                                                                              | `FilterComparisonOperator` | No       | Comparison operator applied to `recordName`.                                        |
| `notAssigneeIds`                                                                                            | `[String!]`                | No       | Exclude records assigned to any of these users.                                     |
| `notTagIds`                                                                                                 | `[String!]`                | No       | Exclude records carrying any of these tag IDs.                                      |
| `notColors`                                                                                                 | `[String!]`                | No       | Exclude records of any of these colors.                                             |
| `notTodoListIds`                                                                                            | `[String!]`                | No       | Exclude records in any of these lists.                                              |
| `hasTag` / `hasColor` / `hasDueDate` / `hasDescription` / `hasChecklist` / `hasDependency` / `hasReference` | `Boolean`                  | No       | Existence filters; when `true`, include only records that have the named attribute. |
| `lastUpdatedByUserIds`                                                                                      | `[String!]`                | No       | Include records last updated by any of these users.                                 |
| `lastUpdatedByAutomationIds`                                                                                | `[String!]`                | No       | Include records last updated by any of these automations.                           |
| `lastUpdatedByActorTypes`                                                                                   | `[ActorType!]`             | No       | Include records last updated by any of these actor types.                           |
| `fields`                                                                                                    | `JSON`                     | No       | Custom-field filter conditions (same shape as above; each entry needs a `type`).    |
| `op`                                                                                                        | `FilterLogicalOperator`    | No       | Logical operator for the `fields` conditions. Defaults to `AND`.                    |
| `groups`                                                                                                    | `[TodoFilterGroupInput!]`  | No       | Nested filter groups; when provided, the flat fields above are ignored.             |
| `groupLinks`                                                                                                | `[FilterLogicalOperator!]` | No       | Operators linking adjacent groups.                                                  |

## Track export progress

Subscribe to `subscribeToImportExportProgress` to receive real-time progress for the current user's imports and exports. The subscription filters server-side by `userId`, so pass the ID of the user who started the export.

```graphql
subscription TrackExportProgress {
  subscribeToImportExportProgress(projectId: "project_123", userId: "user_123")
}
```

The subscription resolves a `JSON` payload. For exports, the worker publishes events shaped like this:

```json
{
  "data": {
    "subscribeToImportExportProgress": {
      "type": "EXPORT",
      "userId": "user_123",
      "username": "Casey",
      "projectName": "Q1 Launch",
      "status": "COMPLETE",
      "progress": 100,
      "fileUrl": "https://api.blue.app/uploads/export-1234/q1-launch-export-1430290126.csv",
      "totalCount": 482
    }
  }
}
```

On failure the event carries `"status": "ERROR"`, `"progress": 0`, and an `error` message instead of `fileUrl`/`totalCount`. See the [Import/Export overview](/api/import-export) for the import-side payload.

## Errors

| Code                      | When                                                                                                                                                                     |
| ------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `PROJECT_NOT_FOUND`       | `exportRecords`: the workspace does not exist, or the requesting user is not a member of it (membership is part of the lookup).                                          |
| `REPORT_NOT_FOUND`        | `exportReport`: the report does not exist or the user has no access to it.                                                                                               |
| `CHART_NOT_FOUND`         | `exportChartCSV`: the chart does not exist, or the user is not a member of the organization that owns the dashboard.                                                     |
| `CHART_ALREADY_EXPORTING` | `exportChartCSV`: an export for the same `(chartId, userId)` pair is already in flight.                                                                                  |
| `RATE_LIMITED`            | `exportRecords` / `exportReport`: more than one export request was made within 50 seconds by the same token.                                                             |
| `FORBIDDEN`               | The caller lacks the required role for the operation (see [Permissions](#permissions)), or `exportReport` is called by an organization without the Reports plan feature. |

`exportReport` also throws a generic (untyped) error — `No valid projects found in report data sources` — when the report's data sources resolve to zero accessible workspaces. This surfaces with no `extensions.code`.

## Permissions

| Operation        | Required access                       | Other requirements                                                                      |
| ---------------- | ------------------------------------- | --------------------------------------------------------------------------------------- |
| `exportRecords`  | Project `OWNER`, `ADMIN`, or `MEMBER` | The workspace must be active; rate-limited to 1 request per 50s.                        |
| `exportReport`   | Company `OWNER`, `ADMIN`, or `MEMBER` | The organization must have the Reports plan feature; rate-limited to 1 request per 50s. |
| `exportChartCSV` | Company `OWNER`, `ADMIN`, or `MEMBER` | `VIEW_ONLY` and `COMMENT_ONLY` company members are excluded.                            |

## Notes

- **Delivery is email-only.** Every export here returns a bare `true`; the CSV is generated in the background and a download link is emailed to the requesting user. There is no programmatic way to fetch the file from the API.
- **Checklists are included.** The `exportTodos` CSV carries a `Checklists` column (after `Project`, before custom-field columns) holding each record's checklist items, with `[x]` for done items and `[ ]` for open ones — for example `Topics: Acid-Base [x], Blood Product [ ]`. Records with no checklists leave the cell empty. This applies to the CSV export only; the PDF export is unaffected.
- **Chart export deduplication.** `exportChartCSV` sets a lock keyed by `(chartId, userId)` for 6 hours (21,600 s) on each accepted request. A repeat request for the same chart by the same user inside that window is rejected with `CHART_ALREADY_EXPORTING`. The 6-hour window restarts from the most recently accepted export.
- **Report exports combine data sources.** `exportReport` resolves every workspace referenced by the report's data sources (with per-user permission checks) and merges their filters before queuing one export.

## Related

- [Export CSV template](/api/import-export/export-csv-template) — download a blank CSV with the workspace's columns.
- [Import/Export overview](/api/import-export) — import records and the full progress-event reference.
- [List records](/api/records/list-records) — read records through the API instead of exporting.
- [Dashboards](/api/dashboards) — create the charts that `exportChartCSV` exports.
