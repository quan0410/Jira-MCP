package tools_test

// Acceptance-criteria coverage for doc/MCP_TOOL_REFERENCE_ARCHITECTURE.md's
// EC1-EC7 and jira-mcp/docs/business/business-rules.md's BR-J1-BR-J6, beyond
// the representative handlers already covered in tools_test.go. Each
// previously-untested tool gets at least one request-shape assertion so a
// regression in path/method/body construction fails loudly.

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/datumbridge/jira-mcp/internal/tools"
)

func callTool(t *testing.T, name string, srv *httptest.Server, extra map[string]interface{}) map[string]interface{} {
	t.Helper()
	args := testCreds(t, srv)
	var m map[string]interface{}
	_ = json.Unmarshal(args, &m)
	for k, v := range extra {
		m[k] = v
	}
	raw, _ := json.Marshal(m)
	_, handlers := tools.Register()
	h, ok := handlers[name]
	if !ok {
		t.Fatalf("no handler registered for %s", name)
	}
	return h(raw)
}

// --- EC4: expired/revoked token surfaces as the provider's own 401, not a
// retry loop or a masked error. ---
func TestHandleGetIssue401NotRetried(t *testing.T) {
	clearToolFilterEnv(t)
	calls := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(401)
		_, _ = w.Write([]byte(`{"errorMessages":["Unauthorized"]}`))
	}))
	defer srv.Close()

	result := callTool(t, "jira_get_issue", srv, map[string]interface{}{"issue_key": "PROJ-1"})
	if result["isError"] != true {
		t.Fatalf("expected error result for 401, got %+v", result)
	}
	if calls != 1 {
		t.Fatalf("expected exactly 1 upstream call (no client-side retry), got %d", calls)
	}
}

// --- EC5: 429 must be distinguishable from an auth error in the surfaced
// message, so an agent doesn't conclude "needs reconnect". ---
func TestHandleGetIssue429DistinctFromAuthError(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		_, _ = w.Write([]byte(`{"errorMessages":["Rate limit exceeded"]}`))
	}))
	defer srv.Close()

	result := callTool(t, "jira_get_issue", srv, map[string]interface{}{"issue_key": "PROJ-1"})
	if result["isError"] != true {
		t.Fatalf("expected error result for 429, got %+v", result)
	}
	text := toolResultText(t, result)
	if !strings.Contains(text, "429") {
		t.Fatalf("429 error message must carry the status so it's distinguishable from a 401: %q", text)
	}
	if strings.Contains(text, "401") {
		t.Fatalf("429 must not be conflated with 401: %q", text)
	}
}

// --- EC1 (this repo's boundary of it): no credentials at all produces a
// clear config-class error, not a vague runtime exception. ---
func TestHandleGetIssueNoCredentialsAtAll(t *testing.T) {
	clearToolFilterEnv(t)
	t.Setenv("JIRA_ACCESS_TOKEN", "")
	t.Setenv("JIRA_CLOUD_ID", "")
	_, handlers := tools.Register()
	result := handlers["jira_get_issue"](json.RawMessage(`{"issue_key":"PROJ-1"}`))
	if result["isError"] != true {
		t.Fatalf("expected error result, got %+v", result)
	}
	text := toolResultText(t, result)
	if !strings.Contains(text, "access_token") {
		t.Fatalf("expected an actionable message naming the missing credential, got %q", text)
	}
}

// --- BR-J5: sprint update only sends the fields the caller actually
// provided (partial update), never clears omitted ones. ---
func TestHandleUpdateSprintPartialBody(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("method = %s, want POST (partial update, not PUT)", r.Method)
		}
		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		_ = json.Unmarshal(body, &payload)
		if len(payload) != 1 {
			t.Fatalf("expected exactly the 1 provided field in body, got %+v", payload)
		}
		if payload["state"] != "active" {
			t.Errorf("state = %v", payload["state"])
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"id":1,"state":"active"}`))
	}))
	defer srv.Close()

	result := callTool(t, "jira_update_sprint", srv, map[string]interface{}{
		"sprint_id": float64(1),
		"state":     "active",
	})
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
}

func TestHandleUpdateSprintRequiresAtLeastOneField(t *testing.T) {
	clearToolFilterEnv(t)
	_, handlers := tools.Register()
	result := handlers["jira_update_sprint"](json.RawMessage(`{"credentials_json":"{\"access_token\":\"t\",\"cloud_id\":\"s\"}","sprint_id":1}`))
	if result["isError"] != true {
		t.Fatalf("expected error when no field to update is given, got %+v", result)
	}
}

// --- Fix verification: jira_search_issues sends a usable default "fields"
// list when the caller omits it, instead of letting Jira return bare {id}
// objects (found via live E2E test against a real Jira Cloud site). ---
func TestHandleSearchIssuesDefaultsFieldsWhenOmitted(t *testing.T) {
	clearToolFilterEnv(t)
	var gotFields []interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		_ = json.Unmarshal(body, &payload)
		gotFields, _ = payload["fields"].([]interface{})
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"issues":[]}`))
	}))
	defer srv.Close()

	result := callTool(t, "jira_search_issues", srv, map[string]interface{}{"jql": "project = PROJ"})
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
	if len(gotFields) == 0 {
		t.Fatal("expected a non-empty default fields list to be sent when caller omits fields")
	}
	found := false
	for _, f := range gotFields {
		if f == "key" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected default fields to include \"key\", got %v", gotFields)
	}
}

