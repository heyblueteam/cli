---
title: Activity, Mentions & Notifications
description: Stream the activity feed, the user's @-mentions, and the mark-seen / mark-read events that drive notification badges.
icon: Bell
order: 7
---

These are the notification streams that keep an inbox, a notification bell, and an activity feed live. This page covers four subscriptions: `subscribeToActivity` (the workspace and organization activity feed), `subscribeToMyMention` (the signed-in user's `@`-mentions), and the two mark-as-read event channels `onMarkAllActivityAsSeen` and `onMarkAllMentionsAsRead` that fire when the user clears their unread counts.

All four run over the same authenticated WebSocket as every other subscription — see [Connect & Authenticate](/api/realtime/connect-and-authenticate) for the `graphql-ws` handshake and credentials. The endpoint is `wss://api.blue.app/graphql`.

<Callout variant="warning" title="subscribeToMention is deprecated">

Use `subscribeToMyMention` for the signed-in user's mentions. The older `subscribeToMention(filter: MentionFilter!)` field is deprecated and returns the legacy `MentionSubscriptionPayload_DEPRECATED` type — do not use it in new integrations.

</Callout>

## subscribeToActivity

Streams the activity feed for an organization or a workspace — every action that produces an activity entry (a record created, a comment posted, a user added to a workspace, an invitation accepted, and so on). Organizations are `Company` objects and workspaces are `Project` objects in the API, so the feed is scoped by `companyId`, `projectId`, or both.

This is an entity change-feed: each event is an `ActivitySubscriptionPayload` with the standard `{ mutation, node, previousValues, updatedFields }` shape described on the [Real-time overview](/api/realtime).

### Request

Scope the feed to a single workspace:

```graphql
subscription OnActivity {
  subscribeToActivity(projectId: "project_123") {
    mutation
    node {
      id
      category
      html
      isSeen
      isRead
      createdAt
      createdBy {
        id
        fullName
      }
    }
    updatedFields
  }
}
```

To watch all activity across an organization, pass `companyId` instead; pass both to scope to a workspace within a specific organization. `companyId` and `projectId` each accept an ID or a slug.

### Parameters

| Parameter   | Type     | Required | Description                                                                                                                                       |
| ----------- | -------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| `companyId` | `String` | No       | Organization whose activity feed to watch. Accepts an ID or a slug.                                                                               |
| `projectId` | `String` | No       | Workspace whose activity feed to watch. Accepts an ID or a slug. Combine with `companyId` to scope to a workspace inside a specific organization. |

### Response

Each delivered event is an `ActivitySubscriptionPayload`. The example below shows a new comment activity (`mutation: CREATED`).

```json
{
  "data": {
    "subscribeToActivity": {
      "mutation": "CREATED",
      "node": {
        "id": "clm4n8qwx000008l0g4oxdqn7",
        "category": "CREATE_COMMENT",
        "html": "<p>Ada commented on <strong>Launch checklist</strong></p>",
        "isSeen": false,
        "isRead": false,
        "createdAt": "2026-05-29T14:02:11.000Z",
        "createdBy": {
          "id": "clm4n8qwx000108l0h2pyerm8",
          "fullName": "Ada Lovelace"
        }
      },
      "updatedFields": null
    }
  }
}
```

#### ActivitySubscriptionPayload

| Field            | Type                     | Description                                                              |
| ---------------- | ------------------------ | ------------------------------------------------------------------------ |
| `mutation`       | `MutationType!`          | `CREATED`, `UPDATED`, or `DELETED`.                                      |
| `node`           | `Activity`               | The activity entry in its current state. `null` on `DELETED`.            |
| `updatedFields`  | `[String!]`              | Names of the properties that changed on an `UPDATED` event.              |
| `previousValues` | `ActivityPreviousValues` | Snapshot of the entry before the change. Present on `UPDATED`/`DELETED`. |

#### Activity

`node` is the full `Activity` type. Select what your feed renders:

| Field          | Type                | Description                                                                 |
| -------------- | ------------------- | --------------------------------------------------------------------------- |
| `id`           | `ID!`               | The activity entry ID.                                                      |
| `category`     | `ActivityCategory!` | What kind of action produced the entry (see below).                         |
| `html`         | `String!`           | The rendered activity message. Prefer this over the deprecated `text`.      |
| `isSeen`       | `Boolean!`          | Whether the entry has been marked seen (drives the unread dot on the bell). |
| `isRead`       | `Boolean!`          | Whether the entry has been opened/read.                                     |
| `createdAt`    | `DateTime!`         | When the action happened.                                                   |
| `createdBy`    | `User!`             | The user who performed the action. Select `id`, `fullName`, `email`.        |
| `affectedBy`   | `User`              | The user the action was performed on, when applicable.                      |
| `company`      | `Company`           | The organization the entry belongs to.                                      |
| `project`      | `Project`           | The workspace the entry belongs to, when scoped to one.                     |
| `todo`         | `Todo`              | The record the entry references, when applicable.                           |
| `todoList`     | `TodoList`          | The list the entry references, when applicable.                             |
| `comment`      | `Comment`           | The comment the entry references, for comment activity.                     |
| `discussion`   | `Discussion`        | The discussion the entry references.                                        |
| `statusUpdate` | `StatusUpdate`      | The status update the entry references.                                     |
| `inviteeEmail` | `String`            | The invited address, for invitation activity.                               |
| `metadata`     | `String`            | Extra context for the entry, as a JSON string.                              |

`previousValues` is the narrower `ActivityPreviousValues` (`id`, `uid`, `category`, `createdAt`, `updatedAt`, `inviteeEmail`, `metadata`, `userAccessLevel`).

#### ActivityCategory

The `category` tells you what happened. The enum values are:

```graphql
enum ActivityCategory {
  CREATE_TODO
  REMOVE_TODO
  COPY_TODO
  MOVE_TODO
  CREATE_TODO_LIST
  REMOVE_TODO_LIST
  CREATE_COMMENT
  CREATE_DISCUSSION
  CREATE_STATUS_UPDATE
  EDIT_NAME
  EDIT_JOB_TITLE
  CREATE_INVITATION
  CANCEL_INVITATION
  ACCEPT_INVITATION
  REJECT_INVITATION
  ADD_USER_TO_PROJECT
  REMOVE_USER_FROM_PROJECT
  REMOVE_USER_FROM_COMPANY
  LEAVE_COMPANY
  LEAVE_PROJECT
  ARCHIVE_PROJECT
  # …and more — see the schema for the complete set.
}
```

## subscribeToMyMention

Streams the `@`-mentions addressed to the **signed-in user** — fired when someone mentions you in a comment, discussion, or status update, and when one of your mentions is updated or removed. There are no arguments: the stream is implicitly scoped to whoever owns the WebSocket connection's credentials. This is what drives the mentions inbox and the mention count badge.

### Request

```graphql
subscription OnMyMention {
  subscribeToMyMention {
    mutation
    node {
      id
      isRead
      createdAt
      mentioner {
        id
        fullName
      }
      ref {
        ... on Comment {
          id
          text
        }
        ... on Discussion {
          id
          title
        }
        ... on StatusUpdate {
          id
        }
      }
      targetWorkspace {
        id
        name
      }
    }
  }
}
```

### Parameters

`subscribeToMyMention` takes no arguments.

### Response

Each event is a `MentionSubscriptionPayload`. Unlike the activity feed, this payload has **no `updatedFields`**, and `node` / `previousValues` are both the `Mention` type.

```json
{
  "data": {
    "subscribeToMyMention": {
      "mutation": "CREATED",
      "node": {
        "id": "clm4n8qwx000208l0i9qzfsn9",
        "isRead": false,
        "createdAt": "2026-05-29T14:05:42.000Z",
        "mentioner": {
          "id": "clm4n8qwx000108l0h2pyerm8",
          "fullName": "Ada Lovelace"
        },
        "ref": {
          "id": "clm4n8qwx000308l0jkr0gtoa",
          "text": "@you can you review the launch checklist?"
        },
        "targetWorkspace": {
          "id": "clm4n8qwx000408l0lst1huob",
          "name": "Product"
        }
      }
    }
  }
}
```

#### MentionSubscriptionPayload

| Field            | Type            | Description                                                    |
| ---------------- | --------------- | -------------------------------------------------------------- |
| `mutation`       | `MutationType!` | `CREATED`, `UPDATED`, or `DELETED`.                            |
| `node`           | `Mention`       | The mention in its current state. `null` on `DELETED`.         |
| `previousValues` | `Mention`       | The mention's prior state. Present on `UPDATED` and `DELETED`. |

#### Mention

| Field             | Type         | Description                                                                                                |
| ----------------- | ------------ | ---------------------------------------------------------------------------------------------------------- |
| `id`              | `ID!`        | The mention ID.                                                                                            |
| `createdAt`       | `DateTime!`  | When the mention was created.                                                                              |
| `isRead`          | `Boolean`    | Whether the mention has been read.                                                                         |
| `mentioner`       | `User`       | The user who wrote the mention. Select `id`, `fullName`, `email`.                                          |
| `mentionee`       | `User`       | The mentioned user — always the signed-in user on this stream.                                             |
| `ref`             | `MentionRef` | What the mention points at: a `Comment`, `Discussion`, or `StatusUpdate` (a union — use inline fragments). |
| `targetWorkspace` | `Project`    | The workspace to open when the user clicks the mention. `null` only when the reference is gone.            |

`ref` is the `MentionRef` union (`Comment | Discussion | StatusUpdate`), so select its fields with inline fragments as shown in the request above.

## onMarkAllMentionsAsRead

Fires when the signed-in user clears their mentions — the "mark all as read" action. The event delivers the list of `Mention` objects that were just marked read, so a connected client can update its inbox and reset the mention badge without refetching. Use it to keep a second open tab or device in sync.

### Request

```graphql
subscription OnMarkAllMentionsAsRead {
  onMarkAllMentionsAsRead {
    id
    isRead
  }
}
```

### Parameters

`onMarkAllMentionsAsRead` takes no arguments — it is scoped to the connection's user.

### Response

Each event is `[Mention!]!` — the array of mentions that were marked read.

```json
{
  "data": {
    "onMarkAllMentionsAsRead": [
      {
        "id": "clm4n8qwx000208l0i9qzfsn9",
        "isRead": true
      },
      {
        "id": "clm4n8qwx000508l0mtu2ivpc",
        "isRead": true
      }
    ]
  }
}
```

## onMarkAllActivityAsSeen

Fires when a user marks their whole activity feed as seen — the action that clears the unread dot on the activity bell. The event is a simple signal: it resolves to `true` whenever the named user's feed is marked seen, so a connected client knows to reset its unread indicator. Scoped by `userId`.

### Request

```graphql
subscription OnMarkAllActivityAsSeen {
  onMarkAllActivityAsSeen(userId: "user_123")
}
```

### Parameters

| Parameter | Type      | Required | Description                                                 |
| --------- | --------- | -------- | ----------------------------------------------------------- |
| `userId`  | `String!` | Yes      | The user whose "mark all activity as seen" events to watch. |

`onMarkAllActivityAsSeen` returns the scalar `Boolean` — it takes no sub-selection. Every delivered event is the value `true`.

### Response

```json
{
  "data": {
    "onMarkAllActivityAsSeen": true
  }
}
```

## Errors

Subscriptions report a rejected connection or an invalid argument when the operation is set up. Permission scoping is otherwise silent — see Permissions below.

| Code              | When                                                                                                                                                                                   |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `UNAUTHENTICATED` | The WebSocket handshake carried no valid credentials. The connection is rejected before any subscription starts. See [Connect & Authenticate](/api/realtime/connect-and-authenticate). |
| `BAD_USER_INPUT`  | A required argument is missing or malformed — e.g. `onMarkAllActivityAsSeen` called without `userId`.                                                                                  |

## Permissions

Events are filtered per delivered event, not just at connection time. Subscribing always succeeds for an authenticated client, but you only receive events you are allowed to see:

- **`subscribeToActivity`** delivers an event only when it matches the `companyId` / `projectId` you named **and** you are a member of that workspace or organization. Activity in workspaces you can't access is never delivered, and an extra field-permission filter drops entries you aren't allowed to see even within a workspace you belong to.
- **`subscribeToMyMention`** delivers only mentions where you are the mentionee (the `node` or `previousValues` `mentionee` is the connection's user). You never receive other users' mentions.
- **`onMarkAllMentionsAsRead`** delivers an event only when every mention in the batch belongs to you, so it fires for your own "mark all read" action.
- **`onMarkAllActivityAsSeen`** delivers an event only when the named `userId` is a member of the workspace where the feed was marked seen — pass your own user ID to watch your own feed.

## Related

- [Real-time (Subscriptions)](/api/realtime) — the shared payload shape and the full subscription catalog.
- [Connect & Authenticate](/api/realtime/connect-and-authenticate) — open an authenticated WebSocket.
- [Comments, Discussions & Chat](/api/realtime/comment-discussion-subscriptions) — the streams that generate most mention and activity events.
- [Record & Todo-List Subscriptions](/api/realtime/record-subscriptions) — record-level change feeds that pair with the activity feed.
