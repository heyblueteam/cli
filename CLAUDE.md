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
# List projects with sorting and search options
go run . read-projects -simple
go run . read-projects -search "CRM" -sort updatedAt_DESC

# Get lists in a project (using Project ID or slug)
go run . read-lists -project PROJECT_ID_OR_SLUG -simple

# Get detailed record information by ID
go run . read-record -record RECORD_ID -project PROJECT_ID -simple

# Advanced record querying with filtering and statistics (NEW ENHANCED)
go run . read-records -project PROJECT_ID -done false -assignee USER_ID -simple
go run . read-records -project PROJECT_ID -custom-field "field_id:GT:1000" -stats
go run . read-records -project PROJECT_ID -custom-field "field_id:CONTAINS:urgent" -limit 10

# List all todos across all lists in a project (overview)
go run . read-project-records -project PROJECT_ID

# List todos in a specific list (detailed with filtering)
go run . read-list-records -list LIST_ID -simple

# List tags in a project
go run . read-tags -project PROJECT_ID

# List custom fields in a project
go run . read-project-custom-fields -project PROJECT_ID

# Count records/todos in a project with optional filtering
go run . read-records-count -project PROJECT_ID
go run . read-records-count -project PROJECT_ID -done false
go run . read-records-count -project PROJECT_ID -list LIST_ID -archived false

# CREATE operations - Add new data
# Create project with options
go run . create-project -name "Demo" -color blue -icon rocket -category ENGINEERING

# Create lists in a project
go run . create-list -project PROJECT_ID -names "To Do,In Progress,Done"

# Create tags in a project
go run . create-tags -project PROJECT_ID -title "Bug" -color "red"

# Create custom fields (all types except reference/lookup)
go run . create-custom-field -name "Priority" -type "SELECT_SINGLE" -description "Task priority" -options "High:red,Medium:yellow,Low:green"
go run . create-custom-field -name "Status" -type "SELECT_MULTI" -options "In Progress,Blocked:red,Review Required:blue"
go run . create-custom-field -name "Story Points" -type "NUMBER" -min 1 -max 13
go run . create-custom-field -name "Cost" -type "CURRENCY" -currency "USD"

# Create records/todos in lists (supports custom fields)
go run . create-record -list LIST_ID -title "Task Name" -description "Description" -simple

# Create records/todos with custom field values
go run . create-record -list LIST_ID -title "Task" -custom-fields "cf123:Priority High;cf456:42"

# Create comments on records/todos
go run . create-comment -record RECORD_ID -text "This is a comment" -project PROJECT_ID -simple
go run . create-comment -record RECORD_ID -text "Progress update" -html "<p><strong>Progress update</strong><br>Making good progress on this task.</p>"

# Add tags to existing records/todos
go run . create-record-tags -record RECORD_ID -tag-ids "tag1,tag2" -simple
go run . create-record-tags -record RECORD_ID -tag-titles "Bug,Priority" -project PROJECT_ID

# UPDATE operations - Modify existing data
# Edit/update project settings and toggle features
go run . update-project -project PROJECT_ID -name "New Name" -features "Chat:true,Files:false"
go run . update-project -project PROJECT_ID -todo-alias "Tasks" -hide-record-count true
go run . update-project -project PROJECT_ID -features "Wiki:true,Docs:false" -simple

# Available feature types: Activity, Todo, Wiki, Chat, Docs, Forms, Files, People
# Features are merged with existing state (partial updates supported)
go run . update-project -project PROJECT_ID -features "Todo:false,People:false"

# Update individual records/todos with comprehensive field support
go run . update-record -record RECORD_ID -title "New Title" -description "Updated description"
go run . update-record -record RECORD_ID -move-to-list LIST_ID -assignees "user1,user2"
go run . update-record -record RECORD_ID -custom-fields "cf123:Updated Value;cf456:42"

# Update existing comments
go run . update-comment -comment COMMENT_ID -text "Updated comment text" -project PROJECT_ID -simple
go run . update-comment -comment COMMENT_ID -text "Updated text" -html "<p><em>Updated with formatting</em></p>"

# DELETE operations - Remove data
# Delete project (requires confirmation and special permissions)
go run . delete-project -project PROJECT_ID -confirm

