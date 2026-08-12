---
title: Saved Views
description: Store and share workspace layouts - board, table, calendar, timeline, and map views with their filters, sorting, grouping, and columns.
icon: Bookmark
order: 0
---

A saved view captures how a workspace is displayed: the layout (board, table, calendar, timeline, or map), plus the filters, sorting, grouping, and column settings applied on top of it. Members switch between views to see the same records through different lenses without re-configuring anything. In the API a saved view is a `SavedView` object, and it belongs to a workspace (`Project`).

Views are either **personal** (visible only to the member who created them) or **shared** (visible to everyone in the workspace). Any workspace member can create personal views; only `OWNER` and `ADMIN` members can create shared ones. The layout details live in the free-form `viewConfig` JSON, which the server validates against a fixed schema before storing.

```graphql
POST https://api.blue.app/graphql
blue-token-id: YOUR_TOKEN_ID
blue-token-secret: YOUR_TOKEN_SECRET
blue-org-id: YOUR_ORG_ID
blue-workspace-id: project_123
```

The `blue-workspace-id` header scopes requests to one workspace. Company and project headers accept either an ID or a slug. Header names are case-insensitive.

## Operations

| Operation                                                 | GraphQL                    | Description                                           |
| --------------------------------------------------------- | -------------------------- | ----------------------------------------------------- |
| Create a saved view                                       | `createSavedView` mutation | Add a view to a workspace. Documented below.          |
| [List saved views](/api/saved-views/list-saved-views)     | `savedViews` query         | Retrieve the views in a workspace, with pagination.   |
| [Update a saved view](/api/saved-views/update-saved-view) | `editSavedView` mutation   | Change a view's name, icon, sharing, or `viewConfig`. |
| [Delete a saved view](/api/saved-views/delete-saved-view) | `deleteSavedView` mutation | Permanently remove a view from a workspace.           |

Three more mutations round out the surface but don't have dedicated pages:

| Operation             | GraphQL                                                                 | Description                                                                      |
| --------------------- | ----------------------------------------------------------------------- | -------------------------------------------------------------------------------- |
| Reorder a view        | `updateSavedViewPosition(id: String!, position: Float!): SavedView!`    | Move a view to a new sidebar position.                                           |
| Set workspace default | `setWorkspaceDefaultView(projectId: String!, viewId: String): Project!` | Set a shared view as the workspace's default; pass `null` for `viewId` to clear. |
| Set personal default  | `setUserDefaultView(projectId: String!, viewId: String): Boolean!`      | Set the calling member's default view for a workspace; `null` clears it.         |

## Create a saved view

Use the `createSavedView` mutation to add a view to a workspace. The minimum is a `name`, a `viewType`, and a `viewConfig` object (which may be empty, `{}`). Everything else - an `icon` and whether the view is shared - is optional.

### Request

The smallest call creates a personal board view with no extra configuration. An empty `viewConfig: {}` is valid; the layout falls back to defaults.

```graphql
mutation CreateSavedView {
  createSavedView(
    input: { projectId: "project_123", name: "My Board", viewType: BOARD, viewConfig: {} }
  ) {
    id
    name
    viewType
    isShared
    position
  }
}
```

### Parameters

#### CreateSavedViewInput

