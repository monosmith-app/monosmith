package cachecommand

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/remotesync"
)

// objectServer imitates a plain HTTP object store that wants a bearer token.
func objectServer(t *testing.T) (*httptest.Server, *int) {
	var mu sync.Mutex
	objects, puts := map[string][]byte{}, 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer team-token" {
			w.WriteHeader(http.StatusForbidden)
			return
		}
		mu.Lock()
		defer mu.Unlock()
		switch r.Method {
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			objects[r.URL.Path] = body
			puts++
		default:
			if _, ok := objects[r.URL.Path]; !ok {
				w.WriteHeader(http.StatusNotFound)
			}
		}
	}))
	t.Cleanup(srv.Close)
	return srv, &puts
}

func withRemote(t *testing.T, srv *httptest.Server, token string) {
	t.Setenv("SMITH_CACHE_REMOTE", srv.URL+"/team")
	t.Setenv("SMITH_CACHE_KEY", "team-key")
	t.Setenv("SMITH_CACHE_TOKEN", token)
	t.Setenv("SMITH_CACHE_READ_ONLY", "")
	for _, ci := range []string{"GITHUB_EVENT_NAME", "BUILD_REASON", "CI_PIPELINE_SOURCE", "BITBUCKET_PR_ID", "CIRCLE_PULL_REQUEST", "CIRCLE_PR_NUMBER"} {
		t.Setenv(ci, "")
	}
}

func TestPushCommandDryRunThenPushThenNothingLeft(t *testing.T) {
	workspaceWithCache(t, "", 2, 1000)
	srv, puts := objectServer(t)
	withRemote(t, srv, "team-token")
	host := strings.TrimPrefix(srv.URL, "http://")

	out, err := run(t, "push", "--dry-run")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Asking http://" + host + "/team which local entries it lacks (dry run: nothing is uploaded).", "  ○ a:build  would upload (2 files,", "✓ Dry run: 2 of 2 entries would be uploaded to http://" + host + "/team", "Nothing was uploaded.", "→ run smith cache push to upload them"} {
		if !strings.Contains(out, want) {
			t.Errorf("dry run: missing %q in:\n%s", want, out)
		}
	}
	if *puts != 0 {
		t.Fatalf("the dry run wrote %d objects", *puts)
	}

	out, err = run(t, "push")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"  ✓ a:build  uploaded (2 files,", "✓ Uploaded 2 entries to http://" + host + "/team", "0 already on the remote."} {
		if !strings.Contains(out, want) {
			t.Errorf("push: missing %q in:\n%s", want, out)
		}
	}

	out, err = run(t, "push", "--json")
	if err != nil {
		t.Fatal(err)
	}
	var result remotesync.PushResult
	if err := json.Unmarshal([]byte(out), &result); err != nil {
		t.Fatalf("%v in:\n%s", err, out)
	}
	if result.Entries != 2 || result.Present != 2 || result.Uploaded != 0 {
		t.Errorf("third push = %+v", result)
	}
}

func TestPushCommandNamesTheRemedyForARefusedToken(t *testing.T) {
	workspaceWithCache(t, "", 1, 10)
	srv, _ := objectServer(t)
	withRemote(t, srv, "old-token")
	_, err := run(t, "push")
	if err == nil || !strings.Contains(err.Error(), "refused the credentials") || !strings.Contains(err.Error(), "→ check SMITH_CACHE_TOKEN") {
		t.Fatalf("err = %v", err)
	}
}

func TestPushCommandWithoutARemoteSaysHowToSetOne(t *testing.T) {
	workspaceWithCache(t, "", 1, 10)
	_, err := run(t, "push")
	if err == nil || !strings.Contains(err.Error(), "smith cache setup") {
		t.Fatalf("err = %v", err)
	}
}
