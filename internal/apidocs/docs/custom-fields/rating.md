---
title: Rating Custom Field
description: Store a numeric rating on a record, with an optional lower bound and an upper bound that caps the scale.
icon: Star
order: 24
---

A rating custom field stores a numeric score on a record — a satisfaction rating, a priority level, an NPS response, or any value on a defined scale. It maps to the `RATING` value of the `CustomFieldType` enum and is backed by a `CustomField` whose value is exposed as a `Float`. It behaves like the [Number field](/api/custom-fields/number) but is conventionally shown as a star scale in the app. For an unbounded number use the [Number field](/api/custom-fields/number); for percentages use the [Percent field](/api/custom-fields/percent).

Records are `Record` objects in the API.

## Overview

- **Type enum:** `RATING`
- **Optional config:** `min` / `max` bounds (`Float`). `max` defines the top of the scale (e.g. 5 for a 5-star rating).
- **Value shape:** a single `Float`. On a record, read it from `CustomField.number` (or `CustomField.value`, which returns the same bare number for this type).
- **Bounds enforcement:** enforced when you create a record with `createRecord` — `min` defaults to **0** if unset, and `max` is enforced only when set (otherwise the value is unbounded above). Out-of-range values throw `CUSTOM_FIELD_VALUE_PARSE_ERROR`. Bounds are **not** enforced by `setRecordCustomField` — see [Notes](#notes).

## Create

Use the `createCustomField` mutation with `type: RATING`. The field is created in the workspace named by your `blue-workspace-id` header — there is no `projectId` argument. Set `max` to define the top of the scale.

```graphql
mutation CreateRatingField {
  createCustomField(input: { name: "Performance Rating", type: RATING, max: 5 }) {
    id
    name
    type
    max
  }
}
```

Add `min`, `max`, and `description` to define a fuller scale:

```graphql
mutation CreateDetailedRatingField {
  createCustomField(
    input: {
      name: "Customer Satisfaction"
      type: RATING
      min: 1
      max: 10
      description: "Rate customer satisfaction from 1 to 10"
    }
  ) {
    id
    name
    type
    min
    max
    description
  }
}
```

```json
{
  "data": {
    "createCustomField": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Customer Satisfaction",
      "type": "RATING",
      "min": 1,
      "max": 10,
      "description": "Rate customer satisfaction from 1 to 10"
    }
  }
}
```

### CreateCustomFieldInput

| Parameter     | Type               | Required | Description                                                                    |
| ------------- | ------------------ | -------- | ------------------------------------------------------------------------------ |
| `name`        | `String!`          | Yes      | Display name of the field.                                                     |
| `type`        | `CustomFieldType!` | Yes      | Must be `RATING`.                                                              |
| `min`         | `Float`            | No       | Lower bound. Enforced by `createRecord`; defaults to `0` when a value is parsed. |
| `max`         | `Float`            | No       | Upper bound / top of the scale. Enforced by `createRecord`; unbounded if unset.  |
| `description` | `String`           | No       | Help text shown to users.                                                      |

The workspace is taken from the `blue-workspace-id` header (ID or slug); `CreateCustomFieldInput` has no `projectId` field.

## Set a value

Use the `setRecordCustomField` mutation with the `number` argument — a rating is numeric, so it is set through `number`, not a generic `value`. It returns `Boolean!` — `true` on success — so it has no sub-selection.

```graphql
mutation SetRatingValue {
  setRecordCustomField(input: { todoId: "todo_123", customFieldId: "field_123", number: 4.5 })
}
```

```json
{ "data": { "setRecordCustomField": true } }
```

### SetRecordCustomFieldInput

| Parameter       | Type      | Required | Description                                             |
| --------------- | --------- | -------- | ------------------------------------------------------- |
| `todoId`        | `String!` | Yes      | ID of the record to update.                             |
| `customFieldId` | `String!` | Yes      | ID of the rating field.                                 |
| `number`        | `Float`   | No       | The rating to store. Omit (or send `null`) to clear it. |

## Read a value

`setRecordCustomField` returns only a boolean, so read the stored value back through the record's `customFields` connection. `Record.customFields` returns `[CustomField!]!` directly — select the value fields on the element. For a `RATING` field, both `number` and `value` resolve to the bare numeric value.

```graphql
query ReadRatingValue {
  recordQueries {
    todos(filter: { companyIds: ["company_123"], todoIds: ["todo_123"] }) {
      items {
        id
        title
        customFields {
          id
          name
          type
          min
          max
          number
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
                "id": "clm4n8qwx000108l0a1b2c3d4",
                "name": "Customer Satisfaction",
                "type": "RATING",
                "min": 1,
                "max": 10,
                "number": 4.5,
                "value": 4.5
              }
            ]
          }
        ]
      }
    }
  }
}
```

The `number` and `value` fields resolve only when the `CustomField` is read in a record context (via `Record.customFields`); they are `null` when the field definition is read on its own. If no value is set, both return `null`.

## Set a value at record creation

`createRecord` accepts custom-field values inline through `CreateRecordInputCustomField`, which carries the value as a **string** in the `value` field (there is no `number` argument here). The string is parsed to a number, and bounds are enforced — a value below `min` (default `0`), above `max`, or non-numeric throws `CUSTOM_FIELD_VALUE_PARSE_ERROR`.

```graphql
mutation CreateRecordWithRating {
  createRecord(
    input: {
      title: "Review customer feedback"
      todoListId: "list_123"
      customFields: [{ customFieldId: "field_123", value: "4.5" }]
    }
  ) {
    id
    title
    customFields {
      name
      type
      number
    }
  }
}
```

## Notes

- **Bounds are enforced only at record creation.** `createRecord` parses the inline string value and rejects anything below `min`, above `max` (when set), or non-numeric with `CUSTOM_FIELD_VALUE_PARSE_ERROR`. If you never set `min`, it defaults to `0` at parse time; if you never set `max`, the value is unbounded above. `setRecordCustomField` writes the `number` argument straight to storage with no bounds check — a value outside `min`/`max` is accepted and stored. Enforce the scale client-side when calling `setRecordCustomField`.
- **No default scale on creation.** Creating a `RATING` field without `min`/`max` leaves both unset on the definition. There is no implicit 0–5 scale; set `max` to define the top of the scale you want.
- **Filtering by value** is done through the `fields` JSON of `TodosFilter` (a `CUSTOM_FIELD` entry with `customFieldType: "RATING"`), not through per-field operators. See [List records](/api/records/list-records) for the full `fields` filter shape.

```graphql
query FilterByRating {
  recordQueries {
    todos(
      filter: {
        companyIds: ["company_123"]
        projectIds: ["project_123"]
        fields: [
          {
            type: "CUSTOM_FIELD"
            customFieldId: "field_123"
            customFieldType: "RATING"
            values: ["4"]
            op: "GTE"
          }
        ]
      }
    ) {
      items {
        id
        title
      }
    }
  }
}
```

## Errors

| Code                             | When                                                                              |
| -------------------------------- | --------------------------------------------------------------------------------- |
| `CUSTOM_FIELD_VALUE_PARSE_ERROR` | A `createRecord` inline value is non-numeric, below `min`, or above `max`.          |
| `CUSTOM_FIELD_NOT_FOUND`         | `customFieldId` does not match a field in the active workspace.                   |
| `TODO_NOT_FOUND`                 | The `todoId` passed to `setRecordCustomField` does not match a record you can edit. |
| `FORBIDDEN`                      | The caller lacks permission for the operation (see Permissions).                  |

## Permissions

- **Create the field:** `createCustomField` requires the `OWNER` or `ADMIN` role on the workspace.
- **Set a value:** `setRecordCustomField` requires an authenticated caller with edit access to the record (any editing company role, or a custom project role granting edit on the field). `VIEW_ONLY` and `COMMENT_ONLY` callers are rejected.

## Related

- [Set custom field values](/api/custom-fields/custom-field-values)
- [Number field](/api/custom-fields/number)
- [Percent field](/api/custom-fields/percent)
- [Single-select field](/api/custom-fields/select-single)
- [List records](/api/records/list-records)
- [Custom Fields overview](/api/custom-fields)
