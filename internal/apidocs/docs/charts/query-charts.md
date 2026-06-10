---
title: Query and subscribe to charts
description: Read a single chart by id, list every chart on a dashboard, and subscribe to real-time chart create, update, and delete events.
icon: ListFilter
order: 4
---

Use the `chart` query to read a single chart by id, the `charts` query to list every chart on a dashboard, and the `subscribeToChart` subscription to receive real-time create, update, and delete events. Charts are `Chart` objects in the API; they always live inside a parent dashboard (a `Dashboard` object), so reads are scoped by the chart's dashboard, not by an organization or workspace.

A chart is one of three card types — `STAT`, `PIE`, or `BAR`. Manual charts carry `chartSegments` (each a `Formula` over named `ChartSegmentValue` aggregates); auto-generated charts instead carry `metadata` describing a bar or pie spec over records. Both shapes come back from these queries; see [Build chart segments and values](/api/charts/chart-segments-and-values) and [Create and manage charts](/api/charts/manage-charts) for how each is constructed.

## Query a single chart

Use the `chart` query to fetch one chart by its id.

### Request

```graphql
query GetChart {
  chart(id: "chart_123") {
    id
    title
    type
    position
    isCalculating
    needCalculation
    chartSegments {
      id
      uid
      title
      color
      formulaResult
    }
    createdAt
    updatedAt
  }
}
```

### Parameters

| Argument | Type      | Required | Description              |
| -------- | --------- | -------- | ------------------------ |
| `id`     | `String!` | Yes      | ID of the chart to read. |

### Response

```json
{
  "data": {
    "chart": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Revenue by stage",
      "type": "STAT",
      "position": 1024,
      "isCalculating": false,
      "needCalculation": false,
      "chartSegments": [
        {
          "id": "clm4n8qwx000108l0a1b2c3d4",
          "uid": "seg_a1b2c3",
          "title": "Closed won",
          "color": "#2563eb",
          "formulaResult": 48250
        }
      ],
      "createdAt": "2026-04-12T09:30:00.000Z",
      "updatedAt": "2026-05-28T14:05:11.000Z"
    }
  }
}
```

## List charts on a dashboard

Use the `charts` query to list every chart on a dashboard. The filter requires a `dashboardId` — there is no organization- or workspace-level chart listing. Results are paginated with offset-based `skip`/`take` and ordered by `sort` (default `[position_ASC]`, the card order on the dashboard).

### Request

The smallest valid call passes a `filter` with the required `dashboardId`:

```graphql
query ListCharts {
  charts(filter: { dashboardId: "dashboard_123" }) {
    items {
      id
      title
      type
      position
      isCalculating
      needCalculation
    }
    pageInfo {
      totalItems
      totalPages
      page
      perPage
      hasNextPage
      hasPreviousPage
    }
  }
}
```

### Parameters

| Argument | Type                | Required | Description                                                            |
| -------- | ------------------- | -------- | ---------------------------------------------------------------------- |
| `filter` | `ChartFilterInput!` | Yes      | Scopes the listing to one dashboard, optionally under a record filter. |
| `sort`   | `[ChartSort!]`      | No       | Ordering applied to the results. Defaults to `[position_ASC]`.         |
| `skip`   | `Int`               | No       | Number of charts to skip before returning results. Defaults to `0`.    |
| `take`   | `Int`               | No       | Number of charts to return per page. Defaults to `20`.                 |

#### ChartFilterInput

| Field         | Type              | Required | Description                                                                                                         |
| ------------- | ----------------- | -------- | ------------------------------------------------------------------------------------------------------------------- |
| `dashboardId` | `String!`         | Yes      | ID of the dashboard whose charts to list.                                                                           |
| `todoFilter`  | `TodoFilterInput` | No       | Recalculates each chart's results against this record filter instead of returning the saved values. See note below. |

#### ChartSort

`sort` takes an array of these enum values; later values break ties from earlier ones.

