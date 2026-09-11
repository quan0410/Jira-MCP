package jira_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/datumbridge/jira-mcp/internal/jira"
)

func newTestClient(t *testing.T, srv *httptest.Server) *jira.Client {
	t.Helper()
	t.Setenv("JIRA_API_BASE", srv.URL)
	return jira.NewClient(&jira.Credentials{AccessToken: "test-token", CloudID: "site-1"})
}

func TestClientCoreSuccess(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("auth header %q", r.Header.Get("Authorization"))
		}
		if r.URL.Path != "/site-1/rest/api/3/issue/PROJ-1" {
			t.Errorf("path %q", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"key":"PROJ-1"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	data, status, err := c.Core(context.Background(), http.MethodGet, "/issue/PROJ-1", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 || !strings.Contains(string(data), "PROJ-1") {
		t.Fatalf("status=%d data=%q", status, data)
	}
}

func TestClientCoreHTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"errorMessages":["Issue does not exist"]}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, _, err := c.Core(context.Background(), http.MethodGet, "/issue/NOPE-1", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "404") {
		t.Fatalf("expected 404 error, got %v", err)
	}
}

func TestClientCorePostBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/site-1/rest/api/3/issue" {
			t.Errorf("path %q", r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		if !strings.Contains(string(body), `"summary":"Test issue"`) {
			t.Errorf("body %s", body)
		}
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"key":"PROJ-2"}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, status, err := c.Core(context.Background(), http.MethodPost, "/issue", nil, map[string]interface{}{
		"fields": map[string]interface{}{"summary": "Test issue"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if status != 201 {
		t.Fatalf("status=%d", status)
	}
}

func TestClientAgilePath(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/site-1/rest/agile/1.0/board/42/sprint" {
			t.Errorf("path %q", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	defer srv.Close()

	c := newTestClient(t, srv)
	_, status, err := c.Agile(context.Background(), http.MethodGet, "/board/42/sprint", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if status != 200 {
		t.Fatalf("status=%d", status)
	}
}

func TestPlainTextToADF(t *testing.T) {
	adf := jira.PlainTextToADF("hello")
	if adf["type"] != "doc" || adf["version"] != 1 {
		t.Fatalf("unexpected ADF envelope: %+v", adf)
	}
}
