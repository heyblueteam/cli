# Blue Demo Builder

A collection of Go scripts for interacting with the Blue GraphQL API to create demo projects programmatically.

## 🚀 Quick Start

### Prerequisites
- Go 1.21+ installed
- Blue API credentials (personal access token, client ID, company ID)

### Setup
1. Clone this directory
2. Install dependencies:
   ```bash
   go mod tidy
   ```
3. Ensure `.env` file exists with your credentials (see Configuration section)

<!-- ─────────────────────────────────────────────────────────────── -->
<!--                                                                -->
<!--                📋   AVAILABLE SCRIPTS   📋                      -->
<!--                                                                -->
<!--  Use the scripts below to interact with the Blue API!           -->
<!--  Each script is designed for a specific demo-building task.     -->
<!--                                                                -->
<!-- ─────────────────────────────────────────────────────────────── -->


### 1. List Projects (`read-projects`)
Lists projects in your Blue company with pagination, search, and filtering.

```bash
# List first 20 projects (default)
go run . read-projects

# List with just names and IDs
go run . read-projects -simple

# Search for projects by name
go run . read-projects -search "marketing"

# Navigate through pages
go run . read-projects -page 2
go run . read-projects -page 3 -size 50

# Include archived and template projects
go run . read-projects -archived    # Include archived
go run . read-projects -templates   # Include templates  
go run . read-projects -all         # Show everything

# Combine options
go run . read-projects -search "CRM" -page 2 -simple
```

**Options:**
- `-simple`: Show only basic information (name and ID)
- `-page`: Page number to display (default: 1)
- `-size`: Number of items per page (default: 20)
- `-search`: Search projects by name
- `-archived`: Include archived projects
- `-templates`: Include template projects
- `-all`: Show all projects including archived and templates

### 2. Create Project (`create-project`)
Creates a new project in your Blue company.

```bash
# Create a basic project
go run . create-project -name "My Demo Project"

# Create with all options
go run . create-project \
  -name "Sprint Planning" \
  -description "Q1 2024 Sprint Planning" \
  -color blue \
  -icon rocket \
  -category ENGINEERING

# Show available options
go run . create-project -list
```

