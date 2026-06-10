---
title: List Custom Fields
description: Read the custom fields in a workspace with type filtering and sorting via the customFields query.
icon: List
order: 2
---

Use the `customFields` query to read the custom fields defined in a workspace, optionally filtered by field type and sorted. Custom fields are `CustomField` objects scoped to a single workspace (a `Project` in the API). The result is a `CustomFieldPagination` page of `items` plus `pageInfo`.

The workspace is resolved from the `filter.projectId` argument, or — when that is omitted — from the `X-Bloo-Project-ID` request header. Only fields visible to the calling user's project role are returned: if the user has a custom role with restricted field access, fields hidden from that role are excluded from `items`.

```
POST https://api.blue.app/graphql
X-Bloo-Token-ID: YOUR_TOKEN_ID
X-Bloo-Token-Secret: YOUR_TOKEN_SECRET
X-Bloo-Company-ID: YOUR_COMPANY_ID
X-Bloo-Project-ID: project_123
```

Company and project headers accept an ID or a slug. Header names are case-insensitive. See [Authentication](/api/start-guide/authentication) for the full header reference.

## Request

The smallest call that works. With the `X-Bloo-Project-ID` header set, `filter` is optional; passing `filter.projectId` overrides the header.

```graphql
query ListCustomFields {
  customFields(filter: { projectId: "project_123" }) {
    items {
      id
      uid
      name
      type
      position
    }
    pageInfo {
      totalItems
      hasNextPage
    }
  }
}
```

## Parameters

The `customFields` query takes these arguments:

| Argument | Type                     | Required | Description                                                      |
| -------- | ------------------------ | -------- | ---------------------------------------------------------------- |
| `filter` | `CustomFieldFilterInput` | No       | Scope the result by workspace and field type.                    |
| `sort`   | `CustomFieldSort`        | No       | A single sort key. Defaults to `position_ASC`.                   |
| `skip`   | `Int`                    | No       | Number of fields to skip for offset pagination. Defaults to `0`. |
| `take`   | `Int`                    | No       | Number of fields to return. Defaults to `500`.                   |

### CustomFieldFilterInput

| Field       | Type                 | Required | Description                                                                                          |
| ----------- | -------------------- | -------- | ---------------------------------------------------------------------------------------------------- |
| `projectId` | `String`             | No       | Workspace ID or slug to read fields from. Falls back to the `X-Bloo-Project-ID` header when omitted. |
| `types`     | `[CustomFieldType!]` | No       | Return only fields of these types.                                                                   |

### CustomFieldSort

A single enum value (not a list on this query).

| Value            | Description                                                                      |
| ---------------- | -------------------------------------------------------------------------------- |
| `position_ASC`   | By workspace position, ascending (default — matches the order shown in the app). |
| `position_DESC`  | By workspace position, descending.                                               |
| `name_ASC`       | By name, A–Z.                                                                    |
| `name_DESC`      | By name, Z–A.                                                                    |
| `createdAt_ASC`  | Oldest first.                                                                    |
| `createdAt_DESC` | Newest first.                                                                    |

### CustomFieldType

The `type` discriminates how a field stores and renders its value. Each value also drives which type-specific fields are populated on the returned `CustomField` (for example `currency` on `CURRENCY`, `customFieldOptions` on the `SELECT_*` types).

| Value                 | Description                                                           |
| --------------------- | --------------------------------------------------------------------- |
| `TEXT_SINGLE`         | Single-line text.                                                     |
| `TEXT_MULTI`          | Multi-line text.                                                      |
| `SELECT_SINGLE`       | Single-select dropdown (one option).                                  |
| `SELECT_MULTI`        | Multi-select dropdown (many options).                                 |
| `CHECKBOX`            | Boolean checkbox.                                                     |
| `RATING`              | Star rating.                                                          |
| `PHONE`               | Phone number.                                                         |
| `NUMBER`              | Numeric value.                                                        |
| `CURRENCY`            | Monetary amount in a fixed currency.                                  |
| `PERCENT`             | Percentage value.                                                     |
| `EMAIL`               | Email address.                                                        |
| `URL`                 | Web URL, optionally rendered as a button.                             |
| `UNIQUE_ID`           | Auto-generated identifier with an optional prefix.                    |
| `LOCATION`            | Geographic point (latitude/longitude).                                |
| `FILE`                | File attachments.                                                     |
| `COUNTRY`             | One or more country codes.                                            |
| `DATE`                | Date or date range.                                                   |
| `FORMULA`             | Value computed from other fields.                                     |
| `REFERENCE`           | Links to records in another workspace.                                |
| `LOOKUP`              | Pulls a value from a referenced record.                               |
| `TIME_DURATION`       | Elapsed time between two events.                                      |
| `BUTTON`              | Actionable button.                                                    |
| `CURRENCY_CONVERSION` | Amount converted between currencies.                                  |
| `ASSIGNEE`            | One or more assigned users.                                           |
| `REFERENCED_BY`       | Reverse side of a `REFERENCE` field (records that point at this one). |
| `ROLLUP`              | Aggregate (sum, count, …) over referenced records.                    |

