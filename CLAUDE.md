# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go module for building demo projects using the Blue GraphQL API. It provides a unified CLI with multiple commands for managing projects, lists, records, tags, and custom fields.

## Development Commands

### Running Commands
All commands follow this pattern:
```bash
go run . <command> [flags]
```

**Note**: Scripts that require a project context can accept either a Project ID or Project slug. The system automatically detects which type is provided.

### Available Scripts & Usage
```bash
# READ operations - List/view data
go run . read-projects -simple
go run . read-projects -search "CRM" -sort updatedAt_DESC
go run . read-lists -project PROJECT_ID_OR_SLUG -simple
go run . read-record -record RECORD_ID -project PROJECT_ID -simple
go run . read-records -project PROJECT_ID -done false -assignee USER_ID -simple
go run . read-records -project PROJECT_ID -custom-field "field_id:GT:1000" -stats
go run . read-records -project PROJECT_ID -custom-field "field_id:CONTAINS:urgent" -limit 10
go run . read-project-records -project PROJECT_ID
go run . read-list-records -list LIST_ID -simple
go run . read-tags -project PROJECT_ID
go run . read-project-custom-fields -project PROJECT_ID
go run . list-custom-fields -project PROJECT_ID -simple
go run . list-custom-fields -project PROJECT_ID -examples
go run . list-custom-fields -project PROJECT_ID -format json
go run . read-records-count -project PROJECT_ID
go run . read-records-count -project PROJECT_ID -done false

# CREATE operations - Add new data
go run . create-project -name "Demo" -color blue -icon rocket -category ENGINEERING
go run . create-list -project PROJECT_ID -names "To Do,In Progress,Done"
go run . create-tags -project PROJECT_ID -title "Bug" -color "red"
go run . create-custom-field -name "Priority" -type "SELECT_SINGLE" -options "High:red,Medium:yellow,Low:green"
go run . create-custom-field -name "Story Points" -type "NUMBER" -min 1 -max 13
go run . create-custom-field -name "Cost" -type "CURRENCY" -currency "USD"
go run . add-custom-field-options -field FIELD_ID -options "High:red,Medium:yellow,Low:green"
go run . create-record -list LIST_ID -title "Task Name" -description "Description" -simple
go run . create-record -list LIST_ID -title "Task" -custom-fields "cf123:option_id_123,;cf456:42"
go run . create-comment -record RECORD_ID -text "This is a comment" -project PROJECT_ID -simple
go run . create-record-tags -record RECORD_ID -tag-ids "tag1,tag2" -simple

# UPDATE operations - Modify existing data
go run . update-project -project PROJECT_ID -name "New Name" -features "Chat:true,Files:false"
go run . update-record -record RECORD_ID -title "New Title" -description "Updated description"
go run . update-record -record RECORD_ID -move-to-list LIST_ID -assignees "user1,user2"
go run . update-record -record RECORD_ID -custom-fields "cf123:Updated Value;cf456:42"
go run . update-comment -comment COMMENT_ID -text "Updated comment text" -project PROJECT_ID -simple
go run . edit-custom-field -field FIELD_ID -project PROJECT_ID -name "New Field Name" -description "Updated description"
go run . edit-list -list LIST_ID -title "New List Name" -position 1000.0 -locked true

# DELETE operations - Remove data
go run . delete-project -project PROJECT_ID -confirm
go run . delete-record -record RECORD_ID -confirm
go run . delete-custom-field -field FIELD_ID -project PROJECT_ID -confirm
go run . delete-custom-field-options -field FIELD_ID -option-ids "id1,id2" -confirm
go run . delete-list -project PROJECT_ID -list LIST_ID -confirm
```

## Key Command Details

### Project Listing (`read-projects`) - ENHANCED
```bash
go run . read-projects -simple
go run . read-projects -search "CRM" -sort updatedAt_DESC -page 1 -size 10
go run . read-projects -all -archived -templates
```
**Options**: `-simple`, `-search`, `-sort` (name_ASC/DESC, createdAt_ASC/DESC, updatedAt_ASC/DESC, position_ASC/DESC), `-page`, `-size`, `-all`, `-archived`, `-templates`

### Advanced Record Querying (`read-records`) - ENHANCED
```bash
go run . read-records -project PROJECT_ID -custom-field "cf123:GT:50000" -calc -simple
go run . read-records -project PROJECT_ID -custom-field "cf456:CONTAINS:urgent" -stats
```
**Standard Options**: `-project` (required), `-list`, `-assignee`, `-tags`, `-done`, `-archived`, `-order`, `-limit`, `-skip`, `-simple`

**Custom Field Filtering (CLIENT-SIDE)**:
- Format: `-custom-field "field_id:operator:value"`
- Operators: `EQ`, `NE`, `GT`, `GTE`, `LT`, `LTE`, `IN`, `NIN`, `CONTAINS`, `IS`, `NOT`
- Examples: `"cf123:GT:50000"`, `"cf456:CONTAINS:urgent"`, `"cf789:EQ:high"`

