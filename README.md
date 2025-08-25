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


### 1. List Projects (`list-projects.go`)
Lists projects in your Blue company with pagination, search, and filtering.

```bash
# List first 20 projects (default)
go run auth.go list-projects.go

# List with just names and IDs
go run auth.go list-projects.go -simple

# Search for projects by name
go run auth.go list-projects.go -search "marketing"

# Navigate through pages
go run auth.go list-projects.go -page 2
go run auth.go list-projects.go -page 3 -size 50

# Include archived and template projects
go run auth.go list-projects.go -archived    # Include archived
go run auth.go list-projects.go -templates   # Include templates  
go run auth.go list-projects.go -all         # Show everything

# Combine options
go run auth.go list-projects.go -search "CRM" -page 2 -simple
```

**Options:**
- `-simple`: Show only basic information (name and ID)
- `-page`: Page number to display (default: 1)
- `-size`: Number of items per page (default: 20)
- `-search`: Search projects by name
- `-archived`: Include archived projects
- `-templates`: Include template projects
- `-all`: Show all projects including archived and templates

### 2. Create Project (`create-project.go`)
Creates a new project in your Blue company.

```bash
# Create a basic project
go run auth.go create-project.go -name "My Demo Project"

# Create with all options
go run auth.go create-project.go \
  -name "Sprint Planning" \
  -description "Q1 2024 Sprint Planning" \
  -color blue \
  -icon rocket \
  -category ENGINEERING

# Show available options
go run auth.go create-project.go -list
```