func TestHandleSearchIssuesRespectsExplicitFields(t *testing.T) {
	clearToolFilterEnv(t)
	var gotFields []interface{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		_ = json.Unmarshal(body, &payload)
		gotFields, _ = payload["fields"].([]interface{})
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"issues":[]}`))
	}))
	defer srv.Close()

	result := callTool(t, "jira_search_issues", srv, map[string]interface{}{
		"jql":    "project = PROJ",
		"fields": []interface{}{"summary"},
	})
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
	if len(gotFields) != 1 || gotFields[0] != "summary" {
		t.Fatalf("expected caller-provided fields to be sent as-is, got %v", gotFields)
	}
}

// --- BR-J4: search cap. Documented as "capped at 100" — verify the actual
// behavior for an out-of-range max_results so a future spec/code mismatch is
// visible rather than silent. Current implementation resets to the 50
// default rather than clamping to 100; this test pins that behavior.
func TestHandleSearchIssuesMaxResultsOutOfRangeResetsToDefault(t *testing.T) {
	clearToolFilterEnv(t)
	var gotMaxResults float64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		_ = json.Unmarshal(body, &payload)
		gotMaxResults, _ = payload["maxResults"].(float64)
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"issues":[]}`))
	}))
	defer srv.Close()

	result := callTool(t, "jira_search_issues", srv, map[string]interface{}{
		"jql":         "project = PROJ",
		"max_results": float64(150),
	})
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
	if gotMaxResults != 50 {
		t.Fatalf("out-of-range max_results=150: got maxResults=%v sent upstream, want 50 (reset-to-default, not clamp-to-100) — "+
			"if the intent is actually clamp-to-100, this is a real mismatch against docs/business/business-rules.md BR-J4, flag to dev_ba/dev_code", gotMaxResults)
	}
}

// --- Remaining tools: one request-shape assertion each so a path/method/body
// regression fails loudly. ---

func TestHandleUpdateIssue(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/site-1/rest/api/3/issue/PROJ-1" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer srv.Close()
	result := callTool(t, "jira_update_issue", srv, map[string]interface{}{
		"issue_key": "PROJ-1",
		"fields":    map[string]interface{}{"summary": "Updated"},
	})
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
}

func TestHandleDeleteIssue(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/site-1/rest/api/3/issue/PROJ-1" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer srv.Close()
	result := callTool(t, "jira_delete_issue", srv, map[string]interface{}{"issue_key": "PROJ-1"})
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
}

func TestHandleListComments(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/site-1/rest/api/3/issue/PROJ-1/comment" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"comments":[]}`))
	}))
	defer srv.Close()
	result := callTool(t, "jira_list_comments", srv, map[string]interface{}{"issue_key": "PROJ-1"})
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
}

func TestHandleAddCommentWrapsADF(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		_ = json.Unmarshal(body, &payload)
		b, _ := payload["body"].(map[string]interface{})
		if b["type"] != "doc" {
			t.Errorf("expected comment body wrapped as ADF doc, got %+v", payload["body"])
		}
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"id":"1"}`))
	}))
	defer srv.Close()
	result := callTool(t, "jira_add_comment", srv, map[string]interface{}{"issue_key": "PROJ-1", "body": "hello"})
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
}

func TestHandleUpdateComment(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut || r.URL.Path != "/site-1/rest/api/3/issue/PROJ-1/comment/10" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"id":"10"}`))
	}))
	defer srv.Close()
	result := callTool(t, "jira_update_comment", srv, map[string]interface{}{
		"issue_key": "PROJ-1", "comment_id": "10", "body": "edited",
	})
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
}

func TestHandleDeleteComment(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete || r.URL.Path != "/site-1/rest/api/3/issue/PROJ-1/comment/10" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(204)
	}))
	defer srv.Close()
	result := callTool(t, "jira_delete_comment", srv, map[string]interface{}{"issue_key": "PROJ-1", "comment_id": "10"})
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
}

func TestHandleListTransitions(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/site-1/rest/api/3/issue/PROJ-1/transitions" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"transitions":[]}`))
	}))
	defer srv.Close()
	result := callTool(t, "jira_list_transitions", srv, map[string]interface{}{"issue_key": "PROJ-1"})
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
}

func TestHandleListWorklogs(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/site-1/rest/api/3/issue/PROJ-1/worklog" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"worklogs":[]}`))
	}))
	defer srv.Close()
	result := callTool(t, "jira_list_worklogs", srv, map[string]interface{}{"issue_key": "PROJ-1"})
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
}

