# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go module for building demo projects using the Blue GraphQL API. It consists of individual command-line utilities that share a centralized authentication module.

## Development Commands

### Running Scripts
All scripts follow this pattern:
```bash
go run auth.go <script-name>.go [flags]
```

### Available Scripts & Usage
```bash
# List projects (first 20)
go run auth.go list-projects.go -simple

# Create project with options
go run auth.go create-project.go -name "Demo" -color blue -icon rocket -category ENGINEERING

# Get lists in a project
go run auth.go get-lists.go -project PROJECT_ID -simple

# Create lists in a project
go run auth.go create-list.go -project PROJECT_ID -names "To Do,In Progress,Done"

# List tags in a project
go run auth.go list-tags.go -project PROJECT_ID
```

### Dependencies
```bash
go mod tidy  # Install/update dependencies
```

## Architecture

### Centralized Authentication (`auth.go`)
All scripts import and use the shared authentication module which provides:
- `Client` struct with GraphQL request method
- Environment variable loading from `.env`
- Standard HTTP headers for Blue API authentication
- 30-second timeout for requests

### GraphQL Integration Pattern
Each script:
1. Imports the auth module
2. Creates a client instance
3. Defines GraphQL query/mutation as a string
4. Makes requests using `client.Request()`
5. Unmarshals JSON response into typed structs

### Required Environment Variables
The `.env` file must contain:
```
API_URL=https://api.blue.cc/graphql
AUTH_TOKEN=your_personal_access_token
CLIENT_ID=your_client_id
COMPANY_ID=your_company_slug
```

## Planned Features (from plan.md)

To implement:
- Create tags
- Add tags to records
- Create custom fields (all types except reference/lookup)
- Create custom field groups
- Create automations
- Create custom user roles
- Create record (simple: name + list)
- Create record (full: name + list + fields)
- Feature toggles for projects

## Implementation Guidelines

When adding new scripts:
1. Use the centralized `auth.go` module for all API calls
2. Follow the existing command-line flag patterns using Go's `flag` package
3. Include both `-simple` and detailed output options where applicable
4. Define proper struct types for GraphQL responses
5. Handle errors consistently with proper context
6. Update README.md with usage examples

## Known Limitations

- Project listing limited to first 20 results (pagination not implemented)
- Maximum 50 lists per project
- No test suite or linting configuration
- Individual script execution (no unified CLI)

## GraphQL API Details

- Endpoint: `https://api.blue.cc/graphql`
- Authentication Headers:
  - `X-Bloo-Token-ID`: Client ID
  - `X-Bloo-Token-Secret`: Auth Token
  - `X-Bloo-Company-ID`: Company slug
- Request timeout: 30 seconds
- All requests use POST method with JSON body