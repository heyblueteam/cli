---
title: Single-Select Field
description: Store one choice from a predefined option list on a record — for status, priority, category, or any controlled single-choice value.
icon: List
order: 16
---

A single-select field stores exactly one option chosen from a predefined list. Use it for status, priority, category, or any value that must come from a controlled set where only one choice is valid at a time. In the API this is a `CustomField` with `type: SELECT_SINGLE`; selectable options are `CustomFieldOption` objects, and the chosen option for a given record is exposed as `selectedOption`.

Custom fields are scoped to a workspace (a `Project`), set with the `X-Bloo-Project-ID` header. Records are `Todo` objects.

## Overview

A single-select field is created in two steps: create the field, then add its options. The field on its own holds no choices — the options you attach to it are what users (and the API) can select.

| Aspect                  | Single-Select                                                         |
| ----------------------- | --------------------------------------------------------------------- |
| `CustomFieldType` value | `SELECT_SINGLE`                                                       |
| Selection limit         | One option                                                            |
| Set value with          | `customFieldOptionId` (or `customFieldOptionIds`, first element only) |
| Reads back as           | `selectedOption` (a single `CustomFieldOption`)                       |

## Create

Create the field with `createCustomField`. There is no inline option list — options are added separately after the field exists.

```graphql
mutation CreateSingleSelectField {
  createCustomField(input: { name: "Priority", type: SELECT_SINGLE }) {
    id
    name
    type
  }
}
```

Then add options to that field with `createCustomFieldOptions`. The `customFieldId` is passed once on the wrapper input; each entry in `customFieldOptions` carries only its own properties.

```graphql
mutation AddPriorityOptions {
  createCustomFieldOptions(
    input: {
      customFieldId: "field_123"
      customFieldOptions: [
        { title: "Low", color: "#28a745" }
        { title: "Medium", color: "#ffc107" }
        { title: "High", color: "#fd7e14" }
        { title: "Critical", color: "#dc3545" }
      ]
    }
  ) {
    id
    title
    color
    position
  }
}
```

To add a single option to an existing field, use `createCustomFieldOption` instead, which takes `customFieldId` directly on its input.

```graphql
mutation AddOneOption {
  createCustomFieldOption(
    input: { customFieldId: "field_123", title: "Urgent", color: "#6f42c1" }
  ) {
    id
    title
    color
    position
  }
}
```

### CreateCustomFieldInput

| Parameter     | Type               | Required | Description                |
| ------------- | ------------------ | -------- | -------------------------- |
| `name`        | `String!`          | Yes      | Display name of the field. |
| `type`        | `CustomFieldType!` | Yes      | Must be `SELECT_SINGLE`.   |
| `description` | `String`           | No       | Help text shown to users.  |

### CreateCustomFieldOptionsInput

| Parameter            | Type                         | Required | Description                                    |
| -------------------- | ---------------------------- | -------- | ---------------------------------------------- |
| `customFieldId`      | `String!`                    | Yes      | The single-select field the options belong to. |
| `customFieldOptions` | `[CustomFieldOptionInput!]!` | Yes      | The options to create.                         |

### CustomFieldOptionInput

| Parameter  | Type      | Required | Description                                          |
| ---------- | --------- | -------- | ---------------------------------------------------- |
| `title`    | `String!` | Yes      | Display text for the option.                         |
| `color`    | `String`  | No       | Color for the option (any string; the app uses hex). |
| `position` | `Float`   | No       | Sort order. Defaults to append order when omitted.   |

## Set a value

Set the selected option on a record with `setTodoCustomField`, passing the chosen option's id as `customFieldOptionId`. Setting a new option replaces any previous selection. `setTodoCustomField` returns `Boolean!` — select no subfields on it.

```graphql
mutation SetPriority {
  setTodoCustomField(
    input: { todoId: "todo_123", customFieldId: "field_123", customFieldOptionId: "option_123" }
  )
}
```

```json
{ "data": { "setTodoCustomField": true } }
```

You can also create a record with the value already set. In `createTodo`, the option id is passed as the string `value` (not a separate option-id field).

```graphql
mutation CreateRecordWithPriority {
  createTodo(
    input: {
      title: "Review onboarding flow"
      todoListId: "list_123"
      customFields: [{ customFieldId: "field_123", value: "option_123" }]
    }
  ) {
    id
    title
  }
}
```

### SetTodoCustomFieldInput

| Parameter              | Type        | Required | Description                                                            |
| ---------------------- | ----------- | -------- | ---------------------------------------------------------------------- |
| `todoId`               | `String!`   | Yes      | The record to update.                                                  |
| `customFieldId`        | `String!`   | Yes      | The single-select field.                                               |
| `customFieldOptionId`  | `String`    | No       | The option to select. Preferred for single-select.                     |
| `customFieldOptionIds` | `[String!]` | No       | Accepted for parity with multi-select; only the first element is used. |

<Callout variant="info" title="Only one id is kept">

