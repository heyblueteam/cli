---
title: Record & Todo-List Subscriptions
description: Real-time GraphQL subscriptions for records and lists — change feeds, action history, the badge counter, and move/done events.
icon: ListChecks
order: 2
---

Stream live changes to records and lists over a WebSocket. Records are `Todo` objects in the API and lists are `TodoList` objects; this page covers the seven subscription fields that report on them: the `subscribeToTodo` change feed, the `subscribeToTodoAction` history feed, `subscribeToTodoList`, the scalar `subscribeToMyTodoCount` badge counter, and the imperative `onMoveTodo`, `onMarkTodoListAsDone`, and `onMarkTodoListAsUndone` event streams.

All subscriptions run over the same graphql-ws WebSocket as the rest of the API. Open and authenticate the connection first — see [Connect & Authenticate](/api/realtime/connect-and-authenticate) — then send any of the operations below as a `subscribe` message. Every example here assumes that authenticated socket is already open at `wss://api.blue.app/graphql`.

<Callout variant="info" title="Permissions are checked per event">

Subscribing always succeeds. You only receive events for records and lists you can actually see — each delivered `subscribeToTodo` event is gated by `permissions.todo(id).view()`, and the list/action/move streams require you to be a member of the workspace the event belongs to. See [Permissions](#permissions).

</Callout>

## subscribeToTodo

The primary record change feed. Streams a `TodoSubscriptionPayload` every time a record is created, updated, or deleted within your scope. This is what keeps a board, list, or record view live as records move, get reassigned, change due dates, or are archived.

### Request

Scope the stream with the `filter` argument. `companyId` is the floor; add `projectId` to narrow to a single workspace. Company and project accept either an ID or a slug.

```graphql
subscription OnRecordChange {
  subscribeToTodo(filter: { companyId: "company_123", projectId: "project_123" }) {
    mutation
    node {
      id
      title
      done
      duedAt
      position
      users {
        id
        fullName
      }
      todoList {
        id
        title
      }
    }
    updatedFields
    previousValues {
      title
      done
      duedAt
    }
  }
}
```

<Callout variant="info" title="filter vs. the top-level id / projectId args">

`subscribeToTodo` also exposes top-level `id` and `projectId` arguments, but the resolver scopes delivery from the `filter` object — pass `companyId` (required for any event to match) and the optional `projectId` / date range there. Use `filter` in new integrations.

</Callout>

### Parameters

| Argument    | Type                    | Required | Description                                                                     |
| ----------- | ----------------------- | -------- | ------------------------------------------------------------------------------- |
| `filter`    | `SubscribeToTodoFilter` | No       | Scoping object. Without a matching `companyId`, no events are delivered.        |
| `id`        | `String`                | No       | Record (`Todo`) ID. Top-level convenience arg.                                  |
| `projectId` | `String`                | No       | Workspace (`Project`) ID. Top-level convenience arg; prefer `filter.projectId`. |

#### SubscribeToTodoFilter

| Field       | Type       | Required | Description                                                                                                                                                                       |
| ----------- | ---------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `companyId` | `String`   | No       | Organization ID or slug. The event's company must match this for any event to be delivered.                                                                                       |
| `projectId` | `String`   | No       | Workspace ID or slug. When set, only events from that workspace are delivered. When omitted, you receive events for every record in the company that you have permission to view. |
| `todoId`    | `String`   | No       | Restrict to a single record.                                                                                                                                                      |
| `userId`    | `String`   | No       | Restrict to events directed at a specific user.                                                                                                                                   |
| `startedAt` | `DateTime` | No       | Lower bound of a date-range filter. Combine with `duedAt` to deliver only records whose start/due window overlaps the range.                                                      |
| `duedAt`    | `DateTime` | No       | Upper bound of the date-range filter. Must be paired with `startedAt` to take effect.                                                                                             |

### Response

Each event is a `TodoSubscriptionPayload`. `mutation` tells you what happened; `node` carries the current record (`null` for a `DELETED` event); `previousValues` carries the pre-change scalar values; `updatedFields` lists the record fields that changed on an `UPDATED` event.

```json
{
  "data": {
    "subscribeToTodo": {
      "mutation": "UPDATED",
      "node": {
        "id": "clm4n8qwx000008l0g4oxdqn7",
        "title": "Draft Q3 launch plan",
        "done": false,
        "duedAt": "2026-06-15T17:00:00.000Z",
        "position": 65536,
        "users": [{ "id": "clm4n8usr000108l0a1b2c3d4", "fullName": "Ada Lovelace" }],
        "todoList": { "id": "clm4n8lst000208l0e5f6g7h8", "title": "In Progress" }
      },
      "updatedFields": ["duedAt"],
      "previousValues": {
        "title": "Draft Q3 launch plan",
        "done": false,
        "duedAt": "2026-06-12T17:00:00.000Z"
      }
    }
  }
}
```

#### TodoSubscriptionPayload

| Field            | Type                 | Description                                                                |
| ---------------- | -------------------- | -------------------------------------------------------------------------- |
| `mutation`       | `MutationType!`      | One of `CREATED`, `UPDATED`, `DELETED`.                                    |
| `node`           | `Todo`               | The record after the change. `null` on `DELETED`. Select any `Todo` field. |
| `updatedFields`  | `[String!]`          | Names of the record fields that changed (`UPDATED` only).                  |
| `previousValues` | `TodoPreviousValues` | The record's scalar values before the change.                              |

#### TodoPreviousValues

Scalar snapshot — relations are not included.

| Field       | Type       | Field       | Type        |
| ----------- | ---------- | ----------- | ----------- |
| `id`        | `ID!`      | `text`      | `String!`   |
| `uid`       | `String!`  | `html`      | `String!`   |
| `position`  | `Float!`   | `createdAt` | `DateTime!` |
| `title`     | `String!`  | `updatedAt` | `DateTime!` |
| `startedAt` | `DateTime` | `archived`  | `Boolean!`  |
| `duedAt`    | `DateTime` | `done`      | `Boolean!`  |
| `timezone`  | `String`   |             |             |

#### MutationType

| Value     | Meaning                                                                                   |
| --------- | ----------------------------------------------------------------------------------------- |
| `CREATED` | The record was created. `previousValues` is `null`.                                       |
| `UPDATED` | The record changed. `updatedFields` lists what.                                           |
| `DELETED` | The record was deleted. `node` is `null`; read `previousValues` for the last-known state. |

## subscribeToTodoAction

Streams the record activity log in real time — every discrete action taken on a record (tag added, assignee changed, due date set, list changed, completed, and so on). Returns a `TodoActionSubscriptionPayload`. Use it to keep a record's history panel live. Unlike `subscribeToTodo`, this is scoped to a single record via a **required** filter.

### Request

```graphql
subscription OnRecordAction {
  subscribeToTodoAction(filter: { todoId: "todo_123" }) {
    mutation
    node {
      id
      type
      newValue
      oldValue
      automated
      createdAt
      user {
        id
        fullName
      }
    }
  }
}
```

### Parameters

#### SubscribeToTodoActionFilter

| Field    | Type      | Required | Description                                      |
| -------- | --------- | -------- | ------------------------------------------------ |
| `todoId` | `String!` | Yes      | The record (`Todo`) whose action feed to stream. |

### Response

```json
{
  "data": {
    "subscribeToTodoAction": {
      "mutation": "CREATED",
      "node": {
        "id": "clm4n9act000308l0i9j0k1l2",
        "type": "CHANGE_DUE_DATE",
        "newValue": "2026-06-15T17:00:00.000Z",
        "oldValue": "2026-06-12T17:00:00.000Z",
        "automated": false,
        "createdAt": "2026-05-29T09:14:22.000Z",
        "user": { "id": "clm4n8usr000108l0a1b2c3d4", "fullName": "Ada Lovelace" }
      }
    }
  }
}
```

#### TodoActionSubscriptionPayload

| Field            | Type                       | Description                                   |
| ---------------- | -------------------------- | --------------------------------------------- |
| `mutation`       | `MutationType!`            | `CREATED`, `UPDATED`, or `DELETED`.           |
| `node`           | `TodoAction`               | The action. `null` on `DELETED`.              |
| `updatedFields`  | `[String!]`                | Changed action fields (`UPDATED` only).       |
| `previousValues` | `TodoActionPreviousValues` | The action's scalar values before the change. |

#### TodoAction

| Field         | Type              | Description                                                  |
| ------------- | ----------------- | ------------------------------------------------------------ |
| `id`          | `ID!`             | Action ID.                                                   |
| `type`        | `TodoActionType!` | What kind of action — see below.                             |
| `newValue`    | `String`          | Value after the action (shape depends on `type`).            |
| `oldValue`    | `String`          | Value before the action.                                     |
| `automated`   | `Boolean`         | `true` when triggered by an automation rather than a person. |
| `createdAt`   | `DateTime!`       | When the action occurred.                                    |
| `affectedBy`  | `User`            | The user the action targeted (e.g. the assignee added).      |
| `customField` | `CustomField`     | The field involved, for `SET_CUSTOM_FIELD` actions.          |
| `user`        | `User`            | The actor who performed the action.                          |
| `todo`        | `Todo!`           | The record the action belongs to.                            |

`TodoActionType` is one of: `ADD_TAG`, `ASSIGN_AN_ASSIGNEE`, `ASSIGN_CHECKLIST_ITEM`, `CHANGE_DUE_DATE`, `CHANGE_TODO_LIST`, `COPY_TODO`, `CREATE_CHECKLIST`, `CREATE_CHECKLIST_ITEM`, `CREATE_DEPENDENCY`, `DELETE_CHECKLIST`, `DELETE_CHECKLIST_ITEM`, `DELETE_DEPENDENCY`, `DELETE_TAG`, `MARK_AS_COMPLETE`, `MARK_AS_INCOMPLETE`, `MARK_CHECKLIST_ITEM_AS_DONE`, `MARK_CHECKLIST_ITEM_AS_UNDONE`, `MOVE_TODO`, `REMOVE_DUE_DATE`, `REMOVE_TAG`, `REPEAT_TODO`, `SET_CHECKLIST_ITEM_DUE_DATE`, `SET_CUSTOM_FIELD`, `SET_DUE_DATE`, `UNASSIGN_AN_ASSIGNEE`, `UNASSIGN_CHECKLIST_ITEM`, `UPDATE_CHECKLIST`, `UPDATE_CHECKLIST_ITEM`, `UPDATE_DEPENDENCY`, `UPDATE_DESCRIPTION`, `UPDATE_TAG`, `UPDATE_TITLE`, `SEND_EMAIL`, `ADD_TO_PROJECT`, `REMOVE_FROM_PROJECT`.

## subscribeToTodoList

Streams list (`TodoList`) lifecycle changes for a workspace: lists created, renamed, reordered, locked/disabled, or deleted. Returns a `TodoListSubscriptionPayload`. Use it to keep a board's column set in sync. Record-level changes within a list come from [`subscribeToTodo`](#subscribetotodo), not this stream.

### Request

```graphql
subscription OnListChange {
  subscribeToTodoList(projectId: "project_123") {
    mutation
    node {
      id
      title
      position
      isDisabled
      isLocked
    }
    updatedFields
    previousValues {
      title
      position
    }
  }
}
```

### Parameters

| Argument    | Type     | Required | Description                                                                                                                                                    |
| ----------- | -------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `projectId` | `String` | No       | Workspace (`Project`) ID. When set, only that workspace's list events are delivered. When omitted, you receive list events from every workspace you belong to. |

### Response

```json
{
  "data": {
    "subscribeToTodoList": {
      "mutation": "CREATED",
      "node": {
        "id": "clm4n8lst000208l0e5f6g7h8",
        "title": "In Review",
        "position": 196608,
        "isDisabled": false,
        "isLocked": false
      },
      "updatedFields": null,
      "previousValues": null
    }
  }
}
```

#### TodoListSubscriptionPayload

| Field            | Type                     | Description                                                                                  |
| ---------------- | ------------------------ | -------------------------------------------------------------------------------------------- |
| `mutation`       | `MutationType!`          | `CREATED`, `UPDATED`, or `DELETED`.                                                          |
| `node`           | `TodoList`               | The list after the change. `null` on `DELETED`.                                              |
| `updatedFields`  | `[String!]`              | Changed list fields (`UPDATED` only).                                                        |
| `previousValues` | `TodoListPreviousValues` | Scalar values before the change: `id`, `uid`, `position`, `title`, `createdAt`, `updatedAt`. |

## subscribeToMyTodoCount

Streams the signed-in user's open-record count for an organization as a bare `Int!`. This is the badge counter behind the nav — it pushes a fresh integer whenever the count changes. There is no payload object and no sub-selection.

### Request

```graphql
subscription OnMyTodoCount {
  subscribeToMyTodoCount(companyId: "company_123")
}
```

### Parameters

| Argument    | Type      | Required | Description                                                                                    |
| ----------- | --------- | -------- | ---------------------------------------------------------------------------------------------- |
| `companyId` | `String!` | Yes      | Organization ID or slug. Events are scoped to this organization and to the authenticated user. |

### Response

```json
{ "data": { "subscribeToMyTodoCount": 12 } }
```

## onMoveTodo

Imperative event stream that fires when a record is moved within a workspace — between lists or repositioned. Returns the moved record as a bare `Todo!` (no payload wrapper, no `mutation` discriminator). Use it when you only care about move events and don't want to filter the full `subscribeToTodo` feed.

### Request

```graphql
subscription OnMoveRecord {
  onMoveTodo(projectId: "project_123") {
    id
    title
    position
    todoList {
      id
      title
    }
  }
}
```

### Parameters

| Argument    | Type      | Required | Description                                                                                  |
| ----------- | --------- | -------- | -------------------------------------------------------------------------------------------- |
| `projectId` | `String!` | Yes      | Workspace (`Project`) ID. You must be a member of this workspace to receive its move events. |

### Response

```json
{
  "data": {
    "onMoveTodo": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Draft Q3 launch plan",
      "position": 131072,
      "todoList": { "id": "clm4n8lst000208l0e5f6g7h8", "title": "In Review" }
    }
  }
}
```

## onMarkTodoListAsDone / onMarkTodoListAsUndone

Two imperative streams that fire when a whole list is marked done or undone — the bulk complete/uncomplete action on a column. Both return a bare `TodoList!`. They are siblings; subscribe to one or both.

### Request

```graphql
subscription OnListDone {
  onMarkTodoListAsDone(projectId: "project_123") {
    id
    title
    completed
  }
}
```

```graphql
subscription OnListUndone {
  onMarkTodoListAsUndone(projectId: "project_123") {
    id
    title
    completed
  }
}
```

### Parameters

| Argument    | Type     | Required | Description                                                                                                                                                                         |
| ----------- | -------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `projectId` | `String` | No       | Workspace (`Project`) ID. When set, only that workspace's events are delivered and you must be a member of it. When omitted, you receive events from every workspace you belong to. |

### Response

```json
{
  "data": {
    "onMarkTodoListAsDone": {
      "id": "clm4n8lst000208l0e5f6g7h8",
      "title": "In Progress",
      "completed": true
    }
  }
}
```

## Errors

Most subscription failures surface at connection time, not per operation — an unauthenticated handshake is rejected before any `subscribe` message is processed (see [Connect & Authenticate](/api/realtime/connect-and-authenticate)). Once connected, a malformed operation returns a standard GraphQL error.

| Code              | When                                                                                                                                                 |
| ----------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------- |
| `UNAUTHENTICATED` | The WebSocket handshake carried no valid credentials, so the connection was rejected.                                                                |
| `BAD_USER_INPUT`  | A required argument is missing or malformed — e.g. `subscribeToTodoAction` without `filter.todoId`, or `subscribeToMyTodoCount` without `companyId`. |

A record or workspace you can't see does not raise an error — the subscription simply never delivers events for it.

## Permissions

Filtering happens on the server, per delivered event:

- **`subscribeToTodo`** — every candidate event is checked with `permissions.todo(id).view()` before delivery. You must be able to see the record (including its workspace membership). Without a `filter.projectId`, you still receive events across the organization, but only for records that pass this per-record check.
- **`subscribeToTodoAction`** — delivered only if you're a member of the record's workspace. Field-level permissions are also honored: actions on fields your role can't view (custom fields or built-in record fields) are dropped from your stream.
- **`subscribeToTodoList`, `onMarkTodoListAsDone`, `onMarkTodoListAsUndone`, `onMoveTodo`** — delivered only for workspaces you're a member of. With `projectId` set, events are further restricted to that workspace.
- **`subscribeToMyTodoCount`** — scoped to the authenticated user; the count is always your own.

## Related

- [Connect & Authenticate](/api/realtime/connect-and-authenticate) — open and authenticate the WebSocket before subscribing.
- [Real-time (Subscriptions)](/api/realtime) — the shared payload shape and the full subscription catalog.
- [Project & Workspace Subscriptions](/api/realtime/project-subscriptions) — workspace lifecycle, membership, folders, and the project `on*` events.
- [Comments, Discussions & Chat](/api/realtime/comment-discussion-subscriptions) — conversational streams on records.
- [Custom Field Subscriptions](/api/realtime/custom-field-subscriptions) — keep a workspace's field schema in sync.
- [List Records](/api/records/list-records) — the request/response counterpart for reading records over HTTP.
