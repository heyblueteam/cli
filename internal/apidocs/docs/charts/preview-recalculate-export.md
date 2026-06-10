---
title: Preview, recalculate, and export charts
description: Compute a throwaway chart preview, force-recalculate saved charts, and export chart data to CSV with previewChart, recalculateCharts, and exportChartCSV.
icon: RefreshCw
order: 3
---

Three operations work over chart _results_ rather than chart structure: `previewChart` computes a chart without saving it (for live editor previews), `recalculateCharts` forces saved charts to recompute, and `exportChartCSV` ships a chart's underlying records to CSV. Charts are `Chart` objects in the API, and every chart belongs to a dashboard (`Dashboard`) — these operations are authorized through that dashboard.

To create, edit, or delete charts, see [Create and manage charts](/api/charts/manage-charts). To read or subscribe to saved charts, see [Query and subscribe to charts](/api/charts/query-charts).

<Callout variant="info" title="Results are asynchronous">

Recalculation and CSV export do not return chart data. `recalculateCharts` republishes updated charts over the [`subscribeToChart`](/api/charts/query-charts) subscription; `exportChartCSV` delivers the file out-of-band. Both return immediately.

</Callout>

## previewChart

`previewChart` is a **query**, not a mutation, even though it accepts `CreateChartInput`. It computes a chart and returns it **without persisting anything** — no chart, segment, or segment value is written to the database. Use it to render a live preview in a chart editor before the user commits to saving.

The returned `Chart` is fabricated: it carries a throwaway `id`/`uid`, `position` is always `65535`, and `createdAt`/`updatedAt` are the current time. Don't store or re-reference these — call [`createChart`](/api/charts/manage-charts) to persist the chart for real.

A preview works two ways, mirroring the two kinds of chart:

- **Manual chart** — pass `chartSegments`. Each segment's formula is evaluated against its segment values (each value is an aggregate over one workspace's records).
- **Auto-generated chart** — omit `chartSegments` and pass `metadata.barChart` (an x-axis/y-axis spec) or `metadata.pieChart` (a group-by/value spec). Blue generates the segments for you.

If you supply neither `chartSegments` nor a `metadata.barChart`/`metadata.pieChart` block, the query fails.

### Request

Preview an auto-generated bar chart that counts records per workspace, without saving it.

```graphql
query PreviewBarChart {
  previewChart(
    input: {
      dashboardId: "dashboard_123"
      title: "Records by workspace"
      type: BAR
      metadata: {
        barChart: {
          xAxis: { type: PROJECT, title: "Workspace" }
          yAxis: { function: COUNT, title: "Records" }
        }
      }
    }
  ) {
    id
    title
    type
    position
    chartSegments {
      uid
      title
      formulaResult
    }
  }
}
```

### Parameters

`previewChart` takes the same `CreateChartInput` as [`createChart`](/api/charts/manage-charts). The fields relevant to a preview:

#### CreateChartInput

| Parameter       | Type                   | Required | Description                                                                                           |
| --------------- | ---------------------- | -------- | ----------------------------------------------------------------------------------------------------- |
| `dashboardId`   | `String!`              | Yes      | Dashboard the preview is scoped to. You must be its creator or an `EDITOR`. Nothing is written to it. |
| `title`         | `String!`              | Yes      | Chart title. Echoed back on the preview.                                                              |
| `type`          | `ChartType!`           | Yes      | `STAT`, `PIE`, or `BAR`.                                                                              |
| `chartSegments` | `[ChartSegmentInput!]` | No       | Manual segments. When present, the chart is manual and `metadata` is ignored.                         |
| `metadata`      | `ChartMetadataInput`   | No       | Auto-generation spec. Provide `barChart` or `pieChart` when `chartSegments` is omitted.               |
| `display`       | `FormulaDisplayInput`  | No       | Number/currency/percentage formatting for the displayed value.                                        |
| `position`      | `Float`                | No       | Ignored by the preview — the returned `position` is always `65535`.                                   |

#### ChartType

| Value  | Description                        |
| ------ | ---------------------------------- |
| `STAT` | A single rolled-up number.         |
| `PIE`  | A pie chart of grouped values.     |
| `BAR`  | A bar chart over an x-axis/y-axis. |

