---
title: Currency Custom Field
description: Store a monetary amount and its currency on a record, with an optional default currency and min/max range.
icon: DollarSign
order: 8
---

A currency custom field stores a monetary amount together with a currency code (`1500.50` + `USD`) on a record. It maps to the `CURRENCY` value of the `CustomFieldType` enum. Records are `Record` objects and custom fields are `CustomField` objects in the API.

The set of currencies a field offers is stored as `CustomFieldOption` rows on the field — one option per currency (`USD`, `EUR`, `GBP`, …), each identified by its three-letter code in `title`. The field's `currency` property is the default code used when a value omits one.

## Overview

|                               |                                                |
| ----------------------------- | ---------------------------------------------- |
| Enum value                    | `CURRENCY`                                     |
| Value fields on `CustomField` | `number` (the amount), `currency` (the code)   |
| Set with `setRecordCustomField` | `number`, `currency`                           |
| Set with `createRecord`         | a single `value` string encoding amount + code |

## Create

Use the `createCustomField` mutation with `type: CURRENCY`. The field is created in the workspace identified by the `blue-workspace-id` header — there is no `projectId` argument. Set `currency` to the default code and, optionally, `min`/`max` to record an intended range.

```graphql
mutation CreateCurrencyField {
  createCustomField(
    input: { name: "Deal Value", type: CURRENCY, currency: "USD", min: 0, max: 1000000 }
  ) {
    id
    name
    type
    currency
    min
    max
  }
}
```

```json
{
  "data": {
    "createCustomField": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Deal Value",
      "type": "CURRENCY",
      "currency": "USD",
      "min": 0,
      "max": 1000000
    }
  }
}
```

### CreateCustomFieldInput

| Parameter     | Type               | Required | Description                                                                       |
| ------------- | ------------------ | -------- | --------------------------------------------------------------------------------- |
| `name`        | `String!`          | Yes      | Display name of the field.                                                        |
| `type`        | `CustomFieldType!` | Yes      | Must be `CURRENCY`.                                                               |
| `currency`    | `String`           | No       | Default three-letter currency code used when a value is set without one.          |
| `min`         | `Float`            | No       | Lowest intended amount. Recorded on the field; not enforced when setting values.  |
| `max`         | `Float`            | No       | Highest intended amount. Recorded on the field; not enforced when setting values. |
| `description` | `String`           | No       | Help text shown alongside the field.                                              |
| `prefix`      | `String`           | No       | Display prefix shown before the amount.                                           |

<Callout variant="info" title="Offer more than one currency">

A new currency field offers only its default `currency`. To let records choose another code, add a `CustomFieldOption` per currency with `createCustomFieldOptions` — each option's `title` is the three-letter code. When a value is set, the code is matched against these options.

</Callout>

## Set a value

Use the `setRecordCustomField` mutation with `number` (the amount) and `currency` (the code). It returns `Boolean!` — there are no subfields to select.

```graphql
mutation SetDealValue {
  setRecordCustomField(
    input: { todoId: "todo_123", customFieldId: "field_123", number: 1500.50, currency: "USD" }
  )
}
```

```json
{ "data": { "setRecordCustomField": true } }
```

### SetRecordCustomFieldInput

| Parameter       | Type      | Required | Description                                                                            |
| --------------- | --------- | -------- | -------------------------------------------------------------------------------------- |
| `todoId`        | `String!` | Yes      | The record to update.                                                                  |
| `customFieldId` | `String!` | Yes      | The currency field to set.                                                             |
| `number`        | `Float`   | No       | The monetary amount.                                                                   |
| `currency`      | `String`  | No       | Three-letter currency code. Falls back to the field's default `currency` when omitted. |

### Setting a value while creating a record

`createRecord` takes custom field values as `CreateRecordInputCustomField` entries, each with only `customFieldId` and a single `value` string — there is no separate `currency` key here. For a currency field, encode the amount and code into one string:

| `value`    | Result                                        |
| ---------- | --------------------------------------------- |
| `"200USD"` | amount `200`, currency `USD`                  |
| `"USD200"` | amount `200`, currency `USD`                  |
| `"5000"`   | amount `5000`, currency = the field's default |

The code (when present) is matched against the field's currency `CustomFieldOption` rows; an unrecognized code falls back to the field's first option.

```graphql
mutation CreateDeal {
  createRecord(
    input: {
      title: "Acme renewal"
      todoListId: "list_123"
      customFields: [{ customFieldId: "field_123", value: "25000.00GBP" }]
    }
  ) {
    id
    title
    customFields {
      name
      type
      number
      currency
    }
  }
}
```

```json
{
  "data": {
    "createRecord": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Acme renewal",
      "customFields": [
        { "name": "Deal Value", "type": "CURRENCY", "number": 25000, "currency": "GBP" }
      ]
    }
  }
}
```

## Read a value

Read currency values from a record's `customFields` connection. Each element is a `CustomField` — select `number` for the amount and `currency` for the code.

```graphql
query ReadDealValue {
  recordQueries {
    todos(filter: { companyIds: ["company_123"], todoListIds: ["list_123"] }) {
      items {
        id
        title
        customFields {
          name
          type
          number
          currency
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
            "title": "Acme renewal",
            "customFields": [
              { "name": "Deal Value", "type": "CURRENCY", "number": 25000, "currency": "GBP" }
            ]
          }
        ]
      }
    }
  }
}
```

## Notes

- **Amount and code are separate columns.** `number` holds the amount, `currency` holds the code. `setRecordCustomField` writes both directly; `createRecord` parses both out of the single `value` string.
- **`min`/`max` are not enforced.** They are stored on the field as a documented range but are not validated when values are set or updated.
- **Conversion is a separate field type.** A `CURRENCY` field does not convert between currencies. To convert a source amount into other currencies automatically, use a [Currency Conversion field](/api/custom-fields/currency-conversion).

## Errors

| Code                             | When                                                          |
| -------------------------------- | ------------------------------------------------------------- |
| `CUSTOM_FIELD_NOT_FOUND`         | `customFieldId` does not resolve to a field in the workspace. |
| `TODO_NOT_FOUND`                 | `todoId` does not resolve to a record you can access.         |
| `CUSTOM_FIELD_VALUE_PARSE_ERROR` | The supplied amount cannot be parsed as a number.             |
| `BAD_USER_INPUT`                 | A required argument is missing or malformed.                  |
| `FORBIDDEN`                      | You lack permission to edit the field on this record.         |

## Related

- [Currency Conversion field](/api/custom-fields/currency-conversion) — automatic conversion using exchange rates.
- [Number field](/api/custom-fields/number) — numeric values without a currency code.
- [Formula field](/api/custom-fields/formula) — calculations across numeric and currency fields.
- [Set custom field values](/api/custom-fields/custom-field-values) — setting values across all field types.
- [Create a custom field](/api/custom-fields/create-custom-fields) — the full `createCustomField` reference.
