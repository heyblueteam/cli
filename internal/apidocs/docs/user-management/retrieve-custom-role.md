---
title: Custom Roles
description: List, create, update, and delete custom project roles with granular permission and field-level access control.
icon: Shield
order: 4
---

Custom roles let you define precise permission sets for members of a workspace, beyond the built-in access levels. A role controls which sections a member sees (records, files, chat, docs, …), what they can do (invite, complete, delete), and—optionally—which records, lists, and fields they can view or edit. Roles are `ProjectUserRole` objects in the API; workspaces are `Project` objects.

You assign a role when you [invite someone](/api/user-management/invite-user) by passing its `roleId` with `accessLevel: MEMBER`. The role must belong to the same workspace as the invite.

This page covers four operations:

- `projectUserRoles` — list the roles in one or more workspaces.
- `createProjectUserRole` — create a role.
- `updateProjectUserRole` — change a role.
- `deleteProjectUserRole` — delete a role.

## Request

Send every request to `https://api.blue.app/graphql` with your authentication headers. Custom roles are scoped to a workspace, so pass the workspace ID (or slug) in the operation input. Company and project arguments accept either an ID or a slug.

```http
POST https://api.blue.app/graphql
X-Bloo-Token-ID: YOUR_TOKEN_ID
X-Bloo-Token-Secret: YOUR_TOKEN_SECRET
X-Bloo-Company-ID: YOUR_COMPANY_ID
```

List the roles in a workspace. The `filter` argument is required; supply `projectId` for a single workspace.

```graphql
query ListProjectRoles {
  projectUserRoles(filter: { projectId: "project_123" }) {
    id
    name
    description
    allowInviteOthers
    canDeleteRecords
    isRecordsEnabled
  }
}
```

To query several workspaces at once, pass `projectIds` instead. Either field may be null, but you must pass the `filter` object itself. Roles are only returned for workspaces the calling user belongs to.

```graphql
query ListRolesAcrossWorkspaces {
  projectUserRoles(filter: { projectIds: ["project_123", "project_456"] }) {
    id
    name
    project {
      id
      name
    }
  }
}
```

## Parameters

### ProjectUserRoleFilter

The required `filter` argument for `projectUserRoles`.

| Parameter    | Type       | Required | Description                                                               |
| ------------ | ---------- | -------- | ------------------------------------------------------------------------- |
| `projectId`  | `String`   | No       | Single workspace ID or slug. Takes precedence over `projectIds` when set. |
| `projectIds` | `[String]` | No       | Multiple workspace IDs or slugs, for a cross-workspace listing.           |

### CreateProjectUserRoleInput

The input for `createProjectUserRole`. Only `projectId` and `name` are required; every other field is optional and falls back to the default shown.

| Parameter                   | Type                                       | Required | Default  | Description                                                         |
| --------------------------- | ------------------------------------------ | -------- | -------- | ------------------------------------------------------------------- |
| `projectId`                 | `String!`                                  | Yes      | —        | Workspace the role belongs to (ID or slug).                         |
| `name`                      | `String!`                                  | Yes      | —        | Display name of the role.                                           |
| `description`               | `String`                                   | No       | `null`   | Optional description.                                               |
| **Action permissions**      |                                            |          |          |                                                                     |
| `allowInviteOthers`         | `Boolean`                                  | No       | `false`  | May invite new users to the workspace.                              |
| `allowMarkRecordsAsDone`    | `Boolean`                                  | No       | `false`  | May mark records complete.                                          |
| `canCreateRecords`          | `Boolean`                                  | No       | `true`   | May create records.                                                 |
| `canDeleteRecords`          | `Boolean`                                  | No       | `true`   | May delete records.                                                 |
| `canDeleteFiles`            | `Boolean`                                  | No       | `true`   | May delete files.                                                   |
| `canCreateLists`            | `Boolean`                                  | No       | `true`   | May create lists.                                                   |
| `canEditLists`              | `Boolean`                                  | No       | `true`   | May rename and reorder lists.                                       |
| **Section access**          |                                            |          |          |                                                                     |
| `isRecordsEnabled`          | `Boolean`                                  | No       | `false`  | Access to the records views.                                        |
| `isActivityEnabled`         | `Boolean`                                  | No       | `false`  | Access to the activity feed.                                        |
| `isChatEnabled`             | `Boolean`                                  | No       | `false`  | Access to chat.                                                     |
| `isDocsEnabled`             | `Boolean`                                  | No       | `false`  | Access to docs.                                                     |
| `isFilesEnabled`            | `Boolean`                                  | No       | `false`  | Access to files.                                                    |
| `isFormsEnabled`            | `Boolean`                                  | No       | `false`  | Access to forms.                                                    |
| `isWikiEnabled`             | `Boolean`                                  | No       | `false`  | Access to the wiki.                                                 |
| `isPeopleEnabled`           | `Boolean`                                  | No       | `false`  | Access to the people/members section.                               |
| **Visibility filters**      |                                            |          |          |                                                                     |
| `showOnlyAssignedTodos`     | `Boolean`                                  | No       | `false`  | Members see only records assigned to them.                          |
| `showOnlyMentionedComments` | `Boolean`                                  | No       | `false`  | Members see only comments that mention them.                        |
| **Field-level scoping**     |                                            |          |          |                                                                     |
| `recordTagFilter`           | `RecordTagFilterInput`                     | No       | disabled | Restrict visible records by tag (allow/deny lists).                 |
| `recordFilter`              | `RecordFilterInput`                        | No       | none     | Restrict visible records to those matching a filter.                |
| `customFields`              | `[CreateProjectUserRoleCustomFieldInput!]` | No       | none     | Per–custom-field view/edit access.                                  |
| `todoLists`                 | `[CreateProjectUserRoleTodoListInput!]`    | No       | none     | Per-list view/edit/delete access.                                   |
| `todoFields`                | `[CreateProjectUserRoleTodoFieldInput!]`   | No       | none     | Per built-in record field (assignee, due date, …) view/edit access. |

