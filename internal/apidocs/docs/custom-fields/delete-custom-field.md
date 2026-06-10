---
title: Delete a Custom Field
description: Permanently remove a custom field from a workspace, clearing its values on every record.
icon: Trash2
order: 4
---

Use the `deleteCustomField` mutation to permanently remove a custom field from a workspace, along with its values on every record. Custom fields are `CustomField` objects in the API; workspaces are `Project` objects. This is irreversible — there is no soft delete or restore. The mutation returns `Boolean!`.

All custom-field requests are sent to `https://api.blue.app/graphql` with your authentication headers. A custom field is scoped to a single workspace, so include the workspace header on this call:

```
X-Bloo-Token-ID: YOUR_TOKEN_ID
X-Bloo-Token-Secret: YOUR_TOKEN_SECRET
X-Bloo-Company-ID: YOUR_COMPANY_ID
X-Bloo-Project-ID: project_123
```

Headers are case-insensitive. `X-Bloo-Company-ID` and `X-Bloo-Project-ID` accept either an ID or a slug.

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

## What gets cleaned up

Deleting a field is a cascade. The resolver performs the following before returning:

- **Field values** — the field's value on every record in the workspace is removed.
- **Field options** — for option-based fields (single-/multi-select, conversion fields), all of the field's `CustomFieldOption` rows are removed with it.
- **Activity** — activity entries tied to the field are deleted and a deletion is published to live clients.
- **Rollup fields that referenced it** — any rollup field that used the deleted field as its source or aggregate has its computed values cleared and is marked broken, so the app can surface the broken state instead of a stale number.
- **Button automations** — if the field's type is `BUTTON`, every automation triggered by that button is deleted.
- **Formula and time-duration fields** — formula results and time-duration calculations across the workspace are recomputed to account for the removed field.
- **Charts** — charts in the workspace are flagged for refresh.
- **Webhook** — a `CUSTOM_FIELD_DELETED` event fires for the workspace (see [Webhooks](/api/webhooks)).

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
