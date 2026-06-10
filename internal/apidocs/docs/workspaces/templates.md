---
title: Templates
description: Reuse a workspace's structure by saving it as a template and creating new workspaces from it, with the templates queries and template mutations.
icon: BookTemplate
order: 8
---

A template is a workspace whose structure is meant to be reused. Saving a workspace as a template lets you stamp out new workspaces — lists, custom fields, automations, forms, and more — without rebuilding them by hand. Templates are `Project` objects in the API with `isTemplate: true`, so every operation on this page maps to the same project queries and mutations you already use, plus a few template-specific ones.

Blue has two kinds of templates:

- **Organization templates** — saved by your organization, visible only to its members. `isOfficialTemplate: false`.
- **Official templates** — curated by Blue and visible to everyone. `isOfficialTemplate: true`. Only Blue employees can create these.

## Request

List your organization's templates with the `templates` query. It returns a paginated `TemplatePagination` — an `items` array of `Project` objects plus a `pageInfo` block — and takes a `TemplateFilterInput`. With no filter it returns your organization's templates (`isOfficialTemplate` defaults to `false`, scoped to the company in your `X-Bloo-Company-ID` header).

```graphql
query ListTemplates {
  templates(filter: { isOfficialTemplate: false, category: MARKETING }, skip: 0, take: 20) {
    items {
      id
      name
      description
      category
      isOfficialTemplate
      icon
      color
      image {
        thumbnail
        small
      }
    }
    pageInfo {
      hasNextPage
      totalItems
    }
  }
}
```

To browse Blue's official templates instead, pass `isOfficialTemplate: true`. Company and project arguments accept either an ID or a slug.

## Parameters

### `templates` arguments

| Parameter | Type                  | Required | Description                                        |
| --------- | --------------------- | -------- | -------------------------------------------------- |
| `filter`  | `TemplateFilterInput` | No       | Narrows the result set. See the field table below. |
| `skip`    | `Int`                 | No       | Number of templates to skip. Default `0`.          |
| `take`    | `Int`                 | No       | Page size. Default `20`.                           |

### TemplateFilterInput

