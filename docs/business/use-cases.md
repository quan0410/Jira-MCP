# Use Cases

1. **Triage a bug report.** Agent creates an issue (`jira_create_issue`), adds a
   reproduction comment (`jira_add_comment`), and transitions it to "In Progress"
   (`jira_list_transitions` → `jira_transition_issue`).
2. **Stand-up summary.** Agent searches active-sprint issues assigned to a user
   (`jira_search_issues` with a JQL like `sprint in openSprints() AND assignee = currentUser()`)
   and summarizes status.
3. **Sprint planning.** Agent lists the backlog (`jira_list_backlog_issues`),
   creates a new sprint (`jira_create_sprint`), and moves selected issues into it
   (`jira_move_issues_to_sprint`), then starts it (`jira_update_sprint` with
   `state: "active"`).
4. **Time tracking.** Agent logs work against an issue after a task is done
   (`jira_add_worklog`).
5. **Assignee lookup.** Agent finds who can be assigned to a ticket
   (`jira_search_assignable_users`) before updating it
   (`jira_update_issue` with `fields.assignee`).
