---
title: URL Field
description: Store website addresses and links on records with a URL custom field — set, read, and optionally render as a button.
icon: Link
order: 11
---

A URL custom field stores a single web address on a record. It's the `URL` value of the `CustomFieldType` enum and is the right type for project websites, documentation links, repository URLs, or any per-record link you want to display and filter.

Custom fields are `CustomField` objects in the API, and records are `Todo` objects. The URL is stored on the field's `text` column and is readable through either `text` or the convenience `value` field. A URL field can optionally render in the app as a clickable button instead of a plain link.

## Overview

|            |                                                                        |
| ---------- | ---------------------------------------------------------------------- |
| Field type | `URL`                                                                  |
| Set with   | `setTodoCustomField` → `text` argument                                 |
| Stored on  | `CustomField.text`                                                     |
| Read with  | `CustomField.text` or `CustomField.value` (both return the URL string) |

## Create

Use the `createCustomField` mutation with `type: URL`. The field is scoped to the workspace you pass in the `X-Bloo-Project-ID` header — there is no `projectId` argument or input field.

```graphql
mutation CreateUrlField {
  createCustomField(
    input: { name: "Project Website", type: URL, description: "Link to the project's website" }
  ) {
    id
    name
    type
  }
}
```

### CreateCustomFieldInput

| Parameter            | Type               | Required | Description                                                                           |
| -------------------- | ------------------ | -------- | ------------------------------------------------------------------------------------- |
| `name`               | `String!`          | Yes      | Display name of the field.                                                            |
| `type`               | `CustomFieldType!` | Yes      | Must be `URL`.                                                                        |
| `description`        | `String`           | No       | Help text shown to users in the app.                                                  |
| `urlDisplayAsButton` | `Boolean`          | No       | When `true`, the app renders the URL as a clickable button instead of an inline link. |
| `urlButtonLabel`     | `String`           | No       | Button text shown when `urlDisplayAsButton` is `true` (e.g. `"Open site"`).           |
| `urlButtonColor`     | `String`           | No       | Button color when `urlDisplayAsButton` is `true` (a hex string, e.g. `"#2563eb"`).    |

### Response

```json
{
  "data": {
    "createCustomField": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Project Website",
      "type": "URL"
    }
  }
}
```

### Render the URL as a button

Set `urlDisplayAsButton: true` and supply a label and color to display the link as a button in the app. These three options are purely presentational — the stored value is still the URL string.

```graphql
mutation CreateUrlButtonField {
  createCustomField(
    input: {
      name: "Live Demo"
      type: URL
      urlDisplayAsButton: true
      urlButtonLabel: "Open demo"
      urlButtonColor: "#2563eb"
    }
  ) {
    id
    name
    type
    urlDisplayAsButton
    urlButtonLabel
    urlButtonColor
  }
}
```

## Set a value

Use `setTodoCustomField` with the `text` argument to set the URL on a record. The mutation returns `Boolean!` — `true` on success — so it takes no sub-selection.

```graphql
mutation SetUrlValue {
  setTodoCustomField(
    input: { todoId: "todo_123", customFieldId: "field_123", text: "https://example.com/docs" }
  )
}
```

```json
{
  "data": {
    "setTodoCustomField": true
  }
}
```

### SetTodoCustomFieldInput

| Parameter       | Type      | Required | Description                                                 |
| --------------- | --------- | -------- | ----------------------------------------------------------- |
| `todoId`        | `String!` | Yes      | ID of the record to update.                                 |
| `customFieldId` | `String!` | Yes      | ID of the URL field.                                        |
| `text`          | `String`  | No       | The URL to store. Omit (or pass `null`) to clear the value. |

### Set the value when creating a record

`createTodo` accepts custom-field values inline through `customFields`. Each entry is a `CreateTodoInputCustomField` whose `value` is the URL string. The elements returned in `customFields` are `CustomField` objects — read the URL back from `text`.

```graphql
mutation CreateRecordWithUrl {
  createTodo(
    input: {
      title: "Review documentation"
      todoListId: "list_123"
      customFields: [{ customFieldId: "field_123", value: "https://example.com/docs" }]
    }
  ) {
    id
    title
    customFields {
      id
      name
      type
      text
    }
  }
}
```

```json
{
  "data": {
    "createTodo": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Review documentation",
      "customFields": [
        {
          "id": "clm4n8qwx000108l0a1b2c3d4",
          "name": "Project Website",
          "type": "URL",
          "text": "https://example.com/docs"
        }
      ]
    }
  }
}
```

## Read a value

Query the record with the top-level `todo(id:)` query and select `customFields`. The field returns `[CustomField!]!` directly — each element is a `CustomField`, with no wrapper object. For a `URL` field, both `text` and `value` resolve to the same URL string (`value` is a convenience accessor that reads the underlying `text` column).

```graphql
query GetRecordWithUrl {
  todo(id: "todo_123") {
    id
    title
    customFields {
      id
      name
      type
      text
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
      "title": "Review documentation",
      "customFields": [
        {
          "id": "clm4n8qwx000108l0a1b2c3d4",
          "name": "Project Website",
          "type": "URL",
          "text": "https://example.com/docs",
          "value": "https://example.com/docs"
        }
      ]
    }
  }
}
```

### Returns

| Field   | Type               | Description                                                                                                    |
| ------- | ------------------ | -------------------------------------------------------------------------------------------------------------- |
| `id`    | `ID!`              | The custom field's ID.                                                                                         |
| `name`  | `String!`          | Display name of the field.                                                                                     |
| `type`  | `CustomFieldType!` | Always `URL` for this field.                                                                                   |
| `text`  | `String`           | The stored URL.                                                                                                |
| `value` | `JSON`             | Convenience accessor returning the same URL string. Only populated when the field is read in a record context. |

The button-display configuration is also readable on the field: `urlDisplayAsButton` (`Boolean`), `urlButtonLabel` (`String`), and `urlButtonColor` (`String`).

## Notes

- The API stores the URL exactly as sent — no format validation, protocol normalization, or trimming. Validate and normalize URLs (e.g. add an `https://` prefix) in your own code before calling `setTodoCustomField` if you need clean data.
- A `URL` field stores and returns the same string as a `TEXT_SINGLE` field — the type difference is semantic and drives the app's UI (clickable link, optional button). Choose `URL` when the value is a web address so it renders correctly in the app.
- `value` is only resolved when the `CustomField` is read through a record (it depends on the record context). Reading a bare field definition returns `null` for `value`; use `text` for the stored URL.

## Errors

| Code                     | When                                                         |
| ------------------------ | ------------------------------------------------------------ |
| `CUSTOM_FIELD_NOT_FOUND` | No custom field matches `customFieldId` in the workspace.    |
| `TODO_NOT_FOUND`         | No record matches `todoId`.                                  |
| `CUSTOM_FIELD_LIMIT`     | The workspace has reached its custom field limit on create.  |
| `BAD_USER_INPUT`         | The input is malformed (e.g. a missing required argument).   |
| `FORBIDDEN`              | The caller lacks permission to edit the field or the record. |

## Related

- [Set custom field values](/api/custom-fields/custom-field-values) — the shared pattern for every field type.
- [Single-line text field](/api/custom-fields/text-single) — general single-line text.
- [Email field](/api/custom-fields/email) — email addresses.
- [Button field](/api/custom-fields/button) — action buttons that trigger automations.
- [Custom fields overview](/api/custom-fields) — concepts and the full type list.
