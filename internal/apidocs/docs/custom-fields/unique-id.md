---
title: Unique ID Field
description: Auto-generate sequential identifiers on records — ticket numbers, order IDs, invoice numbers — with the UNIQUE_ID custom field type.
icon: Fingerprint
order: 28
---

A Unique ID field assigns each record a sequential number — the basis for ticket numbers, order IDs, invoice numbers, or any per-workspace counter. It maps to the `UNIQUE_ID` value of the `CustomFieldType` enum.

Records are `Record` objects and workspaces are `Workspace` objects in the API. A Unique ID field runs in one of two modes: **sequence mode** (`useSequenceUniqueId: true`) auto-numbers records in creation order, or **manual mode** (the default) behaves like a text field you fill in yourself.

## Overview

In sequence mode, Blue maintains a counter per field and assigns the next integer to each record. The raw counter value is stored as an integer; the `prefix` and `sequenceDigits` you configure are applied at **display time** in the app, so the API returns the bare sequence number (`42`), not the formatted string (`TASK-042`). Assignment runs asynchronously in a background worker that holds a row lock while it numbers records, so concurrent record creation never produces duplicates or gaps from races.

## Create

Create a sequence-mode field with the `createCustomField` mutation. The field is scoped to the workspace in the `blue-workspace-id` header — there is no `projectId` argument.

```graphql
mutation CreateUniqueIdField {
  createCustomField(input: { name: "Ticket Number", type: UNIQUE_ID, useSequenceUniqueId: true }) {
    id
    name
    type
    useSequenceUniqueId
  }
}
```

To format the displayed value, add a `prefix` and zero-pad with `sequenceDigits`, and optionally start the counter at a specific number with `sequenceStartingNumber`:

```graphql
mutation CreateFormattedUniqueIdField {
  createCustomField(
    input: {
      name: "Order ID"
      type: UNIQUE_ID
      description: "Auto-generated order identifier"
      useSequenceUniqueId: true
      prefix: "ORD-"
      sequenceDigits: 4
      sequenceStartingNumber: 1000
    }
  ) {
    id
    name
    type
    useSequenceUniqueId
    prefix
    sequenceDigits
    sequenceStartingNumber
  }
}
```

### CreateCustomFieldInput (UNIQUE_ID)

| Parameter                | Type               | Required | Description                                                                                   |
| ------------------------ | ------------------ | -------- | --------------------------------------------------------------------------------------------- |
| `name`                   | `String!`          | Yes      | Display name of the field.                                                                    |
| `type`                   | `CustomFieldType!` | Yes      | Must be `UNIQUE_ID`.                                                                          |
| `description`            | `String`           | No       | Help text shown on the field.                                                                 |
| `useSequenceUniqueId`    | `Boolean`          | No       | Enable auto-numbering. Omit or `false` for manual mode.                                       |
| `prefix`                 | `String`           | No       | Text prepended to the displayed value (e.g. `ORD-`). Display-only.                            |
| `sequenceDigits`         | `Int`              | No       | Zero-pad the displayed number to this width (e.g. `4` → `0042`). Display-only.                |
| `sequenceStartingNumber` | `Int`              | No       | First number in the sequence. Defaults to `1`. Applies only to the first assignment.          |
| `enableBulkAction`       | `Boolean`          | No       | Allow this field in bulk actions. Not all field types support bulk actions; `UNIQUE_ID` does. |

### Response

```json
{
  "data": {
    "createCustomField": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Order ID",
      "type": "UNIQUE_ID",
      "useSequenceUniqueId": true,
      "prefix": "ORD-",
      "sequenceDigits": 4,
      "sequenceStartingNumber": 1000
    }
  }
}
```

When the field is created in sequence mode, a background job numbers every existing record in the workspace and continues numbering new records as they are created.

## Set a value

In **sequence mode** you do not set the value — Blue assigns it automatically. The four lines below apply only to **manual mode** (`useSequenceUniqueId` omitted or `false`), where the field behaves like a text field. Write a value with `setRecordCustomField` using the `text` parameter.

```graphql
mutation SetUniqueIdValue {
  setRecordCustomField(
    input: { todoId: "todo_123", customFieldId: "field_123", text: "CUSTOM-ID-001" }
  )
}
```

`setRecordCustomField` returns a `Boolean!` — there are no subfields to select. Read the value back with a separate query (see below).

```json
{
  "data": {
    "setRecordCustomField": true
  }
}
```

### SetRecordCustomFieldInput (UNIQUE_ID, manual mode)

