---
title: Edit Workspace
description: Rename a workspace and update its description, color, icon, category, and other settings with the editWorkspace mutation.
icon: FileEdit
order: 3
---

Use the `editWorkspace` mutation to update an existing workspace. Despite the "rename" framing, this single mutation edits the full workspace surface — name, slug, description, color, icon, category, the record alias, time-tracking display, bulk-action toggles, the name formula, view types, and more. Pass only the fields you want to change; everything else is left untouched.

Workspaces are `Workspace` objects in the API.

## Request

The minimal call renames a workspace. Only `projectId` is required.

```graphql
mutation RenameWorkspace {
  editWorkspace(input: { projectId: "project_123", name: "Q2 Marketing Campaign" }) {
    id
    name
    slug
  }
}
```

## Parameters

### EditWorkspaceInput

| Parameter                 | Type                      | Required | Description                                                                                                                                    |
| ------------------------- | ------------------------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------- |
| `projectId`               | `String!`                 | Yes      | ID of the workspace to edit.                                                                                                                   |
| `name`                    | `String`                  | No       | New display name.                                                                                                                              |
| `slug`                    | `String`                  | No       | New URL slug. Lowercase letters, numbers, and hyphens only; must not contain a URL or collide with an existing organization or workspace slug. |
| `description`             | `String`                  | No       | Workspace description. HTML is sanitized on save.                                                                                              |
| `image`                   | `ImageInput`              | No       | Workspace cover image. Pass `null` to remove the current image.                                                                                |
| `color`                   | `String`                  | No       | Hex color, e.g. `#3B82F6`.                                                                                                                     |
| `icon`                    | `String`                  | No       | Icon identifier.                                                                                                                               |
| `category`                | `ProjectCategory`         | No       | Workspace category. See the values below.                                                                                                      |
| `todoAlias`               | `String`                  | No       | Custom singular label for records in this workspace (e.g. `Task`, `Deal`).                                                                     |
| `hideRecordCount`         | `Boolean`                 | No       | Hide record counts in the UI.                                                                                                                  |
| `showTimeSpentInTodoList` | `Boolean`                 | No       | Show time tracked per list.                                                                                                                    |
| `showTimeSpentInProject`  | `Boolean`                 | No       | Show time tracked across the workspace.                                                                                                        |
| `duplicateDetection`      | `Boolean`                 | No       | Warn when a record looks like a duplicate.                                                                                                     |
| `compactCards`            | `Boolean`                 | No       | Render board cards in compact mode.                                                                                                            |
| `sequenceCustomFieldId`   | `String`                  | No       | ID of a custom field used to sequence records. Pass `null` to clear it.                                                                        |
| `nameFormula`             | `JSON`                    | No       | Formula configuration that auto-generates record titles. Pass `null` to clear it.                                                              |
| `nameFormulaEnabled`      | `Boolean`                 | No       | Enable or disable the name formula. Enabling requires a configured formula.                                                                    |
| `bulkActionComplete`      | `Boolean`                 | No       | Allow the bulk "complete" action.                                                                                                              |
| `bulkActionAssignee`      | `Boolean`                 | No       | Allow the bulk "assignee" action.                                                                                                              |
| `bulkActionTag`           | `Boolean`                 | No       | Allow the bulk "tag" action.                                                                                                                   |
| `bulkActionDueDate`       | `Boolean`                 | No       | Allow the bulk "due date" action.                                                                                                              |
| `viewTypeSettings`        | `[ViewTypeSettingInput!]` | No       | Which view types are available and how they are labeled. At least one entry must be enabled.                                                   |
| `todoFields`              | `[TodoFieldInput]`        | No       | Ordered field layout for records in this workspace.                                                                                            |
| `features`                | `[ProjectFeatureInput]`   | No       | Per-feature toggles for the workspace.                                                                                                         |
| `coverConfig`             | `TodoCoverConfigInput`    | No       | How record cover images are sourced and displayed.                                                                                             |

### ProjectCategory

| Value              | Description                      |
| ------------------ | -------------------------------- |
| `CRM`              | Customer relationship management |
| `CROSS_FUNCTIONAL` | Cross-functional teams           |
| `CUSTOMER_SUCCESS` | Customer success                 |
| `DESIGN`           | Design and creative              |
| `ENGINEERING`      | Engineering and development      |
| `GENERAL`          | General-purpose (default)        |
| `HR`               | Human resources                  |
| `IT`               | Information technology           |
| `MARKETING`        | Marketing                        |
| `OPERATIONS`       | Operations and logistics         |
| `PRODUCT`          | Product management               |
| `SALES`            | Sales and business development   |

### ViewTypeSettingInput