# Delete records/todos (requires confirmation for safety)
go run . delete-record -record RECORD_ID -confirm
```

### Detailed Script Documentation

#### Project Listing (`read-projects`) - ENHANCED WITH SORTING AND FILTERING
List and search projects with advanced sorting, pagination, and filtering capabilities.

```bash
# Basic project listing
go run . read-projects -simple
go run . read-projects

# Sorting options
go run . read-projects -sort name_ASC -simple
go run . read-projects -sort name_DESC -simple
go run . read-projects -sort createdAt_DESC -simple
go run . read-projects -sort updatedAt_DESC -simple
go run . read-projects -sort position_ASC -simple

# Search projects by name
go run . read-projects -search "CRM" -simple
go run . read-projects -search "Test" -sort updatedAt_DESC

# Pagination
go run . read-projects -page 1 -size 10
go run . read-projects -page 2 -size 5

# Include archived and template projects
go run . read-projects -all
go run . read-projects -archived -simple
go run . read-projects -templates -simple

# Combined options
go run . read-projects -search "Test" -sort name_ASC -archived -page 1 -simple
```

**Options:**
- `-simple`: Show only project names and IDs
- `-page`: Page number (default: 1)
- `-size`: Page size (default: 20)
- `-search`: Search projects by name (case-sensitive)
- `-sort`: Sort projects by field (default: name_ASC)
- `-all`: Show all projects including archived and templates
- `-archived`: Include archived projects
- `-templates`: Include template projects

**Available Sort Options:**
- **Name**: `name_ASC`, `name_DESC`
- **Creation Date**: `createdAt_ASC`, `createdAt_DESC`
- **Update Date**: `updatedAt_ASC`, `updatedAt_DESC`
- **Position**: `position_ASC`, `position_DESC`

**Output Features:**
- **Pagination info**: Shows current page, total pages, and total items
- **Filter indicators**: Displays active search terms, sort options, and filters
- **Navigation help**: Provides next/previous page commands when applicable
- **Simple vs Detailed**: Simple shows ID/name only, detailed includes description, colors, icons, timestamps

#### Advanced Record Querying (`read-records`) - ENHANCED WITH CLIENT-SIDE FILTERING
Query records across projects with advanced filtering, client-side custom field filtering, and automatic numerical statistics.

```bash
# Basic querying with standard filters
go run . read-records -project PROJECT_ID -simple -limit 10
go run . read-records -project PROJECT_ID -done false -list LIST_ID

# Client-side custom field filtering with operators (FIXED - Now Working!)
go run . read-records -project PROJECT_ID -custom-field "cf123:GT:50000" -simple
go run . read-records -project PROJECT_ID -custom-field "cf456:CONTAINS:urgent" -limit 5
go run . read-records -project PROJECT_ID -custom-field "cf789:EQ:high" -done false

# Automatic numerical calculations (NEW)
go run . read-records -project PROJECT_ID -calc -simple
go run . read-records -project PROJECT_ID -custom-field "cf123:GT:1000" -calc

# Manual numerical statistics for custom fields
go run . read-records -project PROJECT_ID -stats -limit 20
go run . read-records -project PROJECT_ID -stats -calc-fields "cf123,cf456"

