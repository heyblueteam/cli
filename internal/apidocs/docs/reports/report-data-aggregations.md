---
title: Report Data & Aggregations
description: Read a report's combined records, counts, and per-field sum/avg/min/max aggregations, and refresh its cached results.
icon: ChartColumn
order: 3
---

A report is a query surface as well as a saved configuration. Once a report has data sources, the `Report` type exposes three read fields — `todos` (the combined record list), `todoCount` (the matching count), and `aggregations` (per-field statistics) — plus the `refreshReportAggregations` mutation, which invalidates the cached aggregation results. Records are `Todo` objects in the API.

All three read fields run with the **report creator's** workspace permissions (not the caller's) and combine every data source on the report, so the data you read is identical for every viewer regardless of which workspaces _they_ can see. See [Query Reports](/api/reports/query-reports) for reading a report's metadata and configuration, and [Create a Report](/api/reports/create-report) for setting up its data sources.

## Request

Read the combined record list, count, and a sum/avg over a numeric custom field in one query:

```graphql
query ReadReportData {
  report(id: "report_123") {
    todoCount
    todos(limit: 50, sort: [duedAt_ASC]) {
      id
      title
      done
      duedAt
      users {
        fullName
      }
    }
    aggregations(fields: [{ field: "field_123", fieldType: "number" }]) {
      field
      fieldName
      sum
      avg
      min
      max
      count
    }
  }
}
```

`report` is identified by the report's `id`. `todos`, `todoCount`, and `aggregations` are fields _on_ the resolved `Report`, so they are always nested under a `report` (or `reports`) query.

## Parameters

These fields share the records `TodosFilter` and `TodosSort` inputs rather than redefining them. See the [Records](/api/records) section for the full field-by-field reference of those inputs — only the report-specific behavior is documented here.

### `Report.todos` arguments

| Parameter | Type            | Required | Description                                                                                                    |
| --------- | --------------- | -------- | -------------------------------------------------------------------------------------------------------------- |
| `filter`  | `TodosFilter`   | No       | Extra record filter applied on top of each data source's stored filter. Its `projectIds` are ignored — workspace scope comes from the data sources, not from this filter. |
| `sort`    | `[TodosSort!]`  | No       | Sort order. Accepts the records `TodosSort` enum values (e.g. `duedAt_ASC`, `title_DESC`, `createdAt_DESC`).    |
| `limit`   | `Int`           | No       | Maximum records to return. Defaults to **100**; capped at a hard maximum of **5000**. Values above 5000 are clamped, not rejected. |
| `skip`    | `Int`           | No       | Number of records to skip before returning results, for paging through large reports.                          |

### `Report.todoCount` arguments

| Parameter | Type          | Required | Description                                                                              |
| --------- | ------------- | -------- | ---------------------------------------------------------------------------------------- |
| `filter`  | `TodosFilter` | No       | Same extra filter as `todos`. Returns the count of matching records across all data sources. |

### `Report.aggregations` arguments

| Parameter | Type                          | Required | Description                                                            |
| --------- | ----------------------------- | -------- | --------------------------------------------------------------------- |
| `fields`  | `[AggregationFieldInput!]!`   | Yes      | The fields to aggregate. One `FieldAggregation` is returned per entry. |
| `filter`  | `TodosFilter`                 | No       | Extra record filter, applied on top of each data source's filter before aggregating. |

### AggregationFieldInput

| Field         | Type         | Required | Description                                                                                                                                  |
| ------------- | ------------ | -------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `field`       | `String!`    | Yes      | The field to aggregate. A custom field's `id`, a default field key (e.g. `timeTracking.totalDuration`), or a `merged_…` key for merged fields. |
| `originalIds` | `[String!]`  | No       | For a merged/mapped field, the underlying custom field IDs it spans. Each record contributes at most one value across them.                  |
| `fieldType`   | `String`     | No       | The field's type hint, echoed back on the result as `FieldAggregation.fieldType`.                                                            |
| `footerType`  | `String`     | No       | Display hint passed through from the app's report footer config. Does not change the computed statistics.                                    |

<Callout variant="info" title="Aggregations are numeric">

`sum`, `avg`, `min`, and `max` are computed over **numeric** values only — number custom fields and numeric default fields such as time tracking. Non-numeric values are skipped, so a field with no numeric data returns `count: 0` and `null` for every statistic.

</Callout>

## Response

```json
{
  "data": {
    "report": {
      "todoCount": 128,
      "todos": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "title": "Renew enterprise contract",
          "done": false,
          "duedAt": "2026-06-02T00:00:00.000Z",
          "users": [{ "fullName": "Dana Reyes" }]
        },
        {
          "id": "clm4r2abc000108l0aaaa1111",
          "title": "Q3 onboarding kickoff",
          "done": true,
          "duedAt": "2026-06-05T00:00:00.000Z",
          "users": []
        }
      ],
      "aggregations": [
        {
          "field": "field_123",
          "fieldName": "Deal Value",
          "sum": 184500,
          "avg": 1441.4,
          "min": 0,
          "max": 50000,
          "count": 128
        }
      ]
    }
  }
}
```

