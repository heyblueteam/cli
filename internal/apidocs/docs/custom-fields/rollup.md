---
title: Rollup Field
description: Aggregate a field across the records reached through a reference or referenced-by field - sum, average, count, min, max, or join.
icon: Sigma
order: 29
---

A rollup field computes an aggregate over the records reached through a [reference](/api/custom-fields/reference) or [referenced-by](/api/custom-fields/referenced-by) field — for example, the sum of order values on every linked deal. Its `CustomFieldType` is `ROLLUP`. Rollup fields require a Pro subscription.

## Overview

A rollup walks one relationship field, reads one field on each record it reaches, and combines those values with an aggregation function. The value is read-only and recomputed by Blue when the underlying records change.

## Create

```graphql
mutation CreateTotalOrderValue {
  createCustomField(
    input: {
      name: "Total Order Value"
      type: ROLLUP
      rollupOption: {
        sourceType: REFERENCE_FIELD
        sourceFieldId: "field_123"
        aggregateFieldId: "field_456"
        aggregationFunction: SUM
      }
    }
  ) {
    id
    name
    type
    customFieldRollupOption {
      sourceType
      aggregationFunction
    }
  }
}
```

| Parameter      | Type                           | Required | Description                 |
| -------------- | ------------------------------ | -------- | --------------------------- |
| `name`         | `String!`                      | Yes      | Display name of the field.  |
| `type`         | `CustomFieldType!`             | Yes      | Must be `ROLLUP`.           |
| `rollupOption` | `CustomFieldRollupOptionInput` | Yes      | How the rollup is computed. |

### CustomFieldRollupOptionInput

| Field                 | Type                         | Required | Description                                                             |
| --------------------- | ---------------------------- | -------- | ----------------------------------------------------------------------- |
| `sourceType`          | `RollupSourceType!`          | Yes      | Which relationship to walk: `REFERENCE_FIELD` or `REFERENCED_BY_FIELD`. |
| `sourceFieldId`       | `String`                     | No       | The reference or referenced-by field to walk.                           |
| `aggregateFieldId`    | `String`                     | No       | The field on the reached records to aggregate.                          |
| `aggregationFunction` | `RollupAggregationFunction!` | Yes      | How to combine the values.                                              |
| `filters`             | `TodoFilterInput`            | No       | Limit the reached records before aggregating.                           |

`RollupAggregationFunction` is one of `SUM`, `AVERAGE`, `COUNT`, `COUNTA`, `MIN`, `MAX`, or `ARRAYJOIN`. Creating a rollup without a Pro subscription returns a `PRO_REQUIRED` error.

## Set a value

Rollup fields are computed — `setRecordCustomField` rejects them with a `BAD_USER_INPUT` error. The value changes only when the underlying records or their aggregated field change.

## Read a value

A rollup result lands in exactly one of three typed columns depending on what it aggregates: `rollupValueNumeric` (a decimal returned as a string to preserve precision), `rollupValueDate`, or `rollupValueText`. `rollupResultType` tells you which one is populated, and `rollupComputeStatus` reports whether the value is current.

```graphql
query RecordRollups {
  recordQueries {
    todos(filter: { companyIds: [], projectIds: ["project_123"] }, limit: 5) {
      items {
        id
        title
        customFields {
          name
          type
          rollupResultType
          rollupValueNumeric
          rollupValueDate
          rollupValueText
          rollupComputeStatus
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
            "title": "Acme Inc.",
            "customFields": [
              {
                "name": "Total Order Value",
                "type": "ROLLUP",
                "rollupResultType": "numeric",
                "rollupValueNumeric": "48000",
                "rollupValueDate": null,
                "rollupValueText": null,
                "rollupComputeStatus": "FRESH"
              }
            ]
          }
        ]
      }
    }
  }
}
```

## Notes

- `rollupComputeStatus` is one of `FRESH`, `COMPUTING`, `FAILED`, or `BROKEN_SOURCE`. A `BROKEN_SOURCE` status means the source or aggregate field was removed — re-point the rollup or recreate it.
- `rollupValueNumeric` is a string so large or fractional totals survive the JavaScript number boundary; parse it on your side.
- `COUNT` counts the reached records; `COUNTA` counts those with a non-empty value; `ARRAYJOIN` concatenates the text values.

## Related

- [Reference field](/api/custom-fields/reference)
- [Referenced by field](/api/custom-fields/referenced-by)
- [Lookup field](/api/custom-fields/lookup)
- [Custom fields overview](/api/custom-fields)
