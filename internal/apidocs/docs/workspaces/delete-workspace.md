---
title: Delete a Workspace
description: Permanently delete a workspace and all of its records, lists, comments, and files using the deleteProject mutation.
icon: Trash2
order: 6
---

Use the `deleteProject` mutation to permanently delete a workspace and everything inside it — records, lists, comments, custom fields, automations, tags, and file attachments. Deletion is irreversible through the API. If you might need the workspace later, [archive it](/api/workspaces/archive-workspace) instead. Workspaces are `Project` objects in the API.

<Callout variant="danger" title="Deletion is permanent">

`deleteProject` cannot be undone through the API. The workspace and all of its records, lists, comments, custom field values, automations, tags, dependencies, and file attachments are removed. Archive the workspace first if there is any chance you will need the data again.

</Callout>

## Request

```graphql
mutation DeleteWorkspace {
  deleteProject(id: "project_123") {
    success
    operationId
  }
}
```

The `id` argument accepts either the workspace ID or its slug.

## Parameters

| Parameter | Type      | Required | Description                                |
| --------- | --------- | -------- | ------------------------------------------ |
| `id`      | `String!` | Yes      | The ID or slug of the workspace to delete. |

## Response

The workspace record is removed immediately and `success: true` is returned. The associated child data (records, lists, comments, and files) is removed by an asynchronous background cleanup job identified by `operationId`.

```json
{
  "data": {
    "deleteProject": {
      "success": true,
      "operationId": "clm4n8qwx000008l0g4oxdqn7"
    }
  }
}
```

### Returns

`deleteProject` returns a `MutationResult`.

| Field         | Type       | Description                                                                                       |
| ------------- | ---------- | ------------------------------------------------------------------------------------------------- |
| `success`     | `Boolean!` | `true` when the workspace was deleted.                                                            |
| `operationId` | `String`   | Identifier for the background cleanup job that removes the workspace's child data. May be `null`. |

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

## Related

- [Archive a workspace](/api/workspaces/archive-workspace) — the reversible alternative.
- [Create a workspace](/api/workspaces/create-workspace)
- [List workspaces](/api/workspaces/list-workspaces)
- [Error codes](/api/start-guide/error-codes)