# Combined filtering, calculations, and statistics
go run . read-records -project PROJECT_ID -custom-field "cf123:GT:1000" -calc -stats -simple
```

**Standard Filtering Options:**
- `-project`: Project ID or slug (required)
- `-list`: Todo List ID to filter records
- `-assignee`: Filter by assignee ID
- `-tags`: Filter by tag IDs (comma-separated)
- `-done`: Filter by completion status (true/false)
- `-archived`: Filter by archived status (true/false)
- `-order`: Order by field (position_ASC/DESC, title_ASC/DESC, createdAt_ASC/DESC, etc.)
- `-limit`: Maximum number of records to return (default: 20)
- `-skip`: Number of records to skip (for pagination)
- `-simple`: Show only basic record information

**Custom Field Filtering (CLIENT-SIDE):**
- `-custom-field`: Filter by custom field using format "field_id:operator:value"
- **Implementation**: Client-side filtering after fetching records (server-side filtering not supported)
- **Performance**: Filters on already-fetched results, shows "Filter Applied: X → Y records (client-side)"

**Available Operators:**
- **Equality**: `EQ` (equal), `NE` (not equal)
- **Comparison**: `GT` (greater than), `GTE` (greater than or equal), `LT` (less than), `LTE` (less than or equal)
- **Membership**: `IN` (value in list), `NIN` (value not in list)
- **Text**: `CONTAINS` (contains substring)
- **Existence**: `IS` (is/is not null), `NOT` (negation)

**Numerical Statistics & Calculations:**
- `-calc`: Automatically detect and calculate stats for all numerical fields in results (NEW)
- `-stats`: Show numerical statistics for custom fields (sum, average, min, max)
- `-calc-fields`: Comma-separated list of custom field IDs to calculate stats for (optional - auto-detects if not specified)

**Automatic Calculation Features:**
- **Auto-detection**: Finds NUMBER, CURRENCY, PERCENT, RATING fields automatically
- **Real-time stats**: Calculates sum, average, min, max for numerical fields
- **Filtered calculations**: Stats apply to filtered results when combined with `-custom-field`
- **Multiple fields**: Handles multiple numerical fields simultaneously
- **Type safety**: Proper number conversion with error handling

**Custom Field Filter Examples:**
- Find records with amount over $50,000: `"cf123:GT:50000"`
- Find records containing 'urgent': `"cf456:CONTAINS:urgent"`
- Find records with specific priority: `"cf789:EQ:high"`
- Find records in value range: `"cf123:IN:1000,2000,3000"`

**Supported Field Types for Filtering:**
- **Numerical**: NUMBER, CURRENCY, PERCENT, RATING (supports GT, LT, GTE, LTE, EQ, NE)
- **Text**: TEXT_SINGLE, TEXT_MULTI, EMAIL, PHONE, URL (supports CONTAINS, EQ, NE)
- **Selection**: SELECT_SINGLE, SELECT_MULTI (supports EQ, NE, IN, NIN)
- **Boolean**: CHECKBOX (supports EQ, IS, NOT)

**Custom Field Display Format:**
- **Simple mode**: Shows field name and ID for disambiguation: `Deal Value (cmeqh9ts21czrsj1othq1rff8)=75000`
- **Detailed mode**: Shows field name, type, and ID: `Deal Value (CURRENCY) [cmeqh9ts21czrsj1othq1rff8]: 75000`
- **Field IDs are always visible** to handle cases where multiple fields might have the same name
- **Empty fields are hidden** by default - only fields with actual values are displayed
- **Value parsing**: Complex value structures are automatically parsed to show the relevant value (e.g., extracts `75000` from `map[currency:<nil> number:75000 text:<nil>]`)

#### Single Record Details (`read-record`)
Get comprehensive details for a specific record by ID, including custom field values.

```bash
# Get detailed record information with custom fields
go run . read-record -record RECORD_ID -project PROJECT_ID

# Get simple record information with custom field count
go run . read-record -record RECORD_ID -project PROJECT_ID -simple
```

**Options:**
- `-record` (required): Record ID to retrieve
- `-project`: Project ID or slug (required for context)
- `-simple`: Show only basic record information and custom field count

**Custom Field Display:**
- **Detailed mode**: Shows all custom field values with names, types, and parsed values (e.g., `Deal Value (CURRENCY) [field_id]: 75000`)
- **Simple mode**: Shows only the count of custom fields attached to the record
- **Field names and types**: Field names and types are fetched from project metadata for user-friendly display
- **Value parsing**: Complex API structures are automatically parsed to show meaningful values (e.g., extracts `75000` from currency fields)
- **Field IDs**: Shown in brackets `[field_id]` for technical reference and disambiguation
- **Empty fields**: Clearly marked as `(empty)` when no value is set

#### Create Comment (`create-comment`)
Creates comments on records/todos with support for both plain text and HTML content.

```bash
# Create a simple text comment
go run . create-comment -record RECORD_ID -text "This is a progress update"

# Create comment with HTML formatting
go run . create-comment -record RECORD_ID -text "Progress update" -html "<p><strong>Progress update</strong><br>Making good progress on this task.</p>"

# Create comment with project context
go run . create-comment -record RECORD_ID -text "Task completed" -project PROJECT_ID -simple
```

**Options:**
- `-record` (required): Record ID to comment on
- `-text` (required): Plain text content of the comment
- `-html`: HTML content of the comment (optional - will use formatted text if not provided)
- `-project`: Project ID or slug for context (optional)
- `-simple`: Show only basic comment information after creation

**Notes:**
- If HTML is not provided, the text will be automatically converted to HTML with basic formatting
- Comments on records use the "TODO" category in the Blue system
- All comments are associated with the authenticated user making the request

#### Update Comment (`update-comment`)
Updates existing comments on records/todos with new text and HTML content.

```bash
# Update comment with new text
go run . update-comment -comment COMMENT_ID -text "Updated progress: task is now 75% complete"

