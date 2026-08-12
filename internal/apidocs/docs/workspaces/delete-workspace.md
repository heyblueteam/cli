---
title: Delete a Workspace
description: Delete a workspace with the deleteWorkspace mutation, moving it to the organization's Trash where an admin can restore it until the retention window expires.
icon: Trash2
order: 6
---

Use the `deleteWorkspace` mutation to delete a workspace and everything inside it — records, lists, comments, custom fields, automations, tags, and file attachments. Deletion moves the workspace to the organization's Trash: nothing inside it is destroyed, the workspace is hidden from every read, and its slug is freed so a "delete and recreate" flow works. An organization admin can [restore it](#restoring-a-deleted-workspace) until the organization's retention window expires, after which it is purged for good. Workspaces are `Workspace` objects in the API.

<Callout variant="warning" title="Purging is permanent">

A workspace in Trash is recoverable, but **Delete forever** (or reaching the end of the retention window) purges it — removing all of its records, lists, comments, custom field values, automations, tags, dependencies, and file attachments with no way to restore them. If you might need the workspace later, leave it in Trash or [archive it](/api/workspaces/archive-workspace) instead of purging.

</Callout>

## Request

```graphql
mutation DeleteWorkspace {
  deleteWorkspace(id: "project_123") {
    success
  }
}
```

The `id` argument accepts either the workspace ID or its slug.

## Parameters

| Parameter | Type      | Required | Description                                |
| --------- | --------- | -------- | ------------------------------------------ |
| `id`      | `String!` | Yes      | The ID or slug of the workspace to delete. |

## Response

The workspace is moved to Trash and `success: true` is returned. Its child data is left intact for a full-fidelity restore; the cleanup that removes it runs later, at purge time.

```json
{
  "data": {
    "deleteWorkspace": {
      "success": true
    }
  }
}
```

### Returns

`deleteWorkspace` returns a `MutationResult`.

| Field         | Type       | Description                                                                                                       |
| ------------- | ---------- | ----------------------------------------------------------------------------------------------------------------- |
| `success`     | `Boolean!` | `true` when the workspace was moved to Trash.                                                                     |
| `operationId` | `String`   | Identifier for an asynchronous follow-up job. `null` for a soft delete — the cleanup now runs at purge, not here. |

## Errors

| Code                | When                                                                                              |
| ------------------- | ------------------------------------------------------------------------------------------------- |
| `PROJECT_NOT_FOUND` | No workspace matches the supplied `id` in the current organization.                               |
| `FORBIDDEN`         | The authenticated user is not allowed to delete this workspace (see [Permissions](#permissions)). |

```json
{
  "errors": [
    {
      "message": "Project was not found.",
      "extensions": { "code": "PROJECT_NOT_FOUND" }
    }
  ]
}
```

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

## Permissions

To delete a workspace, the user must be both:

- An `OWNER`, `ADMIN`, or `MEMBER` at the organization level, **and**
- An `OWNER` or `ADMIN` at the workspace level.

| Workspace role | Can delete |
| -------------- | ---------- |
| `OWNER`        | Yes        |
| `ADMIN`        | Yes        |
| `MEMBER`       | No         |
| `CLIENT`       | No         |
| `COMMENT_ONLY` | No         |
| `VIEW_ONLY`    | No         |

## Restoring a deleted workspace

While the workspace is in the organization's Trash, an organization `OWNER` or `ADMIN` can restore it with the `restoreWorkspace` mutation, which returns the restored `Workspace`. Everything inside it comes back, and its original slug is reinstated (with a `-restored` suffix if the slug has since been reused).

```graphql
mutation RestoreWorkspace {
  restoreWorkspace(id: "project_123") {
    id
    name
    slug
  }
}
```

Once the workspace is purged — via **Delete forever** or by reaching the end of the organization's retention window — it can no longer be restored.

## Related

- [Archive a workspace](/api/workspaces/archive-workspace) — the reversible alternative.
- [Create a workspace](/api/workspaces/create-workspace)
- [List workspaces](/api/workspaces/list-workspaces)
- [Error codes](/api/start-guide/error-codes)
