---
title: Create Automation
description: Create an automation with a trigger and one or more actions that run when the trigger fires.
icon: Plus
order: 1
---

Use the `createAutomation` mutation to build an automation in a workspace: a single **trigger** (the event to watch for) plus one or more **actions** that run when the trigger fires. Automations are workspace-scoped (workspaces are `Workspace` objects in the API) and are created in an active state by default.

The automation is created in the workspace identified by the `blue-workspace-id` header, which accepts a workspace ID or slug.

## Request

The smallest automation needs a trigger `type` and at least one action. This one adds a tag to every record (`Record`) the moment it is created.

```graphql
mutation CreateAutomation {
  createAutomation(
    input: { trigger: { type: TODO_CREATED }, actions: [{ type: ADD_TAG, tagIds: ["tag_123"] }] }
  ) {
    id
    isActive
    trigger {
      id
      type
    }
    actions {
      id
      type
    }
    createdAt
  }
}
```

## Parameters

### CreateAutomationInput

| Parameter | Type                              | Required | Description                                                  |
| --------- | --------------------------------- | -------- | ------------------------------------------------------------ |
| `trigger` | `CreateAutomationTriggerInput!`   | Yes      | The event that starts the automation.                        |
| `actions` | `[CreateAutomationActionInput!]!` | Yes      | One or more actions to run when the trigger fires, in order. |

### CreateAutomationTriggerInput

The trigger fields that apply depend on `type`. Most fields are filters that narrow when the trigger fires; `schedule` is only for `SCHEDULED` triggers, and `conditionMode` is only for `CONDITIONAL` triggers.