**Options:**
- `-name` (required): Project name
- `-description`: Project description
- `-color`: Color name (blue, red, green, etc.) or hex code (#3B82F6)
- `-icon`: Icon name (briefcase, rocket, star, etc.)
- `-category`: Project category (GENERAL, CRM, MARKETING, ENGINEERING, etc.)
- `-template`: Template ID to create from

### 3. Read Lists (`read-lists`)
Gets all lists in a specific project.

```bash
# Get lists with full details
go run . read-lists -project PROJECT_ID

# Get lists with simple output
go run . read-lists -project PROJECT_ID -simple
```

**Options:**
- `-project` (required): Project ID
- `-simple`: Show only basic information

### 4. Read Project Custom Fields (`read-project-custom-fields`)
Lists all custom fields within a specific project with detailed information.

```bash
# List custom fields with full details
go run . read-project-custom-fields -project PROJECT_ID

# List with simple output (name, type, ID, position only)
go run . read-project-custom-fields -project PROJECT_ID -simple

# Navigate through pages
go run . read-project-custom-fields -project PROJECT_ID -page 2
go run . read-project-custom-fields -project PROJECT_ID -page 3 -size 100

# Combine options
go run . read-project-custom-fields -project PROJECT_ID -simple -page 2 -size 25
```

**Options:**
- `-project` (required): Project ID
- `-simple`: Show only basic information (name, type, ID, position)
- `-page`: Page number to display (default: 1)
- `-size`: Number of items per page (default: 50)

**Custom Field Types Supported:**
- Text, Number, Date, Time, Currency, Location
- Button, Checkbox, Dropdown, Multi-select
- Formula, Sequence, Reference, and more

**Options:**
- `-project` (required): Project ID
- `-simple`: Show only basic list information

### 5. Read Records in Project (`read-project-records`)
Lists all records across all lists in a project (project overview).

```bash
# List all records in all lists of a project
go run . read-project-records -project PROJECT_ID
```

**Options:**
- `-project` (required): Project ID

### 6. Read Records in Specific List (`read-list-records`)
Lists records within a specific records list with filtering and sorting options.

```bash
# List records in a list with full details
go run . read-list-records -list LIST_ID

# List records with simple output
go run . read-list-records -list LIST_ID -simple

# Search for specific records
go run . read-list-records -list LIST_ID -search "bug fix"

# Filter by completion status
go run . read-list-records -list LIST_ID -done false  # Show only active records
go run . read-list-records -list LIST_ID -done true   # Show only completed records

# Filter by assignee
go run . read-list-records -list LIST_ID -assignee USER_ID

# Filter by tags
go run . read-list-records -list LIST_ID -tags "tag1,tag2"

# Sort by different fields
go run . read-list-records -list LIST_ID -order title_ASC
go run . read-list-records -list LIST_ID -order duedAt_ASC
go run . read-list-records -list LIST_ID -order createdAt_DESC

# Limit the number of results
go run . read-list-records -list LIST_ID -limit 100
```

**Options:**
- `-list` (required): Todo List ID
- `-simple`: Show only basic record information
- `-search`: Search records by title or description
- `-assignee`: Filter by assignee ID
- `-tags`: Filter by tag IDs (comma-separated)
- `-done`: Filter by completion status (true/false)
- `-order`: Order by field (position_ASC, position_DESC, title_ASC, title_DESC, createdAt_ASC, createdAt_DESC, duedAt_ASC, duedAt_DESC)
- `-limit`: Maximum number of records to return (default: 50)

### 7. Create Lists (`create-list`)
Creates one or more lists in a project.

```bash
# Create multiple lists
go run . create-list -project PROJECT_ID -names "To Do,In Progress,Done"

# Create lists in reverse order (for right-to-left display)
go run . create-list -project PROJECT_ID -names "Done,In Progress,To Do" -reverse

# Create a single list
go run . create-list -project PROJECT_ID -names "Backlog"
```

**Options:**
- `-project` (required): Project ID where lists will be created
- `-names` (required): Comma-separated list names
- `-reverse`: Create lists in reverse order

### 8. Read Tags (`read-tags`)
Lists all tags within a specific project.

```bash
# List tags in a project
go run . read-tags -project PROJECT_ID
```

**Options:**
- `-project` (required): Project ID

### 9. Create Tags (`create-tags`)
Creates tags within a specific project.

```bash
# Create a tag
go run . create-tags -project PROJECT_ID -title "Bug" -color "red"

# Create different types of tags
go run . create-tags -project PROJECT_ID -title "Feature" -color "blue"
go run . create-tags -project PROJECT_ID -title "Urgent" -color "orange"
```

**Options:**
- `-project` (required): Project ID
- `-title` (required): Tag title/name
- `-color` (required): Tag color (e.g., "red", "blue", "green", etc.)

### 10. Create Custom Field (`create-custom-field`)
Creates custom fields for projects with support for 24+ field types.

```bash
# Create a simple text field
go run . create-custom-field -name "Notes" -type "TEXT_MULTI" -description "Additional notes"

# Create a number field with constraints
go run . create-custom-field -name "Story Points" -type "NUMBER" -min 1 -max 13

# Create a currency field
go run . create-custom-field -name "Cost" -type "CURRENCY" -currency "USD" -prefix "$"

# Create a date field
go run . create-custom-field -name "Start Date" -type "DATE" -is-due-date

# Create a unique ID field
go run . create-custom-field -name "Task ID" -type "UNIQUE_ID" -use-sequence -sequence-digits 8

# Show available options
go run . create-custom-field -list
```

**Available Field Types:**
- Text: `TEXT_SINGLE`, `TEXT_MULTI`
- Numbers: `NUMBER`, `CURRENCY`, `PERCENT`
- Dates/Time: `DATE`, `TIME_DURATION`
- Selection: `SELECT_SINGLE`, `SELECT_MULTI`, `CHECKBOX`
- Advanced: `RATING`, `EMAIL`, `PHONE`, `URL`, `LOCATION`, `COUNTRY`, `FILE`
- System: `UNIQUE_ID`, `FORMULA`, `REFERENCE`, `LOOKUP`, `BUTTON`, `CURRENCY_CONVERSION`

**Options:**
- `-name` (required): Custom field name
- `-type` (required): Field type (see available types above)
- `-description`: Field description
- `-min`, `-max`: Min/max values for NUMBER fields
- `-currency`: Currency code (default: USD)
- `-prefix`: Field prefix
- `-is-due-date`: Mark as due date field
- `-use-sequence`: Use sequence for UNIQUE_ID
- `-sequence-digits`: Number of digits in sequence (default: 6)
- `-reference-project`: Reference project ID for REFERENCE fields
- `-list`: List all available options

> **💡 For comprehensive custom field documentation** including detailed examples, field type explanations, and troubleshooting, see [CUSTOM_FIELDS_README.md](CUSTOM_FIELDS_README.md).

### 11. Create Record/Todo (`create-record`)
Creates records (records) within lists.

```bash
# Create a basic record
go run . create-record -list LIST_ID -title "Fix login bug"

# Create with description and assignees
go run . create-record -list LIST_ID -title "User Stories" -description "Create user stories for sprint" -assignees "user1,user2"

# Create with placement
go run . create-record -list LIST_ID -title "Priority Task" -placement "TOP"

# Simple output
go run . create-record -list LIST_ID -title "Task" -simple
```

**Options:**
- `-list` (required): List ID to create the record in
- `-title` (required): Title of the record
- `-description`: Description of the record
- `-placement`: Placement in list (`TOP` or `BOTTOM`)
- `-assignees`: Comma-separated assignee IDs
- `-simple`: Simple output format

### 12. Read Records (`read-records`)
Advanced querying of records across projects with filtering and sorting.

```bash
# List records in a project
go run . read-records -project PROJECT_ID

# List records in a specific list
go run . read-records -list LIST_ID

# Filter by assignee and completion status
go run . read-records -project PROJECT_ID -assignee USER_ID -done false

# Filter by tags
go run . read-records -project PROJECT_ID -tags "tag1,tag2"

# Sort by different fields
go run . read-records -project PROJECT_ID -order "duedAt_ASC"

# Pagination
go run . read-records -project PROJECT_ID -limit 50 -skip 100

# Simple output
go run . read-records -project PROJECT_ID -simple
```

**Options:**
- `-project`: Project ID to filter records
- `-list`: Todo List ID to filter records
- `-assignee`: Filter by assignee ID
- `-tags`: Filter by tag IDs (comma-separated)
- `-done`: Filter by completion status (true/false)
- `-archived`: Filter by archived status (true/false)
- `-order`: Order by field (position_ASC, position_DESC, title_ASC, title_DESC, createdAt_ASC, createdAt_DESC, updatedAt_ASC, updatedAt_DESC, duedAt_ASC, duedAt_DESC)
- `-limit`: Maximum number of records to return (default: 20)
- `-skip`: Number of records to skip (for pagination)
- `-simple`: Show only basic record information

### 13. Count Records (`read-records-count`)
Counts the total number of records in a project with optional filtering.

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
- `-project` (required): Project ID to count records
- `-list`: Todo List ID to filter records (optional)
- `-done`: Filter by completion status (true/false, optional)
- `-archived`: Filter by archived status (true/false, optional)

### 14. Delete Record/Todo (`delete-record`)
Permanently deletes a record from a project. Requires confirmation for safety.

```bash
# Delete a record (confirmation required)
go run . delete-record -record RECORD_ID -confirm

# Example with actual record ID
go run . delete-record -record "clr2x3y4z5a6b7c8d9e0" -confirm
```

**Options:**
- `-record` (required): Record/Todo ID to delete
- `-confirm` (required): Confirmation flag for safety (prevents accidental deletions)

**⚠️ Warning:** This operation permanently deletes the record and cannot be undone.

### 15. Add Tags to Records (`create-record-tags`)
Adds tags to existing records/todos. Supports adding tags by either tag IDs or tag titles.

```bash
# Add tags using tag IDs
go run . create-record-tags -record RECORD_ID -tag-ids "tag1,tag2"

# Add tags using tag titles (requires project context)
go run . create-record-tags -record RECORD_ID -tag-titles "Bug,Priority" -project PROJECT_ID

# Simple output
go run . create-record-tags -record RECORD_ID -tag-ids "tag1,tag2" -simple
```

**Options:**
- `-record` (required): Record/Todo ID to add tags to
- `-tag-ids`: Comma-separated list of tag IDs to add
- `-tag-titles`: Comma-separated list of tag titles to add (requires `-project`)
- `-project`: Project ID (required when using `-tag-titles`)
- `-simple`: Simple output format

**Note:** You must provide either `-tag-ids` or `-tag-titles` (with `-project`), but not both.

### 16. Update Project (`update-project`)
Updates project settings and toggles feature flags. Supports intelligent feature merging.

```bash
# Update project name and description
go run . update-project -project PROJECT_ID -name "New Name" -description "Updated description"

# Change project color and icon
go run . update-project -project PROJECT_ID -color "green" -icon "chart"

# Toggle project features (merged with existing settings)
go run . update-project -project PROJECT_ID -features "Chat:true,Files:false"

# Update todo alias and visibility settings
go run . update-project -project PROJECT_ID -todo-alias "Tasks" -hide-record-count true

# Disable multiple features
go run . update-project -project PROJECT_ID -features "Todo:false,People:false"

# Simple output
go run . update-project -project PROJECT_ID -name "Updated Name" -simple
```

**Options:**
- `-project` (required): Project ID or slug to update
- `-name`: New project name
- `-description`: New project description
- `-color`: New project color
- `-icon`: New project icon
- `-todo-alias`: Custom name for todos (e.g., "Tasks", "Items")
- `-hide-record-count`: Hide record count in UI (true/false)
- `-features`: Feature toggles in format "Feature:true/false,Feature2:true/false"
- `-simple`: Simple output format

**Available Features:**
- `Activity`, `Todo`, `Wiki`, `Chat`, `Docs`, `Forms`, `Files`, `People`

**Note:** Feature updates are merged with existing settings (partial updates supported).

### 17. Delete Project (`delete-project`)
Permanently deletes a project. Requires confirmation and special permissions.

```bash
# Delete a project (confirmation required)
go run . delete-project -project PROJECT_ID -confirm

# Example with project slug
go run . delete-project -project "my-demo-project" -confirm
```

**Options:**
- `-project` (required): Project ID or slug to delete
- `-confirm` (required): Confirmation flag for safety (prevents accidental deletions)

**⚠️ Warning:** 
- This operation permanently deletes the project and ALL its data
- Requires special permissions (may fail with authorization error)
- Cannot be undone

## 🔧 Configuration

Create a `.env` file in the demo-builder directory with the following variables:

```env
# Blue API Configuration
API_URL=https://api.blue.cc/graphql
AUTH_TOKEN=your_personal_access_token
CLIENT_ID=your_client_id
COMPANY_ID=your_company_slug
```

### Getting Your Credentials
1. **Personal Access Token**: Generate from Blue settings
2. **Client ID**: Found in your API settings
3. **Company ID**: Your company's slug (e.g., "heyblueteam")

## 📁 Project Structure

```
demo-builder/
├── .env                          # Your API credentials (git ignored)
├── .gitignore                    # Git ignore file  
├── go.mod                        # Go module file
├── go.sum                        # Go dependencies
├── main.go                       # Main entry point with command router
├── common/                       # Shared code
│   ├── auth.go                   # Centralized authentication and GraphQL client
│   ├── types.go                  # Shared type definitions
│   └── utils.go                  # Utility functions
├── tools/                        # All command implementations
│   ├── create_custom_field.go    # Create custom fields
│   ├── create_list.go            # Create lists in a project
│   ├── create_project.go         # Create new projects
│   ├── create_record_tags.go     # Add tags to records
│   ├── create_record.go          # Create records in lists
│   ├── create_tags.go            # Create tags in a project
│   ├── delete_project.go         # Delete projects
│   ├── delete_record.go          # Delete records
│   ├── read_list_records.go      # List records within a specific list
│   ├── read_lists.go             # Get lists in a project
│   ├── read_project_custom_fields.go # List custom fields in a project
│   ├── read_project_records.go   # List all records in a project
│   ├── read_projects.go          # List all projects
│   ├── read_records_count.go     # Count records in projects
│   ├── read_records.go           # Advanced record querying with filtering
│   ├── read_tags.go              # List tags in a project
│   └── update_project.go         # Update project settings and features
├── test/                         # Test suite
│   └── e2e.go                    # End-to-end test suite
├── README.md                     # This file
├── CLAUDE.md                     # Claude Code configuration
└── CUSTOM_FIELDS_README.md      # Detailed custom fields documentation
```

## 🎯 Example Workflow

Here's a complete example of creating a demo project:

```bash
# 1. List existing projects
go run . read-projects -simple

# 2. Create a new project
go run . create-project -name "Q1 Sprint Demo" -color blue -icon rocket

# 3. Get the project ID from the output, then create lists
go run . create-list -project PROJECT_ID -names "Backlog,To Do,In Progress,Done"

# 4. Verify the lists were created
go run . read-lists -project PROJECT_ID -simple

# 5. List all custom fields in the project
go run . read-project-custom-fields -project PROJECT_ID -simple

# 6. List tags in the project
go run . read-tags -project PROJECT_ID

# 7. Create some tags for the project
go run . create-tags -project PROJECT_ID -title "Bug" -color "red"
go run . create-tags -project PROJECT_ID -title "Feature" -color "blue"

# 8. List all records across all lists in the project
go run . read-project-records -project PROJECT_ID

# 9. Create some custom fields for the project
go run . create-custom-field -name "Priority" -type "SELECT_SINGLE" -description "Task priority level"
go run . create-custom-field -name "Story Points" -type "NUMBER" -min 1 -max 13

# 10. Create some records in the lists
go run . create-record -list LIST_ID -title "Setup project structure" -description "Initialize project with basic structure"
go run . create-record -list LIST_ID -title "Create user authentication" -placement "TOP"

# 11. List records in a specific list with detailed info
go run . read-list-records -list LIST_ID -simple

# 12. Query records across the project with filters
go run . read-records -project PROJECT_ID -done false -limit 10

# 13. Count total records in the project
go run . read-records-count -project PROJECT_ID
```

## 🧪 Testing

### End-to-End Test Suite (`e2e`)
A comprehensive test suite that validates all 17 tool files by executing them in sequence.

```bash
# Run the complete end-to-end test
go run . e2e
```

**What it tests:**
- ✅ Project operations (list, create, update, delete)
- ✅ List operations (create, read)
- ✅ Tag operations (create, read)
- ✅ Custom field operations (create multiple types, read)
- ✅ Record/Todo operations (create, tag, read, query, count, delete)
- ✅ Automatic cleanup (deletes test project)

**Features:**
- Emoji-friendly output (✅ pass, ❌ fail)
- Creates realistic test data
- Tests actual tool execution (not reimplemented logic)
- Generates unique names with timestamps to avoid conflicts
- Complete cleanup after testing
- Exit code 0 for success, 1 for failure (CI/CD friendly)

**Example output:**
```
🚀 Starting End-to-End Tests for Demo Builder
===================================================

📋 Running Tests:
---------------------------------------------------

🏗️  Project Operations:
✅ List existing projects
✅ Create project
✅ Update project settings and features

📝 List Operations:
✅ Create lists
✅ Read project lists

... (continues for all tests)

==================================================
📊 Test Summary:
   ✅ Passed: 22
   ❌ Failed: 0
   📈 Total:  22

✅ All tests passed successfully!
```

## 🛠️ Technical Details

### Architecture
- **auth**: Provides centralized authentication and GraphQL client
- All scripts use the shared `Client` from auth
- Environment variables are loaded from `.env` file
- Project context support via `client.SetProjectID()` method
- GraphQL queries are embedded in each script

### GraphQL API
- Uses Blue's GraphQL API at `https://api.blue.cc/graphql`
- Authentication via custom headers:
  - `X-Bloo-Token-ID`: Client ID
  - `X-Bloo-Token-Secret`: Auth Token  
  - `X-Bloo-Company-ID`: Company slug
  - `X-Bloo-Project-Id`: Project ID (when project context is set)

### Position System
- Lists use a floating-point position system
- Standard increment is 65535 between lists
- Allows for reordering without updating all positions

## 🚧 Limitations
- Maximum 50 lists per project
- Project names are automatically trimmed
- All scripts require authentication
- `list-projects` shows only first 20 projects (pagination not yet implemented)


## 🤝 Contributing
When adding new scripts:
1. Use the centralized auth for all API calls
2. Follow the existing command-line flag patterns
3. Use `client.SetProjectID()` for operations requiring project context
4. Include both simple and detailed output options where applicable
5. Update this README with usage examples

### Project Context Pattern
For scripts that operate within a specific project:

```go
// Create client
client := NewClient(config)

// Set project context for operations that require it
client.SetProjectID(projectID)

// Now mutations like createTag will work within the project context
```

This automatically adds the `X-Bloo-Project-Id` header to requests, enabling project-scoped operations like tag creation.

## 📝 License
Internal use only for Blue team demonstrations.