---
title: Phone Custom Field
description: Create phone fields to store and validate phone numbers with international formatting
category: Custom Fields
icon: Phone
---

Phone custom fields allow you to store phone numbers in records with built-in validation and international formatting. They're ideal for tracking contact information, emergency contacts, or any phone-related data in your projects.

## Basic Example

Create a simple phone field:

```graphql
mutation CreatePhoneField {
  createCustomField(input: { name: "Contact Phone", type: PHONE }) {
    id
    name
    type
  }
}
```

## Advanced Example

Create a phone field with description:

```graphql
mutation CreateDetailedPhoneField {
  createCustomField(
    input: {
      name: "Emergency Contact"
      type: PHONE
      description: "Emergency contact number with country code"
    }
  ) {
    id
    name
    type
    description
  }
}
```

## Input Parameters

### CreateCustomFieldInput

| Parameter     | Type             | Required | Description                     |
| ------------- | ---------------- | -------- | ------------------------------- |
| `name`        | String!          | ✅ Yes   | Display name of the phone field |
| `type`        | CustomFieldType! | ✅ Yes   | Must be `PHONE`                 |
| `description` | String           | No       | Help text shown to users        |

**Note**: Custom fields are automatically associated with the project based on the user's current project context. No `projectId` parameter is required.

## Setting Phone Values

To set or update a phone value on a record:

```graphql
mutation SetPhoneValue {
  setTodoCustomField(
    input: { todoId: "todo_123", customFieldId: "field_456", text: "+1 234 567 8900" }
  )
}
```

### SetTodoCustomFieldInput Parameters

| Parameter       | Type    | Required | Description                           |
| --------------- | ------- | -------- | ------------------------------------- |
| `todoId`        | String! | ✅ Yes   | ID of the record to update            |
| `customFieldId` | String! | ✅ Yes   | ID of the phone custom field          |
| `text`          | String  | No       | Phone number with country code        |
| `regionCode`    | String  | No       | Country code (automatically detected) |

**Note**: While `text` is optional in the schema, a phone number is required for the field to be meaningful. When using `setTodoCustomField`, no validation is performed - you can store any text value and regionCode. The automatic detection only happens during record creation.

## Creating Records with Phone Values

When creating a new record with phone values:

```graphql
mutation CreateRecordWithPhone {
  createTodo(
    input: {
      title: "Call client"
      todoListId: "list_123"
      customFields: [{ customFieldId: "phone_field_id", value: "+1-555-123-4567" }]
    }
  ) {
    id
    title
    customFields {
      id
      name
      type
      text
      regionCode
    }
  }
}
```

## Response Fields

### Returned fields (`CustomField` in record context)

Each element of a record's `customFields` is a `CustomField` carrying the value for that record:

| Field        | Type             | Description                                       |
| ------------ | ---------------- | ------------------------------------------------- |
| `id`         | ID!              | Unique identifier for the field value             |
| `name`       | String           | The custom field's display name                   |
| `type`       | CustomFieldType! | The field type (`PHONE`)                          |
| `text`       | String           | The formatted phone number (international format) |
| `regionCode` | String           | The country code (e.g., "US", "GB", "CA")         |
| `todo`       | Todo!            | The record this value belongs to                  |
| `createdAt`  | DateTime!        | When the value was created                        |
| `updatedAt`  | DateTime!        | When the value was last modified                  |

## Phone Number Validation

**Important**: Phone number validation and formatting only occurs when creating new records via `createTodo`. When updating existing phone values using `setTodoCustomField`, no validation is performed and the values are stored as provided.

### Accepted Formats (During Record Creation)

Phone numbers must include a country code in one of these formats:

- **E.164 format (preferred)**: `+12345678900`
- **International format**: `+1 234 567 8900`
- **International with punctuation**: `+1 (234) 567-8900`
- **Country code with dashes**: `+1-234-567-8900`

**Note**: National formats without country code (like `(234) 567-8900`) will be rejected during record creation.

### Validation Rules (During Record Creation)

- Uses libphonenumber-js for parsing and validation
- Accepts various international phone number formats
- Automatically detects country from the number
- Formats number in international display format (e.g., `+1 234 567 8900`)
- Extracts and stores country code separately (e.g., `US`)