| Value            | Description                                    |
| ---------------- | ---------------------------------------------- |
| `position_ASC`   | Dashboard card order, first to last (default). |
| `position_DESC`  | Dashboard card order, last to first.           |
| `title_ASC`      | Title, A to Z.                                 |
| `title_DESC`     | Title, Z to A.                                 |
| `createdBy_ASC`  | Creator's first name, A to Z.                  |
| `createdBy_DESC` | Creator's first name, Z to A.                  |
| `updatedAt_ASC`  | Least recently updated first.                  |
| `updatedAt_DESC` | Most recently updated first.                   |

### Response

```json
{
  "data": {
    "charts": {
      "items": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "title": "Revenue by stage",
          "type": "STAT",
          "position": 1024,
          "isCalculating": false,
          "needCalculation": false
        },
        {
          "id": "clm4n8qwx000208l0e5f6g7h8",
          "title": "Records by list",
          "type": "BAR",
          "position": 2048,
          "isCalculating": false,
          "needCalculation": false
        }
      ],
      "pageInfo": {
        "totalItems": 2,
        "totalPages": 1,
        "page": 1,
        "perPage": 20,
        "hasNextPage": false,
        "hasPreviousPage": false
      }
    }
  }
}
```

<Callout variant="info" title="todoFilter recomputes in the background — poll until the flags clear">

