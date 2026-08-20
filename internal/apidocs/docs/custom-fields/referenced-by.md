---
title: Referenced By Field
description: Surface the records in another workspace that reference this record through a reference field - the reverse of a reference.
icon: GitMerge
order: 28
---

A referenced-by field shows the records that point at the current record through a [reference field](/api/custom-fields/reference) elsewhere. Its `CustomFieldType` is `REFERENCED_BY`. Where a reference field is the outgoing link you set, a referenced-by field is the incoming view, computed automatically — it is the reverse side of the same relationship.

## Overview

You configure a referenced-by field by naming the workspace and the reference field that points back at this one. Blue then resolves, per record, the set of source records whose reference field includes it. The value is read-only — there is nothing to set.

## Create

```graphql
mutation CreateLinkedDealsField {
  createCustomField(
    input: {
      name: "Linked Deals"
      type: REFERENCED_BY
      referencedByOption: { referencingProjectId: "project_123", referencingFieldId: "field_123" }
    }
  ) {
    id
    name
    type
    customFieldReferencedByOption {
      referencingProject {
        id
        name
      }
      referencingField {
        id
        name
      }
    }
  }
}
```

| Parameter            | Type                                 | Required | Description                               |
| -------------------- | ------------------------------------ | -------- | ----------------------------------------- |
| `name`               | `String!`                            | Yes      | Display name of the field.                |
| `type`               | `CustomFieldType!`                   | Yes      | Must be `REFERENCED_BY`.                  |
| `referencedByOption` | `CustomFieldReferencedByOptionInput` | Yes      | The source workspace and reference field. |

### CustomFieldReferencedByOptionInput

| Field                  | Type              | Required | Description                                                    |
| ---------------------- | ----------------- | -------- | -------------------------------------------------------------- |
| `referencingProjectId` | `String`          | No       | The workspace whose records reference this one.                |
| `referencingFieldId`   | `String`          | No       | The `REFERENCE` field on that workspace that points back here. |
| `filters`              | `TodoFilterInput` | No       | Limit the incoming records to those matching this filter.      |

The referencing field must be a `REFERENCE` field that belongs to the same organization; otherwise create returns a `BAD_USER_INPUT` error. An unknown workspace returns `PROJECT_NOT_FOUND`.

## Set a value

Referenced-by fields are computed — `setRecordCustomField` rejects them with a `BAD_USER_INPUT` error. To change what a record is referenced by, edit the reference field on the source record instead.

## Read a value

The resolved source records come back in `referencedByResult` (a `JSON` value, populated only when the field is read in a record context). The field's configuration is available on `customFieldReferencedByOption`.

```graphql
query RecordReferencedBy {
  recordQueries {
    todos(filter: { companyIds: [], projectIds: ["project_123"] }, limit: 5) {
      items {
        id
        title
        customFields {
          name
          type
          referencedByResult
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
                "name": "Linked Deals",
                "type": "REFERENCED_BY",
                "referencedByResult": [
                  { "id": "clm4z7p1x000508l0a1b2c3d4", "title": "Acme renewal Q3" }
                ]
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

- The relationship is driven entirely by the source reference field. Adding or removing a reference on a source record updates this field automatically.
- Use `filters` to scope the incoming records — for example, only source records that are not done.
- `referencedByResult` is capped at 200 records, ordered newest first.

## Reading values from the source records

Point a [Lookup](/api/custom-fields/lookup) field's `referenceId` at this field to read one value from each record that references this one. That Lookup's `lookupResult` is **positional** over this field's `referencedByResult`: one entry per source record, in the same order, `null` where a source record has no value. Match the two arrays by index only when their lengths agree.

These Lookups are always read-only — write-through editing (`allowEdits`) is rejected for a `REFERENCED_BY` source, because the value belongs to the source record in its own workspace.

## Related

- [Reference field](/api/custom-fields/reference)
- [Lookup field](/api/custom-fields/lookup)
- [Rollup field](/api/custom-fields/rollup)
- [Custom fields overview](/api/custom-fields)
