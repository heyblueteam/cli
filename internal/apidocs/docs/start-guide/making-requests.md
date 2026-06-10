---
title: Making Requests
description: Send your first GraphQL query, mutation, and subscription to the Blue API over HTTP and WebSocket.
icon: Send
order: 2
---

The Blue API is a single GraphQL endpoint. You send every read and write as a POST to `https://api.blue.app/graphql`, and stream real-time updates over `wss://api.blue.app/graphql`. This page walks through a first read, a first write, and a first subscription, using the idiomatic GraphQL **variables** pattern so your queries stay clean and injection-safe.

Before you start, create a [Personal Access Token](/api/tokens) and have your Token ID, Token Secret, and Company ID ready. Every request authenticates with these three headers:

| Header                | Value                                            |
| --------------------- | ------------------------------------------------ |
| `X-Bloo-Token-ID`     | Your token's ID (an unprefixed cuid).            |
| `X-Bloo-Token-Secret` | Your token's secret (carries the `pat_` prefix). |
| `X-Bloo-Company-ID`   | The organization to act in — its ID or slug.     |

Header names are case-insensitive. Add `X-Bloo-Project-ID` only for the operations that require a workspace scope (custom fields, for example). See [Authentication](/api/start-guide/authentication) for the full model. Organizations are `Company` objects, workspaces are `Project` objects, and records are `Todo` objects in the API.

## Reading data

This query lists the workspaces (projects) you can access in an organization. Pass the company ID as a variable rather than inlining it into the query string.

```graphql
query ProjectList($companyId: String!) {
  projectList(filter: { companyIds: [$companyId] }) {
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

Send it with the `variables` object alongside the `query`:

```bash
curl -X POST https://api.blue.app/graphql \
  -H "Content-Type: application/json" \
  -H "X-Bloo-Token-ID: YOUR_TOKEN_ID" \
  -H "X-Bloo-Token-Secret: YOUR_TOKEN_SECRET" \
  -H "X-Bloo-Company-ID: YOUR_COMPANY_ID" \
  -d '{
    "query": "query ProjectList($companyId: String!) { projectList(filter: { companyIds: [$companyId] }) { items { id name updatedAt } pageInfo { totalItems hasNextPage } } }",
    "variables": { "companyId": "company_123" }
  }'
```

```python
import requests

url = "https://api.blue.app/graphql"

headers = {
    "Content-Type": "application/json",
    "X-Bloo-Token-ID": "YOUR_TOKEN_ID",
    "X-Bloo-Token-Secret": "YOUR_TOKEN_SECRET",
    "X-Bloo-Company-ID": "YOUR_COMPANY_ID",
}

query = """
query ProjectList($companyId: String!) {
  projectList(filter: { companyIds: [$companyId] }) {
    items { id name updatedAt }
    pageInfo { totalItems hasNextPage }
  }
}
"""

response = requests.post(
    url,
    json={"query": query, "variables": {"companyId": "company_123"}},
    headers=headers,
)

print(response.json())
```

```javascript
const url = 'https://api.blue.app/graphql'

const headers = {
  'Content-Type': 'application/json',
  'X-Bloo-Token-ID': 'YOUR_TOKEN_ID',
  'X-Bloo-Token-Secret': 'YOUR_TOKEN_SECRET',
  'X-Bloo-Company-ID': 'YOUR_COMPANY_ID',
}

const query = `
  query ProjectList($companyId: String!) {
    projectList(filter: { companyIds: [$companyId] }) {
      items { id name updatedAt }
      pageInfo { totalItems hasNextPage }
    }
  }
`

const response = await fetch(url, {
  method: 'POST',
  headers,
  body: JSON.stringify({ query, variables: { companyId: 'company_123' } }),
})

console.log(await response.json())
```

The API returns standard JSON. `projectList` is paginated, so results come back under `items` with a `pageInfo` block:

```json
{
  "data": {
    "projectList": {
      "items": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "name": "Website Redesign",
          "updatedAt": "2026-05-01T12:34:56.789Z"
        },
        {
          "id": "clm4n8r2k000108l0a1b2c3d4",
          "name": "Customer Onboarding",
          "updatedAt": "2026-04-28T09:21:34.567Z"
        }
      ],
      "pageInfo": { "totalItems": 24, "hasNextPage": true }
    }
  }
}
```

Select only the fields you need — GraphQL returns exactly what you ask for. `projectList` also accepts `skip` and `take` for paging through larger result sets; see [List workspaces](/api/workspaces/list-workspaces) for the full filter and pagination reference.

## Writing data

Use the `createTodo` mutation to create a record. `title` is the only required input; `todoListId` places the record in a specific list and is recommended.

```graphql
mutation CreateRecord($input: CreateTodoInput!) {
  createTodo(input: $input) {
    id
    title
    position
  }
}
```

```bash
curl -X POST https://api.blue.app/graphql \
  -H "Content-Type: application/json" \
  -H "X-Bloo-Token-ID: YOUR_TOKEN_ID" \
  -H "X-Bloo-Token-Secret: YOUR_TOKEN_SECRET" \
  -H "X-Bloo-Company-ID: YOUR_COMPANY_ID" \
  -d '{
    "query": "mutation CreateRecord($input: CreateTodoInput!) { createTodo(input: $input) { id title position } }",
    "variables": { "input": { "todoListId": "list_123", "title": "Draft launch plan" } }
  }'
