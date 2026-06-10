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
API_URL=https://api.blue.app/graphql
AUTH_TOKEN=your_personal_access_token
CLIENT_ID=your_client_id
COMPANY_ID=your_company_slug
```

**Getting Your Credentials:**
1. **Client ID & Auth Token**: Account Settings → API → Generate Token
2. **Company ID**: Your org slug from the URL (e.g., `acme` from `blue.app/org/acme`)

## Commands

```
blue [command]

Available Commands:
  api         (graphql, gql) Run raw Blue API requests
  activity            Show recent activity
  bootstrap           Bootstrap workspaces from JSON
  company             Manage known companies
  workspaces   (ws)    Manage workspaces
  records      (rec)   Manage records
  reports              Manage reports
  saved-views  (views) Manage saved views
  lists                Manage lists
  tags                 Manage tags
  fields       (cf)    Manage custom fields
  automations  (auto)  Manage automations
  checklists           Manage checklists
  comments             Manage comments
  users                Manage users
  dependencies (deps)  Manage record dependencies
  documents    (docs)  Manage documents and wiki pages
  doctor               Check CLI configuration and API access
  domains              Manage custom domains and email settings
  exports              Queue CSV exports
  files                Manage files
  forms        (form)  Manage forms
  ids          (id)    Resolve Blue names to IDs
  open                 Open Blue pages in a browser
  search               Search records by name
  dashboards   (dash)  Manage dashboards
  charts               Manage dashboard charts
  webhooks     (wh)    Manage webhooks
  whoami               Show authenticated identity and context
  completion           Generate shell completions
  version              Print version information
```

Use `blue <command> --help` for details on any command.

### API

```bash
blue api query --raw 'query { __typename }'
blue api query --file query.graphql --variables '{"id":"workspace_123"}'
blue api schema
blue api schema --introspect
blue api docs --print
```

### Activity

```bash
blue activity
blue activity --workspace <id> --since 7d
blue activity --workspace <id> --category CREATE_TODO,CREATE_COMMENT
blue activity --format json
```

### Doctor

```bash
blue doctor
blue doctor --workspace <id-or-slug>
```

### Whoami

```bash
blue whoami
blue whoami --format json
```

### Bootstrap

```bash
blue bootstrap template > workspace.json
blue bootstrap apply --file workspace.json --confirm
blue bootstrap export --workspace <id> > workspace.json
```

### Domains & Email

```bash
blue domains domains list
blue domains domains create --name app.example.com --type APPLICATION
blue domains domains verify --name app.example.com
blue domains domains update --domain <id> --name app2.example.com
blue domains domains delete --domain <id> --confirm
blue domains smtp list
blue domains smtp verify --host smtp.example.com --port 587 --username user --password pass
blue domains templates list
blue domains templates get --type INVITATION
blue domains templates test --template <id> --email you@example.com
```

### Company

```bash
blue company list
blue company add <slug>
blue company use <slug>
blue company remove <slug>
blue workspaces list --company <slug> --simple
```

### IDs

```bash
blue ids workspace --search CRM
blue ids field --workspace <id> --search Priority
blue ids list --workspace <id>
blue ids tag --workspace <id> --format csv
blue ids user --search alex --format json
blue ids record --workspace <id> --search "Launch plan"
```

### Open

```bash
blue open https://blue.app/org/acme
blue open workspace <id-or-slug>
blue open record <id> --workspace <id-or-slug>
blue open form <id> --workspace <id-or-slug>
blue open document <id> --workspace <id-or-slug>
blue open dashboard <id>
blue open report <id>
blue open files --workspace <id-or-slug>
blue open folder <id> --workspace <id-or-slug>
blue open record <id> --workspace <id-or-slug> --print
```

### Search

```bash
blue search "launch" --workspace <id>
blue search "invoice" --workspace <id> --format json
blue search "bug" --workspace <id> --done false --limit 50
```

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

### Reports

```bash
blue reports list --simple
blue reports get --report <id>
blue reports create --title "Open work" --workspaces "ws1,ws2" --filter-json '{"done":false}'
blue reports update --report <id> --title "New title"
blue reports share --report <id> --users "user1:EDITOR,user2:VIEWER"
blue reports data --report <id> --limit 50
blue reports aggregate --report <id> --field field_123:number
blue reports refresh --report <id>
blue reports duplicate --report <id> --title "Working copy"
blue reports export --report <id>
blue reports delete --report <id> --confirm
```

### Saved Views

```bash
blue saved-views list --workspace <id> --simple
blue saved-views get --view <id>
blue saved-views update --view <id> --name "Sprint Board"
blue saved-views update --view <id> --shared true --config-json '{"searchQuery":"launch"}'
blue saved-views apply --view <id>
blue saved-views delete --view <id> --confirm
```

### Lists

```bash
blue lists list --workspace <id> --simple
blue lists create --workspace <id> --names "To Do,In Progress,Done"
blue lists update --list <id> --workspace <id> --title "Backlog" --locked true
blue lists update --list <id> --workspace <id> --color "#ff0000"
blue lists delete --workspace <id> --list <id> --confirm
```

### Tags

```bash
blue tags list --workspace <id>
blue tags create --workspace <id> --title "Bug" --color "#ff0000"
blue tags update --tag <id> --color "#0066ff"
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
blue files inventory > files.csv                            # Company-wide CSV
blue files inventory --workspace <id> --output files.csv
blue files download                                          # Interactive mode
blue files download --use-env --output "backup.zip" --parallel 10
```

### Documents

```bash
blue documents list --workspace <id> --simple
blue documents list --workspace <id> --wiki true
blue documents get --document <id> --content
blue documents create --workspace <id> --title "Runbook" --content '<h1>Runbook</h1>'
blue documents create --workspace <id> --title "Handbook" --wiki --content-file handbook.html
blue documents update --document <id> --title "New title"
blue documents delete --document <id> --confirm
```

### Exports

```bash
blue exports records --workspace <id>
blue exports records --workspace <id> --done false --q launch --assignees "user1,user2"
blue exports records --workspace <id> --filter-json '{"hasDueDate":true}'
blue exports report --report <id>
blue exports chart --chart <id> --filter-json '{"showCompleted":false}'
blue exports template --workspace <id>
```

### Dashboards

```bash
blue dashboards list --simple
blue dashboards create --title "Sales Dashboard"
blue dashboards create --title "Sprint Metrics" --workspace <id>
blue dashboards get --dashboard <id>
blue dashboards share --dashboard <id> --users "user1:EDITOR,user2:VIEWER"
blue dashboards delete --dashboard <id> --confirm
```

### Charts

```bash
blue charts list --dashboard <id>
blue charts create --dashboard <id> --type STAT --title "Total Records" --workspace <id> --function COUNT
blue charts create --dashboard <id> --type BAR --title "By Assignee" --workspace <id> --group-by ASSIGNEE --function COUNT
blue charts create --dashboard <id> --type PIE --title "By Status" --workspace <id> --group-by TODO_STATUS --function COUNT
blue charts recalculate --charts "chart1,chart2"
blue charts delete --chart <id> --confirm
```

### Webhooks

```bash
blue webhooks list --simple
blue webhooks get --webhook <id>
blue webhooks create --name "Production sync" --url https://example.com/webhooks/blue --events TODO_CREATED,COMMENT_CREATED
blue webhooks update --webhook <id> --enabled false
blue webhooks disable --webhook <id>
blue webhooks delete --webhook <id> --confirm
blue webhooks events
blue webhooks verify-signature --secret <secret> --signature <hex> --body-file payload.json
blue webhooks listen --port 8080 --secret <secret>
```

### Forms

Field types: `title`, `description`, `tags`, `startedAt`, `duedAt`, `custom`. Only `custom` carries a `customField` ID.

```bash
# List / get (--workspace required for get)
blue forms list --workspace <ID> --simple
blue forms get --form <ID> --workspace <ID>

