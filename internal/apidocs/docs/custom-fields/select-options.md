---
title: Select Options
description: Add, rename, reorder, limit, remap and retire the options of a record-engine select field — the choices behind SELECT_SINGLE and SELECT_MULTI fields and the columns of a board.
icon: ListChecks
order: 31
---

Options are the choices of a select field. A `SELECT_SINGLE` field picks one option per record, a `SELECT_MULTI` field picks several, and a board grouped by the field renders each option as a column. These mutations manage that option list on **record-engine fields** — fields defined on a table (database) and mounted on it. For the legacy `CustomField` option mutations (`createCustomFieldOption`, `editCustomFieldOption`, `deleteCustomFieldOption`), see [Single-Select Field](/api/custom-fields/select-single) and [Multi-Select Field](/api/custom-fields/select-multi); the REST `manage-custom-field-options` endpoint covers the same legacy rows.

Six mutations cover the whole life cycle:

| Mutation                          | Does                                                                 | Returns                                                 |
| --------------------------------- | -------------------------------------------------------------------- | ------------------------------------------------------- |
| `createRecordFieldOption`         | Add an option (append at the end).                                   | `optionId`, `created`, `position`                       |
| `renameRecordFieldOption`         | Change title, color, icon, or lock the option.                       | `optionId`, `renamed`                                   |
| `moveRecordFieldOption`           | Place the option before or after a sibling.                          | `Boolean`                                               |
| `setRecordFieldOptionWip`         | Set or clear the option's WIP (work-in-progress) limit.              | `Boolean`                                               |
| `remapRecordFieldOption`          | Retire the option, moving its values onto a target or clearing them. | `optionId`, `retired`, `moved`, `cleared`               |
| `remapAndDeleteRecordFieldOption` | Same, through a durable, replay-safe write for Blue's own clients.   | `optionId`, `deleted`, `cleared`, `remappedRecordCount` |

Every mutation takes `workspaceId` and scopes to the caller's workspace; the workspace header selects it, and a mismatch is rejected.

Read options back with the `recordFieldOptions` query:

```graphql
query FieldOptions {
  recordFieldOptions(workspaceId: "ws_123", fieldId: "field_123") {
    id
    title
    color
    position
    icon
    isLocked
    wipLimit
    wipLimitHard
  }
}
```

## Create an option

`createRecordFieldOption` appends one option to a choice field (`SELECT_SINGLE` or `SELECT_MULTI`, plus currency and currency-conversion fields, whose options are the allowed currencies and conversion rules). The caller supplies the option id, so repeating a call with the same ids is a no-op that returns `created: false` instead of adding a duplicate. New options are appended after the existing ones; `position` is the 0-based ordinal the option landed at.

```graphql
mutation AddOption {
  createRecordFieldOption(
    input: {
      workspaceId: "ws_123"
      fieldId: "field_123"
      optionId: "option_blocked"
      title: "Blocked"
      color: "#dc3545"
    }
  ) {
    optionId
    created
    position
  }
}
```

```json
{
  "data": {
    "createRecordFieldOption": { "optionId": "option_blocked", "created": true, "position": 4 }
  }
}
```

### CreateRecordFieldOptionInput

| Parameter     | Type      | Required | Description                                                                              |
| ------------- | --------- | -------- | ---------------------------------------------------------------------------------------- |
| `workspaceId` | `ID!`     | Yes      | The workspace that owns the field.                                                       |
| `fieldId`     | `ID!`     | Yes      | The choice field to add the option to.                                                   |
| `optionId`    | `ID!`     | Yes      | Caller-chosen option id (at most 191 characters). A repeat of an existing id is a no-op. |
| `title`       | `String!` | Yes      | Display text. Must not be blank.                                                         |
| `color`       | `String!` | Yes      | Option color (the app renders hex values; no hex validation).                            |
| `recordId`    | `ID`      | No       | With `databaseId`, the record being edited. Enables the delegated path below.            |
| `databaseId`  | `ID`      | No       | The table holding that record.                                                           |

