---
title: Email Field
description: Store email addresses on records with an EMAIL custom field — set, read, and filter the value.
icon: Mail
order: 10
---

An email custom field stores a single email address on a record. It's the `EMAIL` value of the `CustomFieldType` enum and is the right type for contact addresses, lead emails, or any per-record email you want to filter and display.

Custom fields are `CustomField` objects in the API, and records are `Record` objects. The email value is stored on the field's `text` column and is readable through either `text` or the convenience `value` field.

## Overview

|            |                                                                          |
| ---------- | ------------------------------------------------------------------------ |
| Field type | `EMAIL`                                                                  |
| Set with   | `setRecordCustomField` → `text` argument                                   |
| Stored on  | `CustomField.text`                                                       |
| Read with  | `CustomField.text` or `CustomField.value` (both return the email string) |

## Create

Use the `createCustomField` mutation with `type: EMAIL`. The field is scoped to the workspace you pass in the `blue-workspace-id` header — there is no `projectId` argument or input field.

```graphql
mutation CreateEmailField {
  createCustomField(
    input: { name: "Contact Email", type: EMAIL, description: "Primary contact address" }
  ) {
    id
    name
    type
  }
}
```

### CreateCustomFieldInput

| Parameter     | Type               | Required | Description                          |
| ------------- | ------------------ | -------- | ------------------------------------ |
| `name`        | `String!`          | Yes      | Display name of the field.           |
| `type`        | `CustomFieldType!` | Yes      | Must be `EMAIL`.                     |
| `description` | `String`           | No       | Help text shown to users in the app. |

### Response

```json
{
  "data": {
    "createCustomField": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Contact Email",
      "type": "EMAIL"
    }
  }
}
```

## Set a value

Use `setRecordCustomField` with the `text` argument to set the email on a record. The mutation returns `Boolean!` — `true` on success — so it takes no sub-selection.

```graphql
mutation SetEmailValue {
  setRecordCustomField(
    input: { todoId: "todo_123", customFieldId: "field_123", text: "client@company.com" }
  )
}
```

```json
{
  "data": {
    "setRecordCustomField": true
  }
}
```

### SetRecordCustomFieldInput

| Parameter       | Type      | Required | Description                                                           |
| --------------- | --------- | -------- | --------------------------------------------------------------------- |
| `todoId`        | `String!` | Yes      | ID of the record to update.                                           |
| `customFieldId` | `String!` | Yes      | ID of the email field.                                                |
| `text`          | `String`  | No       | The email address to store. Omit (or pass `null`) to clear the value. |

### Set the value when creating a record

`createRecord` accepts custom-field values inline through `customFields`. Each entry is a `CreateRecordInputCustomField` whose `value` is the email string. The elements returned in `customFields` are `CustomField` objects — read the email back from `text`.

```graphql
mutation CreateRecordWithEmail {
  createRecord(
    input: {
      title: "Follow up with client"
      todoListId: "list_123"
      customFields: [{ customFieldId: "field_123", value: "client@company.com" }]
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
    "createRecord": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Follow up with client",
      "customFields": [
        {
          "id": "clm4n8qwx000108l0a1b2c3d4",
          "name": "Contact Email",
          "type": "EMAIL",
          "text": "client@company.com"
        }
      ]
    }
  }
}
```

## Read a value

Query the record with the top-level `todo(id:)` query and select `customFields`. The field returns `[CustomField!]!` directly — each element is a `CustomField`, with no wrapper object. For an `EMAIL` field, both `text` and `value` resolve to the same email string (`value` is a convenience accessor that reads the underlying `text` column).

```graphql
query GetRecordWithEmail {
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
      "title": "Follow up with client",
      "customFields": [
        {
          "id": "clm4n8qwx000108l0a1b2c3d4",
          "name": "Contact Email",
          "type": "EMAIL",
          "text": "client@company.com",
          "value": "client@company.com"
        }
      ]
    }
  }
}
```

### Returns

| Field   | Type               | Description                                                                                                      |
| ------- | ------------------ | ---------------------------------------------------------------------------------------------------------------- |
| `id`    | `ID!`              | The custom field's ID.                                                                                           |
| `name`  | `String!`          | Display name of the field.                                                                                       |
| `type`  | `CustomFieldType!` | Always `EMAIL` for this field.                                                                                   |
| `text`  | `String`           | The stored email address.                                                                                        |
| `value` | `JSON`             | Convenience accessor returning the same email string. Only populated when the field is read in a record context. |

## Notes

- The API stores the email exactly as sent — no format validation, normalization, or trimming. Validate and lowercase addresses in your own code before calling `setRecordCustomField` if you need clean data.
- `value` is only resolved when the `CustomField` is read through a record (it depends on the record context). Reading a bare field definition returns `null` for `value`; use `text` for the stored address.

## Errors

| Code                     | When                                                         |
| ------------------------ | ------------------------------------------------------------ |
| `CUSTOM_FIELD_NOT_FOUND` | No custom field matches `customFieldId` in the workspace.    |
| `TODO_NOT_FOUND`         | No record matches `todoId`.                                  |
| `BAD_USER_INPUT`         | The input is malformed (e.g. a missing required argument).   |
| `FORBIDDEN`              | The caller lacks permission to edit the field or the record. |

## Related

- [Set custom field values](/api/custom-fields/custom-field-values) — the shared pattern for every field type.
- [Text field](/api/custom-fields/text-single) — general single-line text.
- [URL field](/api/custom-fields/url) — web addresses and links.
- [Phone field](/api/custom-fields/phone) — phone numbers.
- [Custom fields overview](/api/custom-fields) — concepts and the full type list.
