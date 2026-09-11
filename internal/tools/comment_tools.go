package tools

import (
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/datumbridge/jira-mcp/internal/jira"
	"github.com/datumbridge/jira-mcp/internal/mcp"
)

func registerCommentTools(add toolAdder) {
	add("jira_list_comments",
		"List comments on an issue.",
		baseProps(map[string]interface{}{
			"issue_key":   map[string]interface{}{"type": "string"},
			"max_results": map[string]interface{}{"type": "integer", "default": 50},
		}),
		[]string{"issue_key"}, handleListComments)

	add("jira_add_comment",
		"Add a comment to an issue.",
		baseProps(map[string]interface{}{
			"issue_key": map[string]interface{}{"type": "string"},
			"body":      map[string]interface{}{"type": "string", "description": "Plain text; converted to Atlassian Document Format"},
		}),
		[]string{"issue_key", "body"}, handleAddComment)

	add("jira_update_comment",
		"Update an existing comment's body.",
		baseProps(map[string]interface{}{
			"issue_key":  map[string]interface{}{"type": "string"},
			"comment_id": map[string]interface{}{"type": "string"},
			"body":       map[string]interface{}{"type": "string"},
		}),
		[]string{"issue_key", "comment_id", "body"}, handleUpdateComment)

	add("jira_delete_comment",
		"Delete a comment.",
		baseProps(map[string]interface{}{
			"issue_key":  map[string]interface{}{"type": "string"},
			"comment_id": map[string]interface{}{"type": "string"},
		}),
		[]string{"issue_key", "comment_id"}, handleDeleteComment)
}

func handleListComments(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	issueKey := strArg(m, "issue_key")
	if issueKey == "" {
		return mcp.ToolResultError("issue_key is required")
	}
	q := url.Values{}
	if mr := intArg(m, "max_results", 0); mr > 0 {
		q.Set("maxResults", strconv.Itoa(mr))
	}
	cctx, cancel := ctx()
	defer cancel()
	body, _, err := client.Core(cctx, http.MethodGet, "/issue/"+url.PathEscape(issueKey)+"/comment", q, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}

func handleAddComment(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	issueKey := strArg(m, "issue_key")
	text := strArg(m, "body")
	if issueKey == "" || text == "" {
		return mcp.ToolResultError("issue_key and body are required")
	}
	cctx, cancel := ctx()
	defer cancel()
	respBody, _, err := client.Core(cctx, http.MethodPost, "/issue/"+url.PathEscape(issueKey)+"/comment", nil,
		map[string]interface{}{"body": jira.PlainTextToADF(text)})
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(respBody)
}

func handleUpdateComment(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	issueKey := strArg(m, "issue_key")
	commentID := strArg(m, "comment_id")
	text := strArg(m, "body")
	if issueKey == "" || commentID == "" || text == "" {
		return mcp.ToolResultError("issue_key, comment_id and body are required")
	}
	cctx, cancel := ctx()
	defer cancel()
	respBody, _, err := client.Core(cctx, http.MethodPut,
		"/issue/"+url.PathEscape(issueKey)+"/comment/"+url.PathEscape(commentID), nil,
		map[string]interface{}{"body": jira.PlainTextToADF(text)})
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(respBody)
}

func handleDeleteComment(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	issueKey := strArg(m, "issue_key")
	commentID := strArg(m, "comment_id")
	if issueKey == "" || commentID == "" {
		return mcp.ToolResultError("issue_key and comment_id are required")
	}
	cctx, cancel := ctx()
	defer cancel()
	_, status, err := client.Core(cctx, http.MethodDelete,
		"/issue/"+url.PathEscape(issueKey)+"/comment/"+url.PathEscape(commentID), nil, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return mcp.ToolResultText("deleted comment " + commentID + " (HTTP " + strconv.Itoa(status) + ")")
}