# Update comment with custom HTML formatting
go run . update-comment -comment COMMENT_ID -text "Final update" -html "<p><strong>COMPLETED</strong><br>✅ Task finished ahead of schedule</p>"

# Update comment with project context and simple output
go run . update-comment -comment COMMENT_ID -text "Status changed to completed" -project PROJECT_ID -simple
```

**Options:**
- `-comment` (required): Comment ID to update
- `-text` (required): New plain text content for the comment
- `-html`: New HTML content for the comment (optional - will use formatted text if not provided)
- `-project`: Project ID or slug for context (optional)
- `-simple`: Show only basic comment information after update

**Notes:**
- Updates both the text and HTML content of the comment
- If HTML is not provided, the text will be automatically converted to HTML with basic formatting
- The comment's creation timestamp remains unchanged, but updatedAt is set to current time
- Only the comment author or users with appropriate permissions can update comments

#### Create Record (`create-record`)
Creates new records/todos in lists with support for custom field values, assignments, and placement options.

```bash
# Create a simple record
go run . create-record -project PROJECT_ID -list LIST_ID -title "Task Name"

# Create record with description and placement
go run . create-record -project PROJECT_ID -list LIST_ID -title "Task Name" -description "Task description" -placement TOP

# Create record with custom field values
go run . create-record -project PROJECT_ID -list LIST_ID -title "Task Name" -custom-fields "cf123:High Priority;cf456:42.5"

# Create record with assignees and custom fields
go run . create-record -project PROJECT_ID -list LIST_ID -title "Task Name" -assignees "user1,user2" -custom-fields "cf789:true"
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

#### Create Custom Field (`create-custom-field`)
Creates custom fields for projects with support for all field types including SELECT fields with options.

```bash
# Create SELECT_SINGLE field with options and colors
go run . create-custom-field -project PROJECT_ID -name "Priority" -type "SELECT_SINGLE" -options "High:red,Medium:yellow,Low:green"

# Create SELECT_MULTI field with options (some with colors, some without)
go run . create-custom-field -project PROJECT_ID -name "Labels" -type "SELECT_MULTI" -options "Bug:red,Feature,Enhancement:blue"

# Create other field types
go run . create-custom-field -project PROJECT_ID -name "Story Points" -type "NUMBER" -min 1 -max 13
go run . create-custom-field -project PROJECT_ID -name "Budget" -type "CURRENCY" -currency "USD"
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

#### Count Records (`read-records-count`)
Counts the total number of records/todos in a project with optional filtering.

```bash
# Count all records in a project
go run . read-records-count -project PROJECT_ID

# Count only incomplete records
go run . read-records-count -project PROJECT_ID -done false

# Count records in a specific list
go run . read-records-count -project PROJECT_ID -list LIST_ID

# Count non-archived records
go run . read-records-count -project PROJECT_ID -archived false
```

**Options:**
- `-project` (required): Project ID or Project slug to count records
- `-list`: Todo List ID to filter records (optional)
- `-done`: Filter by completion status (true/false, optional)
- `-archived`: Filter by archived status (true/false, optional)

#### Update Record (`update-record`)
Update individual records/todos with comprehensive field support, including moving between lists and updating custom fields.

```bash
# Update basic record information
go run . update-record -record RECORD_ID -title "New Title" -description "Updated description"

# Move record to different list and update assignees
go run . update-record -record RECORD_ID -move-to-list LIST_ID -assignees "user1,user2"

# Update custom field values
go run . update-record -record RECORD_ID -custom-fields "cf123:Updated Priority;cf456:75000"

