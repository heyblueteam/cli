---
title: Build chart segments and values
description: Create, edit, and delete the segments and values that drive a manual chart — formula logic plus project-scoped aggregations.
icon: Layers
order: 2
---

Manual charts are built from **segments** and **values**. A `ChartSegment` holds a `Formula` — display text/HTML that references value identifiers — plus a title and color. Each `ChartSegmentValue` is a named, workspace-scoped aggregation: a `ChartSegmentValueFunctions` rollup (`COUNT`, `SUM`, `AVERAGE`, …) over one workspace's records, optionally narrowed to a single custom field and filtered with a `TodoFilterInput`. The formula combines those values into the number the chart renders.

These mutations apply only to **manual** charts — charts you created with inline `chartSegments`. A chart built from `metadata.query` is **auto-generated**: its values come from a grouping, not from segments, so the segment and value mutations do not apply to it. See [Create and manage charts](/api/charts/manage-charts) for the manual-vs-auto distinction.

Every mutation on this page republishes the parent chart to dashboard subscribers and enqueues a recalculation. Results arrive asynchronously over [`subscribeToChart`](/api/charts/query-charts) — not in the mutation response. Workspaces are `Project` objects and records are `Todo` objects in the API.

## The UID linkage

The non-obvious core mechanic: each `ChartSegmentValue` has a stable `uid`, and the segment's `formula.logic.text` and `formula.logic.html` reference those `uid` tokens. The formula is the expression; the values are its variables. When you copy a chart with [`copyChart`](/api/charts/manage-charts), Blue regenerates each value `uid` and string-replaces it inside the copied formula so the link survives.

Pass `uid` explicitly on a value when you want the formula to reference a known token; omit it and Blue auto-generates one. The same applies to a segment's own `uid`.

## Segments

### Create a segment

Use the `createChartSegment` mutation to add a segment to a manual chart. Provide the chart, a `title`, a `color`, the `formula`, and (optionally) the segment's `chartSegmentValues` inline.

#### Request

```graphql
mutation CreateChartSegment {
  createChartSegment(
    input: {
      chartId: "chart_123"
      title: "Win rate"
      color: "#2563eb"
      formula: {
        logic: {
          text: "{{value_won}} / {{value_total}}"
          html: "<p>{{value_won}} / {{value_total}}</p>"
        }
        display: { type: PERCENTAGE, precision: 1 }
      }
      chartSegmentValues: [
        { uid: "value_won", title: "Won deals", projectId: "project_123", function: COUNT }
        { uid: "value_total", title: "All deals", projectId: "project_123", function: COUNT }
      ]
    }
  ) {
    id
    uid
    title
    color
    chartSegmentValues {
      id
      uid
      title
    }
  }
}
```

#### Parameters

##### CreateChartSegmentInput

| Parameter            | Type                        | Required | Description                                                                                                |
| -------------------- | --------------------------- | -------- | ---------------------------------------------------------------------------------------------------------- |
| `chartId`            | `String!`                   | Yes      | ID of the manual chart to add the segment to. You must own the dashboard or be an `EDITOR`.                |
| `title`              | `String!`                   | Yes      | Display name for the segment.                                                                              |
| `color`              | `String!`                   | Yes      | The segment's color (a CSS color string, e.g. `#2563eb`).                                                  |
| `formula`            | `FormulaInput!`             | Yes      | The expression and display config. Its `logic.text`/`logic.html` reference the segment's value `uid`s.     |
| `uid`                | `String`                    | No       | Stable identifier for the segment. Auto-generated when omitted.                                            |
| `chartSegmentValues` | `[ChartSegmentValueInput!]` | No       | Values to create alongside the segment. Each value can also be added later with `createChartSegmentValue`. |

##### FormulaInput

| Parameter | Type                   | Required | Description                                              |
| --------- | ---------------------- | -------- | -------------------------------------------------------- |
| `logic`   | `FormulaLogicInput!`   | Yes      | The expression that combines value `uid`s into a number. |
| `display` | `FormulaDisplayInput!` | Yes      | How the computed result is formatted.                    |

##### FormulaLogicInput

| Parameter | Type      | Required | Description                                                                                  |
| --------- | --------- | -------- | -------------------------------------------------------------------------------------------- |
| `text`    | `String!` | Yes      | Plain-text form of the formula, containing the segment's value `uid` tokens.                 |
| `html`    | `String!` | Yes      | Rich-text/HTML form of the same formula. Keep it in sync with `text`; both carry the `uid`s. |

