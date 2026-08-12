---
title: Create an Organization
description: Create a new organization with the createOrganization mutation. The caller becomes the owner and starts a 14-day trial.
icon: Plus
order: 1
---

Use the `createOrganization` mutation to create a new organization. Organizations are `Organization` objects in the API and are the top-level tenant that contains workspaces, lists, and records.

Creating an organization has two immediate side effects you should plan for:

- **The authenticated user becomes the organization `OWNER`** — the new organization is owned by whoever makes the call.
- **A 14-day free trial starts automatically**, unless an unattached paid license already exists for the caller's email (for example, a pay-first checkout or an AppSumo lifetime deal). In that case the organization claims the existing license and no trial is started.

<Callout variant="info" title="Billing is per organization">

Each organization is billed separately. Creating organizations from the API can affect your account billing once trials expire.

</Callout>

## Request

This mutation takes a single required `input` argument of type `CreateOrganizationInput`. The smallest valid call supplies a `name` and a `slug`:

```graphql
mutation CreateOrganization {
  createOrganization(input: { name: "Acme Corporation", slug: "acme-corp" }) {
    id
    name
    slug
  }
}
```

Authenticate with your personal access token headers — no `blue-org-id` is required, since the organization does not exist yet:

```
blue-token-id: YOUR_TOKEN_ID
blue-token-secret: YOUR_TOKEN_SECRET
```

The endpoint is `https://api.blue.app/graphql`. Header names are case-insensitive.

## Parameters

### CreateOrganizationInput

| Field   | Type         | Required | Description                                                                                                                       |
| ------- | ------------ | -------- | --------------------------------------------------------------------------------------------------------------------------------- |
| `name`  | `String!`    | Yes      | Display name of the organization. Trimmed and capped at 50 characters. Any URLs and dots in the value are stripped before saving. |
| `slug`  | `String!`    | Yes      | Requested URL identifier. Trimmed and capped at 50 characters, then normalized server-side (see below). Must not contain a URL.   |
| `image` | `ImageInput` | No       | Optional organization logo. Supply the URLs for each pre-generated size.                                                          |

### ImageInput

All sizes are required when you provide an `image`. These are URLs to images you have already uploaded (see [Upload Files](/api/files)).

| Field       | Type      | Required | Description                           |
| ----------- | --------- | -------- | ------------------------------------- |
| `thumbnail` | `String!` | Yes      | URL of the thumbnail-size image.      |
| `small`     | `String!` | Yes      | URL of the small-size image.          |
| `medium`    | `String!` | Yes      | URL of the medium-size image.         |
| `large`     | `String!` | Yes      | URL of the large-size image.          |
| `original`  | `String!` | Yes      | URL of the original, full-size image. |

### Slug normalization

The `slug` you submit is **not stored verbatim**. The server lowercases it, strips symbols, replaces spaces with hyphens, and, if the result is already taken, appends a number (`acme-corp`, `acme-corp0`, `acme-corp1`, …). Two side effects matter for API consumers:

- The canonical, guaranteed-unique identifier for an organization is the returned `id` (a CUID), not the slug. Treat `id` as the primary key.
- Always read the `slug` field back from the response rather than assuming your input persisted unchanged.

To check whether a slug is free before you call `createOrganization`, use the [`isOrganizationSlugAvailable`](/api/organization-management/query-organization) query.

## Response

`createOrganization` returns the created `Organization`. The selection set above resolves to:

```json
{
  "data": {
    "createOrganization": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Acme Corporation",
      "slug": "acme-corp"
    }
  }
}
```

### Returns

`createOrganization` returns an `Organization!`. Commonly selected fields:

| Field                | Type               | Description                                                                                  |
| -------------------- | ------------------ | -------------------------------------------------------------------------------------------- |
| `id`                 | `ID!`              | Canonical unique identifier (CUID). Use this in subsequent calls and as `blue-org-id`. |
| `name`               | `String!`          | Stored display name (after trimming and URL/dot stripping).                                  |
| `slug`               | `String!`          | Normalized, de-duplicated URL identifier — may differ from your input.                       |
| `image`              | `Image`            | Logo, if an `image` was supplied.                                                            |
| `accessLevel`        | `UserAccessLevel!` | The caller's access level — `OWNER` for the user who created the organization.               |
| `freeTrialExpiredAt` | `DateTime`         | When the 14-day trial ends. `null` when the organization claimed an existing paid license.   |

## Full example

Create an organization with a logo and read back the trial and ownership fields:

```graphql
mutation CreateOrganizationWithLogo {
  createOrganization(
    input: {
      name: "Acme Corporation"
      slug: "acme-corp"
      image: {
        thumbnail: "https://blue.app/files/logo-thumb.png"
        small: "https://blue.app/files/logo-sm.png"
        medium: "https://blue.app/files/logo-md.png"
        large: "https://blue.app/files/logo-lg.png"
        original: "https://blue.app/files/logo.png"
      }
    }
  ) {
    id
    name
    slug
    accessLevel
    freeTrialStartedAt
    freeTrialExpiredAt
    image {
      medium
    }
  }
}
```

## Errors

| Code              | When                                                                                                                                    |
| ----------------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| `BAD_USER_INPUT`  | `name` or `slug` is missing, empty after trimming, or exceeds 50 characters.                                                            |
| `UNAUTHENTICATED` | The request has no valid authentication. Any authenticated user can create an organization; no existing company membership is required. |

A `slug` that contains a URL is rejected during normalization. Note that, unlike renaming an existing organization, `createOrganization` does **not** return `COMPANY_URL_ALREADY_EXISTS` for a slug collision — it silently appends a number to make the slug unique. Check availability up front with `isOrganizationSlugAvailable` if you need a specific slug.

## Related

- [Query an Organization](/api/organization-management/query-organization) — read back the organization you created, and check slug availability with `isOrganizationSlugAvailable`.
- [Organization Management overview](/api/organization-management)
- [Upload Files](/api/files) — generate the image URLs for the `image` input.
