# Deployment Architecture

Single Go binary, containerized (`Dockerfile`, multi-stage Alpine build), stateless
except for in-memory MCP sessions (`internal/mcp.SessionStore`, 30-minute TTL —
safe to lose on restart, a client just calls `initialize` again).

- **Port:** `8012` (`JIRA_MCP_PORT`, matches the reference doc's convention of one
  port per provider service; distinct from Bright-Data-MCP's `8011`).
- **No database.** No local persistence at all — every request is either
  stateless (MCP protocol bookkeeping) or proxies straight through to Jira Cloud
  after a credential fetch from `datumbridge-integrations`.
- **Network dependency:** must reach `datumbridge-integrations`'s
  `/internal/v1/inject` (service-to-service, internal network only in
  production) and `api.atlassian.com` (public internet / Atlassian's API
  gateway).
- **Horizontal scaling:** trivially stateless-scalable — session store is
  per-process, but since a lost session only costs one extra `initialize` round
  trip, running multiple replicas behind a load balancer needs no sticky
  sessions.
- **Registration:** per `doc/MCP_TOOL_REFERENCE_ARCHITECTURE.md`'s BR12, adding
  `jira-mcp` to the Weaver Tool Registry (`datumbridge-mcp`) is a coordination
  step outside this repo — not resolved here.
