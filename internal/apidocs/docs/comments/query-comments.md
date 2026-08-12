---
title: Query comments
description: Fetch comments for any record, chat, or status update with the commentList query — scoped by category, paginated, and ordered.
icon: ListFilter
order: 2
---

Use the `commentList` query to read comments from any surface in Blue. There is one comment store across the product, so a single query serves all three targets: pass a `category` (`DISCUSSION`, `STATUS_UPDATE`, or `TODO`) plus the `categoryId` of the thing you're reading. The result is a `CommentList` with the page of comments, pagination metadata, and a `totalCount`.

`commentList` is the canonical way to fetch comments. The older `Chat.comments`, `StatusUpdate.comments`, and `Todo.comments` fields are deprecated in favor of it — see [Comments overview](/api/comments) for the comment model and the `category`/`categoryId` targeting convention.

## Request

Fetch the top-level comments on a record. `category` is `TODO` and `categoryId` is the record (`Todo`) ID.

```graphql
query RecordComments {
  commentList(category: TODO, categoryId: "todo_123") {
    comments {
      id
      text
      createdAt
      user {
        id
        fullName
      }
    }
    totalCount
  }
}
```

The same query shape reads a chat (`category: DISCUSSION`, `categoryId` = the chat ID) or a status update (`category: STATUS_UPDATE`, `categoryId` = the status-update ID). Only `category` and `categoryId` change.

<Callout variant="info" title="Top-level comments only">

`commentList` returns only the root comments of a thread — those with `parentId = null`. Threaded replies are not included in the page. Fetch a comment's replies through its `replies` field (and `replyCount`); see [Comments overview](/api/comments).

</Callout>

## Parameters

### Arguments

| Argument     | Type                     | Required | Description                                                                                         |
| ------------ | ------------------------ | -------- | --------------------------------------------------------------------------------------------------- |
| `category`   | `CommentCategory!`       | Yes      | Which kind of target you're reading. Pairs with `categoryId`.                                       |
| `categoryId` | `String!`                | Yes      | The ID of the target: a chat ID, a status-update ID, or a record (`Todo`) ID, per `category`. |
| `first`      | `Int`                    | No       | Page size — how many comments to return. Omit to return all matching comments (subject to `skip`).  |
| `skip`       | `Int`                    | No       | Number of comments to skip before the page, for offset pagination. Defaults to `0`.                 |
| `after`      | `String`                 | No       | Cursor: return comments whose `id` is greater than or equal to this value. Use a comment `id`.      |
| `orderBy`    | `ChatOrderByInput` | No       | Sort order. Defaults to `createdAt_ASC` (oldest first).                                             |
| `before`     | `String`                 | No       | Accepted by the schema but not applied by this query. Paginate with `first` + `skip` (or `after`).  |
| `last`       | `Int`                    | No       | Accepted by the schema but not applied by this query. Use `first` for the page size.                |

<Callout variant="warning" title="before and last are no-ops here">

The schema accepts `before` and `last`, but `commentList` ignores them. Page with `first` + `skip` for offset pagination, or `after` for an id-cursor scan. Sending `before`/`last` has no effect on the result.

</Callout>

### CommentCategory

| Value           | `categoryId` is the ID of | Read instead via                                     |
| --------------- | ------------------------- | ---------------------------------------------------- |
| `DISCUSSION`    | A chat thread       | [Query chats](/api/comments/query-discussions) |
| `STATUS_UPDATE` | A status update           | The status update's ID                               |
| `TODO`          | A record (`Todo`)         | [List records](/api/records/list-records)            |

### ChatOrderByInput

`orderBy` takes a single `ChatOrderByInput` value (the type is shared with `chatList`). For comments the meaningful sorts are by timestamp:

| Value                              | Sorts comments by                          |
| ---------------------------------- | ------------------------------------------ |
| `createdAt_ASC` / `createdAt_DESC` | Creation time. `createdAt_ASC` is default. |
| `updatedAt_ASC` / `updatedAt_DESC` | Last-edited time.                          |
| `id_ASC` / `id_DESC`               | Comment ID.                                |

## Response

```json
{
  "data": {
    "commentList": {
      "comments": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "text": "Kicking off the launch checklist — see the linked doc.",
          "createdAt": "2026-05-20T09:14:00.000Z",
          "user": {
            "id": "clm4n8qwx000108l0haz1f2pe",
            "fullName": "Ada Lovelace"
          }
        },
        {
          "id": "clm4n8qwx000208l0bk7c9xqz",
          "text": "Pricing copy is approved.",
          "createdAt": "2026-05-21T16:02:00.000Z",
          "user": {
            "id": "clm4n8qwx000308l0e1g5m4ab",
            "fullName": "Grace Hopper"
          }
        }
      ],
      "totalCount": 2
    }
  }
}
```

### CommentList

