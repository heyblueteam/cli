---
title: Recurring Records
description: Attach a repeating schedule to a record so Blue automatically generates copies on a defined cadence.
icon: Repeat
order: 10
---

Use the `createRepeatingTodo`, `updateRepeatingTodo`, and `deleteRepeatingTodo` mutations to manage a repeating schedule on a record. The schedule is attached to an existing record, which acts as the template: each time the schedule fires, Blue copies the template into a target list. Records are `Todo` objects in the API; lists are `TodoList` objects.

You control the cadence with a preset (`DAILY`, `WEEKLY`, `MONTHLY`, …) or a fully custom interval, choose when the schedule ends, and pick which elements (assignees, tags, custom fields, and more) carry over to each copy. All three mutations return `Boolean`.

Send these headers with every request. Company and project accept either an ID or a slug, and header names are case-insensitive.

```http
X-Bloo-Token-ID: YOUR_TOKEN_ID
X-Bloo-Token-Secret: YOUR_TOKEN_SECRET
X-Bloo-Company-ID: YOUR_COMPANY_ID
X-Bloo-Project-ID: project_123
```

## Request

Attach a daily repeating schedule to an existing record, copying assignees and tags to each new occurrence:

```graphql
mutation CreateRecurringRecord {
  createRepeatingTodo(
    input: {
      todoId: "todo_123"
      todoListId: "list_123"
      type: DAILY
      fields: [ASSIGNEES, TAGS]
      from: "2026-06-01T09:00:00Z"
    }
  )
}
```

## Parameters

### CreateRepeatingTodoInput

| Parameter    | Type                           | Required | Description                                                    |
| ------------ | ------------------------------ | -------- | -------------------------------------------------------------- |
| `todoId`     | `String!`                      | Yes      | The existing record to turn into a recurring template.         |
| `todoListId` | `String!`                      | Yes      | The list where each new copy is created.                       |
| `type`       | `RepeatingTodoRepeatType!`     | Yes      | The repeat cadence. Use a preset, or `CUSTOM` with `interval`. |
| `fields`     | `[RepeatingTodoAllowedField]!` | Yes      | Which elements to copy to each occurrence.                     |
| `from`       | `DateTime!`                    | Yes      | The first occurrence date/time the schedule starts from.       |
| `interval`   | `RepeatingTodoIntervalInput`   | No       | Custom interval. Required when `type` is `CUSTOM`.             |
| `end`        | `RepeatingTodoEndInput`        | No       | When the schedule stops. Omit to repeat indefinitely.          |

### UpdateRepeatingTodoInput

Same shape as `CreateRepeatingTodoInput`, plus `repeatCounts`.

| Parameter      | Type                           | Required | Description                                        |
| -------------- | ------------------------------ | -------- | -------------------------------------------------- |
| `todoId`       | `String!`                      | Yes      | The record whose schedule is being updated.        |
| `todoListId`   | `String!`                      | Yes      | The list where each new copy is created.           |
| `type`         | `RepeatingTodoRepeatType!`     | Yes      | The repeat cadence.                                |
| `fields`       | `[RepeatingTodoAllowedField]!` | Yes      | Which elements to copy to each occurrence.         |
| `from`         | `DateTime!`                    | Yes      | The occurrence date/time the schedule runs from.   |
| `interval`     | `RepeatingTodoIntervalInput`   | No       | Custom interval. Required when `type` is `CUSTOM`. |
| `end`          | `RepeatingTodoEndInput`        | No       | When the schedule stops.                           |
| `repeatCounts` | `Int`                          | No       | How many times the record has already repeated.    |

### RepeatingTodoIntervalInput

| Parameter | Type                         | Required | Description                                                                               |
| --------- | ---------------------------- | -------- | ----------------------------------------------------------------------------------------- |
| `count`   | `Int!`                       | Yes      | How many units between occurrences (e.g. `2` for "every 2 weeks").                        |
| `type`    | `RepeatingTodoIntervalType!` | Yes      | The unit of time: `DAYS`, `WEEKS`, `MONTHS`, or `YEARS`.                                  |
| `days`    | `[RepeatingTodoDayType]`     | No       | Days of the week for `WEEKS` intervals (`Sun`, `Mon`, `Tue`, `Wed`, `Thu`, `Fri`, `Sat`). |
| `month`   | `RepeatingTodoMonthType`     | No       | How to anchor `MONTHS` intervals: `BY_DD` or `BY_DDDD`.                                   |

### RepeatingTodoEndInput

| Parameter | Type                    | Required | Description                                             |
| --------- | ----------------------- | -------- | ------------------------------------------------------- |
| `type`    | `RepeatingTodoEndType!` | Yes      | How the schedule ends: `NEVER`, `ON`, or `AFTER`.       |
| `on`      | `DateTime`              | No       | End date. Required when `type` is `ON`.                 |
| `after`   | `Int`                   | No       | Number of occurrences. Required when `type` is `AFTER`. |

### RepeatingTodoRepeatType

| Value      | Description                              |
| ---------- | ---------------------------------------- |
| `DAILY`    | Repeats every day.                       |
| `WEEKDAYS` | Repeats Monday through Friday.           |
| `WEEKLY`   | Repeats every week on the same day.      |
| `MONTHLY`  | Repeats every month on the same date.    |
| `YEARLY`   | Repeats every year on the same date.     |
| `CUSTOM`   | Cadence defined by the `interval` field. |

### RepeatingTodoAllowedField

