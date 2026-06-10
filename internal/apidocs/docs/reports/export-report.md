---
title: Export a Report
description: Queue a CSV export of a report's combined record set and email it to yourself with the exportReport mutation.
icon: Download
order: 6
---

Use the `exportReport` mutation to export a report's combined record set to CSV. A report aggregates records (`Todo` objects) from one or more workspaces; this mutation gathers the records across **every** data source — applying the merged filters and the report creator's workspace permissions — and delivers the result as a CSV.

The export runs in the background. `exportReport` returns `Boolean!` immediately after queuing the job; it does **not** stream or return the CSV inline. When the job finishes, the file is emailed to the calling user's address. To read a report's records over the API instead of by email, use the record list on [Report Data & Aggregations](/api/reports/report-data-aggregations).

## Request

```graphql
mutation ExportReport {
  exportReport(input: { reportId: "report_123" })
}
```

`exportReport` returns a bare `Boolean!`, so the call takes no sub-selection — a `true` result means the export job was queued, not that the CSV has been built or sent.

## Parameters

### ExportReportInput

| Parameter  | Type      | Required | Description                       |
| ---------- | --------- | -------- | --------------------------------- |
| `reportId` | `String!` | Yes      | The `id` of the report to export. |

## Response

```json
{
  "data": {
    "exportReport": true
  }
}
```

`true` confirms the background job was enqueued. The CSV is generated asynchronously and emailed to the address on the calling user's account; nothing is returned in the GraphQL response body.

## How the export is built

When the job runs, it resolves the record set the same way the rest of the report does:

- **Projects** — it unions the workspace IDs across all of the report's data sources, resolved against the **report creator's** permissions (not the caller's). A data source with `projectIds: null` contributes every workspace the creator can access; a data source with a specific array contributes those workspaces. Archived workspaces are excluded. This is the same computed set exposed by `Report.projectIds`.
- **Filters** — it merges the `filters` blob from each data source into a single record filter (the first source to set a given key wins). The CSV then contains the records from the unioned workspaces that match the merged filter.

<Callout variant="info" title="Permissions follow the creator, not the caller">

Like the rest of a report's data surface, the export resolves workspaces with the report **creator's** permissions. A colleague who can export a report still sees the records the creator can see, even if their own workspace access differs.

</Callout>

## Errors

| Code               | When                                                                                                     |
| ------------------ | -------------------------------------------------------------------------------------------------------- |
| `REPORT_NOT_FOUND` | The `reportId` does not exist, or you do not have export access to it (see [Permissions](#permissions)). |
| `UNAUTHENTICATED`  | The request has no valid credentials.                                                                    |

The export also fails with the message `No valid projects found in report data sources` when the report resolves to an empty workspace set — for example, every data source points at workspaces the creator can no longer access, or all referenced workspaces are archived. Fix this by [updating the report's data sources](/api/reports/manage-report-access) to reference workspaces the creator can access. This is a plain error without a dedicated extension code.

## Permissions

Export access is intentionally **broader** than the other report mutations. `exportReport` succeeds if you are any of:

- the report's `createdBy` (creator), **or**
- a user the report is shared with (any entry in `reportUsers`, `EDITOR` or `VIEWER`), **or**
- any member of the report's organization (a `companyUsers` entry on the report's `Company`).

That third tier is unique to export — every member of the organization can email themselves a report's records, whereas viewing, modifying, and deleting are gated more tightly:

| Operation                                                           | Who can call it                                          |
| ------------------------------------------------------------------- | -------------------------------------------------------- |
| **Export** (`exportReport`)                                         | Creator, any shared user, **or** any organization member |
| **View** (`report`, `duplicateReport`, `refreshReportAggregations`) | Creator or any shared user                               |
| **Modify** (`updateReport`)                                         | Creator or a shared user with the `EDITOR` role          |
| **Delete** (`deleteReport`)                                         | Creator only                                             |

If you fall outside the export tier, the mutation raises `REPORT_NOT_FOUND` rather than a distinct authorization error, so a missing report and an inaccessible one are indistinguishable.

## Related

- [Reports overview](/api/reports)
- [Report Data & Aggregations](/api/reports/report-data-aggregations)
- [Query Reports](/api/reports/query-reports)
- [Manage Report Access](/api/reports/manage-report-access)
- [Duplicate & Delete a Report](/api/reports/duplicate-and-delete-report)
- [Export Records](/api/import-export/export-records)
- [Records](/api/records)
