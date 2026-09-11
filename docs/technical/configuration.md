# Configuration

All via environment variables (`.env` loaded in dev via `godotenv`, see
`.env.example`).

| Var | Default | Purpose |
|---|---|---|
| `JIRA_MCP_PORT` / `PORT` | `8012` | HTTP listen port |
| `JIRA_MCP_ALLOWED_ORIGINS` | (none — CORS denied) | Comma-separated allowed CORS origins |
| `LOG_LEVEL` | `INFO` | `DEBUG`/`INFO`/`WARN`/`ERROR` |
| `JIRA_ACCESS_TOKEN` | — | Local/dev credential fallback |
| `JIRA_CLOUD_ID` | — | Local/dev credential fallback |
| `JIRA_CREDENTIALS_DIR` | `/credentials` | Jail root for `credentials_path` |
| `JIRA_TOOLS` | (unset — full scope) | Comma-separated tool-name allowlist |
| `JIRA_API_BASE` | `https://api.atlassian.com/ex/jira` | Override for testing against a mock/proxy |
| `JIRA_HTTP_TIMEOUT_SEC` | `30` | Upstream HTTP client timeout |

No database, no other required infrastructure.
