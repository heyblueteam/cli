---
title: Create and edit documents
description: Create, update, and delete rich-text Documents and Wiki pages in a workspace, including the content/contentBase64 split and the anti-wipe guard.
icon: FileText
order: 1
---

Use the `createDocument`, `updateDocument`, and `deleteDocument` mutations to manage the lifecycle of a rich-text document — a collaboratively-edited page that lives inside a workspace. Workspaces are `Project` objects in the API, and each document is a `Document`. The same `Document` row powers both regular documents and Wiki pages: set `wiki: true` on create to make it a Wiki page.

A document carries two content representations. `content` is the rendered HTML. `contentBase64` is the base64-encoded collaborative-editing snapshot (the Yjs binary state the editor syncs over WebSocket). When you create a document from the API, sending `content` is usually enough; `contentBase64` is for clients that maintain their own collab state.

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
    wiki
    createdAt
  }
}
```

To create a Wiki page instead, set `wiki: true`:

```graphql
mutation CreateWikiPage {
  createDocument(input: { projectId: "project_123", title: "Team handbook", wiki: true }) {
    id
    title
    wiki
  }
}
```

## Parameters

### CreateDocumentInput

| Parameter       | Type      | Required | Description                                                                                                                                                          |
| --------------- | --------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `projectId`     | `String!` | Yes      | ID or slug of the workspace (`Project`) the document belongs to.                                                                                                     |
| `title`         | `String`  | No       | Document title. Defaults to an empty string if omitted.                                                                                                              |
| `content`       | `String`  | No       | Document body as HTML. Defaults to an empty string if omitted.                                                                                                       |
| `contentBase64` | `String`  | No       | Base64-encoded collaborative-editing snapshot. Leave unset unless your client manages its own collab state.                                                          |
| `wiki`          | `Boolean` | No       | When `true`, the document is a Wiki page instead of a regular document. The workspace must have the corresponding feature enabled (see [Permissions](#permissions)). |

## Response

```json
{
  "data": {
    "createDocument": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Onboarding runbook",
      "wiki": null,
      "createdAt": "2026-05-29T10:12:04.000Z"
    }
  }
}
```

### Document

The fields available on the returned `Document` type.

| Field           | Type        | Description                                              |
| --------------- | ----------- | -------------------------------------------------------- |
| `id`            | `ID!`       | Unique document identifier.                              |
| `uid`           | `String!`   | Short stable identifier used in URLs.                    |
| `title`         | `String!`   | Document title.                                          |
| `content`       | `String`    | Document body as HTML.                                   |
| `contentBase64` | `String`    | Base64-encoded collaborative-editing snapshot.           |
| `wiki`          | `Boolean`   | `true` if this is a Wiki page; otherwise `null`/`false`. |
| `project`       | `Project!`  | The workspace the document belongs to.                   |
| `createdBy`     | `User!`     | The user who created the document.                       |
| `createdAt`     | `DateTime!` | When the document was created.                           |
| `updatedAt`     | `DateTime!` | When the document was last modified.                     |

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

| Parameter       | Type      | Required | Description                                                                                                                                                                              |
| --------------- | --------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `id`            | `ID!`     | Yes      | ID of the document to update.                                                                                                                                                            |
| `title`         | `String`  | No       | New title.                                                                                                                                                                               |
| `content`       | `String`  | No       | New HTML body.                                                                                                                                                                           |
| `contentBase64` | `String`  | No       | New collaborative-editing snapshot. Subject to the anti-wipe guard below.                                                                                                                |
| `wiki`          | `Boolean` | No       | Flip the document between document and Wiki page. The caller must have permission for both the current and the resulting type.                                                           |
| `userId`        | `String`  | No       | Service-only attribution field used by Blue's collaboration server. **Not for client use** — it is validated server-side and rejected on regular calls. See [Permissions](#permissions). |

<Callout variant="warning" title="Anti-wipe guard">

If `contentBase64` is provided and decodes to empty (or undecodable) collab state while the document already has real content, `updateDocument` rejects the call with `BAD_USER_INPUT` rather than blanking the document. This protects against a stale or disconnected client overwriting good content. Title-only updates (no `contentBase64`) are never affected.

</Callout>

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

| Code                 | Operation                          | When                                                                                                                                                         |
| -------------------- | ---------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `PROJECT_NOT_FOUND`  | `createDocument`                   | No workspace matches `projectId`.                                                                                                                            |
| `DOCUMENT_NOT_FOUND` | `updateDocument`, `deleteDocument` | No document matches `id`.                                                                                                                                    |
| `FORBIDDEN`          | all                                | The caller isn't an ADMIN/OWNER/MEMBER on the workspace, or the workspace doesn't have the docs (or wiki) feature enabled for the caller's role.             |
| `BAD_USER_INPUT`     | `updateDocument`                   | The anti-wipe guard rejected a `contentBase64` that would blank existing content, or a service call supplied a `userId` that doesn't resolve to a real user. |
| `RATE_LIMITED`       | `createDocument`                   | More than 5 `createDocument` calls in the 60-second window for the same user (default limit).                                                                |
| `UNAUTHENTICATED`    | all                                | The request carries no valid credentials.                                                                                                                    |

## Permissions

- **`createDocument`** requires the caller to be an `ADMIN`, `OWNER`, or `MEMBER` on the workspace, and the workspace must have the relevant feature enabled for that role: the **docs** feature for a regular document, or the **wiki** feature when `wiki: true`. It is rate-limited to a default of **5 requests per 60 seconds per user**.
- **`updateDocument`** and **`deleteDocument`** require the caller to have permission for the document's type (docs for documents, wiki for Wiki pages). For an update that flips `wiki`, the caller needs permission for both the current and the resulting type. `deleteDocument` also succeeds for the document's original creator.
- **`updateDocument`** additionally accepts internal Blue service callers. The `userId` field exists only for that path — Blue's collaboration server passes it to attribute an edit to the right user. The server validates it against a real `User` and rejects it otherwise, so regular API clients should never set it; the editing user is taken from your credentials.

## Related

- [Query and subscribe to documents](/api/documents/query-documents) — fetch a document, list a workspace's documents, and stream real-time changes.
- [Build Portable Document templates](/api/documents/portable-documents) — the separate PDF-template subsystem.
- [Documents overview](/api/documents) — how rich-text Documents and Portable Documents differ.
