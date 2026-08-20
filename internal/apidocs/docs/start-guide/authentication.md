---
title: Authentication
description: Authenticate to the Blue API with a personal access token and the blue-* request headers.
icon: Key
order: 1
---

The Blue API authenticates every request with a **personal access token (PAT)** — a Token ID and a Secret you generate in the app and send as headers on each request. There are no API keys to register; a token belongs to a user and inherits that user's permissions.

Blue also runs an OAuth 2.1 authorization server, but it serves **MCP connectors only** — its tokens are not accepted here. See [Connector authorization (OAuth 2.1)](#connector-authorization-oauth-2-1) below.

The base endpoint is `https://api.blue.app/graphql` for HTTP requests and `wss://api.blue.app/graphql` for subscriptions.

## Authenticate a request

Send your credentials on every request using these headers:

```bash
curl -X POST https://api.blue.app/graphql \
  -H "Content-Type: application/json" \
  -H "blue-token-id: YOUR_TOKEN_ID" \
  -H "blue-token-secret: YOUR_TOKEN_SECRET" \
  -H "blue-org-id: YOUR_ORG_ID" \
  -d '{"query": "query Me { user { id email fullName } }"}'
```

| Header              | Required        | Description                                                                                                                                  |
| ------------------- | --------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `blue-token-id`     | Yes             | Your token's public identifier (the `uid`).                                                                                                  |
| `blue-token-secret` | Yes             | The token secret returned once at creation.                                                                                                  |
| `blue-org-id`       | Most operations | The organization to act in. Accepts the company ID or its slug.                                                                              |
| `blue-workspace-id` | Some operations | The workspace to scope to. Accepts the project ID or its slug. Required for project-scoped reads and writes (records, lists, custom fields). |

Header names are **case-insensitive**, so `blue-token-id` and `Blue-Token-Id` are equivalent.

`blue-org-id` and `blue-workspace-id` both accept either the cuid or the human-readable slug — the slug is the segment you see in the app URL: `blue.app/company/{company-slug}/project/{project-slug}/`.

<Callout variant="info" title="Legacy header names still work">

These headers were previously named `x-bloo-token-id`, `x-bloo-token-secret`, `x-bloo-company-id`, and `x-bloo-project-id` (and the older `x-company-id` / `x-project-id`). All of those are still accepted, so existing integrations keep working unchanged. Use the `blue-*` names for anything new.

</Callout>

## Create a token

Tokens are created in the app, not via the API. (The `createPersonalAccessToken` mutation exists but rejects any request made with token headers, so it can only be called from a logged-in browser session.)

1. Open Blue and click your profile avatar in the top-right corner.
2. Choose **API** from the profile menu.
3. Click **Generate a Token**.
4. Give the token a name, and optionally set an expiration date. Past the expiration date the token stops working (see [Expiration](#expiration)).
5. Click create. Blue shows you the **Token ID** and the **Secret** once.

<youtube url="https://www.youtube.com/watch?v=C-q_AqdFUzE" />

### Token ID and Secret

A token has two parts:

- **Token ID** — your token's public identifier (the `uid` field on `PersonalAccessToken`). It is an unprefixed identifier and is **not** secret. Send it as `blue-token-id`.
- **Secret** — the password half of the credential. It is **prefixed with `pat_`** for easy identification (for example, `pat_clm4n8qwx000008l0g4oxdqn7`). Send it as `blue-token-secret`.

<Callout variant="warning" title="The Secret is shown only once">

Blue stores the secret as a bcrypt hash, so it cannot be retrieved after creation — not even by Blue's team. Copy it somewhere safe the moment it's shown. If you lose it, delete the token and create a new one. Anyone with both the Token ID and Secret can act as you in Blue, so treat them like a password and never commit them to source control.

</Callout>

## Find your Company and Project IDs

Organizations (`Organization` in the API) are the top-level entity; workspaces (`Workspace`) live inside them. Both the company and project values you pass in headers can be the ID **or** the slug, and the slug is visible in the app URL:

```
blue.app/company/{company-slug}/project/{project-slug}/
```

<youtube url="https://www.youtube.com/watch?v=zLEvs6zqGTc" />

## Expiration

If you set an expiration date when creating a token, the API rejects it once that date has passed — a request with an expired token authenticates as if no credentials were sent, returning `UNAUTHENTICATED`. Tokens created without an expiration date never expire and remain valid until you delete them.

## Revoke a token

Delete a token to revoke it immediately. You can do this from the same **API** tab in the app, or with the `deletePersonalAccessToken` mutation — which, like creation, must be called from a logged-in session rather than with token headers.

```graphql
mutation RevokeToken {
  deletePersonalAccessToken(input: { id: "token_123" })
}
```

The mutation returns `true` on success and throws `PERSONAL_ACCESS_TOKEN_NOT_FOUND` if no token with that `id` belongs to you. Deleting a token takes effect on the next request.

## Connector authorization (OAuth 2.1)

Blue exposes an OAuth 2.1 authorization server so an MCP client — Claude.ai, ChatGPT, or one you build — can connect on a user's behalf without that user generating a PAT. It is reached through the MCP endpoint, not the GraphQL API:

| Endpoint                                  | Purpose                                   |
| ----------------------------------------- | ----------------------------------------- |
| `https://mcp.blue.app/mcp`                | The MCP endpoint itself.                  |
| `/.well-known/oauth-protected-resource`   | Resource metadata (RFC 9728).             |
| `/.well-known/oauth-authorization-server` | Authorization server metadata (RFC 8414). |
| `/oauth2/register`                        | Dynamic Client Registration (RFC 7591).   |
| `/oauth2/authorize`                       | Authorization request.                    |
| `/oauth2/token`                           | Token and refresh.                        |
| `/oauth2/revoke`                          | Token revocation (RFC 7009).              |

Discover the endpoints rather than hard-coding them: an unauthenticated request to `/mcp` returns `401` with a `WWW-Authenticate` header pointing at the resource metadata, and the metadata documents point on to the rest.

What a client must do:

- **Register dynamically.** Only public clients are accepted. Redirect URIs must be `https` and are matched exactly; `http://localhost` is accepted for local clients. Registration is rate-limited per IP.
- **Use PKCE with `S256`.** It is required, not optional. `code` is the only response type, and `authorization_code` and `refresh_token` are the only grant types.
- **Use each authorization code once.** A replayed code is refused.
- **Rotate refresh tokens.** The old refresh token stays usable for a 60-second grace window so a racing client is not broken. A replay outside that window revokes the whole grant.

A grant covers every organization the user belongs to — the same reach as their web session, and no more. Pass the organization per call rather than in a header.

<Callout variant="warning" title="Connector tokens do not work on the GraphQL or REST API">

An access token issued by this server is confined to the MCP endpoint. Sending one to `https://api.blue.app/graphql` or the REST API fails as if no credentials were sent. Use a personal access token for direct API integrations.

</Callout>

Access tokens last 24 hours, but revocation is immediate: Blue checks the grant on every MCP request, so disconnecting an app under **Account > Access > Connected apps** takes effect on its next call. For the end-user walkthrough, see [Claude & ChatGPT](/docs/integrations/claude-chatgpt).

## Authentication errors

| Code                              | When                                                                                                                                 |
| --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `UNAUTHENTICATED`                 | The Token ID/Secret pair is missing, the secret does not match, or the token has expired. The request is treated as unauthenticated. |
| `PERSONAL_ACCESS_TOKEN_NOT_FOUND` | A token-management mutation (`deletePersonalAccessToken`) references a token `id` that doesn't belong to you.                        |
| `FORBIDDEN`                       | You attempt to create or delete a token while authenticating with token headers; these mutations require a logged-in session.        |

See [Error Codes](/api/start-guide/error-codes) for the full list.

## Authenticating subscriptions

GraphQL subscriptions connect over WebSocket at `wss://api.blue.app/graphql` and authenticate through `connectionParams` rather than HTTP headers. Pass the same `blue-*` headers in `connectionParams`, or pass a session JWT as `Authorization: Bearer <jwt>`. See [Connect and authenticate](/api/realtime/connect-and-authenticate) for the connection payload and a full example.

## Related

- [Making Requests](/api/start-guide/making-requests)
- [Error Codes](/api/start-guide/error-codes)
- [Connect and authenticate (subscriptions)](/api/realtime/connect-and-authenticate)
