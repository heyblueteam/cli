---
title: Field Library
description: Share one field across every workspace - promote it to the organization library, demote it back to one workspace, detach a workspace-local copy, or merge duplicates.
icon: BookOpen
order: 30
---

A custom field belongs to one workspace. Every other workspace that wants the same field builds its own, and the copies drift apart. The field library fixes this: a field can be promoted so the organization owns it, and any workspace in the organization can then mount that one shared field. One definition, one identity, one set of options - every change to the field shows up in every workspace that mounts it.

```
                 promoteField                 demoteField
 workspace-owned -------------> ORGANIZATION -------------> workspace-owned
 (mounted in its    (reversible)  FIELD      (only when no     (back to one
  own workspace)                  LIBRARY     other workspace   workspace)
                                  FIELD       mounts it)

 detachField: shared field + one table -> new workspace-local copy for that table
 mergeFields: duplicate field          -> folds into one surviving library field
```

The four mutations move a field between these scopes:

| Mutation       | Move                                                        | Reversible                                   |
| -------------- | ----------------------------------------------------------- | -------------------------------------------- |
| `promoteField` | Workspace-owned field becomes an organization library field | Yes - demote it back                         |
| `demoteField`  | Library field becomes one workspace's field again           | Yes - promote it again                       |
| `detachField`  | Library field stays shared; one table gets its own copy     | Not automatic - delete the copy when done    |
| `mergeFields`  | A duplicate field folds into a surviving library field      | No - the source field is retired permanently |

Promote, demote, and merge change who owns a field, so they need organization-level authority. Detach changes nothing for anyone else, so workspace admin rights are enough.

The scope inputs must match the request headers: `promoteField`, `demoteField`, and `mergeFields` require `organizationId` to match the organization in the `blue-org-id` header, and `detachField` requires `workspaceId` to match the workspace in the `blue-workspace-id` header.

```
POST https://api.blue.app/graphql
blue-token-id: YOUR_TOKEN_ID
blue-token-secret: YOUR_TOKEN_SECRET
blue-org-id: YOUR_ORG_ID
blue-workspace-id: project_123
```

Company and project headers accept an ID or a slug. Header names are case-insensitive. See [Authentication](/api/start-guide/authentication) for the full header reference.

## Promote a field to the library

`promoteField` turns a workspace-owned field into an organization library field. Only ownership changes - the field keeps its identity, name, type, options, values, and history, and every table that already shows it keeps showing it. What changes is reach: every other workspace in the organization can now mount the same field, and one definition edit updates all of them.

Calling `promoteField` on a field that is already in the library changes nothing and returns `promoted: false`.

```graphql
mutation PromoteField {
  promoteField(input: { organizationId: "company_123", fieldId: "field_123" }) {
    fieldId
    promoted
  }
}
```

```json
{ "data": { "promoteField": { "fieldId": "field_123", "promoted": true } } }
```

### PromoteFieldInput

| Parameter        | Type  | Required | Description                                                                        |
| ---------------- | ----- | -------- | ---------------------------------------------------------------------------------- |
| `organizationId` | `ID!` | Yes      | Organization that will own the library field. Must match the `blue-org-id` header. |
| `fieldId`        | `ID!` | Yes      | The workspace-owned field to promote.                                              |

### Returns

| Field      | Type       | Description                                         |
| ---------- | ---------- | --------------------------------------------------- |
| `fieldId`  | `ID!`      | The promoted field's ID, echoed from the input.     |
| `promoted` | `Boolean!` | `false` when the field was already a library field. |

## Demote a field to one workspace

`demoteField` is the reverse of promotion: a library field becomes an ordinary field owned by the workspace named in `workspaceId`. As with promotion, the field's identity, values, options, and history are untouched.

A library field can only demote into a workspace that still has it to itself. If the field is mounted in any table outside the target workspace, the call fails - unmount the field there first, or give that workspace its own copy with `detachField`. Blue refuses rather than silently taking the field away from a workspace that is using it.

Calling `demoteField` on a field the target workspace already owns returns `demoted: false`.

