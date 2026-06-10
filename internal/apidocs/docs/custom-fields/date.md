---
title: Date Field
description: Store a single date, a date range, or an all-day event on a record, with timezone-aware storage.
icon: Calendar
order: 19
---

A date field stores a point in time or a span of time on a record — a deadline, an event window, a milestone. Each value has a `startDate`, an `endDate`, and an optional `timezone`. A single date is just a range whose start and end are equal. Dates are stored in UTC; the `timezone` is preserved separately so the value can be rendered correctly per viewer.

Date fields are `CustomField` objects with `type: DATE`. Records are `Todo` objects in the API; a workspace is a `Project`.

## Overview

|                   |                                                           |
| ----------------- | --------------------------------------------------------- |
| `CustomFieldType` | `DATE`                                                    |
| Set with          | `setTodoCustomField` → `startDate`, `endDate`, `timezone` |
| Reads back as     | `CustomField.value` → `{ startDate, endDate, timezone }`  |
| Stores            | UTC timestamps + a timezone identifier                    |

Mark a date field as the record's due date with `isDueDate: true`. Due-date fields can drive automations (`DUE_DATE_CHANGED`, `DUE_DATE_REMOVED`) and scheduled reminders.

## Create

Create a date field with the `createCustomField` mutation. The field is scoped to the workspace in your `X-Bloo-Project-ID` header — there is no `projectId` argument.

```graphql
mutation CreateDateField {
  createCustomField(input: { name: "Deadline", type: DATE }) {
    id
    name
    type
  }
}
```

Create a due-date field that automations can act on:

```graphql
mutation CreateDueDateField {
  createCustomField(
    input: {
      name: "Contract Expiration"
      type: DATE
      isDueDate: true
      description: "When the contract expires and needs renewal"
    }
  ) {
    id
    name
    type
    isDueDate
  }
}
```

### CreateCustomFieldInput

| Parameter     | Type               | Required | Description                                                                     |
| ------------- | ------------------ | -------- | ------------------------------------------------------------------------------- |
| `name`        | `String!`          | Yes      | Display name of the field.                                                      |
| `type`        | `CustomFieldType!` | Yes      | Must be `DATE`.                                                                 |
| `isDueDate`   | `Boolean`          | No       | Treat this field as the record's due date so it can drive due-date automations. |
| `description` | `String`           | No       | Help text shown next to the field.                                              |

## Set a value

