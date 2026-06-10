---
title: Create, edit & delete comments
description: Post, edit, and soft-delete comments on any record, discussion, or status update with one polymorphic set of mutations.
icon: MessageCircle
order: 1
---

Blue has one `Comment` type and one set of write mutations that work on every surface that supports discussion. You target the surface with `category` + `categoryId`: a record (`category: TODO`), a discussion (`category: DISCUSSION`), or a status update (`category: STATUS_UPDATE`). The `categoryId` is the id of that record, discussion, or status update respectively. Records are `Todo` objects in the API. See the [comments overview](/api/comments) for the full mental model.

- Use `createComment` to post a comment (or a threaded reply via `parentId`).
- Use `editComment` to replace the body of a comment you own.
- Use `deleteComment` to soft-delete a comment — the row stays, its body is blanked, and `deletedAt` / `deletedBy` are stamped.

To read comments back, use the [`commentList` query](/api/comments/query-comments) — not the deprecated `Todo.comments` / `Discussion.comments` / `StatusUpdate.comments` fields.

## createComment

Post a comment on the targeted surface. The body is supplied as both `html` (rich content, sanitized server-side) and `text` (the plain-text fallback). The mutation returns the created `Comment`.

### Request

```graphql
mutation CreateComment {
  createComment(
    input: {
      category: TODO
      categoryId: "todo_123"
      html: "<p>Shipping this today.</p>"
      text: "Shipping this today."
    }
  ) {
    id
    html
    text
    category
    createdAt
    user {
      id
      fullName
    }
  }
}
```

### Parameters

#### CreateCommentInput

| Parameter    | Type               | Required | Description                                                                                                                                         |
| ------------ | ------------------ | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------- |
| `html`       | `String!`          | Yes      | The comment body as HTML. Sanitized server-side; embedded `@mentions`, images, and file attachments are extracted from it.                          |
| `text`       | `String!`          | Yes      | The plain-text rendering of the body, used in notifications, search, and clients that don't render HTML.                                            |
| `category`   | `CommentCategory!` | Yes      | Which surface to comment on: `TODO`, `DISCUSSION`, or `STATUS_UPDATE`.                                                                              |
| `categoryId` | `String!`          | Yes      | The id of the target — a record id (`TODO`), discussion id (`DISCUSSION`), or status update id (`STATUS_UPDATE`).                                   |
| `tiptap`     | `Boolean`          | No       | Set `true` if `html` is Tiptap-formatted; the server then runs the Tiptap sanitizer (which also extracts image/file nodes). Omit for standard HTML. |
| `parentId`   | `String`           | No       | The id of an existing comment on the same target to reply to. Creates a threaded reply. Threads are at most two levels deep.                        |

#### CommentCategory

| Value           | `categoryId` refers to | Target type    |
| --------------- | ---------------------- | -------------- |
| `TODO`          | A record id            | `Todo`         |
| `DISCUSSION`    | A discussion id        | `Discussion`   |
| `STATUS_UPDATE` | A status update id     | `StatusUpdate` |

### Response

```json
{
  "data": {
    "createComment": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "html": "<p>Shipping this today.</p>",
      "text": "Shipping this today.",
      "category": "TODO",
      "createdAt": "2026-05-29T14:02:11.000Z",
      "user": {
        "id": "clm4n8a1b000108l0c2d3e4f5",
        "fullName": "Dana Rao"
      }
    }
  }
}
```

#### Returns

Returns the created [`Comment`](/api/comments). Commonly selected fields:

| Field        | Type               | Description                                                                  |
| ------------ | ------------------ | ---------------------------------------------------------------------------- |
| `id`         | `ID!`              | The comment id.                                                              |
| `html`       | `String!`          | The sanitized HTML body.                                                     |
| `text`       | `String!`          | The plain-text body.                                                         |
| `category`   | `CommentCategory!` | The surface the comment lives on.                                            |
| `createdAt`  | `DateTime!`        | When the comment was posted.                                                 |
| `user`       | `User!`            | The author.                                                                  |
| `parentId`   | `String`           | The id of the comment this is a reply to, or `null` for a top-level comment. |
| `replyCount` | `Int!`             | Number of replies (top-level comments only).                                 |

### Mentions, images & attachments

`@mentions`, inline images, and file attachments all travel inside the `html` field — there are no separate arguments for them.

To mention a user, include an anchor whose `href` is `#view-profile-<userId>`. The server rewrites it into a mention, notifies the mentioned user, and emits a mention event:

```graphql
mutation MentionInComment {
  createComment(
    input: {
      category: DISCUSSION
      categoryId: "discussion_123"
      html: "<p>Can you take a look, <a href=\"#view-profile-user_123\">@Dana</a>?</p>"
      text: "Can you take a look, @Dana?"
    }
  ) {
    id
  }
}
```

Images and attachments embedded in the HTML are parsed out, stored against the comment, and linked back automatically. See [Upload files](/api/files) for obtaining file references.

### Errors

| Code              | When                                                                                                           |
| ----------------- | -------------------------------------------------------------------------------------------------------------- |
| `FORBIDDEN`       | The caller is `VIEW_ONLY` in the target's project, isn't a member of that project, or the project is archived. |
| `UNAUTHENTICATED` | No valid token was supplied.                                                                                   |

