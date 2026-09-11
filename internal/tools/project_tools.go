package tools

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/datumbridge/jira-mcp/internal/mcp"
)

func registerProjectTools(add toolAdder) {
	add("jira_list_projects",
		"List projects visible to the connected user.",
		baseProps(map[string]interface{}{
			"query":       map[string]interface{}{"type": "string", "description": "Optional name/key substring filter"},
			"max_results": map[string]interface{}{"type": "integer", "default": 50},
		}),
		nil, handleListProjects)

	add("jira_get_project",
		"Get a single project by key or id.",
		baseProps(map[string]interface{}{
			"project_key": map[string]interface{}{"type": "string"},
		}),
		[]string{"project_key"}, handleGetProject)

	add("jira_list_issue_types",
		"List issue types available for creating issues in a project (use before jira_create_issue to pick a valid issue_type name).",
		baseProps(map[string]interface{}{
			"project_key": map[string]interface{}{"type": "string"},
		}),
		[]string{"project_key"}, handleListIssueTypes)
}

func handleListProjects(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	q := url.Values{}
	if query := strArg(m, "query"); query != "" {
		q.Set("query", query)
	}
	if mr := intArg(m, "max_results", 0); mr > 0 {
		q.Set("maxResults", strconv.Itoa(mr))
	}
	cctx, cancel := ctx()
	defer cancel()
	body, _, err := client.Core(cctx, http.MethodGet, "/project/search", q, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}

func handleGetProject(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	projectKey := strArg(m, "project_key")
	if projectKey == "" {
		return mcp.ToolResultError("project_key is required")
	}
	cctx, cancel := ctx()
	defer cancel()
	body, _, err := client.Core(cctx, http.MethodGet, "/project/"+url.PathEscape(projectKey), nil, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}

func handleListIssueTypes(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	projectKey := strArg(m, "project_key")
	if projectKey == "" {
		return mcp.ToolResultError("project_key is required")
	}
	cctx, cancel := ctx()
	defer cancel()
	body, _, err := client.Core(cctx, http.MethodGet, "/issue/createmeta/"+url.PathEscape(projectKey)+"/issuetypes", nil, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}