<Callout variant="warning" title="Section access defaults to off">

Every `is*Enabled` flag defaults to `false`. A role created with only `projectId` and `name` grants no section access — members assigned to it see nothing until you enable at least one section. The `can*` permissions, by contrast, default to `true`. Set the flags you care about explicitly rather than relying on the mixed defaults.

</Callout>

### RecordTagFilterInput

Limit which records a role can see based on their tags. Tags are `Tag` objects (`Tag.title` in the API).

| Parameter       | Type        | Required | Description                                          |
| --------------- | ----------- | -------- | ---------------------------------------------------- |
| `enabled`       | `Boolean`   | No       | Turn tag-based filtering on for the role.            |
| `allowedTagIds` | `[String!]` | No       | Only records carrying one of these tags are visible. |
| `deniedTagIds`  | `[String!]` | No       | Records carrying one of these tags are hidden.       |

### RecordFilterInput

Restrict visible records to those matching a filter, using the same group structure as record queries.

| Parameter    | Type                       | Required | Description                                                           |
| ------------ | -------------------------- | -------- | --------------------------------------------------------------------- |
| `groups`     | `[TodoFilterGroupInput!]!` | Yes      | One or more filter groups; a record is visible if it matches.         |
| `groupLinks` | `[FilterLogicalOperator!]` | No       | `AND` / `OR` operators joining the groups. Defaults to an empty list. |

### CreateProjectUserRoleCustomFieldInput

Per-field access for a custom field (`CustomField`).

| Parameter       | Type      | Required | Default | Description                           |
| --------------- | --------- | -------- | ------- | ------------------------------------- |
| `customFieldId` | `String!` | Yes      | —       | The custom field to scope.            |
| `viewable`      | `Boolean` | No       | `true`  | Members can read the field's value.   |
| `editable`      | `Boolean` | No       | `false` | Members can change the field's value. |

### CreateProjectUserRoleTodoListInput

Per-list access for a list (`TodoList`).

| Parameter    | Type       | Required | Description                             |
| ------------ | ---------- | -------- | --------------------------------------- |
| `todoListId` | `String!`  | Yes      | The list to scope.                      |
| `viewable`   | `Boolean!` | Yes      | Members can see the list.               |
| `editable`   | `Boolean!` | Yes      | Members can edit records in the list.   |
| `deletable`  | `Boolean!` | Yes      | Members can delete records in the list. |

### CreateProjectUserRoleTodoFieldInput

Per-field access for a built-in record field.

| Parameter   | Type             | Required | Default | Description                            |
| ----------- | ---------------- | -------- | ------- | -------------------------------------- |
| `fieldType` | `TodoFieldType!` | Yes      | —       | Which built-in field (see enum below). |
| `viewable`  | `Boolean!`       | Yes      | —       | Members can read the field.            |
| `editable`  | `Boolean!`       | Yes      | —       | Members can edit the field.            |

