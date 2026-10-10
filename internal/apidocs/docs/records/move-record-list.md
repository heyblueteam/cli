---
title: Move Record to List
description: Move a record to a different list, within the same workspace or across workspaces, with the moveRecord mutation.
icon: Move
order: 4
---

Use the `moveRecord` mutation to move a record into a different list. The record keeps its ID and carries all of its data with it — assignees, tags, custom field values, checklists, comments, sub-items, due date, and attachments — because those are relationships on the same `Record` row. You can move a record between lists in the same workspace or into a list in another workspace (subject to permissions). Records are `Record` objects and lists are `RecordList` objects in the API.

## Request

Pass the record's ID and the destination list's ID. The mutation returns a `Boolean`.

```graphql
mutation MoveRecord {
  moveRecord(input: { todoId: "todo_123", todoListId: "list_123" })
}
```

To move a record into a list that belongs to a different workspace, point `todoListId` at that workspace's list — the call is identical:

```graphql
mutation MoveRecordCrossWorkspace {
  moveRecord(input: { todoId: "todo_123", todoListId: "list_456" })
}
```

## Move to a database

A database has no single list, so you can also send a destination database instead of a list. The record moves into that database, in the same workspace or another one. Optionally set a value of one of its single-select or multi-select fields in the same call:

```graphql
mutation MoveRecordToDatabase {
  moveRecord(
    input: {
      recordId: "todo_123"
      databaseId: "database_456"
      fieldId: "field_789"
      optionId: "option_012"
    }
  )
}
```

A single-select value replaces the record's value. A multi-select value is added to its values. To change only the value in the record's current database, send its own `databaseId` with `fieldId` and `optionId`; that call records no move.

## Values the destination does not have

When the record moves into another database, values in fields that the destination does not have are carried over. The title, description, due date, assignees, and tags go into the destination's fields of the same kind. Other values go into a destination field with the same name and type. If there is none, the field is added to the destination, and missing select options are added. Lists are not carried. The values also stay in the source fields.

## Parameters

### MoveRecordInput

Send exactly one destination: `todoListId` (or `recordListId`), or `databaseId`.

| Parameter          | Type     | Required | Description                                                                              |
| ------------------ | -------- | -------- | ---------------------------------------------------------------------------------------- |
| `todoId`           | `String` | Yes\*    | ID of the record to move. \*Or `recordId`.                                               |
| `todoListId`       | `String` | No       | ID of the destination list. Or `recordListId`.                                           |
| `databaseId`       | `ID`     | No       | ID of the destination database.                                                          |
| `sourceDatabaseId` | `ID`     | No       | The database to move from, when the record is in several databases of the workspace.     |
| `fieldId`          | `ID`     | No       | A single-select or multi-select field of the destination database. Send with `optionId`. |
| `optionId`         | `ID`     | No       | The option of `fieldId` to set.                                                          |

## Response

The mutation returns `true` on success. Failures throw an error rather than returning `false`.

```json
{
  "data": {
    "moveRecord": true
  }
}
```

### Returns

| Field        | Type       | Description                       |
| ------------ | ---------- | --------------------------------- |
| `moveRecord` | `Boolean!` | `true` when the record was moved. |

## Placement

The record lands at the top of the destination list — next to the greatest `rank` there. You cannot choose the spot on this mutation.

## Side effects

Moving a record is a single relocation, not a copy-and-delete — the `Record` keeps its ID and its history. The move also:

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
