---
title: Query an Organization
description: Retrieve an organization's profile, plan, and limits by ID or slug with the company query.
icon: Search
order: 2
---

Use the `company` query to fetch a single organization's profile — its name, owner, plan, access level, and usage limits. Organizations are `Company` objects in the API, and each one maps to a workspace group under a single billing account.

The query resolves an organization by **ID or slug**, but only one the authenticated user is a member of. It is the read counterpart to [Create an Organization](/api/organization-management/create-organization).

## Request

All requests go to `https://api.blue.app/graphql` with your personal access token headers. See [Authentication](/api/start-guide/authentication) for how to obtain a token.

```graphql
query GetOrganization {
  company(id: "company_123") {
    id
    name
    slug
  }
}
```

```bash
curl https://api.blue.app/graphql \
  -H "Content-Type: application/json" \
  -H "X-Bloo-Token-ID: YOUR_TOKEN_ID" \
  -H "X-Bloo-Token-Secret: YOUR_TOKEN_SECRET" \
  -H "X-Bloo-Company-ID: YOUR_COMPANY_ID" \
  -d '{"query": "query GetOrganization { company(id: \"company_123\") { id name slug } }"}'
```

The `id` argument accepts either the organization's ID or its slug, so `company(id: "acme-corp")` and `company(id: "clm4n8qwx000008l0g4oxdqn7")` resolve the same organization. Header names are case-insensitive.

## Parameters

| Argument | Type     | Required | Description                                                                                                                                                                                                                             |
| -------- | -------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `id`     | `String` | No       | The organization ID or slug to look up. Resolves only organizations the authenticated user belongs to. When omitted, the query returns an organization the user is a member of (typically the one in your `X-Bloo-Company-ID` context). |

<Callout variant="info" title="Lookups are scoped to your memberships">

The query never reveals organizations you don't belong to. An ID or slug for an org you aren't a member of resolves the same as a non-existent one — it raises `COMPANY_NOT_FOUND`.

</Callout>

## Response

```json
{
  "data": {
    "company": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Acme Corporation",
      "slug": "acme-corp",
      "description": "Operations and delivery teams at Acme.",
      "plan": "pro",
      "tier": "scale",
      "accessLevel": "ADMIN",
      "image": {
        "medium": "https://blue.app/files/companies/clm4n8qwx/medium.png"
      },
      "createdAt": "2024-01-15T10:30:00.000Z",
      "updatedAt": "2024-03-20T14:45:00.000Z"
    }
  }
}
```

### Returns

The query returns a single `Company` object. The most useful selectable fields:

| Field            | Type               | Description                                                                   |
| ---------------- | ------------------ | ----------------------------------------------------------------------------- |
| `id`             | `ID!`              | The organization's unique identifier.                                         |
| `uid`            | `String!`          | Short human-readable identifier.                                              |
| `name`           | `String!`          | Display name of the organization.                                             |
| `slug`           | `String!`          | URL-friendly identifier, unique per organization.                             |
| `description`    | `String`           | Optional organization description.                                            |
| `image`          | `Image`            | Organization logo, with `thumbnail`/`small`/`medium`/`large`/`original` URLs. |
| `owner`          | `User`             | The organization's owner. Select `fullName`, `email`, etc.                    |
| `accessLevel`    | `UserAccessLevel!` | Your access level within this organization (see below).                       |
| `plan`           | `String`           | Current plan, e.g. `"free"`, `"pro"`, `"enterprise"`.                         |
| `tier`           | `String`           | Current tier label, when set.                                                 |
| `workspaceAlias` | `String`           | The org's term for "workspace" when customized.                               |
| `limits`         | `CompanyLimits`    | Usage limits and consumption (records, users, storage, etc.).                 |
| `isBanned`       | `Boolean`          | Whether the organization is banned.                                           |
| `createdAt`      | `DateTime!`        | ISO 8601 timestamp of creation.                                               |
| `updatedAt`      | `DateTime!`        | ISO 8601 timestamp of last update.                                            |

`UserAccessLevel` is one of `OWNER`, `ADMIN`, `MEMBER`, `CLIENT`, `COMMENT_ONLY`, or `VIEW_ONLY`.

<Callout variant="warning" title="Don't use the deprecated relation fields">

`Company.customFields` and `Company.projects` are deprecated. Query [custom fields](/api/custom-fields) and [workspaces](/api/workspaces) through the top-level `customFields` and `projects` queries instead — they support filtering and pagination.

</Callout>

## Full example

Fetch the full profile including the owner and current usage against limits:

```graphql
query GetOrganizationDetails {
  company(id: "acme-corp") {
    id
    uid
    name
    slug
    description
    plan
    tier
    accessLevel
    workspaceAlias
    owner {
      id
      fullName
      email
    }
    image {
      medium
      large
    }
    limits {
      users {
        used
        limit
      }
      storage {
        used
        limit
      }
    }
    createdAt
    updatedAt
  }
}
```

## Errors

| Code                | When                                                                                                    |
| ------------------- | ------------------------------------------------------------------------------------------------------- |
| `COMPANY_NOT_FOUND` | No organization matches the given ID or slug among the organizations the authenticated user belongs to. |

A request made **as** a banned organization (a banned company in your `X-Bloo-Company-ID` context) is rejected at the auth layer with `COMPANY_BANNED` before the query runs — this is not raised by passing a banned org's ID to `company`. See [Error Codes](/api/start-guide/error-codes) for the full list.

## Related

- [Create an Organization](/api/organization-management/create-organization)
- [List Users](/api/user-management/list-users)
- [List Workspaces](/api/workspaces/list-workspaces)
- [Authentication](/api/start-guide/authentication)