### Two authorization paths

- **Schema authority** — the workspace OWNER/ADMIN who owns the field, or a user holding the field-management grant, may add options to any of the organization's choice fields.
- **Delegated (a member adding an option while editing a record)** — pass `recordId` and `databaseId`. A caller who may edit this field on that record may add an option without schema authority. Without both ids and edit access, the call is refused. The same idempotency and append-at-end rules apply, and no option-count cap applies on either path.

## Rename an option

`renameRecordFieldOption` changes an option's `title`, `color`, `icon`, or lock state. Stored record values reference the option id, so every record holding the option follows the new label without a rewrite. Send only the properties you want to change; a call that changes nothing returns `renamed: false`.

`isLocked: true` locks the option: new records can no longer be created with it (a locked board column refuses quick-add). Existing records keep the option, and moving a record onto it is unaffected. `icon` is a lowercase lucide-style slug — letters, digits and dashes, at most 64 characters.

Renaming or recoloring also refreshes computed fields that snapshot option labels, and a color change drops board group-color overrides copied from the old color so the explicit new color wins.

```graphql
mutation RenameOption {
  renameRecordFieldOption(
    input: {
      workspaceId: "ws_123"
      fieldId: "field_123"
      optionId: "option_123"
      title: "In Review"
      color: "#ffc107"
      icon: "eye"
    }
  ) {
    optionId
    renamed
  }
}
```

```json
{ "data": { "renameRecordFieldOption": { "optionId": "option_123", "renamed": true } } }
```

### RenameRecordFieldOptionInput

| Parameter     | Type      | Required | Description                                                   |
| ------------- | --------- | -------- | ------------------------------------------------------------- |
| `workspaceId` | `ID!`     | Yes      | The workspace that owns the field.                            |
| `fieldId`     | `ID!`     | Yes      | The field the option belongs to.                              |
| `optionId`    | `ID!`     | Yes      | The option to change.                                         |
| `title`       | `String`  | No       | New display text. Must not be blank when sent.                |
| `color`       | `String`  | No       | New color.                                                    |
| `icon`        | `String`  | No       | Option icon slug; `null` clears it.                           |
| `isLocked`    | `Boolean` | No       | `true` locks the option against new records; `false` unlocks. |

## Move an option

`moveRecordFieldOption` places one option before or after a sibling option of the same field. `after: true` puts `optionId` directly after `targetOptionId`; `after: false` puts it directly before. Moving an option to where it already is succeeds without change. The order is shared schema — every table mounting the field sees it.

The `databaseId` must be a table the field is mounted on, in the caller's workspace.

```graphql
mutation MoveOption {
  moveRecordFieldOption(
    input: {
      workspaceId: "ws_123"
      databaseId: "db_123"
      fieldId: "field_123"
      optionId: "option_123"
      targetOptionId: "option_456"
      after: true
    }
  )
}
```

```json
{ "data": { "moveRecordFieldOption": true } }
```

### MoveRecordFieldOptionInput

| Parameter        | Type       | Required | Description                                         |
| ---------------- | ---------- | -------- | --------------------------------------------------- |
| `workspaceId`    | `ID!`      | Yes      | The workspace that owns the field.                  |
| `databaseId`     | `ID!`      | Yes      | A table the field is mounted on.                    |
| `fieldId`        | `ID!`      | Yes      | A `SELECT_SINGLE` or `SELECT_MULTI` field.          |
| `optionId`       | `ID!`      | Yes      | The option to move.                                 |
| `targetOptionId` | `ID!`      | Yes      | The sibling to place next to.                       |
| `after`          | `Boolean!` | Yes      | `true` places after the sibling, `false` before it. |

## Set a WIP limit

A WIP (work-in-progress) limit caps how many records may hold an option — typically a board column. `setRecordFieldOptionWip` sets that cap on one option of a `SELECT_SINGLE` field.

