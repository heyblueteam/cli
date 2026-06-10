---
title: Custom Domains & Email
description: Serve Blue on your own domain and brand the transactional email Blue sends on your behalf.
icon: Globe
order: 0
---

White-label customers can serve Blue on their own domains and brand the email Blue sends on their behalf. This section covers three resources: **custom domains** (point your own hostname at a Blue application type and verify it via DNS/CNAME), **SMTP credentials** (route outbound mail through your own mail server), and **email templates** (override the copy and CTAs of Blue's transactional emails, plus send a test).

These resources map to three API types: custom domains are `CustomDomain` objects, mail-routing settings are `SmtpCredential` objects, and transactional-email overrides are `EmailTemplate` objects. Each lives on an organization (`Company`), selected by the `X-Bloo-Company-ID` header.

## Pages

| Page                                                    | Resource         | Purpose                                                                                                                                                            |
| ------------------------------------------------------- | ---------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| [Custom Domains](/api/domains-email/custom-domains)     | `CustomDomain`   | Point a hostname at a Blue application type (`APPLICATION`, `FORMS`, or `FILES`), provision it on Blue's hosting layer, and verify it with a live DNS/CNAME check. |
| [SMTP Credentials](/api/domains-email/smtp-credentials) | `SmtpCredential` | Route Blue's outbound mail through your own SMTP server, with a live connection test on create and a standalone verify call.                                       |
| [Email Templates](/api/domains-email/email-templates)   | `EmailTemplate`  | Override the subject, body, CTA, and footer of Blue's transactional emails (the invitation email today), and send a one-off test message.                          |

## Authentication and access

Send every request to `https://api.blue.app/graphql` with your [personal access token headers](/api/start-guide/authentication) and an organization context:

```
X-Bloo-Token-ID: YOUR_TOKEN_ID
X-Bloo-Token-Secret: YOUR_TOKEN_SECRET
X-Bloo-Company-ID: YOUR_COMPANY_ID
```

`X-Bloo-Company-ID` accepts an organization ID or its slug. Headers are case-insensitive.

Access splits in two ways across this section:

- **Every write operation** (`createCustomDomain`, `verifyCustomDomain`, `createSmtpCredential`, `createEmailTemplate`, and their update/delete/verify siblings) requires the organization to have the company-level **`white_label`** feature enabled — not just membership. Without it, the mutation fails authorization.
- **The four read queries** — `customDomains`, `smtpCredentials`, `emailTemplate`, and `emailTemplates` — require only that you are a member of the organization (a `CompanyUser` row). They do **not** require the `white_label` feature.

<Callout variant="info" title="White-label is a Pro-plan feature">

White-label is now bundled with the Pro plan rather than sold as a standalone subscription, so there is no white-label billing surface in this section. The deprecated `whiteLabelSubscription` query and the `createWhiteLabelSubscription*URL` mutations are intentionally omitted — billing flows through the standard billing portal.

</Callout>

## Related

- [Custom Domains](/api/domains-email/custom-domains)
- [SMTP Credentials](/api/domains-email/smtp-credentials)
- [Email Templates](/api/domains-email/email-templates)
- [Organization Management](/api/organization-management) — organization context and identifiers
- [Authentication](/api/start-guide/authentication)
