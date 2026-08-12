---
title: Set Custom Field Values
description: Write values to custom fields on a record with the type-specific parameter each field expects.
icon: FileEdit
order: 3
---

Use the `setRecordCustomField` mutation to write a value to a custom field on a single record. The mutation is an upsert: it creates the value if the field is empty on that record, or overwrites the existing value. Records are `Record` objects in the API; custom fields are `CustomField` objects, scoped to a workspace.

Each field type reads a different parameter from the input (text, number, option IDs, assignee IDs, and so on). Send only the parameter that matches the field's `type` — see the [Value reference](#value-reference) below. To set values on many records at once, use `bulkSetCustomField` instead.

## Request

`setRecordCustomField` takes a single `SetRecordCustomFieldInput` and returns `Boolean!`.

```graphql
mutation SetTextField {
  setRecordCustomField(
    input: {
      todoId: "todo_123"
      customFieldId: "field_123"
      text: "Project specification document"
    }
  )
}
```

Send these headers with every request. `blue-workspace-id` accepts a workspace ID or slug.

```
blue-token-id: YOUR_TOKEN_ID
blue-token-secret: YOUR_TOKEN_SECRET
blue-org-id: YOUR_ORG_ID
blue-workspace-id: project_123
```

## Parameters

### SetRecordCustomFieldInput

`todoId` and `customFieldId` are always required. Beyond those, supply only the value parameter that matches the field's type.

| Parameter                     | Type        | Required | Description                                                                         |
| ----------------------------- | ----------- | -------- | ----------------------------------------------------------------------------------- |
| `todoId`                      | `String!`   | Yes      | The record to write to.                                                             |
| `customFieldId`               | `String!`   | Yes      | The custom field to set. Must belong to the same workspace as the record.           |
| `text`                        | `String`    | No       | Value for `TEXT_SINGLE`, `TEXT_MULTI`, `PHONE`, `EMAIL`, `URL`.                     |
| `number`                      | `Float`     | No       | Value for `NUMBER`, `PERCENT`, `RATING`, `CURRENCY`, `TIME_DURATION`.               |
| `currency`                    | `String`    | No       | ISO 4217 currency code for a `CURRENCY` field (e.g. `"USD"`), paired with `number`. |
| `checked`                     | `Boolean`   | No       | Value for `CHECKBOX`.                                                               |
| `startDate`                   | `DateTime`  | No       | Start of a `DATE` field (or the single date).                                       |
| `endDate`                     | `DateTime`  | No       | End of a `DATE` range.                                                              |
| `timezone`                    | `String`    | No       | IANA timezone for a `DATE` field (e.g. `"America/New_York"`).                       |
| `latitude`                    | `Float`     | No       | Latitude for a `LOCATION` field, paired with `longitude`.                           |
| `longitude`                   | `Float`     | No       | Longitude for a `LOCATION` field.                                                   |
| `regionCode`                  | `String`    | No       | ISO region code for a `PHONE` field (e.g. `"US"`).                                  |
| `countryCodes`                | `[String!]` | No       | ISO 3166 country codes for a `COUNTRY` field.                                       |
| `customFieldOptionId`         | `String`    | No       | Selected option for a `SELECT_SINGLE` field.                                        |
| `customFieldOptionIds`        | `[String!]` | No       | Selected options for a `SELECT_MULTI` field.                                        |
| `customFieldReferenceTodoIds` | `[String!]` | No       | Linked record IDs for a `REFERENCE` field.                                          |
| `assigneeUserIds`             | `[String!]` | No       | User IDs for an `ASSIGNEE` field.                                                   |

## Response

The mutation returns `true` on success.

```json
{
  "data": {
    "setRecordCustomField": true
  }
}
```

To read the value back, query the record's `customFields`. `Record.customFields` returns `[CustomField!]!` directly — the value lives on each element, with no wrapper object.

```graphql
query ReadFieldValue {
  recordQueries {
    todos(filter: { companyIds: ["company_123"], todoIds: ["todo_123"] }) {
      items {
        id
        customFields {
          id
          name
          type
          text
          number
          checked
          selectedOption {
            title
          }
          value
        }
      }
    }
  }
}
```

## Field type examples

### Number

```graphql
mutation SetNumber {
  setRecordCustomField(input: { todoId: "todo_123", customFieldId: "field_123", number: 15000.5 })
}
```

### Single-select and multi-select

