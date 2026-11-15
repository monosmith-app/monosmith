package githubrelease

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestDispatchWorkflowPostsTheRefAndInputs(t *testing.T) {
	var path, auth string
	var body map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path, auth = r.Method+" "+r.URL.Path, r.Header.Get("Authorization")
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &body)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	c := &Client{Token: "tok", API: server.URL, HTTP: server.Client()}
	err := c.DispatchWorkflow(Repository{Owner: "acme", Name: "tool"}, "pub.yml", "refs/tags/tool-v1.0.0", map[string]string{"project": "tool", "version": "1.0.0"})
	if err != nil {
		t.Fatal(err)
	}
	if path != "POST /repos/acme/tool/actions/workflows/pub.yml/dispatches" || auth != "Bearer tok" {
		t.Errorf("request = %s, auth %q", path, auth)
	}
	inputs, _ := body["inputs"].(map[string]any)
	if body["ref"] != "refs/tags/tool-v1.0.0" || inputs["project"] != "tool" || inputs["version"] != "1.0.0" {
		t.Errorf("body = %v", body)
	}
}

func TestDispatchWorkflowReportsARefusal(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"message":"Resource not accessible by integration"}`, http.StatusForbidden)
	}))
	defer server.Close()
	c := &Client{Token: "tok", API: server.URL, HTTP: server.Client()}
	if err := c.DispatchWorkflow(Repository{Owner: "acme", Name: "tool"}, "pub.yml", "main", nil); err == nil {
		t.Error("a 403 is not an error")
	}
}

func TestDownloadAssetWritesTheFile(t *testing.T) {
	var accept string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/repos/acme/tool/releases/assets/42" {
			http.NotFound(w, r)
			return
		}
		accept = r.Header.Get("Accept")
		_, _ = w.Write([]byte("zip bytes"))
	}))
	defer server.Close()
	c := &Client{Token: "tok", API: server.URL, HTTP: server.Client()}
	to := filepath.Join(t.TempDir(), "dist", "tool_1.0.0_linux_amd64.zip")
	if err := c.DownloadAsset(Repository{Owner: "acme", Name: "tool"}, Asset{ID: 42, Name: "tool_1.0.0_linux_amd64.zip"}, to); err != nil {
		t.Fatal(err)
	}
	if raw, _ := os.ReadFile(to); string(raw) != "zip bytes" || accept != "application/octet-stream" {
		t.Errorf("file = %q, accept %q", raw, accept)
	}
	if err := c.DownloadAsset(Repository{Owner: "acme", Name: "tool"}, Asset{ID: 7, Name: "x"}, to); err == nil {
		t.Error("a missing asset is not an error")
	}
}
