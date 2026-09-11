package tools

import (
	"os"
	"strings"
)

// ServerConfig controls tool visibility. Unlike Bright-Data-MCP's Rapid/Pro
// split, jira-mcp ships full tool scope by default (BR6 in
// doc/MCP_TOOL_REFERENCE_ARCHITECTURE.md) — JIRA_TOOLS is an optional narrowing
// allowlist, not an opt-in to unlock more tools.
type ServerConfig struct {
	CustomTools []string
}

func LoadServerConfig() ServerConfig {
	cfg := ServerConfig{}
	if v := strings.TrimSpace(os.Getenv("JIRA_TOOLS")); v != "" {
		cfg.CustomTools = splitCSV(v)
	}
	return cfg
}

func splitCSV(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, strings.ToLower(p))
		}
	}
	return out
}

func (c ServerConfig) toolEnabled(name string) bool {
	if len(c.CustomTools) == 0 {
		return true
	}
	for _, t := range c.CustomTools {
		if t == strings.ToLower(name) {
			return true
		}
	}
	return false
}
