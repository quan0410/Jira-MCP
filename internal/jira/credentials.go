package jira

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Credentials is the vault/injected credentials_json shape for Jira Cloud OAuth
// 2.0 (3LO), matching datumbridge-integrations' JiraTokenBundle
// (internal/credentials/vault.go). RefreshToken/TokenURI/ClientID/ClientSecret
// arrive as part of that bundle but are never used here: refresh is
// datumbridge-integrations' responsibility (Vault.RefreshJiraIfNeeded, run before
// /internal/v1/inject hands back a bundle), not this MCP server's — see
// doc/MCP_TOOL_REFERENCE_ARCHITECTURE.md's BR9. Only AccessToken and CloudID are
// actually read.
type Credentials struct {
	Type         string   `json:"type"`
	AccessToken  string   `json:"access_token"`
	RefreshToken string   `json:"refresh_token,omitempty"`
	TokenURI     string   `json:"token_uri,omitempty"`
	ClientID     string   `json:"client_id,omitempty"`
	ClientSecret string   `json:"client_secret,omitempty"`
	CloudID      string   `json:"cloud_id"`
	Scopes       []string `json:"scopes,omitempty"`
}

// ParseCredentials loads Credentials from vault args (credentials_json /
// credentials_path), then falls back to process env for local/dev runs.
// Precedence mirrors Bright-Data-MCP's ParseCredentials: injected bundle first,
// JIRA_ACCESS_TOKEN/JIRA_CLOUD_ID env fills blanks only.
func ParseCredentials(credentialsJSON, credentialsPath string) (*Credentials, error) {
	c := &Credentials{}
	raw := strings.TrimSpace(credentialsJSON)

	if raw == "" && strings.TrimSpace(credentialsPath) != "" {
		path := filepath.Clean(credentialsPath)
		if strings.Contains(path, "..") {
			return nil, fmt.Errorf("credentials_path must not contain ..")
		}
		allowedRoot := strings.TrimSpace(os.Getenv("JIRA_CREDENTIALS_DIR"))
		if allowedRoot == "" {
			allowedRoot = "/credentials"
		}
		allowedRoot = filepath.Clean(allowedRoot)
		absPath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("credentials_path: %w", err)
		}
		absRoot, err := filepath.Abs(allowedRoot)
		if err != nil {
			return nil, fmt.Errorf("JIRA_CREDENTIALS_DIR: %w", err)
		}
		resolvedRoot, err := filepath.EvalSymlinks(absRoot)
		if err != nil {
			resolvedRoot = absRoot
		}
		resolvedPath, err := filepath.EvalSymlinks(absPath)
		if err != nil {
			return nil, fmt.Errorf("credentials_path: %w", err)
		}
		if resolvedPath != resolvedRoot && !strings.HasPrefix(resolvedPath, resolvedRoot+string(os.PathSeparator)) {
			return nil, fmt.Errorf("credentials_path must be under %s", absRoot)
		}
		b, err := os.ReadFile(resolvedPath)
		if err != nil {
			return nil, fmt.Errorf("read credentials_path: %w", err)
		}
		raw = strings.TrimSpace(string(b))
	}

	if raw != "" {
		if err := json.Unmarshal([]byte(raw), c); err != nil {
			return nil, fmt.Errorf("invalid credentials_json: %w", err)
		}
	}

	if strings.TrimSpace(c.AccessToken) == "" {
		c.AccessToken = os.Getenv("JIRA_ACCESS_TOKEN")
	}
	if strings.TrimSpace(c.CloudID) == "" {
		c.CloudID = os.Getenv("JIRA_CLOUD_ID")
	}
	c.AccessToken = strings.TrimSpace(c.AccessToken)
	c.CloudID = strings.TrimSpace(c.CloudID)

	if c.AccessToken == "" {
		return nil, fmt.Errorf("access_token required (credentials_json.access_token or JIRA_ACCESS_TOKEN)")
	}
	if c.CloudID == "" {
		return nil, fmt.Errorf("cloud_id required (credentials_json.cloud_id or JIRA_CLOUD_ID)")
	}
	return c, nil
}

// MaskToken returns a redacted preview for health/debug output.
func MaskToken(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return "(missing)"
	}
	if len(token) <= 8 {
		return "****"
	}
	return token[:4] + "…" + token[len(token)-4:]
}
