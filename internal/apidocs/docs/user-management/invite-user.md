---
title: Invite User
description: Invite people to a workspace or organization with a chosen access level or custom role.
icon: UserPlus
order: 2
---

Use the `inviteUser` mutation to add a person to a workspace or organization by email. If the email already belongs to a registered Blue user who is part of the organization, they are added to the workspace immediately. Otherwise Blue sends them an invitation email; they join once they accept. Workspaces are `Project` objects and organizations are `Company` objects in the API.

`inviteUser` returns a bare `Boolean!` — there is no object to select fields from. A successful call resolves to `true`.

## Request

Send the request to `https://api.blue.app/graphql` with your authentication headers. The `X-Bloo-Project-ID` header (workspace ID or slug) is required for workspace-scoped invites.

```http
POST https://api.blue.app/graphql
X-Bloo-Token-ID: YOUR_TOKEN_ID
X-Bloo-Token-Secret: YOUR_TOKEN_SECRET
X-Bloo-Company-ID: YOUR_COMPANY_ID
X-Bloo-Project-ID: project_123
```

Invite one person to a single workspace as a member:

```graphql
mutation InviteUser {
  inviteUser(
    input: { email: "teammate@example.com", projectId: "project_123", accessLevel: MEMBER }
  )
}
```

You must supply **either** `projectId` **or** `companyId`. Supplying neither throws `BAD_USER_INPUT`.

## Parameters

### InviteUserInput

| Parameter     | Type               | Required | Description                                                                                                         |
| ------------- | ------------------ | -------- | ------------------------------------------------------------------------------------------------------------------- |
| `email`       | `String!`          | Yes      | Email address of the person to invite. Normalized before lookup.                                                    |
| `accessLevel` | `UserAccessLevel!` | Yes      | Access level to grant. See the values below.                                                                        |
| `projectId`   | `String`           | No       | ID or slug of a single workspace to invite into. Mutually exclusive with `companyId`.                               |
| `projectIds`  | `[String!]`        | No       | IDs or slugs of the workspaces to grant access to. Only meaningful alongside `companyId`.                           |
| `companyId`   | `String`           | No       | ID of the organization to invite into. Mutually exclusive with `projectId`. Caller must be an organization `OWNER`. |
| `roleId`      | `String`           | No       | ID of a custom role to assign instead of the plain access level. Must belong to the target workspace.               |

At least one of `projectId` or `companyId` is required.

### UserAccessLevel

| Value          | Description                                                                 |
| -------------- | --------------------------------------------------------------------------- |
| `OWNER`        | Full control of the workspace. Only a workspace `OWNER` may assign `OWNER`. |
| `ADMIN`        | Manage members, settings, and content.                                      |
| `MEMBER`       | Standard access — create and edit records.                                  |
| `CLIENT`       | Limited external-collaborator access.                                       |
| `COMMENT_ONLY` | View records and add comments only.                                         |
| `VIEW_ONLY`    | Read-only access.                                                           |

## Response

```json
{
  "data": {
    "inviteUser": true
  }
}
```

`true` means the invitation was created (or an existing pending invitation was re-sent), or — for an already-registered organization member — that they were added to the workspace directly.

## Workspace invites

