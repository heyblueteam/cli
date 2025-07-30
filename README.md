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

## 📋 Available Scripts

### 1. List Projects (`list-projects.go`)
Lists all projects in your Blue company.

```bash
# List with full details
go run auth.go list-projects.go

# List with just names and IDs
go run auth.go list-projects.go -simple
```

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
- `-simple`: Show only basic list information

### 4. Create Lists (`create-list.go`)
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
├── .env              # Your API credentials (git ignored)
├── .gitignore        # Git ignore file
├── go.mod            # Go module file
├── go.sum            # Go dependencies
├── auth.go           # Centralized authentication and GraphQL client
├── list-projects.go  # List all projects
├── create-project.go # Create new projects
├── get-lists.go      # Get lists in a project
├── create-list.go    # Create lists in a project
└── README.md         # This file
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
```

## 🛠️ Technical Details

### Architecture
- **auth.go**: Provides centralized authentication and GraphQL client
- All scripts use the shared `Client` from auth.go
- Environment variables are loaded from `.env` file
- GraphQL queries are embedded in each script

### GraphQL API
- Uses Blue's GraphQL API at `https://api.blue.cc/graphql`
- Authentication via custom headers:
  - `X-Bloo-Token-ID`
  - `X-Bloo-Token-Secret`
  - `X-Bloo-Company-ID`

### Position System
- Lists use a floating-point position system
- Standard increment is 65535 between lists
- Allows for reordering without updating all positions

## 🚧 Limitations
- Maximum 50 lists per project
- Project names are automatically trimmed
- All scripts require authentication

## 🔮 Future Scripts
- `create-records.go` - Create records/todos within lists
- `bulk-demo.go` - Create complete demo projects from templates

## 🤝 Contributing
When adding new scripts:
1. Use the centralized auth.go for all API calls
2. Follow the existing command-line flag patterns
3. Include both simple and detailed output options where applicable
4. Update this README with usage examples

## 📝 License
Internal use only for Blue team demonstrations.