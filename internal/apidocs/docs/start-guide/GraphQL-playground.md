---
title: GraphQL Playground
description: Explore the Blue API in your browser — run authenticated queries and browse the schema with the built-in GraphQL playground.
icon: Play
order: 6
---

Blue serves an interactive GraphQL playground (Apollo Studio Explorer) directly from the API endpoint at [https://api.blue.app/graphql](https://api.blue.app/graphql). Open that URL in a browser and you get an in-page editor for writing operations, running them against your live data, and browsing the schema. **Introspection is enabled**, so the Explorer's Documentation and Schema panels list every query, mutation, type, and field — the fastest way to discover what the API can do.

To run _authenticated_ operations (anything that reads or writes your data), you must add your API credentials to the playground's **HTTP headers** panel. Until you do, requests run unauthenticated and return an `UNAUTHENTICATED` error.

## Set your headers

Open the **Headers** tab at the bottom of the Explorer and paste your credentials as JSON. The header names are case-insensitive; the canonical form is the lowercase `blue-*` names:

```json
{
  "blue-token-id": "YOUR_TOKEN_ID",
  "blue-token-secret": "YOUR_TOKEN_SECRET",
  "blue-org-id": "YOUR_ORG_ID",
  "blue-workspace-id": "YOUR_PROJECT_ID"
}
```

- `X-Bloo-Token-ID` and `X-Bloo-Token-Secret` authenticate the request and are required on every authenticated operation. Generate them from your profile's **API** tab — see [Authentication](/api/start-guide/authentication).
- `X-Bloo-Company-ID` scopes the request to an organization. It accepts the company ID or its slug, and most operations need it.
- `X-Bloo-Project-ID` scopes the request to a single workspace (the schema type is `Workspace`). Include it only for operations that require a project context, such as custom-field operations.

<Callout variant="warning" title="Keep your secret safe">

The Token Secret is shown only once at creation and grants full API access to your data. Treat it like a password — don't commit it or paste it into a shared playground session.

</Callout>

## Run a sample query

With the headers set, run this query to list the workspaces (`Workspace` objects) you can access in the company. It needs only the Token ID, Token Secret, and Company ID headers:

```graphql
query ProjectListQuery {
  workspaceList(filter: { companyIds: ["company_123"] }) {
    items {
      id
      name
      updatedAt
    }
    pageInfo {
      totalItems
      hasNextPage
    }
  }
}
```

A successful run returns standard GraphQL JSON:

```json
{
  "data": {
    "workspaceList": {
      "items": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "name": "Website Redesign",
          "updatedAt": "2026-05-20T12:34:56.789Z"
        },
        {
          "id": "clm4n8qwx000108l0h2pyerm8",
          "name": "Sales Pipeline",
          "updatedAt": "2026-05-18T09:21:34.567Z"
        }
      ],
      "pageInfo": {
        "totalItems": 12,
        "hasNextPage": false
      }
    }
  }
}
```

If you see an `errors` array instead of `data`, check that all four headers are present and correctly spelled — a missing or mistyped credential is the most common cause of an `UNAUTHENTICATED` or `FORBIDDEN` response.

## Video overview

A short walkthrough of using the playground to test the API:

<youtube url="https://www.youtube.com/watch?v=GKSAHnoFn4s" />

## Related

- [Authentication](/api/start-guide/authentication) — generate a Token ID and Secret, find your company and project IDs.
- [Making Requests](/api/start-guide/making-requests) — the same operations via curl, Python, and Node.
- [Capabilities](/api/start-guide/capabilities) — what the API can do, including schema introspection.
- [Error Codes](/api/start-guide/error-codes) — reference for the codes returned in the `errors` array.
