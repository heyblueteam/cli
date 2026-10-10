---
title: Export a public view to CSV
description: Read a published public view's rows as CSV text with the publicViewCsv query.
icon: Download
order: 0
---

A public view publishes a database's records through a read-only link. The `publicViewCsv` query returns the view's rows as one CSV file, delivered as the response string — there is no download URL and no file storage. The admin must have turned on CSV export for the view.

Anyone with the view's ID can call it. No Blue account and no API token. The ID is the capability.

## Request

Send the query to `https://api.blue.app/graphql`. No token headers, unless the view has a password (see [Access](#access)).

```graphql
query PublicViewCsv {
  publicViewCsv(publicViewId: "clxq3v7n0000108l0d5k2m9tp")
}
```

### Parameters

| Parameter      | Type      | Required | Description                                                    |
| -------------- | --------- | -------- | -------------------------------------------------------------- |
| `publicViewId` | `String!` | Yes      | The public view's ID (`PublicView.id`). See [Access](#access). |

### Where the view ID comes from

The share link carries slugs, not the ID: `https://blue.app/public/your-org/your-workspace/roadmap`. Resolve it with [publicViewData](#related-operations) — its response names the view's `publicViewId`. The public page itself works this way: it resolves the slugs, then calls this query.

## Response

The GraphQL response is the file content as one JSON string. Decoded, an export from a view that publishes every standard column plus two custom fields reads:

```text
Record ID,Title,Description,Due Date,Completed,Tags,Assignees,Status,Budget
clxq3v7n0000108l0d5k2m9tp,Design the launch page,Three layout options for review,2026-05-04 00:00,Yes,Website,Ana Silva,Done,1200
clxq3v7n0000208l0q8h4w6rj,Write the announcement,Blog post plus email draft,2026-05-11 00:00,No,"Launch, Website",Ana Silva,In progress,
clxq3v7n0000308l0j2f7t8vk,Prepare demo data,Import the sample table,,No,,,In progress,
clxq3v7n0000408l0s9g5n3wq,Visitor feedback,'=SUM(A1:A2),,Yes,,,Done,
clxq3v7n0000508l0c4d8x2hm,Refund follow-up,Check the invoice,,No,Finance,Maria Lopez,In progress,'-500
```

Line breaks are `\n` in the raw JSON. The file has no byte-order mark. A cell is double-quoted only when it contains a comma, a double quote, or a line break. A cell that starts with `=`, `+`, `-`, `@`, a tab, or a carriage return gets a `'` prefix — this stops Excel and Google Sheets from running the cell as a formula. Rows 4 and 5 of the example show the prefix.

## Columns

The view's visible-field allow-list decides the columns. An empty allow-list shows all of them.

| Column                       | When                                                 | Content                                                           |
| ---------------------------- | ---------------------------------------------------- | ----------------------------------------------------------------- |
| `Record ID`                  | Always                                               | The record's ID.                                                  |
| `Title`                      | Always                                               | The record title.                                                 |
| `Description`                | When the allow-list names `DESCRIPTION`, or is empty | Record description text.                                          |
| `Due Date`                   | When the allow-list names `DUE_DATE`, or is empty    | `YYYY-MM-DD HH:mm`; a range prints `start-end`; blank when unset. |
| `Completed`                  | Always                                               | `Yes` / `No`.                                                     |
| `Tags`                       | When the allow-list names `TAG`, or is empty         | Tag titles joined with `, `.                                      |
| `Assignees`                  | When the allow-list names `ASSIGNEE`, or is empty    | Full names joined with `, `.                                      |
| One per visible custom field | When the allow-list names the field, or is empty     | Header is the field name; cell per type (below).                  |

### Custom-field cells

| Type                                                                            | Cell content                    |
| ------------------------------------------------------------------------------- | ------------------------------- |
| `CHECKBOX`                                                                      | `Yes` / `No`.                   |
| `NUMBER`, `PERCENT`, `RATING`, `CURRENCY`                                       | The number.                     |
| `SELECT_SINGLE`                                                                 | The selected option's title.    |
| `SELECT_MULTI`                                                                  | Option titles joined with `, `. |
| `DATE`                                                                          | Date, or a `start-end` range.   |
| `ASSIGNEE`                                                                      | Names joined with `, `.         |
| Text types (`TEXT_SINGLE`, `TEXT_MULTI`, `EMAIL`, `URL`, `PHONE`, and the rest) | The plain text.                 |

## Limits

- Rows are sorted `createdAt_ASC` — oldest first. The sort is fixed; there is no sort argument.
- The export pages the view's rows 150 at a time for at most 67 pages: **10,050 rows**. Paging goes through `publicViewRecords`, so the same password gate, record filter, and visible-list allowlist apply to the export.
- When the view holds more rows than the cap, the file ends with a truncation row. `EXPORT TRUNCATED` sits in the Title column:

```text
,EXPORT TRUNCATED,The view holds more records than the 10050-row export cap; only the first 10050 are in this file.,,,,
```

- Rate limit: 2 requests per 60 seconds per client (identified by IP). Far below the other public-view reads (100/60s). See [Rate limits](/api/start-guide/rate-limits).

## Errors

| Code                    | When                                                                                                                                     |
| ----------------------- | ---------------------------------------------------------------------------------------------------------------------------------------- |
| `BAD_USER_INPUT`        | The view does not exist, or CSV export is not enabled for it. One message covers both: "CSV export is not enabled for this public view." |
| `UNAUTHENTICATED`       | The view has a password and the `Authorization` header is missing, or its token was not issued for this view.                            |
| `RATE_LIMITED`          | More than 2 requests in 60 seconds.                                                                                                      |
| `PUBLIC_VIEW_NOT_FOUND` | The view was deleted while the export was reading its pages. Rare.                                                                       |

## Access

- No account or token. The view's ID is the only credential; anyone who has it can export.
- The view decides what leaves: rows always pass the view's record filter and visible-list allowlist, columns pass the visible-field allow-list.
- **Password-protected views** need a token. Call `publicViewData` for the view and send the password in the `x-public-view-password` header — the response carries a `sessionToken`, valid for 1 hour. Send it as `Authorization: Bearer <sessionToken>` on this query.

## Related operations

| Query                      | Returns                  | One line                                                                                  |
| -------------------------- | ------------------------ | ----------------------------------------------------------------------------------------- |
| `publicViewRecords`        | `TodosResult!`           | Pages of live rows. `limit` must be 25 or 150; `skip` a multiple of `limit`, up to 10000. |
| `publicView`               | `PublicView!`            | One view's settings; needs a signed-in workspace member.                                  |
| `publicViewData`           | `PublicViewDataPayload!` | The published schema and settings by URL slug; issues the password session token.         |
| `publicViewRecordComments` | `[PublicComment!]!`      | The published comment thread for one record, as a visitor sees it.                        |

## Related

- [Pagination and filtering](/api/start-guide/pagination-and-filtering) — the paging rules this export follows internally.
- [Rate limits](/api/start-guide/rate-limits) — the 2-per-60-seconds ceiling.
- [Public views](/docs/workspaces/public-views) — create and configure the views.
