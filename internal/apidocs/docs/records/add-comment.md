---
title: Add Comment
description: Add a comment to a record with rich HTML content, file attachments, @mentions, and threaded replies.
icon: MessageSquare
order: 8
---

Use the `createComment` mutation to post a comment on a record. Comments accept rich HTML content, file attachments, and @mentions, and can be posted as threaded replies to an existing comment. Each comment is added to the record's activity feed and triggers the relevant notifications, subscriptions, and webhooks. Records are `Record` objects in the API.

The same mutation comments on a record (`category: TODO`), a chat (`category: DISCUSSION`), or a status update (`category: STATUS_UPDATE`); this page focuses on commenting on records.

## Request

The smallest call posts a comment on one record. Pass the record ID as `categoryId` with `category: TODO`, and supply both the rendered `html` and a plain-text `text` fallback.

```graphql
mutation AddComment {
  createComment(
    input: {
      category: TODO
      categoryId: "todo_123"
      html: "<p>This task is progressing well.</p>"
      text: "This task is progressing well."
    }
  ) {
    id
    html
    createdAt
    user {
      id
      fullName
    }
  }
}
```

## Parameters

### CreateCommentInput

| Parameter    | Type               | Required | Description                                                                                                                                  |
| ------------ | ------------------ | -------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `category`   | `CommentCategory!` | Yes      | The kind of entity being commented on. Use `TODO` for records.                                                                               |
| `categoryId` | `String!`          | Yes      | ID of the entity being commented on. For `TODO`, this is the record (`Record`) ID.                                                           |
| `html`       | `String!`          | Yes      | Rendered HTML of the comment. Sanitized on the server before it is stored.                                                                   |
| `text`       | `String!`          | Yes      | Plain-text version of the comment, used as a fallback and for search.                                                                        |
| `tiptap`     | `Boolean`          | No       | Enables TipTap-aware sanitization. Set `true` to parse embedded file-attachment `<div>` elements. Without it, attachment markup is stripped. |
| `parentId`   | `String`           | No       | ID of the comment to reply to. Omit for a top-level comment; set it to nest this comment as a threaded reply.                                |

### CommentCategory

| Value           | Description                       |
| --------------- | --------------------------------- |
| `TODO`          | A comment on a record.            |
| `DISCUSSION`    | A comment on a chat thread. |
| `STATUS_UPDATE` | A comment on a status update.     |

## Response

```json
{
  "data": {
    "createComment": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "html": "<p>This task is progressing well.</p>",
      "createdAt": "2026-05-29T14:03:11.000Z",
      "user": {
        "id": "clm4n8qwx000108l0d2k1abcd",
        "fullName": "Ada Lovelace"
      }
    }
  }
}
```

### Returns

`createComment` returns the created `Comment`.

| Field          | Type                | Description                                                                                                             |
| -------------- | ------------------- | ----------------------------------------------------------------------------------------------------------------------- |
| `id`           | `ID!`               | Unique identifier for the comment.                                                                                      |
| `uid`          | `String!`           | Short public identifier.                                                                                                |
| `html`         | `String!`           | Sanitized HTML content of the comment.                                                                                  |
| `text`         | `String!`           | Plain-text content.                                                                                                     |
| `category`     | `CommentCategory!`  | The entity type this comment belongs to.                                                                                |
| `createdAt`    | `DateTime!`         | When the comment was created.                                                                                           |
| `updatedAt`    | `DateTime!`         | When the comment was last updated.                                                                                      |
| `deletedAt`    | `DateTime`          | When the comment was deleted, or `null` if active.                                                                      |
| `deletedBy`    | `User`              | The user who deleted the comment, if any.                                                                               |
| `user`         | `User!`             | The author of the comment. Select `id`, `fullName`, and `image { … }` (there is no `name` or `avatar` field on `User`). |
| `activity`     | `Activity`          | The activity-feed entry created for this comment.                                                                       |
| `todo`         | `Record`            | The record this comment is on, when `category` is `TODO`.                                                               |
| `chat`   | `Chat`        | The chat, when `category` is `DISCUSSION`.                                                                        |
| `statusUpdate` | `StatusUpdate`      | The status update, when `category` is `STATUS_UPDATE`.                                                                  |
| `parentId`     | `String`            | ID of the parent comment, or `null` for a top-level comment.                                                            |
| `parent`       | `Comment`           | The parent comment when this is a threaded reply.                                                                       |
| `replies`      | `[Comment!]!`       | The threaded replies to this comment.                                                                                   |
| `replyCount`   | `Int!`              | Number of replies on this comment.                                                                                      |
| `reactions`    | `[ReactionGroup!]!` | Emoji reactions, grouped by emoji (`emoji`, `count`, `users { … }`, `reactedByMe`).                                     |
| `isRead`       | `Boolean`           | Whether the current user has read this comment.                                                                         |
| `isSeen`       | `Boolean`           | Whether the current user has seen this comment.                                                                         |
| `aiSummary`    | `Boolean`           | Whether this comment was generated as an AI summary.                                                                    |

