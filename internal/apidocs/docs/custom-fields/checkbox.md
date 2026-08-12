---
title: Checkbox Field
description: Store a single boolean (checked/unchecked) value on records for flags, approvals, and confirmations.
icon: CheckSquare
order: 9
---

A checkbox custom field stores a single boolean value on a record — checked, unchecked, or unset. Use it for binary flags such as approvals, sign-offs, and confirmations. Checkbox fields are the `CHECKBOX` value of the `CustomFieldType` enum. Custom fields are `CustomField` objects; records are `Record` objects.

## Overview

A checkbox field holds one of three states per record: `true` (checked), `false` (unchecked), or `null` (never set). The value lives on the record, so the same field can be checked on one record and unset on another.

Checkbox fields are scoped to a single workspace by the `blue-workspace-id` request header — there is no `projectId` argument on the create input.

## Create

Use the `createCustomField` mutation with `type: CHECKBOX`. Set the workspace with the `blue-workspace-id` header.

```graphql
mutation CreateCheckboxField {
  createCustomField(input: { name: "Reviewed", type: CHECKBOX }) {
    id
    name
    type
  }
}
```

Add `description` for help text shown in the app:

```graphql
mutation CreateCheckboxFieldWithHelp {
  createCustomField(
    input: {
      name: "Customer Approved"
      type: CHECKBOX
      description: "Check this once the customer signs off on the work."
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

| Parameter     | Type               | Required | Description                                   |
| ------------- | ------------------ | -------- | --------------------------------------------- |
| `name`        | `String!`          | Yes      | Display name of the field.                    |
| `type`        | `CustomFieldType!` | Yes      | Must be `CHECKBOX`.                           |
| `description` | `String`           | No       | Help text shown beneath the field in the app. |

### Response

```json
{
  "data": {
    "createCustomField": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Reviewed",
      "type": "CHECKBOX"
    }
  }
}
```

## Set a value

Use the `setRecordCustomField` mutation with the `checked` argument. The mutation returns `Boolean!` (`true` on success) — it does not return the updated record, so do not select subfields on it.

```graphql
mutation CheckTheBox {
  setRecordCustomField(input: { todoId: "todo_123", customFieldId: "field_123", checked: true })
}
```

Pass `checked: false` to uncheck:

```graphql
mutation UncheckTheBox {
  setRecordCustomField(input: { todoId: "todo_123", customFieldId: "field_123", checked: false })
}
```

### SetRecordCustomFieldInput

Only the fields relevant to a checkbox are shown; the input is shared across all field types.

| Parameter       | Type      | Required | Description                          |
| --------------- | --------- | -------- | ------------------------------------ |
| `todoId`        | `String!` | Yes      | ID of the record to update.          |
| `customFieldId` | `String!` | Yes      | ID of the checkbox field.            |
| `checked`       | `Boolean` | No       | `true` to check, `false` to uncheck. |

### Response

```json
{
  "data": {
    "setRecordCustomField": true
  }
}
```

To read the value back after setting it, query the record (see [Read a value](#read-a-value)).

## Set a value at record creation

Use the `createRecord` mutation to set a checkbox while creating the record. Inside `customFields`, the value is passed as a **string** (`value: "true"`) — `createRecord` takes a single `value` string per field, not the typed `checked` boolean used by `setRecordCustomField`.

```graphql
mutation CreateRecordWithCheckbox {
  createRecord(
    input: {
      title: "Review contract"
      todoListId: "list_123"
      customFields: [{ customFieldId: "field_123", value: "true" }]
    }
  ) {
    id
    title
    customFields {
      name
      type
      checked
    }
  }
}
```

The elements of `Record.customFields` are `CustomField` objects, not a junction type. Select the type-specific value field (`checked`) directly on the element.

### Accepted string values

At record creation the value string is matched **exactly and case-sensitively**: the string must be `"true"`, `"1"`, or `"checked"` to produce a checked state. Any other value (including `"True"`, `"yes"`, or `"false"`) is stored as unchecked.

| String value  | Result    |
| ------------- | --------- |
| `"true"`      | Checked   |
| `"1"`         | Checked   |
| `"checked"`   | Checked   |
| anything else | Unchecked |

<Callout variant="info" title="String matching is exact only at creation">

The case-sensitive `"true"` / `"1"` / `"checked"` parsing applies only to the `value` string on `createRecord`. The `setRecordCustomField` mutation takes a real `Boolean` (`checked`), so there is no string parsing to worry about there.

</Callout>

## Read a value

`Record.customFields` returns `[CustomField!]!` — a list of `CustomField` objects, one per field set on the record. There is no `RecordCustomField` wrapper. Select the `CustomField` fields you need on each element; for a checkbox, `checked` is the typed boolean read.

```graphql
query RecordCheckboxValues {
  recordQueries {
    todos(filter: { companyIds: ["company_123"], todoListIds: ["list_123"] }) {
      items {
        id
        title
        customFields {
          name
          type
          checked
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
            "title": "Review contract",
            "customFields": [
              {
                "name": "Reviewed",
                "type": "CHECKBOX",
                "checked": true,
                "value": { "checked": true }
              }
            ]
          }
        ]
      }
    }
  }
}
```

| Field     | Type               | Description                                                                                                                          |
| --------- | ------------------ | ------------------------------------------------------------------------------------------------------------------------------------ |
| `name`    | `String!`          | Display name of the field.                                                                                                           |
| `type`    | `CustomFieldType!` | Always `CHECKBOX` for this field.                                                                                                    |
| `checked` | `Boolean`          | The boolean state: `true`, `false`, or `null` if never set. The recommended read for a checkbox.                                     |
| `value`   | `JSON`             | The raw stored value, resolved only when the field is read in a record context. For a checkbox use `checked` — it's the typed shape. |

## Notes

- A checkbox stores one value per record. There is no tri-state beyond `true` / `false` / `null` (unset), and no per-field default — a checkbox is `null` until first set.
- Toggling a checkbox fires automation triggers: checking it (`null`/`false` → `true`) behaves like a value being added, and unchecking it (`true` → `false`) behaves like a value being removed. Build automations that react to these in the [Automations API](/api/automations/create-automation). (Trigger behavior is internal to the resolver and not exposed in the schema.)

## Errors

| Code                     | When                                                                        |
| ------------------------ | --------------------------------------------------------------------------- |
| `CUSTOM_FIELD_NOT_FOUND` | The `customFieldId` does not exist or is not in a workspace you can access. |
| `TODO_NOT_FOUND`         | The `todoId` does not exist or you lack access to it.                       |
| `BAD_USER_INPUT`         | The field does not belong to the record's workspace.                        |
| `FORBIDDEN`              | Your custom project role does not grant edit access to this field.          |

## Permissions

| Action                   | Required access                                                                                                                                   |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| Create or edit the field | Company role `OWNER` or `ADMIN`.                                                                                                                  |
| Set a value on a record  | `OWNER`, `ADMIN`, `MEMBER`, or `CLIENT` — `VIEW_ONLY` and `COMMENT_ONLY` are denied. Custom project roles need edit access to the specific field. |
| Read a value             | Any role with record view access in the workspace.                                                                                                |

## Related

- [Set custom field values](/api/custom-fields/custom-field-values) — the shared `setRecordCustomField` reference across all field types.
- [Create a custom field](/api/custom-fields/create-custom-fields) — the full `createCustomField` reference.
- [Single-select field](/api/custom-fields/select-single) — for more than two mutually exclusive options.
- [Custom Fields overview](/api/custom-fields) — all field types and concepts.