Pass `customFieldOptionId` for a `SELECT_SINGLE` field and `customFieldOptionIds` for a `SELECT_MULTI` field.

```graphql
mutation SetSelect {
  single: setRecordCustomField(
    input: { todoId: "todo_123", customFieldId: "field_123", customFieldOptionId: "option_123" }
  )
  multi: setRecordCustomField(
    input: {
      todoId: "todo_123"
      customFieldId: "field_456"
      customFieldOptionIds: ["option_123", "option_456"]
    }
  )
}
```

### Date and date range

A single date uses `startDate` only; a range adds `endDate`.

```graphql
mutation SetDateRange {
  setRecordCustomField(
    input: {
      todoId: "todo_123"
      customFieldId: "field_123"
      startDate: "2026-01-01T00:00:00Z"
      endDate: "2026-03-31T23:59:59Z"
      timezone: "America/New_York"
    }
  )
}
```

### Assignee

`ASSIGNEE` fields take `assigneeUserIds`. Every user must be eligible for the field (an allowed user, role, or access level), and a single-assignee field rejects more than one ID.

```graphql
mutation SetAssignee {
  setRecordCustomField(
    input: {
      todoId: "todo_123"
      customFieldId: "field_123"
      assigneeUserIds: ["user_123", "user_456"]
    }
  )
}
```

### Location

```graphql
mutation SetLocation {
  setRecordCustomField(
    input: {
      todoId: "todo_123"
      customFieldId: "field_123"
      latitude: 37.7749
      longitude: -122.4194
    }
  )
}
```

### Currency

A `CURRENCY` field stores an amount (`number`) and a currency code (`currency`).

```graphql
mutation SetCurrency {
  setRecordCustomField(
    input: { todoId: "todo_123", customFieldId: "field_123", number: 5000, currency: "USD" }
  )
}
```

### Reference

`REFERENCE` fields link to records in the field's target workspace. Send the full set of linked record IDs each time — IDs you omit are unlinked.

```graphql
mutation SetReference {
  setRecordCustomField(
    input: {
      todoId: "todo_123"
      customFieldId: "field_123"
      customFieldReferenceTodoIds: ["todo_456", "todo_789"]
    }
  )
}
```

## Clearing a value

`setRecordCustomField` reconciles option, reference, and assignee links by comparing the IDs you send against what is already stored — anything you omit is removed. To clear a multi-value field, re-send it with an empty array:

```graphql
mutation ClearAssignees {
  setRecordCustomField(input: { todoId: "todo_123", customFieldId: "field_123", assigneeUserIds: [] })
}
```

For scalar fields (`text`, `number`, `checked`), send the empty/false value to reset it.

## Files

File fields use their own mutations rather than `setRecordCustomField`. Upload the file first (see [Upload files](/api/files)) to obtain a `fileUid`, then attach it. Both mutations return `Boolean`.

```graphql
mutation AttachFile {
  createRecordCustomFieldFile(
    input: { todoId: "todo_123", customFieldId: "field_123", fileUid: "file_123" }
  )
}

mutation DetachFile {
  deleteRecordCustomFieldFile(
    input: { todoId: "todo_123", customFieldId: "field_123", fileUid: "file_123" }
  )
}
```

### CreateRecordCustomFieldFileInput / DeleteRecordCustomFieldFileInput

Both inputs share the same three required fields.

| Parameter       | Type      | Required | Description                                    |
| --------------- | --------- | -------- | ---------------------------------------------- |
| `todoId`        | `String!` | Yes      | The record holding the `FILE` custom field.    |
| `customFieldId` | `String!` | Yes      | The `FILE` custom field.                       |
| `fileUid`       | `String!` | Yes      | The uploaded file's UID, from the upload step. |

## Setting values during record creation

`createRecord` accepts a `customFields` array so you can seed values when the record is created. Each entry is a `CreateRecordInputCustomField` with a `customFieldId` and a stringified `value`.

```graphql
mutation CreateRecordWithFields {
  createRecord(
    input: {
      todoListId: "list_123"
      title: "New Feature Development"
      customFields: [
        { customFieldId: "field_123", value: "high" }
        { customFieldId: "field_456", value: "8" }
      ]
    }
  ) {
    id
    title
    customFields {
      id
      name
      value
    }
  }
}
```

For anything beyond simple scalar seeding (options, references, assignees, dates), create the record first, then call `setRecordCustomField` per field with the type-specific parameters above.