## Full example

Post a richly formatted comment as a threaded reply, and read back the threading and reaction fields:

```graphql
mutation AddReply {
  createComment(
    input: {
      category: TODO
      categoryId: "todo_123"
      parentId: "comment_123"
      tiptap: true
      html: "<p>Agreed — here is my <strong>feedback</strong>:</p><ul><li>Great progress on the design</li><li>Need to review the API integration</li></ul>"
      text: "Agreed — here is my feedback: - Great progress on the design - Need to review the API integration"
    }
  ) {
    id
    html
    createdAt
    parentId
    replyCount
    reactions {
      emoji
      count
      reactedByMe
    }
    user {
      id
      fullName
      image {
        thumbnail
        small
      }
    }
  }
}
```

## File attachments

To attach an uploaded file so it renders as a clickable attachment (the same way files appear when attached in the Blue UI), embed it as a `<div>` in the `html` field and set `tiptap: true`. Without `tiptap: true`, the attachment markup is stripped by sanitization.

### Step 1 — Upload the file

Upload the file with the [`uploadFile` mutation](/api/files) to get its metadata (`uid`, `name`, `size`, `type`, `extension`).

### Step 2 — Embed the attachment

Each attachment is a `<div>` carrying a JSON `file` attribute:

```html
<div
  class="attachment"
  file='{"uid":"FILE_UID","name":"FILENAME","size":SIZE_IN_BYTES,"type":"MIME_TYPE","extension":"EXT"}'
></div>
```

### Step 3 — Post the comment

```graphql
mutation AddCommentWithFile {
  createComment(
    input: {
      category: TODO
      categoryId: "todo_123"
      tiptap: true
      html: "<p>Here is the report.</p><div class=\"attachment\" file='{\"uid\":\"cm8qujq3b01p22lrv9xbb2q62\",\"name\":\"report.pdf\",\"size\":204800,\"type\":\"application/pdf\",\"extension\":\"pdf\"}'></div>"
      text: "Here is the report."
    }
  ) {
    id
    html
    createdAt
  }
}
```

For an end-to-end walkthrough (upload + comment, with Python), see [Attaching files to comments](/api/files#attaching-files-to-records-and-comments). Once uploaded, a file's download URL is `https://api.blue.app/uploads/{uid}/{url_encoded_filename}` — see [File URLs](/api/files#serving-files) for thumbnail and inline variants.

<Callout variant="info" title="Side effects">

Posting a comment creates an activity-feed entry, fires any configured project webhooks, publishes to GraphQL subscriptions for live updates, and sends notifications — including targeted notifications for any @mentioned users.

</Callout>

## Errors

| Code             | When                                                                                                                                                                                                                                               |
| ---------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `FORBIDDEN`      | The user is not a member of the record's workspace, has `VIEW_ONLY` access, holds a custom role with records disabled, or the workspace is archived. The same code is returned when `categoryId` does not resolve to a record the user can access. |
| `BAD_USER_INPUT` | A required field is missing, `category` is not a valid `CommentCategory`, or the input is otherwise malformed.                                                                                                                                     |

```json
{
  "errors": [
    {
      "message": "You don't have permission to perform this action.",
      "extensions": { "code": "FORBIDDEN" }
    }
  ]
}
```

## Permissions

Any workspace member can comment on a record except those with `VIEW_ONLY` access. A custom role that has records disabled (`isRecordsEnabled: false`) cannot comment, and no one can comment on a record in an archived workspace.

| Access level   | Can comment |
| -------------- | ----------- |
| `OWNER`        | Yes         |
| `ADMIN`        | Yes         |
| `MEMBER`       | Yes         |
| `CLIENT`       | Yes         |
| `COMMENT_ONLY` | Yes         |
| `VIEW_ONLY`    | No          |

## Related

- [List records](/api/records/list-records)
- [Manage record assignees](/api/records/assignees)
- [Tags](/api/records/tags)
- [Upload files](/api/files)
