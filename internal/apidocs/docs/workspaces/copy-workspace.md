---
title: Copy a Workspace
description: Duplicate a workspace within an organization or across organizations, choosing exactly which elements to include.
icon: Copy
order: 5
---

Use the `copyWorkspace` mutation to duplicate an existing workspace, either within the same organization or into a different one. It is the way to build reusable workspace templates, stand up similar projects quickly, or move work between organizations. Workspaces are `Workspace` objects in the API.

Copying runs asynchronously: the mutation queues a background job and returns immediately. Track progress with the `copyProjectStatus` query.

## Request

The smallest call needs the source workspace `projectId`, a `name` for the copy, and an `options` object selecting what to include.

```graphql
mutation CopyWorkspace {
  copyWorkspace(
    input: {
      projectId: "project_123"
      name: "New workspace copy"
      options: { todos: true, todoLists: true, people: true }
    }
  )
}
```

The `projectId` and `companyId` arguments accept either an ID or a slug.

## Parameters

### CopyWorkspaceInput

| Parameter     | Type                       | Required | Description                                                                                            |
| ------------- | -------------------------- | -------- | ------------------------------------------------------------------------------------------------------ |
| `projectId`   | `String!`                  | Yes      | ID or slug of the workspace to copy.                                                                   |
| `name`        | `String!`                  | Yes      | Name for the new workspace (max 50 characters). Trimmed, and any URLs are stripped out.                |
| `description` | `String`                   | No       | Description for the new workspace (max 500 characters).                                                |
| `imageURL`    | `String`                   | No       | Cover image URL for the new workspace.                                                                 |
| `companyId`   | `String`                   | No       | ID or slug of the organization to create the copy in. Defaults to the source workspace's organization. |
| `options`     | `CopyProjectOptionsInput!` | Yes      | Which elements to copy. See below.                                                                     |

### CopyProjectOptionsInput

