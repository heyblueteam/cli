---
title: Create and manage charts
description: Create STAT, PIE, and BAR charts on a dashboard, then rename, reposition, restyle, duplicate, and delete them.
icon: ChartColumn
order: 1
---

A chart is a single card on a [dashboard](/api/dashboards). This page covers the full lifecycle: create a chart with `createChart`, change its title, position, display, or metadata with `editChart`, duplicate it into the same or another dashboard with `copyChart`, and remove it with `deleteChart`. Charts are `Chart` objects in the API.

Every chart is either a **manual** chart (driven by `chartSegments` — formulas over project-scoped values) or an **auto-generated** chart (driven by `metadata.query`). `createChart` accepts either shape; the [segment and value mutations](/api/charts/chart-segments-and-values) apply only to manual charts. See the [section overview](/api/charts) for the full type map.

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
| `metadata`      | `ChartMetadataInput`   | No       | A `query` spec for an **auto-generated** chart. Mutually exclusive with `chartSegments`.                                               |

#### ChartType

| Value  | Description                                                          |
| ------ | -------------------------------------------------------------------- |
| `STAT` | A single headline number (one segment value, or a formula rollup).   |
| `PIE`  | A pie chart — either manual segments or a grouped `query`.           |
| `BAR`  | A bar chart — either manual segments or a grouped `query`.           |

#### FormulaDisplayInput

