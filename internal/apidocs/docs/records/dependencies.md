---
title: Dependencies
description: Link records into blocking relationships, change a dependency's direction, or remove it with the dependency mutations.
icon: GitBranch
order: 9
---

Dependencies link two records into a blocking relationship — one record blocks another, or is blocked by it. Records are `Record` objects in the API, so these mutations operate on todo IDs.

Use `createRecordDependency` to link two records, `updateRecordDependency` to flip the direction of an existing link, and `deleteRecordDependency` to remove it. Each pair of records can have at most one dependency. To read existing dependencies back, select the `dependOn` and `dependBy` fields on a `Record` (see [Reading dependencies](#reading-dependencies)).

## Request

Create a dependency where one record blocks another:

```graphql
mutation CreateDependency {
  createRecordDependency(input: { type: BLOCKING, todoId: "todo_123", otherTodoId: "todo_456" }) {
    id
    title
  }
}
```

With `type: BLOCKING`, `todoId` blocks `otherTodoId` — the other record can't proceed until this one is done.

## Parameters

### CreateRecordDependencyInput

| Parameter     | Type                  | Required | Description                                                 |
| ------------- | --------------------- | -------- | ----------------------------------------------------------- |
| `type`        | `TodoDependencyType!` | Yes      | Direction of the relationship — `BLOCKING` or `BLOCKED_BY`. |
| `todoId`      | `String!`             | Yes      | ID of the primary record.                                   |
| `otherTodoId` | `String!`             | Yes      | ID of the other record in the relationship.                 |

### UpdateRecordDependencyInput

| Parameter     | Type                  | Required | Description                                 |
| ------------- | --------------------- | -------- | ------------------------------------------- |
| `type`        | `TodoDependencyType!` | Yes      | New direction for the existing dependency.  |
| `todoId`      | `String!`             | Yes      | ID of the primary record.                   |
| `otherTodoId` | `String!`             | Yes      | ID of the other record in the relationship. |

### DeleteRecordDependencyInput

| Parameter     | Type      | Required | Description                                 |
| ------------- | --------- | -------- | ------------------------------------------- |
| `todoId`      | `String!` | Yes      | ID of the primary record.                   |
| `otherTodoId` | `String!` | Yes      | ID of the other record in the relationship. |

Deleting a dependency needs only the two record IDs — direction is not specified, and the order of `todoId` and `otherTodoId` does not matter.

### TodoDependencyType

| Value        | Description                                                                                                                      |
| ------------ | -------------------------------------------------------------------------------------------------------------------------------- |
| `BLOCKING`   | The primary record (`todoId`) blocks the other record (`otherTodoId`). The other record can't proceed until the primary is done. |
| `BLOCKED_BY` | The primary record (`todoId`) is blocked by the other record (`otherTodoId`). The primary can't proceed until the other is done. |

## Response

`createRecordDependency` and `updateRecordDependency` return the primary record (the `Record` named in `todoId`). For the request above:

```json
{
  "data": {
    "createRecordDependency": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Ship the API docs"
    }
  }
}
```

`deleteRecordDependency` returns a `Boolean` — `true` when the dependency was removed.

```json
{
  "data": {
    "deleteRecordDependency": true
  }
}
```

### Returns (createRecordDependency / updateRecordDependency)

Both mutations return the primary `Record`. Commonly selected fields:

| Field       | Type          | Description                                     |
| ----------- | ------------- | ----------------------------------------------- |
| `id`        | `ID!`         | Unique identifier for the record.               |
| `uid`       | `String!`     | Short, human-readable record identifier.        |
| `title`     | `String!`     | Record title.                                   |
| `done`      | `Boolean!`    | Whether the record is complete.                 |
| `startedAt` | `DateTime`    | Start date/time, if set.                        |
| `duedAt`    | `DateTime`    | Due date/time, if set.                          |
| `todoList`  | `RecordList!` | The list the record belongs to.                 |
| `users`     | `[User!]!`    | Assigned users.                                 |
| `tags`      | `[Tag!]!`     | Tags on the record.                             |
| `dependOn`  | `[Record!]`   | Records this record depends on (is blocked by). |
| `dependBy`  | `[Record!]`   | Records that depend on this record (it blocks). |

## Full example

Flip a dependency's direction, then remove it. `updateRecordDependency` targets the pair `(todoId, otherTodoId)` regardless of how it was originally created and changes its direction; setting it to the direction it already has is a no-op that returns the record without logging activity.

```graphql
# Reverse the relationship: now todo_123 is blocked by todo_456
mutation FlipDependency {
  updateRecordDependency(input: { type: BLOCKED_BY, todoId: "todo_123", otherTodoId: "todo_456" }) {
    id
    title
    dependOn {
      id
      title
    }
  }
}

# Remove the dependency entirely
mutation RemoveDependency {
  deleteRecordDependency(input: { todoId: "todo_123", otherTodoId: "todo_456" })
}
```

## Reading dependencies

There is no dedicated dependency query. Read a record's links back through the `dependOn` and `dependBy` fields on the `Record` itself, available wherever you select a record (for example via [List records](/api/records/list-records)):

```graphql
query RecordDependencies {
  recordQueries {
    todos(filter: { companyIds: ["company_123"], todoListIds: ["list_123"] }) {
      items {
        id
        title
        dependOn {
          id
          title
        }
        dependBy {
          id
          title
        }
      }
    }
  }
}
```

`dependOn` lists the records this record is blocked by; `dependBy` lists the records it blocks.

## Errors

| Code                             | When                                                                                                                                                                                 |
| -------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `TODO_NOT_FOUND`                 | Either `todoId` or `otherTodoId` does not exist or is not accessible. Also returned by `deleteRecordDependency` when no dependency exists between the two records. |
| `TODO_DEPENDENCY_ALREADY_EXISTS` | `createRecordDependency` was called for a pair that already has a dependency. Use `updateRecordDependency` to change an existing link's direction.                   |

## Permissions

Managing dependencies requires edit access to the record's dependency field. Per the resolver (`editDependencyOrThrow`), `OWNER`, `ADMIN`, and `MEMBER` can create, update, and delete dependencies; `CLIENT`, `COMMENT_ONLY`, and `VIEW_ONLY` cannot. When the caller has a custom role, that role must grant the `DEPENDENCY` field permission.

## Related

- [List records](/api/records/list-records) — read `dependOn` / `dependBy` to see existing dependencies
- [Update a record](/api/records/update-record) — change other record properties
- [Toggle record status](/api/records/toggle-record-status) — mark the blocking record done
