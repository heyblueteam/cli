---
title: Workspaces
description: Workspaces are the top-level container for records, lists, custom fields, and automations. Create, list, and organize them via the projectList and createProject operations.
icon: FolderKanban
order: 0
---

A **workspace** is the top-level container for work in Blue: it holds lists, records, custom fields, tags, automations, and saved views. Every workspace belongs to one **organization**. In the API, a workspace is a `Project` object and an organization is a `Company` — the GraphQL surface uses those type names throughout.

This section covers creating workspaces, listing and filtering them, renaming, archiving, copying, deleting, and managing the lists inside them.

## Operations

| Operation                                                | GraphQL                        | Description                                     |
| -------------------------------------------------------- | ------------------------------ | ----------------------------------------------- |
| [Create a workspace](/api/workspaces/create-workspace)   | `createProject`                | Create a workspace, optionally from a template. |
| [List workspaces](/api/workspaces/list-workspaces)       | `projectList`                  | Filter, sort, and paginate workspaces.          |
| [Rename a workspace](/api/workspaces/rename-workspace)   | `updateProject`                | Change a workspace's name (and slug).           |
| [Archive a workspace](/api/workspaces/archive-workspace) | `updateProject`                | Archive or unarchive a workspace.               |
| [Copy a workspace](/api/workspaces/copy-workspace)       | `copyProject`                  | Duplicate an existing workspace.                |
| [Delete a workspace](/api/workspaces/delete-workspace)   | `deleteProject`                | Permanently delete a workspace.                 |
| [Lists](/api/workspaces/lists)                           | `todoLists` / `createTodoList` | Manage the lists inside a workspace.            |
| [Templates](/api/workspaces/templates)                   | `projectList`                  | Work with workspace templates.                  |
| [Activity](/api/workspaces/workspace-activity)           | `activityList`                 | Read a workspace's activity log.                |

<Callout variant="info" title="Workspace, project, record">

"Workspace" and "project" are the same thing — the UI says workspace, the schema says `Project`. Likewise a "record" is a `Todo` and an "organization" is a `Company`. This page uses the product nouns in prose and the schema names in code.

</Callout>

## Key concepts

### Structure

A workspace (`Project`) belongs to exactly one organization (`Company`). Inside it:

- **Lists** (`TodoList`) group records into columns or stages.
- **Records** (`Todo`) are the individual items of work, each owned by a list.
- **Custom fields**, **tags**, **automations**, and **saved views** are all scoped to the workspace.

Most workspace-scoped operations require the `X-Bloo-Project-ID` header in addition to the organization headers. Organization (`companyId`) and workspace (`projectId`) arguments accept either an ID or a slug.

### Categories

Each workspace has a `category` (`ProjectCategory`), used for grouping and filtering. It defaults to `GENERAL`. The complete set of values is:

```
CRM
CROSS_FUNCTIONAL
CUSTOMER_SUCCESS
DESIGN
ENGINEERING
GENERAL
HR
IT
MARKETING
OPERATIONS
PRODUCT
SALES
```

### Permissions

A user's access to a workspace is one of the `UserAccessLevel` values:

| Level          | Access                                           |
| -------------- | ------------------------------------------------ |
| `OWNER`        | Full control, including deleting the workspace.  |
| `ADMIN`        | Manage workspace settings, members, and content. |
| `MEMBER`       | Create and edit records and lists.               |
| `CLIENT`       | Restricted access for external collaborators.    |
| `COMMENT_ONLY` | Read content and add comments only.              |
| `VIEW_ONLY`    | Read-only access.                                |

The user whose token makes a `createProject` request is automatically added to the new workspace as its `OWNER`.

## Create a workspace

Use the `createProject` mutation to create a workspace. Only `companyId` and `name` are required; pass `templateId` to seed the workspace from a template, or `category` to classify it.

```graphql
mutation CreateWorkspace {
  createProject(
    input: { companyId: "company_123", name: "Q1 Marketing Campaign", category: MARKETING }
  ) {
    id
    name
    slug
    category
  }
}
```