| Parameter | Type             | Required | Description                                                |
| --------- | ---------------- | -------- | ---------------------------------------------------------- |
| `type`    | `SavedViewType!` | Yes      | One of `BOARD`, `DATABASE`, `CALENDAR`, `TIMELINE`, `MAP`. |
| `enabled` | `Boolean!`       | Yes      | Whether this view type is available in the workspace.      |
| `name`    | `String`         | No       | Custom label for the view (1–100 characters).              |

## Response

```json
{
  "data": {
    "editWorkspace": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Q2 Marketing Campaign",
      "slug": "q2-marketing-campaign"
    }
  }
}
```

### Returns

`editWorkspace` returns the updated `Workspace` (non-null).

| Field                | Type               | Description                         |
| -------------------- | ------------------ | ----------------------------------- |
| `id`                 | `ID!`              | Workspace ID.                       |
| `name`               | `String!`          | Workspace name.                     |
| `slug`               | `String!`          | URL slug.                           |
| `description`        | `String`           | Workspace description.              |
| `color`              | `String`           | Hex color.                          |
| `icon`               | `String`           | Icon identifier.                    |
| `category`           | `ProjectCategory!` | Workspace category.                 |
| `todoAlias`          | `String`           | Custom singular record label.       |
| `hideRecordCount`    | `Boolean`          | Whether record counts are hidden.   |
| `archived`           | `Boolean`          | Whether the workspace is archived.  |
| `nameFormulaEnabled` | `Boolean!`         | Whether the name formula is active. |
| `createdAt`          | `DateTime!`        | Creation timestamp.                 |
| `updatedAt`          | `DateTime!`        | Last update timestamp.              |

## Full example

Update several settings at once. Categories and view types use enum values (no quotes).

```graphql
mutation EditWorkspace {
  editWorkspace(
    input: {
      projectId: "project_123"
      name: "Q2 Marketing Campaign"
      description: "Campaign planning for the Q2 product launch."
      color: "#3B82F6"
      icon: "campaign"
      category: MARKETING
      todoAlias: "Task"
      hideRecordCount: false
      duplicateDetection: true
      compactCards: true
      viewTypeSettings: [
        { type: BOARD, enabled: true, name: "Pipeline" }
        { type: DATABASE, enabled: true }
        { type: CALENDAR, enabled: false }
      ]
    }
  ) {
    id
    name
    slug
    description
    color
    icon
    category
    todoAlias
    hideRecordCount
    updatedAt
  }
}
```

## Errors

| Code                         | When                                                                                                                                                            |
| ---------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `FORBIDDEN`                  | The caller is not an OWNER or ADMIN of the workspace. Message: `You are not authorized.`                                                                        |
| `PROJECT_NOT_FOUND`          | No workspace matches `projectId`.                                                                                                                               |
| `PROJECT_URL_ALREADY_EXISTS` | The requested `slug` collides with an existing organization or workspace slug.                                                                                  |
| `BAD_USER_INPUT`             | The `slug` contains a URL or invalid characters, or a `viewTypeSettings` entry is invalid (none enabled, duplicate types, or an empty/over-100-character name). |
| `CUSTOM_FIELD_NOT_FOUND`     | The `sequenceCustomFieldId` does not reference a custom field in this workspace.                                                                                |

```json
{
  "errors": [
    {
      "message": "You are not authorized.",
      "extensions": { "code": "FORBIDDEN" }
    }
  ]
}
```

## Permissions

Editing a workspace requires the **OWNER** or **ADMIN** access level on that workspace. MEMBER, CLIENT, COMMENT_ONLY, and VIEW_ONLY members cannot call `editWorkspace`; the mutation throws `FORBIDDEN`.

| Workspace access level | Can edit |
| ---------------------- | -------- |
| `OWNER`                | Yes      |
| `ADMIN`                | Yes      |
| `MEMBER`               | No       |
| `CLIENT`               | No       |
| `COMMENT_ONLY`         | No       |
| `VIEW_ONLY`            | No       |

## Notes

- **Renaming does not change the slug.** The slug stays as-is unless you pass a new `slug` explicitly. Provide one if you want the URL to match the new name.
- **Slug rules.** A `slug` must be lowercase letters, numbers, and hyphens, contain no URL, and not collide with an existing organization or workspace slug.
- **Partial updates.** Every field except `projectId` is optional — send only the fields you want to change.
- **Descriptions are sanitized.** HTML in `description` is stripped on save.
- **Removing the cover image.** Pass `image: null` to delete the current cover image.

## Related

- [Create a Workspace](/api/workspaces/create-workspace)
- [List Workspaces](/api/workspaces/list-workspaces)
- [Archive Workspace](/api/workspaces/archive-workspace)
- [Delete a Workspace](/api/workspaces/delete-workspace)
- [Manage Lists](/api/workspaces/lists)
