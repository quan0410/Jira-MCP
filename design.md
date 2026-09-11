# Design — jira-mcp

## Class

**Tool-server** (Streamable HTTP MCP), matching `Bright-Data-MCP`'s shape — not a
WS hub relay, not embedded inside `datumbridge-integrations`.

## Language

Go, matching the platform pattern established by `Bright-Data-MCP` and the
"one provider = one service" decision in
`doc/MCP_TOOL_REFERENCE_ARCHITECTURE.md` (BR1).

## Upstream

- Jira Cloud core REST API v3 — `https://api.atlassian.com/ex/jira/{cloudId}/rest/api/3`
- Jira Software Agile REST API 1.0 — `https://api.atlassian.com/ex/jira/{cloudId}/rest/agile/1.0`
- `{cloudId}` comes from the injected credential bundle (`Credentials.CloudID`),
  never from a tool argument — one connection is bound to exactly one Jira Cloud
  site (enforced upstream by `datumbridge-integrations`' partial unique index on
  `(provider="jira", cloud_id)`).

## Auth

- Primary (production): vault `credentials_json` injected per tool call via
  `POST /internal/v1/inject` on `datumbridge-integrations`.
- Fallback: `JIRA_ACCESS_TOKEN` + `JIRA_CLOUD_ID` env vars (local/dev only).
- `credentials_path` under `JIRA_CREDENTIALS_DIR` supported as a third option,
  mirroring Bright-Data-MCP's path-jail pattern, for deployments that mount a
  credentials file instead of injecting JSON inline.
- Refresh is explicitly **not** this server's responsibility — see
  `doc/MCP_TOOL_REFERENCE_ARCHITECTURE.md`'s BR9/BR10. A stale/expired token
  surfaces as a Jira API 401, which this server reports as a tool error rather
  than retrying with a refreshed token it has no way to obtain.

## Tools

27 tools across 7 groups (`internal/tools/*_tools.go`): issues, comments,
transitions, worklogs, projects, users, agile. Full scope by default per BR6;
`JIRA_TOOLS` env narrows to an explicit allowlist if an operator wants a smaller
surface for a specific deployment.

## Rich text

Jira Cloud v3 requires Atlassian Document Format (ADF) for description/comment
bodies, not plain strings. Tool inputs stay plain text for agent ergonomics;
`jira.PlainTextToADF` wraps them in a minimal single-paragraph ADF `doc` node
before the request goes out. This does not support rich formatting (mentions,
lists, code blocks) — a caller that needs that constructs the ADF object
directly and passes it through `fields` on `jira_update_issue`/`jira_create_issue`.

## Non-goals

Jira Service Management, Confluence (tracked separately — item #5 in the
rollout), attachment upload, admin/webhook configuration, OAuth or token
storage of any kind (that stays in `datumbridge-integrations`).

## Risks

- Jira API shape drift (Atlassian has migrated search endpoints before — this
  server targets the current `/rest/api/3/search/jql` cursor-paginated search,
  not the deprecated `/rest/api/3/search`).
- ADF auto-wrapping loses rich formatting for any caller relying on plain-text
  description round-tripping through `jira_get_issue` → edit → `jira_update_issue`
  (the returned description is full ADF JSON, not the plain string that was sent).
