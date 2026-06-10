---
title: Query and subscribe to status updates
description: Fetch a status update by id, list a workspace's updates with filtering, ordering, and cursor pagination, and subscribe to real-time create and delete events.
icon: ListChecks
order: 2
---

Read status updates with two queries and one subscription. Use the `statusUpdate` query to fetch a single update by id, the `statusUpdateList` query to page through a workspace's updates with user and date filtering plus ordering, and the `subscribeToStatusUpdate` subscription to receive create and delete events in real time. Status updates are `StatusUpdate` objects in the API; in the app they appear as the colored health check-ins on a [workspace](/api/workspaces) (a `Project`).

The `StatusUpdate` type, the `StatusUpdateCategory` health enum, and the immutability rule (there is no edit operation) are documented once on the [section overview](/api/status-updates). To create or delete updates, see [Post and delete status updates](/api/status-updates/manage-status-updates).

## Fetch one status update

Use `statusUpdate(id: String!): StatusUpdate!` to read a single update. The return type is non-null — a missing id raises `STATUS_UPDATE_NOT_FOUND` rather than returning `null`.

### Request

```graphql
query GetStatusUpdate {
  statusUpdate(id: "status_update_123") {
    id
    category
    html
    text
    date
    commentCount
    user {
      id
      fullName
      email
    }
    project {
      id
      name
    }
    createdAt
    updatedAt
  }
}
```

### Parameters

| Argument | Type      | Required | Description                          |
| -------- | --------- | -------- | ------------------------------------ |
| `id`     | `String!` | Yes      | The id of the status update to read. |

### Response

```json
{
  "data": {
    "statusUpdate": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "category": "GREEN",
      "html": "<p>Sprint is on track — all critical paths green.</p>",
      "text": "Sprint is on track — all critical paths green.",
      "date": "2026-05-29T00:00:00.000Z",
      "commentCount": 2,
      "user": {
        "id": "clm4n8qwx000108l0haz1f2pe",
        "fullName": "Dana Reyes",
        "email": "dana@example.com"
      },
      "project": {
        "id": "clm4n8qwx000208l0c1d2e3f4",
        "name": "Q3 Launch"
      },
      "createdAt": "2026-05-29T14:02:11.000Z",
      "updatedAt": "2026-05-29T14:02:11.000Z"
    }
  }
}
```

### StatusUpdate

The fields you'll select most often. The full type is documented on the [section overview](/api/status-updates).

| Field          | Type                    | Description                                                                                      |
| -------------- | ----------------------- | ------------------------------------------------------------------------------------------------ |
| `id`           | `ID!`                   | Unique identifier.                                                                               |
| `html`         | `String!`               | Commentary as sanitized rich-text HTML.                                                          |
| `text`         | `String!`               | Commentary as plain text — the fallback for the same body as `html`.                             |
| `date`         | `DateTime!`             | The reporting date the update is _for_. Distinct from `createdAt` — the UI lets you backdate it. |
| `category`     | `StatusUpdateCategory!` | The health signal: `GREEN`, `ORANGE`, or `RED`.                                                  |
| `user`         | `User!`                 | The author. Select `id`, `fullName`, `email` (a `User` has no `name`).                           |
| `project`      | `Project!`              | The workspace this update belongs to.                                                            |
| `commentCount` | `Int!`                  | Number of comments on this update.                                                               |
| `isRead`       | `Boolean`               | Whether the calling user has read this update.                                                   |
| `isSeen`       | `Boolean`               | Whether the calling user has seen this update in the feed.                                       |
| `createdAt`    | `DateTime!`             | When the update was posted.                                                                      |
| `updatedAt`    | `DateTime!`             | When the row was last touched.                                                                   |

<Callout variant="info" title="Reading comments">

