---
title: Comments overview
description: How comments and discussions work in Blue - one polymorphic Comment type attaches to records, discussions, and status updates, plus threading, reactions, and live updates.
icon: MessageSquare
order: 0
---

Comments and discussions are how teams converse inside Blue. A single, polymorphic `Comment` attaches to a record (a `Todo`), a project discussion, or a status update via a `category` + `categoryId` pair — so the same `createComment` / `editComment` / `deleteComment` mutations and the `commentList` query cover every comment surface. Discussions are standalone, project-scoped conversation threads with their own create/update/delete mutations and queries. This section also covers emoji reactions on comments and the real-time subscriptions (new comments, typing indicators, discussion changes) that power live collaboration.

## The polymorphic comment model

There is **one** `Comment` type and **one** set of comment mutations. A comment doesn't have a separate "todo comment" or "discussion comment" type — instead, every comment names its target with two fields:

- **`category`** — a `CommentCategory` enum: `TODO`, `DISCUSSION`, or `STATUS_UPDATE`.
- **`categoryId`** — the ID of the target. Its meaning depends on the category: the record (`Todo`) ID for `TODO`, the `Discussion` ID for `DISCUSSION`, or the `StatusUpdate` ID for `STATUS_UPDATE`.

Together they answer "what is this comment attached to?". You pass the same pair to write a comment ([createComment](/api/comments/manage-comments)) and to read a target's comments ([commentList](/api/comments/query-comments)). Once you understand this targeting, the rest of the section is just the same operations applied to three surfaces.

```graphql
# A comment on a record, a discussion, and a status update —
# same mutation, only category + categoryId differ.
mutation CommentOnRecord {
  createComment(
    input: { category: TODO, categoryId: "todo_123", html: "<p>On it.</p>", text: "On it." }
  ) {
    id
  }
}
```

<Callout variant="info" title="Discussions vs. discussion comments">

A **discussion** is the thread itself — a `Discussion` row with a title, body, and members, created with `createDiscussion`. The **replies** inside it are `Comment` rows with `category: DISCUSSION` and `categoryId` set to that discussion's ID. Creating a discussion does not create its comments; you add those separately with `createComment`.

</Callout>

## The Comment type

`Comment` is the shape returned by `createComment`, `editComment`, and every element of `CommentList.comments`.

| Field          | Type                | Description                                                                                            |
| -------------- | ------------------- | ------------------------------------------------------------------------------------------------------ |
| `id`           | `ID!`               | Unique identifier for the comment.                                                                     |
| `uid`          | `String!`           | Short reference identifier.                                                                            |
| `html`         | `String!`           | The comment body as sanitized HTML. Blanked to `""` when the comment is soft-deleted.                  |
| `text`         | `String!`           | The comment body as plain text. Blanked to `""` when the comment is soft-deleted.                      |
| `category`     | `CommentCategory!`  | Which surface this comment belongs to: `TODO`, `DISCUSSION`, or `STATUS_UPDATE`.                       |
| `createdAt`    | `DateTime!`         | When the comment was created.                                                                          |
| `updatedAt`    | `DateTime!`         | When the comment was last edited.                                                                      |
| `deletedAt`    | `DateTime`          | When the comment was soft-deleted, or `null` if it is still live. The row is kept even after deletion. |
| `deletedBy`    | `User`              | The user who soft-deleted the comment, or `null`.                                                      |
| `user`         | `User!`             | The author. Select `id`, `fullName`, `email`.                                                          |
| `activity`     | `Activity`          | The activity-feed entry this comment generated, if any.                                                |
| `todo`         | `Todo`              | The parent record, when `category` is `TODO`; otherwise `null`.                                        |
| `discussion`   | `Discussion`        | The parent discussion, when `category` is `DISCUSSION`; otherwise `null`.                              |
| `statusUpdate` | `StatusUpdate`      | The parent status update, when `category` is `STATUS_UPDATE`; otherwise `null`.                        |
| `isRead`       | `Boolean`           | Whether the calling user has read this comment.                                                        |
| `isSeen`       | `Boolean`           | Whether the calling user has seen this comment in the feed.                                            |
| `aiSummary`    | `Boolean`           | Whether the comment was generated as an AI summary.                                                    |
| `parentId`     | `String`            | The ID of the comment this one replies to, or `null` for a top-level comment.                          |
| `parent`       | `Comment`           | The parent comment when this is a threaded reply; `null` for top-level comments.                       |
| `replies`      | `[Comment!]!`       | Direct replies to this comment.                                                                        |
| `replyCount`   | `Int!`              | Number of direct replies.                                                                              |
| `reactions`    | `[ReactionGroup!]!` | Emoji reactions on this comment, grouped by emoji. See [Reactions](/api/comments/reactions).           |
| `files`        | `[File!]`           | Files attached to the comment.                                                                         |

