---
title: File & Link Subscriptions
description: Stream file and link changes over WebSocket — scope by organization, workspace, or folder, and react to batch deletions in real time.
icon: Files
order: 6
---

Keep attachment lists and link rails live without polling. The `subscribeToFile`, `subscribeToLink`, and `onDeleteFiles` subscriptions push file (`File`) and link (`Link`) changes to your client the moment they happen on the server.

These are streamed over WebSocket using the [graphql-ws](https://github.com/enisdenjo/graphql-ws) protocol at `wss://api.blue.app/graphql` — the same path as the HTTP GraphQL endpoint. Open and authenticate the connection first; see [Connect & Authenticate](/api/realtime/connect-and-authenticate) for the handshake and credential forms. The examples below show only the subscription documents.

Like the other entity change-feeds, `subscribeToFile` and `subscribeToLink` use the shared payload shape `{ mutation, node, previousValues, updatedFields }` described on the [Real-time overview](/api/realtime). `onDeleteFiles` is an imperative batch event that returns the deleted files directly as a list.

## subscribeToFile

Stream create, update, and delete events for files. A _file_ is a `File` object — an uploaded attachment that lives under a workspace and, optionally, a folder. Scope the stream by organization (`companyId`), workspace (`projectId`), and optionally folder (`folderId`).

### Request

```graphql
subscription OnFileChange {
  subscribeToFile(companyId: "company_123", projectId: "workspace_123") {
    mutation
    node {
      id
      name
      extension
      size
      status
      folder {
        id
        title
      }
    }
    updatedFields
  }
}
```

Pass `folderId` to narrow the stream to a single folder — you then only receive events for files in that folder:

```graphql
subscription OnFolderFileChange {
  subscribeToFile(companyId: "company_123", projectId: "workspace_123", folderId: "folder_123") {
    mutation
    node {
      id
      name
      folder {
        id
      }
    }
  }
}
```

### Parameters

| Parameter   | Type     | Required | Description                                                                                                                          |
| ----------- | -------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `companyId` | `String` | No       | Organization ID or slug. Only events whose file belongs to this organization are delivered.                                          |
| `projectId` | `String` | No       | Workspace ID or slug. Only events whose file belongs to this workspace are delivered.                                                |
| `folderId`  | `String` | No       | Folder ID. When provided, only events for files in this folder are delivered. Omit to receive events for all files in the workspace. |

<Callout variant="info" title="Scope to a workspace you can access">

Each event is delivered only if you are a member of the file's workspace. Pass the same `companyId` and `projectId` you scope the rest of your client to. Company and workspace arguments accept an ID or a slug.

</Callout>

### Response

`subscribeToFile` returns a `FileSubscriptionPayload` per event.

```json
{
  "data": {
    "subscribeToFile": {
      "mutation": "CREATED",
      "node": {
        "id": "clm4n8qwx000008l0g4oxdqn7",
        "name": "Q3-report.pdf",
        "extension": "pdf",
        "size": 482113,
        "status": "CONFIRMED",
        "folder": {
          "id": "clm4n8qwx000108l0abcd1234",
          "title": "Reports"
        }
      },
      "updatedFields": null
    }
  }
}
```

#### FileSubscriptionPayload

| Field            | Type                 | Description                                                                              |
| ---------------- | -------------------- | ---------------------------------------------------------------------------------------- |
| `mutation`       | `MutationType!`      | The kind of change: `CREATED`, `UPDATED`, or `DELETED`.                                  |
| `node`           | `File`               | The file after the change. `null` on a `DELETED` event — read `previousValues` instead.  |
| `previousValues` | `FilePreviousValues` | The file's scalar values before the change. Populated on `UPDATED` and `DELETED` events. |
| `updatedFields`  | `[String!]`          | Names of the scalar fields that changed on an `UPDATED` event. `null` otherwise.         |

#### File

The most useful fields on the `node`. A `File` also exposes its parent objects (`company`, `project`, `folder`, `comment`, `chat`, `statusUpdate`, `customField`, `document`, `todo`, `user`) — note these are the API field names; select only what you render.

| Field       | Type         | Description                                                                        |
| ----------- | ------------ | ---------------------------------------------------------------------------------- |
| `id`        | `ID!`        | The file's unique identifier.                                                      |
| `uid`       | `String!`    | Storage-layer identifier used to build download URLs.                              |
| `name`      | `String!`    | Display file name.                                                                 |
| `size`      | `Float!`     | File size in bytes.                                                                |
| `type`      | `String!`    | MIME type.                                                                         |
| `extension` | `String!`    | File extension without the dot (e.g. `pdf`).                                       |
| `shared`    | `Boolean!`   | Whether the file is shared publicly.                                               |
| `status`    | `FileStatus` | Upload lifecycle state: `PENDING` until the upload is confirmed, then `CONFIRMED`. |
| `position`  | `Float`      | Sort position within its container.                                                |
| `folder`    | `Folder`     | The folder the file lives in, if any.                                              |
| `createdAt` | `DateTime!`  | When the file was created.                                                         |
| `updatedAt` | `DateTime!`  | When the file was last modified.                                                   |

#### FilePreviousValues

| Field       | Type        | Description                      |
| ----------- | ----------- | -------------------------------- |
| `id`        | `ID!`       | The file's identifier.           |
| `uid`       | `String!`   | Storage-layer identifier.        |
| `name`      | `String!`   | Previous file name.              |
| `size`      | `Float!`    | Previous size in bytes.          |
| `type`      | `String!`   | Previous MIME type.              |
| `extension` | `String!`   | Previous extension.              |
| `shared`    | `Boolean`   | Previous shared flag.            |
| `createdAt` | `DateTime!` | When the file was created.       |
| `updatedAt` | `DateTime!` | When the file was last modified. |

#### FileStatus

| Value       | Description                                          |
| ----------- | ---------------------------------------------------- |
| `PENDING`   | The upload has been initiated but not yet confirmed. |
| `CONFIRMED` | The upload completed and the file is available.      |

## subscribeToLink

Stream create, update, and delete events for links. A _link_ is a `Link` object — a saved URL that belongs to an organization. Links are scoped at the organization level only, so `subscribeToLink` takes just `companyId`.

### Request

```graphql
subscription OnLinkChange {
  subscribeToLink(companyId: "company_123") {
    mutation
    node {
      id
      title
      url
      description
    }
    updatedFields
  }
}
```

### Parameters

| Parameter   | Type     | Required | Description                                                                                                                                    |
| ----------- | -------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `companyId` | `String` | No       | Organization ID or slug. Only events whose link belongs to this organization are delivered, and only if you are a member of that organization. |

### Response

`subscribeToLink` returns a `LinkSubscriptionPayload` per event.

```json
{
  "data": {
    "subscribeToLink": {
      "mutation": "UPDATED",
      "node": {
        "id": "clm4n8qwx000208l0wxyz5678",
        "title": "Brand guidelines",
        "url": "https://example.com/brand",
        "description": "Logo usage and color palette"
      },
      "updatedFields": ["title", "url"]
    }
  }
}
```

#### LinkSubscriptionPayload

| Field            | Type                 | Description                                                                              |
| ---------------- | -------------------- | ---------------------------------------------------------------------------------------- |
| `mutation`       | `MutationType!`      | The kind of change: `CREATED`, `UPDATED`, or `DELETED`.                                  |
| `node`           | `Link`               | The link after the change. `null` on a `DELETED` event — read `previousValues` instead.  |
| `previousValues` | `LinkPreviousValues` | The link's scalar values before the change. Populated on `UPDATED` and `DELETED` events. |
| `updatedFields`  | `[String!]`          | Names of the scalar fields that changed on an `UPDATED` event. `null` otherwise.         |

#### Link

| Field         | Type        | Description                                             |
| ------------- | ----------- | ------------------------------------------------------- |
| `id`          | `ID!`       | The link's unique identifier.                           |
| `uid`         | `String!`   | Internal identifier.                                    |
| `title`       | `String!`   | Display title.                                          |
| `url`         | `String!`   | The target URL.                                         |
| `description` | `String`    | Optional description.                                   |
| `position`    | `Float!`    | Sort position.                                          |
| `membersOnly` | `Boolean`   | Whether the link is restricted to organization members. |
| `createdBy`   | `User!`     | The user who created the link.                          |
| `company`     | `Organization!` | The organization the link belongs to.               |
| `createdAt`   | `DateTime!` | When the link was created.                              |
| `updatedAt`   | `DateTime!` | When the link was last modified.                        |

#### LinkPreviousValues

| Field         | Type        | Description                      |
| ------------- | ----------- | -------------------------------- |
| `id`          | `ID!`       | The link's identifier.           |
| `uid`         | `String!`   | Internal identifier.             |
| `title`       | `String!`   | Previous title.                  |
| `url`         | `String!`   | Previous URL.                    |
| `description` | `String`    | Previous description.            |
| `position`    | `Float!`    | Previous sort position.          |
| `membersOnly` | `Boolean`   | Previous members-only flag.      |
| `createdAt`   | `DateTime!` | When the link was created.       |
| `updatedAt`   | `DateTime!` | When the link was last modified. |

## onDeleteFiles

Receive a single event whenever a batch of files is deleted in a workspace. Unlike `subscribeToFile` — which emits one `DELETED` event per file — `onDeleteFiles` delivers all files removed in one operation as a single `[File!]!` list. Use it to drop several attachments from a rendered list in one update.

### Request

```graphql
subscription OnFilesDeleted {
  onDeleteFiles(projectId: "workspace_123") {
    id
    name
    folder {
      id
    }
  }
}
```

### Parameters

| Parameter   | Type      | Required | Description                                                                                                        |
| ----------- | --------- | -------- | ------------------------------------------------------------------------------------------------------------------ |
| `projectId` | `String!` | Yes      | Workspace ID. Events are delivered only for batch deletions in this workspace, and only if you are a member of it. |

### Response

`onDeleteFiles` returns the deleted files directly as `[File!]!` — there is no `mutation`/`previousValues` wrapper. See the [File](#file) fields above for what you can select.

```json
{
  "data": {
    "onDeleteFiles": [
      {
        "id": "clm4n8qwx000008l0g4oxdqn7",
        "name": "Q3-report.pdf",
        "folder": { "id": "clm4n8qwx000108l0abcd1234" }
      },
      {
        "id": "clm4n8qwx000308l0efgh9012",
        "name": "Q3-appendix.xlsx",
        "folder": { "id": "clm4n8qwx000108l0abcd1234" }
      }
    ]
  }
}
```

## Permissions

Authentication is enforced at the WebSocket handshake — an unauthenticated connection is rejected before any subscription runs. Beyond that, delivery is gated **per event**:

- **`subscribeToFile`** and **`onDeleteFiles`** deliver an event only if you are a member of the file's workspace. `subscribeToFile` additionally matches on `companyId`/`projectId` (and `folderId` when supplied); `onDeleteFiles` matches on `projectId`.
- **`subscribeToLink`** delivers an event only if you are a member of the link's organization, matched on `companyId`.

Subscribing always succeeds; you simply receive events only for files and links you can see. See [Connect & Authenticate](/api/realtime/connect-and-authenticate) for credentials.

## Related

- [Real-time (Subscriptions)](/api/realtime) — the shared payload shape and the full subscription catalog.
- [Connect & Authenticate](/api/realtime/connect-and-authenticate) — open and authenticate the WebSocket.
- [Upload Files](/api/files) — the HTTP side of the files API that produces these events.
- [Project & Workspace Subscriptions](/api/realtime/project-subscriptions) — folder, tag, and membership streams for the same workspace.
