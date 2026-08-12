---
title: Manage Checklists
description: Rename, reposition, and delete checklists on records with the editChecklist and deleteChecklist mutations.
icon: Settings
order: 1
---

A checklist groups related sub-tasks under a record (records are `Record` objects in the API). Use the `editChecklist` mutation to rename a checklist or change its order within a record, and the `deleteChecklist` mutation to remove a checklist together with all of its items.

To add a checklist, see [Create a checklist](/api/checklists). To manage the items inside a checklist, see [Checklist items](/api/checklists/checklist-items).

All requests go to `https://api.blue.app/graphql` and must include your authentication headers:

- `blue-token-id`: your token ID
- `blue-token-secret`: your token secret
- `blue-org-id`: your organization ID or slug

Header names are case-insensitive.

## Request

Rename a checklist with `editChecklist`. Pass `checklistId` plus the fields you want to change — any field you omit is left unchanged.

```graphql
mutation EditChecklist {
  editChecklist(input: { checklistId: "list_123", title: "Pre-Release Verification" }) {
    id
    title
    position
  }
}
```

## Parameters

### EditChecklistInput

| Parameter     | Type      | Required | Description                                                     |
| ------------- | --------- | -------- | --------------------------------------------------------------- |
| `checklistId` | `String!` | Yes      | ID of the checklist to edit.                                    |
| `title`       | `String`  | No       | New title for the checklist.                                    |
| `position`    | `Float`   | No       | New sort position within the record. Lower values appear first. |

<Callout variant="tip" title="Reordering checklists">

`position` is a `Float`, so you can insert a checklist between two others by giving it a value in between — for example `1.5` to slot it between positions `1` and `2` — without renumbering the rest. A `position` of `0` is treated as no change; use a positive value to move a checklist to the top.

</Callout>

## Response

`editChecklist` returns the updated `Checklist`.

```json
{
  "data": {
    "editChecklist": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Pre-Release Verification",
      "position": 1
    }
  }
}
```

### Returns — Checklist

| Field            | Type                | Description                          |
| ---------------- | ------------------- | ------------------------------------ |
| `id`             | `ID!`               | Unique identifier for the checklist. |
| `title`          | `String!`           | Checklist title.                     |
| `position`       | `Float!`            | Sort position within the record.     |
| `createdAt`      | `DateTime!`         | When the checklist was created.      |
| `updatedAt`      | `DateTime!`         | When the checklist was last updated. |
| `todo`           | `Record!`             | The parent record.                   |
| `checklistItems` | `[ChecklistItem!]!` | Items within the checklist.          |
| `createdBy`      | `User!`             | The user who created the checklist.  |

## Full example

Update the title and position together, and select a richer set of fields on the returned checklist.

```graphql
mutation EditChecklistFull {
  editChecklist(
    input: { checklistId: "list_123", title: "Pre-Release Verification", position: 2 }
  ) {
    id
    title
    position
    updatedAt
    createdBy {
      id
      fullName
    }
    todo {
      id
      title
    }
    checklistItems {
      id
      title
      done
    }
  }
}
```

## Delete a checklist

Use the `deleteChecklist` mutation to permanently remove a checklist and every item it contains. It returns `Boolean!`.

```graphql
mutation DeleteChecklist {
  deleteChecklist(id: "list_123")
}
```

| Parameter | Type      | Required | Description                    |
| --------- | --------- | -------- | ------------------------------ |
| `id`      | `String!` | Yes      | ID of the checklist to delete. |

```json
{
  "data": {
    "deleteChecklist": true
  }
}
```

<Callout variant="danger" title="Deletion is permanent">

Deleting a checklist removes all of its items along with their assignees, due dates, and completion state. This cannot be undone.

</Callout>

## Side effects

Editing a checklist:

- Records an `UPDATE_CHECKLIST` activity entry with the old and new titles.
- Fires the `TODO_CHECKLIST_NAME_CHANGED` webhook event when the title changes.
- Pushes a real-time update to everyone viewing the record.

Deleting a checklist:

- Records a `DELETE_CHECKLIST` activity entry.
- Fires the `TODO_CHECKLIST_DELETED` webhook event.
- Pushes a real-time update to everyone viewing the record.

## Errors

| Code                  | When                                                                        |
| --------------------- | --------------------------------------------------------------------------- |
| `CHECKLIST_NOT_FOUND` | No checklist matches the given ID, or you lack access to its parent record. |
| `FORBIDDEN`           | Your role is `VIEW_ONLY` or `COMMENT_ONLY`, or the workspace is archived.   |
| `UNAUTHENTICATED`     | The request is missing or has invalid authentication headers.               |

## Permissions

Editing and deleting checklists requires write access to the parent record's workspace.

| Role           | Edit / delete checklists |
| -------------- | ------------------------ |
| `OWNER`        | Yes                      |
| `ADMIN`        | Yes                      |
| `MEMBER`       | Yes                      |
| `CLIENT`       | Yes                      |
| `COMMENT_ONLY` | No                       |
| `VIEW_ONLY`    | No                       |

## Related

- [Create a checklist](/api/checklists)
- [Checklist items](/api/checklists/checklist-items)
- [Update a record](/api/records/update-record)
