---
title: Reference Field
description: Link records to records in another workspace, with optional filtering and single- or multi-select.
icon: Network
order: 25
---

A Reference field links a record to one or more records in another workspace — for example, connecting a task to the feature request it implements, or a deal to its account. References are the `REFERENCE` value of the `CustomFieldType` enum.

Custom fields are `CustomField` objects in the API; records are `Record` objects and workspaces are `Workspace` objects. A Reference field always points at a target workspace (`referenceProjectId`); the records a user can link are drawn from that workspace, optionally narrowed by a filter.

To pull specific data _out_ of linked records (their tags, assignees, or a single field value), pair a Reference with a [Lookup field](/api/custom-fields/lookup). To aggregate across linked records (sum, count, average), use a [Rollup field](/api/custom-fields/rollup). To see the records that point _at_ a given record, use a [Referenced-by field](/api/custom-fields/referenced-by).

## Overview

| Property          | Value                                                                      |
| ----------------- | -------------------------------------------------------------------------- |
| `CustomFieldType` | `REFERENCE`                                                                |
| Set with          | `setRecordCustomField` → `customFieldReferenceTodoIds: [String!]`          |
| Read with         | `selectedTodos: [Record!]` (always an array) — or the raw `value` JSON     |
| Target workspace  | Required: `referenceProjectId`                                             |
| Cardinality       | `referenceMultiple` — UI selection only; the wire shape is always an array |

`referenceMultiple` controls how many records a user can pick in the app (one vs. many). It does **not** change the API shape: you always set values with the `customFieldReferenceTodoIds` array and read them back from the `selectedTodos` array.

### Creating records from the picker

`referenceAllowCreate` (default `false`) lets people create a record in the target workspace directly from the field's picker in the app, instead of only linking records that already exist.

Read `canCreateReferenceRecords: Boolean` on `CustomField` to know whether the **calling user** gets that affordance. It is `true` only when `referenceAllowCreate` is on **and** the caller could already create a record in `referenceProjectId`. Both are UI signals — `createTodo` in the target workspace remains the enforcement point, so a client that ignores them gains nothing.

## Create

Create a Reference field that links to records in another workspace. `referenceProjectId` is required for `REFERENCE` fields — the resolver rejects the call without it. The field is scoped to the workspace in your `blue-workspace-id` header; there is no `projectId` argument.

```graphql
mutation CreateReferenceField {
  createCustomField(
    input: { name: "Related feature", type: REFERENCE, referenceProjectId: "project_123" }
  ) {
    id
    name
    type
    referenceMultiple
  }
}
```

### Multi-select with a filter

Allow several links and restrict the selectable records with `referenceFilter`. The filter is a `TodoFilterInput` — use its real fields (for example `tagIds` and `showCompleted`), not arbitrary keys.

```graphql
mutation CreateFilteredReferenceField {
  createCustomField(
    input: {
      name: "Dependencies"
      type: REFERENCE
      referenceProjectId: "project_123"
      referenceMultiple: true
      referenceFilter: { tagIds: ["tag_123"], showCompleted: false }
      description: "Open records tagged as a dependency"
    }
  ) {
    id
    name
    type
    referenceMultiple
    referenceProject {
      id
      name
    }
    referenceFilter {
      tagIds
      showCompleted
    }
  }
}
```

### CreateCustomFieldInput (Reference fields)

| Parameter              | Type               | Required | Description                                                                                                                              |
| ---------------------- | ------------------ | -------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| `name`                 | `String!`          | Yes      | Display name of the field.                                                                                                               |
| `type`                 | `CustomFieldType!` | Yes      | Must be `REFERENCE`.                                                                                                                     |
| `referenceProjectId`   | `String`           | Yes      | ID of the workspace whose records can be linked. Schema-optional, but required for `REFERENCE` — omitting it returns `BAD_USER_INPUT`.   |
| `referenceMultiple`    | `Boolean`          | No       | Allow linking more than one record. Defaults to single-select. UI cardinality only.                                                      |
| `referenceAllowCreate` | `Boolean`          | No       | Let people create a record in the target workspace from the field's picker. Defaults to `false`.                                         |
| `referenceFilter`      | `TodoFilterInput`  | No       | Narrows which records appear in the picker. Common fields: `tagIds`, `assigneeIds`, `todoListIds`, `showCompleted`, `dueStart`/`dueEnd`. |
| `description`          | `String`           | No       | Help text shown with the field.                                                                                                          |

