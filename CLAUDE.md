# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is the Blue CLI — a Go CLI tool built with [cobra](https://github.com/spf13/cobra) for interacting with the Blue GraphQL API. It provides a comprehensive set of commands organized as `blue <group> <action> [flags]`.

**Note:** "Workspaces" in the CLI maps to "projects" in the API. The CLI uses workspace terminology throughout, but GraphQL queries still reference the project API.

## Development Commands

### Building
```bash
go build -o blue .         # Build the binary
go mod tidy                # Install/update dependencies
```

### Running Commands
All commands follow this pattern:
```bash
./blue <group> <action> [flags]
```

Workspace ID or slug can be used interchangeably wherever `--workspace` is required.

## Command Reference

### Workspaces (`blue workspaces` / `blue ws`)
```bash
blue ws list --simple
blue ws list --search "CRM" --sort updatedAt_DESC --page 2 --size 50
blue ws create --name "Workspace Name" --color blue --icon rocket --category ENGINEERING
blue ws update --workspace <ID> --name "New Name" --features "Chat:true,Files:false"
blue ws delete --workspace <ID> --confirm
```

### Lists (`blue lists`)
```bash
blue lists list --workspace <ID> --simple
blue lists create --workspace <ID> --names "To Do,In Progress,Done"
blue lists create --workspace <ID> --names "Done,In Progress,To Do" --reverse
blue lists update --list <ID> --workspace <ID> --title "New Title" --locked true
blue lists delete --workspace <ID> --list <ID> --confirm
```

### Records (`blue records` / `blue rec`)
```bash
# List/query records
blue rec list --workspace <ID> --simple
blue rec list --workspace <ID> --done false --assignee <USER_ID> --tags "tag1,tag2"
blue rec list --workspace <ID> --custom-field "field_id:GT:50000" --stats
blue rec list --workspace <ID> --custom-field "field_id:CONTAINS:urgent" --calc
blue rec list --list <ID> --limit 50 --skip 100

# Get single record
blue rec get --record <ID> --workspace <ID>
blue rec get --record <ID> --workspace <ID> --simple

# Create records
blue rec create --workspace <ID> --list <ID> --title "Task Name"
blue rec create -w <ID> -l <ID> -t "Task" --description "Details" --assignees "user1,user2"
blue rec create -w <ID> -l <ID> -t "Task" --custom-fields "cf123:option_id_123,;cf456:42"

# Update records
blue rec update --record <ID> --workspace <ID> --title "New Title"
blue rec update -r <ID> -w <ID> --assignees "user1,user2" --tag-ids "tag1,tag2"
blue rec update -r <ID> -w <ID> --custom-fields "cf123:Updated Value;cf456:42"
blue rec update -r <ID> --due-date "2026-12-31" --timezone "UTC"

# Move records
blue rec move --record <ID> --list <ID> --workspace <ID>

# Count records
blue rec count --workspace <ID>
blue rec count --workspace <ID> --done false --list <ID>

# Delete records
blue rec delete --record <ID> --confirm
```

**Custom Field Filter Operators:** `EQ`, `NE`, `GT`, `GTE`, `LT`, `LTE`, `IN`, `NIN`, `CONTAINS`, `IS`, `NOT`

