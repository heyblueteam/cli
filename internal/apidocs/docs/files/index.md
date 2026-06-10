---
title: Uploading Files
description: Upload files to Blue via a one-step GraphQL mutation or a presigned-PUT REST flow for large files.
icon: Upload
order: 1
---

Blue stores files as `File` objects scoped to an organization (`Company`) and, usually, a workspace (`Project`). There are two ways to get bytes into storage:

- **GraphQL multipart upload** (`uploadFile` / `uploadFiles`) — one request, up to **256 MB** per file. Use this for almost everything.
- **REST presigned-PUT flow** (`GET /uploads` → `PUT` to storage → `PUT /uploads/{uid}/confirm`) — three requests, up to **4.8 GB** per file. Use this only for files larger than 256 MB, where streaming bytes straight to storage avoids holding the whole file in the API.

|                | GraphQL upload                      | REST presigned PUT                  |
| -------------- | ----------------------------------- | ----------------------------------- |
| Requests       | One                                 | Three                               |
| Per-file limit | 256 MB                              | 4.8 GB                              |
| Batch          | Up to 10 files (1 GB total)         | One file per flow                   |
| Transport      | `multipart/form-data` to `/graphql` | `PUT` bytes to a signed storage URL |
| Best for       | Almost everything                   | Files over 256 MB                   |

All examples authenticate with personal access token headers and target `https://api.blue.app`. See [Authentication](/api/start-guide/authentication) for how to generate a token. Headers are case-insensitive; this page uses the canonical casing.

---

## GraphQL upload

