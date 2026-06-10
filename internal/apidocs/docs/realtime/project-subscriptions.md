---
title: Project & Workspace Subscriptions
description: Real-time streams for workspace lifecycle, membership, folders, tags, and saved views over the GraphQL WebSocket.
icon: FolderKanban
order: 3
---

Subscribe to changes in your workspaces and their organizing structure. Workspaces are `Project` objects in the API, so every operation on this page is named `subscribeToProject*`, `subscribeToFolder`, `subscribeToTag`, `subscribeToSavedView`, or an imperative `on*` event. They share the GraphQL WebSocket transport (`wss://api.blue.app/graphql`) and the same authenticated handshake as every other subscription — see [Connect & Authenticate](/api/realtime/connect-and-authenticate) for the `connectionParams` setup.

Two payload shapes appear here:

- **Change-feed subscriptions** (`subscribeToProject`, `subscribeToFolder`, `subscribeToTag`, `subscribeToSavedView`, `onSetProjectFolder`) return a `*SubscriptionPayload` with the shared shape `{ mutation, node, previousValues, updatedFields }`. The `mutation` field is a `MutationType` (`CREATED`, `UPDATED`, or `DELETED`). This shape is explained once on the [Real-time overview](/api/realtime).
- **Bare-entity events** (`onAddedToProject`, `onRemovedFromProject`, `onArchiveProject`, `onUnarchiveProject`, `onConvertToTemplate`, `onRemovedFromTemplate`) resolve directly to a `Project!` — no envelope. `onCopyProjectStarted` / `onCopyProjectFinished` resolve to a `CopyProjectStatus`.

<Callout variant="info" title="Events are gated per delivery">

Subscribing always succeeds. You only receive events for workspaces you can see: each event is filtered server-side against your project membership before it is delivered. Pass company and project IDs as either the ID or the slug.

</Callout>

## subscribeToProject

The primary workspace change-feed. Streams `CREATED` / `UPDATED` / `DELETED` events for every workspace you belong to within a company. Use the optional arguments to narrow the stream to a folder, to archived workspaces, or to templates.

### Request

```graphql
subscription WatchProjects {
  subscribeToProject(companyId: "company_123") {
    mutation
    node {
      id
      name
      archived
      isTemplate
      folder {
        id
        title
      }
    }
    updatedFields
    previousValues {
      name
      archived
    }
  }
}
```

### Parameters

| Argument     | Type      | Required | Description                                                                                                                                                |
| ------------ | --------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `companyId`  | `String!` | Yes      | Organization to watch, by ID or slug. Only workspaces in this company are delivered.                                                                       |
| `folderId`   | `String`  | No       | Restrict to workspaces filed in this folder for the current user. Omit to watch every folder; pass `null` explicitly to watch only un-foldered workspaces. |
| `archived`   | `Boolean` | No       | When `true`, deliver only archived workspaces.                                                                                                             |
| `isTemplate` | `Boolean` | No       | When `true`, deliver only templates. With neither `archived` nor `isTemplate` set, only active (non-archived, non-template) workspaces are delivered.      |

### Response

Each delivered event is a `ProjectSubscriptionPayload`.

```json
{
  "data": {
    "subscribeToProject": {
      "mutation": "UPDATED",
      "node": {
        "id": "clm4n8qwx000008l0g4oxdqn7",
        "name": "Q3 Launch Plan",
        "archived": false,
        "isTemplate": false,
        "folder": { "id": "folder_123", "title": "Marketing" }
      },
      "updatedFields": ["name"],
      "previousValues": { "name": "Q3 Plan", "archived": false }
    }
  }
}
```

#### ProjectSubscriptionPayload

| Field            | Type                    | Description                                                                                                                                                                                       |
| ---------------- | ----------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `mutation`       | `MutationType!`         | `CREATED`, `UPDATED`, or `DELETED`.                                                                                                                                                               |
| `node`           | `Project`               | The workspace's current state. `null` on `DELETED`.                                                                                                                                               |
| `updatedFields`  | `[String!]`             | Names of the workspace fields that changed on `UPDATED`.                                                                                                                                          |
| `previousValues` | `ProjectPreviousValues` | Prior values for the changed fields. Exposes the scalar subset: `id`, `uid`, `slug`, `name`, `description`, `archived`, `isTemplate`, `isOfficialTemplate`, `category`, `createdAt`, `updatedAt`. |

