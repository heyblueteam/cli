---
title: Create and manage charts
description: Create STAT, PIE, and BAR charts on a dashboard, then rename, reposition, restyle, duplicate, and delete them.
icon: ChartColumn
order: 1
---

A chart is a single card on a [dashboard](/api/dashboards). This page covers the full lifecycle: create a chart with `createChart`, change its title, position, display, or metadata with `editChart`, duplicate it into the same or another dashboard with `copyChart`, and remove it with `deleteChart`. Charts are `Chart` objects in the API.

Every chart is either a **manual** chart (driven by `chartSegments` — formulas over project-scoped values) or an **auto-generated** chart (driven by `metadata.barChart` or `metadata.pieChart`). `createChart` accepts either shape; the [segment and value mutations](/api/charts/chart-segments-and-values) apply only to manual charts. See the [section overview](/api/charts) for the full type map.

<Callout variant="warning" title="Results are calculated asynchronously">

`createChart` and `copyChart` enqueue a background calculation rather than returning final chart data. The returned `Chart` reflects pending state via `isCalculating`, `needCalculation`, and `isCalculatingWithFilter`; final results arrive through the [`subscribeToChart`](/api/charts/query-charts) subscription, not synchronously from the mutation.

</Callout>

## createChart

Use the `createChart` mutation to add a chart to a dashboard.

### Request

The smallest call creates an empty `STAT` chart. You build out its segments afterward with [`createChartSegment`](/api/charts/chart-segments-and-values).

```graphql
mutation CreateChart {
  createChart(input: { dashboardId: "dashboard_123", title: "Open records", type: STAT }) {
    id
    title
    type
    position
    isCalculating
    needCalculation
  }
}
```

### Parameters

#### CreateChartInput

