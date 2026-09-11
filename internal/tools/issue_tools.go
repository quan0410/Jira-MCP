package tools

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/datumbridge/jira-mcp/internal/jira"
	"github.com/datumbridge/jira-mcp/internal/mcp"
)

// defaultSearchFields is sent to /search/jql when the caller doesn't specify
// "fields" — that endpoint returns bare {"id": "..."} with no key, summary or
// anything else if "fields" is omitted entirely, verified against a live Jira
// Cloud site. jira_get_issue needs no such default (GET /issue/{key} already
// returns a full field set on its own).
var defaultSearchFields = []string{"key", "summary", "status", "issuetype", "assignee", "priority", "created", "updated"}

func registerIssueTools(add toolAdder) {
	add("jira_get_issue",
		"Get a single Jira issue by key or id.",
		baseProps(map[string]interface{}{
			"issue_key": map[string]interface{}{"type": "string", "description": "e.g. PROJ-123"},
			"fields":    map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Field names to return; omit for Jira's default set"},
		}),
		[]string{"issue_key"}, handleGetIssue)

	add("jira_search_issues",
		"Search issues with JQL (Jira Query Language). Uses Jira Cloud's cursor-paginated /search/jql.",
		baseProps(map[string]interface{}{
			"jql":             map[string]interface{}{"type": "string"},
			"fields":          map[string]interface{}{"type": "array", "items": map[string]interface{}{"type": "string"}, "description": "Defaults to key, summary, status, issuetype, assignee, priority, created, updated — Jira returns only {id} with no fields at all if this is left empty"},
			"max_results":     map[string]interface{}{"type": "integer", "default": 50},
			"next_page_token": map[string]interface{}{"type": "string", "description": "From a previous call's response, to page forward"},
		}),
		[]string{"jql"}, handleSearchIssues)

	add("jira_create_issue",
		"Create a new Jira issue.",
		baseProps(map[string]interface{}{
			"project_key": map[string]interface{}{"type": "string"},
			"issue_type":  map[string]interface{}{"type": "string", "description": "e.g. Task, Bug, Story"},
			"summary":     map[string]interface{}{"type": "string"},
			"description": map[string]interface{}{"type": "string", "description": "Plain text; converted to Atlassian Document Format"},
			"fields":      map[string]interface{}{"type": "object", "description": "Extra/overriding fields merged into the create payload (e.g. labels, assignee, priority)"},
		}),
		[]string{"project_key", "issue_type", "summary"}, handleCreateIssue)

	add("jira_update_issue",
		"Update fields on an existing Jira issue.",
		baseProps(map[string]interface{}{
			"issue_key": map[string]interface{}{"type": "string"},
			"fields":    map[string]interface{}{"type": "object", "description": "Fields to set, Jira REST v3 shape. A plain-string \"description\" is auto-converted to ADF."},
		}),
		[]string{"issue_key", "fields"}, handleUpdateIssue)

	add("jira_delete_issue",
		"Delete a Jira issue.",
		baseProps(map[string]interface{}{
			"issue_key": map[string]interface{}{"type": "string"},
		}),
		[]string{"issue_key"}, handleDeleteIssue)
}

func handleGetIssue(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	issueKey := strArg(m, "issue_key")
	if issueKey == "" {
		return mcp.ToolResultError("issue_key is required")
	}
	q := url.Values{}
	if fields := strSliceArg(m, "fields"); len(fields) > 0 {
		q.Set("fields", strings.Join(fields, ","))
	}
	cctx, cancel := ctx()
	defer cancel()
	body, _, err := client.Core(cctx, http.MethodGet, "/issue/"+url.PathEscape(issueKey), q, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}

func handleSearchIssues(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	jql := strArg(m, "jql")
	if jql == "" {
		return mcp.ToolResultError("jql is required")
	}
	maxResults := intArg(m, "max_results", 50)
	if maxResults <= 0 || maxResults > 100 {
		maxResults = 50
	}
	payload := map[string]interface{}{
		"jql":        jql,
		"maxResults": maxResults,
	}
	fields := strSliceArg(m, "fields")
	if len(fields) == 0 {
		// Jira Cloud's /search/jql returns bare {"id": "..."} with no key or
		// fields at all when "fields" is omitted (unlike the deprecated
		// /search, which defaulted to a useful set) — default to something
		// actually usable rather than making every caller discover this.
		fields = defaultSearchFields
	}
	payload["fields"] = fields
	if token := strArg(m, "next_page_token"); token != "" {
		payload["nextPageToken"] = token
	}
	cctx, cancel := ctx()
	defer cancel()
	body, _, err := client.Core(cctx, http.MethodPost, "/search/jql", nil, payload)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}

func handleCreateIssue(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	projectKey := strArg(m, "project_key")
	issueType := strArg(m, "issue_type")
	summary := strArg(m, "summary")
	if projectKey == "" || issueType == "" || summary == "" {
		return mcp.ToolResultError("project_key, issue_type and summary are required")
	}
	fields := map[string]interface{}{
		"project":   map[string]interface{}{"key": projectKey},
		"issuetype": map[string]interface{}{"name": issueType},
		"summary":   summary,
	}
	if desc := strArg(m, "description"); desc != "" {
		fields["description"] = jira.PlainTextToADF(desc)
	}
	for k, v := range mapArg(m, "fields") {
		fields[k] = v
	}
	cctx, cancel := ctx()
	defer cancel()
	body, _, err := client.Core(cctx, http.MethodPost, "/issue", nil, map[string]interface{}{"fields": fields})
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}

func handleUpdateIssue(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	issueKey := strArg(m, "issue_key")
	fields := mapArg(m, "fields")
	if issueKey == "" || len(fields) == 0 {
		return mcp.ToolResultError("issue_key and a non-empty fields object are required")
	}
	if desc, ok := fields["description"].(string); ok {
		fields["description"] = jira.PlainTextToADF(desc)
	}
	cctx, cancel := ctx()
	defer cancel()
	_, status, err := client.Core(cctx, http.MethodPut, "/issue/"+url.PathEscape(issueKey), nil, map[string]interface{}{"fields": fields})
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return mcp.ToolResultText("updated " + issueKey + " (HTTP " + strconv.Itoa(status) + ")")
}

func handleDeleteIssue(raw json.RawMessage) map[string]interface{} {
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
	_, status, err := client.Core(cctx, http.MethodDelete, "/issue/"+url.PathEscape(issueKey), nil, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return mcp.ToolResultText("deleted " + issueKey + " (HTTP " + strconv.Itoa(status) + ")")
}

// rawJSONResult passes an already-JSON response body through as the tool's text
// result without re-encoding (Jira responses are already well-formed JSON).
func rawJSONResult(body []byte) map[string]interface{} {
	if len(body) == 0 {
		return mcp.ToolResultText("{}")
	}
	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		return mcp.ToolResultText(string(body))
	}
	return jsonResult(v)
}
