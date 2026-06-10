---
title: Lists
description: Create, read, update, and delete the lists that group records inside a workspace, using the TodoList queries and mutations.
icon: SquareKanban
order: 7
---

Lists are the columns that organize records within a workspace — the lanes of a board, the sections of a table. In the API they are `TodoList` objects, and the records they hold are `Todo` objects. Use the queries and mutations on this page to read lists, create and rename them, reorder them by position, mark every record in a list done or undone, and delete a list.

A workspace can hold at most **50 lists**. Lists are ordered by their `position` (ascending) and are scoped to the workspace they belong to. Workspace and organization arguments accept either an ID or a slug.

<youtube url="https://www.youtube.com/watch?v=qziabkn_Gu8" />

## Request

### Read a single list

Use the `todoList` query to fetch one list by ID.

```graphql
query GetTodoList {
  todoList(id: "list_123") {
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

Use the `todoLists` query to fetch all lists in a workspace, ordered by `position`.

```graphql
query GetWorkspaceLists {
  todoLists(projectId: "project_123") {
    id
    title
    position
    isLocked
    todosCount
  }
}
```

### Search lists across an organization

`todoListQueries.todoLists` is the filtered, sorted, paginated reader. It returns a `TodoListsPagination` object (`items` + `pageInfo`). `companyIds` is required; everything else is optional.

```graphql
query SearchTodoLists(
  $filter: TodoListsFilterInput!
  $sort: [TodoListsSort!]
  $skip: Int
  $take: Int
) {
  todoListQueries {
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

Use the `createTodoList` mutation. All three input fields are required — set `position` to control where the new list lands relative to existing ones.

```graphql
mutation CreateTodoList {
  createTodoList(input: { projectId: "project_123", title: "Sprint 1", position: 1 }) {
    id
    title
    position
  }
}
```

### Rename or reorder a list

Use the `editTodoList` mutation to change a list's title, position, or locked state. Only `todoListId` is required; omitted fields are left unchanged.

```graphql
mutation EditTodoList {
  editTodoList(input: { todoListId: "list_123", title: "Sprint 1 — In Progress", isLocked: true }) {
    id
    title
    position
    isLocked
  }
}
```

### Delete a list

Use the `deleteTodoList` mutation. The list and its records are moved to trash; both `projectId` and `todoListId` are required.

```graphql
mutation DeleteTodoList {
  deleteTodoList(input: { projectId: "project_123", todoListId: "list_123" }) {
    success
    operationId
  }
}
```

### Mark every record in a list done or undone

`markTodoListAsDone` and `markTodoListAsUndone` set the `done` state on every record in a list. Pass an optional `filter` (a `TodosFilter`) to scope the change to a subset of records — for example, only records carrying a given tag. Both return a nullable `Boolean`.

```graphql
mutation MarkListDone {
  markTodoListAsDone(todoListId: "list_123")
}

mutation MarkListUndone {
  markTodoListAsUndone(
    todoListId: "list_123"
    filter: { companyIds: ["company_123"], tagIds: ["tag_123"] }
  )
}
```

## Parameters

### CreateTodoListInput

| Field       | Type      | Required | Description                                                |
| ----------- | --------- | -------- | ---------------------------------------------------------- |
| `projectId` | `String!` | Yes      | Workspace the list is created in (ID or slug).             |
| `title`     | `String!` | Yes      | Display name of the list.                                  |
| `position`  | `Float!`  | Yes      | Sort position. Lists render in ascending `position` order. |

### EditTodoListInput

| Field        | Type      | Required | Description                                     |
| ------------ | --------- | -------- | ----------------------------------------------- |
| `todoListId` | `String!` | Yes      | The list to update.                             |
| `title`      | `String`  | No       | New display name.                               |
| `position`   | `Float`   | No       | New sort position.                              |
| `isLocked`   | `Boolean` | No       | Lock the list to prevent further changes to it. |

### DeleteTodoListInput

| Field        | Type      | Required | Description                                           |
| ------------ | --------- | -------- | ----------------------------------------------------- |
| `projectId`  | `String!` | Yes      | Workspace the list belongs to (used to verify scope). |
| `todoListId` | `String!` | Yes      | The list to delete.                                   |

### markTodoListAsDone / markTodoListAsUndone arguments

| Argument     | Type          | Required | Description                                                                                                |
| ------------ | ------------- | -------- | ---------------------------------------------------------------------------------------------------------- |
| `todoListId` | `String!`     | Yes      | The list whose records are updated.                                                                        |
| `filter`     | `TodosFilter` | No       | Restrict the operation to records matching this filter. When omitted, every record in the list is updated. |

`TodosFilter` is the same record-filter input used to query records — it supports `companyIds` (required when supplied), `projectIds`, `tagIds`, `assigneeIds`, `done`, `search`, `duedAt`, and more. See [List records](/api/records/list-records) for the full field set.

### TodoListsFilterInput

Used by `todoListQueries.todoLists`.

| Field        | Type         | Required | Description                      |
| ------------ | ------------ | -------- | -------------------------------- |
| `companyIds` | `[String!]!` | Yes      | Organizations to search within.  |
| `projectIds` | `[String!]`  | No       | Restrict to specific workspaces. |
| `ids`        | `[String!]`  | No       | Restrict to specific list IDs.   |
| `titles`     | `[String!]`  | No       | Match exact list titles.         |
| `search`     | `String`     | No       | Case-insensitive title search.   |

The `todoListQueries.todoLists` field also accepts `sort: [TodoListsSort!]` (default `[position_ASC]`), `skip: Int` (default `0`), `take: Int` (default `20`), and `distinct: [TodoListsFilterDistinct!]` to collapse duplicate rows. `TodoListsFilterDistinct` has a single value, `title`.

### TodoListsSort

| Value            | Description                       |
| ---------------- | --------------------------------- |
| `title_ASC`      | Title, A→Z.                       |
| `title_DESC`     | Title, Z→A.                       |
| `createdAt_ASC`  | Oldest created first.             |
| `createdAt_DESC` | Newest created first.             |
| `updatedAt_ASC`  | Least recently updated first.     |
| `updatedAt_DESC` | Most recently updated first.      |
| `position_ASC`   | By position, ascending (default). |
| `position_DESC`  | By position, descending.          |

## Response

`createTodoList` and `editTodoList` return the `TodoList`:

```json
{
  "data": {
    "createTodoList": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Sprint 1",
      "position": 1
    }
  }
}
```

`deleteTodoList` returns a `MutationResult`:

```json
{
  "data": {
    "deleteTodoList": {
      "success": true,
      "operationId": "clm4n8qwx000108l0h2pyerz9"
    }
  }
}
```

`markTodoListAsDone` and `markTodoListAsUndone` return a nullable `Boolean`:

```json
{
  "data": {
    "markTodoListAsDone": true
  }
}
```

### TodoList

| Field              | Type        | Description                                                               |
| ------------------ | ----------- | ------------------------------------------------------------------------- |
| `id`               | `ID!`       | Unique identifier.                                                        |
| `uid`              | `String!`   | Short human-readable identifier.                                          |
| `title`            | `String!`   | Display name.                                                             |
| `position`         | `Float!`    | Sort position within the workspace.                                       |
| `isDisabled`       | `Boolean!`  | Whether the list is disabled.                                             |
| `isLocked`         | `Boolean`   | Whether the list is locked against changes.                               |
| `completed`        | `Boolean`   | Whether every record in the list is done.                                 |
| `editable`         | `Boolean`   | Whether the current user may edit this list.                              |
| `deletable`        | `Boolean`   | Whether the current user may delete this list.                            |
| `createdAt`        | `DateTime!` | Creation timestamp.                                                       |
| `updatedAt`        | `DateTime!` | Last-update timestamp.                                                    |
| `createdBy`        | `User`      | User who created the list.                                                |
| `project`          | `Project`   | Parent workspace.                                                         |
| `activity`         | `Activity`  | Associated activity-feed entry.                                           |
| `assignees`        | `[User!]!`  | Distinct users assigned to records in this list.                          |
| `tags`             | `[Tag!]!`   | Distinct tags applied to records in this list.                            |
| `todos`            | `[Todo!]!`  | Records in this list (see arguments below).                               |
| `todosCount`       | `Int!`      | Count of records in this list (accepts the same filter arguments).        |
| `todosMaxPosition` | `Float!`    | Highest record position in the list — useful when appending a new record. |

The `todos` and `todosCount` fields accept arguments to filter the records they resolve: `search: String`, `tagIds: [String!]`, `assigneeIds: [String!]`, `unassigned: Boolean`, `done: Boolean`, `duedAt: DateTime`, `duedAtStart: DateTime`, `duedAtEnd: DateTime`, `startedAt: DateTime`, `fields: JSON`, and `op: FilterLogicalOperator`. `todos` additionally accepts `first: Int`, `skip: Int`, and `orderBy: TodoOrderByInput`.

```graphql
query OpenRecordsInList {
  todoList(id: "list_123") {
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

### TodoListsPagination

| Field      | Type           | Description          |
| ---------- | -------------- | -------------------- |
| `items`    | `[TodoList!]!` | The page of lists.   |
| `pageInfo` | `PageInfo!`    | Pagination metadata. |

`PageInfo` exposes `totalItems`, `totalPages`, `page`, `perPage`, `hasNextPage`, and `hasPreviousPage`. `TodoListsPagination` has no `totalCount` sibling field — read the total from `pageInfo.totalItems`.

### MutationResult

| Field         | Type       | Description                                                              |
| ------------- | ---------- | ------------------------------------------------------------------------ |
| `success`     | `Boolean!` | Whether the operation completed.                                         |
| `operationId` | `String`   | Identifier for the operation, echoed by the matching subscription event. |

## Errors

| Code                  | When                                                                                                                                         |
| --------------------- | -------------------------------------------------------------------------------------------------------------------------------------------- |
| `TODO_LIST_NOT_FOUND` | The `todoListId` does not exist. Message: `Todo list was not found.`                                                                         |
| `PROJECT_NOT_FOUND`   | The `projectId` does not exist or you are not a member of it. Message: `Project was not found.`                                              |
| `FORBIDDEN`           | The workspace already has 50 lists (`createTodoList`), or your role lacks permission for the operation (message: `You are not authorized.`). |

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

| Operation              | Allowed roles                |
| ---------------------- | ---------------------------- |
| `createTodoList`       | OWNER, ADMIN, MEMBER         |
| `editTodoList`         | OWNER, ADMIN, MEMBER, CLIENT |
| `deleteTodoList`       | OWNER, ADMIN, MEMBER         |
| `markTodoListAsDone`   | OWNER, ADMIN, MEMBER, CLIENT |
| `markTodoListAsUndone` | OWNER, ADMIN, MEMBER, CLIENT |

Custom roles refine this further: `canCreateLists` gates `createTodoList`, `canEditLists` gates `editTodoList`, and per-list `deletable` flags can override `canEditLists` for `deleteTodoList`.

<Callout variant="info" title="Positioning">

`position` is a `Float`, so you can insert a list between two existing ones without renumbering — give it a value between their positions (e.g. `1.5` to sit between `1` and `2`). Spacing existing lists out as `1`, `2`, `3` leaves room to insert later.

</Callout>

## Related

- [List records](/api/records/list-records) — query the records inside a list
- [Move a record between lists](/api/records/move-record-list)
- [List workspaces](/api/workspaces/list-workspaces)
- [Create a workspace](/api/workspaces/create-workspace)
- [Record subscriptions](/api/realtime/record-subscriptions) — `subscribeToTodoList`, `onMarkTodoListAsDone`
