package jira

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

// DefaultAPIBase is the Jira Cloud API gateway; {cloudId} is filled per-connection
// from the injected bundle (see Credentials.CloudID).
const DefaultAPIBase = "https://api.atlassian.com/ex/jira"

// Client calls Jira Cloud's core (/rest/api/3) and Agile (/rest/agile/1.0) REST
// APIs, authenticated with the connecting user's OAuth 2.0 (3LO) access token.
type Client struct {
	creds      *Credentials
	httpClient *http.Client
	apiBase    string
}

func NewClient(creds *Credentials) *Client {
	timeoutSec := 30
	if v := strings.TrimSpace(os.Getenv("JIRA_HTTP_TIMEOUT_SEC")); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			timeoutSec = n
		}
	}
	base := strings.TrimSpace(os.Getenv("JIRA_API_BASE"))
	if base == "" {
		base = DefaultAPIBase
	}
	return &Client{
		creds:      creds,
		httpClient: &http.Client{Timeout: time.Duration(timeoutSec) * time.Second},
		apiBase:    strings.TrimRight(base, "/"),
	}
}

func (c *Client) coreBase() string {
	return fmt.Sprintf("%s/%s/rest/api/3", c.apiBase, c.creds.CloudID)
}

func (c *Client) agileBase() string {
	return fmt.Sprintf("%s/%s/rest/agile/1.0", c.apiBase, c.creds.CloudID)
}

// Core calls the Jira core REST API (issues, comments, projects, users, …).
func (c *Client) Core(ctx context.Context, method, path string, query url.Values, body interface{}) ([]byte, int, error) {
	return c.do(ctx, c.coreBase(), method, path, query, body)
}

// Agile calls the Jira Software Agile REST API (boards, sprints, backlog).
func (c *Client) Agile(ctx context.Context, method, path string, query url.Values, body interface{}) ([]byte, int, error) {
	return c.do(ctx, c.agileBase(), method, path, query, body)
}

func (c *Client) do(ctx context.Context, base, method, path string, query url.Values, body interface{}) ([]byte, int, error) {
	if c == nil || c.creds == nil {
		return nil, 0, fmt.Errorf("jira client not configured")
	}
	full := base + path
	if len(query) > 0 {
		full += "?" + query.Encode()
	}

	var reader io.Reader
	if body != nil {
		payload, err := json.Marshal(body)
		if err != nil {
			return nil, 0, fmt.Errorf("encode request body: %w", err)
		}
		reader = bytes.NewReader(payload)
	}

	req, err := http.NewRequestWithContext(ctx, method, full, reader)
	if err != nil {
		return nil, 0, err
	}
	req.Header.Set("Authorization", "Bearer "+c.creds.AccessToken)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, 0, fmt.Errorf("jira request failed: %w", err)
	}
	defer resp.Body.Close()

	limited := io.LimitReader(resp.Body, 10<<20) // 10 MiB cap
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, resp.StatusCode, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := sanitizeProviderError(string(data))
		return data, resp.StatusCode, fmt.Errorf("jira HTTP %d: %s", resp.StatusCode, msg)
	}
	return data, resp.StatusCode, nil
}

// AccessTokenMasked returns a redacted access token preview.
func (c *Client) AccessTokenMasked() string {
	if c == nil || c.creds == nil {
		return "(missing)"
	}
	return MaskToken(c.creds.AccessToken)
}

// CloudID returns the connected Jira Cloud site id.
func (c *Client) CloudID() string {
	if c == nil || c.creds == nil {
		return ""
	}
	return c.creds.CloudID
}

func sanitizeProviderError(body string) string {
	body = strings.TrimSpace(body)
	if body == "" {
		return "(empty body)"
	}
	if len(body) > 500 {
		body = body[:500] + "…"
	}
	lower := strings.ToLower(body)
	if strings.Contains(lower, "bearer ") {
		return "(provider error redacted)"
	}
	return body
}

// PlainTextToADF wraps plain text as a minimal single-paragraph Atlassian
// Document Format node — Jira Cloud v3's create/update issue and comment bodies
// require ADF, not plain strings, for any rich-text field (description, comment
// body, …).
func PlainTextToADF(text string) map[string]interface{} {
	return map[string]interface{}{
		"type":    "doc",
		"version": 1,
		"content": []map[string]interface{}{
			{
				"type": "paragraph",
				"content": []map[string]interface{}{
					{"type": "text", "text": text},
				},
			},
		},
	}
}
