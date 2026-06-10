---
title: Remove Users
description: Remove a user from a single workspace or from the entire organization, including all the cleanup each removal triggers.
icon: Trash2
order: 3
---

Removing a user revokes their access and cleans up their assignments. Blue offers two scopes, mapped to two mutations:

- **`removeProjectUser`** detaches a user from a single workspace (`Project`) while keeping their organization (`Company`) membership and access to other workspaces.
- **`removeCompanyUser`** removes a user from the organization entirely, cascading the removal across every workspace they belonged to.

Both are permanent. There is no "undo" — re-granting access means re-inviting the user with [`inviteUser`](/api/user-management/invite-user). Historical data the user created (records, comments, activity) is preserved.

## Authentication

All requests go to `https://api.blue.app/graphql` with your personal access token. The project header is not required for either mutation — both take their scope from the input.

```bash
curl https://api.blue.app/graphql \
  -H "X-Bloo-Token-ID: YOUR_TOKEN_ID" \
  -H "X-Bloo-Token-Secret: YOUR_TOKEN_SECRET" \
  -H "X-Bloo-Company-ID: YOUR_COMPANY_ID" \
  -H "Content-Type: application/json" \
  --data '{"query":"..."}'
```

Header names are case-insensitive. See [Authentication](/api/start-guide/authentication) for token setup.

## Request

### Remove a user from a workspace

Use `removeProjectUser` to detach a user from one workspace. The `projectId` must be the workspace **ID** — this mutation does not accept a slug.

```graphql
mutation RemoveWorkspaceUser {
  removeProjectUser(input: { projectId: "project_123", userId: "user_123" }) {
    success
    operationId
  }
}
```

### Remove a user from the organization

Use `removeCompanyUser` to remove a user from the whole organization. `companyId` accepts either the organization **ID or slug**.

```graphql
mutation RemoveOrganizationUser {
  removeCompanyUser(input: { companyId: "company_123", userId: "user_123" })
}
```

## Parameters

### RemoveProjectUserInput

| Parameter   | Type      | Required | Description                                                                         |
| ----------- | --------- | -------- | ----------------------------------------------------------------------------------- |
| `projectId` | `String!` | Yes      | ID of the workspace to remove the user from. ID only — a slug is not accepted here. |
| `userId`    | `String!` | Yes      | ID of the user to remove.                                                           |

### RemoveCompanyUserInput

| Parameter   | Type      | Required | Description                         |
| ----------- | --------- | -------- | ----------------------------------- |
| `companyId` | `String!` | Yes      | ID **or slug** of the organization. |
| `userId`    | `String!` | Yes      | ID of the user to remove.           |

## Response

`removeProjectUser` returns a `MutationResult`. The resolver returns only `{ success: true }`, so `operationId` is always `null` for this mutation.

```json
{
  "data": {
    "removeProjectUser": {
      "success": true,
      "operationId": null
    }
  }
}
```

`removeCompanyUser` returns a bare `Boolean` — `true` on success. There are no subfields to select.

```json
{
  "data": {
    "removeCompanyUser": true
  }
}
```

### Returns

#### removeProjectUser — `MutationResult!`

| Field         | Type       | Description                       |
| ------------- | ---------- | --------------------------------- |
| `success`     | `Boolean!` | `true` when the user was removed. |
| `operationId` | `String`   | Always `null` for this mutation.  |

#### removeCompanyUser — `Boolean`

`true` when the user was removed from the organization.

## Side effects

These mutations do more than drop a membership row. Know what each one cleans up before you call it.

### removeProjectUser

- Unassigns the user from every record (`Todo`) in the workspace.
- Removes the user from any **assignee** custom-field values on records in the workspace.
- Deletes the user's **personal** saved views in the workspace (shared saved views are kept).
- Deletes the user's per-workspace folders.
- Removes the workspace membership and publishes a real-time update to other members.
- Writes an activity-log entry.

### removeCompanyUser

Cascades across every workspace in the organization:

- Unassigns the user from all records in all workspaces.
- Removes the user from all assignee custom-field values across the organization.
- Deletes the user's workspace folders everywhere.
- Removes every workspace membership.
- Removes the organization membership (which cascades the organization-level folders).
- Decrements the organization's billed seat count by syncing the Stripe subscription quantity.
- Sends a removal-notification email to the removed user.
- Writes an activity-log entry.

<Callout variant="warning" title="Removing the last seat does not cancel billing">

`removeCompanyUser` lowers the billed seat quantity on the Stripe subscription but does not cancel the subscription. Manage the plan itself through billing, not by removing users.

</Callout>

## Errors

| Code                | Returned by         | When                                                                                                                                                          |
| ------------------- | ------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `PROJECT_NOT_FOUND` | `removeProjectUser` | No workspace matches `projectId`.                                                                                                                             |
| `COMPANY_NOT_FOUND` | `removeCompanyUser` | No organization matches the `companyId` ID or slug.                                                                                                           |
| `USER_NOT_FOUND`    | both                | No user matches `userId`.                                                                                                                                     |
| `FORBIDDEN`         | both                | The caller lacks the required role, the target is a workspace `OWNER` (cannot be removed), or the target user is not a member of that workspace/organization. |

A `FORBIDDEN` error covers every authorization failure these mutations raise:

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

### removeProjectUser

Requires a workspace access level of `OWNER` or `ADMIN`. A workspace `OWNER` cannot be removed — transfer ownership first.

| Workspace access level | Can remove users |
| ---------------------- | ---------------- |
| `OWNER`                | Yes              |
| `ADMIN`                | Yes              |
| `MEMBER`               | No               |
| `COMMENT_ONLY`         | No               |
| `VIEW_ONLY`            | No               |

### removeCompanyUser

Requires an organization access level of `OWNER`. Organization `ADMIN`s cannot remove users.

| Organization access level | Can remove users |
| ------------------------- | ---------------- |
| `OWNER`                   | Yes              |
| `ADMIN`                   | No               |
| `MEMBER`                  | No               |
| `COMMENT_ONLY`            | No               |
| `VIEW_ONLY`               | No               |

## Related

- [Invite a User](/api/user-management/invite-user) — re-grant access after a removal.
- [List Users](/api/user-management/list-users) — find a user's ID before removing them.
- [Retrieve Custom Role](/api/user-management/retrieve-custom-role) — inspect a user's role and permissions.
