---
title: Assignee Field
description: Assign one or more users to a record through a custom field, separate from the record's built-in assignees.
icon: UserPlus
order: 26
---

An assignee field lets you attach users to a record through a dedicated custom field, independent of the record's built-in assignees. Its `CustomFieldType` is `ASSIGNEE`. Use it for roles like reviewer, approver, or account owner that you want to track and filter separately from who is doing the work.

## Overview

An assignee field stores a set of users. You can cap it to a single user, choose whether assignment sends a notification, and limit who is eligible to be assigned using the same access rules as any other field.

## Create

```graphql
mutation CreateReviewerField {
  createCustomField(
    input: {
      name: "Reviewer"
      type: ASSIGNEE
      assigneeMultiple: false
      assigneeNotifyOnAssign: true
      allowedLevels: [ADMIN, MEMBER]
    }
  ) {
    id
    name
    type
    assigneeMultiple
    assigneeNotifyOnAssign
  }
}
```

| Parameter                | Type                 | Required | Description                                              |
| ------------------------ | -------------------- | -------- | -------------------------------------------------------- |
| `name`                   | `String!`            | Yes      | Display name of the field.                               |
| `type`                   | `CustomFieldType!`   | Yes      | Must be `ASSIGNEE`.                                      |
| `assigneeMultiple`       | `Boolean`            | No       | Allow more than one user. Defaults to single-user.       |
| `assigneeNotifyOnAssign` | `Boolean`            | No       | Notify a user when they are assigned through this field. |
| `allowedUserIds`         | `[String!]`          | No       | Restrict eligible assignees to these users.              |
| `allowedRoleIds`         | `[String!]`          | No       | Restrict eligible assignees to these custom roles.       |
| `allowedLevels`          | `[UserAccessLevel!]` | No       | Restrict eligible assignees to these access levels.      |

The field is scoped to the workspace in the `blue-workspace-id` header — there is no `projectId` input field.

## Set a value

Pass the user IDs to assign in `assigneeUserIds`. The mutation returns `Boolean!`, so it has no selection set.

```graphql
mutation AssignReviewer {
  setRecordCustomField(
    input: { todoId: "todo_123", customFieldId: "field_123", assigneeUserIds: ["user_123"] }
  )
}
```

```json
{ "data": { "setRecordCustomField": true } }
```

If the field is single-user (`assigneeMultiple: false`), passing more than one ID returns a `BAD_USER_INPUT` error. Assigning a user who is not eligible — outside `allowedUserIds`/`allowedRoleIds`/`allowedLevels` — also returns `BAD_USER_INPUT`.

## Read a value

Read assigned users from `assignedUsers` on the field, and the users who may be assigned from `eligibleAssignees`.

```graphql
query RecordReviewers {
  recordQueries {
    todos(filter: { companyIds: [], projectIds: ["project_123"] }, limit: 5) {
      items {
        id
        title
        customFields {
          name
          type
          assignedUsers {
            id
            fullName
            email
          }
        }
      }
    }
  }
}
```

```json
{
  "data": {
    "recordQueries": {
      "todos": {
        "items": [
          {
            "id": "clm4n8qwx000008l0g4oxdqn7",
            "title": "Acme onboarding",
            "customFields": [
              {
                "name": "Reviewer",
                "type": "ASSIGNEE",
                "assignedUsers": [
                  {
                    "id": "clm4u1a9x000308l0c1d2e3f4",
                    "fullName": "Dana Lee",
                    "email": "dana@acme.example"
                  }
                ]
              }
            ]
          }
        ]
      }
    }
  }
}
```

## Notes

- An assignee field is distinct from the record's built-in assignees (`Record.users`). Setting one does not change the other.
- To clear all assignees, call `setRecordCustomField` with an empty `assigneeUserIds: []`.

## Related

- [Single-select field](/api/custom-fields/select-single)
- [Set field values](/api/custom-fields/custom-field-values)
- [Custom fields overview](/api/custom-fields)
