---
title: Tags
description: Create, list, edit, and delete tags, apply them to records, and trigger AI tagging with the Blue API.
icon: Tag
order: 7
---

Tags are colored labels you attach to records to categorize and filter them. Tags are scoped to a single workspace — the same label in two workspaces is two separate tags. Tags are `Tag` objects and records are `Todo` objects in the API.

Use the `tagList` query to read tags, `createTag` / `editTag` / `deleteTag` to manage them, `setTodoTags` to apply tags to a record, and `aiTag` to trigger AI-suggested tagging in the background.

The workspace is taken from the `X-Bloo-Project-ID` header on tag mutations, so you don't pass a workspace ID to `createTag` or `setTodoTags`.

## List tags

Use the `tagList` query to read tags across one or more workspaces. The only required argument is `filter`.

```graphql
query ListTags {
  tagList(filter: { projectIds: ["project_123"] }, orderBy: title_ASC, first: 50) {
    items {
      id
      title
      color
      project {
        id
        name
      }
    }
    totalCount
    pageInfo {
      page
      perPage
      totalItems
      hasNextPage
    }
  }
}
```

### tagList arguments

| Argument   | Type                       | Required | Description                                                                                                   |
| ---------- | -------------------------- | -------- | ------------------------------------------------------------------------------------------------------------- |
| `filter`   | `TagListFilter!`           | Yes      | Which tags to return. See below.                                                                              |
| `orderBy`  | `TagOrderByInput`          | No       | Sort order. Defaults to `title_ASC`.                                                                          |
| `first`    | `Int`                      | No       | Maximum number of tags to return. Defaults to `500`.                                                          |
| `skip`     | `Int`                      | No       | Number of tags to skip (offset pagination).                                                                   |
| `distinct` | `[TagListFilterDistinct!]` | No       | Collapse duplicates by `title` and/or `color` — useful when the same label exists across multiple workspaces. |
| `after`    | `String`                   | No       | Cursor pagination argument.                                                                                   |
| `before`   | `String`                   | No       | Cursor pagination argument.                                                                                   |
| `last`     | `Int`                      | No       | Cursor pagination argument.                                                                                   |

### TagListFilter

| Parameter                 | Type        | Required | Description                                                                                   |
| ------------------------- | ----------- | -------- | --------------------------------------------------------------------------------------------- |
| `projectIds`              | `[String!]` | No       | Workspace IDs or slugs to read tags from. Omit to read across every workspace you can access. |
| `excludeArchivedProjects` | `Boolean`   | No       | When `true`, skips tags in archived workspaces.                                               |
| `search`                  | `String`    | No       | Case-insensitive substring match on the tag title.                                            |
| `titles`                  | `[String!]` | No       | Return only tags with one of these exact titles.                                              |
| `colors`                  | `[String!]` | No       | Return only tags with one of these hex colors.                                                |
| `tagIds`                  | `[String!]` | No       | Return only tags with one of these IDs.                                                       |

### TagOrderByInput

One of: `id_ASC`, `id_DESC`, `uid_ASC`, `uid_DESC`, `title_ASC`, `title_DESC`, `color_ASC`, `color_DESC`, `createdAt_ASC`, `createdAt_DESC`, `updatedAt_ASC`, `updatedAt_DESC`.

### Response

```json
{
  "data": {
    "tagList": {
      "items": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "title": "High Priority",
          "color": "#ff0000",
          "project": { "id": "clm4n8a1b000008l0abcd1234", "name": "Roadmap" }
        }
      ],
      "totalCount": 1,
      "pageInfo": {
        "page": 1,
        "perPage": 50,
        "totalItems": 1,
        "hasNextPage": false
      }
    }
  }
}
```

### Returns

`tagList` returns a `TagPagination` object.

| Field        | Type        | Description                               |
| ------------ | ----------- | ----------------------------------------- |
| `items`      | `[Tag!]!`   | The tags matching the filter.             |
| `totalCount` | `Int!`      | Total number of tags matching the filter. |
| `pageInfo`   | `PageInfo!` | Pagination metadata.                      |

