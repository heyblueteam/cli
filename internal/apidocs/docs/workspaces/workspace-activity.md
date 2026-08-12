---
title: Workspace Activity
description: Read the activity feed for a workspace or organization with the activityList query, with filtering by user, tag, category, and date.
icon: Activity
order: 9
---

Use the `activityList` query to read the activity feed — the chronological audit trail of actions that happened in a workspace or across an organization (records created, comments posted, users invited, lists moved, and so on). Workspaces are `Workspace` objects and organizations are `Organization` objects in the API; activity entries are `Activity` objects.

Activities are generated automatically by the server — there is no mutation to create one. This query is the read side; for a live stream of new activity over WebSocket, use the [`subscribeToActivity` subscription](/api/realtime/activity-notifications).

## Request

The smallest useful call scopes the feed to one workspace and takes the most recent page. The feed is always scoped to the organization in your `blue-org-id` header — see [Permissions](#permissions).

```graphql
query WorkspaceActivity {
  activityList(projectId: "workspace_123", first: 20) {
    activities {
      id
      category
      html
      createdAt
      createdBy {
        id
        fullName
        email
      }
    }
    pageInfo {
      hasNextPage
    }
    totalCount
  }
}
```

`projectId` accepts a workspace ID or slug. Omit it to read the organization-wide feed across every workspace you can see.

## Parameters

All arguments are optional. With no arguments, `activityList` returns the most recent 20 entries across your accessible workspaces in the organization bound by your `blue-org-id` header.

| Parameter    | Type                   | Description                                                                                                                                                                                     |
| ------------ | ---------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `projectId`  | `String`               | Scope to a single workspace (ID or slug).                                                                                                                                                       |
| `projectIds` | `[String!]`            | Scope to several workspaces by ID.                                                                                                                                                              |
| `companyId`  | `String`               | Must equal the organization in your `blue-org-id` header, or the call is rejected with `FORBIDDEN`. The feed is always scoped to the authenticated organization regardless of this value. |
| `userId`     | `String`               | Filter to a single user (created, affected by, assigned to, or @-mentioned in the activity).                                                                                                    |
| `userIds`    | `[String!]`            | Filter to several users (same matching rules as `userId`).                                                                                                                                      |
| `tagIds`     | `[String!]`            | Filter to activities whose record carries one of these tags.                                                                                                                                    |
| `categories` | `[ActivityCategory!]`  | Filter to specific activity types. See [ActivityCategory](#activitycategory).                                                                                                                   |
| `filter`     | `TodoFilterInput`      | Restrict to activities whose record matches a record filter. Only the `fields` and `unassigned` parts of the filter are applied here.                                                           |
| `startDate`  | `DateTime`             | Only activities at or after this timestamp (ISO 8601).                                                                                                                                          |
| `endDate`    | `DateTime`             | Only activities at or before this timestamp (ISO 8601).                                                                                                                                         |
| `first`      | `Int`                  | Page size. Defaults to `20`.                                                                                                                                                                    |
| `skip`       | `Int`                  | Offset: skip this many entries before returning the page.                                                                                                                                       |
| `after`      | `String`               | Return entries whose `id` sorts before this cursor (use the `id` of the last entry from the previous page).                                                                                     |
| `orderBy`    | `ActivityOrderByInput` | Sort order. Defaults to `createdAt_DESC`. See [ActivityOrderByInput](#activityorderbyinput).                                                                                                    |

<Callout variant="info" title="Pagination">

`activityList` is `id`-cursor paginated. Read a page with `first`, then pass the `id` of the last entry as `after` to get the next page; stop when `pageInfo.hasNextPage` is `false`. `skip` works for simple offset paging. The `last` and `before` arguments are accepted by the schema but ignored by this query, and the `startCursor` / `endCursor` fields on `PageInfo` are deprecated — page by entry `id` instead.

</Callout>

### ActivityCategory

The category of an activity. The full enum:

| Value                         | Meaning                                     |
| ----------------------------- | ------------------------------------------- |
| `CREATE_TODO`                 | A record was created.                       |
| `MARK_TODO_AS_COMPLETE`       | A record was marked complete.               |
| `MOVE_TODO`                   | A record was moved between lists.           |
| `COPY_TODO`                   | A record was copied.                        |
| `REPEAT_TODO`                 | A recurring record produced a new instance. |
| `REMOVE_TODO`                 | A record was deleted.                       |
| `CREATE_TODO_LIST`            | A list was created.                         |
| `REMOVE_TODO_LIST`            | A list was deleted.                         |
| `CREATE_COMMENT`              | A comment was added.                        |
| `CREATE_DISCUSSION`           | A chat was started.                   |
| `CREATE_STATUS_UPDATE`        | A status update was posted.                 |
| `CREATE_CUSTOM_FIELD`         | A custom field was created.                 |
| `RECEIVE_FORM`                | A form submission was received.             |
| `CREATE_INVITATION`           | A user was invited.                         |
| `ACCEPT_INVITATION`           | An invitation was accepted.                 |
| `REJECT_INVITATION`           | An invitation was rejected.                 |
| `CANCEL_INVITATION`           | An invitation was cancelled.                |
| `ADD_USER_TO_PROJECT`         | A user was added to a workspace.            |
| `REMOVE_USER_FROM_PROJECT`    | A user was removed from a workspace.        |
| `LEAVE_PROJECT`               | A user left a workspace.                    |
| `REMOVE_USER_FROM_COMPANY`    | A user was removed from the organization.   |
| `LEAVE_COMPANY`               | A user left the organization.               |
| `UPDATE_PROJECT_ACCESS_LEVEL` | A user's workspace access level changed.    |
| `UPDATE_COMPANY_ACCESS_LEVEL` | A user's organization access level changed. |
| `ARCHIVE_PROJECT`             | A workspace was archived.                   |
| `UNARCHIVE_PROJECT`           | A workspace was unarchived.                 |
| `EDIT_NAME`                   | A user changed their display name.          |
| `EDIT_JOB_TITLE`              | A user changed their job title.             |

### ActivityOrderByInput

Sort order. `createdAt_DESC` (newest first) is the default and the usual choice for a feed.

| Value                            | Meaning                              |
| -------------------------------- | ------------------------------------ |
| `createdAt_DESC`                 | Newest first (default).              |
| `createdAt_ASC`                  | Oldest first.                        |
| `updatedAt_DESC`                 | Most recently updated first.         |
| `updatedAt_ASC`                  | Least recently updated first.        |
| `category_ASC` / `category_DESC` | By category, alphabetical / reverse. |

The enum also accepts `id_*`, `uid_*`, `inviteeEmail_*`, `metadata_*`, and `userAccessLevel_*` (each in `_ASC` / `_DESC`), but those are rarely useful for a chronological feed.

## Response

```json
{
  "data": {
    "activityList": {
      "activities": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "category": "CREATE_TODO",
          "html": "<b>Ada Lovelace</b> created the record <b>Draft Q3 roadmap</b>",
          "createdAt": "2026-05-29T14:22:05.000Z",
          "createdBy": {
            "id": "clm4n8qwx000108l0a1b2c3d4",
            "fullName": "Ada Lovelace",
            "email": "ada@example.com"
          }
        }
      ],
      "pageInfo": { "hasNextPage": true },
      "totalCount": 20
    }
  }
}
```

<Callout variant="warning" title="totalCount is the page count, not the grand total">

`totalCount` reflects the number of entries in the page just returned (capped by `first`), not the total number of activities matching your filters. To know whether more pages exist, read `pageInfo.hasNextPage`.

</Callout>

### ActivityList

| Field        | Type           | Description                                         |
| ------------ | -------------- | --------------------------------------------------- |
| `activities` | `[Activity!]!` | The page of activity entries.                       |
| `pageInfo`   | `PageInfo!`    | Pagination info. Read `hasNextPage` to keep paging. |
| `totalCount` | `Int!`         | Number of entries in this page (see callout above). |

### Activity

| Field          | Type                | Description                                                                  |
| -------------- | ------------------- | ---------------------------------------------------------------------------- |
| `id`           | `ID!`               | Unique identifier. Use it as the `after` cursor for the next page.           |
| `uid`          | `String!`           | Short human-readable identifier.                                             |
| `category`     | `ActivityCategory!` | What happened. See [ActivityCategory](#activitycategory).                    |
| `html`         | `String!`           | Rendered description of the activity, as HTML.                               |
| `text`         | `String!`           | Plain-text description. **Deprecated** — use `html`.                         |
| `isSeen`       | `Boolean!`          | Whether the signed-in user has seen this activity (drives the unread badge). |
| `isRead`       | `Boolean!`          | Whether the signed-in user has marked this activity read.                    |
| `createdAt`    | `DateTime!`         | When the action occurred.                                                    |
| `updatedAt`    | `DateTime!`         | When the entry was last updated.                                             |
| `createdBy`    | `User!`             | The user who performed the action.                                           |
| `affectedBy`   | `User`              | The user the action was performed on (e.g. the invited or removed user).     |
| `inviteeEmail` | `String`            | For invitation activities, the invited email address.                        |
| `company`      | `Organization`      | The organization the activity belongs to.                                    |
| `project`      | `Workspace`         | The workspace, when the activity is workspace-scoped.                        |
| `todo`         | `Record`            | The record, for record activities.                                           |
| `todoList`     | `RecordList`        | The list, for list activities.                                               |
| `comment`      | `Comment`           | The comment, for `CREATE_COMMENT`.                                           |
| `chat`   | `Chat`        | The chat, for `CREATE_DISCUSSION`.                                     |
| `statusUpdate` | `StatusUpdate`      | The status update, for `CREATE_STATUS_UPDATE`.                               |
| `metadata`     | `String`            | Extra context for the activity, as a JSON string.                            |

Select `User` subfields by name — the type exposes `fullName`, `firstName`, `lastName`, `username`, and `email`; there is no `name` field.

## Full example

Read every record-creation and completion event in one workspace from a single user during March 2026, oldest first, paging in batches of 50.

```graphql
query WorkspaceActivityFiltered {
  activityList(
    projectId: "workspace_123"
    userIds: ["user_123"]
    categories: [CREATE_TODO, MARK_TODO_AS_COMPLETE]
    startDate: "2026-03-01T00:00:00Z"
    endDate: "2026-03-31T23:59:59Z"
    first: 50
    orderBy: createdAt_ASC
  ) {
    activities {
      id
      category
      html
      createdAt
      createdBy {
        id
        fullName
      }
      affectedBy {
        id
        fullName
      }
      todo {
        id
        title
      }
      project {
        id
        name
        slug
      }
    }
    pageInfo {
      hasNextPage
    }
    totalCount
  }
}
```

To fetch the next page, pass the `id` of the last entry as `after: "<last-id>"` and keep the other arguments unchanged.

## Errors

| Code                | When                                                                                                       |
| ------------------- | ---------------------------------------------------------------------------------------------------------- |
| `UNAUTHENTICATED`   | The request has no valid token.                                                                            |
| `COMPANY_NOT_FOUND` | The `blue-org-id` header is missing or invalid, or you are not a member of that organization.        |
| `FORBIDDEN`         | A `companyId` argument was passed that does not match the organization in your `blue-org-id` header. |
| `BAD_USER_INPUT`    | An argument is malformed — for example a `startDate` or `endDate` that is not a valid ISO 8601 timestamp.  |

A workspace you cannot see, an archived workspace, or one with activity tracking disabled does **not** error — those activities are silently excluded from the feed.

## Permissions

The feed is scoped to the organization in your `blue-org-id` header and to the workspaces you belong to within it; a `companyId` argument cannot widen that scope.

- You only see activities from workspaces you are a member of and that have activity tracking enabled. Entries from before you joined the organization are excluded.
- Activities that reference a custom field your role cannot view are filtered out, unless you are an organization `OWNER` or `ADMIN`.
- Membership-change categories (`REMOVE_USER_FROM_PROJECT`, `LEAVE_PROJECT`, `REMOVE_USER_FROM_COMPANY`, `LEAVE_COMPANY`) are only visible to workspace or organization admins. In the organization-wide feed, `REMOVE_TODO` entries are also hidden.

## Related

- [Stream live activity](/api/realtime/activity-notifications) — the `subscribeToActivity` subscription for new entries over WebSocket.
- [List records](/api/records/list-records) — query the records the activity feed references.
- [Comments](/api/comments/query-comments) — read the comments behind `CREATE_COMMENT` activities.
- [Workspaces overview](/api/workspaces) — the rest of the workspace operations.
