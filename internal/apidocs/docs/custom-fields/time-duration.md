---
title: Time Duration Custom Field
description: Auto-calculate the elapsed time between two events on a record — cycle times, response times, SLAs.
icon: Clock
order: 27
---

A time duration custom field measures the elapsed time between two events in a record's lifecycle — for example, from creation to completion, or from one status change to another. The value is computed automatically by Blue; you never write it directly. It maps to the `TIME_DURATION` value of the `CustomFieldType` enum and is backed by a `CustomField` whose stored duration (in seconds) is exposed on `CustomField.number`.

Records are `Todo` objects in the API.

## Overview

- **Type enum:** `TIME_DURATION`
- **Configuration:** a start event (`timeDurationStartInput`) and an end event (`timeDurationEndInput`), each a `CustomFieldTimeDurationInput`; a display format (`timeDurationDisplay`); and an optional SLA target (`timeDurationTargetTime`, seconds).
- **Value shape:** the elapsed time in **seconds**, read from `CustomField.number` (a `Float`) in a record context. `CustomField.value` is JSON and is not the duration — use `number`.
- **Read-only:** the value is derived from the configured events. `setTodoCustomField` does not apply to this type; there is no write path for the duration.
- **Schema-optional, practically required:** in `CreateCustomFieldInput` the three time-duration inputs are nullable (no `!`), because the input is shared across all field types. For `TIME_DURATION` the resolver rejects the call if any of `timeDurationDisplay`, `timeDurationStartInput`, or `timeDurationEndInput` is missing.

## Create

Use the `createCustomField` mutation with `type: TIME_DURATION`. The field is created in the workspace named by your `X-Bloo-Project-ID` header (ID or slug) — there is no `projectId` argument.

This example tracks how long a record takes to complete, from creation to the moment it's marked done:

```graphql
mutation CreateDurationField {
  createCustomField(
    input: {
      name: "Completion Time"
      type: TIME_DURATION
      timeDurationDisplay: FULL_DATE_SUBSTRING
      timeDurationStartInput: { type: TODO_CREATED_AT, condition: FIRST }
      timeDurationEndInput: { type: TODO_MARKED_AS_COMPLETE, condition: FIRST }
    }
  ) {
    id
    name
    type
    timeDurationDisplay
    timeDurationStart {
      type
      condition
    }
    timeDurationEnd {
      type
      condition
    }
  }
}
```

```json
{
  "data": {
    "createCustomField": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Completion Time",
      "type": "TIME_DURATION",
      "timeDurationDisplay": "FULL_DATE_SUBSTRING",
      "timeDurationStart": { "type": "TODO_CREATED_AT", "condition": "FIRST" },
      "timeDurationEnd": { "type": "TODO_MARKED_AS_COMPLETE", "condition": "FIRST" }
    }
  }
}
```

### CreateCustomFieldInput

| Parameter                | Type                                 | Required | Description                                                                                 |
| ------------------------ | ------------------------------------ | -------- | ------------------------------------------------------------------------------------------- |
| `name`                   | `String!`                            | Yes      | Display name of the field.                                                                  |
| `type`                   | `CustomFieldType!`                   | Yes      | Must be `TIME_DURATION`.                                                                    |
| `timeDurationDisplay`    | `CustomFieldTimeDurationDisplayType` | For type | How the duration is rendered. Nullable in the schema, but required for `TIME_DURATION`.     |
| `timeDurationStartInput` | `CustomFieldTimeDurationInput`       | For type | The event that starts the clock. Nullable in the schema, but required for `TIME_DURATION`.  |
| `timeDurationEndInput`   | `CustomFieldTimeDurationInput`       | For type | The event that stops the clock. Nullable in the schema, but required for `TIME_DURATION`.   |
| `timeDurationTargetTime` | `Float`                              | No       | Optional SLA target in seconds, returned alongside the value for comparison (e.g. `86400`). |
| `description`            | `String`                             | No       | Help text shown to users.                                                                   |

The workspace is taken from the `X-Bloo-Project-ID` header; `CreateCustomFieldInput` has no `projectId` field. The same time-duration inputs also appear on `EditCustomFieldInput` for updating an existing field.

### CustomFieldTimeDurationInput

Each of the two events is configured with this input.

| Parameter              | Type                                | Required    | Description                                                               |
| ---------------------- | ----------------------------------- | ----------- | ------------------------------------------------------------------------- |
| `type`                 | `CustomFieldTimeDurationType!`      | Yes         | Which kind of event marks this end of the interval.                       |
| `condition`            | `CustomFieldTimeDurationCondition!` | Yes         | Whether to use the `FIRST` or `LAST` occurrence of the event.             |
| `customFieldId`        | `String`                            | Conditional | The field to watch. Required when `type` is `TODO_CUSTOM_FIELD`.          |
| `customFieldOptionIds` | `[String!]`                         | Conditional | Specific option(s) to match for a select field; omit to match any change. |
| `todoListId`           | `String`                            | Conditional | The list to watch. Required when `type` is `TODO_MOVED`.                  |
| `tagId`                | `String`                            | Conditional | The tag to watch. Required when `type` is `TODO_TAG_ADDED`.               |
| `assigneeId`           | `String`                            | Conditional | The user to watch. Required when `type` is `TODO_ASSIGNEE_ADDED`.         |

