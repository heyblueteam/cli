---
title: Custom Field Subscriptions
description: Stream real-time changes to custom field definitions and their options so a client's field schema stays in sync.
icon: SlidersHorizontal
order: 5
---

Keep a client's field schema live as fields and their options change. Custom fields are `CustomField` objects in the API; this page covers three subscriptions: `subscribeToCustomField` (field definition changes), `subscribeToCustomFieldOption` (changes to individual select options), and `onCustomFieldOptionsCreated` (a batch event fired when several options are added at once).

All three run over the same authenticated WebSocket as every other subscription — see [Connect & Authenticate](/api/realtime/connect-and-authenticate) for the `graphql-ws` handshake and credentials. The endpoint is `wss://api.blue.app/graphql`.

The two `subscribeTo*` fields return the standard entity change-feed payload (`{ mutation, node, previousValues, … }`) described on the [Real-time overview](/api/realtime). The `mutation` field is a `MutationType` (`CREATED`, `UPDATED`, or `DELETED`), letting one stream report adds, edits, and removals together.

## subscribeToCustomField

Streams changes to custom field **definitions** in one workspace — a field being created, renamed, retyped, repositioned, or deleted. Workspaces are `Project` objects in the API, so the field is scoped by `projectId`.

### Request

```graphql
subscription OnCustomFieldChanged {
  subscribeToCustomField(projectId: "project_123") {
    mutation
    node {
      id
      name
      type
      position
    }
    updatedFields
    previousValues {
      name
      type
    }
  }
}
```

`projectId` accepts a workspace ID or slug.

### Parameters

| Parameter   | Type      | Required | Description                                                                           |
| ----------- | --------- | -------- | ------------------------------------------------------------------------------------- |
| `projectId` | `String!` | Yes      | Workspace (project) whose custom field definitions to watch. Accepts an ID or a slug. |

### Response

Each delivered event is a `CustomFieldSubscriptionPayload`. The example below shows a field that was renamed — `updatedFields` lists the changed property and `previousValues` carries the pre-change snapshot.

```json
{
  "data": {
    "subscribeToCustomField": {
      "mutation": "UPDATED",
      "node": {
        "id": "clm4n8qwx000008l0g4oxdqn7",
        "name": "Priority",
        "type": "SELECT_SINGLE",
        "position": 3
      },
      "updatedFields": ["name"],
      "previousValues": {
        "name": "Importance",
        "type": "SELECT_SINGLE"
      }
    }
  }
}
```

#### CustomFieldSubscriptionPayload

| Field            | Type                        | Description                                                                  |
| ---------------- | --------------------------- | ---------------------------------------------------------------------------- |
| `mutation`       | `MutationType!`             | `CREATED`, `UPDATED`, or `DELETED`.                                          |
| `node`           | `CustomField`               | The field in its current state. `null` on `DELETED`.                         |
| `updatedFields`  | `[String!]`                 | Names of the field properties that changed on an `UPDATED` event.            |
| `previousValues` | `CustomFieldPreviousValues` | Snapshot of the field before the change. Present on `UPDATED` and `DELETED`. |

`node` is the full `CustomField` type (select `name`, `type`, `position`, `description`, `customFieldOptions { id title }`, and any other field property you track). `previousValues` is a narrower `CustomFieldPreviousValues` with the core scalars only: `id`, `uid`, `name`, `type`, `position`, `description`, `min`, `max`, `currency`, `prefix`, `createdAt`, `updatedAt`.

<Callout variant="info" title="Field options live on a separate stream">

`subscribeToCustomField` fires for the field definition itself. Adding or editing an individual select option does not necessarily emit a field-level event — subscribe to `subscribeToCustomFieldOption` (below) to track option changes granularly.

</Callout>

## subscribeToCustomFieldOption

Streams changes to the individual options of one select-style custom field — for example a single-select or multi-select field's choices being added, renamed, recolored, or removed. Scoped by `customFieldId`.

### Request

```graphql
subscription OnCustomFieldOptionChanged {
  subscribeToCustomFieldOption(input: { customFieldId: "field_123" }) {
    mutation
    node {
      id
      title
      color
      position
    }
    previousValues {
      title
      color
    }
  }
}
```

### Parameters

#### SubscribeToCustomFieldOptionInput

| Parameter       | Type      | Required | Description                              |
| --------------- | --------- | -------- | ---------------------------------------- |
| `customFieldId` | `String!` | Yes      | The custom field whose options to watch. |

