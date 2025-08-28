# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go module for building demo projects using the Blue GraphQL API. It consists of individual command-line utilities that share a centralized authentication module.

## Development Commands

### Running Scripts
All scripts follow this pattern:
```bash
go run auth.go <script-name>.go [flags]
```

**Note**: Scripts that require a project context can accept either a Project ID or Project slug. The system automatically detects which type is provided.

### Available Scripts & Usage
```bash
# READ operations - List/view data
# List projects (first 20)
go run auth.go read-projects.go -simple

# Get lists in a project (using Project ID or slug)
go run auth.go read-lists.go -project PROJECT_ID_OR_SLUG -simple

# List all todos across all lists in a project (overview)
go run auth.go read-project-todos.go -project PROJECT_ID

# List todos in a specific list (detailed with filtering)
go run auth.go read-todos.go -list LIST_ID -simple

# List tags in a project
go run auth.go read-tags.go -project PROJECT_ID

# List custom fields in a project
go run auth.go read-project-custom-fields.go -project PROJECT_ID

# List/query records across projects with advanced filtering
go run auth.go read-records.go -project PROJECT_ID -done false -assignee USER_ID -simple

# Count records/todos in a project with optional filtering
go run auth.go read-records-count.go -project PROJECT_ID
go run auth.go read-records-count.go -project PROJECT_ID -done false
go run auth.go read-records-count.go -project PROJECT_ID -list LIST_ID -archived false

# CREATE operations - Add new data
# Create project with options
go run auth.go create-project.go -name "Demo" -color blue -icon rocket -category ENGINEERING

# Create lists in a project
go run auth.go create-list.go -project PROJECT_ID -names "To Do,In Progress,Done"

# Create tags in a project
go run auth.go create-tags.go -project PROJECT_ID -title "Bug" -color "red"

# Create custom fields (all types except reference/lookup)
go run auth.go create-custom-field.go -name "Priority" -type "SELECT_SINGLE" -description "Task priority" -options "High:red,Medium:yellow,Low:green"
go run auth.go create-custom-field.go -name "Status" -type "SELECT_MULTI" -options "In Progress,Blocked:red,Review Required:blue"
go run auth.go create-custom-field.go -name "Story Points" -type "NUMBER" -min 1 -max 13
go run auth.go create-custom-field.go -name "Cost" -type "CURRENCY" -currency "USD"

# Create records/todos in lists (supports custom fields)
go run auth.go create-record.go -list LIST_ID -title "Task Name" -description "Description" -simple

# Create records/todos with custom field values
go run auth.go create-record.go -list LIST_ID -title "Task" -custom-fields "cf123:Priority High;cf456:42"

# Add tags to existing records/todos
go run auth.go create-record-tags.go -record RECORD_ID -tag-ids "tag1,tag2" -simple
go run auth.go create-record-tags.go -record RECORD_ID -tag-titles "Bug,Priority" -project PROJECT_ID

# UPDATE operations - Modify existing data
# Edit/update project settings and toggle features
go run auth.go update-project.go -project PROJECT_ID -name "New Name" -features "Chat:true,Files:false"
go run auth.go update-project.go -project PROJECT_ID -todo-alias "Tasks" -hide-record-count true
go run auth.go update-project.go -project PROJECT_ID -features "Wiki:true,Docs:false" -simple

# Available feature types: Activity, Todo, Wiki, Chat, Docs, Forms, Files, People
# Features are merged with existing state (partial updates supported)
go run auth.go update-project.go -project PROJECT_ID -features "Todo:false,People:false"

# DELETE operations - Remove data
# Delete project (requires confirmation and special permissions)
go run auth.go delete-project.go -project PROJECT_ID -confirm

# Delete records/todos (requires confirmation for safety)
go run auth.go delete-record.go -record RECORD_ID -confirm
```

### Detailed Script Documentation

#### Create Record (`create-record.go`)
Creates new records/todos in lists with support for custom field values, assignments, and placement options.

