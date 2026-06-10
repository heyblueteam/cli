---
title: Create, update & delete discussions
description: Open project-scoped discussion threads, rename them, and hard-delete them with the createDiscussion, updateDiscussion, and deleteDiscussion mutations.
icon: MessagesSquare
order: 3
---

A discussion is a standalone, project-scoped conversation thread — a place for a topic that isn't tied to a single record. Discussions are `Discussion` objects in the API, and they live in a workspace (`Project`). Once a thread exists, replies are posted as comments with `category: DISCUSSION` (see [Create, edit & delete comments](/api/comments/manage-comments)).

- Use `createDiscussion` to seed a new thread with a title and an opening body.
- Use `updateDiscussion` to rename a thread — it edits the **title only**; the body is fixed once created.
- Use `deleteDiscussion` to remove a thread. Unlike a comment delete, this is a hard delete.

To read discussions back, use the [discussion query paths](/api/comments/query-discussions). The `Discussion.comments` field is deprecated — load replies with [`commentList(category: DISCUSSION)`](/api/comments/query-comments) instead.

## createDiscussion

Create a discussion in a workspace. The opening message is supplied as both `html` (rich content, sanitized server-side) and `text` (the plain-text fallback). The mutation returns the created `Discussion`.

### Request

```graphql
mutation CreateDiscussion {
  createDiscussion(
    input: {
      projectId: "project_123"
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

#### CreateDiscussionInput

| Parameter   | Type      | Required | Description                                                                                                                   |
| ----------- | --------- | -------- | ----------------------------------------------------------------------------------------------------------------------------- |
| `title`     | `String!` | Yes      | The thread title.                                                                                                             |
| `html`      | `String!` | Yes      | The opening message as HTML. Sanitized server-side; embedded `@mentions`, images, and file attachments are extracted from it. |
| `text`      | `String!` | Yes      | The plain-text rendering of the opening message, used in notifications, search, and clients that don't render HTML.           |
| `projectId` | `String!` | Yes      | The workspace the discussion belongs to. Accepts a project id or slug.                                                        |

The body (`html` / `text`) is set once at creation and is immutable afterward — `updateDiscussion` changes only the title. To revise the opening message, delete the thread and create a new one, or post a follow-up comment.

### Response

```json
{
  "data": {
    "createDiscussion": {
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

Returns the created [`Discussion`](#discussion-type).

### Mentions, images & attachments

`@mentions`, inline images, and file attachments all travel inside the `html` field of the opening message — there are no separate arguments for them. To mention a user, include an anchor whose `href` is `#view-profile-<userId>`; the server rewrites it into a mention and notifies that user. The same convention applies when posting comment replies (see [Create, edit & delete comments](/api/comments/manage-comments)).

### Errors

| Code                | When                                                                                            |
| ------------------- | ----------------------------------------------------------------------------------------------- |
| `PROJECT_NOT_FOUND` | No workspace matches the supplied `projectId` (id or slug).                                     |
| `FORBIDDEN`         | The caller is `VIEW_ONLY` or `COMMENT_ONLY` in the workspace, isn't a member, or it's archived. |
| `UNAUTHENTICATED`   | No valid token was supplied.                                                                    |

## updateDiscussion

Rename a discussion. This is the only mutable field: `UpdateDiscussionInput` carries just `id` and an optional `title`. The mutation returns the updated `Discussion`.

### Request

```graphql
mutation UpdateDiscussion {
  updateDiscussion(input: { id: "discussion_123", title: "Q3 launch retro (final)" }) {
    id
    title
    updatedAt
  }
}
```

### Parameters

#### UpdateDiscussionInput

| Parameter | Type      | Required | Description                                                                                      |
| --------- | --------- | -------- | ------------------------------------------------------------------------------------------------ |
| `id`      | `String!` | Yes      | The id of the discussion to rename.                                                              |
| `title`   | `String`  | No       | The new title. Trimmed server-side; omitting it (or sending blank) clears it to an empty string. |

There is no `editDiscussion` mutation, and there is no way to change `html` / `text`, `projectId`, or any other field through this call — only the title.

### Response

```json
{
  "data": {
    "updateDiscussion": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Q3 launch retro (final)",
      "updatedAt": "2026-05-29T15:18:44.000Z"
    }
  }
}
```

#### Returns

Returns the updated [`Discussion`](#discussion-type).

### Errors

| Code                   | When                                                                                                          |
| ---------------------- | ------------------------------------------------------------------------------------------------------------- |
| `DISCUSSION_NOT_FOUND` | No discussion exists with the given `id`.                                                                     |
| `FORBIDDEN`            | The caller isn't the creator or a project `OWNER`, is `VIEW_ONLY`/`COMMENT_ONLY`, or the project is archived. |
| `UNAUTHENTICATED`      | No valid token was supplied.                                                                                  |

## deleteDiscussion

Remove a discussion thread. This is a **hard delete**: before the row is removed, its reply chains are unlinked (threaded comment replies have their `parentId` cleared to release the foreign-key constraint), and a snapshot of the discussion is written to the organization's Trash. The mutation returns a `MutationResult`.

This differs from `deleteComment`, which is a _soft_ delete that keeps the row and blanks its body — see [Create, edit & delete comments](/api/comments/manage-comments).

### Request

`deleteDiscussion` takes the discussion id directly (not an input object) and returns a `MutationResult`.

```graphql
mutation DeleteDiscussion {
  deleteDiscussion(id: "discussion_123") {
    success
    operationId
  }
}
```

### Parameters

| Parameter | Type      | Required | Description                         |
| --------- | --------- | -------- | ----------------------------------- |
| `id`      | `String!` | Yes      | The id of the discussion to delete. |

### Response

```json
{
  "data": {
    "deleteDiscussion": {
      "success": true,
      "operationId": "clm4n8qwx000008l0g4oxdqn7"
    }
  }
}
```

#### Returns

| Field         | Type       | Description                                                          |
| ------------- | ---------- | -------------------------------------------------------------------- |
| `success`     | `Boolean!` | `true` when the discussion was deleted.                              |
| `operationId` | `String`   | Identifier for the mutation, useful for correlating realtime events. |

### Errors

| Code                   | When                                                                                                          |
| ---------------------- | ------------------------------------------------------------------------------------------------------------- |
| `DISCUSSION_NOT_FOUND` | No discussion exists with the given `id`.                                                                     |
| `FORBIDDEN`            | The caller isn't the creator or a project `OWNER`, is `VIEW_ONLY`/`COMMENT_ONLY`, or the project is archived. |
| `UNAUTHENTICATED`      | No valid token was supplied.                                                                                  |

## Discussion type

The fields available on a `Discussion`. The comment thread itself is loaded separately via [`commentList(category: DISCUSSION)`](/api/comments/query-comments) — the `Discussion.comments` field is deprecated and should not be used.

| Field          | Type        | Description                                                                                     |
| -------------- | ----------- | ----------------------------------------------------------------------------------------------- |
| `id`           | `ID!`       | The discussion id.                                                                              |
| `title`        | `String!`   | The thread title (the only editable field).                                                     |
| `description`  | `String`    | A short clipped preview of the body (HTML truncated to ~180 characters), useful for list views. |
| `html`         | `String!`   | The opening message as sanitized HTML. Immutable after creation.                                |
| `text`         | `String!`   | The opening message as plain text.                                                              |
| `createdAt`    | `DateTime!` | When the thread was created.                                                                    |
| `updatedAt`    | `DateTime!` | When the thread was last changed (e.g. renamed).                                                |
| `user`         | `User!`     | The creator of the thread.                                                                      |
| `people`       | `[User!]`   | Users participating in the thread (creator plus commenters).                                    |
| `project`      | `Project!`  | The workspace the discussion belongs to.                                                        |
| `commentCount` | `Int!`      | The number of comments in the thread.                                                           |
| `isRead`       | `Boolean`   | Whether the calling user has read the thread.                                                   |
| `isSeen`       | `Boolean`   | Whether the calling user has seen (surfaced but not necessarily opened) the thread.             |

## Permissions

All three mutations require the caller to be a member of the discussion's workspace, and they deny `VIEW_ONLY` and `COMMENT_ONLY` access levels. The workspace must be active (not archived).

| Mutation           | Who can call it                                                                         |
| ------------------ | --------------------------------------------------------------------------------------- |
| `createDiscussion` | Any member of the workspace above `COMMENT_ONLY` (i.e. not `VIEW_ONLY`/`COMMENT_ONLY`). |
| `updateDiscussion` | The thread's creator, **or** a project `OWNER`.                                         |
| `deleteDiscussion` | The thread's creator, **or** a project `OWNER`.                                         |

## Related

- [Comments overview](/api/comments) — the `Comment` type, `CommentCategory`, threading, and read state.
- [Query discussions](/api/comments/query-discussions) — fetch a single thread, list threads, or page through a workspace's threads.
- [Create, edit & delete comments](/api/comments/manage-comments) — post the replies that hang off a discussion.
- [Query comments](/api/comments/query-comments) — load a discussion's replies with `commentList`.
- [Comment & discussion subscriptions](/api/realtime/comment-discussion-subscriptions) — stream discussion and comment changes in real time.
