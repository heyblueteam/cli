---
title: Lookup Field
description: Pull live data from records linked through a Reference field, with no manual copying.
icon: Search
order: 28
---

A Lookup field surfaces data from records that are linked through a [Reference field](/api/custom-fields/reference) — for example, showing the tags, assignees, or a specific custom field value of every linked record. Lookups are the `LOOKUP` value of the `CustomFieldType` enum. They recompute automatically whenever the referenced data changes. A Lookup is read-only unless `allowEdits` is set, which lets the value be edited through to the source record — see [Write through to the source record](#write-through-to-the-source-record).

Custom fields are `CustomField` objects in the API. A Lookup always points at a `REFERENCE` (or `REFERENCED_BY`) source field on the same workspace and reads one kind of data from the records that source field resolves.

## Overview

A Lookup is configured with two things:

- **`referenceId`** — the ID of the source field to read from. This must be a `REFERENCE` or `REFERENCED_BY` field on the same workspace. You cannot point a Lookup at another Lookup.
- **`lookupType`** — which piece of data to pull from each linked record (tags, assignees, due dates, a custom field value, and so on).

When `lookupType` is `TODO_CUSTOM_FIELD`, you also pass **`lookupId`** — the ID of the specific custom field to read from each linked record.

## Create

Create a Lookup that surfaces the tags of every record linked through a Reference field. Custom fields are scoped by the `blue-workspace-id` header, so no workspace ID appears in the input.

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

| Parameter     | Type                     | Required    | Description                                                                                                                   |
| ------------- | ------------------------ | ----------- | ----------------------------------------------------------------------------------------------------------------------------- |
| `referenceId` | `String!`                | Yes         | ID of the source `REFERENCE` or `REFERENCED_BY` field to read from.                                                           |
| `lookupType`  | `CustomFieldLookupType!` | Yes         | Which data to pull from each linked record.                                                                                   |
| `lookupId`    | `String`                 | Conditional | ID of the custom field to read. Required when `lookupType` is `TODO_CUSTOM_FIELD`; ignored otherwise.                         |
| `allowEdits`  | `Boolean`                | No          | Allow the looked-up value to be edited through to the source record. Defaults to `false`. Accepted only for editable targets. |

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

A Lookup field has no stored value of its own — its result is computed in the context of the record it is rendered on. Read it from `Record.customFields`, which returns `[CustomField!]!` directly (there is no wrapper type), and select `customFieldLookupOption` on the element:

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
| `allowEdits`        | `Boolean!`              | Whether this Lookup is configured for write-through editing.              |
| `canEditSource`     | `Boolean!`              | Whether the calling user may edit the source record through this Lookup.  |

## Response

The shape of `lookupResult` depends on `lookupType`. Because it is a `JSON` field, the structures below are **illustrative** — they show typical output but are not enforced by the schema.

For a Lookup that rides a multi-record Reference, `lookupResult` is **positional**: one entry per referenced record, in the Reference's stored order, with `null` for a record that has no value. This includes `SELECT_SINGLE` targets, which previously arrived compacted and de-duplicated — consumers should skip `null` entries and de-duplicate for display where needed. Tag and assignee entries also carry their `id`.

A Lookup that rides a `REFERENCED_BY` source field is **always** positional, because a reverse relation is always a list. Its entries line up with that field's `referencedByResult`, in the same stored order (newest first), one entry per source record and `null` where a source record has no value:

```
referencedByResult   [ "todo_c", "todo_b", "todo_a" ]
lookupResult         [ "Warehouse B", null, "Warehouse A" ]
                        ^ todo_c        ^ todo_b has no value
```

Before this change these snapshots arrived compacted, so entry `i` did not correspond to source record `i`. Existing stored snapshots are recomputed once after the change; until a record is recomputed its snapshot keeps the old compacted shape. Read entries positionally only when the two arrays have the same length. `REFERENCED_BY`-sourced Lookups are always read-only — see [Referenced by](/api/custom-fields/referenced-by).

<Callout variant="warning" title="This changes the payload of every REFERENCED_BY-sourced Lookup">
It applies to all of them, not only the ones shown in a folder's Grid View. If you read `lookupResult` from the GraphQL API, the REST API or a webhook, check two things:

- **List-valued targets nest one level.** `TODO_TAG`, `TODO_ASSIGNEE`, `SELECT_MULTI` and `FILE` return one **array per source record**, not one flat list: `[[tagA, tagB], null, [tagA]]` rather than `[tagA, tagB]`. Flatten before reading `.title` / `.firstName`.
- **Entries are no longer de-duplicated.** A compacted `TODO_TAG` snapshot returned each distinct tag once. A positional one returns one entry per source record, so a record referenced by 40 others that all carry the same tag now yields that tag 40 times. De-duplicate for display.

Both were already true for Lookups riding a multi-record Reference; this change extends them to `REFERENCED_BY` sources. Blue's own automation, email and webhook rendering flattens and de-duplicates for you — this note is for direct API consumers.
</Callout>

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

## Write through to the source record

Set `allowEdits: true` on `lookupOption` to let the looked-up value be changed from the workspace holding the Lookup. The edit is applied to the **source record in the source workspace**, so it changes that record for every workspace that reads it.

There are no new mutations for this. You write the source record directly, with the source record's ID and the source workspace's `blue-workspace-id` header, using the mutation that matches the target: `setRecordCustomField`, `editRecord`, `setRecordAssignees`, `setRecordTags`, or `moveRecord`. `allowEdits` only tells a client that the value is meant to be editable in place — it grants no permission of its own.

### Editable targets

`allowEdits` is accepted only when the Lookup reads data Blue can write back. Setting it on any other target is rejected with `BAD_USER_INPUT` at create and edit time.

| `lookupType`        | Editable                                                                                                                                                                                                                    |
| ------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `TODO_DUE_DATE`     | Yes                                                                                                                                                                                                                         |
| `TODO_ASSIGNEE`     | Yes                                                                                                                                                                                                                         |
| `TODO_TAG`          | Yes                                                                                                                                                                                                                         |
| `TODO_LIST`         | Yes                                                                                                                                                                                                                         |
| `TODO_DESCRIPTION`  | Yes                                                                                                                                                                                                                         |
| `TODO_CUSTOM_FIELD` | Only for these target field types: `TEXT_SINGLE`, `TEXT_MULTI`, `NUMBER`, `PERCENT`, `RATING`, `EMAIL`, `PHONE`, `URL`, `CHECKBOX`, `DATE`, `COUNTRY`, `LOCATION`, `CURRENCY`, `SELECT_SINGLE`, `SELECT_MULTI`, `REFERENCE` |

`TODO_CREATED_AT` and `TODO_UPDATED_AT` are system timestamps, and `TODO_REFERENCED_BY` has no per-entry source record, so none of them can be edited. Computed target types (`FORMULA`, `ROLLUP`, `LOOKUP`), `FILE`, and `TABLE` are rejected as well. A Lookup that rides a `REFERENCED_BY` source field is always read-only.

### Who may edit

`canEditSource` is resolved per requesting user and is `true` only when `allowEdits` is on **and** that user could make the same edit in the source workspace directly. It checks membership in the source workspace, an access level other than `VIEW_ONLY` or `COMMENT_ONLY`, the user's role allowing records, the workspace being active, and — for `TODO_CUSTOM_FIELD` — the target field still being alive, still an editable type, and editable by the user's role. A `TODO_LIST` Lookup additionally needs `OWNER`, `ADMIN`, or `MEMBER`, matching `moveRecord`.

Permissions come entirely from the **source** workspace. A user with `VIEW_ONLY` access to the workspace holding the Lookup, but full access to the source workspace, can still edit these entries and nothing else on the page.

Treat `canEditSource` as a display signal only. The write mutations run their own checks and remain the enforcement point, so a client that ignores `canEditSource` gets a permission error rather than an unauthorized write.

On `editCustomField`, omitting `allowEdits` from `lookupOption` leaves the stored value unchanged. The new target is still checked against the stored value, so repointing an editable Lookup at a target that cannot be edited is rejected even when the flag is absent from the input.

## Notes

- **No direct value.** You cannot set a Lookup's own value with `setRecordCustomField` — it always reflects the current linked data and recomputes when that data changes. With `allowEdits`, you write the source record instead; see [Write through to the source record](#write-through-to-the-source-record).
- **No aggregation.** A Lookup extracts the linked values as-is — it has no built-in sum, count, or average across linked records.
- **Source must be a Reference.** `referenceId` must point at a `REFERENCE` or `REFERENCED_BY` field. Pointing it at another `LOOKUP` is rejected — Lookup-of-Lookup chains are not supported.
- **Cross-workspace access.** A viewer only sees Lookup results for linked records in workspaces they have access to.

## Errors

| Code                     | When                                                                                                                                                                                                                 |
| ------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `CUSTOM_FIELD_NOT_FOUND` | `referenceId` (or `lookupId`) does not resolve to a field you can access.                                                                                                                                            |
| `PROJECT_NOT_FOUND`      | The referenced workspace does not exist or you lack access.                                                                                                                                                          |
| `BAD_USER_INPUT`         | `lookupOption` is missing on a `LOOKUP` field; `lookupId` is missing for `TODO_CUSTOM_FIELD`; the source field is not a `REFERENCE`/`REFERENCED_BY` field; or `allowEdits` is set on a target that cannot be edited. |
| `FORBIDDEN`              | You lack permission to manage custom fields in this workspace.                                                                                                                                                       |

## Related

- [Reference Field](/api/custom-fields/reference) — the source field a Lookup reads from.
- [Custom Field Values](/api/custom-fields/custom-field-values) — set values on editable fields.
- [Create Custom Fields](/api/custom-fields/create-custom-fields) — the `createCustomField` mutation in full.
- [List Custom Fields](/api/custom-fields/list-custom-fields) — query the custom fields in a workspace.
