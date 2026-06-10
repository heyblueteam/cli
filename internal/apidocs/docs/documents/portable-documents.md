---
title: Build Portable Document templates
description: Turn an uploaded PDF into a Portable Document template, then rename, delete, fetch, or list a project's templates.
icon: FileStack
order: 3
---

A Portable Document is a PDF template scoped to a workspace. You upload a PDF, Blue rasterizes each page into a PNG, and the resulting template becomes a canvas you overlay record fields onto so you can later print a filled PDF for any specific record. Workspaces are `Project` objects in the API, and a template is a `PortableDocument`.

This is a separate subsystem from the collaboratively-edited rich-text [Documents](/api/documents/create-document) — they share the "Documents" section but no model, no fields, and no operations. Portable Documents are about laying record data onto a fixed PDF layout; rich-text Documents are about editing prose.

This page covers the template lifecycle: create one from an uploaded file, rename it, delete it, fetch one, and list a workspace's templates. Placing the overlay fields and printing a filled PDF are covered in [Map fields and print Portable Documents](/api/documents/portable-document-fields).

## Create a template

Use the `createPortableDocument` mutation to build a template from a PDF that has already been uploaded. You pass the workspace and the uploaded file; Blue fetches the file from storage, renders each page to a PNG, and stores the pages on the new template.

```graphql
mutation CreatePortableDocument {
  createPortableDocument(input: { projectId: "project_123", fileId: "file_123" }) {
    id
    uid
    name
    pages {
      id
      url
    }
    createdBy {
      id
      fullName
    }
  }
}
```

The `name` is taken from the source file's name — there is no `name` argument on create. Use `updatePortableDocument` afterwards to rename it.

### CreatePortableDocumentInput

| Parameter   | Type      | Required | Description                                                                                           |
| ----------- | --------- | -------- | ----------------------------------------------------------------------------------------------------- |
| `projectId` | `String!` | Yes      | Workspace the template belongs to. Accepts the project **ID or slug**.                                |
| `fileId`    | `String!` | Yes      | The uploaded PDF to convert. Resolved from the upload, so a file ID, storage key, or URL is accepted. |

<Callout variant="warning" title="Source constraints">

The source file's content type must be <code>application/pdf</code>, and the PDF must be **10 pages or fewer**. Conversion runs with a 15-second timeout — a PDF that is too complex to rasterize in that window fails with a timeout error. After a successful conversion the original uploaded PDF is **deleted from storage**; the template keeps only the rendered page images.

</Callout>

See [Upload files](/api/files) for how to obtain a `fileId`.

## Response

```json
{
  "data": {
    "createPortableDocument": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "uid": "pdf_clm4n8qwx000108l0h2abc456",
      "name": "Invoice template.pdf",
      "pages": [
        { "id": "clm4n8qwx000208l0p1ghi789", "url": "pdf_clm4n8qwx000108l0h2abc456/page_1" },
        { "id": "clm4n8qwx000308l0p2jkl012", "url": "pdf_clm4n8qwx000108l0h2abc456/page_2" }
      ],
      "createdBy": { "id": "clm4n8qwx000408l0u1mno345", "fullName": "Ada Lovelace" }
    }
  }
}
```

## The PortableDocument type

| Field       | Type                       | Description                                                                 |
| ----------- | -------------------------- | --------------------------------------------------------------------------- |
| `id`        | `ID!`                      | Unique template identifier.                                                 |
| `uid`       | `String!`                  | Prefixed unique key (`pdf_…`); also the storage prefix for the page images. |
| `name`      | `String!`                  | Display name, derived from the source file.                                 |
| `createdAt` | `DateTime!`                | When the template was created.                                              |
| `updatedAt` | `DateTime!`                | When the template was last modified.                                        |
| `createdBy` | `User!`                    | The user who created the template.                                          |
| `project`   | `Project!`                 | The workspace the template belongs to.                                      |
| `pages`     | `[PortableDocumentPage!]!` | The rendered pages, in order.                                               |
| `file`      | `File`                     | The source file, if still referenced.                                       |

### PortableDocumentPage

| Field       | Type                        | Description                                                          |
| ----------- | --------------------------- | -------------------------------------------------------------------- |
| `id`        | `ID!`                       | Unique page identifier.                                              |
| `uid`       | `String!`                   | Prefixed unique key for the page.                                    |
| `url`       | `String!`                   | Storage key of the rendered PNG (the `{document-uid}/page_{n}` key). |
| `createdAt` | `DateTime!`                 | When the page was created.                                           |
| `updatedAt` | `DateTime!`                 | When the page was last modified.                                     |
| `document`  | `PortableDocument!`         | The template this page belongs to.                                   |
| `fields`    | `[PortableDocumentField!]!` | Overlay fields placed on this page.                                  |

The `fields` list is populated by `createPortableDocumentField`; see [Map fields and print Portable Documents](/api/documents/portable-document-fields).

## Rename a template

Use the `updatePortableDocument` mutation to change a template's name. The only mutable field is `name`.

```graphql
mutation RenamePortableDocument {
  updatePortableDocument(input: { id: "pdf_123", name: "Q3 invoice" }) {
    id
    name
  }
}
```

### UpdatePortableDocumentInput

