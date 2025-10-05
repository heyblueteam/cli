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
go run . read-custom-fields -project PROJECT_ID -simple
go run . read-custom-fields -project PROJECT_ID -examples
go run . read-custom-fields -project PROJECT_ID -format json
go run . read-records-count -project PROJECT_ID
go run . read-records-count -project PROJECT_ID -done false
go run . read-automations -project PROJECT_ID
go run . read-automations -project PROJECT_ID -simple
go run . read-automations -project PROJECT_ID -page 2 -size 10
go run . read-automations -project PROJECT_ID -skip 20 -limit 5
go run . read-user-profiles -simple
go run . read-user-profiles -search "john" -page 1 -size 10
go run . read-project-user-roles -project PROJECT_ID -simple
go run . read-project-user-roles -projects "PROJECT_ID1,PROJECT_ID2" -format json

# CREATE operations - Add new data
go run . create-project -name "Demo" -color blue -icon rocket -category ENGINEERING
go run . create-list -project PROJECT_ID -names "To Do,In Progress,Done"
go run . create-tags -project PROJECT_ID -title "Bug" -color "red"
go run . create-custom-field -name "Priority" -type "SELECT_SINGLE" -options "High:red,Medium:yellow,Low:green"
go run . create-custom-field -name "Story Points" -type "NUMBER" -min 1 -max 13
go run . create-custom-field -name "Cost" -type "CURRENCY" -currency "USD"
go run . create-custom-field-options -field FIELD_ID -options "High:red,Medium:yellow,Low:green"
go run . create-record -list LIST_ID -title "Task Name" -description "Description" -simple
go run . create-record -list LIST_ID -title "Task" -custom-fields "cf123:option_id_123,;cf456:42"
go run . create-comment -record RECORD_ID -text "This is a comment" -project PROJECT_ID -simple
go run . create-record-tags -record RECORD_ID -tag-ids "tag1,tag2" -simple
go run . create-automation -project PROJECT_ID -trigger-type "TODO_MARKED_AS_COMPLETE" -action-type "SEND_EMAIL" -email-to "user@example.com"
go run . create-automation -project PROJECT_ID -trigger-type "TAG_ADDED" -trigger-tags "TAG_ID" -action-type "MAKE_HTTP_REQUEST" -http-url "https://example.com/webhook"
go run . create-automation-multi -project PROJECT_ID -trigger-type "TAG_ADDED" -action1-type "SEND_EMAIL" -action1-email-to "manager@company.com" -action2-type "ADD_COLOR" -action2-color "#ff0000"
go run . invite-user -email "user@example.com" -access-level "MEMBER" -project PROJECT_ID
go run . invite-user -email "admin@example.com" -access-level "ADMIN" -projects "PROJECT_ID1,PROJECT_ID2"

# UPDATE operations - Modify existing data
go run . update-project -project PROJECT_ID -name "New Name" -features "Chat:true,Files:false"
go run . update-record -record RECORD_ID -title "New Title" -description "Updated description"
go run . update-record -record RECORD_ID -move-to-list LIST_ID -assignees "user1,user2"
go run . update-record -record RECORD_ID -custom-fields "cf123:Updated Value;cf456:42"
# For SELECT fields, MUST use option IDs with trailing comma:
go run . update-record -record RECORD_ID -custom-fields "cf123:option_id_123,;cf456:42"
go run . update-comment -comment COMMENT_ID -text "Updated comment text" -project PROJECT_ID -simple
go run . update-custom-field -field FIELD_ID -project PROJECT_ID -name "New Field Name" -description "Updated description"
go run . update-list -list LIST_ID -title "New List Name" -position 1000.0 -locked true
go run . update-automation -automation AUTOMATION_ID -project PROJECT_ID -active true
go run . update-automation -automation AUTOMATION_ID -project PROJECT_ID -trigger-type "TODO_MARKED_AS_COMPLETE" -action-type "SEND_EMAIL"
go run . update-automation -automation AUTOMATION_ID -project PROJECT_ID -email-to "new@example.com" -email-subject "Updated subject"
go run . update-automation-multi -automation AUTOMATION_ID -project PROJECT_ID -action1-type "SEND_EMAIL" -action1-email-to "manager@company.com" -action2-type "ADD_COLOR" -action2-color "#00ff00"
go run . move-record -record RECORD_ID -list LIST_ID

