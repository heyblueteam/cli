---
title: Automation Groups
description: Organize a workspace's automations into flat groups, and move an automation between them.
icon: Layers
order: 5
---

An automation group is a flat, workspace-scoped section that organizes automations. Groups never change when or how an automation runs — they only affect how the automations are listed.

Groups have one level: a group cannot contain another group. Each group belongs to a single workspace (a `Workspace` in the API), identified by the `blue-workspace-id` header or by an explicit `projectId`.

An automation carries its group in the `group` field and its order inside that group in `position`. An automation with no group is ungrouped.

## Operations

| Operation      | GraphQL                          | Description                                        |
| -------------- | -------------------------------- | -------------------------------------------------- |
| List groups    | `automationGroups` query         | Groups of a workspace, ordered by position         |
| Create a group | `createAutomationGroup` mutation | Add a group to a workspace                         |
| Edit a group   | `editAutomationGroup` mutation   | Rename a group or change its position              |
| Delete a group | `deleteAutomationGroup` mutation | Remove a group; its automations become ungrouped   |
| Move           | `moveAutomation` mutation        | Change an automation's group and order in one call |

## List groups

`automationGroups` returns the groups of a workspace, ordered by `position` ascending, then by `createdAt`.

```graphql
query ListAutomationGroups {
  automationGroups {
    id
    uid
    name
    position
  }
}
```

| Argument    | Type     | Required | Description                                                                        |
| ----------- | -------- | -------- | ---------------------------------------------------------------------------------- |
| `projectId` | `String` | No       | Workspace ID or slug. Defaults to the workspace in the `blue-workspace-id` header. |

Any member of the workspace can list its groups. An explicit `projectId` requires membership of that workspace.

### AutomationGroup

| Field       | Type        | Description                                     |
| ----------- | ----------- | ----------------------------------------------- |
| `id`        | `ID!`       | Unique identifier.                              |
| `uid`       | `String!`   | Short identifier for the group.                 |
| `name`      | `String!`   | The group name.                                 |
| `position`  | `Float!`    | Sort order in the workspace. Lower sorts first. |
| `createdAt` | `DateTime!` | When the group was created.                     |
| `updatedAt` | `DateTime!` | When the group was last changed.                |

## Create a group

```graphql
mutation CreateAutomationGroup {
  createAutomationGroup(input: { name: "Onboarding" }) {
    id
    name
    position
  }
}
```

### CreateAutomationGroupInput

| Parameter   | Type      | Required | Description                                                                   |
| ----------- | --------- | -------- | ----------------------------------------------------------------------------- |
| `name`      | `String!` | Yes      | Group name. Trimmed, 120 characters or fewer. A blank name is rejected.       |
| `projectId` | `String`  | No       | Workspace to create the group in. Defaults to the `blue-workspace-id` header. |

The new group is appended after the workspace's existing groups. The mutation returns the created `AutomationGroup`.

## Edit a group

```graphql
mutation EditAutomationGroup {
  editAutomationGroup(input: { id: "group_123", name: "Client onboarding" }) {
    id
    name
    position
  }
}
```

### EditAutomationGroupInput

| Parameter  | Type      | Required | Description                                                         |
| ---------- | --------- | -------- | ------------------------------------------------------------------- |
| `id`       | `String!` | Yes      | The group to update.                                                |
| `name`     | `String`  | No       | New name. Omit to keep the current one. A blank name is rejected.   |
| `position` | `Float`   | No       | New sort order in the workspace. Omit to keep the current position. |

The mutation returns the updated `AutomationGroup`.

## Delete a group

```graphql
mutation DeleteAutomationGroup {
  deleteAutomationGroup(id: "group_123")
}
```

| Parameter | Type      | Required | Description                    |
| --------- | --------- | -------- | ------------------------------ |
| `id`      | `String!` | Yes      | The ID of the group to delete. |

Deleting a group does not delete its automations. Every automation in the group becomes ungrouped in the same transaction, and each one publishes an update event so connected clients re-section their list.

`deleteAutomationGroup` returns the scalar `Boolean!` — `true` when the group is deleted. It has no sub-fields, so don't add a selection set.

## Move an automation

`moveAutomation` changes an automation's group and order without touching its trigger or actions. Use it instead of `editAutomation` for a reorder or a drag-and-drop, so a move never revalidates or rewrites the automation's configuration.

```graphql
mutation MoveAutomation {
  moveAutomation(input: { automationId: "automation_123", groupId: "group_123" }) {
    id
    position
    group {
      id
      name
    }
  }
}
```

### MoveAutomationInput

| Parameter      | Type      | Required | Description                                                                                    |
| -------------- | --------- | -------- | ---------------------------------------------------------------------------------------------- |
| `automationId` | `String!` | Yes      | The automation to move.                                                                        |
| `groupId`      | `String`  | No       | Target group. Omit to keep the current group; send an explicit `null` to move it to Ungrouped. |
| `position`     | `Float`   | No       | Position inside the target group. Omit to append at the end.                                   |

The target group must belong to the same workspace as the automation. A group from another workspace is rejected rather than silently connected.

The mutation returns the updated `Automation`.

## Errors

| Code              | When                                                                                                            |
| ----------------- | --------------------------------------------------------------------------------------------------------------- |
| `UNAUTHENTICATED` | No valid credentials were provided.                                                                             |
| `FORBIDDEN`       | The caller isn't an `OWNER` or `ADMIN` of the workspace, has no access to it, or the workspace is archived.     |
| `BAD_USER_INPUT`  | The group doesn't exist, the name is blank or longer than 120 characters, or the group is in another workspace. |

## Permissions

| Access level   | List groups | Create, edit, delete, and move |
| -------------- | ----------- | ------------------------------ |
| `OWNER`        | Yes         | Yes                            |
| `ADMIN`        | Yes         | Yes                            |
| `MEMBER`       | Yes         | No                             |
| `CLIENT`       | Yes         | No                             |
| `COMMENT_ONLY` | Yes         | No                             |
| `VIEW_ONLY`    | Yes         | No                             |

The workspace must be active.

## Related

- [Automations overview](/api/automations)
- [Create an automation](/api/automations/create-automation)
- [Edit an automation](/api/automations/edit-automation)
- [Workspaces](/api/workspaces)
