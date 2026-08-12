---
title: Error Codes
description: How the Blue GraphQL API reports errors, the safe vs. masked exposure rules, and a complete reference of every error code.
icon: AlertTriangle
order: 99
---

When a Blue GraphQL operation fails, the response contains a top-level `errors` array instead of (or alongside) `data`. Each entry carries a human-readable `message` and a machine-readable code at `extensions.code`. Switch on `extensions.code` in your client — never parse the `message` string, which is wording that can change.

Blue throws **123 intentional error codes** from `api/src/lib/errors.ts`. They are listed in full below, grouped by area. Two codes are not in that file but you will still see them: `BAD_USER_INPUT` (input validation) and `INTERNAL_SERVER_ERROR` (the masked code applied to unexpected server failures — see [Safe vs. masked errors](#safe-vs-masked-errors)).

## Error response shape

```json
{
  "errors": [
    {
      "message": "Todo was not found.",
      "path": ["updateTodo"],
      "extensions": {
        "code": "TODO_NOT_FOUND"
      }
    }
  ],
  "data": null
}
```

Each error object includes:

- **`message`** — a human-readable description. For display only; do not branch on it.
- **`extensions.code`** — the stable, machine-readable code. This is what you handle programmatically.
- **`path`** — the response path of the field that failed (e.g. the mutation name).
- **`extensions.data`** — present only on some codes; carries structured detail (see below).

A GraphQL response can succeed partially: `data` may hold results for the fields that resolved while `errors` describes the ones that didn't. Always check `errors` even when `data` is non-null.

## Structured error detail (`extensions.data`)

Some errors attach a structured payload so your client can react precisely instead of showing a generic toast.

### `BAD_USER_INPUT`

Thrown for input that fails validation (Blue's `UserInputError`). The validation detail is carried under `extensions.data`:

```json
{
  "errors": [
    {
      "message": "Invalid input",
      "extensions": {
        "code": "BAD_USER_INPUT",
        "data": { "field": "email", "reason": "must be a valid email address" }
      }
    }
  ]
}
```

The `data` shape varies by what was validated — treat it as an opaque object keyed by the offending input. `FORMULA_VALIDATION_ERROR` is a specialization of this error for name-formula validation and uses the same `BAD_USER_INPUT` mechanism with its own `code`.

### `PLAN_LIMIT_REACHED`

This is the **single signal for anything blocked by your plan or tier** — over a resource cap, a feature your plan doesn't include, a locked workspace, or a tier-level rate limit. There are no per-resource limit codes; route on the `kind` inside `extensions`:

```json
{
  "errors": [
    {
      "message": "You've reached the workspace limit for your plan.",
      "extensions": {
        "code": "PLAN_LIMIT_REACHED",
        "kind": "over_limit",
        "resource": "workspaces",
        "current": 10,
        "limit": 10
      }
    }
  ]
}
```

| Field            | When present                 | Meaning                                                          |
| ---------------- | ---------------------------- | ---------------------------------------------------------------- |
| `kind`           | always                       | One of `over_limit`, `missing_feature`, `locked`, `rate_limited` |
| `resource`       | `over_limit`, `rate_limited` | The resource at its cap, e.g. `workspaces`, `records`, `users`   |
| `current`        | `over_limit`                 | Current count                                                    |
| `limit`          | `over_limit`                 | The tier cap for that resource                                   |
| `feature`        | `missing_feature`            | The feature-flag key the plan is missing                         |
| `cooldownEndsAt` | `rate_limited`               | ISO timestamp when the action becomes available again            |

<Callout variant="info" title="Plan limits vs. permissions">

`PLAN_LIMIT_REACHED` means "your plan can't do this — upgrade." `FORBIDDEN` means "your role isn't allowed to do this." They are deliberately distinct so a client can show an upgrade prompt in one case and an access message in the other.

</Callout>

## Safe vs. masked errors

In production Blue runs every error through a formatter (`api/src/lib/format-error.ts`) before returning it. The rule is simple:

- An error is **safe** when the server _intentionally_ threw a `GraphQLError` carrying an explicit code other than `INTERNAL_SERVER_ERROR`. Every code in this page's tables is thrown that way, so its real `message` and `code` reach the client unchanged. There is **no** special-casing of `*_NOT_FOUND` — those are safe only because they're thrown as coded `GraphQLError`s, exactly like every other code here.
- An error is **unsafe** when it wasn't thrown intentionally — an uncaught exception, or any error Apollo auto-wraps. The `message` is replaced with `"Internal server error"` and the `code` becomes `INTERNAL_SERVER_ERROR` (the original code is dropped). The full detail is logged server-side, never sent to the client.
- **Prisma / database errors** are a special unsafe case: the client receives the generic message `"A database error occurred"` with code `INTERNAL_SERVER_ERROR`.

```json
{
  "errors": [
    {
      "message": "Internal server error",
      "extensions": { "code": "INTERNAL_SERVER_ERROR" }
    }
  ]
}
```

Treat `INTERNAL_SERVER_ERROR` as "retry or contact support" — it never carries actionable detail. In non-production environments the formatter is bypassed and the full original error (including stack traces) is returned for debugging.

## Handling errors in your client

Switch on `extensions.code`, with a default branch for anything unrecognized:

```javascript
const errors = response.errors ?? []

for (const err of errors) {
  switch (err.extensions?.code) {
    case 'UNAUTHENTICATED':
      // Token missing or invalid — re-authenticate.
      break
    case 'FORBIDDEN':
      // Authenticated, but the role can't perform this action.
      break
    case 'PLAN_LIMIT_REACHED':
      // Blocked by plan/tier — inspect extensions.kind and show an upgrade path.
      break
    case 'BAD_USER_INPUT':
      // Validation failed — read extensions.data for the offending field.
      break
    case 'RATE_LIMITED':
      // Back off and retry (see Rate Limits).
      break
    case 'TODO_NOT_FOUND':
      // The record doesn't exist or you can't access it.
      break
    default:
    // INTERNAL_SERVER_ERROR or any unhandled code — show a generic message and log it.
  }
}
```

## Rate limiting

Most operations are governed by the per-tier rate-limit middleware documented on [Rate Limits](/api/start-guide/rate-limits). A handful of sensitive or expensive operations have their own per-operation caps enforced by `graphql-rate-limit` (keyed by user, falling back to IP). Exceeding any of these returns the `RATE_LIMITED` code.

| Rule              | Window | Max | Applies to                                                                                                     |
| ----------------- | ------ | --- | -------------------------------------------------------------------------------------------------------------- |
| Default           | 60s    | 5   | Per-operation guard on assorted authenticated mutations (e.g. `submitForm`, `sendTestEmail`, `createDocument`) |
| Request           | 60s    | 3   | High-impact account actions (e.g. `deleteOrganization`, `updateEmail`, `verifySecurityCode`)                   |
| Sign-in           | 60s    | 10  | `signIn`, `socialAuth`                                                                                         |
| Sign-in request   | 60s    | 10  | `signInRequest`, `signUpRequest`                                                                               |
| Export            | 50s    | 1   | `exportRecords`, `exportReport`                                                                                |
| Public view       | 60s    | 100 | `publicViewData`, `publicViewRecords`                                                                          |
| Accept invitation | 60s    | 100 | `verifyAcceptInvitation`                                                                                       |
| Check auth method | 60s    | 20  | `checkAuthMethod`                                                                                              |
| Notify org admin  | 24h    | 3   | `notifyOrgAdminCustomFieldLimit`                                                                               |

These per-operation limits sit on top of the tier limits — a request must pass both. See [Rate Limits](/api/start-guide/rate-limits) for the tier middleware and recommended backoff.

## Error code reference

Every code below is thrown by the API. Messages are the exact strings the server returns (some are templated with context, indicated by `…`).

### Authentication & access

| Code                    | Message                                          |
| ----------------------- | ------------------------------------------------ |
| `UNAUTHENTICATED`       | You are not authenticated.                       |
| `FORBIDDEN`             | You are not authorized.                          |
| `PLAN_LIMIT_REACHED`    | _(varies; see structured detail above)_          |
| `PRO_REQUIRED`          | This feature requires a Pro subscription.        |
| `COMPANY_BANNED`        | Company is banned                                |
| `NOT_AVAILABLE_COUNTRY` | This service is currently not available in your country. |
| `RATE_LIMITED`          | Too many requests, please try again later.       |

### Validation & input

| Code                             | Message                                          |
| -------------------------------- | ------------------------------------------------ |
| `BAD_USER_INPUT`                 | _(varies; carries `extensions.data`)_            |
| `FORMULA_VALIDATION_ERROR`       | _(name-formula validation message)_              |
| `BAD_EMAIL`                      | Whoops, something wrong with your email address. |
| `MISSING_PARAM_ERROR`            | missing param                                    |
| `MISSING_PROP_ERROR`             | Missing the `…` property.                        |
| `CUSTOM_FIELD_VALUE_PARSE_ERROR` | Unable to parse custom field value.              |
| `TITLE_COLUMN_NOT_FOUND`         | No title column found.                           |

### Sign-in, sessions & invitations

| Code                              | Message                                                                                             |
| --------------------------------- | --------------------------------------------------------------------------------------------------- |
| `EMAIL_IS_TAKEN`                  | User is already existed with this email.                                                            |
| `EMAIL_NOT_REGISTERED`            | This email is not registered.                                                                       |
| `DISPOSABLE_EMAIL_NOT_ALLOWED`    | Temporary or disposable email addresses are not allowed.                                            |
| `PENDING_INVITATION`              | You have a pending invitation… Please check your email and accept the invitation before signing in. |
| `INVITATION_NOT_FOUND`            | The invitation has expired or no longer exists.                                                     |
| `INVITATION_EXPIRED`              | Invitation has expired.                                                                             |
| `UNABLE_TO_ACCEPT_INVITATION`     | Unable to accept the invitation.                                                                    |
| `SECURITY_CODE_NOT_FOUND`         | Security code was not found.                                                                        |
| `SECURITY_CODE_EXPIRED`           | Security code has expired.                                                                          |
| `SECURITY_CODE_INCORRECT`         | Security code is incorrect.                                                                         |
| `SOCIAL_TOKEN_INVALID`            | Social login token is invalid or expired                                                            |
| `SOCIAL_EMAIL_NOT_PROVIDED`       | Email not provided by social provider                                                               |
| `SOCIAL_EMAIL_NOT_VERIFIED`       | Email not verified by social provider                                                               |
| `SOCIAL_EMAIL_NOT_REGISTERED`     | This email is not registered.                                                                       |
| `SOCIAL_UID_ALREADY_LINKED`       | This social account is already linked to another user                                               |
| `OAUTH_CONNECTION_NOT_FOUND`      | OAuth connection was not found.                                                                     |
| `PERSONAL_ACCESS_TOKEN_NOT_FOUND` | Personal access token was not found.                                                                |

### Users & membership

| Code                            | Message                                               |
| ------------------------------- | ----------------------------------------------------- |
| `USER_NOT_FOUND`                | User was not found.                                   |
| `USER_ALREADY_IN_THE_COMPANY`   | User is already in the company.                       |
| `USER_ALREADY_IN_THE_PROJECT`   | User is already in the project.                       |
| `USER_NOT_IN_COMPANY`           | User is not in the company.                           |
| `USER_NOT_IN_PROJECT`           | User is not in the project.                           |
| `ASSIGNEE_NOT_FOUND`            | Assignee was not found.                               |
| `ADD_SELF`                      | You are not allowed to add yourself.                  |
| `ADDED_ALREADY`                 | You are not allowed to add again.                     |
| `REMOVE_SELF`                   | You are not allowed to remove yourself.               |
| `UPDATE_LEVEL_SELF`             | You are not allowed to update level for yourself.     |
| `UPDATE_COMPANY_LEVEL_TO_OWNER` | You are not allowed to update anyone to be the owner. |
| `COMPANY_USER_NOT_FOUND`        | Company user was not found.                           |
| `PROJECT_USER_ROLE_NOT_FOUND`   | Project user role was not found.                      |
| `PROJECT_USER_ROLE_LIMIT`       | Project user role limit reached.                      |

### Records, lists & dependencies

| Code                                | Message                                                             |
| ----------------------------------- | ------------------------------------------------------------------- |
| `TODO_NOT_FOUND`                    | Todo was not found.                                                 |
| `TODO_LIST_NOT_FOUND`               | Todo list was not found.                                            |
| `TODO_ALREADY_IN_LIST`              | Todo is already in the specified todo list.                         |
| `TODO_ALREADY_IN_PROJECT`           | Todo is already associated with the target project.                 |
| `TODO_DEPENDENCY_ALREADY_EXISTS`    | Todo dependency already exists                                      |
| `TODO_LIST_CREATE_TODO_LIMIT_ERROR` | Todo list has reached the maximum number of todos.                  |
| `TODO_UDPATE_LIMIT`                 | You are updating too many todos                                     |
| `TITLE_EDIT_NOT_ALLOWED`            | Cannot edit title when name formula is enabled.                     |
| `FIELD_REFERENCED_BY_FORMULA`       | Cannot delete field "`…`" - it is used in the project name formula. |
| `FORMULA_RECOMPUTATION_IN_PROGRESS` | A name formula recomputation is already in progress.                |

### Workspaces, organizations & limits

| Code                         | Message                                                                                  |
| ---------------------------- | ---------------------------------------------------------------------------------------- |
| `PROJECT_NOT_FOUND`          | Project was not found.                                                                   |
| `PROJECT_URL_ALREADY_EXISTS` | Project URL already exists.                                                              |
| `CREATE_PROJECT_LIMIT`       | Project size exceeds maximum limit.                                                      |
| `COMPANY_NOT_FOUND`          | Company was not found.                                                                   |
| `COMPANY_URL_ALREADY_EXISTS` | Company URL already exists.                                                              |
| `COMPANY_LICENSE_NOT_FOUND`  | Company license was not found.                                                           |
| `CUSTOM_FIELD_LIMIT`         | This workspace has reached the 30 custom field limit. Upgrade to Enterprise to add more. |
| `TOO_MANY_LINKS`             | The company has too many links.                                                          |

### Custom fields, tags & templates

| Code                            | Message                            |
| ------------------------------- | ---------------------------------- |
| `CUSTOM_FIELD_NOT_FOUND`        | Custom field was not found.        |
| `CUSTOM_FIELD_OPTION_NOT_FOUND` | Custom field option was not found. |
| `TAG_NOT_FOUND`                 | Tag was not found                  |
| `TEMPLATE_NOT_FOUND`            | Template was not found.            |
| `SAVED_VIEW_NOT_FOUND`          | Saved view was not found.          |
| `PUBLIC_VIEW_NOT_FOUND`         | Public view was not found.         |

### Comments, activity & notifications

| Code                            | Message                            |
| ------------------------------- | ---------------------------------- |
| `COMMENT_NOT_FOUND`             | Comment was not found.             |
| `DISCUSSION_NOT_FOUND`          | Discussion was not found.          |
| `MENTION_NOT_FOUND`             | Mention not found                  |
| `ACTIVITY_NOT_FOUND`            | Activity was not found.            |
| `STATUS_UPDATE_NOT_FOUND`       | Status update was not found.       |
| `NOTIFICATION_OPTION_NOT_FOUND` | Notification option was not found. |
| `CHAT_NOT_FOUND`                | Chat was not found.                |

### Checklists

| Code                       | Message                       |
| -------------------------- | ----------------------------- |
| `CHECKLIST_NOT_FOUND`      | Checklist was not found.      |
| `CHECKLIST_ITEM_NOT_FOUND` | Checklist item was not found. |

### Automations & webhooks

| Code                   | Message                   |
| ---------------------- | ------------------------- |
| `AUTOMATION_NOT_FOUND` | Automation was not found. |
| `WEBHOOK_NOT_FOUND`    | Webhook was not found.    |

### Dashboards & charts

| Code                            | Message                            |
| ------------------------------- | ---------------------------------- |
| `DASHBOARD_NOT_FOUND`           | Dashboard was not found.           |
| `CHART_NOT_FOUND`               | Chart was not found.               |
| `CHART_ALREADY_EXPORTING`       | Chart is already being exported.   |
| `CHART_SEGMENT_NOT_FOUND`       | Chart segment was not found.       |
| `CHART_SEGMENT_VALUE_NOT_FOUND` | Chart segment value was not found. |
| `REPORT_NOT_FOUND`              | Report was not found.              |

### Files, folders & documents

| Code                                | Message                                                    |
| ----------------------------------- | ---------------------------------------------------------- |
| `FILE_NOT_FOUND`                    | File was not found.                                        |
| `UNABLE_TO_DELETE_FILE`             | Unable to delete the file.                                 |
| `FILES_DISABLED_FOR_ROLE`           | File uploads are disabled for your role in this workspace. |
| `FOLDER_NOT_FOUND`                  | Folder was not found.                                      |
| `DOCUMENT_NOT_FOUND`                | Document not found.                                        |
| `PORTABLE_DOCUMENT_NOT_FOUND`       | Portable document was not found.                           |
| `PORTABLE_DOCUMENT_FIELD_NOT_FOUND` | Portable document field was not found.                     |

### Forms

| Code                 | Message                  |
| -------------------- | ------------------------ |
| `FORM_NOT_FOUND`     | Form was not found.      |
| `FORM_TOKEN_INVALID` | Form token is not valid. |

### Imports

| Code                         | Message                                                                                                |
| ---------------------------- | ------------------------------------------------------------------------------------------------------ |
| `IMPORT_ALREADY_IN_PROGRESS` | An import is already running for this workspace. Please wait for it to finish before starting another. |

An import that would exceed your plan's per-workspace record cap is rejected before it starts with `PLAN_LIMIT_REACHED` (`resource: "records"`) — see that section above for the full error shape. There's no import-specific limit code; the cap scales with your plan, same as everywhere else records are counted.

### Custom domains & email

| Code                                    | Message                                                                     |
| --------------------------------------- | --------------------------------------------------------------------------- |
| `CUSTOM_DOMAIN_NOT_FOUND`               | Custom domain `…` was not found.                                            |
| `CUSTOM_DOMAIN_ALREADY_EXISTS`          | Domain already exists.                                                      |
| `CUSTOM_DOMAIN_IN_USE`                  | Domain is already in use.                                                   |
| `CUSTOM_DOMAIN_CREATE_ERROR`            | Couldn't set up `…`. This usually means its DNS isn't pointing to Blue yet… |
| `SMTP_CREDENTIAL_NOT_FOUND`             | SMTP Credential was not found.                                              |
| `SMTP_CREDENTIAL_INVALID`               | Invalid SMTP Credential.                                                    |
| `SMTP_CREDENTIAL_RESERVED_SENDER_NAME`  | Reserved sender name.                                                       |
| `SMTP_CREDENTIAL_RESERVED_SENDER_EMAIL` | Reserved sender email.                                                      |
| `EMAIL_TEMPLATE_NOT_FOUND`              | Email template was not found.                                               |
| `APPLICATION_TYPE_NOT_FOUND`            | Application type was not found.                                             |

### Billing & subscriptions

| Code                                             | Message                                               |
| ------------------------------------------------ | ----------------------------------------------------- |
| `PAYMENT_REQUIRED`                               | Whoops, something went wrong. Try reloading the page. |
| `SUBSCRIPTION_PLAN_MISSING`                      | Subscription plan is missing.                         |
| `COMPANY_SUBSCRIPTION_PLAN_NOT_FOUND`            | The company does not have a subscription plan.        |
| `COMPANY_SUBSCRIPTION_PLAN_ALREADY_EXIST`        | The company already has an active subscription plan.  |
| `COMPANY_SUBSCRIPTION_PLAN_PROMO_CODE_NOT_FOUND` | Invalid promo code                                    |
| `UPGRADE_TO_SAME_PLAN_ERROR`                     | Cannot upgrade to the same plan.                      |
| `CREATE_CHECKOUT_SESSION_FAILED`                 | Unable to create a checkout session.                  |
| `CREATE_STRIPE_CUSTOMER_ERROR`                   | Failed to create a stripe customer.                   |
| `RETRIEVE_STRIPE_CUSTOMER_ERROR`                 | Failed to retrieve a stripe customer.                 |
| `UPDATE_STRIPE_CUSTOMER_ERROR`                   | Failed to update a stripe customer.                   |
| `CREATE_STRIPE_SUBSCRIPTION_ERROR`               | Failed to create a stripe subscription.               |
| `RETRIEVE_STRIPE_SUBSCRIPTION_ERROR`             | Failed to retrieve a stripe subscription.             |
| `UPDATE_STRIPE_SUBSCRIPTION_ERROR`               | Failed to update stripe subscription.                 |
| `RETRIEVE_STRIPE_SUBSCRIPTION_PLAN_ERROR`        | Failed to retrieve a stripe subscription plan.        |
| `RETRIEVE_STRIPE_PRICE_ERROR`                    | Failed to retrieve a stripe price.                    |

### System

| Code                    | Message                                                                                      |
| ----------------------- | -------------------------------------------------------------------------------------------- |
| `RESOURCE_NOT_FOUND`    | Resource was not found.                                                                      |
| `INTERNAL_SERVER_ERROR` | Internal server error. _(masked code; see [Safe vs. masked errors](#safe-vs-masked-errors))_ |

## Related

- [Authentication](/api/start-guide/authentication) — how to authenticate API requests
- [Making Requests](/api/start-guide/making-requests) — request format, headers, and the GraphQL endpoint
- [Rate Limits](/api/start-guide/rate-limits) — per-tier limits and backoff guidance