##### FormulaDisplayInput

| Parameter   | Type                          | Required | Description                                                         |
| ----------- | ----------------------------- | -------- | ------------------------------------------------------------------- |
| `type`      | `FormulaDisplayType!`         | Yes      | Result format: `NUMBER`, `CURRENCY`, or `PERCENTAGE`.               |
| `precision` | `Float`                       | No       | Number of decimal places to render.                                 |
| `currency`  | `FormulaDisplayCurrencyInput` | No       | Currency `code` + `name`, used when `type` is `CURRENCY`.           |
| `function`  | `ChartFunction`               | No       | A display-level rollup over the formula result. See the note below. |

<Callout variant="info" title="ChartFunction vs ChartSegmentValueFunctions">

Two enums share the same members (`AVERAGE`, `AVERAGEA`, `COUNT`, `COUNTA`, `MAX`, `MIN`, `SUM`) but live in different places. `ChartFunction` is the segment-level display rollup on `FormulaDisplayInput.function`. `ChartSegmentValueFunctions` is the per-value aggregation on each `ChartSegmentValue`. They are distinct types — don't substitute one for the other.

</Callout>

##### ChartSegmentValueInput

| Parameter       | Type                         | Required | Description                                                                                         |
| --------------- | ---------------------------- | -------- | --------------------------------------------------------------------------------------------------- |
| `title`         | `String!`                    | Yes      | Display name for the value.                                                                         |
| `projectId`     | `String!`                    | Yes      | ID of the workspace whose records this value aggregates. You must be a member of it.                |
| `uid`           | `String`                     | No       | Stable identifier the formula references. Auto-generated when omitted.                              |
| `customFieldId` | `String`                     | No       | Restrict the aggregation to one custom field's values. Must belong to `projectId`.                  |
| `function`      | `ChartSegmentValueFunctions` | No       | Aggregate to apply: `COUNT`, `COUNTA`, `SUM`, `AVERAGE`, `AVERAGEA`, `MIN`, or `MAX`.               |
| `filter`        | `TodoFilterInput`            | No       | Narrows which records are aggregated. See [List records](/api/records/list-records) for its fields. |

#### Response

```json
{
  "data": {
    "createChartSegment": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "uid": "seg_a1b2c3",
      "title": "Win rate",
      "color": "#2563eb",
      "chartSegmentValues": [
        { "id": "clm4n8qwx000108l0h7yzerm2", "uid": "value_won", "title": "Won deals" },
        { "id": "clm4n8qwx000208l0k9pqwst3", "uid": "value_total", "title": "All deals" }
      ]
    }
  }
}
```

##### ChartSegment

| Field                | Type                    | Description                                                   |
| -------------------- | ----------------------- | ------------------------------------------------------------- |
| `id`                 | `ID!`                   | Unique identifier of the segment.                             |
| `uid`                | `String!`               | Stable short identifier for the segment.                      |
| `title`              | `String`                | The segment's display name.                                   |
| `color`              | `String`                | The segment's color.                                          |
| `chartSegmentValues` | `[ChartSegmentValue!]!` | The values this segment's formula references.                 |
| `formula`            | `Formula!`              | The expression and display config (`logic` + `display`).      |
| `formulaResult`      | `Float`                 | The computed result, or `null` until recalculation completes. |
| `createdAt`          | `DateTime!`             | When the segment was created.                                 |
| `updatedAt`          | `DateTime!`             | When the segment was last modified.                           |

### Edit a segment

Use the `editChartSegment` mutation to change a segment's title, color, or formula. Only the fields you pass are updated; the rest are left unchanged. Passing a new `formula` triggers a recalculation; changing only the title or color republishes the chart without recomputing.

#### Request

```graphql
mutation EditChartSegment {
  editChartSegment(
    input: {
      id: "segment_123"
      title: "Win rate (rounded)"
      formula: {
        logic: {
          text: "{{value_won}} / {{value_total}}"
          html: "<p>{{value_won}} / {{value_total}}</p>"
        }
        display: { type: PERCENTAGE, precision: 0 }
      }
    }
  ) {
    id
    title
    formula {
      display {
        type
        precision
      }
    }
  }
}
```

#### Parameters

##### EditChartSegmentInput

| Parameter | Type           | Required | Description                                                              |
| --------- | -------------- | -------- | ------------------------------------------------------------------------ |
| `id`      | `String!`      | Yes      | ID of the segment to edit. You must own the dashboard or be an `EDITOR`. |
| `title`   | `String`       | No       | New display name.                                                        |
| `color`   | `String`       | No       | New color.                                                               |
| `formula` | `FormulaInput` | No       | Replacement formula. When supplied, the chart is recalculated.           |

