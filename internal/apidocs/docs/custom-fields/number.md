---
title: Number Custom Field
description: Store numeric values on records, with optional min/max bounds and a display prefix.
icon: Hash
order: 20
---

A number custom field stores a numeric value on a record — quantities, scores, measurements, or any plain number. It maps to the `NUMBER` value of the `CustomFieldType` enum and is backed by a `CustomField` whose value is exposed as a `Float`. For monetary values use the [Currency field](/api/custom-fields/currency); for percentages use the [Percent field](/api/custom-fields/percent).

Records are `Todo` objects in the API.

## Overview

- **Type enum:** `NUMBER`
- **Optional config:** `min` / `max` bounds (`Float`) and a display `prefix` (`String`, e.g. `#`).
- **Value shape:** a single `Float`. On a record, read it from `CustomField.number` (or `CustomField.value`, which returns the same bare number for this type).
- **Bounds enforcement:** `min`/`max` are enforced when you create a record with `createTodo` (out-of-range values throw `CUSTOM_FIELD_VALUE_PARSE_ERROR`). They are **not** enforced by `setTodoCustomField` — see [Notes](#notes).

## Create

Use the `createCustomField` mutation with `type: NUMBER`. The field is created in the workspace named by your `X-Bloo-Project-ID` header — there is no `projectId` argument.

```graphql
mutation CreateNumberField {
  createCustomField(input: { name: "Priority Score", type: NUMBER }) {
    id
    name
    type
  }
}
```

Add `min`, `max`, and a `prefix` to bound and decorate the field:

```graphql
mutation CreateConstrainedNumberField {
  createCustomField(
    input: {
      name: "Team Size"
      type: NUMBER
      min: 1
      max: 100
      prefix: "#"
      description: "Number of people assigned to this work"
    }
  ) {
    id
    name
    type
    min
    max
    prefix
    description
  }
}
```

```json
{
  "data": {
    "createCustomField": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Team Size",
      "type": "NUMBER",
      "min": 1,
      "max": 100,
      "prefix": "#",
      "description": "Number of people assigned to this work"
    }
  }
}
```

### CreateCustomFieldInput

| Parameter     | Type               | Required | Description                                       |
| ------------- | ------------------ | -------- | ------------------------------------------------- |
| `name`        | `String!`          | Yes      | Display name of the field.                        |
| `type`        | `CustomFieldType!` | Yes      | Must be `NUMBER`.                                 |
| `min`         | `Float`            | No       | Lower bound. Enforced by `createTodo`.            |
| `max`         | `Float`            | No       | Upper bound. Enforced by `createTodo`.            |
| `prefix`      | `String`           | No       | Display prefix shown before the value (e.g. `#`). |
| `description` | `String`           | No       | Help text shown to users.                         |

The workspace is taken from the `X-Bloo-Project-ID` header (ID or slug); `CreateCustomFieldInput` has no `projectId` field.

## Set a value

Use the `setTodoCustomField` mutation with the `number` argument. It returns `Boolean!` — `true` on success — so it has no sub-selection.

```graphql
mutation SetNumberValue {
  setTodoCustomField(input: { todoId: "todo_123", customFieldId: "field_123", number: 42.5 })
}
```

```json
{ "data": { "setTodoCustomField": true } }
```

### SetTodoCustomFieldInput

| Parameter       | Type      | Required | Description                                                    |
| --------------- | --------- | -------- | -------------------------------------------------------------- |
| `todoId`        | `String!` | Yes      | ID of the record to update.                                    |
| `customFieldId` | `String!` | Yes      | ID of the number field.                                        |
| `number`        | `Float`   | No       | The numeric value to store. Omit (or send `null`) to clear it. |

## Read a value

`setTodoCustomField` returns only a boolean, so read the stored value back through the record's `customFields` connection. `Todo.customFields` returns `[CustomField!]!` directly — select the value fields on the element. For a `NUMBER` field, both `number` and `value` resolve to the bare numeric value.

```graphql
query ReadNumberValue {
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
          value
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
            "title": "Performance Review",
            "customFields": [
              {
                "id": "clm4n8qwx000108l0a1b2c3d4",
                "name": "Priority Score",
                "type": "NUMBER",
                "number": 42.5,
                "value": 42.5
              }
            ]
          }
        ]
      }
    }
  }
}
```

The `number` and `value` fields resolve only when the `CustomField` is read in a record context (via `Todo.customFields`); they are `null` when the field definition is read on its own. If no value is set, both return `null`.

## Set a value at record creation

`createTodo` accepts custom-field values inline through `CreateTodoInputCustomField`, which carries the value as a **string** in the `value` field (there is no `number` argument here). The string is parsed to a number, and `min`/`max` are enforced — an out-of-range or non-numeric value throws `CUSTOM_FIELD_VALUE_PARSE_ERROR`.

```graphql
mutation CreateRecordWithNumber {
  createTodo(
    input: {
      title: "Performance Review"
      todoListId: "list_123"
      customFields: [{ customFieldId: "field_123", value: "85.5" }]
    }
  ) {
    id
    title
    customFields {
      name
      type
      number
    }
  }
}
```

## Notes

- **Bounds are enforced only at record creation.** `createTodo` parses the inline string value and rejects anything below `min`, above `max`, or non-numeric with `CUSTOM_FIELD_VALUE_PARSE_ERROR`. `setTodoCustomField` writes the `number` argument straight to storage with no bounds check — a value outside `min`/`max` is accepted and stored. Enforce bounds client-side when calling `setTodoCustomField`.
- **`prefix` is display-only.** It is returned on the field definition for clients to render; it is not part of the stored value.
- **Filtering by value** is done through the `fields` JSON of `TodosFilter` (a `CUSTOM_FIELD` entry with `customFieldType: "NUMBER"`), not through per-field operators. See [List records](/api/records/list-records) for the full `fields` filter shape.

```graphql
query FilterByNumber {
  todoQueries {
    todos(
      filter: {
        companyIds: ["company_123"]
        projectIds: ["project_123"]
        fields: [
          {
            type: "CUSTOM_FIELD"
            customFieldId: "field_123"
            customFieldType: "NUMBER"
            values: ["80"]
            op: "GTE"
          }
        ]
      }
    ) {
      items {
        id
        title
      }
    }
  }
}
```

## Errors

| Code                             | When                                                                     |
| -------------------------------- | ------------------------------------------------------------------------ |
| `CUSTOM_FIELD_VALUE_PARSE_ERROR` | A `createTodo` inline value is non-numeric, below `min`, or above `max`. |
| `CUSTOM_FIELD_NOT_FOUND`         | `customFieldId` does not match a field in the active workspace.          |
| `FORBIDDEN`                      | The caller lacks permission for the operation (see Permissions).         |

## Permissions

- **Create the field:** `createCustomField` requires the `OWNER` or `ADMIN` role on the workspace.
- **Set a value:** `setTodoCustomField` requires an authenticated caller with edit access to the record (any company role, or a custom project role granting edit on the field).

## Related

- [Set custom field values](/api/custom-fields/custom-field-values)
- [Currency field](/api/custom-fields/currency)
- [Percent field](/api/custom-fields/percent)
- [Rating field](/api/custom-fields/rating)
- [List records](/api/records/list-records)
- [Custom Fields overview](/api/custom-fields)