`TodoFieldType` values: `DUE_DATE`, `ASSIGNEE`, `TAG`, `DEPENDENCY`, `CUSTOM_FIELD`, `DESCRIPTION`, `CHECKLIST`, `REFERENCED_BY`, `CUSTOM_FIELD_GROUP`, `TIME_TRACKING`, `CREATED_DATE`, `UPDATED_DATE`, `COMPLETED_DATE`, `PROJECTS`, `LAST_UPDATED_BY`.

### UpdateProjectUserRoleInput

The input for `updateProjectUserRole`. Same fields as `CreateProjectUserRoleInput`, with two differences:

- `roleId: String!` (required) — the role to update.
- `projectId: String!` (required) — the workspace the role belongs to, unchanged.
- `name` is **optional** here (`String`, not `String!`); omit it to keep the current name.

Any flag you omit keeps its existing value (the update merges over the stored role rather than resetting to defaults).

### DeleteProjectUserRoleInput

| Parameter   | Type      | Required | Description                        |
| ----------- | --------- | -------- | ---------------------------------- |
| `roleId`    | `String!` | Yes      | The role to delete.                |
| `projectId` | `String!` | Yes      | The workspace the role belongs to. |

## Response

`projectUserRoles` returns an array of `ProjectUserRole` objects.

```json
{
  "data": {
    "projectUserRoles": [
      {
        "id": "clm4n8qwx000008l0g4oxdqn7",
        "name": "External Contractor",
        "description": "Limited access for outside collaborators",
        "allowInviteOthers": false,
        "canDeleteRecords": false,
        "isRecordsEnabled": true
      }
    ]
  }
}
```

`createProjectUserRole` and `updateProjectUserRole` both return the `ProjectUserRole`.

```json
{
  "data": {
    "createProjectUserRole": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "External Contractor"
    }
  }
}
```

`deleteProjectUserRole` returns a nullable `Boolean` — `true` on success.

```json
{ "data": { "deleteProjectUserRole": true } }
```

### Returns — ProjectUserRole

| Field                       | Type                           | Description                                                               |
| --------------------------- | ------------------------------ | ------------------------------------------------------------------------- |
| `id`                        | `ID!`                          | Unique role identifier.                                                   |
| `uid`                       | `String!`                      | Stable public identifier.                                                 |
| `name`                      | `String!`                      | Role name.                                                                |
| `description`               | `String`                       | Role description.                                                         |
| `createdAt`                 | `DateTime!`                    | When the role was created.                                                |
| `updatedAt`                 | `DateTime`                     | When the role was last updated.                                           |
| `project`                   | `Project!`                     | The workspace the role belongs to.                                        |
| **Action permissions**      |                                |                                                                           |
| `allowInviteOthers`         | `Boolean`                      | May invite users.                                                         |
| `allowMarkRecordsAsDone`    | `Boolean`                      | May complete records.                                                     |
| `canCreateRecords`          | `Boolean`                      | May create records.                                                       |
| `canDeleteRecords`          | `Boolean`                      | May delete records.                                                       |
| `canDeleteFiles`            | `Boolean`                      | May delete files.                                                         |
| `canCreateLists`            | `Boolean`                      | May create lists.                                                         |
| `canEditLists`              | `Boolean`                      | May edit lists.                                                           |
| **Section access**          |                                |                                                                           |
| `isRecordsEnabled`          | `Boolean`                      | Records access.                                                           |
| `isActivityEnabled`         | `Boolean`                      | Activity access.                                                          |
| `isChatEnabled`             | `Boolean`                      | Chat access.                                                              |
| `isDocsEnabled`             | `Boolean`                      | Docs access.                                                              |
| `isFilesEnabled`            | `Boolean`                      | Files access.                                                             |
| `isFormsEnabled`            | `Boolean`                      | Forms access.                                                             |
| `isWikiEnabled`             | `Boolean`                      | Wiki access.                                                              |
| `isPeopleEnabled`           | `Boolean`                      | People section access.                                                    |
| **Visibility filters**      |                                |                                                                           |
| `showOnlyAssignedTodos`     | `Boolean`                      | Restricts records to those assigned to the member.                        |
| `showOnlyMentionedComments` | `Boolean`                      | Restricts comments to those mentioning the member.                        |
| **Field-level scoping**     |                                |                                                                           |
| `recordTagFilter`           | `RecordTagFilter`              | Tag-based record visibility (`enabled`, `allowedTagIds`, `deniedTagIds`). |
| `recordFilter`              | `ProjectUserRoleRecordFilter`  | Filter-based record visibility (`groups`, `groupLinks`).                  |
| `customFields`              | `[CustomField!]`               | Custom fields scoped by the role.                                         |
| `todoLists`                 | `[ProjectUserRoleTodoList!]`   | Per-list access (`viewable`, `editable`, `deletable`).                    |
| `todoFields`                | `[ProjectUserRoleTodoField!]!` | Per built-in field access (`fieldType`, `viewable`, `editable`).          |