**Numerical Statistics**:
- `-calc`: Auto-detect and calculate stats for all numerical fields
- `-stats`: Show numerical statistics (sum, average, min, max)
- `-calc-fields`: Specify custom field IDs to calculate

### Custom Fields Reference (`list-custom-fields`)
```bash
go run . list-custom-fields -project PROJECT_ID -simple
go run . list-custom-fields -project PROJECT_ID -examples -format json
```
**Options**: `-simple`, `-examples`, `-format` (table, json, csv), `-page`, `-size`

### Creating Records with Custom Fields
```bash
go run . create-record -project PROJECT_ID -list LIST_ID -title "Task" -custom-fields "cf123:option_id_123,;cf456:42"
```

**⚠️ CRITICAL: SELECT Field Values**
For SELECT_SINGLE and SELECT_MULTI fields, MUST use Option IDs with trailing comma:
- ✅ Correct: `"cf123:option_id_123,"` (trailing comma required)
- ❌ Wrong: `"cf123:High"` or `"cf123:option_id_123"` (no comma)

**Custom Field Format Examples**:
- Text: `"cf123:Hello World"`
- Number: `"cf456:42.5"`
- Boolean: `"cf789:true"`
- SELECT (single): `"cf123:option_id_123,"`
- SELECT (multi): `"cf123:option_id_1,option_id_2,"`
- Multiple: `"cf123:option_id_123,;cf456:42;cf789:true"`

**Get Option IDs**:
1. `go run . read-project-custom-fields -project PROJECT_ID` - Shows: `Title [option_id] (color)`
2. `go run . list-custom-fields -project PROJECT_ID -examples`

### Comments
```bash
go run . create-comment -record RECORD_ID -text "Progress update" -html "<p><strong>Update</strong></p>"
go run . update-comment -comment COMMENT_ID -text "Updated text" -html "<p><em>Updated</em></p>"
```

### Dependencies
```bash
go mod tidy  # Install/update dependencies
```

## Architecture

### Project Structure
- `main.go` - Single entry point with command router
- `tools/` - All command implementations  
- `common/` - Shared code (authentication, types, utilities)
- `test/` - End-to-end test suite

### Authentication (`common/auth.go`)
- `Client` struct with GraphQL request method
- Environment variables from `.env` file
- Project context via `X-Bloo-Project-Id` header
- 30-second timeout for requests

### Required Environment Variables
```
API_URL=https://api.blue.cc/graphql
AUTH_TOKEN=your_personal_access_token
CLIENT_ID=your_client_id
COMPANY_ID=your_company_slug
```

## Testing

### End-to-End Test
```bash
go run . e2e
```
- Tests all CRUD operations
- 25+ test cases covering major functionality
- Automatic cleanup
- Exit code 0 for success, 1 for failure

## Implementation Status

### Completed Features ✅
- Enhanced project listing with sorting, pagination, search
- Create/update/delete projects, lists, records, tags, custom fields
- Client-side custom field filtering with all operators
- Automatic numerical calculations and statistics
- Comprehensive record details with custom field display
- Comments creation and updates
- Project features toggle with intelligent merging
- End-to-end test suite

### Planned Features
- Custom field groups
- Automations
- Custom user roles
- Bulk operations
- Advanced export/import
- Real-time notifications

## Implementation Guidelines

1. Create new files in `tools/` directory
2. Use common package with dot imports
3. Follow command-line flag patterns
4. Add commands to `main.go` switch statement
5. Use `client.SetProjectID()` for project context
6. Include `-simple` and detailed output options
7. Handle errors consistently
8. Update CLAUDE.md with usage examples
9. Add test cases to `test/e2e.go`

## GraphQL API Details
- Endpoint: `https://api.blue.cc/graphql`
- Headers: `X-Bloo-Token-ID`, `X-Bloo-Token-Secret`, `X-Bloo-Company-ID`
- 30-second timeout, POST method with JSON body

## Usage Examples

### CRM System Setup
```bash
# Create CRM project and lists
go run . create-project -name "CRM System" -color blue -icon "office-building"
go run . create-list -project PROJECT_ID -names "Leads,Prospects,Customers,Closed Won,Closed Lost"

# Create custom fields
go run . create-custom-field -project PROJECT_ID -name "Deal Value" -type "CURRENCY" -currency "USD"
go run . create-custom-field -project PROJECT_ID -name "Priority" -type "SELECT_SINGLE" -options "High:red,Medium:yellow,Low:green"

# Add records with custom fields
go run . create-record -project PROJECT_ID -list LIST_ID -title "TechCorp Deal" -custom-fields "cf123:75000;cf456:option_id_123,"

# Query and analyze
go run . read-records -project PROJECT_ID -custom-field "cf123:GT:50000" -calc -simple
go run . read-records -project PROJECT_ID -stats -calc-fields "cf123"
```

### Data Analysis
```bash
# Project statistics
go run . read-records-count -project PROJECT_ID
go run . read-records -project PROJECT_ID -calc -simple

# Advanced filtering
go run . read-records -project PROJECT_ID -custom-field "cf456:CONTAINS:urgent" -limit 10
go run . read-projects -search "CRM" -sort updatedAt_DESC -simple
```