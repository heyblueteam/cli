---
title: User Management
description: Invite, list, and remove users, manage pending invitations, and define custom roles across Blue workspaces and organizations.
icon: Users
order: 0
---

The User Management API covers the full lifecycle of a teammate's access: inviting them, listing who has access, managing the invitations they haven't accepted yet, removing them, and defining custom roles with granular permissions.

Access is granted at two scopes:

- **Workspace** — a user belongs to a single workspace (`Workspace`) with an access level scoped to that workspace.
- **Organization** — a user belongs to the organization (`Organization`) and, through it, to one or more of its workspaces.

In the API, workspaces are `Workspace` objects and organizations are `Organization` objects. Access levels are the `UserAccessLevel` enum; custom roles are `ProjectUserRole` objects scoped to a single workspace.

## Authentication

All requests go to the GraphQL endpoint with your personal access token in headers:

```
POST https://api.blue.app/graphql
blue-token-id: YOUR_TOKEN_ID
blue-token-secret: YOUR_TOKEN_SECRET
blue-org-id: YOUR_ORG_ID
blue-workspace-id: project_123   # required for workspace-scoped operations
```

Headers are case-insensitive. The Company and Project headers accept either an ID or a slug. See [Authentication](/api/start-guide/authentication) for how to create a token.

## Operations

### Users

| Operation                                               | Type     | Description                                               |
| ------------------------------------------------------- | -------- | --------------------------------------------------------- |
| [`workspaceUserList`](/api/user-management/list-users)    | Query    | List users in a workspace, with pagination and search.    |
| [`assignees`](/api/records/assignees)                   | Query    | List users who can be assigned to records in a workspace. |
| `currentUser`                                           | Query    | Return the authenticated user.                            |
| [`inviteUser`](/api/user-management/invite-user)        | Mutation | Invite a user to a workspace or organization.             |
| [`removeProjectUser`](/api/user-management/remove-user) | Mutation | Remove a user from a single workspace.                    |
| [`removeOrganizationUser`](/api/user-management/remove-user) | Mutation | Remove a user from the entire organization.               |

### Invitations

| Operation          | Type     | Description                                            |
| ------------------ | -------- | ------------------------------------------------------ |
| `invitations`      | Query    | List pending and accepted invitations for a workspace. |
| `myInvitations`    | Query    | List invitations addressed to the authenticated user.  |
| `resendInvitation` | Mutation | Re-send an invitation email and refresh its expiry.    |
| `cancelInvitation` | Mutation | Revoke a pending invitation.                           |

### Custom roles

| Operation                                                                           | Type     | Description                          |
| ----------------------------------------------------------------------------------- | -------- | ------------------------------------ |
| [`createWorkspaceUserRole`](/api/user-management/retrieve-custom-role)                | Mutation | Create a custom role in a workspace. |
| [`updateWorkspaceUserRole`](/api/user-management/retrieve-custom-role)                | Mutation | Update a custom role.                |
| [`deleteWorkspaceUserRole`](/api/user-management/retrieve-custom-role)                | Mutation | Delete a custom role.                |
| [`workspaceUserRole` / `workspaceUserRoles`](/api/user-management/retrieve-custom-role) | Query    | Read one or many custom roles.       |

## Access levels

`UserAccessLevel` has exactly six values, in descending order of capability:

| Level          | Description                                                                                 |
| -------------- | ------------------------------------------------------------------------------------------- |
| `OWNER`        | Full control over the workspace or organization. Ownership is single-slot per organization. |
| `ADMIN`        | Manages users, settings, and content.                                                       |
| `MEMBER`       | Standard teammate with full workspace functionality.                                        |
| `CLIENT`       | External collaborator with limited visibility.                                              |
| `COMMENT_ONLY` | Can view and comment, but not edit.                                                         |
| `VIEW_ONLY`    | Read-only.                                                                                  |

### Who can invite which level

The level you can assign depends on your own access level in the target workspace. `VIEW_ONLY` and `COMMENT_ONLY` users cannot invite at all.

| Inviter level | Levels they can assign                           |
| ------------- | ------------------------------------------------ |
| `OWNER`       | All six levels, including `OWNER`.               |
| `ADMIN`       | Every level except `OWNER`.                      |
| `MEMBER`      | `MEMBER`, `CLIENT`, `COMMENT_ONLY`, `VIEW_ONLY`. |
| `CLIENT`      | `CLIENT`, `COMMENT_ONLY`, `VIEW_ONLY`.           |

<Callout variant="warning" title="Organization invites require ownership">

Passing `companyId` to <code>inviteUser</code> invites at the organization level, and only the organization <code>OWNER</code> can do this. When an existing organization owner is added to a workspace, they are always added as <code>ADMIN</code> regardless of the <code>accessLevel</code> you request.

</Callout>

## Common patterns

Every example below is valid against the current schema. Inline IDs (`project_123`, `user_123`, …) are placeholders; substitute real IDs or slugs.

### Invite a user to a workspace

`inviteUser` returns a plain `Boolean!` — there is no selection set. See [Invite User](/api/user-management/invite-user) for the full parameter reference.