When you pass `filter.todoFilter`, `charts` does not return the saved (unfiltered) results. Each chart comes back either with its filtered values — when a recent identical filter left them warm in the server cache (results stay cached for about 5 minutes) — or flagged `isCalculatingWithFilter: true` while a background recalculation runs. The recalculation is triggered at most once per dashboard-and-filter combination, so it is safe to poll: re-issue the same `charts` query every second or so until no item is flagged, at which point every chart carries its filtered values. Recomputed values are also published over [`subscribeToChart`](#subscribe-to-chart-changes), but polling the query is the recommended way to consume them.

Within the cache window, repeated identical filters return the cached values rather than recomputing. The cache is invalidated whenever a chart's base results recompute (a config edit, [`recalculateCharts`](/api/charts/preview-recalculate-export), or record changes that trigger a recalculation) — so stale filtered values last at most until the next recalculation or cache expiry, whichever comes first.

</Callout>

## Filtered listing

Pass a `todoFilter` to recompute the dashboard's charts against an ad-hoc record filter — for example, only records updated this quarter:

```graphql
query ListChartsFiltered {
  charts(
    filter: {
      dashboardId: "dashboard_123"
      todoFilter: { projectIds: ["project_123"], showCompleted: false }
    }
    sort: [position_ASC]
    take: 50
  ) {
    items {
      id
      title
      type
      isCalculatingWithFilter
    }
    pageInfo {
      totalItems
      hasNextPage
    }
  }
}
```

`todoFilter` is a `TodoFilterInput` — the same record-filter shape used across the API. See [List records](/api/records) for its full field set.

## Subscribe to chart changes

Use the `subscribeToChart` subscription to receive real-time events when charts on a dashboard are created, updated, or deleted — including the recomputed values that follow a recalculation. Subscriptions run over the WebSocket transport at `wss://api.blue.app/graphql`.

### Request

```graphql
subscription OnChartChange {
  subscribeToChart(filter: { dashboardId: "dashboard_123" }) {
    mutation
    node {
      id
      title
      type
      position
      isCalculating
      needCalculation
      chartSegments {
        id
        uid
        formulaResult
      }
    }
    previousValues {
      id
      title
    }
  }
}
```

### Parameters

| Argument | Type                           | Required | Description                                               |
| -------- | ------------------------------ | -------- | --------------------------------------------------------- |
| `filter` | `SubscribeToChartFilterInput!` | Yes      | The dashboard to watch, optionally under a record filter. |

#### SubscribeToChartFilterInput

| Field         | Type              | Required | Description                                                                                          |
| ------------- | ----------------- | -------- | ---------------------------------------------------------------------------------------------------- |
| `dashboardId` | `String!`         | Yes      | ID of the dashboard to receive `CHART_CREATED` / `CHART_UPDATED` / `CHART_DELETED` events for.       |
| `todoFilter`  | `TodoFilterInput` | No       | Restricts the stream to events recomputed under this exact record filter. See the gating note below. |

### Event payload

Each event delivers a `ChartSubscriptionPayload`.

| Field            | Type            | Description                                                                                           |
| ---------------- | --------------- | ----------------------------------------------------------------------------------------------------- |
| `mutation`       | `MutationType!` | The kind of change: `CREATED`, `UPDATED`, or `DELETED`.                                               |
| `node`           | `Chart`         | The chart's new state. Populated for `CREATED` and `UPDATED`; `null` for `DELETED`.                   |
| `previousValues` | `Chart`         | The chart's prior state. Populated for `DELETED` (and updates); use it to identify the removed chart. |

### Example event

```json
{
  "data": {
    "subscribeToChart": {
      "mutation": "UPDATED",
      "node": {
        "id": "clm4n8qwx000008l0g4oxdqn7",
        "title": "Revenue by stage",
        "type": "STAT",
        "position": 1024,
        "isCalculating": false,
        "needCalculation": false,
        "chartSegments": [
          {
            "id": "clm4n8qwx000108l0a1b2c3d4",
            "uid": "seg_a1b2c3",
            "formulaResult": 48250
          }
        ]
      },
      "previousValues": null
    }
  }
}
```

<Callout variant="warning" title="todoFilter must match exactly to receive filtered events">

The subscription's `todoFilter` is matched against the `todoFilter` a chart was recomputed under, by exact JSON equality. An event is delivered only when both sides carry the same `todoFilter`, or when neither does. If one side has a `todoFilter` and the other doesn't, the event is dropped. To follow the recalculations triggered by a filtered `charts` query, open the subscription with the identical `todoFilter` you passed to `charts`.

</Callout>

## Return types

### Chart

| Field                     | Type               | Description                                                                                                               |
| ------------------------- | ------------------ | ------------------------------------------------------------------------------------------------------------------------- |
| `id`                      | `ID!`              | Unique identifier for the chart.                                                                                          |
| `title`                   | `String!`          | Display name of the chart card.                                                                                           |
| `position`                | `Float!`           | Sort position within the dashboard. Lower values appear first.                                                            |
| `type`                    | `ChartType!`       | The card type: `STAT`, `PIE`, or `BAR`.                                                                                   |
| `chartSegments`           | `[ChartSegment!]!` | The manual segments of the chart. Empty for auto-generated (`metadata`-backed) charts.                                    |
| `display`                 | `FormulaDisplay`   | How the chart's rolled-up value is formatted (number, currency, or percentage) and rolled up.                             |
| `metadata`                | `ChartMetadata`    | The auto-generated bar or pie spec. `null` for manual (segment-backed) charts.                                            |
| `isCalculating`           | `Boolean`          | `true` while a recompute of the saved results is in progress.                                                             |
| `isCalculatingWithFilter` | `Boolean`          | `true` on items returned by `charts` when a `todoFilter` was supplied; the recomputed values arrive via the subscription. |
| `needCalculation`         | `Boolean`          | `true` when the saved results are stale and a recalculation is pending.                                                   |
| `createdAt`               | `DateTime!`        | When the chart was created.                                                                                               |
| `updatedAt`               | `DateTime!`        | When the chart was last modified.                                                                                         |

### ChartSegment

| Field                | Type                    | Description                                                                                   |
| -------------------- | ----------------------- | --------------------------------------------------------------------------------------------- |
| `id`                 | `ID!`                   | Unique identifier for the segment.                                                            |
| `uid`                | `String!`               | Stable public identifier for the segment.                                                     |
| `title`              | `String`                | The segment's label.                                                                          |
| `color`              | `String`                | The segment's display color (hex string).                                                     |
| `chartSegmentValues` | `[ChartSegmentValue!]!` | The named aggregates the segment's formula draws on.                                          |
| `formula`            | `Formula!`              | The formula combining the segment's values. `logic.text` / `logic.html` reference value UIDs. |
| `formulaResult`      | `Float`                 | The computed result of the formula, or `null` until calculated.                               |
| `createdAt`          | `DateTime!`             | When the segment was created.                                                                 |
| `updatedAt`          | `DateTime!`             | When the segment was last modified.                                                           |

### ChartSegmentValue

| Field           | Type                         | Description                                                                               |
| --------------- | ---------------------------- | ----------------------------------------------------------------------------------------- |
| `id`            | `ID!`                        | Unique identifier for the value.                                                          |
| `uid`           | `String!`                    | The identifier a segment's `formula.logic` references this value by.                      |
| `title`         | `String!`                    | The value's label.                                                                        |
| `disabled`      | `Boolean!`                   | Whether the value is excluded from its segment's formula.                                 |
| `projectId`     | `String!`                    | ID of the workspace whose records this value aggregates over.                             |
| `customFieldId` | `String`                     | ID of the custom field being aggregated, when the function targets one.                   |
| `function`      | `ChartSegmentValueFunctions` | The aggregate function: `COUNT`, `COUNTA`, `SUM`, `AVERAGE`, `AVERAGEA`, `MIN`, or `MAX`. |
| `filter`        | `TodoFilter`                 | The record filter scoping which records are aggregated.                                   |
| `createdAt`     | `DateTime!`                  | When the value was created.                                                               |
| `updatedAt`     | `DateTime!`                  | When the value was last modified.                                                         |

### ChartMetadata

`metadata` is a union — `ChartMetadataBarChart` for `BAR` charts and `ChartMetadataPieChart` for `PIE` charts — so select fields through an inline fragment:

```graphql
query GetChartMetadata {
  chart(id: "chart_123") {
    id
    type
    metadata {
      ... on ChartMetadataBarChart {
        barChart {
          xAxis {
            title
            type
            interval
          }
          yAxis {
            title
            function
          }
        }
      }
      ... on ChartMetadataPieChart {
        pieChart {
          groupBy {
            title
            type
          }
          value {
            title
            function
          }
        }
      }
    }
  }
}
```

For `BAR`, `barChart.xAxis` carries the grouping (`type` is a `BarChartXAxisType`, with an optional `interval` for date axes) and `barChart.yAxis` carries the aggregate (`function` is a `ChartSegmentValueFunctions`). For `PIE`, `pieChart.groupBy` and `pieChart.value` mirror those roles. The axis input shapes are documented in [Create and manage charts](/api/charts/manage-charts).

### PageInfo

| Field             | Type       | Description                                                        |
| ----------------- | ---------- | ------------------------------------------------------------------ |
| `totalItems`      | `Int`      | Total number of charts on the dashboard, across all pages.         |
| `totalPages`      | `Int`      | Total number of pages at the current `take`.                       |
| `page`            | `Int`      | The current page number (1-based), derived from `skip` and `take`. |
| `perPage`         | `Int`      | Number of items per page (mirrors `take`).                         |
| `hasNextPage`     | `Boolean!` | Whether another page follows the current one.                      |
| `hasPreviousPage` | `Boolean!` | Whether a page precedes the current one.                           |

`PageInfo` also exposes `startCursor` and `endCursor`, but both are deprecated — charts are not cursor-paginated. Walk pages with `skip`/`take`.

## Errors

| Code                  | When                                                                                                  |
| --------------------- | ----------------------------------------------------------------------------------------------------- |
| `UNAUTHENTICATED`     | The request has no valid credentials.                                                                 |
| `CHART_NOT_FOUND`     | `chart(id:)` references a chart that doesn't exist, or whose dashboard you can't access.              |
| `DASHBOARD_NOT_FOUND` | `charts` (or `subscribeToChart`) references a dashboard that doesn't exist, or that you can't access. |

## Permissions

You must be authenticated. Charts inherit their access from the parent dashboard: you can read a dashboard's charts if you created the dashboard or have been added to it as a `DashboardUser` (any role — `VIEWER` is sufficient for reads). The `subscribeToChart` stream applies the same check and silently drops events for dashboards you can't access. Writing charts requires the dashboard owner or the `EDITOR` role — see [Create and manage charts](/api/charts/manage-charts).

## Related

- [Charts overview](/api/charts) — How charts, segments, values, and metadata fit together.
- [Create and manage charts](/api/charts/manage-charts) — Create, edit, copy, and delete charts.
- [Build chart segments and values](/api/charts/chart-segments-and-values) — Construct the manual building blocks.
- [Preview, recalculate, and export charts](/api/charts/preview-recalculate-export) — Compute results and export CSVs.
- [List dashboards](/api/dashboards) — Find the dashboards whose charts you can query.
