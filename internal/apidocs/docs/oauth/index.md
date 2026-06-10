---
title: OAuth Connections
description: Store OAuth2 credentials at the workspace level so HTTP automation actions can call third-party APIs as an authenticated user.
icon: KeyRound
order: 0
---

An OAuth connection stores an OAuth2 credential — an access token, optionally a refresh token, and provider metadata — inside a workspace, so that [HTTP automation actions](/api/automations/create-automation) can call a third-party API as an authenticated user. You bring tokens you already obtained from the provider; Blue does not run the OAuth authorization flow for you. The supported providers are GitHub and Intuit QuickBooks.

OAuth connections are `OAuthConnection` objects in the API, scoped to a workspace (`Project`). A connection is consumed by an automation's `MAKE_HTTP_REQUEST` action: set the HTTP option's `authorizationType` to `OAUTH2` and reference the connection by `oauthConnectionId`, and the action sends the connection's token on each request.

<Callout variant="warning" title="Tokens are write-only">

`accessToken` and `refreshToken` are accepted only when you [create a connection](/api/oauth/create-oauth-connection). They are **never** returned on the `OAuthConnection` type, and there is no rotation mutation — [`updateOAuthConnection`](/api/oauth/manage-oauth-connection) only renames a connection. To replace a token, delete the connection and create a new one.

</Callout>

## Operations

| Operation                                                            | Mutation / Query / Subscription                                                  | Description                                                                       |
| -------------------------------------------------------------------- | -------------------------------------------------------------------------------- | --------------------------------------------------------------------------------- |
| [Create an OAuth connection](/api/oauth/create-oauth-connection)     | `createOAuthConnection`                                                          | Store an access token (and optional refresh token) for a provider in a workspace. |
| [List OAuth connections](/api/oauth/list-oauth-connections)          | `oauthConnections`                                                               | Page through the connections in a workspace, with optional filtering and sort.    |
| [Update and delete a connection](/api/oauth/manage-oauth-connection) | `updateOAuthConnection` / `deleteOAuthConnection` / `subscribeToOAuthConnection` | Rename a connection, remove one, or subscribe to real-time connection changes.    |

## Providers

`OAuthProvider` has exactly two values:

| Value              | Provider          |
| ------------------ | ----------------- |
| `GITHUB`           | GitHub            |
| `INUIT_QUICKBOOKS` | Intuit QuickBooks |

`INUIT_QUICKBOOKS` is spelled exactly as shown — pass it verbatim.

## Authorization

The two authorization rules differ between the mutations and the list query:

- **Mutations** ([create](/api/oauth/create-oauth-connection), [update, delete](/api/oauth/manage-oauth-connection)) require the caller to be **any member** of the connection's workspace.
- **Listing** ([`oauthConnections`](/api/oauth/list-oauth-connections)) returns only connections in workspaces where the caller is a member at **`ADMIN` or `OWNER`** level — more restrictive than the mutations.

Blue does not refresh tokens for you. `expiredAt` is whatever expiry you supply on create; nothing in these operations renews or rotates a token automatically.

## The OAuthConnection type

| Field       | Type             | Description                                                              |
| ----------- | ---------------- | ------------------------------------------------------------------------ |
| `id`        | `ID!`            | Unique identifier. Used to update or delete the connection.              |
| `uid`       | `String!`        | Short public identifier.                                                 |
| `name`      | `String!`        | Human-readable label.                                                    |
| `provider`  | `OAuthProvider!` | `GITHUB` or `INUIT_QUICKBOOKS`.                                          |
| `expiredAt` | `DateTime`       | Token expiry you supplied on create, or `null`. Never auto-updated.      |
| `metadata`  | `JSON`           | Free-form, provider-specific metadata you supplied on create, or `null`. |
| `project`   | `Project!`       | The workspace the connection belongs to.                                 |
| `createdBy` | `User!`          | The user who created the connection.                                     |
| `createdAt` | `DateTime!`      | Creation timestamp.                                                      |
| `updatedAt` | `DateTime!`      | Last-update timestamp.                                                   |

There is no `accessToken` or `refreshToken` field — those exist only on `CreateOAuthConnectionInput`.

## Related

- [Create an OAuth connection](/api/oauth/create-oauth-connection)
- [List OAuth connections](/api/oauth/list-oauth-connections)
- [Update and delete a connection](/api/oauth/manage-oauth-connection)
- [Create an automation](/api/automations/create-automation) — consume a connection from a `MAKE_HTTP_REQUEST` action via `oauthConnectionId`
- [Authentication](/api/start-guide/authentication)
