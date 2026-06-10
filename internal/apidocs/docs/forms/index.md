---
title: Forms
description: Build public forms with the GraphQL API — create a form, configure its fields, and collect submissions that become records in a workspace.
icon: ClipboardList
order: 1
---

A **form** is a public page that collects responses and turns each submission into a record in a workspace. Forms are `Form` objects scoped to a workspace (`Project` in the API); each visible input is a `FormField`, and a submission creates a `Todo` (record) in the form's target list.

There are two sides to the API:

- **Building & managing forms** (authenticated) — create a form, add and order its fields, theme it, and activate it. Requires project-level `OWNER`, `ADMIN`, or `MEMBER`.
- **Submitting a form** (public) — `submitForm` is unauthenticated and rate-limited, so a respondent never needs an account or token header. This is what Blue's hosted form page calls.

All authenticated examples use personal access token headers and target `https://api.blue.app`. See [Authentication](/api/start-guide/authentication) for how to generate a token.

## List forms

`forms` returns a paginated list of the forms in one workspace.

```graphql
query ListForms {
  forms(filter: { projectId: "project_123" }, sort: updatedAt_DESC, take: 20) {
    items {
      id
      title
      isActive
    }
  }
}
```

| Argument | Type              | Required | Description                                                  |
| -------- | ----------------- | -------- | ------------------------------------------------------------ |
| `filter` | `FormFilterInput` | No       | `{ projectId }` — scope to one workspace (its `id`, a CUID). |
| `sort`   | `FormSort`        | No       | `updatedAt_DESC` (default) or `title_ASC`.                   |
| `skip`   | `Int`             | No       | Offset for pagination. Default `0`.                          |
| `take`   | `Int`             | No       | Page size. Default `20`.                                     |

The result is a `FormPagination` with `items` (an array of `Form`) and `pageInfo`.

## Get one form

`form` fetches a single form — including its ordered `formFields` — by `id`.

```graphql
query GetForm {
  form(id: "form_123") {
    id
    title
    description
    isActive
    primaryColor
    formFields {
      id
      field
      name
      required
      position
    }
  }
}
```

## Create a form

