---
title: Records
description: Create, read, update, and organize records - the core unit of work in Blue, modeled as Record objects in the API.
icon: SquareCheck
order: 0
---

Records are the core unit of work in Blue: the cards you create in a list, assign to people, tag, date, and fill with custom field values. In the API a record is a `Record` object, and it lives inside a list (`RecordList`) within a workspace (`Workspace`).

This section covers creating records and every operation you reach for next - reading, updating, moving, completing, and organizing them. The most common starting point, `createRecord`, is documented in full below.

```graphql
POST https://api.blue.app/graphql
blue-token-id: YOUR_TOKEN_ID
blue-token-secret: YOUR_TOKEN_SECRET
blue-org-id: YOUR_ORG_ID
blue-workspace-id: project_123
```

The `blue-workspace-id` header scopes record operations to one workspace. Company and project headers accept either an ID or a slug. Header names are case-insensitive.

## Operations

| Operation                                              | GraphQL                           | Description                                                                                                          |
| ------------------------------------------------------ | --------------------------------- | -------------------------------------------------------------------------------------------------------------------- |
| Create a record                                        | `createRecord` mutation           | Add a record to a list, with optional tags, assignees, dates, checklists, and custom field values. Documented below. |
| [List records](/api/records/list-records)              | `recordQueries { todos }` query   | Retrieve and filter records in a workspace.                                                                          |
| [Update a record](/api/records/update-record)          | `editRecord` mutation             | Change a record's title, dates, or other fields.                                                                     |
| [Toggle completion](/api/records/toggle-record-status) | `changeTodoDoneStatus` mutation   | Mark a record done or not done.                                                                                      |
| [Move a record](/api/records/move-record-list)         | `moveRecord` mutation             | Move a record to a different list.                                                                                   |
| [Copy a record](/api/records/copy-record)              | `copyRecord` mutation             | Duplicate an existing record.                                                                                        |
| [Assignees](/api/records/assignees)                    | `setRecordAssignees` mutation     | Assign or unassign users.                                                                                            |
| [Tags](/api/records/tags)                              | `setRecordTags` mutation          | Attach or detach tags.                                                                                               |
| [Comments](/api/records/add-comment)                   | `createComment` mutation          | Add a comment to a record.                                                                                           |
| [Dependencies](/api/records/dependencies)              | `createRecordDependency` mutation | Link records as blockers or dependents.                                                                              |
| [Recurring records](/api/records/recurring-records)    | `createRepeatingRecord` mutation  | Configure a repeating schedule.                                                                                      |
| [Archive a record](/api/records/archive-record)        | `archiveRecord` mutation          | Archive a record (and `unarchiveRecord` to restore it), hiding it from active lists while keeping its data.          |

Two lifecycle mutations don't yet have dedicated pages but round out the record surface:

| Operation           | GraphQL                                                    | Description                                                                                                                                            |
| ------------------- | ---------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------ |
| Delete a record     | `deleteRecord(input: DeleteRecordInput!): MutationResult!` | Move a record to Trash. `DeleteRecordInput` takes a single `todoId: String!`. Reversible with `restoreRecord(id: String!)` until the record is purged. |
| Restore a record    | `restoreRecord(id: String!): Record!`                      | Restore a record from Trash to its list (or the first available list if its original list is gone). Workspace `OWNER`/`ADMIN` only.                    |
| Bulk-update records | `updateRecords(input: UpdateRecordsInput!): Boolean!`      | Apply the same change to many records at once.                                                                                                         |

## Create a record

Use the `createRecord` mutation to add a record to a list. The only required input is a `title`; everything else - the target list, dates, assignees, tags, checklists, and custom field values - is optional and can be set in the same call.

### Request

The smallest call creates a record from a title alone. If you omit `todoListId`, the record lands in the workspace's first list (one is created if the workspace has none).

```graphql
mutation CreateRecord {
  createRecord(input: { title: "Draft launch plan" }) {
    id
    title
    position
  }
}
```

## Parameters

### CreateRecordInput

