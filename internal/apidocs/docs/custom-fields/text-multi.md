---
title: Multi-Line Text Field
description: Store longer free-form text — descriptions, notes, logs — on a record with the TEXT_MULTI custom field type.
icon: AlignLeft
order: 6
---

A multi-line text field stores free-form text that can span several lines, such as descriptions, notes, or logs. It is the `TEXT_MULTI` value of the `CustomFieldType` enum. Custom fields are `CustomField` objects scoped to a workspace (a `Workspace` in the API); records are `Record` objects.

`TEXT_MULTI` and [`TEXT_SINGLE`](/api/custom-fields/text-single) store and validate text the same way — the difference is the editor the app renders (a multi-line textarea versus a single-line input). Choose `TEXT_MULTI` when you expect line breaks.

## Create

Create the field with `createCustomField`. The workspace is taken from the `blue-workspace-id` header — there is no `projectId` argument.

```graphql
mutation CreateTextMultiField {
  createCustomField(input: { name: "Notes", type: TEXT_MULTI }) {
    id
    name
    type
  }
}
```

Pass `description` to show help text under the field in the app:

```graphql
mutation CreateDetailedTextMultiField {
  createCustomField(
    input: {
      name: "Notes"
      type: TEXT_MULTI
      description: "Free-form notes and observations about this record."
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
| `type`        | `CustomFieldType!` | Yes      | Must be `TEXT_MULTI`.                       |
| `description` | `String`           | No       | Help text shown under the field in the app. |

A successful call returns the created `CustomField`:

```json
{
  "data": {
    "createCustomField": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Notes",
      "type": "TEXT_MULTI"
    }
  }
}
```

## Set a value

Write text to a record with `setRecordCustomField`, using the `text` argument. The mutation returns `Boolean!` — `true` on success — so it takes no selection set.

```graphql
mutation SetTextMultiValue {
  setRecordCustomField(
    input: {
      todoId: "todo_123"
      customFieldId: "field_123"
      text: "Kickoff call done.\n\nNext: send proposal by Friday."
    }
  )
}
```

```json
{ "data": { "setRecordCustomField": true } }
```

To store formatting (bold, italic, headings, lists, links), pass sanitized HTML via `html` instead of `text`. The server derives a plain-text mirror from it automatically — a `text` sent alongside `html` in the same request is ignored.

```graphql
mutation SetTextMultiRichValue {
  setRecordCustomField(
    input: {
      todoId: "todo_123"
      customFieldId: "field_123"
      html: "<p>Kickoff call done.</p><ul><li>Send proposal by Friday</li></ul>"
    }
  )
}
```

### SetRecordCustomFieldInput

| Parameter       | Type      | Required | Description                                                                                                                                                             |
| --------------- | --------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `todoId`        | `String!` | Yes      | ID of the record to write to.                                                                                                                                           |
| `customFieldId` | `String!` | Yes      | ID of the `TEXT_MULTI` field.                                                                                                                                           |
| `text`          | `String`  | No       | The text to store. Newlines are preserved. Ignored when `html` is also set. Omit both to clear the value.                                                               |
| `html`          | `String`  | No       | Rich-text HTML to store. Sanitized server-side to a formatting-only allowlist (no images, files, embeds, tables, or mentions); `text` is derived from it automatically. |

You can also set the value when creating a record. `CreateRecordInput.customFields` takes `CreateRecordInputCustomField` entries (`{ customFieldId, value }`), where `value` is the text passed as a string:

```graphql
mutation CreateRecordWithNotes {
  createRecord(
    input: {
      title: "Acme onboarding"
      todoListId: "list_123"
      customFields: [
        {
          customFieldId: "field_123"
          value: "Kickoff call done.\n\nNext: send proposal by Friday."
        }
      ]
    }
  ) {
    id
    title
  }
}
```

## Read a value

`Record.customFields` returns `CustomField` objects directly — there is no junction wrapper type. Read the value on each element. For `TEXT_MULTI`, `text` and `value` always resolve to the plain-text string; `html` resolves to the sanitized rich-text HTML when the value was set with formatting (`null` otherwise).

```graphql
query GetRecordNotes {
  recordQueries {
    todos(filter: { companyIds: ["company_123"], todoIds: ["todo_123"] }) {
      items {
        id
        title
        customFields {
          name
          type
          text
          html
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
    "recordQueries": {
      "todos": {
        "items": [
          {
            "id": "clm4n8qwx000008l0g4oxdqn7",
            "title": "Acme onboarding",
            "customFields": [
              {
                "name": "Notes",
                "type": "TEXT_MULTI",
                "text": "Kickoff call done.\nSend proposal by Friday",
                "html": "<p>Kickoff call done.</p><ul><li>Send proposal by Friday</li></ul>",
                "value": "Kickoff call done.\nSend proposal by Friday"
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

When a `CustomField` is read through `Record.customFields`, these fields carry the record's value:

| Field   | Type               | Description                                                                                          |
| ------- | ------------------ | ---------------------------------------------------------------------------------------------------- |
| `name`  | `String!`          | The field's display name.                                                                            |
| `type`  | `CustomFieldType!` | `TEXT_MULTI` for this field.                                                                         |
| `text`  | `String`           | The plain-text mirror, including line breaks. Always plain text, even when the value has formatting. |
| `html`  | `String`           | Sanitized rich-text HTML, `null` if the value has never been set with formatting.                    |
| `value` | `JSON`             | The resolved value — the plain text string for `TEXT_MULTI` fields (same contract as `text`).        |

## Notes

- Text is stored verbatim through the API: no trimming, no length limit beyond the column's storage capacity, and full Unicode support. Forms applied to the field may trim whitespace and enforce required-ness; the direct API does not.
- `text`/`value` are always plain text — CSV export, automations, webhooks, and this `value` field never receive HTML, regardless of how the value was set. `html` is the only field that carries formatting; read it explicitly if you need it.
- `html` accepts a formatting-only allowlist (paragraphs, headings, bold/italic/underline/strikethrough, inline code, bullet/ordered lists, links) — images, files, embeds, tables, and mentions are stripped on write.
- `value` is only populated when the `CustomField` is read in a record context (through `Record.customFields`). On a bare field definition fetched from the [`customFields`](/api/custom-fields/list-custom-fields) query it is `null`.
- `TEXT_MULTI` and `TEXT_SINGLE` share storage and validation; this parity is current behavior, not a guaranteed contract.

## Errors

| Code                     | When                                                                                                |
| ------------------------ | --------------------------------------------------------------------------------------------------- |
| `CUSTOM_FIELD_NOT_FOUND` | The `customFieldId` does not exist or is not in a workspace you can access.                         |
| `TODO_NOT_FOUND`         | The `todoId` does not exist or is not in a workspace you can access.                                |
| `CUSTOM_FIELD_LIMIT`     | The workspace has reached its 30 custom-field limit (raise it by upgrading to Enterprise).          |
| `FORBIDDEN`              | Your role cannot perform the action — `VIEW_ONLY` and `COMMENT_ONLY` roles cannot set field values. |

## Related

- [Single-Line Text Field](/api/custom-fields/text-single) — short, single-line values.
- [Email Field](/api/custom-fields/email) — email addresses.
- [URL Field](/api/custom-fields/url) — links.
- [Set custom field values](/api/custom-fields/custom-field-values) — the `setRecordCustomField` reference.
- [Custom Fields overview](/api/custom-fields) — types, access control, and operations.