Use the `uploadFile` mutation to upload a single file and the `uploadFiles` mutation to upload a batch. Both use the [GraphQL multipart request spec](https://github.com/jaydenseric/graphql-multipart-request-spec): the `file` (or `files`) variable is sent as `null` in the `operations` JSON and mapped to a multipart part.

### Request

A single-file upload returns the created `File`:

```graphql
mutation UploadFile($input: UploadFileInput!) {
  uploadFile(input: $input) {
    id
    uid
    name
    size
    type
    extension
    shared
    status
    createdAt
    project {
      id
      name
    }
    folder {
      id
      title
    }
  }
}
```

A freshly uploaded file is not placed in a folder, so `folder` is `null` unless you later move it. The `status` of a file created through `uploadFile`/`uploadFiles` is `CONFIRMED` immediately — the lifecycle `PENDING` state only applies to the REST flow below.

### Parameters

#### UploadFileInput

| Parameter   | Type      | Required | Description                                                |
| ----------- | --------- | -------- | ---------------------------------------------------------- |
| `file`      | `Upload!` | Yes      | The file to upload, sent as a multipart part.              |
| `companyId` | `String!` | Yes      | Organization ID or slug where the file is stored.          |
| `projectId` | `String`  | No       | Workspace ID or slug. Omit for an organization-level file. |

#### UploadFilesInput

| Parameter   | Type         | Required | Description                                              |
| ----------- | ------------ | -------- | -------------------------------------------------------- |
| `files`     | `[Upload!]!` | Yes      | The files to upload. Maximum 10, 1 GB total per request. |
| `companyId` | `String!`    | Yes      | Organization ID or slug where the files are stored.      |
| `projectId` | `String!`    | Yes      | Workspace ID or slug. Required for batch uploads.        |

<Callout variant="info" title="projectId differs between the two mutations">

`uploadFile` accepts a nullable `projectId` — omit it to attach the file to the organization rather than a workspace. `uploadFiles` requires `projectId`; every file in a batch lands in the same workspace.

</Callout>

### Response

```json
{
  "data": {
    "uploadFile": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "uid": "cm8qujq3b01p22lrv9xbb2q62",
      "name": "report.pdf",
      "size": 204800,
      "type": "application/pdf",
      "extension": "pdf",
      "shared": false,
      "status": "CONFIRMED",
      "createdAt": "2026-05-29T12:00:00.000Z",
      "project": { "id": "project_123", "name": "Acme Onboarding" },
      "folder": null
    }
  }
}
```

#### Returns — File

| Field       | Type         | Description                                                     |
| ----------- | ------------ | --------------------------------------------------------------- |
| `id`        | `ID!`        | The file's database ID.                                         |
| `uid`       | `String!`    | The storage UID, used to build [download URLs](#serving-files). |
| `name`      | `String!`    | Original filename.                                              |
| `size`      | `Float!`     | File size in bytes.                                             |
| `type`      | `String!`    | MIME type.                                                      |
| `extension` | `String!`    | File extension, no leading dot.                                 |
| `shared`    | `Boolean!`   | Whether the file is publicly shareable. Defaults to `false`.    |
| `status`    | `FileStatus` | `CONFIRMED` or `PENDING`. See [File status](#file-status).      |
| `createdAt` | `DateTime!`  | When the file was created.                                      |
| `project`   | `Project!`   | The workspace the file belongs to.                              |
| `folder`    | `Folder`     | The folder the file is in, or `null`.                           |

### Full example

#### cURL

The multipart body has three parts: `operations` (the query and variables, with `file` set to `null`), `map` (which part fills which variable), and the file itself.

```bash
curl https://api.blue.app/graphql \
  -H "X-Bloo-Token-ID: YOUR_TOKEN_ID" \
  -H "X-Bloo-Token-Secret: YOUR_TOKEN_SECRET" \
  -H "X-Bloo-Company-ID: company_123" \
  -H "X-Bloo-Project-ID: project_123" \
  -F 'operations={"query":"mutation UploadFile($input: UploadFileInput!) { uploadFile(input: $input) { id uid name size type extension } }","variables":{"input":{"file":null,"companyId":"company_123","projectId":"project_123"}}}' \
  -F 'map={"0":["variables.input.file"]}' \
  -F '0=@/path/to/report.pdf'
```

#### Python

```python
import json
import os
import requests

BASE_URL = "https://api.blue.app"
HEADERS = {
    "X-Bloo-Token-ID": "YOUR_TOKEN_ID",
    "X-Bloo-Token-Secret": "YOUR_TOKEN_SECRET",
    "X-Bloo-Company-ID": "company_123",
    "X-Bloo-Project-ID": "project_123",
}

def upload_file(filepath):
    filename = os.path.basename(filepath)
    query = """
    mutation UploadFile($input: UploadFileInput!) {
      uploadFile(input: $input) { id uid name size type extension status }
    }
    """
    variables = {"input": {"file": None, "companyId": "company_123", "projectId": "project_123"}}
    with open(filepath, "rb") as f:
        resp = requests.post(
            f"{BASE_URL}/graphql",
            headers=HEADERS,
            data={
                "operations": json.dumps({"query": query, "variables": variables}),
                "map": json.dumps({"0": ["variables.input.file"]}),
            },
            files={"0": (filename, f)},
        )
    resp.raise_for_status()
    result = resp.json()
    if "errors" in result:
        raise Exception(result["errors"])
    return result["data"]["uploadFile"]

print(upload_file("report.pdf"))
```

#### Batch upload

`uploadFiles` maps each file to its own multipart part:

```graphql
mutation UploadFiles($input: UploadFilesInput!) {
  uploadFiles(input: $input) {
    id
    uid
    name
    size
    type
    extension
  }
}
```

```bash
curl https://api.blue.app/graphql \
  -H "X-Bloo-Token-ID: YOUR_TOKEN_ID" \
  -H "X-Bloo-Token-Secret: YOUR_TOKEN_SECRET" \
  -H "X-Bloo-Company-ID: company_123" \
  -H "X-Bloo-Project-ID: project_123" \
  -F 'operations={"query":"mutation UploadFiles($input: UploadFilesInput!) { uploadFiles(input: $input) { id uid name } }","variables":{"input":{"files":[null,null],"companyId":"company_123","projectId":"project_123"}}}' \
  -F 'map={"0":["variables.input.files.0"],"1":["variables.input.files.1"]}' \
  -F '0=@/path/to/first.pdf' \
  -F '1=@/path/to/second.pdf'
```

```json
{
  "data": {
    "uploadFiles": [
      {
        "id": "clm4n8qwx000008l0g4oxdqn7",
        "uid": "cm8qujq3b01p22lrv9xbb2q62",
        "name": "first.pdf",
        "size": 102400,
        "type": "application/pdf",
        "extension": "pdf"
      },
      {
        "id": "clm4n8qwx000108l0h5pyerao8",
        "uid": "cm8qujq3b01p32lrv9xbb2q73",
        "name": "second.pdf",
        "size": 158720,
        "type": "application/pdf",
        "extension": "pdf"
      }
    ]
  }
}
```

### Errors

| Code                      | When                                                                               |
| ------------------------- | ---------------------------------------------------------------------------------- |
| `BAD_USER_INPUT`          | A file exceeds 256 MB, or a batch exceeds 10 files / 1 GB total.                   |
| `PLAN_LIMIT_REACHED`      | The upload would exceed the organization's storage allowance or per-file tier cap. |
| `FILES_DISABLED_FOR_ROLE` | The caller's role in the workspace has file uploads disabled.                      |
| `PROJECT_NOT_FOUND`       | `projectId` does not resolve, or the caller has no access to it.                   |
| `COMPANY_NOT_FOUND`       | `companyId` is missing or does not resolve.                                        |

---

## REST presigned-PUT upload

For files larger than 256 MB (up to **4.8 GB**), use the REST flow. The API never holds the file: you `PUT` the bytes directly to a presigned storage URL.

The flow is three requests:

1. **`GET /uploads`** — ask for a signed URL. The API creates a `File` row with `status: PENDING` and returns where to put the bytes.
2. **`PUT` the file bytes** to the returned signed URL.
3. **`PUT /uploads/{uid}/confirm`** — finalize. The API verifies the object landed and flips the row to `status: CONFIRMED`.

If you never confirm, a background verifier reconciles the row against storage shortly after the signing window closes, so an abandoned upload won't leave a dangling `PENDING` row forever — but you should always confirm so the file is usable immediately.

<Callout variant="warning" title="Both headers are required">

`GET /uploads` requires `X-Bloo-Company-ID` **and** `X-Bloo-Project-ID`. Omitting either returns `400` with `"Company ID and Project ID are required"`.

</Callout>

### Step 1 — Request a signed URL

```
GET https://api.blue.app/uploads?filename=large.zip&size=734003200
```

Query parameters:

| Parameter  | Required    | Description                                                                                       |
| ---------- | ----------- | ------------------------------------------------------------------------------------------------- |
| `filename` | Yes         | The filename, including extension. The MIME type and extension are derived from it.               |
| `size`     | Recommended | File size in bytes. Used to reject oversize uploads up front and to verify the object on confirm. |

Response:

```json
{
  "url": "https://s3.us-west-002.backblazeb2.com/bloo-uploads/companies/acme/projects/onboarding/uploads/2026/05/jdoe/cm8qujq3b01p22lrv9xbb2q62_large.zip?X-Amz-Expires=...&X-Amz-Signature=...",
  "method": "PUT",
  "headers": {
    "Content-Type": "application/zip"
  },
  "fields": {
    "Content-Type": "application/zip",
    "Key": "companies/acme/projects/onboarding/uploads/2026/05/jdoe/cm8qujq3b01p22lrv9xbb2q62_large.zip",
    "X-Key": "cm8qujq3b01p22lrv9xbb2q62"
  }
}
```

`fields["X-Key"]` is the file's `uid` — keep it for Step 3 and for [building download URLs](#serving-files).

### Step 2 — PUT the bytes

`PUT` the raw file bytes to `url`, sending the `Content-Type` from `headers`. This goes straight to storage, not to the Blue API. Do **not** wrap it in `multipart/form-data` — send the body as-is.

### Step 3 — Confirm

```
PUT https://api.blue.app/uploads/cm8qujq3b01p22lrv9xbb2q62/confirm
```

Send the same `X-Bloo-*` auth headers. On success the API returns `200` with `{ "message": "File confirmed successfully", "status": "CONFIRMED" }`. If the stored object's size doesn't match the declared `size`, confirm returns `400` and the file stays `PENDING`.

### Full example

```python
import os
import requests

BASE_URL = "https://api.blue.app"
HEADERS = {
    "X-Bloo-Token-ID": "YOUR_TOKEN_ID",
    "X-Bloo-Token-Secret": "YOUR_TOKEN_SECRET",
    "X-Bloo-Company-ID": "company_123",
    "X-Bloo-Project-ID": "project_123",
}

def upload_large_file(filepath):
    filename = os.path.basename(filepath)
    size = os.path.getsize(filepath)

    # Step 1: request a signed URL
    resp = requests.get(
        f"{BASE_URL}/uploads",
        headers=HEADERS,
        params={"filename": filename, "size": size},
    )
    resp.raise_for_status()
    signed = resp.json()
    uid = signed["fields"]["X-Key"]

    # Step 2: PUT the raw bytes to storage
    with open(filepath, "rb") as f:
        put = requests.put(
            signed["url"],
            data=f,
            headers={"Content-Type": signed["headers"]["Content-Type"]},
        )
    put.raise_for_status()

    # Step 3: confirm
    confirm = requests.put(f"{BASE_URL}/uploads/{uid}/confirm", headers=HEADERS)
    confirm.raise_for_status()
    return uid

print(upload_large_file("large.zip"))
```

### Errors

The REST endpoint returns HTTP status codes (it is not GraphQL):

| Status                            | When                                                                                                                            |
| --------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| `400`                             | `filename` missing, both company and project headers not supplied, the extension can't be determined, or `size` exceeds 4.8 GB. |
| `401`                             | No valid token.                                                                                                                 |
| `403` (`FILES_DISABLED_FOR_ROLE`) | The caller's workspace role has file uploads disabled.                                                                          |
| `404`                             | Company or project not found; or, on confirm, the file row or stored object is missing.                                         |

### Anonymous form uploads

The same endpoint accepts uploads from public forms without a token. Pass `?formId=<form-id>` instead of auth headers; the upload is gated by the form being active and is capped at **500 MB**. Confirm the same way with `PUT /uploads/{uid}/confirm?formId=<form-id>`. See [Forms](/api/forms) for building public forms.

---

## File status

`File.status` is a `FileStatus` enum reflecting the upload lifecycle:

| Value       | Meaning                                                                                                   |
| ----------- | --------------------------------------------------------------------------------------------------------- |
| `PENDING`   | The `File` row exists but the bytes are not yet confirmed in storage. Set during Step 1 of the REST flow. |
| `CONFIRMED` | The file is fully uploaded and usable. GraphQL uploads are `CONFIRMED` immediately.                       |

Treat `PENDING` files as not-yet-available; serve and attach only `CONFIRMED` files.

---

## Serving files

After upload, build a download URL from the file's `uid` and `name`:

```
https://api.blue.app/uploads/{uid}/{url_encoded_filename}
```

For `uid: "cm8qujq3b01p22lrv9xbb2q62"` and `name: "report.pdf"`:

```
https://api.blue.app/uploads/cm8qujq3b01p22lrv9xbb2q62/report.pdf
```

This redirects to a short-lived signed storage URL — no extra authentication needed once you have the link.

| Pattern                                                | Description                                    |
| ------------------------------------------------------ | ---------------------------------------------- |
| `/uploads/{uid}/{filename}`                            | Download the file.                             |
| `/uploads/{uid}/{filename}?content-disposition=inline` | Display in the browser instead of downloading. |
| `/uploads/{uid}/50x50/{filename}`                      | 50×50 thumbnail (images only).                 |
| `/uploads/{uid}/500x500/{filename}`                    | 500×500 preview (images only).                 |

```python
from urllib.parse import quote

file = result["data"]["uploadFile"]
download_url = f"https://api.blue.app/uploads/{file['uid']}/{quote(file['name'])}"
```

<Callout variant="warning" title="Both segments are required">

The URL needs both `uid` and `filename` — the `uid` alone returns `404`. Always URL-encode the filename so spaces and special characters survive.

</Callout>

---

## Attaching files to records and comments

There is no separate "attach file" mutation. To make an uploaded file appear as a clickable attachment inside a comment — exactly as it does in the Blue UI — embed its metadata as an HTML `<div>` in the comment body and set `tiptap: true`.

Comments are created with the `createComment` mutation; records are `Todo` objects, so a record comment uses `category: TODO`.

### How it works

1. Upload the file to get its `uid`, `name`, `size`, `type`, and `extension`.
2. Embed each file as an attachment `<div>` in the comment HTML.
3. Set `tiptap: true` in `CreateCommentInput` — **required** for the API to parse the attachment markup.

Each attachment is one element:

```html
<div
  class="attachment"
  file='{"uid":"FILE_UID","name":"FILENAME","size":SIZE_IN_BYTES,"type":"MIME_TYPE","extension":"EXT"}'
></div>
```

The `file` attribute is a JSON string built from the upload response:

| Field       | Type   | Example                       |
| ----------- | ------ | ----------------------------- |
| `uid`       | string | `"cm8qujq3b01p22lrv9xbb2q62"` |
| `name`      | string | `"report.pdf"`                |
| `size`      | number | `204800`                      |
| `type`      | string | `"application/pdf"`           |
| `extension` | string | `"pdf"`                       |

### CreateCommentInput

| Parameter    | Type               | Required | Description                                                            |
| ------------ | ------------------ | -------- | ---------------------------------------------------------------------- |
| `html`       | `String!`          | Yes      | Comment body as HTML, including any attachment `<div>`s.               |
| `text`       | `String!`          | Yes      | Plain-text fallback of the comment.                                    |
| `category`   | `CommentCategory!` | Yes      | `TODO`, `DISCUSSION`, or `STATUS_UPDATE`.                              |
| `categoryId` | `String!`          | Yes      | ID of the record, discussion, or status update the comment belongs to. |
| `tiptap`     | `Boolean`          | No       | Set `true` so attachment markup is parsed.                             |
| `parentId`   | `String`           | No       | ID of the comment this replies to.                                     |

### Example

```python
import json
import requests

BASE_URL = "https://api.blue.app"
HEADERS = {
    "X-Bloo-Token-ID": "YOUR_TOKEN_ID",
    "X-Bloo-Token-Secret": "YOUR_TOKEN_SECRET",
    "X-Bloo-Company-ID": "company_123",
    "X-Bloo-Project-ID": "project_123",
    "Content-Type": "application/json",
}

def comment_with_attachments(record_id, message, files):
    attachments = "".join(
        f'<div class="attachment" file=\'{json.dumps({k: f[k] for k in ("uid","name","size","type","extension")})}\'></div>'
        for f in files
    )
    query = """
    mutation CreateComment($input: CreateCommentInput!) {
      createComment(input: $input) { id html text createdAt }
    }
    """
    variables = {
        "input": {
            "html": f"<p>{message}</p>{attachments}",
            "text": message,
            "category": "TODO",
            "categoryId": record_id,
            "tiptap": True,
        }
    }
    resp = requests.post(f"{BASE_URL}/graphql", headers=HEADERS, json={"query": query, "variables": variables})
    resp.raise_for_status()
    result = resp.json()
    if "errors" in result:
        raise Exception(result["errors"])
    return result["data"]["createComment"]

file = upload_file("report.pdf")          # from the GraphQL upload example
comment_with_attachments("todo_123", "Here is the report you requested.", [file])
```

```json
{
  "data": {
    "createComment": {
      "id": "clm4n8qwx000208l0i6qzfbp9",
      "html": "<p>Here is the report you requested.</p><div class=\"attachment\" file='{\"uid\":\"cm8qujq3b01p22lrv9xbb2q62\",\"name\":\"report.pdf\",\"size\":204800,\"type\":\"application/pdf\",\"extension\":\"pdf\"}'></div>",
      "text": "Here is the report you requested.",
      "createdAt": "2026-05-29T12:01:00.000Z"
    }
  }
}
```

### Common mistakes

| Mistake                                                                  | Result                                                                   |
| ------------------------------------------------------------------------ | ------------------------------------------------------------------------ |
| Omitting `tiptap: true`                                                  | The attachment `<div>`s are ignored — files won't appear in the comment. |
| Using a placeholder `type` (e.g. `"file"`) instead of the real MIME type | The file may not render correctly in the UI.                             |
| Building the URL with `uid` alone                                        | `404` — the URL needs both `uid` and `filename`.                         |
| Not URL-encoding the filename                                            | Broken links for filenames with spaces or special characters.            |

---

## Attaching files to a file custom field

To populate a record's **file** custom field, upload the file, then link it with the `createTodoCustomFieldFile` mutation.

```graphql
mutation CreateTodoCustomFieldFile($input: CreateTodoCustomFieldFileInput!) {
  createTodoCustomFieldFile(input: $input)
}
```

### CreateTodoCustomFieldFileInput

| Parameter       | Type      | Required | Description                                  |
| --------------- | --------- | -------- | -------------------------------------------- |
| `todoId`        | `String!` | Yes      | ID of the record (`Todo`) holding the field. |
| `customFieldId` | `String!` | Yes      | ID of the file custom field.                 |
| `fileUid`       | `String!` | Yes      | `uid` of the uploaded file.                  |

The mutation returns `Boolean` — `true` on success. Do not request a sub-selection.

```json
{ "data": { "createTodoCustomFieldFile": true } }
```

See [File custom field](/api/custom-fields/file) for creating the field itself.

---

## Registering a pre-existing object (`createFile`)

If you already have a file's storage `uid` and metadata out of band — for example, you completed the presigned-PUT flow yourself and want a `File` row without calling `/uploads/{uid}/confirm` — use the `createFile` mutation. It is **not** part of the standard upload lifecycle; prefer `uploadFile` or the confirm step instead.

```graphql
mutation CreateFile($input: CreateFileInput!) {
  createFile(input: $input) {
    id
    uid
    name
    status
  }
}
```

### CreateFileInput

| Parameter    | Type      | Required | Description                                                  |
| ------------ | --------- | -------- | ------------------------------------------------------------ |
| `uid`        | `String!` | Yes      | Storage UID of the object.                                   |
| `name`       | `String!` | Yes      | Filename.                                                    |
| `size`       | `Float!`  | Yes      | Size in bytes.                                               |
| `type`       | `String!` | Yes      | MIME type.                                                   |
| `extension`  | `String!` | Yes      | Extension, no leading dot.                                   |
| `companyId`  | `String!` | Yes      | Organization ID or slug.                                     |
| `projectId`  | `String!` | Yes      | Workspace ID or slug.                                        |
| `folderId`   | `String`  | No       | Folder to place the file in.                                 |
| `documentId` | `String`  | No       | Document to associate the file with.                         |
| `shared`     | `Boolean` | No       | Whether the file is publicly shareable. Defaults to `false`. |

---

## Related

- [File custom field](/api/custom-fields/file)
- [Manage comments](/api/comments/manage-comments)
- [Add a comment to a record](/api/records/add-comment)
- [Authentication](/api/start-guide/authentication)
- [Error codes](/api/start-guide/error-codes)
