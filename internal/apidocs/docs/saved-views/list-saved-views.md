---
title: List Saved Views
description: Read the saved views in a workspace, or fetch one by ID, via the savedViews and savedView queries.
icon: List
order: 1
---

Use the `savedViews` query to read the saved views in a workspace, and the `savedView` query to fetch a single view by ID. A saved view is a `SavedView` object: a stored workspace layout (board, database, calendar, timeline, or map) with its own filters, sorting, and column configuration. Saved views belong to a workspace (a `Project` in the API).

Both queries return only the views the caller can see: their own personal views plus every shared view in the workspace. Personal views created by other members are never returned — and `savedView` reports them as not found rather than forbidden, so a missing ID and another user's private view are indistinguishable. `savedViews` returns a `SavedViewPagination` page of `items` plus `pageInfo`, ordered by `position` descending.

```
POST https://api.blue.app/graphql
blue-token-id: YOUR_TOKEN_ID
blue-token-secret: YOUR_TOKEN_SECRET
blue-org-id: YOUR_ORG_ID
```

Company headers accept an ID or a slug. Header names are case-insensitive. See [Authentication](/api/start-guide/authentication) for the full header reference.

## Request

List the saved views in a workspace. The smallest call passes only `filter.projectId`.

```graphql
query ListSavedViews {
  savedViews(filter: { projectId: "project_123" }) {
    items {
      id
      name
      viewType
      isShared
      position
    }
    pageInfo {
      totalItems
      hasNextPage
    }
  }
}
```

<Callout variant="warning" title="projectId must be an exact ID">

For this query, `filter.projectId` must be the exact workspace ID — a slug is **not** accepted. Passing a slug returns an empty `items` array rather than an error.

</Callout>

## Parameters

### Arguments

| Argument | Type                    | Required | Description                                                                   |
| -------- | ----------------------- | -------- | ----------------------------------------------------------------------------- |
| `filter` | `SavedViewFilterInput!` | Yes      | Selects the workspace to read views from.                                     |
| `skip`   | `Int`                   | No       | Number of views to skip. Defaults to `0`.                                     |
| `take`   | `Int`                   | No       | Maximum number of views to return. Defaults to `100`. Not capped server-side. |

### SavedViewFilterInput

| Field       | Type      | Required | Description                                                  |
| ----------- | --------- | -------- | ------------------------------------------------------------ |
| `projectId` | `String!` | Yes      | The workspace ID. A slug is **not** accepted for this query. |

## Response

```json
{
  "data": {
    "savedViews": {
      "items": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "name": "Sprint Board",
          "viewType": "BOARD",
          "isShared": true,
          "position": 2
        },
        {
          "id": "clm4n8qwx000108l0h7tzbq19",
          "name": "My Backlog",
          "viewType": "DATABASE",
          "isShared": false,
          "position": 1
        }
      ],
      "pageInfo": {
        "totalItems": 2,
        "hasNextPage": false
      }
    }
  }
}
```

### SavedViewPagination

