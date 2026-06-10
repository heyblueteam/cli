---
title: Disable a Webhook
description: Stop a webhook from receiving events with the disableWebhook mutation without deleting it.
icon: PauseCircle
order: 3
---

Use the `disableWebhook` mutation to stop a webhook from receiving events while keeping its configuration. The webhook's `enabled` flag is set to `false` and its `status` to `UNHEALTHY`, and the webhook's creator is emailed a health-status notification. To remove a webhook entirely, use [`deleteWebhook`](/api/webhooks/delete-webhook); to turn it back on, use [`updateWebhook`](/api/webhooks/update-webhook) with `enabled: true`.

This is the same state Blue applies automatically when a webhook's deliveries fail repeatedly (see [Delivery and health](/api/webhooks#delivery-and-scoping)).

## Request

```graphql
mutation DisableWebhook {
  disableWebhook(input: { webhookId: "webhook_123" }) {
    success
    operationId
  }
}
```

## Parameters

### DisableWebhookInput

| Parameter   | Type      | Required | Description                         |
| ----------- | --------- | -------- | ----------------------------------- |
| `webhookId` | `String!` | Yes      | The `id` of the webhook to disable. |

## Response

`disableWebhook` returns a `MutationResult`.

```json
{
  "data": {
    "disableWebhook": {
      "success": true,
      "operationId": null
    }
  }
}
```

### Returns

| Field         | Type       | Description                                    |
| ------------- | ---------- | ---------------------------------------------- |
| `success`     | `Boolean!` | `true` when the webhook was disabled.          |
| `operationId` | `String`   | Identifier for the operation, when applicable. |

## Errors

| Code                | When                                                                              |
| ------------------- | --------------------------------------------------------------------------------- |
| `WEBHOOK_NOT_FOUND` | No webhook exists with the given `webhookId` (message: `Webhook was not found.`). |

## Permissions

You must be authenticated. Unlike [`updateWebhook`](/api/webhooks/update-webhook) and [`deleteWebhook`](/api/webhooks/delete-webhook), `disableWebhook` does not check that you created the webhook.

## Related

- [Webhooks overview](/api/webhooks)
- [Update a webhook](/api/webhooks/update-webhook)
- [Delete a webhook](/api/webhooks/delete-webhook)
- [List webhooks](/api/webhooks/list-webhooks)
