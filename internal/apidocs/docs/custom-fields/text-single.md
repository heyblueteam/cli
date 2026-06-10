---
title: Single-Line Text Field
description: Store short single-line text — names, codes, references — on a record with the TEXT_SINGLE custom field type.
icon: Type
order: 5
---

A single-line text field stores short, single-line text such as names, titles, codes, or reference numbers. It is the `TEXT_SINGLE` value of the `CustomFieldType` enum. Custom fields are `CustomField` objects scoped to a workspace (a `Project` in the API); records are `Todo` objects.

`TEXT_SINGLE` and [`TEXT_MULTI`](/api/custom-fields/text-multi) store and validate text the same way — the difference is the editor the app renders (a single-line input versus a multi-line textarea). Choose `TEXT_SINGLE` for values meant to stay on one line.

## Create

Create the field with `createCustomField`. The workspace is taken from the `X-Bloo-Project-ID` header — there is no `projectId` argument.

```graphql
mutation CreateTextSingleField {
  createCustomField(input: { name: "Client Name", type: TEXT_SINGLE }) {
    id
    name
    type
  }
}
```

Pass `description` to show help text under the field in the app:

```graphql
mutation CreateDetailedTextSingleField {
  createCustomField(
    input: {
      name: "Product SKU"
      type: TEXT_SINGLE
      description: "Unique product identifier code."
    }
  ) {
    id
    name
    type
    description
  }
}
```

### CreateCustomFieldInput

| Parameter     | Type               | Required | Description                                 |
| ------------- | ------------------ | -------- | ------------------------------------------- |
| `name`        | `String!`          | Yes      | Display name of the field.                  |
| `type`        | `CustomFieldType!` | Yes      | Must be `TEXT_SINGLE`.                      |
| `description` | `String`           | No       | Help text shown under the field in the app. |

A successful call returns the created `CustomField`:

```json
{
  "data": {
    "createCustomField": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Product SKU",
      "type": "TEXT_SINGLE"
    }
  }
}
```

## Set a value

Write text to a record with `setTodoCustomField`, using the `text` argument. The mutation returns `Boolean!` — `true` on success — so it takes no selection set.

```graphql
mutation SetTextSingleValue {
  setTodoCustomField(
    input: { todoId: "todo_123", customFieldId: "field_123", text: "ORD-2024-001" }
  )
}
```

```json
{ "data": { "setTodoCustomField": true } }
```

### SetTodoCustomFieldInput

| Parameter       | Type      | Required | Description                                 |
| --------------- | --------- | -------- | ------------------------------------------- |
| `todoId`        | `String!` | Yes      | ID of the record to write to.               |
| `customFieldId` | `String!` | Yes      | ID of the `TEXT_SINGLE` field.              |
| `text`          | `String`  | No       | The text to store. Omit to clear the value. |

You can also set the value when creating a record. `CreateTodoInput.customFields` takes `CreateTodoInputCustomField` entries (`{ customFieldId, value }`), where `value` is the text passed as a string:

```graphql
mutation CreateRecordWithSku {
  createTodo(
    input: {
      title: "Process order ORD-2024-001"
      todoListId: "list_123"
      customFields: [{ customFieldId: "field_123", value: "ORD-2024-001" }]
    }
  ) {
    id
    title
  }
}
```

## Read a value

`Todo.customFields` returns `CustomField` objects directly — there is no junction wrapper type. Read the value on each element. For `TEXT_SINGLE`, the `value` field resolves to the stored string (the same string is also on the `text` field).

```graphql
query GetRecordSku {
  todoQueries {
    todos(filter: { companyIds: ["company_123"], todoIds: ["todo_123"] }) {
      items {
        id
        title
        customFields {
          name
          type
          text
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
            "title": "Process order ORD-2024-001",
            "customFields": [
              {
                "name": "Product SKU",
                "type": "TEXT_SINGLE",
                "text": "ORD-2024-001",
                "value": "ORD-2024-001"
              }
            ]
          }
        ]
      }
    }
  }
}
```

### CustomField (record context)

When a `CustomField` is read through `Todo.customFields`, these fields carry the record's value:

| Field   | Type               | Description                                                          |
| ------- | ------------------ | -------------------------------------------------------------------- |
| `name`  | `String!`          | The field's display name.                                            |
| `type`  | `CustomFieldType!` | `TEXT_SINGLE` for this field.                                        |
| `text`  | `String`           | The stored text.                                                     |
| `value` | `JSON`             | The resolved value — the plain text string for `TEXT_SINGLE` fields. |

## Notes

- Text is stored verbatim through the API: no trimming, no length limit beyond the column's storage capacity, and full Unicode support. Forms applied to the field may trim whitespace and enforce required-ness; the direct API does not.
- `value` is only populated when the `CustomField` is read in a record context (through `Todo.customFields`). On a bare field definition fetched from the [`customFields`](/api/custom-fields/list-custom-fields) query it is `null`.
- `TEXT_SINGLE` and `TEXT_MULTI` share storage and validation; the only difference is the editor the app renders. This parity is current behavior, not a guaranteed contract.
- To find records by text content, filter the [`todoQueries.todos`](/api/records/list-records) query with the `fields` JSON filter rather than a dedicated text argument — there is no top-level text-search parameter on this field type.

## Errors

| Code                     | When                                                                                                |
| ------------------------ | --------------------------------------------------------------------------------------------------- |
| `CUSTOM_FIELD_NOT_FOUND` | The `customFieldId` does not exist or is not in a workspace you can access.                         |
| `TODO_NOT_FOUND`         | The `todoId` does not exist or is not in a workspace you can access.                                |
| `CUSTOM_FIELD_LIMIT`     | The workspace has reached its 30 custom-field limit (raise it by upgrading to Enterprise).          |
| `FORBIDDEN`              | Your role cannot perform the action — `VIEW_ONLY` and `COMMENT_ONLY` roles cannot set field values. |

## Related

- [Multi-Line Text Field](/api/custom-fields/text-multi) — longer, free-form text.
- [Email Field](/api/custom-fields/email) — email addresses.
- [URL Field](/api/custom-fields/url) — links.
- [Unique ID Field](/api/custom-fields/unique-id) — auto-generated identifiers.
- [Set custom field values](/api/custom-fields/custom-field-values) — the `setTodoCustomField` reference.
- [Custom Fields overview](/api/custom-fields) — types, access control, and operations.
