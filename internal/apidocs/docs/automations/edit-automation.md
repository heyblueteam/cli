---
title: Edit Automation
description: Update an existing automation's trigger, actions, or active state with the editAutomation mutation.
icon: Pencil
order: 2
---

Use the `editAutomation` mutation to update an existing automation's trigger, actions, or active state. Every field except `automationId` is optional, so you can flip an automation on or off, swap its actions, or reconfigure its trigger without resending the parts you aren't changing. Automations are `Automation` objects in the API.

`trigger` and `actions` reuse the same input types as [Create Automation](/api/automations/create-automation) — `CreateAutomationTriggerInput` and `CreateAutomationActionInput`. See that page for the full field reference. The semantics on edit differ in one important way: each block, when supplied, **fully replaces** the existing configuration (see [How replacement works](#how-replacement-works)).

## Request

The smallest call deactivates an automation by id. Pass the workspace context in the `blue-workspace-id` header (ID or slug).

```graphql
mutation EditAutomation {
  editAutomation(input: { automationId: "automation_123", isActive: false }) {
    id
    isActive
    updatedAt
  }
}
```

## Parameters

### EditAutomationInput

| Parameter      | Type                             | Required | Description                                                                                                                                                  |
| -------------- | -------------------------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `automationId` | `String!`                        | Yes      | The id of the automation to update.                                                                                                                          |
| `trigger`      | `CreateAutomationTriggerInput`   | No       | New trigger configuration. Fully replaces the existing trigger configuration (the trigger row keeps its id; its field values are rewritten from this input). |
| `actions`      | `[CreateAutomationActionInput!]` | No       | New ordered list of actions. Replaces all existing actions — there is no way to edit a single action in place.                                               |
| `isActive`     | `Boolean`                        | No       | Set the automation's active state. Deactivating pauses scheduled and due-date triggers; reactivating resumes them.                                           |

For the full `CreateAutomationTriggerInput` and `CreateAutomationActionInput` field tables (including the `schedule` block, trigger types, and action types), see [Create Automation](/api/automations/create-automation).

## Response

```json
{
  "data": {
    "editAutomation": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "isActive": false,
      "updatedAt": "2026-05-29T14:22:07.000Z"
    }
  }
}
```

### Returns

The mutation returns the updated `Automation`.

| Field       | Type                   | Description                                     |
| ----------- | ---------------------- | ----------------------------------------------- |
| `id`        | `ID!`                  | Unique identifier for the automation.           |
| `trigger`   | `AutomationTrigger!`   | The trigger configuration after the update.     |
| `actions`   | `[AutomationAction!]!` | The actions after the update, in order.         |
| `isActive`  | `Boolean!`             | Whether the automation is currently active.     |
| `createdBy` | `User!`                | The user who originally created the automation. |
| `project`   | `Workspace!`           | The workspace the automation belongs to.        |
| `createdAt` | `DateTime!`            | When the automation was created.                |
| `updatedAt` | `DateTime!`            | When the automation was last updated.           |

## Full example

Reconfigure a `TODO_LIST_CHANGED` trigger and replace the action set. Because `trigger` and `actions` each replace the existing configuration wholesale, send the complete desired state for both.

```graphql
mutation EditAutomationAdvanced {
  editAutomation(
    input: {
      automationId: "automation_123"
      isActive: true
      trigger: { type: TODO_LIST_CHANGED, todoListId: "list_123", tagIds: ["tag_123"] }
      actions: [
        { type: ADD_ASSIGNEE, assigneeIds: ["user_123"] }
        { type: CHANGE_DUE_DATE, duedIn: 14 }
        {
          type: SEND_EMAIL
          metadata: {
            email: {
              to: ["team@example.com"]
              from: "noreply@example.com"
              subject: "Updated automation triggered"
              content: "<p>An automation has been triggered.</p>"
            }
          }
        }
      ]
    }
  ) {
    id
    isActive
    trigger {
      id
      type
      todoList {
        id
        title
      }
      tags {
        id
        title
      }
    }
    actions {
      id
      type
      duedIn
      assignees {
        id
        fullName
      }
      metadata {
        ... on AutomationActionMetadataSendEmail {
          email {
            subject
            to
          }
        }
      }
    }
    createdBy {
      id
      fullName
    }
    updatedAt
  }
}
```

## How replacement works

Supplying `trigger` or `actions` replaces — never merges — the existing configuration.

- **Trigger.** The trigger row keeps its id, but its field values are rewritten from the `CreateAutomationTriggerInput` you send. Any scalar you omit reverts to that input's default rather than retaining its prior value. To change one setting, re-send the full desired trigger config (including its `type`).
- **Schedule.** `SCHEDULED` triggers carry a nested `schedule` block. Because the trigger input fully replaces, you must include the complete `schedule` again when editing a scheduled automation — otherwise the schedule is dropped. Changing a trigger to or from `SCHEDULED` automatically registers or removes the underlying schedule job.
- **Actions.** Providing `actions` deletes the existing actions and creates the new list. To keep an existing action, include it again in the array.
- **Due-date triggers.** Edits to a `DUE_DATE_EXPIRED` trigger (and toggling `isActive` on one) automatically update its monitoring job.

<Callout variant="warning" title="Partial trigger edits drop config">

There is no partial trigger update. If you send `trigger` to change a single field, omitted scalars (and a `schedule` block you forgot to include) are cleared. Always send the complete trigger configuration you want the automation to end with.

</Callout>

## Errors

| Code                   | Message                    | When                                                                                                                                                        |
| ---------------------- | -------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `UNAUTHENTICATED`      | You are not authenticated. | No valid authentication token was provided.                                                                                                                 |
| `FORBIDDEN`            | You are not authorized.    | The authenticated user is not an `OWNER` or `ADMIN` of the workspace, the workspace is archived, or the automation belongs to a workspace you can't access. |
| `AUTOMATION_NOT_FOUND` | Automation was not found.  | No automation exists with the given `automationId`.                                                                                                         |

An automation in a workspace you have no access to is looked up successfully but rejected for lack of permission, so it returns `FORBIDDEN` — not `AUTOMATION_NOT_FOUND`. `AUTOMATION_NOT_FOUND` fires only when the id truly does not exist.

## Permissions

Only an `OWNER` or `ADMIN` of an active (non-archived) workspace can edit automations.

| Access level | Can edit automations |
| ------------ | -------------------- |
| `OWNER`      | Yes                  |
| `ADMIN`      | Yes                  |
| `MEMBER`     | No                   |
| `CLIENT`     | No                   |

## Related

- [Create Automation](/api/automations/create-automation)
- [Copy Automation](/api/automations/copy-automation)
- [Delete Automation](/api/automations/delete-automation)
