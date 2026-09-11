package tools

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/datumbridge/jira-mcp/internal/jira"
	"github.com/datumbridge/jira-mcp/internal/mcp"
)

func registerTransitionTools(add toolAdder) {
	add("jira_list_transitions",
		"List the workflow transitions currently available for an issue (the moves its status can make next).",
		baseProps(map[string]interface{}{
			"issue_key": map[string]interface{}{"type": "string"},
		}),
		[]string{"issue_key"}, handleListTransitions)

	add("jira_transition_issue",
		"Move an issue through its workflow (e.g. To Do -> In Progress -> Done). Get transition_id from jira_list_transitions first.",
		baseProps(map[string]interface{}{
			"issue_key":     map[string]interface{}{"type": "string"},
			"transition_id": map[string]interface{}{"type": "string"},
			"comment":       map[string]interface{}{"type": "string", "description": "Optional comment to add along with the transition"},
			"fields":        map[string]interface{}{"type": "object", "description": "Optional fields to set as part of the transition (e.g. resolution)"},
		}),
		[]string{"issue_key", "transition_id"}, handleTransitionIssue)
}

func handleListTransitions(raw json.RawMessage) map[string]interface{} {
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
	body, _, err := client.Core(cctx, http.MethodGet, "/issue/"+url.PathEscape(issueKey)+"/transitions", nil, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}

func handleTransitionIssue(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	issueKey := strArg(m, "issue_key")
	transitionID := strArg(m, "transition_id")
	if issueKey == "" || transitionID == "" {
		return mcp.ToolResultError("issue_key and transition_id are required")
	}
	payload := map[string]interface{}{
		"transition": map[string]interface{}{"id": transitionID},
	}
	if fields := mapArg(m, "fields"); len(fields) > 0 {
		payload["fields"] = fields
	}
	if comment := strArg(m, "comment"); comment != "" {
		payload["update"] = map[string]interface{}{
			"comment": []map[string]interface{}{
				{"add": map[string]interface{}{"body": jira.PlainTextToADF(comment)}},
			},
		}
	}
	cctx, cancel := ctx()
	defer cancel()
	_, status, err := client.Core(cctx, http.MethodPost, "/issue/"+url.PathEscape(issueKey)+"/transitions", nil, payload)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return mcp.ToolResultText("transitioned " + issueKey + " (HTTP " + strconv.Itoa(status) + ")")
}
