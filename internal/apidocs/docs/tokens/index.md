---
title: Personal Access Tokens
description: Create, list, and revoke the personal access tokens that authenticate Blue API requests.
icon: KeyRound
order: 0
---

Personal access tokens (PATs) are the programmatic credentials behind Blue's API authentication. This section is the API-reference complement to the UI walkthrough in [Getting Started > Authentication](/api/start-guide/authentication): it documents how to create, list, and revoke tokens via GraphQL.

A token has two parts:

- **Token ID** — the unprefixed `uid`, sent as the `blue-token-id` header.
- **Secret** — the `pat_`-prefixed value, sent as the `blue-token-secret` header.

The Secret is shown **exactly once**, at creation. Blue stores only a bcrypt hash of it, so it can never be retrieved again — capture it when the create mutation returns it.

<Callout variant="warning" title="Token management needs a user session">

All three operations require an authenticated user session (the Firebase/JWT login used by the app). They **cannot** be performed while authenticating with a token itself — calling them with `blue-token-id` headers present returns `FORBIDDEN`. Generate, audit, and revoke tokens from a logged-in browser session, not from an API integration.

</Callout>

In the API, a token is a `PersonalAccessToken` object. The list query returns a `PersonalAccessTokenPagination` wrapper (`items` + `pageInfo`). Tokens are always scoped to the calling user — you only ever see and manage your own.

## Operations

| Operation                                  | Type     | Description                                                                                        |
| ------------------------------------------ | -------- | -------------------------------------------------------------------------------------------------- |
| [Create a Token](/api/tokens/create-token) | Mutation | `createPersonalAccessToken` — generate a token and read its Secret (the only time it's returned).  |
| [List Tokens](/api/tokens/query-tokens)    | Query    | `personalAccessTokens` — page through your own tokens to audit names, expiry, and last-used dates. |
| [Revoke a Token](/api/tokens/delete-token) | Mutation | `deletePersonalAccessToken` — delete a token by ID; it stops authenticating immediately.           |

## Related

- [Authentication](/api/start-guide/authentication) — the in-app flow for generating a token and the `blue-*` request headers.
- [Making Requests](/api/start-guide/making-requests) — how to send authenticated GraphQL requests with curl, Python, or Node.
