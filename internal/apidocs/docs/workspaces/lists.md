---
title: Lists
description: Create, read, update, and delete the lists that group records inside a workspace, using the RecordList queries and mutations.
icon: SquareKanban
order: 7
---

Lists are the columns that organize records within a workspace — the lanes of a board, the sections of a table. In the API they are `RecordList` objects, and the records they hold are `Record` objects. Use the queries and mutations on this page to read lists, create and rename them, reorder them, mark every record in a list done or undone, and delete a list.

A workspace can hold at most **50 lists**. Lists are ordered by their `rank` (ascending) and are scoped to the workspace they belong to. Workspace and organization arguments accept either an ID or a slug.

<youtube url="https://www.youtube.com/watch?v=qziabkn_Gu8" />

## Request

### Read a single list

Use the `recordList` query to fetch one list by ID.

```graphql
query GetRecordList {
  recordList(id: "list_123") {
    id
    uid
    title
    position
    isLocked
    todosCount
    project {
      id
      name
    }
  }
}
```

### Read every list in a workspace

Use the `recordLists` query to fetch all lists in a workspace, ordered by `rank`.

```graphql
query GetWorkspaceLists {
  recordLists(projectId: "project_123") {
    id
    title
    position
    isLocked
    todosCount
  }
}
```

### Search lists across an organization

`recordListQueries.todoLists` is the filtered, sorted, paginated reader. It returns a `RecordListsPagination` object (`items` + `pageInfo`). `companyIds` is required; everything else is optional.

```graphql
query SearchRecordLists(
  $filter: RecordListsFilterInput!
  $sort: [RecordListsSort!]
  $skip: Int
  $take: Int
) {
  recordListQueries {
    todoLists(filter: $filter, sort: $sort, skip: $skip, take: $take) {
      items {
        id
        title
        position
        project {
          id
          name
        }
      }
      pageInfo {
        totalItems
        hasNextPage
        hasPreviousPage
      }
    }
  }
}
```

```json
{
  "filter": {
    "companyIds": ["company_123"],
    "projectIds": ["project_123"],
    "search": "sprint"
  },
  "sort": ["position_ASC"],
  "skip": 0,
  "take": 20
}
```

### Create a list

Use the `createRecordList` mutation. Send `previousId`/`nextId` to place the new list between two existing lists; without them it lands at the far right end. `position` is deprecated and ignored.

```graphql
mutation CreateRecordList {
  createRecordList(input: { projectId: "project_123", title: "Sprint 1", position: 1 }) {
    id
    title
    position
  }
}
```

### Rename or reorder a list

Use the `editRecordList` mutation to change a list's title or locked state, or to reorder it with `previousId`/`nextId`. Only `todoListId` is required; omitted fields are left unchanged. `position` is deprecated and ignored — a reorder without anchors keeps the list where it is.

```graphql
mutation EditRecordList {
  editRecordList(
    input: { todoListId: "list_123", title: "Sprint 1 — In Progress", isLocked: true }
  ) {
    id
    title
    position
    isLocked
  }
}
```

### Delete a list

Use the `deleteRecordList` mutation. The list and its records are moved to Trash together; both `projectId` and `todoListId` are required. Deletion is reversible — a workspace `OWNER`/`ADMIN` can restore the list (and exactly the records that went down with it) with `restoreRecordList` until it is purged.

```graphql
mutation DeleteRecordList {
  deleteRecordList(input: { projectId: "project_123", todoListId: "list_123" }) {
    success
    operationId
  }
}
```

Restore a deleted list with `restoreTodoList`, which returns the restored `TodoList`:

```graphql
mutation RestoreTodoList {
  restoreTodoList(id: "list_123") {
    id
    title
  }
}
```

### Mark every record in a list done or undone

`markRecordListAsDone` and `markRecordListAsUndone` set the `done` state on every record in a list. Pass an optional `filter` (a `TodosFilter`) to scope the change to a subset of records — for example, only records carrying a given tag. Both return a nullable `Boolean`.

```graphql
mutation MarkListDone {
  markRecordListAsDone(todoListId: "list_123")
}

mutation MarkListUndone {
  markRecordListAsUndone(
    todoListId: "list_123"
    filter: { companyIds: ["company_123"], tagIds: ["tag_123"] }
  )
}
```

## Parameters

### CreateRecordListInput

| Field       | Type      | Required | Description                                                |
| ----------- | --------- | -------- | ---------------------------------------------------------- |
| `projectId` | `String!` | Yes      | Workspace the list is created in (ID or slug).             |
| `title`     | `String!` | Yes      | Display name of the list.                                  |
| `position`  | `Float!`  | Yes      | Deprecated and ignored since the position freeze; it selects nothing. |

