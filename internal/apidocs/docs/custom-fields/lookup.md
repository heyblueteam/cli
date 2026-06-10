---
title: Lookup Field
description: Pull live data from records linked through a Reference field, with no manual copying.
icon: Search
order: 28
---

A Lookup field surfaces data from records that are linked through a [Reference field](/api/custom-fields/reference) — for example, showing the tags, assignees, or a specific custom field value of every linked record. Lookups are the `LOOKUP` value of the `CustomFieldType` enum. They are read-only and recompute automatically whenever the referenced data changes.

Custom fields are `CustomField` objects in the API. A Lookup always points at a `REFERENCE` (or `REFERENCED_BY`) source field on the same workspace and reads one kind of data from the records that source field resolves.

## Overview

A Lookup is configured with two things:

- **`referenceId`** — the ID of the source field to read from. This must be a `REFERENCE` or `REFERENCED_BY` field on the same workspace. You cannot point a Lookup at another Lookup.
- **`lookupType`** — which piece of data to pull from each linked record (tags, assignees, due dates, a custom field value, and so on).

When `lookupType` is `TODO_CUSTOM_FIELD`, you also pass **`lookupId`** — the ID of the specific custom field to read from each linked record.

## Create

Create a Lookup that surfaces the tags of every record linked through a Reference field. Custom fields are scoped by the `X-Bloo-Project-ID` header, so no workspace ID appears in the input.

```graphql
mutation CreateLookupField {
  createCustomField(
    input: {
      name: "Linked record tags"
      type: LOOKUP
      lookupOption: { referenceId: "field_123", lookupType: TODO_TAG }
    }
  ) {
    id
    name
    type
    customFieldLookupOption {
      lookupType
      reference {
        id
        name
      }
    }
  }
}
```

To read a specific custom field from each linked record, use `lookupType: TODO_CUSTOM_FIELD` and pass `lookupId` (the field to read):

```graphql
mutation CreateCustomFieldLookup {
  createCustomField(
    input: {
      name: "Linked budget"
      type: LOOKUP
      lookupOption: {
        referenceId: "field_123"
        lookupId: "field_456"
        lookupType: TODO_CUSTOM_FIELD
      }
    }
  ) {
    id
  }
}
```

### CreateCustomFieldInput

| Parameter      | Type                           | Required              | Description                                                                                                         |
| -------------- | ------------------------------ | --------------------- | ------------------------------------------------------------------------------------------------------------------- |
| `name`         | `String!`                      | Yes                   | Display name of the Lookup field.                                                                                   |
| `type`         | `CustomFieldType!`             | Yes                   | Must be `LOOKUP`.                                                                                                   |
| `lookupOption` | `CustomFieldLookupOptionInput` | Required for `LOOKUP` | Lookup configuration. Optional at the schema level, but `createCustomField` rejects a `LOOKUP` field that omits it. |
| `description`  | `String`                       | No                    | Help text shown to users.                                                                                           |

### CustomFieldLookupOptionInput

| Parameter     | Type                     | Required    | Description                                                                                           |
| ------------- | ------------------------ | ----------- | ----------------------------------------------------------------------------------------------------- |
| `referenceId` | `String!`                | Yes         | ID of the source `REFERENCE` or `REFERENCED_BY` field to read from.                                   |
| `lookupType`  | `CustomFieldLookupType!` | Yes         | Which data to pull from each linked record.                                                           |
| `lookupId`    | `String`                 | Conditional | ID of the custom field to read. Required when `lookupType` is `TODO_CUSTOM_FIELD`; ignored otherwise. |

### CustomFieldLookupType

| Value                | Reads from each linked record                                                                                                   |
| -------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| `TODO_DUE_DATE`      | The record's due date.                                                                                                          |
| `TODO_CREATED_AT`    | The record's creation timestamp.                                                                                                |
| `TODO_UPDATED_AT`    | The record's last-updated timestamp.                                                                                            |
| `TODO_TAG`           | The record's tags.                                                                                                              |
| `TODO_ASSIGNEE`      | The record's assigned users.                                                                                                    |
| `TODO_DESCRIPTION`   | The record's description text (blank values are skipped).                                                                       |
| `TODO_LIST`          | The list the record belongs to.                                                                                                 |
| `TODO_CUSTOM_FIELD`  | A specific custom field value, named by `lookupId`.                                                                             |
| `TODO_REFERENCED_BY` | **Deprecated.** Instead, create a `REFERENCED_BY` field and point a Lookup's `referenceId` at it. Existing usages keep working. |

## Read a value

A Lookup field has no stored value of its own — its result is computed in the context of the record it is rendered on. Read it from `Todo.customFields`, which returns `[CustomField!]!` directly (there is no wrapper type), and select `customFieldLookupOption` on the element:

