---
title: Checklist Items
description: Add, edit, and delete checklist items on a record, manage assignees, and set start and due dates.
icon: ListChecks
order: 2
---

Checklist items are the individual to-dos inside a checklist. Each one can be assigned to users, scheduled with start and due dates, and marked complete. In the API they are `ChecklistItem` objects, attached to a `Checklist`, which in turn belongs to a record (`Todo`).

This page covers the five mutations that operate on checklist items:

- [`createChecklistItem`](#create-a-checklist-item) — add an item to a checklist.
- [`editChecklistItem`](#edit-a-checklist-item) — update title, position, completion, or move it to another checklist.
- [`deleteChecklistItem`](#delete-a-checklist-item) — remove an item.
- [`setChecklistItemAssignees`](#set-assignees) — replace the item's assignees.
- [`updateChecklistItemDueDate`](#set-the-due-date) — set or clear start and due dates.

To create the parent checklist first, see [Create a Checklist](/api/checklists). To find items across your whole organization, use the [`checklistItems` query](/api/checklists#query-checklist-items-across-your-organization).

## Authentication

All requests go to `https://api.blue.app/graphql` and authenticate with your personal access token headers. Headers are case-insensitive; the canonical casing is shown below.

```bash
curl https://api.blue.app/graphql \
  -H "X-Bloo-Token-ID: YOUR_TOKEN_ID" \
  -H "X-Bloo-Token-Secret: YOUR_TOKEN_SECRET" \
  -H "X-Bloo-Company-ID: YOUR_COMPANY_ID" \
  -H "Content-Type: application/json" \
  -d '{"query": "..."}'
```

`X-Bloo-Company-ID` accepts either the organization ID or its slug. See [Authentication](/api/start-guide/authentication) for how to create tokens.

## Create a checklist item

Use the `createChecklistItem` mutation to add a new item to an existing checklist. The item starts incomplete and unassigned; assign users and set dates with the dedicated mutations below.

### Request

```graphql
mutation CreateChecklistItem {
  createChecklistItem(input: { checklistId: "list_123", title: "Write unit tests", position: 1 }) {
    id
    title
    position
    done
  }
}
```

### Parameters

#### CreateChecklistItemInput

| Parameter     | Type      | Required | Description                                                                                    |
| ------------- | --------- | -------- | ---------------------------------------------------------------------------------------------- |
| `checklistId` | `String!` | Yes      | ID of the checklist to add the item to.                                                        |
| `title`       | `String!` | Yes      | Item title. HTML is stripped and stored as plain text (see [Title handling](#title-handling)). |
| `position`    | `Float!`  | Yes      | Position of the item within the checklist, used for ordering.                                  |

### Response

```json
{
  "data": {
    "createChecklistItem": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Write unit tests",
      "position": 1,
      "done": false
    }
  }
}
```

Returns the created `ChecklistItem`. See [ChecklistItem fields](#checklistitem-fields) for the full selection set.

## Edit a checklist item

Use the `editChecklistItem` mutation to update an item's title, position, or completion status, or to move it to a different checklist. Every field except `checklistItemId` is optional; omitted fields are left unchanged.

### Request

Mark an item complete:

```graphql
mutation MarkItemDone {
  editChecklistItem(input: { checklistItemId: "todo_123", done: true }) {
    id
    title
    done
  }
}
```

### Parameters

#### EditChecklistItemInput

| Parameter         | Type      | Required | Description                                               |
| ----------------- | --------- | -------- | --------------------------------------------------------- |
| `checklistItemId` | `String!` | Yes      | ID of the checklist item to edit.                         |
| `checklistId`     | `String`  | No       | ID of a different checklist to move this item into.       |
| `title`           | `String`  | No       | Updated title. HTML is stripped and stored as plain text. |
| `position`        | `Float`   | No       | Updated position within the checklist.                    |
| `done`            | `Boolean` | No       | `true` to mark complete, `false` to mark incomplete.      |

### Response

```json
{
  "data": {
    "editChecklistItem": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Write unit tests",
      "done": true
    }
  }
}
```

Returns the updated `ChecklistItem`.

### Full example

Move an item to another checklist and update its title and position in one call:

```graphql
mutation EditChecklistItem {
  editChecklistItem(
    input: {
      checklistItemId: "todo_123"
      checklistId: "list_456"
      title: "Updated task description"
      position: 2
      done: false
    }
  ) {
    id
    title
    position
    done
    checklist {
      id
      title
    }
    users {
      id
      fullName
    }
  }
}
```

Toggling `done` and changing the title or moving the item fire **distinct** activity actions: a completion change records `MARK_CHECKLIST_ITEM_AS_DONE` / `MARK_CHECKLIST_ITEM_AS_UNDONE`, while a title change records `UPDATE_CHECKLIST_ITEM`. A completion change also triggers any matching automation (`CHECKLIST_ITEM_MARKED_AS_DONE` / `CHECKLIST_ITEM_MARKED_AS_UNDONE`).

## Delete a checklist item

Use the `deleteChecklistItem` mutation to permanently remove an item. This is irreversible.

### Request

```graphql
mutation DeleteChecklistItem {
  deleteChecklistItem(id: "todo_123")
}
```

### Parameters

| Parameter | Type      | Required | Description                         |
| --------- | --------- | -------- | ----------------------------------- |
| `id`      | `String!` | Yes      | ID of the checklist item to delete. |

### Response

Returns `Boolean!` — `true` when the item is deleted.

```json
{
  "data": {
    "deleteChecklistItem": true
  }
}
```

## Set assignees

Use the `setChecklistItemAssignees` mutation to set the complete list of users assigned to an item. This is a **replace** operation: pass every assignee you want each time. Users present in the existing assignment but absent from `assigneeIds` are unassigned; new IDs are added.

### Request

```graphql
mutation AssignChecklistItem {
  setChecklistItemAssignees(
    input: { checklistItemId: "todo_123", assigneeIds: ["user_123", "user_456"] }
  )
}
```

To remove all assignees, pass an empty array:

```graphql
mutation UnassignAllFromItem {
  setChecklistItemAssignees(input: { checklistItemId: "todo_123", assigneeIds: [] })
}
```

### Parameters

#### SetChecklistItemAssigneesInput

| Parameter         | Type         | Required | Description                                                           |
| ----------------- | ------------ | -------- | --------------------------------------------------------------------- |
| `checklistItemId` | `String!`    | Yes      | ID of the checklist item.                                             |
| `assigneeIds`     | `[String!]!` | Yes      | The full set of user IDs to assign. Pass `[]` to clear all assignees. |

### Response

Returns `Boolean!` — `true` when assignees are updated.

```json
{
  "data": {
    "setChecklistItemAssignees": true
  }
}
```

Only users who are members of the record's workspace (and your organization) can be assigned; IDs that don't meet both conditions are silently skipped. Newly assigned users receive an email and push notification.

## Set the due date

Use the `updateChecklistItemDueDate` mutation to set or clear an item's start and due dates. Setting a due date schedules an overdue notification for the item's assignees.

### Request

```graphql
mutation SetItemDueDate {
  updateChecklistItemDueDate(
    input: { checklistItemId: "todo_123", duedAt: "2026-03-15T17:00:00Z" }
  ) {
    id
    title
    startedAt
    duedAt
  }
}
```

### Parameters

#### UpdateChecklistItemDueDateInput

| Parameter         | Type       | Required | Description                                       |
| ----------------- | ---------- | -------- | ------------------------------------------------- |
| `checklistItemId` | `String!`  | Yes      | ID of the checklist item.                         |
| `startedAt`       | `DateTime` | No       | Start date/time (ISO 8601). Pass `null` to clear. |
| `duedAt`          | `DateTime` | No       | Due date/time (ISO 8601). Pass `null` to clear.   |

### Response

```json
{
  "data": {
    "updateChecklistItemDueDate": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Write unit tests",
      "startedAt": null,
      "duedAt": "2026-03-15T17:00:00Z"
    }
  }
}
```

Returns the updated `ChecklistItem`.

### Full example

Set a start/due window, then later clear both dates by passing `null`:

```graphql
mutation SetItemDateRange {
  updateChecklistItemDueDate(
    input: {
      checklistItemId: "todo_123"
      startedAt: "2026-03-10T09:00:00Z"
      duedAt: "2026-03-15T17:00:00Z"
    }
  ) {
    id
    startedAt
    duedAt
    users {
      id
      fullName
    }
  }
}

mutation RemoveItemDueDate {
  updateChecklistItemDueDate(
    input: { checklistItemId: "todo_123", startedAt: null, duedAt: null }
  ) {
    id
    startedAt
    duedAt
  }
}
```

Clearing or changing the dates fires the matching webhook event: `TODO_CHECKLIST_ITEM_DUE_DATE_ADDED` when an item previously had no dates, `TODO_CHECKLIST_ITEM_DUE_DATE_REMOVED` when both fields become `null`, otherwise `TODO_CHECKLIST_ITEM_DUE_DATE_UPDATED`.

## ChecklistItem fields

`createChecklistItem`, `editChecklistItem`, and `updateChecklistItemDueDate` return a `ChecklistItem`. (`deleteChecklistItem` and `setChecklistItemAssignees` return `Boolean!` — see their Response sections.) Select any of these fields:

| Field       | Type         | Description                                                                           |
| ----------- | ------------ | ------------------------------------------------------------------------------------- |
| `id`        | `ID!`        | Unique identifier for the checklist item.                                             |
| `title`     | `String!`    | Item title.                                                                           |
| `position`  | `Float!`     | Position within the checklist.                                                        |
| `done`      | `Boolean!`   | Whether the item is marked complete.                                                  |
| `startedAt` | `DateTime`   | Start date/time, if set.                                                              |
| `duedAt`    | `DateTime`   | Due date/time, if set.                                                                |
| `createdAt` | `DateTime!`  | When the item was created.                                                            |
| `updatedAt` | `DateTime!`  | When the item was last updated.                                                       |
| `checklist` | `Checklist!` | The parent checklist (which exposes its own `todo`).                                  |
| `createdBy` | `User`       | The user who created the item. **Nullable** — null-guard it before reading subfields. |
| `users`     | `[User!]!`   | Users assigned to this item.                                                          |

<Callout variant="info" title="createdBy nullability">

`ChecklistItem.createdBy` is nullable (`User`), so it can come back `null` for items created by deleted users or by automations. This differs from `Checklist.createdBy`, which is non-null (`User!`). Always null-check `createdBy` on a checklist item before selecting `fullName` or `id`.

</Callout>

## Errors

| Code                       | When                                                                                                                            |
| -------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| `FORBIDDEN`                | The caller has `VIEW_ONLY` or `COMMENT_ONLY` access, the workspace is archived, or a custom role lacks the required permission. |
| `CHECKLIST_NOT_FOUND`      | A `checklistId` (on `createChecklistItem`, or the move target on `editChecklistItem`) doesn't exist or isn't accessible.        |
| `CHECKLIST_ITEM_NOT_FOUND` | The `checklistItemId` / `id` doesn't exist or isn't accessible.                                                                 |

```json
{
  "errors": [
    {
      "message": "Checklist item was not found.",
      "extensions": { "code": "CHECKLIST_ITEM_NOT_FOUND" }
    }
  ]
}
```

## Permissions

All five mutations require the same access level on the record's workspace. Read-only roles are denied.

| Access level   | Can manage checklist items |
| -------------- | -------------------------- |
| `OWNER`        | Yes                        |
| `ADMIN`        | Yes                        |
| `MEMBER`       | Yes                        |
| `CLIENT`       | Yes                        |
| `COMMENT_ONLY` | No                         |
| `VIEW_ONLY`    | No                         |

## Title handling

Titles passed to `createChecklistItem` and `editChecklistItem` are sanitized server-side: HTML tags are stripped and only the plain-text content is stored. Send plain text to avoid surprises.

## Related

- [Create a Checklist](/api/checklists) — create the parent checklist and query items across the organization.
- [Manage Checklists](/api/checklists/manage-checklist) — rename, reorder, and delete checklists.
- [Assignees](/api/records/assignees) — assign users to records (not just checklist items).
- [Webhooks](/api/webhooks) — subscribe to checklist item events.