```json
{
  "data": {
    "createProject": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Q1 Marketing Campaign",
      "slug": "q1-marketing-campaign",
      "category": "MARKETING"
    }
  }
}
```

`createProject` returns a nullable `Project`. The workspace `name` is capped at 50 characters. See [Create a workspace](/api/workspaces/create-workspace) for the full input reference, including `folderId`, `description`, `color`, `icon`, and `coverConfig`.

## List workspaces

Use the `projectList` query to filter, sort, and paginate workspaces. The filter's `companyIds` is required; everything else is optional.

```graphql
query ListWorkspaces {
  projectList(
    filter: { companyIds: ["company_123"], archived: false }
    sort: [updatedAt_DESC]
    take: 20
  ) {
    items {
      id
      name
      slug
      category
      archived
      updatedAt
    }
    pageInfo {
      totalItems
      hasNextPage
    }
  }
}
```

```json
{
  "data": {
    "projectList": {
      "items": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "name": "Q1 Marketing Campaign",
          "slug": "q1-marketing-campaign",
          "category": "MARKETING",
          "archived": false,
          "updatedAt": "2026-05-29T10:14:00.000Z"
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

`sort` takes `ProjectSort` enum values (`name_ASC`, `updatedAt_DESC`, `createdAt_DESC`, `position_ASC`, …), and `pageInfo.totalItems` holds the total count. To filter by category, query with `archived` / `isTemplate` / `search` and filter the results client-side — `ProjectListFilter` does not accept a category filter. See [List workspaces](/api/workspaces/list-workspaces) for the full filter and pagination reference.

<Callout variant="warning" title="The projects query is deprecated">

An older top-level `projects(companyId:, archived:, …)` query still exists but is `@deprecated`. Use `projectList` for all new integrations.

</Callout>

## Manage lists in a workspace

There are two ways to read the lists in a workspace.

For a simple, unpaginated read of every list in one workspace, use the `todoLists(projectId:)` query:

```graphql
query WorkspaceLists {
  todoLists(projectId: "project_123") {
    id
    title
    position
    todosCount
  }
}
```

For filtering across workspaces, or for pagination, use `todoListQueries.todoLists` with a `TodoListsFilterInput` (its `companyIds` is required):

```graphql
query PaginatedLists {
  todoListQueries {
    todoLists(filter: { companyIds: ["company_123"], projectIds: ["project_123"] }, take: 20) {
      items {
        id
        title
        position
      }
      pageInfo {
        totalItems
        hasNextPage
      }
    }
  }
}
```

To create a list, use the `createTodoList` mutation. All three inputs — `projectId`, `title`, and `position` — are required:

```graphql
mutation CreateList {
  createTodoList(input: { projectId: "project_123", title: "To Do", position: 1.0 }) {
    id
    title
    position
  }
}
```

A workspace can hold at most 50 lists; creating a 51st fails with `FORBIDDEN`. See [Lists](/api/workspaces/lists) for the full reference.

## Errors

| Code                   | When                                                                                                    |
| ---------------------- | ------------------------------------------------------------------------------------------------------- |
| `BAD_USER_INPUT`       | Invalid input — e.g. a workspace `name` longer than 50 characters, or a `folderId` that does not exist. |
| `COMPANY_NOT_FOUND`    | The `companyId` does not match an organization you can access.                                          |
| `PROJECT_NOT_FOUND`    | The referenced workspace does not exist or you cannot access it.                                        |
| `FORBIDDEN`            | Your access level is too low for the operation, or you have hit the 50-list-per-workspace limit.        |
| `CREATE_PROJECT_LIMIT` | A template-based `createProject` would exceed the maximum workspace size (250,000 records).             |

## Related

- [Records](/api/records) — the records (`Todo`) that live inside a workspace's lists.
- [Custom fields](/api/custom-fields) — add structured fields to a workspace's records.
- [Automations](/api/automations) — automate work within a workspace.
- [User management](/api/user-management) — manage who can access a workspace and at what level.
