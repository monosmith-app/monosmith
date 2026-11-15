package releasepreview

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// fake records requests and answers from a script, by "METHOD path".
type fake struct {
	mu      sync.Mutex
	seen    []string
	bodies  map[string]string
	answers map[string]string
}

func (f *fake) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	key := r.Method + " " + r.URL.Path
	f.seen = append(f.seen, key)
	body, _ := io.ReadAll(r.Body)
	f.bodies[key] = string(body)
	if r.Header.Get("Authorization") != "Bearer secret" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	if a, ok := f.answers[key]; ok {
		_, _ = io.WriteString(w, a)
		return
	}
	_, _ = io.WriteString(w, "{}")
}

func serve(t *testing.T, answers map[string]string) (*fake, string) {
	f := &fake{bodies: map[string]string{}, answers: answers}
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)
	return f, srv.URL
}

func TestGitHubCreatesThenUpdates(t *testing.T) {
	f, url := serve(t, map[string]string{"GET /repos/acme/repo/issues/5/comments": `[{"id": 1, "body": "nice"}]`})
	pr := PullRequest{Provider: "github", Number: 5, Token: "secret", API: url, Repo: "acme/repo"}
	if done, err := Upsert(pr, Marker+"\nplan"); err != nil || done != "created" {
		t.Fatalf("%s %v", done, err)
	}
	if !strings.Contains(f.bodies["POST /repos/acme/repo/issues/5/comments"], "plan") {
		t.Fatalf("post body %q", f.bodies)
	}
	f.answers["GET /repos/acme/repo/issues/5/comments"] = `[{"id": 1, "body": "nice"}, {"id": 9, "body": "` + Marker + `\nold"}]`
	if done, err := Upsert(pr, Marker+"\nnew plan"); err != nil || done != "updated" {
		t.Fatalf("%s %v", done, err)
	}
	var patched map[string]string
	_ = json.Unmarshal([]byte(f.bodies["PATCH /repos/acme/repo/issues/comments/9"]), &patched)
	if !strings.Contains(patched["body"], "new plan") {
		t.Fatalf("patched %q", f.bodies)
	}
}

func TestAzureCreatesAClosedThreadThenUpdates(t *testing.T) {
	threads := "/acme/Platform/_apis/git/repositories/r1/pullRequests/7/threads"
	f, url := serve(t, map[string]string{"GET " + threads: `{"value": [{"id": 3, "comments": [{"id": 1, "content": "lgtm"}]}]}`})
	pr := PullRequest{Provider: "azure", Number: 7, Token: "secret", Collection: url + "/acme/", Project: "Platform", RepoID: "r1"}
	if done, err := Upsert(pr, Marker+"\nplan"); err != nil || done != "created" {
		t.Fatalf("%s %v", done, err)
	}
	if b := f.bodies["POST "+threads]; !strings.Contains(b, `"status":4`) || !strings.Contains(b, "plan") {
		t.Fatalf("thread %s", b)
	}
	f.answers["GET "+threads] = `{"value": [{"id": 8, "comments": [{"id": 2, "content": "` + Marker + `\nold"}]}]}`
	if done, err := Upsert(pr, Marker+"\nnew"); err != nil || done != "updated" {
		t.Fatalf("%s %v", done, err)
	}
	if !strings.Contains(f.bodies["PATCH "+threads+"/8/comments/2"], "new") {
		t.Fatalf("patch %v", f.seen)
	}
}

func TestBitbucketCreatesThenUpdates(t *testing.T) {
	comments := "/repositories/acme/repo/pullrequests/7/comments"
	f, url := serve(t, map[string]string{"GET " + comments: `{"values": [{"id": 1, "content": {"raw": "lgtm"}}]}`})
	pr := PullRequest{Provider: "bitbucket", Number: 7, Token: "secret", API: url, Repo: "acme/repo"}
	if done, err := Upsert(pr, Marker+" plan"); err != nil || done != "created" {
		t.Fatalf("%s %v", done, err)
	}
	if !strings.Contains(f.bodies["POST "+comments], `plan"}}`) {
		t.Fatalf("bodies %v", f.bodies)
	}
	f.answers["GET "+comments] = `{"values": [{"id": 4, "content": {"raw": "` + Marker + ` old"}}]}`
	if done, err := Upsert(pr, Marker+" new plan"); err != nil || done != "updated" {
		t.Fatalf("%s %v", done, err)
	}
	if !strings.Contains(f.bodies["PUT "+comments+"/4"], "new plan") {
		t.Fatalf("bodies %v", f.bodies)
	}
}

func TestAnAPIFailureIsAnError(t *testing.T) {
	_, url := serve(t, nil)
	pr := PullRequest{Provider: "github", Number: 5, Token: "wrong", API: url, Repo: "acme/repo"}
	if _, err := Upsert(pr, "x"); err == nil || !strings.Contains(err.Error(), "401") {
		t.Fatalf("%v", err)
	}
}

func TestGitLabCreatesThenUpdates(t *testing.T) {
	f, url := serve(t, map[string]string{"GET /api/v4/projects/42/merge_requests/7/notes": `[{"id": 1, "body": "lgtm"}]`})
	pr := PullRequest{Provider: "gitlab", Number: 7, Token: "secret", API: url + "/api/v4", ProjectID: "42"}
	if done, err := Upsert(pr, Marker+" plan"); err != nil || done != "created" {
		t.Fatalf("%s %v", done, err)
	}
	f.answers["GET /api/v4/projects/42/merge_requests/7/notes"] = `[{"id": 3, "body": "` + Marker + ` old"}]`
	if done, err := Upsert(pr, Marker+" new plan"); err != nil || done != "updated" {
		t.Fatalf("%s %v", done, err)
	}
	if !strings.Contains(f.bodies["PUT /api/v4/projects/42/merge_requests/7/notes/3"], "new plan") {
		t.Fatalf("bodies %v", f.bodies)
	}
}