# DELETE operations - Remove data
go run . delete-project -project PROJECT_ID -confirm
go run . delete-record -record RECORD_ID -confirm
go run . delete-custom-field -field FIELD_ID -project PROJECT_ID -confirm
go run . delete-custom-field-options -field FIELD_ID -option-ids "id1,id2" -confirm
go run . delete-list -project PROJECT_ID -list LIST_ID -confirm
go run . delete-automation -automation AUTOMATION_ID -project PROJECT_ID -confirm
```

## Key Command Details

### Project Listing (`read-projects`) - ENHANCED
```bash
go run . read-projects -simple
go run . read-projects -search "CRM" -sort updatedAt_DESC -page 1 -size 10
go run . read-projects -all -archived -templates
```
**Options**: `-simple`, `-search`, `-sort` (name_ASC/DESC, createdAt_ASC/DESC, updatedAt_ASC/DESC, position_ASC/DESC), `-page`, `-size`, `-all`, `-archived`, `-templates`

### Automation Listing (`read-automations`) - ENHANCED
```bash
go run . read-automations -project PROJECT_ID
go run . read-automations -project PROJECT_ID -simple
go run . read-automations -project PROJECT_ID -page 2 -size 10
go run . read-automations -project PROJECT_ID -skip 20 -limit 5
```
**Options**: `-project` (required), `-simple`, `-page` (default: 1), `-size` (default: 50, max: 100), `-skip` (overrides page), `-limit` (overrides size)

**Pagination Features**:
- Page-based navigation with `-page` and `-size`
- Direct skip/limit control with `-skip` and `-limit`
- Visual pagination summary with navigation hints
- Enhanced detailed output with icons and tree structure

Shows all automations in a project with comprehensive information including:
- Automation ID, UID, and active status
- Creation and update timestamps  
- **Trigger details with full metadata:**
  - Type and associated todo list
  - Custom fields and field options  
  - Tags and assignees
  - Metadata (e.g., incomplete only flag)
  - Color settings
- **Action details with full metadata:**
  - Type, target todo list, tags, and assignees
  - **Email metadata:** from/to/cc/bcc, subject, content, attachments
  - **Checklist metadata:** checklist items with positions, due dates, assignees
  - **Copy todo metadata:** copy options
  - **HTTP webhook metadata:** URL, method, headers, parameters, auth
  - Custom fields, color settings, assignee triggerer
- Complete automation structure and response format for understanding automation behavior

### User Profiles Listing (`read-user-profiles`) - ENHANCED
```bash
# Project-specific users (from record assignments)
go run . read-user-profiles -project PROJECT_ID -simple
go run . read-user-profiles -project PROJECT_ID -stats
go run . read-user-profiles -project PROJECT_ID -search "john"

# Company-wide users (aggregated from all projects)
go run . read-user-profiles -simple
go run . read-user-profiles -search "developer" -company COMPANY_ID
go run . read-user-profiles -first 100
```
**Options**: `-project`, `-simple`, `-search`, `-first` (default: 50), `-company`

**Implementation Details**:
- **Project Mode** (`-project`): Uses `projectUserList` GraphQL endpoint for project-specific users
- **Company Mode** (no `-project`): Uses `companyUserList` GraphQL endpoint for all company users
- **Search**: Server-side search (company mode) or client-side filtering (project mode)
- **Pagination**: Control with `-first` parameter (default: 50 users)

**✅ Verified Working Endpoints**:
- `companyUserList(companyId, search, first, orderBy)` - Company-wide users with search
- `projectUserList(filter: {projectIds}, first, orderBy)` - Project-specific users  
- Fallback to `userList` with filters if primary endpoints unavailable

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

### Custom Fields Reference (`read-custom-fields`)
```bash
go run . read-custom-fields -project PROJECT_ID -simple
go run . read-custom-fields -project PROJECT_ID -examples -format json
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
2. `go run . read-custom-fields -project PROJECT_ID -examples`

