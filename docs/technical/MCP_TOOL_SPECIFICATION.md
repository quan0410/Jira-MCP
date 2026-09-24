# MCP Tool Specification

| Field | Value |
|---|---|
| Folder | `jira-mcp` |
| mcpServer id | `jira-mcp` |
| Transport | Streamable HTTP `POST /mcp` |
| Protocol | `2024-11-05` |
| Health | `GET /health` |
| Default port | `8012` |
| Upstream | Jira Cloud REST API v3 + Agile REST API 1.0 |

## Tool visibility

Full scope by default (all tools below). `JIRA_TOOLS=tool1,tool2` narrows to an
explicit allowlist — there is no Rapid/Pro split (no cost-tier distinction to
gate on, unlike Bright Data's Direct API billing).

## Tools

| Name | Group | Method | Endpoint |
|---|---|---|---|
| `jira_get_issue` | issues | GET | `/issue/{key}` |
| `jira_search_issues` | issues | POST | `/search/jql` |
| `jira_create_issue` | issues | POST | `/issue` |
| `jira_update_issue` | issues | PUT | `/issue/{key}` |
| `jira_delete_issue` | issues | DELETE | `/issue/{key}` |
| `jira_list_comments` | comments | GET | `/issue/{key}/comment` |
| `jira_add_comment` | comments | POST | `/issue/{key}/comment` |
| `jira_update_comment` | comments | PUT | `/issue/{key}/comment/{id}` |
| `jira_delete_comment` | comments | DELETE | `/issue/{key}/comment/{id}` |
| `jira_list_transitions` | transitions | GET | `/issue/{key}/transitions` |
| `jira_transition_issue` | transitions | POST | `/issue/{key}/transitions` |
| `jira_list_worklogs` | worklogs | GET | `/issue/{key}/worklog` |
| `jira_add_worklog` | worklogs | POST | `/issue/{key}/worklog` |
| `jira_list_projects` | projects | GET | `/project/search` |
| `jira_get_project` | projects | GET | `/project/{key}` |
| `jira_list_issue_types` | projects | GET | `/issue/createmeta/{key}/issuetypes` |
| `jira_get_myself` | users | GET | `/myself` |
| `jira_search_assignable_users` | users | GET | `/user/assignable/search` |
| `jira_list_boards` | agile | GET | `(agile)/board` |
| `jira_get_board` | agile | GET | `(agile)/board/{id}` |
| `jira_create_board` | agile | POST | `(agile)/board` |
| `jira_delete_board` | agile | DELETE | `(agile)/board/{id}` |
| `jira_get_board_configuration` | agile | GET | `(agile)/board/{id}/configuration` |
| `jira_list_sprints` | agile | GET | `(agile)/board/{id}/sprint` |
| `jira_get_sprint` | agile | GET | `(agile)/sprint/{id}` |
| `jira_list_backlog_issues` | agile | GET | `(agile)/board/{id}/backlog` |
| `jira_list_sprint_issues` | agile | GET | `(agile)/sprint/{id}/issue` |
| `jira_create_sprint` | agile | POST | `(agile)/sprint` |
| `jira_update_sprint` | agile | POST | `(agile)/sprint/{id}` (partial update) |
| `jira_move_issues_to_sprint` | agile | POST | `(agile)/sprint/{id}/issue` |
| `jira_get_issue_estimation` | agile | GET | `(agile)/issue/{key}/estimation` |
| `jira_list_epics` | epics | GET | `(agile)/board/{id}/epic` |
| `jira_get_epic` | epics | GET | `(agile)/epic/{idOrKey}` |
| `jira_list_epic_issues` | epics | GET | `(agile)/epic/{idOrKey}/issue` |
| `jira_move_issues_to_epic` | epics | POST | `(agile)/epic/{idOrKey}/issue` |
| `jira_remove_issues_from_epic` | epics | POST | `(agile)/epic/none/issue` |
| `jira_list_components` | projects | GET | `/project/{key}/components` |
| `jira_create_component` | projects | POST | `/component` |
| `jira_list_versions` | projects | GET | `/project/{key}/versions` |
| `jira_create_version` | projects | POST | `/version` |
| `jira_list_priorities` | config | GET | `/priority` |
| `jira_list_statuses` | config | GET | `/status` |
| `jira_list_issue_link_types` | config | GET | `/issueLinkType` |
| `jira_link_issues` | config | POST | `/issueLink` |
| `jira_list_webhooks` | webhooks | GET | `/webhook` |
| `jira_delete_webhooks` | webhooks | DELETE | `/webhook` |

`(agile)` = `/rest/agile/1.0`; everything else is `/rest/api/3`, both under
`https://api.atlassian.com/ex/jira/{cloudId}`.

## Non-goals / not implemented

Jira Service Management, attachment upload/download, bulk issue operations beyond
`jira_move_issues_to_sprint` and `jira_move_issues_to_epic`.
