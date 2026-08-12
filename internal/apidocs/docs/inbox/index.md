---
title: Inbox & Messaging overview
description: How Blue's unified Inbox works - one cross-workspace feed of record comment threads, workspace chats, AI conversations, and direct messages / group chats, backed by a real InboxThread table.
icon: Inbox
order: 0
---

Blue's Inbox is a single, cross-workspace feed of everything that needs your attention: comment threads on records, workspace chat threads, AI panel conversations, and direct messages / group chats. It is a materialized `InboxThread` table, not a live union of comments and chats computed per request — a thread row is created and kept up to date as activity happens, so paginating the feed is an ordinary indexed query rather than a fan-out across several entity types.

Direct messages and group chats are new conversation kinds layered onto the existing `Chat` type (the same type that already backed workspace chat threads) — see [Direct Messages & Group Chats](/api/inbox/direct-messages-and-group-chats) for how `kind` tells them apart.

## The InboxThread type

| Field            | Type               | Description                                                                                                                |
| ---------------- | ------------------ | -------------------------------------------------------------------------------------------------------------------------- |
| `id`             | `ID!`              | Unique identifier for the thread.                                                                                          |
| `type`           | `InboxThreadType!` | Which surface this thread represents. See below.                                                                           |
| `workspace`      | `Project`          | The workspace the thread belongs to. `null` for direct messages and group chats.                                           |
| `title`          | `String!`          | The thread's display title.                                                                                                |
| `preview`        | `String`           | A short preview of the most recent activity, for list rendering.                                                           |
| `lastActivityAt` | `DateTime!`        | When the thread last had activity. Threads are ordered by this field, most recent first.                                   |
| `lastActor`      | `User`             | Who generated the most recent activity.                                                                                    |
| `isRead`         | `Boolean!`         | Whether **you** have read this thread. Evaluated per caller — see [Manage inbox threads](/api/inbox/manage-inbox-threads). |
| `isMuted`        | `Boolean!`         | Your own notification mute state for this thread.                                                                          |
| `pinned`         | `Boolean!`         | Your own "pin to top" state for this thread.                                                                               |
| `todoId`         | `String`           | Set only when `type` is `RECORD_THREAD` and the record is a `Todo`.                                                        |
| `statusUpdateId` | `String`           | Set only when `type` is `RECORD_THREAD` and the record is a `StatusUpdate`.                                                |
| `discussionId`   | `String`           | Set when `type` is `CHAT_THREAD` or `DM_THREAD` — both anchor on a `Chat`.                                           |
| `aiThreadId`     | `String`           | Set only when `type` is `AI_PANEL_THREAD`.                                                                                 |

`isRead`, `isMuted`, and `pinned` are all evaluated against **your own** participation in the thread — the same thread can be read and pinned for one person and unread for another.

## InboxThreadType

`type` tells you which underlying surface a thread represents, and which of `todoId` / `statusUpdateId` / `discussionId` / `aiThreadId` is set.

| Value             | The thread is...                                                       | Follow up with                                                                      |
| ----------------- | ---------------------------------------------------------------------- | ----------------------------------------------------------------------------------- |
| `RECORD_THREAD`   | Comments on a record (a `Todo` or a `StatusUpdate`).                   | [`commentList(category: TODO \| STATUS_UPDATE, ...)`](/api/comments/query-comments) |
| `CHAT_THREAD`     | A workspace chat topic (`Chat` with `kind: CHANNEL`).            | [Query chats](/api/comments/query-discussions)                                |
| `AI_PANEL_THREAD` | A standalone AI panel conversation.                                    | —                                                                                   |
| `DM_THREAD`       | A direct message or group chat (`Chat` with `kind: DM`/`GROUP`). | [Direct Messages & Group Chats](/api/inbox/direct-messages-and-group-chats)         |

## InboxFilter

Passed to [`inbox`](/api/inbox/query-inbox) to narrow the feed to one slice, evaluated against your own participation in each thread.

| Value             | Returns threads where...        |
| ----------------- | ------------------------------- |
| `ALL`             | No filter — the default.        |
| `UNREAD`          | You haven't read the thread.    |
| `ASSIGNED_TO_ME`  | You're the assignee.            |
| `COMMENTED_BY_ME` | You've commented in the thread. |
| `MENTIONED`       | You were @-mentioned in it.     |

## Pages in this section

| Page                                                                        | What it covers                                                                                                                            |
| --------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| [Query your Inbox](/api/inbox/query-inbox)                                  | The `inbox` feed query, `hasUnreadInboxThreads`, and `pinnedInboxThreads`.                                                                |
| [Manage inbox threads](/api/inbox/manage-inbox-threads)                     | Read/unread state, mark-all-read, per-thread mute, and pin-to-top — `markInboxThreadRead`, `setInboxThreadPinned`, and related mutations. |
| [Direct Messages & Group Chats](/api/inbox/direct-messages-and-group-chats) | `startDirectMessage`, `createGroupChat`, membership management, and `messageableUsers`.                                                   |

Live updates for the Inbox — new activity and cross-tab "mark all read" sync — are documented with the rest of the real-time API in [Inbox subscriptions](/api/realtime/inbox-subscriptions).

## Product nouns to schema types

| UI / docs word        | Schema type / field                                        |
| --------------------- | ---------------------------------------------------------- |
| Inbox thread          | `InboxThread`                                              |
| Direct message        | `Chat` with `kind: DM`                               |
| Group chat            | `Chat` with `kind: GROUP`                            |
| Workspace chat        | `Chat` with `kind: CHANNEL`                          |
| Workspace             | `Project` (`InboxThread.workspace` / `Chat.project`) |
| Group chat membership | `ChatMember`                                               |
| Record                | `Todo` or `StatusUpdate`                                   |

## Related

- [Query your Inbox](/api/inbox/query-inbox)
- [Manage inbox threads](/api/inbox/manage-inbox-threads)
- [Direct Messages & Group Chats](/api/inbox/direct-messages-and-group-chats)
- [Inbox subscriptions](/api/realtime/inbox-subscriptions)
- [Comments overview](/api/comments) — the `Comment` type behind `RECORD_THREAD` and `CHAT_THREAD` replies.
- [Query chats](/api/comments/query-discussions) — the read paths for workspace chat `Chat`s.