### EditRecordListInput

| Field        | Type      | Required | Description                                     |
| ------------ | --------- | -------- | ----------------------------------------------- |
| `todoListId` | `String!` | Yes      | The list to update.                             |
| `title`      | `String`  | No       | New display name.                               |
| `position`   | `Float`   | No       | Deprecated and ignored — use `previousId`/`nextId` to reorder. |
| `isLocked`   | `Boolean` | No       | Lock the list to prevent further changes to it. |

### DeleteRecordListInput

| Field        | Type      | Required | Description                                           |
| ------------ | --------- | -------- | ----------------------------------------------------- |
| `projectId`  | `String!` | Yes      | Workspace the list belongs to (used to verify scope). |
| `todoListId` | `String!` | Yes      | The list to delete.                                   |

### markRecordListAsDone / markRecordListAsUndone arguments

| Argument     | Type          | Required | Description                                                                                                |
| ------------ | ------------- | -------- | ---------------------------------------------------------------------------------------------------------- |
| `todoListId` | `String!`     | Yes      | The list whose records are updated.                                                                        |
| `filter`     | `TodosFilter` | No       | Restrict the operation to records matching this filter. When omitted, every record in the list is updated. |

`TodosFilter` is the same record-filter input used to query records — it supports `companyIds` (required when supplied), `projectIds`, `tagIds`, `assigneeIds`, `done`, `search`, `duedAt`, and more. See [List records](/api/records/list-records) for the full field set.

### RecordListsFilterInput

Used by `recordListQueries.todoLists`.

| Field        | Type         | Required | Description                      |
| ------------ | ------------ | -------- | -------------------------------- |
| `companyIds` | `[String!]!` | Yes      | Organizations to search within.  |
| `projectIds` | `[String!]`  | No       | Restrict to specific workspaces. |
| `ids`        | `[String!]`  | No       | Restrict to specific list IDs.   |
| `titles`     | `[String!]`  | No       | Match exact list titles.         |
| `search`     | `String`     | No       | Case-insensitive title search.   |

The `recordListQueries.todoLists` field also accepts `sort: [RecordListsSort!]` (SDL default `[position_ASC]`; for a workspace with stable rank ordering the default resolves to `rank_ASC` — prefer sorting by `rank`, `position` values are deprecated), `skip: Int` (default `0`), `take: Int` (default `20`), and `distinct: [RecordListsFilterDistinct!]` to collapse duplicate rows. `RecordListsFilterDistinct` has a single value, `title`.

### RecordListsSort

| Value            | Description                       |
| ---------------- | --------------------------------- |
| `title_ASC`      | Title, A→Z.                       |
| `title_DESC`     | Title, Z→A.                       |
| `createdAt_ASC`  | Oldest created first.             |
| `createdAt_DESC` | Newest created first.             |
| `updatedAt_ASC`  | Least recently updated first.     |
| `updatedAt_DESC` | Most recently updated first.      |
| `position_ASC`   | Deprecated — by frozen position value, ascending. The default order is `rank_ASC`. |
| `position_DESC`  | Deprecated — by frozen position value, descending. |

## Response

`createRecordList` and `editRecordList` return the `RecordList`:

```json
{
  "data": {
    "createRecordList": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Sprint 1",
      "position": 1
    }
  }
}
```

`deleteRecordList` returns a `MutationResult`:

```json
{
  "data": {
    "deleteRecordList": {
      "success": true,
      "operationId": "clm4n8qwx000108l0h2pyerz9"
    }
  }
}
```

`markRecordListAsDone` and `markRecordListAsUndone` return a nullable `Boolean`:

```json
{
  "data": {
    "markRecordListAsDone": true
  }
}
```

### RecordList

| Field              | Type         | Description                                                               |
| ------------------ | ------------ | ------------------------------------------------------------------------- |
| `id`               | `ID!`        | Unique identifier.                                                        |
| `uid`              | `String!`    | Short human-readable identifier.                                          |
| `title`            | `String!`    | Display name.                                                             |
| `position`         | `Float!`     | Deprecated — frozen legacy value; order lists by `rank`.                  |
| `isDisabled`       | `Boolean!`   | Whether the list is disabled.                                             |
| `isLocked`         | `Boolean`    | Whether the list is locked against changes.                               |
| `completed`        | `Boolean`    | Whether every record in the list is done.                                 |
| `editable`         | `Boolean`    | Whether the current user may edit this list.                              |
| `deletable`        | `Boolean`    | Whether the current user may delete this list.                            |
| `createdAt`        | `DateTime!`  | Creation timestamp.                                                       |
| `updatedAt`        | `DateTime!`  | Last-update timestamp.                                                    |
| `createdBy`        | `User`       | User who created the list.                                                |
| `project`          | `Workspace`  | Parent workspace.                                                         |
| `activity`         | `Activity`   | Associated activity-feed entry.                                           |
| `assignees`        | `[User!]!`   | Distinct users assigned to records in this list.                          |
| `tags`             | `[Tag!]!`    | Distinct tags applied to records in this list.                            |
| `todos`            | `[Record!]!` | Records in this list (see arguments below).                               |
| `todosCount`       | `Int!`       | Count of records in this list (accepts the same filter arguments).        |
| `todosMaxPosition` | `Float!`     | Deprecated — positions no longer track record order; read a record's `rank`. |