| Parameter       | Type      | Required | Description                     |
| --------------- | --------- | -------- | ------------------------------- |
| `todoId`        | `String!` | Yes      | The record to set the value on. |
| `customFieldId` | `String!` | Yes      | The Unique ID field.            |
| `text`          | `String`  | No       | The manual identifier to store. |

## Read a value

Read records and their field values with the `todos` query under `recordQueries`. `TodosFilter.companyIds` is required; scope further with `projectIds`. The query returns a `TodosResult` (`items` + `pageInfo`), and `Record.customFields` is a list of `CustomField` objects — there is no wrapper type.

```graphql
query GetRecordsWithUniqueIds {
  recordQueries {
    todos(filter: { companyIds: ["company_123"], projectIds: ["project_123"] }) {
      items {
        id
        title
        customFields {
          name
          type
          prefix
          sequenceDigits
          sequenceId
          text
          value
        }
      }
      pageInfo {
        totalItems
        hasNextPage
      }
    }
  }
}
```

On each `CustomField` element:

- **Sequence mode:** `sequenceId` (and `value`) hold the raw integer counter. `prefix`/`sequenceDigits` come back as the field's display config — apply them yourself to render `ORD-0042`. `text` is `null`.
- **Manual mode:** `text` holds the value you set; `sequenceId` and `value` are `null`.

```json
{
  "data": {
    "recordQueries": {
      "todos": {
        "items": [
          {
            "id": "clm4n8qwx000008l0g4oxdqn7",
            "title": "Fix login issue",
            "customFields": [
              {
                "name": "Order ID",
                "type": "UNIQUE_ID",
                "prefix": "ORD-",
                "sequenceDigits": 4,
                "sequenceId": 42,
                "text": null,
                "value": 42
              }
            ]
          }
        ],
        "pageInfo": { "totalItems": 1, "hasNextPage": false }
      }
    }
  }
}
```

### Returns (CustomField, UNIQUE_ID)

| Field                    | Type      | Description                                                      |
| ------------------------ | --------- | ---------------------------------------------------------------- |
| `useSequenceUniqueId`    | `Boolean` | Whether auto-numbering is enabled.                               |
| `prefix`                 | `String`  | Display prefix for generated values.                             |
| `sequenceDigits`         | `Int`     | Display zero-padding width.                                      |
| `sequenceStartingNumber` | `Int`     | The configured starting number.                                  |
| `sequenceId`             | `Int`     | The raw sequence number assigned to this record (sequence mode). |
| `text`                   | `String`  | The stored value in manual mode; `null` in sequence mode.        |
| `value`                  | `JSON`    | The record value — the integer sequence number for `UNIQUE_ID`.  |

## Notes

- **The API returns the bare number, not the formatted string.** `prefix` and `sequenceDigits` are display config; combine them client-side (`` `${prefix}${String(sequenceId).padStart(sequenceDigits, '0')}` ``) to reproduce what the app shows.
- **Assignment is asynchronous.** New records get a number shortly after creation, not in the same transaction. Don't assume `sequenceId` is populated the instant a record is created.
- **Sequences never reuse numbers.** Deleting a record leaves a gap; the counter only moves forward. `sequenceStartingNumber` sets the first value only — once the counter advances, later edits to it don't renumber existing records.
- **Sequences are per field.** Each `UNIQUE_ID` field keeps its own counter, scoped to its workspace.
- **`enableBulkAction`** is an optional `CreateCustomFieldInput` flag that exposes a field in bulk actions; `UNIQUE_ID` supports it.

## Permissions

Creating or editing a Unique ID field requires the `OWNER` or `ADMIN` role on the workspace. Setting a manual value and reading values use standard record edit and view permissions.

## Errors

| Code                     | When                                                                             |
| ------------------------ | -------------------------------------------------------------------------------- |
| `FORBIDDEN`              | The caller lacks `OWNER`/`ADMIN` on the workspace, or edit access to the record. |
| `CUSTOM_FIELD_NOT_FOUND` | The `customFieldId` does not exist in this workspace.                            |
| `TODO_NOT_FOUND`         | The `todoId` in `setRecordCustomField` does not exist.                             |
| `BAD_USER_INPUT`         | The input is malformed (e.g. `enableBulkAction` on an unsupported type).         |

## Related

- [Create a Custom Field](/api/custom-fields/create-custom-fields)
- [Set Custom Field Values](/api/custom-fields/custom-field-values)
- [List Custom Fields](/api/custom-fields/list-custom-fields)
- [Number Field](/api/custom-fields/number)
- [Single-Line Text Field](/api/custom-fields/text-single)
- [Custom Fields Overview](/api/custom-fields)