## Full example

Create a contractor role: enable the records and files sections, allow completing records but not deleting them, limit visibility to assigned records, and grant view-only access to one custom field.

```graphql
mutation CreateContractorRole {
  createProjectUserRole(
    input: {
      projectId: "project_123"
      name: "External Contractor"
      description: "Limited access for outside collaborators"
      isRecordsEnabled: true
      isFilesEnabled: true
      allowMarkRecordsAsDone: true
      canDeleteRecords: false
      canDeleteFiles: false
      showOnlyAssignedTodos: true
      customFields: [{ customFieldId: "field_123", viewable: true, editable: false }]
    }
  ) {
    id
    name
    isRecordsEnabled
    canDeleteRecords
  }
}
```

Restrict a role to records tagged `tag_123` only:

```graphql
mutation CreateScopedRole {
  createProjectUserRole(
    input: {
      projectId: "project_123"
      name: "Client Reviewer"
      isRecordsEnabled: true
      recordTagFilter: { enabled: true, allowedTagIds: ["tag_123"] }
    }
  ) {
    id
    name
  }
}
```

Rename an existing role and turn off chat without touching its other settings:

```graphql
mutation UpdateRole {
  updateProjectUserRole(
    input: {
      roleId: "role_123"
      projectId: "project_123"
      name: "Contractor (Phase 2)"
      isChatEnabled: false
    }
  ) {
    id
    name
    isChatEnabled
  }
}
```

Delete a role:

```graphql
mutation DeleteRole {
  deleteProjectUserRole(input: { roleId: "role_123", projectId: "project_123" })
}
```

## Assigning a role

A role is attached to a member at invite time. Set `accessLevel: MEMBER` and pass the role's `roleId` to [`inviteUser`](/api/user-management/invite-user). The role's `projectId` must match the invite's workspace, or the call fails with `PROJECT_USER_ROLE_NOT_FOUND`. Custom roles are treated as `MEMBER`-level for the access hierarchy.

```graphql
mutation InviteWithRole {
  inviteUser(
    input: {
      email: "contractor@example.com"
      projectId: "project_123"
      accessLevel: MEMBER
      roleId: "role_123"
    }
  )
}
```

## Errors

| Code                          | When                                                                                                                                              |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| `FORBIDDEN`                   | The caller is not a workspace `OWNER` or `ADMIN`. Mutations require manage-roles access.                                                          |
| `PROJECT_NOT_FOUND`           | The `projectId` does not match a workspace.                                                                                                       |
| `PROJECT_USER_ROLE_NOT_FOUND` | The `roleId` does not exist, or belongs to a different workspace than `projectId`.                                                                |
| `PLAN_LIMIT_REACHED`          | The workspace has reached its custom-role limit. Message: `Custom roles per workspace limit reached (current/limit). Upgrade your plan for more.` |

The custom-role limit is plan-dependent — roughly 1 on starter plans up to ~50 on enterprise, and unlimited on some plans. Upgrade the organization's plan to raise it.

```json
{
  "errors": [
    {
      "message": "Custom roles per workspace limit reached (1/1). Upgrade your plan for more.",
      "extensions": {
        "code": "PLAN_LIMIT_REACHED",
        "kind": "over_limit",
        "resource": "Custom roles per workspace",
        "current": 1,
        "limit": 1
      }
    }
  ]
}
```

## Permissions

| Operation               | Required access                                                                                       |
| ----------------------- | ----------------------------------------------------------------------------------------------------- |
| `projectUserRoles`      | Any member of the queried workspace (roles for workspaces you don't belong to are silently excluded). |
| `createProjectUserRole` | Workspace `OWNER` or `ADMIN`.                                                                         |
| `updateProjectUserRole` | Workspace `OWNER` or `ADMIN`.                                                                         |
| `deleteProjectUserRole` | Workspace `OWNER` or `ADMIN`.                                                                         |

## Related

- [Invite a user](/api/user-management/invite-user) — assign a role via `roleId`.
- [List users](/api/user-management/list-users) — see who has access and their roles.
- [Remove a user](/api/user-management/remove-user) — revoke workspace or organization access.