**IMPORTANT**: Always use option IDs (not option titles) for SELECT fields. The read-project-custom-fields command shows the actual option IDs in brackets that you must use.

### Automations (`create-automation`) - NEW
```bash
# Simple email automation on task completion
go run . create-automation -project PROJECT_ID \
  -trigger-type "TODO_MARKED_AS_COMPLETE" \
  -action-type "SEND_EMAIL" \
  -email-to "user@example.com" \
  -email-subject "Task completed" \
  -email-content "<p>Task has been completed!</p>"

# HTTP webhook when tag is added
go run . create-automation -project PROJECT_ID \
  -trigger-type "TAG_ADDED" -trigger-tags "TAG_ID" \
  -action-type "MAKE_HTTP_REQUEST" \
  -http-url "https://example.com/webhook" \
  -http-method "POST" \
  -http-body '{"event": "tag_added"}'

# Auto-assign tag and user on task creation
go run . create-automation -project PROJECT_ID \
  -trigger-type "TODO_CREATED" -trigger-todo-list "LIST_ID" \
  -action-type "ADD_TAG" \
  -action-tags "TAG_ID"

# Set color when moving to specific list
go run . create-automation -project PROJECT_ID \
  -trigger-type "TODO_LIST_CHANGED" -trigger-todo-list "LIST_ID" \
  -action-type "ADD_COLOR" \
  -action-color "#ff6b6b"
```

**Supported Trigger Types**: `TODO_CREATED`, `TODO_MARKED_AS_COMPLETE`, `TODO_MARKED_AS_INCOMPLETE`, `TODO_LIST_CHANGED`, `TAG_ADDED`, `TAG_REMOVED`, `ASSIGNEE_ADDED`, `ASSIGNEE_REMOVED`, `CUSTOM_FIELD_CHANGED`, `TODO_OVERDUE`

**Supported Action Types**: `SEND_EMAIL`, `MAKE_HTTP_REQUEST`, `ADD_TAG`, `REMOVE_TAG`, `ADD_ASSIGNEE`, `REMOVE_ASSIGNEE`, `ADD_COLOR`, `CHANGE_TODO_LIST`, `MARK_AS_COMPLETE`, `MARK_AS_INCOMPLETE`, `CREATE_TODO`, `COPY_TODO`, `DELETE_TODO`, `ARCHIVE_TODO`, `CREATE_CHECKLIST`

**Email Options**: Use `-email-from`, `-email-to`, `-email-subject`, `-email-content` for SEND_EMAIL actions.

**HTTP Options**: Use `-http-url`, `-http-method`, `-http-content-type`, `-http-body`, `-http-headers`, `-http-params`, `-http-auth-type`, `-http-auth-value` for MAKE_HTTP_REQUEST actions.

### Multi-Action Automations (`create-automation-multi`) - NEW
```bash
# Create automation with multiple different actions
go run . create-automation-multi -project PROJECT_ID \
  -trigger-type "TAG_ADDED" \
  -trigger-tags "HIGH_PRIORITY_TAG_ID" \
  -action1-type "SEND_EMAIL" \
  -action1-email-to "manager@company.com" \
  -action1-email-subject "🚨 High Priority Alert" \
  -action1-email-content "<h2>Manager Alert</h2><p>High priority issue needs attention</p>" \
  -action2-type "SEND_EMAIL" \
  -action2-email-to "team@company.com" \
  -action2-email-subject "📋 Team Notification" \
  -action2-email-content "<h2>Team Alert</h2><p>New high priority task assigned</p>" \
  -action3-type "ADD_COLOR" \
  -action3-color "#ff4444"

# Single action using clean syntax (no numbering needed)
go run . create-automation-multi -project PROJECT_ID \
  -trigger-type "TODO_COMPLETED" \
  -action-type "SEND_EMAIL" \
  -email-to "success@company.com" \
  -email-subject "Task Completed!" \
  -email-content "<p>Another task completed successfully!</p>"
```

