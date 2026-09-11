package jira_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/datumbridge/jira-mcp/internal/jira"
)

func TestParseCredentialsFromJSON(t *testing.T) {
	c, err := jira.ParseCredentials(`{"access_token":"secrettoken123","cloud_id":"site-abc"}`, "")
	if err != nil {
		t.Fatal(err)
	}
	if c.AccessToken != "secrettoken123" || c.CloudID != "site-abc" {
		t.Fatalf("unexpected creds: %+v", c)
	}
}

func TestParseCredentialsEnvFallback(t *testing.T) {
	t.Setenv("JIRA_ACCESS_TOKEN", "envtoken")
	t.Setenv("JIRA_CLOUD_ID", "env-site")
	c, err := jira.ParseCredentials("", "")
	if err != nil {
		t.Fatal(err)
	}
	if c.AccessToken != "envtoken" || c.CloudID != "env-site" {
		t.Fatalf("unexpected: %+v", c)
	}
}

func TestParseCredentialsMissingCloudID(t *testing.T) {
	if _, err := jira.ParseCredentials(`{"access_token":"t"}`, ""); err == nil {
		t.Fatal("expected error for missing cloud_id")
	}
}

func TestParseCredentialsMissingAccessToken(t *testing.T) {
	if _, err := jira.ParseCredentials(`{"cloud_id":"site-abc"}`, ""); err == nil {
		t.Fatal("expected error for missing access_token")
	}
}

func TestParseCredentialsPathJail(t *testing.T) {
	root := t.TempDir()
	t.Setenv("JIRA_CREDENTIALS_DIR", root)
	path := filepath.Join(root, "creds.json")
	if err := os.WriteFile(path, []byte(`{"access_token":"fromfile","cloud_id":"site-abc"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	c, err := jira.ParseCredentials("", path)
	if err != nil {
		t.Fatal(err)
	}
	if c.AccessToken != "fromfile" {
		t.Fatalf("got %q", c.AccessToken)
	}

	outside := filepath.Join(t.TempDir(), "escape.json")
	_ = os.WriteFile(outside, []byte(`{"access_token":"x","cloud_id":"y"}`), 0o600)
	if _, err := jira.ParseCredentials("", outside); err == nil {
		t.Fatal("expected jail error")
	}
}

func TestMaskToken(t *testing.T) {
	if jira.MaskToken("") != "(missing)" {
		t.Fatal("empty")
	}
	m := jira.MaskToken("abcdefghijklmnop")
	if m == "abcdefghijklmnop" || !strings.Contains(m, "…") {
		t.Fatalf("mask failed: %s", m)
	}
}
