package tools

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/datumbridge/jira-mcp/internal/mcp"
)

func registerEpicTools(add toolAdder) {
	add("jira_list_epics",
		"List epics from a board (requires read:epic:jira-software).",
		baseProps(map[string]interface{}{
			"board_id":    map[string]interface{}{"type": "integer"},
			"done":        map[string]interface{}{"type": "string", "enum": []string{"true", "false"}, "description": "Filter by completion state"},
			"max_results": map[string]interface{}{"type": "integer", "default": 50},
		}),
		[]string{"board_id"}, handleListEpics)

	add("jira_get_epic",
		"Get a single epic by id or key (requires read:epic:jira-software).",
		baseProps(map[string]interface{}{
			"epic_id_or_key": map[string]interface{}{"type": "string"},
		}),
		[]string{"epic_id_or_key"}, handleGetEpic)

	add("jira_list_epic_issues",
		"List issues belonging to an epic (requires read:epic:jira-software).",
		baseProps(map[string]interface{}{
			"epic_id_or_key": map[string]interface{}{"type": "string"},
			"max_results":    map[string]interface{}{"type": "integer", "default": 50},
		}),
		[]string{"epic_id_or_key"}, handleListEpicIssues)

	add("jira_move_issues_to_epic",
		"Move issues into an epic (requires write:epic:jira-software).",
		baseProps(map[string]interface{}{
			"epic_id_or_key": map[string]interface{}{"type": "string"},
			"issue_keys":     map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "maxItems": 50},
		}),
		[]string{"epic_id_or_key", "issue_keys"}, handleMoveIssuesToEpic)

	add("jira_remove_issues_from_epic",
		"Remove issues from any epic (requires write:epic:jira-software).",
		baseProps(map[string]interface{}{
			"issue_keys": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "maxItems": 50},
		}),
		[]string{"issue_keys"}, handleRemoveIssuesFromEpic)
}

func handleListEpics(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	boardID := intArg(m, "board_id", 0)
	if boardID <= 0 {
		return mcp.ToolResultError("board_id is required")
	}
	q := url.Values{}
	if done := strArg(m, "done"); done != "" {
		q.Set("done", done)
	}
	mr := intArg(m, "max_results", 50)
	if mr <= 0 || mr > 100 {
		mr = 50
	}
	q.Set("maxResults", strconv.Itoa(mr))

	cctx, cancel := ctx()
	defer cancel()
	body, _, err := client.Agile(cctx, http.MethodGet, "/board/"+strconv.Itoa(boardID)+"/epic", q, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}

func handleGetEpic(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	epicIDOrKey := strArg(m, "epic_id_or_key")
	if epicIDOrKey == "" {
		return mcp.ToolResultError("epic_id_or_key is required")
	}
	cctx, cancel := ctx()
	defer cancel()
	body, _, err := client.Agile(cctx, http.MethodGet, "/epic/"+url.PathEscape(epicIDOrKey), nil, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}

func handleListEpicIssues(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	epicIDOrKey := strArg(m, "epic_id_or_key")
	if epicIDOrKey == "" {
		return mcp.ToolResultError("epic_id_or_key is required")
	}
	q := url.Values{}
	mr := intArg(m, "max_results", 50)
	if mr <= 0 || mr > 100 {
		mr = 50
	}
	q.Set("maxResults", strconv.Itoa(mr))

	cctx, cancel := ctx()
	defer cancel()
	body, _, err := client.Agile(cctx, http.MethodGet, "/epic/"+url.PathEscape(epicIDOrKey)+"/issue", q, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}

func handleMoveIssuesToEpic(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	epicIDOrKey := strArg(m, "epic_id_or_key")
	issueKeys := strSliceArg(m, "issue_keys")
	if epicIDOrKey == "" || len(issueKeys) == 0 {
		return mcp.ToolResultError("epic_id_or_key and a non-empty issue_keys array are required")
	}
	if len(issueKeys) > 50 {
		issueKeys = issueKeys[:50]
	}
	cctx, cancel := ctx()
	defer cancel()
	_, status, err := client.Agile(cctx, http.MethodPost, "/epic/"+url.PathEscape(epicIDOrKey)+"/issue", nil,
		map[string]interface{}{"issues": issueKeys})
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return mcp.ToolResultText("moved " + strconv.Itoa(len(issueKeys)) + " issue(s) to epic " + epicIDOrKey + " (HTTP " + strconv.Itoa(status) + ")")
}

func handleRemoveIssuesFromEpic(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	issueKeys := strSliceArg(m, "issue_keys")
	if len(issueKeys) == 0 {
		return mcp.ToolResultError("a non-empty issue_keys array is required")
	}
	if len(issueKeys) > 50 {
		issueKeys = issueKeys[:50]
	}
	cctx, cancel := ctx()
	defer cancel()
	_, status, err := client.Agile(cctx, http.MethodPost, "/epic/none/issue", nil,
		map[string]interface{}{"issues": issueKeys})
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return mcp.ToolResultText("removed " + strconv.Itoa(len(issueKeys)) + " issue(s) from epic (HTTP " + strconv.Itoa(status) + ")")
}