The limit is stored on the option, so every table mounting the field sees the same limit, while the count stays local to each table: only that table's active, non-trashed records count, and a multi-select assignment consumes one slot per table. Two modes:

- `hard: true` (**BLOCK**) — once the count reaches `limit`, writes that would move a record into the option are refused with `WIP_LIMIT_REACHED`. Documented exemptions (inbound email, imports) may still exceed the limit.
- `hard: false` (**WARN**) — records may still enter; readers can show that the option is over capacity.

`limit: null` clears the limit and the blocking behavior.

```graphql
mutation SetWipLimit {
  setRecordFieldOptionWip(
    input: {
      workspaceId: "ws_123"
      databaseId: "db_123"
      fieldId: "field_123"
      optionId: "option_123"
      limit: 5
      hard: true
    }
  )
}
```

```json
{ "data": { "setRecordFieldOptionWip": true } }
```

### SetRecordFieldOptionWipInput

| Parameter     | Type       | Required | Description                                                 |
| ------------- | ---------- | -------- | ----------------------------------------------------------- |
| `workspaceId` | `ID!`      | Yes      | The workspace that owns the field.                          |
| `databaseId`  | `ID!`      | Yes      | A table the field is mounted on.                            |
| `fieldId`     | `ID!`      | Yes      | A `SELECT_SINGLE` field.                                    |
| `optionId`    | `ID!`      | Yes      | The option to limit.                                        |
| `limit`       | `Int`      | No       | Maximum record count. `null` (or omitted) clears the limit. |
| `hard`        | `Boolean!` | Yes      | `true` blocks writes past the limit; `false` only warns.    |

## Remap an option (retire it)

`remapRecordFieldOption` retires an option and says what happens to the values holding it: with `targetOptionId`, every affected record's value moves onto that live option of the same field; without one, the values are cleared. Values are never left pointing at a retired option, and a single-select record never ends up holding two options after a merge. History is recorded for every record whose value changes, and dependent computed fields refresh.

The option cannot be remapped onto itself, and the target must be a live option of the same field. Retiring also removes the option from automations that name it.

```graphql
mutation RemapOption {
  remapRecordFieldOption(
    input: {
      workspaceId: "ws_123"
      fieldId: "field_123"
      optionId: "option_old"
      targetOptionId: "option_new"
    }
  ) {
    optionId
    retired
    moved
    cleared
  }
}
```

```json
{
  "data": {
    "remapRecordFieldOption": {
      "optionId": "option_old",
      "retired": true,
      "moved": 12,
      "cleared": false
    }
  }
}
```

### RemapRecordFieldOptionInput

| Parameter        | Type  | Required | Description                                                              |
| ---------------- | ----- | -------- | ------------------------------------------------------------------------ |
| `workspaceId`    | `ID!` | Yes      | The workspace that owns the field.                                       |
| `fieldId`        | `ID!` | Yes      | The field the option belongs to.                                         |
| `optionId`       | `ID!` | Yes      | The option to retire.                                                    |
| `targetOptionId` | `ID`  | No       | A live option of the same field to move values onto. Omit to clear them. |

## Remap and delete an option

`remapAndDeleteRecordFieldOption` is the option-delete dialog's write: it remaps (or clears) exactly like `remapRecordFieldOption`, then retires the definition — but it runs through a durable transaction and stores its result in a workspace receipt keyed by `clientMutationId`. A retry after a lost connection returns the original result instead of failing on an already-retired option. Reusing a `clientMutationId` with a different payload is refused with `That client mutation ID was reused.` The field must be mounted on a table in the workspace.

The caller must be authenticated. Use a unique `clientMutationId` per logical delete (a UUID works).

```graphql
mutation DeleteOption {
  remapAndDeleteRecordFieldOption(
    input: {
      workspaceId: "ws_123"
      fieldId: "field_123"
      optionId: "option_old"
      targetOptionId: "option_new"
      clientMutationId: "0b9e6c1a-2f4e-4a7b-9c3d-1e5f7a9b2c4d"
    }
  ) {
    optionId
    deleted
    cleared
    remappedRecordCount
  }
}
```