```bash
# Create a simple record
go run auth.go create-record.go -project PROJECT_ID -list LIST_ID -title "Task Name"

# Create record with description and placement
go run auth.go create-record.go -project PROJECT_ID -list LIST_ID -title "Task Name" -description "Task description" -placement TOP

# Create record with custom field values
go run auth.go create-record.go -project PROJECT_ID -list LIST_ID -title "Task Name" -custom-fields "cf123:High Priority;cf456:42.5"

# Create record with assignees and custom fields
go run auth.go create-record.go -project PROJECT_ID -list LIST_ID -title "Task Name" -assignees "user1,user2" -custom-fields "cf789:true"
```

**Options:**
- `-project` (required): Project ID or Project slug
- `-list` (required): List ID to create the record in
- `-title` (required): Title of the record
- `-description`: Description of the record (optional)
- `-placement`: Placement in list - TOP or BOTTOM (optional)
- `-assignees`: Comma-separated assignee IDs (optional)
- `-custom-fields`: Custom field values in format "field_id1:value1;field_id2:value2" (optional)
- `-simple`: Simple output format (optional)

#### Create Custom Field (`create-custom-field.go`)
Creates custom fields for projects with support for all field types including SELECT fields with options.

```bash
# Create SELECT_SINGLE field with options and colors
go run auth.go create-custom-field.go -project PROJECT_ID -name "Priority" -type "SELECT_SINGLE" -options "High:red,Medium:yellow,Low:green"

# Create SELECT_MULTI field with options (some with colors, some without)
go run auth.go create-custom-field.go -project PROJECT_ID -name "Labels" -type "SELECT_MULTI" -options "Bug:red,Feature,Enhancement:blue"

# Create other field types
go run auth.go create-custom-field.go -project PROJECT_ID -name "Story Points" -type "NUMBER" -min 1 -max 13
go run auth.go create-custom-field.go -project PROJECT_ID -name "Budget" -type "CURRENCY" -currency "USD"
```

**Options:**
- `-project` (required): Project ID or Project slug
- `-name` (required): Custom field name
- `-type` (required): Custom field type (use -list to see all available types)
- `-description`: Custom field description (optional)
- `-options`: Options for SELECT fields in format "value1:color1,value2:color2" (optional)
  - Format: Comma-separated values, optionally with colors after colon
  - Examples: "High,Medium,Low" or "High:red,Medium:yellow,Low:green"
  - Colors can be omitted for some options: "In Progress,Blocked:red,Complete"
- `-min`: Minimum value for NUMBER fields (optional)
- `-max`: Maximum value for NUMBER fields (optional)
- `-currency`: Currency code for CURRENCY fields (default: USD)
- `-list`: List all available field types and other options

**Custom Fields Format Examples:**
- Text field: `"cf123:Hello World"`
- Number field: `"cf456:42.5"`
- Boolean field: `"cf789:true"`
- Multiple fields: `"cf123:Hello;cf456:42;cf789:true"`

#### Count Records (`read-records-count.go`)
Counts the total number of records/todos in a project with optional filtering.

```bash
# Count all records in a project
go run auth.go read-records-count.go -project PROJECT_ID

# Count only incomplete records
go run auth.go read-records-count.go -project PROJECT_ID -done false

# Count records in a specific list
go run auth.go read-records-count.go -project PROJECT_ID -list LIST_ID

# Count non-archived records
go run auth.go read-records-count.go -project PROJECT_ID -archived false
```

**Options:**
- `-project` (required): Project ID or Project slug to count records
- `-list`: Todo List ID to filter records (optional)
- `-done`: Filter by completion status (true/false, optional)
- `-archived`: Filter by archived status (true/false, optional)

### Dependencies
```bash
go mod tidy  # Install/update dependencies
```

## Architecture

