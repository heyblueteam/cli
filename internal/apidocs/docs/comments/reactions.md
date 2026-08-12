---
title: Reactions
description: Add or remove an emoji reaction on a comment and read back the comment's full set of reaction groups.
icon: SmilePlus
order: 5
---

Use the `addReaction` and `removeReaction` mutations to toggle an emoji reaction on a comment. Both take the same shape — a `targetType`, the `targetId` of the thing being reacted to, and the `emoji` itself — and both return the target's **complete** reaction set as `[ReactionGroup!]!`. Each `ReactionGroup` rolls up one emoji: how many people used it, who they are, and whether the calling user is one of them.

Reactions are stored one row per user, emoji, and target, so a user contributes at most one of any given emoji to a target. Both mutations are idempotent: adding a reaction you already have, or removing one you never had, succeeds and returns the current set unchanged.

The `targetType` is a `ReactionTargetType` enum with a single value — `COMMENT`. Comments are `Comment` objects in the API; see [Comments overview](/api/comments) for how a comment attaches to a record, chat, or status update.

This covers chat and direct messages too: a chat message _is_ a `Comment` (one whose `chat` is set), so react to it with `targetType: COMMENT` and the message's comment id as `targetId`.

<Callout variant="info" title="Not the community forum">

The unrelated <code>toggleCommunityReaction</code> mutation (LIKE / HEART / CELEBRATE on community posts) is a separate subsystem and is not covered here. For comment and chat reactions, use <code>addReaction</code> / <code>removeReaction</code>.

</Callout>

## Request

Add a thumbs-up to a comment. Pass the comment's id as `targetId` and an emoji string. The mutation returns the comment's full reaction set after the change.

```graphql
mutation AddReaction {
  addReaction(input: { targetType: COMMENT, targetId: "comment_123", emoji: "👍" }) {
    emoji
    count
    reactedByMe
    users {
      id
      fullName
    }
  }
}
```

Remove it again with the matching mutation — same input shape:

```graphql
mutation RemoveReaction {
  removeReaction(input: { targetType: COMMENT, targetId: "comment_123", emoji: "👍" }) {
    emoji
    count
    reactedByMe
  }
}
```

## Parameters

### AddReactionInput

| Parameter    | Type                  | Required | Description                                                                                                    |
| ------------ | --------------------- | -------- | -------------------------------------------------------------------------------------------------------------- |
| `targetType` | `ReactionTargetType!` | Yes      | The kind of object being reacted to. Use `COMMENT` for comments.                                               |
| `targetId`   | `String!`             | Yes      | The id of the target. For `COMMENT`, this is the `Comment.id`.                                                 |
| `emoji`      | `String!`             | Yes      | The emoji to add, e.g. `"👍"`. Trimmed of surrounding whitespace; must be non-empty and at most 32 characters. |

### RemoveReactionInput

Identical to `AddReactionInput`. The `emoji` is matched against the calling user's existing reaction on the target.

| Parameter    | Type                  | Required | Description                                                      |
| ------------ | --------------------- | -------- | ---------------------------------------------------------------- |
| `targetType` | `ReactionTargetType!` | Yes      | The kind of object being reacted to. Use `COMMENT` for comments. |
| `targetId`   | `String!`             | Yes      | The id of the target. For `COMMENT`, this is the `Comment.id`.   |
| `emoji`      | `String!`             | Yes      | The emoji to remove from the calling user's reactions.           |

### ReactionTargetType

| Value     | Description                                                                                    |
| --------- | ---------------------------------------------------------------------------------------------- |
| `COMMENT` | React to a `Comment`. `targetId` is the comment id. Chat and direct messages are comments too. |

## Response

Both mutations return the target's reaction set: one `ReactionGroup` per distinct emoji currently on the target, ordered by when each emoji first appeared. After adding 👍 to a comment that already had two 🎉 reactions:

```json
{
  "data": {
    "addReaction": [
      {
        "emoji": "🎉",
        "count": 2,
        "reactedByMe": false,
        "users": [
          { "id": "clm4n8qwx000008l0g4oxdqn7", "fullName": "Ada Lovelace" },
          { "id": "clm4n8qwx000108l0h2pzabc9", "fullName": "Alan Turing" }
        ]
      },
      {
        "emoji": "👍",
        "count": 1,
        "reactedByMe": true,
        "users": [{ "id": "clm4n8qwx000208l0k7rtd5e3", "fullName": "Grace Hopper" }]
      }
    ]
  }
}
```

Removing the last reaction of an emoji drops its group from the array entirely. A target with no reactions returns `[]`.

### ReactionGroup

| Field         | Type       | Description                                                          |
| ------------- | ---------- | -------------------------------------------------------------------- |
| `emoji`       | `String!`  | The emoji this group rolls up.                                       |
| `count`       | `Int!`     | Number of users who reacted with this emoji.                         |
| `users`       | `[User!]!` | The users who reacted with this emoji, ordered by when they reacted. |
| `reactedByMe` | `Boolean!` | `true` if the calling user is one of `users`.                        |

## Reading reactions without changing them

`addReaction` and `removeReaction` are the only operations that mutate reactions — there's no separate "set" or "clear" mutation. To read a comment's reactions without modifying them, select the `reactions` field on the `Comment` itself; it returns the same `[ReactionGroup!]!`. Fetch comments with the [`commentList` query](/api/comments/query-comments):

```graphql
query CommentReactions {
  commentList(category: TODO, categoryId: "todo_123") {
    comments {
      id
      reactions {
        emoji
        count
        reactedByMe
      }
    }
  }
}
```

## Errors

| Code              | When                                                                                                         |
| ----------------- | ------------------------------------------------------------------------------------------------------------ |
| `BAD_USER_INPUT`  | The comment in `targetId` does not exist, or `emoji` is empty (after trimming) or longer than 32 characters. |
| `FORBIDDEN`       | The calling user lacks comment access on the record the target comment belongs to (see Permissions).         |
| `UNAUTHENTICATED` | No valid credentials on the request.                                                                         |

## Permissions

For a comment attached to a **record**, both mutations require comment-level access to that record — the same access required to post a comment there. Callers with view-only access are denied with `FORBIDDEN`.

Comments attached to a **chat** or **status update** are not currently access-checked beyond authentication, so any authenticated member of the organization can react to them. Do not rely on a per-thread guarantee for those targets.

## Related

- [Comments overview](/api/comments) — how a `Comment` attaches to a record, chat, or status update.
- [Create, edit & delete comments](/api/comments/manage-comments) — the comment write lifecycle.
- [Query comments](/api/comments/query-comments) — fetch comments (and their `reactions`) for any target.
