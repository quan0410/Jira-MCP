# Component Diagram

```text
Agent / Hub (Weaver Tool Registry)
    │  POST /mcp (Streamable HTTP)
    ▼
jira-mcp
  mcp.Server ──► tools.Register handlers (issues/comments/transitions/
                  worklogs/projects/users/agile)
                     │
                     ▼
              jira.Client (Core /rest/api/3, Agile /rest/agile/1.0)
                     │  Bearer <access_token>
                     ▼
         api.atlassian.com/ex/jira/{cloudId}/...

Credential source (not pictured above — happens before the tool call reaches
jira.Client):

  jira-mcp  ──POST /internal/v1/inject──►  datumbridge-integrations
            ◄──{access_token, cloud_id}────
```

`GET /health` is process liveness only (no upstream call, no credential check).
