---
title: Toggle Record Status
description: Mark a record complete or incomplete with the updateTodoDoneStatus mutation, either by toggling or by setting an explicit state.
icon: ToggleLeft
order: 3
---

Use the `updateTodoDoneStatus` mutation to change a record's completion status. Records are `Record` objects in the API, and their completion state is the `done: Boolean!` field.

The mutation works in two modes:

- **Toggle** — omit the `done` argument and the record flips to the opposite of its current state (incomplete → complete, complete → incomplete).
- **Explicit set** — pass `done: true` or `done: false` to set the state directly, regardless of the current value. This is idempotent: setting `done: true` on an already-complete record leaves it complete.

To change the status of many records at once, use [`updateRecords`](#related) instead.

## Request

Toggle a single record's completion status:

```graphql
mutation ToggleRecordStatus {
  updateTodoDoneStatus(todoId: "todo_123") {
    id
    title
    done
    completedAt
    updatedAt
  }
}
```

## Parameters

| Parameter | Type      | Required | Description                                                                                                                                                        |
| --------- | --------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `todoId`  | `String!` | Yes      | The ID of the record to update.                                                                                                                                    |
| `done`    | `Boolean` | No       | If provided, sets the completion status explicitly (`true` = complete, `false` = incomplete). If omitted, the record toggles to the opposite of its current state. |

## Response

Returns the updated `Record`. When the record becomes complete, `done` is `true` and `completedAt` carries the completion timestamp; when it becomes incomplete, `done` is `false`.

```json
{
  "data": {
    "updateTodoDoneStatus": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Draft Q3 launch plan",
      "done": true,
      "completedAt": "2026-05-29T14:21:08.000Z",
      "updatedAt": "2026-05-29T14:21:08.000Z"
    }
  }
}
```

### Returns

| Field         | Type        | Description                                                          |
| ------------- | ----------- | -------------------------------------------------------------------- |
| `id`          | `ID!`       | The record's unique identifier.                                      |
| `title`       | `String!`   | The record title.                                                    |
| `done`        | `Boolean!`  | The new completion status.                                           |
| `completedAt` | `DateTime`  | When the record was completed; `null` when the record is incomplete. |
| `updatedAt`   | `DateTime!` | When the record was last updated.                                    |

The full set of [`Record` fields](/api/records/list-records) is available in the selection set.

## Full example

Set an explicit state rather than toggling. This always marks the record complete, even if it already was:

```graphql
mutation MarkRecordComplete {
  updateTodoDoneStatus(todoId: "todo_123", done: true) {
    id
    done
    completedAt
  }
}
```

Toggling and side effects are driven by the resulting state: completing a record creates a `MARK_AS_COMPLETE` activity entry and fires `TODO_MARKED_AS_COMPLETE` automations; marking it incomplete creates `MARK_AS_INCOMPLETE` and fires `TODO_MARKED_AS_INCOMPLETE`. Time-tracking custom fields and connected webhooks are updated either way.

## Errors

| Code             | When                                                                         |
| ---------------- | ---------------------------------------------------------------------------- |
| `TODO_NOT_FOUND` | No record exists with the given `todoId`.                                    |
| `FORBIDDEN`      | The caller lacks permission to change the record's status (see Permissions). |

## Permissions

| Access level   | Can toggle status |
| -------------- | ----------------- |
| `OWNER`        | Yes               |
| `ADMIN`        | Yes               |
| `MEMBER`       | Yes               |
| `CLIENT`       | Yes               |
| `COMMENT_ONLY` | No                |
| `VIEW_ONLY`    | No                |

The record's workspace must be active (not archived). Custom roles can also block this action: a role with `isRecordsEnabled: false` or `allowMarkRecordsAsDone: false` is denied even at an otherwise-permitted access level. Denied callers receive a `FORBIDDEN` error.

## Related

- [Update a record](/api/records/update-record) — edit title, description, dates, and other fields with `editRecord`, and set custom field values. The `editRecord` mutation does not change completion status; use this mutation for that.
- [List records](/api/records/list-records) — query and filter records, including by completion state.

<Callout variant="info" title="Bulk variant">

To change completion status on many records in one call, use the `updateRecords(input: UpdateRecordsInput!)` mutation, which returns `Boolean!`.

</Callout>
