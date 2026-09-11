---
title: Query and subscribe to documents
description: Fetch a single rich-text document, list a workspace's documents with filtering, sorting and pagination, and subscribe to real-time create/update/delete events.
icon: FileSearch
order: 2
---

Read rich-text documents from a workspace and watch them change in real time. Use the `document` query to fetch one document by id, the `documents` query to list a workspace's documents with filtering, sorting, and pagination, and the `subscribeToDocument` subscription to receive create/update/delete events as they happen.

Documents are `Document` objects in the API, and a workspace is a `Project`. Documents and legacy Wiki pages are one page tree — every page is an ordinary `Document`. This page covers the rich-text document subsystem only — Portable Document PDF templates are a separate model, covered in [Build Portable Document templates](/api/documents/portable-documents).

<Callout variant="info" title="Results are constrained to what your role can see">

Docs is a single role feature covering every page. If your role has it disabled, the result set is empty; otherwise you see the whole workspace tree.

</Callout>

## Request

Fetch a single document by id with the `document` query. The `id` is the document's `id` (a cuid).

```graphql
query GetDocument {
  document(id: "document_123") {
    id
    title
    content
    updatedAt
  }
}
```

List a workspace's documents with the `documents` query. Pass the workspace via `filter.projectId`, or omit it to default to the `blue-workspace-id` header.

```graphql
query ListDocuments {
  documents(filter: { projectId: "project_123" }) {
    items {
      id
      title
      createdBy {
        fullName
      }
      updatedAt
    }
    pageInfo {
      totalItems
      page
      perPage
      hasNextPage
    }
  }
}
```

## Parameters

### `document` arguments

| Parameter | Type      | Required | Description          |
| --------- | --------- | -------- | -------------------- |
| `id`      | `String!` | Yes      | The document's `id`. |

### `documents` arguments

| Parameter | Type                   | Required | Description                                                                                      |
| --------- | ---------------------- | -------- | ------------------------------------------------------------------------------------------------ |
| `filter`  | `DocumentFilterInput!` | Yes      | Which workspace's documents to list.                                                             |
| `sort`    | `[DocumentSort!]`      | No       | Sort order. Defaults to `[updatedAt_DESC]`.                                                      |
| `skip`    | `Int`                  | No       | Number of documents to skip (offset). Defaults to `0`.                                           |
| `take`    | `Int`                  | No       | Page size. Schema default is `20`; if you omit `take` entirely the resolver falls back to `200`. |

### DocumentFilterInput

| Field       | Type      | Required | Description                                                                                                                         |
| ----------- | --------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| `projectId` | `String`  | No       | Workspace to list documents from. Falls back to the `blue-workspace-id` header when omitted.                                        |
| `wiki`      | `Boolean` | No       | **Deprecated and ignored.** Documents and Wiki pages are one page tree, so this no longer narrows the result set.                    |

### DocumentSort

| Value            | Description                            |
| ---------------- | -------------------------------------- |
| `title_ASC`      | Title, A→Z.                            |
| `title_DESC`     | Title, Z→A.                            |
| `createdBy_ASC`  | Author first name, A→Z.                |
| `createdBy_DESC` | Author first name, Z→A.                |
| `updatedAt_ASC`  | Oldest update first.                   |
| `updatedAt_DESC` | Most recently updated first (default). |

## Response

`document` returns a single `Document`:

```json
{
  "data": {
    "document": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Onboarding runbook",
      "content": "<h1>Onboarding</h1><p>Welcome to the team.</p>",
      "updatedAt": "2026-05-29T10:14:22.000Z"
    }
  }
}
```

`documents` returns a `DocumentPagination` with `items` and `pageInfo`:

```json
{
  "data": {
    "documents": {
      "items": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "title": "Onboarding runbook",
          "createdBy": { "fullName": "Ada Lovelace" },
          "updatedAt": "2026-05-29T10:14:22.000Z"
        }
      ],
      "pageInfo": {
        "totalItems": 1,
        "page": 1,
        "perPage": 20,
        "hasNextPage": false
      }
    }
  }
}
```

