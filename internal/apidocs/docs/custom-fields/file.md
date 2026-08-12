---
title: File Custom Field
description: Attach uploaded files to records with a file custom field, then read them back via the field's files list.
icon: Paperclip
order: 22
---

A file custom field attaches one or more uploaded files to a record. This is the `FILE` value of the `CustomFieldType` enum. Records are `Record` objects and workspaces are `Workspace` objects in the API.

Unlike most field types, file values are not written with `setRecordCustomField`. The flow is three steps: upload the file with `uploadFile` to get a file `uid`, attach that file to the record's field with `createRecordCustomFieldFile`, then read the attached files back from the field's `files` list.

## Overview

Create the field with `createCustomField` using `type: FILE`. The field is scoped to the workspace in your `X-Bloo-Project-ID` header — there is no `projectId` parameter on the input. A single file field can hold multiple files; each is attached with its own `createRecordCustomFieldFile` call.

## Create

```graphql
mutation CreateFileField {
  createCustomField(input: { name: "Attachments", type: FILE }) {
    id
    name
    type
  }
}
```

### CreateCustomFieldInput

| Parameter     | Type               | Required | Description                |
| ------------- | ------------------ | -------- | -------------------------- |
| `name`        | `String!`          | Yes      | Display name of the field. |
| `type`        | `CustomFieldType!` | Yes      | Must be `FILE`.            |
| `description` | `String`           | No       | Help text shown to users.  |

The field is associated with the workspace in your `blue-workspace-id` header. `CreateCustomFieldInput` has no `projectId` field.

```json
{
  "data": {
    "createCustomField": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Attachments",
      "type": "FILE"
    }
  }
}
```

## Set a value

Setting a file value is a two-step operation: upload, then attach.

### Step 1 — Upload the file

Use the `uploadFile` mutation. This is a GraphQL multipart upload: the `file` variable is the file part of the request, not an inline string. See [Upload Files](/api/files) for the full multipart request format. `uploadFile` returns a `File`; keep its `uid`.

```graphql
mutation UploadFile($file: Upload!) {
  uploadFile(input: { file: $file, companyId: "company_123", projectId: "project_123" }) {
    uid
    name
    size
    type
    status
  }
}
```

#### UploadFileInput

| Parameter   | Type      | Required | Description                                                |
| ----------- | --------- | -------- | ---------------------------------------------------------- |
| `file`      | `Upload!` | Yes      | The file to upload (multipart file part).                  |
| `companyId` | `String!` | Yes      | Organization ID or slug where the file is stored.          |
| `projectId` | `String`  | No       | Workspace ID or slug. Omit for an organization-level file. |

```json
{
  "data": {
    "uploadFile": {
      "uid": "clm4n8qwx000108l0a1b2c3d4",
      "name": "spec.pdf",
      "size": 184320,
      "type": "application/pdf",
      "status": "CONFIRMED"
    }
  }
}
```

### Step 2 — Attach the file to the record

Use `createRecordCustomFieldFile` with the record ID, the file field ID, and the `uid` from the upload. It returns `Boolean` (`true` on success) — it does **not** return the created object, so do not add a sub-selection.

```graphql
mutation AttachFile {
  createRecordCustomFieldFile(
    input: { todoId: "todo_123", customFieldId: "field_123", fileUid: "clm4n8qwx000108l0a1b2c3d4" }
  )
}
```

#### CreateRecordCustomFieldFileInput

| Parameter       | Type      | Required | Description                                   |
| --------------- | --------- | -------- | --------------------------------------------- |
| `todoId`        | `String!` | Yes      | ID of the record.                             |
| `customFieldId` | `String!` | Yes      | ID of the file custom field.                  |
| `fileUid`       | `String!` | Yes      | `uid` of the uploaded file from `uploadFile`. |

```json
{ "data": { "createRecordCustomFieldFile": true } }
```

Attaching an already-attached file is a no-op and still returns `true`. To attach several files, call `createRecordCustomFieldFile` once per file `uid` — there is no bulk-attach mutation.

## Read a value

Read attached files from the field's `files` list. Query the record's `customFields` and select `files` on the `FILE` field — there is no separate junction type.

