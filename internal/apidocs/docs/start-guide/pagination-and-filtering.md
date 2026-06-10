---
title: Pagination & Filtering
description: How list queries page and filter — the offset style (skip/take), legacy cursor args, the PageInfo shape, page sizes, and the shared filter vocabulary.
icon: ListOrdered
order: 7
---

Almost every list query in the Blue API is paginated, and most accept a filter. This page documents the conventions so you don't have to relearn them per query: the two pagination styles, the `PageInfo` shape returned by both, the default and maximum page sizes, and the shared filter input used to narrow record queries. Records are `Todo` objects and workspaces are `Project` objects in the API.

## Two pagination styles

Blue has two pagination idioms. The **offset style** is dominant — roughly 35 queries use it — and is the one to reach for. A handful of older `*List` queries expose **cursor-style arguments** (`first`/`last`/`after`/`before`), but their resolvers only honor a subset of those args. When in doubt, use the offset style.

### Offset style (recommended)

Offset queries take a `skip` (how many rows to skip) and a page-size argument — either `take` or `limit` depending on the query — and return a container object with `items` and a `pageInfo`:

```graphql
query ListForms {
  forms(skip: 0, take: 20) {
    items {
      id
      title
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

```json
{
  "data": {
    "forms": {
      "items": [
        { "id": "clm4n8qwx000008l0g4oxdqn7", "title": "Bug intake" },
        { "id": "clm4n8qwx000108l0h7pterz2", "title": "Customer feedback" }
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

The page-size argument is named **`take`** on most queries (`forms`, `customFields`, `projectList`, `templates`, …) and **`limit`** on the record-oriented queries (`todoQueries { todos }`, `publicViewRecords`). The skip argument is always `skip`. A couple of container types also carry a `totalCount` field, but it's **deprecated** in favor of `pageInfo.totalItems` — don't select it.

### Cursor style (partial — legacy)

Three `*List` queries — `companyUserList`, `commentList`, and `activityList` — expose the relay-flavored `after`, `before`, `first`, `last`, and `orderBy` arguments alongside `skip`. Their resolvers do **not** implement full bidirectional cursoring. What each actually honors:

| Query             | Honors                                          | Ignores                            | Notes                                                                                          |
| ----------------- | ----------------------------------------------- | ---------------------------------- | ---------------------------------------------------------------------------------------------- |
| `companyUserList` | `skip`, `search`, `orderBy`                     | `first`, `last`, `before`, `after` | Page size is hard-capped at **200** server-side regardless of `first`. Page with `skip`.       |
| `commentList`     | `skip`, `first` (page size), `after`, `orderBy` | `last`, `before`                   | `after` is a **keyset** cursor: pass a comment `id`, results start at rows with `id >= after`. |
| `activityList`    | `skip`, `first` (page size), `after`, `orderBy` | `last`, `before`                   | `after` is a keyset cursor on the activity `id`.                                               |

Because `before`/`last` are no-ops on these queries, treat them as offset queries: page forward with `skip` (or with the keyset `after` cursor on `commentList`/`activityList`), and read `pageInfo` to know when to stop. The deprecated `pageInfo.startCursor`/`pageInfo.endCursor` fields are not populated by these resolvers.

<Callout variant="info" title="Container field name varies">

Offset containers expose the page as **`items`** (`FormPagination`, `CustomFieldPagination`, `TodosResult`, `ProjectPagination`, `UserPagination`). The three cursor-style `*List` containers name it differently: `CompanyUserList.users`, `CommentList.comments`, and `ActivityList.activities`.

</Callout>

## PageInfo

Both styles return a `PageInfo` object. Select these fields:

| Field             | Type       | Description                                                            |
| ----------------- | ---------- | ---------------------------------------------------------------------- |
| `totalItems`      | `Int`      | Total rows matching the query across all pages.                        |
| `totalPages`      | `Int`      | Total number of pages, `ceil(totalItems / perPage)`.                   |
| `page`            | `Int`      | Current page number, 1-based. Computed as `floor(skip / perPage) + 1`. |
| `perPage`         | `Int`      | Page size in effect (the resolved `take`/`limit`).                     |
| `hasNextPage`     | `Boolean!` | `true` when `page < totalPages`.                                       |
| `hasPreviousPage` | `Boolean!` | `true` when `page > 1`.                                                |
| `startCursor`     | `String`   | **Deprecated.** Page with `skip` instead.                              |
| `endCursor`       | `String`   | **Deprecated.** Page with `skip` instead.                              |

<Callout variant="warning" title="Don't use total or PaginationInfo">

The total-rows field is `totalItems`, not `total`. An older `PaginationInfo` type (`totalPages!`, `totalItems!`, `page!`, `perPage!`, `hasNextPage!`, `hasPreviousPage!`) is **deprecated** and not used by current queries — use `PageInfo`. The deprecated `startCursor`/`endCursor` are not populated; never page off them.

</Callout>

## Default and maximum page sizes

`skip` defaults to `0` everywhere. The page-size default varies by query:

| Query                                  | Page-size arg | Default  | Maximum / cap                                                        |
| -------------------------------------- | ------------- | -------- | -------------------------------------------------------------------- |
| `todoQueries { todos }`                | `limit`       | `20`     | **500** (hard cap; `limit <= 0` or `> 500` clamps to 500)            |
| `forms`                                | `take`        | `20`     | —                                                                    |
| `customFields`                         | `take`        | `500`    | —                                                                    |
| `publicViewRecords`                    | `limit`       | required | `limit` must be **25 or 150**; `skip` a multiple of `limit`, ≤ 10000 |
| Most other `*List` / paginated queries | `take`        | `20`     | —                                                                    |

<Callout variant="warning" title="The records query caps at 500">

`todoQueries { todos }` silently clamps `limit` to **500**: any value `<= 0` or `> 500` is treated as 500. To read more than 500 records, page with `skip` — the clamp is per-page, not a total ceiling.

</Callout>

`publicViewRecords` is stricter than the others: it accepts a `limit` of only `25` or `150`, requires `skip` to be a non-negative multiple of `limit`, and rejects `skip` above 10000. Out-of-range values return `BAD_USER_INPUT`.

## Driving a pagination loop

Page off `skip` plus `pageInfo.hasNextPage`. Start at `skip: 0`, then add the page size each round until `hasNextPage` is `false`:

```graphql
query RecordsPage($skip: Int!) {
  todoQueries {
    todos(filter: { companyIds: ["company_123"] }, limit: 100, skip: $skip) {
      items {
        id
        title
      }
      pageInfo {
        totalItems
        hasNextPage
      }
    }
  }
}
```

Send the query with `{ "skip": 0 }`, read `hasNextPage`, and re-send with `{ "skip": 100 }`, `{ "skip": 200 }`, and so on while `hasNextPage` is `true`. `totalItems` lets you size the loop up front (`ceil(totalItems / perPage)` requests).

### Counting without fetching

If you only need the size of a result set, use the count-only companion query rather than paging through `items`. The records query has `todosCount`, which takes the same filter input and returns a bare `Int!`:

```graphql
query CountRecords {
  todosCount(filter: { projectIds: ["project_123"], showCompleted: false })
}
```

```json
{ "data": { "todosCount": 128 } }
```

`todosCount` also accepts an optional `refetchedAt: DateTime` argument to bypass cached counts.

## Filtering records

The record queries share a filter vocabulary. `todoQueries { todos }` and `publicViewRecords` take a `TodosFilter!` (its `companyIds` is required); `todosCount` and `activityList` take a `TodoFilterInput`. The two inputs overlap heavily — the table below covers the most useful `TodoFilterInput` fields. (`TodosFilter` adds the required `companyIds` plus `todoIds`, `todoListTitles`, search ranges, and more.)

### TodoFilterInput

| Field                         | Type                       | Description                                                                                  |
| ----------------------------- | -------------------------- | -------------------------------------------------------------------------------------------- |
| `projectIds`                  | `[String!]`                | Restrict to these workspaces (accepts IDs or slugs).                                         |
| `assigneeIds`                 | `[String!]`                | Records assigned to any of these users.                                                      |
| `unassigned`                  | `Boolean`                  | When `true`, match records with no assignee.                                                 |
| `tagIds`                      | `[String!]`                | Records carrying any of these tags.                                                          |
| `todoListIds`                 | `[String!]`                | Records in any of these lists.                                                               |
| `showCompleted`               | `Boolean`                  | Include completed records (default excludes them).                                           |
| `q`                           | `String`                   | Free-text query against record title and content.                                            |
| `search`                      | `String`                   | Alias of `q` (same free-text match).                                                         |
| `fields`                      | `JSON`                     | Per-field conditions (custom fields, time-in-list, due-date). Each entry needs a `type` key. |
| `op`                          | `FilterLogicalOperator`    | Logical operator joining the `fields` conditions. Defaults to `AND`.                         |
| `groups`                      | `[TodoFilterGroupInput!]`  | Nested filter groups. When set, the flat fields above are ignored in favor of the groups.    |
| `groupLinks`                  | `[FilterLogicalOperator!]` | Operators linking adjacent groups (length should be `groups.length - 1`).                    |
| `createdStart` / `createdEnd` | `DateTime`                 | Created-date range.                                                                          |
| `dueStart` / `dueEnd`         | `DateTime`                 | Due-date range.                                                                              |

The input also exposes negation arrays (`notAssigneeIds`, `notTagIds`, `notColors`, `notTodoListIds`), existence booleans (`hasTag`, `hasDueDate`, `hasChecklist`, `hasDependency`, `hasReference`, …), completed-date range (`completedStart`/`completedEnd`), and last-updated-by axes (`lastUpdatedByUserIds`, `lastUpdatedByAutomationIds`, `lastUpdatedByActorTypes`). Grep the schema for the full set.

<Callout variant="info" title="The fields JSON needs a type">

Each entry in the `fields` array **must** include a `type` key — `CUSTOM_FIELD`, `TIME_IN_LIST`, `RELATIVE_DUEDATE`, or `FIELD` — or the query layer silently drops it. A custom-field condition looks like `{ type: "CUSTOM_FIELD", customFieldId: "field_123", customFieldType: "CHECKBOX", op: "IS", values: true }`.

</Callout>

### Filter operators

Two enums drive how conditions combine and compare.

`FilterLogicalOperator` joins conditions or groups:

| Value | Meaning                                |
| ----- | -------------------------------------- |
| `AND` | A record must satisfy every condition. |
| `OR`  | A record must satisfy at least one.    |

`FilterComparisonOperator` is the per-condition comparison (used within `fields` entries):

| Value          | Meaning                     |
| -------------- | --------------------------- |
| `IS`           | Equals (presence-aware).    |
| `NOT`          | Not equal (presence-aware). |
| `EQ`           | Equal.                      |
| `NE`           | Not equal.                  |
| `IN`           | Value in a set.             |
| `NIN`          | Value not in a set.         |
| `GT`           | Greater than.               |
| `GTE`          | Greater than or equal.      |
| `LT`           | Less than.                  |
| `LTE`          | Less than or equal.         |
| `CONTAINS`     | Substring / membership.     |
| `NOT_CONTAINS` | Negated `CONTAINS`.         |

### Filtering example

Active records in two workspaces, assigned to a given user, whose title contains "launch":

```graphql
query FilteredRecords {
  todoQueries {
    todos(
      filter: {
        companyIds: ["company_123"]
        projectIds: ["project_123", "project_456"]
        assigneeIds: ["user_123"]
        showCompleted: false
        q: "launch"
      }
      sort: [duedAt_ASC]
      limit: 50
    ) {
      items {
        id
        title
        done
        duedAt
      }
      pageInfo {
        totalItems
        hasNextPage
      }
    }
  }
}
```

<Callout variant="info" title="Filtering is the only way to search server-side">

The in-app search box filters records on the client after they load — it does not hit the API. To find records across a workspace larger than one page, you must filter server-side with the `todos` query's `TodosFilter` (e.g. `q` for free-text, or `fields` for custom-field conditions).

</Callout>

## Errors

| Code                    | When                                                                                                                   |
| ----------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `BAD_USER_INPUT`        | `publicViewRecords` got an invalid `limit` (not 25/150), a `skip` that isn't a multiple of `limit`, or `skip > 10000`. |
| `FORBIDDEN`             | The filter references a workspace or organization the caller can't access.                                             |
| `COMPANY_NOT_FOUND`     | `companyUserList` got a `companyId` the caller isn't a member of.                                                      |
| `PUBLIC_VIEW_NOT_FOUND` | `publicViewRecords` got a `publicViewId` that doesn't exist.                                                           |
| `UNAUTHENTICATED`       | A password-protected public view was queried without a valid view token.                                               |

## Related

- [List records](/api/records/list-records)
- [List workspaces](/api/workspaces/list-workspaces)
- [List custom fields](/api/custom-fields/list-custom-fields)
- [List users](/api/user-management/list-users)
- [Making requests](/api/start-guide/making-requests)
- [Error codes](/api/start-guide/error-codes)
