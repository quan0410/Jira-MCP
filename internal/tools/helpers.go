package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/datumbridge/jira-mcp/internal/jira"
	"github.com/datumbridge/jira-mcp/internal/mcp"
)

// toolAdder registers one tool's descriptor + handler with the registry.
type toolAdder func(name, desc string, props map[string]interface{}, required []string, h mcp.ToolHandler)

func baseProps(extra map[string]interface{}) map[string]interface{} {
	props := map[string]interface{}{
		"credentials_json": map[string]interface{}{
			"type":        "string",
			"description": "Vault JSON: {access_token, cloud_id}",
		},
		"credentials_path": map[string]interface{}{
			"type":        "string",
			"description": "Optional path under JIRA_CREDENTIALS_DIR",
		},
	}
	for k, v := range extra {
		props[k] = v
	}
	return props
}

func schema(props map[string]interface{}, required []string) map[string]interface{} {
	s := map[string]interface{}{
		"type":       "object",
		"properties": props,
	}
	if len(required) > 0 {
		s["required"] = required
	}
	return s
}

func clientFrom(raw json.RawMessage) (*jira.Client, map[string]interface{}, error) {
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		m = map[string]interface{}{}
	}
	credsJSON, _ := m["credentials_json"].(string)
	credsPath, _ := m["credentials_path"].(string)
	creds, err := jira.ParseCredentials(credsJSON, credsPath)
	if err != nil {
		return nil, m, err
	}
	return jira.NewClient(creds), m, nil
}

func strArg(m map[string]interface{}, key string) string {
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return strings.TrimSpace(s)
	}
	return strings.TrimSpace(fmt.Sprint(v))
}

func intArg(m map[string]interface{}, key string, def int) int {
	v, ok := m[key]
	if !ok || v == nil {
		return def
	}
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case json.Number:
		i, _ := t.Int64()
		return int(i)
	case string:
		var i int
		if _, err := fmt.Sscanf(t, "%d", &i); err == nil {
			return i
		}
	}
	return def
}

func strSliceArg(m map[string]interface{}, key string) []string {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	out := make([]string, 0, len(arr))
	for _, e := range arr {
		s, _ := e.(string)
		s = strings.TrimSpace(s)
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}

func intSliceArg(m map[string]interface{}, key string) []int {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	arr, ok := v.([]interface{})
	if !ok {
		return nil
	}
	out := make([]int, 0, len(arr))
	for _, e := range arr {
		switch t := e.(type) {
		case float64:
			out = append(out, int(t))
		case int:
			out = append(out, t)
		case json.Number:
			if i, err := t.Int64(); err == nil {
				out = append(out, int(i))
			}
		}
	}
	return out
}

func boolPtrArg(m map[string]interface{}, key string) *bool {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	if b, ok := v.(bool); ok {
		return &b
	}
	return nil
}

func mapArg(m map[string]interface{}, key string) map[string]interface{} {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	obj, _ := v.(map[string]interface{})
	return obj
}

func ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 30*time.Second)
}

func jsonResult(v interface{}) map[string]interface{} {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return mcp.ToolResultError(err.Error())
	}
	return mcp.ToolResultText(string(b))
}
