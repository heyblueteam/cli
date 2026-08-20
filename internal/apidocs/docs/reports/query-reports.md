---
title: Query Reports
description: Fetch a single report by id or list the reports in an organization, with pagination and the full Report field reference.
icon: Search
order: 2
---

Use the `report` query to fetch a single report by id, and the `reports` query to list the reports you can see in an organization. A report aggregates records (`Todo` objects) from one or more workspaces into a saved analytical view; these two queries read the report's metadata and configuration. To read the records, counts, and aggregations a report produces, see [Report Data & Aggregations](/api/reports/report-data-aggregations).

Both queries are visibility-scoped: you only ever see a report you created or one that has been shared with you.

## Request

Fetch one report by id with its data sources:

```graphql
query GetReport {
  report(id: "report_123") {
    id
    title
    description
    lastGeneratedAt
    projectIds
    createdBy {
      fullName
    }
    dataSources {
      id
      name
      sourceType
      projectIds
      order
    }
  }
}
```

List the reports you can see in an organization, newest-updated first:

```graphql
query ListReports {
  reports(filter: { companyId: "company_123" }, skip: 0, take: 20) {
    totalCount
    hasNextPage
    hasPreviousPage
    items {
      id
      title
      updatedAt
    }
  }
}
```

The organization is identified by `companyId` (accepts the organization ID or slug). The `report` query is identified by the report's `id`.

## Parameters

### `report` arguments

| Parameter | Type      | Required | Description                                                                                   |
| --------- | --------- | -------- | --------------------------------------------------------------------------------------------- |
| `id`      | `String!` | Yes      | The report's `id`. Returns `null`-equivalent behavior only via error — see [Errors](#errors). |

### `reports` arguments

| Parameter | Type            | Required | Description                                                           |
| --------- | --------------- | -------- | --------------------------------------------------------------------- |
| `filter`  | `ReportFilter!` | Yes      | Scopes the list to one organization, with an optional creator filter. |
| `skip`    | `Int`           | No       | Number of reports to skip. Defaults to `0`.                           |
| `take`    | `Int`           | No       | Number of reports to return. Defaults to `20`.                        |

### ReportFilter

| Field         | Type      | Required | Description                                                                                                                            |
| ------------- | --------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| `companyId`   | `String!` | Yes      | The organization (`Company`) to list reports for. Accepts ID or slug.                                                                  |
| `createdById` | `String`  | No       | Restrict the list to reports created by this user. Applied on top of the visibility scope.                                             |
| `adminView`   | `Boolean` | No       | Return every report in the organization, not only yours and those shared with you. Honoured only for an OWNER or ADMIN of `companyId`. |

<Callout variant="info" title="Pagination is applied after the visibility scope">

`reports` first loads every report you created or that was shared with you in the organization, sorts them by `updatedAt` descending, then applies `createdById`, `skip`, and `take`. `totalCount` reflects the count after filtering but before pagination; `hasNextPage` is `skip + take < totalCount` and `hasPreviousPage` is `skip > 0`.

</Callout>

### Listing every report in an organization

`adminView: true` drops the created-by / shared-with restriction, so the list returns every report in `companyId`. Use it to reach reports left behind by people who have left the organization.

The server proves the caller's membership rather than trusting the flag: it checks for an `OWNER` or `ADMIN` `CompanyUser` row in that same `companyId`. A caller who is neither is **silently narrowed** to their normal scope rather than refused — the query succeeds and simply returns their own visible reports. Because the check is against the requested organization, passing another organization's id reaches nothing.

## Response

`report`:

```json
{
  "data": {
    "report": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Q3 Delivery Pipeline",
      "description": "Open work across the delivery workspaces.",
      "lastGeneratedAt": "2026-05-28T09:14:22.000Z",
      "projectIds": ["clm4projalpha0001", "clm4projbeta00002"],
      "createdBy": { "fullName": "Dana Reyes" },
      "dataSources": [
        {
          "id": "clm4dsrc00000abcd1234",
          "name": "Delivery",
          "sourceType": "TODOS",
          "projectIds": ["clm4projalpha0001", "clm4projbeta00002"],
          "order": 0
        }
      ]
    }
  }
}
```

`reports`:

```json
{
  "data": {
    "reports": {
      "totalCount": 3,
      "hasNextPage": false,
      "hasPreviousPage": false,
      "items": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "title": "Q3 Delivery Pipeline",
          "updatedAt": "2026-05-28T09:14:22.000Z"
        },
        {
          "id": "clm4r2abc000108l0aaaa1111",
          "title": "Support Backlog",
          "updatedAt": "2026-05-21T16:02:10.000Z"
        },
        {
          "id": "clm4r9def000208l0bbbb2222",
          "title": "Marketing Calendar",
          "updatedAt": "2026-05-19T11:48:55.000Z"
        }
      ]
    }
  }
}
```

### ReportPagination

| Field             | Type         | Description                                              |
| ----------------- | ------------ | -------------------------------------------------------- |
| `items`           | `[Report!]!` | The page of reports, sorted by `updatedAt` descending.   |
| `totalCount`      | `Int!`       | Total reports matching the filter, before `skip`/`take`. |
| `hasNextPage`     | `Boolean!`   | `true` when more reports exist after this page.          |
| `hasPreviousPage` | `Boolean!`   | `true` when `skip` is greater than `0`.                  |

### Report

