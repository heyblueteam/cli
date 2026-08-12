---
title: Custom Domains
description: Serve Blue on your own hostname — create, verify via DNS, and list custom domains for your organization.
icon: Globe
order: 1
---

White-label customers can serve Blue's application, public forms, and file links from their own hostname instead of `blue.app`. A custom domain is a `CustomDomain` object: a hostname bound to one [application type](#applicationtype) for one organization, with a DNS [verification status](#the-customdomain-type) Blue refreshes on demand. Organizations are `Company` objects in the API.

This page covers the full lifecycle — create a domain, point it at Blue with a CNAME, verify it, update or delete it, and list the domains for an organization. For routing outbound mail through your own server, see [SMTP Credentials](/api/domains-email/smtp-credentials); to rebrand Blue's transactional emails, see [Email Templates](/api/domains-email/email-templates).

<Callout variant="info" title="White-label is required for writes, not reads">

Every **mutation** on this page requires the organization to have the `white_label` feature enabled (bundled with the Pro plan) — plain membership is not enough. The **`customDomains` query** requires only that you are a member of the organization, no feature gate. That asymmetry is intentional and applies across the whole [Custom Domains & Email](/api/domains-email) section.

</Callout>

## The CustomDomain type

| Field                | Type              | Description                                                                                             |
| -------------------- | ----------------- | ------------------------------------------------------------------------------------------------------- |
| `id`                 | `ID!`             | Unique identifier.                                                                                      |
| `uid`                | `String!`         | Short public identifier.                                                                                |
| `name`               | `String!`         | The hostname, e.g. `app.example.com`.                                                                   |
| `applicationType`    | `ApplicationType` | Which Blue surface this hostname serves. See [ApplicationType](#applicationtype).                       |
| `verificationStatus` | `String`          | `"verified"`, `"failed"`, or `null` if never verified. Set by [`verifyCustomDomain`](#verify-a-domain). |
| `company`            | `Company!`        | The organization that owns the domain.                                                                  |
| `user`               | `User!`           | The user who created the domain.                                                                        |
| `createdAt`          | `DateTime!`       | Creation timestamp.                                                                                     |
| `updatedAt`          | `DateTime!`       | Last-update timestamp.                                                                                  |

### ApplicationType

| Value           | Serves                    | Accepted by `createCustomDomain` |
| --------------- | ------------------------- | -------------------------------- |
| `APPLICATION`   | The main Blue web app.    | Yes                              |
| `FORMS`         | Public forms.             | Yes                              |
| `FILES`         | File links and downloads. | Yes                              |
| `DOCUMENTATION` | Documentation sites.      | **No** — rejected                |

An organization may register **one domain per application type** — e.g. one `APPLICATION` domain and one `FORMS` domain, but not two `APPLICATION` domains. `createCustomDomain` and `verifyCustomDomain` reject `DOCUMENTATION` with `APPLICATION_TYPE_NOT_FOUND`.

## Create a domain

Use the `createCustomDomain` mutation to register a hostname for an organization. Blue provisions the domain on its hosting layer and returns the new `CustomDomain` with `verificationStatus: null` — it is not live until DNS is in place and you [verify](#verify-a-domain) it.

### Request

```graphql
mutation CreateCustomDomain {
  createCustomDomain(
    input: { name: "app.example.com", companyId: "company_123", applicationType: APPLICATION }
  ) {
    id
    uid
    name
    applicationType
    verificationStatus
  }
}
```

### Parameters

#### CreateCustomDomainInput

| Parameter         | Type              | Required | Description                                                                                                      |
| ----------------- | ----------------- | -------- | ---------------------------------------------------------------------------------------------------------------- |
| `name`            | `String!`         | Yes      | The hostname to register, e.g. `app.example.com`.                                                                |
| `companyId`       | `String!`         | Yes      | The organization that will own the domain. Accepts an ID or slug. Must have the `white_label` feature.           |
| `applicationType` | `ApplicationType` | Yes¹     | One of `APPLICATION`, `FORMS`, or `FILES`. `DOCUMENTATION` is rejected. See [ApplicationType](#applicationtype). |

¹ The field is nullable in the schema, but the resolver rejects a missing or `DOCUMENTATION` value with `APPLICATION_TYPE_NOT_FOUND`. Always pass one of the three accepted values.

### Response

```json
{
  "data": {
    "createCustomDomain": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "uid": "cd_8f3a2c1e",
      "name": "app.example.com",
      "applicationType": "APPLICATION",
      "verificationStatus": null
    }
  }
}
```

### Errors

| Code                           | When                                                                                                               |
| ------------------------------ | ------------------------------------------------------------------------------------------------------------------ |
| `APPLICATION_TYPE_NOT_FOUND`   | `applicationType` is omitted or is `DOCUMENTATION`.                                                                |
| `CUSTOM_DOMAIN_IN_USE`         | The hostname is already registered (by any organization). Hostnames are globally unique.                           |
| `CUSTOM_DOMAIN_ALREADY_EXISTS` | This organization already has a domain for that application type.                                                  |
| `CUSTOM_DOMAIN_CREATE_ERROR`   | Blue's hosting layer rejected the domain — usually DNS isn't pointing at Blue yet. Add the CNAME, wait, and retry. |
| `COMPANY_NOT_FOUND`            | `companyId` does not resolve to an organization you belong to, or it lacks the `white_label` feature.              |

## Verify a domain

Use the `verifyCustomDomain` mutation to check that a hostname's DNS points at Blue. It runs a live CNAME lookup against the expected target for the domain's application type (with an IPv4 fallback to Blue's server address), writes the result to `verificationStatus`, and returns `true` when the domain is verified.

This mutation takes the hostname (`name`), not the domain ID.

### Request

```graphql
mutation VerifyCustomDomain {
  verifyCustomDomain(name: "app.example.com")
}
```

### Response

```json
{
  "data": {
    "verifyCustomDomain": true
  }
}
```

A `true` result sets `verificationStatus` to `"verified"`; `false` sets it to `"failed"`. Re-read the domain with [`customDomains`](#list-domains) to see the persisted status.

<Callout variant="info" title="Verification is throttled and live">

Each hostname can be verified at most **once per 10 seconds**. A call within that window returns the last known status without re-running the DNS lookup. The check is a real DNS resolution: if it succeeds and the CNAME (or fallback IP) matches, the domain is `verified`; if the domain resolves but doesn't match, it is `failed`. On a **transient network error** the mutation throws and leaves `verificationStatus` unchanged — retry rather than treating it as a failure.

</Callout>

### Errors

| Code                         | When                                                                                |
| ---------------------------- | ----------------------------------------------------------------------------------- |
| `CUSTOM_DOMAIN_NOT_FOUND`    | No domain is registered with that hostname.                                         |
| `APPLICATION_TYPE_NOT_FOUND` | The domain has no application type, or is a `DOCUMENTATION` domain.                 |
| `COMPANY_NOT_FOUND`          | You don't belong to the owning organization, or it lacks the `white_label` feature. |

## Update a domain

Use the `updateCustomDomain` mutation to change a domain's hostname. Blue removes the old hostname from its hosting layer, provisions the new one, migrates the white-label assets, and **resets `verificationStatus` to `null`** — re-run [`verifyCustomDomain`](#verify-a-domain) after the DNS for the new hostname is in place. The application type cannot be changed; delete and recreate to move to a different type.

### Request

```graphql
mutation UpdateCustomDomain {
  updateCustomDomain(input: { id: "domain_123", name: "app.newdomain.com" }) {
    id
    name
    applicationType
    verificationStatus
  }
}
```

### Parameters

#### UpdateCustomDomainInput

| Parameter | Type      | Required | Description                      |
| --------- | --------- | -------- | -------------------------------- |
| `id`      | `String!` | Yes      | The `CustomDomain` ID to update. |
| `name`    | `String!` | Yes      | The new hostname.                |

### Response

```json
{
  "data": {
    "updateCustomDomain": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "app.newdomain.com",
      "applicationType": "APPLICATION",
      "verificationStatus": null
    }
  }
}
```

### Errors

| Code                         | When                                                                                |
| ---------------------------- | ----------------------------------------------------------------------------------- |
| `CUSTOM_DOMAIN_NOT_FOUND`    | No domain matches `id`.                                                             |
| `CUSTOM_DOMAIN_IN_USE`       | The new hostname is already registered.                                             |
| `CUSTOM_DOMAIN_CREATE_ERROR` | Blue's hosting layer rejected the new hostname — usually a DNS issue. Retry.        |
| `COMPANY_NOT_FOUND`          | You don't belong to the owning organization, or it lacks the `white_label` feature. |

## Delete a domain

Use the `deleteCustomDomain` mutation to remove a domain. Blue tears it down on the hosting layer, clears the organization's white-label logo/favicon/manifest assets, and deletes the record. Returns `true` on success. This is a bare-scalar mutation — it takes no sub-selection.

### Request

```graphql
mutation DeleteCustomDomain {
  deleteCustomDomain(id: "domain_123")
}
```

### Response

```json
{
  "data": {
    "deleteCustomDomain": true
  }
}
```

### Errors

| Code                      | When                                                                                |
| ------------------------- | ----------------------------------------------------------------------------------- |
| `CUSTOM_DOMAIN_NOT_FOUND` | No domain matches `id`.                                                             |
| `COMPANY_NOT_FOUND`       | You don't belong to the owning organization, or it lacks the `white_label` feature. |

## List domains

Use the `customDomains` query to page through an organization's domains. It returns a `CustomDomainPagination` object — `items` plus a [`pageInfo`](#pageinfo). Unlike the mutations, this query requires only organization membership, not the `white_label` feature.

Pass `filter.companyId` to target a specific organization — useful when your token belongs to multiple organizations. Omit it and the query falls back to the organization in your `blue-org-id` header.

### Request

```graphql
query CustomDomains {
  customDomains(filter: { companyId: "company_123" }, skip: 0, take: 20) {
    items {
      id
      name
      applicationType
      verificationStatus
      createdAt
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

| Parameter | Type                       | Required | Description                                                                         |
| --------- | -------------------------- | -------- | ----------------------------------------------------------------------------------- |
| `filter`  | `CustomDomainFilterInput!` | Yes      | Filter object. See below. The argument is required, but its only field is optional. |
| `skip`    | `Int`                      | No       | Number of items to skip. Defaults to `0`.                                           |
| `take`    | `Int`                      | No       | Page size. Defaults to `20`.                                                        |

#### CustomDomainFilterInput

| Field       | Type     | Required | Description                                                                                                                 |
| ----------- | -------- | -------- | --------------------------------------------------------------------------------------------------------------------------- |
| `companyId` | `String` | No       | Organization to list domains for. Accepts an ID or slug. If omitted, falls back to the organization in `blue-org-id`. |

### Response

```json
{
  "data": {
    "customDomains": {
      "items": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "name": "app.example.com",
          "applicationType": "APPLICATION",
          "verificationStatus": "verified",
          "createdAt": "2026-05-20T14:32:09.000Z"
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

### CustomDomainPagination

| Field      | Type               | Description                                       |
| ---------- | ------------------ | ------------------------------------------------- |
| `items`    | `[CustomDomain!]!` | The domains on this page.                         |
| `pageInfo` | `PageInfo!`        | Pagination metadata. See [PageInfo](#pageinfo-1). |

#### PageInfo

| Field             | Type       | Description                        |
| ----------------- | ---------- | ---------------------------------- |
| `totalItems`      | `Int`      | Total domains matching the filter. |
| `totalPages`      | `Int`      | Total number of pages.             |
| `page`            | `Int`      | Current page (1-based).            |
| `perPage`         | `Int`      | Items per page (the `take` value). |
| `hasNextPage`     | `Boolean!` | Whether a next page exists.        |
| `hasPreviousPage` | `Boolean!` | Whether a previous page exists.    |

### Errors

| Code                | When                                                                     |
| ------------------- | ------------------------------------------------------------------------ |
| `COMPANY_NOT_FOUND` | No `filter.companyId` and no `blue-org-id` header to fall back to. |
| `FORBIDDEN`         | You are not a member of the target organization.                         |

## Setting up a custom domain end to end

1. **Create** the domain with [`createCustomDomain`](#create-a-domain), choosing the right `applicationType`. It comes back with `verificationStatus: null`.
2. **Point DNS** at Blue: add the CNAME record for your hostname as instructed in your Blue workspace settings. Newly added DNS records can take a few minutes to propagate.
3. **Verify** with [`verifyCustomDomain`](#verify-a-domain). When it returns `true`, the domain is live and `verificationStatus` is `"verified"`.

If [`createCustomDomain`](#create-a-domain) returns `CUSTOM_DOMAIN_CREATE_ERROR`, the DNS almost certainly isn't pointing at Blue yet — add the CNAME first, wait for propagation, then retry.

## Permissions

- **Mutations** (`createCustomDomain`, `updateCustomDomain`, `deleteCustomDomain`, `verifyCustomDomain`): you must be a member of the owning organization **and** that organization must have the `white_label` feature enabled.
- **`customDomains` query**: organization membership only — no feature gate. With `filter.companyId` you can list domains for any organization you belong to.

## Related

- [Custom Domains & Email](/api/domains-email)
- [SMTP Credentials](/api/domains-email/smtp-credentials)
- [Email Templates](/api/domains-email/email-templates)
- [Authentication](/api/start-guide/authentication)
- [Error codes](/api/start-guide/error-codes)