# Create — minimal
blue forms create -w <ws> --title "Contact us"

# Create — full (composite: createForm -> updateForm -> upsertFormField)
blue forms create -w <ws> --title "Lead intake" \
  --primary-color "#0066ff" --theme dark --hide-branding --active \
  --submit-text "Send" --redirect-url "https://example.com/thanks" \
  --list <list-id> \
  --field "type=title;name=Full name;required=true;position=1000" \
  --field "type=custom;customField=cf_xxx;name=Budget;required=true;position=2000"

# Create — fields defined in JSON
blue forms create -w <ws> --title "Lead intake" --fields-file ./form-fields.json

# Update / copy / delete (copy + delete require --workspace)
blue forms update --form <ID> --active true
blue forms update --form <ID> --primary-color "#ff0000"
blue forms copy --form <ID> --workspace <ID>
blue forms delete --form <ID> --workspace <ID> --confirm

# Public submit URL (default https://blue.app/forms/<uid>)
blue forms url --form <ID> --workspace <ID>
blue forms url --form <ID> --workspace <ID> --base-url https://forms.acme.com

# Field-level ops (all require --workspace)
blue forms fields list   --form <ID> --workspace <ID>
blue forms fields add    --form <ID> --workspace <ID> --type title --name "Full name" --required --position 1000
blue forms fields add    --form <ID> --workspace <ID> --type custom --custom-field <cf-id> --name "Priority" --required
blue forms fields update --field <ff-id> --form <ID> --workspace <ID> --name "New label" --position 1500
blue forms fields delete --field <ff-id> --workspace <ID> --confirm
```

**`--field` inline syntax:** `key=value` pairs separated by `;`. Keys: `type`, `customField`, `name`, `placeholder`, `position`, `required`, `hidden`, `addToDescription`, `extraInfo`.

**`--fields-file` JSON shape:**

```json
[
  { "field": "title",       "name": "Full name",       "required": true,  "position": 1000 },
  { "field": "description", "name": "Project details", "placeholder": "Tell us more", "position": 2000 },
  { "field": "custom", "customFieldId": "cf_xxx", "name": "Budget", "required": true, "position": 3000 }
]
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
│   ├── api/             # blue api *
│   ├── bootstrap/       # blue bootstrap *
│   ├── workspaces/      # blue workspaces *
│   ├── records/         # blue records *
│   ├── reports/         # blue reports *
│   ├── savedviews/      # blue saved-views *
│   ├── lists/           # blue lists *
│   ├── tags/            # blue tags *
│   ├── fields/          # blue fields *
│   │   ├── options/     # blue fields options *
│   │   └── groups/      # blue fields groups *
│   ├── automations/     # blue automations *
│   ├── charts/          # blue charts *
│   ├── checklists/      # blue checklists *
│   │   └── items/       # blue checklists items *
│   ├── company/         # blue company *
│   ├── comments/        # blue comments *
│   ├── dashboards/      # blue dashboards *
│   ├── users/           # blue users *
│   ├── webhooks/        # blue webhooks *
│   ├── dependencies/    # blue dependencies *
│   ├── documents/       # blue documents *
│   ├── doctor/          # blue doctor
│   ├── domains/         # blue domains *
│   ├── exports/         # blue exports *
│   ├── files/           # blue files *
│   └── forms/           # blue forms *
├── common/              # Shared code
│   ├── auth.go          # GraphQL client & authentication
│   ├── types.go         # Shared type definitions
│   └── utils.go         # Utility functions
└── .env                 # API credentials (git ignored)
```

## License

Internal use only for Blue team.
