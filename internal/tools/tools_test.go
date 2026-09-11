package tools_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/datumbridge/jira-mcp/internal/tools"
)

func clearToolFilterEnv(t *testing.T) {
	t.Helper()
	t.Setenv("JIRA_TOOLS", "")
}

func TestRegisterFullScopeByDefault(t *testing.T) {
	clearToolFilterEnv(t)
	descs, handlers := tools.Register()
	if len(descs) < 20 {
		t.Fatalf("expected full tool scope (20+ tools), got %d", len(descs))
	}
	for _, want := range []string{
		"jira_get_issue", "jira_search_issues", "jira_create_issue", "jira_update_issue", "jira_delete_issue",
		"jira_list_comments", "jira_add_comment", "jira_update_comment", "jira_delete_comment",
		"jira_list_transitions", "jira_transition_issue",
		"jira_list_worklogs", "jira_add_worklog",
		"jira_list_projects", "jira_get_project", "jira_list_issue_types",
		"jira_get_myself", "jira_search_assignable_users",
		"jira_list_boards", "jira_get_board", "jira_list_sprints", "jira_get_sprint",
		"jira_list_backlog_issues", "jira_list_sprint_issues",
		"jira_create_sprint", "jira_update_sprint", "jira_move_issues_to_sprint",
	} {
		if _, ok := handlers[want]; !ok {
			t.Fatalf("missing handler %s", want)
		}
	}
}

func TestRegisterCustomAllowlist(t *testing.T) {
	clearToolFilterEnv(t)
	t.Setenv("JIRA_TOOLS", "jira_get_issue,jira_list_projects")
	descs, handlers := tools.Register()
	if len(descs) != 2 {
		t.Fatalf("expected 2 allowlisted tools, got %d", len(descs))
	}
	if _, ok := handlers["jira_get_issue"]; !ok {
		t.Fatal("jira_get_issue should be enabled")
	}
	if _, ok := handlers["jira_create_issue"]; ok {
		t.Fatal("jira_create_issue should be filtered out")
	}
}

func testCreds(t *testing.T, srv *httptest.Server) json.RawMessage {
	t.Helper()
	t.Setenv("JIRA_API_BASE", srv.URL)
	creds := `{"access_token":"test-token","cloud_id":"site-1"}`
	raw, _ := json.Marshal(map[string]string{"credentials_json": creds})
	return raw
}

func TestHandleGetIssueSuccess(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/site-1/rest/api/3/issue/PROJ-1" {
			t.Errorf("path %q", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"key":"PROJ-1","fields":{"summary":"Test"}}`))
	}))
	defer srv.Close()

	args := testCreds(t, srv)
	var m map[string]interface{}
	_ = json.Unmarshal(args, &m)
	m["issue_key"] = "PROJ-1"
	raw, _ := json.Marshal(m)

	_, handlers := tools.Register()
	result := handlers["jira_get_issue"](raw)
	if result["isError"] == true {
		t.Fatalf("unexpected error result: %+v", result)
	}
}

func TestHandleGetIssueMissingKey(t *testing.T) {
	clearToolFilterEnv(t)
	_, handlers := tools.Register()
	result := handlers["jira_get_issue"](json.RawMessage(`{"credentials_json":"{\"access_token\":\"t\",\"cloud_id\":\"s\"}"}`))
	if result["isError"] != true {
		t.Fatalf("expected error for missing issue_key, got %+v", result)
	}
}

func TestHandleCreateIssue(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := map[string]interface{}{}
		_ = json.NewDecoder(r.Body).Decode(&body)
		fields, _ := body["fields"].(map[string]interface{})
		if fields["summary"] != "New bug" {
			t.Errorf("summary = %v", fields["summary"])
		}
		project, _ := fields["project"].(map[string]interface{})
		if project["key"] != "PROJ" {
			t.Errorf("project key = %v", project["key"])
		}
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"key":"PROJ-99"}`))
	}))
	defer srv.Close()

	args := testCreds(t, srv)
	var m map[string]interface{}
	_ = json.Unmarshal(args, &m)
	m["project_key"] = "PROJ"
	m["issue_type"] = "Bug"
	m["summary"] = "New bug"
	m["description"] = "Steps to reproduce"
	raw, _ := json.Marshal(m)

	_, handlers := tools.Register()
	result := handlers["jira_create_issue"](raw)
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
}

func TestHandleSearchIssuesRequiresJQL(t *testing.T) {
	clearToolFilterEnv(t)
	_, handlers := tools.Register()
	result := handlers["jira_search_issues"](json.RawMessage(`{"credentials_json":"{\"access_token\":\"t\",\"cloud_id\":\"s\"}"}`))
	if result["isError"] != true {
		t.Fatalf("expected error for missing jql, got %+v", result)
	}
}

func TestHandleTransitionIssue(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/transitions") {
			t.Errorf("path %q", r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer srv.Close()

	args := testCreds(t, srv)
	var m map[string]interface{}
	_ = json.Unmarshal(args, &m)
	m["issue_key"] = "PROJ-1"
	m["transition_id"] = "31"
	raw, _ := json.Marshal(m)

	_, handlers := tools.Register()
	result := handlers["jira_transition_issue"](raw)
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
}

func TestHandleListBoards(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !strings.Contains(r.URL.Path, "/rest/agile/1.0/board") {
			t.Errorf("path %q", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"values":[{"id":1,"name":"Board 1"}]}`))
	}))
	defer srv.Close()

	args := testCreds(t, srv)
	_, handlers := tools.Register()
	result := handlers["jira_list_boards"](args)
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
}

func TestHandleMoveIssuesToSprintRequiresIssues(t *testing.T) {
	clearToolFilterEnv(t)
	_, handlers := tools.Register()
	result := handlers["jira_move_issues_to_sprint"](json.RawMessage(`{"credentials_json":"{\"access_token\":\"t\",\"cloud_id\":\"s\"}","sprint_id":1}`))
	if result["isError"] != true {
		t.Fatalf("expected error for missing issue_keys, got %+v", result)
	}
}

func TestHandleGetIssueProviderError(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"errorMessages":["Issue does not exist"]}`))
	}))
	defer srv.Close()

	args := testCreds(t, srv)
	var m map[string]interface{}
	_ = json.Unmarshal(args, &m)
	m["issue_key"] = "NOPE-1"
	raw, _ := json.Marshal(m)

	_, handlers := tools.Register()
	result := handlers["jira_get_issue"](raw)
	if result["isError"] != true {
		t.Fatalf("expected error result for 404, got %+v", result)
	}
}