`StatusUpdate.comments` is deprecated — read comments with the [comment operations](/api/comments) instead (a `Comment` with `category: STATUS_UPDATE` and `categoryId` set to the update's id). The `commentCount` field above is the supported way to get the count inline.

</Callout>

## List a workspace's status updates

Use `statusUpdateList` to page through every status update in one workspace, optionally narrowed to a single author or a reporting-date range. The query returns a `StatusUpdateList` envelope: the page of updates plus pagination metadata and a total count.

### Request

`projectId` is the only required argument. Without an `orderBy`, results come back newest-first by creation date.

```graphql
query ListStatusUpdates {
  statusUpdateList(projectId: "project_123") {
    statusUpdates {
      id
      category
      date
      user {
        id
        fullName
      }
    }
    pageInfo {
      totalItems
      hasNextPage
      endCursor
    }
    totalCount
  }
}
```

The `projectId` accepts a workspace **ID or slug**.

### Parameters

| Argument    | Type                       | Required | Description                                                                                                     |
| ----------- | -------------------------- | -------- | --------------------------------------------------------------------------------------------------------------- |
| `projectId` | `String!`                  | Yes      | The workspace whose updates to list (ID or slug).                                                               |
| `userId`    | `String`                   | No       | Return only updates authored by this exact user.                                                                |
| `dateFrom`  | `DateTime`                 | No       | Lower bound (inclusive) on the update's reporting `date`.                                                       |
| `dateTo`    | `DateTime`                 | No       | Upper bound (inclusive) on the update's reporting `date`.                                                       |
| `first`     | `Int`                      | No       | Page size. Omit to get **20**.                                                                                  |
| `after`     | `String`                   | No       | Cursor for forward pagination — pass a previous page's `endCursor` (the last item's id) to fetch the next page. |
| `before`    | `String`                   | No       | Cursor for the relay-style pagination shape.                                                                    |
| `last`      | `Int`                      | No       | Companion to `before` in the relay-style shape.                                                                 |
| `skip`      | `Int`                      | No       | Items to skip — feeds the `page` / `hasPreviousPage` math in `pageInfo`. Defaults to `0`.                       |
| `orderBy`   | `StatusUpdateOrderByInput` | No       | Sort key. Defaults to `createdAt_DESC` (newest first) when omitted.                                             |

<Callout variant="info" title="Cursor pagination">

`statusUpdateList` paginates by cursor, not by `skip`/`limit`. To walk the list, read `pageInfo.endCursor` from one page and pass it as `after` on the next call (the resolver sets the cursor to that id and skips it, so you get the rows that follow). Stop when `pageInfo.hasNextPage` is `false`.

</Callout>

### StatusUpdateOrderByInput

Each value is a `field_DIRECTION` enum member.

| Value                              | Sorts by                                                    |
| ---------------------------------- | ----------------------------------------------------------- |
| `createdAt_ASC` / `createdAt_DESC` | When the update was posted (**default: `createdAt_DESC`**). |
| `date_ASC` / `date_DESC`           | The reporting date.                                         |
| `category_ASC` / `category_DESC`   | Health category.                                            |
| `updatedAt_ASC` / `updatedAt_DESC` | When the row was last touched.                              |
| `html_ASC` / `html_DESC`           | Rich-text body.                                             |
| `text_ASC` / `text_DESC`           | Plain-text body.                                            |
| `id_ASC` / `id_DESC`               | Internal id.                                                |
| `uid_ASC` / `uid_DESC`             | Short identifier.                                           |

### Response

```json
{
  "data": {
    "statusUpdateList": {
      "statusUpdates": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "category": "GREEN",
          "date": "2026-05-29T00:00:00.000Z",
          "user": {
            "id": "clm4n8qwx000108l0haz1f2pe",
            "fullName": "Dana Reyes"
          }
        },
        {
          "id": "clm4n8qwx000308l0g9h8i7j6",
          "category": "ORANGE",
          "date": "2026-05-22T00:00:00.000Z",
          "user": {
            "id": "clm4n8qwx000108l0haz1f2pe",
            "fullName": "Dana Reyes"
          }
        }
      ],
      "pageInfo": {
        "totalItems": 14,
        "hasNextPage": true,
        "endCursor": "clm4n8qwx000308l0g9h8i7j6"
      },
      "totalCount": 14
    }
  }
}
```

### StatusUpdateList

| Field           | Type               | Description                                          |
| --------------- | ------------------ | ---------------------------------------------------- |
| `statusUpdates` | `[StatusUpdate!]!` | The updates on this page.                            |
| `pageInfo`      | `PageInfo!`        | Pagination metadata (below).                         |
| `totalCount`    | `Int!`             | Total updates matching the filter, across all pages. |

### PageInfo

| Field             | Type       | Description                                                                 |
| ----------------- | ---------- | --------------------------------------------------------------------------- |
| `totalItems`      | `Int`      | Total updates matching the filter (mirrors `totalCount`).                   |
| `totalPages`      | `Int`      | Total pages at the current page size.                                       |
| `page`            | `Int`      | Current page, derived from `skip` and the page size.                        |
| `perPage`         | `Int`      | Page size in effect.                                                        |
| `hasNextPage`     | `Boolean!` | Whether a next page exists.                                                 |
| `hasPreviousPage` | `Boolean!` | Whether a previous page exists.                                             |
| `endCursor`       | `String`   | The last item's id on this page. Pass it as `after` to fetch the next page. |
| `startCursor`     | `String`   | Deprecated.                                                                 |

