---
title: Query chats
description: Fetch chats three ways - one by ID, an offset-paginated org-wide list, or a cursor-paginated list scoped to a single workspace you belong to.
icon: Search
order: 4
---

Chats are standalone, workspace-scoped conversation threads (the `Chat` type). There are three read paths, and which one you reach for depends on what you have and what you need:

- **`chat(id)`** — fetch a single thread by ID.
- **`chats(filter, sort, skip, take)`** — the offset-paginated list across the current organization, filtered to one workspace and ordered by `ChatSort` (recently updated or recently commented).
- **`chatList(projectId, ...)`** — the cursor-paginated variant scoped to one workspace, which returns only threads in workspaces you are a member of.

All three return the thread's metadata (title, body, members, `commentCount`, read state) but **not** its replies. The messages inside a chat are `Comment` rows with `category: DISCUSSION`; load them separately with [`commentList`](/api/comments/query-comments). Use `chat(id)` to render a single thread, `chats` when you want simple page-number paging, and `chatList` when you want to walk a workspace's threads with an opaque cursor.

> **Legacy names.** `Chat` was called `Discussion`, and these three queries were `discussion` / `discussions` / `discussionList`. Every legacy name still works and returns identical data — it is marked `@deprecated` in introspection and points at its modern replacement. `CommentCategory.DISCUSSION` is **not** renamed: it is a stored enum value on millions of comment rows.

## `chat(id)`

Use `chat` to fetch one thread by its ID.

### Request

