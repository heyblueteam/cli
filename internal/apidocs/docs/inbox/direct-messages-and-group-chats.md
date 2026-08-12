---
title: Direct Messages & Group Chats
description: Find who you can message, start a 1:1 or group conversation, and manage group membership and admins with startDirectMessage, createGroupChat, and the membership mutations.
icon: MessagesSquare
order: 3
---

Direct messages and group chats are `Chat` rows — the same type that already backed workspace chat threads — distinguished by `kind`: `CHANNEL` for a workspace chat, `DM` for a 1:1, `GROUP` for a named multi-person conversation. A DM/group `Chat` has `project: null` (it has no workspace anchor) and carries a `members: [ChatMember!]!` list that a `CHANNEL` chat does not use. Replies in any of the three are posted the same way — as `Comment` rows with `category: DISCUSSION` — see [Create, edit & delete comments](/api/comments/manage-comments).

<Callout variant="info" title="Use `chatId`, not `discussionId`">

The group membership mutations — `addGroupChatMembers`, `removeGroupChatMember`, `setGroupChatMemberRole`, and `leaveGroupChat` — take `chatId` as the modern spelling. The older `discussionId` field is deprecated but still accepted; if you send both, `chatId` wins. The examples below use `discussionId` and remain valid.

</Callout>

## messageableUsers

Before starting a conversation, use `messageableUsers` to find who you're allowed to message. Organization `OWNER`/`ADMIN` can message anyone in the organization; everyone else can message organization `OWNER`/`ADMIN`s plus anyone they share at least one active workspace with. The same rule is enforced server-side on `startDirectMessage`, `createGroupChat`, and `addGroupChatMembers` — this query only lists it.

### Request

```graphql
query WhoCanIMessage {
  messageableUsers(query: "pri", first: 25) {
    id
    fullName
    email
  }
}
```

### Parameters

| Parameter | Type     | Required | Default | Description                                                      |
| --------- | -------- | -------- | ------- | ---------------------------------------------------------------- |
| `query`   | `String` | No       | —       | Filter by name (first or last). Matched on name only, not email. |
| `first`   | `Int`    | No       | `25`    | Maximum number of users to return, clamped to 100.               |

### Response

```json
{
  "data": {
    "messageableUsers": [
      { "id": "clm4n8qwx000108l0a1b2c3d4", "fullName": "Priya Shah", "email": "priya@example.com" }
    ]
  }
}
```

`messageableUsers` returns `[User!]!` and never includes yourself.

## startDirectMessage

Open (or return) the canonical 1:1 conversation with another user. Calling it again for the same pair of users returns the **same** `Chat` rather than creating a duplicate — DMs are deduplicated per organization, keyed by the two user IDs regardless of who calls first.

### Request

```graphql
mutation MessagePriya {
  startDirectMessage(userId: "user_456") {
    id
    kind
    members {
      user {
        id
        fullName
      }
      role
    }
  }
}
```

### Parameters

| Parameter | Type      | Required | Description                              |
| --------- | --------- | -------- | ---------------------------------------- |
| `userId`  | `String!` | Yes      | The user to start (or resume) a DM with. |

### Response

```json
{
  "data": {
    "startDirectMessage": {
      "id": "clm4n8qwx000308l0i9j0k1l2",
      "kind": "DM",
      "members": [
        { "user": { "id": "clm4n8qwx000108l0h2pyerm8", "fullName": "You" }, "role": "MEMBER" },
        {
          "user": { "id": "clm4n8qwx000108l0a1b2c3d4", "fullName": "Priya Shah" },
          "role": "MEMBER"
        }
      ]
    }
  }
}
```

#### Returns

