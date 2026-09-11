# ADR-0001 — Standalone Go Service, Vault-Injected Credentials

## Context

Jira is item #1 in `doc/MCP_TOOL_REFERENCE_ARCHITECTURE.md`'s rollout: a channel
`datumbridge-integrations` already stores OAuth credentials for, but with no MCP
tool-server yet. That document (BR1–BR13) already settled the architecture
question for all 14 upcoming providers; this ADR records how jira-mcp applies it.

## Decision

1. Own Go service, own repo (`jira-mcp/`), not a package inside
   `datumbridge-integrations` — per BR1.
2. Credentials come exclusively from `POST /internal/v1/inject` on
   `datumbridge-integrations` at tool-call time (`JiraTokenBundle` shape:
   `access_token` + `cloud_id`) — per BR4/BR5. No OAuth flow, no token storage,
   no refresh logic in this repo — per BR9.
3. Transport: Streamable HTTP `POST /mcp`, protocol `2024-11-05`, `GET /health`
   liveness-only — per BR3, copied from `Bright-Data-MCP`'s `internal/mcp`
   package verbatim (it is provider-agnostic).
4. Full tool scope from the first release: core issue tracking (issues,
   comments, transitions, worklogs, projects, users) plus Agile
   (boards/sprints/backlog) — per BR6 and the explicit "Core + Agile" scope
   decision made when this build was scoped (see `dev_pm`/`dev_ba` history,
   2026-09-09).
5. `{cloudId}` is read only from the injected bundle, never accepted as a tool
   argument — per BR8 (tenant-identifying values pinned server-side).

## Alternatives Considered

- Embedding Jira tool handlers inside `datumbridge-integrations` — rejected,
  contradicts BR1 and the credential-injection trust boundary the whole
  ecosystem is built around (see `doc/ARCHITECTURE.md`'s `/internal/v1/inject`
  design).
- Minimal/rapid tool subset first, full coverage later — rejected per the
  explicit "full tool scope" decision for this build; Bright-Data-MCP's
  Rapid/Pro split was considered but not adopted since there is no equivalent
  billing/cost tier distinction for Jira API calls.

## Consequences

Consistent with every other planned MCP server in the rollout; a second
provider's build can copy this repo's `internal/mcp` and general shape with
near-zero changes. Refresh-token bugs, if any, are isolated to
`datumbridge-integrations` and out of this repo's blast radius.

## Trade-offs

Plain-text tool inputs for rich-text fields (description, comments) are
auto-wrapped into minimal ADF — callers needing real rich formatting must pass
ADF objects directly through the `fields` parameter (see `design.md`).

## Risks

Jira REST API shape drift (mitigated by targeting the current, non-deprecated
`/rest/api/3/search/jql` endpoint rather than the sunset `/rest/api/3/search`);
the cross-repo `vaultProviderSpecs`/`datumbridge-mcp` dependency noted in
`doc/MCP_TOOL_REFERENCE_ARCHITECTURE.md`'s BR12 — Jira already has that wiring,
so it is not a blocker for this provider specifically, but the note stands for
every provider after it in the rollout.
