---
title: Edit Dashboard
description: Rename a dashboard and manage who can view or edit it with the editDashboard mutation.
icon: FileEdit
order: 2
---

Use the `editDashboard` mutation to rename a dashboard and manage its access list in a single call. It updates the dashboard `title` and replaces the set of users granted `VIEWER` or `EDITOR` access. There is no separate `renameDashboard` mutation — renaming is done by passing a new `title` here.

Dashboards are `Dashboard` objects in the API. Each shared user is a `DashboardUser` linking a `User` to a `DashboardRole`.

<Callout variant="warning" title="dashboardUsers is a full replacement set">

When you pass `dashboardUsers`, it becomes the complete access list. Any user currently on the dashboard who is **not** included in the array is **removed**. To keep an existing user, include them in every call that passes `dashboardUsers`. Omit `dashboardUsers` entirely to leave the access list untouched.

</Callout>

## Request

Rename a dashboard. Pass only `id` and the new `title`.

```graphql
mutation EditDashboard {
  editDashboard(input: { id: "dashboard_123", title: "Q4 Sales Dashboard" }) {
    id
    title
    updatedAt
  }
}
```

## Parameters

### EditDashboardInput

| Parameter              | Type                        | Required | Description                                                                                                                                               |
| ---------------------- | --------------------------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `id`                   | `String!`                   | Yes      | ID of the dashboard to edit.                                                                                                                              |
| `title`                | `String`                    | No       | New title. Omit to leave the title unchanged. Changing the title requires being the dashboard creator (see [Permissions](#permissions)).                  |
| `dashboardUsers`       | `[EditDashboardUserInput!]` | No       | Full desired access list. Users present are added or updated; existing users omitted from the array are removed. Omit to leave the access list unchanged. |
| `allowViewerChartData` | `Boolean`                   | No       | Whether viewers may inspect and export records behind chart totals. Only the dashboard creator may change it.                                             |

### EditDashboardUserInput

| Parameter | Type             | Required | Description                                                                  |
| --------- | ---------------- | -------- | ---------------------------------------------------------------------------- |
| `userId`  | `String!`        | Yes      | ID of the user to grant access. Must belong to the dashboard's organization. |
| `role`    | `DashboardRole!` | Yes      | Access level to assign.                                                      |

### DashboardRole

| Value    | Description                                                                                                               |
| -------- | ------------------------------------------------------------------------------------------------------------------------- |
| `EDITOR` | Can view the dashboard and edit its charts, filters, and layout. Can also call `editDashboard` to manage the access list. |
| `VIEWER` | Can view the dashboard only.                                                                                              |

## Response

```json
{
  "data": {
    "editDashboard": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Q4 Sales Dashboard",
      "updatedAt": "2026-05-29T14:22:08.000Z"
    }
  }
}
```

### Returns

`editDashboard` returns the updated `Dashboard`.

| Field            | Type               | Description                            |
| ---------------- | ------------------ | -------------------------------------- |
| `id`             | `ID!`              | Unique identifier for the dashboard.   |
| `title`          | `String!`          | Current dashboard title.               |
| `createdBy`      | `User!`            | User who created the dashboard.        |
| `dashboardUsers` | `[DashboardUser!]` | Users granted access to the dashboard. |
| `createdAt`      | `DateTime!`        | When the dashboard was created.        |
| `updatedAt`      | `DateTime!`        | When the dashboard was last modified.  |

Each `DashboardUser` exposes `id` (`ID!`), `uid` (`String!`), `user` (`User!`), `role` (`DashboardRole!`), `createdAt` (`DateTime!`), and `updatedAt` (`DateTime!`).

## Full example

Rename the dashboard and set its access list. After this call the dashboard's shared users are exactly `user_123` (as `EDITOR`) and `user_456` (as `VIEWER`); any previously shared user not listed here is removed.

```graphql
mutation EditDashboardWithUsers {
  editDashboard(
    input: {
      id: "dashboard_123"
      title: "Revenue Overview"
      dashboardUsers: [{ userId: "user_123", role: EDITOR }, { userId: "user_456", role: VIEWER }]
    }
  ) {
    id
    title
    dashboardUsers {
      id
      role
      user {
        id
        email
        firstName
        lastName
      }
    }
    updatedAt
  }
}
```

## Errors

| Code                  | When                                                                                                                                                                                                                                         |
| --------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `DASHBOARD_NOT_FOUND` | No dashboard with the given `id` is modifiable by the caller. The caller must be the creator or a current `EDITOR`. `VIEWER`s and unrelated users get this code — the dashboard is treated as not found. Message: `Dashboard was not found.` |
| `FORBIDDEN`           | A non-creator tried to rename the dashboard or change `allowViewerChartData`. Message: `You are not authorized.`                                                                                                                             |
| `USER_NOT_IN_COMPANY` | A `userId` in `dashboardUsers` does not belong to the dashboard's organization. Message: `User is not in the company.`                                                                                                                       |

## Permissions

`editDashboard` requires **modify access** to the dashboard: the caller must be the dashboard's creator or a `DashboardUser` with the `EDITOR` role. Callers without modify access (including `VIEWER`s) receive `DASHBOARD_NOT_FOUND`.

Within that access:

- **Managing `dashboardUsers`** — any caller with modify access (creator or `EDITOR`) can add, update, or remove shared users.
- **Renaming (`title`)** — only the dashboard **creator** can change the title. An `EDITOR` who passes a different `title` is rejected with `FORBIDDEN`. Passing the current title (or omitting `title`) is always allowed, so an editor can include the unchanged title alongside a `dashboardUsers` update.
- **Viewer chart data (`allowViewerChartData`)** — only the dashboard **creator** can change whether viewers may inspect or export the records behind chart totals. Editors always retain access to those records.

## Related

- [Create a dashboard](/api/dashboards/create-dashboard)
- [List dashboards](/api/dashboards)
- [Copy a dashboard](/api/dashboards/copy-dashboard)
- [Delete a dashboard](/api/dashboards/delete-dashboard)
