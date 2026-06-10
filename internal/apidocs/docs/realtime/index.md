---
title: Real-time (Subscriptions)
description: Push updates over a WebSocket. Subscribe to records, projects, comments, activity, presence, and job progress with the same auth and permissions as the HTTP API.
icon: Radio
order: 0
---

Blue's GraphQL API exposes a large `type Subscription` root that delivers real-time push updates over a WebSocket transport ([graphql-ws](https://github.com/enisdenjo/graphql-ws)) at `wss://api.blue.app/graphql` — the same URL as the HTTP GraphQL endpoint. This section explains how to open and authenticate a subscription, then documents every operation grouped by domain: entity change-feeds (records, projects, comments, custom fields, files, activity), presence and notifications, and progress streams for long-running jobs.

Authentication uses the same credentials as HTTP requests, passed through graphql-ws [`connectionParams`](/api/realtime/connect-and-authenticate) at connection init (a Personal Access Token or a Bearer JWT). The same permission checks apply: subscribing succeeds, but you only receive events for the records and projects you are allowed to see.

## How subscriptions work

A subscription is a long-lived GraphQL operation. You open one WebSocket connection, send one or more `subscription { … }` operations over it, and the server streams a result every time a matching change happens server-side. The connection stays open until you close it or the socket drops.

Three things are consistent across the whole section:

- **One endpoint, one protocol.** All subscriptions run over graphql-ws at `wss://api.blue.app/graphql`. There is no separate subscription path. See [Connect & Authenticate](/api/realtime/connect-and-authenticate).
- **Auth at connect, in `connectionParams`.** Credentials ride in the graphql-ws connection-init payload, not in per-operation arguments. An unauthenticated handshake is rejected before any subscription starts.
- **Permission gating per event.** Each delivered event is filtered against your permissions. You can subscribe broadly (e.g. all records in a workspace) and still only receive the changes you can see.

## The entity-payload shape

Most subscriptions are _change-feeds_ for an entity (a record, project, comment, custom field, file, and so on). They share one payload shape, so once you can read one you can read them all:

```graphql
subscription OnRecordChange {
  subscribeToTodo(projectId: "project_123") {
    mutation # CREATED | UPDATED | DELETED
    node {
      # the entity in its new state (null on DELETED)
      id
      title
      done
    }
    previousValues {
      # the entity's prior state (present on UPDATED/DELETED)
      title
      done
    }
    updatedFields # [String] — names of the fields that changed on UPDATE
  }
}
```

| Field            | Type                              | Description                                                    |
| ---------------- | --------------------------------- | -------------------------------------------------------------- |
| `mutation`       | `MutationType`                    | What happened: `CREATED`, `UPDATED`, or `DELETED`.             |
| `node`           | the entity type                   | The entity in its current state. `null` for a `DELETED` event. |
| `previousValues` | the entity's previous-values type | The entity's prior field values, for diffing on update/delete. |
| `updatedFields`  | `[String!]`                       | Names of the fields that changed (present on `UPDATED`).       |

### MutationType

```graphql
enum MutationType {
  CREATED
  UPDATED
  DELETED
}
```

<Callout variant="info" title="Not every subscription uses this shape">

A handful of operations diverge — read each domain page for the exact return type:

- **Scalars and bare types:** `subscribeToMyTodoCount` returns `Int!`, `subscribeToUserPresence` and `subscribeToCommentTyping` return `User!`, and the import/export and lookup progress streams return `JSON`.
- **Mentions:** `MentionSubscriptionPayload` has a non-nullable `mutation`, a `Mention` for both `node` and `previousValues`, and no `updatedFields`.
- **Community:** community feeds use a different `CommunityMutationType` enum and a `{ mutation, node }` shape with no `previousValues`.
- **`on*` events** are imperative event channels (a record moved, a project archived, a list marked done). They return a bare entity or list, not a change-feed payload.

</Callout>

## Pages in this section

Start with the transport/auth primer, then jump to the domain you need. Each page documents its subscriptions, their filter inputs, and the exact payload type they return.

| Page                                                                              | What it covers                                                                                                                                                                                  |
| --------------------------------------------------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [Connect & Authenticate](/api/realtime/connect-and-authenticate)                  | Open an authenticated graphql-ws connection: `connectionParams`, the three credential forms, company/project scoping, and a runnable client example.                                            |
| [Record & Todo-List Subscriptions](/api/realtime/record-subscriptions)            | `subscribeToTodo`, `subscribeToTodoAction`, `subscribeToTodoList`, the `subscribeToMyTodoCount` badge counter, and the `onMoveTodo` / `onMarkTodoListAsDone` / `onMarkTodoListAsUndone` events. |
| [Project & Workspace Subscriptions](/api/realtime/project-subscriptions)          | `subscribeToProject` plus people-list, role, folder, tag, and saved-view streams, and the full set of project `on*` events (membership, archive, folder, template, copy status).                |
| [Comments, Discussions & Chat](/api/realtime/comment-discussion-subscriptions)    | `subscribeToComment`, the `subscribeToCommentTyping` typing indicator, `subscribeToDiscussion`, `subscribeToStatusUpdate`, `subscribeToChat`, and `subscribeToDocument`.                        |
| [Custom Field Subscriptions](/api/realtime/custom-field-subscriptions)            | `subscribeToCustomField`, `subscribeToCustomFieldOption`, and the `onCustomFieldOptionsCreated` batch event.                                                                                    |
| [File & Link Subscriptions](/api/realtime/file-subscriptions)                     | `subscribeToFile`, `subscribeToLink`, and the `onDeleteFiles` batch deletion event.                                                                                                             |
| [Activity, Mentions & Notifications](/api/realtime/activity-notifications)        | `subscribeToActivity`, `subscribeToMyMention`, and the `onMarkAllActivityAsSeen` / `onMarkAllMentionsAsRead` events.                                                                            |
| [User Presence](/api/realtime/presence-subscriptions)                             | `subscribeToUserPresence` (online/offline driven by the connection lifecycle), `subscribeToCompanyPeopleList`, `subscribeToCompany`, and `subscribeToInvitation`.                               |
| [Progress & Long-Running Jobs](/api/realtime/progress-subscriptions)              | Job-status channels: import/export and lookup progress, AI auto-tagging, cover-image generation, automation config and execution runs, and OAuth connection status.                             |
| [Dashboards, Charts & Community](/api/realtime/analytics-community-subscriptions) | `subscribeToDashboard`, `subscribeToChart`, `subscribeToForm`, and the community feeds (`subscribeToCommunityPost`, `subscribeToCommunityPosts`).                                               |

## Product nouns to schema types

Subscription names use the schema type names, not the UI words. The mapping:

| UI / docs word         | Schema type   |
| ---------------------- | ------------- |
| Record                 | `Todo`        |
| Workspace / Project    | `Project`     |
| Organization / Company | `Company`     |
| List                   | `TodoList`    |
| Custom field           | `CustomField` |
| Saved view             | `SavedView`   |
| Dashboard              | `Dashboard`   |
| @-mention              | `Mention`     |

## Deprecated subscriptions

Three subscriptions remain in the schema for backward compatibility but should not be used in new code. Each is omitted from this reference in favor of its replacement:

| Deprecated                    | Use instead                                                       |
| ----------------------------- | ----------------------------------------------------------------- |
| `subscribeToMention`          | [`subscribeToMyMention`](/api/realtime/activity-notifications)    |
| `subscribeToNetworkStatus`    | [`subscribeToUserPresence`](/api/realtime/presence-subscriptions) |
| `subscribeToRecentActiveUser` | — (never implemented; has no publisher — do not use)              |
