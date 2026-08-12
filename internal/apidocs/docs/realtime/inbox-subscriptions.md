---
title: Inbox subscriptions
description: Stream new Inbox activity in real time and sync "mark all read" across a user's open tabs and devices with subscribeToMyInboxActivity and onMarkAllInboxThreadsRead.
icon: Inbox
order: 8
---

Two subscriptions keep the Inbox live: `subscribeToMyInboxActivity` streams updated `InboxThread` rows as new activity happens, and `onMarkAllInboxThreadsRead` notifies a user's other connections when they clear their Inbox with [`markAllInboxThreadsRead`](/api/inbox/manage-inbox-threads#markallinboxthreadsread). Both run over the same authenticated WebSocket as every other subscription — see [Connect & Authenticate](/api/realtime/connect-and-authenticate). The endpoint is `wss://api.blue.app/graphql`.

## subscribeToMyInboxActivity

Streams the signed-in user's Inbox: an event fires whenever one of your `InboxThread` rows changes — a new comment, a new direct message, a new AI response, and so on. Unlike most entity subscriptions, this isn't a `{ mutation, node, previousValues }` change-feed — each event is the full `InboxThread` in its current state, since fields like `isRead` are evaluated per caller and the underlying publish doesn't carry that.

### Request

```graphql
subscription OnMyInboxActivity {
  subscribeToMyInboxActivity(companyId: "company_123") {
    id
    type
    title
    preview
    isRead
    lastActivityAt
    lastActor {
      id
      fullName
    }
  }
}
```

### Parameters

| Parameter   | Type      | Required | Description                                     |
| ----------- | --------- | -------- | ----------------------------------------------- |
| `companyId` | `String!` | Yes      | The organization whose Inbox activity to watch. |

### Response

```json
{
  "data": {
    "subscribeToMyInboxActivity": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "type": "DM_THREAD",
      "title": "Priya Shah",
      "preview": "Priya: Quick one, can you approve the vendor list?",
      "isRead": false,
      "lastActivityAt": "2026-07-15T09:41:02.000Z",
      "lastActor": { "id": "clm4n8qwx000108l0a1b2c3d4", "fullName": "Priya Shah" }
    }
  }
}
```

Each event is a full [`InboxThread`](/api/inbox#the-inboxthread-type) — select whatever fields your feed renders.

### Errors

| Code              | When                                                  |
| ----------------- | ----------------------------------------------------- |
| `UNAUTHENTICATED` | The WebSocket handshake carried no valid credentials. |

### Permissions

Delivered only for threads you participate in, within the organization you named. The confidentiality restrictions that apply to [`inbox`](/api/inbox/query-inbox) (for example, role-based visibility on a workspace) are re-checked per event, not just once at subscribe time.

## onMarkAllInboxThreadsRead

Fires when the signed-in user calls [`markAllInboxThreadsRead`](/api/inbox/manage-inbox-threads#markallinboxthreadsread) — the "clear my Inbox" action. This is a signal, not a change-feed: it resolves to `true` so a connected client (another tab, another device) knows to reset its own unread state without refetching.

### Request

```graphql
subscription OnInboxCleared {
  onMarkAllInboxThreadsRead(companyId: "company_123")
}
```

### Parameters

| Parameter   | Type      | Required | Description                                             |
| ----------- | --------- | -------- | ------------------------------------------------------- |
| `companyId` | `String!` | Yes      | The organization whose "mark all read" events to watch. |

`onMarkAllInboxThreadsRead` returns a bare `Boolean!` — every delivered event is `true`.

### Response

```json
{
  "data": {
    "onMarkAllInboxThreadsRead": true
  }
}
```

### Errors

| Code              | When                                                  |
| ----------------- | ----------------------------------------------------- |
| `UNAUTHENTICATED` | The WebSocket handshake carried no valid credentials. |

### Permissions

Delivered only when the event's `userId` matches your own connection and the event's `companyId` matches the one you named — you only ever receive your own "mark all read" events, never another user's.

## Related

- [Real-time (Subscriptions)](/api/realtime) — the shared change-feed payload shape and the full subscription catalog.
- [Connect & Authenticate](/api/realtime/connect-and-authenticate) — open an authenticated `graphql-ws` connection.
- [Inbox & Messaging overview](/api/inbox) — the `InboxThread` type and its enums.
- [Query your Inbox](/api/inbox/query-inbox) — the polling equivalent of this stream.
- [Manage inbox threads](/api/inbox/manage-inbox-threads) — the mutation behind `onMarkAllInboxThreadsRead`.
- [Comments, Chats & Chat](/api/realtime/comment-discussion-subscriptions) — the comment/chat streams that generate most Inbox activity.
- [Activity, Mentions & Notifications](/api/realtime/activity-notifications) — the separate @-mention and activity-feed subscriptions.
