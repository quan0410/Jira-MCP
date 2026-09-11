# Jira MCP Server

Go MCP **tool-server** for Jira Cloud — item #1 of the rollout in
[`doc/MCP_TOOL_REFERENCE_ARCHITECTURE.md`](../doc/MCP_TOOL_REFERENCE_ARCHITECTURE.md).
Built to that document's pattern (mirrors [`Bright-Data-MCP`](../Bright-Data-MCP)):
own Go service, own repo, credentials pulled from `datumbridge-integrations`'
vault via `POST /internal/v1/inject`, no OAuth or token storage of its own.

**mcpServer id:** `jira-mcp`

## Scope

Full core issue-tracking + Agile (boards/sprints) coverage — the "full tool scope"
default from the reference doc's BR6, not a minimal subset:

- **Issues:** get, search (JQL via `/search/jql`), create, update, delete
- **Comments:** list, add, update, delete
- **Transitions:** list available transitions, execute one (with optional comment)
- **Worklogs:** list, add
- **Projects:** list, get, list issue types (for `jira_create_issue`)
- **Users:** get connected identity (`myself`), search assignable users
- **Agile (Jira Software):** boards, sprints (list/get/create/update), backlog,
  sprint issues, move issues into a sprint

Not covered (explicit non-goals for this first release): Jira Service Management,
webhook/admin configuration, attachments upload, Confluence (separate provider,
separate future `confluence-mcp` per the rollout order).

## Setup

```bash
JIRA_MCP_PORT=8012
JIRA_ACCESS_TOKEN=...   # local/dev only — production uses injected credentials_json
JIRA_CLOUD_ID=...       # local/dev only
```

Vault `credentials_json` (production path, matches `datumbridge-integrations`'
`JiraTokenBundle`):

```json
{
  "access_token": "...",
  "cloud_id": "..."
}
```

`refresh_token`/`client_id`/`client_secret`/`token_uri` also arrive in that bundle
but are unused here — refresh is `datumbridge-integrations`' job
(`Vault.RefreshJiraIfNeeded`), run before inject hands back a bundle.

## Run

```bash
go run ./cmd/api
# GET  http://localhost:8012/health
# POST http://localhost:8012/mcp
```

```bash
docker build -t jira-mcp .
docker run --rm -p 8012:8012 --env-file .env jira-mcp
```

## Prerequisite: backend wiring

Per the reference doc's BR11, Jira already has its `vaultProviderSpecs` entry in
`datumbridge-integrations/internal/credentials/provider_match.go` (`jira-mcp`,
`jira`), so `POST /internal/v1/inject` already resolves this provider — no backend
change needed before running this server against a real connected user.

## Tests

```bash
go test ./...
```

## Notes

- `internal/mcp` is the generic Streamable HTTP MCP transport, copied verbatim
  from `Bright-Data-MCP` (protocol `2024-11-05`) — it has no Jira-specific code.
- `internal/jira` is the Jira Cloud REST API v3 + Agile 1.0 client and credential
  parsing.
- `internal/tools` registers and implements every `jira_*` tool.
- Tool bodies that touch Jira's rich-text fields (issue description, comments)
  accept plain text and are converted to a minimal single-paragraph Atlassian
  Document Format (ADF) node — see `jira.PlainTextToADF`.

## Non-goals

Jira Service Management, Confluence (separate provider/server), attachment
upload, webhook/admin configuration, replacing this server with any
Atlassian-hosted MCP offering.
