---
title: Location Custom Field
description: Store geographic coordinates (latitude and longitude) on a record, with coordinate-range validation when records are created.
icon: MapPin
order: 20
---

A location custom field stores a geographic point on a record as a latitude/longitude pair. It maps to the `LOCATION` value of the `CustomFieldType` enum. Records are `Record` objects and custom fields are `CustomField` objects in the API.

Location fields store coordinates only. There is no built-in geocoding, reverse geocoding, address validation, or map rendering — convert addresses to coordinates with an external service before writing the value.

## Overview

A location field holds two numeric values on each record:

- `latitude` — north/south position, between `-90` and `90`.
- `longitude` — east/west position, between `-180` and `180`.

Coordinates are stored in decimal degrees. The two mutations that write a value behave differently:

- **`createRecord`** validates the coordinate ranges and rejects out-of-range values.
- **`setRecordCustomField`** stores whatever `latitude`/`longitude` you send, with no range validation.

## Create

Create the field with `createCustomField`, using `type: LOCATION`. The field is scoped to the workspace (`Workspace`) named in the `X-Bloo-Project-ID` header — there is no `projectId` argument on the input.

```graphql
mutation CreateLocationField {
  createCustomField(input: { name: "Meeting Location", type: LOCATION }) {
    id
    name
    type
  }
}
```

```json
{
  "data": {
    "createCustomField": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "name": "Meeting Location",
      "type": "LOCATION"
    }
  }
}
```

Add an optional `description` to show help text alongside the field:

```graphql
mutation CreateLocationFieldWithHelp {
  createCustomField(
    input: {
      name: "Office Location"
      type: LOCATION
      description: "Primary office location coordinates"
    }
  ) {
    id
    name
    type
    description
  }
}
```

### CreateCustomFieldInput

| Parameter     | Type               | Required | Description                |
| ------------- | ------------------ | -------- | -------------------------- |
| `name`        | `String!`          | Yes      | Display name of the field. |
| `type`        | `CustomFieldType!` | Yes      | Must be `LOCATION`.        |
| `description` | `String`           | No       | Help text shown to users.  |

The custom field is attached to the workspace from the `blue-workspace-id` request header, not from an input argument. Company and project headers accept an ID or a slug.

## Set a value

Use `setRecordCustomField` to write a location to a record. Pass `latitude` and `longitude` as floats. This mutation returns `Boolean!` — it does not return the updated value, so read it back with a separate query.

`setRecordCustomField` performs **no** range validation: the values are stored exactly as provided. Send a valid latitude/longitude pair.

```graphql
mutation SetLocation {
  setRecordCustomField(
    input: { todoId: "todo_123", customFieldId: "field_123", latitude: 40.7128, longitude: -74.006 }
  )
}
```

```json
{
  "data": {
    "setRecordCustomField": true
  }
}
```

### SetRecordCustomFieldInput

| Parameter       | Type      | Required | Description                                   |
| --------------- | --------- | -------- | --------------------------------------------- |
| `todoId`        | `String!` | Yes      | The record to update.                         |
| `customFieldId` | `String!` | Yes      | The location field to set.                    |
| `latitude`      | `Float`   | No       | North/south position, between `-90` and `90`. |
| `longitude`     | `Float`   | No       | East/west position, between `-180` and `180`. |

Both `latitude` and `longitude` are optional in the schema, but a valid point needs both. Sending only one stores a partial value; supply the pair together.

### Setting a location when creating a record

`createRecord` is the only mutation that validates the coordinate ranges. Pass the location in `customFields` as a `CreateRecordInputCustomField` (`customFieldId` + a single `value` string), with the coordinates comma-separated as `"latitude,longitude"`. The resolver splits on the comma, parses both numbers, and stores them on `latitude`/`longitude`. If either coordinate is non-numeric or out of range, it throws `CUSTOM_FIELD_VALUE_PARSE_ERROR` with the message `Invalid location coordinates.`

```graphql
mutation CreateRecordWithLocation {
  createRecord(
    input: {
      title: "Site Visit"
      todoListId: "list_123"
      customFields: [{ customFieldId: "field_123", value: "40.7128,-74.0060" }]
    }
  ) {
    id
    title
    customFields {
      name
      type
      latitude
      longitude
    }
  }
}
```

```json
{
  "data": {
    "createRecord": {
      "id": "clm4n8qwx000008l0g4oxdqn7",
      "title": "Site Visit",
      "customFields": [
        {
          "name": "Meeting Location",
          "type": "LOCATION",
          "latitude": 40.7128,
          "longitude": -74.006
        }
      ]
    }
  }
}
```

## Read a value

Location values are read on the `customFields` connection of a record via `recordQueries.todos`, which returns `{ items, pageInfo }`. Coordinate data lives directly on each `CustomField` element — there is no wrapper type. Select `latitude` and `longitude`.

```graphql
query ReadLocation {
  recordQueries {
    todos(filter: { companyIds: ["company_123"], todoIds: ["todo_123"] }) {
      items {
        id
        title
        customFields {
          name
          type
          latitude
          longitude
        }
      }
    }
  }
}
```

### CustomField return fields (location)

| Field       | Type               | Description                                     |
| ----------- | ------------------ | ----------------------------------------------- |
| `name`      | `String!`          | The field's display name.                       |
| `type`      | `CustomFieldType!` | `LOCATION` for this field.                      |
| `latitude`  | `Float`            | North/south position in decimal degrees.        |
| `longitude` | `Float`            | East/west position in decimal degrees.          |
| `value`     | `JSON`             | Context-aware JSON representation of the value. |

## Notes

- Coordinates are decimal degrees, not degrees/minutes/seconds. Six decimal places resolve to roughly 10 cm.
- Only `createRecord` enforces the coordinate ranges (`-90`–`90` latitude, `-180`–`180` longitude). `setRecordCustomField` does not — validate before calling it if you need range checks when updating a record.
- There is no built-in geocoding. To turn an address into coordinates, use an external service (for example Google Maps Geocoding, Mapbox, or OpenStreetMap Nominatim), then write the resulting latitude/longitude with `setRecordCustomField`.

## Errors

| Code                             | When                                                                                        |
| -------------------------------- | ------------------------------------------------------------------------------------------- |
| `CUSTOM_FIELD_VALUE_PARSE_ERROR` | `createRecord` receives a `value` whose latitude or longitude is non-numeric or out of range. |

```json
{
  "errors": [
    {
      "message": "Invalid location coordinates.",
      "extensions": { "code": "CUSTOM_FIELD_VALUE_PARSE_ERROR" }
    }
  ]
}
```

## Related

- [Create a Custom Field](/api/custom-fields/create-custom-fields)
- [Set Custom Field Values](/api/custom-fields/custom-field-values)
- [Country Custom Field](/api/custom-fields/country)
- [Number Custom Field](/api/custom-fields/number)
