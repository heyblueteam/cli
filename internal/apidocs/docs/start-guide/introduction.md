---
title: Introduction
description: Get started with the Blue API — a single GraphQL endpoint for everything in your Blue organization, authenticated with personal access tokens.
icon: BookOpen
order: 0
---

The Blue API is a single [GraphQL](https://graphql.org/) endpoint that exposes the same data and operations you use in the Blue app: organizations, workspaces, records, lists, custom fields, comments, automations, and more. Blue is a process-management platform used by 19,000+ organizations, and anything you can do in the UI you can drive programmatically.

GraphQL means one endpoint, a typed schema you can introspect, and responses that return exactly the fields you ask for. If you're new to GraphQL, the [graphql.org learning guide](https://graphql.org/learn/) is a good primer.

## Base endpoint

All requests are `POST` to a single URL:

```
https://api.blue.app/graphql
```

Real-time subscriptions use the WebSocket endpoint `wss://api.blue.app/graphql`. See [Realtime](/api/realtime/connect-and-authenticate) for the subscription protocol.

## Authentication in one minute

The API authenticates with a **personal access token (PAT)** — a Token ID and a Secret you generate from your profile's **API** tab. You send them as request headers, along with the organization (and, for some operations, the workspace) you want to act in:

| Header                | Required        | Description                                                    |
| --------------------- | --------------- | -------------------------------------------------------------- |
| `X-Bloo-Token-ID`     | Yes             | Your token ID (prefixed `pat_`).                               |
| `X-Bloo-Token-Secret` | Yes             | Your token secret. Shown once at creation — store it securely. |
| `X-Bloo-Company-ID`   | Most operations | The organization ID or slug to act in.                         |
| `X-Bloo-Project-ID`   | Some operations | The workspace ID or slug, when an operation is scoped to one.  |

Header names are case-insensitive. The secret is hashed with bcrypt on our side, so it can never be recovered after creation — treat it like a password.

<Callout variant="warning" title="Keep your secret safe">

Anyone with your Token ID and Secret can read and modify your organization's data. Store the secret in a secrets manager, never in client-side code or version control.

</Callout>

See [Authentication](/api/start-guide/authentication) for how to create a token and find your organization and workspace IDs, and [Making Requests](/api/start-guide/making-requests) for full curl, Python, and Node examples.

## Quick start

Once you have a token, this query lists the workspaces in an organization. Workspaces are `Project` objects in the API, queried with the [`projectList`](/api/workspaces/list-workspaces) query.

```bash
curl -X POST https://api.blue.app/graphql \
  -H "Content-Type: application/json" \
  -H "X-Bloo-Token-ID: YOUR_TOKEN_ID" \
  -H "X-Bloo-Token-Secret: YOUR_TOKEN_SECRET" \
  -H "X-Bloo-Company-ID: YOUR_COMPANY_ID" \
  -d '{
    "query": "query MyWorkspaces($filter: ProjectListFilter!) { projectList(filter: $filter) { items { id name updatedAt } } }",
    "variables": { "filter": { "companyIds": ["company_123"] } }
  }'
```

The same operation as a named GraphQL document:

```graphql
query MyWorkspaces {
  projectList(filter: { companyIds: ["company_123"] }) {
    items {
      id
      name
      updatedAt
    }
  }
}
```

`companyIds` accepts organization IDs or slugs. The response returns the requested fields and nothing else:

```json
{
  "data": {
    "projectList": {
      "items": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "name": "Product Roadmap",
          "updatedAt": "2026-05-20T14:32:11.000Z"
        }
      ]
    }
  }
}
```

Want to experiment without writing code first? Open the [GraphQL Playground](/api/start-guide/GraphQL-playground) in your browser to run authenticated queries and browse the full schema.

## What you'll learn

The rest of this section covers everything you need before working with specific resources:

1. **[Authentication](/api/start-guide/authentication)** — create a personal access token and find your organization and workspace IDs.
2. **[Making Requests](/api/start-guide/making-requests)** — send queries, mutations, and variables over curl, Python, and Node.
3. **[Capabilities](/api/start-guide/capabilities)** — what the API can do, including schema introspection.
4. **[Rate Limits](/api/start-guide/rate-limits)** — request limits by plan and the headers that report your remaining quota.
5. **[Upload Files](/api/files)** — attach documents and images to records and comments.
6. **[GraphQL Playground](/api/start-guide/GraphQL-playground)** — explore and run operations interactively in the browser.
7. **[Error Codes](/api/start-guide/error-codes)** — the codes returned in the `errors` array and what each one means.

## What you can build

- **Sync records both ways.** Create and update records (`Todo` objects) from external systems, and read them back to keep another database, spreadsheet, or warehouse in step.
- **Drive workflows from external events.** Turn inbound forms, webhooks, or emails into records, set custom field values, and move records between lists as work progresses.
- **Connect Blue to your stack.** Push data from a CRM, ERP, or support tool into the right workspace, and surface Blue data in your own dashboards.
- **React in real time.** Subscribe over WebSocket to record, comment, and file events to update integrations the moment something changes — see [Realtime](/api/realtime/connect-and-authenticate).

## Related

- [Authentication](/api/start-guide/authentication) — generate a token and find your IDs.
- [Making Requests](/api/start-guide/making-requests) — the endpoint, headers, and request shape.
- [GraphQL Playground](/api/start-guide/GraphQL-playground) — run operations in the browser.
- [Error Codes](/api/start-guide/error-codes) — the full list of error codes.