## Errors

| Code                     | When                                                                                                                                                                                                          |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `CUSTOM_FIELD_NOT_FOUND` | The `customFieldId` does not exist or is not in a workspace you can access.                                                                                                                                   |
| `TODO_NOT_FOUND`         | The `todoId` does not exist or is not visible to you.                                                                                                                                                          |
| `FORBIDDEN`              | You lack edit access to the record, or your custom role marks this field as non-editable.                                                                                                                      |
| `BAD_USER_INPUT`         | The field is computed (`FORMULA`, `LOOKUP`, `ROLLUP`, `REFERENCED_BY`), the field belongs to a different workspace than the record, an option/reference/assignee ID is invalid, or an assignee is ineligible. |

```json
{
  "errors": [
    {
      "message": "ROLLUP fields are computed and cannot be set directly.",
      "extensions": { "code": "BAD_USER_INPUT" }
    }
  ]
}
```

## Permissions

Writing a value requires edit access to the record. Members and clients can set values unless a custom role restricts them. When the caller has a custom project role, the API runs a two-tier check: it confirms workspace access, then confirms the specific field is marked editable in that role. A user can therefore have general workspace access yet be blocked from editing a particular field.

## Value reference

Send the parameter listed for the field's `type`. Computed types reject writes with `BAD_USER_INPUT`.

| Field type            | Parameter                             | Notes                                                          |
| --------------------- | ------------------------------------- | -------------------------------------------------------------- |
| `TEXT_SINGLE`         | `text`                                | Plain string.                                                  |
| `TEXT_MULTI`          | `text`                                | String; newlines preserved.                                    |
| `NUMBER`              | `number`                              | Float.                                                         |
| `CURRENCY`            | `number` + `currency`                 | Amount plus ISO 4217 code.                                     |
| `PERCENT`             | `number`                              | Float.                                                         |
| `RATING`              | `number`                              | Within the field's `min`/`max`.                                |
| `TIME_DURATION`       | `number`                              | Duration value.                                                |
| `CHECKBOX`            | `checked`                             | Boolean.                                                       |
| `DATE`                | `startDate` (+ `endDate`, `timezone`) | ISO 8601; add `endDate` for a range.                           |
| `SELECT_SINGLE`       | `customFieldOptionId`                 | One option ID.                                                 |
| `SELECT_MULTI`        | `customFieldOptionIds`                | Array of option IDs.                                           |
| `PHONE`               | `text` (+ `regionCode`)               | E.164 string; `regionCode` for formatting.                     |
| `EMAIL`               | `text`                                | Email string.                                                  |
| `URL`                 | `text`                                | URL string.                                                    |
| `LOCATION`            | `latitude` + `longitude`              | Floats.                                                        |
| `COUNTRY`             | `countryCodes`                        | Array of ISO 3166 codes.                                       |
| `REFERENCE`           | `customFieldReferenceTodoIds`         | Array of record IDs in the target workspace.                   |
| `ASSIGNEE`            | `assigneeUserIds`                     | Array of user IDs; eligibility enforced.                       |
| `FILE`                | —                                     | Use `createRecordCustomFieldFile` / `deleteRecordCustomFieldFile`. |
| `BUTTON`              | —                                     | Triggers automations when clicked; not set via this mutation.  |
| `UNIQUE_ID`           | —                                     | System-generated; read-only.                                   |
| `FORMULA`             | —                                     | Computed from other fields; read-only.                         |
| `LOOKUP`              | —                                     | Pulled from a referenced record; read-only.                    |
| `ROLLUP`              | —                                     | Aggregated from linked records; read-only.                     |
| `REFERENCED_BY`       | —                                     | The inverse of a `REFERENCE` link; read-only.                  |
| `CURRENCY_CONVERSION` | —                                     | Computed from a referenced `CURRENCY` field; read-only.        |

`CURRENCY` does not auto-convert: it just stores an amount and a code. Conversion happens only when a separate `CURRENCY_CONVERSION` field references that `CURRENCY` field.

## Related

- [Create a Custom Field](/api/custom-fields/create-custom-fields)
- [List all Custom Fields](/api/custom-fields/list-custom-fields)
- [Delete Custom Field](/api/custom-fields/delete-custom-field)
- [Reference Field](/api/custom-fields/reference)
- [File Field](/api/custom-fields/file)
- [List Records](/api/records/list-records)