#### Tag

| Field       | Type        | Description                                        |
| ----------- | ----------- | -------------------------------------------------- |
| `id`        | `ID!`       | Unique identifier for the tag.                     |
| `uid`       | `String!`   | Short human-readable identifier.                   |
| `title`     | `String!`   | Tag label.                                         |
| `color`     | `String!`   | Tag color as a hex string (for example `#ff0000`). |
| `project`   | `Project!`  | The workspace this tag belongs to.                 |
| `todos`     | `[Todo!]!`  | Records this tag is applied to.                    |
| `createdAt` | `DateTime!` | When the tag was created.                          |
| `updatedAt` | `DateTime!` | When the tag was last modified.                    |

#### PageInfo

| Field             | Type       | Description                     |
| ----------------- | ---------- | ------------------------------- |
| `totalItems`      | `Int`      | Total items across all pages.   |
| `totalPages`      | `Int`      | Total number of pages.          |
| `page`            | `Int`      | Current page number.            |
| `perPage`         | `Int`      | Items per page.                 |
| `hasNextPage`     | `Boolean!` | Whether another page follows.   |
| `hasPreviousPage` | `Boolean!` | Whether a previous page exists. |

## Create a tag

Use the `createTag` mutation to add a tag to the workspace set in the `X-Bloo-Project-ID` header. Only `color` is required; if you omit `title`, the tag is created with an empty label that you can set later with `editTag`.

```graphql
mutation CreateTag {
  createTag(input: { title: "High Priority", color: "#ff0000" }) {
    id
    title
    color
  }
}
```

### CreateTagInput

| Parameter | Type      | Required | Description                                        |
| --------- | --------- | -------- | -------------------------------------------------- |
| `title`   | `String`  | No       | Tag label. Defaults to an empty string if omitted. |
| `color`   | `String!` | Yes      | Tag color as a hex string, for example `#ff0000`.  |

### Response

```json
{
  "data": {
    "createTag": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "High Priority",
      "color": "#ff0000"
    }
  }
}
```

