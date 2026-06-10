---
title: Currency Conversion Field
description: Create fields that automatically convert a record's currency value into other currencies using daily exchange rates.
icon: ArrowLeftRight
order: 25
---

A Currency Conversion field automatically converts the value of a source [Currency field](/api/custom-fields/currency) into one or more target currencies. Whenever the source value changes, Blue looks up the exchange rate and writes the converted amount onto the conversion field — you never set its value by hand.

Currency Conversion fields are `CustomField` objects with `type: CURRENCY_CONVERSION`. You create them with the `createCustomField` mutation and define their currency pairs with `createCustomFieldOptions`. Exchange rates come from the [Frankfurter API](https://frankfurter.dev), which blends daily reference rates from the European Central Bank and 50+ other central banks and institutions, with history back to 1999.

All examples call `https://api.blue.app/graphql` and require the standard auth headers (`X-Bloo-Token-ID`, `X-Bloo-Token-Secret`, `X-Bloo-Company-ID`) plus `X-Bloo-Project-ID` — custom fields are scoped to the workspace named in that header. Headers are case-insensitive. See [Authentication](/api/start-guide/authentication) for details.

## Overview

A working conversion setup has three pieces:

1. A source field of type `CURRENCY` that holds the original amount and currency.
2. A `CURRENCY_CONVERSION` field whose `currencyFieldId` points at that source field.
3. One or more conversion **options** on the conversion field, each defining a `from → to` currency pair.

When a record's source Currency value is set, Blue finds every Currency Conversion field that references it, matches an option by source currency (falling back to an `Any` option if present), fetches the rate, and stores the converted amount on the record.

## Create

Creating the conversion field and its options takes three calls.

### Step 1 — Create the source Currency field

```graphql
mutation CreateSourceCurrencyField {
  createCustomField(input: { name: "Contract Value", type: CURRENCY, currency: "USD" }) {
    id
    name
    type
  }
}
```

Save the returned `id` — it becomes the `currencyFieldId` in the next step.

### Step 2 — Create the Currency Conversion field

```graphql
mutation CreateConversionField {
  createCustomField(
    input: {
      name: "Contract Value (Converted)"
      type: CURRENCY_CONVERSION
      currencyFieldId: "field_123"
      conversionDateType: "currentDate"
    }
  ) {
    id
    name
    type
    currencyFieldId
    conversionDateType
  }
}
```

`currencyFieldId` is the `id` of the source Currency field from Step 1. `conversionDateType` controls which day's rate is used (see [Conversion date types](#conversion-date-types) below). Save the returned `id` for Step 3.

### Step 3 — Define the conversion options

Each option declares a currency pair. Use `createCustomFieldOptions` to add several at once: the `customFieldId` is given once on the wrapper, and each entry in `customFieldOptions` is a `CustomFieldOptionInput` (which carries no `customFieldId` of its own).

```graphql
mutation CreateConversionOptions {
  createCustomFieldOptions(
    input: {
      customFieldId: "field_123"
      customFieldOptions: [
        { title: "USD to EUR", currencyConversionFrom: "USD", currencyConversionTo: "EUR" }
        { title: "USD to GBP", currencyConversionFrom: "USD", currencyConversionTo: "GBP" }
        { title: "Any to JPY", currencyConversionFrom: "Any", currencyConversionTo: "JPY" }
      ]
    }
  ) {
    id
    title
    currencyConversionFrom
    currencyConversionTo
  }
}
```

To add a single option later, use `createCustomFieldOption` instead — its input (`CreateCustomFieldOptionInput`) carries the `customFieldId` on each call:

```graphql
mutation AddFallbackOption {
  createCustomFieldOption(
    input: {
      customFieldId: "field_123"
      title: "Any currency to EUR"
      currencyConversionFrom: "Any"
      currencyConversionTo: "EUR"
    }
  ) {
    id
    title
  }
}
```

<Callout variant="info" title="Two option input shapes">

`createCustomFieldOptions` (plural) wraps one `customFieldId` plus an array of `CustomFieldOptionInput`. `createCustomFieldOption` (singular) takes a single `CreateCustomFieldOptionInput` with `customFieldId` on each call. Don't put a per-option `customFieldId` inside the plural call's array — it isn't a field there.

</Callout>

### Parameters

#### CreateCustomFieldInput (Currency Conversion fields)

Only the fields relevant to `CURRENCY_CONVERSION` are listed. Custom fields are scoped to the workspace in the `X-Bloo-Project-ID` header — there is no `projectId` parameter.

| Parameter            | Type               | Required             | Description                                                                                                                            |
| -------------------- | ------------------ | -------------------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| `name`               | `String!`          | Yes                  | Display name of the field.                                                                                                             |
| `type`               | `CustomFieldType!` | Yes                  | Must be `CURRENCY_CONVERSION`.                                                                                                         |
| `currencyFieldId`    | `String`           | Yes (for conversion) | `id` of the source `CURRENCY` field to convert from. Stored as an empty string if omitted, leaving the field with nothing to convert.  |
| `conversionDateType` | `String`           | No                   | Which day's rate to use. See [Conversion date types](#conversion-date-types). Defaults to current-date behavior when omitted.          |
| `conversionDate`     | `String`           | No                   | Companion value for `conversionDateType` — an ISO date for `specificDate`, or `"todoDueDate"` / a DATE field `id` for `fromDateField`. |
| `description`        | `String`           | No                   | Help text shown alongside the field.                                                                                                   |

#### CreateCustomFieldOptionsInput (the `createCustomFieldOptions` wrapper)

| Parameter            | Type                         | Required | Description                                                      |
| -------------------- | ---------------------------- | -------- | ---------------------------------------------------------------- |
| `customFieldId`      | `String!`                    | Yes      | `id` of the `CURRENCY_CONVERSION` field these options belong to. |
| `customFieldOptions` | `[CustomFieldOptionInput!]!` | Yes      | The options to create.                                           |

#### CustomFieldOptionInput (each entry in `customFieldOptions`)

| Parameter                | Type      | Required             | Description                                                                                                                               |
| ------------------------ | --------- | -------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `title`                  | `String!` | Yes                  | Display label for the conversion pair.                                                                                                    |
| `currencyConversionFrom` | `String`  | Yes (for conversion) | Source ISO currency code (e.g. `"USD"`), or the literal `"Any"` for a fallback. Schema-optional; required for a usable conversion option. |
| `currencyConversionTo`   | `String`  | Yes (for conversion) | Target ISO currency code (e.g. `"EUR"`). Schema-optional; required for a usable conversion option.                                        |
| `color`                  | `String`  | No                   | Hex color for the option.                                                                                                                 |
| `position`               | `Float`   | No                   | Sort position; appended to the end if omitted.                                                                                            |

#### CreateCustomFieldOptionInput (the `createCustomFieldOption` singular input)

Same as `CustomFieldOptionInput`, plus:

| Parameter       | Type      | Required | Description                                  |
| --------------- | --------- | -------- | -------------------------------------------- |
| `customFieldId` | `String!` | Yes      | `id` of the `CURRENCY_CONVERSION` field.     |
| `todoId`        | `String`  | No       | Record context for the option-created event. |

### Conversion date types

`conversionDateType` is a plain `String` in the schema, not an enum. The following values are the ones the conversion engine recognizes; anything else (or omitting it) falls back to current-date behavior.

| Value           | Rate used                                                 | `conversionDate`                                                                |
| --------------- | --------------------------------------------------------- | ------------------------------------------------------------------------------- |
| `currentDate`   | The rate as of the day the source value was last updated. | Not used.                                                                       |
| `specificDate`  | The rate on a fixed historical date.                      | ISO date string, e.g. `"2024-01-01T00:00:00Z"` (only the date portion is used). |
| `fromDateField` | The rate on a date read from another field.               | `"todoDueDate"` for the record's due date, or the `id` of a DATE custom field.  |

### Response

`createCustomField` returns the created `CustomField`:

```json
{
  "data": {
    "createCustomField": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Contract Value (Converted)",
      "type": "CURRENCY_CONVERSION",
      "currencyFieldId": "clm4n8abc000008l0sourcefld",
      "conversionDateType": "currentDate"
    }
  }
}
```

`createCustomFieldOptions` returns the created `CustomFieldOption` list:

```json
{
  "data": {
    "createCustomFieldOptions": [
      {
        "id": "clm4n8opt000008l0usd2eur00",
        "title": "USD to EUR",
        "currencyConversionFrom": "USD",
        "currencyConversionTo": "EUR"
      }
    ]
  }
}
```

## Set a value

You never write the converted amount directly — a Currency Conversion field is read-only. Instead, set the **source `CURRENCY` field** with `setTodoCustomField`, passing `number` and `currency`. Every conversion field that references the source then recomputes automatically.

```graphql
mutation SetSourceCurrency {
  setTodoCustomField(
    input: { todoId: "todo_123", customFieldId: "field_123", number: 1000, currency: "USD" }
  )
}
```

`setTodoCustomField` returns `Boolean!`, so it takes no sub-selection:

```json
{ "data": { "setTodoCustomField": true } }
```

After this call, a `USD → EUR` conversion field on the same record holds the EUR equivalent of 1000 USD; a `USD → GBP` field holds the GBP equivalent, and so on.

## Read a value

The converted amount is read off the conversion `CustomField` in record context — there is no separate value type. Select `number` (the converted amount), `currency` (the target code), and optionally the structured `value` JSON via `Todo.customFields`:

```graphql
query ReadConvertedValue {
  todoQueries {
    todos(
      filter: { companyIds: ["company_123"], projectIds: ["project_123"], todoIds: ["todo_123"] }
    ) {
      items {
        id
        customFields {
          id
          name
          type
          number
          currency
          currencyFieldId
          conversionDateType
          conversionDate
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
            "customFields": [
              {
                "id": "clm4n8opt000008l0convfield",
                "name": "Contract Value (Converted)",
                "type": "CURRENCY_CONVERSION",
                "number": 925.4,
                "currency": "EUR",
                "currencyFieldId": "clm4n8abc000008l0sourcefld",
                "conversionDateType": "currentDate",
                "conversionDate": "",
                "value": {
                  "number": 925.4,
                  "currency": "EUR",
                  "currencyFieldId": "clm4n8abc000008l0sourcefld",
                  "conversionDateType": "currentDate",
                  "conversionDate": ""
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

## Errors

These codes apply to the create and set operations on this page (each verified against the API's error definitions).

| Code                     | Operation                                          | When                                                                           |
| ------------------------ | -------------------------------------------------- | ------------------------------------------------------------------------------ |
| `CUSTOM_FIELD_NOT_FOUND` | `createCustomFieldOption(s)`, `setTodoCustomField` | The `customFieldId` doesn't exist or isn't in a workspace you can access.      |
| `TODO_NOT_FOUND`         | `setTodoCustomField`                               | The `todoId` doesn't exist or isn't accessible.                                |
| `FORBIDDEN`              | `createCustomField`, `createCustomFieldOption`     | The caller isn't an OWNER or ADMIN on the workspace.                           |
| `BAD_USER_INPUT`         | `setTodoCustomField`                               | The field type can't be set this way, or a referenced option/value is invalid. |
| `PLAN_LIMIT_REACHED`     | `createCustomField`                                | The workspace has hit its custom-field limit, or the organization is locked.   |

## Notes

- **Read-only.** A `CURRENCY_CONVERSION` field is computed from its source Currency field. `setTodoCustomField` on the conversion field itself does nothing — set the source field instead. See [Set Custom Field Values](/api/custom-fields/custom-field-values).
- **`Any` fallback.** An option with `currencyConversionFrom: "Any"` is used when no option matches the source currency exactly, letting one conversion field handle records denominated in different currencies.
- **Same-currency shortcut.** When the matched pair has identical `from` and `to` currencies, the amount is copied through without calling the rate API.
- **On failure, the amount is 0.** If the rate API is unavailable or returns no rate (e.g. an unsupported currency code), the converted `number` is stored as `0` with the target `currency` still set. No error is surfaced to the caller.
- **No conversion runs** when the source field has no amount, or when no option matches and no `Any` fallback exists.
- **Crypto and custom rates aren't supported.** Frankfurter covers fiat currencies only; there's no way to supply your own rate, and crypto codes (BTC, ETH) can be entered in a Currency field but not converted.

## Related

- [Currency Field](/api/custom-fields/currency) — the source field these conversions read from.
- [Date Field](/api/custom-fields/date) — used by the `fromDateField` conversion date type.
- [Set Custom Field Values](/api/custom-fields/custom-field-values) — how `setTodoCustomField` works across field types.
- [Create Custom Fields](/api/custom-fields/create-custom-fields) — the full `createCustomField` reference.
- [Custom Fields Overview](/api/custom-fields) — all field types and concepts.