Pass `projectId` to invite into a single workspace. The caller must have invite permission in that workspace (`VIEW_ONLY` and `COMMENT_ONLY` members cannot invite), and the access level they assign is capped by their own role — see [Permissions](#permissions).

If the invitee is already a registered Blue user who belongs to the workspace's organization, they are added to the workspace right away with the requested access level. If they are an organization `OWNER`, they are granted `ADMIN` in that workspace regardless of the `accessLevel` you request. Everyone else receives an invitation email and joins on acceptance.

## Organization invites

Pass `companyId` to invite into an organization. Only an organization `OWNER` can do this — otherwise the call throws `FORBIDDEN`. Use `projectIds` to list the workspaces the invitee should be added to:

```graphql
mutation InviteUserToOrganization {
  inviteUser(
    input: {
      email: "contractor@example.com"
      companyId: "company_123"
      projectIds: ["project_123", "project_456"]
      accessLevel: MEMBER
    }
  )
}
```

Each listed workspace produces its own invitation, consolidated into a single email. Omitting `projectIds` adds the invitee to **zero** workspaces — a brand-new invitee gains no membership until they are invited into at least one workspace, so include the workspaces you want them in.

## Custom roles

To assign a custom role instead of a built-in access level, set `accessLevel: MEMBER` and pass the role's `roleId`. The role must belong to the target workspace, and the invitee inherits the permissions defined on it. Retrieve available roles with the [`projectUserRoles` query](/api/user-management/retrieve-custom-role).

```graphql
mutation InviteUserWithRole {
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

## Re-inviting

Calling `inviteUser` again for an email that already has a pending invitation does **not** error. Blue reuses the existing invitation and re-sends it, so links in earlier emails stay valid. Invitations expire 7 days after they are sent.

## Errors

| Code                          | When                                                                                                                                                                                              |
| ----------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `BAD_USER_INPUT`              | Neither `projectId` nor `companyId` was supplied (`projectId or companyId is required`), or the email is missing/invalid.                                                                         |
| `FORBIDDEN`                   | Your role can't assign the requested access level, you lack invite permission in the workspace, or you used `companyId` without being an organization `OWNER`. Message: `You are not authorized.` |
| `ADD_SELF`                    | You tried to invite your own email. Message: `You are not allowed to add yourself.`                                                                                                               |
| `USER_ALREADY_IN_THE_PROJECT` | The invitee is already a member of the workspace. Message: `User is already in the project.`                                                                                                      |
| `PROJECT_NOT_FOUND`           | No workspace matches `projectId`. Message: `Project was not found.`                                                                                                                               |
| `COMPANY_NOT_FOUND`           | No organization matches `companyId`. Message: `Company was not found.`                                                                                                                            |
| `PROJECT_USER_ROLE_NOT_FOUND` | `roleId` doesn't match a role in the target workspace. Message: `Project user role was not found.`                                                                                                |
| `PLAN_LIMIT_REACHED`          | The organization is at its seat limit.                                                                                                                                                            |
| `COMPANY_BANNED`              | The organization is banned. Message: `Company is banned`.                                                                                                                                         |

A seat-limit error includes the limit context in `extensions`:

```json
{
  "errors": [
    {
      "message": "Users limit reached (5/5). Upgrade your plan for more.",
      "extensions": {
        "code": "PLAN_LIMIT_REACHED",
        "kind": "over_limit",
        "resource": "Users",
        "current": 5,
        "limit": 5
      }
    }
  ]
}
```

## Permissions

For workspace invites, the access level you can assign is capped by your own role in that workspace:

| Your role      | Can assign                                                        |
| -------------- | ----------------------------------------------------------------- |
| `OWNER`        | `OWNER`, `ADMIN`, `MEMBER`, `CLIENT`, `COMMENT_ONLY`, `VIEW_ONLY` |
| `ADMIN`        | `ADMIN`, `MEMBER`, `CLIENT`, `COMMENT_ONLY`, `VIEW_ONLY`          |
| `MEMBER`       | `MEMBER`, `CLIENT`, `COMMENT_ONLY`, `VIEW_ONLY`                   |
| `CLIENT`       | `CLIENT`                                                          |
| `COMMENT_ONLY` | — (cannot invite)                                                 |
| `VIEW_ONLY`    | — (cannot invite)                                                 |

Only a workspace `OWNER` may assign `OWNER`. Organization invites (`companyId`) require the caller to be an organization `OWNER`.

## Related

- [List users](/api/user-management/list-users) — see who already has access.
- [Remove users](/api/user-management/remove-user) — revoke workspace or organization access.
- [Custom roles](/api/user-management/retrieve-custom-role) — list the roles you can pass as `roleId`.

Manage the invitation that an invite produces with the `resendInvitation` and `cancelInvitation` mutations.