| Parameter              | Type                                   | Required | Description                                                                                                                            |
| ---------------------- | -------------------------------------- | -------- | -------------------------------------------------------------------------------------------------------------------------------------- |
| `type`                 | `AutomationTriggerType!`               | Yes      | The event that triggers the automation. See [AutomationTriggerType](#automationtriggertype).                                           |
| `metadata`             | `AutomationTriggerMetadataInput`       | No       | Trigger-specific settings (e.g. due-date `offset` for `DUE_DATE_EXPIRED`).                                                             |
| `customFieldId`        | `String`                               | No       | Custom field to watch (for custom-field triggers).                                                                                     |
| `customFieldOptionIds` | `[String!]`                            | No       | Specific custom-field option IDs to match.                                                                                             |
| `todoListId`           | `String`                               | No       | Restrict the trigger to a single list (`RecordList`).                                                                                  |
| `todoListIds`          | `[String!]`                            | No       | Restrict the trigger to multiple lists.                                                                                                |
| `tagIds`               | `[String!]`                            | No       | Tag IDs that must match.                                                                                                               |
| `todoIds`              | `[String!]`                            | No       | Restrict the trigger to specific records.                                                                                              |
| `assigneeIds`          | `[String!]`                            | No       | Assignee (user) IDs that must match.                                                                                                   |
| `color`                | `String`                               | No       | Color to match (hex, e.g. `#FF5733`).                                                                                                  |
| `colors`               | `[String!]`                            | No       | Multiple colors to match.                                                                                                              |
| `dueStart`             | `DateTime`                             | No       | Start of a due-date range to match.                                                                                                    |
| `dueEnd`               | `DateTime`                             | No       | End of a due-date range to match.                                                                                                      |
| `showCompleted`        | `Boolean`                              | No       | Include completed records when matching.                                                                                               |
| `fields`               | `JSON`                                 | No       | Opaque, internally structured field-condition payload, mirrored from the workspace filter UI. Build it in the app rather than by hand. |
| `op`                   | `FilterLogicalOperator`                | No       | How to combine filter conditions: `AND` or `OR`.                                                                                       |
| `conditionMode`        | `ConditionalAutomationMode`            | No       | For `CONDITIONAL` triggers only: `STARTS_MATCHING` or `STOPS_MATCHING`.                                                                |
| `filterGroups`         | `JSON`                                 | No       | Opaque, internally structured filter-group payload for `CONDITIONAL` triggers. Build it in the app.                                    |
| `filterGroupLinks`     | `JSON`                                 | No       | Opaque, internally structured links between filter groups. Build it in the app.                                                        |
| `schedule`             | `CreateAutomationTriggerScheduleInput` | No       | Required for `SCHEDULED` triggers; ignored otherwise.                                                                                  |

### CreateAutomationTriggerScheduleInput

Configures a `SCHEDULED` trigger. `hour`, `minute`, and `timezone` are always required; the other fields are used by specific `type` values (e.g. `dayOfWeek` for `WEEKLY`, `dayOfMonth` for `MONTHLY`, `cronExpression` for `CUSTOM`).

| Parameter        | Type                      | Required | Description                                                                |
| ---------------- | ------------------------- | -------- | -------------------------------------------------------------------------- |
| `type`           | `AutomationScheduleType!` | Yes      | Schedule frequency. See [AutomationScheduleType](#automationscheduletype). |
| `hour`           | `Int!`                    | Yes      | Hour of day to run, `0`–`23`.                                              |
| `minute`         | `Int!`                    | Yes      | Minute of hour to run, `0`–`59`.                                           |
| `timezone`       | `String!`                 | Yes      | IANA timezone, e.g. `"America/New_York"`.                                  |
| `days`           | `[Int!]`                  | No       | Specific days for the schedule.                                            |
| `dayOfWeek`      | `[Int!]`                  | No       | Days of week, `0`=Sunday through `6`=Saturday. Used by `WEEKLY`.           |
| `dayOfMonth`     | `Int`                     | No       | Day of month, `1`–`31`. Used by `MONTHLY`/`YEARLY`.                        |
| `month`          | `Int`                     | No       | Month of year, `1`–`12`. Used by `YEARLY`.                                 |
| `cronExpression` | `String`                  | No       | Custom cron expression. Used by `CUSTOM`.                                  |
| `isActive`       | `Boolean`                 | No       | Whether the schedule is active.                                            |

### AutomationTriggerMetadataInput

| Parameter        | Type      | Required | Description                                                                                            |
| ---------------- | --------- | -------- | ------------------------------------------------------------------------------------------------------ |
| `incompleteOnly` | `Boolean` | No       | Only fire for incomplete records (for due-date triggers).                                              |
| `offset`         | `Int`     | No       | Minutes before/after the due date to fire. Read by `DUE_DATE_EXPIRED` to schedule its monitoring jobs. |

### CreateAutomationActionInput

The fields that apply depend on `type` — e.g. `tagIds` for `ADD_TAG`/`REMOVE_TAG`, `assigneeIds` for `ADD_ASSIGNEE`, `duedIn` for `CHANGE_DUE_DATE`, `metadata.email` for `SEND_EMAIL`, `httpOption` for `MAKE_HTTP_REQUEST`.

| Parameter              | Type                              | Required | Description                                                                                            |
| ---------------------- | --------------------------------- | -------- | ------------------------------------------------------------------------------------------------------ |
| `type`                 | `AutomationActionType!`           | Yes      | The action to run. See [AutomationActionType](#automationactiontype).                                  |
| `duedIn`               | `Int`                             | No       | Days from the trigger to set the due date (for `CHANGE_DUE_DATE`).                                     |
| `customFieldId`        | `String`                          | No       | Target custom field (for `ADD_CUSTOM_FIELD`/`REMOVE_CUSTOM_FIELD`).                                    |
| `customFieldOptionIds` | `[String!]`                       | No       | Custom-field option IDs to set.                                                                        |
| `todoListId`           | `String`                          | No       | Target list for `CHANGE_TODO_LIST`.                                                                    |
| `metadata`             | `AutomationActionMetadataInput`   | No       | Action-specific payload (email, checklists, copy options, field values).                               |
| `tagIds`               | `[String!]`                       | No       | Tags to add or remove.                                                                                 |
| `assigneeIds`          | `[String!]`                       | No       | Assignees to add or remove.                                                                            |
| `projectIds`           | `[String!]`                       | No       | Target workspaces for `ADD_PROJECT`.                                                                   |
| `color`                | `String`                          | No       | Color to apply (hex), for `ADD_COLOR`.                                                                 |
| `assigneeTriggerer`    | `String`                          | No       | Used by `ADD_ASSIGNEE_TRIGGERER`/`REMOVE_ASSIGNEE_TRIGGERER` to act on the user who fired the trigger. |
| `portableDocumentId`   | `String`                          | No       | Document template ID for `GENERATE_PDF`.                                                               |
| `httpOption`           | `AutomationActionHttpOptionInput` | No       | HTTP request configuration for `MAKE_HTTP_REQUEST`.                                                    |

### AutomationActionMetadataInput

| Parameter         | Type                                     | Required | Description                                                              |
| ----------------- | ---------------------------------------- | -------- | ------------------------------------------------------------------------ |
| `email`           | `AutomationActionSendEmailInput`         | No       | Email configuration for `SEND_EMAIL`.                                    |
| `checklists`      | `[AutomationActionCreateChecklistInput]` | No       | Checklists to create for `CREATE_CHECKLIST`.                             |
| `copyTodoOptions` | `[CopyRecordOption!]`                    | No       | What to copy for `COPY_TODO`. See [CopyRecordOption](#copyrecordoption). |
| `number`          | `Float`                                  | No       | Numeric value for custom-field actions.                                  |
| `text`            | `String`                                 | No       | Text value for custom-field actions.                                     |

### AutomationActionSendEmailInput

Used by `SEND_EMAIL` actions. `to`, `content`, and `subject` are required; the rest are optional.

| Parameter     | Type                                          | Required | Description                    |
| ------------- | --------------------------------------------- | -------- | ------------------------------ |
| `to`          | `[String!]!`                                  | Yes      | Recipient email addresses.     |
| `content`     | `String!`                                     | Yes      | Email body. HTML is supported. |
| `subject`     | `String!`                                     | Yes      | Email subject line.            |
| `from`        | `String`                                      | No       | Sender address.                |
| `cc`          | `[String!]`                                   | No       | CC recipients.                 |
| `bcc`         | `[String!]`                                   | No       | BCC recipients.                |
| `replyTo`     | `[String!]`                                   | No       | Reply-to addresses.            |
| `attachments` | `[AutomationActionSendEmailAttachmentInput!]` | No       | File attachments.              |

### AutomationActionSendEmailAttachmentInput

| Parameter   | Type      | Required | Description         |
| ----------- | --------- | -------- | ------------------- |
| `uid`       | `String!` | Yes      | Uploaded file UID.  |
| `name`      | `String!` | Yes      | File name.          |
| `size`      | `Float!`  | Yes      | File size in bytes. |
| `type`      | `String!` | Yes      | MIME type.          |
| `extension` | `String!` | Yes      | File extension.     |

### AutomationActionCreateChecklistInput

Used by `CREATE_CHECKLIST` actions. `title` and `position` are required.

| Parameter        | Type                                          | Required | Description                           |
| ---------------- | --------------------------------------------- | -------- | ------------------------------------- |
| `title`          | `String!`                                     | Yes      | Checklist title.                      |
| `position`       | `Float!`                                      | Yes      | Sort position among checklists.       |
| `checklistItems` | `[AutomationActionCreateChecklistItemInput!]` | No       | Items to create inside the checklist. |

### AutomationActionCreateChecklistItemInput

| Parameter     | Type        | Required | Description                                    |
| ------------- | ----------- | -------- | ---------------------------------------------- |
| `title`       | `String!`   | Yes      | Item title.                                    |
| `position`    | `Float!`    | Yes      | Sort position among items.                     |
| `assigneeIds` | `[String!]` | No       | Users to assign to the item.                   |
| `duedIn`      | `Int`       | No       | Days from creation to set the item's due date. |

### AutomationActionHttpOptionInput

Used by `MAKE_HTTP_REQUEST` actions.

| Parameter                  | Type                              | Required | Description                                               |
| -------------------------- | --------------------------------- | -------- | --------------------------------------------------------- |
| `url`                      | `String!`                         | Yes      | The URL to call.                                          |
| `method`                   | `HttpMethod!`                     | Yes      | One of `GET`, `POST`, `PUT`, `DELETE`, `PATCH`.           |
| `headers`                  | `[HttpHeaderInput]`               | No       | Request headers.                                          |
| `parameters`               | `[HttpParameterInput]`            | No       | Query parameters.                                         |
| `contentType`              | `HttpContentType`                 | No       | `JSON` or `TEXT`.                                         |
| `body`                     | `String`                          | No       | Request body.                                             |
| `authorizationType`        | `HttpAuthorizationType`           | No       | `BASIC_AUTH`, `BEARER`, `API_KEY`, or `OAUTH2`.           |
| `authorizationBasicAuth`   | `HttpAuthorizationBasicAuthInput` | No       | Credentials when `authorizationType` is `BASIC_AUTH`.     |
| `authorizationBearerToken` | `String`                          | No       | Token when `authorizationType` is `BEARER`.               |
| `authorizationApiKey`      | `HttpAuthorizationApiKeyInput`    | No       | Key config when `authorizationType` is `API_KEY`.         |
| `oauthConnectionId`        | `String`                          | No       | OAuth connection ID when `authorizationType` is `OAUTH2`. |

### HttpHeaderInput / HttpParameterInput

| Parameter | Type      | Required | Description                |
| --------- | --------- | -------- | -------------------------- |
| `key`     | `String!` | Yes      | Header or parameter name.  |
| `value`   | `String!` | Yes      | Header or parameter value. |

### HttpAuthorizationBasicAuthInput

| Parameter  | Type      | Required | Description          |
| ---------- | --------- | -------- | -------------------- |
| `username` | `String!` | Yes      | Basic-auth username. |
| `password` | `String!` | Yes      | Basic-auth password. |

### HttpAuthorizationApiKeyInput

| Parameter | Type                            | Required | Description                                 |
| --------- | ------------------------------- | -------- | ------------------------------------------- |
| `key`     | `String!`                       | Yes      | API key name.                               |
| `value`   | `String!`                       | Yes      | API key value.                              |
| `passBy`  | `HttpAuthorizationApiKeyPassBy` | No       | Where to send the key: `HEADER` or `QUERY`. |

### AutomationTriggerType

| Value                                     | Fires when                                                         |
| ----------------------------------------- | ------------------------------------------------------------------ |
| `TODO_CREATED`                            | A record is created.                                               |
| `TODO_LIST_CHANGED`                       | A record moves to a different list.                                |
| `TODO_MARKED_AS_COMPLETE`                 | A record is completed.                                             |
| `TODO_MARKED_AS_INCOMPLETE`               | A record is reopened.                                              |
| `ASSIGNEE_ADDED`                          | An assignee is added.                                              |
| `ASSIGNEE_REMOVED`                        | An assignee is removed.                                            |
| `DUE_DATE_CHANGED`                        | A record's due date changes.                                       |
| `DUE_DATE_REMOVED`                        | A record's due date is cleared.                                    |
| `TAG_ADDED`                               | A tag is added.                                                    |
| `TAG_REMOVED`                             | A tag is removed.                                                  |
| `CHECKLIST_ITEM_MARKED_AS_DONE`           | A checklist item is checked.                                       |
| `CHECKLIST_ITEM_MARKED_AS_UNDONE`         | A checklist item is unchecked.                                     |
| `DUE_DATE_EXPIRED`                        | A record's due date passes (uses `metadata.offset`).               |
| `TODO_COPIED_OR_MOVED_FROM_OTHER_PROJECT` | A record is copied or moved in from another workspace.             |
| `CUSTOM_FIELD_ADDED`                      | A custom-field value is set.                                       |
| `CUSTOM_FIELD_REMOVED`                    | A custom-field value is cleared.                                   |
| `COLOR_ADDED`                             | A color is added.                                                  |
| `COLOR_REMOVED`                           | A color is removed.                                                |
| `CUSTOM_FIELD_BUTTON_CLICKED`             | A button custom field is clicked.                                  |
| `PROJECT_ADDED_TO_TODO`                   | A record is added to a workspace.                                  |
| `SCHEDULED`                               | A schedule elapses (requires `schedule`).                          |
| `CONDITIONAL`                             | A record starts or stops matching a filter (uses `conditionMode`). |

### AutomationActionType

| Value                           | Action                                            |
| ------------------------------- | ------------------------------------------------- |
| `CHANGE_TODO_LIST`              | Move the record to `todoListId`.                  |
| `MARK_AS_COMPLETE`              | Complete the record.                              |
| `MARK_AS_INCOMPLETE`            | Reopen the record.                                |
| `ADD_ASSIGNEE`                  | Add `assigneeIds`.                                |
| `REMOVE_ASSIGNEE`               | Remove `assigneeIds`.                             |
| `CHANGE_DUE_DATE`               | Set the due date `duedIn` days out.               |
| `REMOVE_DUE_DATE`               | Clear the due date.                               |
| `ADD_TAG`                       | Add `tagIds`.                                     |
| `REMOVE_TAG`                    | Remove `tagIds`.                                  |
| `MARK_CHECKLIST_ITEM_AS_DONE`   | Check matching checklist items.                   |
| `MARK_CHECKLIST_ITEM_AS_UNDONE` | Uncheck matching checklist items.                 |
| `CREATE_CHECKLIST`              | Create checklists from `metadata.checklists`.     |
| `COPY_TODO`                     | Copy the record using `metadata.copyTodoOptions`. |
| `SEND_EMAIL`                    | Send the email in `metadata.email`.               |
| `ADD_CUSTOM_FIELD`              | Set a custom-field value.                         |
| `REMOVE_CUSTOM_FIELD`           | Clear a custom-field value.                       |
| `ADD_COLOR`                     | Apply `color`.                                    |
| `REMOVE_COLOR`                  | Remove the color.                                 |
| `GENERATE_PDF`                  | Generate a PDF from `portableDocumentId`.         |
| `MAKE_HTTP_REQUEST`             | Call an external URL via `httpOption`.            |
| `ADD_ASSIGNEE_TRIGGERER`        | Assign the user who fired the trigger.            |
| `REMOVE_ASSIGNEE_TRIGGERER`     | Unassign the user who fired the trigger.          |
| `ARCHIVE_TODO`                  | Archive the record (removes it from the board).   |
| `ADD_PROJECT`                   | Add the record to `projectIds`.                   |

### AutomationScheduleType

| Value      | Meaning                               |
| ---------- | ------------------------------------- |
| `DAILY`    | Every day at `hour`:`minute`.         |
| `WEEKLY`   | On the given `dayOfWeek` each week.   |
| `WEEKDAYS` | Monday through Friday.                |
| `MONTHLY`  | On `dayOfMonth` each month.           |
| `YEARLY`   | On `dayOfMonth` of `month` each year. |
| `CUSTOM`   | On a `cronExpression`.                |

### CopyRecordOption

What to carry over when a `COPY_TODO` action runs: `DESCRIPTION`, `DUE_DATE`, `ASSIGNEES`, `TAGS`, `COMMENTS`, `CHECKLISTS`, `CUSTOM_FIELDS`.

## Response

`createAutomation` returns the created `Automation`. The example above resolves to:

```json
{
  "data": {
    "createAutomation": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "isActive": true,
      "trigger": {
        "id": "clm4n8r2k000108l0b7y2f3qd",
        "type": "TODO_CREATED"
      },
      "actions": [
        {
          "id": "clm4n8r7p000208l0c9k4h5wz",
          "type": "ADD_TAG"
        }
      ],
      "createdAt": "2026-05-29T14:21:08.000Z"
    }
  }
}
```

### Returns

| Field       | Type                   | Description                                                   |
| ----------- | ---------------------- | ------------------------------------------------------------- |
| `id`        | `ID!`                  | Unique identifier for the automation.                         |
| `trigger`   | `AutomationTrigger!`   | The configured trigger.                                       |
| `actions`   | `[AutomationAction!]!` | The configured actions.                                       |
| `isActive`  | `Boolean!`             | Whether the automation is active. New automations are active. |
| `createdBy` | `User!`                | The user who created the automation.                          |
| `project`   | `Workspace!`           | The workspace the automation belongs to.                      |
| `createdAt` | `DateTime!`            | Creation timestamp.                                           |
| `updatedAt` | `DateTime!`            | Last update timestamp.                                        |

## Full example

A `TODO_LIST_CHANGED` trigger scoped to one list, with four actions: set a due date, add assignees, send an email, and call a webhook. Note `contentType: JSON`, `authorizationType: BEARER`, and the required email fields (`to`, `subject`, `content`).

```graphql
mutation CreateAutomationAdvanced {
  createAutomation(
    input: {
      trigger: {
        type: TODO_LIST_CHANGED
        todoListId: "list_123"
        tagIds: ["tag_123"]
        assigneeIds: ["user_123"]
        color: "#FF5733"
      }
      actions: [
        { type: CHANGE_DUE_DATE, duedIn: 7 }
        { type: ADD_ASSIGNEE, assigneeIds: ["user_123", "user_456"] }
        {
          type: SEND_EMAIL
          metadata: {
            email: {
              to: ["team@example.com"]
              subject: "Record moved to In Progress"
              content: "<p>A record has been moved.</p>"
              cc: ["manager@example.com"]
              replyTo: ["ops@example.com"]
            }
          }
        }
        {
          type: MAKE_HTTP_REQUEST
          httpOption: {
            url: "https://api.example.com/webhook"
            method: POST
            contentType: JSON
            body: "{\"event\": \"record_moved\"}"
            headers: [{ key: "X-Custom-Header", value: "automation" }]
            authorizationType: BEARER
            authorizationBearerToken: "YOUR_BEARER_TOKEN"
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
      color
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
        email
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
            to
            subject
          }
        }
      }
      httpOption {
        url
        method
        contentType
      }
    }
    createdBy {
      id
      fullName
      email
    }
    createdAt
  }
}
```

### Scheduled automation

A `SCHEDULED` trigger requires a `schedule`. This one fires every Monday at 9:00 AM Eastern (`dayOfWeek: [1]`) and copies the record with its assignees and tags.

```graphql
mutation CreateScheduledAutomation {
  createAutomation(
    input: {
      trigger: {
        type: SCHEDULED
        schedule: { type: WEEKLY, hour: 9, minute: 0, timezone: "America/New_York", dayOfWeek: [1] }
      }
      actions: [{ type: COPY_TODO, metadata: { copyTodoOptions: [ASSIGNEES, TAGS, CHECKLISTS] } }]
    }
  ) {
    id
    trigger {
      id
      type
      schedule {
        type
        hour
        minute
        timezone
        dayOfWeek
        nextExecutionAt
      }
    }
  }
}
```

### Conditional automation

A `CONDITIONAL` trigger fires when a record starts (or stops) matching a saved filter. Set `conditionMode` to `STARTS_MATCHING` or `STOPS_MATCHING`. The filter itself lives in the opaque `filterGroups`/`filterGroupLinks` payload — build it in the workspace automation UI and read it back; it is not meant to be hand-authored.

```graphql
mutation CreateConditionalAutomation {
  createAutomation(
    input: {
      trigger: { type: CONDITIONAL, conditionMode: STARTS_MATCHING, op: AND }
      actions: [{ type: ADD_TAG, tagIds: ["tag_123"] }]
    }
  ) {
    id
    trigger {
      id
      type
      conditionMode
      op
    }
  }
}
```

### Due-date automation

A `DUE_DATE_EXPIRED` trigger fires when a record's due date passes. `metadata.offset` shifts that moment in minutes (negative = before, positive = after); the resolver uses it to schedule the due-date monitoring jobs. This one emails the assignees 60 minutes before the due date, only for incomplete records.

```graphql
mutation CreateDueDateAutomation {
  createAutomation(
    input: {
      trigger: { type: DUE_DATE_EXPIRED, metadata: { offset: -60, incompleteOnly: true } }
      actions: [
        {
          type: SEND_EMAIL
          metadata: {
            email: {
              to: ["owner@example.com"]
              subject: "A record is due in 1 hour"
              content: "<p>Heads up — this record is due soon.</p>"
            }
          }
        }
      ]
    }
  ) {
    id
    trigger {
      id
      type
      metadata {
        ... on AutomationTriggerMetadataTodoOverdue {
          offset
          incompleteOnly
        }
      }
    }
  }
}
```

## Errors

| Code                | When                                                                                                                 |
| ------------------- | -------------------------------------------------------------------------------------------------------------------- |
| `UNAUTHENTICATED`   | No valid authentication was provided. Message: `You are not authenticated.`                                          |
| `FORBIDDEN`         | The user is not an OWNER or ADMIN of the workspace, or the workspace is archived. Message: `You are not authorized.` |
| `PROJECT_NOT_FOUND` | The `blue-workspace-id` header is missing or does not resolve. Message: `Project was not found.`                     |

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

Only `OWNER` and `ADMIN` members of an **active** workspace can create automations. `MEMBER` and `CLIENT` cannot, and any role is blocked on an archived workspace.

## Related

- [Edit an automation](/api/automations/edit-automation)
- [Copy an automation](/api/automations/copy-automation)
- [Delete an automation](/api/automations/delete-automation)
- [Automations overview](/api/automations)
