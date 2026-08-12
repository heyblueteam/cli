---
title: Create, update & delete chats
description: Open workspace-scoped chat threads, rename them, and hard-delete them with the createChat, updateChat, and deleteChat mutations.
icon: MessagesSquare
order: 3
---

A chat is a standalone, workspace-scoped conversation thread — a place for a topic that isn't tied to a single record. Chats are `Chat` objects in the API, and they live in a workspace. Once a thread exists, replies are posted as comments with `category: DISCUSSION` (see [Create, edit & delete comments](/api/comments/manage-comments)).

- Use `createChat` to seed a new thread with a title and an opening body.
- Use `updateChat` to rename a thread — it edits the **title only**; the body is fixed once created.
- Use `deleteChat` to remove a thread. Unlike a comment delete, this is a hard delete.

To read chats back, use the [chat query paths](/api/comments/query-discussions). The `Chat.comments` field is deprecated — load replies with [`commentList(category: DISCUSSION)`](/api/comments/query-comments) instead.

> **Legacy names.** `Chat` was called `Discussion`, and these three mutations were `createDiscussion` / `updateDiscussion` / `deleteDiscussion`. Every legacy name still works and returns identical data — it is marked `@deprecated` in introspection and points at its modern replacement. `CommentCategory.DISCUSSION` is **not** renamed: it is a stored enum value on millions of comment rows.

## createChat

Create a chat in a workspace. The opening message is supplied as both `html` (rich content, sanitized server-side) and `text` (the plain-text fallback). The mutation returns the created `Chat`.

### Request

```graphql
mutation CreateChat {
  createChat(
    input: {
      projectId: "workspace_123"
      title: "Q3 launch retro"
      html: "<p>What went well, what didn't?</p>"
      text: "What went well, what didn't?"
    }
  ) {
    id
    title
    html
    text
    createdAt
    user {
      id
      fullName
    }
    project {
      id
      name
    }
  }
}
```

### Parameters

#### CreateChatInput

| Parameter   | Type      | Required | Description                                                                                                                   |
| ----------- | --------- | -------- | ----------------------------------------------------------------------------------------------------------------------------- |
| `title`     | `String!` | Yes      | The thread title.                                                                                                             |
| `html`      | `String!` | Yes      | The opening message as HTML. Sanitized server-side; embedded `@mentions`, images, and file attachments are extracted from it. |
| `text`      | `String!` | Yes      | The plain-text rendering of the opening message, used in notifications, search, and clients that don't render HTML.           |
| `projectId` | `String!` | Yes      | The workspace the chat belongs to. Accepts a workspace id or slug.                                                            |

The body (`html` / `text`) is set once at creation and is immutable afterward — `updateChat` changes only the title. To revise the opening message, delete the thread and create a new one, or post a follow-up comment.

### Response

```json
{
  "data": {
    "createChat": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Q3 launch retro",
      "html": "<p>What went well, what didn't?</p>",
      "text": "What went well, what didn't?",
      "createdAt": "2026-05-29T14:02:11.000Z",
      "user": {
        "id": "clm4n8a1b000108l0c2d3e4f5",
        "fullName": "Dana Rao"
      },
      "project": {
        "id": "clm4n7zzz000008l0a1b2c3d4",
        "name": "Marketing"
      }
    }
  }
}
```

#### Returns

