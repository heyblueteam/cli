---
title: Query record time tracking
description: Read a record's derived time-tracking data — total time in seconds per workspace and list, plus time to completion.
icon: Timer
order: 11
---

Blue tracks how long each record spends in every list automatically. Read this data through the `timeTracking` field on a record (records are `Record` objects in the API). There is no dedicated query and no mutation — select `timeTracking` inside any of the existing record queries (`todo`, `recordQueries.todos`, or `todoList`) and the values are computed on read.

Every value is a duration in **seconds** (a `Float`). Time tracking is read-only: it is derived from the record's list-movement history, not from a manual timer you start and stop.

## Request

Select `timeTracking` on a single record with the `todo(id:)` query:

```graphql
query RecordTimeTracking {
  todo(id: "todo_123") {
    id
    title
    timeTracking {
      timeInCurrentProject
      timeInCurrentTodoList
      timeToCompletion
      timeInLists {
        todoListId
        todoListTitle
        totalTime
        percentage
      }
    }
  }
}
```

The same selection works inside `recordQueries.todos` to read time tracking for many records at once — see [List records](/api/records/list-records).

## Parameters

`timeTracking` takes no arguments. It is a nested field selected on a `Record`; the record is identified by the enclosing query (`todo(id:)`, `recordQueries.todos(filter:)`, or a `todoList` selection).

## Response

```json
{
  "data": {
    "todo": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Ship onboarding revamp",
      "timeTracking": {
        "timeInCurrentProject": 432000,
        "timeInCurrentTodoList": 86400,
        "timeToCompletion": 604800,
        "timeInLists": [
          {
            "todoListId": "clm4n8qwx000108l0a1b2c3d4",
            "todoListTitle": "Backlog",
            "totalTime": 345600,
            "percentage": 57.14
          },
          {
            "todoListId": "clm4n8qwx000208l0e5f6g7h8",
            "todoListTitle": "In progress",
            "totalTime": 86400,
            "percentage": 14.29
          }
        ]
      }
    }
  }
}
```

When the record has no list history yet, the numeric fields resolve to `0` and `timeInLists` is an empty array.

### TimeTracking

| Field                   | Type            | Description                                                                                                                  |
| ----------------------- | --------------- | ---------------------------------------------------------------------------------------------------------------------------- |
| `timeInCurrentProject`  | `Float`         | Total seconds the record has spent across all lists in its **current workspace**.                                            |
| `timeInCurrentTodoList` | `Float`         | Total seconds the record has spent in the list it is **currently in**.                                                       |
| `timeToCompletion`      | `Float`         | Total seconds from when the record was created to when it was completed. Continues to accrue against "now" while still open. |
| `timeInLists`           | `[TimeInList!]` | Per-list breakdown of time spent, one entry per list the record has passed through in the current workspace.                 |

### TimeInList

| Field           | Type     | Description                                                           |
| --------------- | -------- | --------------------------------------------------------------------- |
| `todoListId`    | `String` | ID of the list. `null` if the list has since been deleted.            |
| `todoListTitle` | `String` | Title of the list at the time the record was in it.                   |
| `totalTime`     | `Float`  | Total seconds the record has spent in this list.                      |
| `percentage`    | `Float`  | This list's share of `timeInCurrentProject`, as a percentage (0–100). |

<Callout variant="info" title="Where the numbers come from">

Each time a record is created in or moved between lists, Blue writes a list-history entry. `timeTracking` sums the open and closed intervals from that history. A record still sitting in a list accrues time against the current moment, so repeated reads of an open record return increasing values. Time tracking is scoped to the record's **current workspace** — moving a record to a different workspace resets the `timeInCurrentProject` and `timeInLists` view to that workspace's history.

</Callout>

## Filter records by time in list

You can find records by how long they have spent in a list using a `TIME_IN_LIST` filter entry. This works in the `fields` array of [`recordQueries.todos`](/api/records/list-records) and in `todosCount` (via `TodoFilterInput`), routed by the `TIME_IN_LIST` value of `TodoFilterFieldType`. The list and matching options come from `TimeInListFilterInput`.

```graphql
query RecordsSittingTooLong {
  recordQueries {
    todos(
      filter: {
        companyIds: ["company_123"]
        projectIds: ["project_123"]
        fields: [
          {
            type: "TIME_IN_LIST"
            op: "GTE"
            values: 604800
            timeInList: { todoListId: "list_123" }
          }
        ]
      }
    ) {
      items {
        id
        title
      }
      pageInfo {
        totalItems
      }
    }
  }
}
```

This returns records that have spent **at least** 604,800 seconds (7 days) in list `list_123`.

### TimeInListFilterInput

| Field           | Type      | Description                                                                                                                          |
| --------------- | --------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `todoListId`    | `String`  | Match records by this list's ID. Supply this or `todoListTitle`.                                                                     |
| `todoListTitle` | `String`  | Match records by list title — useful for lists that have since been deleted, which keep their title in history but lose their ID.    |
| `cumulative`    | `Boolean` | When `true`, sum every interval the record has spent in the list. When `false` (the default), measure only the **most recent** stay. |

### The enclosing filter entry

The entry sits inside the filter's `fields` array (typed `JSON`); supply these keys alongside `timeInList`:

| Key          | Type     | Description                                                                                         |
| ------------ | -------- | --------------------------------------------------------------------------------------------------- |
| `type`       | `String` | Must be `"TIME_IN_LIST"`. Entries without a `type` are silently dropped.                            |
| `op`         | `String` | The comparison. Only `"GTE"` (at least) and `"LTE"` (at most) are applied to time-in-list duration. |
| `values`     | `Number` | The threshold **in seconds**. Negative values are ignored.                                          |
| `timeInList` | object   | A `TimeInListFilterInput` (above) naming the list and the cumulative/most-recent mode.              |

## Errors

| Code             | When                                                                 |
| ---------------- | -------------------------------------------------------------------- |
| `TODO_NOT_FOUND` | No record exists with the given `id` (when reading via `todo(id:)`). |
| `FORBIDDEN`      | The caller cannot access the record or its workspace.                |

`timeTracking` itself does not raise — if the time-tracking field is disabled or not viewable, it resolves to `null` rather than erroring (see Permissions).

## Permissions

`timeTracking` resolves to `null` unless **both** conditions hold:

- The record's workspace has the **Time Tracking** system field enabled (the `TIME_TRACKING` value of `TodoFieldType`).
- The field is viewable by the caller. A custom role that hides this field, or an access level without view rights to the record, sees `null`.

A `null` result is not an error — select the field and check for `null` in your response handling.

## Related

- [List records](/api/records/list-records) — read `timeTracking` for many records at once and filter by time in list.
- [Move a record between lists](/api/records/move-record-list) — every move writes the list-history entry that time tracking is computed from.
- [Toggle record status](/api/records/toggle-record-status) — completing a record fixes its `timeToCompletion`.
- [Duration field](/api/custom-fields/time-duration) — the separate `TIME_DURATION` custom field, for storing a duration value you set yourself (distinct from this automatic time tracking).
- [Records overview](/api/records) — the full `Record` field surface and the record query namespaces.