The `todos` and `todosCount` fields accept arguments to filter the records they resolve: `search: String`, `tagIds: [String!]`, `assigneeIds: [String!]`, `unassigned: Boolean`, `done: Boolean`, `duedAt: DateTime`, `duedAtStart: DateTime`, `duedAtEnd: DateTime`, `startedAt: DateTime`, `fields: JSON`, and `op: FilterLogicalOperator`. `todos` additionally accepts `first: Int`, `skip: Int`, and `orderBy: TodoOrderByInput`.

```graphql
query OpenRecordsInList {
  recordList(id: "list_123") {
    title
    todosCount(done: false)
    todos(done: false, first: 25) {
      id
      title
      duedAt
    }
  }
}
```

### RecordListsPagination

| Field      | Type             | Description          |
| ---------- | ---------------- | -------------------- |
| `items`    | `[RecordList!]!` | The page of lists.   |
| `pageInfo` | `PageInfo!`      | Pagination metadata. |

`PageInfo` exposes `totalItems`, `totalPages`, `page`, `perPage`, `hasNextPage`, and `hasPreviousPage`. `RecordListsPagination` has no `totalCount` sibling field — read the total from `pageInfo.totalItems`.

### MutationResult

| Field         | Type       | Description                                                              |
| ------------- | ---------- | ------------------------------------------------------------------------ |
| `success`     | `Boolean!` | Whether the operation completed.                                         |
| `operationId` | `String`   | Identifier for the operation, echoed by the matching subscription event. |

## Errors

| Code                  | When                                                                                                                                           |
| --------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `TODO_LIST_NOT_FOUND` | The `todoListId` does not exist. Message: `Todo list was not found.`                                                                           |
| `PROJECT_NOT_FOUND`   | The `projectId` does not exist or you are not a member of it. Message: `Project was not found.`                                                |
| `FORBIDDEN`           | The workspace already has 50 lists (`createRecordList`), or your role lacks permission for the operation (message: `You are not authorized.`). |

```json
{
  "errors": [
    {
      "message": "You have reached the maximum number of todo lists for this project.",
      "extensions": { "code": "FORBIDDEN" }
    }
  ]
}
```

```json
{
  "errors": [
    {
      "message": "You are not authorized.",
      "extensions": { "code": "FORBIDDEN" }
    }
  ]
}
```

## Permissions

All list mutations require an active (non-archived) workspace and deny `VIEW_ONLY` and `COMMENT_ONLY` roles. Creating and deleting lists additionally deny `CLIENT`.

| Operation                | Allowed roles                |
| ------------------------ | ---------------------------- |
| `createRecordList`       | OWNER, ADMIN, MEMBER         |
| `editRecordList`         | OWNER, ADMIN, MEMBER, CLIENT |
| `deleteRecordList`       | OWNER, ADMIN, MEMBER         |
| `markRecordListAsDone`   | OWNER, ADMIN, MEMBER, CLIENT |
| `markRecordListAsUndone` | OWNER, ADMIN, MEMBER, CLIENT |

Custom roles refine this further: `canCreateLists` gates `createRecordList`, `canEditLists` gates `editRecordList`, and per-list `deletable` flags can override `canEditLists` for `deleteRecordList`.

<Callout variant="info" title="Positioning">

To insert a list between two existing ones, send `previousId` (the list to the left) and `nextId` (the list to the right) on the create or edit mutation. To move a list to an edge, send only one anchor: `nextId` alone places it at the far left, `previousId` alone at the far right.

</Callout>

## Related

- [List records](/api/records/list-records) — query the records inside a list
- [Move a record between lists](/api/records/move-record-list)
- [List workspaces](/api/workspaces/list-workspaces)
- [Create a workspace](/api/workspaces/create-workspace)
- [Record subscriptions](/api/realtime/record-subscriptions) — `subscribeToRecordList`, `onMarkRecordListAsDone`
