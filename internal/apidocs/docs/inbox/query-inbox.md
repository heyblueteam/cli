---
title: Query your Inbox
description: Page through the unified Inbox feed with the inbox query, check hasUnreadInboxThreads for the sidebar dot, and list pinned threads.
icon: Search
order: 1
---

Three read paths cover the Inbox: `inbox` for the paginated, filterable feed; `hasUnreadInboxThreads` for a cheap existence check to drive an unread indicator; and `pinnedInboxThreads` for the threads you've pinned to the top. See [Inbox & Messaging overview](/api/inbox) for the shared `InboxThread` type and its `type`/`InboxFilter` enums.

## `inbox`

The paginated Inbox feed for the current organization, ordered by `lastActivityAt` descending (most recent first). Uses keyset (cursor) pagination.

### Request

```graphql
query MyInbox {
  inbox(filter: UNREAD, first: 20) {
    items {
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
    pageInfo {
      hasNextPage
      nextCursor
    }
  }
}
```

### Parameters

| Parameter | Type              | Required | Default | Description                                                                   |
| --------- | ----------------- | -------- | ------- | ----------------------------------------------------------------------------- |
| `filter`  | `InboxFilter`     | No       | `ALL`   | Narrow the feed to one slice. See [InboxFilter](/api/inbox#inboxfilter).      |
| `type`    | `InboxThreadType` | No       | —       | Restrict to one thread type (e.g. only `DM_THREAD`). Omit for all types.      |
| `after`   | `String`          | No       | —       | Cursor to page forward from — pass the previous page's `pageInfo.nextCursor`. |
| `first`   | `Int`             | No       | `20`    | Maximum number of threads to return on this page.                             |

<Callout variant="info" title="Pinned threads are excluded here">

A thread you've pinned (see [`setInboxThreadPinned`](/api/inbox/manage-inbox-threads#setinboxthreadpinned)) is left out of the `inbox` feed entirely — it surfaces only through [`pinnedInboxThreads`](#pinnedinboxthreads), so it isn't rendered twice.

</Callout>

### Response

```json
{
  "data": {
    "inbox": {
      "items": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "type": "DM_THREAD",
          "title": "Priya Shah",
          "preview": "Priya: Quick one, can you approve the vendor list?",
          "isRead": false,
          "lastActivityAt": "2026-07-15T09:41:02.000Z",
          "lastActor": { "id": "clm4n8qwx000108l0a1b2c3d4", "fullName": "Priya Shah" }
        }
      ],
      "pageInfo": {
        "hasNextPage": true,
        "nextCursor": "clm4n8qwx000008l0g4oxdqn7"
      }
    }
  }
}
```

#### Returns — InboxThreadList

| Field      | Type              | Description                                                                                                                                   |
| ---------- | ----------------- | --------------------------------------------------------------------------------------------------------------------------------------------- |
| `items`    | `[InboxThread!]!` | The threads on this page.                                                                                                                     |
| `pageInfo` | `PageInfo!`       | Cursor pagination metadata. There is no `totalCount` — an exact total would reintroduce the full-scan cost cursor pagination exists to avoid. |

To fetch the next page, pass the current page's `pageInfo.nextCursor` as the next call's `after`, and stop when `pageInfo.hasNextPage` is `false`.

<Callout variant="warning" title="A page can come back shorter than `first`">

Visibility filtering (for example, `showOnlyMentionedComments`-style role restrictions on a workspace, or a thread whose workspace was since trashed) is applied **after** the page is fetched from the database. A page can legitimately return fewer than `first` items — including zero — even when `hasNextPage` is `true`. Keep paging on `nextCursor` rather than assuming a short page means you've reached the end.

</Callout>

## `hasUnreadInboxThreads`

A cheap existence check — whether **any** thread in your Inbox is unread. Use this to drive a sidebar unread dot without paginating the full feed.

### Request

```graphql
query InboxUnreadDot {
  hasUnreadInboxThreads
}
```

`hasUnreadInboxThreads` takes no arguments and returns a bare `Boolean!`.

### Response

```json
{
  "data": {
    "hasUnreadInboxThreads": true
  }
}
```

## `pinnedInboxThreads`

Your pinned threads for the current organization, unpaginated (capped at 20 — see [`setInboxThreadPinned`](/api/inbox/manage-inbox-threads#setinboxthreadpinned)), most recently pinned first.

### Request

```graphql
query MyPinnedThreads {
  pinnedInboxThreads(type: DM_THREAD) {
    id
    type
    title
    preview
    lastActivityAt
  }
}
```

### Parameters

| Parameter | Type              | Required | Description                                      |
| --------- | ----------------- | -------- | ------------------------------------------------ |
| `type`    | `InboxThreadType` | No       | Restrict to one thread type. Omit for all types. |

### Response

```json
{
  "data": {
    "pinnedInboxThreads": [
      {
        "id": "clm4n8qwx000208l0e5f6g7h8",
        "type": "DM_THREAD",
        "title": "Launch crew",
        "preview": "Sarah: Works for me, I'll bring the customer feedback doc",
        "lastActivityAt": "2026-07-15T08:12:44.000Z"
      }
    ]
  }
}
```

`pinnedInboxThreads` returns `[InboxThread!]!` — see [Inbox & Messaging overview](/api/inbox#the-inboxthread-type) for the field list.

<Callout variant="info" title="Pin display is per-organization; the 20-item cap is not">

`pinnedInboxThreads` only shows pins that belong to the organization you're currently in. The 20-pin cap enforced by `setInboxThreadPinned`, however, counts pins **across every organization you belong to** — see the callout on [Manage inbox threads](/api/inbox/manage-inbox-threads#setinboxthreadpinned) before assuming an empty list here means you have room to pin more.

</Callout>

## Errors

None of the three queries throw beyond the standard authentication/authorization errors below — an unrecognized or omitted filter simply yields the default (`ALL` / all types).

| Code                | When                                                                    |
| ------------------- | ----------------------------------------------------------------------- |
| `UNAUTHENTICATED`   | No valid token was supplied.                                            |
| `COMPANY_NOT_FOUND` | No active organization for the request (missing/invalid `blue-org-id`). |
| `PAYMENT_REQUIRED`  | The organization has no active license or subscription.                 |

## Permissions

All three require authentication and an active organization membership (`isAuth`, `isAuthInCompany`, `isActiveCompany`) — the same scoping as the record queries.

## Related

- [Inbox & Messaging overview](/api/inbox) — the `InboxThread` type and its enums.
- [Manage inbox threads](/api/inbox/manage-inbox-threads) — mark read/unread, mute, and pin.
- [Direct Messages & Group Chats](/api/inbox/direct-messages-and-group-chats) — start and manage DMs/group chats.
- [Inbox subscriptions](/api/realtime/inbox-subscriptions) — stream new Inbox activity in real time.
