---
title: Automations
description: Automate work in a workspace — list automations and reach their create, edit, copy, and delete operations.
icon: Zap
order: 0
---

Automations run actions automatically when something happens in a workspace — for example, moving a record to a list when it's marked complete, or assigning the person who triggered the change. Each automation is **project-scoped** and pairs one **trigger** (the event to watch) with one or more **actions** (what to do).

Workspaces are `Project` objects in the API; records are `Todo` objects. Every automation belongs to a single project, identified by the `X-Bloo-Project-ID` header or the `projectId` filter argument.

## Operations

| Operation                                                  | GraphQL                     | Description                                              |
| ---------------------------------------------------------- | --------------------------- | -------------------------------------------------------- |
| List automations                                           | `automationList` query      | Documented on this page                                  |
| [Create an automation](/api/automations/create-automation) | `createAutomation` mutation | Define a trigger and its actions                         |
| [Edit an automation](/api/automations/edit-automation)     | `editAutomation` mutation   | Change the trigger, actions, or active state             |
| [Copy an automation](/api/automations/copy-automation)     | `copyAutomation` mutation   | Duplicate an automation, optionally into another project |
| [Delete an automation](/api/automations/delete-automation) | `deleteAutomation` mutation | Remove an automation                                     |

## Request

Use the `automationList` query to retrieve the automations for a project. With no arguments it returns the current project's automations (from the `X-Bloo-Project-ID` header), newest first.

```graphql
query ListAutomations {
  automationList {
    items {
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
    pageInfo {
      totalItems
      hasNextPage
    }
    totalCount
  }
}
```

## Parameters

`automationList` takes pagination arguments plus an optional `filter` input.

| Argument | Type                    | Required | Description                              |
| -------- | ----------------------- | -------- | ---------------------------------------- |
| `filter` | `AutomationFilterInput` | No       | Narrows the result set. See below.       |
| `skip`   | `Int`                   | No       | Number of items to skip. Default `0`.    |
| `take`   | `Int`                   | No       | Number of items to return. Default `20`. |

### AutomationFilterInput

| Field            | Type       | Required | Description                                                                                                                      |
| ---------------- | ---------- | -------- | -------------------------------------------------------------------------------------------------------------------------------- |
| `projectId`      | `String`   | No       | List automations from a specific project by id or slug instead of the header context. Access is re-validated against the caller. |
| `customFieldIds` | `[String]` | No       | Only return automations whose trigger references one of these custom field ids.                                                  |
| `customFieldId`  | `String`   | No       | **Deprecated.** Single-field version of `customFieldIds`. Use `customFieldIds` instead.                                          |

<Callout variant="warning" title="skip and take are ignored when filtering by custom field">

When `customFieldIds` (or the deprecated `customFieldId`) is supplied, the query returns every matching automation and ignores `skip`/`take`. Pagination only applies to the unfiltered listing.

</Callout>

## Response

```json
{
  "data": {
    "automationList": {
      "items": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "isActive": true,
          "trigger": { "type": "TODO_MARKED_AS_COMPLETE" },
          "actions": [{ "type": "CHANGE_TODO_LIST" }],
          "createdAt": "2026-05-20T14:32:10.000Z"
        }
      ],
      "pageInfo": {
        "totalItems": 1,
        "hasNextPage": false
      },
      "totalCount": 1
    }
  }
}
```

### AutomationPagination

| Field        | Type             | Description                                                                                            |
| ------------ | ---------------- | ------------------------------------------------------------------------------------------------------ |
| `items`      | `[Automation!]!` | The automations for this page, newest first.                                                           |
| `pageInfo`   | `PageInfo!`      | Pagination metadata (`totalItems`, `totalPages`, `page`, `perPage`, `hasNextPage`, `hasPreviousPage`). |
| `totalCount` | `Int!`           | Total automations matching the filter, ignoring pagination.                                            |

### Automation

| Field       | Type                   | Description                                            |
| ----------- | ---------------------- | ------------------------------------------------------ |
| `id`        | `ID!`                  | Unique identifier.                                     |
| `trigger`   | `AutomationTrigger!`   | The event that starts this automation.                 |
| `actions`   | `[AutomationAction!]!` | Ordered list of actions to run when the trigger fires. |
| `isActive`  | `Boolean!`             | Whether the automation is currently enabled.           |
| `createdBy` | `User!`                | The user who created the automation.                   |
| `project`   | `Project!`             | The project this automation belongs to.                |
| `createdAt` | `DateTime!`            | When the automation was created.                       |
| `updatedAt` | `DateTime!`            | When the automation was last changed.                  |

