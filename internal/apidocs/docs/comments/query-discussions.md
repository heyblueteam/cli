---
title: Query discussions
description: Fetch discussions three ways - one by ID, an offset-paginated org-wide list, or a cursor-paginated list scoped to a single project you belong to.
icon: Search
order: 4
---

Discussions are standalone, project-scoped conversation threads (the `Discussion` type). There are three read paths, and which one you reach for depends on what you have and what you need:

- **`discussion(id)`** — fetch a single thread by ID.
- **`discussions(filter, sort, skip, take)`** — the offset-paginated list across the current organization, filtered to one project and ordered by `DiscussionSort` (recently updated or recently commented).
- **`discussionList(projectId, ...)`** — the cursor-paginated variant scoped to one project, which returns only threads in projects you are a member of.

All three return the thread's metadata (title, body, members, `commentCount`, read state) but **not** its replies. The messages inside a discussion are `Comment` rows with `category: DISCUSSION`; load them separately with [`commentList`](/api/comments/query-comments). Use `discussion(id)` to render a single thread, `discussions` when you want simple page-number paging, and `discussionList` when you want to walk a project's threads with an opaque cursor.

## `discussion(id)`

Use `discussion` to fetch one thread by its ID.

### Request

```graphql
query GetDiscussion {
  discussion(id: "discussion_123") {
    id
    title
    text
    commentCount
    createdAt
    user {
      id
      fullName
    }
    project {
      id
      name
    }
  }
}
```

### Parameters

| Parameter | Type      | Required | Description                       |
| --------- | --------- | -------- | --------------------------------- |
| `id`      | `String!` | Yes      | The ID of the discussion to load. |

### Response

```json
{
  "data": {
    "discussion": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Q3 launch retro",
      "text": "What worked, what didn't, and what we change next quarter.",
      "commentCount": 14,
      "createdAt": "2026-05-20T09:12:44.000Z",
      "user": {
        "id": "clm4n8qwx000108l0a1b2c3d4",
        "fullName": "Dana Okafor"
      },
      "project": {
        "id": "clm4n8qwx000208l0e5f6g7h8",
        "name": "Marketing"
      }
    }
  }
}
```