`createTag` returns the created [Tag](#tag).

## Edit a tag

Use the `editTag` mutation to rename a tag or change its color. Both `title` and `color` are optional — only the fields you pass are updated. Renaming a tag updates it on every record it is applied to.

```graphql
mutation EditTag {
  editTag(input: { id: "tag_123", title: "Critical", color: "#cc0000" }) {
    id
    title
    color
    updatedAt
  }
}
```

### EditTagInput

| Parameter | Type      | Required | Description                                         |
| --------- | --------- | -------- | --------------------------------------------------- |
| `id`      | `String!` | Yes      | ID of the tag to update.                            |
| `title`   | `String`  | No       | New label. Leave out to keep the current title.     |
| `color`   | `String`  | No       | New hex color. Leave out to keep the current color. |

### Response

```json
{
  "data": {
    "editTag": {
      "id": "tag_123",
      "title": "Critical",
      "color": "#cc0000",
      "updatedAt": "2026-05-29T10:42:00.000Z"
    }
  }
}
```

`editTag` returns the updated [Tag](#tag).

## Delete a tag

Use the `deleteTag` mutation to permanently delete a tag. It is removed from every record it was applied to, and any automation that uses the tag as a trigger or action is deleted along with it. Returns `true` on success.

```graphql
mutation DeleteTag {
  deleteTag(id: "tag_123")
}
```

### Response

```json
{
  "data": {
    "deleteTag": true
  }
}
```

<Callout variant="warning" title="Deletion is permanent">

Deleting a tag also deletes any automation whose trigger or action references it. There is no undo — recreate the tag and re-apply it if you delete one by mistake.

</Callout>

## Apply tags to a record

Use the `setTodoTags` mutation to set the complete tag set on a record. This **replaces** the record's current tags in the workspace — any tag not included is removed, and any new tag is added. Returns `true` on success.

Pass tag IDs in `tagIds`, or create-and-apply tags by title in `tagTitles`. Titles that match an existing tag in the workspace reuse it; titles that don't are created as new tags with the default color `#4a9fff`.

```graphql
mutation SetRecordTags {
  setTodoTags(input: { todoId: "todo_123", tagIds: ["tag_123", "tag_456"] })
}
```

To create tags on the fly while applying them, supply `tagTitles`:

```graphql
mutation SetRecordTagsByTitle {
  setTodoTags(
    input: { todoId: "todo_123", tagIds: ["tag_123"], tagTitles: ["Needs Review", "Blocked"] }
  )
}
```

To clear all tags from a record, call `setTodoTags` with an empty `tagIds` array.

### SetTodoTagsInput

| Parameter   | Type        | Required | Description                                                                           |
| ----------- | ----------- | -------- | ------------------------------------------------------------------------------------- |
| `todoId`    | `String!`   | Yes      | ID of the record to tag.                                                              |
| `tagIds`    | `[String!]` | No       | IDs of existing tags to apply. Tags omitted here are removed from the record.         |
| `tagTitles` | `[String!]` | No       | Titles to apply. Matching tags are reused; new titles are created at color `#4a9fff`. |

### Response

`setTodoTags` returns `Boolean!`. Read the record back with `tagList` or a record query to see the resulting tags.

```json
{
  "data": {
    "setTodoTags": true
  }
}
```

## AI tagging

Use the `aiTag` mutation to trigger AI-suggested tagging for one or more records. The work runs in the background, so the mutation returns immediately with an `operationId` rather than the tags themselves. The AI reads each record's content, then creates and applies tags asynchronously.

```graphql
mutation AiTagRecords {
  aiTag(input: { todoIds: ["todo_123", "todo_456"] }) {
    success
    operationId
  }
}
```

You can also target an entire workspace or a single list instead of specific records.

### AITagInput

| Parameter    | Type        | Required | Description                                                                                     |
| ------------ | ----------- | -------- | ----------------------------------------------------------------------------------------------- |
| `todoIds`    | `[String!]` | No       | Records to tag.                                                                                 |
| `projectId`  | `String`    | No       | Tag all records in this workspace. Defaults to the workspace in the `X-Bloo-Project-ID` header. |
| `todoListId` | `String`    | No       | Tag all records in this list.                                                                   |

### Response

`aiTag` returns a `MutationResult`. Because tagging happens in the background, the applied tags are not in the response — read them back with `tagList` or a record query once the operation completes.

```json
{
  "data": {
    "aiTag": {
      "success": true,
      "operationId": "clm4n8qwx000008l0g4oxdqn7"
    }
  }
}
```

| Field         | Type       | Description                                      |
| ------------- | ---------- | ------------------------------------------------ |
| `success`     | `Boolean!` | Whether the tagging job was queued.              |
| `operationId` | `String`   | Identifier for the background tagging operation. |

## Errors

| Code                | When                                                                                                  |
| ------------------- | ----------------------------------------------------------------------------------------------------- |
| `TAG_NOT_FOUND`     | The `id` passed to `editTag` or `deleteTag` does not match a tag you can access.                      |
| `TODO_NOT_FOUND`    | The `todoId` passed to `setTodoTags` does not match a record you can access.                          |
| `PROJECT_NOT_FOUND` | No valid workspace context — set the `X-Bloo-Project-ID` header (or a valid `projectId` for `aiTag`). |
| `FORBIDDEN`         | Your access level is too low for the action, or the workspace is archived.                            |
| `BAD_USER_INPUT`    | A required argument is missing or malformed (for example a missing `color` on `createTag`).           |

## Permissions

Reading tags with `tagList` requires access to the workspace. Creating, editing, deleting, and applying tags require an access level of `OWNER`, `ADMIN`, `MEMBER`, or `CLIENT`; `COMMENT_ONLY` and `VIEW_ONLY` users cannot modify tags. The workspace must be active — tag mutations are rejected on archived workspaces.

## Related

- [List records](/api/records/list-records)
- [Record assignees](/api/records/assignees)
- [Create an automation](/api/automations/create-automation) — use tags as triggers or actions
- [Workspaces](/api/workspaces)
