---
title: Archive Workspace
description: Archive and unarchive workspaces to hide them from active lists without deleting their data.
icon: Archive
order: 4
---

Archiving a workspace hides it from your active workspace lists without deleting any of its data. Use the `archiveProject` mutation to archive a workspace and `unarchiveProject` to restore it. Archiving is the reversible alternative to [deleting a workspace](/api/workspaces/delete-workspace) — all records, lists, comments, and files are preserved and become visible again the moment you unarchive. Workspaces are `Project` objects in the API.

Both mutations target a single workspace and return `Boolean!` (`true` on success). The target workspace is resolved from the `id` argument, or from the `X-Bloo-Project-ID` request header when `id` is omitted.

## Request

Archive a workspace by passing its ID:

```graphql
mutation ArchiveWorkspace {
  archiveProject(id: "project_123")
}
```

Restore it later with `unarchiveProject`:

```graphql
mutation UnarchiveWorkspace {
  unarchiveProject(id: "project_123")
}
```

You can omit `id` and let the workspace be resolved from the `X-Bloo-Project-ID` header instead. This is the only operation set in the reference that documents header-based workspace resolution — it lets you reuse a single request configuration across calls without threading the ID into every mutation:

```graphql
# Sent with header: X-Bloo-Project-ID: project_123
mutation ArchiveCurrentWorkspace {
  archiveProject
}
```

If both the `id` argument and the header are present, the `id` argument takes precedence. The `id` accepts either a workspace ID or its slug.

## Parameters

### archiveProject

| Parameter | Type     | Required | Description                                                                                                          |
| --------- | -------- | -------- | -------------------------------------------------------------------------------------------------------------------- |
| `id`      | `String` | No       | ID or slug of the workspace to archive. When omitted, the workspace is resolved from the `X-Bloo-Project-ID` header. |

### unarchiveProject

| Parameter | Type     | Required | Description                                                                                                            |
| --------- | -------- | -------- | ---------------------------------------------------------------------------------------------------------------------- |
| `id`      | `String` | No       | ID or slug of the workspace to unarchive. When omitted, the workspace is resolved from the `X-Bloo-Project-ID` header. |

## Response

Both mutations return `Boolean!` — `true` when the operation succeeds.

```json
{
  "data": {
    "archiveProject": true
  }
}
```

### Returns

| Field                                 | Type       | Description                                          |
| ------------------------------------- | ---------- | ---------------------------------------------------- |
| `archiveProject` / `unarchiveProject` | `Boolean!` | `true` when the workspace is archived or unarchived. |

## What archiving does

When you archive a workspace, the API:

1. Marks the workspace as archived and hides it from active workspace lists.
2. Clears template status — if the workspace was a [template](/api/workspaces/templates), `isTemplate` is set to `false`.
3. Moves the workspace to the end of every member's workspace ordering.
4. Removes it from any workspace folders it was filed under.
5. Records the archive in the [workspace activity log](/api/workspaces/workspace-activity).
6. Notifies connected clients in real time so the workspace disappears from open sessions.

`unarchiveProject` reverses the archived flag and re-positions the workspace at the end of the ordering. It does not restore template status or folder membership — set those again explicitly after unarchiving if you need them.

## Errors

| Code                | When                                                                                                                                                                                           |
| ------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `PROJECT_NOT_FOUND` | No workspace matches the resolved `id` (or no `id` was provided and no `X-Bloo-Project-ID` header was sent).                                                                                   |
| `FORBIDDEN`         | The caller is not an `OWNER` or `ADMIN` of the workspace, or the precondition is not met — `archiveProject` requires an active workspace, `unarchiveProject` requires an already-archived one. |

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

<Callout variant="warning" title="Preconditions are enforced as authorization">

`archiveProject` only accepts a workspace that is currently active, and `unarchiveProject` only accepts one that is currently archived. Calling either on a workspace in the wrong state fails the authorization check and returns `FORBIDDEN` — archiving is not idempotent, so archiving an already-archived workspace errors rather than silently returning `true`.

</Callout>

## Permissions

| Workspace role | Archive / Unarchive |
| -------------- | ------------------- |
| `OWNER`        | Yes                 |
| `ADMIN`        | Yes                 |
| `MEMBER`       | No                  |
| `CLIENT`       | No                  |
| `COMMENT_ONLY` | No                  |
| `VIEW_ONLY`    | No                  |

The caller must also hold an `OWNER`, `ADMIN`, or `MEMBER` role at the organization level.

## Related

- [Delete a workspace](/api/workspaces/delete-workspace) — the permanent, non-reversible alternative.
- [List workspaces](/api/workspaces/list-workspaces) — filter active versus archived workspaces.
- [Workspace activity](/api/workspaces/workspace-activity) — read the archive/unarchive entries.
- [Workspace templates](/api/workspaces/templates) — how template status interacts with archiving.
