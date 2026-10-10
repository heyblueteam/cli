---
title: Record groups
description: Group records by list, assignee, due date, or tag, with a count per group, using the recordGroups query.
icon: LayoutList
order: 13
---

`recordGroups` returns the groups for one grouping type, with the number of matching records in each group. It powers grouped views such as board columns and assignee buckets. The query returns groups, not records: each item carries a group `id`, `name`, and `todoCount`. To read the records inside a group, run a second [List records](/api/records/list-records) query with the matching filter.

## Request

```graphql
query RecordGroups {
  recordGroups(type: TODO_LIST, filter: { companyIds: ["company_123"] }) {
    items {
      id
      name
      type
      todoCount
    }
    pageInfo {
      totalItems
      hasNextPage
    }
  }
}
```

## Parameters

| Argument | Type               | Required | Description                                                                                                      |
| -------- | ------------------ | -------- | ---------------------------------------------------------------------------------------------------------------- |
| `type`   | `RecordGroupType!` | Yes      | What to group by — see [RecordGroupType](#recordgrouptype).                                                      |
| `filter` | `RecordsFilter`    | No       | Same filter the [todos query](/api/records/list-records) takes. Omit it to read the caller's whole organization. |
| `sort`   | `[RecordsSort!]`   | No       | Orders the groups. Only the first sort key applies — see [Ordering](#ordering).                                  |
| `skip`   | `Int`              | No       | Groups to skip for offset pagination. Defaults to `0`.                                                           |
| `take`   | `Int`              | No       | Groups per page. Defaults to `20`. There is no page-size ceiling.                                                |

`skip` and `take` paginate **groups**, not records. A group counts as one item however many records it holds.

### RecordGroupType

| Value       | Groups by                | Group `id`                                                   | Group `name`                                                                         |
| ----------- | ------------------------ | ------------------------------------------------------------ | ------------------------------------------------------------------------------------ |
| `TODO_LIST` | The list a record is in. | The list's ID.                                               | The list title.                                                                      |
| `DUE_DATE`  | The record's due date.   | The due instant. Empty string when a record has no due date. | The calendar date, for example `2026-07-15`. Empty string for the no-due-date group. |
| `ASSIGNEES` | The record's assignees.  | The user's ID.                                               | The user's first name.                                                               |
| `TAGS`      | The record's tags.       | The tag's ID.                                                | The tag title.                                                                       |

The `id` of a `TODO_LIST`, `ASSIGNEES`, or `TAGS` group is a real ID you can pass back to [List records](/api/records/list-records) as `todoListIds`, `assigneeIds`, or `tagIds`. A `DUE_DATE` group's `id` is the due instant itself.

Group facts that hold for every type:

- Only groups with at least one matching record are returned. An empty list, tag, or date never appears.
- `todoCount` is the number of distinct matching records in the group, after the filter and after visibility.
- A record with several assignees or several tags counts once in each of those groups.
- When several workspaces hold the same group — the same list, tag, user, or due instant — it appears once and the counts are summed.

## Response

```json
{
  "data": {
    "recordGroups": {
      "items": [
        { "id": "list_123", "name": "To do", "type": "TODO_LIST", "todoCount": 14 },
        { "id": "list_456", "name": "In progress", "type": "TODO_LIST", "todoCount": 3 }
      ],
      "pageInfo": {
        "totalItems": 2,
        "totalPages": 1,
        "page": 1,
        "perPage": 20,
        "hasNextPage": false,
        "hasPreviousPage": false
      }
    }
  }
}
```

### RecordGroup

| Field       | Type               | Description                                                                      |
| ----------- | ------------------ | -------------------------------------------------------------------------------- |
| `id`        | `String`           | The group's identifier (see [RecordGroupType](#recordgrouptype)). Can be `null`. |
| `name`      | `String`           | The group's display name. Can be `null`.                                         |
| `type`      | `RecordGroupType!` | The requested grouping type.                                                     |
| `todoCount` | `Int!`             | Distinct matching records in the group.                                          |

### PageInfo

Same shape as the [todos query](/api/records/list-records#pageinfo). `totalItems` counts **groups**, not records; `perPage` is the resolved `take`.

## Ordering

Without `sort`, groups keep their natural order: lists and tags by their position, assignees by name, due dates by date ascending with the no-due-date group last.

With `sort`, the groups are ordered by the sort value of a representative record in each group — the group's lowest matching record ID and its value for that key. Only the **first** key in the array applies; later keys are ignored. `rank_ASC` and `rank_DESC` are rejected with a `BAD_USER_INPUT` error, and any other key the query cannot translate falls back to the natural order.

## Full example

Open records across two workspaces, grouped by due date, oldest due first:

```graphql
query RecordGroupsByDueDate {
  recordGroups(
    type: DUE_DATE
    filter: {
      companyIds: ["company_123"]
      projectIds: ["project_123", "project_456"]
      showCompleted: false
    }
    sort: [duedAt_ASC]
    take: 50
  ) {
    items {
      id
      name
      todoCount
    }
    pageInfo {
      totalItems
      hasNextPage
    }
  }
}
```

Then read one group's records. The group's `id` goes back in as `todoListIds`, `assigneeIds`, or `tagIds` (or, for `DUE_DATE`, as `dueStart`/`dueEnd` around the group's instant):

```graphql
query RecordsInGroup {
  recordQueries {
    todos(
      filter: { companyIds: ["company_123"], assigneeIds: ["user_123"], showCompleted: false }
    ) {
      items {
        id
        title
        duedAt
      }
      pageInfo {
        totalItems
      }
    }
  }
}
```

## Scope and visibility

The query reads one organization: the one the request is authenticated against. `filter.companyIds` and `filter.projectIds` accept IDs or slugs and narrow that scope further.

- Organization owners and admins group every workspace. Everyone else groups only the workspaces they belong to.
- Workspaces without a migrated records table, and workspaces whose role grants the caller no table access, contribute no groups. This is silent — no error.
- Role limits still apply inside each workspace: assigned-only roles count only records naming the caller, hidden lists are excluded and never appear as `TODO_LIST` groups, and tag rules narrow every count.
- Deleted records are never counted. Archived records are counted only when `filter.archived` is `true` and the caller's role may view archived records; otherwise the query counts active records only.

See [List records — Permissions](/api/records/list-records#permissions) for the general visibility model.

## Errors

| Code                | When                                                                                                       |
| ------------------- | ---------------------------------------------------------------------------------------------------------- |
| `UNAUTHENTICATED`   | The request carries no valid credentials.                                                                  |
| `COMPANY_NOT_FOUND` | The credentials do not resolve to an organization membership.                                              |
| `PAYMENT_REQUIRED`  | The organization has neither a license nor an active subscription.                                         |
| `BAD_USER_INPUT`    | `sort` contains `rank_ASC` or `rank_DESC`, or an `updatedAt_gt`/`updatedAt_gte` value is not a valid date. |

Invisible workspaces and records never error — they are absent from `items`.

## Related

- [List records](/api/records/list-records) — read the records behind a group, and the full filter reference
- [Manage assignees](/api/records/assignees) — the records behind `ASSIGNEES` groups
- [Manage tags](/api/records/tags) — the records behind `TAGS` groups