```json
{
  "data": {
    "remapAndDeleteRecordFieldOption": {
      "optionId": "option_old",
      "deleted": true,
      "cleared": false,
      "remappedRecordCount": 12
    }
  }
}
```

Omit `targetOptionId` to delete the option and clear its values: `cleared` comes back `true`.

### RemapAndDeleteRecordFieldOptionInput

| Parameter          | Type  | Required | Description                                                                                                  |
| ------------------ | ----- | -------- | ------------------------------------------------------------------------------------------------------------ |
| `workspaceId`      | `ID!` | Yes      | The workspace that owns the field.                                                                           |
| `fieldId`          | `ID!` | Yes      | The field the option belongs to.                                                                             |
| `optionId`         | `ID!` | Yes      | The option to delete.                                                                                        |
| `targetOptionId`   | `ID`  | No       | A live option of the same field to move every affected record's value onto. Omit to clear the value instead. |
| `clientMutationId` | `ID!` | Yes      | Caller-chosen id for replay protection. A retry returns the original result.                                 |

## Errors

All six return `UNAUTHENTICATED` when the caller is not signed in and `FORBIDDEN` when the workspace does not match, the field does not exist in the caller's organization, or the caller lacks the authority described above.

| Code                | When                                                                                                                                                                                    |
| ------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `BAD_USER_INPUT`    | The field is not a choice field ("Only choice fields have options.").                                                                                                                   |
| `BAD_USER_INPUT`    | The option is not live on that field ("That option is not available on this field.").                                                                                                   |
| `BAD_USER_INPUT`    | `createRecordFieldOption`: missing or over-long `optionId`, blank `title`, or the id already belongs to a different field.                                                              |
| `BAD_USER_INPUT`    | `renameRecordFieldOption`: blank `title` or an invalid `icon` slug.                                                                                                                     |
| `BAD_USER_INPUT`    | `moveRecordFieldOption` / `setRecordFieldOptionWip`: a missing or over-long id ("Invalid field option scope."), an unknown option, or a WIP `limit` that is not a non-negative integer. |
| `BAD_USER_INPUT`    | `remapRecordFieldOption` / `remapAndDeleteRecordFieldOption`: the option remapped onto itself, or the target is not a live option of the same field.                                    |
| `BAD_USER_INPUT`    | `remapAndDeleteRecordFieldOption`: a `clientMutationId` reused with a different payload.                                                                                                |
| `WIP_LIMIT_REACHED` | A later write tried to move a record into an option whose hard WIP limit is full (raised by the value write, not by `setRecordFieldOptionWip` itself).                                  |

## Permissions

- **Schema authority** — the organization's OWNER and ADMIN users, the OWNER/ADMIN of the workspace that owns the field, and holders of the field-management grant. `setRecordFieldOptionWip`, `remapRecordFieldOption`, and `remapAndDeleteRecordFieldOption` require it.
- **Field editors** — a workspace OWNER/ADMIN, or a MEMBER with edit access to the field in a table they belong to (a custom role needs an EDIT table grant and an editable permission on the field). `createRecordFieldOption`, `renameRecordFieldOption`, and `moveRecordFieldOption` accept field editors in place of schema authority. Viewer-level members never qualify.
- **Delegated create** — any caller with edit access to the field on the record named by `recordId` + `databaseId` may add an option, no schema authority needed.

## Related

- [Single-Select Field](/api/custom-fields/select-single) — the legacy option mutations and how values are set.
- [Multi-Select Field](/api/custom-fields/select-multi) — several options per record.
- [Custom Field Values](/api/custom-fields/custom-field-values) — set and read values across all field types.
- [Custom Fields Overview](/api/custom-fields) — concepts shared by every field type.