`referenceFilter` accepts the same `TodoFilterInput` used by record queries — see [List records](/api/records/list-records) for the full field set. The filter shapes the picker only; it is **not** enforced when you set a value through the API.

<Callout variant="info" title="Editing a Reference field">

To change a Reference field's target workspace, filter, or cardinality after creation, call `editCustomField` with the same `referenceProjectId` / `referenceFilter` / `referenceMultiple` arguments plus the `customFieldId`. See the <a href="/api/custom-fields">custom fields overview</a> for the edit operation.

</Callout>

## Set a value

Link records with `setRecordCustomField`, passing the target record IDs in `customFieldReferenceTodoIds`. This is always an array — pass one ID for a single-select field, several for a multi-select field. The set call **replaces** the current link set: any IDs you omit are unlinked.

The mutation returns `Boolean!` — it has no selectable sub-fields. Read the value back with a separate `recordQueries { todos }` query (see [Read a value](#read-a-value)).

```graphql
mutation LinkRecords {
  setRecordCustomField(
    input: {
      todoId: "todo_123"
      customFieldId: "field_123"
      customFieldReferenceTodoIds: ["todo_456", "todo_789"]
    }
  )
}
```

```json
{ "data": { "setRecordCustomField": true } }
```

To clear all links, pass an empty array: `customFieldReferenceTodoIds: []`.

### SetRecordCustomFieldInput (Reference fields)

| Parameter                     | Type        | Required              | Description                                                                                                                                                         |
| ----------------------------- | ----------- | --------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `todoId`                      | `String!`   | Yes                   | ID of the record whose field you are setting.                                                                                                                       |
| `customFieldId`               | `String!`   | Yes                   | ID of the `REFERENCE` field.                                                                                                                                        |
| `customFieldReferenceTodoIds` | `[String!]` | Yes (for `REFERENCE`) | IDs of the records to link. Schema-optional — the resolver dispatches on whichever value field is present — but it is the only value field a Reference field reads. |

<Callout variant="warning" title="Links must live in the target workspace">

Only IDs whose records belong to the field's `referenceProjectId` are linked. IDs from any other workspace are silently dropped — the call still returns `true`, but those records do not appear in `selectedTodos`.

</Callout>

## Read a value

A Reference field's value comes back on `Record.customFields` as a `CustomField` rendered in record context. `Record.customFields` returns `[CustomField!]!` directly — there is no wrapper type. The linked records are exposed on `selectedTodos` (always an array) and, identically, on the raw `value` JSON.

```graphql
query GetRecordWithReferences {
  recordQueries {
    todos(filter: { companyIds: ["company_123"], todoIds: ["todo_123"] }) {
      items {
        id
        title
        customFields {
          id
          name
          type
          referenceMultiple
          selectedTodos {
            id
            title
            done
            duedAt
            users {
              id
              fullName
            }
            tags {
              id
              title
            }
          }
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
            "title": "Ship onboarding flow",
            "customFields": [
              {
                "id": "clm5a1b2c000108l0h1n2k3p4",
                "name": "Dependencies",
                "type": "REFERENCE",
                "referenceMultiple": true,
                "selectedTodos": [
                  {
                    "id": "clm6d7e8f000208l0m4q5r6s7",
                    "title": "Set up auth service",
                    "done": false,
                    "duedAt": "2026-06-01T17:00:00.000Z",
                    "users": [{ "id": "clm7g0h1i000308l0t8u9v0w1", "fullName": "Ada Lovelace" }],
                    "tags": [{ "id": "clm8j2k3l000408l0x2y3z4a5", "title": "dependency" }]
                  }
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

### CustomField (record context)

When a `CustomField` is resolved through `Record.customFields`, these are the relevant fields for a Reference field:

| Field           | Type               | Description                                                                                            |
| --------------- | ------------------ | ------------------------------------------------------------------------------------------------------ |
| `id`            | `ID!`              | The field definition ID.                                                                               |
| `name`          | `String!`          | Field display name.                                                                                    |
| `type`          | `CustomFieldType!` | `REFERENCE`.                                                                                           |
| `selectedTodos` | `[Record!]`        | The linked records. Always an array, regardless of `referenceMultiple`. `null` outside record context. |
| `value`         | `JSON`             | The same linked records as raw JSON — `selectedTodos` is the typed accessor for it.                    |
| `todo`          | `Record`           | The record this value belongs to (the host record).                                                    |
| `createdAt`     | `DateTime!`        | When the field was created.                                                                            |
| `updatedAt`     | `DateTime!`        | When the field was last modified.                                                                      |

The value object is a `CustomField` carrying record-aware fields — there is no separate `RecordCustomField` join type and no `customField { … }` sub-object. Select linked-record fields directly on `selectedTodos`.

<Callout variant="warning" title="Use real Record fields">

A linked `Record` has no `status`, `dueDate`, or `assignees` field, and there is no `RecordStatus` enum. Completion is `done: Boolean!`; the due date is `duedAt`; assignees are `users` (`User.fullName` / `firstName` / `lastName`). A `Tag` exposes `title`, not `name`.

</Callout>

## Create a record with references

`createRecord` accepts a `customFields` array of `CreateRecordInputCustomField` entries. Each entry takes a single `value` string, so it sets exactly one linked record at create time. To link several records at once, create the record first, then call `setRecordCustomField` with the full `customFieldReferenceTodoIds` array.

```graphql
mutation CreateRecordWithReference {
  createRecord(
    input: {
      title: "Implement onboarding flow"
      todoListId: "list_123"
      customFields: [{ customFieldId: "field_123", value: "todo_456" }]
    }
  ) {
    id
    title
    customFields {
      id
      name
      type
      selectedTodos {
        id
        title
      }
    }
  }
}
```

## Notes

- **Cardinality is wire-shape neutral.** Whether `referenceMultiple` is `true` or `false`, you set values with the `customFieldReferenceTodoIds` array and read them from the `selectedTodos` array. The flag only limits how many records the app's picker allows.
- **Set replaces, not appends.** `setRecordCustomField` overwrites the link set with exactly the IDs you pass. Include every link you want to keep.
- **Filters are advisory.** `referenceFilter` shapes the in-app picker but is not enforced on API writes — you can link any record in the target workspace.
- **Cross-workspace visibility.** A viewer only sees linked records in workspaces they can access. The link itself persists regardless of the viewer's access.
- **REFERENCE vs. REFERENCED_BY.** A Reference field stores the links you set. A [Referenced-by field](/api/custom-fields/referenced-by) is the computed reverse view — it lists the records that point at the host record — and cannot be set with `setRecordCustomField`.

## Errors

| Code                     | When                                                                                                  |
| ------------------------ | ----------------------------------------------------------------------------------------------------- |
| `BAD_USER_INPUT`         | `referenceProjectId` is missing on a `REFERENCE` field at create time.                                |
| `PROJECT_NOT_FOUND`      | `referenceProjectId` does not resolve to a workspace you can access.                                  |
| `CUSTOM_FIELD_NOT_FOUND` | `customFieldId` does not resolve to a field in a workspace you belong to (on `setRecordCustomField`). |
| `TODO_NOT_FOUND`         | `todoId` does not resolve to a record you can edit.                                                   |
| `FORBIDDEN`              | You lack permission to manage custom fields in this workspace, or your role cannot edit this field.   |

## Related

- [Lookup Field](/api/custom-fields/lookup) — pull a value out of linked records.
- [Rollup Field](/api/custom-fields/rollup) — aggregate across linked records.
- [Referenced-by Field](/api/custom-fields/referenced-by) — the computed reverse view of incoming references.
- [Set field values](/api/custom-fields/custom-field-values) — the full `setRecordCustomField` input mapping.
- [Create custom fields](/api/custom-fields/create-custom-fields) — the `createCustomField` mutation in full.
- [List records](/api/records/list-records) — the `TodoFilterInput` fields available to `referenceFilter`.