Returns a single [`Discussion`](#the-discussion-type). The query throws `DISCUSSION_NOT_FOUND` when no thread matches the ID.

## `discussions`

Use `discussions` for the offset-paginated list. It is filtered to a single project (via the required `DiscussionFilterInput`) but always scoped to the calling organization, so a `projectId` from another organization returns no results. Results default to most-recently-updated first.

### Request

```graphql
query ListDiscussions {
  discussions(
    filter: { projectId: "project_123" }
    sort: [lastCommentedAt_DESC]
    skip: 0
    take: 20
  ) {
    items {
      id
      title
      commentCount
      updatedAt
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

### Parameters

| Parameter | Type                     | Required | Default            | Description                                                                                   |
| --------- | ------------------------ | -------- | ------------------ | --------------------------------------------------------------------------------------------- |
| `filter`  | `DiscussionFilterInput!` | Yes      | —                  | Which project's discussions to return. See [`DiscussionFilterInput`](#discussionfilterinput). |
| `sort`    | `[DiscussionSort!]`      | No       | `[updatedAt_DESC]` | Sort order. See [`DiscussionSort`](#discussionsort).                                          |
| `skip`    | `Int`                    | No       | `0`                | Number of discussions to skip from the start of the result set.                               |
| `take`    | `Int`                    | No       | `20`               | Maximum number of discussions to return on this page.                                         |

#### `DiscussionFilterInput`

| Field       | Type      | Required | Description                                      |
| ----------- | --------- | -------- | ------------------------------------------------ |
| `projectId` | `String!` | Yes      | The ID of the project whose discussions to list. |

#### `DiscussionSort`

`sort` takes a list of `DiscussionSort` values. Pass a single value for a single ordering.

| Value                  | Orders by                                         |
| ---------------------- | ------------------------------------------------- |
| `updatedAt_DESC`       | Most recently updated thread first (the default). |
| `updatedAt_ASC`        | Least recently updated thread first.              |
| `lastCommentedAt_DESC` | Thread with the most recent reply first.          |
| `lastCommentedAt_ASC`  | Thread with the oldest most-recent reply first.   |

### Response

`discussions` returns a `DiscussionPagination` — the threads on this page in `items`, plus offset pagination metadata in `pageInfo`.

```json
{
  "data": {
    "discussions": {
      "items": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "title": "Q3 launch retro",
          "commentCount": 14,
          "updatedAt": "2026-05-28T16:40:02.000Z"
        },
        {
          "id": "clm4n8qwx000308l0i9j0k1l2",
          "title": "Brand refresh kickoff",
          "commentCount": 3,
          "updatedAt": "2026-05-27T11:05:18.000Z"
        }
      ],
      "pageInfo": {
        "totalItems": 42,
        "totalPages": 3,
        "page": 1,
        "perPage": 20,
        "hasNextPage": true,
        "hasPreviousPage": false
      }
    }
  }
}
```

#### `DiscussionPagination`

| Field      | Type             | Description                                    |
| ---------- | ---------------- | ---------------------------------------------- |
| `items`    | `[Discussion!]!` | The discussions on this page.                  |
| `pageInfo` | `PageInfo!`      | Offset pagination metadata for the result set. |

To page through results, advance `skip` by `take` (for example `skip: 20, take: 20` for the second page) until `pageInfo.hasNextPage` is `false`.

## `discussionList`

Use `discussionList` for cursor-paginated access to one project's threads. Unlike `discussions`, it returns **only threads in a project the caller is a member of** — if you are not a member of `projectId`, the list comes back empty. Step through pages with the `after` cursor rather than a `skip` offset.

### Request

```graphql
query ListProjectDiscussions {
  discussionList(projectId: "project_123", first: 20, orderBy: updatedAt_DESC) {
    discussions {
      id
      title
      commentCount
      updatedAt
    }
    totalCount
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
```

### Parameters

| Parameter   | Type                     | Required | Default | Description                                                                                    |
| ----------- | ------------------------ | -------- | ------- | ---------------------------------------------------------------------------------------------- |
| `projectId` | `String!`                | Yes      | —       | The project whose discussions to list. Only returns threads if the caller is a project member. |
| `first`     | `Int`                    | No       | `20`    | Maximum number of discussions to return on this page.                                          |
| `after`     | `String`                 | No       | —       | Cursor to page forward from — pass the previous page's `pageInfo.endCursor`.                   |
| `before`    | `String`                 | No       | —       | Cursor to page backward from.                                                                  |
| `last`      | `Int`                    | No       | —       | Maximum number of discussions to return when paging backward.                                  |
| `skip`      | `Int`                    | No       | `0`     | Number of discussions to skip. Reported in `pageInfo`; prefer `after` for forward paging.      |
| `orderBy`   | `DiscussionOrderByInput` | No       | —       | Sort order. See [`DiscussionOrderByInput`](#discussionorderbyinput).                           |

#### `DiscussionOrderByInput`

A single enum value of the form `<field>_<direction>`. The fields and directions available:

| Field       | Ascending       | Descending       |
| ----------- | --------------- | ---------------- |
| `id`        | `id_ASC`        | `id_DESC`        |
| `uid`       | `uid_ASC`       | `uid_DESC`       |
| `title`     | `title_ASC`     | `title_DESC`     |
| `html`      | `html_ASC`      | `html_DESC`      |
| `text`      | `text_ASC`      | `text_DESC`      |
| `createdAt` | `createdAt_ASC` | `createdAt_DESC` |
| `updatedAt` | `updatedAt_ASC` | `updatedAt_DESC` |

<Callout variant="info" title="Same orderBy type as commentList">

`discussionList` and [`commentList`](/api/comments/query-comments) share the `DiscussionOrderByInput` type for their `orderBy` argument — the name is shared even though `commentList` orders comments. Sort by `updatedAt_DESC` or `createdAt_DESC` in both.

</Callout>

### Response

`discussionList` returns a `DiscussionList` — the threads in `discussions`, a `totalCount` of all matching threads, and a `pageInfo` carrying the `endCursor` for the next page.

```json
{
  "data": {
    "discussionList": {
      "discussions": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "title": "Q3 launch retro",
          "commentCount": 14,
          "updatedAt": "2026-05-28T16:40:02.000Z"
        }
      ],
      "totalCount": 42,
      "pageInfo": {
        "hasNextPage": true,
        "endCursor": "clm4n8qwx000008l0g4oxdqn7"
      }
    }
  }
}
```

#### `DiscussionList`

| Field         | Type             | Description                                                                    |
| ------------- | ---------------- | ------------------------------------------------------------------------------ |
| `discussions` | `[Discussion!]!` | The discussions on this page.                                                  |
| `totalCount`  | `Int!`           | Total number of discussions in the project the caller can see.                 |
| `pageInfo`    | `PageInfo!`      | Cursor pagination metadata. `endCursor` is the ID of the last thread returned. |

To fetch the next page, pass the current page's `pageInfo.endCursor` as the next call's `after`, and stop when `pageInfo.hasNextPage` is `false`.

## The `Discussion` type

All three queries return the same `Discussion` shape. Replies are not included here — load them with [`commentList`](/api/comments/query-comments).

| Field          | Type        | Description                                                              |
| -------------- | ----------- | ------------------------------------------------------------------------ |
| `id`           | `ID!`       | Unique identifier for the discussion.                                    |
| `title`        | `String!`   | The thread title.                                                        |
| `description`  | `String`    | Optional short description.                                              |
| `html`         | `String!`   | The opening body as sanitized HTML.                                      |
| `text`         | `String!`   | The opening body as plain text.                                          |
| `createdAt`    | `DateTime!` | When the discussion was created.                                         |
| `updatedAt`    | `DateTime!` | When the discussion was last updated.                                    |
| `user`         | `User!`     | The thread's creator. Select `id`, `fullName`, `email`.                  |
| `people`       | `[User!]`   | Members of the discussion.                                               |
| `project`      | `Project!`  | The project the discussion belongs to.                                   |
| `commentCount` | `Int!`      | Number of replies in the thread.                                         |
| `isRead`       | `Boolean`   | Whether the calling user has read this discussion. Evaluated per caller. |
| `isSeen`       | `Boolean`   | Whether the calling user has seen this discussion. Evaluated per caller. |

<Callout variant="warning" title="Discussion.comments is deprecated">

The `Discussion.comments` field is deprecated. Load a thread's replies with [`commentList`](/api/comments/query-comments) using `category: DISCUSSION` and `categoryId` set to the discussion's ID — it returns top-level comments with cursor and offset pagination.

</Callout>

## Loading a thread's replies

A discussion query gives you the thread; to render the conversation, follow up with `commentList` scoped to that discussion. `commentList` returns only top-level comments — read each comment's `replies` for the threaded answers.

```graphql
query DiscussionWithComments {
  discussion(id: "discussion_123") {
    id
    title
    commentCount
  }
  commentList(category: DISCUSSION, categoryId: "discussion_123", first: 20) {
    comments {
      id
      text
      user {
        id
        fullName
      }
      replyCount
    }
    totalCount
  }
}
```

## Errors

| Code                   | When                                                                                                  |
| ---------------------- | ----------------------------------------------------------------------------------------------------- |
| `DISCUSSION_NOT_FOUND` | `discussion(id)` — no discussion matches the given ID.                                                |
| `UNAUTHENTICATED`      | The request carries no valid authentication.                                                          |
| `COMPANY_NOT_FOUND`    | `discussion(id)` — the caller is not a member of the organization the discussion belongs to.          |
| `PROJECT_NOT_FOUND`    | `discussion(id)` — the caller is not a member of the discussion's project, or the project is missing. |

## Permissions

- **`discussion(id)`** requires the caller to be authenticated and a member of both the organization and the project the discussion lives in. Non-members get `COMPANY_NOT_FOUND` / `PROJECT_NOT_FOUND`.
- **`discussions`** returns threads only for the calling organization. A `projectId` belonging to another organization yields an empty `items` list rather than an error.
- **`discussionList`** returns threads only for projects the caller is a member of. Passing a `projectId` you do not belong to returns an empty `discussions` list, not an error.

All three calls authenticate with the standard token headers (`X-Bloo-Token-ID`, `X-Bloo-Token-Secret`, `X-Bloo-Company-ID`). The organization and project arguments accept an ID or a slug.

## Related

- [Query comments](/api/comments/query-comments)
- [Create, update & delete discussions](/api/comments/manage-discussions)
- [Comments overview](/api/comments)
- [Reactions](/api/comments/reactions)
- [Comments, Discussions & Chat subscriptions](/api/realtime/comment-discussion-subscriptions)