**Custom Field Values for SELECT fields MUST use option IDs with trailing comma:**
- Correct: `"cf123:option_id_123,"`
- Wrong: `"cf123:High"` (don't use display names)

### Tags (`blue tags`)
```bash
blue tags list --workspace <ID>
blue tags create --workspace <ID> --title "Bug" --color red
blue tags add --record <ID> --tag-ids "tag1,tag2"
blue tags add --record <ID> --tag-titles "Bug,Priority" --workspace <ID>
```

### Custom Fields (`blue fields` / `blue cf`)
```bash
# List fields
blue cf list --workspace <ID> --simple
blue cf list --workspace <ID> --detailed --examples --format json

# Create fields
blue cf create --workspace <ID> --name "Priority" --type "SELECT_SINGLE" --options "High:red,Medium:yellow,Low:green"
blue cf create --workspace <ID> --name "Story Points" --type "NUMBER" --min 1 --max 13
blue cf create --workspace <ID> --name "Cost" --type "CURRENCY" --currency "USD"

# Update/delete fields
blue cf update --field <ID> --workspace <ID> --name "New Name" --description "Updated"
blue cf delete --field <ID> --workspace <ID> --confirm

# Field options
blue cf options create --field <ID> --workspace <ID> --options "High:red,Medium:yellow,Low:green"
blue cf options delete --field <ID> --workspace <ID> --option-ids "id1,id2" --confirm

# Field groups
blue cf groups list --workspace <ID>
blue cf groups manage --workspace <ID> --action create --name "Group Name" --color "#ff0000"
```

**Available Field Types:** `TEXT_SINGLE`, `TEXT_MULTI`, `NUMBER`, `CURRENCY`, `PERCENT`, `DATE`, `TIME_DURATION`, `SELECT_SINGLE`, `SELECT_MULTI`, `CHECKBOX`, `RATING`, `EMAIL`, `PHONE`, `URL`, `LOCATION`, `COUNTRY`, `FILE`, `UNIQUE_ID`, `FORMULA`, `REFERENCE`, `LOOKUP`, `BUTTON`, `CURRENCY_CONVERSION`

### Automations (`blue automations` / `blue auto`)
```bash
# List
blue auto list --workspace <ID> --simple
blue auto list --workspace <ID> --page 2 --size 10

# Create (single action)
blue auto create --workspace <ID> --trigger-type "TODO_MARKED_AS_COMPLETE" --action-type "SEND_EMAIL" --email-to "user@example.com" --email-subject "Task done"

# Create (multi-action with numbered flags)
blue auto create --workspace <ID> --trigger-type "TAG_ADDED" --trigger-tags "TAG_ID" \
  --action1-type "SEND_EMAIL" --action1-email-to "mgr@co.com" \
  --action2-type "ADD_COLOR" --action2-color "#ff0000"

# Update
blue auto update --automation <ID> --workspace <ID> --active true
blue auto update --automation <ID> --workspace <ID> --action-type "SEND_EMAIL" --email-to "new@example.com"

# Delete
blue auto delete --automation <ID> --workspace <ID> --confirm
```

**Trigger Types:** `TODO_CREATED`, `TODO_MARKED_AS_COMPLETE`, `TODO_MARKED_AS_INCOMPLETE`, `TODO_LIST_CHANGED`, `TAG_ADDED`, `TAG_REMOVED`, `ASSIGNEE_ADDED`, `ASSIGNEE_REMOVED`, `CUSTOM_FIELD_CHANGED`, `TODO_OVERDUE`

**Action Types:** `SEND_EMAIL`, `MAKE_HTTP_REQUEST`, `ADD_TAG`, `REMOVE_TAG`, `ADD_ASSIGNEE`, `REMOVE_ASSIGNEE`, `ADD_COLOR`, `CHANGE_TODO_LIST`, `MARK_AS_COMPLETE`, `MARK_AS_INCOMPLETE`, `CREATE_TODO`, `COPY_TODO`, `DELETE_TODO`, `ARCHIVE_TODO`, `CREATE_CHECKLIST`

### Checklists (`blue checklists`)
```bash
blue checklists list --record <ID> --workspace <ID> --simple
blue checklists create --record <ID> --title "Pre-launch Tasks" --workspace <ID>
blue checklists delete --checklist <ID> --confirm

# Checklist items
blue checklists items create --checklist <ID> --title "Review docs" --position 1000.0
blue checklists items update --item <ID> --done true
blue checklists items update --item <ID> --title "Updated" --move-to-checklist <ID>
blue checklists items delete --item <ID> --confirm
```

### Comments (`blue comments`)
```bash
blue comments create --record <ID> --workspace <ID> --text "Comment text"
blue comments create --record <ID> --workspace <ID> --text "Comment" --html "<p><strong>Comment</strong></p>"
blue comments update --comment <ID> --workspace <ID> --text "Updated text"
```

### Users (`blue users`)
```bash
# List users
blue users list --simple                                     # Company-wide
blue users list --workspace <ID> --simple                    # Workspace-specific
blue users list --search "john" --first 100

# Invite users
blue users invite --email "user@example.com" --access-level "MEMBER" --workspace <ID>
blue users invite --email "admin@example.com" --access-level "ADMIN" --workspaces "ID1,ID2"

# List roles
blue users roles --workspace <ID> --simple
blue users roles --workspaces "ID1,ID2" --format json
```

**Access Levels:** `OWNER`, `ADMIN`, `MEMBER`, `CLIENT`, `COMMENT_ONLY`

### Dependencies (`blue dependencies` / `blue deps`)
```bash
blue deps create --record <ID> --other-record <ID> --type "BLOCKED_BY" --workspace <ID>
blue deps update --record <ID> --other-record <ID> --type "BLOCKS" --workspace <ID>
blue deps delete --record <ID> --other-record <ID> --confirm --workspace <ID>
```

### Files (`blue files`)
```bash
blue files download                                          # Interactive mode
blue files download --use-env --output "backup.zip" --parallel 10
```

## Architecture

### Project Structure
```
cli/
├── main.go              # Entry point — calls cmd.Execute()
├── cmd/                 # All cobra command definitions
│   ├── root.go          # Root command, version, global setup
│   ├── workspaces/      # blue workspaces *
│   ├── records/         # blue records *
│   ├── lists/           # blue lists *
│   ├── tags/            # blue tags *
│   ├── fields/          # blue fields *
│   │   ├── options/     # blue fields options *
│   │   └── groups/      # blue fields groups *
│   ├── automations/     # blue automations *
│   ├── checklists/      # blue checklists *
│   │   └── items/       # blue checklists items *
│   ├── comments/        # blue comments *
│   ├── users/           # blue users *
│   ├── dependencies/    # blue dependencies *
│   └── files/           # blue files *
├── common/              # Shared code (auth, types, utils)
│   ├── auth.go          # GraphQL client & authentication
│   ├── types.go         # Shared type definitions
│   └── utils.go         # Utility functions
```

### Authentication (`common/auth.go`)
- `Client` struct with GraphQL request method
- Environment variables from `.env` file
- Workspace context via `X-Bloo-Project-Id` header (API still uses project terminology)
- 30-second timeout for requests

### Required Environment Variables
```
API_URL=https://api.blue.cc/graphql
AUTH_TOKEN=your_personal_access_token
CLIENT_ID=your_client_id
COMPANY_ID=your_company_slug
```

### GraphQL API Details
- Endpoint: `https://api.blue.cc/graphql`
- Headers: `X-Bloo-Token-ID`, `X-Bloo-Token-Secret`, `X-Bloo-Company-ID`, `X-Bloo-Project-Id`
- 30-second timeout, POST method with JSON body

## Implementation Guidelines

When adding new commands:
1. Create a new directory under `cmd/` for the command group
2. Create a parent command file (e.g., `mygroup.go`) with exported `Cmd` variable
3. Create individual command files (e.g., `list.go`, `create.go`)
4. Register the group in `cmd/root.go` with `rootCmd.AddCommand(mygroup.Cmd)`
5. Use `--workspace`/`-w` for workspace context (maps to `client.SetProject()`)
6. Use `--simple`/`-s` for simplified output
7. Use `--confirm`/`-y` for destructive operations
8. Import `blue-cli/common` for shared types and auth

## Key Technologies
- Go with [cobra](https://github.com/spf13/cobra) for CLI framework
- GraphQL (raw queries, no client library)
- godotenv for `.env` configuration
- promptui for interactive prompts
