---
title: Multi-Select Field
description: Store multiple choices from a predefined list of options on a record with a SELECT_MULTI custom field.
icon: ListChecks
order: 8
---

A multi-select custom field stores one or more choices from a predefined list of options on a record. Use it for categories, skills, affected products, or any field where several values from a controlled set apply at once. Multi-select fields are `CustomField` objects with `type: SELECT_MULTI`; each choice is a `CustomFieldOption`, and records are `Todo` objects in the API.

Custom fields are scoped to a workspace by the `X-Bloo-Project-ID` header, so you do not pass a project ID in the field input.

## Overview

A `SELECT_MULTI` field is created in two steps: create the field, then add its options. Set a value on a record by referencing one or more option IDs. The value reads back as the `selectedOptions` array on the record's `CustomField`.

## Create

Create the field with `createCustomField`, then add options with `createCustomFieldOptions`. The field carries the `SELECT_MULTI` type; the options carry the titles and colors.

```graphql
mutation CreateMultiSelectField {
  createCustomField(
    input: {
      name: "Required Skills"
      type: SELECT_MULTI
      description: "All skills this task needs"
    }
  ) {
    id
    name
    type
  }
}
```

Then add options in one call. `createCustomFieldOptions` takes a single `CreateCustomFieldOptionsInput` — the `customFieldId` lives on the wrapper, and each item in `customFieldOptions` is a `CustomFieldOptionInput` with no `customFieldId` of its own.

