---
title: Phone Field
description: Store phone numbers on records with a PHONE custom field — international-format validation on create, raw storage on update.
icon: Phone
order: 15
---

A phone custom field stores a single phone number on a record. It's the `PHONE` value of the `CustomFieldType` enum and is the right type for contact numbers, emergency contacts, or any per-record phone you want to display and filter. Custom fields are `CustomField` objects in the API, and records are `Record` objects.

The two mutations that write a phone value behave differently, and this is the single most important thing to know about this field type:

- **`createRecord`** validates and formats the number — it must include a country code, and the country is derived and stored automatically.
- **`setRecordCustomField`** stores exactly what you send, with no validation or formatting.

## Overview

|            |                                                                              |
| ---------- | ---------------------------------------------------------------------------- |
| Field type | `PHONE`                                                                      |
| Set with   | `setRecordCustomField` → `text` argument (plus optional `regionCode`)        |
| Stored on  | `CustomField.text` (the number), `CustomField.regionCode` (ISO country code) |
| Read with  | `CustomField.text`, `CustomField.regionCode`, or `CustomField.value`         |
| Validated  | Only by `createRecord`/`createTodo`; `setRecordCustomField` stores as-is     |

## Create

Use the `createCustomField` mutation with `type: PHONE`. The field is scoped to the workspace you pass in the `X-Bloo-Project-ID` header — there is no `projectId` argument or input field.

```graphql
mutation CreatePhoneField {
  createCustomField(
    input: { name: "Contact Phone", type: PHONE, description: "Include the country code" }
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
| `type`        | `CustomFieldType!` | Yes      | Must be `PHONE`.                     |
| `description` | `String`           | No       | Help text shown to users in the app. |

### Response

```json
{
  "data": {
    "createCustomField": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Contact Phone",
      "type": "PHONE"
    }
  }
}
```

## Set a value

Use `setRecordCustomField` with the `text` argument to set the phone number on a record. The mutation returns `Boolean!` — `true` on success — so it takes no sub-selection.

```graphql
mutation SetPhoneValue {
  setRecordCustomField(
    input: { todoId: "todo_123", customFieldId: "field_123", text: "+1 234 567 8900" }
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

| Parameter       | Type      | Required | Description                                                                                                                     |
| --------------- | --------- | -------- | ------------------------------------------------------------------------------------------------------------------------------- |
| `todoId`        | `String!` | Yes      | ID of the record to update.                                                                                                     |
| `customFieldId` | `String!` | Yes      | ID of the phone field.                                                                                                          |
| `text`          | `String`  | No       | Phone number to store. No validation or formatting is applied — send it pre-formatted if you need international display format. |
| `regionCode`    | `String`  | No       | ISO country code to store alongside the number. Not derived automatically by this mutation.                                     |

### Set the value when creating a record

`createRecord` accepts custom-field values inline through `customFields`. Each entry is a `CreateRecordInputCustomField` whose `value` is the phone number. Unlike `setRecordCustomField`, this path validates: the resolver parses the value with libphonenumber-js, stores it on `text` in international display format, and derives `regionCode` from the number automatically. A value that cannot be parsed is rejected with `CUSTOM_FIELD_VALUE_PARSE_ERROR`.

```graphql
mutation CreateRecordWithPhone {
  createRecord(
    input: {
      title: "Call client"
      todoListId: "list_123"
      customFields: [{ customFieldId: "field_123", value: "+1 234 567 8900" }]
    }
  ) {
    id
    title
    customFields {
      id
      name
      type
      text
      regionCode
    }
  }
}
```

```json
{
  "data": {
    "createRecord": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Call client",
      "customFields": [
        {
          "id": "clm4n8qwx000108l0a1b2c3d4",
          "name": "Contact Phone",
          "type": "PHONE",
          "text": "+1 234 567 8900",
          "regionCode": "US"
        }
      ]
    }
  }
}
```

## Read a value

Query the record with the top-level `todo(id:)` query and select `customFields`. The field returns `[CustomField!]!` directly — each element is a `CustomField`, with no wrapper object. For a `PHONE` field, the number is on `text` and the country on `regionCode`; `value` is a convenience accessor returning the same stored number.

```graphql
query GetRecordWithPhone {
  todo(id: "todo_123") {
    id
    title
    customFields {
      id
      name
      type
      text
      regionCode
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
      "title": "Call client",
      "customFields": [
        {
          "id": "clm4n8qwx000108l0a1b2c3d4",
          "name": "Contact Phone",
          "type": "PHONE",
          "text": "+1 234 567 8900",
          "regionCode": "US",
          "value": "+1 234 567 8900"
        }
      ]
    }
  }
}
```

### Returns

| Field        | Type               | Description                                                                                                                    |
| ------------ | ------------------ | ------------------------------------------------------------------------------------------------------------------------------ |
| `id`         | `ID!`              | The custom field's ID.                                                                                                         |
| `name`       | `String!`          | Display name of the field.                                                                                                     |
| `type`       | `CustomFieldType!` | Always `PHONE` for this field.                                                                                                 |
| `text`       | `String`           | The stored phone number.                                                                                                       |
| `regionCode` | `String`           | ISO country code stored alongside the number (e.g. `US`, `GB`). Only populated if set explicitly or derived by `createRecord`. |
| `value`      | `JSON`             | Convenience accessor returning the same phone string. Only populated when the field is read in a record context.               |

## Notes

- `createRecord`/`createTodo` requires the number to include a country code — E.164 (`+12345678900`) and international formats with punctuation (`+1 (234) 567-8900`, `+1-234-567-8900`) all work. National formats without a country code (e.g. `(234) 567-8900`) are rejected with `CUSTOM_FIELD_VALUE_PARSE_ERROR`.
- `createRecord` formats the accepted number to international display form (`+1 234 567 8900`) and derives `regionCode` from it automatically using libphonenumber-js.
- `setRecordCustomField` does neither — it stores any string in `text` and any value in `regionCode`, unchanged. Validate and format numbers in your own code before calling it if you need clean data.

## Errors

| Code                             | When                                                                                                           |
| -------------------------------- | -------------------------------------------------------------------------------------------------------------- |
| `CUSTOM_FIELD_VALUE_PARSE_ERROR` | `createRecord`/`createTodo` receives a phone value it cannot parse — missing country code or malformed number. |
| `CUSTOM_FIELD_NOT_FOUND`         | No custom field matches `customFieldId` in the workspace.                                                      |
| `TODO_NOT_FOUND`                 | No record matches `todoId`.                                                                                    |
| `FORBIDDEN`                      | The caller lacks permission to edit the field or the record.                                                   |

## Related

- [Set custom field values](/api/custom-fields/custom-field-values) — the shared pattern for every field type.
- [Text field](/api/custom-fields/text-single) — general single-line text.
- [Email field](/api/custom-fields/email) — email addresses.
- [URL field](/api/custom-fields/url) — web addresses and links.
- [Custom fields overview](/api/custom-fields) — concepts and the full type list.