### FieldAggregation

| Field       | Type      | Description                                                                                            |
| ----------- | --------- | ------------------------------------------------------------------------------------------------------ |
| `field`     | `String!` | Echoes the `field` from the matching `AggregationFieldInput`, so you can correlate inputs to results.  |
| `fieldName` | `String`  | Human-readable name derived from the field. `null` when it cannot be resolved.                         |
| `fieldType` | `String`  | The `fieldType` hint passed in, echoed back. `null` when not supplied.                                 |
| `sum`       | `Float`   | Sum of the numeric values. `null` when no numeric values matched.                                      |
| `avg`       | `Float`   | Mean of the numeric values. `null` when no numeric values matched.                                     |
| `min`       | `Float`   | Smallest numeric value. `null` when no numeric values matched.                                         |
| `max`       | `Float`   | Largest numeric value. `null` when no numeric values matched.                                          |
| `count`     | `Int!`    | Number of records that contributed a numeric value. `0` when none matched.                             |

`todos` returns `[Todo!]!` — select any field on `Todo` (see the [Records](/api/records) section for the full type). `todoCount` returns a bare `Int!`. `aggregations` returns `[FieldAggregation!]!`, one entry per requested field, in the same order.

## Empty reports

A report with **no data sources** is valid but has nothing to read:

- `todos` returns an empty list `[]`.
- `todoCount` returns `0`.
- `aggregations` returns an empty list `[]`.

No error is raised in any of these cases. Add data sources with [Create a Report](/api/reports/create-report) or [Manage Report Access](/api/reports/manage-report-access) before expecting results.

## Refreshing cached aggregations

`aggregations` results are cached in Redis (in 5-minute buckets) to keep large reports fast. Use the `refreshReportAggregations` mutation to discard that cache when the underlying records have changed and you need the next read to recompute.

```graphql
mutation RefreshReportAggregations {
  refreshReportAggregations(reportId: "report_123") {
    id
    lastGeneratedAt
  }
}
```

<Callout variant="warning" title="Refresh is not a synchronous recompute">

`refreshReportAggregations` does **not** recompute aggregations itself. It invalidates the report's cached aggregation results and sets `lastGeneratedAt` to the current time, then returns the updated `Report`. The numbers are recomputed lazily on the **next** `aggregations` query against the report. The returned `Report` does not contain fresh aggregation values — query `aggregations` afterwards to read them.

</Callout>

### `refreshReportAggregations` arguments

| Parameter  | Type      | Required | Description                                       |
| ---------- | --------- | -------- | ------------------------------------------------- |
| `reportId` | `String!` | Yes      | The report whose aggregation cache to invalidate. |

```json
{
  "data": {
    "refreshReportAggregations": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "lastGeneratedAt": "2026-05-29T12:00:00.000Z"
    }
  }
}
```

## Errors

| Code               | When                                                                                     |
| ------------------ | ---------------------------------------------------------------------------------------- |
| `REPORT_NOT_FOUND` | The `report`/`reportId` does not exist, or you are neither the creator nor a shared user of the report. |
| `UNAUTHENTICATED`  | The request has no valid credentials.                                                    |
| `FORBIDDEN`        | The request is authenticated but carries no organization context.                        |

Reading data via `report { todos / todoCount / aggregations }` is gated by the parent `report` query's visibility check, which raises `REPORT_NOT_FOUND` when the report is invisible to you. `refreshReportAggregations` performs the same view-access check before touching the cache.

## Permissions

- **Reading `todos` / `todoCount` / `aggregations`** — requires view access to the parent report: you must be its `createdBy` or appear in its `reportUsers` (either `EDITOR` or `VIEWER`). The records themselves are resolved with the **report creator's** workspace permissions, so a `VIEWER` who cannot see a given workspace directly still sees its records through the report.
- **`refreshReportAggregations`** — requires only **view** access (creator or any shared user, `EDITOR` or `VIEWER`). It is intentionally looser than modifying the report; invalidating the cache is a read-side operation.

Modifying or deleting a report requires higher tiers — see [Manage Report Access](/api/reports/manage-report-access) and [Duplicate & Delete a Report](/api/reports/duplicate-and-delete-report).

## Related

- [Reports overview](/api/reports)
- [Create a Report](/api/reports/create-report)
- [Query Reports](/api/reports/query-reports)
- [Manage Report Access](/api/reports/manage-report-access)
- [Export a Report](/api/reports/export-report)
- [Records](/api/records)
