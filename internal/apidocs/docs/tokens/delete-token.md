---
title: Revoke a Token
description: Permanently revoke a personal access token with the deletePersonalAccessToken mutation. It stops authenticating immediately.
icon: Trash2
order: 3
---

Use the `deletePersonalAccessToken` mutation to permanently revoke a personal access token (a `PersonalAccessToken` object). Revocation is immediate and irreversible: the moment the mutation returns, the token stops authenticating and any request carrying its credentials begins failing.

You identify the token by its **Token ID** — the unprefixed `uid` you also send in the `X-Bloo-Token-ID` header, and the `id` returned by [List Tokens](/api/tokens/query-tokens). It is **not** the `pat_`-prefixed Secret; the Secret is never accepted as an argument and, after creation, is never recoverable.

You can only revoke your own tokens. The mutation is scoped to the calling user, so an ID belonging to another user is indistinguishable from one that does not exist — both raise `PERSONAL_ACCESS_TOKEN_NOT_FOUND`.

<Callout variant="warning" title="Token management needs a user session">

Like [creating a token](/api/tokens/create-token), revoking one requires an authenticated user session (the Firebase/JWT login the app uses). It **cannot** be performed while authenticating with a token: if the request carries `X-Bloo-Token-ID` headers, the mutation returns `FORBIDDEN`. Revoke tokens from a logged-in browser session, not from an API integration.

</Callout>

## Request

```graphql
mutation RevokeToken {
  deletePersonalAccessToken(input: { id: "pat_123" })
}
```

`deletePersonalAccessToken` returns the scalar `Boolean`, so the call takes no sub-selection. It resolves to `true` on success.

Authenticate with a user session, not token headers (see the callout above):

```
Authorization: Bearer YOUR_SESSION_JWT
X-Bloo-Company-ID: YOUR_COMPANY_ID
```

## Parameters

### DeletePersonalAccessTokenInput

| Parameter | Type  | Required | Description                                                                                                               |
| --------- | ----- | -------- | ------------------------------------------------------------------------------------------------------------------------- |
| `id`      | `ID!` | Yes      | The Token ID (the `PersonalAccessToken.id` from [List Tokens](/api/tokens/query-tokens)), not the `pat_`-prefixed Secret. |

## Response

```json
{
  "data": {
    "deletePersonalAccessToken": true
  }
}
```

### Returns

| Field                       | Type      | Description                                                       |
| --------------------------- | --------- | ----------------------------------------------------------------- |
| `deletePersonalAccessToken` | `Boolean` | `true` when the token was revoked. Errors are returned otherwise. |

## Errors

| Code                              | When                                                                                                                                                        |
| --------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `PERSONAL_ACCESS_TOKEN_NOT_FOUND` | No token owned by the caller matches `id` — the ID is unknown, already revoked, or belongs to another user. Message: `Personal access token was not found.` |
| `FORBIDDEN`                       | The request authenticated with a token instead of a user session (an `X-Bloo-Token-ID` header was present). Message: `You are not authorized.`              |

```json
{
  "errors": [
    {
      "message": "Personal access token was not found.",
      "extensions": { "code": "PERSONAL_ACCESS_TOKEN_NOT_FOUND" }
    }
  ]
}
```

## Permissions

You must be authenticated with a user session, and you can only revoke tokens you own. Both the foreign-ID and unknown-ID cases return `PERSONAL_ACCESS_TOKEN_NOT_FOUND` — the mutation never reveals whether a token exists under another user.

Revocation is permanent: Blue stores only a bcrypt hash of the Secret, so a revoked token cannot be restored or re-enabled. To replace it, [create a new token](/api/tokens/create-token) and capture its Secret. If you only want a token to stop working at a future date rather than now, set `expiredAt` when you create it instead of revoking it manually.

## Related

- [List Tokens](/api/tokens/query-tokens) — find the `id` of the token to revoke and audit which tokens are active.
- [Create a Token](/api/tokens/create-token) — generate a replacement and read its Secret.
- [Personal Access Tokens overview](/api/tokens)
