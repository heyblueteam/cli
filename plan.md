# Blue Demo Builder - Implementation Plan

## Overview
Build a set of Go scripts to interact with Blue's GraphQL API for creating demo projects programmatically.

## Authentication Details
- **API URL**: `https://api.blue.cc/graphql`
- **Auth Token**: `pat_73bb4f43a75748cfa15703460e1b8a90`
- **Client ID**: `21dbe637fc8f4c979cc9f79b4758f3bb`
- **Company Name**: `heyblueteam`
- **Required Headers**:
  - `X-Bloo-Token-ID`
  - `X-Bloo-Token-Secret`
  - `X-Bloo-Company-ID`

## Project Structure
```
/demo-builder/
├── plan.md
├── go.mod
├── go.sum
├── cmd/
│   ├── create-project/
│   │   └── main.go
│   ├── create-list/
│   │   └── main.go
│   └── create-records/
│       └── main.go
├── pkg/
│   ├── client/
│   │   └── client.go
│   └── models/
│       └── models.go
└── scripts/
    └── demo.sh
```

## Implementation Steps

### Phase 1: Core Infrastructure
- **Create GraphQL client wrapper**
  - Handle authentication headers
  - Implement request/response handling
  - Error handling and retries
  - Rate limiting awareness

### Phase 2: API Research
- **Discover GraphQL mutations**
  - Project creation mutation
  - List creation mutation (with ordering)
  - Record/Todo creation mutation
  - Query project structure

### Phase 3: Individual Scripts
- **create-project script**
  - Accept project name as parameter
  - Create project via GraphQL mutation
  - Return project ID for chaining

- **create-list script**
  - Accept project ID and list names
  - Support ordering (left to right)
  - Create multiple lists in sequence
  - Return list IDs

- **create-records script**
  - Accept list ID and record data
  - Support bulk creation
  - Handle positioning

### Phase 4: Integration Script
- **Master demo script**
  - Orchestrate all three operations
  - Accept JSON or YAML config
  - Support templates for common demo scenarios

## Technical Decisions

### Go Package Structure
- **cmd/**: Individual executable commands
- **pkg/client/**: Shared GraphQL client
- **pkg/models/**: Data structures for API

### Dependencies
- `github.com/machinebox/graphql` - GraphQL client
- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration

### Configuration Options
- Environment variables for credentials
- Command-line flags for parameters
- Config file support for complex scenarios

## Usage Examples

### Individual Commands
```bash
# Create a project
./create-project --name "Demo Project"

# Create lists in order
./create-list --project-id "xxx" --names "To Do,In Progress,Done"

# Create records
./create-records --list-id "xxx" --records "Task 1,Task 2,Task 3"
```

### Integrated Demo Builder
```bash
# Run full demo setup
./demo-builder create --config demo-template.yaml
```

### Config Template Example
```yaml
project:
  name: "Sprint Planning Demo"
lists:
  - name: "Backlog"
  - name: "Sprint 1"
  - name: "In Progress"
  - name: "Done"
records:
  "Backlog":
    - "User authentication"
    - "Dashboard design"
  "Sprint 1":
    - "Setup CI/CD"
    - "Database schema"
```

## Error Handling
- Validate authentication before operations
- Handle API rate limits gracefully
- Provide clear error messages
- Support retry mechanisms

## Future Enhancements
- WebSocket subscriptions for real-time updates
- Batch operations for performance
- Template library for common scenarios
- Interactive mode for Claude integration
- Undo/rollback functionality

## Testing Strategy
- Unit tests for GraphQL client
- Integration tests with API
- Mock server for development
- Example demo scenarios

## Documentation
- README with setup instructions
- API operation examples
- Troubleshooting guide
- Claude integration guide