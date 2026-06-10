---
title: Update Saved View
description: Edit a saved view, reorder it, and set workspace or personal default views in a Blue workspace.
icon: Pencil
order: 2
---

Use the `editSavedView` mutation to change a saved view's name, icon, sharing status, or configuration. This page also covers the three related mutations that operate on existing views: `updateSavedViewPosition` (reorder), `setWorkspaceDefaultView` (set the workspace-wide default), and `setUserDefaultView` (set your personal default). Saved views are `SavedView` objects scoped to a workspace (`Project`).

All examples are sent to `https://api.blue.app/graphql` with your authentication headers (`X-Bloo-Token-ID`, `X-Bloo-Token-Secret`, `X-Bloo-Company-ID`). Header names are case-insensitive.

<Callout variant="info" title="viewType is immutable">

`EditSavedViewInput` has no `viewType` field — a view's layout type (`BOARD`, `DATABASE`, etc.) is fixed at creation. To switch layout, create a new view with `createSavedView` and delete the old one.

</Callout>

## Request

Rename a saved view. Only the `id` is required; include only the fields you want to change.

```graphql
mutation EditSavedView {
  editSavedView(input: { id: "view_123", name: "Q1 Roadmap" }) {
    id
    name
    updatedAt
  }
}
```

## Parameters

### EditSavedViewInput

