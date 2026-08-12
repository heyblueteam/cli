---
title: Map fields and print Portable Documents
description: Place overlay fields that bind a record's attributes onto a Portable Document page, then print a filled PDF for any record.
icon: Printer
order: 4
---

A Portable Document template is only useful once you tell it _where_ each piece of record data should appear. Use the `createPortableDocumentField` mutation to place an overlay field on a [template](/api/documents/portable-documents) page: each field binds one of a record's attributes — its title, description, due date, assignees, tags, or a specific custom field — to an x/y position with an optional width and height. Then call `printPortableDocument` to render a filled PDF for any specific record and get back its storage key.

Records are `Todo` objects in the API, and workspaces are `Project` objects. Overlay fields are `PortableDocumentField` objects, each attached to a `PortableDocumentPage` of a `PortableDocument` template. Placing, moving, and deleting fields requires **contribute** permission on the template's workspace.

## Create a field

Use `createPortableDocumentField` to place a field on a page. The smallest call binds a record attribute (`field`) to a position (`positionX` / `positionY`) on a page (`pageId`).

```graphql
mutation PlaceTitleField {
  createPortableDocumentField(
    input: { pageId: "page_123", field: "title", positionX: 120.0, positionY: 64.0 }
  ) {
    id
    field
    positionX
    positionY
  }
}
```

To bind a custom field, set `field: "customField"` and supply the `customFieldId` — omitting it raises `CUSTOM_FIELD_NOT_FOUND`. You can also pass `width` and `height` to define a bounding box (text wraps inside it; without bounds the value is drawn on a single line at the anchor point).

```graphql
mutation PlaceCustomFieldBox {
  createPortableDocumentField(
    input: {
      pageId: "page_123"
      field: "customField"
      customFieldId: "field_123"
      positionX: 120.0
      positionY: 220.0
      width: 240.0
      height: 48.0
    }
  ) {
    id
    field
    width
    height
    customField {
      id
      name
    }
  }
}
```

### CreatePortableDocumentFieldInput

| Parameter       | Type      | Required | Description                                                                                                                    |
| --------------- | --------- | -------- | ------------------------------------------------------------------------------------------------------------------------------ |
| `pageId`        | `String!` | Yes      | The `PortableDocumentPage` this field is placed on. The page's parent template determines which workspace's permissions apply. |
| `field`         | `String!` | Yes      | Which record attribute to bind. One of `title`, `description`, `dueEnd`, `assignees`, `tags`, or `customField`.                |
| `positionX`     | `Float!`  | Yes      | Horizontal anchor of the field on the page.                                                                                    |
| `positionY`     | `Float!`  | Yes      | Vertical anchor of the field on the page.                                                                                      |
| `width`         | `Float`   | No       | Bounding-box width. When set together with `height`, the printed value wraps to fit.                                           |
| `height`        | `Float`   | No       | Bounding-box height.                                                                                                           |
| `customFieldId` | `String`  | No       | Required when `field` is `customField` — the `CustomField` to read for each record. Ignored for other `field` values.          |
| `id`            | `String`  | No       | Pass an existing field id to upsert (reuse) that id; omit it to create a new field.                                            |

#### `field` values

| Value         | Binds to the record's                                             |
| ------------- | ----------------------------------------------------------------- |
| `title`       | Title (`Todo.title`).                                             |
| `description` | Description text (`Todo.text`).                                   |
| `dueEnd`      | Due date (`Todo.duedAt`), formatted in the record's timezone.     |
| `assignees`   | Assignee names, comma-separated.                                  |
| `tags`        | Tag titles, comma-separated.                                      |
| `customField` | The custom field named by `customFieldId`, formatted by its type. |

## Response

```json
{
  "data": {
    "createPortableDocumentField": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "field": "title",
      "positionX": 120.0,
      "positionY": 64.0
    }
  }
}
```

### PortableDocumentField

| Field         | Type                    | Description                                                                                    |
| ------------- | ----------------------- | ---------------------------------------------------------------------------------------------- |
| `id`          | `ID!`                   | Unique identifier of the field.                                                                |
| `uid`         | `String!`               | Short opaque identifier.                                                                       |
| `field`       | `String!`               | The bound attribute (`title`, `description`, `dueEnd`, `assignees`, `tags`, or `customField`). |
| `positionX`   | `Float!`                | Horizontal anchor on the page.                                                                 |
| `positionY`   | `Float!`                | Vertical anchor on the page.                                                                   |
| `width`       | `Float`                 | Bounding-box width, if set.                                                                    |
| `height`      | `Float`                 | Bounding-box height, if set.                                                                   |
| `createdAt`   | `DateTime!`             | When the field was created.                                                                    |
| `updatedAt`   | `DateTime!`             | When the field was last changed.                                                               |
| `page`        | `PortableDocumentPage!` | The page the field is placed on.                                                               |
| `customField` | `CustomField`           | The bound custom field, when `field` is `customField`; otherwise `null`.                       |

