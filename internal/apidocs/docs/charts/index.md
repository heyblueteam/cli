---
title: Charts
description: Create, build, preview, recalculate, export, and query the STAT, PIE, and BAR charts that live inside a dashboard.
icon: ChartColumn
order: 0
---

Charts are the visualizations that live inside a [dashboard](/api/dashboards). Each chart is a `STAT`, `PIE`, or `BAR` card built either from **auto-generated metadata** (an x-axis/y-axis or group-by/value spec computed over records) or from **manual segments** containing formula-driven values. This section covers creating and managing charts, building their segments and values, previewing results before saving, recalculating and CSV-exporting chart data, and querying or subscribing to charts on a dashboard.

Charts are always scoped to a parent dashboard. Permissions derive from dashboard access: viewer access for reads, and dashboard ownership or the `EDITOR` role for any write. There is no organization- or workspace-level chart query — you reach charts through their dashboard.

## Concepts and types

| Product noun         | Schema type                        | What it is                                                                                                         |
| -------------------- | ---------------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| Chart                | `Chart`                            | A single STAT/PIE/BAR card on a dashboard. `type` is `ChartType`.                                                  |
| Manual chart         | `Chart.chartSegments`              | A chart driven by `ChartSegment`s and formulas.                                                                    |
| Auto-generated chart | `Chart.metadata` (`ChartMetadata`) | A chart driven by a `query` — what to group by, what to measure.                                                   |
| Segment              | `ChartSegment`                     | A `Formula` plus a color and title; the building block of a manual chart.                                          |
| Segment value        | `ChartSegmentValue`                | A named, project-scoped aggregation referenced by a segment's formula via its `uid`.                               |
| Aggregate function   | `ChartSegmentValueFunctions`       | `COUNT`, `COUNTA`, `SUM`, `AVERAGE`, `AVERAGEA`, `MIN`, `MAX` — applied to a value.                                |
| Display rollup       | `FormulaDisplay` (`ChartFunction`) | The segment-level display function. Same member names as `ChartSegmentValueFunctions`, but a distinct type.        |
| Record               | `Todo`                             | Records are `Todo` objects; values aggregate over a project's records, optionally narrowed by a `TodoFilterInput`. |

<Callout variant="info" title="Manual vs auto-generated">

A chart carries **either** `chartSegments` (manual) **or** `metadata` (auto-generated), not both. The segment and value mutations apply only to manual charts; `query` metadata is supplied at create time. `ChartFunction` (on `FormulaDisplay`) and `ChartSegmentValueFunctions` (on each value and each chart metric) share member names but are different enums — keep them straight in examples.

</Callout>

## In this section

| Page                                                                       | What it covers                                                                                                                                                                                                                                               |
| -------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| [Create and manage charts](/api/charts/manage-charts)                      | The chart lifecycle: `createChart` (with optional inline `chartSegments` or `query` metadata), `editChart` to rename/reposition/restyle, `copyChart` to duplicate with UID-remapped formulas, and `deleteChart`.                                             |
| [Build chart segments and values](/api/charts/chart-segments-and-values)   | The building blocks of manual charts: create/edit/delete for `ChartSegment` and `ChartSegmentValue`, the formula-UID linkage, `ChartSegmentValueFunctions`, and project/custom-field validation.                                                             |
| [Preview, recalculate, and export](/api/charts/preview-recalculate-export) | Data operations: `previewChart` (a Query that computes a throwaway chart without persisting), `recalculateCharts` (force recomputation by id), and `exportChartCSV` (async, fire-and-forget CSV export).                                                     |
| [Query and subscribe to charts](/api/charts/query-charts)                  | Read one chart with `chart(id:)` or list a dashboard's charts with `charts(filter:)`, including the `todoFilter` recalculation path, `chartData` for the records behind a chart's numbers, and `subscribeToChart` for real-time create/update/delete events. |

<Callout variant="warning" title="Results are asynchronous">

Creating, copying, and recalculating charts enqueue a background calculation rather than returning final numbers. The `Chart` flags `isCalculating`, `needCalculation`, and `isCalculatingWithFilter` reflect that pending state; final results arrive via [`subscribeToChart`](/api/charts/query-charts), not synchronously from the mutation.

</Callout>

## Related

- [Dashboards](/api/dashboards) — Charts always belong to a dashboard; access is inherited from it.
- [Custom fields](/api/custom-fields) — Segment values and metadata axes can aggregate over a `customFieldId`.
- [List records](/api/records/list-records) — The `TodoFilterInput` used to scope a chart's records is the same filter used to query records.