### Valid Phone Examples

```
+12345678900           # E.164 format
+1 234 567 8900        # International format
+1 (234) 567-8900      # With parentheses
+1-234-567-8900        # With dashes
+44 20 7946 0958       # UK number
+33 1 42 86 83 26      # French number
```

### Invalid Phone Examples

```
(234) 567-8900         # Missing country code
234-567-8900           # Missing country code
123                    # Too short
invalid-phone          # Not a number
+1 234                 # Incomplete number
```

## Storage Format

When creating records with phone numbers:

- **text**: Stored in international format (e.g., `+1 234 567 8900`) after validation
- **regionCode**: Stored as ISO country code (e.g., `US`, `GB`, `CA`) automatically detected

When updating via `setTodoCustomField`:

- **text**: Stored exactly as provided (no formatting)
- **regionCode**: Stored exactly as provided (no validation)

## Required Permissions

| Action             | Required Permission                      |
| ------------------ | ---------------------------------------- |
| Create phone field | `OWNER` or `ADMIN` role at project level |
| Update phone field | `OWNER` or `ADMIN` role at project level |
| Set phone value    | Standard record edit permissions         |
| View phone value   | Standard record view permissions         |

## Error Responses

### Invalid Phone Format

```json
{
  "errors": [
    {
      "message": "Invalid phone number format.",
      "extensions": {
        "code": "CUSTOM_FIELD_VALUE_PARSE_ERROR"
      }
    }
  ]
}
```

### Field Not Found

```json
{
  "errors": [
    {
      "message": "Custom field not found",
      "extensions": {
        "code": "CUSTOM_FIELD_NOT_FOUND"
      }
    }
  ]
}
```

### Missing Country Code

```json
{
  "errors": [
    {
      "message": "Invalid phone number format.",
      "extensions": {
        "code": "CUSTOM_FIELD_VALUE_PARSE_ERROR"
      }
    }
  ]
}
```

## Best Practices

### Data Entry

- Always include country code in phone numbers
- Use E.164 format for consistency
- Validate numbers before storing for important operations
- Consider regional preferences for display formatting

### Data Quality

- Store numbers in international format for global compatibility
- Use regionCode for country-specific features
- Validate phone numbers before critical operations (SMS, calls)
- Consider time zone implications for contact timing

### International Considerations

- Country code is automatically detected and stored
- Numbers are formatted in international standard
- Regional display preferences can use regionCode
- Consider local dialing conventions when displaying

## Common Use Cases

1. **Contact Management**
   - Client phone numbers
   - Vendor contact information
   - Team member phone numbers
   - Support contact details

2. **Emergency Contacts**
   - Emergency contact numbers
   - On-call contact information
   - Crisis response contacts
   - Escalation phone numbers

3. **Customer Support**
   - Customer phone numbers
   - Support callback numbers
   - Verification phone numbers
   - Follow-up contact numbers

4. **Sales & Marketing**
   - Lead phone numbers
   - Campaign contact lists
   - Partner contact information
   - Referral source phones

## Integration Features

### With Automations

- Trigger actions when phone fields are updated
- Send SMS notifications to stored phone numbers
- Create follow-up tasks based on phone changes
- Route calls based on phone number data

### With Lookups

- Reference phone data from other records
- Aggregate phone lists from multiple sources
- Find records by phone number
- Cross-reference contact information

### With Forms

- Automatic phone validation
- International format checking
- Country code detection
- Real-time format feedback

## Limitations

- Requires country code for all numbers
- No built-in SMS or calling capabilities
- No phone number verification beyond format checking
- No storage of phone metadata (carrier, type, etc.)
- National format numbers without country code are rejected
- No automatic phone number formatting in UI beyond international standard

## Related Resources

- [Text Fields](/api/custom-fields/text-single) - For non-phone text data
- [Email Fields](/api/custom-fields/email) - For email addresses
- [URL Fields](/api/custom-fields/url) - For website addresses
- [Custom Fields Overview](/custom-fields/list-custom-fields) - General concepts
