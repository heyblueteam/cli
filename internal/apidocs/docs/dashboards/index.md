---
title: List Dashboards
description: Retrieve a paginated list of the dashboards you can access in an organization using the dashboards query.
icon: ChartBar
order: 0
---

Use the `dashboards` query to retrieve a paginated list of the dashboards you can access in an organization, optionally scoped to a single workspace. The query returns only dashboards you created or that have been shared with you. Dashboards are `Dashboard` objects in the API; organizations are `Company` objects and workspaces are `Project` objects.

Results are paginated with offset-based `skip`/`take` — there is no cursor pagination. Read `pageInfo.totalItems` and `pageInfo.totalPages` to size and walk the pages.

## Request

The smallest valid call passes a `filter` with the required `companyId`:

```graphql
query ListDashboards {
  dashboards(filter: { companyId: "company_123" }) {
    items {
      id
      title
      createdBy {
        id
        fullName
        email
      }
      createdAt
      updatedAt
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

`companyId` accepts the organization's ID or its slug.

## Parameters

| Argument | Type                    | Required | Description                                                             |
| -------- | ----------------------- | -------- | ----------------------------------------------------------------------- |
| `filter` | `DashboardFilterInput!` | Yes      | Scopes the results to an organization (and optionally a workspace).     |
| `sort`   | `[DashboardSort!]`      | No       | Ordering applied to the results. Defaults to `[updatedAt_DESC]`.        |
| `skip`   | `Int`                   | No       | Number of dashboards to skip before returning results. Defaults to `0`. |
| `take`   | `Int`                   | No       | Number of dashboards to return per page. Defaults to `20`.              |

### DashboardFilterInput

| Field       | Type      | Required | Description                                                                                    |
| ----------- | --------- | -------- | ---------------------------------------------------------------------------------------------- |
| `companyId` | `String!` | Yes      | ID or slug of the organization whose dashboards to list.                                       |
| `projectId` | `String`  | No       | ID or slug of a workspace. When set, only dashboards belonging to that workspace are returned. |

### DashboardSort

`sort` takes an array of these enum values; later values break ties from earlier ones.

| Value            | Description                            |
| ---------------- | -------------------------------------- |
| `title_ASC`      | Title, A to Z.                         |
| `title_DESC`     | Title, Z to A.                         |
| `createdBy_ASC`  | Creator's first name, A to Z.          |
| `createdBy_DESC` | Creator's first name, Z to A.          |
| `updatedAt_ASC`  | Least recently updated first.          |
| `updatedAt_DESC` | Most recently updated first (default). |

## Response

```json
{
  "data": {
    "dashboards": {
      "items": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "title": "Q2 Delivery Overview",
          "createdBy": {
            "id": "clm4n8qwx000108l0a1b2c3d4",
            "fullName": "Ada Lovelace",
            "email": "ada@example.com"
          },
          "createdAt": "2026-04-12T09:30:00.000Z",
          "updatedAt": "2026-05-28T14:05:11.000Z"
        }
      ],
      "pageInfo": {
        "totalItems": 1,
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

### DashboardPagination

| Field      | Type            | Description                                          |
| ---------- | --------------- | ---------------------------------------------------- |
| `items`    | `[Dashboard!]!` | The dashboards on the current page.                  |
| `pageInfo` | `PageInfo!`     | Offset-based pagination metadata for the result set. |

### Dashboard

| Field                  | Type               | Description                                                                             |
| ---------------------- | ------------------ | --------------------------------------------------------------------------------------- |
| `id`                   | `ID!`              | Unique identifier for the dashboard.                                                    |
| `title`                | `String!`          | Display name of the dashboard.                                                          |
| `createdBy`            | `User!`            | The user who created the dashboard.                                                     |
| `dashboardUsers`       | `[DashboardUser!]` | Users the dashboard has been shared with. Null when no sharing records exist.           |
| `allowViewerChartData` | `Boolean!`         | Whether viewers may inspect and export records behind chart totals. Editors always can. |
| `createdAt`            | `DateTime!`        | When the dashboard was created.                                                         |
| `updatedAt`            | `DateTime!`        | When the dashboard was last modified.                                                   |

### DashboardUser

| Field       | Type             | Description                                           |
| ----------- | ---------------- | ----------------------------------------------------- |
| `id`        | `ID!`            | Unique identifier for the sharing record.             |
| `uid`       | `String!`        | Stable public identifier for the sharing record.      |
| `user`      | `User!`          | The user the dashboard is shared with.                |
| `role`      | `DashboardRole!` | The shared user's access level: `VIEWER` or `EDITOR`. |
| `createdAt` | `DateTime!`      | When the dashboard was shared with this user.         |
| `updatedAt` | `DateTime!`      | When this sharing record was last modified.           |

### PageInfo

| Field             | Type       | Description                                                        |
| ----------------- | ---------- | ------------------------------------------------------------------ |
| `totalItems`      | `Int`      | Total number of dashboards matching the filter, across all pages.  |
| `totalPages`      | `Int`      | Total number of pages at the current `take`.                       |
| `page`            | `Int`      | The current page number (1-based), derived from `skip` and `take`. |
| `perPage`         | `Int`      | Number of items per page (mirrors `take`).                         |
| `hasNextPage`     | `Boolean!` | Whether another page follows the current one.                      |
| `hasPreviousPage` | `Boolean!` | Whether a page precedes the current one.                           |

<Callout variant="info" title="Offset pagination only">

`PageInfo` also exposes `startCursor` and `endCursor`, but both are deprecated — Blue dashboards are not cursor-paginated. Walk pages with `skip`/`take` and use `totalItems`/`totalPages` to size the result set.

</Callout>

## Full example

Scope to a single workspace, sort by most recently updated then title, take the second page of 10, and expand the sharing list:

```graphql
query ListWorkspaceDashboards {
  dashboards(
    filter: { companyId: "company_123", projectId: "project_123" }
    sort: [updatedAt_DESC, title_ASC]
    skip: 10
    take: 10
  ) {
    items {
      id
      title
      createdBy {
        id
        fullName
        email
      }
      dashboardUsers {
        id
        role
        user {
          id
          fullName
          email
        }
      }
      createdAt
      updatedAt
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

## Permissions

You must be authenticated. The query returns only dashboards where you are the creator or have been added as a `DashboardUser`. An unknown or inaccessible `companyId` (or `projectId`) does not raise an error — it simply returns an empty `items` list, so this query never throws `COMPANY_NOT_FOUND` or `PROJECT_NOT_FOUND`.

## Related

- [Create Dashboard](/api/dashboards/create-dashboard) — Create a new dashboard.
- [Edit Dashboard](/api/dashboards/rename-dashboard) — Rename a dashboard and manage shared users.
- [Copy Dashboard](/api/dashboards/copy-dashboard) — Duplicate an existing dashboard.
- [Delete Dashboard](/api/dashboards/delete-dashboard) — Permanently remove a dashboard.
