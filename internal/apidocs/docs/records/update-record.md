---
title: Update Record
description: Edit a record's title, description, dates, color, cover, and tags with editTodo, and write custom field values with setTodoCustomField.
icon: FileEdit
order: 2
---

Use the `editTodo` mutation to change a record's built-in fields — title, description, due dates, color, cover, and tags — and use the `setTodoCustomField` mutation to write a value to one custom field on a record. Records are `Todo` objects in the API and custom fields are `CustomField` objects.

`editTodo` returns the updated `Todo`. `setTodoCustomField` returns `Boolean!` — it does not return the field value, so read it back with [List Records](/api/records/list-records) when you need to confirm the result.

## Request

Update a record's title and description. `todoId` is the only required field; include only the fields you want to change.

```graphql
mutation UpdateRecord {
  editTodo(
    input: {
      todoId: "todo_123"
      title: "Finalize Q3 proposal"
      html: "<p>Send to legal by Friday.</p>"
      text: "Send to legal by Friday."
    }
  ) {
    id
    title
    text
    html
  }
}
```

<Callout variant="info" title="Set the workspace header">

`editTodo` resolves the record against the workspace in the `X-Bloo-Project-ID` header. A record can live in more than one workspace; the header determines which one the move/position change applies to.

</Callout>

## Parameters

### EditTodoInput

