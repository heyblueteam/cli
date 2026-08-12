---
title: SMTP Credentials
description: Route Blue's outbound mail through your own SMTP server, with a live connection test and a standalone verify call.
icon: Server
order: 2
---

Route the transactional email Blue sends on your behalf through your own SMTP server instead of Blue's. Store your mail server's connection details once, and Blue uses them to deliver invitations, notifications, and other outbound mail with your own sending domain and reputation.

Mail-routing settings are `SmtpCredential` objects in the API, scoped to an organization (`Company`). Use [`createSmtpCredential`](#create-credentials) to store a server, [`smtpCredentials`](#list-credentials) to read them back, [`updateSmtpCredential`](#update-credentials) / [`deleteSmtpCredential`](#delete-credentials) to manage them, and [`verifySmtpCredential`](#verify-without-saving) to test a connection without saving anything.

<Callout variant="warning" title="The password is returned in plaintext on read">

The stored password is **encrypted at rest**, but the [`smtpCredentials`](#list-credentials) query **decrypts it and returns `SmtpCredential.password` in plaintext** in the response. This is only ever exposed to authorized members of the organization, but treat the query's output as a secret: do not log it, cache it, or surface it in a client you don't control.

</Callout>

All four mutations on this page require the organization to have the company-level **`white_label`** feature enabled — membership alone is not enough. The [`smtpCredentials`](#list-credentials) query requires only that you are a member of the organization. See [Permissions](#permissions).

## The SmtpCredential type

| Field         | Type       | Description                                                                                                                                  |
| ------------- | ---------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `id`          | `ID!`      | Unique identifier.                                                                                                                           |
| `uid`         | `String!`  | Short public identifier.                                                                                                                     |
| `host`        | `String`   | SMTP server hostname (e.g. `smtp.example.com`).                                                                                              |
| `port`        | `Int`      | SMTP port. Defaults to `587` when omitted on a connection test.                                                                              |
| `username`    | `String`   | SMTP auth username.                                                                                                                          |
| `password`    | `String`   | SMTP auth password. Encrypted at rest; **decrypted and returned in plaintext** by the [`smtpCredentials`](#list-credentials) query.          |
| `senderName`  | `String`   | Display name on outbound mail. The literal value `blue` is reserved.                                                                         |
| `senderEmail` | `String`   | From-address on outbound mail. Any address containing `@blue.cc` or `@blue.app` is reserved.                                                                |
| `verifiedAt`  | `DateTime` | When the connection was last verified successfully, or `null` if the most recent attempt failed. Stamped automatically on create and update. |
| `createdAt`   | `DateTime` | Creation timestamp.                                                                                                                          |
| `updatedAt`   | `DateTime` | Last-update timestamp.                                                                                                                       |

## Create credentials

Use the `createSmtpCredential` mutation to store a mail server for an organization. `companyId`, `host`, `port`, `username`, and `password` are required; `senderName` and `senderEmail` are optional.

On create, Blue encrypts the password, then immediately **attempts a live connection** to the server with the supplied credentials. If the connection succeeds, `verifiedAt` is set to the current time; if it fails, `verifiedAt` is left `null`. The mutation still returns the created `SmtpCredential` either way — inspect `verifiedAt` to know whether the server is actually usable.

<Callout variant="info" title="Reserved sender values">

`senderName` equal to `blue` (case-insensitive) is rejected with `SMTP_CREDENTIAL_RESERVED_SENDER_NAME`, and any `senderEmail` containing `@blue.cc` or `@blue.app` (case-insensitive) is rejected with `SMTP_CREDENTIAL_RESERVED_SENDER_EMAIL`. These keep white-label mail from impersonating Blue's own domain.

</Callout>

### Request

```graphql
mutation CreateSmtpCredential {
  createSmtpCredential(
    input: {
      companyId: "company_123"
      host: "smtp.example.com"
      port: 587
      username: "postmaster@acme.example.com"
      password: "YOUR_SMTP_PASSWORD"
      senderName: "Acme"
      senderEmail: "no-reply@acme.example.com"
    }
  ) {
    id
    uid
    host
    port
    senderName
    senderEmail
    verifiedAt
  }
}
```

`companyId` accepts an organization ID or its slug.

### Parameters

#### CreateSmtpCredentialInput

| Parameter     | Type      | Required | Description                                                                  |
| ------------- | --------- | -------- | ---------------------------------------------------------------------------- |
| `companyId`   | `String!` | Yes      | Organization that owns the credential. Must have the `white_label` feature.  |
| `host`        | `String!` | Yes      | SMTP server hostname.                                                        |
| `port`        | `Int!`    | Yes      | SMTP port (commonly `587` for STARTTLS or `465` for SMTPS).                  |
| `username`    | `String!` | Yes      | SMTP auth username.                                                          |
| `password`    | `String!` | Yes      | SMTP auth password. Encrypted before storage.                                |
| `senderName`  | `String`  | No       | Display name on outbound mail. The value `blue` is reserved.                 |
| `senderEmail` | `String`  | No       | From-address on outbound mail. Addresses containing `@blue.cc` or `@blue.app` are reserved. |

### Response

A successful connection stamps `verifiedAt`:

```json
{
  "data": {
    "createSmtpCredential": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "uid": "smtp_8f3a2c1e",
      "host": "smtp.example.com",
      "port": 587,
      "senderName": "Acme",
      "senderEmail": "no-reply@acme.example.com",
      "verifiedAt": "2026-05-29T14:22:05.000Z"
    }
  }
}
```

If the live connection fails, the credential is still created but `verifiedAt` is `null`:

```json
{
  "data": {
    "createSmtpCredential": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "uid": "smtp_8f3a2c1e",
      "host": "smtp.example.com",
      "port": 587,
      "senderName": "Acme",
      "senderEmail": "no-reply@acme.example.com",
      "verifiedAt": null
    }
  }
}
```

#### Returns

`SmtpCredential!` — see [The SmtpCredential type](#the-smtpcredential-type).

### Errors

| Code                                    | When                                                              |
| --------------------------------------- | ----------------------------------------------------------------- |
| `SMTP_CREDENTIAL_RESERVED_SENDER_NAME`  | `senderName` is `blue` (case-insensitive).                        |
| `SMTP_CREDENTIAL_RESERVED_SENDER_EMAIL` | `senderEmail` contains `@blue.cc` or `@blue.app` (case-insensitive).             |
| `FORBIDDEN`                             | You are not an `OWNER` or `ADMIN` of the organization.            |
| `PRO_REQUIRED`                          | The organization does not have the `white_label` feature enabled. |

## List credentials

Use the `smtpCredentials` query to page through the SMTP credentials stored for the organization in the `blue-org-id` header. Results are scoped to that organization; only its members can read them.

<Callout variant="warning" title="This query returns the password in plaintext">

`SmtpCredential.password` comes back **decrypted** from this query. Request it only when you need it, and never expose the response to an untrusted client.

</Callout>

### Request

```graphql
query SmtpCredentials {
  smtpCredentials(skip: 0, take: 20) {
    items {
      id
      host
      port
      username
      senderName
      senderEmail
      verifiedAt
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

### Parameters

| Parameter | Type  | Required | Description                                 |
| --------- | ----- | -------- | ------------------------------------------- |
| `skip`    | `Int` | No       | Number of records to skip. Defaults to `0`. |
| `take`    | `Int` | No       | Page size. Defaults to `20`.                |

### Response

```json
{
  "data": {
    "smtpCredentials": {
      "items": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "host": "smtp.example.com",
          "port": 587,
          "username": "postmaster@acme.example.com",
          "senderName": "Acme",
          "senderEmail": "no-reply@acme.example.com",
          "verifiedAt": "2026-05-29T14:22:05.000Z"
        }
      ],
      "pageInfo": {
        "totalItems": 1,
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

#### SmtpCredentialPagination

| Field      | Type                 | Description                                                                                            |
| ---------- | -------------------- | ------------------------------------------------------------------------------------------------------ |
| `items`    | `[SmtpCredential!]!` | The credentials on this page.                                                                          |
| `pageInfo` | `PageInfo!`          | Pagination metadata (`totalItems`, `totalPages`, `page`, `perPage`, `hasNextPage`, `hasPreviousPage`). |

### Errors

| Code        | When                                                             |
| ----------- | ---------------------------------------------------------------- |
| `FORBIDDEN` | You are not a member of the organization in `blue-org-id`. |

## Update credentials

Use the `updateSmtpCredential` mutation to change a stored server. Only `id` is required; every other field is optional and only the fields you pass are changed. As with create, Blue runs a **live connection test** after the update and re-stamps `verifiedAt` (success) or clears it to `null` (failure).

Password handling on update is specific:

- Pass a non-empty `password` to set a new one (it is re-encrypted).
- Pass `""` (empty string) or `null` to clear the stored password.
- Omit `password` entirely to leave the existing one unchanged.

The same reserved-value rules apply: `senderName` of `blue` and any `senderEmail` containing `@blue.cc` or `@blue.app` are rejected.

### Request

```graphql
mutation UpdateSmtpCredential {
  updateSmtpCredential(
    input: { id: "smtp_123", host: "smtp.new-provider.example.com", port: 465 }
  ) {
    id
    host
    port
    verifiedAt
  }
}
```

### Parameters

#### UpdateSmtpCredentialInput

| Parameter     | Type      | Required | Description                                                                   |
| ------------- | --------- | -------- | ----------------------------------------------------------------------------- |
| `id`          | `String!` | Yes      | The credential to update.                                                     |
| `host`        | `String`  | No       | New SMTP hostname.                                                            |
| `port`        | `Int`     | No       | New SMTP port.                                                                |
| `username`    | `String`  | No       | New SMTP auth username.                                                       |
| `password`    | `String`  | No       | New password. Empty string or `null` clears it; omitting leaves it unchanged. |
| `senderName`  | `String`  | No       | New display name. The value `blue` is reserved.                               |
| `senderEmail` | `String`  | No       | New from-address. Addresses containing `@blue.cc` or `@blue.app` are reserved.               |

### Response

```json
{
  "data": {
    "updateSmtpCredential": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "host": "smtp.new-provider.example.com",
      "port": 465,
      "verifiedAt": "2026-05-29T15:01:40.000Z"
    }
  }
}
```

#### Returns

`SmtpCredential!` — see [The SmtpCredential type](#the-smtpcredential-type).

### Errors

| Code                                    | When                                                                            |
| --------------------------------------- | ------------------------------------------------------------------------------- |
| `SMTP_CREDENTIAL_NOT_FOUND`             | No credential with the given `id` exists.                                       |
| `SMTP_CREDENTIAL_RESERVED_SENDER_NAME`  | `senderName` is `blue` (case-insensitive).                                      |
| `SMTP_CREDENTIAL_RESERVED_SENDER_EMAIL` | `senderEmail` contains `@blue.cc` or `@blue.app` (case-insensitive).                           |
| `FORBIDDEN`                             | You are not an `OWNER` or `ADMIN` of the organization that owns the credential. |
| `PRO_REQUIRED`                          | The organization does not have the `white_label` feature enabled.               |

## Delete credentials

Use the `deleteSmtpCredential` mutation to permanently remove a stored server. Once deleted, Blue falls back to sending mail from its own infrastructure for that organization. This mutation returns `Boolean` (`true` on success) and takes no sub-selection.

### Request

```graphql
mutation DeleteSmtpCredential {
  deleteSmtpCredential(id: "smtp_123")
}
```

### Parameters

| Parameter | Type      | Required | Description               |
| --------- | --------- | -------- | ------------------------- |
| `id`      | `String!` | Yes      | The credential to delete. |

### Response

```json
{ "data": { "deleteSmtpCredential": true } }
```

### Errors

| Code                        | When                                                                            |
| --------------------------- | ------------------------------------------------------------------------------- |
| `SMTP_CREDENTIAL_NOT_FOUND` | No credential with the given `id` exists.                                       |
| `FORBIDDEN`                 | You are not an `OWNER` or `ADMIN` of the organization that owns the credential. |
| `PRO_REQUIRED`              | The organization does not have the `white_label` feature enabled.               |

## Verify without saving

Use the `verifySmtpCredential` mutation to test a set of connection details **without storing anything**. It opens a connection to the server with the supplied values and returns `Boolean!` — `true` if the server accepts the connection and credentials, `false` otherwise. Nothing is persisted and no `SmtpCredential` is created, so this is the right call for validating a form before you save.

Every field on the input is optional. If you omit `port`, the test defaults to **`587`**. This mutation returns `Boolean!` and takes no sub-selection.

### Request

```graphql
mutation VerifySmtpCredential {
  verifySmtpCredential(
    input: {
      host: "smtp.example.com"
      port: 587
      username: "postmaster@acme.example.com"
      password: "YOUR_SMTP_PASSWORD"
    }
  )
}
```

### Parameters

#### VerifySmtpCredentialInput

| Parameter  | Type     | Required | Description                                |
| ---------- | -------- | -------- | ------------------------------------------ |
| `host`     | `String` | No       | SMTP server hostname.                      |
| `port`     | `Int`    | No       | SMTP port. Defaults to `587` when omitted. |
| `username` | `String` | No       | SMTP auth username.                        |
| `password` | `String` | No       | SMTP auth password.                        |

### Response

```json
{ "data": { "verifySmtpCredential": true } }
```

A failed connection returns `false` rather than erroring:

```json
{ "data": { "verifySmtpCredential": false } }
```

### Errors

| Code           | When                                                              |
| -------------- | ----------------------------------------------------------------- |
| `FORBIDDEN`    | You are not an `OWNER` or `ADMIN` of the organization.            |
| `PRO_REQUIRED` | The organization does not have the `white_label` feature enabled. |

## Permissions

Access to this resource is asymmetric:

- **Mutations** (`createSmtpCredential`, `updateSmtpCredential`, `deleteSmtpCredential`, `verifySmtpCredential`) require you to be an `OWNER` or `ADMIN` of the organization **and** require the organization to have the **`white_label`** feature enabled. Missing membership/role yields `FORBIDDEN`; a missing feature yields `PRO_REQUIRED`. White-label is bundled with the Pro plan — see the [section overview](/api/domains-email).
- **The `smtpCredentials` query** requires only that you are a member of the organization in `blue-org-id` (any `CompanyUser` row). It does **not** require the `white_label` feature. Because it returns the decrypted password, treat its output as sensitive.

## Related

- [Custom Domains](/api/domains-email/custom-domains) — serve Blue on your own hostname
- [Email Templates](/api/domains-email/email-templates) — override the copy of Blue's transactional emails
- [Custom Domains & Email overview](/api/domains-email)
- [Authentication](/api/start-guide/authentication)
- [Error codes](/api/start-guide/error-codes)
