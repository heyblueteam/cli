---
title: List Webhooks
description: Query your webhooks with the webhooks and webhook queries — paginate the list or fetch a single webhook by ID.
icon: List
order: 2
---

Use the `webhooks` query to retrieve a paginated list of the webhooks you created, and the `webhook` query to fetch a single webhook by its ID. Both are **top-level `Query` fields** — they are not nested under a query group (unlike `recordQueries` or `customFieldQueries` elsewhere in the API).

Webhooks are user-scoped: each query returns only the webhooks created by the authenticated token's user.

## Request

The smallest call lists your webhooks with the default page size:

```graphql
query ListWebhooks {
  webhooks {
    items {
      id
      uid
      name
      url
      status
      enabled
      createdAt
    }
    pageInfo {
      totalItems
      totalPages
      hasNextPage
      hasPreviousPage
    }
  }
}
```

## Parameters

### `webhooks` arguments

| Parameter | Type            | Default | Description                                       |
| --------- | --------------- | ------- | ------------------------------------------------- |
| `filter`  | `WebhookFilter` | —       | Optional filter to narrow results.                |
| `skip`    | `Int`           | `0`     | Number of items to skip before returning results. |
| `take`    | `Int`           | `20`    | Number of items to return per page.               |

### `webhook` arguments

| Parameter | Type      | Required | Description                        |
| --------- | --------- | -------- | ---------------------------------- |
| `id`      | `String!` | Yes      | The ID of the webhook to retrieve. |

### WebhookFilter

| Parameter | Type      | Required | Description                                                       |
| --------- | --------- | -------- | ----------------------------------------------------------------- |
| `enabled` | `Boolean` | No       | When `true`, returns only enabled webhooks. See the caveat below. |

<Callout variant="warning" title="`enabled: false` returns all webhooks">

The resolver applies the filter as `enabled || undefined`, so only `enabled: true` filters anything. Passing `enabled: false` (or omitting `filter` entirely) coerces to `undefined` and returns **all** of your webhooks — both enabled and disabled. Filtering for disabled-only webhooks is not currently supported; fetch all and check the `enabled` field client-side.

</Callout>

## Response

```json
{
  "data": {
    "webhooks": {
      "items": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "uid": "whk_8qwx000008l0",
          "name": "Production sync",
          "url": "https://example.com/hooks/blue",
          "status": "HEALTHY",
          "enabled": true,
          "createdAt": "2026-05-21T14:02:11.000Z"
        }
      ],
      "pageInfo": {
        "totalItems": 1,
        "totalPages": 1,
        "hasNextPage": false,
        "hasPreviousPage": false
      }
    }
  }
}
```

### Webhook

| Field        | Type                 | Description                                                                                       |
| ------------ | -------------------- | ------------------------------------------------------------------------------------------------- |
| `id`         | `ID!`                | Unique identifier for the webhook.                                                                |
| `uid`        | `String!`            | Short, user-friendly identifier.                                                                  |
| `name`       | `String`             | Human-readable name for the webhook.                                                              |
| `url`        | `String!`            | The endpoint URL that receives event POST requests.                                               |
| `secret`     | `String`             | The HMAC signing secret. Always `null` when queried — it is returned only once, at creation time. |
| `status`     | `WebhookStatusType!` | Delivery health: `HEALTHY` or `UNHEALTHY`.                                                        |
| `events`     | `[WebhookEvent!]`    | The event types this webhook is subscribed to.                                                    |
| `projectIds` | `[String!]`          | The workspace IDs this webhook is scoped to. Empty means all workspaces.                          |
| `enabled`    | `Boolean`            | Whether the webhook is currently active.                                                          |
| `metadata`   | `JSON`               | Additional metadata stored with the webhook.                                                      |
| `createdAt`  | `DateTime!`          | When the webhook was created.                                                                     |
| `updatedAt`  | `DateTime!`          | When the webhook was last updated.                                                                |

### PageInfo

| Field             | Type       | Description                                                   |
| ----------------- | ---------- | ------------------------------------------------------------- |
| `totalItems`      | `Int`      | Total number of webhooks matching the query.                  |
| `totalPages`      | `Int`      | Total number of pages at the current `take` size.             |
| `page`            | `Int`      | Current page number.                                          |
| `perPage`         | `Int`      | Number of items per page.                                     |
| `hasNextPage`     | `Boolean!` | Whether more results exist after this page.                   |
| `hasPreviousPage` | `Boolean!` | Whether results exist before this page.                       |
| `startCursor`     | `String`   | Deprecated. Not used by `skip`/`take` pagination — ignore it. |
| `endCursor`       | `String`   | Deprecated. Not used by `skip`/`take` pagination — ignore it. |

## Full example

Filter for enabled webhooks, page through the results, and select the full field set:

```graphql
query ListEnabledWebhooks($filter: WebhookFilter, $skip: Int, $take: Int) {
  webhooks(filter: $filter, skip: $skip, take: $take) {
    items {
      id
      uid
      name
      url
      status
      events
      projectIds
      enabled
      metadata
      createdAt
      updatedAt
    }
    pageInfo {
      totalItems
      totalPages
      page
      perPage
      hasNextPage
      hasPreviousPage
    }
  }
}
```

Variables:

```json
{
  "filter": { "enabled": true },
  "skip": 0,
  "take": 10
}
```

To fetch a single webhook by ID, use the `webhook` query:

```graphql
query GetWebhook($id: String!) {
  webhook(id: $id) {
    id
    uid
    name
    url
    status
    events
    projectIds
    enabled
    metadata
    createdAt
    updatedAt
  }
}
```

Variables:

```json
{
  "id": "clm4n8qwx000008l0g4oxdqn7"
}
```

## Errors

| Code                | When                                                                                            |
| ------------------- | ----------------------------------------------------------------------------------------------- |
| `WEBHOOK_NOT_FOUND` | The `webhook` query is called with an ID that does not exist. Message: `Webhook was not found.` |
| `FORBIDDEN`         | You request a webhook created by another user. Message: `You are not authorized.`               |

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

Both queries are user-scoped. The `webhooks` list returns only the webhooks created by the authenticated token's user, ordered by creation date, newest first. The `webhook` query returns `FORBIDDEN` if the requested webhook belongs to another user.

## Related

- [Create a webhook](/api/webhooks/create-webhook)
- [Update a webhook](/api/webhooks/update-webhook)
- [Disable a webhook](/api/webhooks/disable-webhook)
- [Delete a webhook](/api/webhooks/delete-webhook)
- [Webhooks overview](/api/webhooks)