| Parameter      | Type                                 | Required | Description                                                                                                                                 |
| -------------- | ------------------------------------ | -------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `title`        | `String!`                            | Yes      | The record's title.                                                                                                                         |
| `todoListId`   | `String`                             | No       | List to add the record to. If omitted, the workspace's first list is used.                                                                  |
| `position`     | `Float`                              | No       | Explicit sort position within the list. Overrides `placement` when both are given.                                                          |
| `placement`    | `CreateRecordInputPlacement`         | No       | Where to insert the record when `position` is not set. Defaults to the top of the list.                                                     |
| `startedAt`    | `DateTime`                           | No       | Start date/time.                                                                                                                            |
| `duedAt`       | `DateTime`                           | No       | Due date/time.                                                                                                                              |
| `timezone`     | `String`                             | No       | IANA timezone used to interpret `startedAt`/`duedAt` (e.g. `America/New_York`). Only affects inference when `granularity` is omitted.       |
| `granularity`  | `DateGranularity`                    | No       | `ALL_DAY` or `TIMED`. See [All-day vs. timed dates](/api/records/update-record#all-day-vs-timed-dates). Omit to infer from the values sent. |
| `notify`       | `Boolean`                            | No       | Send creation notifications to assignees and watchers.                                                                                      |
| `description`  | `String`                             | No       | HTML body for the record. Sanitized server-side; stored as both `html` and plain `text`.                                                    |
| `assigneeIds`  | `[String!]`                          | No       | User IDs to assign. IDs that aren't workspace members are silently dropped.                                                                 |
| `checklists`   | `[CreateChecklistWithoutTodoInput!]` | No       | Checklists to create on the record.                                                                                                         |
| `customFields` | `[CreateRecordInputCustomField]`     | No       | Initial custom field values.                                                                                                                |
| `tags`         | `[CreateRecordTagInput!]`            | No       | Tags to attach.                                                                                                                             |

### CreateRecordInputPlacement

| Value    | Description                                                   |
| -------- | ------------------------------------------------------------- |
| `TOP`    | Insert above the current first record (the default behavior). |
| `BOTTOM` | Insert after the current last record.                         |

### CreateRecordTagInput

Provide either `id` (to attach an existing tag) or `title` (to find or create a tag by title).

| Parameter | Type     | Required | Description                                                          |
| --------- | -------- | -------- | -------------------------------------------------------------------- |
| `id`      | `String` | No       | ID of an existing tag in this workspace.                             |
| `title`   | `String` | No       | Tag title. A matching tag is reused; otherwise a new one is created. |
| `color`   | `String` | No       | Hex color for a newly created tag. Defaults to `#4a9fff`.            |

### CreateRecordInputCustomField

| Parameter       | Type              | Required | Description                                                                                          |
| --------------- | ----------------- | -------- | ---------------------------------------------------------------------------------------------------- |
| `customFieldId` | `String`          | No       | ID of the custom field to set. The field must belong to the same workspace.                          |
| `value`         | `String`          | No       | Value as a string, parsed according to the field's type (see below).                                 |
| `granularity`   | `DateGranularity` | No       | For a `DATE` field, `ALL_DAY` or `TIMED`. Omit to infer from `value`. Ignored for other field types. |

### CreateChecklistWithoutTodoInput

| Parameter  | Type      | Required | Description                                                                     |
| ---------- | --------- | -------- | ------------------------------------------------------------------------------- |
| `title`    | `String!` | Yes      | Checklist title.                                                                |
| `position` | `Float`   | No       | Sort position within the record. Defaults to gap-based ordering by array index. |

### Custom field value formats

The `value` string is parsed based on the target field's `CustomFieldType`. Common types:

| Field type                  | Format                                                           | Example                                     |
| --------------------------- | ---------------------------------------------------------------- | ------------------------------------------- |
| `TEXT_SINGLE`, `TEXT_MULTI` | Plain text                                                       | `"Any text content"`                        |
| `CHECKBOX`                  | `"true"`, `"1"`, or `"checked"` to check; anything else unchecks | `"true"`                                    |
| `NUMBER`                    | Numeric value (validated against the field's min/max)            | `"42"`                                      |
| `PERCENT`                   | Number, with optional `%`                                        | `"75"` or `"75%"`                           |
| `RATING`                    | Number within the field's range                                  | `"4"`                                       |
| `CURRENCY`                  | Amount with optional currency code                               | `"50000 USD"`                               |
| `DATE`                      | `YYYY-MM-DD`, or two comma-separated dates for a range           | `"2025-01-15"` or `"2025-01-15,2025-01-20"` |
| `PHONE`                     | Any parseable number; stored in international format             | `"+1 415 555 0132"`                         |
| `COUNTRY`                   | Country name or ISO alpha-2 code                                 | `"United States"` or `"US"`                 |
| `LOCATION`                  | `latitude,longitude`                                             | `"40.7128,-74.0060"`                        |
| `EMAIL`                     | Email address                                                    | `"user@example.com"`                        |
| `URL`                       | Web link                                                         | `"https://example.com"`                     |
| `SELECT_SINGLE`             | A custom field option ID                                         | `"option_123"`                              |
| `SELECT_MULTI`              | Comma-separated option IDs                                       | `"option_123,option_456"`                   |
| `ASSIGNEE`                  | Comma-separated user IDs (non-members and extras dropped)        | `"user_123,user_456"`                       |

Computed types - `FORMULA`, `LOOKUP`, `ROLLUP`, `REFERENCED_BY`, `TIME_DURATION` - derive their own values and ignore any value you pass at creation. See the [custom field types](/api/custom-fields) reference for full per-type details.

## Response

`createRecord` returns the created `Record`.

```json
{
  "data": {
    "createRecord": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Draft launch plan",
      "position": 131070.42
    }
  }
}
```

### Returns

The mutation returns the `Record` object. Common fields to select:

| Field                | Type              | Description                                                                                              |
| -------------------- | ----------------- | -------------------------------------------------------------------------------------------------------- |
| `id`                 | `ID!`             | Unique identifier for the record.                                                                        |
| `uid`                | `String!`         | Short reference identifier.                                                                              |
| `title`              | `String!`         | Record title. If custom fields drive a name formula, this reflects the computed title.                   |
| `position`           | `Float!`          | Sort position within the list.                                                                           |
| `text`               | `String!`         | Plain-text body.                                                                                         |
| `html`               | `String!`         | HTML body.                                                                                               |
| `done`               | `Boolean!`        | Whether the record is completed.                                                                         |
| `startedAt`          | `DateTime`        | Start date/time.                                                                                         |
| `duedAt`             | `DateTime`        | Due date/time.                                                                                           |
| `dueDateGranularity` | `DateGranularity` | `ALL_DAY` or `TIMED` — see [All-day vs. timed dates](/api/records/update-record#all-day-vs-timed-dates). |
| `todoList`           | `RecordList!`     | The list the record belongs to.                                                                          |
| `users`              | `[User!]!`        | Assigned users (select `id`, `fullName`, `email`).                                                       |
| `tags`               | `[Tag!]!`         | Attached tags (select `id`, `title`, `color`).                                                           |
| `checklists`         | `[Checklist!]!`   | Checklists on the record.                                                                                |
| `customFields`       | `[CustomField!]!` | Custom field definitions with this record's values.                                                      |

## Full example

Create a record with a target list, dates, assignees, tags, a checklist, and custom field values, then read back the resolved relations.

```graphql
mutation CreateRecordAdvanced {
  createRecord(
    input: {
      todoListId: "list_123"
      title: "Product launch planning"
      placement: BOTTOM
      description: "<p>Coordinate marketing, docs, and release.</p>"
      startedAt: "2025-01-15T09:00:00Z"
      duedAt: "2025-02-01T17:00:00Z"
      notify: true
      assigneeIds: ["user_123", "user_456"]
      tags: [{ id: "tag_123" }, { title: "Priority", color: "#ff4b4b" }, { title: "Marketing" }]
      customFields: [
        { customFieldId: "field_123", value: "50000 USD" }
        { customFieldId: "field_456", value: "option_123" }
      ]
      checklists: [{ title: "Pre-launch checklist", position: 1 }]
    }
  ) {
    id
    uid
    title
    position
    startedAt
    duedAt
    todoList {
      id
      title
    }
    users {
      id
      fullName
    }
    tags {
      id
      title
      color
    }
  }
}
```

## Errors

| Code                                | When                                                                                                                                      |
| ----------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------- |
| `PROJECT_NOT_FOUND`                 | No workspace context - the `blue-workspace-id` header is missing or invalid.                                                              |
| `FORBIDDEN`                         | The caller's workspace role is `VIEW_ONLY` or `COMMENT_ONLY`, or records are disabled for the workspace.                                  |
| `PLAN_LIMIT_REACHED`                | The workspace or organization is at its record limit for the current plan.                                                                |
| `TODO_LIST_NOT_FOUND`               | The `todoListId` doesn't exist or the caller can't access it.                                                                             |
| `TODO_LIST_CREATE_TODO_LIMIT_ERROR` | The target list already holds the maximum of 100,000 records.                                                                             |
| `CUSTOM_FIELD_VALUE_PARSE_ERROR`    | A `customFields` value fails validation for its field type (e.g. an unparseable phone number, a rating out of range, an invalid country). |

```json
{
  "errors": [
    {
      "message": "Invalid phone number format.",
      "extensions": { "code": "CUSTOM_FIELD_VALUE_PARSE_ERROR" }
    }
  ]
}
```

## Behavior

- **Position.** When `position` is set it wins. Otherwise `placement` decides: the default (no placement) and `TOP` insert above the current first record; `BOTTOM` inserts after the last. The gap constant `65535` is only used directly when the list is empty.
- **Dates.** Passing only `duedAt` sets `startedAt` to the start of that day; passing only `startedAt` sets `duedAt` to the same point in time. Pass both to control each independently.
- **Tags.** A `CreateRecordTagInput` with a `title` reuses an existing tag of the same title and color, or creates a new one (default color `#4a9fff`). A `title` with no match creates the tag.
- **Custom fields.** Each value is parsed by the field's type; computed fields ignore supplied values. If the record's workspace uses a name-formula field, the returned `title` reflects the computed name rather than the literal input.

## Related

- [List records](/api/records/list-records)
- [Update a record](/api/records/update-record)
- [Toggle completion](/api/records/toggle-record-status)
- [Custom field values](/api/custom-fields/custom-field-values)
- [Tags](/api/records/tags)
