---
title: Update and Delete a Connection
description: Rename a connection, delete it, and subscribe to real-time create/update/delete events for a workspace.
icon: Settings
order: 3
---

Use the `updateOAuthConnection` and `deleteOAuthConnection` mutations to maintain an existing OAuth connection, and the `subscribeToOAuthConnection` subscription to react to changes in real time. Connections store OAuth2 credentials at the workspace level (workspaces are `Project` objects in the API) so HTTP automation actions can call third-party APIs as an authenticated user.

`updateOAuthConnection` only renames a connection. The stored tokens are write-only and immutable through these operations — see [Rotating tokens](#rotating-tokens) below.

## Rename a connection

Use `updateOAuthConnection` to change a connection's display name. The `name` is the only mutable field; the operation returns the full `OAuthConnection`.

### Request

```graphql
mutation RenameConnection {
  updateOAuthConnection(input: { id: "oauth_123", name: "GitHub (bot account)" }) {
    id
    name
    provider
    updatedAt
  }
}
```

### Parameters

#### UpdateOAuthConnectionInput

| Parameter | Type     | Required | Description                                                                             |
| --------- | -------- | -------- | --------------------------------------------------------------------------------------- |
| `id`      | `ID!`    | Yes      | The ID of the connection to rename. Matched by `id` only — slugs are not accepted here. |
| `name`    | `String` | No       | The new display name. Omitting it (or sending `null`) leaves the name unchanged.        |

There is no input field for the access token, refresh token, provider, expiry, or metadata — `name` is the only thing this mutation can change.

### Response

```json
{
  "data": {
    "updateOAuthConnection": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "GitHub (bot account)",
      "provider": "GITHUB",
      "updatedAt": "2026-05-29T10:42:00.000Z"
    }
  }
}
```

#### Returns

Returns the updated `OAuthConnection`. Tokens are never included on this type. See [List OAuth Connections](/api/oauth/list-oauth-connections) for the full field reference.

| Field       | Type             | Description                                                                                        |
| ----------- | ---------------- | -------------------------------------------------------------------------------------------------- |
| `id`        | `ID!`            | The connection's unique identifier.                                                                |
| `uid`       | `String!`        | Short human-readable identifier.                                                                   |
| `name`      | `String!`        | The current display name.                                                                          |
| `provider`  | `OAuthProvider!` | `GITHUB` or `INUIT_QUICKBOOKS`.                                                                    |
| `expiredAt` | `DateTime`       | Token expiry you supplied on create. Not changed by this mutation; nothing here auto-refreshes it. |
| `metadata`  | `JSON`           | Free-form, provider-specific metadata.                                                             |
| `project`   | `Project!`       | The workspace the connection belongs to.                                                           |
| `createdBy` | `User!`          | The user who created the connection.                                                               |
| `createdAt` | `DateTime!`      | When the connection was created.                                                                   |
| `updatedAt` | `DateTime!`      | When the connection was last updated.                                                              |

## Delete a connection

Use `deleteOAuthConnection` to permanently remove a connection. This is a hard delete — the row and its stored tokens are gone immediately, and any HTTP automation action still referencing it will fail to authenticate on its next run.

### Request

```graphql
mutation DeleteConnection {
  deleteOAuthConnection(id: "oauth_123")
}
```

`deleteOAuthConnection` returns a bare `Boolean`, so it takes no sub-selection. On success the resolver always returns `true`.

### Parameters

| Parameter | Type      | Required | Description                                               |
| --------- | --------- | -------- | --------------------------------------------------------- |
| `id`      | `String!` | Yes      | The ID of the connection to delete. Matched by `id` only. |

### Response

```json
{
  "data": {
    "deleteOAuthConnection": true
  }
}
```

<Callout variant="warning" title="Deletion is permanent">

There is no undo and no archive state. Before deleting, repoint any HTTP automation action that uses this connection (the action's <code>oauthConnectionId</code>) to a replacement, or those actions will stop authenticating.

</Callout>

## Rotating tokens

There is no mutation to change a connection's access or refresh token. `accessToken` and `refreshToken` are accepted only on [`createOAuthConnection`](/api/oauth/create-oauth-connection) and are never returned on the `OAuthConnection` type, and `updateOAuthConnection` edits `name` only.

To rotate credentials, delete the connection and create a new one with the fresh tokens:

```graphql
mutation RotateConnection {
  deleteOAuthConnection(id: "oauth_123")
}
```

```graphql
mutation CreateRotated {
  createOAuthConnection(
    input: {
      projectId: "project_123"
      name: "GitHub (bot account)"
      provider: GITHUB
      accessToken: "NEW_ACCESS_TOKEN"
      refreshToken: "NEW_REFRESH_TOKEN"
    }
  ) {
    id
    name
  }
}
```

If you create the replacement first, update any HTTP automation actions to the new connection's `id` before deleting the old one to avoid a window where actions can't authenticate.

## Subscribe to changes

Use the `subscribeToOAuthConnection` subscription to receive a real-time event whenever a connection in a workspace is created, updated, or deleted. Subscriptions run over WebSocket at `wss://api.blue.app/graphql`.

### Request

```graphql
subscription OnConnectionChange {
  subscribeToOAuthConnection(input: { projectId: "project_123" }) {
    mutation
    node {
      id
      name
      provider
      updatedAt
    }
    previousValues {
      id
      name
    }
  }
}
```

### Parameters

#### SubscribeToOAuthConnectionInput

| Parameter   | Type      | Required | Description                                                                                                       |
| ----------- | --------- | -------- | ----------------------------------------------------------------------------------------------------------------- |
| `projectId` | `String!` | Yes      | The workspace to watch. Only events for connections in this workspace are delivered, and only to project members. |

### Response

Each event is an `OAuthConnectionSubscriptionPayload`.

| Field            | Type              | Description                                                          |
| ---------------- | ----------------- | -------------------------------------------------------------------- |
| `mutation`       | `MutationType!`   | The kind of change: `CREATED`, `UPDATED`, or `DELETED`.              |
| `node`           | `OAuthConnection` | The connection's current state. Present for `CREATED` and `UPDATED`. |
| `previousValues` | `OAuthConnection` | The connection's prior state, for `UPDATED` and `DELETED` events.    |

```json
{
  "data": {
    "subscribeToOAuthConnection": {
      "mutation": "UPDATED",
      "node": {
        "id": "clm4n8qwx000008l0g4oxdqn7",
        "name": "GitHub (bot account)",
        "provider": "GITHUB",
        "updatedAt": "2026-05-29T10:42:00.000Z"
      },
      "previousValues": {
        "id": "clm4n8qwx000008l0g4oxdqn7",
        "name": "GitHub"
      }
    }
  }
}
```

As with the rest of the type, `node` and `previousValues` never carry token fields. Events are scoped to the `projectId` you subscribe to — a change in another workspace is never delivered to this stream.

## Errors

| Code                         | When                                                                                              |
| ---------------------------- | ------------------------------------------------------------------------------------------------- |
| `UNAUTHENTICATED`            | The request is missing or carries invalid credentials.                                            |
| `OAUTH_CONNECTION_NOT_FOUND` | No connection matches the supplied `id` (update and delete only).                                 |
| `FORBIDDEN`                  | The caller is not a member of the connection's (or, for the subscription, the input's) workspace. |

## Permissions

`updateOAuthConnection` and `deleteOAuthConnection` require an authenticated caller who is a member of the connection's workspace — any membership level is sufficient (the resolver checks the workspace has at least one membership row for the caller). The `subscribeToOAuthConnection` stream only delivers events to members of the subscribed workspace.

This is more permissive than [listing connections](/api/oauth/list-oauth-connections), which is restricted to `ADMIN` and `OWNER` members. Any member can rename or delete a connection, but only admins and owners can see the list.

## Related

- [OAuth Connections](/api/oauth) — section overview.
- [Create an OAuth Connection](/api/oauth/create-oauth-connection) — store a credential from tokens you already hold.
- [List OAuth Connections](/api/oauth/list-oauth-connections) — query connections in a workspace (full `OAuthConnection` field reference).
- [Create Automation](/api/automations/create-automation) — set an HTTP action's `authorizationType` to `OAUTH2` and `oauthConnectionId` to consume a connection.
