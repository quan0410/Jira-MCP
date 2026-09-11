# Security Architecture

## Trust model

- The Jira access token authorizes API calls as the **connecting Weaver user's own
  Jira identity** — there is no shared/bot Jira account.
- Credentials are injected per tool call from `datumbridge-integrations`'
  encrypted vault; this server never persists them.
- `cloud_id` is pinned from the injected bundle — a tool cannot redirect a call
  at a different Jira site than the one the connecting user actually connected.

## Controls

| Control | Behavior |
|---|---|
| `cloud_id` from credentials only | Never accepted as a free-form tool argument |
| Path jail | `credentials_path` must stay under `JIRA_CREDENTIALS_DIR` |
| Logging | Never log the `Authorization` header or full request/response bodies at info level |
| Error sanitization | Provider error bodies are capped at 500 chars and redacted if they contain "bearer " |
| Response size cap | 10 MiB read limit per upstream response |
| Health | `/health` never makes an upstream call or reads credentials |

## Operator duties

Comply with the connecting user's own Jira permissions — this server does not
grant any access the user's own Jira account doesn't already have; it exercises
their existing OAuth scopes.