**Options:**
- `-name` (required): Project name
- `-description`: Project description
- `-color`: Color name (blue, red, green, etc.) or hex code (#3B82F6)
- `-icon`: Icon name (briefcase, rocket, star, etc.)
- `-category`: Project category (GENERAL, CRM, MARKETING, ENGINEERING, etc.)
- `-template`: Template ID to create from

### 3. Get Lists (`get-lists.go`)
Gets all lists in a specific project.

```bash
# Get lists with full details
go run auth.go get-lists.go -project PROJECT_ID

# Get lists with simple output
go run auth.go get-lists.go -project PROJECT_ID -simple
```

**Options:**
- `-project` (required): Project ID
- `-simple`: Show only basic information

### 4. List Project Custom Fields (`list-project-custom-fields.go`)
Lists all custom fields within a specific project with detailed information.

```bash
# List custom fields with full details
go run auth.go list-project-custom-fields.go -project PROJECT_ID

# List with simple output (name, type, ID, position only)
go run auth.go list-project-custom-fields.go -project PROJECT_ID -simple

# Navigate through pages
go run auth.go list-project-custom-fields.go -project PROJECT_ID -page 2
go run auth.go list-project-custom-fields.go -project PROJECT_ID -page 3 -size 100

# Combine options
go run auth.go list-project-custom-fields.go -project PROJECT_ID -simple -page 2 -size 25
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

### 4. List Todos in Project (`list-project-todos.go`)
Lists all todos across all lists in a project (project overview).

```bash
# List all todos in all lists of a project
go run auth.go list-project-todos.go -project PROJECT_ID
```

**Options:**
- `-project` (required): Project ID

### 5. List Todos in Specific List (`list-todos.go`)
Lists todos within a specific todo list with filtering and sorting options.

```bash
# List todos in a list with full details
go run auth.go list-todos.go -list LIST_ID

# List todos with simple output
go run auth.go list-todos.go -list LIST_ID -simple

# Search for specific todos
go run auth.go list-todos.go -list LIST_ID -search "bug fix"

# Filter by completion status
go run auth.go list-todos.go -list LIST_ID -done false  # Show only active todos
go run auth.go list-todos.go -list LIST_ID -done true   # Show only completed todos

# Filter by assignee
go run auth.go list-todos.go -list LIST_ID -assignee USER_ID

# Filter by tags
go run auth.go list-todos.go -list LIST_ID -tags "tag1,tag2"

# Sort by different fields
go run auth.go list-todos.go -list LIST_ID -order title_ASC
go run auth.go list-todos.go -list LIST_ID -order duedAt_ASC
go run auth.go list-todos.go -list LIST_ID -order createdAt_DESC

# Limit the number of results
go run auth.go list-todos.go -list LIST_ID -limit 100
```

**Options:**
- `-list` (required): Todo List ID
- `-simple`: Show only basic todo information
- `-search`: Search todos by title or description
- `-assignee`: Filter by assignee ID
- `-tags`: Filter by tag IDs (comma-separated)
- `-done`: Filter by completion status (true/false)
- `-order`: Order by field (position_ASC, position_DESC, title_ASC, title_DESC, createdAt_ASC, createdAt_DESC, duedAt_ASC, duedAt_DESC)
- `-limit`: Maximum number of todos to return (default: 50)

### 6. Create Lists (`create-list.go`)
Creates one or more lists in a project.

```bash
# Create multiple lists
go run auth.go create-list.go -project PROJECT_ID -names "To Do,In Progress,Done"

# Create lists in reverse order (for right-to-left display)
go run auth.go create-list.go -project PROJECT_ID -names "Done,In Progress,To Do" -reverse

# Create a single list
go run auth.go create-list.go -project PROJECT_ID -names "Backlog"
```

**Options:**
- `-project` (required): Project ID where lists will be created
- `-names` (required): Comma-separated list names
- `-reverse`: Create lists in reverse order

### 7. List Tags (`list-tags.go`)
Lists all tags within a specific project.

```bash
# List tags in a project
go run auth.go list-tags.go -project PROJECT_ID
```

**Options:**
- `-project` (required): Project ID

### 8. Create Tags (`create-tags.go`)
Creates tags within a specific project.

```bash
# Create a tag
go run auth.go create-tags.go -project PROJECT_ID -title "Bug" -color "red"

# Create different types of tags
go run auth.go create-tags.go -project PROJECT_ID -title "Feature" -color "blue"
go run auth.go create-tags.go -project PROJECT_ID -title "Urgent" -color "orange"
```

**Options:**
- `-project` (required): Project ID
- `-title` (required): Tag title/name
- `-color` (required): Tag color (e.g., "red", "blue", "green", etc.)

### 9. Create Custom Field (`create-custom-field.go`)
Creates custom fields for projects with support for 24+ field types.

```bash
# Create a simple text field
go run auth.go create-custom-field.go -name "Notes" -type "TEXT_MULTI" -description "Additional notes"

# Create a number field with constraints
go run auth.go create-custom-field.go -name "Story Points" -type "NUMBER" -min 1 -max 13

# Create a currency field
go run auth.go create-custom-field.go -name "Cost" -type "CURRENCY" -currency "USD" -prefix "$"

# Create a date field
go run auth.go create-custom-field.go -name "Start Date" -type "DATE" -is-due-date

# Create a unique ID field
go run auth.go create-custom-field.go -name "Task ID" -type "UNIQUE_ID" -use-sequence -sequence-digits 8

# Show available options
go run auth.go create-custom-field.go -list
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

### 10. Create Record/Todo (`create-record.go`)
Creates todos (records) within todo lists.

```bash
# Create a basic record
go run auth.go create-record.go -list LIST_ID -title "Fix login bug"

# Create with description and assignees
go run auth.go create-record.go -list LIST_ID -title "User Stories" -description "Create user stories for sprint" -assignees "user1,user2"

# Create with placement
go run auth.go create-record.go -list LIST_ID -title "Priority Task" -placement "TOP"

# Simple output
go run auth.go create-record.go -list LIST_ID -title "Task" -simple
```

**Options:**
- `-list` (required): List ID to create the record in
- `-title` (required): Title of the record
- `-description`: Description of the record
- `-placement`: Placement in list (`TOP` or `BOTTOM`)
- `-assignees`: Comma-separated assignee IDs
- `-simple`: Simple output format

### 11. List Records (`list-records.go`)
Advanced querying of records/todos across projects with filtering and sorting.

```bash
# List records in a project
go run auth.go list-records.go -project PROJECT_ID

# List records in a specific list
go run auth.go list-records.go -list LIST_ID

# Filter by assignee and completion status
go run auth.go list-records.go -project PROJECT_ID -assignee USER_ID -done false

# Filter by tags
go run auth.go list-records.go -project PROJECT_ID -tags "tag1,tag2"

# Sort by different fields
go run auth.go list-records.go -project PROJECT_ID -order "duedAt_ASC"

# Pagination
go run auth.go list-records.go -project PROJECT_ID -limit 50 -skip 100

# Simple output
go run auth.go list-records.go -project PROJECT_ID -simple
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

### 12. Delete Record/Todo (`delete-record.go`)
Permanently deletes a record/todo from a project. Requires confirmation for safety.

```bash
# Delete a record (confirmation required)
go run auth.go delete-record.go -record RECORD_ID -confirm

# Example with actual record ID
go run auth.go delete-record.go -record "clr2x3y4z5a6b7c8d9e0" -confirm
```

**Options:**
- `-record` (required): Record/Todo ID to delete
- `-confirm` (required): Confirmation flag for safety (prevents accidental deletions)

**⚠️ Warning:** This operation permanently deletes the record and cannot be undone.

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
├── auth.go                       # Centralized authentication and GraphQL client
├── list-projects.go              # List all projects
├── create-project.go             # Create new projects
├── get-lists.go                  # Get lists in a project
├── create-list.go                # Create lists in a project
├── list-tags.go                  # List tags in a project
├── create-tags.go                # Create tags in a project
├── list-project-custom-fields.go # List custom fields in a project
├── list-project-todos.go         # List all todos in a project
├── list-todos.go                 # List todos within a specific list
├── create-custom-field.go        # Create custom fields
├── create-record.go              # Create todos/records in lists
├── list-records.go               # Advanced record querying with filtering
├── delete-record.go              # Delete records/todos
└── README.md                     # This file
```

## 🎯 Example Workflow

Here's a complete example of creating a demo project:

```bash
# 1. List existing projects
go run auth.go list-projects.go -simple

# 2. Create a new project
go run auth.go create-project.go -name "Q1 Sprint Demo" -color blue -icon rocket

# 3. Get the project ID from the output, then create lists
go run auth.go create-list.go -project PROJECT_ID -names "Backlog,To Do,In Progress,Done"

# 4. Verify the lists were created
go run auth.go get-lists.go -project PROJECT_ID -simple

# 5. List all custom fields in the project
go run auth.go list-project-custom-fields.go -project PROJECT_ID -simple

# 6. List tags in the project
go run auth.go list-tags.go -project PROJECT_ID

# 7. Create some tags for the project
go run auth.go create-tags.go -project PROJECT_ID -title "Bug" -color "red"
go run auth.go create-tags.go -project PROJECT_ID -title "Feature" -color "blue"

# 8. List all todos across all lists in the project
go run auth.go list-project-todos.go -project PROJECT_ID

# 9. Create some custom fields for the project
go run auth.go create-custom-field.go -name "Priority" -type "SELECT_SINGLE" -description "Task priority level"
go run auth.go create-custom-field.go -name "Story Points" -type "NUMBER" -min 1 -max 13

# 10. Create some records in the lists
go run auth.go create-record.go -list LIST_ID -title "Setup project structure" -description "Initialize project with basic structure"
go run auth.go create-record.go -list LIST_ID -title "Create user authentication" -placement "TOP"

# 11. List todos in a specific list with detailed info
go run auth.go list-todos.go -list LIST_ID -simple

# 12. Query records across the project with filters
go run auth.go list-records.go -project PROJECT_ID -done false -limit 10

## 🛠️ Technical Details

### Architecture
- **auth.go**: Provides centralized authentication and GraphQL client
- All scripts use the shared `Client` from auth.go
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
- `list-projects.go` shows only first 20 projects (pagination not yet implemented)

## 🔮 Future Scripts
- `create-todos.go` - Create todos within lists
- `bulk-demo.go` - Create complete demo projects from templates

## 🤝 Contributing
When adding new scripts:
1. Use the centralized auth.go for all API calls
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