### Threaded replies

Comments thread one level deep. A reply is an ordinary comment created with a `parentId` pointing at the comment it answers; from the parent's side you read `replies` and `replyCount`. Top-level comments have `parentId: null`. The [commentList](/api/comments/query-comments) query returns only top-level comments — fetch a thread's replies through the `replies` field on each comment.

### Read state and soft deletion

`isRead` and `isSeen` are evaluated per calling user, so the same comment can come back read for one member and unread for another. Deletion is a **soft delete**: [deleteComment](/api/comments/manage-comments) blanks `html`/`text` and stamps `deletedAt` + `deletedBy` rather than removing the row, which keeps reply chains and thread structure intact.

## The CommentList type

`CommentList` is returned by the [commentList](/api/comments/query-comments) query — a page of top-level comments for one target, plus pagination metadata.

| Field        | Type          | Description                                            |
| ------------ | ------------- | ------------------------------------------------------ |
| `comments`   | `[Comment!]!` | The top-level comments on this page.                   |
| `pageInfo`   | `PageInfo!`   | Cursor/offset pagination metadata for the result.      |
| `totalCount` | `Int!`        | Total number of top-level comments matching the query. |

## The CommentCategory enum

`CommentCategory` is the discriminator that makes the comment model polymorphic. Its value pairs with `categoryId` to identify the comment's target.

| Value           | `categoryId` refers to | Surface                                                                         |
| --------------- | ---------------------- | ------------------------------------------------------------------------------- |
| `TODO`          | a `Todo` (record) ID   | A comment on a record. See [Records › Add a comment](/api/records/add-comment). |
| `DISCUSSION`    | a `Discussion` ID      | A reply inside a project discussion.                                            |
| `STATUS_UPDATE` | a `StatusUpdate` ID    | A comment on a [status update](/api/status-updates).                            |

## Pages in this section

| Page                                                                    | What it covers                                                                                                                           |
| ----------------------------------------------------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| [Create, edit & delete comments](/api/comments/manage-comments)         | `createComment`, `editComment`, and the soft-delete `deleteComment` — the write lifecycle on any surface, with mentions and permissions. |
| [Query comments](/api/comments/query-comments)                          | The `commentList` query: scope to a record, discussion, or status update with `category` + `categoryId`, paginate, and order results.    |
| [Create, update & delete discussions](/api/comments/manage-discussions) | `createDiscussion`, `updateDiscussion` (title only), and the hard-delete `deleteDiscussion` for project-scoped threads.                  |
| [Query discussions](/api/comments/query-discussions)                    | The three read paths: `discussion(id)`, the offset-paginated `discussions`, and the project-scoped cursor list `discussionList`.         |
| [Reactions](/api/comments/reactions)                                    | `addReaction` / `removeReaction` for emoji reactions on a comment, returning the full `[ReactionGroup!]!` set.                           |

Live updates for comments and discussions — new comments, typing indicators, and discussion changes — are documented with the rest of the real-time API in [Comments, Discussions & Chat subscriptions](/api/realtime/comment-discussion-subscriptions).

## Product nouns to schema types

Use the UI word in prose and the schema name in code. The mapping for this section:

| UI / docs word      | Schema type / field              |
| ------------------- | -------------------------------- |
| Comment             | `Comment`                        |
| Discussion          | `Discussion`                     |
| Status update       | `StatusUpdate`                   |
| Record              | `Todo`                           |
| Workspace / Project | `Project`                        |
| Comment target kind | `CommentCategory` + `categoryId` |
| Reaction (grouped)  | `ReactionGroup`                  |

## Related

- [Create, edit & delete comments](/api/comments/manage-comments)
- [Query comments](/api/comments/query-comments)
- [Create, update & delete discussions](/api/comments/manage-discussions)
- [Reactions](/api/comments/reactions)
- [Add a comment to a record](/api/records/add-comment)
- [Comments, Discussions & Chat subscriptions](/api/realtime/comment-discussion-subscriptions)
