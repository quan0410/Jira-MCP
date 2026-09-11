package tools

import (
	"github.com/datumbridge/jira-mcp/internal/mcp"
)

// Register returns the full jira-mcp tool catalogue: issues, comments,
// transitions, worklogs, projects, users (core Jira Cloud REST API v3), plus
// boards/sprints/backlog (Jira Software Agile REST API). Full scope is the
// default (BR6); JIRA_TOOLS narrows it to an explicit allowlist if set.
func Register() ([]mcp.ToolDesc, map[string]mcp.ToolHandler) {
	cfg := LoadServerConfig()
	handlers := map[string]mcp.ToolHandler{}
	var descs []mcp.ToolDesc

	add := func(name, desc string, props map[string]interface{}, required []string, h mcp.ToolHandler) {
		if !cfg.toolEnabled(name) {
			return
		}
		descs = append(descs, mcp.ToolDesc{
			Name:        name,
			Description: desc,
			InputSchema: schema(props, required),
		})
		handlers[name] = h
	}

	registerIssueTools(add)
	registerCommentTools(add)
	registerTransitionTools(add)
	registerWorklogTools(add)
	registerProjectTools(add)
	registerUserTools(add)
	registerAgileTools(add)

	return descs, handlers
}