| Parameter       | Type                   | Required | Description                                                                                                                            |
| --------------- | ---------------------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| `dashboardId`   | `String!`              | Yes      | Dashboard the chart belongs to. You must be its creator or hold the `EDITOR` role.                                                     |
| `title`         | `String!`              | Yes      | Display title of the chart.                                                                                                            |
| `type`          | `ChartType!`           | Yes      | `STAT`, `PIE`, or `BAR`.                                                                                                               |
| `position`      | `Float`                | No       | Sort position within the dashboard. Omit to append after the last chart (see [Position defaults](#position-defaults)).                 |
| `display`       | `FormulaDisplayInput`  | No       | How the chart's value is formatted (number/currency/percentage) and the optional display rollup function.                              |
| `chartSegments` | `[ChartSegmentInput!]` | No       | Inline segments for a **manual** chart. Mutually exclusive with `metadata`. Each segment carries a formula and its values — see below. |
| `metadata`      | `ChartMetadataInput`   | No       | A `barChart` or `pieChart` spec for an **auto-generated** chart. Mutually exclusive with `chartSegments`.                              |

#### ChartType

| Value  | Description                                                          |
| ------ | -------------------------------------------------------------------- |
| `STAT` | A single headline number (one segment value, or a formula rollup).   |
| `PIE`  | A pie chart — either manual segments or a `pieChart` group-by/value. |
| `BAR`  | A bar chart — either manual segments or a `barChart` x-axis/y-axis.  |

#### FormulaDisplayInput

| Parameter   | Type                          | Required | Description                                                                                                 |
| ----------- | ----------------------------- | -------- | ----------------------------------------------------------------------------------------------------------- |
| `type`      | `FormulaDisplayType!`         | Yes      | `NUMBER`, `CURRENCY`, or `PERCENTAGE`.                                                                      |
| `currency`  | `FormulaDisplayCurrencyInput` | No       | `{ code, name }`, e.g. `{ code: "USD", name: "US Dollar" }`. Use with `CURRENCY`.                           |
| `precision` | `Float`                       | No       | Number of decimal places to show.                                                                           |
| `function`  | `ChartFunction`               | No       | Display rollup applied across the chart. See [ChartFunction](#chartfunction-vs-chartsegmentvaluefunctions). |

#### ChartMetadataInput

Supply exactly one of `barChart` or `pieChart` for an auto-generated chart.

| Parameter  | Type                    | Required | Description                                  |
| ---------- | ----------------------- | -------- | -------------------------------------------- |
| `barChart` | `BarChartMetadataInput` | No       | An x-axis/y-axis spec. Use for `BAR` charts. |
| `pieChart` | `PieChartMetadataInput` | No       | A group-by/value spec. Use for `PIE` charts. |

`BarChartMetadataInput` is `{ xAxis: BarChartXAxisInput!, yAxis: BarChartYAxisInput! }`; `PieChartMetadataInput` is `{ groupBy: PieChartGroupByInput!, value: PieChartValueInput! }`. The axis inputs share a common shape:

| Field                           | Type                         | Where           | Description                                                                                                                                                                      |
| ------------------------------- | ---------------------------- | --------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `title`                         | `String`                     | all axes        | Optional axis label.                                                                                                                                                             |
| `type`                          | `BarChartXAxisType!`         | xAxis / groupBy | What to bucket records by: `PROJECT`, `ASSIGNEE`, `TAG`, `CUSTOM_FIELD`, `TODO_LIST`, `TODO_STATUS`, `TODO_DUE_DATE`, `TODO_CREATED_AT`, `TODO_UPDATED_AT`, `TODO_COMPLETED_AT`. |
| `interval`                      | `BarChartXAxisInterval`      | xAxis only      | Bucket size for date axes: `DAY`, `WEEK`, `MONTH`, `QUARTER`, `YEAR`.                                                                                                            |
| `function`                      | `ChartSegmentValueFunctions` | yAxis / value   | Aggregate applied to the measured records.                                                                                                                                       |
| `filter`                        | `TodoFilterInput`            | yAxis / value   | Narrows which records are measured. Same filter used by [list records](/api/records/list-records).                                                                               |
| `customFieldName`               | `String`                     | all axes        | Custom-field name when `type` is `CUSTOM_FIELD`.                                                                                                                                 |
| `customFieldType`               | `CustomFieldType`            | all axes        | Type of that custom field.                                                                                                                                                       |
| `customFieldReferenceProjectId` | `String`                     | all axes        | For reference custom fields, the referenced project.                                                                                                                             |

For `chartSegments` (manual charts), the segment and value input fields are documented on [Build chart segments and values](/api/charts/chart-segments-and-values); the example below shows the inline form.

### Response

```json
{
  "data": {
    "createChart": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Open records",
      "type": "STAT",
      "position": 65535,
      "isCalculating": true,
      "needCalculation": true
    }
  }
}
```

#### Returns

`createChart` returns the new `Chart`.

| Field                     | Type               | Description                                                                            |
| ------------------------- | ------------------ | -------------------------------------------------------------------------------------- |
| `id`                      | `ID!`              | Unique identifier of the chart.                                                        |
| `title`                   | `String!`          | Chart title.                                                                           |
| `position`                | `Float!`           | Sort position within the dashboard.                                                    |
| `type`                    | `ChartType!`       | `STAT`, `PIE`, or `BAR`.                                                               |
| `chartSegments`           | `[ChartSegment!]!` | Manual segments. Empty for an auto-generated chart.                                    |
| `metadata`                | `ChartMetadata`    | The `barChart`/`pieChart` spec for an auto-generated chart. `null` for a manual chart. |
| `display`                 | `FormulaDisplay`   | Display formatting (`type`, `currency`, `precision`, `function`).                      |
| `isCalculating`           | `Boolean`          | A recompute is currently in progress.                                                  |
| `needCalculation`         | `Boolean`          | Results are stale or pending — a calculation is queued.                                |
| `isCalculatingWithFilter` | `Boolean`          | Set transiently when the chart is read through `charts(filter.todoFilter)`.            |
| `createdAt`               | `DateTime!`        | When the chart was created.                                                            |
| `updatedAt`               | `DateTime!`        | When the chart was last modified.                                                      |

### Full example

Create an auto-generated `BAR` chart that counts records by workspace, formatted as a whole number.

```graphql
mutation CreateBarChart {
  createChart(
    input: {
      dashboardId: "dashboard_123"
      title: "Records per workspace"
      type: BAR
      display: { type: NUMBER, precision: 0 }
      metadata: {
        barChart: {
          xAxis: { title: "Workspace", type: PROJECT }
          yAxis: { title: "Records", function: COUNT }
        }
      }
    }
  ) {
    id
    title
    type
    metadata {
      ... on ChartMetadataBarChart {
        barChart {
          xAxis {
            title
            type
          }
          yAxis {
            title
            function
          }
        }
      }
    }
    needCalculation
  }
}
```

To create a manual chart instead, pass `chartSegments` (and omit `metadata`). Each segment carries a `formula` whose `logic.text`/`logic.html` reference its values by their `uid`:

```graphql
mutation CreateManualChart {
  createChart(
    input: {
      dashboardId: "dashboard_123"
      title: "Won vs. lost"
      type: PIE
      chartSegments: [
        {
          title: "Won deals"
          color: "#22c55e"
          uid: "seg_won"
          formula: {
            logic: { text: "val_won", html: "<span>val_won</span>" }
            display: { type: NUMBER }
          }
          chartSegmentValues: [
            { uid: "val_won", title: "Won", projectId: "project_123", function: COUNT }
          ]
        }
      ]
    }
  ) {
    id
    title
    chartSegments {
      id
      uid
      title
    }
  }
}
```

### Errors

| Code                  | When                                                                                                                 |
| --------------------- | -------------------------------------------------------------------------------------------------------------------- |
| `DASHBOARD_NOT_FOUND` | No dashboard matches `dashboardId` for the caller, including when the caller lacks creator or `EDITOR` access to it. |

```json
{
  "errors": [
    {
      "message": "Dashboard was not found.",
      "extensions": { "code": "DASHBOARD_NOT_FOUND" }
    }
  ]
}
```

## editChart

Use the `editChart` mutation to rename, reposition, restyle, or re-spec a chart. Only the fields you pass are changed; omitted fields are left untouched. Segments and values are **not** edited here — use the [segment and value mutations](/api/charts/chart-segments-and-values).

### Request

```graphql
mutation EditChart {
  editChart(input: { id: "chart_123", title: "Open records (this quarter)" }) {
    id
    title
    updatedAt
  }
}
```

### Parameters

#### EditChartInput

| Parameter  | Type                  | Required | Description                                                                                       |
| ---------- | --------------------- | -------- | ------------------------------------------------------------------------------------------------- |
| `id`       | `String!`             | Yes      | ID of the chart to edit.                                                                          |
| `title`    | `String`              | No       | New title.                                                                                        |
| `position` | `Float`               | No       | New sort position within the dashboard.                                                           |
| `display`  | `FormulaDisplayInput` | No       | New display formatting. Replaces the existing `display`.                                          |
| `metadata` | `ChartMetadataInput`  | No       | New `barChart`/`pieChart` spec. Changing metadata re-triggers an asynchronous result calculation. |

<Callout variant="info" title="Editing metadata recalculates; editing title/position/display does not">

`editChart` only enqueues a recalculation when the input includes `metadata.barChart` or `metadata.pieChart`. Renaming, repositioning, or restyling a chart updates it and publishes the change to subscribers, but does not recompute its data. To force a recompute without changing the spec, use [`recalculateCharts`](/api/charts/preview-recalculate-export).

</Callout>

### Response

```json
{
  "data": {
    "editChart": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Open records (this quarter)",
      "updatedAt": "2026-05-29T14:07:42.000Z"
    }
  }
}
```

`editChart` returns the updated `Chart` (same shape as [`createChart`](#returns)).

### Errors

| Code              | When                                                                                                               |
| ----------------- | ------------------------------------------------------------------------------------------------------------------ |
| `CHART_NOT_FOUND` | No chart matches `id` for the caller, including when the caller lacks creator or `EDITOR` access to its dashboard. |

## copyChart

Use the `copyChart` mutation to duplicate a chart into the same dashboard or a different one. The copy is a deep copy: segments, values, formulas, type, display, and metadata are recreated with fresh identifiers.

### Request

```graphql
mutation CopyChart {
  copyChart(input: { chartId: "chart_123", dashboardId: "dashboard_456" }) {
    id
    title
    type
    position
  }
}
```

### Parameters

#### CopyChartInput

| Parameter     | Type      | Required | Description                                                                               |
| ------------- | --------- | -------- | ----------------------------------------------------------------------------------------- |
| `chartId`     | `String!` | Yes      | Chart to copy. You must be the creator or an `EDITOR` of its dashboard.                   |
| `dashboardId` | `String!` | Yes      | Destination dashboard. You must be its creator or an `EDITOR`. May be the same dashboard. |
| `title`       | `String`  | No       | Title for the copy. If omitted, the copy keeps the original chart's title.                |

### Response

```json
{
  "data": {
    "copyChart": {
      "id": "clm4n8qwx000409l0d1weq2bk",
      "title": "Open records",
      "type": "STAT",
      "position": 131070
    }
  }
}
```

`copyChart` returns the newly created `Chart`.

<Callout variant="info" title="Formula UIDs are remapped">

For a manual chart, each copied `ChartSegmentValue` gets a fresh `uid`, and the new UID is string-replaced into the copied segment's `formula.logic.text` and `formula.logic.html`. The formula keeps working against the copied values rather than pointing back at the originals. The formula-UID linkage is detailed on [Build chart segments and values](/api/charts/chart-segments-and-values).

</Callout>

The copy is appended after the last chart in the destination dashboard (see [Position defaults](#position-defaults)) and its results are recalculated asynchronously, so it briefly reports `needCalculation: true`.

### Errors

| Code                  | When                                                                                                                       |
| --------------------- | -------------------------------------------------------------------------------------------------------------------------- |
| `DASHBOARD_NOT_FOUND` | No destination dashboard matches `dashboardId` for the caller, or the caller lacks creator or `EDITOR` access to it.       |
| `CHART_NOT_FOUND`     | No chart matches `chartId` for the caller, or the caller lacks creator or `EDITOR` access to the source chart's dashboard. |

## deleteChart

Use the `deleteChart` mutation to remove a chart and its segments. `deleteChart` takes the chart `id` directly, not an input object.

### Request

```graphql
mutation DeleteChart {
  deleteChart(id: "chart_123") {
    success
  }
}
```

### Parameters

| Parameter | Type      | Required | Description                |
| --------- | --------- | -------- | -------------------------- |
| `id`      | `String!` | Yes      | ID of the chart to delete. |

### Response

`deleteChart` returns a `MutationResult`. It exposes only `success` and `operationId` — there is no `message` field.

```json
{
  "data": {
    "deleteChart": {
      "success": true
    }
  }
}
```

#### MutationResult

| Field         | Type       | Description                                                    |
| ------------- | ---------- | -------------------------------------------------------------- |
| `success`     | `Boolean!` | `true` if the chart was deleted; `false` if the delete failed. |
| `operationId` | `String`   | Correlation ID for the operation, when set.                    |

### Errors

| Code              | When                                                                                                               |
| ----------------- | ------------------------------------------------------------------------------------------------------------------ |
| `CHART_NOT_FOUND` | No chart matches `id` for the caller, including when the caller lacks creator or `EDITOR` access to its dashboard. |

If the chart is found but the delete itself fails, the mutation returns `{ "success": false }` rather than throwing.

## ChartFunction vs. ChartSegmentValueFunctions

Two enums share the same seven members but live in different places — keep them straight in examples:

| Enum                         | Where it appears                                                    | Members                                                       |
| ---------------------------- | ------------------------------------------------------------------- | ------------------------------------------------------------- |
| `ChartFunction`              | `FormulaDisplayInput.function` (the chart-level display rollup)     | `COUNT`, `COUNTA`, `SUM`, `AVERAGE`, `AVERAGEA`, `MIN`, `MAX` |
| `ChartSegmentValueFunctions` | Each segment value and each `barChart`/`pieChart` axis (`function`) | `COUNT`, `COUNTA`, `SUM`, `AVERAGE`, `AVERAGEA`, `MIN`, `MAX` |

## Position defaults

`position` is a `Float!` controlling order within the dashboard. When you omit `position` on `createChart` (and always on `copyChart`), the server appends the chart after the current last one: it takes the highest existing `position` and adds a fixed gap, or uses that gap as the starting position if the dashboard has no charts yet. Pass an explicit `position` to slot a chart between two others. Reorder later with [`editChart`](#editchart).

## Permissions

All four mutations are writes and require the caller to either have created the chart's dashboard or hold the `EDITOR` role on it. `copyChart` checks this on **both** the source chart's dashboard and the destination dashboard.

| Access to the dashboard | Can create / edit / copy / delete |
| ----------------------- | --------------------------------- |
| Creator                 | Yes                               |
| `EDITOR`                | Yes                               |
| `VIEWER`                | No                                |
| No access               | No                                |

A caller without sufficient access receives `CHART_NOT_FOUND` (or `DASHBOARD_NOT_FOUND` for the dashboard argument) rather than a distinct permission error — the chart or dashboard simply does not match the access-scoped lookup. Reads, by contrast, require only viewer access; see [Query and subscribe to charts](/api/charts/query-charts).

## Related

- [Charts](/api/charts) — section overview and the full chart type map.
- [Build chart segments and values](/api/charts/chart-segments-and-values) — add the formulas and aggregations that power a manual chart.
- [Preview, recalculate, and export](/api/charts/preview-recalculate-export) — preview before saving, force a recompute, or export to CSV.
- [Query and subscribe to charts](/api/charts/query-charts) — read charts and listen for real-time updates.
- [Dashboards](/api/dashboards) — charts always belong to a dashboard; access is inherited from it.
