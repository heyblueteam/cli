# Blue CLI

A command-line interface for managing Blue workspaces, records, lists, tags, custom fields, automations, and more.

## Install

```bash
brew install heyblueteam/tap/blue-cli
```

## Setup

```bash
blue init
```

This prompts for your API credentials (get them from Account Settings > API > Generate Token) and saves them to `~/.config/blue/config.env`.

## Verify

```bash
blue workspaces list --simple
```

## Development

If you're working on the CLI itself:

```bash
go mod tidy              # Install dependencies
go build -o blue ./cmd/blue  # Build the binary
./blue --help            # Verify it works
```

You can also create a `.env` file in the project directory — it takes priority over the global config:

```env
API_URL=https://api.blue.cc/graphql
AUTH_TOKEN=your_personal_access_token
CLIENT_ID=your_client_id
COMPANY_ID=your_company_slug
```

**Getting Your Credentials:**
1. **Client ID & Auth Token**: Account Settings → API → Generate Token
2. **Company ID**: Your org slug from the URL (e.g., `acme` from `blue.cc/org/acme`)

## Commands

```
blue [command]

Available Commands:
  workspaces   (ws)    Manage workspaces
  records      (rec)   Manage records
  lists                Manage lists
  tags                 Manage tags
  fields       (cf)    Manage custom fields
  automations  (auto)  Manage automations
  checklists           Manage checklists
  comments             Manage comments
  users                Manage users
  dependencies (deps)  Manage record dependencies
  files                Manage files
  completion           Generate shell completions
  version              Print version information
```

Use `blue <command> --help` for details on any command.

### Workspaces

```bash
blue workspaces list --simple
blue workspaces list --search "CRM" --sort updatedAt_DESC
blue workspaces create --name "Sprint Planning" --color blue --icon rocket --category ENGINEERING
blue workspaces update --workspace <id> --name "New Name" --features "Chat:true,Files:false"
blue workspaces delete --workspace <id> --confirm
```

### Records

```bash
blue records list --workspace <id> --simple
blue records list --workspace <id> --done false --assignee <user_id>
blue records list --workspace <id> --custom-field "cf123:GT:50000" --stats
blue records get --record <id> --workspace <id>
blue records create --workspace <id> --list <id> --title "Fix login bug"
blue records create -w <id> -l <id> -t "Task" --custom-fields "cf123:option_id,"
blue records update --record <id> --workspace <id> --title "New Title"
blue records update -r <id> -w <id> --assignees "user1,user2" --tag-ids "tag1,tag2"
blue records move --record <id> --list <id> --workspace <id>
blue records count --workspace <id> --done false
blue records delete --record <id> --confirm
```

### Lists

```bash
blue lists list --workspace <id> --simple
blue lists create --workspace <id> --names "To Do,In Progress,Done"
blue lists update --list <id> --workspace <id> --title "Backlog" --locked true
blue lists delete --workspace <id> --list <id> --confirm
```

### Tags

```bash
blue tags list --workspace <id>
blue tags create --workspace <id> --title "Bug" --color red
blue tags add --record <id> --tag-ids "tag1,tag2"
blue tags add --record <id> --tag-titles "Bug,Priority" --workspace <id>
```

### Custom Fields

```bash
blue fields list --workspace <id> --simple
blue fields list --workspace <id> --detailed --examples
blue fields create --workspace <id> --name "Priority" --type "SELECT_SINGLE" --options "High:red,Medium:yellow,Low:green"
blue fields create --workspace <id> --name "Story Points" --type "NUMBER" --min 1 --max 13
blue fields update --field <id> --workspace <id> --name "New Name"
blue fields delete --field <id> --workspace <id> --confirm

# Field options
blue fields options create --field <id> --workspace <id> --options "High:red,Medium:yellow"
blue fields options delete --field <id> --workspace <id> --option-ids "id1,id2" --confirm

# Field groups
blue fields groups list --workspace <id>
blue fields groups manage --workspace <id> --action create --name "Group Name"
```

### Automations

```bash
blue automations list --workspace <id> --simple
blue automations create --workspace <id> --trigger-type "TODO_MARKED_AS_COMPLETE" --action-type "SEND_EMAIL" --email-to "user@example.com"
blue automations create --workspace <id> --trigger-type "TAG_ADDED" --action1-type "SEND_EMAIL" --action1-email-to "mgr@co.com" --action2-type "ADD_COLOR" --action2-color "#ff0000"
blue automations update --automation <id> --workspace <id> --active true
blue automations delete --automation <id> --workspace <id> --confirm
```

### Checklists

```bash
blue checklists list --record <id> --workspace <id>
blue checklists create --record <id> --title "Pre-launch Tasks" --workspace <id>
blue checklists delete --checklist <id> --confirm

# Checklist items
blue checklists items create --checklist <id> --title "Review docs" --position 1000.0
blue checklists items update --item <id> --done true
blue checklists items update --item <id> --title "Updated title" --position 2000.0
blue checklists items delete --item <id> --confirm
```

### Comments

```bash
blue comments create --record <id> --workspace <id> --text "Progress update"
blue comments update --comment <id> --workspace <id> --text "Updated comment"
```

### Users

```bash
blue users list --simple                                    # Company-wide
blue users list --workspace <id> --simple                   # Workspace-specific
blue users list --search "john" --first 100
blue users invite --email "user@example.com" --access-level "MEMBER" --workspace <id>
blue users roles --workspace <id> --simple
```

### Dependencies

```bash
blue dependencies create --record <id> --other-record <id> --type "BLOCKED_BY" --workspace <id>
blue dependencies update --record <id> --other-record <id> --type "BLOCKS" --workspace <id>
blue dependencies delete --record <id> --other-record <id> --confirm
```

### Files

```bash
blue files download                                          # Interactive mode
blue files download --use-env --output "backup.zip" --parallel 10
```

### Shell Completions

```bash
blue completion bash > /etc/bash_completion.d/blue
blue completion zsh > "${fpath[1]}/_blue"
blue completion fish > ~/.config/fish/completions/blue.fish
```

## Custom Field Values

When creating or updating records with custom fields, use this format:

```bash
--custom-fields "field_id1:value1;field_id2:value2"
```

**Important:** For SELECT fields, use option IDs with trailing comma:
- Single select: `"cf123:option_id_123,"`
- Multi select: `"cf123:option_id_1,option_id_2,"`
- Text: `"cf123:Hello World"`
- Number: `"cf456:42.5"`
- Boolean: `"cf789:true"`

Get option IDs with: `blue fields list --workspace <id> --detailed`

## Project Structure

```
cli/
├── cmd/                 # Cobra command definitions
│   ├── blue/            # main package — `go install` target
│   │   └── main.go      # Entry point
│   ├── root.go          # Root command, global config
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
├── common/              # Shared code
│   ├── auth.go          # GraphQL client & authentication
│   ├── types.go         # Shared type definitions
│   └── utils.go         # Utility functions
└── .env                 # API credentials (git ignored)
```

## License

Internal use only for Blue team.
