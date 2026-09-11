# Workflows

## Credential injection (every tool call)

```
Agent → jira-mcp: tools/call (jira_get_issue, args={credentials_json omitted})
jira-mcp → datumbridge-integrations: POST /internal/v1/inject
  {user_id, mcp_server:"jira-mcp", args:{}}
datumbridge-integrations → jira-mcp: {injected:true, credentials_json:"{...}"}
jira-mcp → api.atlassian.com: GET /ex/jira/{cloudId}/rest/api/3/issue/{key}
  Authorization: Bearer <access_token>
jira-mcp → Agent: tool result
```

In practice, the Weaver Tool Registry performs the inject step and passes the
resulting `credentials_json` as a tool argument — `jira-mcp` itself only
*parses* the bundle (`internal/jira.ParseCredentials`); it does not call
`/internal/v1/inject` directly. This mirrors `Bright-Data-MCP`'s
`credentials_json`/`credentials_path` argument pattern exactly.

## Issue-not-found / not-connected

```
Tool call → Jira 404 → jira-mcp returns isError:true with the sanitized message
Tool call → inject 404 (no jira credentials in vault) → the Tool Registry layer
  surfaces "connect Jira in Weaver first" before jira-mcp is even reached
```

## Sprint lifecycle

```
jira_create_sprint (state=future, implicit)
  → jira_move_issues_to_sprint (add backlog issues)
  → jira_update_sprint (state=active)   # starts the sprint
  → ... work happens, worklogs/comments/transitions accrue ...
  → jira_update_sprint (state=closed)   # closes the sprint
```