```graphql
mutation AddOptions {
  createCustomFieldOptions(
    input: {
      customFieldId: "field_123"
      customFieldOptions: [
        { title: "JavaScript", color: "#f7df1e" }
        { title: "React", color: "#61dafb" }
        { title: "Node.js", color: "#339933" }
        { title: "GraphQL", color: "#e10098" }
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

To add a single option to an existing field instead, use `createCustomFieldOption`, whose input does carry `customFieldId`:

```graphql
mutation AddOneOption {
  createCustomFieldOption(
    input: { customFieldId: "field_123", title: "Python", color: "#3776ab" }
  ) {
    id
    title
    color
  }
}
```

### CreateCustomFieldInput

| Parameter     | Type               | Required | Description                     |
| ------------- | ------------------ | -------- | ------------------------------- |
| `name`        | `String!`          | Yes      | Display name of the field.      |
| `type`        | `CustomFieldType!` | Yes      | Must be `SELECT_MULTI`.         |
| `description` | `String`           | No       | Help text shown with the field. |

### CreateCustomFieldOptionsInput

Used by `createCustomFieldOptions` to add several options at once.

| Parameter            | Type                         | Required | Description                                     |
| -------------------- | ---------------------------- | -------- | ----------------------------------------------- |
| `customFieldId`      | `String!`                    | Yes      | The `SELECT_MULTI` field the options belong to. |
| `customFieldOptions` | `[CustomFieldOptionInput!]!` | Yes      | The options to create.                          |

### CustomFieldOptionInput

Each item in `customFieldOptions`. Note there is **no** `customFieldId` here — it is set once on the wrapper.

| Parameter  | Type      | Required | Description                                              |
| ---------- | --------- | -------- | -------------------------------------------------------- |
| `title`    | `String!` | Yes      | Display text for the option.                             |
| `color`    | `String`  | No       | Color for the option (any string; not validated as hex). |
| `position` | `Float`   | No       | Sort order; defaults to the end of the list.             |

### CreateCustomFieldOptionInput

Used by `createCustomFieldOption` to add a single option. This input **does** carry `customFieldId`.

| Parameter       | Type      | Required | Description                        |
| --------------- | --------- | -------- | ---------------------------------- |
| `customFieldId` | `String!` | Yes      | The field the option belongs to.   |
| `title`         | `String!` | Yes      | Display text for the option.       |
| `color`         | `String`  | No       | Color for the option (any string). |
| `position`      | `Float`   | No       | Sort order.                        |

## Set a value

Set the selected options on a record with `setTodoCustomField`, passing the option IDs in `customFieldOptionIds`. This replaces the record's current selection. Pass an empty array to clear all selections.

```graphql
mutation SetSkills {
  setTodoCustomField(
    input: {
      todoId: "todo_123"
      customFieldId: "field_123"
      customFieldOptionIds: ["option_123", "option_456"]
    }
  )
}
```

`setTodoCustomField` returns `Boolean!` — it does not return the updated field. Read the value back with a separate `todoQueries { todos }` query (see [Read a value](#read-a-value)).

```json
{ "data": { "setTodoCustomField": true } }
```

### SetTodoCustomFieldInput

| Parameter              | Type        | Required | Description                                                                         |
| ---------------------- | ----------- | -------- | ----------------------------------------------------------------------------------- |
| `todoId`               | `String!`   | Yes      | The record to update.                                                               |
| `customFieldId`        | `String!`   | Yes      | The `SELECT_MULTI` field to set.                                                    |
| `customFieldOptionIds` | `[String!]` | No       | Option IDs to select. Replaces the current selection; pass an empty array to clear. |

You can also set the value at record creation time. In `createTodo`, the `CreateTodoInputCustomField.value` is a **comma-separated string of option IDs** (no spaces required; whitespace is trimmed) — the resolver splits it and matches the IDs against the field's options.

```graphql
mutation CreateRecordWithSkills {
  createTodo(
    input: {
      title: "Build the new feature"
      todoListId: "list_123"
      customFields: [{ customFieldId: "field_123", value: "option_123,option_456,option_789" }]
    }
  ) {
    id
    title
    customFields {
      name
      type
      selectedOptions {
        id
        title
        color
      }
    }
  }
}
```

## Read a value

A record's fields come back as `Todo.customFields`, which is `[CustomField!]!` — there is no wrapper object. For a `SELECT_MULTI` field, read the chosen options from `selectedOptions`.

```graphql
query GetRecordSkills {
  todoQueries {
    todos(filter: { companyIds: ["company_123"], todoListIds: ["list_123"] }) {
      items {
        id
        title
        customFields {
          name
          type
          selectedOptions {
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
            "title": "Build the new feature",
            "customFields": [
              {
                "name": "Required Skills",
                "type": "SELECT_MULTI",
                "selectedOptions": [
                  { "id": "option_123", "title": "JavaScript", "color": "#f7df1e" },
                  { "id": "option_456", "title": "React", "color": "#61dafb" }
                ]
              }
            ]
          }
        ]
      }
    }
  }
}
```

### CustomField (selected return fields)

| Field                | Type                   | Description                              |
| -------------------- | ---------------------- | ---------------------------------------- |
| `id`                 | `ID!`                  | The field's unique identifier.           |
| `name`               | `String!`              | Display name of the field.               |
| `type`               | `CustomFieldType!`     | `SELECT_MULTI` for a multi-select field. |
| `customFieldOptions` | `[CustomFieldOption!]` | All options defined on the field.        |
| `selectedOptions`    | `[CustomFieldOption!]` | The options chosen on this record.       |

### CustomFieldOption

| Field      | Type      | Description                     |
| ---------- | --------- | ------------------------------- |
| `id`       | `ID!`     | The option's unique identifier. |
| `title`    | `String!` | Display text for the option.    |
| `color`    | `String!` | The option's color.             |
| `position` | `Float!`  | Sort order within the field.    |

## Manage options

Update an option with `editCustomFieldOption`. It identifies the option by `customFieldId` **and** `optionId` — there is no top-level `id` argument.

```graphql
mutation RenameOption {
  editCustomFieldOption(
    input: {
      customFieldId: "field_123"
      optionId: "option_123"
      title: "TypeScript"
      color: "#3178c6"
    }
  ) {
    id
    title
    color
  }
}
```

Reorder options by editing their `position` (there is no dedicated reorder mutation). Use a fractional value to slot an option between two others:

```graphql
mutation ReorderOption {
  editCustomFieldOption(
    input: { customFieldId: "field_123", optionId: "option_123", position: 1.5 }
  ) {
    id
    position
  }
}
```

Delete an option with `deleteCustomFieldOption`, identifying it by `customFieldId` and `optionId`. It returns `Boolean!`.

```graphql
mutation DeleteOption {
  deleteCustomFieldOption(customFieldId: "field_123", optionId: "option_123")
}
```

### EditCustomFieldOptionInput

| Parameter       | Type      | Required | Description                      |
| --------------- | --------- | -------- | -------------------------------- |
| `customFieldId` | `String!` | Yes      | The field the option belongs to. |
| `optionId`      | `String!` | Yes      | The option to edit.              |
| `title`         | `String`  | No       | New display text.                |
| `color`         | `String`  | No       | New color.                       |
| `position`      | `Float`   | No       | New sort order.                  |

### deleteCustomFieldOption arguments

| Argument        | Type      | Required | Description                                                                        |
| --------------- | --------- | -------- | ---------------------------------------------------------------------------------- |
| `customFieldId` | `String!` | Yes      | The field the option belongs to.                                                   |
| `optionId`      | `String!` | Yes      | The option to delete.                                                              |
| `todoId`        | `String`  | No       | Scope the delete to a single record's selection rather than the option definition. |

## Notes

- The `value` in `createTodo` is a comma-separated string of option IDs; `setTodoCustomField` and `bulkSetCustomField` take an array (`customFieldOptionIds`).
- `setTodoCustomField` replaces the record's full selection on each call — it is not additive. Send the complete set of option IDs you want, or an empty array to clear.
- Option IDs must belong to the same field; unknown IDs cause the call to fail.
- `color` accepts any string (no hex validation).

## Errors

| Code                     | When                                                                               |
| ------------------------ | ---------------------------------------------------------------------------------- |
| `CUSTOM_FIELD_NOT_FOUND` | The `customFieldId` does not exist in the workspace.                               |
| `BAD_USER_INPUT`         | An option ID in `customFieldOptionIds` is unknown or does not belong to the field. |
| `FORBIDDEN`              | The caller cannot edit the field or the record's value.                            |

## Related

- [Single-Select Field](/api/custom-fields/select-single) — choose exactly one option.
- [Checkbox Field](/api/custom-fields/checkbox) — a single boolean choice.
- [Custom Field Values](/api/custom-fields/custom-field-values) — set and read values across all field types.
- [Custom Fields Overview](/api/custom-fields) — concepts and the full field-type list.
