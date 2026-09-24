package tools

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/datumbridge/jira-mcp/internal/mcp"
)

func registerWebhookTools(add toolAdder) {
	add("jira_list_webhooks",
		"List dynamic webhooks registered by the app (manage:jira-webhook).",
		baseProps(nil), nil, handleListWebhooks)

	add("jira_delete_webhooks",
		"Delete dynamic webhooks by id list (manage:jira-webhook).",
		baseProps(map[string]interface{}{
			"webhook_ids": map[string]interface{}{
				"type":  "array",
				"items": map[string]interface{}{"type": "integer"},
			},
		}),
		[]string{"webhook_ids"}, handleDeleteWebhooks)
}

func handleListWebhooks(raw json.RawMessage) map[string]interface{} {
	client, _, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	cctx, cancel := ctx()
	defer cancel()
	body, _, err := client.Core(cctx, http.MethodGet, "/webhook", nil, nil)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return rawJSONResult(body)
}

func handleDeleteWebhooks(raw json.RawMessage) map[string]interface{} {
	client, m, err := clientFrom(raw)
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	ids := intSliceArg(m, "webhook_ids")
	if len(ids) == 0 {
		return mcp.ToolResultError("webhook_ids must be a non-empty array of integer IDs")
	}
	cctx, cancel := ctx()
	defer cancel()
	_, status, err := client.Core(cctx, http.MethodDelete, "/webhook", nil, map[string]interface{}{
		"webhookIds": ids,
	})
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return mcp.ToolResultText("deleted " + strconv.Itoa(len(ids)) + " webhook(s) [HTTP " + strconv.Itoa(status) + "]")
}
