---
title: Delete Automation
description: Permanently delete an automation and clean up its schedules, due-date jobs, and caches.
icon: Trash2
order: 4
---

Use the `deleteAutomation` mutation to permanently remove an automation from a workspace. Deleting an automation also tears down everything attached to it: an active schedule is unscheduled, due-date monitoring jobs are removed, the conditional-evaluation cache is invalidated, and a deletion event is published to subscribers. This action cannot be undone.

Automations belong to a workspace (a `Workspace` in the API) and require the `X-Bloo-Project-ID` header. Only owners and admins of an active workspace can delete them.

## Request

```graphql
mutation DeleteAutomation {
  deleteAutomation(id: "automation_123")
}
```

## Parameters

| Parameter | Type      | Required | Description                         |
| --------- | --------- | -------- | ----------------------------------- |
| `id`      | `String!` | Yes      | The ID of the automation to delete. |

## Response

`deleteAutomation` returns the scalar `Boolean!` — `true` when the automation is deleted. It has no sub-fields, so don't add a selection set.

```json
{
  "data": {
    "deleteAutomation": true
  }
}
```

### Returns

| Field  | Type       | Description                                                          |
| ------ | ---------- | -------------------------------------------------------------------- |
| (root) | `Boolean!` | `true` when the automation and its associated resources are deleted. |

## Full example

Delete several automations in one request with GraphQL aliases. Each alias resolves independently and returns its own boolean.

```graphql
mutation DeleteAutomations {
  removeWorkflow: deleteAutomation(id: "automation_123")
  removeNotifier: deleteAutomation(id: "automation_456")
}
```

```json
{
  "data": {
    "removeWorkflow": true,
    "removeNotifier": true
  }
}
```

## Errors

| Code                   | When                                                                                                                                                                                             |
| ---------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ |
| `UNAUTHENTICATED`      | No valid credentials were provided. Message: `You are not authenticated.`                                                                                                                        |
| `FORBIDDEN`            | The automation exists, but the caller isn't an `OWNER` or `ADMIN` of its workspace, the caller has no access to that workspace, or the workspace is archived. Message: `You are not authorized.` |
| `AUTOMATION_NOT_FOUND` | No automation exists with the given `id`. Message: `Automation was not found.`                                                                                                                   |

<Callout variant="info" title="Not found vs. forbidden">

A non-existent `id` returns `AUTOMATION_NOT_FOUND`. An `id` that exists but lives in a workspace you can't access returns `FORBIDDEN` — the automation is looked up globally, then checked against your workspace membership.

</Callout>

## Permissions

| Access level   | Can delete automations |
| -------------- | ---------------------- |
| `OWNER`        | Yes                    |
| `ADMIN`        | Yes                    |
| `MEMBER`       | No                     |
| `CLIENT`       | No                     |
| `COMMENT_ONLY` | No                     |
| `VIEW_ONLY`    | No                     |

The workspace must be active — automations in an archived workspace can't be deleted.

## Cleanup behavior

Deletion is irreversible and cascades automatically:

- **Schedules** — an active schedule is unscheduled before the automation is removed.
- **Due-date jobs** — for a `DUE_DATE_EXPIRED` trigger, the monitoring jobs are removed (record due dates, or custom-field due dates when the trigger watches a custom field).
- **Conditional cache** — for a `CONDITIONAL` trigger, the evaluation cache and the workspace's active-automation flag are invalidated.
- **Subscriptions** — a deletion event is published so connected clients can update in real time.

## Related

- [Create an automation](/api/automations/create-automation)
- [Edit an automation](/api/automations/edit-automation)
- [Copy an automation](/api/automations/copy-automation)
- [Automations overview](/api/automations)
