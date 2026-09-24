package tools

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/datumbridge/jira-mcp/internal/jira"
	"github.com/datumbridge/jira-mcp/internal/mcp"
)

func registerConfigTools(add toolAdder) {
	add("jira_list_priorities",
		"List issue priorities defined in Jira (manage:jira-configuration).",
		baseProps(nil), nil, handleListPriorities)

	add("jira_list_statuses",
		"List all issue statuses defined in Jira (manage:jira-configuration).",
		baseProps(nil), nil, handleListStatuses)

	add("jira_list_issue_link_types",
		"List available issue link types (e.g. Blocks, Relates to, Duplicates) — manage:jira-configuration.",
		baseProps(nil), nil, handleListIssueLinkTypes)

	add("jira_link_issues",
		"Create a link between two issues (e.g. PROJ-1 blocks PROJ-2) — manage:jira-configuration.",
		baseProps(map[string]interface{}{
			"link_type":         map[string]interface{}{"type": "string", "description": "Link type name, e.g. Blocks or Relates"},
			"inward_issue_key":  map[string]interface{}{"type": "string", "description": "Key of the inward issue"},
			"outward_issue_key": map[string]interface{}{"type": "string", "description": "Key of the outward issue"},
			"comment":           map[string]interface{}{"type": "string", "description": "Optional comment on the link"},
		}),
		[]string{"link_type", "inward_issue_key", "outward_issue_key"}, handleLinkIssues)
}

func handleListPriorities(raw json.RawMessage) map[string]interface{} {
	client, _, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	cctx, cancel := ctx()
	defer cancel()
	body, _, err := client.Core(cctx, http.MethodGet, "/priority", nil, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}

func handleListStatuses(raw json.RawMessage) map[string]interface{} {
	client, _, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	cctx, cancel := ctx()
	defer cancel()
	body, _, err := client.Core(cctx, http.MethodGet, "/status", nil, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}

func handleListIssueLinkTypes(raw json.RawMessage) map[string]interface{} {
	client, _, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	cctx, cancel := ctx()
	defer cancel()
	body, _, err := client.Core(cctx, http.MethodGet, "/issueLinkType", nil, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}

func handleLinkIssues(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	linkType := strArg(m, "link_type")
	inwardKey := strArg(m, "inward_issue_key")
	outwardKey := strArg(m, "outward_issue_key")
	if linkType == "" || inwardKey == "" || outwardKey == "" {
		return mcp.ToolResultError("link_type, inward_issue_key, and outward_issue_key are required")
	}

	payload := map[string]interface{}{
		"type":         map[string]interface{}{"name": linkType},
		"inwardIssue":  map[string]interface{}{"key": inwardKey},
		"outwardIssue": map[string]interface{}{"key": outwardKey},
	}
	if comment := strArg(m, "comment"); comment != "" {
		payload["comment"] = map[string]interface{}{
			"body": jira.PlainTextToADF(comment),
		}
	}

	cctx, cancel := ctx()
	defer cancel()
	_, status, err := client.Core(cctx, http.MethodPost, "/issueLink", nil, payload)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return mcp.ToolResultText("linked " + inwardKey + " to " + outwardKey + " (" + linkType + ") [HTTP " + strconv.Itoa(status) + "]")
}
