package release

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/monosmith-app/monosmith/apps/smith/githubrelease"
)

// fakeGitHub is a GitHub in memory: releases by tag, their assets, and a log
// of every request, so a test can say what was asked and in which order.
type fakeGitHub struct {
	mu       sync.Mutex
	releases map[string]*githubrelease.Release
	content  map[int64]string
	nextID   int64
	// log holds "METHOD path" of every request, in order.
	log []string
	// uploads holds "tag/name" of every uploaded asset, in order.
	uploads []string
	// deleted counts the deleted assets.
	deleted int
	// failUpload, when it returns true for an asset name, answers 500 once.
	failUpload func(name string) bool
}

// newFakeGitHub starts the fake and points the environment smith reads at it.
func newFakeGitHub(t *testing.T) *fakeGitHub {
	t.Helper()
	f := &fakeGitHub{releases: map[string]*githubrelease.Release{}, content: map[int64]string{}, nextID: 100}
	srv := httptest.NewServer(f)
	t.Cleanup(srv.Close)
	t.Setenv("GITHUB_API_URL", srv.URL)
	t.Setenv("GITHUB_TOKEN", "tok")
	t.Setenv("GH_TOKEN", "")
	t.Setenv("GITHUB_REPOSITORY", "o/r")
	t.Setenv("SMITH_RELEASE_ANY_BRANCH", "1")
	t.Setenv("TF_BUILD", "")
	t.Setenv("GITLAB_CI", "")
	return f
}

func (f *fakeGitHub) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.log = append(f.log, r.Method+" "+r.URL.Path)
	if r.URL.Path != "" && !strings.HasPrefix(r.URL.Path, "/upload/") && r.Header.Get("Authorization") != "Bearer tok" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	switch {
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/repos/o/r/releases/tags/"):
		tag := strings.TrimPrefix(r.URL.Path, "/repos/o/r/releases/tags/")
		rel, ok := f.releases[tag]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		rel.UploadURL = "http://" + r.Host + "/upload/" + tag + "{?name,label}"
		_ = json.NewEncoder(w).Encode(rel)
	case r.Method == http.MethodPost && r.URL.Path == "/repos/o/r/releases":
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		tag, _ := body["tag_name"].(string)
		rel := &githubrelease.Release{ID: int64(len(f.releases) + 1), HTMLURL: "https://github.com/o/r/releases/tag/" + tag}
		rel.UploadURL = "http://" + r.Host + "/upload/" + tag + "{?name,label}"
		f.releases[tag] = rel
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(rel)
	case r.Method == http.MethodPost && strings.HasPrefix(r.URL.Path, "/upload/"):
		tag := strings.TrimPrefix(r.URL.Path, "/upload/")
		name := r.URL.Query().Get("name")
		if f.failUpload != nil && f.failUpload(name) {
			f.failUpload = nil
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		raw, _ := io.ReadAll(r.Body)
		f.nextID++
		f.content[f.nextID] = string(raw)
		f.releases[tag].Assets = append(f.releases[tag].Assets, githubrelease.Asset{ID: f.nextID, Name: name, Size: int64(len(raw))})
		f.uploads = append(f.uploads, tag+"/"+name)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{}`))
	case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/repos/o/r/releases/assets/"):
		var id int64
		_, _ = fmt.Sscanf(strings.TrimPrefix(r.URL.Path, "/repos/o/r/releases/assets/"), "%d", &id)
		_, _ = w.Write([]byte(f.content[id]))
	case r.Method == http.MethodDelete && strings.HasPrefix(r.URL.Path, "/repos/o/r/releases/assets/"):
		var id int64
		_, _ = fmt.Sscanf(strings.TrimPrefix(r.URL.Path, "/repos/o/r/releases/assets/"), "%d", &id)
		for _, rel := range f.releases {
			kept := rel.Assets[:0]
			for _, a := range rel.Assets {
				if a.ID != id {
					kept = append(kept, a)
				}
			}
			rel.Assets = kept
		}
		f.deleted++
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusTeapot)
	}
}

// assetNames lists the names attached to a release, in attach order.
func (f *fakeGitHub) assetNames(tag string) []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []string
	if rel, ok := f.releases[tag]; ok {
		for _, a := range rel.Assets {
			out = append(out, a.Name)
		}
	}
	return out
}

// body is the content of an attached asset.
func (f *fakeGitHub) body(tag, name string) string {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, a := range f.releases[tag].Assets {
		if a.Name == name {
			return f.content[a.ID]
		}
	}
	return ""
}

// requests is how many requests the fake has answered.
func (f *fakeGitHub) requests() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.log)
}