| Field                | Type              | Required | Description                                                                                                                                             |
| -------------------- | ----------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `companyId`          | `String`          | No       | Organization ID or slug to read templates from. Defaults to the company in the `X-Bloo-Company-ID` header. Ignored when `isOfficialTemplate` is `true`. |
| `templateId`         | `String`          | No       | Used only by the `template` query (see [Get a single template](#get-a-single-template)). Ignored by `templates`.                                        |
| `isOfficialTemplate` | `Boolean`         | No       | `true` returns Blue's official templates; `false` (the default) returns your organization's templates.                                                  |
| `category`           | `ProjectCategory` | No       | Filter by workspace category. See [Template categories](#template-categories).                                                                          |

## Response

```json
{
  "data": {
    "templates": {
      "items": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "name": "Marketing Campaign",
          "description": "Plan and track a multi-channel campaign.",
          "category": "MARKETING",
          "isOfficialTemplate": false,
          "icon": "rocket",
          "color": "#10B981",
          "image": {
            "thumbnail": "https://blue.app/files/clm4n8qwx-thumbnail.png",
            "small": "https://blue.app/files/clm4n8qwx-small.png"
          }
        }
      ],
      "pageInfo": {
        "hasNextPage": false,
        "totalItems": 1
      }
    }
  }
}
```

### Returns

`templates` returns a `TemplatePagination`.

| Field      | Type          | Description                                                                                            |
| ---------- | ------------- | ------------------------------------------------------------------------------------------------------ |
| `items`    | `[Project!]!` | The templates on this page, each a `Project` with `isTemplate: true`.                                  |
| `pageInfo` | `PageInfo!`   | Pagination metadata (`totalItems`, `totalPages`, `page`, `perPage`, `hasNextPage`, `hasPreviousPage`). |

## Listing templates with projectList

`templates` is the focused way to read templates, but you can also use the general [`projectList`](/api/workspaces/list-workspaces) query with `isTemplate: true` on its `ProjectListFilter`. This is handy when you want templates and active workspaces from the same query surface, or need `projectList`-only filters such as `folderId`.

```graphql
query ListTemplateProjects {
  projectList(
    filter: { companyIds: ["company_123"], isTemplate: true }
    sort: [updatedAt_DESC]
    take: 20
    skip: 0
  ) {
    items {
      id
      slug
      name
      description
      category
      isTemplate
      isOfficialTemplate
      color
      icon
      createdAt
      updatedAt
    }
    pageInfo {
      hasNextPage
      totalItems
    }
    totalCount
  }
}
```

Note that `projectList`'s filter requires `companyIds` (a list) and does not split organization vs. official templates the way `templates` does.

## Get a single template

Fetch one template by ID or slug with the `template` query. It takes a `TemplateFilterInput`, but only `templateId` is meaningful here — it is **required**, and omitting it returns a `TEMPLATE_NOT_FOUND` error. For an organization template, the caller must belong to the owning company; pass `companyId` to read a template from a specific organization, otherwise it defaults to your header company.

```graphql
query GetTemplate {
  template(filter: { templateId: "project_123" }) {
    id
    name
    description
    category
    isOfficialTemplate
  }
}
```

To inspect a template's lists and records, fetch them with the [`todoLists`](/api/workspaces/lists) query against the returned project ID — `Project` does not expose its lists or records inline.

## Create a workspace from a template

Pass `templateId` to the [`createProject`](/api/workspaces/create-workspace) mutation to stamp out a new workspace from a template. The template can be one of your organization's templates or an official template; you supply the new workspace's `name` and `companyId`, and may override `description`, `color`, `icon`, and `folderId`.

```graphql
mutation CreateFromTemplate {
  createProject(
    input: {
      templateId: "project_123"
      name: "Q1 Marketing Campaign"
      companyId: "company_123"
      description: "Marketing initiatives for Q1"
      color: "#10B981"
    }
  ) {
    id
    name
    slug
  }
}
```

<Callout variant="warning" title="Creating from a template returns null">

Unlike a blank `createProject` (which returns the new `Project` immediately), creating **from a template** copies content in a background job and the mutation returns `null`. The new workspace is built asynchronously. Poll the [`copyProjectStatus`](#track-copy-progress) query to follow progress, or subscribe to [project events](/api/realtime/project-subscriptions) to be notified when it lands.

</Callout>

### What gets copied

The background copy includes the template's full structure and content:

- Lists and records, with their positions and due dates
- Custom fields and tags
- Automations
- Forms
- Documents and wiki pages
- Comments and discussions
- Status updates
- Cover and display configuration
- Project user-role definitions

What is **not** carried over:

- User assignments (the creating user is added as `OWNER`; other members are not assigned)
- Activity history
- Completed state — records start un-done in the new workspace

## Track copy progress

After creating a workspace from a template, poll `copyProjectStatus` to follow the background copy. It takes no arguments and returns the status of your most recent copy job (or `null` if none is in flight).

```graphql
query CopyStatus {
  copyProjectStatus {
    newProjectName
    isTemplate
    isActive
    queuePosition
    totalQueues
  }
}
```

### CopyProjectStatus

| Field            | Type      | Description                                                      |
| ---------------- | --------- | ---------------------------------------------------------------- |
| `newProjectName` | `String`  | Name of the workspace being created.                             |
| `isTemplate`     | `Boolean` | `true` when the source is a template.                            |
| `isActive`       | `Boolean` | `true` once the copy job is actively running (vs. still queued). |
| `queuePosition`  | `Int`     | This job's position in the copy queue.                           |
| `totalQueues`    | `Int`     | Total jobs currently queued.                                     |
| `oldProject`     | `Project` | The source workspace (the template).                             |
| `newCompany`     | `Company` | The destination organization.                                    |

## Convert a workspace to a template

Save an existing workspace as a template with `convertProjectToTemplate`. It sets `isTemplate: true` and returns the updated `Project`. Converting also **unarchives** the workspace if it was archived and removes it from any folder.

```graphql
mutation ConvertToTemplate {
  convertProjectToTemplate(input: { projectId: "project_123", isOfficialTemplate: false }) {
    id
    name
    isTemplate
    isOfficialTemplate
  }
}
```

### ConvertProjectToTemplateInput

| Field                | Type       | Required | Description                                                                                                                                                 |
| -------------------- | ---------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `projectId`          | `String!`  | Yes      | The workspace ID or slug to convert.                                                                                                                        |
| `isOfficialTemplate` | `Boolean!` | Yes      | Mark as an official Blue template. Only Blue employees can set this `true`; non-employees get a `FORBIDDEN` error. Pass `false` for organization templates. |

## Remove template status

Convert a template back into a regular workspace with `removeProjectFromTemplates`. It sets `isTemplate: false` (and `isOfficialTemplate: false`) and returns the updated `Project`.

```graphql
mutation RemoveTemplateStatus {
  removeProjectFromTemplates(input: { projectId: "project_123" }) {
    id
    name
    isTemplate
  }
}
```

### RemoveProjectFromTemplatesInput

| Field       | Type      | Required | Description                                    |
| ----------- | --------- | -------- | ---------------------------------------------- |
| `projectId` | `String!` | Yes      | The template's workspace ID or slug to revert. |

## Template categories

`category` is a `ProjectCategory` enum, shared with `createProject` and used to group templates in the UI.

| Value              | Description                      |
| ------------------ | -------------------------------- |
| `CRM`              | Customer relationship management |
| `CROSS_FUNCTIONAL` | Cross-functional team workspaces |
| `CUSTOMER_SUCCESS` | Customer success initiatives     |
| `DESIGN`           | Design and creative work         |
| `ENGINEERING`      | Engineering and development      |
| `GENERAL`          | General workspaces (default)     |
| `HR`               | Human Resources                  |
| `IT`               | Information Technology           |
| `MARKETING`        | Marketing campaigns              |
| `OPERATIONS`       | Operations and logistics         |
| `PRODUCT`          | Product management               |
| `SALES`            | Sales and business development   |

## Errors

| Code                   | When                                                                                                                                                              |
| ---------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `TEMPLATE_NOT_FOUND`   | The `template` query is called without `templateId`, or the template does not exist / is not accessible to your organization.                                     |
| `PROJECT_NOT_FOUND`    | `convertProjectToTemplate` / `removeProjectFromTemplates` references a workspace that does not exist, or `createProject` references a template you cannot access. |
| `CREATE_PROJECT_LIMIT` | A workspace with more than 250,000 records is used as a template source — too large to copy.                                                                      |
| `FORBIDDEN`            | A non-Blue-employee passes `isOfficialTemplate: true`, or the caller lacks the required workspace role.                                                           |

```json
{
  "errors": [
    {
      "message": "Template was not found.",
      "extensions": { "code": "TEMPLATE_NOT_FOUND" }
    }
  ]
}
```

## Permissions

| Action                                                   | Required role                     |
| -------------------------------------------------------- | --------------------------------- |
| Convert a workspace to a template                        | Workspace `OWNER` or `ADMIN`      |
| Create an official template (`isOfficialTemplate: true`) | Blue employee                     |
| Remove template status                                   | Workspace `OWNER` or `ADMIN`      |
| Use an organization template                             | Member of the owning organization |
| Use an official template                                 | Any Blue user                     |

## Related

- [Create a workspace](/api/workspaces/create-workspace) — `createProject`, including the `templateId` path
- [Copy a workspace](/api/workspaces/copy-workspace) — `copyProject` for one-off duplication
- [List workspaces](/api/workspaces/list-workspaces) — `projectList` and its filters
- [Lists](/api/workspaces/lists) — read a template's lists and records
- [Project & workspace subscriptions](/api/realtime/project-subscriptions) — `subscribeToProject` for live template events
