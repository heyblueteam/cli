---
title: Delete a Webhook
description: Permanently delete a webhook with the deleteWebhook mutation. Queued deliveries are skipped once the webhook is gone.
icon: Trash2
order: 5
---

Use the `deleteWebhook` mutation to permanently remove a webhook. Once it's deleted, Blue stops delivering events to its URL: any deliveries already queued for that webhook are skipped at send time (the delivery worker re-fetches the webhook before each POST and no-ops if it's missing).

Deletion is permanent and cannot be undone. If you only want to pause deliveries while keeping the configuration, use [`disableWebhook`](/api/webhooks/disable-webhook) instead — it flips the webhook to disabled without destroying it.

You can only delete a webhook that you created.

## Request

```graphql
mutation DeleteWebhook {
  deleteWebhook(input: { webhookId: "webhook_123" }) {
    success
  }
}
```

Authenticate with your personal access token headers:

```
X-Bloo-Token-ID: YOUR_TOKEN_ID
X-Bloo-Token-Secret: YOUR_TOKEN_SECRET
X-Bloo-Company-ID: YOUR_COMPANY_ID
```

## Parameters

### DeleteWebhookInput

| Parameter   | Type      | Required | Description                      |
| ----------- | --------- | -------- | -------------------------------- |
| `webhookId` | `String!` | Yes      | The ID of the webhook to delete. |

## Response

`deleteWebhook` returns a `MutationResult`.

```json
{
  "data": {
    "deleteWebhook": {
      "success": true
    }
  }
}
```

### Returns

| Field         | Type       | Description                                                                                |
| ------------- | ---------- | ------------------------------------------------------------------------------------------ |
| `success`     | `Boolean!` | `true` when the webhook was deleted.                                                       |
| `operationId` | `String`   | Identifier for the operation when one is tracked asynchronously. Null for `deleteWebhook`. |

## Errors

| Code                | When                                                                                   |
| ------------------- | -------------------------------------------------------------------------------------- |
| `WEBHOOK_NOT_FOUND` | No webhook matches the provided `webhookId`. Message: `Webhook was not found.`         |
| `FORBIDDEN`         | The webhook exists but was created by another user. Message: `You are not authorized.` |

```json
{
  "errors": [
    {
      "message": "Webhook was not found.",
      "extensions": { "code": "WEBHOOK_NOT_FOUND" }
    }
  ]
}
```

## Permissions

You must be authenticated, and you can only delete a webhook you created. Deleting a webhook owned by another user returns `FORBIDDEN`.

Webhooks are also tied to their creator's lifecycle: deleting a user cascades to delete every webhook that user created.

## Related

- [List webhooks](/api/webhooks/list-webhooks)
- [Disable a webhook](/api/webhooks/disable-webhook) — pause deliveries without deleting
- [Update a webhook](/api/webhooks/update-webhook)
- [Webhooks overview](/api/webhooks)
