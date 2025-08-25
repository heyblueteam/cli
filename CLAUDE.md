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

### Available Scripts & Usage
```bash
# List projects (first 20)
go run auth.go list-projects.go -simple

# Create project with options
go run auth.go create-project.go -name "Demo" -color blue -icon rocket -category ENGINEERING

# Delete project (requires confirmation and special permissions)
go run auth.go delete-project.go -project PROJECT_ID -confirm

# Get lists in a project
go run auth.go get-lists.go -project PROJECT_ID -simple

# Create lists in a project
go run auth.go create-list.go -project PROJECT_ID -names "To Do,In Progress,Done"

# List all todos across all lists in a project (overview)
go run auth.go list-project-todos.go -project PROJECT_ID

# List todos in a specific list (detailed with filtering)
go run auth.go list-todos.go -list LIST_ID -simple

# List tags in a project
go run auth.go list-tags.go -project PROJECT_ID

# Create tags in a project
go run auth.go create-tags.go -project PROJECT_ID -title "Bug" -color "red"

# Create custom fields (all types except reference/lookup)
go run auth.go create-custom-field.go -name "Priority" -type "SELECT_SINGLE" -description "Task priority"
go run auth.go create-custom-field.go -name "Story Points" -type "NUMBER" -min 1 -max 13
go run auth.go create-custom-field.go -name "Cost" -type "CURRENCY" -currency "USD"

# Create records/todos in lists
go run auth.go create-record.go -list LIST_ID -title "Task Name" -description "Description" -simple

# List/query records across projects with advanced filtering
go run auth.go list-records.go -project PROJECT_ID -done false -assignee USER_ID -simple

# Count records/todos in a project with optional filtering
go run auth.go count-records.go -project PROJECT_ID
go run auth.go count-records.go -project PROJECT_ID -done false
go run auth.go count-records.go -project PROJECT_ID -list LIST_ID -archived false

# Delete records/todos (requires confirmation for safety)
go run auth.go delete-record.go -record RECORD_ID -confirm

# Add tags to existing records/todos
go run auth.go add-tags-to-record.go -record RECORD_ID -tag-ids "tag1,tag2" -simple
go run auth.go add-tags-to-record.go -record RECORD_ID -tag-titles "Bug,Priority" -project PROJECT_ID

# Edit/update project settings and toggle features
go run auth.go edit-project.go -project PROJECT_ID -name "New Name" -features "Chat:true,Files:false"
go run auth.go edit-project.go -project PROJECT_ID -todo-alias "Tasks" -hide-record-count true
go run auth.go edit-project.go -project PROJECT_ID -features "Wiki:true,Docs:false" -simple

# Available feature types: Activity, Todo, Wiki, Chat, Docs, Forms, Files, People
# Features are merged with existing state (partial updates supported)
go run auth.go edit-project.go -project PROJECT_ID -features "Todo:false,People:false"
```

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
- Project context support via `X-Bloo-Project-Id` header
- 30-second timeout for requests

### GraphQL Integration Pattern
Each script:
1. Imports the auth module
2. Creates a client instance
3. Sets project context using `client.SetProjectID()` when needed
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

The `edit-project.go` script implements intelligent feature merging:
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