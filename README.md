# Blue Demo Builder


Wat is working

1. Create project
2. Get List
3. Get Tag
4. Create List
5. List tag

Go scripts for interacting with the Blue GraphQL API to create demo projects programmatically.

## Setup

1. Ensure you have Go 1.21+ installed
2. Run `go mod tidy` to install dependencies
3. Create a `.env` file with your Blue API credentials (already created)

## Scripts

### list-projects.go

Lists all projects in your Blue company.

**Usage:**

```bash
# List all projects with full details
go run list-projects.go

# List only project names and IDs (simple mode)
go run list-projects.go -simple
```

**Example output (simple mode):**
```
=== Projects in heyblueteam ===
Total projects: 20

1. Tech CRM
   ID: clhh41xu50cy5t51e6c5qjve6

2. Bakery Sales CRM
   ID: clhh4bl1d0d8jt51ed5rda0tg

...
```

## Environment Variables

The following environment variables are required in `.env`:

- `API_URL`: Blue GraphQL API endpoint
- `AUTH_TOKEN`: Your personal access token
- `CLIENT_ID`: Your client ID
- `COMPANY_ID`: Your company ID (slug)

### create-project.go

Creates a new project in your Blue company.

**Usage:**

```bash
# Create a basic project
go run create-project.go -name "My Demo Project"

# Create with all options
go run create-project.go \
  -name "Sprint Planning" \
  -description "Q1 2024 Sprint Planning" \
  -color blue \
  -icon rocket \
  -category ENGINEERING

# List available options
go run create-project.go -list
```

**Flags:**
- `-name` (required): Project name
- `-description`: Project description
- `-color`: Color name (blue, red, green, etc.) or hex code
- `-icon`: Icon name (briefcase, rocket, star, etc.)
- `-category`: Project category (GENERAL, CRM, MARKETING, etc.)
- `-template`: Template ID to create from
- `-list`: Show available colors, icons, and categories

### get-lists.go

Gets all lists in a project.

**Usage:**

```bash
# Get all lists with full details
go run get-lists.go -project PROJECT_ID

# Get lists with simple output
go run get-lists.go -project PROJECT_ID -simple
```

**Flags:**
- `-project` (required): Project ID
- `-simple`: Show only basic list information

### create-list.go

Creates one or more lists in a project.

**Usage:**

```bash
# Create multiple lists
go run create-list.go -project PROJECT_ID -names "To Do,In Progress,Done"

# Create lists in reverse order
go run create-list.go -project PROJECT_ID -names "Done,In Progress,To Do" -reverse

# Create a single list
go run create-list.go -project PROJECT_ID -names "Backlog"
```

**Flags:**
- `-project` (required): Project ID where lists will be created
- `-names` (required): Comma-separated list names
- `-reverse`: Create lists in reverse order (useful for right-to-left ordering)

**Notes:**
- Lists are created with positions spaced by 65535 units
- Maximum 50 lists per project
- Empty names are automatically filtered out

## Upcoming Scripts

- `create-records.go` - Create records/todos within lists