func TestHandleAddWorklog(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		_ = json.Unmarshal(body, &payload)
		if payload["timeSpent"] != "3h" {
			t.Errorf("timeSpent = %v", payload["timeSpent"])
		}
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"id":"1"}`))
	}))
	defer srv.Close()
	result := callTool(t, "jira_add_worklog", srv, map[string]interface{}{"issue_key": "PROJ-1", "time_spent": "3h"})
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
}

func TestHandleListProjects(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/site-1/rest/api/3/project/search" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	defer srv.Close()
	result := callTool(t, "jira_list_projects", srv, nil)
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
}

func TestHandleGetProject(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/site-1/rest/api/3/project/PROJ" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"key":"PROJ"}`))
	}))
	defer srv.Close()
	result := callTool(t, "jira_get_project", srv, map[string]interface{}{"project_key": "PROJ"})
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
}

func TestHandleListIssueTypes(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/site-1/rest/api/3/issue/createmeta/PROJ/issuetypes" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"issueTypes":[]}`))
	}))
	defer srv.Close()
	result := callTool(t, "jira_list_issue_types", srv, map[string]interface{}{"project_key": "PROJ"})
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
}

func TestHandleGetMyself(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/site-1/rest/api/3/myself" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"accountId":"abc"}`))
	}))
	defer srv.Close()
	result := callTool(t, "jira_get_myself", srv, nil)
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
}

func TestHandleSearchAssignableUsersRequiresProjectOrIssue(t *testing.T) {
	clearToolFilterEnv(t)
	_, handlers := tools.Register()
	result := handlers["jira_search_assignable_users"](json.RawMessage(`{"credentials_json":"{\"access_token\":\"t\",\"cloud_id\":\"s\"}"}`))
	if result["isError"] != true {
		t.Fatalf("expected error when neither project_key nor issue_key given, got %+v", result)
	}
}

func TestHandleSearchAssignableUsers(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/site-1/rest/api/3/user/assignable/search" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.URL.Query().Get("project") != "PROJ" {
			t.Errorf("project query = %s", r.URL.Query().Get("project"))
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`[]`))
	}))
	defer srv.Close()
	result := callTool(t, "jira_search_assignable_users", srv, map[string]interface{}{"project_key": "PROJ"})
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
}

func TestHandleGetBoard(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/site-1/rest/agile/1.0/board/42" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"id":42}`))
	}))
	defer srv.Close()
	result := callTool(t, "jira_get_board", srv, map[string]interface{}{"board_id": float64(42)})
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
}

func TestHandleListSprints(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/site-1/rest/agile/1.0/board/42/sprint" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"values":[]}`))
	}))
	defer srv.Close()
	result := callTool(t, "jira_list_sprints", srv, map[string]interface{}{"board_id": float64(42)})
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
}

func TestHandleGetSprint(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/site-1/rest/agile/1.0/sprint/7" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"id":7}`))
	}))
	defer srv.Close()
	result := callTool(t, "jira_get_sprint", srv, map[string]interface{}{"sprint_id": float64(7)})
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
}

func TestHandleListBacklogIssues(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/site-1/rest/agile/1.0/board/42/backlog" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"issues":[]}`))
	}))
	defer srv.Close()
	result := callTool(t, "jira_list_backlog_issues", srv, map[string]interface{}{"board_id": float64(42)})
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
}

func TestHandleListSprintIssues(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/site-1/rest/agile/1.0/sprint/7/issue" {
			t.Errorf("path = %s", r.URL.Path)
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"issues":[]}`))
	}))
	defer srv.Close()
	result := callTool(t, "jira_list_sprint_issues", srv, map[string]interface{}{"sprint_id": float64(7)})
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
}

func TestHandleCreateSprint(t *testing.T) {
	clearToolFilterEnv(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/site-1/rest/agile/1.0/sprint" {
			t.Errorf("method/path = %s %s", r.Method, r.URL.Path)
		}
		body, _ := io.ReadAll(r.Body)
		var payload map[string]interface{}
		_ = json.Unmarshal(body, &payload)
		if payload["originBoardId"] != float64(42) {
			t.Errorf("originBoardId = %v", payload["originBoardId"])
		}
		w.WriteHeader(201)
		_, _ = w.Write([]byte(`{"id":8,"state":"future"}`))
	}))
	defer srv.Close()
	result := callTool(t, "jira_create_sprint", srv, map[string]interface{}{
		"board_id": float64(42), "name": "Sprint 9",
	})
	if result["isError"] == true {
		t.Fatalf("unexpected error: %+v", result)
	}
}

// toolResultText extracts the text of a tool result's first content block.
func toolResultText(t *testing.T, result map[string]interface{}) string {
	t.Helper()
	content, ok := result["content"].([]map[string]string)
	if !ok || len(content) == 0 {
		t.Fatalf("tool result has no text content: %+v", result)
	}
	return content[0]["text"]
}
