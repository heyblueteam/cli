---
title: Reports
description: Aggregate records across one or more workspaces into a single saved analytical view with counts, field aggregations, and CSV export.
icon: FileChartColumn
order: 0
---

Reports aggregate records across one or more workspaces into a single saved analytical view. A report holds one or more **data sources**, each scoping a set of workspaces and a JSON filter, and exposes record lists, counts, and field aggregations (sum, avg, min, max, count) computed across every source. Use this section to create and configure reports, read the data and aggregations they produce, share them with `EDITOR` and `VIEWER` collaborators, duplicate and delete them, and export the combined record set to CSV.

Reports are distinct from [Dashboards](/api/dashboards): a report is a record-level query surface, while a dashboard is a visualization container.

## Concepts to types

| Docs word           | Schema type / field                                        |
| ------------------- | ---------------------------------------------------------- |
| Report              | `Report`                                                   |
| Data source         | `ReportDataSource`                                         |
| Shared collaborator | `ReportUser` (role `ReportRole`: `EDITOR` / `VIEWER`)      |
| Field aggregation   | `FieldAggregation` (requested via `AggregationFieldInput`) |
| Record              | `Todo`                                                     |
| Workspace           | `Project`                                                  |
| Organization        | `Company`                                                  |

A report's data sources currently support a single `sourceType`, `TODOS` — every report aggregates records (`Todo` objects). The `config` and `filters` fields are opaque `JSONObject` blobs whose internal shape is defined by the app, not the GraphQL schema; the filters mirror the [records filter](/api/records/list-records) structure.

## Pages in this section

| Page                                                                    | Purpose                                                                                                    |
| ----------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------- |
| [Create a Report](/api/reports/create-report)                           | Create a report scoped to one or more data sources with `createReport`.                                    |
| [Query Reports](/api/reports/query-reports)                             | Fetch a single report (`report`) or list a company's reports (`reports`).                                  |
| [Report Data & Aggregations](/api/reports/report-data-aggregations)     | Read a report's records, counts, and field aggregations, and refresh cached aggregation results.           |
| [Manage Report Access](/api/reports/manage-report-access)               | Update a report's config and data sources and share it with collaborators via `updateReport`.              |
| [Duplicate & Delete a Report](/api/reports/duplicate-and-delete-report) | Copy a report with `duplicateReport` or permanently remove it with `deleteReport`.                         |
| [Export a Report](/api/reports/export-report)                           | Queue an async CSV export of the report's combined record set, emailed to the caller, with `exportReport`. |

## Access at a glance

Report operations enforce four distinct permission tiers — see each page for the exact rule.

| Tier   | Who                                                                          | Operations                                               |
| ------ | ---------------------------------------------------------------------------- | -------------------------------------------------------- |
| View   | Creator or any shared `ReportUser` (`EDITOR` or `VIEWER`)                    | `report`, `duplicateReport`, `refreshReportAggregations` |
| Modify | Creator or shared `ReportUser` with role `EDITOR`                            | `updateReport`                                           |
| Delete | Creator only                                                                 | `deleteReport`                                           |
| Export | Creator, any shared `ReportUser`, or any member of the report's organization | `exportReport`                                           |

<Callout variant="info" title="Export is intentionally broader">

`exportReport` is reachable by any member of the report's organization, not just the creator and shared users. This is looser than every other report operation by design.

</Callout>
