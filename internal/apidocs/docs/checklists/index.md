---
title: Create a Checklist
description: Add a checklist to a record with createChecklist, and query checklist items across your whole organization.
icon: CheckSquare
order: 0
---

A checklist groups related sub-tasks under a record (records are `Todo` objects in the API). Use the `createChecklist` mutation to add a new checklist to a record, then use [Checklist items](/api/checklists/checklist-items) to populate it. This page also documents the organization-wide `checklistItems` query for finding items across every record.

All requests go to `https://api.blue.app/graphql` and must include your authentication headers:

- `X-Bloo-Token-ID`: your token ID
- `X-Bloo-Token-Secret`: your token secret
- `X-Bloo-Company-ID`: your organization ID or slug

Header names are case-insensitive.

## Request

Create a checklist by attaching it to a record. All three fields are required.

```graphql
mutation CreateChecklist {
  createChecklist(input: { todoId: "todo_123", title: "Launch Checklist", position: 1 }) {
    id
    title
    position
  }
}
```

## Parameters

### CreateChecklistInput

| Parameter  | Type      | Required | Description                                                 |
| ---------- | --------- | -------- | ----------------------------------------------------------- |
| `todoId`   | `String!` | Yes      | ID of the record to attach the checklist to.                |
| `title`    | `String!` | Yes      | Title of the checklist.                                     |
| `position` | `Float!`  | Yes      | Sort position within the record. Lower values appear first. |

<Callout variant="tip" title="Positioning checklists">

`position` is a `Float`, so you can slot a new checklist between two existing ones by giving it an in-between value — for example `1.5` to land it between positions `1` and `2` — without renumbering the rest.

</Callout>

## Response

`createChecklist` returns the newly created `Checklist`. It starts with no items — add them with [`createChecklistItem`](/api/checklists/checklist-items).

