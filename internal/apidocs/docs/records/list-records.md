---
title: List Records
description: Query, filter, sort, and paginate records across your organization with the todos query.
icon: List
order: 1
---

Use the `todoQueries.todos` query to read records from Blue with filtering, sorting, and pagination. Records are `Todo` objects in the API; in the app they appear as the cards inside a workspace's lists. A single call can span one organization, several workspaces, or a precise slice defined by assignees, tags, dates, and custom-field conditions.

This is the only operation that filters records server-side. (In-app search filters client-side after records load, so it can't reach records beyond what the client has fetched.)

## Request

The query lives under the `todoQueries` namespace. The only required input is `companyIds` — one or more organization IDs or slugs.

```graphql
query ListRecords {
  todoQueries {
    todos(filter: { companyIds: ["company_123"] }) {
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

The `companyIds`, `projectIds`, and `todoListIds` arrays each accept **IDs or slugs**.

## Parameters

### Arguments

| Argument | Type           | Required | Description                                                                                                      |
| -------- | -------------- | -------- | ---------------------------------------------------------------------------------------------------------------- |
| `filter` | `TodosFilter!` | Yes      | Filter criteria (below). `companyIds` is the only required field.                                                |
| `sort`   | `[TodosSort!]` | No       | Ordering. Defaults to `[]` (the engine's natural order). Multiple keys apply in sequence.                        |
| `limit`  | `Int`          | No       | Items per page. Omit to get **20**. Any value over 500 — and any value `<= 0` — is coerced to **500** (the max). |
| `skip`   | `Int`          | No       | Items to skip for offset pagination. Defaults to `0`.                                                            |

<Callout variant="warning" title="limit: 0 is not the default">

Omitting `limit` yields 20, but passing `0` (or a negative number) yields 500, not 20. Pass the explicit page size you want.

</Callout>

### TodosFilter

The flat fields below combine with `AND` by default. For mixed AND/OR logic and the full set of conditions (record-name match, completion-date range, last-updated-by actor), use `groups` instead — see [Group filtering](#group-filtering).

| Field                     | Type                       | Description                                                                                                                                                                    |
| ------------------------- | -------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `companyIds`              | `[String!]!`               | **Required.** Organization IDs or slugs to query.                                                                                                                              |
| `projectIds`              | `[String!]`                | Restrict to these workspace IDs or slugs.                                                                                                                                      |
| `todoIds`                 | `[String!]`                | Return only these specific records.                                                                                                                                            |
| `assigneeIds`             | `[String!]`                | Records assigned to any of these users.                                                                                                                                        |
| `unassigned`              | `Boolean`                  | When `true`, only records with no assignees.                                                                                                                                   |
| `tagIds`                  | `[String!]`                | Records carrying any of these tags.                                                                                                                                            |
| `tagColors`               | `[String!]`                | Filter by tag color.                                                                                                                                                           |
| `tagTitles`               | `[String!]`                | Filter by tag title.                                                                                                                                                           |
| `todoListIds`             | `[String!]`                | Records in these lists.                                                                                                                                                        |
| `todoListTitles`          | `[String!]`                | Records in lists with these titles.                                                                                                                                            |
| `done`                    | `Boolean`                  | Filter by completion state. Prefer `showCompleted` for the common "hide completed" case.                                                                                       |
| `showCompleted`           | `Boolean`                  | When `false`, completed records are excluded.                                                                                                                                  |
| `startedAt`               | `DateTime`                 | Records with this exact start date.                                                                                                                                            |
| `duedAt`                  | `DateTime`                 | Records with this exact due date.                                                                                                                                              |
| `dueStart`                | `DateTime`                 | Due date on or after this (inclusive).                                                                                                                                         |
| `dueEnd`                  | `DateTime`                 | Due date on or before this (inclusive).                                                                                                                                        |
| `duedAtStart`             | `DateTime`                 | Alias of `dueStart`.                                                                                                                                                           |
| `duedAtEnd`               | `DateTime`                 | Alias of `dueEnd`.                                                                                                                                                             |
| `createdStart`            | `DateTime`                 | Created on or after this.                                                                                                                                                      |
| `createdEnd`              | `DateTime`                 | Created on or before this.                                                                                                                                                     |
| `updatedAt_gt`            | `DateTime`                 | Updated strictly after this.                                                                                                                                                   |
| `updatedAt_gte`           | `DateTime`                 | Updated on or after this.                                                                                                                                                      |
| `search`                  | `String`                   | Match title and text content (case-insensitive).                                                                                                                               |
| `q`                       | `String`                   | Alias of `search`.                                                                                                                                                             |
| `excludeArchivedProjects` | `Boolean`                  | Exclude records in archived workspaces.                                                                                                                                        |
| `archived`                | `Boolean`                  | Archive scope. Omitted/`false` → active records (archived excluded); `true` → only archived records (needs the view-archived permission). See [Archive scope](#archive-scope). |
| `recordName`              | `String`                   | Full-text record-title search (indexed; `CONTAINS` by default, override with `recordNameOp`).                                                                                  |
| `coordinates`             | `JSON`                     | Polygon coordinates for map-view geo filtering.                                                                                                                                |
| `notAssigneeIds`          | `[String!]`                | Exclude records assigned to any of these users.                                                                                                                                |
| `notTagIds`               | `[String!]`                | Exclude records carrying any of these tags.                                                                                                                                    |
| `notColors`               | `[String!]`                | Exclude records with these tag colors.                                                                                                                                         |
| `notTodoListIds`          | `[String!]`                | Exclude records in these lists.                                                                                                                                                |
| `hasTag`                  | `Boolean`                  | Records that have (or, when `false`, lack) any tag.                                                                                                                            |
| `hasColor`                | `Boolean`                  | Records that have (or lack) a tag color.                                                                                                                                       |
| `hasDueDate`              | `Boolean`                  | Records that have (or lack) a due date.                                                                                                                                        |
| `hasDescription`          | `Boolean`                  | Records that have (or lack) description text.                                                                                                                                  |
| `hasChecklist`            | `Boolean`                  | Records that have (or lack) a checklist.                                                                                                                                       |
| `hasDependency`           | `Boolean`                  | Records that have (or lack) a dependency.                                                                                                                                      |
| `hasReference`            | `Boolean`                  | Records that have (or lack) a reference.                                                                                                                                       |
| `fields`                  | `JSON`                     | Custom-field conditions. See [Custom-field filtering](#custom-field-filtering).                                                                                                |
| `op`                      | `FilterLogicalOperator`    | Logical operator applied across `fields` entries. Defaults to `AND`.                                                                                                           |
| `groups`                  | `[TodoFilterGroupInput!]`  | Nested filter groups. **When provided, every flat field above is ignored** in favor of the groups. See [Group filtering](#group-filtering).                                    |
| `groupLinks`              | `[FilterLogicalOperator!]` | Operators linking adjacent groups; length should be `groups.length - 1`. A shorter list repeats its last operator.                                                             |

### TodosSort

Each value is a `field_DIRECTION` enum member. Pass an array to sort by multiple keys in order.

| Value                                                                  | Sorts by                                           |
| ---------------------------------------------------------------------- | -------------------------------------------------- |
| `position_ASC` / `position_DESC`                                       | Position within the list (the default list order). |
| `title_ASC` / `title_DESC`                                             | Record title, alphabetically.                      |
| `createdAt_ASC` / `createdAt_DESC`                                     | Creation date.                                     |
| `updatedAt_ASC` / `updatedAt_DESC`                                     | Last-updated date.                                 |
| `completedAt_ASC` / `completedAt_DESC`                                 | Completion date.                                   |
| `startedAt_ASC` / `startedAt_DESC`                                     | Start date.                                        |
| `duedAt_ASC` / `duedAt_DESC`                                           | Due date.                                          |
| `createdBy_ASC` / `createdBy_DESC`                                     | Creator.                                           |
| `assignees_ASC` / `assignees_DESC`                                     | Assignees.                                         |
| `todoTags_ASC` / `todoTags_DESC`                                       | Tags.                                              |
| `todoListTitle_ASC` / `todoListTitle_DESC`                             | List title.                                        |
| `todoListPosition_ASC` / `todoListPosition_DESC`                       | List position.                                     |
| `projectName_ASC` / `projectName_DESC`                                 | Workspace name.                                    |
| `checklistTitle_ASC` / `checklistTitle_DESC`                           | Checklist title.                                   |
| `checklistItemTitle_ASC` / `checklistItemTitle_DESC`                   | Checklist-item title.                              |
| `todoCustomFieldDate_ASC` / `todoCustomFieldDate_DESC`                 | A date custom field.                               |
| `todoCustomFieldSelectSingle_ASC` / `todoCustomFieldSelectSingle_DESC` | A single-select custom field.                      |
| `todoCustomFieldSelectMulti_ASC` / `todoCustomFieldSelectMulti_DESC`   | A multi-select custom field.                       |
| `todoCustomFieldAssignee_ASC` / `todoCustomFieldAssignee_DESC`         | An assignee custom field.                          |
| `todoCustomFieldRollup_ASC` / `todoCustomFieldRollup_DESC`             | A rollup custom field.                             |

### FilterLogicalOperator

| Value | Meaning                    |
| ----- | -------------------------- |
| `AND` | All conditions must match. |
| `OR`  | Any condition may match.   |

## Response

```json
{
  "data": {
    "todoQueries": {
      "todos": {
        "items": [
          {
            "id": "clm4n8qwx000008l0g4oxdqn7",
            "title": "Draft Q3 launch plan",
            "done": false,
            "duedAt": "2026-07-15T17:00:00.000Z"
          },
          {
            "id": "clm4n8qwx000108l0haz1f2pe",
            "title": "Review pricing page copy",
            "done": true,
            "duedAt": null
          }
        ],
        "pageInfo": {
          "totalItems": 2,
          "hasNextPage": false
        }
      }
    }
  }
}
```

### TodosResult

| Field      | Type        | Description               |
| ---------- | ----------- | ------------------------- |
| `items`    | `[Todo!]!`  | The records on this page. |
| `pageInfo` | `PageInfo!` | Pagination metadata.      |

### PageInfo

| Field             | Type       | Description                                                                   |
| ----------------- | ---------- | ----------------------------------------------------------------------------- |
| `totalItems`      | `Int`      | Total records matching the filter, across all pages.                          |
| `totalPages`      | `Int`      | Total pages at the current `limit`.                                           |
| `page`            | `Int`      | Current page, derived from `skip` and `limit`.                                |
| `perPage`         | `Int`      | Page size in effect (the resolved `limit`).                                   |
| `hasNextPage`     | `Boolean!` | Whether a next page exists.                                                   |
| `hasPreviousPage` | `Boolean!` | Whether a previous page exists.                                               |
| `startCursor`     | `String`   | Deprecated. Cursor pagination is not used here; paginate with `skip`/`limit`. |
| `endCursor`       | `String`   | Deprecated. See `startCursor`.                                                |

### Todo

The most useful fields on each returned record. The `Todo` type exposes more — these are the ones you'll select most often.

| Field                     | Type              | Description                                                                                                                                                               |
| ------------------------- | ----------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `id`                      | `ID!`             | Unique identifier.                                                                                                                                                        |
| `uid`                     | `String!`         | Short, human-facing identifier.                                                                                                                                           |
| `position`                | `Float!`          | Position within its list.                                                                                                                                                 |
| `title`                   | `String!`         | Record title.                                                                                                                                                             |
| `text`                    | `String!`         | Description as plain text.                                                                                                                                                |
| `html`                    | `String!`         | Description as HTML.                                                                                                                                                      |
| `startedAt`               | `DateTime`        | Start date.                                                                                                                                                               |
| `duedAt`                  | `DateTime`        | Due date.                                                                                                                                                                 |
| `timezone`                | `String`          | Timezone for the record's dates.                                                                                                                                          |
| `color`                   | `String`          | Color label.                                                                                                                                                              |
| `cover`                   | `String`          | Cover image URL.                                                                                                                                                          |
| `coverLocked`             | `Boolean`         | Whether the cover is locked from auto-updates.                                                                                                                            |
| `done`                    | `Boolean!`        | Completion state. There is no status enum — completion is this boolean.                                                                                                   |
| `completedAt`             | `DateTime`        | When the record was last marked complete.                                                                                                                                 |
| `archived`                | `Boolean!`        | Whether the record is archived.                                                                                                                                           |
| `createdAt`               | `DateTime!`       | Creation timestamp.                                                                                                                                                       |
| `updatedAt`               | `DateTime!`       | Last-update timestamp.                                                                                                                                                    |
| `createdBy`               | `User`            | The user who created the record. Select `fullName`/`email`, not `name`.                                                                                                   |
| `lastUpdatedActor`        | `TodoActor`       | Polymorphic actor (user, automation, API key, system, or form) that last updated the record. Nullable for older rows.                                                     |
| `commentCount`            | `Int!`            | Number of comments.                                                                                                                                                       |
| `checklistCount`          | `Int!`            | Total checklist items.                                                                                                                                                    |
| `checklistCompletedCount` | `Int!`            | Completed checklist items.                                                                                                                                                |
| `isRepeating`             | `Boolean!`        | Whether the record recurs.                                                                                                                                                |
| `repeating`               | `JSON`            | Recurrence configuration.                                                                                                                                                 |
| `reminder`                | `JSON`            | Reminder configuration.                                                                                                                                                   |
| `isRead`                  | `Boolean`         | Read state for the current user.                                                                                                                                          |
| `isSeen`                  | `Boolean`         | Seen state for the current user.                                                                                                                                          |
| `todoList`                | `TodoList!`       | The list this record belongs to.                                                                                                                                          |
| `todoLists`               | `[TodoList!]`     | All lists this record appears in (for cross-list records).                                                                                                                |
| `repeatingTodoList`       | `TodoList`        | The list a recurring record was generated into.                                                                                                                           |
| `users`                   | `[User!]!`        | Assigned users. Select `fullName`/`email`, not `name`.                                                                                                                    |
| `tags`                    | `[Tag!]!`         | Tags. Select `title`/`color` (a `Tag` has no `name`).                                                                                                                     |
| `checklists`              | `[Checklist!]!`   | Checklists on the record.                                                                                                                                                 |
| `customFields`            | `[CustomField!]!` | Custom-field values. Select value fields on each element (`name`, `type`, `value`, `text`, `number`, `checked`, `selectedOption { title }`, `selectedOptions { title }`). |
| `dependOn`                | `[Todo!]`         | Records this one depends on (blocking).                                                                                                                                   |
| `dependBy`                | `[Todo!]`         | Records that depend on this one.                                                                                                                                          |
| `timeTracking`            | `TimeTracking`    | Time-tracking summary.                                                                                                                                                    |

## Full example

Filter to open, high-priority records in two workspaces, due in 2026, sorted by due date then list position, with a rich selection set.

```graphql
query ListRecordsAdvanced {
  todoQueries {
    todos(
      filter: {
        companyIds: ["company_123"]
        projectIds: ["project_123", "project_456"]
        assigneeIds: ["user_123"]
        tagIds: ["tag_123"]
        showCompleted: false
        dueStart: "2026-01-01T00:00:00Z"
        dueEnd: "2026-12-31T23:59:59Z"
        search: "launch"
        excludeArchivedProjects: true
        fields: [
          {
            type: "CUSTOM_FIELD"
            customFieldId: "field_123"
            customFieldType: "SELECT_SINGLE"
            op: "IN"
            values: ["option_123", "option_456"]
          }
        ]
        op: AND
      }
      sort: [duedAt_ASC, todoListPosition_ASC]
      limit: 50
      skip: 0
    ) {
      items {
        id
        uid
        title
        text
        duedAt
        done
        completedAt
        todoList {
          id
          title
        }
        users {
          id
          fullName
          email
        }
        tags {
          id
          title
          color
        }
        customFields {
          id
          name
          type
          value
          selectedOption {
            title
          }
        }
        createdBy {
          id
          fullName
        }
      }
      pageInfo {
        totalPages
        totalItems
        page
        perPage
        hasNextPage
        hasPreviousPage
      }
    }
  }
}
```

## Archive scope

Archiving takes a record out of its list without deleting it (see [Archive a record](/api/records/archive-record)). The `archived` filter selects which side of that partition you read:

| `archived` value | Returns                                                           |
| ---------------- | ----------------------------------------------------------------- |
| omitted          | Active records only — archived records are excluded. **Default.** |
| `false`          | Active records only — identical to omitting it.                   |
| `true`           | Archived records only. Active records are excluded.               |

There is **no value that returns active and archived records together** in a single call — the filter is a partition, not an inclusion toggle, so each query reads one side or the other. To assemble a complete picture, run the query twice (once with `archived: false` or omitted, once with `archived: true`) and merge the two result sets client-side.

```graphql
query ArchivedRecords {
  todoQueries {
    todos(filter: { companyIds: ["company_123"], projectIds: ["project_123"], archived: true }) {
      items {
        id
        title
        archived
        archivedAt
      }
      pageInfo {
        totalItems
      }
    }
  }
}
```

<Callout variant="warning" title="Archived results need the view-archived permission">

`archived: true` is honored only for roles permitted to view archived records. A role without that permission is silently downgraded to the active scope — it gets active records back rather than an error — so `archived: true` can never leak archived rows to a caller who shouldn't see them.

</Callout>

## List all records in a workspace

To read every record in one workspace, filter by that workspace and page through the results. `companyIds` is always required; add the workspace to `projectIds` and omit `todoListIds` so the query spans every list:

```graphql
query AllWorkspaceRecords {
  todoQueries {
    todos(
      filter: { companyIds: ["company_123"], projectIds: ["project_123"] }
      limit: 500
      skip: 0
    ) {
      items {
        id
        title
        todoList {
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
}
```

A few things to keep in mind:

- **Paginate to completion.** A page is capped at 500 (the `limit` ceiling). Increment `skip` by your page size until `pageInfo.hasNextPage` is `false`; `pageInfo.totalItems` gives you the full count up front so you can size the loop.
- **This returns active records only.** Archived records are excluded by default — see [Archive scope](#archive-scope). To include them, run the same query a second time with `archived: true` and combine the pages.
- **One query covers the whole workspace.** There's no per-list call to make — a single query with no `todoListIds` returns records from every list in the workspace.
- **Span workspaces or the whole org.** `projectIds` accepts several IDs to cover multiple workspaces at once; drop `projectIds` entirely to read across every workspace in the organizations named in `companyIds`.
- **Visibility still applies.** Records the caller can't see — restricted lists, assigned-only roles, tag-based rules — are filtered out silently rather than erroring.

## Custom-field filtering

`TodosFilter.fields` is typed `JSON`. Each entry must include a `type` discriminator — entries without one are silently dropped. The query layer routes on `type`:

| `type`             | Filters by                                                    |
| ------------------ | ------------------------------------------------------------- |
| `CUSTOM_FIELD`     | A custom field's value.                                       |
| `FIELD`            | A built-in record field.                                      |
| `TIME_IN_LIST`     | How long a record has sat in its current list.                |
| `RELATIVE_DUEDATE` | A due date relative to now (e.g. "overdue", "due this week"). |

A `CUSTOM_FIELD` entry has this shape:

```json
{
  "type": "CUSTOM_FIELD",
  "customFieldId": "field_123",
  "customFieldType": "SELECT_SINGLE",
  "op": "IN",
  "values": ["option_123", "option_456"]
}
```

| Key               | Type            | Description                                                                                                                                                                |
| ----------------- | --------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `type`            | `String`        | Must be `"CUSTOM_FIELD"`.                                                                                                                                                  |
| `customFieldId`   | `String`        | The custom field's ID.                                                                                                                                                     |
| `customFieldType` | `String`        | One of the `CustomFieldType` enum values (below). Determines how the value is matched.                                                                                     |
| `op`              | `String`        | A `FilterComparisonOperator`: `IS`, `NOT`, `EQ`, `NE`, `IN`, `NIN`, `GT`, `GTE`, `LT`, `LTE`, `CONTAINS`, `NOT_CONTAINS`. Which operators apply depends on the field type. |
| `values`          | depends on type | Option IDs for selects, a number for numeric types, a boolean for `CHECKBOX`, a string for text types.                                                                     |

### Filterable custom-field types

`customFieldType` must be a `CustomFieldType` enum value. These types can be filtered:

`TEXT_SINGLE`, `TEXT_MULTI`, `URL`, `EMAIL`, `PHONE`, `UNIQUE_ID` (text matching), `NUMBER`, `CURRENCY`, `PERCENT`, `RATING`, `FORMULA` (numeric comparison), `CHECKBOX` (`IS true` / `IS false`), `SELECT_SINGLE`, `SELECT_MULTI` (option matching), `COUNTRY`, `DATE`, `ASSIGNEE`, `REFERENCE`, `LOOKUP`, and `ROLLUP`.

The remaining `CustomFieldType` values are not filterable: `LOCATION`, `FILE`, `TIME_DURATION`, `BUTTON`, `CURRENCY_CONVERSION`, and `REFERENCED_BY`.

## Group filtering

For mixed AND/OR logic, supply `groups`. **When `groups` is present, the flat `TodosFilter` fields are ignored.** Each group is a `TodoFilterGroupInput` with its own internal operator (`op`), and adjacent groups are linked by `groupLinks`.

```graphql
query ListRecordsByGroups {
  todoQueries {
    todos(
      filter: {
        companyIds: ["company_123"]
        groups: [
          { op: AND, tagIds: ["tag_123"], showCompleted: false }
          { op: OR, recordName: "launch", recordNameOp: CONTAINS }
        ]
        groupLinks: [AND]
      }
      limit: 50
    ) {
      items {
        id
        title
      }
      pageInfo {
        totalItems
      }
    }
  }
}
```

`TodoFilterGroupInput` accepts the same assignee/tag/list/date/existence fields as the flat filter, plus group-only conditions:

| Field                        | Type                       | Description                                                                                               |
| ---------------------------- | -------------------------- | --------------------------------------------------------------------------------------------------------- |
| `op`                         | `FilterLogicalOperator`    | How conditions combine within this group.                                                                 |
| `completedStart`             | `DateTime`                 | Completed on or after this (latest mark-complete timestamp on currently-done records).                    |
| `completedEnd`               | `DateTime`                 | Completed on or before this.                                                                              |
| `recordName`                 | `String`                   | Match against the record title (full-text).                                                               |
| `recordNameOp`               | `FilterComparisonOperator` | Operator for `recordName` (e.g. `CONTAINS`, `EQ`).                                                        |
| `lastUpdatedByUserIds`       | `[String!]`                | Records last updated by any of these users.                                                               |
| `lastUpdatedByAutomationIds` | `[String!]`                | Records last updated by any of these automations.                                                         |
| `lastUpdatedByActorTypes`    | `[ActorType!]`             | Records whose last-update actor is one of these types: `USER`, `AUTOMATION`, `API_KEY`, `SYSTEM`, `FORM`. |

The three `lastUpdatedBy*` axes compose with OR — a record matches if it satisfies any supplied axis. `fields` works exactly as in [Custom-field filtering](#custom-field-filtering).

## Errors

This query degrades gracefully: an inaccessible workspace or a non-existent record ID inside an organization you can access simply contributes no rows, rather than raising an error. Errors are reserved for authentication and authorization.

| Code              | When                                                                                                   |
| ----------------- | ------------------------------------------------------------------------------------------------------ |
| `UNAUTHENTICATED` | The request carries no valid credentials.                                                              |
| `FORBIDDEN`       | The token is valid but is not a member of the requested organization, or the organization is inactive. |

## Permissions

The caller must be authenticated and a member of every organization in `companyIds`; the organization must be active. Beyond that, visibility is scoped per user:

- Records in workspaces the user can't access are excluded.
- Users restricted to assigned-only records see only their own.
- Hidden lists (per role configuration) are excluded automatically.
- Tag-based access rules further narrow results.

This filtering is applied silently — restricted records are absent from `items`, not flagged.

## Related

- [Update a record](/api/records/update-record)
- [Toggle record status](/api/records/toggle-record-status)
- [Move a record between lists](/api/records/move-record-list)
- [Manage assignees](/api/records/assignees)
- [Manage tags](/api/records/tags)
- [List custom fields](/api/custom-fields/list-custom-fields)
- [Custom field values](/api/custom-fields/custom-field-values)