## subscribeToFolder

Streams folder `CREATED` / `UPDATED` / `DELETED` events. Folders come in two kinds, selected by the required `type` argument: workspace folders (`PROJECT`, shared org-wide) and file folders (`FILE`, scoped to a workspace).

### Request

```graphql
subscription WatchProjectFolders {
  subscribeToFolder(companyId: "company_123", type: PROJECT) {
    mutation
    node {
      id
      title
      type
      parent {
        id
        title
      }
    }
    updatedFields
  }
}
```

### Parameters

| Argument    | Type          | Required | Description                                                                                                  |
| ----------- | ------------- | -------- | ------------------------------------------------------------------------------------------------------------ |
| `companyId` | `String!`     | Yes      | Organization to watch, by ID or slug.                                                                        |
| `type`      | `FolderType!` | Yes      | `PROJECT` (workspace folders, delivered to any company member) or `FILE` (file folders inside a workspace).  |
| `projectId` | `String`      | No       | Required for `FILE` folders — the workspace whose file folders you want. Ignored for `PROJECT`.              |
| `parentId`  | `String`      | No       | Restrict to children of this folder. Pass `null` to watch only top-level folders; omit to watch every level. |

#### FolderType

| Value     | Description                                                             |
| --------- | ----------------------------------------------------------------------- |
| `PROJECT` | Folders that organize workspaces. Shared across the whole organization. |
| `FILE`    | Folders that organize files inside a single workspace.                  |

### Response

```json
{
  "data": {
    "subscribeToFolder": {
      "mutation": "CREATED",
      "node": {
        "id": "folder_123",
        "title": "Marketing",
        "type": "PROJECT",
        "parent": null
      },
      "updatedFields": null
    }
  }
}
```

`FolderSubscriptionPayload` carries `{ mutation, node: Folder, updatedFields, previousValues: FolderPreviousValues }`.

## subscribeToTag

Streams tag `CREATED` / `UPDATED` / `DELETED` events for a single workspace. Tags carry a `title` and `color` (there is no `Tag.name`).

### Request

```graphql
subscription WatchTags {
  subscribeToTag(projectId: "project_123") {
    mutation
    node {
      id
      title
      color
    }
    updatedFields
  }
}
```

### Parameters

| Argument    | Type      | Required | Description                        |
| ----------- | --------- | -------- | ---------------------------------- |
| `projectId` | `String!` | Yes      | Workspace to watch, by ID or slug. |

### Response

```json
{
  "data": {
    "subscribeToTag": {
      "mutation": "UPDATED",
      "node": { "id": "tag_123", "title": "Urgent", "color": "#FF4D4D" },
      "updatedFields": ["color"]
    }
  }
}
```

`TagSubscriptionPayload` carries `{ mutation, node: Tag, updatedFields, previousValues: TagPreviousValues }`.

## subscribeToSavedView

Streams saved-view `CREATED` / `UPDATED` / `DELETED` events for a workspace. Saved views are `SavedView` objects. Shared views (`isShared: true`) are delivered to every workspace member; personal views are delivered only to their creator.

### Request

```graphql
subscription WatchSavedViews {
  subscribeToSavedView(projectId: "project_123") {
    mutation
    node {
      id
      name
      viewType
      isShared
    }
  }
}
```

### Parameters

| Argument    | Type      | Required | Description                        |
| ----------- | --------- | -------- | ---------------------------------- |
| `projectId` | `String!` | Yes      | Workspace to watch, by ID or slug. |

### Response

```json
{
  "data": {
    "subscribeToSavedView": {
      "mutation": "CREATED",
      "node": {
        "id": "clm4n8qwx000008l0g4oxdqn7",
        "name": "My open records",
        "viewType": "DATABASE",
        "isShared": false
      }
    }
  }
}
```

`SavedViewSubscriptionPayload` carries `{ mutation, node: SavedView, previousValues: SavedView }` — note there is no `updatedFields` on this payload, and `previousValues` is a full `SavedView` rather than a scalar subset.

See [Saved Views](/api/saved-views) for the matching read and write operations.

## subscribeToProjectPeopleList

```graphql
subscription WatchProjectPeople {
  subscribeToProjectPeopleList(projectId: "project_123") {
    mutation
    node {
      id
      fullName
      email
    }
  }
}
```

