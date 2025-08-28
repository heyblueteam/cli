# Custom Fields Creator

This Go application allows you to create custom fields in your Blue project management system. Custom fields provide additional data points for your todos and projects beyond the standard fields.

## Features

- **24 Field Types**: Support for all available custom field types including text, numbers, dates, selects, checkboxes, and more
- **Flexible Configuration**: Extensive options for each field type
- **Validation**: Built-in validation for field types and parameters
- **Command Line Interface**: Easy-to-use CLI with comprehensive help

## Available Field Types

| Type | Description | Use Case |
|------|-------------|----------|
| `TEXT_SINGLE` | Single line text | Names, titles, short descriptions |
| `TEXT_MULTI` | Multi-line text | Detailed descriptions, notes |
| `NUMBER` | Numeric value | Quantities, scores, measurements |
| `CURRENCY` | Monetary value | Costs, budgets, prices |
| `DATE` | Date field | Due dates, start dates, deadlines |
| `SELECT_SINGLE` | Single choice dropdown | Status, priority, category |
| `SELECT_MULTI` | Multiple choice dropdown | Tags, skills, labels |
| `CHECKBOX` | Boolean field | Flags, toggles, yes/no |
| `RATING` | Rating scale | Difficulty, satisfaction, quality |
| `EMAIL` | Email address | Contact information |
| `PHONE` | Phone number | Contact information |
| `URL` | Web link | References, resources |
| `LOCATION` | Geographic location | Address, coordinates |
| `COUNTRY` | Country selection | Geographic data |
| `FILE` | File attachment | Documents, images |
| `PERCENT` | Percentage value | Progress, completion rates |
| `UNIQUE_ID` | Auto-generated ID | Reference numbers, codes |
| `FORMULA` | Calculated field | Computed values |
| `REFERENCE` | Link to other items | Related todos, projects |
| `LOOKUP` | Advanced lookup | Complex relationships |
| `TIME_DURATION` | Time period | Duration, estimates |
| `BUTTON` | Action button | Workflow triggers |
| `CURRENCY_CONVERSION` | Currency conversion | Multi-currency support |

## Installation

1. Ensure you have Go installed (version 1.16 or later)
2. Clone or download this repository
3. Set up your environment variables (see Configuration section)

## Configuration

Create a `.env` file in the project root with your Blue API credentials:

```bash
API_URL=https://api.blue.com/graphql
AUTH_TOKEN=your_auth_token_here
CLIENT_ID=your_client_id_here
COMPANY_ID=your_company_id_here
```

## Usage

### Basic Syntax

```bash
go run create-custom-field.go auth.go -name "Field Name" -type "FIELD_TYPE"
```

### Required Parameters

- `-name`: The name of the custom field (required)
- `-type`: The type of custom field (required)

### Common Examples

#### 1. Simple Text Field
```bash
go run create-custom-field.go auth.go \
  -name "Notes" \
  -type "TEXT_MULTI" \
  -description "Additional notes for the task"
```

#### 2. Number Field with Constraints
```bash
go run create-custom-field.go auth.go \
  -name "Story Points" \
  -type "NUMBER" \
  -description "Agile story points (1-13)" \
  -min 1 \
  -max 13
```

#### 3. Currency Field
```bash
go run create-custom-field.go auth.go \
  -name "Cost" \
  -type "CURRENCY" \
  -description "Task cost in USD" \
  -currency "USD" \
  -prefix "$"
```

#### 4. Select Field
```bash
go run create-custom-field.go auth.go \
  -name "Priority" \
  -type "SELECT_SINGLE" \
  -description "Task priority level"
```

#### 5. Date Field
```bash
go run create-custom-field.go auth.go \
  -name "Start Date" \
  -type "DATE" \
  -description "When the task should start" \
  -is-due-date
```

#### 6. Unique ID Field
```bash
go run create-custom-field.go auth.go \
  -name "Task ID" \
  -type "UNIQUE_ID" \
  -description "Unique identifier for the task" \
  -use-sequence \
  -sequence-digits 8 \
  -sequence-start 1000
```

### Advanced Options

#### Time Duration Fields
```bash
go run create-custom-field.go auth.go \
  -name "Estimated Time" \
  -type "TIME_DURATION" \
  -description "Estimated time to complete the task" \
  -time-duration-display "HOURS" \
  -time-duration-target 8.0
```

#### Reference Fields
```bash
go run create-custom-field.go auth.go \
  -name "Related Project" \
  -type "REFERENCE" \
  -description "Reference to related project" \
  -reference-project "project-id-here" \
  -reference-multiple
```

#### Button Fields
```bash
go run create-custom-field.go auth.go \
  -name "Approve" \
  -type "BUTTON" \
  -description "Approve the task" \
  -button-type "APPROVE" \
  -button-confirm-text "Are you sure you want to approve this task?"
```

## All Available Flags

| Flag | Description | Default |
|------|-------------|---------|
| `-name` | Custom field name (required) | - |
| `-type` | Custom field type (required) | - |
| `-description` | Field description | - |
| `-button-type` | Button type for BUTTON fields | - |
| `-button-confirm-text` | Button confirmation text | - |
| `-currency-field-id` | Currency field ID for conversion | - |
| `-conversion-date` | Conversion date | - |
| `-conversion-date-type` | Conversion date type | - |
| `-min` | Minimum value for NUMBER fields | 0 |
| `-max` | Maximum value for NUMBER fields | 0 |
| `-currency` | Currency code | USD |
| `-prefix` | Field prefix | - |
| `-is-due-date` | Whether field represents due date | false |
| `-time-duration-display` | Time duration display type | - |
| `-time-duration-target` | Time duration target time | 0 |
| `-reference-project` | Reference project ID | - |
| `-reference-multiple` | Allow multiple references | false |
| `-use-sequence` | Use sequence unique ID | false |
| `-sequence-digits` | Number of digits in sequence | 6 |
| `-sequence-start` | Starting number for sequence | 1 |
| `-list` | List available options | false |

## Getting Help

### List Available Options
```bash
go run create-custom-field.go auth.go -list
```

This will show:
- All available custom field types
- Supported currencies
- Time duration types
- Time duration conditions

### Show Help
```bash
go run create-custom-field.go auth.go -h
```

## Examples Script

Run the included examples script to create multiple custom fields at once:

```bash
./examples/create-custom-fields.sh
```

**Note**: Make sure to update the project ID in the script before running.

## After Creating Custom Fields

1. **For SELECT fields**: You'll need to create custom field options separately using the `createCustomFieldOption` mutation
2. **For REFERENCE fields**: Ensure the referenced project exists and you have access to it
3. **For LOOKUP fields**: Set up the appropriate lookup configuration
4. **For FORMULA fields**: Define the calculation formula in the metadata

## Troubleshooting

### Common Issues

1. **"You are not authorized"**: Check your `.env` file and ensure credentials are correct
2. **"Invalid field type"**: Use `-list` flag to see valid field types
3. **"Missing required environment variables"**: Ensure your `.env` file is properly configured

### Validation

The application validates:
- Required field names and types
- Field type compatibility with provided options
- Numeric constraints (min/max values)
- Reference project existence

## API Reference

This tool uses the Blue GraphQL API with the `createCustomField` mutation. The full schema can be found in `schema.graphql`.

## Contributing

Feel free to extend this tool with additional features:
- Custom field option creation
- Field editing capabilities
- Bulk field creation
- Field deletion
- Field listing and search

## License

This tool is part of the Blue demo builder project.
