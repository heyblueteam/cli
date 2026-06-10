---
title: Copy Dashboard
description: Duplicate a dashboard with all its charts, segments, and viewer permissions using the copyDashboard mutation.
icon: Copy
order: 3
---

Use the `copyDashboard` mutation to duplicate an existing dashboard. The copy is a deep copy: every chart, chart segment, segment value, and formula is recreated with fresh identifiers, and the caller becomes the creator of the new dashboard. Dashboards are `Dashboard` objects in the API.

The mutation returns the new `Dashboard`. Its charts are recreated server-side but are not exposed on the `Dashboard` type — fetch them afterward with the [`charts`](/api/charts) query.

## Request

Copy a dashboard and give the copy a new title.

```graphql
mutation CopyDashboard {
  copyDashboard(input: { dashboardId: "dashboard_123", title: "Q4 Sales Dashboard" }) {
    id
    title
    createdAt
  }
}
```

Omit `title` to fall back to the original title with `" (Copy)"` appended.

```graphql
mutation CopyDashboardWithDefaultTitle {
  copyDashboard(input: { dashboardId: "dashboard_123" }) {
    id
    title
  }
}
```

## Parameters

### CopyDashboardInput

| Parameter     | Type      | Required | Description                                                                           |
| ------------- | --------- | -------- | ------------------------------------------------------------------------------------- |
| `dashboardId` | `String!` | Yes      | ID of the dashboard to copy. You must be its creator or have the `EDITOR` role on it. |
| `title`       | `String`  | No       | Title for the copy. If omitted, the copy is titled `"<original title> (Copy)"`.       |

## Response

```json
{
  "data": {
    "copyDashboard": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Q4 Sales Dashboard",
      "createdAt": "2026-05-29T14:03:11.000Z"
    }
  }
}
```

### Returns

The mutation returns the newly created `Dashboard`.

| Field            | Type               | Description                                                                                                        |
| ---------------- | ------------------ | ------------------------------------------------------------------------------------------------------------------ |
| `id`             | `ID!`              | Unique identifier of the new dashboard.                                                                            |
| `title`          | `String!`          | Title of the copy.                                                                                                 |
| `createdBy`      | `User!`            | The user who ran the copy. The caller always becomes the creator.                                                  |
| `dashboardUsers` | `[DashboardUser!]` | Viewer and editor permissions carried over from the original, excluding the calling user (who is now the creator). |
| `createdAt`      | `DateTime!`        | When the copy was created.                                                                                         |
| `updatedAt`      | `DateTime!`        | When the copy was last modified.                                                                                   |

Charts are deep-copied server-side but are **not** fields on the `Dashboard` type, so you cannot select them on the `copyDashboard` result. Read them with the [`charts`](/api/charts) query once the copy exists. Chart results are recalculated asynchronously after the copy, so chart data may briefly appear empty.

### DashboardUser

| Field       | Type             | Description                                         |
| ----------- | ---------------- | --------------------------------------------------- |
| `id`        | `ID!`            | Unique identifier of the dashboard-user assignment. |
| `uid`       | `String!`        | Stable UID for the assignment.                      |
| `user`      | `User!`          | The user granted access.                            |
| `role`      | `DashboardRole!` | `EDITOR` or `VIEWER`.                               |
| `createdAt` | `DateTime!`      | When the assignment was created.                    |
| `updatedAt` | `DateTime!`      | When the assignment was last modified.              |

### DashboardRole

| Value    | Description                                                    |
| -------- | -------------------------------------------------------------- |
| `EDITOR` | Can view and edit the dashboard's charts, filters, and layout. |
| `VIEWER` | Can only view the dashboard.                                   |

## Full example

Copy a dashboard and read back its carried-over permissions.

```graphql
mutation CopyDashboardWithUsers {
  copyDashboard(input: { dashboardId: "dashboard_123", title: "Marketing — Sales Team Copy" }) {
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

```json
{
  "data": {
    "copyDashboard": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Marketing — Sales Team Copy",
      "createdBy": {
        "id": "clm4n8qwx000108l0c2abf9rt",
        "email": "you@example.com",
        "fullName": "Alex Rivera"
      },
      "dashboardUsers": [
        {
          "id": "clm4n8qwx000208l0h7zk1msd",
          "role": "VIEWER",
          "user": {
            "id": "clm4n8qwx000308l0p3qe5xby",
            "email": "teammate@example.com",
            "fullName": "Jordan Lee"
          }
        }
      ],
      "createdAt": "2026-05-29T14:03:11.000Z",
      "updatedAt": "2026-05-29T14:03:11.000Z"
    }
  }
}
```

## What gets copied

- **The dashboard** — a new dashboard with a new ID and UID. The caller becomes the creator; the copy stays in the same organization as the original (you cannot copy a dashboard into a different organization).
- **Charts, segments, and segment values** — recreated with fresh identifiers, preserving titles, positions, types, and configuration.
- **Formulas** — segment formula references are rewritten to point at the copied segment values, so calculations keep working against the copied data.
- **Permissions** — existing `dashboardUsers` are carried over, **except** the calling user, who becomes the creator rather than a listed dashboard user.

## Errors

| Code                  | When                                                                                                                                                                            |
| --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `DASHBOARD_NOT_FOUND` | No dashboard matches `dashboardId` for the caller. This also covers insufficient access — a user who is neither the creator nor an `EDITOR` simply finds no matching dashboard. |

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

Omitting `dashboardId` is rejected by GraphQL as a non-null validation error before the resolver runs — it is not a Blue error code.

## Permissions

You can copy a dashboard only if you are its creator or hold the `EDITOR` role on it. `VIEWER` access is not enough.

| Access to the original dashboard | Can copy |
| -------------------------------- | -------- |
| Creator                          | Yes      |
| `EDITOR`                         | Yes      |
| `VIEWER`                         | No       |
| No access                        | No       |

A user without sufficient access receives `DASHBOARD_NOT_FOUND` rather than a distinct permission error.

## Related

- [List dashboards](/api/dashboards) — retrieve dashboards for an organization or workspace.
- [Create dashboard](/api/dashboards/create-dashboard) — create a dashboard from scratch.
- [Rename dashboard](/api/dashboards/rename-dashboard) — update a dashboard title or its users.
- [Delete dashboard](/api/dashboards/delete-dashboard) — remove a dashboard.
- [Charts](/api/charts) — query the charts on the copied dashboard.
