# Business Rules

Inherits BR1–BR13 from `doc/MCP_TOOL_REFERENCE_ARCHITECTURE.md` (architecture,
credential contract, security baseline, refresh responsibility). Jira-specific
rules below.

- **BR-J1 (identity):** every action runs as the connecting Weaver user's own
  Jira identity (their OAuth token) — there is no shared service account. Two
  Weaver users with access to the same Jira site each act as themselves.
- **BR-J2 (single site):** a connection is bound to exactly one Jira Cloud
  `cloud_id`. `jira-mcp` never queries or switches sites — that's decided once,
  at OAuth-connect time, in `datumbridge-integrations`.
- **BR-J3 (rich text):** `description`/comment `body` inputs are plain text,
  auto-converted to a single-paragraph ADF node. Callers needing real formatting
  (lists, mentions, code blocks) pass a full ADF object via `fields` on
  `jira_create_issue`/`jira_update_issue` instead of the `description`
  convenience field.
- **BR-J4 (search pagination):** `jira_search_issues` uses cursor-based paging
  (`next_page_token`), not `startAt` — Jira Cloud deprecated offset paging on
  `/search`. Out-of-range `max_results` (≤0 or >100) resets to the 50 default
  rather than clamping to 100.
- **BR-J7 (search default fields):** `/search/jql` returns bare `{"id": "..."}`
  — no `key`, no fields at all — when the caller omits `fields`, unlike the
  deprecated `/search`'s useful default set (verified against a live Jira Cloud
  site). `jira_search_issues` defaults to `["key","summary","status",
  "issuetype","assignee","priority","created","updated"]` when the caller
  doesn't specify `fields`. `jira_get_issue` needs no such default — `GET
  /issue/{key}` already returns a full field set on its own.
- **BR-J5 (sprint update is partial):** `jira_update_sprint` only sends the
  fields the caller provided (Jira Agile API's partial-update semantics via
  `POST /sprint/{id}`) — it will not silently clear fields the caller omitted.
- **BR-J6 (no client-side refresh):** a 401 from Jira (expired/revoked token) is
  returned as a tool error, not retried — see the reference doc's BR9/BR10. The
  caller (agent) is expected to prompt the user to reconnect Jira in Weaver.
