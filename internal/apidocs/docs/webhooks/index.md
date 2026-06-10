---
title: Webhooks
description: Receive signed real-time HTTP notifications when records, comments, tags, and other entities change in Blue.
icon: Webhook
order: 0
---

Webhooks deliver real-time HTTP notifications when events occur in your workspaces — a record is created, a comment is posted, a tag is removed, and more. When a subscribed event fires, Blue sends a signed `POST` request to your endpoint with the event type and the entity's before/after state.

Webhooks are `Webhook` objects in the API, owned by the user who created them. Each webhook subscribes to a set of [event types](#webhook-events) and is scoped to a set of workspaces (`Project` objects).

## Operations

| Operation                                          | Mutation / Query       | Description                                                     |
| -------------------------------------------------- | ---------------------- | --------------------------------------------------------------- |
| [Create a webhook](/api/webhooks/create-webhook)   | `createWebhook`        | Register an endpoint and choose which events it receives.       |
| [List webhooks](/api/webhooks/list-webhooks)       | `webhooks` / `webhook` | Page through your webhooks or fetch one by ID.                  |
| [Update a webhook](/api/webhooks/update-webhook)   | `updateWebhook`        | Change URL, events, scope, or enable/disable a webhook.         |
| [Disable a webhook](/api/webhooks/disable-webhook) | `disableWebhook`       | Stop delivery and mark a webhook unhealthy without deleting it. |
| [Delete a webhook](/api/webhooks/delete-webhook)   | `deleteWebhook`        | Permanently remove a webhook.                                   |

## Authenticating webhook requests

Manage webhooks with your [personal access token headers](/api/start-guide/authentication). [`createWebhook`](/api/webhooks/create-webhook) additionally requires `X-Bloo-Company-ID` because the webhook count is enforced per organization; the other webhook operations are user-scoped.

```bash
curl -X POST https://api.blue.app/graphql \
  -H "Content-Type: application/json" \
  -H "X-Bloo-Token-ID: YOUR_TOKEN_ID" \
  -H "X-Bloo-Token-Secret: YOUR_TOKEN_SECRET" \
  -H "X-Bloo-Company-ID: YOUR_COMPANY_ID" \
  -d '{"query": "query { webhooks { items { id url status enabled } } }"}'
```

## Delivery and scoping

Each event is enqueued with a **5-second delay** before the first delivery attempt. Blue `POST`s the [payload](#webhook-payload) to your URL with a 7-second request timeout.

A webhook only receives an event when **all** of the following hold:

- The webhook is `enabled`.
- Its `events` list contains the event type — **or** the list is empty, which subscribes it to every event.
- Its `projectIds` list contains the event's workspace — **or** the list is empty, which scopes it to all workspaces.
- The webhook's **creator is currently a member of the event's workspace**. Project scoping is additionally gated by live membership, so removing the creator from a workspace silently stops delivery for events from it.

If a delivery attempt fails (non-2xx response, timeout, or connection error), BullMQ retries with backoff — **5 attempts by default**. After all retries are exhausted, the webhook is automatically disabled (`enabled: false`, `status: UNHEALTHY`) and its creator is emailed. This is what `HEALTHY` versus `UNHEALTHY` reflects: a healthy webhook is delivering; an unhealthy one was auto-disabled (or [disabled manually](/api/webhooks/disable-webhook)) and must be re-enabled via [`updateWebhook`](/api/webhooks/update-webhook).

## Webhook payload

Blue sends a JSON body with this exact shape:

```json
{
  "event": "TODO_DONE_STATUS_UPDATED",
  "webhook": {
    "id": "clm4n8qwx000008l0g4oxdqn7",
    "uid": "wh_8f3a2c1e",
    "url": "https://example.com/webhooks/blue",
    "status": "HEALTHY",
    "enabled": true,
    "metadata": { "events": [], "projectIds": [] }
  },
  "previousValue": { "id": "todo_123", "title": "Draft proposal", "done": false },
  "currentValue": { "id": "todo_123", "title": "Draft proposal", "done": true }
}
```

| Field           | Description                                                                                                                                                         |
| --------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `event`         | The `WebhookEvent` value that fired (e.g. `TODO_CREATED`).                                                                                                          |
| `webhook`       | The webhook's own configuration. **The `secret` is stripped** from this object — it never appears in a delivery.                                                    |
| `previousValue` | The entity's state before the change. `null` for create events.                                                                                                     |
| `currentValue`  | The entity's state after the change. `null` for delete events. The shape matches the underlying entity (record, comment, tag, etc.) with its related rows expanded. |

The delivered `event` may also be `WEBHOOK_HEALTH_CHECK`, sent when a webhook is re-enabled to verify reachability. That value is **not** part of the `WebhookEvent` enum and cannot be subscribed to — it is delivery-only, so handle unknown event strings defensively.

## Verifying signatures

Every delivery includes an `X-Signature` header: the HMAC-SHA256 of the raw request body, keyed by the webhook's `secret` (the value returned once from [`createWebhook`](/api/webhooks/create-webhook)), encoded as a hex string.

<Callout variant="warning" title="Hash the raw bytes, not a re-serialized object">

Compute the HMAC over the **exact bytes you received**, before parsing. If you `JSON.parse` the body and re-`stringify` it, key ordering and formatting may differ from what Blue signed and the signature will not match.

</Callout>

```javascript
const crypto = require('crypto')

// `rawBody` is the unparsed request body string (e.g. from a raw body parser).
function verifySignature(rawBody, signatureHeader, secret) {
  const expected = crypto.createHmac('sha256', secret).update(rawBody).digest('hex')
  return crypto.timingSafeEqual(Buffer.from(expected), Buffer.from(signatureHeader))
}
```

## Webhook events

Subscribe to these `WebhookEvent` values via the `events` array on [`createWebhook`](/api/webhooks/create-webhook) or [`updateWebhook`](/api/webhooks/update-webhook). An empty array subscribes to all of them.

### Record events

| Event                       | Description                               |
| --------------------------- | ----------------------------------------- |
| `TODO_CREATED`              | A record was created.                     |
| `TODO_DELETED`              | A record was deleted.                     |
| `TODO_MOVED`                | A record was moved to a different list.   |
| `TODO_NAME_CHANGED`         | A record's title was changed.             |
| `TODO_DONE_STATUS_UPDATED`  | A record was marked done or undone.       |
| `TODO_DUE_DATE_ADDED`       | A due date was added to a record.         |
| `TODO_DUE_DATE_UPDATED`     | A record's due date was changed.          |
| `TODO_DUE_DATE_REMOVED`     | A due date was removed from a record.     |
| `TODO_ASSIGNEE_ADDED`       | An assignee was added to a record.        |
| `TODO_ASSIGNEE_REMOVED`     | An assignee was removed from a record.    |
| `TODO_TAG_ADDED`            | A tag was added to a record.              |
| `TODO_TAG_REMOVED`          | A tag was removed from a record.          |
| `TODO_CUSTOM_FIELD_UPDATED` | A custom field value changed on a record. |

### Checklist events

| Event                                     | Description                                    |
| ----------------------------------------- | ---------------------------------------------- |
| `TODO_CHECKLIST_CREATED`                  | A checklist was created on a record.           |
| `TODO_CHECKLIST_NAME_CHANGED`             | A checklist was renamed.                       |
| `TODO_CHECKLIST_DELETED`                  | A checklist was deleted.                       |
| `TODO_CHECKLIST_ITEM_CREATED`             | A checklist item was created.                  |
| `TODO_CHECKLIST_ITEM_NAME_CHANGED`        | A checklist item was renamed.                  |
| `TODO_CHECKLIST_ITEM_DELETED`             | A checklist item was deleted.                  |
| `TODO_CHECKLIST_ITEM_DUE_DATE_ADDED`      | A due date was added to a checklist item.      |
| `TODO_CHECKLIST_ITEM_DUE_DATE_UPDATED`    | A checklist item's due date was changed.       |
| `TODO_CHECKLIST_ITEM_DUE_DATE_REMOVED`    | A due date was removed from a checklist item.  |
| `TODO_CHECKLIST_ITEM_ASSIGNEE_ADDED`      | An assignee was added to a checklist item.     |
| `TODO_CHECKLIST_ITEM_ASSIGNEE_REMOVED`    | An assignee was removed from a checklist item. |
| `TODO_CHECKLIST_ITEM_DONE_STATUS_UPDATED` | A checklist item was marked done or undone.    |

### List events

| Event                    | Description                        |
| ------------------------ | ---------------------------------- |
| `TODO_LIST_CREATED`      | A list was created in a workspace. |
| `TODO_LIST_DELETED`      | A list was deleted.                |
| `TODO_LIST_NAME_CHANGED` | A list was renamed.                |

### Custom field events

| Event                  | Description                            |
| ---------------------- | -------------------------------------- |
| `CUSTOM_FIELD_CREATED` | A custom field was created.            |
| `CUSTOM_FIELD_DELETED` | A custom field was deleted.            |
| `CUSTOM_FIELD_UPDATED` | A custom field definition was updated. |

### Tag events

| Event         | Description        |
| ------------- | ------------------ |
| `TAG_CREATED` | A tag was created. |
| `TAG_DELETED` | A tag was deleted. |
| `TAG_UPDATED` | A tag was updated. |

### Comment events

| Event             | Description            |
| ----------------- | ---------------------- |
| `COMMENT_CREATED` | A comment was created. |
| `COMMENT_DELETED` | A comment was deleted. |
| `COMMENT_UPDATED` | A comment was updated. |

## The Webhook type

| Field        | Type                 | Description                                                                                                              |
| ------------ | -------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| `id`         | `ID!`                | Unique identifier.                                                                                                       |
| `uid`        | `String!`            | Short public identifier.                                                                                                 |
| `name`       | `String`             | Human-readable label, or `null`.                                                                                         |
| `url`        | `String!`            | The endpoint receiving events.                                                                                           |
| `secret`     | `String`             | HMAC signing secret. Returned **only** from [`createWebhook`](/api/webhooks/create-webhook); `null` on every other read. |
| `status`     | `WebhookStatusType!` | `HEALTHY` or `UNHEALTHY`.                                                                                                |
| `events`     | `[WebhookEvent!]`    | Subscribed event types. An empty array means all events.                                                                 |
| `projectIds` | `[String!]`          | Scoped workspace IDs. An empty array means all your workspaces.                                                          |
| `enabled`    | `Boolean`            | Whether the webhook is delivering.                                                                                       |
| `metadata`   | `JSON`               | Stored scoping metadata (`events` and `projectIds`).                                                                     |
| `createdAt`  | `DateTime!`          | Creation timestamp.                                                                                                      |
| `updatedAt`  | `DateTime!`          | Last-update timestamp.                                                                                                   |

`WebhookStatusType` is `HEALTHY` or `UNHEALTHY`.

## Related

- [Create a webhook](/api/webhooks/create-webhook)
- [List webhooks](/api/webhooks/list-webhooks)
- [Update a webhook](/api/webhooks/update-webhook)
- [Disable a webhook](/api/webhooks/disable-webhook)
- [Delete a webhook](/api/webhooks/delete-webhook)
- [Authentication](/api/start-guide/authentication)
