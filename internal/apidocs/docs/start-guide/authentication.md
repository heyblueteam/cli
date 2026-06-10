---
title: Authentication
description: Authenticate to the Blue API with a personal access token and the X-Bloo-* request headers.
icon: Key
order: 1
---

The Blue API authenticates every request with a **personal access token (PAT)** — a Token ID and a Secret you generate in the app and send as headers on each request. There are no API keys or OAuth apps to register; a token belongs to a user and inherits that user's permissions.

The base endpoint is `https://api.blue.app/graphql` for HTTP requests and `wss://api.blue.app/graphql` for subscriptions.

## Authenticate a request

Send your credentials on every request using these headers:

```bash
curl -X POST https://api.blue.app/graphql \
  -H "Content-Type: application/json" \
  -H "X-Bloo-Token-ID: YOUR_TOKEN_ID" \
  -H "X-Bloo-Token-Secret: YOUR_TOKEN_SECRET" \
  -H "X-Bloo-Company-ID: YOUR_COMPANY_ID" \
  -d '{"query": "query Me { user { id email fullName } }"}'
```

| Header                | Required        | Description                                                                                                                                  |
| --------------------- | --------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `X-Bloo-Token-ID`     | Yes             | Your token's public identifier (the `uid`).                                                                                                  |
| `X-Bloo-Token-Secret` | Yes             | The token secret returned once at creation.                                                                                                  |
| `X-Bloo-Company-ID`   | Most operations | The organization to act in. Accepts the company ID or its slug.                                                                              |
| `X-Bloo-Project-ID`   | Some operations | The workspace to scope to. Accepts the project ID or its slug. Required for project-scoped reads and writes (records, lists, custom fields). |

Header names are **case-insensitive**, so `x-bloo-token-id` and `X-Bloo-Token-ID` are equivalent.

`X-Bloo-Company-ID` and `X-Bloo-Project-ID` both accept either the cuid or the human-readable slug — the slug is the segment you see in the app URL: `blue.app/company/{company-slug}/project/{project-slug}/`.

<Callout variant="info" title="Deprecated header aliases">

`x-company-id` and `x-project-id` are still accepted as aliases for `X-Bloo-Company-ID` and `X-Bloo-Project-ID`, but are deprecated. Use the `X-Bloo-*` names.

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

- **Token ID** — your token's public identifier (the `uid` field on `PersonalAccessToken`). It is an unprefixed identifier and is **not** secret. Send it as `X-Bloo-Token-ID`.
- **Secret** — the password half of the credential. It is **prefixed with `pat_`** for easy identification (for example, `pat_clm4n8qwx000008l0g4oxdqn7`). Send it as `X-Bloo-Token-Secret`.

<Callout variant="warning" title="The Secret is shown only once">

Blue stores the secret as a bcrypt hash, so it cannot be retrieved after creation — not even by Blue's team. Copy it somewhere safe the moment it's shown. If you lose it, delete the token and create a new one. Anyone with both the Token ID and Secret can act as you in Blue, so treat them like a password and never commit them to source control.

</Callout>

## Find your Company and Project IDs

Organizations (`Company` in the API) are the top-level entity; workspaces (`Project`) live inside them. Both the company and project values you pass in headers can be the ID **or** the slug, and the slug is visible in the app URL:

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

## Authentication errors

| Code                              | When                                                                                                                                 |
| --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `UNAUTHENTICATED`                 | The Token ID/Secret pair is missing, the secret does not match, or the token has expired. The request is treated as unauthenticated. |
| `PERSONAL_ACCESS_TOKEN_NOT_FOUND` | A token-management mutation (`deletePersonalAccessToken`) references a token `id` that doesn't belong to you.                        |
| `FORBIDDEN`                       | You attempt to create or delete a token while authenticating with token headers; these mutations require a logged-in session.        |

See [Error Codes](/api/start-guide/error-codes) for the full list.

## Authenticating subscriptions

GraphQL subscriptions connect over WebSocket at `wss://api.blue.app/graphql` and authenticate through `connectionParams` rather than HTTP headers. Pass the same `X-Bloo-*` headers in `connectionParams`, or pass a session JWT as `Authorization: Bearer <jwt>`. See [Connect and authenticate](/api/realtime/connect-and-authenticate) for the connection payload and a full example.

## Related

- [Making Requests](/api/start-guide/making-requests)
- [Error Codes](/api/start-guide/error-codes)
- [Connect and authenticate (subscriptions)](/api/realtime/connect-and-authenticate)
