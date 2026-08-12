---
title: Comments, Chats & Chat
description: Real-time GraphQL subscriptions for comments, typing indicators, chats, status updates, AI chat, and collaborative documents.
icon: MessageSquare
order: 4
---

These subscriptions stream the conversational surfaces of a workspace: comments on records, chats, and status updates; the typing indicators that appear while someone is writing; the AI chat assistant; and collaborative documents. Open one over the same authenticated WebSocket you use for every other subscription — see [Connect & Authenticate](/api/realtime/connect-and-authenticate) for the `graphql-ws` handshake. The endpoint is `wss://api.blue.app/graphql`.

Workspaces are `Workspace` objects in the API, and records are `Record` objects. Comments, chats, and status updates all live inside a workspace, so every subscription here is scoped by a `projectId` (or, for `subscribeToComment`, the ID of the thing being commented on).

<Callout variant="info" title="Shared payload shape">

The entity subscriptions on this page (`subscribeToComment`, `subscribeToChat`, `subscribeToStatusUpdate`, `subscribeToDocument`) deliver the standard change-feed payload — `{ mutation, node, previousValues, updatedFields }` where `mutation` is `CREATED`, `UPDATED`, or `DELETED`. That shape is documented once on the [Real-time overview](/api/realtime). `subscribeToCommentTyping` deviates; it is called out below.

</Callout>

## subscribeToComment

Streams comment changes (create, edit, delete) for one thread. A thread is identified by a `category` and the ID of the parent it hangs off — a chat, a status update, or a record.

### Request

```graphql
subscription OnRecordComments {
  subscribeToComment(category: TODO, categoryId: "todo_123") {
    mutation
    node {
      id
      html
      text
      user {
        id
        fullName
      }
      parentId
      replyCount
      createdAt
    }
    updatedFields
  }
}
```

### Parameters

| Argument                    | Type               | Required | Description                                                                                                                                               |
| --------------------------- | ------------------ | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `category`                  | `CommentCategory!` | Yes      | Which kind of parent the comments belong to. See the enum below.                                                                                          |
| `categoryId`                | `String!`          | Yes      | The ID of the parent — a `Chat` ID, a `StatusUpdate` ID, or a `Record` ID, matching `category`.                                                     |
| `showOnlyMentionedComments` | `Boolean`          | No       | When `true`, deliver only comments where you are the mentioner or the mentionee. The `category`/`categoryId` filter is then ignored. Defaults to `false`. |

#### CommentCategory

| Value           | Parent type    | `categoryId` is the ID of |
| --------------- | -------------- | ------------------------- |
| `DISCUSSION`    | `Chat`   | A chat thread       |
| `STATUS_UPDATE` | `StatusUpdate` | A status update           |
| `TODO`          | `Record`       | A record                  |

### Response

```json
{
  "data": {
    "subscribeToComment": {
      "mutation": "CREATED",
      "node": {
        "id": "clm4n8qwx000008l0g4oxdqn7",
        "html": "<p>Looks good — shipping it.</p>",
        "text": "Looks good — shipping it.",
        "user": { "id": "clm4n8qwx000108l0a1b2c3d4", "fullName": "Ada Lovelace" },
        "parentId": null,
        "replyCount": 0,
        "createdAt": "2026-05-29T14:02:11.000Z"
      },
      "updatedFields": []
    }
  }
}
```

#### Returns — CommentSubscriptionPayload

| Field            | Type                    | Description                                                       |
| ---------------- | ----------------------- | ----------------------------------------------------------------- |
| `mutation`       | `MutationType!`         | `CREATED`, `UPDATED`, or `DELETED`.                               |
| `node`           | `Comment`               | The comment after the change. `null` for some `DELETED` events.   |
| `previousValues` | `CommentPreviousValues` | The comment's prior state on an `UPDATED`/`DELETED` event.        |
| `updatedFields`  | `[String!]`             | Names of the `Comment` fields that changed on an `UPDATED` event. |

A `Comment` exposes `id`, `html`, `text`, `category`, `user`, `parent`/`parentId`, `replies`, `replyCount`, `reactions`, `isRead`, `isSeen`, `createdAt`, and `updatedAt`, plus the `chat`/`statusUpdate`/`todo` it belongs to.

### Mentions-only mode

Set `showOnlyMentionedComments: true` to receive comments across any thread where you were @-mentioned (or where you authored the mention), regardless of `category`/`categoryId`:

```graphql
subscription OnMyMentionedComments {
  subscribeToComment(category: TODO, categoryId: "todo_123", showOnlyMentionedComments: true) {
    mutation
    node {
      id
      text
      todo {
        id
        title
      }
    }
  }
}
```

`category` and `categoryId` are still required by the schema, but the resolver bypasses them in this mode — pass any valid values.