#### ChartMetadataInput

| Parameter  | Type                    | Required | Description                                |
| ---------- | ----------------------- | -------- | ------------------------------------------ |
| `barChart` | `BarChartMetadataInput` | No       | Bar-chart spec: an `xAxis` and a `yAxis`.  |
| `pieChart` | `PieChartMetadataInput` | No       | Pie-chart spec: a `groupBy` and a `value`. |

#### BarChartMetadataInput

| Parameter | Type                  | Required | Description                   |
| --------- | --------------------- | -------- | ----------------------------- |
| `xAxis`   | `BarChartXAxisInput!` | Yes      | What the bars are grouped by. |
| `yAxis`   | `BarChartYAxisInput!` | Yes      | What each bar measures.       |

#### BarChartXAxisInput

| Parameter                       | Type                    | Required | Description                                        |
| ------------------------------- | ----------------------- | -------- | -------------------------------------------------- |
| `type`                          | `BarChartXAxisType!`    | Yes      | Dimension to group by (see values below).          |
| `title`                         | `String`                | No       | Axis label.                                        |
| `interval`                      | `BarChartXAxisInterval` | No       | Bucket size for date-based axes (`DAY`…`YEAR`).    |
| `customFieldName`               | `String`                | No       | Field name when `type` is `CUSTOM_FIELD`.          |
| `customFieldType`               | `CustomFieldType`       | No       | Field type when `type` is `CUSTOM_FIELD`.          |
| `customFieldReferenceProjectId` | `String`                | No       | Referenced workspace for a reference custom field. |

#### BarChartYAxisInput

| Parameter                       | Type                         | Required | Description                                        |
| ------------------------------- | ---------------------------- | -------- | -------------------------------------------------- |
| `function`                      | `ChartSegmentValueFunctions` | No       | Aggregate applied per bucket (`COUNT`, `SUM`, …).  |
| `filter`                        | `TodoFilterInput`            | No       | Restrict the records each bar measures.            |
| `title`                         | `String`                     | No       | Axis label.                                        |
| `customFieldName`               | `String`                     | No       | Field name when measuring a custom field.          |
| `customFieldType`               | `CustomFieldType`            | No       | Field type when measuring a custom field.          |
| `customFieldReferenceProjectId` | `String`                     | No       | Referenced workspace for a reference custom field. |

`PieChartMetadataInput` is the same shape with `groupBy` (a `PieChartGroupByInput`, no `interval`) and `value` (a `PieChartValueInput`, identical to `BarChartYAxisInput`).

#### BarChartXAxisType

| Value               | Groups by                                   |
| ------------------- | ------------------------------------------- |
| `PROJECT`           | Workspace.                                  |
| `ASSIGNEE`          | Assigned user.                              |
| `TAG`               | Tag.                                        |
| `CUSTOM_FIELD`      | A custom field's value.                     |
| `TODO_LIST`         | List.                                       |
| `TODO_STATUS`       | Record completion status.                   |
| `TODO_DUE_DATE`     | Due date (bucketed by `interval`).          |
| `TODO_CREATED_AT`   | Creation date (bucketed by `interval`).     |
| `TODO_UPDATED_AT`   | Last-updated date (bucketed by `interval`). |
| `TODO_COMPLETED_AT` | Completion date (bucketed by `interval`).   |

#### ChartSegmentValueFunctions

| Value      | Aggregate                                  |
| ---------- | ------------------------------------------ |
| `COUNT`    | Count of records with a numeric value.     |
| `COUNTA`   | Count of records with any non-empty value. |
| `SUM`      | Sum of values.                             |
| `AVERAGE`  | Mean of numeric values.                    |
| `AVERAGEA` | Mean treating non-numeric values as `0`.   |
| `MIN`      | Smallest value.                            |
| `MAX`      | Largest value.                             |

<Callout variant="info" title="Two function enums, identical members">

