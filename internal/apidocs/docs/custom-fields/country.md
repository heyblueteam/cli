---
title: Country Custom Field
description: Store one or more countries on a record as ISO 3166-1 alpha-2 codes, with name-to-code conversion when records are created.
icon: Globe
order: 21
---

A country custom field stores one or more countries on a record as ISO 3166-1 alpha-2 codes (`US`, `GB`, `FR`). It maps to the `COUNTRY` value of the `CustomFieldType` enum. Records are `Todo` objects and custom fields are `CustomField` objects in the API.

The two mutations that write a country value behave differently, and this is the single most important thing to know about this field type:

- **`createTodo`** validates and converts the value — country names become codes, and unrecognized values are rejected.
- **`setTodoCustomField`** stores whatever you send, with no validation or conversion.

## Overview

A country field holds two pieces of data:

- `countryCodes` — the ISO alpha-2 codes. The API returns them as an array of strings; internally they are persisted as a comma-joined string.
- `text` — the display text (for example a country name or a label like `"North American Markets"`).

## Create

Create the field with `createCustomField`, using `type: COUNTRY`. The field is scoped to the workspace (`Project`) named in the `X-Bloo-Project-ID` header — there is no `projectId` argument on the input.

```graphql
mutation CreateCountryField {
  createCustomField(input: { name: "Country of Origin", type: COUNTRY }) {
    id
    name
    type
  }
}
```

```json
{
  "data": {
    "createCustomField": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Country of Origin",
      "type": "COUNTRY"
    }
  }
}
```

Add an optional `description` to show help text alongside the field:

```graphql
mutation CreateCountryFieldWithHelp {
  createCustomField(
    input: {
      name: "Customer Location"
      type: COUNTRY
      description: "Primary country where the customer is located"
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

| Parameter     | Type               | Required | Description                |
| ------------- | ------------------ | -------- | -------------------------- |
| `name`        | `String!`          | Yes      | Display name of the field. |
| `type`        | `CustomFieldType!` | Yes      | Must be `COUNTRY`.         |
| `description` | `String`           | No       | Help text shown to users.  |

The custom field is attached to the workspace from the `X-Bloo-Project-ID` request header, not from an input argument. Company and project headers accept an ID or a slug.

## Set a value

Use `setTodoCustomField` to write a country value to a record. Pass `countryCodes` (an array of codes), `text`, or both. This mutation returns `Boolean!` — it does not return the updated value, so read it back with a separate query.

`setTodoCustomField` performs **no** validation: the values are stored exactly as provided. Send well-formed ISO alpha-2 codes.

```graphql
mutation SetCountry {
  setTodoCustomField(
    input: {
      todoId: "todo_123"
      customFieldId: "field_123"
      countryCodes: ["US"]
      text: "United States"
    }
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

Store multiple countries by passing more codes:

```graphql
mutation SetMultipleCountries {
  setTodoCustomField(
    input: {
      todoId: "todo_123"
      customFieldId: "field_123"
      countryCodes: ["US", "CA", "MX"]
      text: "North American Markets"
    }
  )
}
```

### SetTodoCustomFieldInput

| Parameter       | Type        | Required | Description                                            |
| --------------- | ----------- | -------- | ------------------------------------------------------ |
| `todoId`        | `String!`   | Yes      | The record to update.                                  |
| `customFieldId` | `String!`   | Yes      | The country field to set.                              |
| `countryCodes`  | `[String!]` | No       | ISO alpha-2 codes. Persisted as a comma-joined string. |
| `text`          | `String`    | No       | Display text, stored independently of `countryCodes`.  |

`countryCodes` and `text` are stored independently — set one, the other, or both. Omitting one leaves it `null` rather than deriving it from the other.

### Setting a country when creating a record

`createTodo` is the only mutation that validates and converts country input. Pass the country in `customFields` as a `CreateTodoInputCustomField` (`customFieldId` + a single `value` string). The value may be a country name or an ISO alpha-2 code; the resolver converts the name to a code, stores the code in `countryCodes`, and keeps the original input in `text`. An unrecognized value throws `CUSTOM_FIELD_VALUE_PARSE_ERROR`.

```graphql
mutation CreateRecordWithCountry {
  createTodo(
    input: {
      title: "International Client"
      todoListId: "list_123"
      customFields: [{ customFieldId: "field_123", value: "France" }]
    }
  ) {
    id
    title
    customFields {
      name
      type
      text
      countryCodes
    }
  }
}
```

```json
{
  "data": {
    "createTodo": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "International Client",
      "customFields": [
        {
          "name": "Country of Origin",
          "type": "COUNTRY",
          "text": "France",
          "countryCodes": ["FR"]
        }
      ]
    }
  }
}
```

Only a single value is converted per record at create time. A comma-separated string like `"US, CA"` is treated as one value and fails conversion. To store several countries, create the record first, then call `setTodoCustomField` with a `countryCodes` array.

## Read a value

Country values are read on the `customFields` connection of a record via `todoQueries.todos`, which returns `{ items, pageInfo }`. Country data lives directly on each `CustomField` element — there is no wrapper type. Select `text` and `countryCodes` (the codes come back as an array).

```graphql
query ReadCountry {
  todoQueries {
    todos(filter: { companyIds: ["company_123"], todoIds: ["todo_123"] }) {
      items {
        id
        title
        customFields {
          name
          type
          text
          countryCodes
        }
      }
    }
  }
}
```

### CustomField return fields (country)

| Field          | Type               | Description                                     |
| -------------- | ------------------ | ----------------------------------------------- |
| `name`         | `String!`          | The field's display name.                       |
| `type`         | `CustomFieldType!` | `COUNTRY` for this field.                       |
| `text`         | `String`           | The display text stored with the value.         |
| `countryCodes` | `[String!]`        | The ISO alpha-2 codes.                          |
| `value`        | `JSON`             | Context-aware JSON representation of the value. |

## Notes

- Blue uses **ISO 3166-1 alpha-2** codes (two letters, e.g. `US`, `GB`, `FR`, `DE`, `JP`). For the full list, see the [ISO Online Browsing Platform](https://www.iso.org/obp/ui/#search/code/).
- `createTodo` accepts either an English country name (`"United States"`) or a code (`"GB"`) and stores the uppercased code. Invalid input throws `CUSTOM_FIELD_VALUE_PARSE_ERROR`.
- `setTodoCustomField` does not validate. If you need validation when updating a record, validate in your application before calling it.
- Codes are persisted internally as a comma-joined string but are always returned as an array.

## Errors

| Code                             | When                                                                          |
| -------------------------------- | ----------------------------------------------------------------------------- |
| `CUSTOM_FIELD_VALUE_PARSE_ERROR` | `createTodo` receives a country value it cannot match to an ISO alpha-2 code. |

```json
{
  "errors": [
    {
      "message": "Invalid country value.",
      "extensions": { "code": "CUSTOM_FIELD_VALUE_PARSE_ERROR" }
    }
  ]
}
```

## Related

- [Create a Custom Field](/api/custom-fields/create-custom-fields)
- [Set Custom Field Values](/api/custom-fields/custom-field-values)
- [Location Custom Field](/api/custom-fields/location)
- [Lookup Custom Field](/api/custom-fields/lookup)