### Errors

| Code              | When                                                  |
| ----------------- | ----------------------------------------------------- |
| `UNAUTHENTICATED` | The WebSocket handshake carried no valid credentials. |

### Permissions

Subscribing always succeeds. You only receive events for comments you can see — the resolver matches each published comment against your `category`/`categoryId` filter (or your mentions) before delivering it.

---

## subscribeToCommentTyping

Streams typing indicators for one thread: each event is the `User` who is currently typing into that chat, status update, or record. This is how the app renders "Ada is typing…" beneath a comment box.

The producer side is the `commentTyping` mutation, which a client calls (debounced) while its user types. That publish fans out to everyone subscribed to the same thread.

### Request

```graphql
subscription OnChatTyping {
  subscribeToCommentTyping(id: "chat_123", name: DISCUSSION) {
    id
    fullName
  }
}
```

### Parameters

| Argument | Type                            | Required | Description                                                                      |
| -------- | ------------------------------- | -------- | -------------------------------------------------------------------------------- |
| `id`     | `String!`                       | Yes      | The thread ID — a `Chat`, `StatusUpdate`, or `Record` ID, matching `name`. |
| `name`   | `CommentTypingSubscriptionName` | No       | Which thread surface to watch. Pair it with the matching `id`.                   |

#### CommentTypingSubscriptionName

| Value           | `id` is the ID of   |
| --------------- | ------------------- |
| `DISCUSSION`    | A chat thread |
| `STATUS_UPDATE` | A status update     |
| `TODO`          | A record            |

### Response

This subscription returns a bare `User!`, not a change-feed payload — there is no `mutation` or `node` wrapper. Each event is the typing user.

```json
{
  "data": {
    "subscribeToCommentTyping": {
      "id": "clm4n8qwx000208l0e5f6g7h8",
      "fullName": "Grace Hopper"
    }
  }
}
```

### Permissions

Events are keyed strictly by thread (`name` + `id`), so you receive typing pings only for threads you actively subscribe to. Subscribe to the same thread you are reading comments from.

---

## subscribeToChat

Streams chat lifecycle changes (created, edited, deleted) within a workspace. Use it to keep a chat list live as threads are opened and renamed.

### Request

```graphql
subscription OnChats {
  subscribeToChat(projectId: "workspace_123") {
    mutation
    node {
      id
      title
      html
      commentCount
      user {
        id
        fullName
      }
      updatedAt
    }
    updatedFields
  }
}
```

### Parameters

| Argument    | Type      | Required | Description                                        |
| ----------- | --------- | -------- | -------------------------------------------------- |
| `projectId` | `String!` | Yes      | The workspace whose chats you want to watch. |

### Response

```json
{
  "data": {
    "subscribeToChat": {
      "mutation": "UPDATED",
      "node": {
        "id": "clm4n8qwx000308l0i9j0k1l2",
        "title": "Q3 launch checklist",
        "html": "<p>Final review before we ship.</p>",
        "commentCount": 4,
        "user": { "id": "clm4n8qwx000108l0a1b2c3d4", "fullName": "Ada Lovelace" },
        "updatedAt": "2026-05-29T14:05:40.000Z"
      },
      "updatedFields": ["title"]
    }
  }
}
```

#### Returns — ChatSubscriptionPayload

| Field            | Type                       | Description                            |
| ---------------- | -------------------------- | -------------------------------------- |
| `mutation`       | `MutationType!`            | `CREATED`, `UPDATED`, or `DELETED`.    |
| `node`           | `Chat`               | The chat after the change.       |
| `previousValues` | `ChatPreviousValues` | Prior state on `UPDATED`/`DELETED`.    |
| `updatedFields`  | `[String!]`                | Field names that changed on `UPDATED`. |

### Errors

| Code              | When                                                  |
| ----------------- | ----------------------------------------------------- |
| `UNAUTHENTICATED` | The WebSocket handshake carried no valid credentials. |

### Permissions

You only receive events for the workspace if you are a member of it. Events for workspaces you do not belong to are filtered out before delivery.

---

## subscribeToStatusUpdate

Streams status update changes within a workspace — the posts that appear in a workspace's status feed.

### Request

```graphql
subscription OnStatusUpdates {
  subscribeToStatusUpdate(projectId: "workspace_123") {
    mutation
    node {
      id
      html
      text
      date
      category
      user {
        id
        fullName
      }
    }
    updatedFields
  }
}
```

### Parameters

| Argument    | Type      | Required | Description                                           |
| ----------- | --------- | -------- | ----------------------------------------------------- |
| `projectId` | `String!` | Yes      | The workspace whose status updates you want to watch. |

### Response

