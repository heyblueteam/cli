---
title: Create Dashboard
description: Create a dashboard in an organization (optionally scoped to a workspace) using the createDashboard mutation.
icon: Plus
order: 1
---

Use the `createDashboard` mutation to create a new dashboard. A dashboard is a container for charts and reports; you add charts to it after it exists. The caller automatically becomes the dashboard's creator and is added as its first `DashboardUser`. Dashboards are `Dashboard` objects in the API.

A dashboard always belongs to an organization (`companyId`). You can optionally scope it to a single workspace by passing `projectId`; otherwise it spans the whole organization.

## Request

```graphql
mutation CreateDashboard {
  createDashboard(input: { companyId: "company_123", title: "Sales Performance" }) {
    id
    title
    createdAt
  }
}
```

Send the request to `https://api.blue.app/graphql` with your authentication headers:

```bash
curl https://api.blue.app/graphql \
  -H "Content-Type: application/json" \
  -H "blue-token-id: YOUR_TOKEN_ID" \
  -H "blue-token-secret: YOUR_TOKEN_SECRET" \
  -H "blue-org-id: YOUR_ORG_ID" \
  -d '{"query": "mutation { createDashboard(input: { companyId: \"company_123\", title: \"Sales Performance\" }) { id title createdAt } }"}'
```

Headers are case-insensitive. `companyId` (and `projectId`) accept either the object's ID or its slug.

## Parameters

### CreateDashboardInput

| Parameter   | Type      | Required | Description                                                                                                                                           |
| ----------- | --------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------- |
| `companyId` | `String!` | Yes      | ID or slug of the organization to create the dashboard in. You must be an `OWNER`, `ADMIN`, or `MEMBER` of it.                                        |
| `title`     | `String!` | Yes      | Display name for the dashboard.                                                                                                                       |
| `projectId` | `String`  | No       | ID or slug of a workspace to scope the dashboard to. You must be an `OWNER`, `ADMIN`, or `MEMBER` of it. Omit to make an organization-wide dashboard. |

## Response

```json
{
  "data": {
    "createDashboard": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Sales Performance",
      "createdAt": "2026-05-29T14:21:08.000Z"
    }
  }
}
```

### Dashboard

| Field            | Type               | Description                                                                |
| ---------------- | ------------------ | -------------------------------------------------------------------------- |
| `id`             | `ID!`              | Unique identifier of the dashboard.                                        |
| `title`          | `String!`          | The dashboard title.                                                       |
| `createdBy`      | `User!`            | The user who created the dashboard.                                        |
| `dashboardUsers` | `[DashboardUser!]` | Users with access to the dashboard. On creation this contains the creator. |
| `createdAt`      | `DateTime!`        | When the dashboard was created.                                            |
| `updatedAt`      | `DateTime!`        | When the dashboard was last modified.                                      |

### DashboardUser

| Field       | Type             | Description                                     |
| ----------- | ---------------- | ----------------------------------------------- |
| `id`        | `ID!`            | Unique identifier of the dashboard-user record. |
| `uid`       | `String!`        | Stable short identifier for the record.         |
| `user`      | `User!`          | The user with access to the dashboard.          |
| `role`      | `DashboardRole!` | The user's access level: `EDITOR` or `VIEWER`.  |
| `createdAt` | `DateTime!`      | When the user was granted access.               |
| `updatedAt` | `DateTime!`      | When the access record was last modified.       |

## Full example

Create a workspace-scoped dashboard and read back the creator and access list:

```graphql
mutation CreateWorkspaceDashboard {
  createDashboard(
    input: { companyId: "company_123", projectId: "project_123", title: "Q4 Project Metrics" }
  ) {
    id
    title
    createdBy {
      id
      email
      fullName
    }
    dashboardUsers {
      id
      role
      user {
        id
        email
        fullName
      }
    }
    createdAt
    updatedAt
  }
}
```

## Errors

| Code                 | When                                                                                                                                        |
| -------------------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `COMPANY_NOT_FOUND`  | The `companyId` doesn't resolve, or you aren't an `OWNER`/`ADMIN`/`MEMBER` of that organization. Message: `Company was not found.`          |
| `PROJECT_NOT_FOUND`  | A `projectId` was given but doesn't resolve, or you aren't an `OWNER`/`ADMIN`/`MEMBER` of that workspace. Message: `Project was not found.` |
| `PLAN_LIMIT_REACHED` | The organization is at its plan's dashboard limit. Message: `Dashboards limit reached (current/limit). Upgrade your plan for more.`         |

## Permissions

The caller must be an `OWNER`, `ADMIN`, or `MEMBER` of the organization — a `GUEST` or non-member is rejected with `COMPANY_NOT_FOUND`. When `projectId` is supplied, the same role requirement applies to that workspace.

Dashboard creation is also subject to a per-plan dashboard cap. Once an organization reaches the limit for its tier, further `createDashboard` calls fail with `PLAN_LIMIT_REACHED` until a dashboard is deleted or the plan is upgraded.

## Related

- [List Dashboards](/api/dashboards) — retrieve dashboards you can access.
- [Rename Dashboard](/api/dashboards/rename-dashboard) — update a dashboard's title and manage its users.
- [Copy Dashboard](/api/dashboards/copy-dashboard) — duplicate a dashboard with its charts.
- [Delete Dashboard](/api/dashboards/delete-dashboard) — permanently remove a dashboard.
