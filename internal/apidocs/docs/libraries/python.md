---
title: Python
description: The bluepm Python SDK wraps the Blue GraphQL API with typed, schema-generated models built on sgqlc.
icon: Code
order: 1
---

[`bluepm`](https://pypi.org/project/bluepm/) is the official Python SDK for the Blue API. It is generated from the GraphQL schema with [`sgqlc`](https://github.com/profusion/sgqlc), so it ships a typed `schema` object that covers the full API surface — every query, mutation, object, and field — with autocompletion in any modern IDE.

The SDK gives you two ways to call the API:

- **Convenience helpers** for the most common reads. Two ship today: `get_project_list()` and `get_todo_lists()`.
- **The escape hatch** — `query()`, `mutation()`, and `execute()` — which lets you build and run any operation the schema supports. This is how you reach everything the helpers don't cover.

Requires Python 3.6 or newer.

## Install

```bash
pip install bluepm
```

```python
from bluepm import BlueAPIClient
```

## Authentication

The client authenticates with a Blue personal access token. You can supply credentials either by setting environment variables or by passing them directly to the constructor — constructor arguments take priority when both are present.

`token_id`, `secret_id`, and `company_id` are required. The constructor raises `ValueError` if any of them is missing (from both the argument and its environment variable). `project_id` is optional and only needed for project-scoped operations such as custom fields.

| Constructor arg | Environment variable | Required | Request header        | Description                                                          |
| --------------- | -------------------- | -------- | --------------------- | -------------------------------------------------------------------- |
| `token_id`      | `BLUE_TOKEN_ID`      | Yes      | `blue-token-id`     | The token ID from your personal access token.                        |
| `secret_id`     | `BLUE_SECRET_ID`     | Yes      | `blue-token-secret` | The secret half of the token (the value prefixed with `pat_`).       |
| `company_id`    | `BLUE_COMPANY_ID`    | Yes      | `blue-org-id`   | The organization the token operates within. Accepts an ID or slug.   |
| `project_id`    | `BLUE_PROJECT_ID`    | No       | `blue-workspace-id`   | A default workspace for project-scoped calls. Accepts an ID or slug. |

See [Authentication](/api/start-guide/authentication) for how to create a token ID and secret.

### Environment variables

```bash
export BLUE_TOKEN_ID="YOUR_TOKEN_ID"
export BLUE_SECRET_ID="YOUR_TOKEN_SECRET"
export BLUE_COMPANY_ID="YOUR_ORG_ID"
export BLUE_PROJECT_ID="YOUR_PROJECT_ID"  # optional
```

```python
from bluepm import BlueAPIClient

client = BlueAPIClient()  # reads the BLUE_* environment variables
```

### Constructor arguments

```python
from bluepm import BlueAPIClient

client = BlueAPIClient(
    token_id="YOUR_TOKEN_ID",
    secret_id="YOUR_TOKEN_SECRET",
    company_id="YOUR_ORG_ID",
    project_id="YOUR_PROJECT_ID",  # optional
)
```

<Callout variant="warning" title="Keep your secret safe">

The secret half of your token grants full API access to your organization. Load it from an environment variable or secret manager — never commit it to source control.

</Callout>

## Convenience helpers

### get_project_list

Lists the workspaces the token can access. Workspaces are `Project` objects in the API, so this calls the [`workspaceList`](/api/workspaces/list-workspaces) query under the hood.

```python
def get_project_list(self, company_ids=None) -> list
```

`company_ids` is an optional **list** of organization IDs. When omitted, it defaults to `[client.company_id]`. The helper pre-selects each project's `id` and `name`.

```python
projects = client.get_project_list()              # workspaces in the client's organization
projects = client.get_project_list(company_ids=["company_123"])

for project in projects:
    print(project.id, project.name)
```

Each item is a typed `Project` exposing the selected fields:

```python
project.id    # "clm4n8qwx000008l0g4oxdqn7"
project.name  # "Product Roadmap"
```

### get_todo_lists

Lists the lists inside a workspace. Lists are `TodoList` objects in the API, so this calls the [`todoLists`](/api/workspaces/lists) query.

```python
def get_todo_lists(self, project_id) -> list
```

`project_id` is required (it accepts a workspace ID or slug). The helper pre-selects each list's `id`, `title`, and `rank`.

```python
todo_lists = client.get_todo_lists("project_123")

for todo_list in todo_lists:
    print(todo_list.rank, todo_list.title)
```

```python
todo_list.id        # "clm4n8qwx000008l0g4oxdqn7"
todo_list.title     # "In Progress"
todo_list.rank      # "a1"
```

## Building any operation

For anything the helpers don't cover, build the operation yourself. `client.query()` returns an `sgqlc` `Operation` over the `Query` root, `client.mutation()` returns one over `Mutation`, and `client.execute(op)` runs it and merges the response back onto the operation as typed objects.

The pattern is: start an operation, select the fields you want, run it, then read the results off the returned object.

```python
from bluepm import BlueAPIClient

client = BlueAPIClient()

# 1. Build a query against the schema
op = client.query()
projects = op.project_list(
    filter={"companyIds": [client.company_id]},
    take=10,
)
projects.items.id()
projects.items.name()
projects.page_info.total_items()
projects.page_info.has_next_page()

# 2. Execute it
result = client.execute(op)

# 3. Read typed results
for project in result.project_list.items:
    print(project.id, project.name)

print("total:", result.project_list.page_info.total_items)
```

Field and argument names use Python's `snake_case` — `project_list` maps to the GraphQL `workspaceList` field, `page_info` to `pageInfo`, `total_items` to `totalItems`. The full set of queryable fields and their arguments is the Blue GraphQL schema; browse the operation reference (for example [List Workspaces](/api/workspaces/list-workspaces)) for what each field returns.

Mutations work the same way through `client.mutation()`:

```python
op = client.mutation()
op.create_todo(input={"todoListId": "list_123", "title": "New task"}).id()
result = client.execute(op)

print(result.create_todo.id)
```

## Response shape

`execute()` returns the operation with the server's data merged in, so you read fields directly as Python attributes — there is no raw JSON to parse. The underlying GraphQL response for the `project_list` example above looks like this:

```json
{
  "data": {
    "workspaceList": {
      "items": [
        { "id": "clm4n8qwx000008l0g4oxdqn7", "name": "Product Roadmap" },
        { "id": "clm4n8qwx000008l0g4oxdqn8", "name": "Customer Onboarding" }
      ],
      "pageInfo": {
        "totalItems": 2,
        "hasNextPage": false
      }
    }
  }
}
```

## Errors

| Exception      | When                                                                                                                                                     |
| -------------- | -------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `ValueError`   | The constructor is missing `token_id`, `secret_id`, or `company_id` (from both the argument and its environment variable).                               |
| `RuntimeError` | An executed operation fails — for example invalid credentials, a malformed selection set, or an API error. The original error is wrapped in the message. |

```python
from bluepm import BlueAPIClient

try:
    client = BlueAPIClient()  # raises ValueError if BLUE_* vars are unset
    todo_lists = client.get_todo_lists("project_123")
except ValueError as e:
    print("Missing credentials:", e)
except RuntimeError as e:
    print("Request failed:", e)
```

## Related

- [Authentication](/api/start-guide/authentication) — create a token ID and secret
- [Making Requests](/api/start-guide/making-requests) — the raw GraphQL endpoint and headers
- [List Workspaces](/api/workspaces/list-workspaces) — the `workspaceList` query `get_project_list` wraps
- [Lists](/api/workspaces/lists) — the `todoLists` query `get_todo_lists` wraps
