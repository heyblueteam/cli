---
title: Create a Custom Field
description: Add a custom field to a workspace with type-specific configuration via the createCustomField mutation.
icon: Plus
order: 1
---

Use the `createCustomField` mutation to add a structured data field to a workspace. Custom fields are `CustomField` objects scoped to a single workspace (a `Project` in the API). The new field is created at the end of the workspace's field list and becomes available on every record (`Todo`) immediately.

The field's behavior is driven by its `type` (a `CustomFieldType` enum value). Each type reads a different subset of the input — for example a `CURRENCY` field uses `currency`/`min`/`max`, a `REFERENCE` field uses `referenceProjectId`, and a `LOOKUP` field uses `lookupOption`. See [Field Types](#field-types) below for the per-type configuration.

The target workspace comes from the `X-Bloo-Project-ID` request header, **not** from the input — there is no `projectId` field on `CreateCustomFieldInput`.

```
POST https://api.blue.app/graphql
X-Bloo-Token-ID: YOUR_TOKEN_ID
X-Bloo-Token-Secret: YOUR_TOKEN_SECRET
X-Bloo-Company-ID: YOUR_COMPANY_ID
X-Bloo-Project-ID: project_123
```

Company and project headers accept an ID or a slug. Header names are case-insensitive. See [Authentication](/api/start-guide/authentication) for the full header reference.

## Request

The smallest call that works: a single-line text field. Only `name` and `type` are required.

```graphql
mutation CreateTextField {
  createCustomField(input: { name: "Customer Name", type: TEXT_SINGLE }) {
    id
    uid
    name
    type
    position
  }
}
```

## Parameters

### CreateCustomFieldInput