The response is the updated `ChartSegment` (same shape as `createChartSegment` above).

### Delete a segment

Use the `deleteChartSegment` mutation to remove a segment and its values from the chart. The chart is republished to dashboard subscribers.

#### Request

```graphql
mutation DeleteChartSegment {
  deleteChartSegment(id: "segment_123") {
    success
  }
}
```

`deleteChartSegment` returns a `MutationResult`, which exposes only `success` and `operationId`. It does not take a sub-selection on a segment object.

#### Response

```json
{
  "data": {
    "deleteChartSegment": {
      "success": true
    }
  }
}
```

##### MutationResult

| Field         | Type       | Description                                                  |
| ------------- | ---------- | ------------------------------------------------------------ |
| `success`     | `Boolean!` | `true` when the segment was deleted, `false` if it failed.   |
| `operationId` | `String`   | Identifier for correlating the operation with subscriptions. |

## Values

You can manage values inline on a segment (above) or independently with the value mutations below.

### Create a value

Use the `createChartSegmentValue` mutation to add a value to an existing segment. The new value's `uid` is what you reference from the segment's formula.

#### Request

```graphql
mutation CreateChartSegmentValue {
  createChartSegmentValue(
    input: {
      chartSegmentId: "segment_123"
      uid: "value_won"
      title: "Won deals"
      projectId: "project_123"
      customFieldId: "field_123"
      function: SUM
      filter: { showCompleted: false, todoListIds: ["list_123"] }
    }
  ) {
    id
    uid
    title
    projectId
    customFieldId
    function
  }
}
```

#### Parameters

##### CreateChartSegmentValueInput

| Parameter        | Type                         | Required | Description                                                                                         |
| ---------------- | ---------------------------- | -------- | --------------------------------------------------------------------------------------------------- |
| `chartSegmentId` | `String!`                    | Yes      | ID of the segment to add the value to. You must own the dashboard or be an `EDITOR`.                |
| `title`          | `String!`                    | Yes      | Display name for the value.                                                                         |
| `projectId`      | `String!`                    | Yes      | ID of the workspace whose records this value aggregates. You must be a member of it.                |
| `uid`            | `String`                     | No       | Stable identifier the formula references. Auto-generated when omitted.                              |
| `customFieldId`  | `String`                     | No       | Restrict the aggregation to one custom field's values. Must belong to `projectId`.                  |
| `function`       | `ChartSegmentValueFunctions` | No       | Aggregate to apply: `COUNT`, `COUNTA`, `SUM`, `AVERAGE`, `AVERAGEA`, `MIN`, or `MAX`.               |
| `filter`         | `TodoFilterInput`            | No       | Narrows which records are aggregated. See [List records](/api/records/list-records) for its fields. |

##### ChartSegmentValueFunctions

| Value      | Aggregation                                                         |
| ---------- | ------------------------------------------------------------------- |
| `COUNT`    | Count of records with a numeric value in the field.                 |
| `COUNTA`   | Count of records with any non-empty value (or any matching record). |
| `SUM`      | Sum of the field's numeric values.                                  |
| `AVERAGE`  | Mean of the field's numeric values.                                 |
| `AVERAGEA` | Mean treating non-numeric/empty values as zero.                     |
| `MIN`      | Minimum of the field's numeric values.                              |
| `MAX`      | Maximum of the field's numeric values.                              |

#### Response

```json
{
  "data": {
    "createChartSegmentValue": {
      "id": "clm4n8qwx000308l0v2bcdef4",
      "uid": "value_won",
      "title": "Won deals",
      "projectId": "clm4n8qwx000408l0p5ghij5k",
      "customFieldId": "clm4n8qwx000508l0q6klmn6l",
      "function": "SUM"
    }
  }
}
```

##### ChartSegmentValue

| Field           | Type                         | Description                                                |
| --------------- | ---------------------------- | ---------------------------------------------------------- |
| `id`            | `ID!`                        | Unique identifier of the value.                            |
| `uid`           | `String!`                    | Stable identifier the segment's formula references.        |
| `title`         | `String!`                    | The value's display name.                                  |
| `disabled`      | `Boolean!`                   | Whether the value is excluded from the formula.            |
| `projectId`     | `String!`                    | The workspace whose records are aggregated.                |
| `customFieldId` | `String`                     | The custom field the aggregation is restricted to, if any. |
| `function`      | `ChartSegmentValueFunctions` | The aggregate applied.                                     |
| `filter`        | `TodoFilter`                 | The record filter applied, if any.                         |
| `createdAt`     | `DateTime!`                  | When the value was created.                                |
| `updatedAt`     | `DateTime!`                  | When the value was last modified.                          |

