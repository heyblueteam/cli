---
title: Delete Dashboard
description: Permanently delete a dashboard and all of its charts using the Blue API.
icon: Trash2
order: 4
---

Use the `deleteDashboard` mutation to permanently delete a dashboard, along with the charts and chart segments it contains. Dashboards are `Dashboard` objects in the API. This is a hard delete — there is no soft-delete or undo — and only the dashboard's creator can perform it.

## Request

```graphql
mutation DeleteDashboard {
  deleteDashboard(id: "dashboard_123") {
    success
    operationId
  }
}
```

## Parameters

| Parameter | Type      | Required | Description                            |
| --------- | --------- | -------- | -------------------------------------- |
| `id`      | `String!` | Yes      | Identifier of the dashboard to delete. |

## Response

```json
{
  "data": {
    "deleteDashboard": {
      "success": true,
      "operationId": null
    }
  }
}
```

### Returns

`deleteDashboard` returns a `MutationResult`.

| Field         | Type       | Description                                                                                                                                        |
| ------------- | ---------- | -------------------------------------------------------------------------------------------------------------------------------------------------- |
| `success`     | `Boolean!` | `true` when the dashboard was deleted. Returns `false` (with no error) if the delete fails internally — see [Failure behavior](#failure-behavior). |
| `operationId` | `String`   | Identifier of the resulting async operation, if one was queued. `null` for this mutation, which deletes synchronously.                             |

## Permissions

The resolver enforces access in two stages, which produce two distinct outcomes:

1. **Access check.** The caller must be the creator of the dashboard _or_ have the `EDITOR` role on it. A caller with neither (including viewers and callers who can't see the dashboard at all) receives `DASHBOARD_NOT_FOUND`.
2. **Creator check.** Of those who pass the access check, only the **creator** may delete. An `EDITOR` who is not the creator receives `FORBIDDEN`.

There is no role-based override: company owners and admins cannot delete a dashboard created by someone else.

## Failure behavior

If the underlying delete fails internally after the permission checks pass, the mutation does **not** throw — it returns `{ "success": false }`. Always check the `success` field rather than assuming a non-error response means the dashboard was removed.

Deleting a dashboard publishes a `DashboardDeleted` event to anyone subscribed to that dashboard's company, so live clients drop it from their view. It does not affect workspaces, records, or any other company data — only the dashboard, its charts, and its chart segments are removed.

## Errors

| Code                  | When                                                                                                                                            |
| --------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| `DASHBOARD_NOT_FOUND` | No dashboard with that `id` is visible to the caller, or the caller is neither the creator nor an `EDITOR`. Message: `Dashboard was not found.` |
| `FORBIDDEN`           | The caller can modify the dashboard (is an `EDITOR`) but is not its creator. Message: `You are not authorized.`                                 |
| `UNAUTHENTICATED`     | The request is missing or has invalid credentials. Message: `You are not authenticated.`                                                        |

```json
{
  "errors": [
    {
      "message": "You are not authorized.",
      "extensions": { "code": "FORBIDDEN" }
    }
  ]
}
```

## Related

- [Create a dashboard](/api/dashboards/create-dashboard)
- [Rename a dashboard](/api/dashboards/rename-dashboard)
- [Copy a dashboard](/api/dashboards/copy-dashboard)
- [Dashboards overview](/api/dashboards)