## Response

```json
{
  "data": {
    "customFields": {
      "items": [
        {
          "id": "clm4n8qwx000008l0g4oxdqn7",
          "uid": "cf_qZ1r8",
          "name": "Stage",
          "type": "SELECT_SINGLE",
          "position": 1
        },
        {
          "id": "clm4n8r2p000108l0d3k1f9aa",
          "uid": "cf_7Hb2k",
          "name": "Deal Size",
          "type": "CURRENCY",
          "position": 2
        }
      ],
      "pageInfo": {
        "totalItems": 2,
        "hasNextPage": false
      }
    }
  }
}
```

### CustomFieldPagination

| Field        | Type              | Description                                 |
| ------------ | ----------------- | ------------------------------------------- |
| `items`      | `[CustomField!]!` | The page of custom fields.                  |
| `pageInfo`   | `PageInfo!`       | Pagination metadata.                        |
| `totalCount` | `Int`             | **Deprecated** — use `pageInfo.totalItems`. |

### CustomField

The fields below cover the common cases. `CustomField` exposes many more type-specific outputs (see [Field-type outputs](#field-type-outputs)).

| Field                | Type                   | Description                                             |
| -------------------- | ---------------------- | ------------------------------------------------------- |
| `id`                 | `ID!`                  | Unique identifier.                                      |
| `uid`                | `String!`              | Short human-readable identifier.                        |
| `name`               | `String!`              | Display name.                                           |
| `type`               | `CustomFieldType!`     | The field type.                                         |
| `position`           | `Float!`               | Order within the workspace.                             |
| `description`        | `String`               | Optional help text.                                     |
| `min`                | `Float`                | Minimum (`NUMBER`, `RATING`, `PERCENT`).                |
| `max`                | `Float`                | Maximum (`NUMBER`, `RATING`, `PERCENT`).                |
| `currency`           | `String`               | Currency code (`CURRENCY`).                             |
| `prefix`             | `String`               | Prefix for generated IDs (`UNIQUE_ID`).                 |
| `isDueDate`          | `Boolean`              | Whether a `DATE` field marks the record's due date.     |
| `formula`            | `JSON`                 | Formula configuration (`FORMULA`).                      |
| `customFieldOptions` | `[CustomFieldOption!]` | Selectable options (`SELECT_*`, `BUTTON`).              |
| `viewable`           | `Boolean`              | Whether the calling user can see this field's values.   |
| `editable`           | `Boolean`              | Whether the calling user can edit this field's values.  |
| `projectUserRole`    | `ProjectUserRole`      | The role context used to resolve `viewable`/`editable`. |
| `createdAt`          | `DateTime!`            | When the field was created.                             |
| `updatedAt`          | `DateTime!`            | When the field was last changed.                        |
| `metadata`           | `JSON`                 | **Deprecated** — no longer in use.                      |

#### Field-type outputs

When you fetch fields in the context of a specific record, these populate with that record's value. They are otherwise null.

| Field                                                        | Type                             | Description                                                            |
| ------------------------------------------------------------ | -------------------------------- | ---------------------------------------------------------------------- |
| `text`                                                       | `String`                         | Text value (`TEXT_*`, `EMAIL`, `URL`, `PHONE`).                        |
| `number`                                                     | `Float`                          | Numeric value (`NUMBER`, `CURRENCY`, `PERCENT`, `RATING`).             |
| `checked`                                                    | `Boolean`                        | Checkbox state (`CHECKBOX`).                                           |
| `selectedOption`                                             | `CustomFieldOption`              | Chosen option (`SELECT_SINGLE`).                                       |
| `selectedOptions`                                            | `[CustomFieldOption!]`           | Chosen options (`SELECT_MULTI`).                                       |
| `selectedTodos`                                              | `[Todo!]`                        | Linked records (`REFERENCE`).                                          |
| `files`                                                      | `[File!]`                        | Attached files (`FILE`).                                               |
| `latitude` / `longitude`                                     | `Float`                          | Coordinates (`LOCATION`).                                              |
| `startDate` / `endDate`                                      | `DateTime`                       | Date range (`DATE`).                                                   |
| `timezone`                                                   | `String`                         | Timezone (`DATE`).                                                     |
| `regionCode`                                                 | `String`                         | Region (`PHONE`).                                                      |
| `countryCodes`                                               | `[String!]`                      | Country codes (`COUNTRY`).                                             |
| `referenceMultiple`                                          | `Boolean`                        | Whether a `REFERENCE` field allows multiple records.                   |
| `customFieldLookupOption`                                    | `CustomFieldLookupOption`        | Lookup configuration (`LOOKUP`).                                       |
| `rollupValueNumeric` / `rollupValueDate` / `rollupValueText` | `String` / `DateTime` / `String` | Computed rollup value (`ROLLUP`); exactly one is populated per record. |
| `value`                                                      | `JSON`                           | The record's value as a type-agnostic JSON blob.                       |

### CustomFieldOption

Selectable options on `SELECT_*` and `BUTTON` fields. Commonly selected fields:

| Field      | Type      | Description             |
| ---------- | --------- | ----------------------- |
| `id`       | `ID!`     | Unique identifier.      |
| `title`    | `String!` | Display label.          |
| `color`    | `String!` | Hex color.              |
| `position` | `Float!`  | Order within the field. |

The type also exposes `buttonType`, `buttonConfirmText`, `currencyConversionFrom`, `currencyConversionTo`, `createdAt`, and `updatedAt` for button and currency-conversion fields.

### PageInfo

| Field             | Type       | Description                                             |
| ----------------- | ---------- | ------------------------------------------------------- |
| `totalItems`      | `Int`      | Total fields matching the filter.                       |
| `totalPages`      | `Int`      | Total pages at the current `take`.                      |
| `page`            | `Int`      | Current page number.                                    |
| `perPage`         | `Int`      | Page size (`take`).                                     |
| `hasNextPage`     | `Boolean!` | Whether more pages follow.                              |
| `hasPreviousPage` | `Boolean!` | Whether earlier pages exist.                            |
| `startCursor`     | `String`   | **Deprecated** — use offset pagination (`skip`/`take`). |
| `endCursor`       | `String`   | **Deprecated** — use offset pagination (`skip`/`take`). |

## Full example

Filter by type, page with `skip`/`take`, and select the options for select fields plus access flags.

```graphql
query ListSelectFields {
  customFields(
    filter: { projectId: "project_123", types: [SELECT_SINGLE, SELECT_MULTI] }
    sort: name_ASC
    skip: 0
    take: 50
  ) {
    items {
      id
      uid
      name
      type
      position
      description
      viewable
      editable
      customFieldOptions {
        id
        title
        color
        position
      }
    }
    pageInfo {
      totalItems
      totalPages
      page
      hasNextPage
      hasPreviousPage
    }
  }
}
```

To read fields across multiple workspaces, or to search fields by name, use the nested `customFieldQueries.customFields` query instead. It takes a `CustomFieldsFilterInput` where `companyIds` is **required** and `projectIds` is an optional plural list, adds a `search` argument, accepts `sort` as a **list** (`[CustomFieldSort!]`), supports a `distinct` argument, and defaults `take` to `20`.

```graphql
query SearchCustomFields {
  customFieldQueries {
    customFields(
      filter: { companyIds: ["company_123"], projectIds: ["project_123"], search: "amount" }
      sort: [name_ASC]
      take: 20
    ) {
      items {
        id
        name
        type
      }
      pageInfo {
        totalItems
        hasNextPage
      }
    }
  }
}
```

## Errors

| Code                | When                                                                                  |
| ------------------- | ------------------------------------------------------------------------------------- |
| `PROJECT_NOT_FOUND` | No workspace matches the resolved ID/slug, or the calling user is not a member of it. |
| `UNAUTHENTICATED`   | The request is missing or has invalid token headers.                                  |

## Permissions

Any project member can list a workspace's custom fields. Field-level visibility is enforced by the caller's project role: a field is included only when it is visible to that role. The returned `viewable` and `editable` flags report whether the caller can see and edit each field's values, and `projectUserRole` exposes the role used to resolve them.

## Related

- [Create a custom field](/api/custom-fields/create-custom-fields) — add a field to a workspace.
- [Set custom field values](/api/custom-fields/custom-field-values) — write values onto records.
- [Delete a custom field](/api/custom-fields/delete-custom-field) — remove a field.
- [Custom fields overview](/api/custom-fields) — the full list of field types and guides.
- [Authentication](/api/start-guide/authentication) — request headers and tokens.