### Edit a value

Use the `editChartSegmentValue` mutation to change a value's title, workspace, custom field, function, or filter. Only the fields you pass are updated. Editing a value always recalculates the chart.

If you pass `customFieldId`, it is validated against the new `projectId` when you also pass one — otherwise against the value's existing workspace.

#### Request

```graphql
mutation EditChartSegmentValue {
  editChartSegmentValue(
    input: { id: "value_123", function: AVERAGE, filter: { dueStart: "2026-01-01T00:00:00.000Z" } }
  ) {
    id
    title
    function
  }
}
```

#### Parameters

##### EditChartSegmentValueInput

| Parameter       | Type                         | Required | Description                                                              |
| --------------- | ---------------------------- | -------- | ------------------------------------------------------------------------ |
| `id`            | `String!`                    | Yes      | ID of the value to edit. You must own the dashboard or be an `EDITOR`.   |
| `title`         | `String`                     | No       | New display name.                                                        |
| `projectId`     | `String`                     | No       | New workspace to aggregate. You must be a member of it.                  |
| `customFieldId` | `String`                     | No       | New custom field to restrict to. Must belong to the effective workspace. |
| `function`      | `ChartSegmentValueFunctions` | No       | New aggregate.                                                           |
| `filter`        | `TodoFilterInput`            | No       | New record filter.                                                       |

The response is the updated `ChartSegmentValue` (same shape as `createChartSegmentValue` above).

### Delete a value

Use the `deleteChartSegmentValue` mutation to remove a value. The chart is republished and recalculated.

<Callout variant="warning" title="Update the formula too">

Deleting a value does not rewrite the segment's formula. If the formula still references the deleted value's `uid`, edit the segment with `editChartSegment` to remove that token, or the formula will reference a value that no longer exists.

</Callout>

#### Request

```graphql
mutation DeleteChartSegmentValue {
  deleteChartSegmentValue(id: "value_123") {
    success
  }
}
```

`deleteChartSegmentValue` returns a `MutationResult` — select `success` and/or `operationId`, with no sub-selection on a value object.

#### Response

```json
{
  "data": {
    "deleteChartSegmentValue": {
      "success": true
    }
  }
}
```

## Errors

| Code                            | When                                                                                                                     |
| ------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| `CHART_NOT_FOUND`               | `createChartSegment`: the `chartId` doesn't resolve, or you don't own its dashboard and aren't an `EDITOR`.              |
| `CHART_SEGMENT_NOT_FOUND`       | `editChartSegment`, `deleteChartSegment`, `createChartSegmentValue`: the segment doesn't resolve or you can't modify it. |
| `CHART_SEGMENT_VALUE_NOT_FOUND` | `editChartSegmentValue`, `deleteChartSegmentValue`: the value doesn't resolve or you can't modify it.                    |
| `PROJECT_NOT_FOUND`             | A value's `projectId` doesn't resolve or you aren't a member of that workspace.                                          |
| `CUSTOM_FIELD_NOT_FOUND`        | A value's `customFieldId` doesn't belong to the referenced workspace.                                                    |

## Permissions

Segments and values are governed by the parent chart's **dashboard**. Every mutation here requires that you either created the dashboard or are a `DashboardUser` on it with the `EDITOR` role. A `VIEWER` (or anyone without dashboard access) is rejected with the corresponding not-found error — the API does not distinguish "doesn't exist" from "you can't see it" for these resources.

Beyond dashboard access, creating or editing a value also checks the data it points at: you must be a member of every referenced `projectId`, and any `customFieldId` must belong to that workspace. This prevents a value from aggregating records you can't otherwise access.

## Related

- [Charts](/api/charts) — overview of charts and how they live on a dashboard.
- [Create and manage charts](/api/charts/manage-charts) — create a manual or auto-generated chart, then `copyChart` (which remaps value `uid`s in formulas).
- [Preview, recalculate, and export charts](/api/charts/preview-recalculate-export) — preview a chart before saving and force recalculation.
- [Query and subscribe to charts](/api/charts/query-charts) — read chart results and subscribe for recalculation updates.
- [List records](/api/records/list-records) — the `TodoFilterInput` fields used by a value's `filter`.
