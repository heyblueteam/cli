---
title: Create a Webhook
description: Register a webhook endpoint with the createWebhook mutation to receive event notifications.
icon: Plus
order: 1
---

Use the `createWebhook` mutation to register an endpoint that receives event notifications from your organization. When a subscribed event occurs in a workspace you belong to, Blue POSTs a signed JSON payload to your URL. See [Webhooks](/api/webhooks) for the delivery model and payload shape.

The mutation returns a [`Webhook`](/api/webhooks#the-webhook-type). The `secret` used to sign payloads is **only** returned in this response — store it immediately.

## Request

The smallest valid call needs only a `url`. Omitting `events` subscribes the webhook to **all** event types (see [Parameters](#parameters)).

```graphql
mutation CreateWebhook {
  createWebhook(input: { url: "https://example.com/webhooks/blue" }) {
    id
    uid
    url
    secret
    status
    enabled
    createdAt
  }
}
```

`createWebhook` requires the `blue-org-id` header in addition to your token credentials — the webhook count is enforced per organization. The other webhook operations are user-scoped and do not need it.

```bash
curl -X POST https://api.blue.app/graphql \
  -H "Content-Type: application/json" \
  -H "blue-token-id: YOUR_TOKEN_ID" \
  -H "blue-token-secret: YOUR_TOKEN_SECRET" \
  -H "blue-org-id: YOUR_ORG_ID" \
  -d '{"query": "mutation { createWebhook(input: { url: \"https://example.com/webhooks/blue\" }) { id secret } }"}'
```

## Parameters

### CreateWebhookInput

| Parameter    | Type              | Required | Description                                                                                                                                                                                                                              |
| ------------ | ----------------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `url`        | `String!`         | Yes      | The HTTPS endpoint that receives webhook POST requests. Must resolve to a public address — URLs resolving to loopback, RFC-1918, link-local, or cloud-metadata ranges are rejected.                                                      |
| `name`       | `String`          | No       | A human-readable label for the webhook.                                                                                                                                                                                                  |
| `events`     | `[WebhookEvent!]` | No       | Event types to subscribe to. **Omitting this (or passing `[]`) subscribes the webhook to every event type.** Pass an explicit, non-empty array to scope delivery to specific events. See [Webhook events](/api/webhooks#webhook-events). |
| `projectIds` | `[String!]`       | No       | Workspace IDs to scope delivery to. You must be a member of every workspace listed. Omitting it (or passing `[]`) scopes the webhook to all workspaces you belong to.                                                                    |

<Callout variant="warning" title="Empty events = all events">

Leaving `events` unset does **not** create a silent webhook — it subscribes to the full event catalog. If you only want a few events, you must pass them explicitly. The same rule applies to `projectIds`: empty means all your workspaces.

</Callout>

## Response

```json
{
  "data": {
    "createWebhook": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "uid": "wh_8f3a2c1e",
      "url": "https://example.com/webhooks/blue",
      "secret": "pat_a1b2c3d4e5f6g7h8i9j0",
      "status": "HEALTHY",
      "enabled": true,
      "createdAt": "2026-05-29T14:02:11.000Z"
    }
  }
}
```

A new webhook is `enabled: true` with `status: HEALTHY` immediately and starts receiving events. `createWebhook` does **not** make a test request to your URL — there is no creation-time health check. (Health checks run only on [`updateWebhook`](/api/webhooks/update-webhook) when you re-enable a webhook.) If deliveries later fail, the webhook is auto-disabled and set to `UNHEALTHY`.

### Returns

`createWebhook` returns a `Webhook`:

| Field        | Type                 | Description                                                                    |
| ------------ | -------------------- | ------------------------------------------------------------------------------ |
| `id`         | `ID!`                | The webhook's unique identifier.                                               |
| `uid`        | `String!`            | Short public identifier.                                                       |
| `name`       | `String`             | The label you set, or `null`.                                                  |
| `url`        | `String!`            | The endpoint receiving events.                                                 |
| `secret`     | `String`             | HMAC signing secret. **Returned only here** — `null` on every subsequent read. |
| `status`     | `WebhookStatusType!` | `HEALTHY` or `UNHEALTHY`.                                                      |
| `events`     | `[WebhookEvent!]`    | Subscribed event types (`[]` means all).                                       |
| `projectIds` | `[String!]`          | Scoped workspace IDs (`[]` means all).                                         |
| `enabled`    | `Boolean`            | Whether delivery is active.                                                    |
| `metadata`   | `JSON`               | Stored `events`/`projectIds` scoping metadata.                                 |
| `createdAt`  | `DateTime!`          | Creation timestamp.                                                            |
| `updatedAt`  | `DateTime!`          | Last-update timestamp.                                                         |

## Full example

Subscribe a named webhook to a few events across two specific workspaces:

```graphql
mutation CreateScopedWebhook($input: CreateWebhookInput!) {
  createWebhook(input: $input) {
    id
    uid
    name
    url
    secret
    status
    events
    projectIds
    enabled
    createdAt
  }
}
```

```json
{
  "input": {
    "name": "Slack notifications",
    "url": "https://example.com/webhooks/blue",
    "events": ["TODO_CREATED", "TODO_DONE_STATUS_UPDATED", "COMMENT_CREATED"],
    "projectIds": ["project_123", "project_456"]
  }
}
```

## Errors

| Code                    | When                                                                                                                                |
| ----------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| `COMPANY_NOT_FOUND`     | The `blue-org-id` header is missing or does not match an organization you belong to.                                          |
| `PLAN_LIMIT_REACHED`    | Your organization is at its plan's enabled-webhook limit. The message reports the count and cap; upgrade to add more.               |
| `BAD_USER_INPUT`        | `url` is missing or is not a valid URL.                                                                                             |
| `INTERNAL_SERVER_ERROR` | `url` resolves to a private/internal address (message: `URL resolves to a private/internal address`) or uses a disallowed protocol. |
| `FORBIDDEN`             | One or more `projectIds` reference a workspace you are not a member of (message: `You are not authorized.`).                        |

## Permissions

You must be authenticated and send `blue-org-id`. You can only scope a webhook to workspaces you currently belong to. The webhook is owned by your user; only you can update or delete it, and only events from workspaces you are a member of are delivered (see [Delivery scoping](/api/webhooks#delivery-and-scoping)).

## Related

- [Webhooks overview](/api/webhooks)
- [List webhooks](/api/webhooks/list-webhooks)
- [Update a webhook](/api/webhooks/update-webhook)
- [Disable a webhook](/api/webhooks/disable-webhook)
- [Delete a webhook](/api/webhooks/delete-webhook)