```json
{
  "data": {
    "createChecklist": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Launch Checklist",
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
| `todo`           | `Todo!`             | The parent record.                   |
| `checklistItems` | `[ChecklistItem!]!` | Items within the checklist.          |
| `createdBy`      | `User!`             | The user who created the checklist.  |

## Full example

Create a checklist and select a richer set of fields, including the parent record and creator.

```graphql
mutation CreateChecklistFull {
  createChecklist(input: { todoId: "todo_123", title: "QA Verification Steps", position: 2 }) {
    id
    title
    position
    createdAt
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

## Side effects

Creating a checklist:

- Records a `CREATE_CHECKLIST` activity entry on the record.
- Fires the `TODO_CHECKLIST_CREATED` webhook event.
- Pushes a real-time update to everyone viewing the record.

## Errors

| Code              | When                                                                      |
| ----------------- | ------------------------------------------------------------------------- |
| `TODO_NOT_FOUND`  | No record matches `todoId`, or you lack access to it.                     |
| `FORBIDDEN`       | Your role is `VIEW_ONLY` or `COMMENT_ONLY`, or the workspace is archived. |
| `UNAUTHENTICATED` | The request is missing or has invalid authentication headers.             |

## Permissions

Creating a checklist requires write access to the parent record's workspace.

| Role           | Create checklists |
| -------------- | ----------------- |
| `OWNER`        | Yes               |
| `ADMIN`        | Yes               |
| `MEMBER`       | Yes               |
| `CLIENT`       | Yes               |
| `COMMENT_ONLY` | No                |
| `VIEW_ONLY`    | No                |

---

## Query checklist items across your organization

The `checklistItems` query searches and paginates checklist items across every record in your organization, regardless of which record or checklist they belong to. Use it to build "my open items" dashboards or to find items by keyword. To read the items of a single checklist, select `checklistItems` on the `Checklist` object instead.

```graphql
query MyChecklistItems {
  checklistItems(filter: { assigneeIds: ["user_123"], done: false }, skip: 0, take: 20) {
    items {
      id
      title
      done
      duedAt
      checklist {
        id
        title
      }
    }
    pageInfo {
      totalItems
      hasNextPage
    }
  }
}
```

<Callout variant="warning" title="filter is required">

`filter` is a required argument (`ChecklistItemFilterInput!`). You must pass it even to fetch everything — use an empty object, `filter: {}`, to apply no filtering. Calling `checklistItems(take: 20)` with no `filter` fails GraphQL validation.

</Callout>

### Arguments

| Argument | Type                        | Required | Description                                                                  |
| -------- | --------------------------- | -------- | ---------------------------------------------------------------------------- |
| `filter` | `ChecklistItemFilterInput!` | Yes      | Filtering criteria. Pass `{}` to apply none.                                 |
| `sort`   | `[TodosSort!]`              | No       | Ordering rules, applied in order. See [Sorting](#sorting). Defaults to none. |
| `skip`   | `Int`                       | No       | Number of items to skip (offset). Defaults to `0`.                           |
| `take`   | `Int`                       | No       | Maximum number of items to return (page size). Defaults to `20`.             |

### ChecklistItemFilterInput

| Field                     | Type        | Description                                                                                               |
| ------------------------- | ----------- | --------------------------------------------------------------------------------------------------------- |
| `q`                       | `String`    | Case-insensitive substring match on the checklist item's title.                                           |
| `done`                    | `Boolean`   | Filter by the item's completion state. Omit to include both done and not-done items.                      |
| `todoDone`                | `Boolean`   | Filter by the parent record's completion state. Omit to include items on both open and completed records. |
| `excludeArchivedProjects` | `Boolean`   | Tri-state workspace filter — see the note below.                                                          |
| `assigneeIds`             | `[String!]` | Return only items assigned to any of these user IDs.                                                      |

<Callout variant="warning" title="excludeArchivedProjects is tri-state">

Despite the name, this field is not a simple on/off toggle:

- **Omit it** to include items from **all** workspaces (archived and active).
- **`true`** returns items from **non-archived** workspaces only.
- **`false`** returns items from **archived** workspaces only.

To get everything, leave the field out — don't set it to `false`.

</Callout>

### Sorting

The `sort` argument is a list of `TodosSort!` values applied in order; append `_ASC` or `_DESC` to each. The `TodosSort` enum is large (it is shared with the records query), but the `checklistItems` resolver only honors the nine values below. Any other value is silently ignored.

| Sort value           | Orders by               |
| -------------------- | ----------------------- |
| `title`              | Parent record title     |
| `todoListTitle`      | List title              |
| `todoListPosition`   | List position           |
| `position`           | Checklist item position |
| `startedAt`          | Item start date         |
| `duedAt`             | Item due date           |
| `projectName`        | Workspace name          |
| `checklistTitle`     | Checklist title         |
| `checklistItemTitle` | Checklist item title    |

### Full example

Find open items containing "deploy", assigned to either of two users, on active records and non-archived workspaces, sorted by due date.

```graphql
query SearchChecklistItems {
  checklistItems(
    filter: {
      q: "deploy"
      done: false
      todoDone: false
      excludeArchivedProjects: true
      assigneeIds: ["user_123", "user_456"]
    }
    sort: [duedAt_ASC]
    skip: 0
    take: 50
  ) {
    items {
      id
      title
      done
      startedAt
      duedAt
      createdAt
      users {
        id
        fullName
      }
      checklist {
        id
        title
        todo {
          id
          title
        }
      }
    }
    pageInfo {
      totalItems
      totalPages
      page
      perPage
      hasNextPage
      hasPreviousPage
    }
  }
}
```

### Response

`checklistItems` returns a `ChecklistItemPagination` object: an `items` array plus a `pageInfo` block describing the page.

```json
{
  "data": {
    "checklistItems": {
      "items": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "title": "Deploy to staging",
          "done": false,
          "duedAt": "2026-06-01T17:00:00.000Z",
          "checklist": {
            "id": "clm4n8qwx000009l0checklist1",
            "title": "Release Checklist"
          }
        }
      ],
      "pageInfo": {
        "totalItems": 1,
        "hasNextPage": false
      }
    }
  }
}
```

### Returns — ChecklistItemPagination

| Field      | Type                | Description                         |
| ---------- | ------------------- | ----------------------------------- |
| `items`    | `[ChecklistItem!]!` | The checklist items on this page.   |
| `pageInfo` | `PageInfo!`         | Pagination metadata for the result. |

#### PageInfo

| Field             | Type       | Description                                        |
| ----------------- | ---------- | -------------------------------------------------- |
| `totalItems`      | `Int`      | Total items matching the filter, across all pages. |
| `totalPages`      | `Int`      | Total number of pages at the current `take`.       |
| `page`            | `Int`      | The current page number.                           |
| `perPage`         | `Int`      | Items per page (mirrors `take`).                   |
| `hasNextPage`     | `Boolean!` | Whether another page follows the current one.      |
| `hasPreviousPage` | `Boolean!` | Whether a page precedes the current one.           |

## Related

- [Checklist items](/api/checklists/checklist-items) — create, edit, assign, and schedule items.
- [Manage checklists](/api/checklists/manage-checklist) — rename, reposition, and delete checklists.
- [Update a record](/api/records/update-record)
- [Webhooks](/api/webhooks)
