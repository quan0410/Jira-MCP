package tools

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/datumbridge/jira-mcp/internal/mcp"
)

// Agile tools cover Jira Software's boards/sprints/backlog — the Agile REST API
// (/rest/agile/1.0), distinct from core issue tracking (/rest/api/3). See
// doc/MCP_TOOL_REFERENCE_ARCHITECTURE.md's "full tool scope" rule (BR6): this
// project's rollout deliberately includes Agile alongside core issue tracking.
func registerAgileTools(add toolAdder) {
	add("jira_list_boards",
		"List Jira Software boards, optionally filtered to one project.",
		baseProps(map[string]interface{}{
			"project_key_or_id": map[string]interface{}{"type": "string"},
			"max_results":       map[string]interface{}{"type": "integer", "default": 50},
		}),
		nil, handleListBoards)

	add("jira_get_board",
		"Get a single board by id.",
		baseProps(map[string]interface{}{
			"board_id": map[string]interface{}{"type": "integer"},
		}),
		[]string{"board_id"}, handleGetBoard)

	add("jira_list_sprints",
		"List sprints on a board.",
		baseProps(map[string]interface{}{
			"board_id": map[string]interface{}{"type": "integer"},
			"state":    map[string]interface{}{"type": "string", "enum": []string{"active", "future", "closed"}, "description": "Optional; omit for all"},
		}),
		[]string{"board_id"}, handleListSprints)

	add("jira_get_sprint",
		"Get a single sprint by id.",
		baseProps(map[string]interface{}{
			"sprint_id": map[string]interface{}{"type": "integer"},
		}),
		[]string{"sprint_id"}, handleGetSprint)

	add("jira_list_backlog_issues",
		"List issues in a board's backlog (not yet assigned to a sprint).",
		baseProps(map[string]interface{}{
			"board_id":    map[string]interface{}{"type": "integer"},
			"max_results": map[string]interface{}{"type": "integer", "default": 50},
		}),
		[]string{"board_id"}, handleListBacklogIssues)

	add("jira_list_sprint_issues",
		"List issues assigned to a sprint.",
		baseProps(map[string]interface{}{
			"sprint_id":   map[string]interface{}{"type": "integer"},
			"max_results": map[string]interface{}{"type": "integer", "default": 50},
		}),
		[]string{"sprint_id"}, handleListSprintIssues)

	add("jira_create_sprint",
		"Create a new sprint on a board (created in the \"future\" state; start it with jira_update_sprint).",
		baseProps(map[string]interface{}{
			"board_id":   map[string]interface{}{"type": "integer"},
			"name":       map[string]interface{}{"type": "string"},
			"start_date": map[string]interface{}{"type": "string", "description": "ISO-8601, e.g. 2026-01-15T09:00:00.000Z"},
			"end_date":   map[string]interface{}{"type": "string"},
			"goal":       map[string]interface{}{"type": "string"},
		}),
		[]string{"board_id", "name"}, handleCreateSprint)

	add("jira_update_sprint",
		"Partially update a sprint — set name/goal/dates, or change state to start (\"active\") or close (\"closed\") it.",
		baseProps(map[string]interface{}{
			"sprint_id":  map[string]interface{}{"type": "integer"},
			"name":       map[string]interface{}{"type": "string"},
			"state":      map[string]interface{}{"type": "string", "enum": []string{"active", "closed"}},
			"start_date": map[string]interface{}{"type": "string"},
			"end_date":   map[string]interface{}{"type": "string"},
			"goal":       map[string]interface{}{"type": "string"},
		}),
		[]string{"sprint_id"}, handleUpdateSprint)

	add("jira_move_issues_to_sprint",
		"Move issues into a sprint (from the backlog or another sprint).",
		baseProps(map[string]interface{}{
			"sprint_id":  map[string]interface{}{"type": "integer"},
			"issue_keys": map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "maxItems": 50},
		}),
		[]string{"sprint_id", "issue_keys"}, handleMoveIssuesToSprint)
}

