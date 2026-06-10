---
title: Formula Field
description: Aggregate custom field values (SUM, AVERAGE, COUNT, MIN, MAX) for charts and dashboards.
icon: Calculator
order: 26
---

A Formula field defines an aggregation — `SUM`, `AVERAGE`, `COUNT`, `MIN`, `MAX` — over the values of other custom fields. Formula fields power chart segments and dashboards: they roll values up across the records in a workspace rather than computing a per-record value. Records are `Todo` objects and workspaces are `Project` objects in the API.

Create a Formula field with the `createCustomField` mutation, passing `type: FORMULA`. The field is created in the workspace identified by the `X-Bloo-Project-ID` header.

<Callout variant="info" title="Formulas aggregate; they don't compute per record">

A Formula field has no per-record value. It produces a single aggregated number that surfaces on chart segments (`ChartSegment.formulaResult`), not on each `Todo`. Use a [Rollup field](/api/custom-fields/rollup) if you need per-record computed values.

</Callout>

## Create

The `formula` argument on `createCustomField` is typed `JSON`, so you pass the formula definition as a free-form object. Use the same shape as the typed `FormulaInput` (a `logic` block plus a `display` block) so the field renders identically to one configured in the app.

```graphql
mutation CreateFormulaField {
  createCustomField(
    input: {
      name: "Budget Total"
      type: FORMULA
      formula: {
        logic: { text: "Budget Total", html: "<span>Budget Total</span>" }
        display: { type: NUMBER, precision: 2, function: SUM }
      }
    }
  ) {
    id
    name
    type
    formula
  }
}
```

### CreateCustomFieldInput

| Parameter     | Type               | Required | Description                                             |
| ------------- | ------------------ | -------- | ------------------------------------------------------- |
| `name`        | `String!`          | Yes      | Display name of the field.                              |
| `type`        | `CustomFieldType!` | Yes      | Must be `FORMULA`.                                      |
| `formula`     | `JSON`             | No       | Formula definition. Use the `FormulaInput` shape below. |
| `description` | `String`           | No       | Help text shown next to the field.                      |

The workspace is taken from the `X-Bloo-Project-ID` header, not from the input. There is no `projectId` field on `CreateCustomFieldInput`.

### Formula shape

`createCustomField` accepts `formula` as `JSON`; `editCustomField` accepts it as the typed `FormulaInput`. Both expect the same structure:

| Field               | Type                          | Required | Description                                               |
| ------------------- | ----------------------------- | -------- | --------------------------------------------------------- |
| `logic`             | `FormulaLogicInput!`          | Yes      | Label for the formula.                                    |
| `logic.text`        | `String!`                     | Yes      | Plain-text label.                                         |
| `logic.html`        | `String!`                     | Yes      | HTML label used in the app UI.                            |
| `display`           | `FormulaDisplayInput!`        | Yes      | How the result is computed and formatted.                 |
| `display.type`      | `FormulaDisplayType!`         | Yes      | `NUMBER`, `CURRENCY`, or `PERCENTAGE`.                    |
| `display.function`  | `ChartFunction`               | No       | Aggregation function (see below).                         |
| `display.precision` | `Float`                       | No       | Number of decimal places.                                 |
| `display.currency`  | `FormulaDisplayCurrencyInput` | No       | Currency code + name; only with `display.type: CURRENCY`. |

#### ChartFunction

| Value      | Aggregation                           |
| ---------- | ------------------------------------- |
| `SUM`      | Sum of all values.                    |
| `AVERAGE`  | Mean of all values.                   |
| `AVERAGEA` | Mean, treating empty values as zero.  |
| `COUNT`    | Count of non-empty numeric values.    |
| `COUNTA`   | Count of all values, including empty. |
| `MAX`      | Largest value.                        |
| `MIN`      | Smallest value.                       |

#### FormulaDisplayType

| Value        | Formatting                              | Example     |
| ------------ | --------------------------------------- | ----------- |
| `NUMBER`     | Plain number with optional `precision`. | `1250.75`   |
| `CURRENCY`   | Formatted with `currency.code`.         | `$1,250.75` |
| `PERCENTAGE` | Number with a `%` suffix.               | `87.5%`     |

## Response

`createCustomField` returns the created `CustomField`. The stored definition comes back on `formula` (typed `JSON`).

```json
{
  "data": {
    "createCustomField": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Budget Total",
      "type": "FORMULA",
      "formula": {
        "logic": { "text": "Budget Total", "html": "<span>Budget Total</span>" },
        "display": { "type": "NUMBER", "precision": 2, "function": "SUM" }
      }
    }
  }
}
```

### Returns

| Field         | Type               | Description                      |
| ------------- | ------------------ | -------------------------------- |
| `id`          | `ID!`              | The custom field's identifier.   |
| `name`        | `String!`          | Display name.                    |
| `type`        | `CustomFieldType!` | Always `FORMULA` for this field. |
| `formula`     | `JSON`             | The stored formula definition.   |
| `description` | `String`           | Help text, if set.               |

A Formula field has no per-record `value`. The aggregated result is computed asynchronously and exposed on `ChartSegment.formulaResult` (a `Float`) when the field is used in a chart — it is not returned on the `CustomField` itself or on any record.

## Full example

A currency formula that averages a numeric field, formatted in US dollars to two decimal places.

```graphql
mutation CreateRevenueAverage {
  createCustomField(
    input: {
      name: "Average Deal Size"
      description: "Mean value across all deals in this workspace"
      type: FORMULA
      formula: {
        logic: { text: "Average Deal Size", html: "<span>Average Deal Size</span>" }
        display: {
          type: CURRENCY
          currency: { code: "USD", name: "US Dollar" }
          precision: 2
          function: AVERAGE
        }
      }
    }
  ) {
    id
    name
    formula
  }
}
```

## Edit

Use `editCustomField` to change an existing Formula field. Here `formula` is the typed `FormulaInput` (same shape as create, but type-checked by the schema). Pass the field's `id` as `customFieldId`.

```graphql
mutation UpdateFormulaField {
  editCustomField(
    input: {
      customFieldId: "field_123"
      formula: {
        logic: { text: "Completion Rate", html: "<span>Completion Rate</span>" }
        display: { type: PERCENTAGE, precision: 1, function: AVERAGE }
      }
    }
  ) {
    id
    name
    formula
  }
}
```

## Errors

| Code                     | When                                                              |
| ------------------------ | ----------------------------------------------------------------- |
| `BAD_USER_INPUT`         | The input is invalid (for example, a malformed `formula` block).  |
| `CUSTOM_FIELD_NOT_FOUND` | `editCustomField` was given a `customFieldId` that doesn't exist. |
| `FORBIDDEN`              | The caller isn't an OWNER or ADMIN of the workspace.              |

## Permissions

Creating and editing custom fields requires the OWNER or ADMIN role on the workspace. Reading a field's aggregated result requires view access to the chart or dashboard it appears on.

## Related

- [Number Field](/api/custom-fields/number) — store a numeric value per record.
- [Currency Field](/api/custom-fields/currency) — store a monetary value per record.
- [Lookup Field](/api/custom-fields/lookup) — pull and roll up values from referenced records.
- [Create a Custom Field](/api/custom-fields/create-custom-fields) — the full `createCustomField` reference.
- [Custom Fields Overview](/api/custom-fields) — all field types and shared concepts.
