---
title: User Presence
description: Stream who is online and live organization membership — presence status, company changes, and project invitations over the realtime WebSocket.
icon: UserRoundCheck
order: 9
---

Know who is online and keep organization membership in sync as it changes. This page covers four subscriptions: `subscribeToUserPresence` (a user coming online or going offline), `subscribeToCompany` (organization record changes), `subscribeToCompanyPeopleList` (the org's member roster), and `subscribeToInvitation` (workspace invitations created or revoked). Organizations are `Organization` objects in the API and workspaces are `Workspace` objects.

All four run over the same authenticated WebSocket as every other subscription — see [Connect & Authenticate](/api/realtime/connect-and-authenticate) for the `graphql-ws` handshake and credentials. The endpoint is `wss://api.blue.app/graphql`.

`subscribeToUserPresence` returns a bare `User!` (no payload wrapper), because presence is a status snapshot rather than an entity change-feed. `subscribeToCompany` and `subscribeToInvitation` return the standard payload (`{ mutation, node, previousValues, updatedFields }`) described on the [Real-time overview](/api/realtime), where `mutation` is a `MutationType` (`CREATED`, `UPDATED`, or `DELETED`). Note: `subscribeToCompany` streams changes to the underlying `Organization` record.

## subscribeToUserPresence

Streams online/offline transitions for the people you share an organization with. An event fires when a user **comes online** (their first WebSocket connection opens) or **goes offline** (their last connection closes). Read `isOnline` on the delivered `User` to tell which way the transition went.

Presence is wired directly to the WebSocket connection lifecycle. When a client completes the `onConnect` handshake the server runs `handleUserConnect`; when the socket closes it runs `handleUserDisconnect`. A user is "online" while they hold at least one open connection, so opening a second tab does not re-fire an online event and closing one of two tabs does not fire an offline event — only the first connect and the last disconnect publish. Presence keys carry a 5-minute TTL, so a connection that dies without a clean close is reaped shortly after.

This replaces the deprecated `subscribeToNetworkStatus`.

### Request

```graphql
subscription OnUserPresenceChanged {
  subscribeToUserPresence(filter: { companyId: "company_123" }) {
    id
    fullName
    email
    isOnline
    lastActiveAt
    image {
      medium
    }
  }
}
```

Omit `filter` to watch presence across **every** organization you belong to. Pass `filter.companyId` to narrow to one organization — you only receive events for users who are members of that organization alongside you.

### Parameters

#### UserPresenceFilter

| Parameter   | Type     | Required | Description                                                                                                                             |
| ----------- | -------- | -------- | --------------------------------------------------------------------------------------------------------------------------------------- |
| `companyId` | `String` | No       | Restrict events to one organization. When omitted, presence for users in any shared organization is delivered. Accepts an ID or a slug. |

### Response

Each event is a `User`. `isOnline` reflects the transition that triggered the event — `true` when the user just came online, `false` when they just went offline.

```json
{
  "data": {
    "subscribeToUserPresence": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "fullName": "Ada Lovelace",
      "email": "ada@example.com",
      "isOnline": true,
      "lastActiveAt": "2026-05-29T14:02:11.000Z",
      "image": {
        "medium": "https://blue.app/files/clm4n8qwx000008l0g4oxdqn7/medium.png"
      }
    }
  }
}
```

#### Returns: `User`

| Field          | Type       | Description                                                              |
| -------------- | ---------- | ------------------------------------------------------------------------ |
| `id`           | `ID!`      | The user whose presence changed.                                         |
| `fullName`     | `String!`  | Display name. (`firstName` / `lastName` are also available.)             |
| `email`        | `String!`  | The user's email address.                                                |
| `isOnline`     | `Boolean!` | `true` if this event marks them coming online, `false` if going offline. |
| `lastActiveAt` | `DateTime` | When the user was last active.                                           |
| `image`        | `Image`    | Avatar; select a size such as `thumbnail`, `medium`, or `original`.      |

`User` exposes many more fields — select only what your presence UI needs. See [User Management](/api/user-management) for the full type.

<Callout variant="info" title="Use this, not subscribeToNetworkStatus">

`subscribeToNetworkStatus` is deprecated and exists only for backward compatibility. Build presence on `subscribeToUserPresence`.

</Callout>

## subscribeToCompany

Streams changes to the organization record itself — for example a rename, a plan or settings change, or other `Organization` field edits. Returns the standard entity payload.

### Request

```graphql
subscription OnCompanyChanged {
  subscribeToCompany(filter: { companyId: "company_123" }) {
    mutation
    node {
      id
      name
      slug
    }
    updatedFields
    previousValues {
      name
    }
  }
}
```

Pass `filter.companyId` to watch a single organization. With no `companyId`, you receive change events for every organization you are a member of.

### Parameters

#### CompanyFilter

| Parameter     | Type                  | Required | Description                                                   |
| ------------- | --------------------- | -------- | ------------------------------------------------------------- |
| `companyId`   | `String`              | No       | Restrict events to one organization. Accepts an ID or a slug. |
| `active`      | `Boolean`             | No       | Filter to active organizations.                               |
| `search`      | `String`              | No       | Free-text match on organization name.                         |
| `status`      | `CompanyFilterStatus` | No       | Filter by subscription status.                                |
| `activatable` | `Boolean`             | No       | Organizations that are applicable for license activation.     |

#### CompanyFilterStatus

`ACTIVE`, `CANCELED`, `EXPIRED`, `TRIALING`, `PAST_DUE`, `UNKNOWN`.

### Response

Each event is a `CompanySubscriptionPayload`.

```json
{
  "data": {
    "subscribeToCompany": {
      "mutation": "UPDATED",
      "node": {
        "id": "clm4n8qwx000008l0g4oxdqn7",
        "name": "Acme, Inc.",
        "slug": "acme"
      },
      "updatedFields": ["name"],
      "previousValues": {
        "name": "Acme Co."
      }
    }
  }
}
```

#### CompanySubscriptionPayload

| Field            | Type                         | Description                                                              |
| ---------------- | ---------------------------- | ------------------------------------------------------------------------ |
| `mutation`       | `MutationType!`              | `CREATED`, `UPDATED`, or `DELETED`.                                      |
| `node`           | `Organization`               | The organization in its current state. `null` on `DELETED`.              |
| `updatedFields`  | `[String!]`                  | Names of the organization properties that changed on an `UPDATED` event. |
| `previousValues` | `OrganizationPreviousValues` | Snapshot of the organization before the change.                          |

## subscribeToCompanyPeopleList

Notifies a client when the organization's member roster changes — someone added, removed, or having their role updated — so a "People" view can refresh. Scoped to a single organization by `companyId`.

### Request

```graphql
subscription OnCompanyPeopleListChanged {
  subscribeToCompanyPeopleList(companyId: "company_123") {
    mutation
    node {
      id
      fullName
      email
    }
  }
}
```

### Parameters

| Parameter   | Type      | Required | Description                                                |
| ----------- | --------- | -------- | ---------------------------------------------------------- |
| `companyId` | `String!` | Yes      | The organization whose member roster to watch. ID or slug. |

<Callout variant="warning" title="Treat as a refresh signal">

`subscribeToCompanyPeopleList` is delivered as a notification that the roster changed; the payload `node` is not currently populated with the changed user. When you receive an event, re-fetch the organization's people list (see [User Management](/api/user-management)) rather than relying on the event body. The return type is `UserSubscriptionPayload`.

</Callout>

#### UserSubscriptionPayload

| Field            | Type                 | Description                                |
| ---------------- | -------------------- | ------------------------------------------ |
| `mutation`       | `MutationType!`      | `CREATED`, `UPDATED`, or `DELETED`.        |
| `node`           | `User`               | The affected user.                         |
| `updatedFields`  | `[String!]`          | Names of the user properties that changed. |
| `previousValues` | `UserPreviousValues` | Snapshot of the user before the change.    |

## subscribeToInvitation

Streams workspace invitations being **created** or **revoked**. Invitations are `Invitation` objects. There are two ways to use it:

- **Watch one workspace** — pass `projectId` to receive every invitation created or deleted for that workspace (an admin view of pending invites).
- **Watch your own invitations** — omit `projectId` to receive only invitations addressed to _your_ email, so a logged-in user is notified the moment they're invited somewhere (matched on `node.email` for `CREATED` and `previousValues.email` for `DELETED`).

### Request

```graphql
subscription OnInvitationChanged {
  subscribeToInvitation(projectId: "project_123") {
    mutation
    node {
      id
      email
      accessLevel
      invitedBy {
        fullName
      }
      project {
        id
        name
      }
      expiredAt
    }
    previousValues {
      email
    }
  }
}
```

### Parameters

| Parameter   | Type     | Required | Description                                                                                                      |
| ----------- | -------- | -------- | ---------------------------------------------------------------------------------------------------------------- |
| `projectId` | `String` | No       | Watch all invitations for one workspace. When omitted, you receive only invitations addressed to your own email. |

### Response

Each event is an `InvitationSubscriptionPayload`. On a `CREATED` event `node` carries the new invitation; on `DELETED` (the invite was revoked or accepted) inspect `previousValues`.

```json
{
  "data": {
    "subscribeToInvitation": {
      "mutation": "CREATED",
      "node": {
        "id": "clm4n8qwx000008l0g4oxdqn7",
        "email": "newhire@example.com",
        "accessLevel": "MEMBER",
        "invitedBy": {
          "fullName": "Ada Lovelace"
        },
        "project": {
          "id": "clm4n8qwx000108l0h2pyerm8",
          "name": "Q3 Roadmap"
        },
        "expiredAt": "2026-06-05T14:02:11.000Z"
      },
      "previousValues": null
    }
  }
}
```

#### InvitationSubscriptionPayload

| Field            | Type                       | Description                                                |
| ---------------- | -------------------------- | ---------------------------------------------------------- |
| `mutation`       | `MutationType!`            | `CREATED` (invited) or `DELETED` (revoked/accepted).       |
| `node`           | `Invitation`               | The invitation. Present on `CREATED`; `null` on `DELETED`. |
| `updatedFields`  | `[String!]`                | Changed properties on an update.                           |
| `previousValues` | `InvitationPreviousValues` | Snapshot before the change. Carries `email` on `DELETED`.  |

#### Invitation

| Field         | Type              | Description                                    |
| ------------- | ----------------- | ---------------------------------------------- |
| `id`          | `ID!`             | The invitation ID.                             |
| `email`       | `String!`         | The invited person's email address.            |
| `accessLevel` | `UserAccessLevel` | The role the invitee will receive.             |
| `invitedBy`   | `User!`           | The user who sent the invitation.              |
| `project`     | `Workspace!`      | The workspace the invitation grants access to. |
| `expiredAt`   | `DateTime!`       | When the invitation expires.                   |
| `metadata`    | `JSON`            | Arbitrary invitation metadata.                 |
| `createdAt`   | `DateTime!`       | When the invitation was created.               |

## Errors

Subscriptions report invalid arguments and missing scopes when the operation is set up; permission scoping is otherwise silent (see Permissions below).

| Code                | When                                                                                                                                                                                   |
| ------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `UNAUTHENTICATED`   | The WebSocket handshake carried no valid credentials. The connection is rejected before any subscription starts. See [Connect & Authenticate](/api/realtime/connect-and-authenticate). |
| `COMPANY_NOT_FOUND` | A `companyId` passed to `subscribeToUserPresence`, `subscribeToCompany`, or `subscribeToCompanyPeopleList` matches no organization you can access.                                     |
| `PROJECT_NOT_FOUND` | The `projectId` passed to `subscribeToInvitation` matches no workspace you can access.                                                                                                 |

## Permissions

Events are filtered per delivered event, not just at connection time. Subscribing always succeeds for an authenticated client, but you only receive events you are allowed to see:

- **`subscribeToUserPresence`** delivers a presence event only for a user who shares an organization with you. With `filter.companyId`, the user must be a member of _that_ organization (and so must you); without a filter, any shared organization qualifies.
- **`subscribeToCompany`** delivers an event only for an organization you belong to. With `filter.companyId`, the event must also match that organization.
- **`subscribeToInvitation`** with `projectId` delivers events for that workspace; without `projectId` it delivers only invitations addressed to your own email.

## Related

- [Real-time (Subscriptions)](/api/realtime) — the shared payload shape and the full subscription catalog.
- [Connect & Authenticate](/api/realtime/connect-and-authenticate) — open an authenticated WebSocket.
- [Project & Workspace Subscriptions](/api/realtime/project-subscriptions) — membership, roles, and workspace lifecycle streams.
- [User Management](/api/user-management) — invite and remove members, list the people in an organization.