### AutomationTrigger

| Field                          | Type                          | Description                                                                                             |
| ------------------------------ | ----------------------------- | ------------------------------------------------------------------------------------------------------- |
| `id`                           | `ID!`                         | Unique identifier.                                                                                      |
| `type`                         | `AutomationTriggerType!`      | The kind of event to watch. See [Trigger types](#trigger-types).                                        |
| `metadata`                     | `AutomationTriggerMetadata`   | Trigger-specific config. Currently `AutomationTriggerMetadataTodoOverdue` (`incompleteOnly`, `offset`). |
| `customField`                  | `CustomField`                 | The custom field this trigger watches, when applicable.                                                 |
| `customFieldOptions`           | `[CustomFieldOption!]`        | The select options the trigger reacts to.                                                               |
| `todos(first: Int, skip: Int)` | `[CustomFieldReferenceTodo!]` | Referenced records, with optional pagination args.                                                      |
| `todoList`                     | `TodoList`                    | A single related list.                                                                                  |
| `todoLists`                    | `[TodoList!]!`                | Related lists for the trigger.                                                                          |
| `tags`                         | `[Tag!]`                      | Related tags.                                                                                           |
| `assignees`                    | `[User!]`                     | Related assignees.                                                                                      |
| `color`                        | `String`                      | A single related color (hex).                                                                           |
| `colors`                       | `[String!]`                   | Related colors (hex).                                                                                   |
| `dueStart`                     | `DateTime`                    | Start of the due-date window the trigger matches.                                                       |
| `dueEnd`                       | `DateTime`                    | End of the due-date window the trigger matches.                                                         |
| `showCompleted`                | `Boolean`                     | Whether completed records are included.                                                                 |
| `fields`                       | `JSON`                        | Raw field-condition config for filter-based triggers.                                                   |
| `op`                           | `FilterLogicalOperator`       | How filter conditions combine (`AND` / `OR`).                                                           |
| `conditionMode`                | `ConditionalAutomationMode`   | For `CONDITIONAL` triggers: `STARTS_MATCHING` or `STOPS_MATCHING`.                                      |
| `filterGroups`                 | `JSON`                        | Raw grouped filter conditions.                                                                          |
| `filterGroupLinks`             | `JSON`                        | Raw links between filter groups.                                                                        |
| `schedule`                     | `AutomationTriggerSchedule`   | Schedule config for `SCHEDULED` triggers. See [AutomationTriggerSchedule](#automationtriggerschedule).  |
| `createdAt`                    | `DateTime!`                   | When the trigger was created.                                                                           |
| `updatedAt`                    | `DateTime!`                   | When the trigger was last changed.                                                                      |

### AutomationAction

| Field                | Type                         | Description                                                                                                        |
| -------------------- | ---------------------------- | ------------------------------------------------------------------------------------------------------------------ |
| `id`                 | `ID!`                        | Unique identifier.                                                                                                 |
| `type`               | `AutomationActionType!`      | The action to perform. See [Action types](#action-types).                                                          |
| `duedIn`             | `Int`                        | Days from execution to set as the due date (for due-date actions).                                                 |
| `customField`        | `CustomField`                | The target custom field.                                                                                           |
| `customFieldOptions` | `[CustomFieldOption!]`       | Select options to set.                                                                                             |
| `todoList`           | `TodoList`                   | The destination list (for `CHANGE_TODO_LIST`).                                                                     |
| `metadata`           | `AutomationActionMetadata`   | Action-specific config (email, checklist, copy, or custom-field value).                                            |
| `tags`               | `[Tag!]`                     | Tags to add or remove.                                                                                             |
| `assignees`          | `[User!]`                    | Assignees to add or remove.                                                                                        |
| `projects`           | `[Project!]`                 | Target projects (for `ADD_PROJECT`).                                                                               |
| `color`              | `String`                     | Color to add or remove (hex).                                                                                      |
| `assigneeTriggerer`  | `String`                     | Identifier for the user who triggered the automation (for `ADD_ASSIGNEE_TRIGGERER` / `REMOVE_ASSIGNEE_TRIGGERER`). |
| `portableDocument`   | `PortableDocument`           | The document template used by `GENERATE_PDF`.                                                                      |
| `httpOption`         | `AutomationActionHttpOption` | Request config for `MAKE_HTTP_REQUEST`.                                                                            |
| `createdAt`          | `DateTime!`                  | When the action was created.                                                                                       |
| `updatedAt`          | `DateTime!`                  | When the action was last changed.                                                                                  |

### AutomationTriggerSchedule

Returned on `AutomationTrigger.schedule` for `SCHEDULED` triggers.

| Field             | Type                      | Description                                                      |
| ----------------- | ------------------------- | ---------------------------------------------------------------- |
| `id`              | `ID!`                     | Unique identifier.                                               |
| `type`            | `AutomationScheduleType!` | `DAILY`, `WEEKLY`, `WEEKDAYS`, `MONTHLY`, `YEARLY`, or `CUSTOM`. |
| `hour`            | `Int!`                    | Hour of day (0–23) the schedule fires.                           |
| `minute`          | `Int!`                    | Minute of the hour the schedule fires.                           |
| `timezone`        | `String!`                 | IANA timezone the schedule is evaluated in.                      |
| `days`            | `[Int!]`                  | Specific days the schedule runs.                                 |
| `dayOfWeek`       | `[Int!]`                  | Days of the week for `WEEKLY` schedules.                         |
| `dayOfMonth`      | `Int`                     | Day of the month for `MONTHLY` schedules.                        |
| `month`           | `Int`                     | Month for `YEARLY` schedules.                                    |
| `cronExpression`  | `String`                  | Cron expression for `CUSTOM` schedules.                          |
| `lastExecutedAt`  | `DateTime`                | When the schedule last ran.                                      |
| `nextExecutionAt` | `DateTime`                | When the schedule will next run.                                 |
| `isActive`        | `Boolean!`                | Whether the schedule is enabled.                                 |
| `createdAt`       | `DateTime!`               | When the schedule was created.                                   |
| `updatedAt`       | `DateTime!`               | When the schedule was last changed.                              |

## Trigger types

`AutomationTriggerType` enumerates the events that can start an automation.

| Value                                     | Fires when                                                        |
| ----------------------------------------- | ----------------------------------------------------------------- |
| `TODO_CREATED`                            | A record is created.                                              |
| `TODO_LIST_CHANGED`                       | A record moves between lists.                                     |
| `TODO_MARKED_AS_COMPLETE`                 | A record is marked complete.                                      |
| `TODO_MARKED_AS_INCOMPLETE`               | A record is marked incomplete.                                    |
| `ASSIGNEE_ADDED`                          | An assignee is added to a record.                                 |
| `ASSIGNEE_REMOVED`                        | An assignee is removed from a record.                             |
| `DUE_DATE_CHANGED`                        | A record's due date is set or changed.                            |
| `DUE_DATE_REMOVED`                        | A record's due date is cleared.                                   |
| `DUE_DATE_EXPIRED`                        | A record's due date passes.                                       |
| `TAG_ADDED`                               | A tag is added to a record.                                       |
| `TAG_REMOVED`                             | A tag is removed from a record.                                   |
| `CHECKLIST_ITEM_MARKED_AS_DONE`           | A checklist item is completed.                                    |
| `CHECKLIST_ITEM_MARKED_AS_UNDONE`         | A checklist item is uncompleted.                                  |
| `TODO_COPIED_OR_MOVED_FROM_OTHER_PROJECT` | A record is copied or moved in from another project.              |
| `CUSTOM_FIELD_ADDED`                      | A custom field value is set on a record.                          |
| `CUSTOM_FIELD_REMOVED`                    | A custom field value is cleared on a record.                      |
| `CUSTOM_FIELD_BUTTON_CLICKED`             | A button custom field is clicked.                                 |
| `COLOR_ADDED`                             | A color is applied to a record.                                   |
| `COLOR_REMOVED`                           | A color is removed from a record.                                 |
| `PROJECT_ADDED_TO_TODO`                   | A project is added to a record.                                   |
| `SCHEDULED`                               | A time-based schedule fires. Requires `schedule` config.          |
| `CONDITIONAL`                             | A record starts or stops matching a filter (see `conditionMode`). |

## Action types

`AutomationActionType` enumerates what an automation can do when its trigger fires.

| Value                           | Effect                                        |
| ------------------------------- | --------------------------------------------- |
| `CHANGE_TODO_LIST`              | Move the record to a different list.          |
| `MARK_AS_COMPLETE`              | Mark the record complete.                     |
| `MARK_AS_INCOMPLETE`            | Mark the record incomplete.                   |
| `ADD_ASSIGNEE`                  | Add one or more assignees.                    |
| `REMOVE_ASSIGNEE`               | Remove one or more assignees.                 |
| `ADD_ASSIGNEE_TRIGGERER`        | Add the user who triggered the automation.    |
| `REMOVE_ASSIGNEE_TRIGGERER`     | Remove the user who triggered the automation. |
| `CHANGE_DUE_DATE`               | Set or update the due date.                   |
| `REMOVE_DUE_DATE`               | Clear the due date.                           |
| `ADD_TAG`                       | Add one or more tags.                         |
| `REMOVE_TAG`                    | Remove one or more tags.                      |
| `ADD_COLOR`                     | Apply a color.                                |
| `REMOVE_COLOR`                  | Remove a color.                               |
| `ADD_CUSTOM_FIELD`              | Set a custom field value.                     |
| `REMOVE_CUSTOM_FIELD`           | Clear a custom field value.                   |
| `ADD_PROJECT`                   | Add the record to another project.            |
| `CREATE_CHECKLIST`              | Create a checklist on the record.             |
| `MARK_CHECKLIST_ITEM_AS_DONE`   | Complete checklist items.                     |
| `MARK_CHECKLIST_ITEM_AS_UNDONE` | Uncomplete checklist items.                   |
| `COPY_TODO`                     | Duplicate the record.                         |
| `SEND_EMAIL`                    | Send an email.                                |
| `GENERATE_PDF`                  | Generate a PDF from a document template.      |
| `MAKE_HTTP_REQUEST`             | Call an external API.                         |

## Full example

This selects the full trigger and action detail, including the metadata unions, the schedule, and the creator.

```graphql
query ListAutomationsDetailed {
  automationList(filter: { projectId: "project_123" }, skip: 0, take: 50) {
    items {
      id
      isActive
      createdAt
      updatedAt
      createdBy {
        id
        fullName
        email
      }
      trigger {
        id
        type
        color
        op
        conditionMode
        customField {
          id
          name
          type
        }
        customFieldOptions {
          id
          title
          color
        }
        todoLists {
          id
          title
        }
        tags {
          id
          title
          color
        }
        assignees {
          id
          fullName
          email
        }
        schedule {
          id
          type
          hour
          minute
          timezone
          nextExecutionAt
          isActive
        }
        metadata {
          ... on AutomationTriggerMetadataTodoOverdue {
            incompleteOnly
            offset
          }
        }
      }
      actions {
        id
        type
        color
        duedIn
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
        projects {
          id
          name
        }
        portableDocument {
          id
          name
        }
        metadata {
          ... on AutomationActionMetadataSendEmail {
            email {
              subject
              to
              from
              content
            }
          }
          ... on AutomationActionMetadataCreateChecklist {
            checklists {
              title
              checklistItems {
                title
                duedIn
              }
            }
          }
        }
      }
    }
    pageInfo {
      totalItems
      totalPages
      hasNextPage
      hasPreviousPage
    }
    totalCount
  }
}
```

## Errors

| Code                | When                                                                                                         |
| ------------------- | ------------------------------------------------------------------------------------------------------------ |
| `UNAUTHENTICATED`   | The request has no valid credentials. Message: `You are not authenticated.`                                  |
| `PROJECT_NOT_FOUND` | A `projectId` filter was supplied but no project matches that id or slug. Message: `Project was not found.`  |
| `FORBIDDEN`         | A `projectId` filter resolved to a project the caller is not a member of. Message: `You are not authorized.` |

```json
{
  "errors": [
    {
      "message": "Project was not found.",
      "extensions": { "code": "PROJECT_NOT_FOUND" }
    }
  ]
}
```

## Permissions

Any member of the project can list its automations, regardless of access level — `OWNER`, `ADMIN`, `MEMBER`, `CLIENT`, `COMMENT_ONLY`, and `VIEW_ONLY` all return the full list. When you pass a `projectId` filter, the caller must be a member of that project or the query fails with `FORBIDDEN`.

## Related

- [Create an automation](/api/automations/create-automation)
- [Edit an automation](/api/automations/edit-automation)
- [Copy an automation](/api/automations/copy-automation)
- [Delete an automation](/api/automations/delete-automation)
- [Custom fields](/api/custom-fields)
- [Workspaces](/api/workspaces)