| Parameter    | Type                    | Required | Description                                                                                                                                               |
| ------------ | ----------------------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `todoId`     | `String!`               | Yes      | ID of the record to update.                                                                                                                               |
| `title`      | `String`                | No       | New record title. Rejected if the workspace has a name formula enabled (see [Errors](#errors)).                                                           |
| `html`       | `String`                | No       | New description as sanitized HTML. The matching `text` is derived automatically — you do not need to send both.                                           |
| `text`       | `String`                | No       | New description as plain text.                                                                                                                            |
| `startedAt`  | `DateTime`              | No       | Start of the record's due-date range, ISO 8601.                                                                                                           |
| `duedAt`     | `DateTime`              | No       | End of the record's due-date range (the due date), ISO 8601.                                                                                              |
| `color`      | `String`                | No       | Record color as a hex code (see [Color options](#color-options)).                                                                                         |
| `cover`      | `String`                | No       | Cover image URL for the record.                                                                                                                           |
| `position`   | `Float`                 | No       | New sort position within its list. Applies only in the record's original workspace.                                                                       |
| `todoListId` | `String`                | No       | Move the record to this list. To move across workspaces, prefer [Move Record to List](/api/records/move-record-list).                                     |
| `tags`       | `[CreateTodoTagInput!]` | No       | **Replaces** the record's entire tag set with these tags. Pass `[]` to clear all tags. See [Tags](/api/records/tags) for an add/remove-style alternative. |

### CreateTodoTagInput

Each entry either references an existing tag by `id`, or creates/matches one by `title` (and optional `color`).

| Parameter | Type     | Required | Description                                                                   |
| --------- | -------- | -------- | ----------------------------------------------------------------------------- |
| `id`      | `String` | No       | ID of an existing tag in the same workspace.                                  |
| `title`   | `String` | No       | Tag text. If no tag with this title (and color) exists, a new one is created. |
| `color`   | `String` | No       | Hex color for a newly created tag. Defaults to `#4a9fff`.                     |

#### Color options

`color` accepts any of the record colors used in the app:

```json
// Light theme
["#ffc2d4", "#ed8285", "#ffb55e", "#ffe885", "#ccf07d",
 "#91e38c", "#a1f7fa", "#91cfff", "#c29ee0", "#e8bd91"]

// Dark theme
["#ff8ebe", "#ff4b4b", "#ff9e4b", "#ffdc6b", "#b4e051",
 "#66d37e", "#4fd2ff", "#4a9fff", "#a17ee8", "#e89e64"]
```

## Response

```json
{
  "data": {
    "editTodo": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Finalize Q3 proposal",
      "text": "Send to legal by Friday.",
      "html": "<p>Send to legal by Friday.</p>"
    }
  }
}
```

### Returns

`editTodo` returns the updated `Todo`. Commonly selected fields:

| Field       | Type       | Description                                          |
| ----------- | ---------- | ---------------------------------------------------- |
| `id`        | `ID!`      | Record ID.                                           |
| `title`     | `String!`  | Record title.                                        |
| `text`      | `String!`  | Description as plain text.                           |
| `html`      | `String!`  | Description as HTML.                                 |
| `position`  | `Float!`   | Sort position within the list.                       |
| `color`     | `String`   | Record color hex code.                               |
| `cover`     | `String`   | Cover image URL.                                     |
| `startedAt` | `DateTime` | Start of the due-date range.                         |
| `duedAt`    | `DateTime` | Due date (end of the range).                         |
| `done`      | `Boolean!` | Whether the record is complete.                      |
| `tags`      | `[Tag!]!`  | The record's tags (each has `id`, `title`, `color`). |

## Full example

Replace the tag set, move the record to a new list, set a position, and set a due-date range in one call.

```graphql
mutation UpdateRecordFull {
  editTodo(
    input: {
      todoId: "todo_123"
      todoListId: "list_123"
      position: 65535
      color: "#91cfff"
      startedAt: "2026-06-01T09:00:00Z"
      duedAt: "2026-06-05T17:00:00Z"
      tags: [{ id: "tag_123" }, { title: "urgent", color: "#ff4b4b" }]
    }
  ) {
    id
    position
    color
    startedAt
    duedAt
    tags {
      id
      title
      color
    }
  }
}
```

## Update custom fields

Use the `setTodoCustomField` mutation to write one custom field value on a record. Send the field-specific parameter for the field's `CustomFieldType` — passing the wrong parameter is silently ignored, so match the table below. The mutation returns `Boolean!`.

```graphql
mutation SetCustomFieldValue {
  setTodoCustomField(input: { todoId: "todo_123", customFieldId: "field_123", text: "Acme Corp" })
}
```

```json
{ "data": { "setTodoCustomField": true } }
```

### SetTodoCustomFieldInput

| Parameter                     | Type        | Required | Description                                                                                        |
| ----------------------------- | ----------- | -------- | -------------------------------------------------------------------------------------------------- |
| `todoId`                      | `String!`   | Yes      | The record to write to.                                                                            |
| `customFieldId`               | `String!`   | Yes      | The custom field to set.                                                                           |
| `text`                        | `String`    | No       | Text value (`TEXT_SINGLE`, `TEXT_MULTI`, `URL`, `EMAIL`, `PHONE`).                                 |
| `number`                      | `Float`     | No       | Numeric value (`NUMBER`, `PERCENT`, `RATING`, `CURRENCY`).                                         |
| `currency`                    | `String`    | No       | Three-letter currency code for a `CURRENCY` field. Falls back to the field's default when omitted. |
| `checked`                     | `Boolean`   | No       | Checkbox state (`CHECKBOX`).                                                                       |
| `regionCode`                  | `String`    | No       | ISO 3166-1 region code paired with `text` for a `PHONE` field.                                     |
| `countryCodes`                | `[String!]` | No       | ISO 3166-1 alpha-2 codes for a `COUNTRY` field.                                                    |
| `latitude`                    | `Float`     | No       | Latitude for a `LOCATION` field (with `longitude` and `text`).                                     |
| `longitude`                   | `Float`     | No       | Longitude for a `LOCATION` field.                                                                  |
| `startDate`                   | `DateTime`  | No       | Start of a `DATE` field's range, ISO 8601.                                                         |
| `endDate`                     | `DateTime`  | No       | End of a `DATE` field's range (the date itself for a single date), ISO 8601.                       |
| `timezone`                    | `String`    | No       | IANA timezone for a `DATE` field (e.g. `America/New_York`).                                        |
| `customFieldOptionId`         | `String`    | No       | Selected option ID for a `SELECT_SINGLE` field.                                                    |
| `customFieldOptionIds`        | `[String!]` | No       | Selected option IDs for a `SELECT_MULTI` field.                                                    |
| `customFieldReferenceTodoIds` | `[String!]` | No       | Linked record IDs for a `REFERENCE` field. Replaces the field's link set.                          |
| `assigneeUserIds`             | `[String!]` | No       | User IDs for an `ASSIGNEE` field. Replaces the field's assignees.                                  |

<Callout variant="warning" title="Computed fields can't be set">

`FORMULA`, `LOOKUP`, `ROLLUP`, `REFERENCED_BY`, `UNIQUE_ID`, and `CURRENCY_CONVERSION` fields are derived from other data. Calling `setTodoCustomField` on them returns a `BAD_USER_INPUT` error.

</Callout>

### Examples by field type

**Text, URL, email** (`TEXT_SINGLE`, `TEXT_MULTI`, `URL`, `EMAIL`)

```graphql
mutation {
  setTodoCustomField(
    input: { todoId: "todo_123", customFieldId: "field_123", text: "hello@acme.com" }
  )
}
```

**Number, percent, rating** (`NUMBER`, `PERCENT`, `RATING`) — `number` is a `Float`, unquoted.

```graphql
mutation {
  setTodoCustomField(input: { todoId: "todo_123", customFieldId: "field_123", number: 42 })
}
```

**Currency** (`CURRENCY`) — amount plus a currency code. See [Currency Field](/api/custom-fields/currency).

```graphql
mutation {
  setTodoCustomField(
    input: { todoId: "todo_123", customFieldId: "field_123", number: 1500.50, currency: "USD" }
  )
}
```

**Checkbox** (`CHECKBOX`)

```graphql
mutation {
  setTodoCustomField(input: { todoId: "todo_123", customFieldId: "field_123", checked: true })
}
```

**Single-select** (`SELECT_SINGLE`) — one option ID.

```graphql
mutation {
  setTodoCustomField(
    input: { todoId: "todo_123", customFieldId: "field_123", customFieldOptionId: "option_123" }
  )
}
```

**Multi-select** (`SELECT_MULTI`) — the full set of selected option IDs.

```graphql
mutation {
  setTodoCustomField(
    input: {
      todoId: "todo_123"
      customFieldId: "field_123"
      customFieldOptionIds: ["option_123", "option_456"]
    }
  )
}
```

**Date** (`DATE`) — `endDate` is the date; add `startDate` for a range and `timezone` for timezone-aware storage. See [Date Field](/api/custom-fields/date).

```graphql
mutation {
  setTodoCustomField(
    input: {
      todoId: "todo_123"
      customFieldId: "field_123"
      startDate: "2026-06-01T00:00:00Z"
      endDate: "2026-06-05T00:00:00Z"
      timezone: "America/New_York"
    }
  )
}
```

**Phone** (`PHONE`) — E.164 number in `text` with the matching `regionCode`.

```graphql
mutation {
  setTodoCustomField(
    input: {
      todoId: "todo_123"
      customFieldId: "field_123"
      text: "+33642526644"
      regionCode: "FR"
    }
  )
}
```

**Country** (`COUNTRY`) — ISO 3166-1 alpha-2 codes.

```graphql
mutation {
  setTodoCustomField(
    input: { todoId: "todo_123", customFieldId: "field_123", countryCodes: ["AF", "AL", "DZ"] }
  )
}
```

**Location** (`LOCATION`) — coordinates plus a display `text`.

```graphql
mutation {
  setTodoCustomField(
    input: {
      todoId: "todo_123"
      customFieldId: "field_123"
      latitude: 42.2923323
      longitude: 12.1266212
      text: "Capranica, Italy"
    }
  )
}
```

**Reference** (`REFERENCE`) — the records to link. Replaces the existing link set. See [Reference Field](/api/custom-fields/reference).

```graphql
mutation {
  setTodoCustomField(
    input: {
      todoId: "todo_123"
      customFieldId: "field_123"
      customFieldReferenceTodoIds: ["todo_456", "todo_789"]
    }
  )
}
```

**Assignee** (`ASSIGNEE`) — the users assigned through the field. Replaces the existing assignees. See [Assignee Field](/api/custom-fields/assignee).

```graphql
mutation {
  setTodoCustomField(
    input: {
      todoId: "todo_123"
      customFieldId: "field_123"
      assigneeUserIds: ["user_123", "user_456"]
    }
  )
}
```

## Errors

`editTodo`:

| Code                     | When                                                                                                  |
| ------------------------ | ----------------------------------------------------------------------------------------------------- |
| `FORBIDDEN`              | Caller has `VIEW_ONLY` or `COMMENT_ONLY` access, or lacks permission to edit the field being changed. |
| `TODO_NOT_FOUND`         | No record with `todoId` is accessible to the caller.                                                  |
| `TODO_LIST_NOT_FOUND`    | `todoListId` does not resolve to a list the caller can access.                                        |
| `TITLE_EDIT_NOT_ALLOWED` | `title` was set on a record in a workspace that has a name formula enabled.                           |

`setTodoCustomField`:

| Code                     | When                                                                                                                                                               |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `FORBIDDEN`              | Caller has `VIEW_ONLY`/`COMMENT_ONLY` access, or the field is not editable for the caller's role.                                                                  |
| `CUSTOM_FIELD_NOT_FOUND` | No custom field with `customFieldId` is accessible to the caller.                                                                                                  |
| `TODO_NOT_FOUND`         | No record with `todoId` is accessible to the caller.                                                                                                               |
| `BAD_USER_INPUT`         | The field is computed and cannot be set; the field does not belong to the record's workspace; an option ID is unknown; or an `assigneeUserIds` user is ineligible. |

## Permissions

| Access level   | Can update records |
| -------------- | ------------------ |
| `OWNER`        | Yes                |
| `ADMIN`        | Yes                |
| `MEMBER`       | Yes                |
| `CLIENT`       | Yes                |
| `COMMENT_ONLY` | No                 |
| `VIEW_ONLY`    | No                 |

`setTodoCustomField` additionally honors per-field role permissions: if a custom field is marked non-editable for the caller's role, the call is rejected with `FORBIDDEN`.

## Related

- [List Records](/api/records/list-records) — read records and custom field values back.
- [Toggle Record Status](/api/records/toggle-record-status) — mark a record complete (`done` is not an `editTodo` field).
- [Move Record to List](/api/records/move-record-list) — move a record between lists and workspaces.
- [Tags](/api/records/tags) — add or remove individual tags instead of replacing the whole set.
- [Set Custom Field Values](/api/custom-fields/custom-field-values) — the full per-type guide for `setTodoCustomField`.