| Parameter    | Type             | Required | Description                                                                                                               |
| ------------ | ---------------- | -------- | ------------------------------------------------------------------------------------------------------------------------- |
| `projectId`  | `String!`        | Yes      | ID or slug of the workspace to create the view in. The caller must be a member of the workspace.                          |
| `name`       | `String!`        | Yes      | Display name. Must be non-empty and at most 100 characters.                                                               |
| `viewType`   | `SavedViewType!` | Yes      | The layout. One of the `SavedViewType` values below.                                                                      |
| `viewConfig` | `JSON!`          | Yes      | Layout configuration object. May be empty (`{}`). Validated server-side - see [View configuration](#view-configuration).  |
| `icon`       | `String`         | No       | Icon identifier. At most 50 characters.                                                                                   |
| `isShared`   | `Boolean`        | No       | `true` shares the view with all workspace members. Defaults to `false` (personal). Only `OWNER` / `ADMIN` may set `true`. |

#### SavedViewType

| Value      | Layout                    |
| ---------- | ------------------------- |
| `BOARD`    | Kanban-style board.       |
| `DATABASE` | Table / spreadsheet grid. |
| `CALENDAR` | Calendar.                 |
| `TIMELINE` | Gantt-style timeline.     |
| `MAP`      | Geographic map.           |

### Response

```json
{
  "data": {
    "createSavedView": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "My Board",
      "viewType": "BOARD",
      "isShared": false,
      "position": 65535
    }
  }
}
```

#### Returns

The mutation returns the created `SavedView`.

| Field        | Type             | Description                                                        |
| ------------ | ---------------- | ------------------------------------------------------------------ |
| `id`         | `ID!`            | Unique identifier for the view.                                    |
| `uid`        | `String!`        | Short public identifier.                                           |
| `name`       | `String!`        | Display name.                                                      |
| `icon`       | `String`         | Icon identifier, or `null` if none was set.                        |
| `position`   | `Float!`         | Ordering position in the sidebar. See [Positioning](#positioning). |
| `isShared`   | `Boolean!`       | Whether the view is shared with all workspace members.             |
| `viewType`   | `SavedViewType!` | The layout.                                                        |
| `viewConfig` | `JSON!`          | The stored configuration object.                                   |
| `createdBy`  | `User!`          | The member who created the view.                                   |
| `project`    | `Project!`       | The workspace the view belongs to.                                 |
| `createdAt`  | `DateTime!`      | When the view was created.                                         |
| `updatedAt`  | `DateTime!`      | When the view was last updated.                                    |

### Full example

A shared table view filtered to a tag, sorted by due date, grouped by a custom field, with two columns hidden.

```graphql
mutation CreateSprintView {
  createSavedView(
    input: {
      projectId: "project_123"
      name: "Sprint Planning"
      icon: "layout-grid"
      isShared: true
      viewType: DATABASE
      viewConfig: {
        searchQuery: "launch"
        tagIds: ["tag_123"]
        tagMode: "or"
        sort: { type: "builtin", field: "duedAt", direction: "asc" }
        groupBy: { type: "customField", fieldId: "field_123" }
        hiddenColumns: ["field_456", "field_789"]
        pinnedColumns: { left: ["field_123"], right: [] }
      }
    }
  ) {
    id
    name
    icon
    viewType
    isShared
    position
    viewConfig
    createdBy {
      id
      fullName
    }
    project {
      id
      name
    }
    createdAt
  }
}
```

## View configuration

`viewConfig` is a JSON object, but it is **not** free-form - the server validates it against a fixed schema and rejects unknown shapes. The whole object must serialize to under 50,000 bytes. Every key below is optional; an empty `{}` is valid and the layout uses defaults. Unrecognized keys are stored but ignored on read.

### Filtering

| Key               | Type       | Description                                                                                            |
| ----------------- | ---------- | ------------------------------------------------------------------------------------------------------ |
| `searchQuery`     | `String`   | Text to match records against. At most 1,000 characters.                                               |
| `assigneeIds`     | `[String]` | Restrict to records assigned to these user IDs. At most 100 entries.                                   |
| `assigneeMode`    | `String`   | How to combine `assigneeIds`: `"and"` or `"or"`.                                                       |
| `tagIds`          | `[String]` | Restrict to records carrying these tag IDs. At most 100 entries.                                       |
| `tagMode`         | `String`   | How to combine `tagIds`: `"and"` or `"or"`.                                                            |
| `dateStart`       | `String`   | Start of a date range. At most 1,000 characters.                                                       |
| `dateEnd`         | `String`   | End of a date range. At most 1,000 characters.                                                         |
| `datePreset`      | `String`   | Named relative range (e.g. a "this week" preset). At most 1,000 characters.                            |
| `advancedFilters` | `JSON`     | Nested filter expression for complex queries. Loosely validated - only the overall size limit applies. |

### Sorting

`sort` is a single **object** (not an array), with all three fields required:

| Key         | Type     | Description                                                   |
| ----------- | -------- | ------------------------------------------------------------- |
| `type`      | `String` | One of `"builtin"`, `"custom"`, or `"customField"`.           |
| `field`     | `String` | The field to sort by (e.g. `"duedAt"`, or a custom field ID). |
| `direction` | `String` | `"asc"` or `"desc"` (lowercase).                              |

### Grouping

`groupBy` is an **object** with a required `type`:

| Key         | Type     | Description                                                                                                                                                             |
| ----------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `type`      | `String` | One of `none`, `list`, `status`, `assignee`, `priority`, `date`, `tag`, `select`, `createdDate`, `customField`, `customDateField`, `customLookupField`, `referencedBy`. |
| `fieldId`   | `String` | Custom field ID to group by (used with `customField` / `customDateField` / `customLookupField`).                                                                        |
| `fieldName` | `String` | Human-readable field label.                                                                                                                                             |
| `mode`      | `String` | For date-based grouping: `"actual"` or `"relative"`.                                                                                                                    |

### Columns

| Key             | Type                                  | Description                                                                         |
| --------------- | ------------------------------------- | ----------------------------------------------------------------------------------- |
| `hiddenColumns` | `[String]`                            | Column / field IDs to hide. At most 100 entries.                                    |
| `pinnedColumns` | `{ left: [String], right: [String] }` | Column / field IDs pinned to the left or right edge; each side at most 100 entries. |

### Timeline

These keys apply when `viewType` is `TIMELINE`:

| Key                        | Type      | Description                                                                                                                               |
| -------------------------- | --------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `timelineZoom`             | `String`  | One of `day`, `week`, `month`, `quarter`, `year`.                                                                                         |
| `timelineDepDisplay`       | `String`  | Dependency line display: `always`, `selected`, or `hidden`.                                                                               |
| `timelineShowCascade`      | `Boolean` | Whether dependent records shift when a record moves.                                                                                      |
| `timelineShowCriticalPath` | `Boolean` | Whether to highlight the critical path.                                                                                                   |
| `timelineDateConfig`       | `object`  | `{ startSource, endSource }`, each one of `createdAt`, `startedAt`, `duedAt`, `updatedAt`, `completedAt`, or `{ customFieldId: String }`. |

<Callout variant="info" title="viewType inside viewConfig">

`viewConfig` may also carry its own `viewType` key, validated against the same `SavedViewType` set. This is separate from the top-level `viewType` argument, which is the one that determines the view's layout. You can omit the nested key entirely.

</Callout>

## Errors

| Code                | When                                                                                                                    |
| ------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| `PROJECT_NOT_FOUND` | The `projectId` doesn't exist, or the caller is not a member of the workspace.                                          |
| `FORBIDDEN`         | A non-`OWNER` / non-`ADMIN` member tried to create a shared view (`isShared: true`). Message: `You are not authorized.` |

`viewConfig` validation failures are thrown as a plain `ViewConfigValidationError`, **not** an Apollo error, so the response carries no specific `code` (it surfaces as a generic server error). Validation rejects a `viewConfig` that isn't an object, exceeds 50,000 bytes, a `name` that is empty or over 100 characters, an `icon` over 50 characters, or a malformed `sort`, `groupBy`, `pinnedColumns`, or timeline shape (see [View configuration](#view-configuration)). The message describes the offending field, e.g. `sort.direction must be 'asc' or 'desc'`.

## Permissions

- **Personal views** (`isShared: false`, the default): any workspace member may create one.
- **Shared views** (`isShared: true`): only members with the `OWNER` or `ADMIN` access level may create one. Others receive `FORBIDDEN`.

## Positioning

New views are assigned a `position` of the last position in their scope plus `65535`, or `65535` for the first view in that scope. Scope is per-workspace and split by sharing: a shared view is positioned after the last shared view; a personal view is positioned after the caller's last personal view. Use `updateSavedViewPosition` to reorder afterward.

## Related

- [List saved views](/api/saved-views/list-saved-views)
- [Update a saved view](/api/saved-views/update-saved-view)
- [Delete a saved view](/api/saved-views/delete-saved-view)
- [Workspaces](/api/workspaces)
- [Custom fields](/api/custom-fields)
