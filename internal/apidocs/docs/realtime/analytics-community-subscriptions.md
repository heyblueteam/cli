---
title: Dashboards, Charts & Community
description: Live streams for dashboards, charts, form changes, and the community feed — the remaining real-time channels in one place.
icon: ChartBar
order: 10
---

The remaining real-time channels, grouped because each is low-traffic on its own: live **dashboards** and their **charts** (`subscribeToDashboard`, `subscribeToChart`), **form** changes (`subscribeToForm`), and the public **community** feed (`subscribeToCommunityPost` for a post's replies, `subscribeToCommunityPosts` for the post feed).

All run over the same authenticated WebSocket as every other subscription — see [Connect & Authenticate](/api/realtime/connect-and-authenticate) for the `graphql-ws` handshake and credentials. The endpoint is `wss://api.blue.app/graphql`, the same URL as the HTTP GraphQL API.

Dashboards (`Dashboard`), charts (`Chart`), and forms (`Form`) use the standard entity change-feed payload (`{ mutation, node, previousValues, … }`) described on the [Real-time overview](/api/realtime), where `mutation` is a `MutationType` (`CREATED`, `UPDATED`, `DELETED`). The two community channels are the exception: they use a distinct **`CommunityMutationType`** enum and a slimmer payload of just `{ mutation, node }` — no `previousValues`, no `updatedFields`. Both enums share the same `CREATED` / `UPDATED` / `DELETED` values; only the type name differs.

## subscribeToDashboard

Streams dashboard lifecycle changes in one organization — a dashboard being created, renamed, shared, or deleted. Organizations are `Company` objects in the API, so the stream is scoped by `companyId`, with an optional `projectId` to narrow it to one workspace.

### Request

```graphql
subscription OnDashboardChanged {
  subscribeToDashboard(filter: { companyId: "company_123" }) {
    mutation
    node {
      id
      title
      createdBy {
        id
        fullName
      }
    }
    previousValues {
      title
    }
  }
}
```

### Parameters

#### SubscribeToDashboardFilterInput

| Parameter   | Type      | Required | Description                                                              |
| ----------- | --------- | -------- | ------------------------------------------------------------------------ |
| `companyId` | `String!` | Yes      | Organization whose dashboards to watch. Accepts an ID or a slug.         |
| `projectId` | `String`  | No       | Restrict to dashboards scoped to one workspace. Accepts an ID or a slug. |

### Response

Each delivered event is a `DashboardSubscriptionPayload`. The example below shows a dashboard that was renamed.

```json
{
  "data": {
    "subscribeToDashboard": {
      "mutation": "UPDATED",
      "node": {
        "id": "clm4n8qwx000008l0g4oxdqn7",
        "title": "Q3 Revenue",
        "createdBy": {
          "id": "clm4n8qwx000108l0h2pyerm8",
          "fullName": "Dana Lee"
        }
      },
      "previousValues": {
        "title": "Q3 Dashboard"
      }
    }
  }
}
```

#### DashboardSubscriptionPayload

| Field            | Type            | Description                                              |
| ---------------- | --------------- | -------------------------------------------------------- |
| `mutation`       | `MutationType!` | `CREATED`, `UPDATED`, or `DELETED`.                      |
| `node`           | `Dashboard`     | The dashboard in its current state. `null` on `DELETED`. |
| `previousValues` | `Dashboard`     | The dashboard before the change. Present on `UPDATED`.   |

`node` and `previousValues` are full `Dashboard` objects — select `title`, `createdBy { id fullName }`, `dashboardUsers { user { id fullName } role }`, `createdAt`, and `updatedAt` as needed.

## subscribeToChart

Streams changes to the charts inside one dashboard — a chart being added, edited, recalculated, repositioned, or removed. Scoped by `dashboardId`.

If your client renders a chart with a record filter applied (the `todoFilter` argument), pass the **same** filter to the subscription so you only receive recomputed values for that exact filtered view. The server matches your `todoFilter` against the published event's filter by value: pass it to receive filtered events, omit it to receive unfiltered events. A mismatch (one side has a filter, the other does not) is dropped.

### Request

```graphql
subscription OnChartChanged {
  subscribeToChart(filter: { dashboardId: "dashboard_123" }) {
    mutation
    node {
      id
      title
      type
      isCalculating
      needCalculation
    }
    previousValues {
      title
    }
  }
}
```

### Parameters

#### SubscribeToChartFilterInput

| Parameter     | Type              | Required | Description                                                                                                     |
| ------------- | ----------------- | -------- | --------------------------------------------------------------------------------------------------------------- |
| `dashboardId` | `String!`         | Yes      | Dashboard whose charts to watch.                                                                                |
| `todoFilter`  | `TodoFilterInput` | No       | The record (todo) filter applied to the chart view. Must match the published event's filter exactly to deliver. |

### Response

Each delivered event is a `ChartSubscriptionPayload`. The example below shows a chart that finished recalculating after underlying records changed.

```json
{
  "data": {
    "subscribeToChart": {
      "mutation": "UPDATED",
      "node": {
        "id": "clm4n8qwx000008l0g4oxdqn7",
        "title": "Open by assignee",
        "type": "BAR",
        "isCalculating": false,
        "needCalculation": false
      },
      "previousValues": {
        "title": "Open by assignee"
      }
    }
  }
}
```

#### ChartSubscriptionPayload

| Field            | Type            | Description                                          |
| ---------------- | --------------- | ---------------------------------------------------- |
| `mutation`       | `MutationType!` | `CREATED`, `UPDATED`, or `DELETED`.                  |
| `node`           | `Chart`         | The chart in its current state. `null` on `DELETED`. |
| `previousValues` | `Chart`         | The chart before the change. Present on `UPDATED`.   |

`node` and `previousValues` are full `Chart` objects. Useful fields to select: `title`, `type` (a `ChartType` — `STAT`, `PIE`, or `BAR`), `position`, `chartSegments`, `display`, `metadata`, and the recalculation flags `isCalculating`, `isCalculatingWithFilter`, and `needCalculation`.

<Callout variant="info" title="Recalculation is asynchronous">

When records that feed a chart change, the chart is queued for recalculation. You will typically see an `UPDATED` event with `needCalculation: true` or `isCalculating: true`, followed by another `UPDATED` event with both back to `false` once the new values are computed.

</Callout>

## subscribeToForm

Streams changes to forms — a form being created, edited, or deleted. The `filter` argument is optional; pass `companyId` and/or `projectId` to document the intended scope. Forms map to `Form` objects in the API.

### Request

```graphql
subscription OnFormChanged {
  subscribeToForm(filter: { projectId: "project_123" }) {
    mutation
    node {
      id
      title
      isActive
    }
    updatedFields
    previousValues {
      title
      isActive
    }
  }
}
```

### Parameters

#### SubscribeToFormFilterInput

| Parameter   | Type     | Required | Description                                  |
| ----------- | -------- | -------- | -------------------------------------------- |
| `companyId` | `String` | No       | Organization scope. Accepts an ID or a slug. |
| `projectId` | `String` | No       | Workspace scope. Accepts an ID or a slug.    |

### Response

Each delivered event is a `FormSubscriptionPayload`. The example below shows a form that was deactivated — `updatedFields` lists the changed property and `previousValues` carries the pre-change snapshot.

```json
{
  "data": {
    "subscribeToForm": {
      "mutation": "UPDATED",
      "node": {
        "id": "clm4n8qwx000008l0g4oxdqn7",
        "title": "Bug report",
        "isActive": false
      },
      "updatedFields": ["isActive"],
      "previousValues": {
        "title": "Bug report",
        "isActive": true
      }
    }
  }
}
```

#### FormSubscriptionPayload

| Field            | Type                 | Description                                                      |
| ---------------- | -------------------- | ---------------------------------------------------------------- |
| `mutation`       | `MutationType!`      | `CREATED`, `UPDATED`, or `DELETED`.                              |
| `node`           | `Form`               | The form in its current state. `null` on `DELETED`.              |
| `updatedFields`  | `[String!]`          | Names of the form properties that changed on an `UPDATED` event. |
| `previousValues` | `FormPreviousValues` | Snapshot of the form before the change. Present on `UPDATED`.    |

`node` is the full `Form` type (select `title`, `description`, `isActive`, `theme`, `primaryColor`, `hideBranding`, `formFields { id name required }`, `todoList { id title }`, and any other form property you track). `previousValues` is the narrower `FormPreviousValues` with the core scalars only: `id`, `uid`, `title`, `description`, `isActive`, `theme`, `primaryColor`, `hideBranding`, `responseText`, `submitText`, `imageURL`, `redirectURL`, `snapshotURL`, `createdAt`, `updatedAt`.

## subscribeToCommunityPosts

Streams the public community feed — a top-level post being created, edited, or deleted across every category. Posts are `CommunityPost` objects. This channel takes no arguments; it watches the whole feed.

### Request

```graphql
subscription OnCommunityPostChanged {
  subscribeToCommunityPosts {
    mutation
    node {
      id
      slug
      title
      category
      repliesCount
      author {
        id
        fullName
      }
    }
  }
}
```

### Parameters

This subscription takes no arguments.

### Response

Each delivered event is a `CommunityPostSubscriptionPayload`. Community payloads use the **`CommunityMutationType`** enum and carry only `mutation` and `node` — there is no `previousValues` and no `updatedFields`.

```json
{
  "data": {
    "subscribeToCommunityPosts": {
      "mutation": "CREATED",
      "node": {
        "id": "clm4n8qwx000008l0g4oxdqn7",
        "slug": "dark-mode-please",
        "title": "Dark mode, please",
        "category": "FEATURE_REQUESTS",
        "repliesCount": 0,
        "author": {
          "id": "clm4n8qwx000108l0h2pyerm8",
          "fullName": "Dana Lee"
        }
      }
    }
  }
}
```

#### CommunityPostSubscriptionPayload

| Field      | Type                     | Description                                         |
| ---------- | ------------------------ | --------------------------------------------------- |
| `mutation` | `CommunityMutationType!` | `CREATED`, `UPDATED`, or `DELETED`. Always present. |
| `node`     | `CommunityPost`          | The post in its current state. `null` on `DELETED`. |

`node` is the full `CommunityPost` type. Useful fields: `slug`, `title`, `html`, `text`, `category` (a `CommunityCategory` — `GENERAL`, `BUGS`, `FEATURE_REQUESTS`, or `ANNOUNCEMENTS`), `isPinned`, `isLocked`, `repliesCount`, `author { id fullName }`, `displayName`, `reactions { type count }`, and `viewerReactions`.

## subscribeToCommunityPost

Streams the **replies** to one community post — a reply being added, edited, or deleted under the post you name. Use this alongside `subscribeToCommunityPosts`: subscribe to the feed for new posts, and subscribe to a post for the live reply thread when a reader opens it. Replies are `CommunityReply` objects, scoped by `postId`.

### Request

```graphql
subscription OnCommunityReply {
  subscribeToCommunityPost(postId: "post_123") {
    mutation
    node {
      id
      html
      text
      parentId
      author {
        id
        fullName
      }
    }
  }
}
```

### Parameters

| Parameter | Type      | Required | Description                                |
| --------- | --------- | -------- | ------------------------------------------ |
| `postId`  | `String!` | Yes      | The community post whose replies to watch. |

### Response

Each delivered event is a `CommunityReplySubscriptionPayload` — the same slim `{ mutation, node }` shape as the post feed, with `mutation` a `CommunityMutationType`.

```json
{
  "data": {
    "subscribeToCommunityPost": {
      "mutation": "CREATED",
      "node": {
        "id": "clm4n8qwx000208l0i9qzfsn9",
        "html": "<p>+1, this would be huge.</p>",
        "text": "+1, this would be huge.",
        "parentId": null,
        "author": {
          "id": "clm4n8qwx000108l0h2pyerm8",
          "fullName": "Dana Lee"
        }
      }
    }
  }
}
```

#### CommunityReplySubscriptionPayload

| Field      | Type                     | Description                                          |
| ---------- | ------------------------ | ---------------------------------------------------- |
| `mutation` | `CommunityMutationType!` | `CREATED`, `UPDATED`, or `DELETED`. Always present.  |
| `node`     | `CommunityReply`         | The reply in its current state. `null` on `DELETED`. |

`node` is the full `CommunityReply` type. Useful fields: `html`, `text`, `replyCount`, `parentId` (set when the reply is nested under another reply), `author { id fullName }`, `displayName`, `reactions { type count }`, and `viewerReactions`.

## Errors

Subscriptions report invalid arguments and missing scopes when the operation is set up. Per-event scoping is otherwise silent — see Permissions below.

| Code              | When                                                                                                                                                                                   |
| ----------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `UNAUTHENTICATED` | The WebSocket handshake carried no valid credentials. The connection is rejected before any subscription starts. See [Connect & Authenticate](/api/realtime/connect-and-authenticate). |
| `BAD_USER_INPUT`  | A required filter field is missing — `companyId` on `subscribeToDashboard`, `dashboardId` on `subscribeToChart`, or `postId` on `subscribeToCommunityPost`.                            |

## Permissions

Events are filtered per delivered event, not just at connection time. Subscribing always succeeds for an authenticated client, but you only receive events you are allowed to see:

- **`subscribeToDashboard`** delivers an event only when the dashboard belongs to the `companyId` you named **and** you either created it or are one of its `dashboardUsers`.
- **`subscribeToChart`** delivers an event only for charts in a dashboard you created or are a member of, and only when your `todoFilter` matches the published event's filter.
- **`subscribeToForm`** delivers form events for the connection's scope.
- **`subscribeToCommunityPosts`** and **`subscribeToCommunityPost`** stream the public community feed; post events are global and reply events are filtered to the `postId` you subscribed with.

## Related

- [Real-time (Subscriptions)](/api/realtime) — the shared payload shape and the full subscription catalog.
- [Connect & Authenticate](/api/realtime/connect-and-authenticate) — open an authenticated WebSocket.
- [File & Link Subscriptions](/api/realtime/file-subscriptions) — keep attachment and link lists live.
- [Activity, Mentions & Notifications](/api/realtime/activity-notifications) — the notification firehose.