# Complete comprehensive update
go run . update-record -record RECORD_ID -title "Updated Title" -description "New description" -move-to-list LIST_ID -assignees "user1" -custom-fields "cf123:High;cf456:50000" -due-date "2025-12-31" -simple
```

**Options:**
- `-record` (required): Record ID to update
- `-title`: New title for the record
- `-description`: New description for the record
- `-move-to-list`: List ID to move the record to
- `-assignees`: Comma-separated list of user IDs to assign
- `-custom-fields`: Custom field updates in format "field_id1:value1;field_id2:value2"
- `-due-date`: Due date in ISO format (YYYY-MM-DD or YYYY-MM-DDTHH:MM:SS)
- `-color`: Record color
- `-done`: Mark as done (true/false)
- `-archived`: Archive status (true/false)
- `-simple`: Simple output format

**Custom Field Update Examples:**
- Text field: `"cf123:Updated text value"`
- Number field: `"cf456:99.5"`
- Boolean field: `"cf789:false"`
- Multiple fields: `"cf123:Priority High;cf456:75000;cf789:true"`

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

### Centralized Authentication (`common/auth.go`)
Provides shared authentication and client functionality:
- `Client` struct with GraphQL request method
- Environment variable loading from `.env`
- Standard HTTP headers for Blue API authentication
- Project context support via `X-Bloo-Project-Id` header (accepts both Project ID and Project slug)
- Automatic detection of Project ID vs Project slug format
- 30-second timeout for requests

### Shared Types (`common/types.go`)
Centralized type definitions to eliminate duplication:
- User, Tag, Project, TodoList, Record, CustomField types
- Input types for mutations (CreateProjectInput, CreateTodoInput, etc.)
- Separate pagination types (CursorPageInfo, OffsetPageInfo)

### GraphQL Integration Pattern
Each tool:
1. Imports the common package using dot imports
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

### End-to-End Test (`test/e2e.go`)
Comprehensive test suite that validates all 20+ commands:

```bash
# Run the end-to-end test
go run . e2e
```

**Coverage:**
- Tests all CRUD operations (Create, Read, Update, Delete)
- Validates project → lists → tags → custom fields → records workflow
- Tests advanced record querying with custom field filtering
- Uses actual command execution through the main router
- Automatic cleanup (deletes test project)
- 25+ test cases covering all major functionality including new enhanced features

**Output:**
- Emoji-friendly status indicators (✅/❌)
- Detailed progress reporting
- Summary with pass/fail counts
- Exit code 0 for success, 1 for failure (CI/CD compatible)

## Implemented Features

Completed:
- ✅ **ENHANCED**: List projects with advanced sorting, pagination, search, and filtering
  - Sorting by name, created/updated dates, position (ASC/DESC)
  - Search by project name with real-time filtering
  - Include/exclude archived and template projects
- ✅ Create projects with customization options
- ✅ Delete projects (with safety confirmation)
- ✅ List and create todo lists in projects
- ✅ List todos with filtering and pagination
- ✅ List and create tags in projects
- ✅ List custom fields in projects
- ✅ Create custom fields (24+ types including reference/lookup)
- ✅ Create records/todos with custom field values and assignments
- ✅ **ENHANCED**: Single record details with comprehensive custom field display
  - Field names, types, and parsed values with proper formatting
  - Field IDs shown for technical reference and disambiguation
  - Empty field handling and value extraction from complex structures
- ✅ **FIXED**: Client-side custom field filtering and automatic numerical calculations
  - **Client-side filtering**: Works with all operators (GT, LT, EQ, CONTAINS, IN, etc.)
  - **Auto calculations**: Automatic detection and stats for numerical fields (-calc flag)
  - **Real-time statistics**: Sum, average, min, max for CURRENCY, NUMBER, PERCENT, RATING fields
  - **Filtered calculations**: Stats apply to filtered subsets of data
  - **Intelligent value parsing**: Extracts meaningful values from complex data structures
  - Support for all custom field types with appropriate operators
- ✅ Count records/todos in projects with filtering options
- ✅ Delete records/todos with safety confirmation
- ✅ Add tags to records/todos (by tag ID or title)
- ✅ Update individual records with full field support
- ✅ Edit/update project settings and toggle features (with intelligent feature merging)
- ✅ End-to-end test suite with full coverage

## Planned Features

To implement:
- Create custom field groups
- Create automations
- Create custom user roles
- Bulk record operations (bulk update, bulk delete)
- Advanced export/import functionality
- Real-time record watching and notifications

## Implementation Guidelines

When adding new commands:
1. Create a new file in the `tools/` directory
2. Use the common package with dot imports for shared functionality
3. Follow the existing command-line flag patterns using Go's `flag` package
4. Add the command to the switch statement in `main.go`
5. Use `client.SetProjectID()` for operations that require project context
6. Include both `-simple` and detailed output options where applicable
7. Define proper struct types for GraphQL responses
8. Handle errors consistently with proper context
9. For operations that modify arrays/lists, implement proper merging logic to preserve existing data
10. Update CLAUDE.md with usage examples
11. Add test cases to `test/e2e.go`

### Feature Toggle Implementation Notes

The `update-project` command implements intelligent feature merging:
- Fetches current project state before making changes
- Merges user-specified feature toggles with existing features
- Sends complete feature array to prevent data loss
- Supports 8 feature types: Activity, Todo, Wiki, Chat, Docs, Forms, Files, People
- All features default to enabled=true for new projects

## Known Limitations

- Project listing limited to first 20 results (pagination not implemented)
- Maximum 50 lists per project
- Project deletion requires special permissions (may fail with authorization error)
- No linting configuration

## GraphQL API Details

- Endpoint: `https://api.blue.cc/graphql`
- Authentication Headers:
  - `X-Bloo-Token-ID`: Client ID
  - `X-Bloo-Token-Secret`: Auth Token
  - `X-Bloo-Company-ID`: Company slug