Every option is an optional `Boolean`; omit one or set it to `false` to skip that element. Some options only take effect when a related option is also enabled — see [Notes](#notes).

| Parameter              | Type      | Description                                                      |
| ---------------------- | --------- | ---------------------------------------------------------------- |
| `todos`                | `Boolean` | Copy records (the `Record` objects in each list).                |
| `todoLists`            | `Boolean` | Copy lists (the `RecordList` columns/sections).                  |
| `todoActions`          | `Boolean` | Copy record actions.                                             |
| `todoComments`         | `Boolean` | Copy comments on records.                                        |
| `checklists`           | `Boolean` | Copy record checklists.                                          |
| `dueDates`             | `Boolean` | Copy record due dates.                                           |
| `coverConfig`          | `Boolean` | Copy record cover-image configuration.                           |
| `assignees`            | `Boolean` | Copy record assignees. Requires `people: true`.                  |
| `people`               | `Boolean` | Copy workspace members.                                          |
| `projectUserRoles`     | `Boolean` | Copy member roles and permissions. Requires `people: true`.      |
| `tags`                 | `Boolean` | Copy workspace tags.                                             |
| `customFields`         | `Boolean` | Copy custom field definitions and their values.                  |
| `files`                | `Boolean` | Copy file attachments.                                           |
| `forms`                | `Boolean` | Copy workspace forms.                                            |
| `automations`          | `Boolean` | Copy workspace automations.                                      |
| `chats`                | `Boolean` | Copy workspace chats. Legacy spelling: `discussions`.            |
| `chatComments`         | `Boolean` | Copy comments on chats. Requires `chats: true`. Legacy spelling: `discussionComments`. |
| `statusUpdates`        | `Boolean` | Copy workspace status updates.                                   |
| `statusUpdateComments` | `Boolean` | Copy comments on status updates. Requires `statusUpdates: true`. |

## Response

`copyWorkspace` returns a nullable `Boolean`. It returns `true` once the copy job has been queued.

```json
{
  "data": {
    "copyWorkspace": true
  }
}
```

## Full example

Copy a workspace into a different organization with a richer set of options.

```graphql
mutation CopyWorkspace {
  copyWorkspace(
    input: {
      projectId: "project_123"
      name: "Q2 Marketing Campaign"
      description: "Copy of Q1 campaign with an updated timeline"
      imageURL: "https://example.com/campaign-cover.png"
      companyId: "company_123"
      options: {
        todos: true
        todoLists: true
        todoActions: true
        checklists: true
        dueDates: true
        coverConfig: true
        people: true
        assignees: true
        projectUserRoles: true
        tags: true
        customFields: true
        files: true
        forms: true
        automations: true
        chats: false
        chatComments: false
        statusUpdates: false
        statusUpdateComments: false
      }
    }
  )
}
```

## Checking copy status

Because the copy runs in the background, poll the `copyProjectStatus` query to follow progress. It returns the status of the calling user's current copy job, or `null` when none is in progress.

```graphql
query CheckCopyStatus {
  copyProjectStatus {
    queuePosition
    totalQueues
    isActive
    isTemplate
    newProjectName
    oldProject {
      id
      name
    }
    oldCompany {
      id
      name
    }
    newCompany {
      id
      name
    }
  }
}
```

### CopyProjectStatus fields

| Field            | Type           | Description                                                            |
| ---------------- | -------------- | ---------------------------------------------------------------------- |
| `queuePosition`  | `Int`          | Position of this copy job in the queue.                                |
| `totalQueues`    | `Int`          | Total number of copy jobs currently queued.                            |
| `isActive`       | `Boolean`      | Whether the copy is actively running (vs. still waiting in the queue). |
| `oldProject`     | `Workspace`    | The source workspace being copied.                                     |
| `newProjectName` | `String`       | Name of the new workspace being created.                               |
| `isTemplate`     | `Boolean`      | Whether the copy is being created as a template.                       |
| `oldCompany`     | `Organization` | Source organization.                                                   |
| `newCompany`     | `Organization` | Target organization.                                                   |

```json
{
  "data": {
    "copyProjectStatus": {
      "queuePosition": 1,
      "totalQueues": 3,
      "isActive": true,
      "isTemplate": false,
      "newProjectName": "Q2 Marketing Campaign",
      "oldProject": { "id": "clm4n8qwx000008l0g4oxdqn7", "name": "Q1 Marketing Campaign" },
      "oldCompany": { "id": "clm4n8qwx000108l0a1b2c3d4", "name": "Acme Inc" },
      "newCompany": { "id": "clm4n8qwx000208l0e5f6g7h8", "name": "Acme EU" }
    }
  }
}
```

## Errors

| Code                   | When                                                                                                                |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------- |
| `FORBIDDEN`            | You lack `OWNER` or `ADMIN` access to the source workspace, or it is archived.                                      |
| `PROJECT_NOT_FOUND`    | The source workspace doesn't exist or you can't access it. Message: `Project was not found.`                        |
| `COMPANY_NOT_FOUND`    | The target organization (`companyId`) doesn't exist or you aren't a member of it. Message: `Company was not found.` |
| `CREATE_PROJECT_LIMIT` | The source workspace has more than 250,000 records. Message: `Project size exceeds maximum limit.`                  |

A copy that is rejected because you already have one in progress returns a generic error with the message `Oops!` and no extension code.

```json
{
  "errors": [
    {
      "message": "Project size exceeds maximum limit.",
      "extensions": {
        "code": "CREATE_PROJECT_LIMIT"
      }
    }
  ]
}
```

## Permissions

You must hold an `OWNER` or `ADMIN` role in the **source** workspace, and the source workspace must be active (not archived).

| Scenario                           | Required access                                                                        |
| ---------------------------------- | -------------------------------------------------------------------------------------- |
| Copy within the same organization  | `OWNER` or `ADMIN` in the source workspace.                                            |
| Copy into a different organization | `OWNER` or `ADMIN` in the source workspace, and membership of the target organization. |

## Notes

- **One copy at a time.** Each user can have a single copy job in progress. Starting another while one is running fails with `Oops!`.
- **Size limit.** Workspaces with more than 250,000 records cannot be copied.
- **Dependent options.** A few options only apply when their parent is also enabled: `assignees` and `projectUserRoles` require `people: true`; `chatComments` requires `chats: true`; `statusUpdateComments` requires `statusUpdates: true`.
- **Name handling.** The new workspace name is trimmed and has any URLs removed before the copy runs.
- **Status retention.** A copy job's status stays available via `copyProjectStatus` for up to 6 hours.

## Related

- [Create a workspace](/api/workspaces/create-workspace)
- [List workspaces](/api/workspaces/list-workspaces)
- [Rename a workspace](/api/workspaces/rename-workspace)
- [Archive a workspace](/api/workspaces/archive-workspace)
- [Workspace templates](/api/workspaces/templates)