| Parameter   | Type                          | Required | Description                                                                                                 |
| ----------- | ----------------------------- | -------- | ----------------------------------------------------------------------------------------------------------- |
| `type`      | `FormulaDisplayType!`         | Yes      | `NUMBER`, `CURRENCY`, or `PERCENTAGE`.                                                                      |
| `currency`  | `FormulaDisplayCurrencyInput` | No       | `{ code, name }`, e.g. `{ code: "USD", name: "US Dollar" }`. Use with `CURRENCY`.                           |
| `precision` | `Float`                       | No       | Number of decimal places to show.                                                                           |
| `function`  | `ChartFunction`               | No       | Display rollup applied across the chart. See [ChartFunction](#chartfunction-vs-chartsegmentvaluefunctions). |

#### ChartMetadataInput

One shape for every chart. Supply `query` for an auto-generated chart; omit it for a
manual one, whose numbers come from its `chartSegments` instead.

| Parameter      | Type                       | Required | Description                                                                        |
| -------------- | -------------------------- | -------- | ---------------------------------------------------------------------------------- |
| `query`        | `ChartQueryInput`          | No       | What to group by and what to measure. Omit for a manual chart.                     |
| `presentation` | `ChartPresentationInput`   | No       | How the result is dressed — stacking, target, thresholds, context band.            |

`ChartQueryInput`:

| Parameter    | Type                     | Required | Description                                                                                        |
| ------------ | ------------------------ | -------- | -------------------------------------------------------------------------------------------------- |
| `dimensions` | `[ChartDimensionInput!]!` | Yes      | What records are grouped into. One entry today.                                                    |
| `metrics`    | `[ChartMetricInput!]!`    | Yes      | What is measured per group. More than one plots several series over the same grouping.             |
| `breakout`   | `ChartBreakoutInput`      | No       | A second grouping that splits every bucket into slices. See [stacked charts](#stacked-bar-charts). |
| `filters`    | `TodoFilterInput`         | No       | Narrows the whole chart. Same filter used by [list records](/api/records/list-records).            |

`ChartDimensionInput`:

| Field                           | Type                    | Required | Description                                                                                                                                                                      |
| ------------------------------- | ----------------------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `type`                          | `BarChartXAxisType!`    | Yes      | What to bucket records by: `PROJECT`, `ASSIGNEE`, `TAG`, `CUSTOM_FIELD`, `TODO`, `TODO_LIST`, `TODO_STATUS`, `TODO_DUE_DATE`, `TODO_CREATED_AT`, `TODO_UPDATED_AT`, `TODO_COMPLETED_AT`. |
| `title`                         | `String`                | No       | Optional axis label.                                                                                                                                                             |
| `interval`                      | `BarChartXAxisInterval` | No       | Bucket size for date dimensions: `DAY`, `WEEK`, `MONTH`, `QUARTER`, `YEAR`.                                                                                                      |
| `customFieldName`               | `String`                | No       | Custom-field name when `type` is `CUSTOM_FIELD`.                                                                                                                                 |
| `customFieldType`               | `CustomFieldType`       | No       | Type of that custom field.                                                                                                                                                       |
| `customFieldReferenceProjectId` | `String`                | No       | For reference custom fields, the referenced project.                                                                                                                             |

`ChartMetricInput`:

| Field                           | Type                         | Required | Description                                                                                                    |
| ------------------------------- | ---------------------------- | -------- | ---------------------------------------------------------------------------------------------------------------- |
| `key`                           | `String`                     | No       | Unique within the chart. Omit and Blue assigns one from the metric's position.                                  |
| `title`                         | `String`                     | No       | Optional series label.                                                                                         |
| `function`                      | `ChartSegmentValueFunctions` | No       | Aggregate applied to the measured records. Omit for a plain record count.                                       |
| `filter`                        | `TodoFilterInput`            | No       | Narrows this metric only, so two metrics can measure different slices of the same grouping.                    |
| `color`                         | `String`                     | No       | Explicit series colour. Omit to take the palette colour for the metric's position.                             |
| `axis`                          | `ChartMetricAxis`            | No       | `LEFT` (default) or `RIGHT`, so two metrics in different units can share a chart.                              |
| `customFieldName`               | `String`                     | No       | Custom-field name to aggregate.                                                                                |
| `customFieldType`               | `CustomFieldType`            | No       | Type of that custom field.                                                                                     |
| `customFieldReferenceProjectId` | `String`                     | No       | For reference custom fields, the referenced project.                                                           |

`ChartPresentationInput` carries `stackMode`, `target`, `direction`, `bands` and `context`. None of it affects which records the chart covers.

#### Stacked bar charts

A `BAR` chart can add a second grouping dimension so each bucket is split into stacked slices — records per assignee within each workspace, for example. Set `breakout` on `ChartQueryInput` to turn this on, and `presentation.stackMode` to choose how the slices combine.

| Parameter                 | Type                 | Required | Description                                                                            |
| ------------------------- | -------------------- | -------- | -------------------------------------------------------------------------------------- |
| `query.breakout`          | `ChartBreakoutInput` | No       | The dimension each bucket is split by. Omit for a single-series chart.                 |
| `presentation.stackMode`  | `ChartStackMode`     | No       | How the slices combine. Ignored when there is no `breakout`. Omit for a stacked chart. |

`ChartBreakoutInput` mirrors the dimension input, minus `interval` — bucketing a date axis by another date is not a composition:

| Field                           | Type                  | Required | Description                                                                                                |
| ------------------------------- | --------------------- | -------- | ---------------------------------------------------------------------------------------------------------- |
| `type`                          | `BarChartSeriesType!` | Yes      | What to split each bucket by: `PROJECT`, `ASSIGNEE`, `TAG`, `TODO_LIST`, `TODO_STATUS`, or `CUSTOM_FIELD`. |
| `title`                         | `String`              | No       | Optional label for the breakdown.                                                                          |
| `customFieldName`               | `String`              | No       | Custom-field name when `type` is `CUSTOM_FIELD`.                                                           |
| `customFieldType`               | `CustomFieldType`     | No       | Type of that custom field.                                                                                 |
| `customFieldReferenceProjectId` | `String`              | No       | For reference custom fields, the referenced project.                                                       |

`ChartStackMode` is either `STACKED` (slices are stacked and the bar's height is the bucket total) or `PERCENT` (each bucket is normalised to 100%, so only the mix is shown). A breakdown always stacks; side-by-side grouped bars are not offered.

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
| `metadata`                | `ChartMetadata`    | The `query`/`presentation` config. `query` is `null` on a manual chart.                |
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
        query: {
          dimensions: [{ title: "Workspace", type: PROJECT }]
          metrics: [{ title: "Records" }]
        }
      }
    }
  ) {
    id
    title
    type
    displayType
    metadata {
      query {
        dimensions {
          title
          type
        }
        metrics {
          key
          title
          function
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
| `metadata` | `ChartMetadataInput`  | No       | New `query`/`presentation` config. Changing the query re-triggers an asynchronous result calculation. |

<Callout variant="info" title="Editing metadata recalculates; editing title/position/display does not">

`editChart` only enqueues a recalculation when the input includes a `metadata.query`. Renaming, repositioning, or restyling a chart updates it and publishes the change to subscribers, but does not recompute its data. To force a recompute without changing the spec, use [`recalculateCharts`](/api/charts/preview-recalculate-export).

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
| `ChartSegmentValueFunctions` | Each segment value and each chart metric (`function`)              | `COUNT`, `COUNTA`, `SUM`, `AVERAGE`, `AVERAGEA`, `MIN`, `MAX` |

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
