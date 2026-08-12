---
title: Documents
description: Two project-scoped document subsystems - collaboratively-edited rich-text Documents (and Wiki pages) and Portable Document PDF templates you fill from record data.
icon: FileText
order: 0
---

Blue has two distinct document subsystems, and they share little beyond living inside a workspace (a `Project` in the API). Don't conflate them:

- **Rich-text Documents** are collaboratively-edited pages with a `title`, HTML `content`, and an optional base64-encoded collab snapshot. The same `Document` type powers **Wiki pages** - a document with `wiki: true` is a Wiki page, everything else is identical.
- **Portable Documents** are PDF templates. You upload a PDF (max 10 pages), overlay fields that map a record's attributes - title, due date, assignees, tags, custom fields - onto its pages, then print a filled PDF for any specific record. This is a completely separate model (`PortableDocument`) created from an uploaded file.

This section covers creating, editing, deleting, querying, and subscribing to rich-text documents, plus building Portable Document templates and printing them.

```graphql
POST https://api.blue.app/graphql
blue-token-id: YOUR_TOKEN_ID
blue-token-secret: YOUR_TOKEN_SECRET
blue-org-id: YOUR_ORG_ID
blue-workspace-id: project_123
```

Both subsystems are scoped by the `blue-workspace-id` header (or a `projectId` in the input). Company and project headers accept either an ID or a slug. Header names are case-insensitive.

## Type mapping

The product nouns map to GraphQL types as follows. Use the product word in prose, the schema name in code.

| Product noun           | Schema type                                                   |
| ---------------------- | ------------------------------------------------------------- |
| Document / Wiki page   | `Document` (a Wiki page is a `Document` with `wiki: true`)    |
| Workspace / Project    | `Project`                                                     |
| Portable Document      | `PortableDocument`                                            |
| Portable Document page | `PortableDocumentPage`                                        |
| Overlay field          | `PortableDocumentField`                                       |
| Record                 | `Todo` (the row whose data fills a printed Portable Document) |

## Pages

| Page                                                                               | Covers                                                                                                                                    |
| ---------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| [Create and edit documents](/api/documents/create-document)                        | Create a rich-text document or Wiki page, update it, and delete it - `createDocument`, `updateDocument`, `deleteDocument`.                |
| [Query and subscribe to documents](/api/documents/query-documents)                 | Fetch one document, list a project's documents with filtering/sorting/pagination, and subscribe to real-time create/update/delete events. |
| [Build Portable Document templates](/api/documents/portable-documents)             | Create a Portable Document template from an uploaded PDF, rename it, delete it, fetch one, and list a project's templates.                |
| [Map fields and print Portable Documents](/api/documents/portable-document-fields) | Place, update, and delete overlay fields that bind record attributes to a page, then print a filled PDF for a given record.               |

## Related

- [Records](/api/records) - the `Todo` objects whose data fills a printed Portable Document.
- [Custom Fields](/api/custom-fields) - the `CustomField` values an overlay field can map onto a page.
- [Files](/api/files) - uploading the source PDF you turn into a Portable Document template.