<Callout variant="info" title="Permission enforcement by category">

For `category: TODO`, the resolver runs an explicit permission check (the caller must be able to comment on that record). For `DISCUSSION` and `STATUS_UPDATE` the data layer still requires the caller to be a non-`VIEW_ONLY` member of the target's project — but there is no extra resolver-level guard beyond that membership check today.

</Callout>

## editComment

Replace the body of an existing comment. Editing is owner-only: you can edit only the comments you authored. The new `html` is re-sanitized and mentions are re-evaluated.

### Request

```graphql
mutation EditComment {
  editComment(
    input: {
      id: "comment_123"
      html: "<p>Shipping this tomorrow — blocked on review.</p>"
      text: "Shipping this tomorrow — blocked on review."
    }
  ) {
    id
    html
    text
    updatedAt
  }
}
```

### Parameters

#### EditCommentInput

| Parameter | Type      | Required | Description                                                                     |
| --------- | --------- | -------- | ------------------------------------------------------------------------------- |
| `id`      | `String!` | Yes      | The id of the comment to edit.                                                  |
| `html`    | `String!` | Yes      | The replacement HTML body. If it sanitizes to empty, the previous body is kept. |
| `text`    | `String!` | Yes      | The replacement plain-text body.                                                |

You can't change a comment's `category`, `categoryId`, or `parentId` — those are fixed when the comment is created. To move a comment, delete it and post a new one.

### Response

```json
{
  "data": {
    "editComment": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "html": "<p>Shipping this tomorrow — blocked on review.</p>",
      "text": "Shipping this tomorrow — blocked on review.",
      "updatedAt": "2026-05-29T15:18:44.000Z"
    }
  }
}
```

#### Returns

Returns the updated [`Comment`](/api/comments).

### Errors

| Code                | When                                                                               |
| ------------------- | ---------------------------------------------------------------------------------- |
| `COMMENT_NOT_FOUND` | No comment exists with the given `id`.                                             |
| `FORBIDDEN`         | The caller isn't the comment's author, is `VIEW_ONLY`, or the project is archived. |
| `UNAUTHENTICATED`   | No valid token was supplied.                                                       |

## deleteComment

Soft-delete a comment. The row is **not** removed: its `html` and `text` are blanked, `deletedAt` is set to the current time, and `deletedBy` is set to the caller. Any files attached to the comment are deleted. Threaded replies remain in place and still reference the now-emptied parent.

This differs from `deleteDiscussion`, which is a hard delete — see [Manage discussions](/api/comments/manage-discussions).

### Request

`deleteComment` takes the comment id directly (not an input object) and returns a `MutationResult`.

```graphql
mutation DeleteComment {
  deleteComment(id: "comment_123") {
    success
    operationId
  }
}
```

### Parameters

| Parameter | Type      | Required | Description                           |
| --------- | --------- | -------- | ------------------------------------- |
| `id`      | `String!` | Yes      | The id of the comment to soft-delete. |

### Response

```json
{
  "data": {
    "deleteComment": {
      "success": true,
      "operationId": "clm4n8qwx000008l0g4oxdqn7"
    }
  }
}
```

#### Returns

| Field         | Type       | Description                                                          |
| ------------- | ---------- | -------------------------------------------------------------------- |
| `success`     | `Boolean!` | `true` when the comment was soft-deleted.                            |
| `operationId` | `String`   | Identifier for the mutation, useful for correlating realtime events. |

A deleted comment still appears in [`commentList`](/api/comments/query-comments) results with an empty body and a populated `deletedAt` / `deletedBy` — clients render it as "this comment was deleted."

### Errors

| Code                | When                                                                                                                 |
| ------------------- | -------------------------------------------------------------------------------------------------------------------- |
| `COMMENT_NOT_FOUND` | No comment exists with the given `id`.                                                                               |
| `FORBIDDEN`         | The caller isn't the author and isn't an `ADMIN`/`OWNER` of the project, is `VIEW_ONLY`, or the project is archived. |
| `UNAUTHENTICATED`   | No valid token was supplied.                                                                                         |

## Permissions

All three mutations require the caller to be a member of the target's project and deny `VIEW_ONLY` access. The project must be active (not archived).

| Mutation        | Who can call it                                                                                                 |
| --------------- | --------------------------------------------------------------------------------------------------------------- |
| `createComment` | Any non-`VIEW_ONLY` member of the target's project. For records, the role's record access must also be enabled. |
| `editComment`   | Only the comment's author.                                                                                      |
| `deleteComment` | The comment's author, **or** an `ADMIN`/`OWNER` of the project.                                                 |

## Related

- [Comments overview](/api/comments) — the `Comment` type, `CommentCategory`, threading, and read state.
- [Query comments](/api/comments/query-comments) — fetch comments for any target with `commentList`.
- [Reactions](/api/comments/reactions) — add or remove emoji reactions on a comment.
- [Manage discussions](/api/comments/manage-discussions) — create the discussions that comments attach to.
- [Comment & discussion subscriptions](/api/realtime/comment-discussion-subscriptions) — stream comment changes in real time.