| Parameter                | Type                                                                        | Required | Applies to                                | Description                                                     |
| ------------------------ | --------------------------------------------------------------------------- | -------- | ----------------------------------------- | --------------------------------------------------------------- |
| `name`                   | `String!`                                                                   | Yes      | all                                       | Display name of the field, as shown in the workspace.           |
| `type`                   | `CustomFieldType!`                                                          | Yes      | all                                       | The field type. See [Field Types](#field-types).                |
| `description`            | `String`                                                                    | No       | all                                       | Optional help text explaining the field.                        |
| `min`                    | `Float`                                                                     | No       | `NUMBER`, `CURRENCY`, `PERCENT`, `RATING` | Minimum allowed value.                                          |
| `max`                    | `Float`                                                                     | No       | `NUMBER`, `CURRENCY`, `PERCENT`, `RATING` | Maximum allowed value.                                          |
| `currency`               | `String`                                                                    | No       | `CURRENCY`                                | ISO 4217 currency code, e.g. `"USD"`.                           |
| `prefix`                 | `String`                                                                    | No       | `UNIQUE_ID`                               | Text prepended to each generated identifier.                    |
| `useSequenceUniqueId`    | `Boolean`                                                                   | No       | `UNIQUE_ID`                               | Generate sequential numbers instead of random IDs.              |
| `sequenceDigits`         | `Int`                                                                       | No       | `UNIQUE_ID`                               | Zero-padded width of the sequence (e.g. `5` → `00001`).         |
| `sequenceStartingNumber` | `Int`                                                                       | No       | `UNIQUE_ID`                               | First number in the sequence.                                   |
| `isDueDate`              | `Boolean`                                                                   | No       | `DATE`                                    | Treat this date field as the record's due date.                 |
| `formula`                | `JSON`                                                                      | No       | `FORMULA`                                 | Formula configuration.                                          |
| `urlDisplayAsButton`     | `Boolean`                                                                   | No       | `URL`                                     | Render the URL as a clickable button.                           |
| `urlButtonLabel`         | `String`                                                                    | No       | `URL`                                     | Button label when `urlDisplayAsButton` is true.                 |
| `urlButtonColor`         | `String`                                                                    | No       | `URL`                                     | Button color (hex) when displayed as a button.                  |
| `buttonType`             | `String`                                                                    | No       | `BUTTON`                                  | Action the button performs.                                     |
| `buttonConfirmText`      | `String`                                                                    | No       | `BUTTON`                                  | Confirmation prompt shown before the action runs.               |
| `buttonColor`            | `String`                                                                    | No       | `BUTTON`                                  | Button color (hex).                                             |
| `currencyFieldId`        | `String`                                                                    | No       | `CURRENCY_CONVERSION`                     | The source `CURRENCY` field to convert from.                    |
| `conversionDate`         | `String`                                                                    | No       | `CURRENCY_CONVERSION`                     | Exchange-rate date.                                             |
| `conversionDateType`     | `String`                                                                    | No       | `CURRENCY_CONVERSION`                     | How the conversion date is resolved.                            |
| `referenceProjectId`     | `String`                                                                    | No\*     | `REFERENCE`                               | Workspace to link records from. **Required for `REFERENCE`.**   |
| `referenceMultiple`      | `Boolean`                                                                   | No       | `REFERENCE`                               | Allow linking more than one record.                             |
| `referenceFilter`        | [`TodoFilterInput`](#todofilterinput)                                       | No       | `REFERENCE`                               | Restrict which records can be linked.                           |
| `lookupOption`           | [`CustomFieldLookupOptionInput`](#customfieldlookupoptioninput)             | No\*     | `LOOKUP`                                  | Lookup configuration. **Required for `LOOKUP`.**                |
| `referencedByOption`     | [`CustomFieldReferencedByOptionInput`](#customfieldreferencedbyoptioninput) | No       | `REFERENCED_BY`                           | Reverse-reference configuration.                                |
| `rollupOption`           | [`CustomFieldRollupOptionInput`](#customfieldrollupoptioninput)             | No\*     | `ROLLUP`                                  | Rollup aggregation configuration. **Required for `ROLLUP`.**    |
| `assigneeMultiple`       | `Boolean`                                                                   | No       | `ASSIGNEE`                                | Allow assigning more than one person.                           |
| `assigneeNotifyOnAssign` | `Boolean`                                                                   | No       | `ASSIGNEE`                                | Notify users when assigned via this field.                      |
| `allowedUserIds`         | `[String!]`                                                                 | No       | `ASSIGNEE`                                | Restrict assignable users to these user IDs.                    |
| `allowedRoleIds`         | `[String!]`                                                                 | No       | `ASSIGNEE`                                | Restrict assignable users to these custom-role IDs.             |
| `allowedLevels`          | `[UserAccessLevel!]`                                                        | No       | `ASSIGNEE`                                | Restrict assignable users to these access levels.               |
| `timeDurationDisplay`    | `CustomFieldTimeDurationDisplayType`                                        | No\*     | `TIME_DURATION`                           | Display format. **Required for `TIME_DURATION`.**               |
| `timeDurationTargetTime` | `Float`                                                                     | No       | `TIME_DURATION`                           | Target duration, in seconds.                                    |
| `timeDurationStartInput` | [`CustomFieldTimeDurationInput`](#customfieldtimedurationinput)             | No\*     | `TIME_DURATION`                           | Start trigger. **Required for `TIME_DURATION`.**                |
| `timeDurationEndInput`   | [`CustomFieldTimeDurationInput`](#customfieldtimedurationinput)             | No\*     | `TIME_DURATION`                           | End trigger. **Required for `TIME_DURATION`.**                  |
| `enableBulkAction`       | `Boolean`                                                                   | No       | supported types                           | Enable bulk-set across records. Rejected for unsupported types. |

\* Schema-optional, but the resolver requires it for the listed `type` and returns `BAD_USER_INPUT` if omitted. See [Errors](#errors).

> `SELECT_SINGLE` and `SELECT_MULTI` take no inline options — create the field first, then add options with [`createCustomFieldOptions`](#adding-select-options).

## Response

```json
{
  "data": {
    "createCustomField": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "uid": "cf_8h2k9d",
      "name": "Customer Name",
      "type": "TEXT_SINGLE",
      "position": 16384
    }
  }
}
```

### Returns

`createCustomField` returns the created `CustomField`. Commonly selected fields:

| Field                           | Type                            | Description                                                    |
| ------------------------------- | ------------------------------- | -------------------------------------------------------------- |
| `id`                            | `ID!`                           | Unique identifier of the field.                                |
| `uid`                           | `String!`                       | Short public identifier.                                       |
| `name`                          | `String!`                       | Display name.                                                  |
| `type`                          | `CustomFieldType!`              | The field type.                                                |
| `position`                      | `Float!`                        | Sort position within the workspace (auto-assigned at the end). |
| `description`                   | `String`                        | Help text, if set.                                             |
| `createdAt`                     | `DateTime!`                     | Creation timestamp.                                            |
| `updatedAt`                     | `DateTime!`                     | Last update timestamp.                                         |
| `min` / `max`                   | `Float`                         | Configured bounds (numeric types).                             |
| `currency`                      | `String`                        | Currency code (`CURRENCY` type).                               |
| `prefix`                        | `String`                        | ID prefix (`UNIQUE_ID` type).                                  |
| `isDueDate`                     | `Boolean`                       | Due-date flag (`DATE` type).                                   |
| `formula`                       | `JSON`                          | Formula config (`FORMULA` type).                               |
| `referenceProject`              | `Project`                       | Linked workspace (`REFERENCE` type).                           |
| `customFieldLookupOption`       | `CustomFieldLookupOption`       | Lookup config (`LOOKUP` type).                                 |
| `customFieldReferencedByOption` | `CustomFieldReferencedByOption` | Reverse-reference config (`REFERENCED_BY` type).               |
| `customFieldRollupOption`       | `CustomFieldRollupOption`       | Rollup config (`ROLLUP` type).                                 |

## Field Types

`CustomFieldType` has 26 values. Most types are created with just `name` + `type`; the table notes the extra input each type reads.

| `type`                | Stores                                | Extra input                                                                                       |
| --------------------- | ------------------------------------- | ------------------------------------------------------------------------------------------------- |
| `TEXT_SINGLE`         | Single-line text                      | —                                                                                                 |
| `TEXT_MULTI`          | Multi-line text                       | —                                                                                                 |
| `SELECT_SINGLE`       | One choice from a list                | Options created separately                                                                        |
| `SELECT_MULTI`        | Multiple choices from a list          | Options created separately                                                                        |
| `CHECKBOX`            | Boolean                               | —                                                                                                 |
| `RATING`              | Star rating                           | `max` (defaults to 5)                                                                             |
| `PHONE`               | Phone number                          | —                                                                                                 |
| `NUMBER`              | Number                                | `min`, `max`                                                                                      |
| `CURRENCY`            | Money amount                          | `currency`, `min`, `max`                                                                          |
| `PERCENT`             | Percentage                            | `min`, `max`                                                                                      |
| `EMAIL`               | Email address                         | —                                                                                                 |
| `URL`                 | Web link                              | `urlDisplayAsButton`, `urlButtonLabel`, `urlButtonColor`                                          |
| `UNIQUE_ID`           | Auto-generated identifier             | `prefix`, `useSequenceUniqueId`, `sequenceDigits`, `sequenceStartingNumber`                       |
| `LOCATION`            | Geographic coordinates                | —                                                                                                 |
| `FILE`                | File attachments                      | —                                                                                                 |
| `COUNTRY`             | Country                               | —                                                                                                 |
| `DATE`                | Date                                  | `isDueDate`                                                                                       |
| `FORMULA`             | Calculated value                      | `formula` (required)                                                                              |
| `REFERENCE`           | Links to records in another workspace | `referenceProjectId` (required), `referenceMultiple`, `referenceFilter`                           |
| `LOOKUP`              | Pulled value from a referenced record | `lookupOption` (required)                                                                         |
| `REFERENCED_BY`       | Records that reference this one       | `referencedByOption`                                                                              |
| `ROLLUP`              | Aggregate over referenced records     | `rollupOption` (required)                                                                         |
| `TIME_DURATION`       | Elapsed time between two events       | `timeDurationDisplay`, `timeDurationStartInput`, `timeDurationEndInput` (all required)            |
| `BUTTON`              | Action trigger                        | `buttonType`, `buttonConfirmText`, `buttonColor`                                                  |
| `CURRENCY_CONVERSION` | Converted currency value              | `currencyFieldId`, `conversionDate`, `conversionDateType`                                         |
| `ASSIGNEE`            | Assigned people                       | `assigneeMultiple`, `assigneeNotifyOnAssign`, `allowedUserIds`, `allowedRoleIds`, `allowedLevels` |

Each type has a dedicated page with set/read examples — see [Related](#related).

## Full examples

### Number field with bounds

```graphql
mutation CreateQuantityField {
  createCustomField(input: { name: "Quantity", type: NUMBER, min: 1, max: 999 }) {
    id
    name
    min
    max
  }
}
```

### Currency field

```graphql
mutation CreateBudgetField {
  createCustomField(input: { name: "Budget", type: CURRENCY, currency: "EUR", min: 0 }) {
    id
    name
    currency
  }
}
```

### Date field flagged as the due date

```graphql
mutation CreateDeadlineField {
  createCustomField(input: { name: "Deadline", type: DATE, isDueDate: true }) {
    id
    name
    isDueDate
  }
}
```

### Reference field

Links records to records in another workspace, restricted with `referenceFilter` (a `TodoFilterInput`).

```graphql
mutation CreateDependenciesField {
  createCustomField(
    input: {
      name: "Dependencies"
      type: REFERENCE
      referenceProjectId: "project_123"
      referenceMultiple: true
      referenceFilter: { todoListIds: ["list_123"], showCompleted: false }
    }
  ) {
    id
    name
    referenceMultiple
    referenceProject {
      id
      name
    }
  }
}
```

### Lookup field

Pulls a value from records linked by a `REFERENCE` field. Point `referenceId` at the `REFERENCE` field's ID and `lookupType` at what to read; for `TODO_CUSTOM_FIELD`, also set `lookupId` to the source field's ID.

```graphql
mutation CreateLookupField {
  createCustomField(
    input: {
      name: "Customer Email"
      type: LOOKUP
      lookupOption: {
        referenceId: "field_123"
        lookupId: "field_456"
        lookupType: TODO_CUSTOM_FIELD
      }
    }
  ) {
    id
    name
    customFieldLookupOption {
      lookupType
      reference {
        id
        name
      }
      lookup {
        id
        name
      }
    }
  }
}
```

### Rollup field

Aggregates a field across the records reached through a `REFERENCE` or `REFERENCED_BY` field. Rollup fields require a Pro subscription.

```graphql
mutation CreateTotalField {
  createCustomField(
    input: {
      name: "Total Order Value"
      type: ROLLUP
      rollupOption: {
        sourceType: REFERENCE_FIELD
        sourceFieldId: "field_123"
        aggregateFieldId: "field_456"
        aggregationFunction: SUM
      }
    }
  ) {
    id
    name
    customFieldRollupOption {
      sourceType
      aggregationFunction
    }
  }
}
```

### Assignee field

A custom assignee field, optionally limited to a set of users, roles, or access levels.

```graphql
mutation CreateReviewerField {
  createCustomField(
    input: {
      name: "Reviewer"
      type: ASSIGNEE
      assigneeMultiple: false
      assigneeNotifyOnAssign: true
      allowedLevels: [ADMIN, MEMBER]
    }
  ) {
    id
    name
    assigneeMultiple
    allowedLevels
  }
}
```

### Unique ID with a sequence

```graphql
mutation CreateOrderNumberField {
  createCustomField(
    input: {
      name: "Order Number"
      type: UNIQUE_ID
      prefix: "ORD-"
      useSequenceUniqueId: true
      sequenceDigits: 6
      sequenceStartingNumber: 1000
    }
  ) {
    id
    name
    prefix
  }
}
```

Sequence numbers are assigned by a background job, so existing records receive their IDs shortly after creation.

### Time-duration field

Measures elapsed time between a start and end event. Both ends take a `CustomFieldTimeDurationInput`.

```graphql
mutation CreateTimeToResolution {
  createCustomField(
    input: {
      name: "Time to Resolution"
      type: TIME_DURATION
      timeDurationDisplay: FULL_DATE_STRING
      timeDurationStartInput: { type: TODO_CREATED_AT, condition: FIRST }
      timeDurationEndInput: { type: TODO_MARKED_AS_COMPLETE, condition: LAST }
    }
  ) {
    id
    name
  }
}
```

## Adding select options

`SELECT_SINGLE` and `SELECT_MULTI` fields have no choices until you add them. Create the field, then call `createCustomFieldOptions` with the new field's `id`. It takes one `CreateCustomFieldOptionsInput` and returns `[CustomFieldOption!]`.

```graphql
mutation CreateSelectOptions {
  createCustomFieldOptions(
    input: {
      customFieldId: "field_123"
      customFieldOptions: [
        { title: "High", color: "#FF0000", position: 1 }
        { title: "Medium", color: "#FFA500", position: 2 }
        { title: "Low", color: "#00FF00", position: 3 }
      ]
    }
  ) {
    id
    title
    color
    position
  }
}
```

`color` and `position` are optional on each `CustomFieldOptionInput`. To add a single option, use the singular `createCustomFieldOption` mutation, which takes `customFieldId` directly in its input.

## Nested input reference

### TodoFilterInput

Used by `referenceFilter` to restrict which records a `REFERENCE` field can link. All fields are optional; common ones:

| Field           | Type        | Description                                        |
| --------------- | ----------- | -------------------------------------------------- |
| `todoListIds`   | `[String!]` | Limit to records in these lists.                   |
| `tagIds`        | `[String!]` | Limit to records carrying these tags.              |
| `assigneeIds`   | `[String!]` | Limit to records assigned to these users.          |
| `showCompleted` | `Boolean`   | Include completed records (default excludes them). |
| `recordName`    | `String`    | Match by record name.                              |

`TodoFilterInput` carries many more fields (date ranges, negation arrays, nested groups). See [List Records](/api/records/list-records) for the full filter surface.

### CustomFieldLookupOptionInput

| Field         | Type                     | Required | Description                                                          |
| ------------- | ------------------------ | -------- | -------------------------------------------------------------------- |
| `referenceId` | `String!`                | Yes      | ID of the `REFERENCE` field this lookup reads through.               |
| `lookupType`  | `CustomFieldLookupType!` | Yes      | What to pull from the referenced record.                             |
| `lookupId`    | `String`                 | No       | Source field ID — required when `lookupType` is `TODO_CUSTOM_FIELD`. |

**CustomFieldLookupType** values: `TODO_DUE_DATE`, `TODO_CREATED_AT`, `TODO_UPDATED_AT`, `TODO_TAG`, `TODO_ASSIGNEE`, `TODO_DESCRIPTION`, `TODO_LIST`, `TODO_CUSTOM_FIELD`. The `TODO_REFERENCED_BY` value is **deprecated** — create a `REFERENCED_BY` field and point a `LOOKUP` at it instead.

### CustomFieldReferencedByOptionInput

Configures a `REFERENCED_BY` field. All fields are optional; omit them to surface every record that references the current one.

| Field                  | Type              | Description                                                                                       |
| ---------------------- | ----------------- | ------------------------------------------------------------------------------------------------- |
| `referencingProjectId` | `String`          | Limit to records in this workspace. Must belong to the same organization.                         |
| `referencingFieldId`   | `String`          | Limit to records linked through this specific `REFERENCE` field. Requires `referencingProjectId`. |
| `filters`              | `TodoFilterInput` | Restrict which referencing records count.                                                         |

### CustomFieldRollupOptionInput

| Field                 | Type                         | Required | Description                                                        |
| --------------------- | ---------------------------- | -------- | ------------------------------------------------------------------ |
| `sourceType`          | `RollupSourceType!`          | Yes      | `REFERENCE_FIELD` or `REFERENCED_BY_FIELD`.                        |
| `aggregationFunction` | `RollupAggregationFunction!` | Yes      | `SUM`, `AVERAGE`, `COUNT`, `COUNTA`, `MIN`, `MAX`, or `ARRAYJOIN`. |
| `sourceFieldId`       | `String`                     | No       | The `REFERENCE`/`REFERENCED_BY` field to traverse.                 |
| `aggregateFieldId`    | `String`                     | No       | The field on the related records to aggregate.                     |
| `filters`             | `TodoFilterInput`            | No       | Restrict which related records are included.                       |

### CustomFieldTimeDurationInput

| Field                  | Type                                | Required | Description                                         |
| ---------------------- | ----------------------------------- | -------- | --------------------------------------------------- |
| `type`                 | `CustomFieldTimeDurationType!`      | Yes      | The event that triggers this end of the duration.   |
| `condition`            | `CustomFieldTimeDurationCondition!` | Yes      | `FIRST` or `LAST` occurrence of the event.          |
| `customFieldId`        | `String`                            | No       | Field to watch when `type` is `TODO_CUSTOM_FIELD`.  |
| `customFieldOptionIds` | `[String!]`                         | No       | Option(s) that trigger the event.                   |
| `todoListId`           | `String`                            | No       | List to watch when `type` is `TODO_MOVED`.          |
| `tagId`                | `String`                            | No       | Tag to watch when `type` is `TODO_TAG_ADDED`.       |
| `assigneeId`           | `String`                            | No       | User to watch when `type` is `TODO_ASSIGNEE_ADDED`. |

**CustomFieldTimeDurationType** values: `TODO_CREATED_AT`, `TODO_CUSTOM_FIELD`, `TODO_DUE_DATE`, `TODO_MARKED_AS_COMPLETE`, `TODO_MOVED`, `TODO_TAG_ADDED`, `TODO_ASSIGNEE_ADDED`.

**CustomFieldTimeDurationCondition** values: `FIRST`, `LAST`.

**CustomFieldTimeDurationDisplayType** values: `FULL_DATE`, `FULL_DATE_STRING`, `FULL_DATE_SUBSTRING`.

## Errors

| Code                 | When                                                                                                                                                                                                                                                |
| -------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `FORBIDDEN`          | The caller is not an OWNER or ADMIN of the workspace, or the workspace is inactive.                                                                                                                                                                 |
| `PLAN_LIMIT_REACHED` | The workspace has hit its custom-field limit, or the organization is locked.                                                                                                                                                                        |
| `PRO_REQUIRED`       | A `ROLLUP` field was requested without a Pro subscription.                                                                                                                                                                                          |
| `BAD_USER_INPUT`     | A type-specific requirement was not met — e.g. `REFERENCE` without `referenceProjectId`, `LOOKUP` without `lookupOption`, `ROLLUP` without `rollupOption`, `TIME_DURATION` missing display/start/end, or `enableBulkAction` on an unsupported type. |
| `PROJECT_NOT_FOUND`  | `referenceProjectId` (or a `REFERENCED_BY` referencing project) does not exist or the caller has no access to it.                                                                                                                                   |

## Permissions

Creating a custom field requires **OWNER** or **ADMIN** access to the workspace, and the workspace must be active. MEMBER, CLIENT, COMMENT_ONLY, and VIEW_ONLY roles cannot create fields. `ROLLUP` fields additionally require the organization to be on a Pro plan.

## Related

- [List custom fields](/api/custom-fields/list-custom-fields) — read the fields in a workspace.
- [Set custom field values](/api/custom-fields/custom-field-values) — write values onto records.
- [Delete a custom field](/api/custom-fields/delete-custom-field) — remove a field.
- [Reference field](/api/custom-fields/reference) · [Lookup field](/api/custom-fields/lookup) · [Currency conversion field](/api/custom-fields/currency-conversion) — type-specific guides.
- [Authentication](/api/start-guide/authentication) — request headers and tokens.