### Full example

List the `ORANGE` and `RED` updates a single user reported across May 2026, oldest-first, ten per page.

```graphql
query ListStatusUpdatesFiltered {
  statusUpdateList(
    projectId: "project_123"
    userId: "user_123"
    dateFrom: "2026-05-01T00:00:00Z"
    dateTo: "2026-05-31T23:59:59Z"
    orderBy: date_ASC
    first: 10
  ) {
    statusUpdates {
      id
      category
      html
      text
      date
      commentCount
      isRead
      user {
        id
        fullName
        email
      }
      createdAt
    }
    pageInfo {
      totalItems
      totalPages
      page
      perPage
      hasNextPage
      hasPreviousPage
      endCursor
    }
    totalCount
  }
}
```

The `dateFrom` / `dateTo` bounds filter the reporting `date`, not `createdAt` — so backdated updates are matched by the date they report _for_.

## Subscribe to status updates

Use the `subscribeToStatusUpdate(projectId: String!): StatusUpdateSubscriptionPayload` subscription to receive an event whenever a status update is created or deleted in a workspace. Subscriptions run over WebSocket at `wss://api.blue.app/graphql`.

Events are scoped to one workspace and delivered only to project members — the server filters each event by whether the subscriber is in that project's user list. A non-member receives nothing.

### Request

```graphql
subscription OnStatusUpdate {
  subscribeToStatusUpdate(projectId: "project_123") {
    mutation
    node {
      id
      category
      date
      html
      user {
        id
        fullName
      }
    }
    previousValues {
      id
      category
      date
    }
    updatedFields
  }
}
```

### Parameters

| Argument    | Type      | Required | Description                                        |
| ----------- | --------- | -------- | -------------------------------------------------- |
| `projectId` | `String!` | Yes      | The workspace to receive status-update events for. |

### Payload

Each event is a `StatusUpdateSubscriptionPayload`.

| Field            | Type                         | Description                                                                                                            |
| ---------------- | ---------------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `mutation`       | `MutationType!`              | The event kind: `CREATED` or `DELETED`. (See note below.)                                                              |
| `node`           | `StatusUpdate`               | The affected status update. Populated on create; `null` once the row is gone on delete.                                |
| `previousValues` | `StatusUpdatePreviousValues` | The prior scalar values (`id`, `uid`, `html`, `text`, `date`, `category`, `createdAt`, `updatedAt`). Useful on delete. |
| `updatedFields`  | `[String!]`                  | Names of changed fields, when applicable.                                                                              |

<Callout variant="info" title="Create and delete only">

`MutationType` includes `UPDATED`, but the public API has no edit operation for status updates (they're immutable — delete and re-create to change one), so in practice this subscription fires only `CREATED` and `DELETED` events. Don't build on an `UPDATED` event arriving.

</Callout>

### Example event

```json
{
  "data": {
    "subscribeToStatusUpdate": {
      "mutation": "CREATED",
      "node": {
        "id": "clm4n8qwx000408l0k1l2m3n4",
        "category": "RED",
        "date": "2026-05-29T00:00:00.000Z",
        "html": "<p>Blocked on the vendor API outage.</p>",
        "user": {
          "id": "clm4n8qwx000108l0haz1f2pe",
          "fullName": "Dana Reyes"
        }
      },
      "previousValues": null,
      "updatedFields": null
    }
  }
}
```

## Errors

| Code                      | When                                                            |
| ------------------------- | --------------------------------------------------------------- |
| `STATUS_UPDATE_NOT_FOUND` | `statusUpdate(id)` was called with an id that doesn't exist.    |
| `UNAUTHENTICATED`         | The request carries no valid credentials.                       |
| `FORBIDDEN`               | The token is valid but lacks access to the workspace or update. |

## Permissions

All three operations require an authenticated caller with access to the workspace the update belongs to. The `subscribeToStatusUpdate` subscription additionally filters events to project members only: the server checks that the subscriber is in the project's user list before delivering each event, so a non-member's subscription stays silent. There is no separate read role for status updates beyond workspace membership.

## Related

- [Status updates overview](/api/status-updates)
- [Post and delete status updates](/api/status-updates/manage-status-updates)
- [Comments and discussions](/api/comments)
- [Workspaces](/api/workspaces)
- [List records](/api/records/list-records)
