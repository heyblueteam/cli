---
title: Create a Workspace
description: Create a new workspace in an organization, optionally seeded from a template, with the createProject mutation.
icon: Plus
order: 1
---

Use the `createProject` mutation to create a new workspace. Workspaces are `Project` objects in the API, and live inside an organization (`Company`). The user whose token makes the request is automatically added to the new workspace as its `OWNER`.

Creating a workspace from scratch returns the new `Project` synchronously. Creating one from a template is queued and runs in the background — the mutation returns `null`, and you poll `copyProjectStatus` to track progress.

## Request

The smallest call provides an organization and a name.

```graphql
mutation CreateWorkspace {
  createProject(input: { companyId: "company_123", name: "Q1 Marketing Campaign" }) {
    id
    slug
    name
  }
}
```

Include your authentication headers (header names are case-insensitive):

- `X-Bloo-Token-ID` — your API token ID
- `X-Bloo-Token-Secret` — your API token secret
- `X-Bloo-Company-ID` — your organization ID or slug

`companyId` accepts either the organization's ID or its slug.

## Parameters

### CreateProjectInput

| Parameter     | Type                   | Required | Description                                                                                                                                 |
| ------------- | ---------------------- | -------- | ------------------------------------------------------------------------------------------------------------------------------------------- |
| `companyId`   | `String!`              | Yes      | ID or slug of the organization where the workspace is created.                                                                              |
| `name`        | `String!`              | Yes      | Workspace name. Trimmed of surrounding whitespace, limited to 50 characters, and any URLs are stripped out.                                 |
| `description` | `String`               | No       | A description of the workspace.                                                                                                             |
| `color`       | `String`               | No       | Workspace color as a hex string (for example `#3B82F6`).                                                                                    |
| `icon`        | `String`               | No       | Icon identifier for the workspace (for example `briefcase`).                                                                                |
| `category`    | `ProjectCategory`      | No       | Workspace category. Defaults to `GENERAL`.                                                                                                  |
| `templateId`  | `String`               | No       | ID or slug of a template workspace to copy. Creation becomes asynchronous when set — see [Create from a template](#create-from-a-template). |
| `folderId`    | `String`               | No       | ID of a folder to place the new workspace inside. The folder must exist in the same organization.                                           |
| `coverConfig` | `TodoCoverConfigInput` | No       | Record cover-image configuration. Only applied when creating from a template — see [Cover image configuration](#cover-image-configuration). |

### ProjectCategory

| Value              | Description                         |
| ------------------ | ----------------------------------- |
| `CRM`              | Customer relationship management    |
| `CROSS_FUNCTIONAL` | Cross-functional teams              |
| `CUSTOMER_SUCCESS` | Customer success initiatives        |
| `DESIGN`           | Design and creative work            |
| `ENGINEERING`      | Engineering and development         |
| `GENERAL`          | General workspaces (default)        |
| `HR`               | Human resources                     |
| `IT`               | Information technology              |
| `MARKETING`        | Marketing campaigns and initiatives |
| `OPERATIONS`       | Operations and logistics            |
| `PRODUCT`          | Product management                  |
| `SALES`            | Sales and business development      |

## Response

The mutation returns the newly created `Project`. (The return type is nullable — see [Create from a template](#create-from-a-template) for the case where it returns `null`.)

```json
{
  "data": {
    "createProject": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "slug": "q1-marketing-campaign",
      "name": "Q1 Marketing Campaign"
    }
  }
}
```

### Returns

`createProject` returns a `Project`. Common fields:

| Field         | Type               | Description                                                                            |
| ------------- | ------------------ | -------------------------------------------------------------------------------------- |
| `id`          | `ID!`              | Unique identifier for the workspace.                                                   |
| `slug`        | `String!`          | URL-friendly workspace identifier, derived from the name.                              |
| `name`        | `String!`          | Workspace name.                                                                        |
| `description` | `String`           | Workspace description.                                                                 |
| `color`       | `String`           | Workspace color in hex format.                                                         |
| `icon`        | `String`           | Icon identifier.                                                                       |
| `category`    | `ProjectCategory!` | Workspace category.                                                                    |
| `company`     | `Company!`         | The organization that owns the workspace. Select subfields such as `{ id slug name }`. |
| `folder`      | `Folder`           | The folder the workspace belongs to, if any.                                           |
| `archived`    | `Boolean`          | Whether the workspace is archived.                                                     |
| `isTemplate`  | `Boolean!`         | Whether this workspace is a template.                                                  |
| `createdAt`   | `DateTime!`        | Creation timestamp.                                                                    |
| `updatedAt`   | `DateTime!`        | Last-update timestamp.                                                                 |

There is no `companyId` scalar on `Project` — the organization is exposed as the `company` relation. Select `company { id slug name }` when you need its identifiers.

## Create from a template

Pass a `templateId` to seed the new workspace from an existing template. This copies the template's lists, records, custom fields, automations, tags, files, and other content.

```graphql
mutation CreateFromTemplate {
  createProject(
    input: { companyId: "company_123", name: "Q1 Marketing Campaign", templateId: "project_123" }
  ) {
    id
  }
}
```

<Callout variant="warning" title="Template creation is asynchronous">

When `templateId` is set, the mutation enqueues a background copy job and returns `null` — the workspace does not exist yet when the response comes back. Poll <code>copyProjectStatus</code> (below) to learn when it is ready, or subscribe to the copy-progress events in <a href="/api/realtime">Realtime updates</a>.

</Callout>

### Track creation status

Use the `copyProjectStatus` query to check where your workspace sits in the copy queue.

```graphql
query WorkspaceCopyStatus {
  copyProjectStatus {
    newProjectName
    isTemplate
    isActive
    queuePosition
    totalQueues
  }
}
```

```json
{
  "data": {
    "copyProjectStatus": {
      "newProjectName": "Q1 Marketing Campaign",
      "isTemplate": true,
      "isActive": false,
      "queuePosition": 2,
      "totalQueues": 5
    }
  }
}
```

`copyProjectStatus` returns a `CopyProjectStatus`:

| Field            | Type      | Description                                |
| ---------------- | --------- | ------------------------------------------ |
| `newProjectName` | `String`  | Name of the workspace being created.       |
| `isTemplate`     | `Boolean` | Whether the source is a template.          |
| `isActive`       | `Boolean` | Whether the copy job is currently running. |
| `queuePosition`  | `Int`     | Position of this job in the copy queue.    |
| `totalQueues`    | `Int`     | Total number of copy jobs queued.          |

## Cover image configuration

The `coverConfig` input controls how record cover images are chosen in the new workspace. It is only applied when creating from a template; for workspaces created from scratch, configure cover images afterward with [Rename / edit a workspace](/api/workspaces/rename-workspace).

```graphql
mutation CreateWithCoverConfig {
  createProject(
    input: {
      companyId: "company_123"
      name: "Q1 Marketing Campaign"
      templateId: "project_123"
      description: "Marketing initiatives for Q1"
      color: "#10B981"
      icon: "megaphone"
      category: MARKETING
      coverConfig: { enabled: true, fit: COVER, imageSelectionType: FIRST, source: DESCRIPTION }
    }
  ) {
    id
  }
}
```

### TodoCoverConfigInput

| Parameter            | Type                  | Required | Description                                                                        |
| -------------------- | --------------------- | -------- | ---------------------------------------------------------------------------------- |
| `enabled`            | `Boolean!`            | Yes      | Whether record cover images are enabled.                                           |
| `fit`                | `ImageFit!`           | Yes      | How the image fits the cover area.                                                 |
| `imageSelectionType` | `ImageSelectionType!` | Yes      | Which image to use when several are available.                                     |
| `source`             | `ImageSource!`        | Yes      | Where to pull the cover image from.                                                |
| `sourceId`           | `String`              | No       | Source identifier (for example a custom field ID) when `source` is `CUSTOM_FIELD`. |

- **`ImageFit`** — `COVER`, `CONTAIN`, `FILL`, `SCALE_DOWN`
- **`ImageSelectionType`** — `FIRST`, `LAST`
- **`ImageSource`** — `DESCRIPTION`, `COMMENTS`, `CUSTOM_FIELD`

## Errors

| Code                   | When                                                                                                       |
| ---------------------- | ---------------------------------------------------------------------------------------------------------- |
| `FORBIDDEN`            | The caller is not `OWNER`, `ADMIN`, or `MEMBER` of the organization.                                       |
| `COMPANY_NOT_FOUND`    | No organization matches the supplied `companyId`.                                                          |
| `PROJECT_NOT_FOUND`    | The `templateId` does not resolve to a template you can access.                                            |
| `CREATE_PROJECT_LIMIT` | The template contains more than 250,000 records and cannot be copied.                                      |
| `BAD_USER_INPUT`       | `name` is missing or longer than 50 characters, or `folderId` does not match a folder in the organization. |

```json
{
  "errors": [
    {
      "message": "Company was not found.",
      "extensions": { "code": "COMPANY_NOT_FOUND" }
    }
  ]
}
```

A plan limit on the number of workspaces (or, for template copies, the number of records) is reported separately with an upgrade-oriented payload rather than one of the codes above.

## Permissions

You must be an `OWNER`, `ADMIN`, or `MEMBER` of the organization to create a workspace. The creating user is added to the new workspace as its `OWNER`.

## Related

- [List workspaces](/api/workspaces/list-workspaces)
- [Rename / edit a workspace](/api/workspaces/rename-workspace)
- [Copy a workspace](/api/workspaces/copy-workspace)
- [Templates](/api/workspaces/templates)
- [Error codes](/api/start-guide/error-codes)
