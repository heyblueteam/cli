---
title: Create a Report
description: Create a report scoped to one or more data sources, each targeting a set of workspaces and a JSON filter.
icon: Plus
order: 1
---

Use the `createReport` mutation to create a report — a saved analytical view that aggregates records across one or more workspaces. A report holds one or more **data sources**, and each data source scopes a set of workspaces plus a JSON filter. The calling user becomes the report's creator (`createdBy`) and is its only owner until they [share it](/api/reports/manage-report-access).

Records are `Todo` objects, and workspaces are `Project` objects in the API.

## Request

A call needs a `title` and a `dataSources` list (which may be empty, though a report with no sources aggregates nothing). The smallest useful call passes one data source; a data source with no `projectIds` targets **all** workspaces the report creator can access.

```graphql
mutation CreateReport {
  createReport(input: { title: "Q3 Pipeline", dataSources: [{ sourceType: TODOS }] }) {
    id
    title
    createdAt
  }
}
```

To scope a data source to specific workspaces and attach a filter, pass `projectIds` and `filters`:

```graphql
mutation CreateScopedReport {
  createReport(
    input: {
      title: "Open sales records"
      description: "Active deals across the EMEA and US workspaces."
      dataSources: [
        {
          name: "EMEA + US"
          sourceType: TODOS
          projectIds: ["project_123", "project_456"]
          filters: { done: false }
          order: 0
        }
      ]
    }
  ) {
    id
    title
    dataSources {
      id
      name
      projectIds
    }
  }
}
```

## Parameters

### CreateReportInput

| Parameter     | Type                              | Required | Description                                                                          |
| ------------- | --------------------------------- | -------- | ------------------------------------------------------------------------------------ |
| `title`       | `String!`                         | Yes      | The report's display name.                                                           |
| `description` | `String`                          | No       | Free-text description. Maximum 10,000 characters.                                    |
| `dataSources` | `[CreateReportDataSourceInput!]!` | Yes      | One or more data sources. The list may be empty, but at most 50 entries are allowed. |

### CreateReportDataSourceInput