| Value           | Description                                 |
| --------------- | ------------------------------------------- |
| `ASSIGNEES`     | Copy assigned users to the new record.      |
| `TAGS`          | Copy tags to the new record.                |
| `CUSTOM_FIELDS` | Copy custom field values to the new record. |
| `DESCRIPTION`   | Copy the description to the new record.     |
| `CHECKLISTS`    | Copy checklists to the new record.          |
| `COMMENTS`      | Copy comments to the new record.            |

### RepeatingTodoIntervalType

| Value    | Description                  |
| -------- | ---------------------------- |
| `DAYS`   | Interval measured in days.   |
| `WEEKS`  | Interval measured in weeks.  |
| `MONTHS` | Interval measured in months. |
| `YEARS`  | Interval measured in years.  |

### RepeatingTodoMonthType

| Value     | Description                                                |
| --------- | ---------------------------------------------------------- |
| `BY_DD`   | Repeat on the same date of the month (e.g. the 15th).      |
| `BY_DDDD` | Repeat on the same weekday position (e.g. the 2nd Monday). |

### RepeatingTodoEndType

| Value   | Description                                       |
| ------- | ------------------------------------------------- |
| `NEVER` | Repeats indefinitely.                             |
| `ON`    | Ends on a specific date (set `on`).               |
| `AFTER` | Ends after a number of occurrences (set `after`). |

## Response

Each mutation returns a `Boolean` under the operation name.

```json
{
  "data": {
    "createRepeatingTodo": true
  }
}
```

### Returns

| Field                 | Type      | Description                                                                                                         |
| --------------------- | --------- | ------------------------------------------------------------------------------------------------------------------- |
| `createRepeatingTodo` | `Boolean` | `true` when the schedule is attached.                                                                               |
| `updateRepeatingTodo` | `Boolean` | `true` when the schedule is updated and the next copy is created; `false` if creating the copy failed (see Errors). |
| `deleteRepeatingTodo` | `Boolean` | `true` when the schedule is removed.                                                                                |

To read the active schedule's target list back, select `repeatingTodoList` on the record. It is `null` when no schedule is set.

```graphql
query RecurringTarget {
  todoQueries {
    todos(filter: { companyIds: ["company_123"], todoIds: ["todo_123"] }) {
      items {
        id
        title
        repeatingTodoList {
          id
          title
        }
      }
    }
  }
}
```

## Full example

Repeat every 2 weeks on Monday and Wednesday, end after 10 occurrences, and copy every supported element:

```graphql
mutation CreateRecurringRecordCustom {
  createRepeatingTodo(
    input: {
      todoId: "todo_123"
      todoListId: "list_123"
      type: CUSTOM
      fields: [ASSIGNEES, TAGS, CUSTOM_FIELDS, DESCRIPTION, CHECKLISTS, COMMENTS]
      from: "2026-06-01T09:00:00Z"
      interval: { count: 2, type: WEEKS, days: [Mon, Wed] }
      end: { type: AFTER, after: 10 }
    }
  )
}
```

Update a record's schedule to repeat monthly by date, ending on a fixed date:

```graphql
mutation UpdateRecurringRecord {
  updateRepeatingTodo(
    input: {
      todoId: "todo_123"
      todoListId: "list_123"
      type: CUSTOM
      fields: [ASSIGNEES, TAGS, DESCRIPTION]
      from: "2026-07-01T09:00:00Z"
      interval: { count: 1, type: MONTHS, month: BY_DD }
      end: { type: ON, on: "2026-12-31T23:59:59Z" }
      repeatCounts: 5
    }
  )
}
```

Remove a schedule from a record (the `id` is the record's `todoId`):

```graphql
mutation DeleteRecurringRecord {
  deleteRepeatingTodo(id: "todo_123")
}
```

## How it works

- The schedule lives on the template record; the template itself is never altered when a copy is generated.
- Each occurrence is a fresh copy created in `todoListId` via the same mechanism as [copying a record](/api/records/copy-record). The target list can be in the same workspace or a different one.
- `fields` controls exactly what carries over. Anything not listed is left off the copy.
- Preset cadences (`DAILY`, `WEEKDAYS`, `WEEKLY`, `MONTHLY`, `YEARLY`) need no `interval`. `CUSTOM` requires one — use `interval.days` to pin specific weekdays on `WEEKS` intervals, and `interval.month` to anchor `MONTHS` intervals to a date (`BY_DD`) or a weekday position (`BY_DDDD`).
- When `updateRepeatingTodo` generates a copy, it logs a `REPEAT_TODO` activity, records a todo action, notifies the copy's assignees, and publishes the new record in real time. If creating the copy fails, the mutation clears the schedule to avoid a stuck state and returns `false` instead of throwing — always check the return value.

## Errors

| Code              | When                                                                                 |
| ----------------- | ------------------------------------------------------------------------------------ |
| `TODO_NOT_FOUND`  | `todoId` (or `id` on delete) does not exist, or the caller cannot access the record. |
| `FORBIDDEN`       | The caller lacks edit-level access to the record (see Permissions).                  |
| `UNAUTHENTICATED` | The request carries no valid token.                                                  |

## Permissions

Managing a recurring schedule requires edit-level access to the record. Only `OWNER`, `ADMIN`, and `MEMBER` access levels qualify; `CLIENT`, `COMMENT_ONLY`, and `VIEW_ONLY` cannot. A custom role with records disabled (`isRecordsEnabled: false`) is also denied, and the record's workspace must be active (not archived).

## Related

- [Copy a record](/api/records/copy-record)
- [Update a record](/api/records/update-record)
- [List records](/api/records/list-records)
- [Assignees](/api/records/assignees)
- [Tags](/api/records/tags)