| Parameter    | Type      | Required | Description                                                                                                                                                       |
| ------------ | --------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `id`         | `String!` | Yes      | ID of the saved view to update.                                                                                                                                   |
| `name`       | `String`  | No       | New display name. Cannot be empty; max 100 characters.                                                                                                            |
| `icon`       | `String`  | No       | New icon identifier. Max 50 characters.                                                                                                                           |
| `isShared`   | `Boolean` | No       | Whether the view is shared with all workspace members. Changing this is permission-gated (see [Permissions](#permissions)).                                       |
| `viewConfig` | `JSON`    | No       | New view configuration. **Replaces the entire stored config** — send the complete object, not a partial patch. See [viewConfig structure](#viewconfig-structure). |

Omitted fields are left unchanged. `viewType` is not part of this input and cannot be edited.

## Response

`editSavedView` returns the full updated `SavedView`.

```json
{
  "data": {
    "editSavedView": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Q1 Roadmap",
      "updatedAt": "2026-05-29T14:02:11.000Z"
    }
  }
}
```

### Returns

`editSavedView`, `updateSavedViewPosition`, and `createSavedView` all return the same `SavedView!` type.

| Field        | Type             | Description                                                         |
| ------------ | ---------------- | ------------------------------------------------------------------- |
| `id`         | `ID!`            | Unique identifier of the saved view.                                |
| `uid`        | `String!`        | Short human-friendly identifier.                                    |
| `name`       | `String!`        | Display name.                                                       |
| `icon`       | `String`         | Icon identifier.                                                    |
| `position`   | `Float!`         | Sort position in the sidebar.                                       |
| `isShared`   | `Boolean!`       | Whether the view is shared with all workspace members.              |
| `viewType`   | `SavedViewType!` | Layout type: `BOARD`, `DATABASE`, `CALENDAR`, `TIMELINE`, or `MAP`. |
| `viewConfig` | `JSON!`          | The view's configuration object.                                    |
| `createdBy`  | `User!`          | The user who created the view.                                      |
| `project`    | `Project!`       | The workspace the view belongs to.                                  |
| `createdAt`  | `DateTime!`      | When the view was created.                                          |
| `updatedAt`  | `DateTime!`      | When the view was last updated.                                     |

## Full example

Update the name, icon, sharing status, and configuration in one call. The `viewConfig` below uses the fields the server actually validates.

```graphql
mutation EditSavedViewAdvanced {
  editSavedView(
    input: {
      id: "view_123"
      name: "Sprint Board"
      icon: "kanban"
      isShared: true
      viewConfig: {
        searchQuery: "design"
        assigneeIds: ["user_123"]
        assigneeMode: "or"
        tagIds: ["tag_123"]
        tagMode: "and"
        hiddenColumns: ["field_123"]
        pinnedColumns: { left: ["title"], right: [] }
        sort: { type: "customField", field: "field_123", direction: "asc" }
        groupBy: { type: "customField", fieldId: "field_123" }
        advancedFilters: { op: "and", conditions: [] }
      }
    }
  ) {
    id
    name
    icon
    isShared
    viewType
    viewConfig
    createdBy {
      id
      fullName
    }
    updatedAt
  }
}
```

### viewConfig structure

`viewConfig` is a free-form `JSON` object, but the server validates known keys and rejects the whole object if any are malformed. The total payload must stay under 50 KB; string fields max 1,000 characters; arrays max 100 items.

| Key                                  | Type       | Notes                                                                                                                                                                                                                       |
| ------------------------------------ | ---------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `searchQuery`                        | `String`   | Free-text filter.                                                                                                                                                                                                           |
| `dateStart`, `dateEnd`, `datePreset` | `String`   | Date-range filter inputs.                                                                                                                                                                                                   |
| `assigneeIds`, `tagIds`              | `String[]` | Filter by assignee / tag IDs.                                                                                                                                                                                               |
| `assigneeMode`, `tagMode`            | `String`   | How multiple filters combine — `"and"` or `"or"`.                                                                                                                                                                           |
| `hiddenColumns`                      | `String[]` | Column IDs to hide. (There is no `columns` key; list what to hide, not what to show.)                                                                                                                                       |
| `pinnedColumns`                      | `Object`   | `{ "left": String[], "right": String[] }`.                                                                                                                                                                                  |
| `sort`                               | `Object`   | `{ "type": "builtin" \| "custom" \| "customField", "field": String, "direction": "asc" \| "desc" }`. All three keys are required; `direction` is lowercase.                                                                 |
| `groupBy`                            | `Object`   | `{ "type": String, "fieldId"?: String }`. `type` is one of `none`, `list`, `status`, `assignee`, `priority`, `date`, `tag`, `select`, `createdDate`, `customField`, `customDateField`, `customLookupField`, `referencedBy`. |
| `advancedFilters`                    | `Object`   | Nested filter tree. Loosely validated (size-limited only).                                                                                                                                                                  |
| `timelineZoom`                       | `String`   | `day`, `week`, `month`, `quarter`, or `year`. Timeline views only.                                                                                                                                                          |
| `timelineDepDisplay`                 | `String`   | `always`, `selected`, or `hidden`. Timeline views only.                                                                                                                                                                     |
| `timelineDateConfig`                 | `Object`   | `{ "startSource", "endSource" }`; each is a date source (`createdAt`, `startedAt`, `duedAt`, `updatedAt`, `completedAt`) or `{ "customFieldId": String }`.                                                                  |

## Update view position

Use the `updateSavedViewPosition` mutation to reorder a view in the sidebar. It takes the view `id` and a new `position: Float!`, and returns the updated `SavedView!`.

```graphql
mutation UpdateSavedViewPosition {
  updateSavedViewPosition(id: "view_123", position: 98302.5) {
    id
    name
    position
  }
}
```

```json
{
  "data": {
    "updateSavedViewPosition": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Sprint Board",
      "position": 98302.5
    }
  }
}
```

Positions are spaced by `65535` when views are created. To place a view between two others, set `position` to a value between theirs (a fractional value is fine). Reordering requires the same modify access as editing — see [Permissions](#permissions).

## Set the workspace default view

Use the `setWorkspaceDefaultView` mutation to set a **shared** view as the default for everyone in the workspace. It takes a `projectId: String!` (ID or slug) and an optional `viewId: String`, and returns the updated `Project!` with its `defaultSavedView`.

```graphql
mutation SetWorkspaceDefaultView {
  setWorkspaceDefaultView(projectId: "project_123", viewId: "view_123") {
    id
    name
    defaultSavedView {
      id
      name
      viewType
    }
  }
}
```

```json
{
  "data": {
    "setWorkspaceDefaultView": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Marketing",
      "defaultSavedView": {
        "id": "clm4n8qwx000009l0g4oxdqn7",
        "name": "Sprint Board",
        "viewType": "BOARD"
      }
    }
  }
}
```

To clear the workspace default, omit `viewId` or pass `null`:

```graphql
mutation ClearWorkspaceDefaultView {
  setWorkspaceDefaultView(projectId: "project_123", viewId: null) {
    id
    defaultSavedView {
      id
    }
  }
}
```

Both setting and clearing require the caller to be an `OWNER` or `ADMIN` of the workspace. A `MEMBER` (or lower) gets `PROJECT_NOT_FOUND`, not `FORBIDDEN` — the workspace lookup is scoped to OWNER/ADMIN membership, so a member appears to have "no such project." The target view must be shared and belong to this workspace, or the call returns `FORBIDDEN`.

## Set your personal default view

Use the `setUserDefaultView` mutation to set your own default view for a workspace. It takes a `projectId: String!` (ID or slug) and an optional `viewId: String`, and returns `Boolean!`. The view may be personal or shared, but it must belong to the workspace. Any workspace member can set their own default.

```graphql
mutation SetUserDefaultView {
  setUserDefaultView(projectId: "project_123", viewId: "view_123")
}
```

```json
{ "data": { "setUserDefaultView": true } }
```

To clear your personal default, pass `null` for `viewId`:

```graphql
mutation ClearUserDefaultView {
  setUserDefaultView(projectId: "project_123", viewId: null)
}
```

Passing a `viewId` for a view that belongs to a different workspace returns `FORBIDDEN`.

## Errors

| Code                    | When                                                                                                                                                                                                                                                                                                                                           |
| ----------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `SAVED_VIEW_NOT_FOUND`  | The view ID does not exist, or you lack access to its workspace. Applies to `editSavedView` and `updateSavedViewPosition`.                                                                                                                                                                                                                     |
| `FORBIDDEN`             | You don't have modify access to the view; you tried to share a view without OWNER/ADMIN; or (for the default-view mutations) the target view is not shared / not in the workspace. Message: `You are not authorized.`                                                                                                                          |
| `PROJECT_NOT_FOUND`     | The workspace doesn't exist, or — for `setWorkspaceDefaultView` — you are not an OWNER/ADMIN of it. Also raised by `setUserDefaultView` when you're not a member. Message: `Project was not found.`                                                                                                                                            |
| `INTERNAL_SERVER_ERROR` | The `viewConfig` you supplied to `editSavedView` failed validation (wrong shape, over the 50 KB limit, or a bad enum value). Validation failures aren't modeled as a dedicated code, and the message is masked to `Internal server error` in production — validate `viewConfig` against the [structure](#viewconfig-structure) before sending. |

```json
{
  "errors": [
    {
      "message": "You are not authorized.",
      "extensions": { "code": "FORBIDDEN" }
    }
  ]
}
```

## Permissions

`editSavedView` and `updateSavedViewPosition` both require **modify access** to the view first (`findModifiableSavedView`):

- The **creator** can modify their own view (personal or shared).
- An `OWNER` or `ADMIN` can modify a view **only if it is already shared**.

An admin who is not the creator **cannot edit a private personal view** that belongs to another user — the modify-access check throws `FORBIDDEN` before any field-level check runs. So an admin cannot convert someone else's personal view to shared.

Changing `isShared` adds a second check on top of modify access:

| Change                                | Who can do it                                                                              |
| ------------------------------------- | ------------------------------------------------------------------------------------------ |
| Set `isShared: true` (make shared)    | `OWNER` or `ADMIN` only — and they must already have modify access (i.e. they created it). |
| Set `isShared: false` (make personal) | The creator, or an `OWNER`/`ADMIN`.                                                        |

For the default-view mutations:

- `setWorkspaceDefaultView` — `OWNER` or `ADMIN` only (both setting and clearing). Others receive `PROJECT_NOT_FOUND`.
- `setUserDefaultView` — any workspace member, for their own default.

## Related

- [Create a saved view](/api/saved-views)
- [List saved views](/api/saved-views/list-saved-views)
- [Delete a saved view](/api/saved-views/delete-saved-view)
- [List records](/api/records/list-records)
