---
title: Email Templates
description: Override the subject, body, and CTA of Blue's transactional emails, and send a one-off test message to preview your copy.
icon: Mail
order: 3
---

White-label organizations can rewrite the copy of the transactional email Blue sends on their behalf. An email template is an `EmailTemplate` object, keyed to one `EmailTemplateType` per organization. You override the pieces you care about — subject line, body, call-to-action, footer — and Blue renders them in place of its defaults.

Today there is exactly one template type: `INVITATION`, the email a new member receives when invited to a workspace. The customization surface is intentionally narrow — this is not a general transactional-email engine, just an override for the invitation copy.

Use [`createEmailTemplate`](#create-a-template) and [`updateEmailTemplate`](#update-a-template) to manage the copy, [`sendTestEmail`](#send-a-test-email) to preview it against sample data, and [`emailTemplate`](#fetch-a-template-by-type) / [`emailTemplates`](#list-templates) to read it back. Branding email this way requires the company-level `white_label` feature for every write; the two read queries only require organization membership.

## Create a template

Use the `createEmailTemplate` mutation to define the override for a template type. It takes the target organization (`companyId`) and the `type`, plus any of the copy fields you want to set. Omitted copy fields fall back to Blue's defaults at render time.

```graphql
mutation CreateEmailTemplate {
  createEmailTemplate(
    input: {
      companyId: "company_123"
      type: INVITATION
      enabled: true
      subject: "You've been invited to {{company}}"
      body: "{{name}}, join your team on {{company}} to start collaborating."
      ctaText: "Accept invitation"
      footer: "Sent by {{company}} via Blue."
    }
  ) {
    id
    type
    enabled
    subject
  }
}
```

The `companyId` argument accepts an organization ID or slug.

### CreateEmailTemplateInput

| Parameter    | Type                 | Required | Description                                                                  |
| ------------ | -------------------- | -------- | ---------------------------------------------------------------------------- |
| `companyId`  | `String!`            | Yes      | The organization that owns the template. ID or slug.                         |
| `type`       | `EmailTemplateType!` | Yes      | Which transactional email to override. Only `INVITATION` is supported today. |
| `enabled`    | `Boolean`            | No       | Whether the override is applied. When `false`, Blue uses its default copy.   |
| `subject`    | `String`             | No       | Subject line.                                                                |
| `body`       | `String`             | No       | Main body text.                                                              |
| `ctaText`    | `String`             | No       | Label of the call-to-action button.                                          |
| `ctaLink`    | `String`             | No       | URL the call-to-action button points at.                                     |
| `footer`     | `String`             | No       | Footer line below the body.                                                  |
| `disclaimer` | `String`             | No       | Fine-print disclaimer line.                                                  |
| `signature`  | `String`             | No       | Sign-off line.                                                               |

### EmailTemplateType

| Value        | Description                                                                 |
| ------------ | --------------------------------------------------------------------------- |
| `INVITATION` | The email sent when a new member is invited to a workspace. The only value. |

## Update a template

Use the `updateEmailTemplate` mutation to change the copy of an existing template. Identify it by `id` (the `type` is fixed at creation and cannot be changed here). Only the fields you pass are written.

```graphql
mutation UpdateEmailTemplate {
  updateEmailTemplate(
    input: { id: "template_123", subject: "Your invitation to {{company}} is ready", enabled: true }
  ) {
    id
    subject
    enabled
    updatedAt
  }
}
```

### UpdateEmailTemplateInput

| Parameter    | Type      | Required | Description                              |
| ------------ | --------- | -------- | ---------------------------------------- |
| `id`         | `String!` | Yes      | The `EmailTemplate` to update.           |
| `enabled`    | `Boolean` | No       | Whether the override is applied.         |
| `name`       | `String`  | No       | Internal label for the template.         |
| `subject`    | `String`  | No       | Subject line.                            |
| `body`       | `String`  | No       | Main body text.                          |
| `ctaText`    | `String`  | No       | Label of the call-to-action button.      |
| `ctaLink`    | `String`  | No       | URL the call-to-action button points at. |
| `footer`     | `String`  | No       | Footer line below the body.              |
| `disclaimer` | `String`  | No       | Fine-print disclaimer line.              |
| `signature`  | `String`  | No       | Sign-off line.                           |

## Delete a template

Use the `deleteEmailTemplate` mutation to remove an override by `id`. Once deleted, the corresponding email reverts to Blue's default copy. This mutation returns a bare `Boolean` (`true` on success), so it takes no sub-selection.

```graphql
mutation DeleteEmailTemplate {
  deleteEmailTemplate(id: "template_123")
}
```

```json
{ "data": { "deleteEmailTemplate": true } }
```

## Send a test email

Use the `sendTestEmail` mutation to dispatch a one-off rendering of a template so you can preview your copy before it goes live. Blue renders the template with **sample data** — the recipient's first name (falling back to `"User"`), their email, the organization's name, and the name of the organization's first workspace — then sends it as a real message to `email`. It returns `Boolean` (`true` once dispatched).

```graphql
mutation SendTestEmail {
  sendTestEmail(input: { email: "you@example.com", emailTemplateId: "template_123" })
}
```

```json
{ "data": { "sendTestEmail": true } }
```

### SendTestEmailInput

| Parameter         | Type      | Required | Description                          |
| ----------------- | --------- | -------- | ------------------------------------ |
| `email`           | `String!` | Yes      | Address to send the test message to. |
| `emailTemplateId` | `String!` | Yes      | The `EmailTemplate` to render.       |

<Callout variant="info" title="Test renders use placeholder data">

The test email is built from sample values, not a live invitation: the recipient name comes from the Blue user that owns `email` (defaulting to `"User"` if that address has no account), and the company and project names are taken from the template's organization and its first workspace. The message is tagged internally as a test send. Use it to check formatting and copy, not to validate a specific recipient's real invitation.

</Callout>

## Fetch a template by type

Use the `emailTemplate` query to read the single template configured for the context organization by `type`. The organization is taken from your `blue-org-id` header. Returns `null` if no override exists for that type. This query requires only organization membership — not the `white_label` feature.

```graphql
query InvitationTemplate {
  emailTemplate(type: INVITATION) {
    id
    type
    enabled
    subject
    body
    ctaText
    ctaLink
    footer
  }
}
```

```json
{
  "data": {
    "emailTemplate": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "type": "INVITATION",
      "enabled": true,
      "subject": "You've been invited to {{company}}",
      "body": "{{name}}, join your team on {{company}} to start collaborating.",
      "ctaText": "Accept invitation",
      "ctaLink": null,
      "footer": "Sent by {{company}} via Blue."
    }
  }
}
```

## List templates

Use the `emailTemplates` query to page through the templates for the context organization, optionally filtered by `type`. Like `emailTemplate`, it requires only organization membership.

```graphql
query EmailTemplates {
  emailTemplates(filter: { type: INVITATION }, skip: 0, take: 20) {
    items {
      id
      type
      enabled
      subject
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

```json
{
  "data": {
    "emailTemplates": {
      "items": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "type": "INVITATION",
          "enabled": true,
          "subject": "You've been invited to {{company}}"
        }
      ],
      "pageInfo": {
        "totalItems": 1,
        "totalPages": 1,
        "page": 1,
        "perPage": 20,
        "hasNextPage": false,
        "hasPreviousPage": false
      }
    }
  }
}
```

### Arguments

| Argument | Type                       | Default | Description                                                   |
| -------- | -------------------------- | ------- | ------------------------------------------------------------- |
| `filter` | `EmailTemplateFilterInput` | —       | Optional filter. Set `type` to restrict to one template type. |
| `skip`   | `Int`                      | `0`     | Number of items to skip.                                      |
| `take`   | `Int`                      | `20`    | Number of items to return.                                    |

### EmailTemplateFilterInput

| Field  | Type                | Description                            |
| ------ | ------------------- | -------------------------------------- |
| `type` | `EmailTemplateType` | Restrict results to one template type. |

## The EmailTemplate type

| Field        | Type                 | Description                                                                |
| ------------ | -------------------- | -------------------------------------------------------------------------- |
| `id`         | `ID!`                | Unique identifier.                                                         |
| `uid`        | `String!`            | Short public identifier.                                                   |
| `type`       | `EmailTemplateType!` | Which transactional email this overrides (`INVITATION`).                   |
| `enabled`    | `Boolean`            | Whether the override is applied. When `false`, Blue uses its default copy. |
| `subject`    | `String`             | Subject line.                                                              |
| `body`       | `String`             | Main body text.                                                            |
| `ctaText`    | `String`             | Label of the call-to-action button.                                        |
| `ctaLink`    | `String`             | URL the call-to-action button points at.                                   |
| `footer`     | `String`             | Footer line below the body.                                                |
| `disclaimer` | `String`             | Fine-print disclaimer line.                                                |
| `signature`  | `String`             | Sign-off line.                                                             |
| `createdAt`  | `DateTime`           | Creation timestamp.                                                        |
| `updatedAt`  | `DateTime`           | Last-update timestamp.                                                     |

The `EmailTemplatePagination` type returned by [`emailTemplates`](#list-templates) has the standard shape: `items` (a list of `EmailTemplate`) and `pageInfo` (a [`PageInfo`](/api/start-guide/capabilities)).

## Errors

| Code                       | When                                                                                                     |
| -------------------------- | -------------------------------------------------------------------------------------------------------- |
| `EMAIL_TEMPLATE_NOT_FOUND` | `updateEmailTemplate`, `deleteEmailTemplate`, or `sendTestEmail` references an `id` that does not exist. |
| `PRO_REQUIRED`             | A write was attempted on an organization without the `white_label` feature.                              |
| `FORBIDDEN`                | The caller is not an `OWNER`/`ADMIN` of the organization (writes), or not a member of it at all (reads). |

## Permissions

The two read queries and the four write operations have different access rules:

- **Writes** — `createEmailTemplate`, `updateEmailTemplate`, `deleteEmailTemplate`, and `sendTestEmail` require that you are an `OWNER` or `ADMIN` of the organization **and** that the organization has the `white_label` feature. For `update`, `delete`, and `sendTestEmail` the organization is resolved from the template's own `id`; for `create` it is the `companyId` you pass.
- **Reads** — `emailTemplate` and `emailTemplates` require only that you are a member of the context organization (any `CompanyUser` row). They do **not** require `white_label`, so non-white-label members can inspect whether overrides exist.

White-label is bundled with the Pro plan — there is no standalone white-label SKU or subscription to manage through the API. Billing flows through the [standard billing portal](/api/organization-management).

## Related

- [Custom Domains & Email overview](/api/domains-email)
- [Custom Domains](/api/domains-email/custom-domains)
- [SMTP Credentials](/api/domains-email/smtp-credentials)
- [Authentication](/api/start-guide/authentication)