| Field        | Type          | Description                                               |
| ------------ | ------------- | --------------------------------------------------------- |
| `comments`   | `[Comment!]!` | The top-level comments on this page, in `orderBy` order.  |
| `pageInfo`   | `PageInfo!`   | Pagination metadata for the page (below).                 |
| `totalCount` | `Int!`        | Total top-level comments on the target, across all pages. |

### PageInfo

| Field             | Type       | Description                                                               |
| ----------------- | ---------- | ------------------------------------------------------------------------- |
| `totalItems`      | `Int`      | Total matching comments, across all pages (mirrors `totalCount`).         |
| `totalPages`      | `Int`      | Total pages at the current `first`.                                       |
| `page`            | `Int`      | Current page, derived from `skip` and `first`.                            |
| `perPage`         | `Int`      | Page size in effect (the resolved `first`).                               |
| `hasNextPage`     | `Boolean!` | Whether a next page exists.                                               |
| `hasPreviousPage` | `Boolean!` | Whether a previous page exists.                                           |
| `startCursor`     | `String`   | Deprecated. Paginate with `first`/`skip` (or `after`), not these cursors. |
| `endCursor`       | `String`   | Deprecated. See `startCursor`.                                            |

### Comment

The fields you'll select most often. The `Comment` type exposes more — these are the common ones; the full set is documented in [Comments overview](/api/comments).

| Field          | Type                | Description                                                                          |
| -------------- | ------------------- | ------------------------------------------------------------------------------------ |
| `id`           | `ID!`               | Unique identifier.                                                                   |
| `uid`          | `String!`           | Short, human-facing identifier.                                                      |
| `html`         | `String!`           | Comment body as HTML.                                                                |
| `text`         | `String!`           | Comment body as plain text.                                                          |
| `category`     | `CommentCategory!`  | The target kind this comment belongs to.                                             |
| `createdAt`    | `DateTime!`         | Creation timestamp.                                                                  |
| `updatedAt`    | `DateTime!`         | Last-edited timestamp.                                                               |
| `deletedAt`    | `DateTime`          | Set when the comment was soft-deleted; the body is blanked but the row remains.      |
| `deletedBy`    | `User`              | Who soft-deleted the comment. Select `fullName`/`email`, not `name`.                 |
| `user`         | `User!`             | The author. Select `fullName`/`email`, not `name`.                                   |
| `parentId`     | `String`            | The parent comment's ID when this is a reply. Always `null` for items returned here. |
| `parent`       | `Comment`           | The parent comment, when this is a reply.                                            |
| `replies`      | `[Comment!]!`       | Threaded replies to this comment.                                                    |
| `replyCount`   | `Int!`              | Number of replies.                                                                   |
| `isRead`       | `Boolean`           | Read state for the current user.                                                     |
| `isSeen`       | `Boolean`           | Seen state for the current user.                                                     |
| `reactions`    | `[ReactionGroup!]!` | Emoji reactions on the comment. See [Reactions](/api/comments/reactions).            |
| `chat`   | `Chat`        | The parent chat, when `category` is `DISCUSSION`.                              |
| `statusUpdate` | `StatusUpdate`      | The parent status update, when `category` is `STATUS_UPDATE`.                        |
| `todo`         | `Todo`              | The parent record, when `category` is `TODO`.                                        |

## Full example

Read the second page of a chat's comments (10 per page, newest first), with reply counts and reactions.

```graphql
query ChatComments {
  commentList(
    category: DISCUSSION
    categoryId: "chat_123"
    first: 10
    skip: 10
    orderBy: createdAt_DESC
  ) {
    comments {
      id
      uid
      html
      text
      createdAt
      updatedAt
      replyCount
      isRead
      user {
        id
        fullName
        email
      }
      reactions {
        emoji
        count
        reactedByMe
      }
    }
    pageInfo {
      totalItems
      totalPages
      page
      perPage
      hasNextPage
      hasPreviousPage
    }
    totalCount
  }
}
```

## Errors

This query degrades gracefully: a `categoryId` that doesn't exist (or points at a target you can't see) yields an empty `comments` list and a `totalCount` of `0`, rather than an error. Errors are reserved for authentication.

| Code              | When                                      |
| ----------------- | ----------------------------------------- |
| `UNAUTHENTICATED` | The request carries no valid credentials. |

## Permissions

Any authenticated caller can run `commentList`. Visibility is then scoped to the caller:

- A role configured with **show-only-mentioned-comments** sees a filtered subset — only comments the user authored or in which the user is mentioned. `totalCount` reflects that filtered set, not the thread's full comment count.
- Comments on targets the caller can't access return nothing rather than erroring.

## Related

- [Comments overview](/api/comments)
- [Create, edit & delete comments](/api/comments/manage-comments)
- [Reactions](/api/comments/reactions)
- [Query chats](/api/comments/query-discussions)
- [List records](/api/records/list-records)