```graphql
mutation InviteToWorkspace {
  inviteUser(
    input: { email: "teammate@example.com", projectId: "project_123", accessLevel: MEMBER }
  )
}
```

```json
{ "data": { "inviteUser": true } }
```

### Invite a user to the organization

Supply `companyId` (requires organization `OWNER`) and, optionally, `projectIds` to add them to specific workspaces in the same call.

```graphql
mutation InviteToOrganization {
  inviteUser(
    input: {
      email: "manager@example.com"
      companyId: "company_123"
      projectIds: ["project_123", "project_456"]
      accessLevel: ADMIN
    }
  )
}
```

### List users in a workspace

`workspaceUserList` returns a `ProjectUserList` with `users`, a `pageInfo`, and a `totalCount`.

```graphql
query WorkspaceUsers {
  workspaceUserList(projectId: "project_123", first: 25) {
    users {
      id
      fullName
      email
      image {
        medium
      }
      projectLevel(projectId: "project_123")
    }
    pageInfo {
      totalItems
      hasNextPage
    }
    totalCount
  }
}
```

```json
{
  "data": {
    "workspaceUserList": {
      "users": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "fullName": "Ada Lovelace",
          "email": "ada@example.com",
          "image": { "medium": "https://blue.app/files/clm4n8qwx/medium.png" },
          "projectLevel": "MEMBER"
        }
      ],
      "pageInfo": { "totalItems": 12, "hasNextPage": false },
      "totalCount": 12
    }
  }
}
```

### List pending invitations

Pass `pending: true` to return only invitations that haven't been accepted yet.

```graphql
query PendingInvitations {
  invitations(projectId: "project_123", pending: true) {
    id
    email
    accessLevel
    expiredAt
    invitedBy {
      fullName
    }
  }
}
```

### Resend or cancel an invitation

`resendInvitation` refreshes the invitation's expiry and returns the `Invitation`. `cancelInvitation` revokes it and returns a `Boolean!`.

```graphql
mutation RefreshInvite {
  resendInvitation(id: "invitation_123") {
    id
    email
    expiredAt
  }
}
```

```graphql
mutation RevokeInvite {
  cancelInvitation(id: "invitation_123")
}
```

### Remove a user

Use `removeProjectUser` to revoke a single workspace, or `removeOrganizationUser` to remove the user from the whole organization. `removeProjectUser` returns a `MutationResult` (`success`, `operationId`); `removeOrganizationUser` returns a nullable `Boolean`.

```graphql
mutation RemoveFromWorkspace {
  removeProjectUser(input: { projectId: "project_123", userId: "user_123" }) {
    success
    operationId
  }
}
```

```graphql
mutation RemoveFromOrganization {
  removeOrganizationUser(input: { companyId: "company_123", userId: "user_123" })
}
```

You cannot remove a user whose level is `OWNER`; transfer ownership first. See [Remove Users](/api/user-management/remove-user) for the cleanup each removal performs.

### Create a custom role

`createWorkspaceUserRole` takes flat boolean toggles (no nested `permissions` object) and returns a `ProjectUserRole`. Select only the fields the role exposes.

```graphql
mutation CreateRole {
  createWorkspaceUserRole(
    input: {
      projectId: "project_123"
      name: "Content Reviewer"
      isRecordsEnabled: true
      canCreateRecords: false
      canDeleteRecords: false
    }
  ) {
    id
    name
    canCreateRecords
    canDeleteRecords
  }
}
```

```json
{
  "data": {
    "createWorkspaceUserRole": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Content Reviewer",
      "canCreateRecords": false,
      "canDeleteRecords": false
    }
  }
}
```

See [Custom Roles](/api/user-management/retrieve-custom-role) for the full toggle list and per-field/per-list visibility controls.

## Errors

| Code                          | When                                                                                                                                                       |
| ----------------------------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `FORBIDDEN`                   | The caller lacks permission for the requested action (e.g. inviting a level above their own, or inviting at the organization level without being `OWNER`). |
| `PLAN_LIMIT_REACHED`          | The organization's seat limit is reached: _"Users limit reached (current/limit). Upgrade your plan for more."_                                             |
| `USER_ALREADY_IN_THE_PROJECT` | The invitee is already a member of the workspace.                                                                                                          |
| `ADD_SELF`                    | The caller tried to invite their own email address.                                                                                                        |
| `PROJECT_NOT_FOUND`           | The `projectId` does not resolve to a workspace the caller can access.                                                                                     |
| `COMPANY_NOT_FOUND`           | The `companyId` does not resolve to an organization.                                                                                                       |
| `USER_NOT_FOUND`              | The `userId` being removed does not exist or is not in the target scope.                                                                                   |
| `PROJECT_USER_ROLE_NOT_FOUND` | The `roleId` passed to `inviteUser` does not exist in the target workspace.                                                                                |
| `COMPANY_BANNED`              | The organization account is suspended. Contact support.                                                                                                    |

## Related

- [Invite User](/api/user-management/invite-user)
- [List Users](/api/user-management/list-users)
- [Remove Users](/api/user-management/remove-user)
- [Custom Roles](/api/user-management/retrieve-custom-role)
- [Workspaces](/api/workspaces)
- [Record assignees](/api/records/assignees)
