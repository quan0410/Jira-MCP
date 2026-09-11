# API Specification

This server exposes only the MCP Streamable HTTP surface — there is no separate
REST API of its own.

## `GET /health`

Liveness only. No upstream call, no credential check.

```json
{"status": "ok", "service": "jira-mcp"}
```

## `POST /mcp`

JSON-RPC 2.0 over HTTP, methods: `initialize`, `notifications/initialized`,
`tools/list`, `tools/call`. See `internal/mcp/http.go` for the exact envelope —
identical to `Bright-Data-MCP`'s, since the MCP protocol layer is
provider-agnostic.

`tools/call` args always accept `credentials_json` (preferred, injected by the
Tool Registry) or `credentials_path` (jailed-file alternative); see each tool's
`inputSchema` from `tools/list`, or `docs/technical/MCP_TOOL_SPECIFICATION.md`.

## Upstream (Jira Cloud)

All calls go to `https://api.atlassian.com/ex/jira/{cloudId}/rest/{api/3|agile/1.0}/...`
with `Authorization: Bearer <access_token>`. See Atlassian's own REST API v3 and
Agile 1.0 documentation for request/response shapes — this server passes Jira's
JSON responses through largely unmodified (`rawJSONResult` in
`internal/tools/issue_tools.go`), so Jira's own field names appear as-is in tool
results.