When `customFieldOptionIds` is supplied for a `TODO_CUSTOM_FIELD` event, every ID must belong to the field named by `customFieldId`, or the call throws `CUSTOM_FIELD_OPTION_NOT_FOUND`.

### CustomFieldTimeDurationType

| Value                     | Description                           | Extra input required |
| ------------------------- | ------------------------------------- | -------------------- |
| `TODO_CREATED_AT`         | The record was created.               | —                    |
| `TODO_CUSTOM_FIELD`       | A custom field on the record changed. | `customFieldId`      |
| `TODO_DUE_DATE`           | The record's due date was reached.    | —                    |
| `TODO_MARKED_AS_COMPLETE` | The record was marked complete.       | —                    |
| `TODO_MOVED`              | The record was moved into a list.     | `todoListId`         |
| `TODO_TAG_ADDED`          | A tag was added to the record.        | `tagId`              |
| `TODO_ASSIGNEE_ADDED`     | A user was assigned to the record.    | `assigneeId`         |

### CustomFieldTimeDurationCondition

| Value   | Description                            |
| ------- | -------------------------------------- |
| `FIRST` | Use the first occurrence of the event. |
| `LAST`  | Use the most recent occurrence.        |

### CustomFieldTimeDurationDisplayType

| Value                 | Description                                      |
| --------------------- | ------------------------------------------------ |
| `FULL_DATE`           | Compact, fixed-width numeric format.             |
| `FULL_DATE_STRING`    | Duration spelled out in words.                   |
| `FULL_DATE_SUBSTRING` | Abbreviated human-readable units (e.g. `1h 2m`). |

The display format controls only how clients render the value — the stored duration is always seconds. There are no per-unit display values (no `DAYS`/`HOURS`/`MINUTES`/`SECONDS`); convert from seconds yourself if you need a single unit.

## Read a value

The duration is computed by Blue and read back through the record's `customFields` connection. `Todo.customFields` returns `[CustomField!]!` directly — select the value fields on the element. For a `TIME_DURATION` field, the elapsed seconds are on `CustomField.number`.

```graphql
query ReadDuration {
  todoQueries {
    todos(filter: { companyIds: ["company_123"], todoIds: ["todo_123"] }) {
      items {
        id
        title
        customFields {
          id
          name
          type
          number
          timeDurationDisplay
          timeDurationTargetTime
        }
      }
    }
  }
}
```

```json
{
  "data": {
    "todoQueries": {
      "todos": {
        "items": [
          {
            "id": "clm4n8qwx000008l0g4oxdqn7",
            "title": "Onboard new client",
            "customFields": [
              {
                "id": "clm4n8qwx000108l0a1b2c3d4",
                "name": "Completion Time",
                "type": "TIME_DURATION",
                "number": 93784,
                "timeDurationDisplay": "FULL_DATE_SUBSTRING",
                "timeDurationTargetTime": 86400
              }
            ]
          }
        ]
      }
    }
  }
}
```

