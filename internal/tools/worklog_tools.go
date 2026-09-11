package tools

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/datumbridge/jira-mcp/internal/jira"
	"github.com/datumbridge/jira-mcp/internal/mcp"
)

func registerWorklogTools(add toolAdder) {
	add("jira_list_worklogs",
		"List work log entries on an issue.",
		baseProps(map[string]interface{}{
			"issue_key": map[string]interface{}{"type": "string"},
		}),
		[]string{"issue_key"}, handleListWorklogs)

	add("jira_add_worklog",
		"Log time spent on an issue.",
		baseProps(map[string]interface{}{
			"issue_key":  map[string]interface{}{"type": "string"},
			"time_spent": map[string]interface{}{"type": "string", "description": "Jira duration shorthand, e.g. \"3h\", \"1d 4h\", \"30m\""},
			"comment":    map[string]interface{}{"type": "string"},
			"started":    map[string]interface{}{"type": "string", "description": "ISO-8601 with offset, e.g. 2026-01-15T09:00:00.000+0000; defaults to now"},
		}),
		[]string{"issue_key", "time_spent"}, handleAddWorklog)
}

func handleListWorklogs(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	issueKey := strArg(m, "issue_key")
	if issueKey == "" {
		return mcp.ToolResultError("issue_key is required")
	}
	cctx, cancel := ctx()
	defer cancel()
	body, _, err := client.Core(cctx, http.MethodGet, "/issue/"+url.PathEscape(issueKey)+"/worklog", nil, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}

func handleAddWorklog(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	issueKey := strArg(m, "issue_key")
	timeSpent := strArg(m, "time_spent")
	if issueKey == "" || timeSpent == "" {
		return mcp.ToolResultError("issue_key and time_spent are required")
	}
	payload := map[string]interface{}{"timeSpent": timeSpent}
	if comment := strArg(m, "comment"); comment != "" {
		payload["comment"] = jira.PlainTextToADF(comment)
	}
	if started := strArg(m, "started"); started != "" {
		payload["started"] = started
	}
	cctx, cancel := ctx()
	defer cancel()
	body, _, err := client.Core(cctx, http.MethodPost, "/issue/"+url.PathEscape(issueKey)+"/worklog", nil, payload)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}
