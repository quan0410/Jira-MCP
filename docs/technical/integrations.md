# Integrations

## datumbridge-integrations (credential source)

- Route: `POST /internal/v1/inject` (service-to-service, `serviceAPIKeyOK` gated).
- Bundle: `JiraTokenBundle` (`internal/credentials/vault.go`) —
  `access_token` + `cloud_id` used here; `refresh_token`/`client_id`/
  `client_secret`/`token_uri` ignored (refresh stays server-side).
- `vaultProviderSpecs` (`internal/credentials/provider_match.go`) already has
  `jira-mcp`/`jira` tokens wired — no backend change needed for this provider
  (see `doc/MCP_TOOL_REFERENCE_ARCHITECTURE.md`'s BR11 gap table; Jira is one of
  the 4 already-wired rows).

## datumbridge-integrations-ui (connect UI)

- Not touched by this build — the Jira `IntegrationDef` (`src/registry.ts`) and
  connect flow already exist and are out of scope; this repo only consumes an
  already-connected credential.

## Weaver Tool Registry (`datumbridge-mcp`, out of workspace)

- Per `doc/MCP_TOOL_REFERENCE_ARCHITECTURE.md`'s BR12, registering `jira-mcp` as
  a callable MCP server (and keeping its own provider-matching table in
  lockstep) is a coordination step with that repo's owners — not resolved by
  this repo.

## Jira Cloud (Atlassian)

- OAuth 2.0 (3LO) app / API scopes are configured on the
  `datumbridge-integrations` side (`oauth_jira.go`) — this repo has no
  Atlassian app registration of its own, it only consumes the resulting token.
