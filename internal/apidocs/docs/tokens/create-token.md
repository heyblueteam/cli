---
title: Create a Token
description: Generate a personal access token with createPersonalAccessToken and capture its one-time Secret.
icon: KeyRound
order: 1
---

Use the `createPersonalAccessToken` mutation to generate a new personal access token (PAT) for the Blue API. The token is created against your own user account and is the credential you supply in the `X-Bloo-Token-ID` and `X-Bloo-Token-Secret` request headers. In the API a token is a `PersonalAccessToken` object.

The create response is the **only** place the plaintext Secret is ever returned — see [Capture the Secret now](#capture-the-secret-now) below before you run this.

## Request

The smallest call passes only a `name`. Select `uid` (the Token ID for your headers) and `secret` (the plaintext value, returned only here).

```graphql
mutation CreateToken {
  createPersonalAccessToken(input: { name: "CI deploy bot" }) {
    id
    uid
    name
    secret
    createdAt
  }
}
```

To make a token auto-expire, pass an `expiredAt` timestamp:

```graphql
mutation CreateExpiringToken {
  createPersonalAccessToken(
    input: { name: "Quarterly export job", expiredAt: "2026-12-31T23:59:59Z" }
  ) {
    uid
    name
    secret
    expiredAt
  }
}
```

This mutation requires an authenticated user session (Firebase/JWT login), so it needs only the `X-Bloo-Company-ID` header alongside your session token — not the `X-Bloo-Token-*` headers. See [Permissions](#permissions).

## Parameters

### CreatePersonalAccessTokenInput

| Parameter   | Type       | Required | Description                                                                                                                                                                                                                            |
| ----------- | ---------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `name`      | `String!`  | Yes      | A label for the token, shown in your token list. Max **50 characters**. Trimmed, HTML-escaped, and stripped of links and tags on save.                                                                                                 |
| `expiredAt` | `DateTime` | No       | Optional expiry timestamp (ISO 8601). After this moment the token stops authenticating — auth rejects it and the request is treated as unauthenticated. Omit for a token that never expires.                                           |
| `scopes`    | `String`   | No       | **Reserved — not yet enforced.** The value is validated but is not stored or checked anywhere; it grants no scope-limiting today. Every token currently has the full access of its owning user. Do not rely on it to restrict a token. |

## Response

The mutation returns the created `PersonalAccessToken`. The `secret` field is populated **only on this response**.

```json
{
  "data": {
    "createPersonalAccessToken": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "uid": "clm4n8qwx000108l0h2pzeabc",
      "name": "CI deploy bot",
      "secret": "pat_clm4n8qwx000208l0k7rytxyz",
      "createdAt": "2026-05-29T14:02:11.000Z"
    }
  }
}
```

### Returns

`PersonalAccessToken` fields relevant to this mutation:

| Field        | Type        | Description                                                                                                                                                           |
| ------------ | ----------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `id`         | `ID!`       | Internal record ID of the token. Use this with [`deletePersonalAccessToken`](/api/tokens/delete-token) to revoke it. **Not** a header value.                          |
| `uid`        | `String!`   | The unprefixed **Token ID**. Send this as the `X-Bloo-Token-ID` header.                                                                                               |
| `name`       | `String!`   | The sanitized label you supplied.                                                                                                                                     |
| `secret`     | `String`    | The `pat_`-prefixed plaintext **Secret**. Send this as the `X-Bloo-Token-Secret` header. **Non-null only on this create response** — every other read returns `null`. |
| `scopes`     | `String`    | Always `null` in practice (the input value is not persisted). Reserved.                                                                                               |
| `expiredAt`  | `DateTime`  | The expiry you set, or `null` if the token never expires.                                                                                                             |
| `lastUsedAt` | `DateTime`  | Last-used timestamp for auditing. `null` on a freshly created token.                                                                                                  |
| `createdAt`  | `DateTime!` | When the token was created.                                                                                                                                           |
| `updatedAt`  | `DateTime!` | When the token was last modified.                                                                                                                                     |
| `user`       | `User!`     | The owning user — always you.                                                                                                                                         |

<Callout variant="danger" title="Capture the Secret now">

The `secret` (`pat_...`) is shown **exactly once**, here. Blue stores only a bcrypt hash of it, so it is unrecoverable afterward — not even Blue can show it to you again. Copy it into your secret store immediately. If you lose it, [revoke the token](/api/tokens/delete-token) and create a new one. The [List Tokens](/api/tokens/query-tokens) query always returns `secret: null`.

</Callout>

### Mapping fields to headers

The two values you read here map directly to the request headers used by every authenticated API call:

| Create response field | Request header        | Example value                   |
| --------------------- | --------------------- | ------------------------------- |
| `uid`                 | `X-Bloo-Token-ID`     | `clm4n8qwx000108l0h2pzeabc`     |
| `secret`              | `X-Bloo-Token-Secret` | `pat_clm4n8qwx000208l0k7rytxyz` |

The `pat_` prefix belongs to the **Secret**, not the Token ID. See [Making Requests](/api/start-guide/making-requests) for a full authenticated request.

## Errors

| Code              | When                                                                                                                                                          |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `FORBIDDEN`       | The request is authenticating with a token (`X-Bloo-Token-ID` header present) instead of a user session. Token creation requires a session — see Permissions. |
| `BAD_USER_INPUT`  | `name` is missing, not a string, or exceeds 50 characters; or `expiredAt` is not a valid date.                                                                |
| `UNAUTHENTICATED` | No valid session credential was supplied at all.                                                                                                              |

## Permissions

Creating a token requires an **authenticated user session** — the Firebase/JWT login the app uses, not token auth. The resolver rejects the request with `FORBIDDEN` if it detects PAT headers (`X-Bloo-Token-ID`), so **you cannot mint a new token using an existing token**. Generate tokens from a logged-in session. The same guard applies to [revoking a token](/api/tokens/delete-token).

Tokens are always owned by the calling user. A token grants that user's full access; there is no scope narrowing today (see the `scopes` note above).

## Related

- [Personal Access Tokens](/api/tokens) — overview of the Token ID / Secret split and the `X-Bloo-*` headers.
- [List Tokens](/api/tokens/query-tokens) — page through and audit your tokens (`secret` is always `null` there).
- [Revoke a Token](/api/tokens/delete-token) — delete a token by `id`; it stops authenticating immediately.
- [Authentication](/api/start-guide/authentication) — the in-app flow for generating a token.
- [Making Requests](/api/start-guide/making-requests) — send an authenticated request with the Token ID and Secret.
