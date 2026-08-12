---
title: Delete a Custom Field
description: Delete a custom field from a workspace, moving it to Trash where an admin can restore it with its values intact.
icon: Trash2
order: 4
---

Use the `deleteCustomField` mutation to delete a custom field from a workspace. Custom fields are `CustomField` objects in the API; workspaces are `Project` objects. Deletion is reversible: the field moves to Trash, keeping the values it held on every record, and an `OWNER` or `ADMIN` can [restore it](#restoring-a-deleted-field) until the organization's retention window expires. The mutation returns `Boolean!`.

All custom-field requests are sent to `https://api.blue.app/graphql` with your authentication headers. A custom field is scoped to a single workspace, so include the workspace header on this call:

```
blue-token-id: YOUR_TOKEN_ID
blue-token-secret: YOUR_TOKEN_SECRET
blue-org-id: YOUR_ORG_ID
blue-workspace-id: project_123
```

Headers are case-insensitive. `blue-org-id` and `blue-workspace-id` accept either an ID or a slug.

## Request

The only argument is `id`, the `CustomField.id` (the GraphQL `ID` for the field — the same value returned by [List Custom Fields](/api/custom-fields/list-custom-fields)). It is distinct from `CustomField.uid`; pass the `id`, not the `uid`.

```graphql
mutation DeleteCustomField {
  deleteCustomField(id: "field_123")
}
```

## Parameters

| Parameter | Type      | Required | Description                                  |
| --------- | --------- | -------- | -------------------------------------------- |
| `id`      | `String!` | Yes      | The `CustomField.id` of the field to delete. |

## Response

`deleteCustomField` returns the scalar `Boolean!`. It is `true` when the field was deleted.

```json
{
  "data": {
    "deleteCustomField": true
  }
}
```

### Returns

| Field               | Type       | Description                                        |
| ------------------- | ---------- | -------------------------------------------------- |
| `deleteCustomField` | `Boolean!` | `true` when the field and its values were removed. |

## What happens when you delete

Deletion soft-deletes the field and hides it from every read, but preserves what's needed for a full-fidelity restore:

- **Field values are kept.** The value the field held on every record stays in place so a restore brings them all back. They are removed only when the field is purged (by **Delete forever** or at the end of the retention window).
- **Field options are kept** for option-based fields (single-/multi-select, conversion fields) for the same reason.
- **Button automations go dormant** — if the field's type is `BUTTON`, the button can no longer render, so its automations stop firing. They are not deleted, and resume when the field is restored.
- **Rollup fields that referenced it** — any rollup that used the field as its source or aggregate has its computed values cleared and is marked `BROKEN_SOURCE`, so the app surfaces the broken state instead of a stale number. Restoring the field re-enqueues a recompute that clears it.
- **Activity** — activity entries tied to the field are deleted and a deletion is published to live clients.
- **Formula and time-duration fields** — formula results and time-duration calculations across the workspace are recomputed to account for the hidden field.
- **Charts** — charts in the workspace are flagged for refresh.
- **Webhook** — a `CUSTOM_FIELD_DELETED` event fires for the workspace (see [Webhooks](/api/webhooks)).

## Restoring a deleted field

While the field is in Trash, an `OWNER` or `ADMIN` can restore it with the `restoreCustomField` mutation, which returns the restored `CustomField`. Its values come back, dependent rollups recompute, and button automations resume.

```graphql
mutation RestoreCustomField {
  restoreCustomField(id: "field_123") {
    id
    name
  }
}
```

Once the field is purged — via **Delete forever** or by reaching the end of the organization's retention window — it can no longer be restored.

## Errors

| Code                          | When                                                                                                     |
| ----------------------------- | -------------------------------------------------------------------------------------------------------- |
| `CUSTOM_FIELD_NOT_FOUND`      | No custom field exists with the given `id`.                                                              |
| `FIELD_REFERENCED_BY_FORMULA` | The field is used in the workspace's record-name formula. Remove it from the formula first, then delete. |
| `FORBIDDEN`                   | The caller is not an `OWNER` or `ADMIN` of the workspace, or the workspace is archived.                  |

```json
{
  "errors": [
    {
      "message": "Cannot delete field \"Priority\" - it is used in the project name formula.",
      "extensions": { "code": "FIELD_REFERENCED_BY_FORMULA" }
    }
  ]
}
```

## Permissions

Only an `OWNER` or `ADMIN` of the workspace can delete custom fields, and the workspace must be active. `MEMBER` and `CLIENT` roles cannot.

## Related

- [Create a Custom Field](/api/custom-fields/create-custom-fields) — re-add a field after deletion.
- [List Custom Fields](/api/custom-fields/list-custom-fields) — find the `id` of the field to delete.
- [Set Custom Field Values](/api/custom-fields/custom-field-values) — write values to a field.
- [Delete a single option](#delete-a-single-option) — remove one choice without deleting the field.

## Delete a single option

To remove one choice from a single-/multi-select or conversion field without deleting the field itself, use `deleteCustomFieldOption` instead. It returns `Boolean!`.

```graphql
mutation DeleteCustomFieldOption {
  deleteCustomFieldOption(customFieldId: "field_123", optionId: "option_123")
}
```

| Parameter       | Type      | Required | Description                                        |
| --------------- | --------- | -------- | -------------------------------------------------- |
| `customFieldId` | `String!` | Yes      | The `CustomField.id` the option belongs to.        |
| `optionId`      | `String!` | Yes      | The `CustomFieldOption.id` to remove.              |
| `todoId`        | `String`  | No       | Optional record (`Todo.id`) scope for the removal. |