| Argument    | Type      | Required | Description                               |
| ----------- | --------- | -------- | ----------------------------------------- |
| `projectId` | `String!` | Yes      | Workspace whose membership list you want. |

Returns a `UserSubscriptionPayload` (`{ mutation, node: User, updatedFields, previousValues: UserPreviousValues }`).

<Callout variant="warning" title="No publisher wired">

This subscription accepts a connection but its resolver currently always resolves `null` — no membership events are emitted. To react to membership changes today, use `onAddedToProject` / `onRemovedFromProject` below, which are fully wired. Treat `subscribeToProjectPeopleList` as reserved.

</Callout>

## subscribeToProjectUserRole

```graphql
subscription WatchMyProjectRole {
  subscribeToProjectUserRole(input: { projectId: "project_123", userId: "user_123" }) {
    mutation
    node {
      id
      name
    }
  }
}
```

#### SubscribeToProjectUserRoleInput

| Field       | Type      | Required | Description                                         |
| ----------- | --------- | -------- | --------------------------------------------------- |
| `projectId` | `String!` | Yes      | Workspace to watch.                                 |
| `userId`    | `String!` | Yes      | The member whose role assignment you want to track. |

Returns a `ProjectUserRoleSubscriptionPayload` (`{ mutation, node: ProjectUserRole, previousValues: ProjectUserRole }`). Like the people-list stream, this channel is a stub with no production publisher — prefer the wired `on*` events for membership changes.

## Imperative project events (`on*`)

These channels fire on specific lifecycle actions rather than streaming a generic change feed. Each is delivered only to members of the affected workspace (and, where a `companyId` argument exists, only for that organization).

### onAddedToProject

Fires for the current user when they are added to a workspace. Resolves to the `Project!` they joined. Takes no arguments — it is implicitly scoped to the authenticated user.

```graphql
subscription OnJoinedProject {
  onAddedToProject {
    id
    name
    company {
      id
      slug
    }
  }
}
```

```json
{
  "data": {
    "onAddedToProject": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Q3 Launch Plan",
      "company": { "id": "company_123", "slug": "acme" }
    }
  }
}
```

### onRemovedFromProject

Mirror of `onAddedToProject`: fires for the current user when they are removed from a workspace, resolving to the `Project!` they left. No arguments.

```graphql
subscription OnLeftProject {
  onRemovedFromProject {
    id
    name
  }
}
```

### onArchiveProject

Fires when a workspace in the given organization is archived. Resolves to the archived `Project!`.

```graphql
subscription OnArchive {
  onArchiveProject(companyId: "company_123") {
    id
    name
    archived
  }
}
```

| Argument    | Type      | Required | Description                           |
| ----------- | --------- | -------- | ------------------------------------- |
| `companyId` | `String!` | Yes      | Organization to watch, by ID or slug. |

### onUnarchiveProject

Inverse of `onArchiveProject` — fires when a workspace is restored from the archive. Same `companyId` argument, resolves to the restored `Project!`.

```graphql
subscription OnUnarchive {
  onUnarchiveProject(companyId: "company_123") {
    id
    name
    archived
  }
}
```

### onSetProjectFolder

Fires when a workspace is moved into (or out of) a folder. Unlike the other project `on*` events, this one returns the full change-feed envelope, `ProjectSubscriptionPayload!`, so you can read the workspace's new state from `node`.

```graphql
subscription OnFolderChange {
  onSetProjectFolder(companyId: "company_123", folderId: "folder_123") {
    mutation
    node {
      id
      name
      folder {
        id
        title
      }
    }
  }
}
```

| Argument    | Type      | Required | Description                                                                                                                  |
| ----------- | --------- | -------- | ---------------------------------------------------------------------------------------------------------------------------- |
| `companyId` | `String!` | Yes      | Organization to watch (matched by ID).                                                                                       |
| `folderId`  | `String`  | No       | Deliver only when the workspace is filed into this folder. Pass `null` to watch moves out of all folders (to the top level). |

### onConvertToTemplate

Fires when a workspace in the organization is converted into a template. Resolves to the affected `Project!` (now `isTemplate: true`).

```graphql
subscription OnConvertToTemplate {
  onConvertToTemplate(companyId: "company_123") {
    id
    name
    isTemplate
  }
}
```