**Key Features**:
- **Per-Action Settings**: Each action can have different email/HTTP settings
- **Unnumbered Flags**: Use `-action-type` for single actions (clean syntax)
- **Numbered Flags**: Use `-action1-type`, `-action2-type`, etc. for multiple actions
- **Mixed Support**: Can combine unnumbered and numbered flags
- **Up to 3 Actions**: Support for action1, action2, and action3

### Automations Update (`update-automation`) - NEW
```bash
# Enable/disable automation
go run . update-automation -automation AUTOMATION_ID -project PROJECT_ID \
  -active true

# Update trigger settings
go run . update-automation -automation AUTOMATION_ID -project PROJECT_ID \
  -trigger-type "TODO_MARKED_AS_COMPLETE" \
  -trigger-todo-list "NEW_LIST_ID"

# Update email action settings
go run . update-automation -automation AUTOMATION_ID -project PROJECT_ID \
  -email-to "newemail@example.com" \
  -email-subject "Updated subject" \
  -email-content "<p>Updated content</p>"

# Update HTTP webhook settings
go run . update-automation -automation AUTOMATION_ID -project PROJECT_ID \
  -http-url "https://newwebhook.com/endpoint" \
  -http-method "PUT" \
  -http-body '{"updated": true}'

# Update action type and settings
go run . update-automation -automation AUTOMATION_ID -project PROJECT_ID \
  -action-type "ADD_TAG" \
  -action-tags "new_tag_id"
```

**Required Parameters**: `-automation` (automation ID), `-project` (project ID for context)

**Update Options**:
- **Active Status**: `-active` (true/false)
- **Trigger Updates**: All trigger options from create-automation
- **Action Updates**: All action options from create-automation
- **Partial Updates**: Only specified fields are updated; others remain unchanged

**Important Notes**:
- At least one field must be provided to update
- Partial updates are supported - only specified fields are changed  
- Use same trigger/action types and options as create-automation
- Project context required for proper authorization
- **⚠️ CRITICAL**: Action updates REPLACE the entire actions array. For multi-action automations, you must specify ALL existing actions to avoid losing any

### Multi-Action Automation Updates (`update-automation-multi`) - NEW
```bash
# Update single action using clean syntax
go run . update-automation-multi -automation AUTOMATION_ID -project PROJECT_ID \
  -action-type "SEND_EMAIL" \
  -email-subject "Updated Subject" \
  -email-content "<h2>Updated Content</h2>"

# Update multiple actions with different settings
go run . update-automation-multi -automation AUTOMATION_ID -project PROJECT_ID \
  -action1-type "SEND_EMAIL" \
  -action1-email-to "manager@company.com" \
  -action1-email-subject "Updated Manager Alert" \
  -action2-type "SEND_EMAIL" \
  -action2-email-to "team@company.com" \
  -action2-email-subject "Updated Team Alert" \
  -action3-type "ADD_COLOR" \
  -action3-color "#00ff00"

# Update only automation status (safe)
go run . update-automation-multi -automation AUTOMATION_ID -project PROJECT_ID \
  -active false
```

**Enhanced Features**:
- **Per-Action Email/HTTP Settings**: Different settings for each action
- **Unnumbered + Numbered Flags**: Same flexibility as create-automation-multi
- **Partial Updates**: Update only specified actions
- **⚠️ WARNING**: Displays warning when updating actions to prevent data loss

**Critical Warning**: Action updates replace the entire actions array. For multi-action automations, specify ALL existing actions to avoid losing any.