Returns the created [`Chat`](#chat-type).

### Mentions, images & attachments

`@mentions`, inline images, and file attachments all travel inside the `html` field of the opening message — there are no separate arguments for them. To mention a user, include an anchor whose `href` is `#view-profile-<userId>`; the server rewrites it into a mention and notifies that user. The same convention applies when posting comment replies (see [Create, edit & delete comments](/api/comments/manage-comments)).

### Errors

| Code                  | When                                                                                            |
| --------------------- | ----------------------------------------------------------------------------------------------- |
| `WORKSPACE_NOT_FOUND` | No workspace matches the supplied `projectId` (id or slug).                                     |
| `FORBIDDEN`           | The caller is `VIEW_ONLY` or `COMMENT_ONLY` in the workspace, isn't a member, or it's archived. |
| `UNAUTHENTICATED`     | No valid token was supplied.                                                                    |

## updateChat

Rename a chat. This is the only mutable field: `UpdateChatInput` carries just `id` and an optional `title`. The mutation returns the updated `Chat`.

### Request

```graphql
mutation UpdateChat {
  updateChat(input: { id: "chat_123", title: "Q3 launch retro (final)" }) {
    id
    title
    updatedAt
  }
}
```

### Parameters

#### UpdateChatInput

| Parameter | Type      | Required | Description                                                                                      |
| --------- | --------- | -------- | ------------------------------------------------------------------------------------------------ |
| `id`      | `String!` | Yes      | The id of the chat to rename.                                                                    |
| `title`   | `String`  | No       | The new title. Trimmed server-side; omitting it (or sending blank) clears it to an empty string. |

There is no `editChat` mutation, and there is no way to change `html` / `text`, `projectId`, or any other field through this call — only the title.

### Response

```json
{
  "data": {
    "updateChat": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Q3 launch retro (final)",
      "updatedAt": "2026-05-29T15:18:44.000Z"
    }
  }
}
```

#### Returns

Returns the updated [`Chat`](#chat-type).

### Errors

| Code              | When                                                                                                              |
| ----------------- | ----------------------------------------------------------------------------------------------------------------- |
| `CHAT_NOT_FOUND`  | No chat exists with the given `id`. Callers using the legacy `updateDiscussion` receive `DISCUSSION_NOT_FOUND`.   |
| `FORBIDDEN`       | The caller isn't the creator or a workspace `OWNER`, is `VIEW_ONLY`/`COMMENT_ONLY`, or the workspace is archived. |
| `UNAUTHENTICATED` | No valid token was supplied.                                                                                      |

## deleteChat

Remove a chat thread. This is a **hard delete**: before the row is removed, its reply chains are unlinked (threaded comment replies have their `parentId` cleared to release the foreign-key constraint), and a snapshot of the chat is written to the organization's Trash. The mutation returns a `MutationResult`.

This differs from `deleteComment`, which is a _soft_ delete that keeps the row and blanks its body — see [Create, edit & delete comments](/api/comments/manage-comments).

### Request

`deleteChat` takes the chat id directly (not an input object) and returns a `MutationResult`.

```graphql
mutation DeleteChat {
  deleteChat(id: "chat_123") {
    success
    operationId
  }
}
```

### Parameters

| Parameter | Type      | Required | Description                   |
| --------- | --------- | -------- | ----------------------------- |
| `id`      | `String!` | Yes      | The id of the chat to delete. |

### Response

```json
{
  "data": {
    "deleteChat": {
      "success": true,
      "operationId": "clm4n8qwx000008l0g4oxdqn7"
    }
  }
}
```

#### Returns

| Field         | Type       | Description                                                          |
| ------------- | ---------- | -------------------------------------------------------------------- |
| `success`     | `Boolean!` | `true` when the chat was deleted.                                    |
| `operationId` | `String`   | Identifier for the mutation, useful for correlating realtime events. |

### Errors

| Code              | When                                                                                                              |
| ----------------- | ----------------------------------------------------------------------------------------------------------------- |
| `CHAT_NOT_FOUND`  | No chat exists with the given `id`. Callers using the legacy `deleteDiscussion` receive `DISCUSSION_NOT_FOUND`.   |
| `FORBIDDEN`       | The caller isn't the creator or a workspace `OWNER`, is `VIEW_ONLY`/`COMMENT_ONLY`, or the workspace is archived. |
| `UNAUTHENTICATED` | No valid token was supplied.                                                                                      |

## Chat type

The fields available on a `Chat`. The comment thread itself is loaded separately via [`commentList(category: DISCUSSION)`](/api/comments/query-comments) — the `Chat.comments` field is deprecated and should not be used.

| Field          | Type             | Description                                                                                                                                                                                                |
| -------------- | ---------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `id`           | `ID!`            | The chat id.                                                                                                                                                                                               |
| `title`        | `String!`        | The thread title (the only editable field).                                                                                                                                                                |
| `description`  | `String`         | A short clipped preview of the body (HTML truncated to ~180 characters), useful for list views.                                                                                                            |
| `html`         | `String!`        | The opening message as sanitized HTML. Immutable after creation.                                                                                                                                           |
| `text`         | `String!`        | The opening message as plain text.                                                                                                                                                                         |
| `kind`         | `ChatKind!`      | `CHANNEL` for a thread created here with `createChat`. See [Direct Messages & Group Chats](/api/inbox/direct-messages-and-group-chats) for the `DM`/`GROUP` kinds.                                         |
| `createdAt`    | `DateTime!`      | When the thread was created.                                                                                                                                                                               |
| `updatedAt`    | `DateTime!`      | When the thread was last changed (e.g. renamed).                                                                                                                                                           |
| `user`         | `User!`          | The creator of the thread.                                                                                                                                                                                 |
| `people`       | `[User!]`        | Users participating in the thread (creator plus commenters).                                                                                                                                               |
| `project`      | `Workspace`      | The workspace the chat belongs to. Always set for a `createChat` thread; `null` only for the `DM`/`GROUP` kinds documented on [Direct Messages & Group Chats](/api/inbox/direct-messages-and-group-chats). |
| `members`      | `[ChatMember!]!` | Membership rows for `DM`/`GROUP` conversations. Unused (empty) for a `CHANNEL` thread created here.                                                                                                        |
| `commentCount` | `Int!`           | The number of comments in the thread.                                                                                                                                                                      |
| `isRead`       | `Boolean`        | Whether the calling user has read the thread.                                                                                                                                                              |
| `isSeen`       | `Boolean`        | Whether the calling user has seen (surfaced but not necessarily opened) the thread.                                                                                                                        |

## Permissions

All three mutations require the caller to be a member of the chat's workspace, and they deny `VIEW_ONLY` and `COMMENT_ONLY` access levels. The workspace must be active (not archived).

| Mutation     | Who can call it                                                                         |
| ------------ | --------------------------------------------------------------------------------------- |
| `createChat` | Any member of the workspace above `COMMENT_ONLY` (i.e. not `VIEW_ONLY`/`COMMENT_ONLY`). |
| `updateChat` | The thread's creator, **or** a workspace `OWNER`.                                       |
| `deleteChat` | The thread's creator, **or** a workspace `OWNER`.                                       |

## Related

- [Comments overview](/api/comments) — the `Comment` type, `CommentCategory`, threading, and read state.
- [Query chats](/api/comments/query-discussions) — fetch a single thread, list threads, or page through a workspace's threads.
- [Create, edit & delete comments](/api/comments/manage-comments) — post the replies that hang off a chat.
- [Query comments](/api/comments/query-comments) — load a chat's replies with `commentList`.
- [Comment & chat subscriptions](/api/realtime/comment-discussion-subscriptions) — stream chat and comment changes in real time.
- [Direct Messages & Group Chats](/api/inbox/direct-messages-and-group-chats) — `startDirectMessage`/`createGroupChat` create the other two `Chat` kinds (`DM`/`GROUP`), started differently from the `createChat` flow on this page.