```graphql
mutation DemoteField {
  demoteField(
    input: { organizationId: "company_123", fieldId: "field_123", workspaceId: "project_123" }
  ) {
    fieldId
    demoted
  }
}
```

```json
{ "data": { "demoteField": { "fieldId": "field_123", "demoted": true } } }
```

### DemoteFieldInput

| Parameter        | Type  | Required | Description                                                                                     |
| ---------------- | ----- | -------- | ----------------------------------------------------------------------------------------------- |
| `organizationId` | `ID!` | Yes      | Organization that owns the library field. Must match the `blue-org-id` header.                  |
| `fieldId`        | `ID!` | Yes      | The library field to demote.                                                                    |
| `workspaceId`    | `ID!` | Yes      | Workspace that takes ownership. Must be an active, non-archived workspace in that organization. |

### Returns

| Field     | Type       | Description                                                      |
| --------- | ---------- | ---------------------------------------------------------------- |
| `fieldId` | `ID!`      | The demoted field's ID, echoed from the input.                   |
| `demoted` | `Boolean!` | `false` when the field already belonged to the target workspace. |

## Detach a field into a workspace-local copy

`detachField` forks a library field into a new, independent field owned by the workspace - "make your own copy". Use it when a workspace needs a field that starts life identical to the shared one but must evolve separately, and the caller has no edit rights on the shared definition.

The copy is per table, not per workspace. `databaseId` names the one record table that switches to the new field:

- The new field gets the source's type and configuration, and its name unless `name` says otherwise.
- Select options are cloned with fresh option IDs.
- The values and value history this table's records hold under the field are carried onto the new field.
- Table designations that pointed at the source (title, completion, color source) now point at the copy.
- The table's mount moves from the source field to the copy.

The source library field, its other mounts, and every other table are untouched. No other workspace sees any change.

`newFieldId` is the identity you choose for the new field - a client-generated cuid2, the same pattern as the optional `recordId` on `createRecordInDatabase`. Supplying it makes retries safe: replaying the call with the same `newFieldId` after the first detach committed changes nothing and returns `detached: false` with zero counts. Use a fresh ID for each new copy.

```graphql
mutation DetachField {
  detachField(
    input: {
      workspaceId: "project_123"
      databaseId: "table_123"
      fieldId: "field_123"
      newFieldId: "field_abc123def456"
      name: "Customer (local)"
    }
  ) {
    detached
    newFieldId
    copiedValues
    copiedOptions
    copiedHistory
  }
}
```

```json
{
  "data": {
    "detachField": {
      "detached": true,
      "newFieldId": "field_abc123def456",
      "copiedValues": 1284,
      "copiedOptions": 4,
      "copiedHistory": 3620
    }
  }
}
```

### DetachFieldInput

| Parameter     | Type     | Required | Description                                                                                     |
| ------------- | -------- | -------- | ----------------------------------------------------------------------------------------------- |
| `workspaceId` | `ID!`    | Yes      | Workspace that owns the table. Must match the `blue-workspace-id` header.                       |
| `databaseId`  | `ID!`    | Yes      | The record table that will use the new copy.                                                    |
| `fieldId`     | `ID!`    | Yes      | The library field to fork. Must be mounted on the table, and must be a library field.           |
| `newFieldId`  | `ID!`    | Yes      | Caller-supplied ID for the new field (cuid2 shape). Replaying with the same ID is a safe no-op. |
| `name`        | `String` | No       | Name for the copy. Omit it to keep the source field's name.                                     |

### Returns

| Field           | Type       | Description                                                          |
| --------------- | ---------- | -------------------------------------------------------------------- |
| `detached`      | `Boolean!` | `false` when this `newFieldId` was already detached onto this table. |
| `newFieldId`    | `ID!`      | The new field's ID, echoed from the input.                           |
| `copiedValues`  | `Int!`     | Field values carried onto the new field.                             |
| `copiedOptions` | `Int!`     | Select options cloned onto the new field (0 for non-select types).   |
| `copiedHistory` | `Int!`     | Value-history intervals carried onto the new field.                  |

## Merge fields