| Field      | Type            | Description                                     |
| ---------- | --------------- | ----------------------------------------------- |
| `items`    | `[SavedView!]!` | The saved views on this page.                   |
| `pageInfo` | `PageInfo!`     | Pagination metadata. See [PageInfo](#pageinfo). |

### SavedView

| Field        | Type             | Description                                                                          |
| ------------ | ---------------- | ------------------------------------------------------------------------------------ |
| `id`         | `ID!`            | Unique identifier for the saved view.                                                |
| `uid`        | `String!`        | Short human-readable identifier.                                                     |
| `name`       | `String!`        | Display name of the view.                                                            |
| `icon`       | `String`         | Icon identifier, or `null` if none is set.                                           |
| `position`   | `Float!`         | Sort position. Views are returned in descending `position` order.                    |
| `isShared`   | `Boolean!`       | `true` if the view is visible to all workspace members; `false` for a personal view. |
| `viewType`   | `SavedViewType!` | Layout type. See [SavedViewType](#savedviewtype).                                    |
| `viewConfig` | `JSON!`          | The view's filters, grouping, sorting, and column settings, as a JSON object.        |
| `createdBy`  | `User!`          | The member who created the view.                                                     |
| `project`    | `Project!`       | The workspace the view belongs to.                                                   |
| `createdAt`  | `DateTime!`      | When the view was created.                                                           |
| `updatedAt`  | `DateTime!`      | When the view was last updated.                                                      |

### SavedViewType

| Value      | Description                              |
| ---------- | ---------------------------------------- |
| `BOARD`    | Kanban-style board grouped into columns. |
| `DATABASE` | Spreadsheet-style table of records.      |
| `CALENDAR` | Records placed on a calendar by date.    |
| `TIMELINE` | Gantt-style timeline.                    |
| `MAP`      | Records plotted on a map by location.    |

### PageInfo

| Field             | Type       | Description                                  |
| ----------------- | ---------- | -------------------------------------------- |
| `totalItems`      | `Int`      | Total number of views matching the filter.   |
| `totalPages`      | `Int`      | Total number of pages at the current `take`. |
| `page`            | `Int`      | Current page number.                         |
| `perPage`         | `Int`      | Number of items per page.                    |
| `hasNextPage`     | `Boolean!` | `true` if more views follow this page.       |
| `hasPreviousPage` | `Boolean!` | `true` if views precede this page.           |

`PageInfo` also exposes `startCursor` and `endCursor`, both **deprecated** — use `skip`/`take` offset pagination instead.

## Full example

Read every field of every view in a workspace, with full pagination metadata and nested relationships.

```graphql
query ListSavedViewsDetailed {
  savedViews(filter: { projectId: "project_123" }, skip: 0, take: 50) {
    items {
      id
      uid
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
      updatedAt
    }
    pageInfo {
      totalItems
      totalPages
      page
      perPage
      hasNextPage
      hasPreviousPage
    }
  }
}
```

## Get a single saved view

Use the `savedView` query to fetch one view by its `id`. The view must be one the caller can see — their own, or a shared view in a workspace they belong to. Any other ID (nonexistent, in a workspace they don't belong to, or another member's personal view) returns the `SAVED_VIEW_NOT_FOUND` error.

```graphql
query GetSavedView {
  savedView(id: "clm4n8qwx000008l0g4oxdqn7") {
    id
    uid
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
    createdAt
    updatedAt
  }
}
```

Returns a single `SavedView` (see the [field table](#savedview) above).

```json
{
  "data": {
    "savedView": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "uid": "sv_8qwx",
      "name": "Sprint Board",
      "icon": "kanban",
      "viewType": "BOARD",
      "isShared": true,
      "position": 2,
      "viewConfig": { "groupBy": "list", "sort": [{ "field": "duedAt", "direction": "asc" }] },
      "createdBy": {
        "id": "user_123",
        "fullName": "Jordan Lee"
      },
      "createdAt": "2026-05-12T09:30:00.000Z",
      "updatedAt": "2026-05-20T14:05:00.000Z"
    }
  }
}
```

## Errors

| Code                   | When                                                                                                                                                           |
| ---------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `SAVED_VIEW_NOT_FOUND` | (`savedView` only) No visible view matches the `id`: it does not exist, is in a workspace the caller is not a member of, or is another member's personal view. |

`savedViews` does not raise a not-found error — a workspace the caller cannot see, or one with no matching views, returns an empty `items` array.

## Permissions

Any workspace member can read saved views, regardless of project role. Both queries scope results to the caller: their own personal views plus all shared views in workspaces they belong to. Personal views created by other members are never returned.

To find which view is the workspace default, read `defaultSavedView` on the `Project` type; for the caller's personal default, read `userDefaultSavedView`.

## Related

- [Create a Saved View](/api/saved-views) — create a new view.
- [Update Saved View](/api/saved-views/update-saved-view) — edit, reorder, or set a default view.
- [Delete Saved View](/api/saved-views/delete-saved-view) — remove a view.
- [List Workspaces](/api/workspaces/list-workspaces) — resolve workspace IDs.
