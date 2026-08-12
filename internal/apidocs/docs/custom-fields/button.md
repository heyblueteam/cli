---
title: Button Field
description: Create button fields that fire a CUSTOM_FIELD_BUTTON_CLICKED automation trigger when a user clicks them on a record.
icon: MousePointerClick
order: 24
---

A button field renders a clickable button on a record. Unlike every other field type, it stores no value — clicking it fires a `CUSTOM_FIELD_BUTTON_CLICKED` event that your automations can listen for. Use it to give people a one-click way to kick off a workflow (send an invoice, sync to a CRM, notify a channel) from inside a record.

Button fields are `CustomField` objects with `type: BUTTON`. Records are `Record` objects, and workspaces are `Workspace` objects in the API.

## Overview

`BUTTON` is a value of the `CustomFieldType` enum. A button field carries presentation config (`buttonType`, `buttonConfirmText`, `buttonColor`) and, optionally, the `enableBulkAction` flag that lets it be clicked across many records at once. None of these affect the API contract — the click itself always executes immediately.

## Create

Custom fields are scoped to a workspace by the `blue-workspace-id` header — there is no `projectId` input field. Pass `type: BUTTON` and a `name`:

```graphql
mutation CreateButtonField {
  createCustomField(input: { name: "Send Invoice", type: BUTTON }) {
    id
    name
    type
  }
}
```

Add the optional confirmation and styling config. `buttonType` and `buttonConfirmText` are UI hints for clients that want to show a confirmation dialog; `buttonColor` is a hex string for the button's color:

```graphql
mutation CreateConfirmingButton {
  createCustomField(
    input: {
      name: "Delete All Attachments"
      type: BUTTON
      description: "Permanently removes every attachment on this record"
      buttonType: "hardConfirmation"
      buttonConfirmText: "DELETE"
      buttonColor: "#E5484D"
      enableBulkAction: true
    }
  ) {
    id
    name
    type
    buttonType
    buttonConfirmText
    buttonColor
  }
}
```

### CreateCustomFieldInput (button fields)

| Parameter           | Type               | Required | Description                                                                                        |
| ------------------- | ------------------ | -------- | -------------------------------------------------------------------------------------------------- |
| `name`              | `String!`          | Yes      | Display label for the button.                                                                      |
| `type`              | `CustomFieldType!` | Yes      | Must be `BUTTON`.                                                                                  |
| `description`       | `String`           | No       | Help text shown alongside the field.                                                               |
| `buttonType`        | `String`           | No       | Free-form UI hint for confirmation behavior (e.g. `"hardConfirmation"`). Not validated by the API. |
| `buttonConfirmText` | `String`           | No       | Text a UI client may require the user to type before confirming. Not enforced by the API.          |
| `buttonColor`       | `String`           | No       | Hex color for the button (e.g. `"#0091FF"`).                                                       |
| `enableBulkAction`  | `Boolean`          | No       | When `true`, the button can be clicked across many records with `bulkClickButton`.                 |

`CreateCustomFieldInput` carries config for every field type; only the parameters above apply to buttons. See [Create a Custom Field](/api/custom-fields/create-custom-fields) for the full input.

## Set a value

Buttons have no value to set. "Setting" a button means clicking it. Call `setRecordCustomField` with just the `todoId` and `customFieldId` — clicking fires the `CUSTOM_FIELD_BUTTON_CLICKED` automation trigger for that record. The mutation returns `Boolean!`:

```graphql
mutation ClickButton {
  setRecordCustomField(input: { todoId: "todo_123", customFieldId: "field_123" })
}
```

```json
{ "data": { "setRecordCustomField": true } }
```

<Callout variant="warning" title="Clicks execute immediately">

The API never enforces confirmation. `buttonType` and `buttonConfirmText` are stored only for UI clients to build confirmation dialogs — the API does not check them, so any `setRecordCustomField` call fires the event right away. Confirmation is purely a client-side safety feature.

</Callout>

### Click across many records

When the field was created with `enableBulkAction: true`, use `bulkClickButton` to click it on every record matching a filter or an explicit list of IDs. It returns a `BulkOperationResult` with how many records were processed:

```graphql
mutation BulkClickButton {
  bulkClickButton(input: { customFieldId: "field_123", todoIds: ["todo_123", "todo_456"] }) {
    success
    count
  }
}
```

```json
{ "data": { "bulkClickButton": { "success": true, "count": 2 } } }
```

Provide either `todoIds`, a `filter`, or both — at least one is required. Bulk clicks are capped at 1,000 records per call.

#### BulkClickButtonInput

| Parameter       | Type                     | Required | Description                                                           |
| --------------- | ------------------------ | -------- | --------------------------------------------------------------------- |
| `customFieldId` | `String!`                | Yes      | ID of the button field. The field must have `enableBulkAction: true`. |
| `todoIds`       | `[String!]`              | No       | Explicit record IDs to click the button on.                           |
| `filter`        | `UpdateTodosInputFilter` | No       | Match records by list and filter instead of listing IDs.              |

## Read a value

A button field never stores data. When you read it on a record, `CustomField.value` is `null` and there are no selected options, numbers, or text to read back. Clicks are recorded in the record's action history, not as a field value.

```graphql
query ReadButton {
  recordQueries {
    todos(filter: { companyIds: ["company_123"], todoListIds: ["list_123"] }) {
      items {
        id
        customFields {
          name
          type
          value
        }
      }
    }
  }
}
```

```json
{
  "data": {
    "recordQueries": {
      "todos": {
        "items": [
          {
            "id": "clm4n8qwx000008l0g4oxdqn7",
            "customFields": [{ "name": "Send Invoice", "type": "BUTTON", "value": null }]
          }
        ]
      }
    }
  }
}
```

## Notes

- **No stored value.** Each click emits an event and records an action in the record's history; it never modifies a field value. Don't expect `value`, `selectedOption`, or `text` to be populated.
- **Automation linkage is behavior, not schema.** The `CUSTOM_FIELD_BUTTON_CLICKED` trigger and the `button_clicked` action are wired up in the automation engine — they aren't fields you set on the button itself. Build the automation that listens for the click with [Create an automation](/api/automations/create-automation).
- **Bulk requires opt-in.** `bulkClickButton` rejects a field that wasn't created (or edited) with `enableBulkAction: true`, and rejects any field whose `type` isn't `BUTTON`.

## Errors

| Code                     | When                                                                                                             |
| ------------------------ | ---------------------------------------------------------------------------------------------------------------- |
| `CUSTOM_FIELD_NOT_FOUND` | No button field with that `customFieldId` exists in the workspace.                                               |
| `TODO_NOT_FOUND`         | The `todoId` doesn't exist or you can't access it.                                                              |
| `FORBIDDEN`              | You lack permission to create the field, or the button is marked non-editable for your role.                     |
| `BAD_USER_INPUT`         | Bulk click on a non-button field, on a field without `enableBulkAction`, or with neither `todoIds` nor `filter`. |

## Related

- [Create a Custom Field](/api/custom-fields/create-custom-fields)
- [Set Custom Field Values](/api/custom-fields/custom-field-values)
- [Custom Fields Overview](/api/custom-fields)
- [Create an automation](/api/automations/create-automation)