If you pass `customFieldOptionIds` to a `SELECT_SINGLE` field, the resolver keeps only the first id and ignores the rest. To clear the selection, call `setTodoCustomField` with neither `customFieldOptionId` nor `customFieldOptionIds`.

</Callout>

## Read a value

A record's custom field values come back on `Todo.customFields`, which returns `[CustomField!]!` directly. For a single-select field, read the chosen option from `selectedOption`. There is no wrapper type — select option fields straight off the element.

```graphql
query GetRecordPriority {
  todoQueries {
    todos(filter: { companyIds: ["company_123"], todoIds: ["todo_123"] }) {
      items {
        id
        title
        customFields {
          id
          name
          type
          selectedOption {
            id
            title
            color
          }
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
            "title": "Review onboarding flow",
            "customFields": [
              {
                "id": "clm4n8field0000priority01",
                "name": "Priority",
                "type": "SELECT_SINGLE",
                "selectedOption": {
                  "id": "clm4n8opt0000highprio0001",
                  "title": "High",
                  "color": "#fd7e14"
                }
              }
            ]
          }
        ]
      }
    }
  }
}
```

`CustomField.value` also returns a JSON representation of the selection, but `selectedOption` is the typed, schema-checked path and the one to prefer.

### Returns

`CustomField` fields relevant to a single-select field:

| Field                | Type                   | Description                                  |
| -------------------- | ---------------------- | -------------------------------------------- |
| `id`                 | `ID!`                  | The field id.                                |
| `name`               | `String!`              | Display name.                                |
| `type`               | `CustomFieldType!`     | `SELECT_SINGLE`.                             |
| `customFieldOptions` | `[CustomFieldOption!]` | All options defined on the field.            |
| `selectedOption`     | `CustomFieldOption`    | The option chosen on this record, or `null`. |

`CustomFieldOption` fields:

| Field      | Type      | Description                  |
| ---------- | --------- | ---------------------------- |
| `id`       | `ID!`     | The option id.               |
| `title`    | `String!` | Display text.                |
| `color`    | `String!` | The option color.            |
| `position` | `Float!`  | Sort order within the field. |

## Manage options

Edit an option with `editCustomFieldOption`, identifying it by `customFieldId` plus `optionId` (not a single `id`). It returns the updated `CustomFieldOption`.

```graphql
mutation RenameOption {
  editCustomFieldOption(
    input: {
      customFieldId: "field_123"
      optionId: "option_123"
      title: "Top Priority"
      color: "#ff6b6b"
    }
  ) {
    id
    title
    color
  }
}
```

Reorder options by setting `position` — there is no separate reorder mutation. Use a value between two existing positions to slot an option in.

```graphql
mutation MoveOption {
  editCustomFieldOption(
    input: { customFieldId: "field_123", optionId: "option_123", position: 1.5 }
  ) {
    id
    position
  }
}
```

Delete an option with `deleteCustomFieldOption`, which takes `customFieldId` and `optionId` and returns `Boolean!`. Deleting an option clears it from every record where it was selected.

```graphql
mutation RemoveOption {
  deleteCustomFieldOption(customFieldId: "field_123", optionId: "option_123")
}
```

```json
{ "data": { "deleteCustomFieldOption": true } }
```

### EditCustomFieldOptionInput

| Parameter       | Type      | Required | Description                      |
| --------------- | --------- | -------- | -------------------------------- |
| `customFieldId` | `String!` | Yes      | The field the option belongs to. |
| `optionId`      | `String!` | Yes      | The option to edit.              |
| `title`         | `String`  | No       | New display text.                |
| `color`         | `String`  | No       | New color.                       |
| `position`      | `Float`   | No       | New sort order.                  |

## Notes

- Setting a new option replaces the previous one — single-select selections are not additive.
- Options live on the field, not the record, so they are shared by every record that uses the field.
- `color` on a created option is stored as a string; the app renders hex values, but no hex validation is enforced.

## Errors

| Code                             | When                                                                      |
| -------------------------------- | ------------------------------------------------------------------------- |
| `CUSTOM_FIELD_NOT_FOUND`         | The `customFieldId` does not resolve to a field in the current workspace. |
| `CUSTOM_FIELD_OPTION_NOT_FOUND`  | The option id does not exist, or does not belong to the given field.      |
| `TODO_NOT_FOUND`                 | The `todoId` does not resolve to a record.                                |
| `CUSTOM_FIELD_VALUE_PARSE_ERROR` | The supplied value cannot be parsed for this field type.                  |
| `FORBIDDEN`                      | The caller lacks permission to modify the field or set the value.         |

## Related

- [Multi-Select Field](/api/custom-fields/select-multi) — choose several options at once.
- [Checkbox Field](/api/custom-fields/checkbox) — a single boolean value.
- [Set Custom Field Values](/api/custom-fields/custom-field-values) — the full `setTodoCustomField` reference across field types.
- [Custom Fields Overview](/api/custom-fields) — concepts shared by every field type.
