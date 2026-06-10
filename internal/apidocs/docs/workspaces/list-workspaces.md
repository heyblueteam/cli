---
title: List Workspaces
description: Retrieve workspaces in an organization with filtering, sorting, and pagination using the projectList query.
icon: List
order: 2
---

Use the `projectList` query to retrieve the workspaces in one or more organizations, with filtering, sorting, and pagination. Workspaces are `Project` objects in the API, and an organization is a `Company`. By default `projectList` returns the workspaces the authenticated user is a member of; organization owners can also list workspaces they are _not_ in (see [Permissions](#permissions)).

## Request

The smallest call passes the organizations to search in via the required `companyIds` filter. Each entry is an organization ID or slug.

```graphql
query ListWorkspaces {
  projectList(filter: { companyIds: ["company_123"] }) {
    items {
      id
      name
      slug
      archived
    }
    pageInfo {
      totalItems
      hasNextPage
    }
  }
}
```

## Parameters

`projectList` takes a required `filter`, plus optional `sort`, `skip`, and `take`.

| Argument | Type                 | Default | Description                                     |
| -------- | -------------------- | ------- | ----------------------------------------------- |
| `filter` | `ProjectListFilter!` | —       | Which workspaces to return. See below.          |
| `sort`   | `[ProjectSort!]`     | —       | Sort order. Multiple keys are applied in order. |
| `skip`   | `Int`                | `0`     | Number of workspaces to skip (offset).          |
| `take`   | `Int`                | `20`    | Number of workspaces to return (page size).     |

<Callout variant="warning" title="Deprecated arguments">

`orderBy`, `after`, `before`, `first`, and `last` are deprecated. Use `sort` for ordering and `skip`/`take` for pagination.

</Callout>

### ProjectListFilter

| Field        | Type         | Required | Description                                                                                                             |
| ------------ | ------------ | -------- | ----------------------------------------------------------------------------------------------------------------------- |
| `companyIds` | `[String!]!` | Yes      | Organization IDs or slugs to search within.                                                                             |
| `ids`        | `[String!]`  | No       | Return only the workspaces with these IDs.                                                                              |
| `archived`   | `Boolean`    | No       | Filter by archived status.                                                                                              |
| `isTemplate` | `Boolean`    | No       | Filter by template status.                                                                                              |
| `search`     | `String`     | No       | Match workspaces whose name contains this substring (case-insensitive).                                                 |
| `folderId`   | `String`     | No       | Return only workspaces in this folder. Pass `null` for root-level workspaces (those not in any folder).                 |
| `inProject`  | `Boolean`    | No       | Membership scope. See the note below.                                                                                   |
| `userIds`    | `[String!]`  | No       | Return only workspaces where every listed user is a member. Also populates `Project.members` with the matching members. |

**`inProject` membership scope**

- `true` or omitted: return workspaces the authenticated user is a member of.
- `false`: return workspaces the user is **not** a member of. Requires the user to be an owner of at least one of the supplied organizations, otherwise the query throws `FORBIDDEN`.

### ProjectSort

| Value                              | Description                                                                                                   |
| ---------------------------------- | ------------------------------------------------------------------------------------------------------------- |
| `id_ASC` / `id_DESC`               | Sort by ID.                                                                                                   |
| `name_ASC` / `name_DESC`           | Sort by name (A–Z / Z–A).                                                                                     |
| `createdAt_ASC` / `createdAt_DESC` | Sort by creation date.                                                                                        |
| `updatedAt_ASC` / `updatedAt_DESC` | Sort by last update.                                                                                          |
| `position_ASC` / `position_DESC`   | Sort by the user's manual ordering. Only honored when listing the user's own workspaces — see the note below. |

## Response

`projectList` returns a `ProjectPagination`: the matched workspaces in `items`, page metadata in `pageInfo`, and the total match count in `totalCount` (a sibling of `pageInfo`, not a field inside it).

```json
{
  "data": {
    "projectList": {
      "items": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "name": "Marketing Campaigns",
          "slug": "marketing-campaigns",
          "archived": false
        },
        {
          "id": "clm4n9b2y000108l0a1b2c3d4",
          "name": "Product Roadmap",
          "slug": "product-roadmap",
          "archived": false
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

### ProjectPagination

| Field        | Type          | Description                                                       |
| ------------ | ------------- | ----------------------------------------------------------------- |
| `items`      | `[Project!]!` | The workspaces on this page.                                      |
| `pageInfo`   | `PageInfo!`   | Pagination metadata.                                              |
| `totalCount` | `Int`         | Total number of workspaces matching the filter, across all pages. |

### PageInfo

| Field             | Type       | Description                                  |
| ----------------- | ---------- | -------------------------------------------- |
| `totalItems`      | `Int`      | Total workspaces matching the filter.        |
| `totalPages`      | `Int`      | Total number of pages at the current `take`. |
| `page`            | `Int`      | Current page number.                         |
| `perPage`         | `Int`      | Page size (mirrors `take`).                  |
| `hasNextPage`     | `Boolean!` | Whether another page follows.                |
| `hasPreviousPage` | `Boolean!` | Whether a previous page exists.              |

### Workspace fields

Each entry in `items` is a `Project`. Commonly requested fields are below; request any combination your integration needs.

| Field                                 | Type                | Description                                                                      |
| ------------------------------------- | ------------------- | -------------------------------------------------------------------------------- |
| `id`                                  | `ID!`               | Unique identifier.                                                               |
| `uid`                                 | `String!`           | Short unique identifier.                                                         |
| `slug`                                | `String!`           | URL-friendly identifier.                                                         |
| `name`                                | `String!`           | Display name.                                                                    |
| `description`                         | `String`            | Workspace description.                                                           |
| `archived`                            | `Boolean`           | Whether the workspace is archived.                                               |
| `color`                               | `String`            | Display color.                                                                   |
| `icon`                                | `String`            | Display icon.                                                                    |
| `image`                               | `Image`             | Cover image object.                                                              |
| `category`                            | `ProjectCategory!`  | Workspace category (e.g. `MARKETING`, `ENGINEERING`, `GENERAL`).                 |
| `createdAt`                           | `DateTime!`         | When the workspace was created.                                                  |
| `updatedAt`                           | `DateTime!`         | When the workspace was last updated.                                             |
| `lastAccessedAt`                      | `DateTime`          | When the user last opened the workspace.                                         |
| `position`                            | `Float!`            | The user's manual ordering position.                                             |
| `allowNotification`                   | `Boolean!`          | Whether notifications are enabled.                                               |
| `unseenActivityCount`                 | `Int!`              | Number of unseen activity items.                                                 |
| `isTemplate`                          | `Boolean!`          | Whether this workspace is a template.                                            |
| `isOfficialTemplate`                  | `Boolean!`          | Whether this is an official Blue template.                                       |
| `automationsCount(isActive: Boolean)` | `Int!`              | Number of automations; pass `isActive` to count only active (or inactive) ones.  |
| `totalFileCount`                      | `Int`               | Number of files in the workspace.                                                |
| `totalFileSize`                       | `Float`             | Combined size of all files, in bytes.                                            |
| `todoAlias`                           | `String`            | Custom label used in place of "record" in this workspace.                        |
| `company`                             | `Company!`          | The organization that owns the workspace.                                        |
| `folder`                              | `Folder`            | The folder containing the workspace, if any.                                     |
| `accessLevel(userId: String)`         | `UserAccessLevel`   | The access level of the given user (defaults to the caller) in this workspace.   |
| `members`                             | `[ProjectMember!]!` | Workspace members. Only populated when the query passed `userIds` in the filter. |
| `nameFormulaEnabled`                  | `Boolean!`          | Whether auto-generated record titles are enabled.                                |
| `nameFormulaDisplay`                  | `String`            | Human-readable form of the record name formula.                                  |

<Callout variant="info" title="Deprecated field">

`Project.unseenActivity` is deprecated. Use `unseenActivityCount` instead. `Project.customFields` on the workspace is also deprecated — read custom fields via the [`customFields` query](/api/custom-fields/list-custom-fields).

</Callout>

## Full example

This call lists active, non-template workspaces in two organizations whose names contain "marketing", sorted by manual position then name, and reads the first 50 plus the total count.

```graphql
query FilteredWorkspaces {
  projectList(
    filter: {
      companyIds: ["company_123", "acme-corp"]
      archived: false
      isTemplate: false
      search: "marketing"
      inProject: true
      folderId: null # root-level workspaces only (not in any folder)
    }
    sort: [position_ASC, name_ASC]
    skip: 0
    take: 50
  ) {
    items {
      id
      name
      slug
      position
      archived
      category
      unseenActivityCount
      automationsCount(isActive: true)
      company {
        id
        name
      }
    }
    totalCount
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

### Behavior notes

- **Folder filtering requires membership.** `folderId` is only valid when listing the user's own workspaces (`inProject` omitted or `true`). Combining `folderId` (including `folderId: null`) with `inProject: false` returns an error, because the user has no folder placement for workspaces they don't belong to.
- **Position sort needs membership.** `position_ASC` / `position_DESC` reflect the user's manual ordering. When `inProject: false`, position sorting falls back to sorting by name.
- **Non-member defaults.** With `inProject: false`, archived and template workspaces are excluded unless you set `archived` or `isTemplate` explicitly.

## Errors

| Code        | When                                                                                                |
| ----------- | --------------------------------------------------------------------------------------------------- |
| `FORBIDDEN` | `inProject: false` was requested but the user is not an owner of any of the supplied organizations. |

## Permissions

By default `projectList` returns only workspaces the authenticated user belongs to within the given organizations. Listing workspaces the user is _not_ in (`inProject: false`) requires the user to be an owner of at least one of the organizations in `companyIds`.

## Related

- [Create a workspace](/api/workspaces/create-workspace)
- [Rename a workspace](/api/workspaces/rename-workspace)
- [Archive a workspace](/api/workspaces/archive-workspace)
- [Lists](/api/workspaces/lists)
- [List records](/api/records/list-records)
