---
title: Status Updates
description: Post, delete, query, and subscribe to project status updates — a colored health category plus rich-text commentary, dated and attributed to a user.
icon: MessageSquarePlus
order: 0
---

Status updates are project-level health check-ins: a traffic-light category — `GREEN`, `ORANGE`, or `RED` — paired with rich-text commentary, stamped with a reporting date, and attributed to the user who posted it. They give a workspace a running record of "how is this going?" alongside its records. This section covers posting and deleting status updates, fetching a single update or a paginated, filterable list for a project, and subscribing to real-time create and delete events.

Status updates always belong to a [workspace](/api/workspaces) (a `Project` in the API). A project can hide the feature entirely via its `hideStatusUpdate` setting, but that's a UI toggle — it does not block the API operations described here.

<Callout variant="warning" title="Status updates are immutable">

There is no edit operation in the public API. Once a status update is posted, its category and text are fixed — to change one, delete it with `deleteStatusUpdate` and post a fresh one with `createStatusUpdate`.

</Callout>

## The StatusUpdate type

`StatusUpdate` is the shape returned by `createStatusUpdate`, the `statusUpdate` query, and every element of `StatusUpdateList.statusUpdates`.

| Field          | Type                    | Description                                                                                      |
| -------------- | ----------------------- | ------------------------------------------------------------------------------------------------ |
| `id`           | `ID!`                   | Unique identifier for the status update.                                                         |
| `html`         | `String!`               | The commentary as sanitized rich-text HTML.                                                      |
| `text`         | `String!`               | The commentary as plain text (the fallback for the same body as `html`).                         |
| `date`         | `DateTime!`             | The reporting date the update is _for_. Distinct from `createdAt` — the UI lets you backdate it. |
| `category`     | `StatusUpdateCategory!` | The health signal: `GREEN`, `ORANGE`, or `RED`.                                                  |
| `createdAt`    | `DateTime!`             | When the update was posted.                                                                      |
| `updatedAt`    | `DateTime!`             | When the row was last touched.                                                                   |
| `user`         | `User!`                 | The author. Select `id`, `fullName`, `email`.                                                    |
| `project`      | `Project!`              | The workspace this update belongs to.                                                            |
| `commentCount` | `Int!`                  | Number of comments on this update.                                                               |
| `isRead`       | `Boolean`               | Whether the calling user has read this update.                                                   |
| `isSeen`       | `Boolean`               | Whether the calling user has seen this update in the feed.                                       |

Comments attach to a status update through the polymorphic `Comment` model (`category: STATUS_UPDATE`, `categoryId` set to the update's `id`) — read and write them with the [comment operations](/api/comments), and read the count here via `commentCount`.

## The StatusUpdateCategory enum

`StatusUpdateCategory` is the traffic-light health signal carried by every status update. It is set at creation and never changes.

| Value    | Meaning                    |
| -------- | -------------------------- |
| `GREEN`  | On track / healthy.        |
| `ORANGE` | At risk / needs attention. |
| `RED`    | Off track / blocked.       |

## Pages in this section

| Page                                                                            | What it covers                                                                                                                                                                                                |
| ------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| [Post & delete status updates](/api/status-updates/manage-status-updates)       | `createStatusUpdate` (with `CreateStatusUpdateInput`) and the `deleteStatusUpdate` mutation, plus the auth rules and the immutability gotcha.                                                                 |
| [Query & subscribe to status updates](/api/status-updates/query-status-updates) | Fetch one with `statusUpdate(id)`, list a project's updates with `statusUpdateList` (filtering, ordering, cursor pagination), and subscribe to real-time create/delete events with `subscribeToStatusUpdate`. |

## Product nouns to schema types

Use the UI word in prose and the schema name in code. The mapping for this section:

| UI / docs word      | Schema type / field                                 |
| ------------------- | --------------------------------------------------- |
| Status update       | `StatusUpdate`                                      |
| Health category     | `StatusUpdateCategory` (`GREEN` / `ORANGE` / `RED`) |
| Workspace / Project | `Project`                                           |
| Author              | `User`                                              |
| Update list page    | `StatusUpdateList`                                  |

## Related

- [Post & delete status updates](/api/status-updates/manage-status-updates)
- [Query & subscribe to status updates](/api/status-updates/query-status-updates)
- [Comments & Discussions](/api/comments)
- [Workspaces](/api/workspaces)