| Field             | Type                   | Description                                                                                                                                                                   |
| ----------------- | ---------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `id`              | `ID!`                  | Unique report identifier.                                                                                                                                                     |
| `uid`             | `String!`              | Short public-facing identifier.                                                                                                                                               |
| `title`           | `String!`              | Report name.                                                                                                                                                                  |
| `description`     | `String`               | Optional description (up to 10,000 characters).                                                                                                                               |
| `config`          | `JSONObject`           | Opaque key/value blob holding display configuration (selected fields, layout). Its shape is defined by the app, not the GraphQL schema.                                       |
| `company`         | `Company!`             | The organization the report belongs to.                                                                                                                                       |
| `createdBy`       | `User!`                | The user who created the report. Their workspace permissions are used to resolve the report's records and project list.                                                       |
| `createdAt`       | `DateTime!`            | When the report was created.                                                                                                                                                  |
| `updatedAt`       | `DateTime!`            | When the report was last modified. Drives the `reports` sort order.                                                                                                           |
| `lastGeneratedAt` | `DateTime`             | When the report's aggregations were last invalidated/regenerated. `null` until first refresh. Bumped by [`refreshReportAggregations`](/api/reports/report-data-aggregations). |
| `dataSources`     | `[ReportDataSource!]!` | The data sources that scope the report's records.                                                                                                                             |
| `reportUsers`     | `[ReportUser!]!`       | Users the report is shared with, with their role. See [Manage Report Access](/api/reports/manage-report-access).                                                              |
| `projectIds`      | `[String!]`            | Computed union of every workspace ID across all data sources that the report creator can access. See the note below.                                                          |

The `Report` type also exposes the `todos`, `todoCount`, and `aggregations` fields. Those are the report's data surface and are documented on [Report Data & Aggregations](/api/reports/report-data-aggregations), which also covers the `TodosFilter`/`TodosSort` inputs (shared with the [Records](/api/records) section).

<Callout variant="info" title="projectIds is computed, not stored">

`Report.projectIds` is derived at read time: it unions the workspaces from every data source, resolved against the report creator's permissions. If **any** data source has `projectIds: null` (meaning "all workspaces"), the field returns every workspace the creator can access. Archived workspaces are excluded. This differs from `ReportDataSource.projectIds`, which reflects a single source's stored scope.

</Callout>

### ReportDataSource

| Field        | Type                | Description                                                                                                                                             |
| ------------ | ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `id`         | `ID!`               | Unique data-source identifier.                                                                                                                          |
| `uid`        | `String!`           | Short public-facing identifier.                                                                                                                         |
| `name`       | `String`            | Optional label for the data source.                                                                                                                     |
| `report`     | `Report!`           | The parent report.                                                                                                                                      |
| `sourceType` | `ReportSourceType!` | The kind of records this source pulls. Only `TODOS` exists today.                                                                                       |
| `projectIds` | `[String!]`         | The workspaces this source scopes to. `null` means **all** workspaces the report creator can access; a non-empty array means those specific workspaces. |
| `projects`   | `[Project!]`        | The resolved workspaces for this source (all accessible workspaces when `projectIds` is `null`), sorted by name.                                        |
| `filters`    | `JSONObject`        | Opaque record filter applied to this source. Mirrors the Records filter structure; its keys are defined by the app, not the GraphQL schema.             |
| `order`      | `Int!`              | Position of the source within the report.                                                                                                               |
| `createdAt`  | `DateTime!`         | When the data source was created.                                                                                                                       |
| `updatedAt`  | `DateTime!`         | When the data source was last modified.                                                                                                                 |

### ReportSourceType

| Value   | Description                                                                          |
| ------- | ------------------------------------------------------------------------------------ |
| `TODOS` | The source pulls records (`Todo` objects) from its workspaces. The only value today. |

## Errors

| Code               | When                                                                                     |
| ------------------ | ---------------------------------------------------------------------------------------- |
| `REPORT_NOT_FOUND` | The `id` does not exist, or you are neither the creator nor a shared user of the report. |
| `UNAUTHENTICATED`  | The request has no valid credentials.                                                    |

Both queries require an authenticated user. `report` raises `REPORT_NOT_FOUND` rather than returning `null` when the report is invisible to you, so a missing report and a forbidden one are indistinguishable by design. `reports` never errors on visibility — it simply omits reports you cannot see.

## Permissions

Report visibility is per-report, independent of workspace membership:

- **`report` (view)** — succeeds only if you are the report's `createdBy` **or** appear in its `reportUsers` (with either `EDITOR` or `VIEWER` role).
- **`reports` (list)** — returns every report in `filter.companyId` that you created **or** that was shared with you, sorted by `updatedAt` descending. Reports you cannot see are silently excluded; the `createdById` filter narrows this set further. An organization OWNER or ADMIN can pass [`adminView: true`](#listing-every-report-in-an-organization) to widen the list to every report in that organization.

Modifying, deleting, and exporting reports have their own (different) access tiers — see [Manage Report Access](/api/reports/manage-report-access), [Duplicate & Delete a Report](/api/reports/duplicate-and-delete-report), and [Export a Report](/api/reports/export-report).

## Related

- [Reports overview](/api/reports)
- [Create a Report](/api/reports/create-report)
- [Report Data & Aggregations](/api/reports/report-data-aggregations)
- [Manage Report Access](/api/reports/manage-report-access)
- [Duplicate & Delete a Report](/api/reports/duplicate-and-delete-report)
- [Export a Report](/api/reports/export-report)
- [Records](/api/records)
