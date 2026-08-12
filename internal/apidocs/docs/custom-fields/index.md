---
title: Custom Fields
description: Extend records with typed custom fields - 26 field types, field-level access control, and the operations to create, read, and write them.
icon: Settings
order: 0
---

Custom fields add typed, structured data to your records beyond the built-in title, dates, and assignees. A custom field is a `CustomField` object scoped to a single workspace (a `Workspace` in the API); once created, it appears on every record (`Record`) in that workspace, and each record holds its own value.

This section covers the operations to manage field definitions and write values, the full catalog of field types, and the access model that governs who can see and edit each field.

```graphql
POST https://api.blue.app/graphql
blue-token-id: YOUR_TOKEN_ID
blue-token-secret: YOUR_TOKEN_SECRET
blue-org-id: YOUR_ORG_ID
blue-workspace-id: project_123
```

The `X-Bloo-Project-ID` header (also accepted as `X-Bloo-Workspace-ID`) scopes custom field reads and writes to one workspace. Company and project headers accept either an ID or a slug. Header names are case-insensitive.

## Operations

| Operation                                                        | GraphQL                       | Description                                                                                                                                                              |
| ---------------------------------------------------------------- | ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| [List custom fields](/api/custom-fields/list-custom-fields)      | `customFields` query          | Retrieve a paginated list of field definitions, filtered by project and type.                                                                                            |
| [Create a custom field](/api/custom-fields/create-custom-fields) | `createCustomField` mutation  | Add a typed field to a workspace.                                                                                                                                        |
| Edit a custom field                                              | `editCustomField` mutation    | Change a field's name, configuration, or access rules. Takes `EditCustomFieldInput` (a `customFieldId` plus the fields to change) and returns the updated `CustomField`. |
| [Set field values](/api/custom-fields/custom-field-values)       | `setRecordCustomField` mutation | Write a value to a field on one record.                                                                                                                                  |
| [Delete a custom field](/api/custom-fields/delete-custom-field)  | `deleteCustomField` mutation  | Remove a field and all of its values.                                                                                                                                    |

## Field types

Every custom field has a `type` drawn from the `CustomFieldType` enum. The type fixes how values are stored, validated, and read back. Each type below links to its own reference page with create and set-value examples.

### Text

| Type          | Stores                                 | Page                                               |
| ------------- | -------------------------------------- | -------------------------------------------------- |
| `TEXT_SINGLE` | Single-line text                       | [Single-line text](/api/custom-fields/text-single) |
| `TEXT_MULTI`  | Multi-line text                        | [Multi-line text](/api/custom-fields/text-multi)   |
| `EMAIL`       | Email address                          | [Email](/api/custom-fields/email)                  |
| `URL`         | Web link, optionally shown as a button | [URL](/api/custom-fields/url)                      |
| `PHONE`       | Phone number with country code         | [Phone](/api/custom-fields/phone)                  |

### Selection

| Type            | Stores                               | Page                                              |
| --------------- | ------------------------------------ | ------------------------------------------------- |
| `SELECT_SINGLE` | One option from a defined list       | [Single-select](/api/custom-fields/select-single) |
| `SELECT_MULTI`  | Multiple options from a defined list | [Multi-select](/api/custom-fields/select-multi)   |
| `CHECKBOX`      | A boolean                            | [Checkbox](/api/custom-fields/checkbox)           |

### Numeric

| Type                  | Stores                                 | Page                                                          |
| --------------------- | -------------------------------------- | ------------------------------------------------------------- |
| `NUMBER`              | A number                               | [Number](/api/custom-fields/number)                           |
| `CURRENCY`            | A monetary amount in a fixed currency  | [Currency](/api/custom-fields/currency)                       |
| `CURRENCY_CONVERSION` | An amount converted between currencies | [Currency conversion](/api/custom-fields/currency-conversion) |
| `PERCENT`             | A percentage                           | [Percent](/api/custom-fields/percent)                         |
| `RATING`              | A star rating on a custom scale        | [Rating](/api/custom-fields/rating)                           |
| `FORMULA`             | A value computed from other fields     | [Formula](/api/custom-fields/formula)                         |

### Date and time

| Type            | Stores                      | Page                                              |
| --------------- | --------------------------- | ------------------------------------------------- |
| `DATE`          | A date or date range        | [Date](/api/custom-fields/date)                   |
| `TIME_DURATION` | An elapsed-time measurement | [Time duration](/api/custom-fields/time-duration) |