`createForm` creates an empty form in a workspace. Only `projectId` is needed to start; everything else has sensible defaults you can fill in later with [`updateForm`](#update-a-form).

```graphql
mutation CreateForm {
  createForm(
    input: {
      projectId: "project_123"
      title: "Contact us"
      description: "Tell us how we can help."
      primaryColor: "#2563eb"
      hideBranding: false
    }
  ) {
    id
    title
    isActive
  }
}
```

### CreateFormInput

| Field          | Type        | Required | Description                                       |
| -------------- | ----------- | -------- | ------------------------------------------------- |
| `projectId`    | `String`    | Yes      | Workspace `id` (CUID) the form belongs to.        |
| `title`        | `String`    | No       | Form title shown to respondents.                  |
| `description`  | `String`    | No       | Intro text under the title.                       |
| `primaryColor` | `String`    | No       | Accent color, hex (e.g. `#2563eb`).               |
| `hideBranding` | `Boolean`   | No       | Hide the "Powered by Blue" footer.                |
| `assigneeIds`  | `[String!]` | No       | Users auto-assigned to records this form creates. |

A new form is **not active** until you set `isActive: true` with `updateForm`.

## Add and update fields

A form's inputs are `FormField` rows. `upsertFormField` both creates and edits a field: pass an existing `formFieldId` to update it, or a new id to create one. `field` selects what the input maps to — a built-in record property or a custom field.

```graphql
mutation AddNameField {
  upsertFormField(
    input: {
      formId: "form_123"
      formFieldId: "field_new_1"
      field: title
      name: "Your name"
      placeholder: "Ada Lovelace"
      required: true
      position: 1
    }
  ) {
    id
    field
    name
    required
    position
  }
}
```

### UpsertFormFieldInput

| Field              | Type               | Required             | Description                                                    |
| ------------------ | ------------------ | -------------------- | -------------------------------------------------------------- |
| `formId`           | `String!`          | Yes                  | The form to add the field to.                                  |
| `formFieldId`      | `String!`          | Yes                  | Existing field id to update, or a new id to create.            |
| `field`            | `FormFieldsField!` | Yes                  | What the input maps to (see below).                            |
| `customFieldId`    | `String`           | When `field: custom` | The workspace custom field this input writes to.               |
| `name`             | `String`           | No                   | Label shown above the input.                                   |
| `placeholder`      | `String`           | No                   | Placeholder text inside the input.                             |
| `position`         | `Float`            | No                   | Order within the form (ascending).                             |
| `required`         | `Boolean`          | No                   | Whether a response is mandatory.                               |
| `hidden`           | `Boolean`          | No                   | Hide the field from the public form (e.g. a pre-filled value). |
| `addToDescription` | `Boolean`          | No                   | Append the field's value to the created record's description.  |
| `extraInfo`        | `String`           | No                   | Helper text shown under the input.                             |

`field` accepts: `title`, `description`, `tags`, `assignees`, `startedAt`, `duedAt`, or `custom`. Use `custom` together with `customFieldId` to collect a value for one of the workspace's custom fields.

List a form's fields with `formFields`:

```graphql
query FormFields {
  formFields(filter: { formId: "form_123" }) {
    id
    field
    name
    required
    position
  }
}
```

Remove a field with `deleteFormField(id: "field_123")`, which returns `Boolean`.

## Update a form

`updateForm` edits any form property and is how you **activate** a form (`isActive: true`) once its fields are in place.

```graphql
mutation ActivateForm {
  updateForm(
    input: {
      id: "form_123"
      isActive: true
      submitText: "Send"
      responseText: "Thanks — we'll be in touch."
      redirectURL: "https://example.com/thank-you"
    }
  ) {
    id
    isActive
  }
}
```

`UpdateFormInput` covers the visible form: `title`, `description`, `footerText`, `showFooter`, `submitText`, `responseText`, `theme`, `primaryColor`, `hideBranding`, `imageURL`, `redirectURL`, `isActive`, plus `assigneeIds`, `tagIds`, and `todoListId` (the list new records land in).

<Callout variant="info" title="Manage fields with upsertFormField">

`UpdateFormInput` also has a `formFields` array, but the supported way to add, edit, and order fields is one [`upsertFormField`](#add-and-update-fields) call per field — it's explicit about create-vs-update and field mapping.

</Callout>

## Copy a form

`copyForm` duplicates a form — and its fields — into a workspace.

```graphql
mutation CopyForm {
  copyForm(input: { formId: "form_123", projectId: "project_456", title: "Contact us (copy)" }) {
    id
    title
  }
}
```

`projectId` is required and may be the same workspace or a different one. `title` is optional; omit it to keep the source title.

## Delete a form

`deleteForm(id: "form_123")` deletes a form and returns `Boolean`.

```graphql
mutation DeleteForm {
  deleteForm(id: "form_123")
}
```

## Submit a form

`submitForm` is **public** — no authentication header, rate-limited. Each submission creates a record in the form's target list. This is the call Blue's hosted form page makes; use it directly only when you render your own form UI.

```graphql
mutation SubmitForm {
  submitForm(
    input: {
      formId: "form_123"
      formToken: "form_123_token"
      formFieldInput: [
        { formFieldId: "field_1", value: "Ada Lovelace" }
        { formFieldId: "field_2", value: "ada@example.com" }
      ]
    }
  )
}
```

### SubmitFormInput

| Field            | Type                       | Required | Description                   |
| ---------------- | -------------------------- | -------- | ----------------------------- |
| `formId`         | `String!`                  | Yes      | The form being submitted.     |
| `formToken`      | `String!`                  | Yes      | The form's submission token.  |
| `formFieldInput` | `[SubmitFormFieldInput!]!` | Yes      | One entry per answered field. |

Each `SubmitFormFieldInput` is `{ formFieldId, customFieldId, value }`, where `value` is `JSON` (a string, number, or structured value matching the field). The mutation returns `Boolean` — `true` once the record is created.

<Callout variant="info" title="File uploads on public forms">

To accept files on a public form, upload to the same presigned-PUT endpoint without a token by passing `?formId=<form-id>` — the upload is gated by the form being active. See [Anonymous form uploads](/api/files#anonymous-form-uploads).

</Callout>

## Watch for changes

`subscribeToForm` streams create/update/delete events for the forms in a workspace.

```graphql
subscription WatchForms {
  subscribeToForm(filter: { projectId: "project_123" }) {
    mutation
    node {
      id
      title
      isActive
    }
    updatedFields
  }
}
```

`mutation` is `CREATED`, `UPDATED`, or `DELETED`; `node` is the affected `Form`.

## Permissions

| Operation                                                                                | Requirement                                                   |
| ---------------------------------------------------------------------------------------- | ------------------------------------------------------------- |
| `form`, `forms`, `createForm`, `updateForm`, `copyForm`, `deleteForm`, `upsertFormField` | Project-level `OWNER`, `ADMIN`, or `MEMBER` on the workspace. |
| `deleteFormField`                                                                        | Any authenticated user.                                       |
| `submitForm`                                                                             | Public — no authentication, rate-limited.                     |

## Related

- [Uploading files](/api/files) — accept file uploads on public forms.
- [Custom fields](/api/custom-fields) — back a form field with a workspace custom field.
- [Records](/api/records) — what a submission creates.
- [Error codes](/api/start-guide/error-codes)