### Centralized Authentication (`auth.go`)
All scripts import and use the shared authentication module which provides:
- `Client` struct with GraphQL request method
- Environment variable loading from `.env`
- Standard HTTP headers for Blue API authentication
- Project context support via `X-Bloo-Project-Id` header (accepts both Project ID and Project slug)
- Automatic detection of Project ID vs Project slug format
- 30-second timeout for requests

### GraphQL Integration Pattern
Each script:
1. Imports the auth module
2. Creates a client instance
3. Sets project context using `client.SetProjectID()`, `client.SetProjectSlug()`, or `client.SetProject()` when needed
4. Defines GraphQL query/mutation as a string
5. Makes requests using `client.ExecuteQueryWithResult()`
6. Unmarshals JSON response into typed structs

### Required Environment Variables
The `.env` file must contain:
```
API_URL=https://api.blue.cc/graphql
AUTH_TOKEN=your_personal_access_token
CLIENT_ID=your_client_id
COMPANY_ID=your_company_slug
```

## Testing

### End-to-End Test (`e2e.go`)
Comprehensive test suite that validates all 17 tool files:

```bash
# Run the end-to-end test
go run auth.go e2e.go
```

**Coverage:**
- Tests all CRUD operations (Create, Read, Update, Delete)
- Validates project → lists → tags → custom fields → records workflow
- Uses actual tool execution (calls the real Go files)
- Automatic cleanup (deletes test project)
- 22 test cases covering all major functionality

**Output:**
- Emoji-friendly status indicators (✅/❌)
- Detailed progress reporting
- Summary with pass/fail counts
- Exit code 0 for success, 1 for failure (CI/CD compatible)

## Implemented Features

Completed:
- ✅ List projects with pagination and search
- ✅ Create projects with customization options
- ✅ Delete projects (with safety confirmation)
- ✅ List and create todo lists in projects
- ✅ List todos with filtering and pagination
- ✅ List and create tags in projects
- ✅ List custom fields in projects
- ✅ Create custom fields (24+ types including reference/lookup)
- ✅ Create records/todos (simple: name + list + description)
- ✅ Advanced record querying with filtering and sorting
- ✅ Count records/todos in projects with filtering options
- ✅ Delete records/todos with safety confirmation
- ✅ Add tags to records/todos (by tag ID or title)
- ✅ Edit/update project settings and toggle features (with intelligent feature merging)
- ✅ End-to-end test suite with full coverage

## Planned Features

To implement:
- Create custom field groups
- Create automations
- Create custom user roles
- Create record (full: name + list + custom field values)

## Implementation Guidelines

When adding new scripts:
1. Use the centralized `auth.go` module for all API calls
2. Follow the existing command-line flag patterns using Go's `flag` package
3. Use `client.SetProjectID()` for operations that require project context
4. Include both `-simple` and detailed output options where applicable
5. Define proper struct types for GraphQL responses
6. Handle errors consistently with proper context
7. For operations that modify arrays/lists, implement proper merging logic to preserve existing data
8. Update CLAUDE.md with usage examples

### Feature Toggle Implementation Notes

The `update-project.go` script implements intelligent feature merging:
- Fetches current project state before making changes
- Merges user-specified feature toggles with existing features
- Sends complete feature array to prevent data loss
- Supports 8 feature types: Activity, Todo, Wiki, Chat, Docs, Forms, Files, People
- All features default to enabled=true for new projects

## Known Limitations

- Project listing limited to first 20 results (pagination not implemented)
- Maximum 50 lists per project
- Project deletion requires special permissions (may fail with authorization error)
- No test suite or linting configuration
- Individual script execution (no unified CLI)

## GraphQL API Details

- Endpoint: `https://api.blue.cc/graphql`
- Authentication Headers:
  - `X-Bloo-Token-ID`: Client ID
  - `X-Bloo-Token-Secret`: Auth Token
  - `X-Bloo-Company-ID`: Company slug
- Request timeout: 30 seconds
- All requests use POST method with JSON body