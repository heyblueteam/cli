---
title: List Users
description: List the people in an organization or workspace, look up a single user, or fetch the current authenticated user.
icon: List
order: 1
---

Blue exposes several queries for reading people. Use `organizationUserList` to list everyone in an organization, `workspaceUserList` to list the members of a single workspace, `user` to look up one person by ID, and `currentUser` to identify the authenticated caller. People are `User` objects in the API; organizations are `Organization` and workspaces are `Workspace`.

All of these queries require [authentication](/api/start-guide/authentication) headers. Workspace-scoped reads also need the `blue-workspace-id` header.

## Request

List the users in an organization. `companyId` accepts either the organization ID or its slug.

```graphql
query ListCompanyUsers {
  organizationUserList(companyId: "company_123") {
    users {
      id
      fullName
      email
      jobTitle
      lastActiveAt
    }
    totalCount
    pageInfo {
      totalItems
      hasNextPage
    }
  }
}
```

## Parameters

### `organizationUserList`

| Parameter        | Type               | Required | Description                                                                                                 |
| ---------------- | ------------------ | -------- | ----------------------------------------------------------------------------------------------------------- |
| `companyId`      | `String!`          | Yes      | Organization ID or slug. The caller must be a member of it.                                                 |
| `notInProjectId` | `String`           | No       | Exclude users who are already members of this workspace. Useful for building an "add member" picker.        |
| `search`         | `String`           | No       | Filter by first or last name (substring match). Does not match email.                                       |
| `skip`           | `Int`              | No       | Number of users to skip. This is the only parameter that advances the page (see [Pagination](#pagination)). |
| `after`          | `String`           | No       | Accepted but ignored by the resolver.                                                                       |
| `before`         | `String`           | No       | Accepted but ignored by the resolver.                                                                       |
| `first`          | `Int`              | No       | Accepted but ignored — every request returns at most 200 users.                                             |
| `last`           | `Int`              | No       | Accepted but ignored by the resolver.                                                                       |
| `orderBy`        | `UserOrderByInput` | No       | Accepted but ignored — results are always sorted by first name (A→Z).                                       |

### `workspaceUserList`

Lists the members of a single workspace. Same return shape and the same pagination behavior as `organizationUserList`.

| Parameter                                         | Type      | Required | Description                                           |
| ------------------------------------------------- | --------- | -------- | ----------------------------------------------------- |
| `projectId`                                       | `String!` | Yes      | Workspace ID or slug.                                 |
| `search`                                          | `String`  | No       | Filter by first or last name (substring match).       |
| `skip`                                            | `Int`     | No       | Number of users to skip.                              |
| `after` / `before` / `first` / `last` / `orderBy` | —         | No       | Accepted but ignored, as with `organizationUserList`. |

### `user`

Looks up a single user by ID.

| Parameter | Type      | Required | Description                                          |
| --------- | --------- | -------- | ---------------------------------------------------- |
| `id`      | `String!` | Yes      | User ID. Throws `USER_NOT_FOUND` if no user matches. |

### `currentUser`

Takes no arguments. Returns the `User!` for the authenticated caller — the canonical "who am I" query.

<Callout variant="info" title="Pagination is offset-based">

Despite accepting cursor arguments, both `organizationUserList` and `workspaceUserList` ignore `first`, `last`, `after`, `before`, and `orderBy`. Each request returns at most **200** users sorted by first name. To page through a larger organization, increase `skip` by 200 on each call and stop when `pageInfo.hasNextPage` is `false`.

</Callout>

## Response

```json
{
  "data": {
    "organizationUserList": {
      "users": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "fullName": "Ada Lovelace",
          "email": "ada@example.com",
          "jobTitle": "Engineering Lead",
          "lastActiveAt": "2026-05-28T09:14:00.000Z"
        }
      ],
      "totalCount": 42,
      "pageInfo": {
        "totalItems": 42,
        "hasNextPage": true
      }
    }
  }
}
```

### Returns

`organizationUserList` and `workspaceUserList` both return the same object.

| Field        | Type        | Description                                             |
| ------------ | ----------- | ------------------------------------------------------- |
| `users`      | `[User!]!`  | The page of users (see [User object](#user-object)).    |
| `totalCount` | `Int!`      | Total users matching the query across all pages.        |
| `pageInfo`   | `PageInfo!` | Offset-pagination metadata (see [PageInfo](#pageinfo)). |

### User object

The most useful fields on `User`. Full definition is in the schema.

| Field                                | Type              | Description                                                                                                     |
| ------------------------------------ | ----------------- | --------------------------------------------------------------------------------------------------------------- |
| `id`                                 | `ID!`             | Unique user identifier.                                                                                         |
| `uid`                                | `String!`         | Firebase authentication UID.                                                                                    |
| `username`                           | `String!`         | The user's username.                                                                                            |
| `email`                              | `String!`         | Email address. Blanked to an empty string unless the caller is the organization OWNER or a project OWNER/ADMIN. |
| `firstName`                          | `String!`         | First name.                                                                                                     |
| `lastName`                           | `String!`         | Last name.                                                                                                      |
| `fullName`                           | `String!`         | First and last name combined.                                                                                   |
| `jobTitle`                           | `String`          | Professional title.                                                                                             |
| `phoneNumber`                        | `String`          | Contact number.                                                                                                 |
| `locale`                             | `Locale!`         | Language preference (an enum, e.g. `EN`, `ES`, `FR`, `DE`, `JA`).                                               |
| `timezone`                           | `String`          | IANA timezone, e.g. `Europe/Rome`.                                                                              |
| `image`                              | `Image`           | Profile picture. Select a size: `thumbnail`, `small`, `medium`, `large`, or `original`.                         |
| `isEmailVerified`                    | `Boolean!`        | Whether the email address is verified.                                                                          |
| `isOnline`                           | `Boolean!`        | Whether the user is currently connected.                                                                        |
| `lastActiveAt`                       | `DateTime`        | Timestamp of last activity.                                                                                     |
| `createdAt`                          | `DateTime!`       | Account creation timestamp.                                                                                     |
| `updatedAt`                          | `DateTime!`       | Last profile update timestamp.                                                                                  |
| `companyLevel(companyId: String)`    | `UserAccessLevel` | The user's access level in the given organization.                                                              |
| `projectLevel(projectId: String)`    | `UserAccessLevel` | The user's access level in the given workspace.                                                                 |
| `projectUserRole(projectId: String)` | `ProjectUserRole` | The user's [custom role](/api/user-management/retrieve-custom-role) in the given workspace, if one is assigned. |

Per-workspace access and role are not separate node types — they are computed fields on `User` that take the workspace as an argument. `UserAccessLevel` is one of `OWNER`, `ADMIN`, `MEMBER`, `CLIENT`, `COMMENT_ONLY`, `VIEW_ONLY`.

### PageInfo

| Field             | Type       | Description                                       |
| ----------------- | ---------- | ------------------------------------------------- |
| `totalItems`      | `Int`      | Total users matching the query.                   |
| `totalPages`      | `Int`      | Total pages, derived from the 200-per-page limit. |
| `page`            | `Int`      | Current page number.                              |
| `perPage`         | `Int`      | Page size (200).                                  |
| `hasNextPage`     | `Boolean!` | Whether more users remain after this page.        |
| `hasPreviousPage` | `Boolean!` | Whether a previous page exists.                   |
| `startCursor`     | `String`   | Deprecated. Do not use.                           |
| `endCursor`       | `String`   | Deprecated. Do not use.                           |

## Full example

List the members of one workspace, reading each person's access level and custom role for that workspace, and page past the first 200 with `skip`.

```graphql
query ListProjectUsers {
  workspaceUserList(projectId: "project_123", search: "ada", skip: 0) {
    users {
      id
      fullName
      email
      projectLevel(projectId: "project_123")
      projectUserRole(projectId: "project_123") {
        id
        name
      }
    }
    totalCount
    pageInfo {
      hasNextPage
    }
  }
}
```

To build an "add member to workspace" picker, list organization users who are not yet in the workspace:

```graphql
query AddableMembers {
  organizationUserList(companyId: "company_123", notInProjectId: "project_123") {
    users {
      id
      fullName
      email
    }
    totalCount
  }
}
```

The `currentUser` query identifies the authenticated caller and is the right way to resolve the calling token's identity and access level:

```graphql
query Me {
  currentUser {
    id
    fullName
    email
    companyLevel(companyId: "company_123")
  }
}
```

The `assignees` query returns the users who can be assigned to records in a workspace. It reads the workspace from the `blue-workspace-id` header unless you pass `filter.projectId`.

```graphql
query AssignableUsers {
  assignees(filter: { projectId: "project_123", search: "ada" }) {
    id
    fullName
    email
  }
}
```

## Errors

| Code                | When                                                                                                  |
| ------------------- | ----------------------------------------------------------------------------------------------------- |
| `COMPANY_NOT_FOUND` | `organizationUserList` was called with a `companyId` you are not a member of, or that does not exist. |
| `USER_NOT_FOUND`    | `user(id:)` was called with an ID that matches no user.                                               |
| `FORBIDDEN`         | `assignees` was called for a workspace you are not a member of.                                       |
| `UNAUTHENTICATED`   | The request is missing or has invalid authentication headers.                                         |

## Permissions

- `organizationUserList` requires membership in the target organization. A non-OWNER caller only sees users who share at least one workspace with them.
- `workspaceUserList` requires membership in the target workspace.
- Email addresses are returned only to the organization OWNER and to project OWNER/ADMIN users; everyone else receives an empty string for `email`.

## Related

- [Invite a user](/api/user-management/invite-user)
- [Remove a user](/api/user-management/remove-user)
- [Retrieve a custom role](/api/user-management/retrieve-custom-role)
- [Record assignees](/api/records/assignees)
- [User Management overview](/api/user-management)