## Update a field

Use `updatePortableDocumentField` to reposition or resize an existing field, or to move it to a different page with `pageId`. `positionX` and `positionY` are required on every update; `width`, `height`, and `pageId` are optional.

```graphql
mutation MoveField {
  updatePortableDocumentField(
    input: { id: "field_123", positionX: 96.0, positionY: 300.0, width: 200.0, height: 40.0 }
  ) {
    id
    positionX
    positionY
    width
    height
  }
}
```

### UpdatePortableDocumentFieldInput

| Parameter   | Type      | Required | Description                                                              |
| ----------- | --------- | -------- | ------------------------------------------------------------------------ |
| `id`        | `String!` | Yes      | The `PortableDocumentField` to update.                                   |
| `positionX` | `Float!`  | Yes      | New horizontal anchor.                                                   |
| `positionY` | `Float!`  | Yes      | New vertical anchor.                                                     |
| `width`     | `Float`   | No       | New bounding-box width.                                                  |
| `height`    | `Float`   | No       | New bounding-box height.                                                 |
| `pageId`    | `String`  | No       | Move the field to a different page. Omit to keep it on its current page. |

`updatePortableDocumentField` returns the updated `PortableDocumentField`. The bound attribute (`field`) and any `customField` are fixed at creation — to change what a field binds to, delete it and create a new one.

## Delete a field

Use `deletePortableDocumentField` to remove a field by id. It returns `Boolean` (`true` on success), so it takes no sub-selection.

```graphql
mutation RemoveField {
  deletePortableDocumentField(id: "field_123")
}
```

```json
{
  "data": {
    "deletePortableDocumentField": true
  }
}
```

## Print a record

Use `printPortableDocument` to render a filled PDF: it takes the template id (`pdfId`) and a record id (`todoId`), draws every overlay field's value for that record onto the template's pages, and returns the new file's storage key.

```graphql
mutation PrintRecord {
  printPortableDocument(pdfId: "pdf_123", todoId: "todo_123")
}
```

```json
{
  "data": {
    "printPortableDocument": "clm4n8qwx000008l0g4oxdqn7/acme-invoice.pdf"
  }
}
```

<Callout variant="info" title="The return value is a storage key, not a URL">

`printPortableDocument` returns a legacy storage key string of the form <code>{file.uid}/{file.name}</code> — not a downloadable URL — or `null` if no file was produced. Resolve it through Blue's file APIs to fetch the rendered PDF. See [Upload files](/api/files).

</Callout>

## Errors

### createPortableDocumentField

| Code                          | When                                                                                          |
| ----------------------------- | --------------------------------------------------------------------------------------------- |
| `PORTABLE_DOCUMENT_NOT_FOUND` | No template owns the given `pageId`.                                                          |
| `CUSTOM_FIELD_NOT_FOUND`      | `field` is `customField` but `customFieldId` is missing or doesn't resolve to a custom field. |
| `FORBIDDEN`                   | The caller lacks contribute permission on the template's workspace.                           |
| `UNAUTHENTICATED`             | No valid credentials on the request.                                                          |

### updatePortableDocumentField / deletePortableDocumentField

| Code                                | When                                                                         |
| ----------------------------------- | ---------------------------------------------------------------------------- |
| `PORTABLE_DOCUMENT_FIELD_NOT_FOUND` | No field matches the given `id`.                                             |
| `FORBIDDEN`                         | (delete) The caller lacks contribute permission on the template's workspace. |
| `UNAUTHENTICATED`                   | No valid credentials on the request.                                         |

### printPortableDocument

| Code                          | When                                 |
| ----------------------------- | ------------------------------------ |
| `TODO_NOT_FOUND`              | No record matches `todoId`.          |
| `PORTABLE_DOCUMENT_NOT_FOUND` | No template matches `pdfId`.         |
| `UNAUTHENTICATED`             | No valid credentials on the request. |

## Permissions

`createPortableDocumentField` and `deletePortableDocumentField` require **contribute** permission on the workspace that owns the field's template — the same level needed to edit records there. `updatePortableDocumentField` and `printPortableDocument` require an authenticated caller. All four are scoped by the `blue-org-id` and `blue-workspace-id` headers (or an ID/slug in the input). Header names are case-insensitive.

## Related

- [Build Portable Document templates](/api/documents/portable-documents) — upload a PDF and create the template and pages these fields attach to.
- [Documents overview](/api/documents) — how the two document subsystems differ.
- [Custom fields](/api/custom-fields) — define the `CustomField` a `customField` overlay reads.
- [List records](/api/records/list-records) — find the `todoId` to print.
- [Upload files](/api/files) — resolve the storage key returned by `printPortableDocument`.