```json
{
  "data": {
    "subscribeToStatusUpdate": {
      "mutation": "CREATED",
      "node": {
        "id": "clm4n8qwx000408l0m3n4o5p6",
        "html": "<p>On track for Friday.</p>",
        "text": "On track for Friday.",
        "date": "2026-05-29T00:00:00.000Z",
        "category": "GREEN",
        "user": { "id": "clm4n8qwx000108l0a1b2c3d4", "fullName": "Ada Lovelace" }
      },
      "updatedFields": []
    }
  }
}
```

#### Returns — StatusUpdateSubscriptionPayload

| Field            | Type                         | Description                            |
| ---------------- | ---------------------------- | -------------------------------------- |
| `mutation`       | `MutationType!`              | `CREATED`, `UPDATED`, or `DELETED`.    |
| `node`           | `StatusUpdate`               | The status update after the change.    |
| `previousValues` | `StatusUpdatePreviousValues` | Prior state on `UPDATED`/`DELETED`.    |
| `updatedFields`  | `[String!]`                  | Field names that changed on `UPDATED`. |

### Errors

| Code              | When                                                  |
| ----------------- | ----------------------------------------------------- |
| `UNAUTHENTICATED` | The WebSocket handshake carried no valid credentials. |

### Permissions

Delivered only for workspaces you belong to: the resolver matches the published workspace ID against `projectId` and confirms you are a member before forwarding the event.

---

## subscribeToDocument

Streams collaborative document changes (`Document` objects) within a workspace — created, edited, deleted. Optionally narrow to wiki documents only.

### Request

```graphql
subscription OnDocuments {
  subscribeToDocument(input: { projectId: "workspace_123" }) {
    mutation
    node {
      id
      title
      wiki
      createdBy {
        id
        fullName
      }
      updatedAt
    }
    updatedFields
  }
}
```

### Parameters

#### SubscribeToDocumentInput

| Parameter   | Type      | Required | Description                                                                                                                                         |
| ----------- | --------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| `projectId` | `String!` | Yes      | The workspace whose documents you want to watch.                                                                                                    |
| `wiki`      | `Boolean` | No       | When set, deliver only documents whose `wiki` flag matches this value — `true` for wiki pages, `false` for regular documents. Omit to receive both. |

### Response

```json
{
  "data": {
    "subscribeToDocument": {
      "mutation": "UPDATED",
      "node": {
        "id": "clm4n8qwx000608l0u1v2w3x4",
        "title": "Onboarding runbook",
        "wiki": true,
        "createdBy": { "id": "clm4n8qwx000108l0a1b2c3d4", "fullName": "Ada Lovelace" },
        "updatedAt": "2026-05-29T14:11:27.000Z"
      },
      "updatedFields": ["title"]
    }
  }
}
```

#### Returns — DocumentSubscriptionPayload

| Field            | Type                     | Description                                                                                                                                 |
| ---------------- | ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `mutation`       | `MutationType!`          | `CREATED`, `UPDATED`, or `DELETED`.                                                                                                         |
| `node`           | `Document`               | The document after the change. Exposes `id`, `title`, `content`, `contentBase64`, `wiki`, `project`, `createdBy`, `createdAt`, `updatedAt`. |
| `previousValues` | `DocumentPreviousValues` | Prior state on `UPDATED`/`DELETED`.                                                                                                         |
| `updatedFields`  | `[String!]`              | Field names that changed on `UPDATED`.                                                                                                      |

<Callout variant="info" title="Document metadata vs. live editing">

`subscribeToDocument` reports document-level changes (title, wiki flag, create/delete). It is not the character-by-character editing channel — live collaborative editing runs over a separate Hocuspocus connection at `/collab`, not the GraphQL WebSocket.

</Callout>

### Errors

| Code              | When                                                  |
| ----------------- | ----------------------------------------------------- |
| `UNAUTHENTICATED` | The WebSocket handshake carried no valid credentials. |

### Permissions

Delivered only for workspaces you belong to. When `wiki` is supplied, only documents matching that flag are forwarded; otherwise both wiki and non-wiki documents stream through.

---

## Related

- [Real-time (Subscriptions)](/api/realtime) — the shared change-feed payload shape and the full subscription map.
- [Connect & Authenticate](/api/realtime/connect-and-authenticate) — open an authenticated `graphql-ws` connection.
- [Add a comment](/api/records/add-comment) — the mutation that produces the comment events you subscribe to.
- [Comments & Chats](/api/comments) — query and manage comment threads.
- [Status Updates](/api/status-updates) — query and post status updates.
- [Documents](/api/documents) — create and manage documents.
- [Activity, Mentions & Notifications](/api/realtime/activity-notifications) — the @-mention firehose (`subscribeToMyMention`).
