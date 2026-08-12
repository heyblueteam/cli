---
title: Manage inbox threads
description: Mark threads read or unread, clear your whole Inbox at once, mute notifications on a thread, and pin threads to the top with markInboxThreadRead, markAllInboxThreadsRead, setInboxThreadMuted, and setInboxThreadPinned.
icon: CheckCheck
order: 2
---

These five mutations manage your own relationship to a thread — read state, notification mute, and pin — without touching the thread's content. Every one of them writes to **your own** participation record for the thread; nobody else's read/mute/pin state is affected.

## markInboxThreadRead / markInboxThreadUnread

Mark a single thread read or unread for the caller.

### Request

```graphql
mutation MarkThreadRead {
  markInboxThreadRead(threadId: "thread_123") {
    id
    isRead
  }
}
```

```graphql
mutation MarkThreadUnread {
  markInboxThreadUnread(threadId: "thread_123") {
    id
    isRead
  }
}
```

### Parameters

| Parameter  | Type  | Required | Description                  |
| ---------- | ----- | -------- | ---------------------------- |
| `threadId` | `ID!` | Yes      | The `InboxThread` to update. |

### Response

```json
{
  "data": {
    "markInboxThreadRead": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "isRead": true
    }
  }
}
```

#### Returns

Both mutations return the updated [`InboxThread`](/api/inbox#the-inboxthread-type).

### Errors

| Code              | When                                                                                                                                                                                         |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `FORBIDDEN`       | No thread matches `threadId` in your current organization, **or** you have no participant row on it. Both cases return the same error — there is no separate "not found" code for this pair. |
| `UNAUTHENTICATED` | No valid token was supplied.                                                                                                                                                                 |

## markAllInboxThreadsRead

Marks every unread thread in your Inbox **for the current organization** as read in one call. Always succeeds for an authenticated, in-organization caller — it's a bulk update, not a per-thread lookup, so there's nothing to 404 on.

### Request

```graphql
mutation ClearInbox {
  markAllInboxThreadsRead
}
```

`markAllInboxThreadsRead` takes no arguments and returns a bare `Boolean!`.

### Response

```json
{
  "data": {
    "markAllInboxThreadsRead": true
  }
}
```

<Callout variant="info" title="Scoped to the current organization">

Unlike the older `markAllMentionsAsRead`, this clears unread state only within the organization you're currently in — threads in your other organizations are untouched. A connected client (another tab, another device) is notified over [`onMarkAllInboxThreadsRead`](/api/realtime/inbox-subscriptions#onmarkallinboxthreadsread) so it can clear its own badge without refetching.

</Callout>

## setInboxThreadMuted

Mute or unmute push/email notifications for one thread, without changing its read state or hiding it from your feed.

### Request

```graphql
mutation MuteThread {
  setInboxThreadMuted(inboxThreadId: "thread_123", muted: true)
}
```

### Parameters

| Parameter       | Type       | Required | Description                        |
| --------------- | ---------- | -------- | ---------------------------------- |
| `inboxThreadId` | `ID!`      | Yes      | The `InboxThread` to mute/unmute.  |
| `muted`         | `Boolean!` | Yes      | `true` to mute, `false` to unmute. |

`setInboxThreadMuted` returns a bare `Boolean!` (the new muted state), not the thread.

### Response

```json
{
  "data": {
    "setInboxThreadMuted": true
  }
}
```

<Callout variant="info" title="Silences notifications only">

Muting never hides the thread from your Inbox feed and never changes `isRead` — it only suppresses push/email delivery for that thread going forward.

</Callout>

### Errors

| Code              | When                                                                                                                                                                               |
| ----------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `BAD_USER_INPUT`  | You have no participant row on `inboxThreadId` (message: `Thread not found.`) — note this differs from the `FORBIDDEN` error `markInboxThreadRead` throws for the equivalent case. |
| `UNAUTHENTICATED` | No valid token was supplied.                                                                                                                                                       |

## setInboxThreadPinned

Pin or unpin a thread to keep it at the top of your Inbox, outside the normal `lastActivityAt` ordering.

### Request

```graphql
mutation PinThread {
  setInboxThreadPinned(inboxThreadId: "thread_123", pinned: true)
}
```

### Parameters

| Parameter       | Type       | Required | Description                      |
| --------------- | ---------- | -------- | -------------------------------- |
| `inboxThreadId` | `ID!`      | Yes      | The `InboxThread` to pin/unpin.  |
| `pinned`        | `Boolean!` | Yes      | `true` to pin, `false` to unpin. |

`setInboxThreadPinned` returns a bare `Boolean!` (the new pinned state).

### Response

```json
{
  "data": {
    "setInboxThreadPinned": true
  }
}
```

<Callout variant="warning" title="Cap of 20 pins, counted across every organization you belong to">

You can pin at most 20 threads. Pinning a 21st throws rather than silently evicting an older pin — unpin one first. The cap is enforced **globally across all organizations you're a member of**, not per-organization: if you've already pinned 20 threads in one organization, pinning anything in a different one also fails, even though [`pinnedInboxThreads`](/api/inbox/query-inbox#pinnedinboxthreads) in that other organization looks empty. Unpinning and re-pinning the same thread, or pinning an already-pinned thread, never counts against the cap.

</Callout>

### Errors

| Code              | When                                                                                                                                                                                     |
| ----------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `BAD_USER_INPUT`  | You have no participant row on `inboxThreadId` (message: `Thread not found.`), or pinning would exceed the 20-pin cap (message: `You can only pin up to 20 threads — unpin one first.`). |
| `UNAUTHENTICATED` | No valid token was supplied.                                                                                                                                                             |

## Permissions

| Mutation                  | Requires                                                                                          |
| ------------------------- | ------------------------------------------------------------------------------------------------- |
| `markInboxThreadRead`     | Authenticated, active organization membership.                                                    |
| `markInboxThreadUnread`   | Authenticated, active organization membership.                                                    |
| `markAllInboxThreadsRead` | Authenticated, active organization membership.                                                    |
| `setInboxThreadMuted`     | Authenticated. (Not scoped to a specific organization — works for any thread you participate in.) |
| `setInboxThreadPinned`    | Authenticated. (Same as above.)                                                                   |

None of these five require the organization's license to be active — reducing your own notification load or catching up on unread threads is always allowed.

## Related

- [Inbox & Messaging overview](/api/inbox) — the `InboxThread` type.
- [Query your Inbox](/api/inbox/query-inbox) — the `inbox`, `hasUnreadInboxThreads`, and `pinnedInboxThreads` queries these mutations feed into.
- [Inbox subscriptions](/api/realtime/inbox-subscriptions) — `onMarkAllInboxThreadsRead` for cross-tab sync.
- [Direct Messages & Group Chats](/api/inbox/direct-messages-and-group-chats)