### Response

Each event is a `CustomFieldOptionSubscriptionPayload`. Unlike the field-level payload, it has **no `updatedFields`** — compare `node` against `previousValues` to see what changed.

```json
{
  "data": {
    "subscribeToCustomFieldOption": {
      "mutation": "UPDATED",
      "node": {
        "id": "clm4n8qwx000008l0g4oxdqn7",
        "title": "Urgent",
        "color": "#e5484d",
        "position": 0
      },
      "previousValues": {
        "title": "High",
        "color": "#f76808"
      }
    }
  }
}
```

#### CustomFieldOptionSubscriptionPayload

| Field            | Type                | Description                                                       |
| ---------------- | ------------------- | ----------------------------------------------------------------- |
| `mutation`       | `MutationType!`     | `CREATED`, `UPDATED`, or `DELETED`.                               |
| `node`           | `CustomFieldOption` | The option in its current state. `null` on `DELETED`.             |
| `previousValues` | `CustomFieldOption` | The option before the change. Present on `UPDATED` and `DELETED`. |

Both `node` and `previousValues` are full `CustomFieldOption` objects — select `title`, `color`, `position`, `buttonType`, `buttonConfirmText`, and `currencyConversionFrom` / `currencyConversionTo` as needed.

## onCustomFieldOptionsCreated

A batch event for the common case of adding several options to a field at once (the `createCustomFieldOptions` mutation creates options in bulk). Rather than emitting one `subscribeToCustomFieldOption` `CREATED` event per option, this channel delivers the whole newly-created batch as a single array. Scoped by `customFieldId`.

### Request

```graphql
subscription OnCustomFieldOptionsCreated {
  onCustomFieldOptionsCreated(customFieldId: "field_123") {
    id
    title
    color
    position
  }
}
```

### Parameters

| Parameter       | Type      | Required | Description                                          |
| --------------- | --------- | -------- | ---------------------------------------------------- |
| `customFieldId` | `String!` | Yes      | The custom field to watch for batch option creation. |

### Response

Each event is `[CustomFieldOption!]!` — the array of options created in that batch.

```json
{
  "data": {
    "onCustomFieldOptionsCreated": [
      {
        "id": "clm4n8qwx000008l0g4oxdqn7",
        "title": "Low",
        "color": "#46a758",
        "position": 0
      },
      {
        "id": "clm4n8qwx000108l0h2pyerm8",
        "title": "Medium",
        "color": "#f76808",
        "position": 1
      },
      {
        "id": "clm4n8qwx000208l0i9qzfsn9",
        "title": "High",
        "color": "#e5484d",
        "position": 2
      }
    ]
  }
}
```

## Errors

Subscriptions report invalid arguments and missing scopes when the operation is set up. Permission scoping is otherwise silent — see Permissions below.

| Code                     | When                                                                                                                                                                                   |
| ------------------------ | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `UNAUTHENTICATED`        | The WebSocket handshake carried no valid credentials. The connection is rejected before any subscription starts. See [Connect & Authenticate](/api/realtime/connect-and-authenticate). |
| `PROJECT_NOT_FOUND`      | The `projectId` passed to `subscribeToCustomField` matches no workspace you can access.                                                                                                |
| `CUSTOM_FIELD_NOT_FOUND` | The `customFieldId` passed to `subscribeToCustomFieldOption` or `onCustomFieldOptionsCreated` matches no field you can access.                                                         |

## Permissions

Events are filtered per delivered event, not just at connection time. Subscribing always succeeds for an authenticated client, but you only receive events you are allowed to see:

- **`subscribeToCustomField`** delivers an event only when the changed field belongs to the `projectId` you named **and** you are a member of that workspace. Changes in other workspaces, or in a workspace you can't access, are never delivered.
- **`subscribeToCustomFieldOption`** and **`onCustomFieldOptionsCreated`** deliver only events whose `customFieldId` matches the one you subscribed with — they are inherently scoped to a single field.

## Related

- [Real-time (Subscriptions)](/api/realtime) — the shared payload shape and the full subscription catalog.
- [Connect & Authenticate](/api/realtime/connect-and-authenticate) — open an authenticated WebSocket.
- [Custom Fields](/api/custom-fields) — create, list, and update field definitions.
- [Custom Field Values](/api/custom-fields/custom-field-values) — read and write a record's field values.
