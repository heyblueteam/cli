---
title: Archive a Record
description: Archive and restore records to hide them from active lists without deleting their data.
icon: Archive
order: 12
---

Archiving a record takes it out of its list without deleting it. The data — title, description, dates, assignees, tags, comments, checklists, and custom-field values — is preserved in full and reappears the moment you restore the record. Use the `archiveTodo` mutation to archive a record and `unarchiveTodo` to restore it. Records are `Todo` objects in the API.

Archiving is the reversible alternative to [deleting a record](/api/records#operations), which is permanent. It is also distinct from completion: a [completed](/api/records/toggle-record-status) record (`done: true`) still lives in its list, whereas an archived record is hidden from it.

Both mutations take the record `id` and return the updated `Todo`.

## Request

Archive a record by passing its ID:

```graphql
mutation ArchiveRecord {
  archiveTodo(id: "todo_123") {
    id
    archived
    archivedAt
  }
}
```

Restore it later with `unarchiveTodo`:

```graphql
mutation RestoreRecord {
  unarchiveTodo(id: "todo_123") {
    id
    archived
    todoList {
      id
      title
    }
  }
}
```

## Parameters

### archiveTodo

| Parameter | Type      | Required | Description                  |
| --------- | --------- | -------- | ---------------------------- |
| `id`      | `String!` | Yes      | ID of the record to archive. |

### unarchiveTodo

| Parameter | Type      | Required | Description                  |
| --------- | --------- | -------- | ---------------------------- |
| `id`      | `String!` | Yes      | ID of the record to restore. |

## Response

Both mutations return the updated `Todo`. After `archiveTodo`, `archived` is `true` and `archivedAt` carries the archive timestamp; after `unarchiveTodo`, `archived` is `false` and `archivedAt` is `null`.

```json
{
  "data": {
    "archiveTodo": {
      "id": "todo_123",
      "archived": true,
      "archivedAt": "2026-06-05T14:21:00.000Z"
    }
  }
}
```

### Returns

| Field        | Type        | Description                                                                             |
| ------------ | ----------- | --------------------------------------------------------------------------------------- |
| `archived`   | `Boolean!`  | `true` after archiving, `false` after restoring.                                        |
| `archivedAt` | `DateTime`  | When the record was archived. `null` once restored.                                     |
| `todoList`   | `TodoList!` | The record's list. Unchanged by archiving — a restored record returns to the same list. |

See [List records](/api/records/list-records#todo) for the full set of selectable `Todo` fields.

## What archiving does

When you archive a record, the API:

1. Marks it archived and records the timestamp and the acting user. The record drops out of its list's active view and out of [list-records](/api/records/list-records) results unless you explicitly query the archive (see [Archive scope](/api/records/list-records#archive-scope)).
2. Notifies connected clients in real time so the card disappears from open sessions.
3. Excludes the record from charts and reports, whose counts recalculate accordingly.

The record's list and position are left untouched. `unarchiveTodo` clears the archived flag and timestamp, and the record snaps back to exactly the spot it occupied before — there is no need to move or re-position it.

<Callout variant="info" title="Both mutations are idempotent">

Re-archiving an already-archived record (or restoring an already-active one) is a no-op: the mutation returns the current record unchanged rather than erroring, and the original `archivedAt` / archived-by audit data is preserved. This is unlike [archiving a workspace](/api/workspaces/archive-workspace), where calling the mutation in the wrong state returns `FORBIDDEN`.

</Callout>

## Errors

| Code             | When                                                                      |
| ---------------- | ------------------------------------------------------------------------- |
| `TODO_NOT_FOUND` | No record exists with the given `id`, or the caller can't access it.      |
| `FORBIDDEN`      | The caller lacks permission to archive or restore the record (see below). |

```json
{
  "errors": [
    {
      "message": "Todo was not found.",
      "extensions": { "code": "TODO_NOT_FOUND" }
    }
  ]
}
```

## Permissions

| Access level   | Archive / Restore |
| -------------- | ----------------- |
| `OWNER`        | Yes               |
| `ADMIN`        | Yes               |
| `MEMBER`       | Yes               |
| `CLIENT`       | No                |
| `COMMENT_ONLY` | No                |
| `VIEW_ONLY`    | No                |

The record's workspace must be active (not archived). Custom roles can further restrict these actions independently: a role with `canArchiveRecords: false` is denied archiving, and one with `canUnarchiveRecords: false` is denied restoring, even at an otherwise-permitted access level. Restoring deliberately does **not** require the view-archived permission — a role that can reach an archived record by direct link or reference can restore it without being able to browse the archive.

## Related

- [List records](/api/records/list-records) — query the archive with the `archived` filter; see [Archive scope](/api/records/list-records#archive-scope).
- [Toggle completion](/api/records/toggle-record-status) — mark a record done without removing it from its list.
- [Archive a workspace](/api/workspaces/archive-workspace) — the workspace-level equivalent.
- [Records overview](/api/records) — the full record operation surface.