### Document fields

| Field       | Type        | Description                                                           |
| ----------- | ----------- | --------------------------------------------------------------------- |
| `id`        | `ID!`       | Stable document identifier.                                           |
| `uid`       | `String!`   | Short unique id.                                                      |
| `title`     | `String!`   | Document title.                                                       |
| `content`   | `String`    | Rendered HTML body.                                                   |
| `wiki`      | `Boolean`   | **Deprecated.** Always `null` — the legacy storage field is retired. |
| `project`   | `Project!`  | The workspace the document belongs to.                                |
| `createdBy` | `User!`     | The user who created the document (select `fullName`, `email`, etc.). |
| `createdAt` | `DateTime!` | Creation timestamp.                                                   |
| `updatedAt` | `DateTime!` | Last-update timestamp.                                                |

### DocumentPagination fields

| Field      | Type          | Description                                                                                           |
| ---------- | ------------- | ----------------------------------------------------------------------------------------------------- |
| `items`    | `[Document!]` | The page of documents.                                                                                |
| `pageInfo` | `PageInfo`    | Pagination metadata: `totalItems`, `totalPages`, `page`, `perPage`, `hasNextPage`, `hasPreviousPage`. |

## Full example

List a workspace's pages, sorted by title, taking the second page of 10:

```graphql
query ListPages {
  documents(filter: { projectId: "project_123" }, sort: [title_ASC], skip: 10, take: 10) {
    items {
      id
      title
      createdBy {
        fullName
      }
    }
    pageInfo {
      totalItems
      totalPages
      page
      hasNextPage
      hasPreviousPage
    }
  }
}
```

## Subscribe to documents

Use the `subscribeToDocument` subscription over a WebSocket (`wss://api.blue.app/graphql`) to receive an event whenever a document in a workspace is created, updated, or deleted. Pass the workspace in `input.projectId`.

```graphql
subscription OnDocumentChange {
  subscribeToDocument(input: { projectId: "project_123" }) {
    mutation
    node {
      id
      title
      updatedAt
    }
    updatedFields
    previousValues {
      title
    }
  }
}
```

### SubscribeToDocumentInput

| Field       | Type      | Required | Description                                                                                   |
| ----------- | --------- | -------- | --------------------------------------------------------------------------------------------- |
| `projectId` | `String!` | Yes      | Workspace to watch.                                                                           |
| `wiki`      | `Boolean` | No       | **Deprecated and ignored.** The stream carries every page in the workspace.                   |

### DocumentSubscriptionPayload fields

| Field            | Type                     | Description                                                         |
| ---------------- | ------------------------ | ------------------------------------------------------------------- |
| `mutation`       | `MutationType!`          | The event type: `CREATED`, `UPDATED`, or `DELETED`.                 |
| `node`           | `Document`               | The current state of the document (null on `DELETED`).              |
| `updatedFields`  | `[String!]`              | Names of the fields that changed (on `UPDATED`).                    |
| `previousValues` | `DocumentPreviousValues` | The document's prior field values (`title`, timestamps, …).         |

Each payload only reaches you if you're a member of the target workspace. Events fire for every page in the workspace.

## Errors

| Code                 | When                                                                                                                   |
| -------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `DOCUMENT_NOT_FOUND` | `document` was called with an id that doesn't exist, or your role cannot read the workspace's pages.                   |
| `UNAUTHENTICATED`    | Missing or invalid credentials.                                                                                        |
| `FORBIDDEN`          | You aren't a member of the workspace you're querying or subscribing to.                                                |

## Permissions

- The `documents` query is gated on one role feature, **docs**, which covers every page in the workspace. If it is disabled the query returns an empty page.
- `document` applies the same gate and returns `DOCUMENT_NOT_FOUND` when it denies. Internal service callers (API-key auth, e.g. the collaboration server) skip this per-role gate.

## Related

- [Documents overview](/api/documents)
- [Create and edit documents](/api/documents/create-document)
- [Build Portable Document templates](/api/documents/portable-documents)
- [Map fields and print Portable Documents](/api/documents/portable-document-fields)
