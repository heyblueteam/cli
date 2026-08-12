---
title: Percent Field
description: Store percentage values on records. Blue strips the % symbol on input and keeps the raw number internally.
icon: Percent
order: 13
---

A percent field stores a percentage value on a record as a plain number — `75.5` rather than `"75.5%"`. The `%` symbol is a display convention only; Blue strips it on input and the API always returns the raw number. Use percent fields for completion rates, success rates, margins, or any percentage-based metric.

Percent fields are `CustomField` objects with `type: PERCENT`. Records are `Record` objects. Custom fields are scoped to a workspace (`Workspace`) by the `X-Bloo-Project-ID` header.

## Create

Use the `createCustomField` mutation with `type: PERCENT`. No type-specific configuration is required.

```graphql
mutation CreatePercentField {
  createCustomField(input: { name: "Completion Rate", type: PERCENT }) {
    id
    name
    type
  }
}
```

### CreateCustomFieldInput

| Parameter     | Type               | Required | Description                |
| ------------- | ------------------ | -------- | -------------------------- |
| `name`        | `String!`          | Yes      | Display name of the field. |
| `type`        | `CustomFieldType!` | Yes      | Must be `PERCENT`.         |
| `description` | `String`           | No       | Help text shown to users.  |

The workspace is taken from the `blue-workspace-id` header — there is no `projectId` input field. Unlike `NUMBER`, percent fields ignore `min`/`max`: the create resolver never enforces a range, so values above 100 or below 0 are accepted (see [Notes](#notes)).

### Response

```json
{
  "data": {
    "createCustomField": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Completion Rate",
      "type": "PERCENT"
    }
  }
}
```

## Set a value

Use the `setRecordCustomField` mutation with the `number` argument. Pass the percentage as a plain number (`75.5`, not `"75.5%"`). This mutation returns `Boolean!`, so it has no selection set.

```graphql
mutation SetPercentValue {
  setRecordCustomField(input: { todoId: "todo_123", customFieldId: "field_123", number: 75.5 })
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

| Parameter       | Type      | Required | Description                                             |
| --------------- | --------- | -------- | ------------------------------------------------------- |
| `todoId`        | `String!` | Yes      | ID of the record to update.                             |
| `customFieldId` | `String!` | Yes      | ID of the percent field.                                |
| `number`        | `Float`   | No       | The percentage as a raw number (e.g. `75.5` for 75.5%). |

You can also set the value when creating a record with `createRecord`. Here the value is a **string** and may include a `%` symbol — the resolver strips it and `parseFloat`s the rest before storing.

```graphql
mutation CreateRecordWithPercent {
  createRecord(
    input: {
      title: "Marketing Campaign"
      todoListId: "list_123"
      customFields: [{ customFieldId: "field_123", value: "85.5%" }]
    }
  ) {
    id
    title
    customFields {
      name
      type
      number
      value
    }
  }
}
```

`"85.5%"`, `"85.5"`, `"100"`, and `"33.333"` all parse to the same number they represent; the `%` is optional. A value that does not parse to a number raises `CUSTOM_FIELD_VALUE_PARSE_ERROR`.

## Read a value

`Record.customFields` returns `[CustomField!]!` directly — select the value fields on each element. For a percent field the value lives on `number`, and the JSON `value` field returns the same bare number (not an object).

```graphql
query GetRecordPercent {
  todo(id: "todo_123") {
    id
    title
    customFields {
      name
      type
      number
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
      "title": "Marketing Campaign",
      "customFields": [
        {
          "name": "Completion Rate",
          "type": "PERCENT",
          "number": 75.5,
          "value": 75.5
        }
      ]
    }
  }
}
```

Append the `%` symbol in your own UI — the API never returns it.

## Notes

- **The `%` is a display convention, not stored.** Input strings may include it (`createRecord`); the resolver strips it. The stored and returned value is always the raw number.
- **No min/max enforcement.** Unlike `NUMBER` and `RATING`, the create resolver does not validate percent values against `min`/`max`, and does not constrain them to 0–100. Negative values and values over 100 are accepted as-is — enforce ranges in your own application if you need them.
- **Value shape.** `CustomField.number` and `CustomField.value` both return the same bare `Float`. There is no `{ number: … }` wrapper and no `TodoCustomField` wrapper type.

## Errors

| Code                             | When                                                                  |
| -------------------------------- | --------------------------------------------------------------------- |
| `CUSTOM_FIELD_VALUE_PARSE_ERROR` | The `value` string passed to `createRecord` does not parse to a number. |
| `CUSTOM_FIELD_NOT_FOUND`         | No custom field matches `customFieldId`.                              |
| `TODO_NOT_FOUND`                 | No record matches `todoId`.                                           |

## Related

- [Number Field](/api/custom-fields/number) — raw numbers with optional min/max and prefix.
- [Currency Field](/api/custom-fields/currency) — monetary amounts with a currency code.
- [Setting Custom Field Values](/api/custom-fields/custom-field-values) — the full `setRecordCustomField` reference.
- [Custom Fields Overview](/api/custom-fields) — all field types and shared concepts.
