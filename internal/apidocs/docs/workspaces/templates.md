---
title: Templates
description: Reuse a workspace's structure by saving it as a template and creating new workspaces from it, with the templates queries and template mutations.
icon: BookTemplate
order: 8
---

A template is a workspace whose structure is meant to be reused. Saving a workspace as a template lets you stamp out new workspaces — lists, custom fields, automations, forms, and more — without rebuilding them by hand. Templates are `Workspace` objects in the API with `isTemplate: true`, so every operation on this page maps to the same workspace queries and mutations you already use, plus a few template-specific ones.

Blue has two kinds of templates:

- **Organization templates** — saved by your organization, visible only to its members. `isOfficialTemplate: false`.
- **Official templates** — curated by Blue and visible to everyone. `isOfficialTemplate: true`. Only Blue employees can create these.

A template can also be a **bundle** — a folder of several interconnected workspaces saved and instantiated together, with their cross-workspace links preserved. See [Template bundles](#template-bundles).

## Request

List your organization's templates with the `templates` query. It returns a paginated `TemplatePagination` — an `items` array of `Workspace` objects plus a `pageInfo` block — and takes a `TemplateFilterInput`. With no filter it returns your organization's templates (`isOfficialTemplate` defaults to `false`, scoped to the company in your `X-Bloo-Company-ID` header).

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

| Field                | Type              | Required | Description                                                                                                                                       |
| -------------------- | ----------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------- |
| `companyId`          | `String`          | No       | Organization ID or slug to read templates from. Defaults to the company in the `blue-org-id` header. Ignored when `isOfficialTemplate` is `true`. |
| `templateId`         | `String`          | No       | Used only by the `template` query (see [Get a single template](#get-a-single-template)). Ignored by `templates`.                                  |
| `isOfficialTemplate` | `Boolean`         | No       | `true` returns Blue's official templates; `false` (the default) returns your organization's templates.                                            |
| `category`           | `ProjectCategory` | No       | Filter by workspace category. See [Template categories](#template-categories).                                                                    |

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

| Field      | Type            | Description                                                                                            |
| ---------- | --------------- | ------------------------------------------------------------------------------------------------------ |
| `items`    | `[Workspace!]!` | The templates on this page, each a `Workspace` with `isTemplate: true`.                                |
| `pageInfo` | `PageInfo!`     | Pagination metadata (`totalItems`, `totalPages`, `page`, `perPage`, `hasNextPage`, `hasPreviousPage`). |

## Listing templates with workspaceList

`templates` is the focused way to read templates, but you can also use the general [`workspaceList`](/api/workspaces/list-workspaces) query with `isTemplate: true` on its `ProjectListFilter`. This is handy when you want templates and active workspaces from the same query surface, or need `workspaceList`-only filters such as `folderId`.

```graphql
query ListTemplateProjects {
  workspaceList(
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

Note that `workspaceList`'s filter requires `companyIds` (a list) and does not split organization vs. official templates the way `templates` does.

## Get a single template

Fetch one template by ID or slug with the `template` query. It takes a `TemplateFilterInput`, but only `templateId` is meaningful here — it is **required**, and omitting it returns a `TEMPLATE_NOT_FOUND` error. For an organization template, the caller must belong to the owning company; pass `companyId` to read a template from a specific organization, otherwise it defaults to your header company.

```graphql
query GetTemplate {
  template(filter: { templateId: "workspace_123" }) {
    id
    name
    description
    category
    isOfficialTemplate
  }
}
```

To inspect a template's lists and records, fetch them with the [`recordLists`](/api/workspaces/lists) query against the returned workspace ID — `Workspace` does not expose its lists or records inline.

## Create a workspace from a template

Pass `templateId` to the [`createWorkspace`](/api/workspaces/create-workspace) mutation to stamp out a new workspace from a template. The template can be one of your organization's templates or an official template; you supply the new workspace's `name` and `companyId`, and may override `description`, `color`, `icon`, and `folderId`.

```graphql
mutation CreateFromTemplate {
  createWorkspace(
    input: {
      templateId: "workspace_123"
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

Unlike a blank `createWorkspace` (which returns the new `Workspace` immediately), creating **from a template** copies content in a background job and the mutation returns `null`. The new workspace is built asynchronously. Poll the [`copyProjectStatus`](#track-copy-progress) query to follow progress, or subscribe to [project events](/api/realtime/project-subscriptions) to be notified when it lands.

</Callout>

### What gets copied

The background copy includes the template's full structure and content:

- Lists and records, with their positions and due dates
- Custom fields and tags
- Automations
- Forms
- Documents and wiki pages
- Comments and chats
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

| Field            | Type           | Description                                                      |
| ---------------- | -------------- | ---------------------------------------------------------------- |
| `newProjectName` | `String`       | Name of the workspace being created.                             |
| `isTemplate`     | `Boolean`      | `true` when the source is a template.                            |
| `isActive`       | `Boolean`      | `true` once the copy job is actively running (vs. still queued). |
| `queuePosition`  | `Int`          | This job's position in the copy queue.                           |
| `totalQueues`    | `Int`          | Total jobs currently queued.                                     |
| `oldProject`     | `Workspace`    | The source workspace (the template).                             |
| `newCompany`     | `Organization` | The destination organization.                                    |

## Convert a workspace to a template

Save an existing workspace as a template with `convertProjectToTemplate`. It sets `isTemplate: true` and returns the updated `Workspace`. Converting also **unarchives** the workspace if it was archived and removes it from any folder.

```graphql
mutation ConvertToTemplate {
  convertProjectToTemplate(input: { projectId: "workspace_123", isOfficialTemplate: false }) {
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

Convert a template back into a regular workspace with `removeProjectFromTemplates`. It sets `isTemplate: false` (and `isOfficialTemplate: false`) and returns the updated `Workspace`.

```graphql
mutation RemoveTemplateStatus {
  removeProjectFromTemplates(input: { projectId: "workspace_123" }) {
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

## Template bundles

A **template bundle** is a folder of two or more interconnected workspaces saved together as a single template. Instantiating one creates a folder containing a copy of every member workspace, with their **cross-workspace links preserved and re-pointed at the copies** — reference, lookup, and rollup fields, record dependencies, and cross-workspace automation targets all resolve against the new workspaces rather than the originals. This lets a template demonstrate a connected set of processes, not just one board.

A bundle is a `Folder` with `isTemplate: true`. Like single-workspace templates, bundles are either **organization** bundles (`isOfficialTemplate: false`, visible only to the owning organization) or **official** bundles (`isOfficialTemplate: true`, curated by Blue, visible to everyone — only Blue employees can create them).

### List template bundles

Use the `templateFolders` query to list bundles. It returns a `[TemplateFolder!]!` — a lightweight summary of each bundle plus its members, suitable for a gallery — and takes a `TemplateFolderFilterInput`. With no filter it returns your organization's bundles (scoped to the company in your `blue-org-id` header); pass `isOfficialTemplate: true` for Blue's official bundles.

```graphql
query ListTemplateBundles {
  templateFolders(filter: { isOfficialTemplate: false }) {
    id
    title
    color
    isOfficialTemplate
    memberCount
    members {
      id
      slug
      name
      icon
      color
    }
  }
}
```

#### TemplateFolderFilterInput

| Field                | Type      | Required | Description                                                                                                                                                                        |
| -------------------- | --------- | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `companyId`          | `String`  | No       | Organization ID or slug to read bundles from. Defaults to your `blue-org-id` header company. Only a company you belong to is honored. Ignored when `isOfficialTemplate` is `true`. |
| `isOfficialTemplate` | `Boolean` | No       | `true` returns Blue's official bundles; `false` (the default) returns your organization's bundles.                                                                                 |

#### TemplateFolder

| Field                | Type                       | Description                                                                 |
| -------------------- | -------------------------- | --------------------------------------------------------------------------- |
| `id`                 | `ID!`                      | The bundle folder's ID.                                                     |
| `title`              | `String!`                  | The bundle's display name.                                                  |
| `color`              | `String`                   | Optional folder color.                                                      |
| `isOfficialTemplate` | `Boolean`                  | `true` for Blue's official bundles.                                         |
| `memberCount`        | `Int!`                     | Number of member workspaces in the bundle.                                  |
| `members`            | `[TemplateFolderMember!]!` | Member summaries (`id`, `slug`, `name`, `icon`, `color`) for gallery cards. |

### Convert a folder to a template bundle

`convertFolderToTemplate` flags an existing folder — and every workspace inside it — as a template bundle. It returns the updated `Folder`. The folder must be **flat**: a bundle cannot contain sub-folders.

```graphql
mutation ConvertFolderToTemplate {
  convertFolderToTemplate(input: { folderId: "folder_123", isOfficialTemplate: false }) {
    id
    title
    isTemplate
    isOfficialTemplate
  }
}
```

#### ConvertFolderToTemplateInput

| Field                | Type       | Required | Description                                                                                                                                             |
| -------------------- | ---------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `folderId`           | `String!`  | Yes      | The folder ID to convert to a bundle.                                                                                                                   |
| `isOfficialTemplate` | `Boolean!` | Yes      | Mark as an official Blue bundle. Only Blue employees can set this `true`; non-employees get a `FORBIDDEN` error. Pass `false` for organization bundles. |

### Instantiate a template bundle

`instantiateTemplateFolder` spins up a new folder of interconnected workspaces from a bundle. Unlike creating from a single template (which returns `null`), this mutation returns the destination `Folder` **immediately** — the member workspaces are copied in a background job and stream in as each completes. Follow progress with the [`copyProjectStatus`](#track-copy-progress) query or [project subscriptions](/api/realtime/project-subscriptions).

```graphql
mutation InstantiateBundle {
  instantiateTemplateFolder(
    input: { companyId: "company_123", templateFolderId: "folder_123", name: "Q1 Launch" }
  ) {
    id
    title
  }
}
```

#### InstantiateTemplateFolderInput

| Field              | Type      | Required | Description                                                                                                  |
| ------------------ | --------- | -------- | ------------------------------------------------------------------------------------------------------------ |
| `companyId`        | `String!` | Yes      | Destination organization ID or slug.                                                                         |
| `templateFolderId` | `String!` | Yes      | The bundle folder to instantiate (ID or slug). May be one of your organization's bundles or an official one. |
| `name`             | `String`  | No       | Name for the destination folder. Defaults to the bundle's title.                                             |

Each member is copied with the same content as a single-workspace template (see [What gets copied](#what-gets-copied)). In addition, links **between members** are re-pointed at the copies, so the new workspaces reference each other rather than the originals. Links to a workspace **outside** the bundle are dropped — never left pointing at the source.

### Remove bundle status

`removeFolderFromTemplates` reverts a bundle folder back into a regular folder (`isTemplate: false`, `isOfficialTemplate: false`). It returns the updated `Folder`.

```graphql
mutation RemoveFolderFromTemplates {
  removeFolderFromTemplates(input: { folderId: "folder_123" }) {
    id
    title
    isTemplate
  }
}
```

#### RemoveFolderFromTemplatesInput

| Field      | Type      | Required | Description                     |
| ---------- | --------- | -------- | ------------------------------- |
| `folderId` | `String!` | Yes      | The bundle folder ID to revert. |

## Template categories

`category` is a `ProjectCategory` enum, shared with `createWorkspace` and used to group templates in the UI.

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

| Code                   | When                                                                                                                                                                 |
| ---------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `TEMPLATE_NOT_FOUND`   | The `template` query is called without `templateId`, or the template does not exist / is not accessible to your organization.                                        |
| `PROJECT_NOT_FOUND`    | `convertProjectToTemplate` / `removeProjectFromTemplates` references a workspace that does not exist, or `createWorkspace` references a template you cannot access.  |
| `CREATE_PROJECT_LIMIT` | A workspace with more than 250,000 records is used as a template source — too large to copy. For a bundle, the combined record count of all members exceeds 250,000. |
| `FORBIDDEN`            | A non-Blue-employee passes `isOfficialTemplate: true`, or the caller lacks the required workspace or folder role.                                                    |
| `BAD_USER_INPUT`       | `instantiateTemplateFolder` references a bundle that does not exist or is not accessible, a folder with sub-folders, or one with no workspaces.                      |
| `COMPANY_NOT_FOUND`    | A folder or bundle mutation references an organization that does not exist.                                                                                          |

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

| Action                                                   | Required role                                               |
| -------------------------------------------------------- | ----------------------------------------------------------- |
| Convert a workspace to a template                        | Workspace `OWNER` or `ADMIN`                                |
| Create an official template (`isOfficialTemplate: true`) | Blue employee                                               |
| Remove template status                                   | Workspace `OWNER` or `ADMIN`                                |
| Use an organization template                             | Member of the owning organization                           |
| Use an official template                                 | Any Blue user                                               |
| Convert a folder to / remove a bundle                    | Organization `OWNER`, or folder management permission       |
| Instantiate a bundle                                     | Same as creating a workspace (`OWNER` / `ADMIN` / `MEMBER`) |

## Related

- [Create a workspace](/api/workspaces/create-workspace) — `createWorkspace`, including the `templateId` path
- [Copy a workspace](/api/workspaces/copy-workspace) — `copyWorkspace` for one-off duplication
- [List workspaces](/api/workspaces/list-workspaces) — `workspaceList` and its filters
- [Lists](/api/workspaces/lists) — read a template's lists and records
- [Project & workspace subscriptions](/api/realtime/project-subscriptions) — `subscribeToProject` and the bundle template events (`onConvertFolderToTemplate` / `onRemoveFolderFromTemplates`)