| Parameter    | Type               | Required | Description                                                                                                                                                                 |
| ------------ | ------------------ | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `name`       | `String`           | No       | Optional label for the data source.                                                                                                                                         |
| `sourceType` | `ReportSourceType` | No       | What the data source aggregates. Defaults to `TODOS` (the only value today).                                                                                                |
| `projectIds` | `[String!]`        | No       | Workspace IDs to scope this source to. See the convention below — `null` or an empty array means **all** workspaces; a non-empty array means **those specific** workspaces. |
| `filters`    | `JSONObject`       | No       | A filter applied to records in this source. See [The `filters` field](#the-filters-field).                                                                                  |
| `order`      | `Int`              | No       | Sort position of the data source within the report. Defaults to its index in the `dataSources` list.                                                                        |

### ReportSourceType

| Value   | Description                                                                                         |
| ------- | --------------------------------------------------------------------------------------------------- |
| `TODOS` | Aggregate records (`Todo` objects). The only source type today; it is the default for `sourceType`. |

### The `projectIds` convention

`projectIds` controls which workspaces a data source draws from:

- **`null` or omitted** — the data source targets **every** workspace the report creator can access.
- **A non-empty array** — the data source targets exactly **those** workspaces.
- **An empty array `[]`** — coerced to `null`, so it also means **all** workspaces.

Workspace access is always resolved against the **report creator**, not the caller of later read queries. The `projectIds` you supply must be workspace (`Project`) IDs — slugs are not resolved.

The [`Report.projectIds`](/api/reports/query-reports) field returns the computed _union_ of project IDs across every data source. If any one source resolves to all workspaces, that field returns the creator's full accessible set.

### The `filters` field

`filters` is an opaque `JSONObject` — a free-form key/value blob whose shape is defined by the application, not by the GraphQL schema. It mirrors the record filter structure used elsewhere in the API; see the [Records filter inputs](/api/records/list-records) for the available keys. The example above uses `{ "done": false }` to keep a source to open records.

<Callout variant="info" title="Limits enforced at create time">

A report may declare at most **50 data sources**, and `description` may be at most **10,000 characters**. Exceeding either limit fails the whole mutation. (The `config` field — settable via [`updateReport`](/api/reports/manage-report-access) — has a separate 100&nbsp;KB cap and is not part of `CreateReportInput`.)

</Callout>

## Response

The mutation returns the newly created `Report`. The example below corresponds to the scoped request above.

```json
{
  "data": {
    "createReport": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Open sales records",
      "dataSources": [
        {
          "id": "clm4n8r1a000108l0a2bcde90",
          "name": "EMEA + US",
          "projectIds": ["project_123", "project_456"]
        }
      ]
    }
  }
}
```

### Returns

The mutation returns a `Report!`. Selected fields below; the full `Report` and `ReportDataSource` field reference lives on [Query Reports](/api/reports/query-reports).

| Field             | Type                   | Description                                                                                                                     |
| ----------------- | ---------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| `id`              | `ID!`                  | The report's unique identifier.                                                                                                 |
| `uid`             | `String!`              | Public unique identifier.                                                                                                       |
| `title`           | `String!`              | The report's display name.                                                                                                      |
| `description`     | `String`               | The description, or `null` if none was set.                                                                                     |
| `config`          | `JSONObject`           | Application-defined view configuration. `null` at creation — set later via [`updateReport`](/api/reports/manage-report-access). |
| `createdBy`       | `User!`                | The user who created the report. Always the calling user.                                                                       |
| `company`         | `Company!`             | The organization the report belongs to (the caller's active company).                                                           |
| `createdAt`       | `DateTime!`            | When the report was created.                                                                                                    |
| `updatedAt`       | `DateTime!`            | When the report was last modified.                                                                                              |
| `lastGeneratedAt` | `DateTime`             | When the report's aggregations were last computed. `null` until first generated.                                                |
| `dataSources`     | `[ReportDataSource!]!` | The report's data sources, in `order`.                                                                                          |
| `reportUsers`     | `[ReportUser!]!`       | Collaborators the report is shared with. Empty at creation.                                                                     |
| `projectIds`      | `[String!]`            | Computed union of all workspace IDs across data sources that the creator can access.                                            |

## Transactional backfill

Report creation runs inside a single database transaction. After the `Report` and its data sources are written, the API backfills the list-item rows for every workspace referenced by an explicit `projectIds` array (this is what makes multi-workspace record queries work for the report).

If the backfill fails for any referenced workspace, the **entire mutation rolls back** — no report is created — and the call fails with:

```
Cannot create report: Failed to backfill N of M projects. Failed projects: project_123, …
```

Data sources that target all workspaces (`projectIds: null`) reference no specific IDs, so they are not backfilled and cannot trigger this error.

## Errors

| Code              | When                                                                   |
| ----------------- | ---------------------------------------------------------------------- |
| `UNAUTHENTICATED` | No authenticated user on the request.                                  |
| `FORBIDDEN`       | The request has no active company context.                             |
| `BAD_USER_INPUT`  | Malformed input (e.g. a `JSONObject` field that isn't a valid object). |

The data-source and description limits, and the backfill rollback, throw a plain error (no machine-readable extension code) — the message states the cause, as shown above.

## Permissions

Any authenticated user with an active company context can create a report. The new report is scoped to the caller's company, and the caller becomes its `createdBy` / owner. The report is private until shared — see the report access model on [Manage Report Access](/api/reports/manage-report-access).

## Related

- [Reports overview](/api/reports)
- [Query Reports](/api/reports/query-reports) — full `Report` and `ReportDataSource` field reference, including the computed `projectIds` union.
- [Report Data & Aggregations](/api/reports/report-data-aggregations) — read the records, counts, and aggregations a report produces.
- [Manage Report Access](/api/reports/manage-report-access) — update the title, description, config, and data sources, and share with collaborators.
- [Duplicate & Delete a Report](/api/reports/duplicate-and-delete-report)
- [Export a Report](/api/reports/export-report)
- [List Records](/api/records/list-records) — the filter structure mirrored by `filters`.