| Argument    | Type      | Required | Description                           |
| ----------- | --------- | -------- | ------------------------------------- |
| `companyId` | `String!` | Yes      | Organization to watch, by ID or slug. |

### onRemovedFromTemplate

Inverse of `onConvertToTemplate` — fires when a workspace is converted back from a template into a regular workspace. Resolves to the affected `Project!`.

```graphql
subscription OnRemovedFromTemplate {
  onRemovedFromTemplate(companyId: "company_123") {
    id
    name
    isTemplate
  }
}
```

| Argument    | Type      | Required | Description                           |
| ----------- | --------- | -------- | ------------------------------------- |
| `companyId` | `String!` | Yes      | Organization to watch, by ID or slug. |

### onCopyProjectStarted / onCopyProjectFinished

Copying a workspace (or a large template) is a background job. These two channels report its lifecycle to the user who started the copy: `onCopyProjectStarted` fires when the job is queued, `onCopyProjectFinished` when it completes. Both resolve to a `CopyProjectStatus` and take no arguments — each is implicitly scoped to the authenticated user who triggered the copy.

```graphql
subscription OnCopyProgress {
  onCopyProjectStarted {
    newProjectName
    isTemplate
    isActive
    queuePosition
    totalQueues
    newCompany {
      id
      slug
    }
  }
}
```

```json
{
  "data": {
    "onCopyProjectStarted": {
      "newProjectName": "Q3 Launch Plan (copy)",
      "isTemplate": false,
      "isActive": true,
      "queuePosition": 2,
      "totalQueues": 3,
      "newCompany": { "id": "company_123", "slug": "acme" }
    }
  }
}
```

#### CopyProjectStatus

| Field            | Type      | Description                                |
| ---------------- | --------- | ------------------------------------------ |
| `oldCompany`     | `Company` | Source organization.                       |
| `oldProject`     | `Project` | Source workspace being copied.             |
| `newCompany`     | `Company` | Destination organization.                  |
| `newProjectName` | `String`  | Name of the copy being created.            |
| `isTemplate`     | `Boolean` | Whether the copy is a template.            |
| `isActive`       | `Boolean` | Whether the copy job is currently running. |
| `queuePosition`  | `Int`     | Position in the copy queue while waiting.  |
| `totalQueues`    | `Int`     | Total jobs in the queue.                   |

Pair these with the `copyProject` mutation in [Workspaces](/api/workspaces) to drive a progress UI.

## Errors

Subscription errors surface either at the WebSocket handshake (connection-level) or on the individual subscription operation.

| Code              | When                                                                                                                                                                             |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `UNAUTHENTICATED` | The `connectionParams` carry no valid credential. The handshake is rejected before any subscription runs — see [Connect & Authenticate](/api/realtime/connect-and-authenticate). |
| `BAD_USER_INPUT`  | A required argument is missing or malformed (e.g. `companyId` omitted on `subscribeToProject`, or `type` omitted on `subscribeToFolder`).                                        |
| `FORBIDDEN`       | The authenticated user lacks access to the scoped organization or workspace.                                                                                                     |

## Permissions

You may open any of these subscriptions, but delivery is filtered per event:

- **Workspace-scoped streams** (`subscribeToProject`, `subscribeToTag`, `subscribeToSavedView`, the project `on*` events) deliver only for workspaces where you are a member. `subscribeToSavedView` additionally hides personal views you did not create.
- **`subscribeToFolder` with `type: PROJECT`** delivers to any member of the organization, because workspace folders are shared org-wide. With `type: FILE` it is scoped to members of the named workspace.
- **`onCopyProjectStarted` / `onCopyProjectFinished`** deliver only to the user who initiated the copy.

## Related

- [Real-time overview](/api/realtime) — the shared `{ mutation, node, previousValues, updatedFields }` payload and the full subscription catalog.
- [Connect & Authenticate](/api/realtime/connect-and-authenticate) — opening an authenticated WebSocket with `connectionParams`.
- [Record & Todo-List Subscriptions](/api/realtime/record-subscriptions) — `subscribeToTodo` and the record-level `on*` events.
- [Workspaces](/api/workspaces) — create, archive, copy, and template the workspaces these streams report on.
- [Saved Views](/api/saved-views) — read and write the views from `subscribeToSavedView`.