- Request timeout: 30 seconds
- All requests use POST method with JSON body

## Practical Usage Examples

### CRM System Management
```bash
# Set up a CRM project
go run . create-project -name "CRM System" -color blue -icon "office-building"

# Create pipeline lists
go run . create-list -project PROJECT_ID -names "Leads,Prospects,Customers,Closed Won,Closed Lost"

# Create custom fields for deal tracking
go run . create-custom-field -project PROJECT_ID -name "Deal Value" -type "CURRENCY" -currency "USD"
go run . create-custom-field -project PROJECT_ID -name "Priority" -type "SELECT_SINGLE" -options "High:red,Medium:yellow,Low:green"

# Add prospects with deal values
go run . create-record -project PROJECT_ID -list LIST_ID -title "TechCorp - Enterprise Deal" -custom-fields "cf123:75000"

# Find high-value deals (CLIENT-SIDE FILTERING - NOW WORKING!)
go run . read-records -project cmeqh9ceq1j4er41oseriuxhv -custom-field "cmeqh9ts21czrsj1othq1rff8:GT:50000" -simple

# Automatic deal statistics (NEW FEATURE)
go run . read-records -project cmeqh9ceq1j4er41oseriuxhv -calc -simple

# Manual statistics for specific fields
go run . read-records -project cmeqh9ceq1j4er41oseriuxhv -stats -calc-fields "cmeqh9ts21czrsj1othq1rff8"

# Combined filtering + calculations
go run . read-records -project cmeqh9ceq1j4er41oseriuxhv -custom-field "cmeqh9ts21czrsj1othq1rff8:EQ:75000" -calc
```

### Project Portfolio Management
```bash
# Query incomplete tasks across projects
go run . read-records -project PROJECT_ID -done false -limit 20

# Find companies containing "ABC" (CLIENT-SIDE TEXT FILTERING)
go run . read-records -project cmeqh9ceq1j4er41oseriuxhv -custom-field "cmeqhal6a1d1msj1o1qpco9wm:CONTAINS:ABC" -simple

# Update task priorities in bulk
go run . read-records -project PROJECT_ID -custom-field "cf789:EQ:low" -simple | \
  grep "ID:" | awk '{print $2}' | \
  xargs -I {} go run . update-record -record {} -custom-fields "cf789:medium"
```

### Data Analysis and Reporting
```bash
# Get comprehensive project statistics with new auto-calculations
go run . read-records-count -project cmeqh9ceq1j4er41oseriuxhv
go run . read-records -project cmeqh9ceq1j4er41oseriuxhv -calc -simple

# Advanced project analysis with sorting and search
go run . read-projects -search "CRM" -sort updatedAt_DESC -simple
go run . read-projects -all -sort createdAt_ASC

# Export filtered data
go run . read-records -project cmeqh9ceq1j4er41oseriuxhv -custom-field "cmeqh9ts21czrsj1othq1rff8:GT:50000" > high_value_deals.txt

# Detailed record inspection
go run . read-record -record dc41f3cf00b94040946c2a4cd1aac356 -project cmeqh9ceq1j4er41oseriuxhv
```