`mergeFields` folds a duplicate field into one surviving field, permanently. `sourceFieldId` is the duplicate that disappears; `targetFieldId` is the survivor. Use it to clean up the drift that predates the library - two same-type fields with overlapping values that should be one curated field.

If the target is not yet a library field, the merge promotes it first (`promotedTarget: true`), so the result is always a single shared field. Then the merge:

- Merges select options by matching title; a source option with no match is cloned onto the target.
- Moves source values onto the target, filling gaps only - a record that already holds a target value keeps it. A source value never overwrites what the survivor already held.
- Appends the source's full value history onto the target's.
- Repoints every table that mounted the source onto the target, along with the table designations (title, completion, color source) and the legacy compatibility rows.
- Retires the source field - the same soft delete as a field deletion, with the caller recorded as the deleter. This is not reversible.

The merge refuses to retire a field that other fields still depend on: a formula, lookup, or rollup that reads it, a referenced-by definition built on it, or a unique-ID field that uses it as its numbering scope. Update or remove the dependents first.

```graphql
mutation MergeFields {
  mergeFields(
    input: { organizationId: "company_123", sourceFieldId: "field_456", targetFieldId: "field_123" }
  ) {
    merged
    targetFieldId
    promotedTarget
    mergedOptions
    migratedValues
    migratedHistory
    repointedMounts
  }
}
```

```json
{
  "data": {
    "mergeFields": {
      "merged": true,
      "targetFieldId": "field_123",
      "promotedTarget": false,
      "mergedOptions": 3,
      "migratedValues": 512,
      "migratedHistory": 1487,
      "repointedMounts": 6
    }
  }
}
```

### MergeFieldsInput

| Parameter        | Type  | Required | Description                                                                  |
| ---------------- | ----- | -------- | ---------------------------------------------------------------------------- |
| `organizationId` | `ID!` | Yes      | Organization both fields belong to. Must match the `blue-org-id` header.     |
| `sourceFieldId`  | `ID!` | Yes      | The duplicate field to retire.                                               |
| `targetFieldId`  | `ID!` | Yes      | The surviving field. Promoted to the library first if it is workspace-owned. |

### Returns

| Field             | Type       | Description                                                             |
| ----------------- | ---------- | ----------------------------------------------------------------------- |
| `merged`          | `Boolean!` | Always `true` on success.                                               |
| `targetFieldId`   | `ID!`      | The surviving field's ID, echoed from the input.                        |
| `promotedTarget`  | `Boolean!` | `true` when the merge promoted a workspace-owned target to the library. |
| `mergedOptions`   | `Int!`     | Source options folded into the target (0 for non-select types).         |
| `migratedValues`  | `Int!`     | Source values moved onto the target.                                    |
| `migratedHistory` | `Int!`     | History intervals appended to the target's history.                     |
| `repointedMounts` | `Int!`     | Tables that now mount the target where they mounted the source.         |

A merge runs in one transaction: either every value, option, mount, and history row moves and the source retires, or nothing changes.

## Choosing the right operation

| Situation                                                        | Use            |
| ---------------------------------------------------------------- | -------------- |
| One workspace built a field that every workspace now needs       | `promoteField` |
| A shared field is really one workspace's concern again           | `demoteField`  |
| A workspace must change a field without touching the shared one  | `detachField`  |
| Two same-type duplicates should become one curated library field | `mergeFields`  |

Promote and demote never touch data; detach and merge move values. Detach is the safe escape hatch - it changes nothing for anyone else. Merge is the one destructive operation: read its refusal conditions before calling it.

Detach and merge carry values only for the 17 stored field types - single-line and multi-line text, number, percent, rating, currency, email, URL, phone, country, date, checkbox, location, single select, multi select, assignee, and reference. Computed types (formula, lookup, rollup, referenced by, currency conversion, time duration), unique ID, file, table, and button cannot be detached or merged yet. Promote and demote have no type restrictions - they move only ownership.

## Errors

