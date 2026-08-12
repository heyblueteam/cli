---
title: Capabilities
description: What you can do with the Blue API — read, write, paginate, subscribe, and introspect, each with a runnable example.
icon: Layers
order: 3
---

The Blue API is a single GraphQL endpoint that exposes the same data and actions you use in the app: workspaces, records, custom fields, comments, files, and more. Every capability below is reachable from one endpoint — `https://api.blue.app/graphql` for queries and mutations, and `wss://api.blue.app/graphql` for real-time subscriptions.

Records are `Record` objects in the API and workspaces are `Workspace` objects; the table maps each capability to a concrete operation and a page that covers it in depth.

| Capability | What it does                                        | Example operation                   | Deep dive                                          |
| ---------- | --------------------------------------------------- | ----------------------------------- | -------------------------------------------------- |
| Read       | Fetch exactly the fields you ask for in one request | `recordQueries { todos }`           | [List records](/api/records/list-records)          |
| Paginate   | Page through large result sets with `skip`/`take`   | `workspaceList` → `ProjectPagination` | [List workspaces](/api/workspaces/list-workspaces) |
| Write      | Create, update, and delete data with mutations      | `createRecord`                      | [Records](/api/records)                            |
| Subscribe  | Stream changes in real time over a WebSocket        | `onMoveTodo`                        | [Real-time overview](/api/realtime)                |
| Introspect | Discover the schema programmatically                | `__schema`                          | —                                                  |

## Read

Queries fetch only the fields you select, so a single request can return a precise slice of data instead of a fixed payload. Records are read through `recordQueries { todos(filter: TodosFilter!) }` — the `filter` requires `companyIds`, and you narrow further with optional fields like `projectIds` or `done`.

```graphql
query ListRecords {
  recordQueries {
    todos(filter: { companyIds: ["company_123"], projectIds: ["project_123"] }) {
      items {
        id
        title
        done
        duedAt
      }
    }
  }
}
```

```json
{
  "data": {
    "recordQueries": {
      "todos": {
        "items": [
          {
            "id": "clm4n8qwx000008l0g4oxdqn7",
            "title": "Draft launch plan",
            "done": false,
            "duedAt": "2026-06-15T17:00:00.000Z"
          }
        ]
      }
    }
  }
}
```

## Paginate

List queries return a paginated result, not a bare array. Each paginated type exposes `items`, a `pageInfo` object, and (where supported) a `totalCount`. Use the `skip` and `take` arguments to page through results, and read `pageInfo.totalItems` and `pageInfo.hasNextPage` to drive your loop.

```graphql
query ListWorkspaces {
  workspaceList(filter: { companyIds: ["company_123"] }, skip: 0, take: 20) {
    items {
      id
      name
    }
    pageInfo {
      totalItems
      page
      perPage
      hasNextPage
    }
    totalCount
  }
}
```

```json
{
  "data": {
    "workspaceList": {
      "items": [{ "id": "clm4n8qwx000008l0g4oxdqn7", "name": "Marketing" }],
      "pageInfo": {
        "totalItems": 42,
        "page": 1,
        "perPage": 20,
        "hasNextPage": true
      },
      "totalCount": 42
    }
  }
}
```

<Callout variant="info" title="PageInfo fields">

`PageInfo` exposes `totalItems`, `totalPages`, `page`, `perPage`, `hasNextPage`, and `hasPreviousPage`. (`startCursor`/`endCursor` are deprecated — page with `skip`/`take`.)

</Callout>

## Write

Mutations create, update, and delete data. The example below creates a record with `createRecord`; `title` is the only required input field, and the call returns the new `Record` so you can read fields back in the same request.

```graphql
mutation CreateRecord {
  createRecord(input: { todoListId: "list_123", title: "Draft launch plan" }) {
    id
    title
    done
  }
}
```

```json
{
  "data": {
    "createRecord": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Draft launch plan",
      "done": false
    }
  }
}
```

There is no generic "bulk write" endpoint, but several mutations operate on multiple items at once — for example `createCustomFieldOptions(input: CreateCustomFieldOptionsInput!)`, `uploadFiles(input: UploadFilesInput!)`, and `deleteFiles(ids: [String!]!)` (note `deleteFiles` takes a positional `ids` array, not an input object). For everything else, batch by sending multiple named mutations in a single GraphQL request.

## Subscribe

Subscriptions stream changes in real time so you never have to poll. They run over a WebSocket at `wss://api.blue.app/graphql` using the [`graphql-ws`](https://github.com/enisdenjo/graphql-ws) protocol — the same URL as the HTTP API, with credentials passed once in `connectionParams` at connection time. The example below pushes a `Record` every time a record moves within a workspace.

```graphql
subscription OnRecordMoved {
  onMoveTodo(projectId: "project_123") {
    id
    title
    position
  }
}
```

See [Connect & authenticate](/api/realtime/connect-and-authenticate) for opening the socket, and the [Real-time overview](/api/realtime) for the full list of subscriptions.

## Introspect

GraphQL is self-describing: you can query the schema itself to discover every type, query, mutation, and field at runtime. This powers tooling like the [GraphQL Playground](/api/start-guide/GraphQL-playground) and lets you generate typed clients without hand-maintaining a separate spec.

```graphql
query Introspect {
  __schema {
    queryType {
      name
    }
    types {
      name
      kind
    }
  }
}
```

<Callout variant="info" title="Query depth limit">

Queries are limited to a maximum depth of 10 nested levels. This prevents pathologically nested queries from degrading performance; well-formed reads stay comfortably under the limit.

</Callout>

## Related

- [Making requests](/api/start-guide/making-requests)
- [List records](/api/records/list-records)
- [List workspaces](/api/workspaces/list-workspaces)
- [Real-time overview](/api/realtime)
- [Upload files](/api/files)
