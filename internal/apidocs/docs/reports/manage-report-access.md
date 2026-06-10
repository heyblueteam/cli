---
title: Manage Report Access
description: Update a report's title, description, config, and data sources, and share it with EDITOR/VIEWER collaborators.
icon: Users
order: 4
---

Use the `updateReport` mutation to change a report's `title`, `description`, or `config`, replace its data sources, and manage who it's shared with. Sharing lives inside the same mutation: the `reportUsers` field on `UpdateReportInput` is the report's collaborator list, where each entry grants a user the `EDITOR` or `VIEWER` role.

Reports are scoped to one organization (`Company` in the API) and start private to their creator. Workspaces are `Project` objects and records are `Todo` objects.

`updateReport` only updates the fields you supply — but the list-valued fields (`dataSources`, `reportUsers`) are **wholesale replacements**, not merges. See [Replacement semantics](#replacement-semantics) before you call it.

## Request

The smallest useful call updates a single scalar. Omitted fields are left untouched.

```graphql
mutation RenameReport {
  updateReport(id: "report_123", input: { title: "Q3 Pipeline (revised)" }) {
    id
    title
    updatedAt
  }
}
```

To share a report, pass `reportUsers`. This call shares the report with two collaborators — one editor, one viewer:

```graphql
mutation ShareReport {
  updateReport(
    id: "report_123"
    input: {
      reportUsers: [{ userId: "user_123", role: EDITOR }, { userId: "user_456", role: VIEWER }]
    }
  ) {
    id
    reportUsers {
      role
      user {
        id
        fullName
        email
      }
    }
  }
}
```

## Parameters

The `id` argument is the report's ID. The caller must be the report's creator or an `EDITOR` — see [Permissions](#permissions).

### UpdateReportInput

| Parameter     | Type                             | Required | Description                                                                                                                                                                         |
| ------------- | -------------------------------- | -------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `title`       | `String`                         | No       | New display name. Only applied if supplied.                                                                                                                                         |
| `description` | `String`                         | No       | New free-text description. Maximum 10,000 characters.                                                                                                                               |
| `config`      | `JSONObject`                     | No       | Application-defined view configuration (opaque key/value blob). Maximum 100&nbsp;KB once JSON-serialized.                                                                           |
| `dataSources` | `[UpdateReportDataSourceInput!]` | No       | If supplied, **replaces every** existing data source (delete-then-recreate). Omit to leave them unchanged. At most 50 entries. See [Replacement semantics](#replacement-semantics). |
| `reportUsers` | `[EditReportUserInput!]`         | No       | If supplied, the full reconciled collaborator list — adds new users, updates roles, and removes anyone not present. Omit to leave sharing unchanged.                                |

### UpdateReportDataSourceInput

| Parameter    | Type               | Required | Description                                                                                                                                                                        |
| ------------ | ------------------ | -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `id`         | `String`           | No       | Accepted for symmetry with the read type but **ignored** — every source is recreated with a new identifier. See the note below.                                                    |
| `name`       | `String`           | No       | Optional label for the data source.                                                                                                                                                |
| `sourceType` | `ReportSourceType` | No       | What the data source aggregates. Defaults to `TODOS` (the only value today).                                                                                                       |
| `projectIds` | `[String!]`        | No       | Workspace IDs (or slugs) to scope this source to. `null` or an empty array means **all** workspaces the creator can access; a non-empty array means **those specific** workspaces. |
| `filters`    | `JSONObject`       | No       | Filter applied to records in this source — an opaque blob mirroring the [Records filter inputs](/api/records/list-records).                                                        |
| `order`      | `Int`              | No       | Sort position of the data source within the report. Defaults to its index in the supplied list.                                                                                    |

### EditReportUserInput

| Parameter | Type          | Required | Description                                                                 |
| --------- | ------------- | -------- | --------------------------------------------------------------------------- |
| `userId`  | `String!`     | Yes      | The user to grant access to. Must be a member of the report's organization. |
| `role`    | `ReportRole!` | Yes      | The access level to grant — `EDITOR` or `VIEWER`.                           |

### ReportRole

| Value    | Grants                                                                                                                                                                                                                                   |
| -------- | ---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `EDITOR` | View the report **and** call `updateReport` (rename, reconfigure, re-share, replace data sources). Cannot delete the report.                                                                                                             |
| `VIEWER` | View the report, read its [data and aggregations](/api/reports/report-data-aggregations), [duplicate](/api/reports/duplicate-and-delete-report) it, and [refresh aggregations](/api/reports/report-data-aggregations). Cannot modify it. |

## Replacement semantics

The two list fields are reconciled wholesale, not patched. Plan your input accordingly.

- **`dataSources` is a full replace.** Supplying it deletes **all** existing data sources and recreates the report's data sources from the array you pass. There is no way to patch a single source through this path: even though `UpdateReportDataSourceInput` carries an `id`, the resolver ignores it and assigns every recreated source a fresh identifier. To "edit one source," resend the complete set with your change applied. Omit `dataSources` entirely to leave the existing sources alone.
- **`reportUsers` is a full reconciled share list.** Supplying it makes the array the complete set of collaborators: users in the array are added or have their role updated, and **any current collaborator not in the array is removed**. Send an empty array `[]` to unshare the report from everyone. Omit `reportUsers` to leave sharing unchanged.

<Callout variant="warning" title="dataSources replaces, never merges">

If you send `dataSources` to change one source, include every source you want to keep. Sending a single-element array drops all the others. The same applies to `reportUsers` — omit the field to preserve the current share list; sending a partial array removes everyone you left out.

</Callout>

The `projectIds` convention is identical to [Create a Report](/api/reports/create-report#the-projectids-convention): `null` or `[]` means all workspaces the **report creator** can access (workspace access is always resolved against the creator, not the editor making the change), and a non-empty array means those specific workspaces. The computed [`Report.projectIds`](/api/reports/query-reports) field returns the union across all sources.

## Response

`updateReport` returns the updated `Report!`. The example below corresponds to the [ShareReport](#request) call above.

```json
{
  "data": {
    "updateReport": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "reportUsers": [
        {
          "role": "EDITOR",
          "user": {
            "id": "clm4n8r1a000108l0a2bcde90",
            "fullName": "Dana Editor",
            "email": "dana@example.com"
          }
        },
        {
          "role": "VIEWER",
          "user": {
            "id": "clm4n8r1a000208l0f5ghij12",
            "fullName": "Vic Viewer",
            "email": "vic@example.com"
          }
        }
      ]
    }
  }
}
```

### Report

`updateReport` returns the full `Report`. The complete `Report` and `ReportDataSource` field reference lives on [Query Reports](/api/reports/query-reports); the fields relevant to this page are below.

| Field         | Type                   | Description                                                     |
| ------------- | ---------------------- | --------------------------------------------------------------- |
| `id`          | `ID!`                  | The report's unique identifier.                                 |
| `title`       | `String!`              | The report's display name.                                      |
| `description` | `String`               | The description, or `null`.                                     |
| `config`      | `JSONObject`           | Application-defined view configuration, or `null`.              |
| `dataSources` | `[ReportDataSource!]!` | The report's data sources after the update, in `order`.         |
| `reportUsers` | `[ReportUser!]!`       | The current collaborator list. See [`ReportUser`](#reportuser). |
| `updatedAt`   | `DateTime!`            | Bumped to now on a successful update.                           |

### ReportUser

Each entry in `Report.reportUsers` is a `ReportUser` — one collaborator the report is shared with.

| Field       | Type          | Description                                          |
| ----------- | ------------- | ---------------------------------------------------- |
| `id`        | `ID!`         | The share record's unique identifier.                |
| `uid`       | `String!`     | Public unique identifier for the share record.       |
| `user`      | `User!`       | The collaborator. Select `id`, `fullName`, `email`.  |
| `role`      | `ReportRole!` | The granted access level — `EDITOR` or `VIEWER`.     |
| `createdAt` | `DateTime!`   | When the report was first shared with this user.     |
| `updatedAt` | `DateTime!`   | When this share record was last changed (e.g. role). |

## Full example

Replace the report's data sources and update its config in one call. This drops any previously configured sources and rebuilds them from the array, while leaving the share list untouched (because `reportUsers` is omitted):

```graphql
mutation ReconfigureReport {
  updateReport(
    id: "report_123"
    input: {
      title: "Open deals — EMEA & US"
      config: { layout: "table", groupBy: "stage" }
      dataSources: [
        {
          name: "EMEA"
          sourceType: TODOS
          projectIds: ["project_123"]
          filters: { done: false }
          order: 0
        }
        {
          name: "US"
          sourceType: TODOS
          projectIds: ["project_456"]
          filters: { done: false }
          order: 1
        }
      ]
    }
  ) {
    id
    title
    config
    dataSources {
      id
      name
      projectIds
      order
    }
  }
}
```

Supplying `dataSources` triggers a transactional backfill of list items for the referenced workspaces, exactly as on create. If the backfill fails for any workspace, the **entire update rolls back** with `Failed to backfill N of M projects. Update aborted. Failed projects: …` — no partial change is applied. Sources scoped to all workspaces (`projectIds: null`) reference no specific IDs and never trigger this.

## Errors

| Code                  | When                                                                                                                                                                                   |
| --------------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `UNAUTHENTICATED`     | No authenticated user on the request.                                                                                                                                                  |
| `FORBIDDEN`           | The request has no active company context.                                                                                                                                             |
| `REPORT_NOT_FOUND`    | No report with that `id` is visible to the caller as creator or `EDITOR` — also returned when the report exists but the caller only has `VIEWER` access (no editable report is found). |
| `USER_NOT_IN_COMPANY` | A `reportUsers` entry references a `userId` that is not a member of the report's organization.                                                                                         |
| `BAD_USER_INPUT`      | Malformed input (e.g. a `JSONObject` field that isn't a valid object).                                                                                                                 |

The data-source count (max 50), `description` length (max 10,000 chars), and `config` size (max 100&nbsp;KB serialized) limits, plus the backfill rollback, throw a plain error with no machine-readable extension code — the message states the cause (e.g. `Config size (… bytes) exceeds maximum of 102400 bytes`).

## Permissions

Report access has four tiers. `updateReport` sits in the **modify** tier:

| Tier       | Who                                                                           | Operations                                                                                                                                                                  |
| ---------- | ----------------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **View**   | Creator **or** any collaborator (`EDITOR` or `VIEWER`)                        | [`report`](/api/reports/query-reports), [`refreshReportAggregations`](/api/reports/report-data-aggregations), [`duplicateReport`](/api/reports/duplicate-and-delete-report) |
| **Modify** | Creator **or** collaborators with role `EDITOR`                               | **`updateReport`** (this page)                                                                                                                                              |
| **Delete** | Creator **only**                                                              | [`deleteReport`](/api/reports/duplicate-and-delete-report)                                                                                                                  |
| **Export** | Creator, **any** collaborator, **or** any member of the report's organization | [`exportReport`](/api/reports/export-report) — intentionally the broadest tier                                                                                              |

A `VIEWER` cannot call `updateReport` — the lookup returns `REPORT_NOT_FOUND` rather than a distinct authorization error, because the resolver finds no report the caller is permitted to modify. To let someone reconfigure or re-share a report, grant them `EDITOR`. Only the creator can delete a report; transfer of ownership is not available through this API.

## Related

- [Reports overview](/api/reports)
- [Create a Report](/api/reports/create-report) — `CreateReportInput`, the `projectIds` convention, and create-time limits.
- [Query Reports](/api/reports/query-reports) — full `Report` and `ReportDataSource` field reference, and the report visibility model.
- [Report Data & Aggregations](/api/reports/report-data-aggregations) — read the records, counts, and aggregations a report produces.
- [Duplicate & Delete a Report](/api/reports/duplicate-and-delete-report) — copy a report (drops the share list) or hard-delete it (creator only).
- [Export a Report](/api/reports/export-report) — async CSV export with broader access.
- [List Records](/api/records/list-records) — the filter structure mirrored by `filters`.
