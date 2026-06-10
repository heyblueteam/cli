---
title: List OAuth Connections
description: Query the OAuth connections in a workspace with the oauthConnections query — filter by workspace, sort, and paginate.
icon: List
order: 2
---

Use the `oauthConnections` query to retrieve a paginated list of the OAuth connections stored in your organization. Each connection holds an OAuth2 credential that HTTP automation actions use to call a third-party API as an authenticated user. Filter by workspace, sort by name or last-updated, and page through results with `skip`/`take`.

Workspaces are `Project` objects in the API, so the `filter.projectId` argument scopes the query to a single workspace.

<Callout variant="warning" title="Admins and owners only">

This query is more restrictive than the create/update/delete mutations. Those allow **any** member of the workspace, but `oauthConnections` returns a connection only when you are a member of its workspace at the `ADMIN` or `OWNER` level. Members at lower levels get an empty list, not an error.

</Callout>

## Request

The smallest call lists every connection you can see across your organization, newest-updated first:

```graphql
query ListOAuthConnections {
  oauthConnections {
    items {
      id
      uid
      name
      provider
      expiredAt
      updatedAt
    }
    pageInfo {
      totalItems
      totalPages
      hasNextPage
      hasPreviousPage
    }
  }
}
```

## Parameters

### `oauthConnections` arguments

| Parameter | Type                         | Default          | Description                                              |
| --------- | ---------------------------- | ---------------- | -------------------------------------------------------- |
| `filter`  | `OAuthConnectionFilterInput` | —                | Optional filter to narrow results to a single workspace. |
| `sort`    | `OAuthConnectionSort`        | `updatedAt_DESC` | Result ordering. See the enum below.                     |
| `skip`    | `Int`                        | `0`              | Number of items to skip before returning results.        |
| `take`    | `Int`                        | `20`             | Number of items to return per page.                      |

### OAuthConnectionFilterInput

| Field       | Type     | Required | Description                                                                                |
| ----------- | -------- | -------- | ------------------------------------------------------------------------------------------ |
| `projectId` | `String` | No       | Restrict results to a single workspace by its ID. Omit to list across all your workspaces. |

### OAuthConnectionSort

| Value            | Description                                        |
| ---------------- | -------------------------------------------------- |
| `name_ASC`       | Sort alphabetically by connection name, ascending. |
| `updatedAt_DESC` | Sort by last-updated time, newest first. Default.  |

Any other value falls back to creation time, newest first.

## Response

```json
{
  "data": {
    "oauthConnections": {
      "items": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "uid": "oac_8qwx000008l0",
          "name": "GitHub deploy bot",
          "provider": "GITHUB",
          "expiredAt": "2026-08-21T14:02:11.000Z",
          "updatedAt": "2026-05-21T14:02:11.000Z"
        }
      ],
      "pageInfo": {
        "totalItems": 1,
        "totalPages": 1,
        "hasNextPage": false,
        "hasPreviousPage": false
      }
    }
  }
}
```

### OAuthConnection

| Field       | Type             | Description                                                                                                  |
| ----------- | ---------------- | ------------------------------------------------------------------------------------------------------------ |
| `id`        | `ID!`            | Unique identifier for the connection.                                                                        |
| `uid`       | `String!`        | Short, user-friendly identifier.                                                                             |
| `name`      | `String!`        | Human-readable name for the connection.                                                                      |
| `provider`  | `OAuthProvider!` | The third-party provider: `GITHUB` or `INUIT_QUICKBOOKS`.                                                    |
| `expiredAt` | `DateTime`       | Token expiry supplied at creation. Informational only — nothing in the API auto-refreshes it. May be `null`. |
| `metadata`  | `JSON`           | Free-form, provider-specific JSON stored with the connection. May be `null`.                                 |
| `project`   | `Project!`       | The workspace this connection belongs to.                                                                    |
| `createdBy` | `User!`          | The user who created the connection.                                                                         |
| `createdAt` | `DateTime!`      | When the connection was created.                                                                             |
| `updatedAt` | `DateTime!`      | When the connection was last updated.                                                                        |

<Callout variant="info" title="Tokens are write-only">

The stored `accessToken` and `refreshToken` are **never** returned on `OAuthConnection` — they exist only on `CreateOAuthConnectionInput`. There is no field to read a token back and no mutation to rotate one. To change tokens, delete the connection and create a new one.

</Callout>

### PageInfo

| Field             | Type       | Description                                                   |
| ----------------- | ---------- | ------------------------------------------------------------- |
| `totalItems`      | `Int`      | Total number of connections matching the query.               |
| `totalPages`      | `Int`      | Total number of pages at the current `take` size.             |
| `page`            | `Int`      | Current page number.                                          |
| `perPage`         | `Int`      | Number of items per page.                                     |
| `hasNextPage`     | `Boolean!` | Whether more results exist after this page.                   |
| `hasPreviousPage` | `Boolean!` | Whether results exist before this page.                       |
| `startCursor`     | `String`   | Deprecated. Not used by `skip`/`take` pagination — ignore it. |
| `endCursor`       | `String`   | Deprecated. Not used by `skip`/`take` pagination — ignore it. |

## Full example

Scope to one workspace, sort by name, page through results, and select the full field set:

```graphql
query ListWorkspaceOAuthConnections(
  $filter: OAuthConnectionFilterInput
  $sort: OAuthConnectionSort
  $skip: Int
  $take: Int
) {
  oauthConnections(filter: $filter, sort: $sort, skip: $skip, take: $take) {
    items {
      id
      uid
      name
      provider
      expiredAt
      metadata
      project {
        id
        name
      }
      createdBy {
        id
        fullName
        email
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

Variables:

```json
{
  "filter": { "projectId": "project_123" },
  "sort": "name_ASC",
  "skip": 0,
  "take": 10
}
```

## Permissions

The query has no separate visibility for connections you cannot reach: it filters at the database level to connections whose workspace lists you as a member at the `ADMIN` or `OWNER` level. If you are a member at a lower level — or not a member of the workspace at all — those connections simply do not appear in `items`, and the query returns successfully with a shorter (or empty) list.

This is intentionally stricter than the connection mutations, where any workspace member can create, rename, or delete a connection.

## Related

- [Create an OAuth connection](/api/oauth/create-oauth-connection)
- [Update and delete a connection](/api/oauth/manage-oauth-connection)
- [Create an automation](/api/automations/create-automation) — consume a connection as the `OAUTH2` credential on an HTTP action
- [OAuth Connections overview](/api/oauth)
