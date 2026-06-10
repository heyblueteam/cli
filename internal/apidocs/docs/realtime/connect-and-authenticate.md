---
title: Connect & Authenticate
description: Open an authenticated graphql-ws WebSocket to stream real-time updates over wss://api.blue.app/graphql.
icon: PlugZap
order: 1
---

Blue's real-time API speaks the [`graphql-ws`](https://github.com/enisdenjo/graphql-ws) protocol over a single WebSocket. The endpoint is the **same URL as the HTTP GraphQL API** — `wss://api.blue.app/graphql` — so there is no separate subscriptions path to configure.

Authentication happens once, at connection time, through the `connectionParams` your client sends in the `connection_init` message. The server's `onConnect` handler reads those params, authenticates the caller, and either accepts the socket or rejects the handshake. Every subscription you open on the socket then runs as that user.

This page covers opening and authenticating the socket. For the subscriptions you can run once connected, see the [Real-time overview](/api/realtime).

## Request

Open the socket with the official `graphql-ws` client and pass your credentials in `connectionParams`. The recommended credential for API clients is a [Personal Access Token](/api/tokens): a Token ID and Secret pair.

```javascript
import { createClient } from 'graphql-ws'
import WebSocket from 'ws' // Node.js only; browsers have a global WebSocket

const client = createClient({
  url: 'wss://api.blue.app/graphql',
  webSocketImpl: WebSocket,
  connectionParams: {
    'x-bloo-token-id': 'YOUR_TOKEN_ID',
    'x-bloo-token-secret': 'YOUR_TOKEN_SECRET',
    'x-bloo-company-id': 'YOUR_COMPANY_ID',
  },
})

const unsubscribe = client.subscribe(
  {
    query: `
      subscription ActivityUpdates($companyId: String!) {
        subscribeToActivity(companyId: $companyId) {
          mutation
          node {
            id
            category
            html
            createdAt
            createdBy {
              fullName
              email
            }
          }
        }
      }
    `,
    variables: { companyId: 'company_123' },
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

`connectionParams` is a flat object of string values that the client serializes into the `connection_init` payload. The server reads it as the request "headers", so the keys are the same `X-Bloo-*` names you use over HTTP. Header names are case-insensitive.

## Parameters

### connectionParams

Pass exactly one of the three credential forms below, plus optional scoping params. The server resolves credentials in a fixed order — **Firebase ID token first, then Bearer JWT, then Personal Access Token** — so do not mix forms in one connection. In particular, if you set `Authorization`, the Token ID / Secret pair is ignored.

| Param                 | Type     | Required    | Description                                                                                                                                                                            |
| --------------------- | -------- | ----------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `x-bloo-token-id`     | `String` | Conditional | Personal Access Token ID — an unprefixed cuid. Pair with `x-bloo-token-secret`. The recommended credential for external/API clients.                                                   |
| `x-bloo-token-secret` | `String` | Conditional | The Personal Access Token Secret (prefixed `pat_`). Compared with bcrypt and checked for expiry. Shown only once at token creation.                                                    |
| `Authorization`       | `String` | Conditional | A session access token as `Bearer <JWT>`. The `Bearer ` prefix is stripped before verification. Used by first-party web sessions; prefer a Personal Access Token for API integrations. |
| `x-bloo-id-token`     | `String` | Conditional | A Firebase ID token. First-party web app only — external consumers should use a Personal Access Token instead.                                                                         |
| `x-bloo-company-id`   | `String` | No          | Scope the connection to an organization. Accepts the organization ID or slug. Required by company-scoped subscriptions such as `subscribeToActivity` and `subscribeToMyTodoCount`.     |
| `x-bloo-project-id`   | `String` | No          | Scope the connection to a workspace. Accepts the workspace ID or slug. Set alongside `x-bloo-company-id` when a subscription is workspace-scoped.                                      |

Organizations are `Company` objects and workspaces are `Project` objects in the API; both scoping params take an ID **or** a slug.

<Callout variant="tip" title="Use a Personal Access Token">

A Personal Access Token authenticates without a login session and never expires unless you set an expiry, which makes it the right credential for servers and integrations. Create one under your profile's API tab — see [Authentication](/api/start-guide/authentication). Send the Token ID as `x-bloo-token-id` and the Secret as `x-bloo-token-secret`, and **do not** also set `Authorization`, or the JWT path takes precedence and your token is skipped.

</Callout>

## Response

`graphql-ws` performs the handshake transparently. When `onConnect` accepts the credentials, the server replies with `connection_ack` and the socket is ready; you receive events through the `next` callback of each `client.subscribe(...)` call. Each delivered event is a normal GraphQL execution result:

```json
{
  "data": {
    "subscribeToActivity": {
      "mutation": "CREATED",
      "node": {
        "id": "clm4n8qwx000008l0g4oxdqn7",
        "category": "CREATE_TODO",
        "html": "<p>created a record</p>",
        "createdAt": "2026-05-29T14:21:08.512Z",
        "createdBy": {
          "fullName": "Ada Lovelace",
          "email": "ada@example.com"
        }
      }
    }
  }
}
```

Most entity subscriptions deliver this `{ mutation, node, previousValues, updatedFields }` change-feed shape, where `mutation` is one of `CREATED`, `UPDATED`, or `DELETED`. A few subscriptions return scalars or bare types instead. The shared payload and the exceptions are documented once on the [Real-time overview](/api/realtime).

## Errors

A failed handshake is rejected at the transport layer rather than returned as a GraphQL `errors` array.

| Condition                                              | Result                                                                                                                                                                 |
| ------------------------------------------------------ | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| No credentials, or credentials that resolve to no user | `onConnect` returns `false` — the server closes the socket and the handshake is **rejected**. `graphql-ws` surfaces this through the client's `error`/`closed` events. |
| Expired Personal Access Token                          | Treated as no user; the handshake is rejected.                                                                                                                         |
| Invalid or expired Bearer JWT / Firebase ID token      | Treated as no user; the handshake is rejected.                                                                                                                         |
| Banned organization, or an account being deleted       | Treated as no user; the handshake is rejected.                                                                                                                         |

Because authentication is resolved once at connect time, refresh a short-lived JWT before it expires and reconnect — `graphql-ws` re-sends `connectionParams` on every reconnect. A Personal Access Token avoids this entirely.

## Permissions

The handshake authenticates **who** you are; it does not pre-authorize **what** you can see. Permissions are enforced **per delivered event**: a subscription such as `subscribeToTodo` checks `permissions.todo(id).view()` before each record event reaches you. Subscribing always succeeds for an authenticated user, but you only receive events for the records, workspaces, and organizations you have access to. Scoping with `x-bloo-company-id` / `x-bloo-project-id` narrows which events are considered in the first place; it does not grant access you would not otherwise have.

## Related

- [Real-time overview](/api/realtime) — the shared payload shape and the full list of subscriptions
- [Record & Todo-List subscriptions](/api/realtime/record-subscriptions) — stream the core entity
- [Activity, mentions & notifications](/api/realtime/activity-notifications) — the `subscribeToActivity` channel used above
- [Authentication](/api/start-guide/authentication) — create a Personal Access Token
- [Making requests](/api/start-guide/making-requests) — the HTTP GraphQL endpoint
