---
title: Post and delete status updates
description: Post a status update to a workspace with createStatusUpdate, and remove one with deleteStatusUpdate.
icon: MessageSquarePlus
order: 1
---

A status update is a project-level health check-in: a traffic-light `category` (`GREEN`, `ORANGE`, or `RED`) plus dated rich-text commentary, attributed to the user who posted it. Use the `createStatusUpdate` mutation to post one to a workspace, and `deleteStatusUpdate` to remove one.

Status updates belong to a workspace (a `Project` in the API). They are **immutable**: there is no edit operation. To change a posted update, delete it and post a new one.

<Callout variant="info" title="No edit operation">

The API does not expose a mutation to edit a status update. Changing one means `deleteStatusUpdate` followed by a fresh `createStatusUpdate`.

</Callout>

## Request

Post a status update. All five input fields are required.

```graphql
mutation PostStatusUpdate {
  createStatusUpdate(
    input: {
      projectId: "project_123"
      category: ORANGE
      html: "<p>Migration is <strong>behind schedule</strong> — slipping to next sprint.</p>"
      text: "Migration is behind schedule — slipping to next sprint."
      date: "2026-05-29T00:00:00.000Z"
    }
  ) {
    id
    category
    text
    date
    createdAt
    user {
      id
      fullName
    }
  }
}
```

Include your authentication headers on every request:

- `X-Bloo-Token-ID` — your API token ID
- `X-Bloo-Token-Secret` — your API token secret
- `X-Bloo-Company-ID` — your organization ID or slug

`projectId` accepts a workspace ID or slug.

## Parameters

### CreateStatusUpdateInput

| Parameter   | Type                    | Required | Description                                                                                                                                                    |
| ----------- | ----------------------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `projectId` | `String!`               | Yes      | ID or slug of the workspace to post the update to.                                                                                                             |
| `category`  | `StatusUpdateCategory!` | Yes      | Traffic-light health indicator. One of `GREEN`, `ORANGE`, `RED`.                                                                                               |
| `html`      | `String!`               | Yes      | The rich-text body, as HTML. This is what renders in the app.                                                                                                  |
| `text`      | `String!`               | Yes      | Plain-text version of the body, used for previews, search, and notifications. Supply both `html` and `text` (the same dual-field convention used by comments). |
| `date`      | `DateTime!`             | Yes      | The as-of date the update reports for, as an ISO 8601 string. This is distinct from `createdAt` (when the record was posted) — it lets you backdate an update. |

### StatusUpdateCategory

| Value    | Meaning                    |
| -------- | -------------------------- |
| `GREEN`  | On track / healthy.        |
| `ORANGE` | At risk / needs attention. |
| `RED`    | Off track / blocked.       |

## Response

`createStatusUpdate` returns the created `StatusUpdate`.

```json
{
  "data": {
    "createStatusUpdate": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "category": "ORANGE",
      "text": "Migration is behind schedule — slipping to next sprint.",
      "date": "2026-05-29T00:00:00.000Z",
      "createdAt": "2026-05-29T14:22:05.118Z",
      "user": {
        "id": "clm4n8qwx000108l0a1b2c3d4",
        "fullName": "Jordan Lee"
      }
    }
  }
}
```

### Returns

`StatusUpdate` — selected fields:

| Field          | Type                    | Description                                        |
| -------------- | ----------------------- | -------------------------------------------------- |
| `id`           | `ID!`                   | Unique identifier of the status update.            |
| `category`     | `StatusUpdateCategory!` | The health indicator (`GREEN` / `ORANGE` / `RED`). |
| `html`         | `String!`               | The rich-text body as HTML.                        |
| `text`         | `String!`               | The plain-text body.                               |
| `date`         | `DateTime!`             | The reported as-of date.                           |
| `commentCount` | `Int!`                  | Number of comments on this update.                 |
| `isRead`       | `Boolean`               | Whether the requesting user has read it.           |
| `isSeen`       | `Boolean`               | Whether the requesting user has seen it.           |
| `user`         | `User!`                 | The user who posted it.                            |
| `project`      | `Project!`              | The workspace it belongs to.                       |
| `createdAt`    | `DateTime!`             | When the update was posted.                        |
| `updatedAt`    | `DateTime!`             | When the record was last modified.                 |

For the full field list and how to read updates back, see [Query and subscribe to status updates](/api/status-updates/query-status-updates).

### Side effects

Posting an update fires the usual collaboration side effects: it records an entry in the workspace activity feed, sends notifications to relevant members, and resolves any `@`-mentions in the `html` body into mention notifications.

## Delete a status update

Pass the status update's `id`. `deleteStatusUpdate` returns a `MutationResult` — a bare result envelope, so select only `success` and `operationId` (it has no other fields).

```graphql
mutation RemoveStatusUpdate {
  deleteStatusUpdate(id: "status_update_123") {
    success
    operationId
  }
}
```

```json
{
  "data": {
    "deleteStatusUpdate": {
      "success": true,
      "operationId": "clm4n9zzz000208l0p7q8r9s0"
    }
  }
}
```

Deleting an update also removes its comments and activity entries, and publishes a delete event to anyone subscribed to the workspace's status updates.

### MutationResult

| Field         | Type       | Description                          |
| ------------- | ---------- | ------------------------------------ |
| `success`     | `Boolean!` | `true` when the update was deleted.  |
| `operationId` | `String`   | Identifier for the delete operation. |

## Errors

| Code                      | Operation            | When                                                                                                       |
| ------------------------- | -------------------- | ---------------------------------------------------------------------------------------------------------- |
| `PROJECT_NOT_FOUND`       | `createStatusUpdate` | No workspace matches `projectId`.                                                                          |
| `STATUS_UPDATE_NOT_FOUND` | `deleteStatusUpdate` | No status update matches `id`.                                                                             |
| `FORBIDDEN`               | both                 | The caller's workspace role isn't allowed to post (or isn't the creator or a workspace OWNER, for delete). |
| `UNAUTHENTICATED`         | both                 | The request is missing or has invalid credentials.                                                         |

## Permissions

Both operations require the workspace to be **active** (not archived). Roles are `UserAccessLevel` values on the workspace.

- **Post** (`createStatusUpdate`): any role except `VIEW_ONLY` and `COMMENT_ONLY`.
- **Delete** (`deleteStatusUpdate`): you must be the update's **creator**, or hold the `OWNER` role on the workspace. `VIEW_ONLY` and `COMMENT_ONLY` are denied regardless.

Workspaces can hide the status-update feature entirely via the project-level `hideStatusUpdate` setting; the API operations still work even when the feature is hidden in the UI.

## Related

- [Status Updates overview](/api/status-updates)
- [Query and subscribe to status updates](/api/status-updates/query-status-updates)
- [Workspaces](/api/workspaces)