```

```python
import requests

url = "https://api.blue.app/graphql"

headers = {
    "Content-Type": "application/json",
    "X-Bloo-Token-ID": "YOUR_TOKEN_ID",
    "X-Bloo-Token-Secret": "YOUR_TOKEN_SECRET",
    "X-Bloo-Company-ID": "YOUR_COMPANY_ID",
}

query = """
mutation CreateRecord($input: CreateTodoInput!) {
  createTodo(input: $input) {
    id
    title
    position
  }
}
"""

variables = {"input": {"todoListId": "list_123", "title": "Draft launch plan"}}

response = requests.post(url, json={"query": query, "variables": variables}, headers=headers)

print(response.json())
```

```javascript
const url = 'https://api.blue.app/graphql'

const headers = {
  'Content-Type': 'application/json',
  'X-Bloo-Token-ID': 'YOUR_TOKEN_ID',
  'X-Bloo-Token-Secret': 'YOUR_TOKEN_SECRET',
  'X-Bloo-Company-ID': 'YOUR_COMPANY_ID',
}

const query = `
  mutation CreateRecord($input: CreateTodoInput!) {
    createTodo(input: $input) { id title position }
  }
`

const variables = { input: { todoListId: 'list_123', title: 'Draft launch plan' } }

const response = await fetch(url, {
  method: 'POST',
  headers,
  body: JSON.stringify({ query, variables }),
})

console.log(await response.json())
```

`createTodo` returns the new `Todo` directly:

```json
{
  "data": {
    "createTodo": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Draft launch plan",
      "position": 65535
    }
  }
}
```

### Useful CreateTodoInput fields

`CreateTodoInput` accepts more than `title` and `todoListId`. The most useful optional fields:

| Field          | Type                                 | Description                                                      |
| -------------- | ------------------------------------ | ---------------------------------------------------------------- |
| `title`        | `String!`                            | The record title. The only required field.                       |
| `todoListId`   | `String`                             | List to create the record in. Omit to use the workspace default. |
| `assigneeIds`  | `[String!]`                          | User IDs to assign to the record.                                |
| `tags`         | `[CreateTodoTagInput!]`              | Tags to attach on creation.                                      |
| `customFields` | `[CreateTodoInputCustomField]`       | Custom field values to set on creation.                          |
| `checklists`   | `[CreateChecklistWithoutTodoInput!]` | Checklists to add to the record.                                 |
| `startedAt`    | `DateTime`                           | Start date.                                                      |
| `duedAt`       | `DateTime`                           | Due date.                                                        |
| `placement`    | `CreateTodoInputPlacement`           | `TOP` or `BOTTOM` of the list.                                   |

For the full surface — updating, assigning, tagging, and listing records — see the [Records](/api/records) section.

### Deleting a record

Use `deleteTodo` to delete a record. It takes a `DeleteTodoInput` with a single `todoId` and returns a `MutationResult`.

```graphql
mutation DeleteRecord($input: DeleteTodoInput!) {
  deleteTodo(input: $input) {
    success
    operationId
  }
}
```

```bash
curl -X POST https://api.blue.app/graphql \
  -H "Content-Type: application/json" \
  -H "X-Bloo-Token-ID: YOUR_TOKEN_ID" \
  -H "X-Bloo-Token-Secret: YOUR_TOKEN_SECRET" \
  -H "X-Bloo-Company-ID: YOUR_COMPANY_ID" \
  -d '{
    "query": "mutation DeleteRecord($input: DeleteTodoInput!) { deleteTodo(input: $input) { success operationId } }",
    "variables": { "input": { "todoId": "todo_123" } }
  }'
```

`deleteTodo` returns whether the delete succeeded. `operationId` is an optional handle for the background work the delete kicks off (cleanup of related rows), useful for correlating logs — most clients only read `success`:

```json
{
  "data": {
    "deleteTodo": {
      "success": true,
      "operationId": "clm4n8qwx000208l0e5f6g7h8"
    }
  }
}
```

## Subscriptions

Subscriptions stream real-time updates over a WebSocket — useful for live activity feeds, collaborative editing, or keeping local state in sync. Blue speaks the [`graphql-ws`](https://github.com/enisdenjo/graphql-ws) protocol on the same URL as the HTTP API, at `wss://api.blue.app/graphql`.

Authentication happens once, at connection time, through `connectionParams`. The server reads those params as request headers, so an API client authenticates over WebSocket with **the same `X-Bloo-*` Personal Access Token credentials** it uses over HTTP — not a `Bearer` token.