Returns the [`Chat`](#the-chat-type) — either newly created or the existing DM between you and `userId` in this organization.

### Errors

| Code              | When                                                                                                                                            |
| ----------------- | ----------------------------------------------------------------------------------------------------------------------------------------------- |
| `FORBIDDEN`       | Direct messaging is disabled for the organization, or you don't share a workspace with `userId` (and neither of you is an org `OWNER`/`ADMIN`). |
| `BAD_USER_INPUT`  | `userId` is your own ID, or doesn't belong to a member of your organization.                                                                    |
| `UNAUTHENTICATED` | No valid token was supplied.                                                                                                                    |

## createGroupChat

Create a named (or unnamed) group conversation. You become its first admin automatically.

### Request

```graphql
mutation StartLaunchCrew {
  createGroupChat(input: { title: "Launch crew", userIds: ["user_456", "user_789"] }) {
    id
    kind
    title
    members {
      user {
        id
        fullName
      }
      role
    }
  }
}
```

### Parameters

#### CreateGroupChatInput

| Parameter | Type         | Required | Description                                                                                                        |
| --------- | ------------ | -------- | ------------------------------------------------------------------------------------------------------------------ |
| `title`   | `String`     | No       | Display name. Left blank (or omitted), the group has no title — clients typically render the member names instead. |
| `userIds` | `[String!]!` | Yes      | The other people to add. Your own ID is stripped automatically if included; duplicates are de-duplicated.          |

<Callout variant="info" title="Size limits and no dedup">

A group needs at least one other person besides you, and at most 100 total members (a hard-coded ceiling — there's no technical reason it couldn't be higher, it's there so a group chat can't quietly become a second all-company channel). Unlike DMs, groups are **never deduplicated** by membership — calling `createGroupChat` twice with the same `userIds` creates two separate groups, told apart by title or recency.

</Callout>

### Response

```json
{
  "data": {
    "createGroupChat": {
      "id": "clm4n8qwx000408l0m3n4o5p6",
      "kind": "GROUP",
      "title": "Launch crew",
      "members": [
        { "user": { "id": "clm4n8qwx000108l0h2pyerm8", "fullName": "You" }, "role": "ADMIN" },
        {
          "user": { "id": "clm4n8qwx000108l0a1b2c3d4", "fullName": "Priya Shah" },
          "role": "MEMBER"
        },
        {
          "user": { "id": "clm4n8qwx000508l0mtu2ivpc", "fullName": "Sarah Mensah" },
          "role": "MEMBER"
        }
      ]
    }
  }
}
```

#### Returns

Returns the created [`Chat`](#the-chat-type) with `kind: GROUP`.

### Errors

| Code              | When                                                                                                                                                    |
| ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `FORBIDDEN`       | Direct messaging is disabled for the organization, or you don't share a workspace with one of `userIds` (and neither of you is an org `OWNER`/`ADMIN`). |
| `BAD_USER_INPUT`  | `userIds` (after removing your own ID and duplicates) is empty, exceeds 100 total members, or references someone outside your organization.             |
| `UNAUTHENTICATED` | No valid token was supplied.                                                                                                                            |

## addGroupChatMembers

Add people to an existing group. Admin-only. A person re-added after being removed regains access to the **entire** message history, exactly like a brand-new member — their role also resets to `MEMBER`, even if they were an admin before being removed.

### Request

```graphql
mutation AddToLaunchCrew {
  addGroupChatMembers(input: { discussionId: "chat_123", userIds: ["user_999"] }) {
    id
    members {
      user {
        id
        fullName
      }
      role
      removedAt
    }
  }
}
```

### Parameters

#### GroupChatMembersInput

| Parameter      | Type         | Required | Description                       |
| -------------- | ------------ | -------- | --------------------------------- |
| `discussionId` | `String!`    | Yes      | The group chat to add members to. |
| `userIds`      | `[String!]!` | Yes      | The people to add.                |

### Response

```json
{
  "data": {
    "addGroupChatMembers": {
      "id": "clm4n8qwx000408l0m3n4o5p6",
      "members": [
        {
          "user": { "id": "clm4n8qwx000608l0u1v2w3x4", "fullName": "New Person" },
          "role": "MEMBER",
          "removedAt": null
        }
      ]
    }
  }
}
```

#### Returns

Returns the updated [`Chat`](#the-chat-type).

### Errors

| Code                   | When                                                                                                                                                                                                            |
| ---------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `DISCUSSION_NOT_FOUND` | No group chat matches `discussionId`.                                                                                                                                                                           |
| `FORBIDDEN`            | You're not an active admin of the group, `discussionId` isn't a `GROUP` (e.g. it's a DM), or direct messaging is disabled for the organization.                                                                 |
| `BAD_USER_INPUT`       | `userIds` (after removing your own ID) is empty, the addition would exceed 100 total members, or a target isn't messageable (outside your organization, or no shared workspace and not an org `OWNER`/`ADMIN`). |
| `UNAUTHENTICATED`      | No valid token was supplied.                                                                                                                                                                                    |

## removeGroupChatMember

Remove someone from a group. Admin-only, and the row is soft-removed rather than deleted — the removed member keeps read-only access to everything up to the moment they were removed, but nothing after.

### Request

```graphql
mutation RemoveFromLaunchCrew {
  removeGroupChatMember(input: { discussionId: "chat_123", userId: "user_999" }) {
    id
    members {
      user {
        id
      }
      removedAt
    }
  }
}
```

### Parameters

#### GroupChatMemberInput

| Parameter      | Type      | Required | Description                            |
| -------------- | --------- | -------- | -------------------------------------- |
| `discussionId` | `String!` | Yes      | The group chat to remove someone from. |
| `userId`       | `String!` | Yes      | The person to remove.                  |

To remove **yourself**, use [`leaveGroupChat`](#leavegroupchat) instead — passing your own `userId` here is rejected.

### Response

```json
{
  "data": {
    "removeGroupChatMember": {
      "id": "clm4n8qwx000408l0m3n4o5p6",
      "members": [
        { "user": { "id": "clm4n8qwx000608l0u1v2w3x4" }, "removedAt": "2026-07-15T10:02:00.000Z" }
      ]
    }
  }
}
```

#### Returns

Returns the updated [`Chat`](#the-chat-type).

### Errors

| Code                   | When                                                                                    |
| ---------------------- | --------------------------------------------------------------------------------------- |
| `DISCUSSION_NOT_FOUND` | No group chat matches `discussionId`.                                                   |
| `FORBIDDEN`            | You're not an active admin of the group.                                                |
| `BAD_USER_INPUT`       | `userId` is your own ID (use `leaveGroupChat`), or isn't an active member of the group. |
| `UNAUTHENTICATED`      | No valid token was supplied.                                                            |

## leaveGroupChat

Leave a group chat you're a member of. Only group chats can be left — passing a DM's `discussionId` is rejected, which is what makes DMs permanent for as long as both people remain in the organization.

If you're the last active admin, the longest-standing remaining member (by when they joined this conversation) is automatically promoted to admin. If you're the last member of any role, the conversation is deleted entirely.

### Request

```graphql
mutation LeaveLaunchCrew {
  leaveGroupChat(discussionId: "chat_123")
}
```

### Parameters

| Parameter      | Type      | Required | Description              |
| -------------- | --------- | -------- | ------------------------ |
| `discussionId` | `String!` | Yes      | The group chat to leave. |

`leaveGroupChat` returns a bare `Boolean!`.

### Response

```json
{
  "data": {
    "leaveGroupChat": true
  }
}
```

<Callout variant="info" title="Auto-promotion order">

When the last admin leaves, the replacement is the active member with the oldest `ChatMember` row in **this** conversation (i.e. whoever joined earliest), not the oldest Blue account. Ties are broken by an internal ordering, not join time, since a batch of members added in the same `createGroupChat`/`addGroupChatMembers` call can share an identical timestamp.

</Callout>

### Errors

| Code                   | When                                                                                                        |
| ---------------------- | ----------------------------------------------------------------------------------------------------------- |
| `DISCUSSION_NOT_FOUND` | No group chat matches `discussionId`.                                                                       |
| `FORBIDDEN`            | `discussionId` refers to a DM, not a group chat (only group chats can be left), or you were never a member. |
| `BAD_USER_INPUT`       | You're not currently an active member (you already left or were removed).                                   |
| `UNAUTHENTICATED`      | No valid token was supplied.                                                                                |

## setGroupChatMemberRole

Promote a member to admin, or demote an admin to member. Admin-only.

### Request

```graphql
mutation PromoteToAdmin {
  setGroupChatMemberRole(input: { discussionId: "chat_123", userId: "user_456", role: ADMIN }) {
    id
    members {
      user {
        id
      }
      role
    }
  }
}
```

### Parameters

#### SetGroupChatMemberRoleInput

| Parameter      | Type              | Required | Description                      |
| -------------- | ----------------- | -------- | -------------------------------- |
| `discussionId` | `String!`         | Yes      | The group chat.                  |
| `userId`       | `String!`         | Yes      | The member whose role to change. |
| `role`         | `ChatMemberRole!` | Yes      | `ADMIN` or `MEMBER`.             |

Setting a member's role to the role they already have is a no-op — it returns the unchanged `Chat` rather than erroring.

### Response

```json
{
  "data": {
    "setGroupChatMemberRole": {
      "id": "clm4n8qwx000408l0m3n4o5p6",
      "members": [{ "user": { "id": "clm4n8qwx000108l0a1b2c3d4" }, "role": "ADMIN" }]
    }
  }
}
```

#### Returns

Returns the updated [`Chat`](#the-chat-type).

### Errors

| Code                   | When                                                                                                                               |
| ---------------------- | ---------------------------------------------------------------------------------------------------------------------------------- |
| `DISCUSSION_NOT_FOUND` | No group chat matches `discussionId`.                                                                                              |
| `FORBIDDEN`            | You're not an active admin of the group.                                                                                           |
| `BAD_USER_INPUT`       | `userId` isn't an active member, or the change would demote the group's only remaining admin — every group must keep at least one. |
| `UNAUTHENTICATED`      | No valid token was supplied.                                                                                                       |

## The Chat type

The fields relevant to DMs and group chats. `Chat` also backs workspace chat threads (`kind: CHANNEL`) — see [Create, update & delete chats](/api/comments/manage-discussions) for that surface's full field list.

| Field          | Type             | Description                                                                                                        |
| -------------- | ---------------- | ------------------------------------------------------------------------------------------------------------------ |
| `id`           | `ID!`            | The conversation's unique identifier.                                                                              |
| `title`        | `String!`        | Display name. Empty for an unnamed group — clients typically render the member names instead.                      |
| `kind`         | `ChatKind!`      | `CHANNEL` (workspace chat), `DM` (1:1), or `GROUP`.                                                                |
| `project`      | `Project`        | `null` for `DM`/`GROUP` — they have no workspace anchor. Set for `CHANNEL`.                                        |
| `members`      | `[ChatMember!]!` | Membership rows for `DM`/`GROUP` conversations, including removed members (`removedAt` set). Unused for `CHANNEL`. |
| `people`       | `[User!]`        | Participants in the thread.                                                                                        |
| `commentCount` | `Int!`           | Number of replies (`Comment` rows with `category: DISCUSSION`) in the thread.                                      |
| `isRead`       | `Boolean`        | Whether you've read this thread. Evaluated per caller.                                                             |
| `isSeen`       | `Boolean`        | Whether you've seen this thread in a feed. Evaluated per caller.                                                   |
| `createdAt`    | `DateTime!`      | When the conversation was created.                                                                                 |
| `updatedAt`    | `DateTime!`      | When the conversation was last changed.                                                                            |

### ChatKind

```graphql
enum ChatKind {
  CHANNEL
  DM
  GROUP
}
```

### ChatMember

| Field       | Type              | Description                                                           |
| ----------- | ----------------- | --------------------------------------------------------------------- |
| `id`        | `ID!`             | The membership row's identifier.                                      |
| `user`      | `User!`           | The member.                                                           |
| `role`      | `ChatMemberRole!` | `ADMIN` or `MEMBER`.                                                  |
| `createdAt` | `DateTime!`       | When this person joined the conversation.                             |
| `removedAt` | `DateTime`        | Set when the member was removed or left. `null` for an active member. |
| `addedBy`   | `User`            | Who added them, when known.                                           |

```graphql
enum ChatMemberRole {
  ADMIN
  MEMBER
}
```

## Permissions

| Operation                | Requires                                                                                                                                    |
| ------------------------ | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `messageableUsers`       | Authenticated, active organization membership.                                                                                              |
| `startDirectMessage`     | Authenticated, active organization membership. Target must be messageable (see above).                                                      |
| `createGroupChat`        | Authenticated, active organization membership. Every invitee must be messageable.                                                           |
| `addGroupChatMembers`    | Active admin of the group. Every invitee must be messageable.                                                                               |
| `removeGroupChatMember`  | Active admin of the group.                                                                                                                  |
| `leaveGroupChat`         | Authenticated, in-organization member of the group. (Does not require the organization's license to be active — leaving is always allowed.) |
| `setGroupChatMemberRole` | Active admin of the group.                                                                                                                  |

An organization can disable direct messaging entirely; when disabled, `startDirectMessage`, `createGroupChat`, and `addGroupChatMembers` all reject with `FORBIDDEN`.

## Related

- [Inbox & Messaging overview](/api/inbox) — how DM/group threads surface in the unified Inbox (`DM_THREAD`).
- [Query your Inbox](/api/inbox/query-inbox) — read the conversations these mutations create.
- [Manage inbox threads](/api/inbox/manage-inbox-threads) — mute, pin, and mark read/unread.
- [Comments overview](/api/comments) — post replies with `createComment(category: DISCUSSION, ...)`.
- [Query chats](/api/comments/query-discussions) — the read paths shared with workspace chat threads.
- [Inbox subscriptions](/api/realtime/inbox-subscriptions) — stream new activity in real time.
