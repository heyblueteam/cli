---
title: Copy Automation
description: Duplicate an existing automation, including its trigger and all actions, into your current workspace.
icon: Copy
order: 3
---

Use the `copyAutomation` mutation to duplicate an existing automation. The copy reproduces the source automation's trigger configuration and every action, then attaches it to the workspace in your current request context. Copies are **always created inactive** — activate them with [`editAutomation`](/api/automations/edit-automation) once you're ready for them to fire.

Automations are `Automation` objects in the API. Workspaces are `Project` objects.

<Callout variant="info" title="The copy lands in your current workspace">

`copyAutomation` looks up the source automation by ID without scoping to a workspace, but it always creates the copy in the workspace named by your `X-Bloo-Project-ID` header. If the source belongs to a different workspace, the copy is still created in your current one — so this is a workspace-to-workspace copy whenever the IDs differ.

</Callout>

## Request

The only field you need is the ID of the automation to copy.

```graphql
mutation CopyAutomation {
  copyAutomation(input: { automationId: "automation_123" }) {
    id
    isActive
    trigger {
      type
    }
    actions {
      type
    }
    createdAt
  }
}
```

## Parameters

### CopyAutomationInput

| Parameter      | Type      | Required | Description                                                                                                                                                                    |
| -------------- | --------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `automationId` | `String!` | Yes      | ID of the automation to copy. Looked up globally, not scoped to a workspace.                                                                                                   |
| `isActive`     | `Boolean` | No       | **Currently ignored.** The resolver always creates the copy with `isActive: false`. To activate the copy, call [`editAutomation`](/api/automations/edit-automation) afterward. |

<Callout variant="warning" title="isActive is a no-op">

`CopyAutomationInput` accepts `isActive`, but the mutation hard-codes the copy to `isActive: false` and never reads the value you pass. Sending `isActive: true` does **not** produce an active copy. Plan to activate via `editAutomation`.

</Callout>

## Response

The mutation returns the newly created `Automation`. Note `isActive` is `false` regardless of the input.

```json
{
  "data": {
    "copyAutomation": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "isActive": false,
      "trigger": {
        "type": "TODO_CREATED"
      },
      "actions": [
        {
          "type": "ADD_TAG"
        }
      ],
      "createdAt": "2026-05-29T14:21:08.000Z"
    }
  }
}
```

### Returns

The `Automation` object:

| Field       | Type                   | Description                                                       |
| ----------- | ---------------------- | ----------------------------------------------------------------- |
| `id`        | `ID!`                  | Unique identifier of the new copy.                                |
| `trigger`   | `AutomationTrigger!`   | The duplicated trigger configuration.                             |
| `actions`   | `[AutomationAction!]!` | The duplicated actions, in order.                                 |
| `isActive`  | `Boolean!`             | Always `false` for a fresh copy.                                  |
| `createdBy` | `User!`                | The user who ran the copy — not the original automation's author. |
| `project`   | `Project!`             | The workspace the copy was created in (your current context).     |
| `createdAt` | `DateTime!`            | When the copy was created.                                        |
| `updatedAt` | `DateTime!`            | When the copy was last updated.                                   |

## Full example

A deeper selection that reads back the duplicated trigger, every action field, and all three commonly populated `AutomationActionMetadata` members via inline fragments.

```graphql
mutation CopyAutomationDetailed {
  copyAutomation(input: { automationId: "automation_123" }) {
    id
    isActive
    trigger {
      id
      type
      color
      conditionMode
      filterGroups
      filterGroupLinks
      todoList {
        id
        title
      }
      tags {
        id
        title
      }
      assignees {
        id
        fullName
      }
      customField {
        id
        name
        type
      }
      customFieldOptions {
        id
        title
      }
    }
    actions {
      id
      type
      duedIn
      color
      assigneeTriggerer
      todoList {
        id
        title
      }
      tags {
        id
        title
      }
      assignees {
        id
        fullName
      }
      customField {
        id
        name
      }
      customFieldOptions {
        id
        title
      }
      projects {
        id
        name
      }
      portableDocument {
        id
      }
      httpOption {
        url
        method
        contentType
        authorizationType
      }
      metadata {
        ... on AutomationActionMetadataSendEmail {
          email {
            subject
            to
            from
            cc
            bcc
            content
          }
        }
        ... on AutomationActionMetadataCreateChecklist {
          checklists {
            title
            position
            checklistItems {
              title
              position
              duedIn
            }
          }
        }
        ... on AutomationActionMetadataCopyTodo {
          copyTodoOptions
        }
      }
    }
    createdBy {
      id
      fullName
      email
    }
    createdAt
    updatedAt
  }
}
```

## What gets copied

- **Trigger** — type, color, condition mode, filter groups, target list, tags, assignees, custom field, and custom-field options.
- **Actions** — type, due offset (`duedIn`), color, target list, tags, assignees, target workspaces, linked PDF, HTTP request configuration, and all action metadata (email content, checklist templates, copy-record options).
- **Assignee fallback** — for `ASSIGNEE_ADDED` and `ASSIGNEE_REMOVED` triggers, if the source has no assignees configured, the user running the copy is added as the default assignee.

What does **not** carry over:

- **Active state** — the copy is always inactive (see above).
- **Schedule** — `SCHEDULED` triggers are not rescheduled on copy. Activating the copy via `editAutomation` starts its schedule.

## Errors

| Code              | When                                                                                                                                                                                                                                                                                                                                                                       |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `UNAUTHENTICATED` | No valid credentials were provided. Message: `You are not authenticated.`                                                                                                                                                                                                                                                                                                  |
| `FORBIDDEN`       | The caller isn't an OWNER or ADMIN of the current workspace, or the workspace is archived. Message: `You are not authorized.`                                                                                                                                                                                                                                              |
| _(none)_          | The `automationId` doesn't exist, or the automation has no trigger. The resolver throws a plain `Error("Automation or its trigger not found")` with **no** `extensions.code`, so it surfaces as a generic server error rather than a coded one. The permission check runs first, so an unknown ID from an unauthorized context returns `FORBIDDEN` before this is reached. |

```json
{
  "errors": [
    {
      "message": "Automation or its trigger not found"
    }
  ],
  "data": null
}
```

## Permissions

Only OWNER or ADMIN members of an **active** (non-archived) workspace can copy automations. The check runs against your current workspace context (`X-Bloo-Project-ID`), not the source automation's workspace.

| Access level | Can copy automations |
| ------------ | -------------------- |
| `OWNER`      | Yes                  |
| `ADMIN`      | Yes                  |
| `MEMBER`     | No                   |
| `CLIENT`     | No                   |

## Related

- [Create an automation](/api/automations/create-automation)
- [Edit an automation](/api/automations/edit-automation)
- [Delete an automation](/api/automations/delete-automation)
- [Automations overview](/api/automations)