| Code              | Mutation(s)                  | When                                                                                                                                                                                                                                                                                                                                                                                                                                         |
| ----------------- | ---------------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `UNAUTHENTICATED` | All                          | The request carries no valid token.                                                                                                                                                                                                                                                                                                                                                                                                          |
| `FORBIDDEN`       | All                          | The caller lacks the authority described in [Permissions](#permissions), or an `organizationId`/`workspaceId` in the input does not match the request headers. For `promoteField`, `demoteField`, and `mergeFields`, a field that does not exist or belongs to a different organization reports the same way. Retrying a `mergeFields` call that already completed also returns this - the source field is retired, so it no longer matches. |
| `BAD_USER_INPUT`  | `demoteField`, `detachField` | Field management is temporarily unavailable for the target workspace.                                                                                                                                                                                                                                                                                                                                                                        |
| `BAD_USER_INPUT`  | `demoteField`                | The target workspace does not exist, is archived, or belongs to a different organization.                                                                                                                                                                                                                                                                                                                                                    |
| `BAD_USER_INPUT`  | `demoteField`                | The field is not a library field.                                                                                                                                                                                                                                                                                                                                                                                                            |
| `BAD_USER_INPUT`  | `demoteField`                | The field is still mounted in another workspace. Unmount it there first, or detach a local copy.                                                                                                                                                                                                                                                                                                                                             |
| `BAD_USER_INPUT`  | `detachField`                | A supplied ID is empty or too long.                                                                                                                                                                                                                                                                                                                                                                                                          |
| `BAD_USER_INPUT`  | `detachField`                | The field is not mounted on that table.                                                                                                                                                                                                                                                                                                                                                                                                      |
| `BAD_USER_INPUT`  | `detachField`                | The field is not a library field - only a shared field can be detached.                                                                                                                                                                                                                                                                                                                                                                      |
| `BAD_USER_INPUT`  | `detachField`                | `newFieldId` is already used by another field.                                                                                                                                                                                                                                                                                                                                                                                               |
| `BAD_USER_INPUT`  | `detachField`                | The field resolves record IDs for a linked GitHub repository. Unlink the repository first.                                                                                                                                                                                                                                                                                                                                                   |
| `BAD_USER_INPUT`  | `mergeFields`                | `sourceFieldId` and `targetFieldId` are the same field.                                                                                                                                                                                                                                                                                                                                                                                      |
| `BAD_USER_INPUT`  | `mergeFields`                | The two fields have different types.                                                                                                                                                                                                                                                                                                                                                                                                         |
| `BAD_USER_INPUT`  | `mergeFields`                | Other fields still reference the source - a formula, lookup, rollup, referenced-by definition, or unique-ID numbering scope.                                                                                                                                                                                                                                                                                                                 |
| `BAD_USER_INPUT`  | `detachField`, `mergeFields` | The field's type cannot be detached or merged yet.                                                                                                                                                                                                                                                                                                                                                                                           |
| `BAD_USER_INPUT`  | `detachField`, `mergeFields` | The field is mirrored from an external database and must be changed at the source, or it is the table's Tags field.                                                                                                                                                                                                                                                                                                                          |

## Permissions

`promoteField`, `demoteField`, and `mergeFields` are organization-management operations. The caller must be an OWNER or ADMIN of the organization, or a user the organization has delegated org-wide field management to. Workspace-level admin rights are not enough - promoting a field changes every workspace that can mount it.

`detachField` needs only admin rights where the copy lands: OWNER or ADMIN of the organization, or OWNER or ADMIN of the workspace that owns the table. It deliberately does not require rights over the shared field's own definition - detaching is how a workspace stops depending on a definition it cannot edit.

All four mutations require an authenticated user. See [Access control](/api/custom-fields#access-control) for the per-field visibility and edit rules that govern normal field reads and writes.

## Related

- [Create a custom field](/api/custom-fields/create-custom-fields) - add a field to one workspace
- [List custom fields](/api/custom-fields/list-custom-fields) - read the fields mounted in a workspace
- [Set field values](/api/custom-fields/custom-field-values) - write values onto records
- [Delete a custom field](/api/custom-fields/delete-custom-field) - remove a field and its values
- [Records](/api/records) - the caller-supplied-ID retry pattern `newFieldId` follows