### Location

| Type       | Stores                                  | Page                                    |
| ---------- | --------------------------------------- | --------------------------------------- |
| `LOCATION` | A geographic point (latitude/longitude) | [Location](/api/custom-fields/location) |
| `COUNTRY`  | One or more country codes               | [Country](/api/custom-fields/country)   |

### People

| Type       | Stores                                   | Page                                    |
| ---------- | ---------------------------------------- | --------------------------------------- |
| `ASSIGNEE` | One or more users assigned via the field | [Assignee](/api/custom-fields/assignee) |

### Relationships

| Type            | Stores                                               | Page                                              |
| --------------- | ---------------------------------------------------- | ------------------------------------------------- |
| `REFERENCE`     | Links to records in another workspace                | [Reference](/api/custom-fields/reference)         |
| `REFERENCED_BY` | Records in another workspace that reference this one | [Referenced by](/api/custom-fields/referenced-by) |
| `LOOKUP`        | A value pulled from a referenced record              | [Lookup](/api/custom-fields/lookup)               |
| `ROLLUP`        | An aggregate over referenced or referencing records  | [Rollup](/api/custom-fields/rollup)               |

### Files, identifiers, and actions

| Type        | Stores                                 | Page                                      |
| ----------- | -------------------------------------- | ----------------------------------------- |
| `FILE`      | File attachments                       | [File](/api/custom-fields/file)           |
| `UNIQUE_ID` | An auto-generated identifier           | [Unique ID](/api/custom-fields/unique-id) |
| `BUTTON`    | An action trigger rendered as a button | [Button](/api/custom-fields/button)       |

<Callout variant="info" title="Computed types are read-only">

`FORMULA`, `LOOKUP`, `ROLLUP`, and `REFERENCED_BY` fields are computed by Blue. You read their values like any other field, but `setRecordCustomField` rejects them with a `BAD_USER_INPUT` error — they have no value to set directly.

</Callout>

## Reading field values

A record exposes its fields through `Record.customFields`, which returns `[CustomField!]!` — a flat list of `CustomField` objects, each already carrying the value for that record. Select the value field that matches the type directly on the array element; there is no wrapper object.

Records are read through `recordQueries`. The `companyIds` filter field is required by the schema but may be an empty array — pass `[]` to fall back to the company in the `X-Bloo-Company-ID` header.

```graphql
query RecordWithCustomFields {
  recordQueries {
    todos(filter: { companyIds: [], projectIds: ["project_123"] }, limit: 5) {
      items {
        id
        title
        customFields {
          id
          name
          type
          text # TEXT_SINGLE, TEXT_MULTI, EMAIL, URL, PHONE
          number # NUMBER, CURRENCY, PERCENT, RATING
          checked # CHECKBOX
          selectedOption {
            title
          } # SELECT_SINGLE
          selectedOptions {
            title
          } # SELECT_MULTI
          startDate # DATE (single date or range start)
          endDate # DATE range end
          value # JSON fallback for any type
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
                "id": "clm4paopt000108l0d2b7f1zr",
                "name": "Priority",
                "type": "SELECT_SINGLE",
                "text": null,
                "number": null,
                "checked": null,
                "selectedOption": { "title": "High" },
                "selectedOptions": null,
                "startDate": null,
                "endDate": null,
                "value": "High"
              },
              {
                "id": "clm4pb12x000208l0h9c3e8ws",
                "name": "Contract value",
                "type": "CURRENCY",
                "text": null,
                "number": 24000,
                "checked": null,
                "selectedOption": null,
                "selectedOptions": null,
                "startDate": null,
                "endDate": null,
                "value": 24000
              }
            ]
          }
        ]
      }
    }
  }
}
```

The `value` field is a `JSON` scalar that returns the field's value in a type-appropriate shape — useful when you don't want to branch on `type`. For per-type details, see each field type's page.

## Writing field values

Use `setRecordCustomField` to write a value to one field on one record. It is an upsert and returns `Boolean!` — it has no selection set, so do not query subfields on the result. Read the value back with the `Record.customFields` query shown above.

```graphql
mutation SetPriority {
  setRecordCustomField(
    input: {
      todoId: "todo_123"
      customFieldId: "field_123"
      customFieldOptionId: "option_123" # SELECT_SINGLE
    }
  )
}
```

