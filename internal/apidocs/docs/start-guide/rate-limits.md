---
title: Rate Limits
description: How Blue rate-limits API requests — per-minute tier limits, the three checked layers, response headers, and per-operation caps.
icon: Gauge
order: 4
---

The Blue API rate-limits every request authenticated with a personal access token (PAT) or session JWT. Limits protect platform stability and keep one integration from starving the rest of an organization. Limits are applied **per minute** using a 60-second sliding window and scale with your organization's tier. Adding the **Pro** add-on doubles your tier's limits.

<Callout variant="info" title="Per minute, not per second">

All tier limits below are counts **per 60-second sliding window** — i.e. requests per minute. A `20` limit means 20 requests in any rolling 60 seconds, not 20 per second.

</Callout>

## Limits by tier

Each request is checked against three layers (per key, per user, per organization). The numbers are requests per minute. The per-organization layer is the hard ceiling on all API activity in the org.

| Tier         | Per API key | Per user | Per org |
| ------------ | ----------- | -------- | ------- |
| **Starter**  | 20/min      | 40/min   | 60/min  |
| **Starter+** | 30/min      | 60/min   | 90/min  |
| **Growth**   | 50/min      | 100/min  | 150/min |
| **Growth+**  | 100/min     | 200/min  | 300/min |
| **Scale**    | 200/min     | 400/min  | 600/min |

**Legacy tiers** (pre-2026 AppSumo and direct licenses) all resolve to `200/min` per key, `400/min` per user, `600/min` per org.

### Plan overlays

Plans (`Pro`, `Enterprise`) sit on top of whatever tier the organization is on and override its rate limits:

| Plan             | Per API key | Per user  | Per org   |
| ---------------- | ----------- | --------- | --------- |
| **+ Pro**        | 2× tier     | 2× tier   | 2× tier   |
| **+ Enterprise** | 1,000/min   | 2,000/min | 3,000/min |

Pro doubles all three of the tier's limits. Enterprise replaces them with fixed `1,000 / 2,000 / 3,000` per minute regardless of the underlying tier.

### Default fallback

If an organization has no resolvable tier (for example, an unconfigured or locked org), requests fall back to `200/min` per key, `400/min` per user, and `600/min` per org.

## How it works

Every request runs through up to three layers before any resolver executes. All applicable layers must pass; if any one is exceeded, the request is rejected with `429 Too Many Requests`.

1. **Per API key** — limits a single token's usage so one integration can't consume the whole org's capacity. This layer applies **only when both PAT headers** (`X-Bloo-Token-ID` and `X-Bloo-Token-Secret`) are present. Requests authenticated with a session JWT (`Authorization: Bearer`) skip this layer.
2. **Per user** — limits total usage across all of a user's tokens, so rotating keys can't bypass the per-key limit.
3. **Per organization** — the hard ceiling on all API activity within the organization.

<Callout variant="info" title="Fail-open">

Rate limiting depends on Redis. If Redis or authentication is unavailable, the limiter **fails open** — requests pass through rather than being blocked. You should never assume a request was rejected for capacity reasons unless you see an explicit `429`.

</Callout>

## Response headers

The tier limiter sets the following headers, based on the **tightest** layer (the one closest to its limit):

| Header                  | Description                                                               |
| ----------------------- | ------------------------------------------------------------------------- |
| `X-RateLimit-Limit`     | Maximum requests allowed in the current window for the tightest layer.    |
| `X-RateLimit-Remaining` | Requests remaining in the current window for the tightest layer.          |
| `X-RateLimit-Reset`     | Unix timestamp (seconds) when the window resets. Sent on `429` responses. |
| `Retry-After`           | Seconds to wait before retrying. Sent on `429` responses only.            |

<Callout variant="warning" title="Headers come from the tier limiter only">

`X-RateLimit-*` and `Retry-After` headers are emitted by the tier limiter (the per-key / per-user / per-org layers). The per-operation limits described below run inside GraphQL and return a normal `200` response with a `RATE_LIMITED` entry in the `errors` array — they do **not** set these headers or return a `429`.

</Callout>

## Error response

When a tier limit is exceeded, you receive a `429 Too Many Requests` response with this JSON body:

```json
{
  "error": "RATE_LIMITED",
  "message": "Too many requests. Limit: 20 per minute.",
  "retryAfter": 12
}
```

`retryAfter` is the number of seconds until the oldest request in the window ages out, freeing a slot.

## Per-operation limits

Some sensitive operations carry their own limit, independent of the tier limits above. These run inside GraphQL: exceeding one returns a normal HTTP `200` with a `RATE_LIMITED` error in the response's `errors` array (no `429`, no rate-limit headers). The limiter keys on the authenticated user, falling back to the client IP for unauthenticated calls.

| Operation                        | Limit        | Window     |
| -------------------------------- | ------------ | ---------- |
| `signIn`                         | 10 requests  | 60 seconds |
| `socialAuth`                     | 10 requests  | 60 seconds |
| `signInRequest`                  | 10 requests  | 60 seconds |
| `signUpRequest`                  | 10 requests  | 60 seconds |
| `checkAuthMethod`                | 20 requests  | 60 seconds |
| `verifyAcceptInvitation`         | 100 requests | 60 seconds |
| `verifySecurityCode`             | 3 requests   | 60 seconds |
| `createDocument`                 | 5 requests   | 60 seconds |
| `sendTestEmail`                  | 5 requests   | 60 seconds |
| `submitForm`                     | 5 requests   | 60 seconds |
| `exportTodos`                    | 1 request    | 50 seconds |
| `exportReport`                   | 1 request    | 50 seconds |
| `publicViewData`                 | 100 requests | 60 seconds |
| `publicViewRecords`              | 100 requests | 60 seconds |
| `deleteCompany`                  | 3 requests   | 60 seconds |
| `deleteCompanyRequest`           | 3 requests   | 60 seconds |
| `updateEmail`                    | 3 requests   | 60 seconds |
| `updateEmailRequest`             | 3 requests   | 60 seconds |
| `createCommunityPost`            | 5 requests   | 1 hour     |
| `createCommunityReply`           | 30 requests  | 1 hour     |
| `notifyOrgAdminCustomFieldLimit` | 3 requests   | 24 hours   |

`createCommunityPost` is additionally capped at 100 requests per hour across all users combined.

## Best practices

- **Use one API key per integration.** Spreading requests across multiple keys to dodge the per-key limit doesn't help — the per-user and per-org limits still apply.
- **Watch `X-RateLimit-Remaining`.** Slow down before you hit the limit rather than after.
- **Respect `Retry-After`.** On a `429`, wait the specified number of seconds before retrying.
- **Batch with GraphQL.** Fetch multiple resources in a single request instead of many small ones — see [Making Requests](/api/start-guide/making-requests).
- **Add Pro for 2× headroom.** The Pro add-on doubles all three layers on any tier.

For Enterprise limits, [contact support](mailto:support@blue.app).

## Related

- [Making Requests](/api/start-guide/making-requests)
- [Authentication](/api/start-guide/authentication)
- [Error Codes](/api/start-guide/error-codes)