```graphql
query GetChat {
  chat(id: "chat_123") {
    id
    title
    text
    commentCount
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

| Parameter | Type      | Required | Description                 |
| --------- | --------- | -------- | --------------------------- |
| `id`      | `String!` | Yes      | The ID of the chat to load. |

### Response

```json
{
  "data": {
    "chat": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Q3 launch retro",
      "text": "What worked, what didn't, and what we change next quarter.",
      "commentCount": 14,
      "createdAt": "2026-05-20T09:12:44.000Z",
      "user": {
        "id": "clm4n8qwx000108l0a1b2c3d4",
        "fullName": "Dana Okafor"
      },
      "project": {
        "id": "clm4n8qwx000208l0e5f6g7h8",
        "name": "Marketing"
      }
    }
  }
}
```

Returns a single [`Chat`](#the-chat-type). The query throws `CHAT_NOT_FOUND` when no thread matches the ID.

## `chats`

Use `chats` for the offset-paginated list. It is filtered to a single workspace (via the required `ChatFilterInput`) but always scoped to the calling organization, so a `projectId` from another organization returns no results. Results default to most-recently-updated first.

### Request

```graphql
query ListChats {
  chats(filter: { projectId: "workspace_123" }, sort: [lastCommentedAt_DESC], skip: 0, take: 20) {
    items {
      id
      title
      commentCount
      updatedAt
    }
    pageInfo {
      totalItems
      totalPages
      page
      perPage
      hasNextPage
      hasPreviousPage
    }
  }
}
```

### Parameters

| Parameter | Type               | Required | Default            | Description                                                                 |
| --------- | ------------------ | -------- | ------------------ | --------------------------------------------------------------------------- |
| `filter`  | `ChatFilterInput!` | Yes      | —                  | Which workspace's chats to return. See [`ChatFilterInput`](#chatfilterinput). |
| `sort`    | `[ChatSort!]`      | No       | `[updatedAt_DESC]` | Sort order. See [`ChatSort`](#chatsort).                                    |
| `skip`    | `Int`              | No       | `0`                | Number of chats to skip from the start of the result set.                   |
| `take`    | `Int`              | No       | `20`               | Maximum number of chats to return on this page.                             |

#### `ChatFilterInput`

| Field       | Type      | Required | Description                                |
| ----------- | --------- | -------- | ------------------------------------------ |
| `projectId` | `String!` | Yes      | The ID of the workspace whose chats to list. |

#### `ChatSort`

`sort` takes a list of `ChatSort` values. Pass a single value for a single ordering.

| Value                  | Orders by                                         |
| ---------------------- | ------------------------------------------------- |
| `updatedAt_DESC`       | Most recently updated thread first (the default). |
| `updatedAt_ASC`        | Least recently updated thread first.              |
| `lastCommentedAt_DESC` | Thread with the most recent reply first.          |
| `lastCommentedAt_ASC`  | Thread with the oldest most-recent reply first.   |

### Response

`chats` returns a `ChatPagination` — the threads on this page in `items`, plus offset pagination metadata in `pageInfo`.

```json
{
  "data": {
    "chats": {
      "items": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "title": "Q3 launch retro",
          "commentCount": 14,
          "updatedAt": "2026-05-28T16:40:02.000Z"
        },
        {
          "id": "clm4n8qwx000308l0i9j0k1l2",
          "title": "Brand refresh kickoff",
          "commentCount": 3,
          "updatedAt": "2026-05-27T11:05:18.000Z"
        }
      ],
      "pageInfo": {
        "totalItems": 42,
        "totalPages": 3,
        "page": 1,
        "perPage": 20,
        "hasNextPage": true,
        "hasPreviousPage": false
      }
    }
  }
}
```

#### `ChatPagination`

| Field      | Type        | Description                                    |
| ---------- | ----------- | ---------------------------------------------- |
| `items`    | `[Chat!]!`  | The chats on this page.                        |
| `pageInfo` | `PageInfo!` | Offset pagination metadata for the result set. |

To page through results, advance `skip` by `take` (for example `skip: 20, take: 20` for the second page) until `pageInfo.hasNextPage` is `false`.

## `chatList`

Use `chatList` for cursor-paginated access to one workspace's threads. Unlike `chats`, it returns **only threads in a workspace the caller is a member of** — if you are not a member of `projectId`, the list comes back empty. Step through pages with the `after` cursor rather than a `skip` offset.

### Request

```graphql
query ListWorkspaceChats {
  chatList(projectId: "workspace_123", first: 20, orderBy: updatedAt_DESC) {
    discussions {
      id
      title
      commentCount
      updatedAt
    }
    totalCount
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
```

### Parameters

| Parameter   | Type               | Required | Default | Description                                                                              |
| ----------- | ------------------ | -------- | ------- | ---------------------------------------------------------------------------------------- |
| `projectId` | `String!`          | Yes      | —       | The workspace whose chats to list. Only returns threads if the caller is a workspace member. |
| `first`     | `Int`              | No       | `20`    | Maximum number of chats to return on this page.                                          |
| `after`     | `String`           | No       | —       | Cursor to page forward from — pass the previous page's `pageInfo.endCursor`.             |
| `before`    | `String`           | No       | —       | Cursor to page backward from.                                                            |
| `last`      | `Int`              | No       | —       | Maximum number of chats to return when paging backward.                                  |
| `skip`      | `Int`              | No       | `0`     | Number of chats to skip. Reported in `pageInfo`; prefer `after` for forward paging.      |
| `orderBy`   | `ChatOrderByInput` | No       | —       | Sort order. See [`ChatOrderByInput`](#chatorderbyinput).                                 |

#### `ChatOrderByInput`

A single enum value of the form `<field>_<direction>`. The fields and directions available:

| Field       | Ascending       | Descending       |
| ----------- | --------------- | ---------------- |
| `id`        | `id_ASC`        | `id_DESC`        |
| `uid`       | `uid_ASC`       | `uid_DESC`       |
| `title`     | `title_ASC`     | `title_DESC`     |
| `html`      | `html_ASC`      | `html_DESC`      |
| `text`      | `text_ASC`      | `text_DESC`      |
| `createdAt` | `createdAt_ASC` | `createdAt_DESC` |
| `updatedAt` | `updatedAt_ASC` | `updatedAt_DESC` |

<Callout variant="info" title="Same orderBy type as commentList">

`chatList` and [`commentList`](/api/comments/query-comments) share the `ChatOrderByInput` type for their `orderBy` argument — the name is shared even though `commentList` orders comments. Its legacy spelling is `DiscussionOrderByInput`. Sort by `updatedAt_DESC` or `createdAt_DESC` in both.

</Callout>

### Response

`chatList` returns a `ChatList` — the threads in `discussions`, a `totalCount` of all matching threads, and a `pageInfo` carrying the `endCursor` for the next page.

```json
{
  "data": {
    "chatList": {
      "discussions": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "title": "Q3 launch retro",
          "commentCount": 14,
          "updatedAt": "2026-05-28T16:40:02.000Z"
        }
      ],
      "totalCount": 42,
      "pageInfo": {
        "hasNextPage": true,
        "endCursor": "clm4n8qwx000008l0g4oxdqn7"
      }
    }
  }
}
```

#### `ChatList`

| Field        | Type        | Description                                                                    |
| ------------ | ----------- | ------------------------------------------------------------------------------ |
| `discussions` | `[Chat!]!` | The chats on this page. Keeps its legacy name — wrapper types have no resolver to delegate a field rename through. |
| `totalCount` | `Int!`      | Total number of chats in the workspace the caller can see.                     |
| `pageInfo`   | `PageInfo!` | Cursor pagination metadata. `endCursor` is the ID of the last thread returned. |

To fetch the next page, pass the current page's `pageInfo.endCursor` as the next call's `after`, and stop when `pageInfo.hasNextPage` is `false`.

## The `Chat` type

All three queries return the same `Chat` shape. Replies are not included here — load them with [`commentList`](/api/comments/query-comments).

| Field          | Type             | Description                                                                                                                                                                                                             |
| -------------- | ---------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `id`           | `ID!`            | Unique identifier for the chat.                                                                                                                                                                                         |
| `title`        | `String!`        | The thread title.                                                                                                                                                                                                       |
| `description`  | `String`         | Optional short description.                                                                                                                                                                                             |
| `html`         | `String!`        | The opening body as sanitized HTML.                                                                                                                                                                                     |
| `text`         | `String!`        | The opening body as plain text.                                                                                                                                                                                         |
| `kind`         | `ChatKind!`      | `CHANNEL` for the workspace threads this page's `chats`/`chatList` return. `chat(id)` can also return a `DM` or `GROUP` conversation — see [Direct Messages & Group Chats](/api/inbox/direct-messages-and-group-chats). |
| `createdAt`    | `DateTime!`      | When the chat was created.                                                                                                                                                                                              |
| `updatedAt`    | `DateTime!`      | When the chat was last updated.                                                                                                                                                                                         |
| `user`         | `User!`          | The thread's creator. Select `id`, `fullName`, `email`.                                                                                                                                                                 |
| `people`       | `[User!]`        | Members of the chat.                                                                                                                                                                                                    |
| `project`      | `Workspace`      | The workspace the chat belongs to. `null` for the `DM`/`GROUP` kinds, which have no workspace.                                                                                                                          |
| `members`      | `[ChatMember!]!` | Membership rows for `DM`/`GROUP` conversations. Empty for `CHANNEL`.                                                                                                                                                    |
| `commentCount` | `Int!`           | Number of replies in the thread.                                                                                                                                                                                        |
| `isRead`       | `Boolean`        | Whether the calling user has read this chat. Evaluated per caller.                                                                                                                                                      |
| `isSeen`       | `Boolean`        | Whether the calling user has seen this chat. Evaluated per caller.                                                                                                                                                      |

<Callout variant="warning" title="Chat.comments is deprecated">

The `Chat.comments` field is deprecated. Load a thread's replies with [`commentList`](/api/comments/query-comments) using `category: DISCUSSION` and `categoryId` set to the chat's ID — it returns top-level comments with cursor and offset pagination.

</Callout>

## Loading a thread's replies

A chat query gives you the thread; to render the conversation, follow up with `commentList` scoped to that chat. `commentList` returns only top-level comments — read each comment's `replies` for the threaded answers.

```graphql
query ChatWithComments {
  chat(id: "chat_123") {
    id
    title
    commentCount
  }
  commentList(category: DISCUSSION, categoryId: "chat_123", first: 20) {
    comments {
      id
      text
      user {
        id
        fullName
      }
      replyCount
    }
    totalCount
  }
}
```

## Errors

| Code              | When                                                                                                                                 |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------ |
| `CHAT_NOT_FOUND`  | `chat(id)` — no chat matches the given ID.                                                                                           |
| `FORBIDDEN`       | `chat(id)` — the caller can't view this thread (not a member of its workspace for `CHANNEL`, or no membership row for `DM`/`GROUP`). |
| `UNAUTHENTICATED` | The request carries no valid authentication.                                                                                         |

## Permissions

- **`chat(id)`** is not scoped by organization or workspace at the shield layer — it works for a plain client/DM-only user with no resolvable workspace. Access is resolved per thread `kind`: `CHANNEL` requires workspace membership, `DM`/`GROUP` requires a `ChatMember` row (including a removed one — see [Direct Messages & Group Chats](/api/inbox/direct-messages-and-group-chats)). Either case failing throws `FORBIDDEN`.
- **`chats`** returns threads only for the calling organization. A `projectId` belonging to another organization yields an empty `items` list rather than an error.
- **`chatList`** returns threads only for projects the caller is a member of. Passing a `projectId` you do not belong to returns an empty `discussions` list, not an error.

All three calls authenticate with the standard token headers (`blue-token-id`, `blue-token-secret`, `blue-org-id`). The organization and project arguments accept an ID or a slug. `chats` and `chatList` only ever return `CHANNEL` threads, since their filters require a `projectId`; only `chat(id)` can return a `DM`/`GROUP` conversation.

## Related

- [Query comments](/api/comments/query-comments)
- [Create, update & delete chats](/api/comments/manage-discussions)
- [Comments overview](/api/comments)
- [Reactions](/api/comments/reactions)
- [Comment & chat subscriptions](/api/realtime/comment-discussion-subscriptions)
- [Direct Messages & Group Chats](/api/inbox/direct-messages-and-group-chats) — the `DM`/`GROUP` kinds `chat(id)` can also return.
