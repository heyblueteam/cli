---
title: Progress & Long-Running Jobs
description: Stream progress on async server jobs — imports/exports, lookups, AI tagging, cover generation, automation runs, and OAuth connections.
icon: LoaderCircle
order: 9
---

Some Blue operations run asynchronously on the server — bulk imports and exports, reference-field lookups, AI auto-tagging, cover-image generation, automation executions. These subscriptions push **job status** to your client as the work proceeds, so you can render a progress bar or refresh a view the moment a job finishes instead of polling.

These are streamed over WebSocket using the [graphql-ws](https://github.com/enisdenjo/graphql-ws) protocol at `wss://api.blue.app/graphql` — the same path as the HTTP GraphQL endpoint. Open and authenticate the connection first; see [Connect & Authenticate](/api/realtime/connect-and-authenticate) for the handshake and credential forms. The examples below show only the subscription documents.

Unlike the entity change-feeds (records, comments, files, …) these are **job-status channels**, so their payloads vary. Some return a free-form `JSON` blob (`subscribeToImportExportProgress`, `subscribeToLookupProgress`), some a typed progress object (`subscribeToAITagProgress`, `subscribeToCoverGenerationProgress`), and the automation/OAuth streams reuse the standard `{ mutation, node, previousValues }` shape described on the [Real-time overview](/api/realtime). Each section below states which.

## subscribeToImportExportProgress

Track a bulk import or export as it runs. Blue runs these as background jobs; this stream reports their lifecycle so you can show a spinner and reload the workspace when the job is `DONE`.

### Request

```graphql
subscription OnImportExportProgress {
  subscribeToImportExportProgress(projectId: "project_123", userId: "user_123")
}
```

### Parameters

| Parameter   | Type      | Required | Description                                                                                                                |
| ----------- | --------- | -------- | -------------------------------------------------------------------------------------------------------------------------- |
| `projectId` | `String!` | Yes      | Workspace ID the import/export targets.                                                                                    |
| `userId`    | `String!` | Yes      | The user running the job. Events are delivered only when this matches the authenticated user — pass your own user ID here. |

### Response

This field returns `JSON` — there is no sub-selection. Each event is an object of the form `{ status: 'IN_PROGRESS' | 'DONE' | 'ERROR' }`.

```json
{
  "data": {
    "subscribeToImportExportProgress": { "status": "IN_PROGRESS" }
  }
}
```

A terminal `DONE` (or `ERROR`) event is the signal to stop your spinner and refetch:

```json
{
  "data": {
    "subscribeToImportExportProgress": { "status": "DONE" }
  }
}
```

| Status        | Meaning                                          |
| ------------- | ------------------------------------------------ |
| `IN_PROGRESS` | The import or export is still running.           |
| `DONE`        | The job finished successfully.                   |
| `ERROR`       | The job failed; nothing further will be emitted. |

## subscribeToLookupProgress

Track progress while Blue resolves a reference (lookup) field across a workspace — for example, recomputing linked-record values after a field or list changes. Scope it to a single workspace with the `filter`.

### Request

```graphql
subscription OnLookupProgress {
  subscribeToLookupProgress(filter: { projectId: "project_123" })
}
```

### Parameters

#### LookupProgressFilter

| Parameter   | Type      | Required | Description                                                                |
| ----------- | --------- | -------- | -------------------------------------------------------------------------- |
| `projectId` | `String!` | Yes      | Workspace ID. Events are delivered only for lookup jobs in this workspace. |

### Response

This field returns `JSON` — there is no sub-selection. The payload is the progress object published by the server for the matching workspace.

```json
{
  "data": {
    "subscribeToLookupProgress": {
      "projectId": "project_123",
      "progress": 60,
      "processed": 120,
      "total": 200
    }
  }
}
```

<Callout variant="info" title="JSON payload, not a typed object">

Because the field's type is `JSON`, the server controls the shape and you select no subfields. Read the keys defensively in your client rather than relying on a fixed schema.

</Callout>

## subscribeToAITagProgress

Track AI auto-tagging of records as it runs. When you trigger AI tagging for a workspace, the server processes records in the background and publishes progress to this stream.

### Request

```graphql
subscription OnAITagProgress {
  subscribeToAITagProgress(projectId: "project_123") {
    progress
    count
    maxCount
    operationId
    error
    errorCode
  }
}
```

### Parameters

| Parameter   | Type      | Required | Description                                                                                      |
| ----------- | --------- | -------- | ------------------------------------------------------------------------------------------------ |
| `projectId` | `String!` | Yes      | Workspace ID the tagging job runs in. Events are delivered only to the user who started the job. |

### Response

`subscribeToAITagProgress` returns an `AITagProgress` object per event.

```json
{
  "data": {
    "subscribeToAITagProgress": {
      "progress": 0.75,
      "count": 75,
      "maxCount": 100,
      "operationId": "clm4n8qwx000008l0g4oxdqn7",
      "error": false,
      "errorCode": null
    }
  }
}
```

#### AITagProgress

| Field         | Type      | Description                                                                     |
| ------------- | --------- | ------------------------------------------------------------------------------- |
| `progress`    | `Float`   | Fractional completion from `0` to `1`.                                          |
| `count`       | `Float`   | Number of records processed so far.                                             |
| `maxCount`    | `Float`   | Total number of records in the job.                                             |
| `operationId` | `String`  | Identifier for this tagging run; correlate it with the operation you triggered. |
| `error`       | `Boolean` | `true` if the job failed.                                                       |
| `errorCode`   | `String`  | A machine-readable error code when `error` is `true`; otherwise `null`.         |

## subscribeToCoverGenerationProgress

Track AI cover-image generation for records in a workspace. Scope it with the `filter`.

### Request

```graphql
subscription OnCoverGenerationProgress {
  subscribeToCoverGenerationProgress(filter: { projectId: "project_123" }) {
    mutation
    node {
      projectId
      progress
      processedCount
      totalCount
      status
    }
  }
}
```

### Parameters

#### CoverGenerationProgressFilter

| Parameter   | Type      | Required | Description                                                                          |
| ----------- | --------- | -------- | ------------------------------------------------------------------------------------ |
| `projectId` | `String!` | Yes      | Workspace ID. Events are delivered only for cover-generation jobs in this workspace. |

### Response

`subscribeToCoverGenerationProgress` returns a `CoverGenerationProgressPayload` per event.

```json
{
  "data": {
    "subscribeToCoverGenerationProgress": {
      "mutation": "UPDATED",
      "node": {
        "projectId": "project_123",
        "progress": 40,
        "processedCount": 4,
        "totalCount": 10,
        "status": "IN_PROGRESS"
      }
    }
  }
}
```

#### CoverGenerationProgressPayload

| Field      | Type                      | Description                                             |
| ---------- | ------------------------- | ------------------------------------------------------- |
| `mutation` | `MutationType!`           | The kind of change: `CREATED`, `UPDATED`, or `DELETED`. |
| `node`     | `CoverGenerationProgress` | The progress snapshot for this event.                   |

This payload has **no** `previousValues` or `updatedFields` — it is a progress feed, not an entity change-feed.

#### CoverGenerationProgress

| Field            | Type      | Description                                 |
| ---------------- | --------- | ------------------------------------------- |
| `projectId`      | `String!` | Workspace the cover-generation job runs in. |
| `progress`       | `Int!`    | Completion percentage from `0` to `100`.    |
| `processedCount` | `Int!`    | Number of records processed so far.         |
| `totalCount`     | `Int!`    | Total number of records in the job.         |
| `status`         | `String!` | Job status string published by the server.  |

## subscribeToAutomation

Stream create, update, and delete events for **automation definitions** in a workspace — keep an automations panel in sync as automations are added, toggled active/inactive, or removed. This is a config feed (the automation itself), not a feed of automation _runs_ — for runs, use [`subscribeToAutomationExecution`](#subscribetoautomationexecution).

### Request

```graphql
subscription OnAutomationChange {
  subscribeToAutomation(projectId: "project_123") {
    mutation
    node {
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
    }
    updatedFields
  }
}
```

### Parameters

| Parameter   | Type      | Required | Description                                                                                                    |
| ----------- | --------- | -------- | -------------------------------------------------------------------------------------------------------------- |
| `projectId` | `String!` | Yes      | Workspace ID. Events are delivered only for automations in this workspace, and only if you are a member of it. |

### Response

`subscribeToAutomation` returns an `AutomationSubscriptionPayload` per event.

```json
{
  "data": {
    "subscribeToAutomation": {
      "mutation": "UPDATED",
      "node": {
        "id": "clm4n8qwx000008l0g4oxdqn7",
        "isActive": false,
        "trigger": { "id": "clm4n8qwx000108l0abcd1234", "type": "CUSTOM_FIELD_ADDED" },
        "actions": [{ "id": "clm4n8qwx000208l0efgh5678", "type": "ADD_CUSTOM_FIELD" }]
      },
      "updatedFields": ["isActive"]
    }
  }
}
```

#### AutomationSubscriptionPayload

| Field            | Type                       | Description                                                    |
| ---------------- | -------------------------- | -------------------------------------------------------------- |
| `mutation`       | `MutationType!`            | The kind of change: `CREATED`, `UPDATED`, or `DELETED`.        |
| `node`           | `Automation`               | The automation after the change. `null` on a `DELETED` event.  |
| `previousValues` | `AutomationPreviousValues` | The automation's scalar values before the change.              |
| `updatedFields`  | `[String!]`                | Names of the scalar fields that changed on an `UPDATED` event. |

#### Automation

The most useful fields on the `node`. See [Create an automation](/api/automations/create-automation) for the full trigger/action model.

| Field       | Type                   | Description                                  |
| ----------- | ---------------------- | -------------------------------------------- |
| `id`        | `ID!`                  | The automation's unique identifier.          |
| `trigger`   | `AutomationTrigger!`   | What fires the automation.                   |
| `actions`   | `[AutomationAction!]!` | What the automation does when triggered.     |
| `isActive`  | `Boolean!`             | Whether the automation is currently enabled. |
| `createdBy` | `User!`                | The user who created the automation.         |
| `createdAt` | `DateTime!`            | When the automation was created.             |
| `updatedAt` | `DateTime!`            | When the automation was last modified.       |

## subscribeToAutomationExecution

Stream **automation runs** for a workspace. Every time an automation fires — by event, on a schedule, or because records started/stopped matching a condition — the server records an `AutomationExecution` and publishes its status here as it progresses through its batches.

### Request

```graphql
subscription OnAutomationExecution {
  subscribeToAutomationExecution(filter: { projectId: "project_123" }) {
    id
    automationId
    type
    status
    totalTodos
    completedBatches
    failedBatches
    todosAffected
    startedAt
    completedAt
    errorMessage
  }
}
```

### Parameters

#### AutomationExecutionSubscriptionFilter

| Parameter   | Type      | Required | Description                                                        |
| ----------- | --------- | -------- | ------------------------------------------------------------------ |
| `projectId` | `String!` | Yes      | Workspace ID. An execution matches via its automation's workspace. |

### Response

`subscribeToAutomationExecution` returns an `AutomationExecution` per event.

```json
{
  "data": {
    "subscribeToAutomationExecution": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "automationId": "clm4n8qwx000108l0abcd1234",
      "type": "EVENT",
      "status": "IN_PROGRESS",
      "totalTodos": 50,
      "completedBatches": 1,
      "failedBatches": 0,
      "todosAffected": 25,
      "startedAt": "2026-05-29T12:00:00.000Z",
      "completedAt": null,
      "errorMessage": null
    }
  }
}
```

#### AutomationExecution

| Field              | Type                         | Description                                                        |
| ------------------ | ---------------------------- | ------------------------------------------------------------------ |
| `id`               | `ID!`                        | The execution's unique identifier.                                 |
| `automationId`     | `String!`                    | ID of the automation that ran.                                     |
| `automation`       | `Automation!`                | The automation that ran.                                           |
| `type`             | `AutomationExecutionType!`   | How the run was triggered: `EVENT`, `SCHEDULED`, or `CONDITIONAL`. |
| `status`           | `AutomationExecutionStatus!` | Current run state (see below).                                     |
| `totalTodos`       | `Int!`                       | Number of records in scope for the run.                            |
| `totalBatches`     | `Int`                        | Number of batches the run is split into.                           |
| `completedBatches` | `Int!`                       | Batches finished so far.                                           |
| `failedBatches`    | `Int!`                       | Batches that failed.                                               |
| `batchSize`        | `Int`                        | Records per batch.                                                 |
| `todosAffected`    | `Int`                        | Records actually modified so far.                                  |
| `startedAt`        | `DateTime!`                  | When the run started.                                              |
| `completedAt`      | `DateTime`                   | When the run finished; `null` while still running.                 |
| `errorMessage`     | `String`                     | Failure detail when the run failed or partially failed.            |

#### AutomationExecutionStatus

| Value              | Description                                        |
| ------------------ | -------------------------------------------------- |
| `PENDING`          | Queued, not yet started.                           |
| `IN_PROGRESS`      | Currently running.                                 |
| `COMPLETED`        | Finished successfully.                             |
| `FAILED`           | Finished with every batch failed.                  |
| `PARTIALLY_FAILED` | Finished with some batches succeeded, some failed. |

#### AutomationExecutionType

| Value         | Description                                                        |
| ------------- | ------------------------------------------------------------------ |
| `EVENT`       | Triggered by a record event (e.g. a field changed).                |
| `SCHEDULED`   | Triggered on a schedule.                                           |
| `CONDITIONAL` | Triggered because records started or stopped matching a condition. |

## subscribeToOAuthConnection

Stream create, update, and delete events for **OAuth connections** in a workspace — third-party integrations (e.g. GitHub) that a workspace has authorized. Use it to reflect connection state in an integrations settings panel as connections are added, renamed, expired, or removed.

### Request

```graphql
subscription OnOAuthConnectionChange {
  subscribeToOAuthConnection(input: { projectId: "project_123" }) {
    mutation
    node {
      id
      name
      provider
      expiredAt
    }
    previousValues {
      id
      name
    }
  }
}
```

### Parameters

#### SubscribeToOAuthConnectionInput

| Parameter   | Type      | Required | Description                                                                                                    |
| ----------- | --------- | -------- | -------------------------------------------------------------------------------------------------------------- |
| `projectId` | `String!` | Yes      | Workspace ID. Events are delivered only for connections in this workspace, and only if you are a member of it. |

### Response

`subscribeToOAuthConnection` returns an `OAuthConnectionSubscriptionPayload` per event.

```json
{
  "data": {
    "subscribeToOAuthConnection": {
      "mutation": "CREATED",
      "node": {
        "id": "clm4n8qwx000008l0g4oxdqn7",
        "name": "Acme GitHub",
        "provider": "GITHUB",
        "expiredAt": null
      },
      "previousValues": null
    }
  }
}
```

#### OAuthConnectionSubscriptionPayload

| Field            | Type              | Description                                                          |
| ---------------- | ----------------- | -------------------------------------------------------------------- |
| `mutation`       | `MutationType!`   | The kind of change: `CREATED`, `UPDATED`, or `DELETED`.              |
| `node`           | `OAuthConnection` | The connection after the change. `null` on a `DELETED` event.        |
| `previousValues` | `OAuthConnection` | The connection before the change, on `UPDATED` and `DELETED` events. |

#### OAuthConnection

| Field       | Type             | Description                                         |
| ----------- | ---------------- | --------------------------------------------------- |
| `id`        | `ID!`            | The connection's unique identifier.                 |
| `uid`       | `String!`        | Internal identifier.                                |
| `name`      | `String!`        | Display name for the connection.                    |
| `provider`  | `OAuthProvider!` | The integration provider (see below).               |
| `expiredAt` | `DateTime`       | When the connection's credentials expire, if known. |
| `metadata`  | `JSON`           | Provider-specific connection metadata.              |
| `project`   | `Project!`       | The workspace the connection belongs to.            |
| `createdBy` | `User!`          | The user who created the connection.                |
| `createdAt` | `DateTime!`      | When the connection was created.                    |
| `updatedAt` | `DateTime!`      | When the connection was last modified.              |

#### OAuthProvider

| Value              | Description        |
| ------------------ | ------------------ |
| `GITHUB`           | GitHub.            |
| `INUIT_QUICKBOOKS` | Intuit QuickBooks. |

## Permissions

Authentication is enforced at the WebSocket handshake — an unauthenticated connection is rejected before any subscription runs. Beyond that, delivery is gated **per event**:

- **`subscribeToImportExportProgress`** delivers an event only when the job's `userId` matches the authenticated user. Pass your own user ID as the `userId` argument.
- **`subscribeToAITagProgress`** delivers an event only to the user who started the tagging job.
- **`subscribeToAutomation`** and **`subscribeToOAuthConnection`** deliver an event only if you are a member of the workspace the automation or connection belongs to.
- **`subscribeToLookupProgress`**, **`subscribeToCoverGenerationProgress`**, and **`subscribeToAutomationExecution`** deliver events matched on the `projectId` you supply.

Subscribing always succeeds; you simply receive events only for jobs you can see. See [Connect & Authenticate](/api/realtime/connect-and-authenticate) for credentials.

## Related

- [Real-time (Subscriptions)](/api/realtime) — the shared payload shape and the full subscription catalog.
- [Connect & Authenticate](/api/realtime/connect-and-authenticate) — open and authenticate the WebSocket.
- [Automations](/api/automations) — create, edit, and delete the automations these streams report on.
- [Import & Export](/api/import-export) — the HTTP side of the jobs `subscribeToImportExportProgress` tracks.