`number` resolves only when the `CustomField` is read in a record context (via `Todo.customFields`); it is `null` when the field definition is read on its own, and also `null` when the interval is incomplete (see [Notes](#notes)). Here `93784` seconds is 1 day, 2 hours, 3 minutes, 4 seconds — and it exceeds the `86400`-second SLA target.

### CustomField (TIME_DURATION fields)

| Field                    | Type                                 | Description                                            |
| ------------------------ | ------------------------------------ | ------------------------------------------------------ |
| `number`                 | `Float`                              | The elapsed duration in seconds (record context only). |
| `timeDurationDisplay`    | `CustomFieldTimeDurationDisplayType` | The configured display format.                         |
| `timeDurationStart`      | `CustomFieldTimeDuration`            | The start-event configuration.                         |
| `timeDurationEnd`        | `CustomFieldTimeDuration`            | The end-event configuration.                           |
| `timeDurationTargetTime` | `Float`                              | The configured SLA target in seconds, if any.          |

### CustomFieldTimeDuration

The shape returned for `timeDurationStart` and `timeDurationEnd`.

| Field                | Type                               | Description                                          |
| -------------------- | ---------------------------------- | ---------------------------------------------------- |
| `type`               | `CustomFieldTimeDurationType!`     | The event kind.                                      |
| `condition`          | `CustomFieldTimeDurationCondition` | `FIRST` or `LAST`.                                   |
| `customField`        | `CustomField`                      | The watched field, for a `TODO_CUSTOM_FIELD` event.  |
| `customFieldOptions` | `[CustomFieldOption!]`             | The matched options, if option IDs were supplied.    |
| `todoList`           | `TodoList`                         | The watched list, for a `TODO_MOVED` event.          |
| `tag`                | `Tag`                              | The watched tag, for a `TODO_TAG_ADDED` event.       |
| `assignee`           | `User`                             | The watched user, for a `TODO_ASSIGNEE_ADDED` event. |

## Full example

Track the time a record spends in review by watching two values of a single-select status field, with a 24-hour SLA target:

```graphql
mutation CreateReviewCycleField {
  createCustomField(
    input: {
      name: "Review Cycle Time"
      type: TIME_DURATION
      description: "Time from review request to approval"
      timeDurationDisplay: FULL_DATE_STRING
      timeDurationTargetTime: 86400
      timeDurationStartInput: {
        type: TODO_CUSTOM_FIELD
        condition: FIRST
        customFieldId: "field_123"
        customFieldOptionIds: ["option_123"]
      }
      timeDurationEndInput: {
        type: TODO_CUSTOM_FIELD
        condition: LAST
        customFieldId: "field_123"
        customFieldOptionIds: ["option_456"]
      }
    }
  ) {
    id
    name
    timeDurationTargetTime
    timeDurationStart {
      type
      condition
      customField {
        name
      }
      customFieldOptions {
        title
      }
    }
    timeDurationEnd {
      type
      condition
      customFieldOptions {
        title
      }
    }
  }
}
```

A few more common shapes:

```graphql
# Time spent in a specific list (start when moved in, end when moved out)
timeDurationStartInput: { type: TODO_MOVED, condition: FIRST, todoListId: "list_123" }
timeDurationEndInput:   { type: TODO_MOVED, condition: FIRST, todoListId: "list_456" }

# Time from assignment to first status change
timeDurationStartInput: { type: TODO_ASSIGNEE_ADDED, condition: FIRST, assigneeId: "user_123" }
timeDurationEndInput:   { type: TODO_CUSTOM_FIELD, condition: FIRST, customFieldId: "field_123", customFieldOptionIds: ["option_123"] }
```

## Notes

- **The value is read-only and asynchronous.** Blue recomputes the duration in the background when a configured event fires (record created, custom field changed, tag/assignee added, record moved or marked complete). There is no mutation to set it — `setTodoCustomField` is not used for this type.
- **`CustomField.number` holds the duration, not `CustomField.value`.** `value` is the generic JSON value column; for a time-duration field, read the seconds from `number`.
- **`number` is `null` until the interval completes.** It returns `null` when the start event hasn't occurred, the end event hasn't occurred, or a referenced field/list/tag/user no longer exists.
- **`FIRST` vs `LAST` matters for repeated events.** A status field can be set multiple times; `FIRST` pins the clock to the earliest match and `LAST` to the most recent. Pick deliberately when an event can recur.
- **`timeDurationTargetTime` is data, not enforcement.** It's stored and returned so a client or dashboard can compare it against the elapsed `number`; Blue does not raise alerts on its own.

## Errors

| Code                            | When                                                                                                       |
| ------------------------------- | ---------------------------------------------------------------------------------------------------------- |
| `BAD_USER_INPUT`                | `timeDurationDisplay`, `timeDurationStartInput`, or `timeDurationEndInput` is missing for `TIME_DURATION`. |
| `CUSTOM_FIELD_NOT_FOUND`        | A `TODO_CUSTOM_FIELD` event omits `customFieldId` or references a field not in the workspace.              |
| `CUSTOM_FIELD_OPTION_NOT_FOUND` | A supplied `customFieldOptionIds` entry doesn't belong to the referenced field.                            |
| `TODO_LIST_NOT_FOUND`           | A `TODO_MOVED` event omits `todoListId` or references a list not in the workspace.                         |
| `TAG_NOT_FOUND`                 | A `TODO_TAG_ADDED` event omits `tagId` or references a tag not in the workspace.                           |
| `USER_NOT_IN_PROJECT`           | A `TODO_ASSIGNEE_ADDED` event omits `assigneeId` or references a user not in the workspace.                |
| `FORBIDDEN`                     | The caller lacks permission to create or edit fields (see Permissions).                                    |

## Permissions

- **Create or edit the field:** `createCustomField` / `editCustomField` require the `OWNER` or `ADMIN` role on the workspace.
- **Read the value:** any member who can view the record can read its time-duration value through `todoQueries`.

## Related

- [Set custom field values](/api/custom-fields/custom-field-values)
- [Date field](/api/custom-fields/date)
- [Number field](/api/custom-fields/number)
- [Formula field](/api/custom-fields/formula)
- [List records](/api/records/list-records)
- [Custom Fields overview](/api/custom-fields)