```json
{ "data": { "setRecordCustomField": true } }
```

The input parameter you send depends on the field type — `text`, `number`, `checked`, `customFieldOptionId`/`customFieldOptionIds`, `startDate`/`endDate`, `countryCodes`, `latitude`/`longitude`, `assigneeUserIds`, or `customFieldReferenceTodoIds`. See [Set field values](/api/custom-fields/custom-field-values) for the full mapping.

To seed values when a record is first created, pass `customFields` to `createRecord`. Each entry is a `CreateRecordInputCustomField` of `{ customFieldId, value }`, where `value` is a string the API parses according to the field's type.

```graphql
mutation CreateRecordWithFields {
  createRecord(
    input: {
      todoListId: "list_123"
      title: "New deal"
      customFields: [
        { customFieldId: "field_123", value: "option_123" }
        { customFieldId: "field_456", value: "24000" }
      ]
    }
  ) {
    id
    title
    customFields {
      name
      type
      value
    }
  }
}
```

## Access control

Visibility and edit rights for a field follow the caller's role in the workspace, refined by optional per-field access rules.

### Workspace roles

A user's access level in a workspace is one of the `UserAccessLevel` values:

| Access level   | Can see field values   | Can set field values   | Can create/edit/delete fields |
| -------------- | ---------------------- | ---------------------- | ----------------------------- |
| `OWNER`        | Yes                    | Yes                    | Yes                           |
| `ADMIN`        | Yes                    | Yes                    | Yes                           |
| `MEMBER`       | Yes                    | Yes                    | No                            |
| `COMMENT_ONLY` | Yes                    | No                     | No                            |
| `VIEW_ONLY`    | Yes                    | No                     | No                            |
| `CLIENT`       | Subject to field rules | Subject to field rules | No                            |

### Field-level rules

When creating or editing a field, you can restrict it to specific users, custom roles, or access levels:

| Input field      | Type                 | Description                                       |
| ---------------- | -------------------- | ------------------------------------------------- |
| `allowedUserIds` | `[String!]`          | Restrict the field to these users.                |
| `allowedRoleIds` | `[String!]`          | Restrict the field to these custom project roles. |
| `allowedLevels`  | `[UserAccessLevel!]` | Restrict the field to these access levels.        |

These rules drive the per-caller booleans on `CustomField`: `viewable` (the caller may read this field) and `editable` (the caller may set values on it). `workspaceUserRole` returns the custom role resolved for the caller. A field with no rules is governed entirely by the workspace role table above.

## Errors

These are the error codes the custom field operations return.

| Code                             | When                                                                                        |
| -------------------------------- | ------------------------------------------------------------------------------------------- |
| `CUSTOM_FIELD_NOT_FOUND`         | The `customFieldId` does not exist or is not in the header's workspace.                     |
| `PROJECT_NOT_FOUND`              | A referenced project (for `REFERENCE`/`REFERENCED_BY`/`ROLLUP`) does not exist.             |
| `BAD_USER_INPUT`                 | The value is malformed, or the target field is a computed type that cannot be set directly. |
| `CUSTOM_FIELD_VALUE_PARSE_ERROR` | A value passed to `createRecord`'s `customFields` cannot be parsed for the field's type.      |
| `FORBIDDEN`                      | The caller's role or the field's access rules forbid the read or write.                     |
| `CUSTOM_FIELD_LIMIT`             | The workspace has reached its custom field limit (upgrade to add more).                     |
| `PRO_REQUIRED`                   | The field type requires a Pro subscription (for example, `ROLLUP`).                         |
| `FIELD_REFERENCED_BY_FORMULA`    | The field cannot be deleted because the workspace's name formula uses it.                   |

<Callout variant="warning" title="metadata is deprecated">

`CustomField.metadata` (and the matching `metadata` input on create/edit) is `@deprecated` and no longer used. Do not write configuration into it; use the typed configuration fields documented on each field type's page.

</Callout>

## Related

- [Create a custom field](/api/custom-fields/create-custom-fields)
- [List custom fields](/api/custom-fields/list-custom-fields)
- [Set field values](/api/custom-fields/custom-field-values)
- [Delete a custom field](/api/custom-fields/delete-custom-field)
- [Records](/api/records) — the records that carry custom field values
- [Workspaces](/api/workspaces) — the workspaces custom fields are scoped to
