---
title: Delete Saved View
description: Permanently delete a saved view from a Blue workspace and clear any default-view references that pointed to it.
icon: Trash2
order: 3
---

Use the `deleteSavedView` mutation to permanently remove a saved view from a workspace. Saved views are `SavedView` objects in the API. Before the view is deleted, Blue clears every default-view reference that pointed to it — both the workspace default (`Project.defaultViewId`) and every user's personal default (`ProjectUser.defaultViewId`) — so no record layout is left pointing at a missing view.

Deleting a view publishes a real-time `DELETED` event to anyone subscribed to the workspace's views (see [Related](#related)). This action is permanent and cannot be undone.

## Request

```graphql
mutation DeleteSavedView {
  deleteSavedView(id: "view_123") {
    success
    operationId
  }
}
```

Send the request to `https://api.blue.app/graphql` with your authentication headers:

```bash
curl https://api.blue.app/graphql \
  -H "Content-Type: application/json" \
  -H "X-Bloo-Token-ID: YOUR_TOKEN_ID" \
  -H "X-Bloo-Token-Secret: YOUR_TOKEN_SECRET" \
  -H "X-Bloo-Company-ID: YOUR_COMPANY_ID" \
  -d '{"query":"mutation { deleteSavedView(id: \"view_123\") { success operationId } }"}'
```

Header names are case-insensitive.

## Parameters

| Parameter | Type      | Required | Description                      |
| --------- | --------- | -------- | -------------------------------- |
| `id`      | `String!` | Yes      | ID of the `SavedView` to delete. |

## Response

```json
{
  "data": {
    "deleteSavedView": {
      "success": true,
      "operationId": "clm4n8qwx000008l0g4oxdqn7"
    }
  }
}
```

### Returns

`deleteSavedView` returns a `MutationResult`.

| Field         | Type       | Description                                                                                                                                                                                           |
| ------------- | ---------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `success`     | `Boolean!` | `true` when the view was deleted. `false` when deletion failed after authorization (see the note below).                                                                                              |
| `operationId` | `String`   | Identifier for this mutation. Use it to correlate the resulting real-time event on the `subscribeToSavedView` subscription so a client that triggered the delete can skip re-applying its own update. |

<Callout variant="warning" title="Always check success">

Permission and lookup failures (`SAVED_VIEW_NOT_FOUND`, `FORBIDDEN`) are thrown as GraphQL errors before any deletion is attempted. But if the delete itself fails *after* authorization succeeds — for example a database error while clearing default references — the mutation returns `{ "success": false }` instead of an error. Treat `success: false` as a real outcome and check it on every call; do not assume a missing `errors` array means the view is gone.

</Callout>

## Errors

| Code                   | When                                                                                                          |
| ---------------------- | ------------------------------------------------------------------------------------------------------------- |
| `SAVED_VIEW_NOT_FOUND` | No view with that `id` exists in a workspace you can access. Message: `Saved view was not found.`             |
| `FORBIDDEN`            | You are not allowed to delete this view (see [Permissions](#permissions)). Message: `You are not authorized.` |

```json
{
  "errors": [
    {
      "message": "You are not authorized.",
      "extensions": { "code": "FORBIDDEN" }
    }
  ]
}
```

## Permissions

Who may delete a view depends on whether it is shared and who created it:

| View                                  | Who can delete                                                                |
| ------------------------------------- | ----------------------------------------------------------------------------- |
| Personal view you created             | You only                                                                      |
| Shared view you created               | You, or any workspace `OWNER`/`ADMIN`                                         |
| Shared view created by someone else   | Workspace `OWNER`/`ADMIN` only                                                |
| Personal view created by someone else | Nobody — even an `OWNER`/`ADMIN` cannot delete another member's personal view |

A workspace admin's elevated rights apply to **shared** views only. A personal (non-shared) view can be deleted exclusively by its creator. Attempting to delete a view you may not modify raises `FORBIDDEN`; if you cannot see the view's workspace at all, the API returns `SAVED_VIEW_NOT_FOUND` so the view's existence is not leaked.

## Related

- [List saved views](/api/saved-views/list-saved-views)
- [Create a saved view](/api/saved-views)
- [Update a saved view](/api/saved-views/update-saved-view)
- [Real-time subscriptions](/api/realtime/project-subscriptions) — `subscribeToSavedView` delivers the `DELETED` event this mutation publishes.
