---
title: Query and subscribe to documents
description: Fetch a single rich-text document, list a workspace's documents with filtering, sorting and pagination, and subscribe to real-time create/update/delete events.
icon: FileSearch
order: 2
---

Read rich-text documents from a workspace and watch them change in real time. Use the `document` query to fetch one document by id, the `documents` query to list a workspace's documents with filtering, sorting, and pagination, and the `subscribeToDocument` subscription to receive create/update/delete events as they happen.

Documents are `Document` objects in the API, and a workspace is a `Project`. A document with `wiki: true` is a Wiki page; everything else about the type is identical. This page covers the rich-text document subsystem only — Portable Document PDF templates are a separate model, covered in [Build Portable Document templates](/api/documents/portable-documents).

<Callout variant="info" title="Results are constrained to what your role can see">

Docs and Wiki are independent role features. The `documents` query always limits results to the document types your role is permitted to read — a docs-only role never sees Wiki pages, and a Wiki-only role never sees regular documents — regardless of the `wiki` filter you pass. If neither feature is enabled for your role, the result set is empty.

</Callout>

## Request

Fetch a single document by id with the `document` query. The `id` is the document's `id` (a cuid).

```graphql
query GetDocument {
  document(id: "document_123") {
    id
    title
    content
    wiki
    updatedAt
  }
}
```

List a workspace's documents with the `documents` query. Pass the workspace via `filter.projectId`, or omit it to default to the `X-Bloo-Project-ID` header.

```graphql
query ListDocuments {
  documents(filter: { projectId: "project_123" }) {
    items {
      id
      title
      wiki
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
| `filter`  | `DocumentFilterInput!` | Yes      | Which workspace's documents to list, and an optional `wiki` flag.                                |
| `sort`    | `[DocumentSort!]`      | No       | Sort order. Defaults to `[updatedAt_DESC]`.                                                      |
| `skip`    | `Int`                  | No       | Number of documents to skip (offset). Defaults to `0`.                                           |
| `take`    | `Int`                  | No       | Page size. Schema default is `20`; if you omit `take` entirely the resolver falls back to `200`. |

### DocumentFilterInput

| Field       | Type      | Required | Description                                                                                                                         |
| ----------- | --------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| `projectId` | `String`  | No       | Workspace to list documents from. Falls back to the `X-Bloo-Project-ID` header when omitted.                                        |
| `wiki`      | `Boolean` | No       | `true` returns only Wiki pages; `false` returns only regular documents. Omit to return both (subject to the role constraint above). |

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
      "wiki": false,
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
          "wiki": false,
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

| Field           | Type        | Description                                                                                 |
| --------------- | ----------- | ------------------------------------------------------------------------------------------- |
| `id`            | `ID!`       | Stable document identifier.                                                                 |
| `uid`           | `String!`   | Short unique id.                                                                            |
| `title`         | `String!`   | Document title.                                                                             |
| `content`       | `String`    | Rendered HTML body.                                                                         |
| `contentBase64` | `String`    | Base64-encoded Yjs collaboration snapshot (the binary state shared with the collab server). |
| `wiki`          | `Boolean`   | `true` if this document is a Wiki page.                                                     |
| `project`       | `Project!`  | The workspace the document belongs to.                                                      |
| `createdBy`     | `User!`     | The user who created the document (select `fullName`, `email`, etc.).                       |
| `createdAt`     | `DateTime!` | Creation timestamp.                                                                         |
| `updatedAt`     | `DateTime!` | Last-update timestamp.                                                                      |

### DocumentPagination fields

| Field      | Type          | Description                                                                                           |
| ---------- | ------------- | ----------------------------------------------------------------------------------------------------- |
| `items`    | `[Document!]` | The page of documents.                                                                                |
| `pageInfo` | `PageInfo`    | Pagination metadata: `totalItems`, `totalPages`, `page`, `perPage`, `hasNextPage`, `hasPreviousPage`. |

## Full example

List only the Wiki pages in a workspace, sorted by title, taking the second page of 10:

```graphql
query ListWikiPages {
  documents(
    filter: { projectId: "project_123", wiki: true }
    sort: [title_ASC]
    skip: 10
    take: 10
  ) {
    items {
      id
      title
      wiki
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

Use the `subscribeToDocument` subscription over a WebSocket (`wss://api.blue.app/graphql`) to receive an event whenever a document in a workspace is created, updated, or deleted. Pass the workspace in `input.projectId`; optionally pass `input.wiki` to scope the stream to only Wiki pages (`true`) or only regular documents (`false`).

```graphql
subscription OnDocumentChange {
  subscribeToDocument(input: { projectId: "project_123" }) {
    mutation
    node {
      id
      title
      wiki
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
| `wiki`      | `Boolean` | No       | Scope the stream to Wiki pages (`true`) or regular documents (`false`). Omit to receive both. |

### DocumentSubscriptionPayload fields

| Field            | Type                     | Description                                                                    |
| ---------------- | ------------------------ | ------------------------------------------------------------------------------ |
| `mutation`       | `MutationType!`          | The event type: `CREATED`, `UPDATED`, or `DELETED`.                            |
| `node`           | `Document`               | The current state of the document (null on `DELETED`).                         |
| `updatedFields`  | `[String!]`              | Names of the fields that changed (on `UPDATED`).                               |
| `previousValues` | `DocumentPreviousValues` | The document's prior field values (`title`, `content`, `wiki`, timestamps, …). |

Each payload only reaches you if you're a member of the target workspace. Events fire for both document and Wiki changes within the workspace unless you narrow with `wiki`.

## Errors

| Code                 | When                                                                                                                   |
| -------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `DOCUMENT_NOT_FOUND` | `document` was called with an id that doesn't exist, or the document's type (doc vs Wiki) is one your role can't read. |
| `UNAUTHENTICATED`    | Missing or invalid credentials.                                                                                        |
| `FORBIDDEN`          | You aren't a member of the workspace you're querying or subscribing to.                                                |

## Permissions

- The `documents` query auto-constrains results to the document types your role can read: a docs-enabled, Wiki-disabled role never sees Wiki pages, and vice versa — even if you set the `wiki` filter to the type you can't access. A `wiki` filter can only narrow the result set further, never widen it. If neither feature is enabled, the query returns an empty page.
- `document` applies the same per-type gate: requesting a Wiki page with a docs-only role (or vice versa) returns `DOCUMENT_NOT_FOUND` rather than the document. Internal service callers (API-key auth, e.g. the collaboration server) skip this per-role gate.

## Related

- [Documents overview](/api/documents)
- [Create and edit documents](/api/documents/create-document)
- [Build Portable Document templates](/api/documents/portable-documents)
- [Map fields and print Portable Documents](/api/documents/portable-document-fields)