Set a date with `setTodoCustomField`, supplying `startDate`, `endDate`, and an optional `timezone`. The mutation returns `Boolean!` — it does not return the updated value, so select no subfields. Read the value back with a follow-up query (see [Read a value](#read-a-value)).

A **single date** has the same `startDate` and `endDate`:

```graphql
mutation SetSingleDate {
  setTodoCustomField(
    input: {
      todoId: "todo_123"
      customFieldId: "field_123"
      startDate: "2025-01-15T10:00:00Z"
      endDate: "2025-01-15T10:00:00Z"
      timezone: "America/New_York"
    }
  )
}
```

A **date range** spans two distinct timestamps:

```graphql
mutation SetDateRange {
  setTodoCustomField(
    input: {
      todoId: "todo_123"
      customFieldId: "field_123"
      startDate: "2025-01-01T09:00:00Z"
      endDate: "2025-01-31T17:00:00Z"
      timezone: "Europe/London"
    }
  )
}
```

An **all-day event** spans the full day in the field's timezone (00:00 to 23:59):

```graphql
mutation SetAllDayEvent {
  setTodoCustomField(
    input: {
      todoId: "todo_123"
      customFieldId: "field_123"
      startDate: "2025-01-15T00:00:00Z"
      endDate: "2025-01-15T23:59:59Z"
      timezone: "Asia/Tokyo"
    }
  )
}
```

To clear a date, set both `startDate` and `endDate` to `null`.

### SetTodoCustomFieldInput

| Parameter       | Type       | Required | Description                                                                                                      |
| --------------- | ---------- | -------- | ---------------------------------------------------------------------------------------------------------------- |
| `todoId`        | `String!`  | Yes      | ID of the record to update.                                                                                      |
| `customFieldId` | `String!`  | Yes      | ID of the date field.                                                                                            |
| `startDate`     | `DateTime` | No       | Start of the range, in ISO 8601 (UTC stored).                                                                    |
| `endDate`       | `DateTime` | No       | End of the range, in ISO 8601. For a single date, set it equal to `startDate`.                                   |
| `timezone`      | `String`   | No       | IANA timezone identifier (e.g. `America/New_York`). Defaults to the calling user's detected timezone if omitted. |

<Callout variant="warning" title="endDate is not auto-filled">

`setTodoCustomField` writes `startDate` and `endDate` exactly as given — supplying only `startDate` leaves `endDate` empty. Always send both (set them equal for a single date). The auto-fill behaviour described below applies only to the `createTodo` shorthand.

</Callout>

### Response

```json
{
  "data": {
    "setTodoCustomField": true
  }
}
```

| Field                | Type       | Description                        |
| -------------------- | ---------- | ---------------------------------- |
| `setTodoCustomField` | `Boolean!` | `true` when the value was written. |

## Set a value at record creation

When creating a record, pass date fields inline through `customFields`. Each entry takes a single `value` string. For a date field, `value` is `startDate` and `endDate` joined by a comma — `"start,end"`. Supplying a single date (no comma) sets `startDate` to that date and `endDate` to the end of the same day.

```graphql
mutation CreateRecordWithDate {
  createTodo(
    input: {
      title: "Project Milestone"
      todoListId: "list_123"
      customFields: [{ customFieldId: "field_123", value: "2025-02-15,2025-02-28" }]
    }
  ) {
    id
    title
  }
}
```

### CreateTodoInputCustomField

| Parameter       | Type     | Required | Description                                                                                                   |
| --------------- | -------- | -------- | ------------------------------------------------------------------------------------------------------------- |
| `customFieldId` | `String` | No       | ID of the date field.                                                                                         |
| `value`         | `String` | No       | `"startDate,endDate"`. A single date with no comma sets `startDate` to that date and `endDate` to end-of-day. |

| `value` string            | Result                                                    |
| ------------------------- | --------------------------------------------------------- |
| `"2025-01-15"`            | Single date: start = `2025-01-15`, end = end of that day. |
| `"2025-01-15T10:00:00Z"`  | Single timestamp.                                         |
| `"2025-01-01,2025-01-31"` | Range from start to end.                                  |

## Read a value

A date value comes back on the `CustomField.value` field as `{ startDate, endDate, timezone }`. `startDate` and `endDate` are ISO 8601 strings (or `null`); `value` itself is `null` until the field is read in a record context. There is no `TodoCustomField` wrapper — `Todo.customFields` returns `CustomField` objects directly.

```graphql
query GetRecordDates {
  todo(id: "todo_123") {
    id
    title
    customFields {
      id
      name
      type
      value
    }
  }
}
```

```json
{
  "data": {
    "todo": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Project Milestone",
      "customFields": [
        {
          "id": "clm4n8qwx000108l0abcd1234",
          "name": "Deadline",
          "type": "DATE",
          "value": {
            "startDate": "2025-01-15T10:00:00.000Z",
            "endDate": "2025-01-15T10:00:00.000Z",
            "timezone": "America/New_York"
          }
        }
      ]
    }
  }
}
```

### CustomField date fields

| Field       | Type       | Description                                                                                  |
| ----------- | ---------- | -------------------------------------------------------------------------------------------- |
| `value`     | `JSON`     | `{ startDate, endDate, timezone }` for a `DATE` field in a record context; `null` otherwise. |
| `startDate` | `DateTime` | Start of the range (also exposed directly on `CustomField`).                                 |
| `endDate`   | `DateTime` | End of the range.                                                                            |
| `timezone`  | `String`   | IANA timezone identifier stored with the value.                                              |
| `isDueDate` | `Boolean`  | Whether this field is configured as the record's due date.                                   |

## Filter records by date

To filter records on a date, use the record query `todoQueries { todos(filter: TodosFilter!) }`. For a field marked `isDueDate: true`, the dedicated `TodosFilter` due-date fields are the simplest path.

```graphql
query DueThisQuarter {
  todoQueries {
    todos(
      filter: {
        companyIds: ["company_123"]
        projectIds: ["project_123"]
        duedAtStart: "2025-01-01T00:00:00Z"
        duedAtEnd: "2025-03-31T23:59:59Z"
      }
    ) {
      items {
        id
        title
        duedAt
      }
    }
  }
}
```

Find records that have no due date set with the `hasDueDate` existence flag:

```graphql
query MissingDueDate {
  todoQueries {
    todos(filter: { companyIds: ["company_123"], hasDueDate: false }) {
      items {
        id
        title
      }
    }
  }
}
```

Relevant `TodosFilter` fields: `duedAt`, `duedAtStart`, `duedAtEnd`, `hasDueDate`, plus `createdStart` / `createdEnd` for creation-date ranges. For non-due-date custom fields, filter through the `fields` JSON argument. See [List records](/api/records/list-records) for the full filter surface.

## Notes

- **Storage is UTC.** Timestamps are normalised to UTC on write; the `timezone` is stored alongside so values render correctly per viewer.
- **All-day detection.** A value spanning 00:00 to 23:59 in its timezone is treated as an all-day event and rendered without a time.
- **No end > start validation.** A range with `endDate` before `startDate` is accepted as-is; validate ordering in your client.
- **Time requires a date.** You cannot store a time without a date — every value has at least a `startDate`.
- **`createTodo` end-of-day default.** A single date passed to `createTodo` sets `endDate` to the end of that day. `setTodoCustomField` does not auto-fill; send both dates.

## Errors

| Code                             | When                                                                                    |
| -------------------------------- | --------------------------------------------------------------------------------------- |
| `CUSTOM_FIELD_VALUE_PARSE_ERROR` | The supplied date string can't be parsed as ISO 8601.                                   |
| `CUSTOM_FIELD_NOT_FOUND`         | No date field matches `customFieldId` in the record's workspace.                        |
| `TODO_NOT_FOUND`                 | No record matches `todoId`.                                                             |
| `BAD_USER_INPUT`                 | The field does not belong to the record's workspace, or another input constraint fails. |
| `FORBIDDEN`                      | Your project role lacks edit access to this field or record.                            |

## Permissions

- **Create or update the field:** `OWNER` or `ADMIN` role on an active workspace.
- **Set a value:** standard record edit access (`VIEW_ONLY` and `COMMENT_ONLY` roles are denied).
- **Read a value:** standard record view access.

## Related

- [Custom field values](/api/custom-fields/custom-field-values) — setting and reading values across field types.
- [Create a custom field](/api/custom-fields/create-custom-fields) — the full `createCustomField` surface.
- [List records](/api/records/list-records) — querying and filtering records.
- [Number field](/api/custom-fields/number) · [Country field](/api/custom-fields/country) — sibling field types.