`ChartSegmentValueFunctions` (above) applies to a segment value or a metadata axis. The separate `ChartFunction` enum has the same member names but lives only on `FormulaDisplayInput.function` (the segment-level display rollup). They are distinct types — don't substitute one for the other. The manual-chart inputs (`chartSegments`) are documented in full on [Build chart segments and values](/api/charts/chart-segments-and-values).

</Callout>

### Response

```json
{
  "data": {
    "previewChart": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Records by workspace",
      "type": "BAR",
      "position": 65535,
      "chartSegments": [
        { "uid": "ck0seg1", "title": "Sales", "formulaResult": 128 },
        { "uid": "ck0seg2", "title": "Marketing", "formulaResult": 74 }
      ]
    }
  }
}
```

#### Returns

`previewChart` returns a `Chart`. Every field is computed in memory; the chart is not saved.

| Field                     | Type               | Description                                                       |
| ------------------------- | ------------------ | ----------------------------------------------------------------- |
| `id`                      | `ID!`              | Throwaway identifier. Do not persist or reuse it.                 |
| `title`                   | `String!`          | Echoes the input title.                                           |
| `type`                    | `ChartType!`       | `STAT`, `PIE`, or `BAR`.                                          |
| `position`                | `Float!`           | Always `65535` for a preview.                                     |
| `chartSegments`           | `[ChartSegment!]!` | Computed segments. `formulaResult` carries each segment's result. |
| `display`                 | `FormulaDisplay`   | The formatting passed in (or empty).                              |
| `metadata`                | `ChartMetadata`    | The bar/pie spec used to generate the chart, when auto-generated. |
| `createdAt` / `updatedAt` | `DateTime!`        | The current time — not a real persisted timestamp.                |

`ChartSegment.formulaResult` is the per-segment number; for a `STAT` chart there is typically one segment whose `formulaResult` is the headline value.

### Errors

| Code                  | When                                                                                                     |
| --------------------- | -------------------------------------------------------------------------------------------------------- |
| `DASHBOARD_NOT_FOUND` | No dashboard matches `dashboardId` for the caller, or the caller is neither its creator nor an `EDITOR`. |

Two failure modes raise a generic error rather than a coded Blue error: a manual `chartSegments` value that references a workspace the caller doesn't belong to (the preview throws `Access denied to projects: …`), and supplying neither `chartSegments` nor a `metadata.barChart`/`metadata.pieChart` block (there is nothing to compute).

---

## recalculateCharts

`recalculateCharts` forces one or more **saved** charts to recompute their results immediately, then republishes them so subscribers see fresh numbers. Use it after data changes that wouldn't otherwise trigger a recalculation, or to refresh a stale chart on demand.

The call is **all-or-nothing on access**: every id in `chartIds` must resolve to a chart the caller can view. If any one is inaccessible, the whole call is rejected with `FORBIDDEN` and **no** chart is recalculated.

### Request

```graphql
mutation RecalculateCharts {
  recalculateCharts(input: { chartIds: ["chart_123", "chart_456"] })
}
```

`recalculateCharts` returns a nullable `Boolean` (`true` on success), so it takes no sub-selection.

### Parameters

#### RecalculateChartsInput

| Parameter  | Type         | Required | Description                                                                                |
| ---------- | ------------ | -------- | ------------------------------------------------------------------------------------------ |
| `chartIds` | `[String!]!` | Yes      | IDs of saved charts to recalculate. You must have view access to **every** id in the list. |

### Response

```json
{
  "data": {
    "recalculateCharts": true
  }
}
```

Each recalculated chart is republished over [`subscribeToChart`](/api/charts/query-charts) as a `ChartSubscriptionPayload` with `mutation: UPDATED`, carrying the recomputed `isCalculating`/`needCalculation`/`formulaResult` values on `node`. The mutation itself returns only the boolean — read the new numbers from the subscription or by re-querying [`chart`/`charts`](/api/charts/query-charts).

### Errors

| Code        | When                                                                                                           |
| ----------- | -------------------------------------------------------------------------------------------------------------- |
| `FORBIDDEN` | The caller cannot view at least one chart in `chartIds`. The entire call is rejected and nothing recalculates. |

An unknown id is treated the same as an inaccessible one — both reduce the viewable count below the requested count and trigger `FORBIDDEN`.