### Automation Deletion (`delete-automation`) - NEW
```bash
# Delete an automation (requires confirmation)
go run . delete-automation -automation AUTOMATION_ID -project PROJECT_ID -confirm

# Simple output format
go run . delete-automation -automation AUTOMATION_ID -project PROJECT_ID -confirm -simple
```

**Required Parameters**: `-automation` (automation ID), `-project` (project ID), `-confirm` (safety confirmation)

**Safety Features**:
- **Confirmation Required**: Must use `-confirm` flag to prevent accidents
- **Project Context**: Requires project ID for authorization
- **Permanent Action**: Cannot be undone once deleted
- **Clear Feedback**: Shows success/failure status

### Record Moving (`move-record`) - NEW
```bash
# Move record to different list in same or different project
go run . move-record -record RECORD_ID -list LIST_ID

# Move record with simple output
go run . move-record -record RECORD_ID -list LIST_ID -simple
```

**Required Parameters**: `-record` (record ID), `-list` (destination list ID)

**Key Features**:
- **Cross-Project Moves**: Automatically handles moving records between different projects
- **Same-Project Moves**: Move records between lists within the same project
- **Position Management**: Record position is automatically set in destination list
- **Project Detection**: Automatically detects destination project from list
- **Simple Interface**: Clean, focused command for just moving records

**Implementation Details**:
- Uses the `editTodo` mutation with just `todoId` and `todoListId` parameters
- More focused than `update-record -list` command - dedicated to moving operations
- Returns updated record information including destination list and project details
- No project context required - system handles cross-project authorization automatically

### User Invitations (`invite-user`) - NEW
```bash
# Invite user to company with basic member access
go run . invite-user -email "user@example.com" -access-level "MEMBER"

# Invite user to specific project with admin access
go run . invite-user -email "admin@example.com" -access-level "ADMIN" \
  -project PROJECT_ID

# Invite user to multiple projects with client access
go run . invite-user -email "client@example.com" -access-level "CLIENT" \
  -projects "PROJECT_ID1,PROJECT_ID2,PROJECT_ID3"

# Invite user with custom project role
go run . invite-user -email "user@example.com" -access-level "MEMBER" \
  -project PROJECT_ID -role "CUSTOM_ROLE_ID"
```

**Required Parameters**: `-email` (email address), `-access-level` (OWNER, ADMIN, MEMBER, CLIENT, COMMENT_ONLY)

**Optional Parameters**:
- `-project` (single project ID)
- `-projects` (comma-separated project IDs)
- `-company` (company ID, uses default if not specified)
- `-role` (custom role ID for project-specific invitations)

**Access Levels**:
- `OWNER`: Full company/project access
- `ADMIN`: Administrative access
- `MEMBER`: Standard member access
- `CLIENT`: Client-level access (limited permissions)
- `COMMENT_ONLY`: Can only view and comment

### Project User Roles (`read-project-user-roles`) - NEW
```bash
# List custom roles for a single project
go run . read-project-user-roles -project PROJECT_ID -simple

# List roles across multiple projects with detailed info
go run . read-project-user-roles -projects "PROJECT_ID1,PROJECT_ID2"

# Export role data as JSON
go run . read-project-user-roles -project PROJECT_ID -format json

# Export role data as CSV
go run . read-project-user-roles -project PROJECT_ID -format csv
```

**Required Parameters**: Either `-project` (single project) or `-projects` (comma-separated list)

**Output Options**: 
- `-simple`: Basic role information only
- `-format`: Output format (table, json, csv)

**Role Information Displayed**:
- Role ID, UID, name, and description
- Permission flags (invite others, mark records done, delete records, etc.)
- Feature access (activity, chat, discussions, forms, wiki, files, records, people)
- Visibility settings (assigned todos only, mentioned comments only)
- Associated custom fields and todo lists
- Creation and update timestamps

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

# Add records with custom fields (note: SELECT fields require option IDs with trailing comma)
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