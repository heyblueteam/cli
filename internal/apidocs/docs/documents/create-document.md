---
title: Create and edit documents
description: Create, update, and delete rich-text Documents in a workspace using HTML.
icon: FileText
order: 1
---

Use the `createDocument`, `updateDocument`, and `deleteDocument` mutations to manage the lifecycle of a rich-text document — a collaboratively-edited page that lives inside a workspace. Workspaces are `Project` objects in the API, and each document is a `Document`. Documents and legacy Wiki pages are one page tree — every page is an ordinary `Document`.

Use `content` to read or write the document body as HTML. Blue converts it to the same collaborative state used by the editor. The raw collaboration state is internal and is not part of the public API.

This page covers rich-text documents only. The separate Portable Document subsystem (PDF templates for printing record data) is documented under [Build Portable Document templates](/api/documents/portable-documents).

## Request

Create a document in a workspace. `projectId` is the only required field; everything else is optional.

```graphql
mutation CreateDocument {
  createDocument(
    input: {
      projectId: "project_123"
      title: "Onboarding runbook"
      content: "<h1>Onboarding runbook</h1><p>Step 1…</p>"
    }
  ) {
    id
    title
    createdAt
  }
}
```

## Parameters

### CreateDocumentInput

| Parameter   | Type      | Required | Description                                                                                                                                                          |
| ----------- | --------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `projectId` | `String!` | Yes      | ID or slug of the workspace (`Project`) the document belongs to.                                                                                                     |
| `title`     | `String`  | No       | Document title. Defaults to an empty string if omitted.                                                                                                              |
| `content`   | `String`  | No       | Document body as HTML. Defaults to an empty document if omitted.                                                                                                     |
| `wiki`      | `Boolean` | No       | **Deprecated and ignored.** Documents and Wiki pages are one page tree; this field no longer has any effect. |

## Response

```json
{
  "data": {
    "createDocument": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Onboarding runbook",
      "createdAt": "2026-05-29T10:12:04.000Z"
    }
  }
}
```

### Document

The fields available on the returned `Document` type.

| Field       | Type        | Description                                              |
| ----------- | ----------- | -------------------------------------------------------- |
| `id`        | `ID!`       | Unique document identifier.                              |
| `uid`       | `String!`   | Short stable identifier used in URLs.                    |
| `title`     | `String!`   | Document title.                                          |
| `content`   | `String`    | Document body as HTML.                                   |
| `wiki`      | `Boolean`   | **Deprecated.** Always `null` — the legacy storage field is retired. |
| `project`   | `Project!`  | The workspace the document belongs to.                   |
| `createdBy` | `User!`     | The user who created the document.                       |
| `createdAt` | `DateTime!` | When the document was created.                           |
| `updatedAt` | `DateTime!` | When the document was last modified.                     |

## Update a document

Use `updateDocument` to change a document's fields. Only `id` is required — include just the fields you want to change. A title-only edit passes through without touching content.

```graphql
mutation UpdateDocument {
  updateDocument(
    input: {
      id: "document_123"
      title: "Onboarding runbook (v2)"
      content: "<h1>Onboarding runbook</h1><p>Updated.</p>"
    }
  ) {
    id
    title
    updatedAt
  }
}
```

### UpdateDocumentInput

| Parameter | Type      | Required | Description                                                                                                                    |
| --------- | --------- | -------- | ------------------------------------------------------------------------------------------------------------------------------ |
| `id`      | `ID!`     | Yes      | ID of the document to update.                                                                                                  |
| `title`   | `String`  | No       | New title.                                                                                                                     |
| `content` | `String`  | No       | New HTML body. An explicit empty string clears the body. If omitted, the existing body is unchanged.                           |
| `wiki`    | `Boolean` | No       | **Deprecated and ignored.** Documents and Wiki pages are one page tree; this field no longer has any effect.                  |

## Delete a document

Use `deleteDocument` to delete a document by ID. It returns `Boolean` — `true` on success — so it takes no sub-selection. The deleted document is moved to trash, not hard-deleted.

```graphql
mutation DeleteDocument {
  deleteDocument(id: "document_123")
}
```

```json
{ "data": { "deleteDocument": true } }
```

## Errors

| Code                 | Operation                          | When                                                                                                         |
| -------------------- | ---------------------------------- | ------------------------------------------------------------------------------------------------------------ |
| `PROJECT_NOT_FOUND`  | `createDocument`                   | No workspace matches `projectId`.                                                                            |
| `DOCUMENT_NOT_FOUND` | `updateDocument`, `deleteDocument` | No document matches `id`.                                                                                    |
| `FORBIDDEN`          | all                                | The caller is not an ADMIN/OWNER/MEMBER on the workspace, or their role cannot use the docs feature.         |
| `RATE_LIMITED`       | `createDocument`                   | More than 5 `createDocument` calls in the 60-second window for the same user (default limit).                |
| `UNAUTHENTICATED`    | all                                | The request carries no valid credentials.                                                                    |

## Permissions

- **`createDocument`** requires the caller to be an `ADMIN`, `OWNER`, or `MEMBER` on the workspace, and their role must have the **docs** feature enabled. It is rate-limited to a default of **5 requests per 60 seconds per user**.
- **`updateDocument`** and **`deleteDocument`** require the same **docs** permission — one gate covers every page in the tree. `deleteDocument` also succeeds for the document's original creator.

## Related

- [Query and subscribe to documents](/api/documents/query-documents) — fetch a document, list a workspace's documents, and stream real-time changes.
- [Build Portable Document templates](/api/documents/portable-documents) — the separate PDF-template subsystem.
- [Documents overview](/api/documents) — how rich-text Documents and Portable Documents differ.
