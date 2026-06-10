---
title: Update a Webhook
description: Update a webhook's URL, name, events, workspace scope, and enabled status with the updateWebhook mutation.
icon: RefreshCw
order: 3
---

Use the `updateWebhook` mutation to change an existing webhook's URL, name, subscribed events, workspace scope, or enabled status. Every field except `webhookId` is optional, so you can patch a single attribute or several at once. Only the user who created a webhook can update it.

Webhooks are scoped to an organization (a `Company` in the API) and the request runs with your personal access token headers. Re-enabling a webhook (`enabled: true`) triggers a live health check before the change is committed — see [Health check](#health-check) below.

## Request

This example renames a webhook and points it at a new endpoint:

```graphql
mutation UpdateWebhook {
  updateWebhook(
    input: {
      webhookId: "clm4n8qwx000008l0g4oxdqn7"
      name: "Production notifications"
      url: "https://example.com/webhooks/blue"
    }
  ) {
    id
    name
    url
    status
    enabled
    updatedAt
  }
}
```

Send your credentials as headers (header names are case-insensitive):

```http
POST https://api.blue.app/graphql
X-Bloo-Token-ID: YOUR_TOKEN_ID
X-Bloo-Token-Secret: YOUR_TOKEN_SECRET
X-Bloo-Company-ID: YOUR_COMPANY_ID
Content-Type: application/json
```

## Parameters

### UpdateWebhookInput

| Parameter    | Type             | Required | Description                                                                                                                       |
| ------------ | ---------------- | -------- | --------------------------------------------------------------------------------------------------------------------------------- |
| `webhookId`  | `String!`        | Yes      | The ID of the webhook to update.                                                                                                  |
| `name`       | `String`         | No       | New human-readable name.                                                                                                          |
| `url`        | `String`         | No       | New endpoint URL. Must be a valid, publicly accessible URL — private/internal addresses are rejected.                             |
| `events`     | `[WebhookEvent]` | No       | New list of event types to subscribe to. **Replaces** the existing list entirely. Omit it to leave events unchanged.              |
| `projectIds` | `[String]`       | No       | New list of workspace IDs to scope the webhook to. **Replaces** the existing list entirely. Omit it to leave the scope unchanged. |
| `enabled`    | `Boolean`        | No       | `true` re-enables the webhook (and triggers a health check); `false` disables it. Omit it to leave the status unchanged.          |

<Callout variant="info" title="Nullable list elements — intentional schema asymmetry">

On `UpdateWebhookInput`, `events` is `[WebhookEvent]` and `projectIds` is `[String]` — the list **elements** are nullable. On `CreateWebhookInput` the same fields require non-null elements (`[WebhookEvent!]`, `[String!]`). This difference is intentional, not a typo. In practice, pass concrete values; do not include `null` entries in the list.

</Callout>

An empty `events` list (or omitting it on create) subscribes the webhook to **all** event types. The full set of `WebhookEvent` values is documented on the [Webhooks overview](/api/webhooks).

## Response

`updateWebhook` returns the updated `Webhook` object (`Webhook!`, non-null).

```json
{
  "data": {
    "updateWebhook": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Production notifications",
      "url": "https://example.com/webhooks/blue",
      "status": "HEALTHY",
      "enabled": true,
      "updatedAt": "2026-05-29T14:22:08.000Z"
    }
  }
}
```

### Returns — Webhook

| Field        | Type                 | Description                                                                                     |
| ------------ | -------------------- | ----------------------------------------------------------------------------------------------- |
| `id`         | `ID!`                | Unique identifier for the webhook.                                                              |
| `uid`        | `String!`            | User-friendly unique identifier.                                                                |
| `name`       | `String`             | Human-readable name.                                                                            |
| `url`        | `String!`            | Endpoint URL receiving event deliveries.                                                        |
| `secret`     | `String`             | Always `null` on update. The signing secret is returned only once, when the webhook is created. |
| `status`     | `WebhookStatusType!` | Health status: `HEALTHY` or `UNHEALTHY`.                                                        |
| `events`     | `[WebhookEvent!]`    | Subscribed event types.                                                                         |
| `projectIds` | `[String!]`          | Workspace IDs this webhook is scoped to.                                                        |
| `enabled`    | `Boolean`            | Whether the webhook is currently active and delivering events.                                  |
| `metadata`   | `JSON`               | Additional metadata associated with the webhook.                                                |
| `createdAt`  | `DateTime!`          | When the webhook was created.                                                                   |
| `updatedAt`  | `DateTime!`          | When the webhook was last updated.                                                              |

## Full example

Replace the subscribed events and workspace scope, and re-enable the webhook, using variables:

```graphql
mutation UpdateWebhook($input: UpdateWebhookInput!) {
  updateWebhook(input: $input) {
    id
    uid
    name
    url
    status
    events
    projectIds
    enabled
    updatedAt
  }
}
```

```json
{
  "input": {
    "webhookId": "clm4n8qwx000008l0g4oxdqn7",
    "events": [
      "TODO_CREATED",
      "TODO_DELETED",
      "TODO_DONE_STATUS_UPDATED",
      "TODO_ASSIGNEE_ADDED",
      "TODO_ASSIGNEE_REMOVED",
      "COMMENT_CREATED"
    ],
    "projectIds": ["project_123"],
    "enabled": true
  }
}
```

## Health check

When the request sets `enabled: true`, Blue verifies the endpoint is reachable **before** committing the change. The check runs against the _effective URL_ — the new `url` if you supplied one in the same call, otherwise the webhook's current URL. So updating `url` and `enabled: true` together validates the new endpoint.

The check is a `POST` to the effective URL with this body:

```json
{
  "event": "WEBHOOK_HEALTH_CHECK",
  "webhook": { "...": "the webhook object", "secret": "" }
}
```

- The `secret` field in the body is blanked to `""`, but the request is still signed with the webhook's **real** secret via an HMAC-SHA256 `X-Signature` header — verify the signature with your stored secret.
- The health-check body has no `previousValue` / `currentValue` keys; it differs from real event deliveries by design. Use the `event` value to distinguish a probe from a real delivery.
- The request has a **7-second timeout**. Any `2xx` response passes.

If the check passes, the webhook is enabled and `status` becomes `HEALTHY`. If it times out or returns a non-`2xx` status, the update still succeeds but the webhook is left **disabled** with `status` set to `UNHEALTHY`. Setting `enabled: true` always runs the check, even if the webhook was already enabled.

<Callout variant="warning" title="updateWebhook does not re-check your plan limit">

The webhook plan limit is enforced only at creation (`createWebhook`), and the limit counts enabled webhooks. Re-enabling webhooks via `updateWebhook` is **not** limit-checked, so it can take you over your plan's enabled-webhook count.

</Callout>

## Errors

| Code                    | Message                                      | When                                                          |
| ----------------------- | -------------------------------------------- | ------------------------------------------------------------- |
| `WEBHOOK_NOT_FOUND`     | `Webhook was not found.`                     | No webhook matches the provided `webhookId`.                  |
| `FORBIDDEN`             | `You are not authorized.`                    | You are not the user who created the webhook.                 |
| `BAD_USER_INPUT`        | URL validation message                       | The provided `url` is not a valid URL.                        |
| `INTERNAL_SERVER_ERROR` | `URL resolves to a private/internal address` | The provided `url` resolves to a private or internal address. |

## Permissions

You can only update a webhook you created (`createdById` must match the authenticated user). There is no organization-level override — even an organization owner cannot update another user's webhook through this mutation.

Note that `disableWebhook` does **not** apply this ownership check: any authenticated user with access can disable a webhook by ID. To disable a webhook you own, either send `enabled: false` here or use `disableWebhook`.

## Related

- [List webhooks](/api/webhooks/list-webhooks)
- [Delete a webhook](/api/webhooks/delete-webhook)
- [Webhooks overview](/api/webhooks) — create a webhook, the event catalog, and signature verification