func handleListBoards(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	q := url.Values{}
	if p := strArg(m, "project_key_or_id"); p != "" {
		q.Set("projectKeyOrId", p)
	}
	if mr := intArg(m, "max_results", 0); mr > 0 {
		q.Set("maxResults", strconv.Itoa(mr))
	}
	cctx, cancel := ctx()
	defer cancel()
	body, _, err := client.Agile(cctx, http.MethodGet, "/board", q, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}

func handleGetBoard(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	boardID := intArg(m, "board_id", 0)
	if boardID <= 0 {
		return mcp.ToolResultError("board_id is required")
	}
	cctx, cancel := ctx()
	defer cancel()
	body, _, err := client.Agile(cctx, http.MethodGet, "/board/"+strconv.Itoa(boardID), nil, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}

func handleListSprints(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	boardID := intArg(m, "board_id", 0)
	if boardID <= 0 {
		return mcp.ToolResultError("board_id is required")
	}
	q := url.Values{}
	if state := strArg(m, "state"); state != "" {
		q.Set("state", state)
	}
	cctx, cancel := ctx()
	defer cancel()
	body, _, err := client.Agile(cctx, http.MethodGet, "/board/"+strconv.Itoa(boardID)+"/sprint", q, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}

func handleGetSprint(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	sprintID := intArg(m, "sprint_id", 0)
	if sprintID <= 0 {
		return mcp.ToolResultError("sprint_id is required")
	}
	cctx, cancel := ctx()
	defer cancel()
	body, _, err := client.Agile(cctx, http.MethodGet, "/sprint/"+strconv.Itoa(sprintID), nil, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}

func handleListBacklogIssues(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	boardID := intArg(m, "board_id", 0)
	if boardID <= 0 {
		return mcp.ToolResultError("board_id is required")
	}
	q := url.Values{}
	if mr := intArg(m, "max_results", 0); mr > 0 {
		q.Set("maxResults", strconv.Itoa(mr))
	}
	cctx, cancel := ctx()
	defer cancel()
	body, _, err := client.Agile(cctx, http.MethodGet, "/board/"+strconv.Itoa(boardID)+"/backlog", q, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}

func handleListSprintIssues(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	sprintID := intArg(m, "sprint_id", 0)
	if sprintID <= 0 {
		return mcp.ToolResultError("sprint_id is required")
	}
	q := url.Values{}
	if mr := intArg(m, "max_results", 0); mr > 0 {
		q.Set("maxResults", strconv.Itoa(mr))
	}
	cctx, cancel := ctx()
	defer cancel()
	body, _, err := client.Agile(cctx, http.MethodGet, "/sprint/"+strconv.Itoa(sprintID)+"/issue", q, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}

func handleCreateSprint(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	boardID := intArg(m, "board_id", 0)
	name := strArg(m, "name")
	if boardID <= 0 || name == "" {
		return mcp.ToolResultError("board_id and name are required")
	}
	payload := map[string]interface{}{
		"name":          name,
		"originBoardId": boardID,
	}
	if v := strArg(m, "start_date"); v != "" {
		payload["startDate"] = v
	}
	if v := strArg(m, "end_date"); v != "" {
		payload["endDate"] = v
	}
	if v := strArg(m, "goal"); v != "" {
		payload["goal"] = v
	}
	cctx, cancel := ctx()
	defer cancel()
	body, _, err := client.Agile(cctx, http.MethodPost, "/sprint", nil, payload)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}

func handleUpdateSprint(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	sprintID := intArg(m, "sprint_id", 0)
	if sprintID <= 0 {
		return mcp.ToolResultError("sprint_id is required")
	}
	payload := map[string]interface{}{}
	if v := strArg(m, "name"); v != "" {
		payload["name"] = v
	}
	if v := strArg(m, "state"); v != "" {
		payload["state"] = v
	}
	if v := strArg(m, "start_date"); v != "" {
		payload["startDate"] = v
	}
	if v := strArg(m, "end_date"); v != "" {
		payload["endDate"] = v
	}
	if v := strArg(m, "goal"); v != "" {
		payload["goal"] = v
	}
	if len(payload) == 0 {
		return mcp.ToolResultError("at least one field to update is required")
	}
	cctx, cancel := ctx()
	defer cancel()
	// Partial update — Jira Agile API's POST /sprint/{id} (not PUT, which is a
	// full replace requiring every field).
	body, _, err := client.Agile(cctx, http.MethodPost, "/sprint/"+strconv.Itoa(sprintID), nil, payload)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}

func handleMoveIssuesToSprint(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	sprintID := intArg(m, "sprint_id", 0)
	issueKeys := strSliceArg(m, "issue_keys")
	if sprintID <= 0 || len(issueKeys) == 0 {
		return mcp.ToolResultError("sprint_id and a non-empty issue_keys array are required")
	}
	if len(issueKeys) > 50 {
		issueKeys = issueKeys[:50]
	}
	cctx, cancel := ctx()
	defer cancel()
	_, status, err := client.Agile(cctx, http.MethodPost, "/sprint/"+strconv.Itoa(sprintID)+"/issue", nil,
		map[string]interface{}{"issues": issueKeys})
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return mcp.ToolResultText("moved " + strconv.Itoa(len(issueKeys)) + " issue(s) to sprint " + strconv.Itoa(sprintID) + " (HTTP " + strconv.Itoa(status) + ")")
}
