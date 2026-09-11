package tools

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/datumbridge/jira-mcp/internal/mcp"
)

func registerUserTools(add toolAdder) {
	add("jira_get_myself",
		"Get the connected user's own Jira account (identity, timezone, account id) — useful for confirming which account a connection maps to.",
		baseProps(nil), nil, handleGetMyself)

	add("jira_search_assignable_users",
		"Search users assignable to a project or issue (for picking an assignee).",
		baseProps(map[string]interface{}{
			"project_key": map[string]interface{}{"type": "string", "description": "One of project_key or issue_key is required"},
			"issue_key":   map[string]interface{}{"type": "string"},
			"query":       map[string]interface{}{"type": "string", "description": "Name/email substring filter"},
		}),
		nil, handleSearchAssignableUsers)
}

func handleGetMyself(raw json.RawMessage) map[string]interface{} {
	client, _, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	cctx, cancel := ctx()
	defer cancel()
	body, _, err := client.Core(cctx, http.MethodGet, "/myself", nil, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}

func handleSearchAssignableUsers(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	projectKey := strArg(m, "project_key")
	issueKey := strArg(m, "issue_key")
	if projectKey == "" && issueKey == "" {
		return mcp.ToolResultError("project_key or issue_key is required")
	}
	q := url.Values{}
	if projectKey != "" {
		q.Set("project", projectKey)
	}
	if issueKey != "" {
		q.Set("issueKey", issueKey)
	}
	if query := strArg(m, "query"); query != "" {
		q.Set("query", query)
	}
	cctx, cancel := ctx()
	defer cancel()
	body, _, err := client.Core(cctx, http.MethodGet, "/user/assignable/search", q, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}
