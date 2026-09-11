# Changelog

## Unreleased

- Initial build: full core issue-tracking + Agile (boards/sprints) tool scope,
  27 tools, per `doc/MCP_TOOL_REFERENCE_ARCHITECTURE.md` item #1 (Jira).
- Fix: `jira_search_issues` now defaults `fields` to a usable set (key,
  summary, status, issuetype, assignee, priority, created, updated) when the
  caller omits it — found via live E2E test: Jira Cloud's `/search/jql`
  otherwise returns bare `{"id": "..."}` with nothing else. See BR-J7.
- Note: Agile tools (boards/sprints/backlog) require `datumbridge-integrations`
  to request Jira Software granular OAuth scopes
  (`read:board-scope:jira-software`, `read:project:jira`,
  `read:sprint:jira-software`, `write:sprint:jira-software`) — classic scopes
  (`read:jira-work`/`write:jira-work`) do not cover them. Existing connections
  must reconnect after that backend change ships.