```javascript
import { createClient } from 'graphql-ws'

const client = createClient({
  url: 'wss://api.blue.app/graphql',
  connectionParams: {
    'x-bloo-token-id': 'YOUR_TOKEN_ID',
    'x-bloo-token-secret': 'YOUR_TOKEN_SECRET',
    'x-bloo-company-id': 'YOUR_COMPANY_ID',
  },
})

const unsubscribe = client.subscribe(
  {
    query: `
      subscription ActivityUpdates($companyId: String, $projectId: String) {
        subscribeToActivity(companyId: $companyId, projectId: $projectId) {
          mutation
          node {
            id
            html
            category
            createdAt
            createdBy {
              fullName
              email
            }
          }
        }
      }
    `,
    variables: {
      companyId: 'company_123',
      projectId: null, // optional: scope to a single workspace
    },
  },
  {
    next: (event) => console.log('activity:', event.data),
    error: (err) => console.error('subscription error:', err),
    complete: () => console.log('subscription complete'),
  },
)

// Later, stop receiving events:
// unsubscribe();
```

Each event is an `ActivitySubscriptionPayload`. The `mutation` field tells you the type of change — `CREATED`, `UPDATED`, or `DELETED` — and `node` is the affected `Activity`. Both `companyId` and `projectId` are nullable in the schema, but you must supply at least one so the server knows which feed to stream.

<Callout variant="info" title="Bearer JWT is the session path">

Setting `Authorization: Bearer <jwt>` in `connectionParams` authenticates a logged-in user **session**, validated as a Blue access-token JWT — not a Personal Access Token. The server resolves credentials in a fixed order (Firebase ID token, then Bearer JWT, then PAT), and the PAT path runs only when `Authorization` is absent. For API integrations, use the PAT params shown above and leave `Authorization` unset.

</Callout>

This is one subscription of many. For the handshake details, reconnection, and the full catalogue of channels, see [Connect & Authenticate](/api/realtime/connect-and-authenticate) and the [Real-time overview](/api/realtime).

## Error handling

Errors come back in the standard GraphQL shape: a top-level `errors` array, each entry with a `message` and an `extensions.code`. Always check for `errors` before reading `data` — when `errors` is present, the operation failed and `data` may be `null`. The full catalogue is on the [Error Codes](/api/start-guide/error-codes) page; the codes below are the ones you will hit first.

### Authentication error

Your credentials are missing or invalid:

```json
{
  "errors": [
    {
      "message": "You are not authenticated.",
      "extensions": { "code": "UNAUTHENTICATED" }
    }
  ]
}
```

### Permission error

You are authenticated but lack access to the resource:

```json
{
  "errors": [
    {
      "message": "You are not authorized.",
      "extensions": { "code": "FORBIDDEN" }
    }
  ]
}
```

### Not found error

The resource does not exist. Not-found errors are resource-specific — for example a missing workspace returns `PROJECT_NOT_FOUND`, a missing record returns `TODO_NOT_FOUND`. There is no generic `NOT_FOUND` code, so switch on the specific code:

```json
{
  "errors": [
    {
      "message": "Project was not found.",
      "extensions": { "code": "PROJECT_NOT_FOUND" }
    }
  ]
}
```

### Validation error

An input argument is invalid. Bad input is coded `BAD_USER_INPUT`, with per-field detail under `extensions.data`:

```json
{
  "errors": [
    {
      "message": "Title is required.",
      "extensions": {
        "code": "BAD_USER_INPUT",
        "data": { "field": "title" }
      }
    }
  ]
}
```

### Rate limit error

The tier rate limiter is enforced **before** GraphQL runs, so exceeding it returns an HTTP **429** with a plain JSON body — not a GraphQL `errors` array:

```json
{
  "error": "RATE_LIMITED",
  "message": "Too many requests. Limit: 600 per minute.",
  "retryAfter": 42
}
```

The response also carries a `Retry-After` header (seconds) and `X-RateLimit-*` headers. Separately, a few expensive operations enforce per-operation limits inside GraphQL and surface a `RATE_LIMITED` entry in the `errors` array instead. Limits are per **minute**. See [Rate Limits](/api/start-guide/rate-limits) for the tiers and headers.

<Callout variant="tip" title="Mind the query depth limit">

Queries are capped at a maximum nesting depth of 10 levels. Well-formed reads stay comfortably under it, but deeply nested selection sets are rejected — see [Capabilities](/api/start-guide/capabilities) for details.

</Callout>

## Next steps

- [Authentication](/api/start-guide/authentication) — the full credential model and how to create a token.
- [Records](/api/records) — create, update, assign, tag, comment on, and list records.
- [Real-time overview](/api/realtime) — every subscription channel and the payload shape.
- [Error Codes](/api/start-guide/error-codes) — the complete list of codes the API emits.
- [Rate Limits](/api/start-guide/rate-limits) — per-minute tiers and the headers to watch.
