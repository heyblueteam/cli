---
title: Copy Record
description: Duplicate a record into any list, choosing exactly which elements carry over.
icon: Copy
order: 5
---

Use the `copyTodo` mutation to duplicate an existing record into a target list, choosing exactly which elements (description, due date, checklists, assignees, tags, custom fields, comments) carry over to the copy. Records are `Todo` objects in the API; lists are `TodoList` objects.

The copy is created in the destination list — which can belong to the same workspace or a different one — and is placed at the bottom of that list. Subtasks (todo actions) are always copied. The mutation returns `Boolean!`; on failure it throws an error rather than returning `false`.

## Request

```graphql
mutation CopyRecord {
  copyTodo(
    input: {
      todoId: "todo_123"
      todoListId: "list_123"
      title: "Copy of Q3 onboarding"
      options: [DESCRIPTION, DUE_DATE, CHECKLISTS, ASSIGNEES, TAGS, CUSTOM_FIELDS, COMMENTS]
    }
  )
}
```

Send these headers with the request. Company and project accept either an ID or a slug, and header names are case-insensitive.

```http
X-Bloo-Token-ID: YOUR_TOKEN_ID
X-Bloo-Token-Secret: YOUR_TOKEN_SECRET
X-Bloo-Company-ID: YOUR_COMPANY_ID
X-Bloo-Project-ID: project_123
```

## Parameters

### CopyTodoInput

| Parameter    | Type                 | Required | Description                                                                                                          |
| ------------ | -------------------- | -------- | -------------------------------------------------------------------------------------------------------------------- |
| `todoId`     | `String!`            | Yes      | ID of the record to copy.                                                                                            |
| `todoListId` | `String!`            | Yes      | ID of the list to create the copy in. May be in the same workspace or a different one.                               |
| `options`    | `[CopyTodoOption!]!` | Yes      | Which elements to copy from the source record. Pass an empty list `[]` to copy only the record's title and subtasks. |
| `title`      | `String`             | No       | Title for the copy. Defaults to the source record's title when omitted.                                              |

### CopyTodoOption

Each value in `options` opts a category of data into the copy. Anything not listed is not copied.

| Value           | Description                                                                                                              |
| --------------- | ------------------------------------------------------------------------------------------------------------------------ |
| `DESCRIPTION`   | The record's description (`text` / `html`).                                                                              |
| `DUE_DATE`      | The due date and its timezone.                                                                                           |
| `ASSIGNEES`     | Assigned users. On a cross-workspace copy, assignees are filtered to users who are members of the destination workspace. |
| `TAGS`          | Associated tags.                                                                                                         |
| `COMMENTS`      | Comments and their replies.                                                                                              |
| `CHECKLISTS`    | Checklists and their items.                                                                                              |
| `CUSTOM_FIELDS` | Custom field values, including file attachments (duplicated to new storage references).                                  |

## Response

```json
{
  "data": {
    "copyTodo": true
  }
}
```

### Returns

| Field      | Type       | Description                                                                          |
| ---------- | ---------- | ------------------------------------------------------------------------------------ |
| `copyTodo` | `Boolean!` | `true` when the copy succeeds. Failures throw an error instead of returning `false`. |

## Errors

| Code                  | When                                                                                                                   |
| --------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `TODO_NOT_FOUND`      | `todoId` does not exist, or you lack `OWNER` / `ADMIN` / `MEMBER` access to its workspace.                             |
| `TODO_LIST_NOT_FOUND` | `todoListId` does not exist, or you lack `OWNER` / `ADMIN` / `MEMBER` access to the destination workspace.             |
| `FORBIDDEN`           | Your role does not permit copying (e.g. `VIEW_ONLY` / `COMMENT_ONLY`), or a `MEMBER` attempted a cross-workspace copy. |
| `PLAN_LIMIT_REACHED`  | The destination workspace or organization has reached its record limit for the current plan.                           |

```json
{
  "errors": [
    {
      "message": "You are not authorized.",
      "extensions": {
        "code": "FORBIDDEN"
      }
    }
  ]
}
```

## Permissions

You need edit access to both the source and destination workspaces. The role on the source workspace also governs whether the copy can cross workspaces.

| Role           | Same workspace | Cross-workspace |
| -------------- | -------------- | --------------- |
| `OWNER`        | Yes            | Yes             |
| `ADMIN`        | Yes            | Yes             |
| `MEMBER`       | Yes            | No              |
| `COMMENT_ONLY` | No             | No              |
| `VIEW_ONLY`    | No             | No              |

A `MEMBER` can only copy within the same workspace; a cross-workspace copy from a `MEMBER` returns `FORBIDDEN`. Cross-workspace copies also run any automations configured for records arriving from another workspace.

## Related

- [Move a record to a list](/api/records/move-record-list)
- [Update a record](/api/records/update-record)
- [List records](/api/records/list-records)
- [Custom field values](/api/custom-fields/custom-field-values)