```graphql
query RecordFiles {
  recordQueries {
    todos(filter: { companyIds: ["company_123"], todoIds: ["todo_123"] }) {
      items {
        id
        title
        customFields(filter: { todoId: "todo_123" }) {
          id
          name
          type
          files {
            uid
            name
            size
            type
            extension
            createdAt
          }
        }
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
            "id": "todo_123",
            "title": "Project deliverables",
            "customFields": [
              {
                "id": "clm4n8qwx000008l0g4oxdqn7",
                "name": "Attachments",
                "type": "FILE",
                "files": [
                  {
                    "uid": "clm4n8qwx000108l0a1b2c3d4",
                    "name": "spec.pdf",
                    "size": 184320,
                    "type": "application/pdf",
                    "extension": "pdf",
                    "createdAt": "2026-05-29T10:14:00.000Z"
                  }
                ]
              }
            ]
          }
        ]
      }
    }
  }
}
```

### File

The `files` list returns `File` objects.

| Field       | Type         | Description                                                                |
| ----------- | ------------ | -------------------------------------------------------------------------- |
| `id`        | `ID!`        | Database ID.                                                               |
| `uid`       | `String!`    | Stable file identifier; use this as `fileUid` when attaching or detaching. |
| `name`      | `String!`    | Original filename.                                                         |
| `size`      | `Float!`     | File size in bytes.                                                        |
| `type`      | `String!`    | MIME type.                                                                 |
| `extension` | `String!`    | File extension.                                                            |
| `shared`    | `Boolean!`   | Whether the file is shared.                                                |
| `status`    | `FileStatus` | `CONFIRMED` or `PENDING`.                                                  |
| `createdAt` | `DateTime!`  | Upload timestamp.                                                          |

## Remove a file

Detach a file with `deleteRecordCustomFieldFile`, identifying it by the same `fileUid`. It also returns `Boolean`.

```graphql
mutation DetachFile {
  deleteRecordCustomFieldFile(
    input: { todoId: "todo_123", customFieldId: "field_123", fileUid: "clm4n8qwx000108l0a1b2c3d4" }
  )
}
```

`DeleteRecordCustomFieldFileInput` has the same shape as the attach input: `todoId`, `customFieldId`, and `fileUid` (all `String!`).

```json
{ "data": { "deleteRecordCustomFieldFile": true } }
```

## Notes

- **Per-file size limits are plan-dependent.** The cap is enforced at upload time against your plan's maximum upload size; exceeding it raises `PLAN_LIMIT_REACHED`. Hitting your total storage allowance raises the same code.
- **Batch uploads** (`uploadFiles`) are capped at 1 GB total per request; upload files larger than that individually.
- Files are stored on Blue's managed object storage (Backblaze B2 as the primary store) and served through short-lived signed URLs. The storage backend is an implementation detail — work with `uid` and the `File` fields, not storage paths.

## Errors

| Code                     | When                                                                                                 |
| ------------------------ | ---------------------------------------------------------------------------------------------------- |
| `CUSTOM_FIELD_NOT_FOUND` | The `customFieldId` does not exist.                                                                  |
| `FILE_NOT_FOUND`         | The `fileUid` does not match an uploaded file, or (on detach) the file is not attached to the field. |
| `TODO_NOT_FOUND`         | The `todoId` does not exist.                                                                         |
| `PLAN_LIMIT_REACHED`     | Upload exceeds the plan's per-file size or total storage limit.                                      |
| `FORBIDDEN`              | The caller lacks permission to modify the record.                                                    |

## Permissions

Creating or updating a file custom field requires the `OWNER` or `ADMIN` workspace role. Attaching and detaching files requires an `ADMIN`, `OWNER`, `MEMBER`, or `CLIENT` role on the record's workspace. Reading files follows standard record-view permissions.

## Related

- [Create a Custom Field](/api/custom-fields/create-custom-fields)
- [Set Custom Field Values](/api/custom-fields/custom-field-values)
- [Upload Files](/api/files)
- [Custom Fields Overview](/api/custom-fields)