```graphql
query GetLookupValues {
  todo(id: "todo_123") {
    customFields {
      id
      name
      type
      customFieldLookupOption {
        lookupType
        lookupResult
        reference {
          id
          name
        }
        lookup {
          id
          name
          type
        }
      }
    }
  }
}
```

<Callout variant="warning" title="Read lookupResult, not lookupValues">

`CustomFieldLookupOption.lookupValues` is deprecated and no longer populated. Always read the computed data from `lookupResult`.

</Callout>

### CustomField (Lookup fields)

| Field                     | Type                      | Description                               |
| ------------------------- | ------------------------- | ----------------------------------------- |
| `id`                      | `ID!`                     | Unique identifier for the field.          |
| `name`                    | `String!`                 | Display name of the Lookup field.         |
| `type`                    | `CustomFieldType!`        | Always `LOOKUP`.                          |
| `customFieldLookupOption` | `CustomFieldLookupOption` | Lookup configuration and computed result. |

### CustomFieldLookupOption

| Field               | Type                    | Description                                                               |
| ------------------- | ----------------------- | ------------------------------------------------------------------------- |
| `lookupType`        | `CustomFieldLookupType` | Which data this Lookup reads.                                             |
| `lookupResult`      | `JSON`                  | The data pulled from the linked records, in the current record's context. |
| `reference`         | `CustomField`           | The source Reference field.                                               |
| `lookup`            | `CustomField`           | The specific field being read (only for `TODO_CUSTOM_FIELD`).             |
| `parentCustomField` | `CustomField`           | The parent Lookup field.                                                  |
| `parentLookup`      | `CustomField`           | Parent Lookup in a chain.                                                 |
| `lookupValues`      | `JSON`                  | **Deprecated** — no longer populated. Use `lookupResult`.                 |

## Response

The shape of `lookupResult` depends on `lookupType`. Because it is a `JSON` field, the structures below are **illustrative** — they show typical output but are not enforced by the schema.

A `TODO_TAG` Lookup returns the linked records' tags (tags are `Tag` objects, keyed by `title`):

```json
{
  "data": {
    "todo": {
      "customFields": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "name": "Linked record tags",
          "type": "LOOKUP",
          "customFieldLookupOption": {
            "lookupType": "TODO_TAG",
            "lookupResult": [
              { "id": "tag_123", "title": "urgent", "color": "#ff0000" },
              { "id": "tag_456", "title": "blocked", "color": "#facc15" }
            ],
            "reference": { "id": "field_123", "name": "Linked records" },
            "lookup": null
          }
        }
      ]
    }
  }
}
```

A `TODO_ASSIGNEE` Lookup returns user objects (users expose `fullName` and `email`):

```json
{
  "data": {
    "todo": {
      "customFields": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "name": "Linked assignees",
          "type": "LOOKUP",
          "customFieldLookupOption": {
            "lookupType": "TODO_ASSIGNEE",
            "lookupResult": [
              { "id": "user_123", "fullName": "Ada Lovelace", "email": "ada@example.com" }
            ],
            "reference": { "id": "field_123", "name": "Linked records" },
            "lookup": null
          }
        }
      ]
    }
  }
}
```

For `TODO_CUSTOM_FIELD`, `lookupResult` mirrors the value shape of the field named by `lookupId` — e.g. a Currency field yields amount/currency pairs.

## Notes

- **Read-only.** You cannot set a Lookup value with `setTodoCustomField`; it always reflects the current linked data and recomputes when that data changes.
- **No aggregation.** A Lookup extracts the linked values as-is — it has no built-in sum, count, or average across linked records.
- **Source must be a Reference.** `referenceId` must point at a `REFERENCE` or `REFERENCED_BY` field. Pointing it at another `LOOKUP` is rejected — Lookup-of-Lookup chains are not supported.
- **Cross-workspace access.** A viewer only sees Lookup results for linked records in workspaces they have access to.

## Errors

| Code                     | When                                                                                                                                                          |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `CUSTOM_FIELD_NOT_FOUND` | `referenceId` (or `lookupId`) does not resolve to a field you can access.                                                                                     |
| `PROJECT_NOT_FOUND`      | The referenced workspace does not exist or you lack access.                                                                                                   |
| `BAD_USER_INPUT`         | `lookupOption` is missing on a `LOOKUP` field; `lookupId` is missing for `TODO_CUSTOM_FIELD`; or the source field is not a `REFERENCE`/`REFERENCED_BY` field. |
| `FORBIDDEN`              | You lack permission to manage custom fields in this workspace.                                                                                                |

## Related

- [Reference Field](/api/custom-fields/reference) — the source field a Lookup reads from.
- [Custom Field Values](/api/custom-fields/custom-field-values) — set values on editable fields.
- [Create Custom Fields](/api/custom-fields/create-custom-fields) — the `createCustomField` mutation in full.
- [List Custom Fields](/api/custom-fields/list-custom-fields) — query the custom fields in a workspace.