| Parameter | Type      | Required | Description                                                                            |
| --------- | --------- | -------- | -------------------------------------------------------------------------------------- |
| `id`      | `String!` | Yes      | The template to rename.                                                                |
| `name`    | `String`  | No       | New name. A `.pdf` suffix is appended automatically; omit it to keep the current name. |

```json
{
  "data": {
    "updatePortableDocument": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Q3 invoice.pdf"
    }
  }
}
```

## Delete a template

Use the `deletePortableDocument` mutation to permanently remove a template and its rendered page images from storage. It returns `Boolean` — `true` on success — so it takes no sub-selection.

```graphql
mutation DeletePortableDocument {
  deletePortableDocument(id: "pdf_123")
}
```

```json
{
  "data": {
    "deletePortableDocument": true
  }
}
```

| Argument | Type      | Required | Description             |
| -------- | --------- | -------- | ----------------------- |
| `id`     | `String!` | Yes      | The template to delete. |

## Fetch one template

Use the `portableDocument` query to read a single template by ID, including its pages and the overlay fields placed on each page.

```graphql
query GetPortableDocument {
  portableDocument(id: "pdf_123") {
    id
    name
    createdBy {
      fullName
    }
    pages {
      id
      url
      fields {
        id
        field
        positionX
        positionY
      }
    }
  }
}
```

| Argument | Type      | Required | Description            |
| -------- | --------- | -------- | ---------------------- |
| `id`     | `String!` | Yes      | The template to fetch. |

```json
{
  "data": {
    "portableDocument": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Q3 invoice.pdf",
      "createdBy": { "fullName": "Ada Lovelace" },
      "pages": [
        {
          "id": "clm4n8qwx000208l0p1ghi789",
          "url": "pdf_clm4n8qwx000108l0h2abc456/page_1",
          "fields": [
            {
              "id": "clm4n8qwx000508l0f1pqr678",
              "field": "title",
              "positionX": 72,
              "positionY": 540
            }
          ]
        }
      ]
    }
  }
}
```

## List a project's templates

Use the `portableDocuments` query to list every template in a workspace, newest-updated first, with offset pagination. It returns a `PortableDocumentPagination` of `{ items, pageInfo }`.

```graphql
query ListPortableDocuments {
  portableDocuments(filter: { projectId: "project_123" }, skip: 0, take: 20) {
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

### Arguments

| Argument | Type                      | Required | Description                                                              |
| -------- | ------------------------- | -------- | ------------------------------------------------------------------------ |
| `filter` | `PortableDocumentFilter!` | Yes      | Scopes the list to one workspace (below).                                |
| `skip`   | `Int`                     | No       | Items to skip for offset pagination. Defaults to `0`.                    |
| `take`   | `Int`                     | No       | Items per page. Schema default is `20`; omitting it falls back to `200`. |

### PortableDocumentFilter

| Field       | Type      | Required | Description                                                             |
| ----------- | --------- | -------- | ----------------------------------------------------------------------- |
| `projectId` | `String!` | Yes      | Workspace to list templates from. This filter is matched by project ID. |

```json
{
  "data": {
    "portableDocuments": {
      "items": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "name": "Q3 invoice.pdf",
          "updatedAt": "2026-05-29T14:02:11.000Z"
        },
        {
          "id": "clm4n8qwx000608l0g5stu901",
          "name": "NDA.pdf",
          "updatedAt": "2026-05-21T09:18:44.000Z"
        }
      ],
      "pageInfo": {
        "totalItems": 2,
        "hasNextPage": false
      }
    }
  }
}
```

### PageInfo

| Field             | Type       | Description                          |
| ----------------- | ---------- | ------------------------------------ |
| `totalItems`      | `Int`      | Total templates matching the filter. |
| `totalPages`      | `Int`      | Total pages at the current `take`.   |
| `page`            | `Int`      | Current page number.                 |
| `perPage`         | `Int`      | Items per page.                      |
| `hasNextPage`     | `Boolean!` | Whether a further page exists.       |
| `hasPreviousPage` | `Boolean!` | Whether a previous page exists.      |

## Errors

| Code                          | When                                                                                                              |
| ----------------------------- | ----------------------------------------------------------------------------------------------------------------- |
| `PROJECT_NOT_FOUND`           | `createPortableDocument` — `projectId` matches no workspace.                                                      |
| `FILE_NOT_FOUND`              | `createPortableDocument` — `fileId` can't be resolved to an uploaded file.                                        |
| `BAD_USER_INPUT`              | `createPortableDocument` — the source is not `application/pdf`, exceeds 10 pages, or conversion times out at 15s. |
| `PORTABLE_DOCUMENT_NOT_FOUND` | `updatePortableDocument`, `deletePortableDocument`, or `portableDocument` — no template with that `id`.           |
| `FORBIDDEN`                   | The caller is not a member of the template's workspace (on delete and single-fetch).                              |

## Permissions

All Portable Document operations require an authenticated caller. Beyond authentication:

- `portableDocument` and `deletePortableDocument` additionally require the caller to be a member of the workspace the template belongs to; otherwise they return `FORBIDDEN`.
- `createPortableDocument` resolves the workspace from `projectId` and attributes the new template to the caller via `createdBy`.

## Related

- [Map fields and print Portable Documents](/api/documents/portable-document-fields)
- [Create and edit documents](/api/documents/create-document)
- [Query and subscribe to documents](/api/documents/query-documents)
- [Upload files](/api/files)
- [Documents overview](/api/documents)