---

## exportChartCSV

`exportChartCSV` enqueues an asynchronous CSV export of the records behind a chart and returns `true` immediately. The export runs as a background job; the finished CSV is delivered **out-of-band** (emailed / served from the calling host), not in the GraphQL response. Pass an optional `filter` to scope which records are exported.

A per-chart, per-user lock prevents duplicate exports: once you start an export, a second `exportChartCSV` call for the **same chart by the same user** fails with `CHART_ALREADY_EXPORTING` until the lock expires after **6 hours** (21,600 seconds).

### Request

Export every record behind a chart.

```graphql
mutation ExportChartCSV {
  exportChartCSV(chartId: "chart_123")
}
```

`exportChartCSV` returns `Boolean!`, so it takes no sub-selection. `true` means the job was accepted — not that the file is ready.

### Parameters

| Parameter | Type              | Required | Description                                                                     |
| --------- | ----------------- | -------- | ------------------------------------------------------------------------------- |
| `chartId` | `ID!`             | Yes      | Chart whose records to export. Must belong to a dashboard in your organization. |
| `filter`  | `TodoFilterInput` | No       | Restrict the exported rows. Omit to export all records behind the chart.        |

`TodoFilterInput` is the standard record filter used across the API; see [List records](/api/records/list-records) for its full field set.

### Response

```json
{
  "data": {
    "exportChartCSV": true
  }
}
```

The boolean confirms the job was enqueued. The CSV arrives separately once the worker finishes; there is no operation that returns the file URL synchronously.

### Full example

Export only the records that match a filter — here, records on a specific list.

```graphql
mutation ExportFilteredChartCSV {
  exportChartCSV(chartId: "chart_123", filter: { todoListIds: ["list_123"] })
}
```

### Errors

| Code                      | When                                                                                                                  |
| ------------------------- | --------------------------------------------------------------------------------------------------------------------- |
| `CHART_NOT_FOUND`         | No chart matches `chartId`, or the chart's dashboard is in an organization the caller is not a member of.             |
| `CHART_ALREADY_EXPORTING` | An export for the same chart by the same user is already in flight. The lock clears 6 hours after the export started. |

```json
{
  "errors": [
    {
      "message": "Chart is already being exported.",
      "extensions": { "code": "CHART_ALREADY_EXPORTING" }
    }
  ]
}
```

<Callout variant="warning" title="The export lock is per user, per chart, for 6 hours">

The dedupe lock is keyed on `chartId` + the calling user and lives in Redis for 21,600 seconds. A second export of the same chart by the same user is blocked for up to 6 hours; a different user, or a different chart, is unaffected. There is no operation to release the lock early.

</Callout>

## Permissions

All three operations derive their permissions from the chart's parent dashboard. There is no organization- or workspace-level chart access — a chart is reachable only through a dashboard you can see.

| Operation           | Access required                                                                                    |
| ------------------- | -------------------------------------------------------------------------------------------------- |
| `previewChart`      | Creator of the target dashboard, or an `EDITOR` on it. A `VIEWER` cannot preview.                  |
| `recalculateCharts` | View access (creator or any dashboard user — `EDITOR` or `VIEWER`) to **every** chart's dashboard. |
| `exportChartCSV`    | Membership in the organization that owns the chart's dashboard.                                    |

See [Dashboards](/api/dashboards) for how dashboard creators and `dashboardUsers` (with `EDITOR`/`VIEWER` roles) are managed.

## Related

- [Create and manage charts](/api/charts/manage-charts) — persist a chart with `createChart`; `previewChart` shares its `CreateChartInput`.
- [Build chart segments and values](/api/charts/chart-segments-and-values) — the manual `chartSegments`/`chartSegmentValues` building blocks.
- [Query and subscribe to charts](/api/charts/query-charts) — read recalculated results and stream `UPDATED` events over `subscribeToChart`.
- [Charts](/api/charts) — section overview.
- [Dashboards](/api/dashboards) — the parent objects charts and their permissions hang off.
- [List records](/api/records/list-records) — the `TodoFilterInput` used by `exportChartCSV`.
