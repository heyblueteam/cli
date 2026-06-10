---
title: Create an OAuth Connection
description: Store an OAuth2 credential in a workspace by supplying tokens you already obtained, so HTTP automations can call third-party APIs as you.
icon: KeyRound
order: 1
---

Use the `createOAuthConnection` mutation to store an OAuth2 credential at the workspace level. You supply tokens you have already obtained from the provider; Blue saves them against the workspace so an [HTTP automation action](/api/automations/create-automation) can later call the provider's API as an authenticated user. Workspaces are `Project` objects in the API, so a connection always belongs to one project.

This mutation does not perform an OAuth flow for you — there is no redirect, no authorization-code exchange. You bring the `accessToken` (and optionally a `refreshToken` and its `expiredAt` expiry) that you obtained yourself, and Blue persists them.

<Callout variant="warning" title="Tokens are write-only">

`accessToken` and `refreshToken` are accepted here but are <strong>never returned</strong> — they do not appear on the `OAuthConnection` type, so you cannot read a stored token back. There is no token-rotation mutation either: `updateOAuthConnection` only changes the `name`. To replace a token, delete the connection and create a new one. See <a href="/api/oauth/manage-oauth-connection">Update and Delete a Connection</a>.

</Callout>

## Request

The smallest call supplies a workspace, a name, a provider, and the access token. `provider` is the [`OAuthProvider`](#oauthprovider) enum.

```graphql
mutation CreateOAuthConnection {
  createOAuthConnection(
    input: {
      projectId: "project_123"
      name: "GitHub deploy bot"
      provider: GITHUB
      accessToken: "YOUR_ACCESS_TOKEN"
    }
  ) {
    id
    name
    provider
    createdAt
  }
}
```

`projectId` accepts either the workspace's ID **or** its slug — the resolver matches on whichever you pass.

## Parameters

### CreateOAuthConnectionInput

| Parameter      | Type             | Required | Description                                                                                                                   |
| -------------- | ---------------- | -------- | ----------------------------------------------------------------------------------------------------------------------------- |
| `projectId`    | `String!`        | Yes      | The workspace to attach the connection to. Accepts the project ID **or** the project slug.                                    |
| `name`         | `String!`        | Yes      | A human-readable label for the connection (shown when listing connections and when picking one in an HTTP automation action). |
| `provider`     | `OAuthProvider!` | Yes      | Which third-party provider these tokens are for. See [`OAuthProvider`](#oauthprovider).                                       |
| `accessToken`  | `String!`        | Yes      | The OAuth2 access token. **Write-only** — stored but never returned.                                                          |
| `refreshToken` | `String`         | No       | The OAuth2 refresh token, if the provider issued one. **Write-only** — stored but never returned.                             |
| `expiredAt`    | `DateTime`       | No       | When the `accessToken` expires. Informational only — nothing in this API auto-refreshes the token when it lapses.             |
| `metadata`     | `JSON`           | No       | Free-form, provider-specific JSON (for example a scope list, account ID, or realm ID). Returned as-is on reads.               |

### OAuthProvider

| Value              | Description                                                                                                      |
| ------------------ | ---------------------------------------------------------------------------------------------------------------- |
| `GITHUB`           | A GitHub credential.                                                                                             |
| `INUIT_QUICKBOOKS` | An Intuit QuickBooks credential. (The enum value is spelled `INUIT_QUICKBOOKS` in the schema — use it verbatim.) |

## Response

The mutation returns the created `OAuthConnection`. Note that the response contains no token fields — only the metadata Blue stores about the connection.

```json
{
  "data": {
    "createOAuthConnection": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "GitHub deploy bot",
      "provider": "GITHUB",
      "createdAt": "2026-05-29T14:02:11.000Z"
    }
  }
}
```

### Returns

`createOAuthConnection` returns a non-null `OAuthConnection`. Its fields:

| Field       | Type             | Description                                                |
| ----------- | ---------------- | ---------------------------------------------------------- |
| `id`        | `ID!`            | The connection's unique identifier.                        |
| `uid`       | `String!`        | Short public identifier.                                   |
| `name`      | `String!`        | The label you supplied.                                    |
| `provider`  | `OAuthProvider!` | The provider the tokens are for.                           |
| `expiredAt` | `DateTime`       | The access-token expiry you supplied on create, or `null`. |
| `metadata`  | `JSON`           | The provider-specific JSON you supplied, or `null`.        |
| `project`   | `Project!`       | The workspace the connection belongs to.                   |
| `createdBy` | `User!`          | The user who created the connection.                       |
| `createdAt` | `DateTime!`      | When the connection was created.                           |
| `updatedAt` | `DateTime!`      | When the connection was last renamed.                      |

There is intentionally no `accessToken` or `refreshToken` field on this type.

## Full example

A QuickBooks connection with a refresh token, an expiry, and provider-specific metadata. This is the form you'll use for a credential you intend to keep refreshed out-of-band.

```graphql
mutation CreateQuickBooksConnection {
  createOAuthConnection(
    input: {
      projectId: "finance-ops"
      name: "QuickBooks — production realm"
      provider: INUIT_QUICKBOOKS
      accessToken: "YOUR_ACCESS_TOKEN"
      refreshToken: "YOUR_REFRESH_TOKEN"
      expiredAt: "2026-06-29T00:00:00.000Z"
      metadata: { realmId: "9341452094008632", scope: "com.intuit.quickbooks.accounting" }
    }
  ) {
    id
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
    }
  }
}
```

Here `projectId` is given as a slug (`finance-ops`) rather than an ID — both resolve to the same workspace.

## Errors

| Code                | When                                                              |
| ------------------- | ----------------------------------------------------------------- |
| `PROJECT_NOT_FOUND` | No workspace matches the given `projectId` (neither ID nor slug). |
| `FORBIDDEN`         | You are not a member of the target workspace.                     |
| `UNAUTHENTICATED`   | The request carries no valid credentials.                         |

## Permissions

Any member of the target workspace can create a connection — there is no minimum role. The resolver looks up the workspace by ID or slug and requires at least one project-membership row for the caller; otherwise it throws `FORBIDDEN`.

This is more permissive than reading connections back: only `ADMIN` and `OWNER` members can [list a workspace's connections](/api/oauth/list-oauth-connections). Renaming and deleting, like creating, are open to any workspace member.

## Related

- [OAuth Connections overview](/api/oauth)
- [List OAuth Connections](/api/oauth/list-oauth-connections)
- [Update and Delete a Connection](/api/oauth/manage-oauth-connection)
- [Create an automation](/api/automations/create-automation) — consume a connection as the `OAUTH2` credential on an HTTP action
