---
title: Move Record to List
description: Move a record to a different list, within the same workspace or across workspaces, with the moveTodo mutation.
icon: Move
order: 4
---

Use the `moveTodo` mutation to move a record into a different list. The record keeps its ID and carries all of its data with it — assignees, tags, custom field values, checklists, comments, sub-items, due date, and attachments — because those are relationships on the same `Todo` row. You can move a record between lists in the same workspace or into a list in another workspace (subject to permissions). Records are `Todo` objects and lists are `TodoList` objects in the API.

## Request

Pass the record's ID and the destination list's ID. The mutation returns a `Boolean`.

```graphql
mutation MoveRecord {
  moveTodo(input: { todoId: "todo_123", todoListId: "list_123" })
}
```

To move a record into a list that belongs to a different workspace, point `todoListId` at that workspace's list — the call is identical:

```graphql
mutation MoveRecordCrossWorkspace {
  moveTodo(input: { todoId: "todo_123", todoListId: "list_456" })
}
```

## Parameters

### MoveTodoInput

| Parameter    | Type      | Required | Description                 |
| ------------ | --------- | -------- | --------------------------- |
| `todoId`     | `String!` | Yes      | ID of the record to move.   |
| `todoListId` | `String!` | Yes      | ID of the destination list. |

## Response

The mutation returns `true` on success. Failures throw an error rather than returning `false`.

```json
{
  "data": {
    "moveTodo": true
  }
}
```

### Returns

| Field      | Type       | Description                       |
| ---------- | ---------- | --------------------------------- |
| `moveTodo` | `Boolean!` | `true` when the record was moved. |

## Position

The record is placed at the end of the destination list. You cannot specify a position — it is computed as the highest existing position in that list plus `65535` (the standard gap Blue leaves between records so later reorders don't need to renumber the list).

## Side effects

Moving a record is a single relocation, not a copy-and-delete — the `Todo` keeps its ID and its history. The move also:

- Records a `MOVE_TODO` action in the record's activity, capturing the source and destination workspace names.
- Notifies relevant members of the destination workspace.
- Publishes a real-time update to connected clients. A same-workspace move emits a single `UPDATED` event; a cross-workspace move emits `CREATED` in the destination and `DELETED` in the source, since the two have different subscribers.
- Marks the affected workspace charts for recalculation (both workspaces on a cross-workspace move).
- Runs automations triggered by a record arriving from another workspace, on cross-workspace moves only.

## Errors

| Code                  | When                                                                   |
| --------------------- | ---------------------------------------------------------------------- |
| `TODO_NOT_FOUND`      | No record matches `todoId`, or it isn't in a workspace you can access. |
| `TODO_LIST_NOT_FOUND` | No list matches `todoListId`.                                          |
| `FORBIDDEN`           | You lack permission to move this record (see Permissions).             |

## Permissions

You need `OWNER`, `ADMIN`, or `MEMBER` access to the source workspace, and the workspace must not be archived. `VIEW_ONLY` and `COMMENT_ONLY` roles, and roles whose custom permission disables record editing, are denied with `FORBIDDEN`. For a cross-workspace move you must also have access to the destination workspace.

## Related

- [Copy a record](/api/records/copy-record)
- [Update a record](/api/records/update-record)
- [Workspace lists](/api/workspaces/lists)
- [Records overview](/api/records)
