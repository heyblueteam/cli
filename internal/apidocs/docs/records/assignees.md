---
title: Record Assignees
description: Set, add, or remove the users assigned to a record, and list the users eligible to be assigned.
icon: Users
order: 6
---

Records can be assigned to one or more users. Use the `setRecordAssignees`, `addRecordAssignees`, and `removeRecordAssignees` mutations to change who is assigned, and the `assignees` query to list the users who are eligible to be assigned in a workspace. Records are `Record` objects and workspaces are `Workspace` objects in the API; assignees are `User` objects.

Pick the mutation by intent:

- **`setRecordAssignees`** — replace the full assignee list. Computes the diff against the current assignees, then adds and removes only what changed. Passing an empty `assigneeIds` array clears all assignees.
- **`addRecordAssignees`** — add users without touching existing assignees. Users already assigned are skipped.
- **`removeRecordAssignees`** — remove specific users, leaving the rest assigned.

<Callout variant="info" title="Only set triggers full side effects">

`setRecordAssignees` is the only one that records activity-log entries, sends assignment notifications, fires `TODO_ASSIGNEE_ADDED` / `TODO_ASSIGNEE_REMOVED` webhooks, runs assignee automations, and refreshes charts. `addRecordAssignees` and `removeRecordAssignees` only change the assignment rows (remove also publishes a real-time record update). If you need notifications, webhooks, or automations to fire, use `setRecordAssignees`.

</Callout>

## Request

Replace the full assignee list on a record:

```graphql
mutation SetRecordAssignees {
  setRecordAssignees(input: { todoId: "todo_123", assigneeIds: ["user_123", "user_456"] }) {
    success
    operationId
  }
}
```

Add assignees without removing the existing ones:

```graphql
mutation AddRecordAssignees {
  addRecordAssignees(input: { todoId: "todo_123", assigneeIds: ["user_789"] }) {
    success
    operationId
  }
}
```

Remove specific assignees:

```graphql
mutation RemoveRecordAssignees {
  removeRecordAssignees(input: { todoId: "todo_123", assigneeIds: ["user_456"] }) {
    success
    operationId
  }
}
```

## Parameters

All three mutations take a single `input` object. The three input types are identical in shape.

### SetRecordAssigneesInput

| Parameter     | Type         | Required | Description                                                                                                                      |
| ------------- | ------------ | -------- | -------------------------------------------------------------------------------------------------------------------------------- |
| `todoId`      | `String!`    | Yes      | ID of the record to update.                                                                                                      |
| `assigneeIds` | `[String!]!` | Yes      | The complete set of user IDs the record should be assigned to. Replaces all current assignees; pass `[]` to clear all assignees. |

### AddRecordAssigneesInput

| Parameter     | Type         | Required | Description                                          |
| ------------- | ------------ | -------- | ---------------------------------------------------- |
| `todoId`      | `String!`    | Yes      | ID of the record to update.                          |
| `assigneeIds` | `[String!]!` | Yes      | User IDs to add. Users already assigned are skipped. |

### RemoveRecordAssigneesInput

| Parameter     | Type         | Required | Description                                                          |
| ------------- | ------------ | -------- | -------------------------------------------------------------------- |
| `todoId`      | `String!`    | Yes      | ID of the record to update.                                          |
| `assigneeIds` | `[String!]!` | Yes      | User IDs to remove. IDs that are not currently assigned are ignored. |

## Response

Each mutation returns a `MutationResult`. `operationId` is `null` for these operations.

```json
{
  "data": {
    "setRecordAssignees": {
      "success": true,
      "operationId": null
    }
  }
}
```

### Returns

`MutationResult`

| Field         | Type       | Description                                                           |
| ------------- | ---------- | --------------------------------------------------------------------- |
| `success`     | `Boolean!` | `true` when the assignment change was applied.                        |
| `operationId` | `String`   | Identifier for tracking the operation. `null` for assignee mutations. |

## List eligible assignees

Use the `assignees` query to list the users who can be assigned to records in a workspace — its members. It returns `[User!]!`.

```graphql
query GetAssignees {
  assignees(filter: { projectId: "project_123" }) {
    id
    fullName
    email
    image {
      thumbnail
      medium
    }
  }
}
```

```json
{
  "data": {
    "assignees": [
      {
        "id": "clm4n8qwx000008l0g4oxdqn7",
        "fullName": "Ada Lovelace",
        "email": "ada@example.com",
        "image": {
          "thumbnail": "https://blue.app/files/u/clm4n8qwx-thumb.jpg",
          "medium": "https://blue.app/files/u/clm4n8qwx-medium.jpg"
        }
      }
    ]
  }
}
```

### AssigneesFilterInput

The `filter` argument is optional. When omitted, the query uses the workspace from the `blue-workspace-id` request header.

| Parameter   | Type     | Required | Description                                                                                                                |
| ----------- | -------- | -------- | -------------------------------------------------------------------------------------------------------------------------- |
| `projectId` | `String` | No       | Workspace to list members from. Falls back to the `blue-workspace-id` header when omitted.                                 |
| `search`    | `String` | No       | Accepted by the schema but not currently applied server-side — filter the returned list by name or email in your own code. |

### User fields

`assignees` returns `User` objects. Common fields:

| Field      | Type      | Description                                                                                |
| ---------- | --------- | ------------------------------------------------------------------------------------------ |
| `id`       | `ID!`     | User ID — use this value in `assigneeIds`.                                                 |
| `fullName` | `String!` | Display name. `firstName` and `lastName` are also available.                               |
| `email`    | `String!` | User's email address.                                                                      |
| `image`    | `Image`   | Avatar, or `null`. Select sized URLs: `thumbnail`, `small`, `medium`, `large`, `original`. |

## Errors

These codes apply to all three mutations. The `assignees` query returns `FORBIDDEN` when the authenticated user is not a member of the requested workspace.

| Code                        | When                                                                                                |
| --------------------------- | --------------------------------------------------------------------------------------------------- |
| `TODO_NOT_FOUND`            | No record exists for the given `todoId`.                                                            |
| `FORBIDDEN`                 | The caller lacks edit access to the record (or, for `assignees`, is not a member of the workspace). |
| `GRAPHQL_VALIDATION_FAILED` | A required argument is missing or malformed (e.g. `assigneeIds` is `null`).                         |

## Permissions

All three mutations require edit access to the record. `VIEW_ONLY` and `COMMENT_ONLY` members are denied; `OWNER`, `ADMIN`, `MEMBER`, and `CLIENT` are allowed. A custom role with records disabled is also denied, and the workspace must be active (not archived).

| Role           | Can change assignees |
| -------------- | -------------------- |
| `OWNER`        | Yes                  |
| `ADMIN`        | Yes                  |
| `MEMBER`       | Yes                  |
| `CLIENT`       | Yes                  |
| `VIEW_ONLY`    | No                   |
| `COMMENT_ONLY` | No                   |

## Notes

- **Duplicates are prevented.** Assigning a user who is already assigned is a no-op; there is one assignment row per user per record.
- **No assignment limit.** A record can have any number of assignees.
- **Read assignees back** on a record via its `users` field — see [List Records](/api/records/list-records).

## Related

- [List Records](/api/records/list-records) — read records and their current assignees.
- [Update a Record](/api/records/update-record) — change other record properties.
- [Record Tags](/api/records/tags) — set, add, or remove tags on a record.
- [List Members](/api/user-management/list-users) — list the users in an organization.
