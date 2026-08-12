---
title: List Tokens
description: Page through your own personal access tokens to audit names, scopes, expiry, and last-used dates.
icon: List
order: 2
---

Use the `personalAccessTokens` query to list the personal access tokens you own. It returns a paginated set of `PersonalAccessToken` objects ordered newest-first, which is what you'd render in a token-management screen or use to audit which tokens are still active and when they expire.

The query is always scoped to the calling user — you only ever see your own tokens, never another user's, and there is no organization-wide listing. The `secret` field is always `null` here: a token's plaintext Secret is returned **only once**, by [`createPersonalAccessToken`](/api/tokens/create-token), and is never recoverable afterward.

## Request

The query takes two optional pagination arguments and returns a `PersonalAccessTokenPagination` wrapper (`items` + `pageInfo`).

```graphql
query ListTokens {
  personalAccessTokens(skip: 0, take: 20) {
    items {
      id
      uid
      name
      scopes
      expiredAt
      lastUsedAt
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

This is a user-session operation — send it with your login session (Firebase/JWT), not with `blue-token-*` headers. See [Token management needs a user session](#permissions) below.

## Parameters

| Argument | Type  | Required | Description                                                      |
| -------- | ----- | -------- | ---------------------------------------------------------------- |
| `skip`   | `Int` | No       | Number of tokens to skip for offset pagination. Defaults to `0`. |
| `take`   | `Int` | No       | Page size — how many tokens to return. Defaults to `20`.         |

## Response

```json
{
  "data": {
    "personalAccessTokens": {
      "items": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "uid": "clm4n8qwx000108l0a1b2c3d4",
          "name": "ci-deploy-bot",
          "scopes": null,
          "expiredAt": "2026-12-31T23:59:59.000Z",
          "lastUsedAt": "2026-05-28T09:14:02.000Z",
          "createdAt": "2026-05-01T12:00:00.000Z",
          "updatedAt": "2026-05-01T12:00:00.000Z"
        },
        {
          "id": "clm4n8qwx000208l0e5f6g7h8",
          "uid": "clm4n8qwx000308l0i9j0k1l2",
          "name": "zapier-integration",
          "scopes": null,
          "expiredAt": null,
          "lastUsedAt": null,
          "createdAt": "2026-04-15T08:30:00.000Z",
          "updatedAt": "2026-04-15T08:30:00.000Z"
        }
      ],
      "pageInfo": {
        "totalItems": 2,
        "totalPages": 1,
        "page": 1,
        "perPage": 20,
        "hasNextPage": false,
        "hasPreviousPage": false
      }
    }
  }
}
```

### PersonalAccessToken

| Field        | Type        | Description                                                                                                                        |
| ------------ | ----------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| `id`         | `ID!`       | Internal database ID of the token. Pass this to [`deletePersonalAccessToken`](/api/tokens/delete-token) to revoke the token.       |
| `uid`        | `String!`   | The unprefixed **Token ID**, sent as the `blue-token-id` header when authenticating with this token.                             |
| `name`       | `String!`   | The label you gave the token at creation (max 50 characters).                                                                      |
| `secret`     | `String`    | Always `null` from this query. The plaintext Secret is returned only on the create response; Blue stores only a bcrypt hash of it. |
| `scopes`     | `String`    | Reserved. Not currently persisted or enforced — see [Scopes](#scopes). Always `null` in practice.                                  |
| `expiredAt`  | `DateTime`  | When the token stops authenticating. `null` means the token never expires. A past value means the token is already rejected.       |
| `lastUsedAt` | `DateTime`  | When the token last authenticated a request, for auditing. `null` if it has never been used.                                       |
| `createdAt`  | `DateTime!` | When the token was created. Results are ordered by this field, descending (newest first).                                          |
| `updatedAt`  | `DateTime!` | When the token row was last modified.                                                                                              |
| `user`       | `User!`     | The token's owner — always the calling user. Select `id`, `email`, or `fullName`.                                                  |

### PageInfo

| Field             | Type       | Description                                                        |
| ----------------- | ---------- | ------------------------------------------------------------------ |
| `totalItems`      | `Int`      | Total number of tokens you own across all pages.                   |
| `totalPages`      | `Int`      | Total number of pages at the current `take`.                       |
| `page`            | `Int`      | The current page number (1-based), derived from `skip` and `take`. |
| `perPage`         | `Int`      | The page size in effect (equals `take`).                           |
| `hasNextPage`     | `Boolean!` | `true` if another page follows.                                    |
| `hasPreviousPage` | `Boolean!` | `true` if a previous page exists.                                  |

## Scopes

The `scopes` field exists on the type but is **reserved and not yet enforced**. A value passed to `createPersonalAccessToken` is validated and then discarded — it is never written to the database — and authentication never reads it. Every token therefore carries the full permissions of its owning user, and `scopes` reads back as `null`. Do not rely on it to scope-limit a token.

## Permissions

This query returns only the calling user's own tokens; there is no way to list another user's tokens or every token in an organization.

<Callout variant="warning" title="Token management needs a user session">

`personalAccessTokens`, like the create and revoke mutations, requires an authenticated user session (the Firebase/JWT login the app uses). It cannot be called while authenticating with a token itself — sending it with `blue-token-id` headers present returns `FORBIDDEN`. Audit and manage tokens from a logged-in session, not from an API integration.

</Callout>

## Related

- [Personal Access Tokens](/api/tokens) — section overview: the Token ID / Secret split and the request headers.
- [Create a Token](/api/tokens/create-token) — generate a token and capture its Secret (the only time it's returned).
- [Revoke a Token](/api/tokens/delete-token) — delete a token by `id`; it stops authenticating immediately.
- [Authentication](/api/start-guide/authentication) — the in-app flow for generating a token and the `blue-*` request headers.
